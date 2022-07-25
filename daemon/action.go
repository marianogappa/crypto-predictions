package daemon

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/marianogappa/predictions/core"
	"github.com/marianogappa/predictions/metadatafetcher/twitter"
	"github.com/marianogappa/predictions/printer"
	"github.com/rs/zerolog/log"
)

type actionType int

const (
	actionTypeUnknown actionType = iota
	actionTypeBecameFinal
	actionTypePredictionCreated
)

var (
	// ErrTweetingDisabled is returned when Daemon.ActionPrediction is called but tweeting is disabled
	ErrTweetingDisabled = errors.New("tweeting is not enabled for the daemon; to enable set the PREDICTIONS_DAEMON_ENABLE_TWEETING env to any value")
)

// actionTypeFromString constructs an actionType from a String
func actionTypeFromString(s string) actionType {
	switch s {
	case "ACTION_TYPE_BECAME_FINAL":
		return actionTypeBecameFinal
	case "ACTION_TYPE_PREDICTION_CREATED":
		return actionTypePredictionCreated
	default:
		return actionTypeUnknown
	}
}
func (a actionType) String() string {
	switch a {
	case actionTypeBecameFinal:
		return "ACTION_TYPE_BECAME_FINAL"
	case actionTypePredictionCreated:
		return "ACTION_TYPE_PREDICTION_CREATED"
	default:
		return "ACTION_TYPE_UNKNOWN"
	}
}

// ActionPendingInteractions actions all pending social media interactions for all predictions that change status.
func (r *Daemon) ActionPendingInteractions(timeNowFunc func() time.Time) error {
	pendingInteractions, err := r.store.GetPendingPredictionInteractions()
	if err != nil {
		log.Error().Err(err).Msgf("Daemon.ActionPendingInteractions: error actioning pending interactions.")
		return err
	}
	if len(pendingInteractions) > 3 {
		// Maximum of 3 action interactions per Daemon run, to not spam the Twitter handle with posts.
		pendingInteractions = pendingInteractions[:3]
	}
	for _, interaction := range pendingInteractions {
		tweetURL, err := r.ActionPendingInteraction(interaction, timeNowFunc)

		interaction.Status = "POSTED"
		interaction.Error = ""
		interaction.InteractionPostURL = tweetURL
		if err != nil {
			interaction.Status = "ERROR"
			interaction.Error = err.Error()
		}

		if err := r.store.UpdatePredictionInteractionStatus(interaction); err != nil {
			log.Error().Err(err).Msg("Daemon.ActionPendingInteractions: error actioning pending interaction...ignoring.")
		}
	}
	return nil
}

// ActionPendingInteraction actions a pending social media interaction.
func (r *Daemon) ActionPendingInteraction(interaction core.PredictionInteraction, timeNowFunc func() time.Time) (string, error) {
	preds, err := r.store.GetPredictions(core.APIFilters{UUIDs: []string{interaction.PredictionUUID}}, []string{}, 1, 0)
	if err != nil {
		log.Error().Err(err).Msgf("Daemon.ActionPendingInteractions: error actioning pending interactions.")
		return "", err
	}
	if len(preds) == 0 {
		log.Error().Msgf("Daemon.ActionPendingInteractions: could not find prediction for interaction.")
		return "", fmt.Errorf("could not find prediction for interaction with UUID: [%v]", interaction.PredictionUUID)
	}
	actType := actionTypeFromString(interaction.ActionType)
	if actType == actionTypeUnknown {
		log.Error().Msgf("Daemon.ActionPendingInteractions: unknown action type.")
		return "", fmt.Errorf("unknown action type: [%v]", interaction.ActionType)
	}

	return r.ActionPrediction(preds[0], actType, int(timeNowFunc().Unix()))
}

// ActionPrediction currently tweets when a prediction is created and when it becomes correct or incorrect.
//
// TODO: this should be extracted into a separate PredictionPublisher component that takes Twitter, Store & Market,
// because BackOffice will probably end up using it, which means API needs to run it.
func (r *Daemon) ActionPrediction(prediction core.Prediction, actType actionType, nowTs int) (string, error) {
	if !r.enableTweeting {
		return "", ErrTweetingDisabled
	}
	// TODO eventually we want to action Youtube predictions as well, possibly by just not replying to a tweet.
	if !strings.HasPrefix(prediction.PostURL, "https://twitter.com/") {
		return "", ErrOnlyTwitterPredictionActioningSupported
	}
	if core.UIUnsupportedPredictionTypes[prediction.Type] {
		return "", ErrUIUnsupportedPredictionType
	}
	if actType != actionTypeBecameFinal && actType != actionTypePredictionCreated {
		return "", ErrUnkownActionType
	}
	// TODO this will have to change for prediction types where ANNULLED is a possible final state
	if actType == actionTypeBecameFinal && prediction.State.Value != core.CORRECT && prediction.State.Value != core.INCORRECT {
		return "", fmt.Errorf("daemon.ActionPrediction: prediction is not CORRECT nor INCORRECT, and I was asked to action ACTION_TYPE_BECAME_FINAL")
	}
	if prediction.State.LastTs > nowTs {
		return "", fmt.Errorf("daemon.ActionPrediction: now() seems to be newer than the prediction lastTs (%v)...that shouldn't happen, bailing", time.Unix(int64(prediction.State.LastTs), 0).Format(time.RFC1123))
	}
	if nowTs-prediction.State.LastTs > 60*60*24 {
		return "", fmt.Errorf("daemon.ActionPrediction: prediction's lastTs is older than 24hs, so I won't action it anymore")
	}
	interaction := core.PredictionInteraction{PredictionUUID: prediction.UUID, PostURL: prediction.PostURL, ActionType: actType.String()}
	exists, err := r.store.NonPendingPredictionInteractionExists(interaction)
	if err != nil {
		return "", err
	}
	if exists {
		return "", ErrPredictionAlreadyActioned
	}
	if prediction.PostAuthorURL == "" {
		return "", errors.New("daemon.ActionPrediction: prediction has no PostAuthorURL, so I cannot make an image")
	}
	accounts, err := r.store.GetAccounts(core.APIAccountFilters{URLs: []string{prediction.PostAuthorURL}}, nil, 1, 0)
	if err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrStorageErrorRetrievingAccounts, err)
	}
	if len(accounts) == 0 {
		return "", fmt.Errorf("daemon.ActionPrediction: there is no account for %v", prediction.PostAuthorURL)
	}
	account := accounts[0]

	tweetURL := ""
	switch actType {
	case actionTypeBecameFinal:
		tweetURL, err = r.tweetActionBecameFinal(prediction, account)
	case actionTypePredictionCreated:
		tweetURL, err = r.tweetActionPredictionCreated(prediction, account)
	}
	if err != nil {
		return "", err
	}
	return tweetURL, nil
}

