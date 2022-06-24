package daemon

import (
	"errors"
	"fmt"
	"time"

	"github.com/marianogappa/predictions/imagebuilder"
	"github.com/marianogappa/predictions/market"
	"github.com/marianogappa/predictions/printer"
	"github.com/marianogappa/predictions/statestorage"
	"github.com/marianogappa/predictions/types"
	"github.com/rs/zerolog/log"
)

var (
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
	market           market.IMarket
	predImageBuilder imagebuilder.PredictionImageBuilder
	enableTweeting   bool
	enableReplying   bool
	websiteURL       string

	errs []error
}

// NewDaemon is the constructor for the Daemon component.
func NewDaemon(market market.IMarket, store statestorage.StateStorage, imgBuilder imagebuilder.PredictionImageBuilder, enableTweeting, enableReplying bool, websiteURL string) *Daemon {
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
		prediction         types.Prediction
	)

	for predictionsScanner.Scan(&prediction) {
		r.maybeActionPredictionCreated(prediction, nowTs)
		r.evolvePrediction(&prediction, r.market, nowTs)
		r.maybeActionPredictionFinal(prediction, nowTs)
		r.storeEvolvedPrediction(prediction)
	}
	r.addErrs(nil, predictionsScanner.Error)

	log.Info().Msgf("Daemon.Run: finished with cache hit ratio of %.2f\n", r.market.(market.Market).CalculateCacheHitRatio())
	if len(r.errs) > 0 {
		log.Info().Errs("errs", r.errs).Msg("Daemon.Run: finished with errors")
	}

	return r.errs
}

func (r *Daemon) maybeActionPredictionCreated(prediction types.Prediction, nowTs int) {
	if prediction.State.Status != types.UNSTARTED {
		return
	}
	err := r.ActionPrediction(prediction, actionTypePredictionCreated, nowTs)
	r.addErrs(&prediction, err)
}

func (r *Daemon) evolvePrediction(prediction *types.Prediction, m market.IMarket, nowTs int) {
	predRunner, errs := NewPredEvolver(prediction, r.market, nowTs)
	r.addErrs(prediction, errs...)
	if len(errs) > 0 {
		return
	}
	errs = predRunner.Run(false)
	r.addErrs(predRunner.prediction, errs...)
}

func (r *Daemon) maybeActionPredictionFinal(prediction types.Prediction, nowTs int) {
	if !prediction.Evaluate().IsFinal() {
		return
	}

	err := r.store.LogPredictionStateValueChange(types.PredictionStateValueChange{
		PredictionUUID: prediction.UUID,
		StateValue:     prediction.State.Value.String(),
		CreatedAt:      types.ISO8601(time.Now().Format(time.RFC3339)),
	})
	r.addErrs(&prediction, err)

	description := printer.NewPredictionPrettyPrinter(prediction).Default()
	log.Info().Msgf("Prediction just finished: [%v] with value [%v]!\n", description, prediction.State.Value)

	if prediction.State.Value == types.CORRECT || prediction.State.Value == types.INCORRECT {
		err := r.ActionPrediction(prediction, actionTypeBecameFinal, nowTs)
		r.addErrs(&prediction, err)
	}
}

func (r *Daemon) storeEvolvedPrediction(prediction types.Prediction) {
	_, err := r.store.UpsertPredictions([]*types.Prediction{&prediction})
	r.addErrs(&prediction, err)
}

func (r *Daemon) addErrs(prediction *types.Prediction, errs ...error) {
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
