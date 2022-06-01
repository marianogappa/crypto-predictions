package daemon

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/marianogappa/predictions/metadatafetcher/twitter"
	"github.com/marianogappa/predictions/printer"
	"github.com/marianogappa/predictions/types"
	"github.com/rs/zerolog/log"
)

type actionType int

const (
	actionTypeBecameFinal actionType = iota
	actionTypePredictionCreated
)

var (
	// ErrTweetingDisabled is returned when Daemon.ActionPrediction is called but tweeting is disabled
	ErrTweetingDisabled = errors.New("tweeting is not enabled for the daemon; to enable set the PREDICTIONS_DAEMON_ENABLE_TWEETING env to any value")
)

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

// ActionPrediction currently tweets when a prediction is created and when it becomes correct or incorrect.
//
// TODO: this should be extracted into a separate PredictionPublisher component that takes Twitter, Store & Market,
// because BackOffice will probably end up using it, which means API needs to run it.
func (r *Daemon) ActionPrediction(prediction types.Prediction, actType actionType, nowTs int) error {
	if !r.enableTweeting {
		return ErrTweetingDisabled
	}
	// TODO eventually we want to action Youtube predictions as well, possibly by just not replying to a tweet.
	if !strings.HasPrefix(prediction.PostUrl, "https://twitter.com/") {
		return ErrOnlyTwitterPredictionActioningSupported
	}
	if actType != actionTypeBecameFinal && actType != actionTypePredictionCreated {
		return ErrUnkownActionType
	}
	if actType == actionTypeBecameFinal && prediction.State.Value != types.CORRECT && prediction.State.Value != types.INCORRECT {
		return fmt.Errorf("daemon.ActionPrediction: prediction is not CORRECT nor INCORRECT, and I was asked to action ACTION_TYPE_BECAME_FINAL")
	}
	if prediction.State.LastTs > nowTs {
		return fmt.Errorf("daemon.ActionPrediction: now() seems to be newer than the prediction lastTs (%v)...that shouldn't happen, bailing", time.Unix(int64(prediction.State.LastTs), 0).Format(time.RFC1123))
	}
	if nowTs-prediction.State.LastTs > 60*60*24 {
		return fmt.Errorf("daemon.ActionPrediction: prediction's lastTs is older than 24hs, so I won't action it anymore")
	}
	exists, err := r.store.PredictionInteractionExists(prediction.UUID, prediction.PostUrl, actType.String())
	if err != nil {
		return err
	}
	if exists {
		return ErrPredictionAlreadyActioned
	}
	if prediction.PostAuthorURL == "" {
		return errors.New("daemon.ActionPrediction: prediction has no PostAuthorURL, so I cannot make an image")
	}
	accounts, err := r.store.GetAccounts(types.APIAccountFilters{URLs: []string{prediction.PostAuthorURL}}, nil, 1, 0)
	if err != nil {
		return fmt.Errorf("%w: %v", types.ErrStorageErrorRetrievingAccounts, err)
	}
	if len(accounts) == 0 {
		return fmt.Errorf("daemon.ActionPrediction: there is no account for %v", prediction.PostAuthorURL)
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
		return err
	}

	if err := r.store.InsertPredictionInteraction(prediction.UUID, prediction.PostUrl, actType.String(), tweetURL); err != nil {
		return err
	}
	return nil
}

func (r *Daemon) tweetActionBecameFinal(prediction types.Prediction, account types.Account) (string, error) {
	var (
		description      = printer.NewPredictionPrettyPrinter(prediction).Default()
		predictionResult = map[types.PredictionStateValue]string{
			types.CORRECT:   "CORRECT ✅",
			types.INCORRECT: "INCORRECT ❌",
		}[prediction.State.Value]

		text = fmt.Sprintf("Prediction by %v became %v!\n\n\"%v\"", `"%v"`, predictionResult, description)
	)
	return r.doTweet(text, prediction, account)
}

func (r *Daemon) tweetActionPredictionCreated(prediction types.Prediction, account types.Account) (string, error) {
	var (
		description = printer.NewPredictionPrettyPrinter(prediction).Default()
		text        = fmt.Sprintf("👀 Now tracking prediction by %v 👀\n\n\"%v\"", `%v`, description)
	)
	return r.doTweet(text, prediction, account)
}

func (r *Daemon) doTweet(rawText string, prediction types.Prediction, account types.Account) (string, error) {
	imageURL, err := r.predImageBuilder.BuildImage(prediction, account)
	if err != nil {
		log.Error().Err(err).Msg("Daemon.tweetActionBecameFinal: silently ignoring error with building image...")
	}
	if imageURL != "" {
		defer os.Remove(imageURL)
	}

	inReplyToStatusID, err := getStatusIDFromTweetURL(prediction.PostUrl)
	handle := fmt.Sprintf("@%v", account.Handle)

	if !r.enableReplying || err != nil || account.Handle == "" {
		inReplyToStatusID = 0
		handle = prediction.PostAuthor
	}

	text := rawText
	if strings.Contains(text, "%v") {
		text = fmt.Sprintf(rawText, handle)
	}

	tweetURL, err := twitter.NewTwitter("").Tweet(text, imageURL, inReplyToStatusID)
	if err != nil {
		return "", err
	}
	return tweetURL, nil
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
