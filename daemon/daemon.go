package daemon

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/marianogappa/predictions/market"
	"github.com/marianogappa/predictions/metadatafetcher/twitter"
	"github.com/marianogappa/predictions/printer"
	"github.com/marianogappa/predictions/statestorage"
	"github.com/marianogappa/predictions/types"
	"github.com/rs/zerolog/log"
)

var (
	ErrOnlyTwitterPredictionActioningSupported = errors.New("only Twitter-based prediction actioning is supported")
	ErrPredictionAlreadyActioned               = errors.New("prediction has already been actioned")
	ErrUnkownActionType                        = errors.New("unknown action type")
)

type Daemon struct {
	store  statestorage.StateStorage
	market market.IMarket
}

type DaemonResult struct {
	Errors      []error
	Predictions []*types.Prediction
}

func NewDaemon(market market.IMarket, store statestorage.StateStorage) *Daemon {
	return &Daemon{store: store, market: market}
}

func (r *Daemon) BlockinglyRunEvery(dur time.Duration) DaemonResult {
	log.Info().Msgf("Daemon started and will run again every: %v", dur)
	for {
		result := r.Run(int(time.Now().Unix()))
		if len(result.Errors) > 0 {
			log.Info().Msg("Daemon run finished with errors:")
			for _, err := range result.Errors {
				log.Error().Err(err).Msg("")
			}
		}
		time.Sleep(dur)
	}
}

func pBool(b bool) *bool { return &b }

func (r *Daemon) Run(nowTs int) DaemonResult {
	var result = DaemonResult{Predictions: []*types.Prediction{}, Errors: []error{}}

	// Get ongoing predictions from storage
	predictions, err := r.store.GetPredictions(
		types.APIFilters{
			PredictionStateValues: []string{
				types.ONGOING_PRE_PREDICTION.String(),
				types.ONGOING_PREDICTION.String(),
			},
			Paused:  pBool(false),
			Deleted: pBool(false),
		},
		[]string{types.CREATED_AT_DESC.String()},
		0, 0,
	)
	if err != nil {
		result.Errors = append(result.Errors, err)
		return result
	}

	// Create prediction runners from all ongoing predictions
	predRunners := []*PredRunner{}
	for _, prediction := range predictions {
		pred := prediction
		if pred.State.Status == types.UNSTARTED {
			err := r.ActionPrediction(&pred, ACTION_TYPE_PREDICTION_CREATED)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("for %v: %w", pred.UUID, err))
			}
		}
		predRunner, errs := NewPredRunner(&pred, r.market, nowTs)
		for _, err := range errs {
			if !errors.Is(err, errPredictionAtFinalStateAtCreation) {
				result.Errors = append(result.Errors, fmt.Errorf("for %v: %w", pred.UUID, err))
			}
		}
		if len(errs) > 0 {
			continue
		}
		predRunners = append(predRunners, predRunner)
	}

	// log.Info().Msgf("Daemon.Run: %v active prediction runners\n", len(predRunners))
	for _, predRunner := range predRunners {
		if errs := predRunner.Run(false); len(errs) > 0 {
			for _, err := range errs {
				result.Errors = append(result.Errors, fmt.Errorf("for %v: %w", predRunner.prediction.UUID, err))
			}
		}
		result.Predictions = append(result.Predictions, predRunner.prediction)
	}

	for _, prediction := range result.Predictions {
		if prediction.Evaluate().IsFinal() {
			err := r.store.LogPredictionStateValueChange(types.PredictionStateValueChange{
				PredictionUUID: prediction.UUID,
				StateValue:     prediction.State.Value.String(),
				CreatedAt:      types.ISO8601(time.Now().Format(time.RFC3339)),
			})
			if err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("for %v: %w", prediction.UUID, err))
			}
			description := printer.NewPredictionPrettyPrinter(*prediction).Default()
			log.Info().Msgf("Prediction just finished: [%v] with value [%v]!\n", description, prediction.State.Value)
			if prediction.State.Value == types.CORRECT || prediction.State.Value == types.INCORRECT {
				err := r.ActionPrediction(prediction, ACTION_TYPE_BECAME_FINAL)
				if err != nil {
					result.Errors = append(result.Errors, fmt.Errorf("for %v: %w", prediction.UUID, err))
				}
			}
		}
	}

	log.Info().Msgf("Daemon.Run: finished with cache hit ratio of %.2f\n", r.market.(market.Market).CalculateCacheHitRatio())

	// Upsert state with changed predictions
	_, err = r.store.UpsertPredictions(result.Predictions)
	if err != nil {
		result.Errors = append(result.Errors, err)
		return result
	}
	return result
}

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

func (r *Daemon) ActionPrediction(prediction *types.Prediction, actionType ActionType) error {
	if !strings.HasPrefix(prediction.PostUrl, "https://twitter.com/") {
		return ErrOnlyTwitterPredictionActioningSupported
	}
	if actionType != ACTION_TYPE_BECAME_FINAL && actionType != ACTION_TYPE_PREDICTION_CREATED {
		return ErrUnkownActionType
	}
	exists, err := r.store.PredictionInteractionExists(prediction.UUID, prediction.PostUrl, actionType.String())
	if err != nil {
		return err
	}
	if exists {
		return ErrPredictionAlreadyActioned
	}

	tweetURL := ""
	switch actionType {
	case ACTION_TYPE_BECAME_FINAL:
		tweetURL, err = r.tweetActionBecameFinal(prediction)
	case ACTION_TYPE_PREDICTION_CREATED:
		tweetURL, err = r.tweetActionPredictionCreated(prediction)
	}
	if err != nil {
		return err
	}

	// TODO eventually we'll need an image & a reply tweet id here
	if err := r.store.InsertPredictionInteraction(prediction.UUID, prediction.PostUrl, actionType.String(), tweetURL); err != nil {
		return err
	}
	return nil
}

func (r *Daemon) tweetActionBecameFinal(prediction *types.Prediction) (string, error) {
	twitter := twitter.NewTwitter("")

	description := printer.NewPredictionPrettyPrinter(*prediction).Default()
	postAuthor := prediction.PostAuthor
	predictionStateValue := prediction.State.Value.String()

	tweetURL, err := twitter.Tweet(fmt.Sprintf(`%v made the following prediction: "%v" and it just became %v!`, postAuthor, description, predictionStateValue), "", 0)
	if err != nil {
		return "", err
	}

	return tweetURL, err
}

func (r *Daemon) tweetActionPredictionCreated(prediction *types.Prediction) (string, error) {
	twitter := twitter.NewTwitter("")

	description := printer.NewPredictionPrettyPrinter(*prediction).Default()
	postAuthor := prediction.PostAuthor

	tweetURL, err := twitter.Tweet(fmt.Sprintf(`Now tracking the following prediction made by %v: "%v"`, postAuthor, description), "", 0)
	if err != nil {
		return "", err
	}

	return tweetURL, err
}