func (r *Daemon) tweetActionBecameFinal(prediction core.Prediction, account core.Account) (string, error) {
	var urlPart string
	if r.websiteURL != "" {
		urlPart = fmt.Sprintf("\n\nSee it here: %v/p/{POST_UUID}\n ${BASE_ASSET} {HASHTAG}", r.websiteURL)
	}
	var (
		description      = printer.NewPredictionPrettyPrinter(prediction).String()
		predictionResult = map[core.PredictionStateValue]string{
			core.CORRECT:   "CORRECT ✅",
			core.INCORRECT: "INCORRECT ❌",
		}[prediction.State.Value]

		text = fmt.Sprintf("Prediction by {HANDLE} became %v!\n\n\"%v\"%v", predictionResult, description, urlPart)
	)
	return r.doTweet(text, prediction, account)
}

func (r *Daemon) tweetActionPredictionCreated(prediction core.Prediction, account core.Account) (string, error) {
	var urlPart string
	if r.websiteURL != "" {
		urlPart = fmt.Sprintf("\n\nFollow it here: %v/p/{POST_UUID}\n ${BASE_ASSET} {HASHTAG}", r.websiteURL)
	}

	var (
		description = printer.NewPredictionPrettyPrinter(prediction).String()
		text        = fmt.Sprintf("👀 Now tracking prediction by {HANDLE} 👀\n\n\"%v\"%v", description, urlPart)
	)
	return r.doTweet(text, prediction, account)
}

func (r *Daemon) doTweet(text string, prediction core.Prediction, account core.Account) (string, error) {
	imageURL, err := r.predImageBuilder.BuildImage(prediction, account)
	if err != nil {
		log.Error().Err(err).Msg("Daemon.tweetActionBecameFinal: silently ignoring error with building image...")
	}
	if imageURL != "" {
		defer os.Remove(imageURL)
	}

	inReplyToStatusID, err := getStatusIDFromTweetURL(prediction.PostURL)
	handle := fmt.Sprintf("@%v", account.Handle)

	if !r.enableReplying || err != nil || account.Handle == "" {
		inReplyToStatusID = 0
		handle = prediction.PostAuthor
	}

	text = strings.Replace(text, "{HANDLE}", handle, -1)
	text = strings.Replace(text, "{POST_URL}", url.QueryEscape(prediction.PostURL), -1)
	text = strings.Replace(text, "{POST_UUID}", url.QueryEscape(prediction.UUID), -1)
	text = strings.Replace(text, "{BASE_ASSET}", prediction.CalculateMainCoin().BaseAsset, -1)
	text = strings.Replace(text, "{HASHTAG}", calculateHashtag(prediction), -1)

	tweetURL, err := twitter.NewTwitter("").Tweet(text, imageURL, inReplyToStatusID)
	if err != nil {
		return "", err
	}
	return tweetURL, nil
}

func calculateHashtag(prediction core.Prediction) string {
	hashtagFunctions := []func(core.Prediction) string{
		func(p core.Prediction) string { return "#crypto" },
		func(p core.Prediction) string { return "#CryptoTwitter" },
		func(p core.Prediction) string { return "#cryptocurrency" },
		func(p core.Prediction) string { return "#investing" },
		func(p core.Prediction) string { return "#finance" },
	}

	return hashtagFunctions[int(prediction.UUID[len(prediction.UUID)-1])%len(hashtagFunctions)](prediction)
}

func getStatusIDFromTweetURL(url string) (int, error) {
	rxStatusID := regexp.MustCompile("^https://twitter.com/[^/]+/status/([0-9]+)$")
	matches := rxStatusID.FindStringSubmatch(url)
	if len(matches) != 2 {
		return 0, errors.New("getStatusIDFromTweetURL: url did not pass regex check")
	}
	id, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("getStatusIDFromTweetURL: %v", err)
	}
	return id, nil
}
