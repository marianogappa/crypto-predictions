package daemon

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/marianogappa/predictions/metadatafetcher/twitter"
	"github.com/marianogappa/predictions/printer"
	"github.com/marianogappa/predictions/types"
	"github.com/rs/zerolog/log"
)

type ActionType int

const (
	ACTION_TYPE_BECAME_FINAL ActionType = iota
	ACTION_TYPE_PREDICTION_CREATED
)

func (a ActionType) String() string {
	switch a {
	case ACTION_TYPE_BECAME_FINAL:
		return "ACTION_TYPE_BECAME_FINAL"
	case ACTION_TYPE_PREDICTION_CREATED:
		return "ACTION_TYPE_PREDICTION_CREATED"
	default:
		return "ACTION_TYPE_UNKNOWN"
	}
}

func (r *Daemon) ActionPrediction(prediction types.Prediction, actionType ActionType, nowTs int) error {
	// TODO eventually we want to action Youtube predictions as well, possibly by just not replying to a tweet.
	if !strings.HasPrefix(prediction.PostUrl, "https://twitter.com/") {
		return ErrOnlyTwitterPredictionActioningSupported
	}
	if actionType != ACTION_TYPE_BECAME_FINAL && actionType != ACTION_TYPE_PREDICTION_CREATED {
		return ErrUnkownActionType
	}
	if prediction.State.LastTs > nowTs {
		return fmt.Errorf("Daemon.ActionPrediction: now() seems to be newer than the prediction lastTs (%v)...that shouldn't happen. Bailing.", time.Unix(int64(prediction.State.LastTs), 0).Format(time.RFC1123))
	}
	if nowTs-prediction.State.LastTs > 60*60*24 {
		return fmt.Errorf("Daemon.ActionPrediction: prediction's lastTs is older than 24hs, so I won't action it anymore.")
	}
	exists, err := r.store.PredictionInteractionExists(prediction.UUID, prediction.PostUrl, actionType.String())
	if err != nil {
		return err
	}
	if exists {
		return ErrPredictionAlreadyActioned
	}
	if prediction.PostAuthorURL == "" {
		return errors.New("Daemon.ActionPrediction: prediction has no PostAuthorURL, so I cannot make an image!")
	}
	accounts, err := r.store.GetAccounts(types.APIAccountFilters{URLs: []string{prediction.PostAuthorURL}}, nil, 1, 0)
	if err != nil {
		return fmt.Errorf("%w: %v", types.ErrStorageErrorRetrievingAccounts, err)
	}
	if len(accounts) == 0 {
		return fmt.Errorf("Daemon.ActionPrediction: there is no account for %v", prediction.PostAuthorURL)
	}
	account := accounts[0]

	tweetURL := ""
	switch actionType {
	case ACTION_TYPE_BECAME_FINAL:
		tweetURL, err = r.tweetActionBecameFinal(prediction, account)
	case ACTION_TYPE_PREDICTION_CREATED:
		tweetURL, err = r.tweetActionPredictionCreated(prediction, account)
	}
	if err != nil {
		return err
	}

	if err := r.store.InsertPredictionInteraction(prediction.UUID, prediction.PostUrl, actionType.String(), tweetURL); err != nil {
		return err
	}
	return nil
}

func (r *Daemon) tweetActionBecameFinal(prediction types.Prediction, account types.Account) (string, error) {
	twitter := twitter.NewTwitter("")

	description := printer.NewPredictionPrettyPrinter(prediction).Default()
	postAuthor := prediction.PostAuthor
	predictionStateValue := prediction.State.Value.String()

	imageURL, err := r.predImageBuilder.BuildImage(prediction, account)
	if err != nil {
		log.Error().Err(err).Msg("Daemon.tweetActionBecameFinal: silently ignoring error with building image...")
	}
	if imageURL != "" {
		defer os.Remove(imageURL)
	}

	// TODO eventually we'll need a reply tweet id here
	tweetURL, err := twitter.Tweet(fmt.Sprintf(`%v made the following prediction: "%v" and it just became %v!`, postAuthor, description, predictionStateValue), imageURL, 0)
	if err != nil {
		return "", err
	}

	return tweetURL, nil
}

func (r *Daemon) tweetActionPredictionCreated(prediction types.Prediction, account types.Account) (string, error) {
	twitter := twitter.NewTwitter("")

	description := printer.NewPredictionPrettyPrinter(prediction).Default()
	postAuthor := prediction.PostAuthor

	imageURL, err := r.predImageBuilder.BuildImage(prediction, account)
	if err != nil {
		log.Error().Err(err).Msg("Daemon.tweetActionBecameFinal: silently ignoring error with building image...")
	}
	if imageURL != "" {
		defer os.Remove(imageURL)
	}

	// TODO eventually we'll need a reply tweet id here
	tweetURL, err := twitter.Tweet(fmt.Sprintf(`Now tracking the following prediction made by %v: "%v"`, postAuthor, description), imageURL, 0)
	if err != nil {
		return "", err
	}

	return tweetURL, err
}
