package daemon

import (
	"errors"
	"fmt"
	"time"

	"github.com/marianogappa/crypto-candles/candles"
	"github.com/marianogappa/predictions/core"
	"github.com/marianogappa/predictions/imagebuilder"
	"github.com/marianogappa/predictions/printer"
	"github.com/marianogappa/predictions/statestorage"
	"github.com/rs/zerolog/log"
)

var (
	// ErrUIUnsupportedPredictionType means: this prediction type is not supported in the UI
	ErrUIUnsupportedPredictionType = errors.New("this prediction type is not supported in the UI")

	// ErrOnlyTwitterPredictionActioningSupported returned when a Youtube-based prediction is triggered to be actioned.
	ErrOnlyTwitterPredictionActioningSupported = errors.New("only Twitter-based prediction actioning is supported")

	// ErrPredictionAlreadyActioned returned when the prediction_interactions table has an entry for that (uuid, actionType).
	ErrPredictionAlreadyActioned = errors.New("prediction has already been actioned")

	// ErrUnkownActionType returned when an action is triggered with an unknown actionType.
	ErrUnkownActionType = errors.New("unknown action type")
)

// Daemon is the main struct for the Daemon component.
type Daemon struct {
	store            statestorage.StateStorage
	market           core.IMarket
	predImageBuilder imagebuilder.PredictionImageBuilder
	enableTweeting   bool
	enableReplying   bool
	websiteURL       string

	errs []error
}

// NewDaemon is the constructor for the Daemon component.
func NewDaemon(market core.IMarket, store statestorage.StateStorage, imgBuilder imagebuilder.PredictionImageBuilder, enableTweeting, enableReplying bool, websiteURL string) *Daemon {
	return &Daemon{store: store, market: market, predImageBuilder: imgBuilder, enableTweeting: enableTweeting, enableReplying: enableReplying, websiteURL: websiteURL}
}

// BlockinglyRunEvery does infinite Daemon runs, separated by a specified time.Sleep.
func (r *Daemon) BlockinglyRunEvery(dur time.Duration) {
	log.Info().Msgf("Daemon scheduler started and will run again every: %v", dur)
	for {
		r.Run(int(time.Now().Unix()))
		time.Sleep(dur)
	}
}

// Run sequentially evolves all evolvable predictions.
func (r *Daemon) Run(nowTs int) []error {
	r.errs = []error{}
	var (
		predictionsScanner = statestorage.NewEvolvablePredictionsScanner(r.store)
		prediction         core.Prediction
	)

	for predictionsScanner.Scan(&prediction) {
		r.maybeActionPredictionCreated(prediction, nowTs)
		r.evolvePrediction(&prediction, r.market, nowTs)
		r.maybeActionPredictionFinal(prediction, nowTs)
		r.storeEvolvedPrediction(prediction)
	}
	r.addErrs(nil, predictionsScanner.Error)

	r.ActionPendingInteractions(time.Now)

	log.Info().Msgf("Daemon.Run: finished with cache hit ratio of %.2f\n", r.market.(candles.Market).CalculateCacheHitRatio())
	if len(r.errs) > 0 {
		log.Info().Errs("errs", r.errs).Msg("Daemon.Run: finished with errors")
	}

	return r.errs
}

func (r *Daemon) maybeActionPredictionCreated(prediction core.Prediction, nowTs int) {
	if prediction.State.Status != core.UNSTARTED {
		return
	}
	err := r.store.LogPredictionStateValueChange(core.PredictionStateValueChange{
		PredictionUUID: prediction.UUID,
		StateValue:     prediction.State.Value.String(),
		CreatedAt:      core.ISO8601(time.Unix(int64(nowTs), 0).Format(time.RFC3339)),
	})
	r.addErrs(&prediction, err)

	err = r.store.InsertPredictionInteraction(core.PredictionInteraction{
		PostURL:        prediction.PostURL,
		PredictionUUID: prediction.UUID,
		ActionType:     actionTypePredictionCreated.String(),
		Status:         "PENDING",
	})
	r.addErrs(&prediction, err)
}

func (r *Daemon) evolvePrediction(prediction *core.Prediction, m core.IMarket, nowTs int) {
	predRunner, errs := NewPredEvolver(prediction, r.market, nowTs)
	r.addErrs(prediction, errs...)
	if len(errs) > 0 {
		return
	}
	errs = predRunner.Run(false)
	r.addErrs(predRunner.prediction, errs...)
}

func (r *Daemon) maybeActionPredictionFinal(prediction core.Prediction, nowTs int) {
	if !prediction.Evaluate().IsFinal() {
		return
	}

	err := r.store.LogPredictionStateValueChange(core.PredictionStateValueChange{
		PredictionUUID: prediction.UUID,
		StateValue:     prediction.State.Value.String(),
		CreatedAt:      core.ISO8601(time.Unix(int64(nowTs), 0).Format(time.RFC3339)),
	})
	r.addErrs(&prediction, err)

	description := printer.NewPredictionPrettyPrinter(prediction).String()
	log.Info().Msgf("Prediction just finished: [%v] with value [%v]!\n", description, prediction.State.Value)

	// TODO this will have to change for prediction types where ANNULLED is a possible final state
	if prediction.State.Value == core.CORRECT || prediction.State.Value == core.INCORRECT {
		err := r.store.InsertPredictionInteraction(core.PredictionInteraction{
			PostURL:        prediction.PostURL,
			PredictionUUID: prediction.UUID,
			ActionType:     actionTypeBecameFinal.String(),
			Status:         "PENDING",
		})
		r.addErrs(&prediction, err)
	}
}

func (r *Daemon) storeEvolvedPrediction(prediction core.Prediction) {
	_, err := r.store.UpsertPredictions([]*core.Prediction{&prediction})
	r.addErrs(&prediction, err)
}

func (r *Daemon) addErrs(prediction *core.Prediction, errs ...error) {
	for _, err := range errs {
		if err == nil {
			continue
		}
		if errors.Is(err, errPredictionAtFinalStateAtCreation) {
			continue
		}
		if prediction == nil {
			r.errs = append(r.errs, err)
		} else {
			r.errs = append(r.errs, fmt.Errorf("for %v: %w", prediction.UUID, err))
		}
	}
}
