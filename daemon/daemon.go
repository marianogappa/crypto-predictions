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
	ErrOnlyTwitterPredictionActioningSupported = errors.New("only Twitter-based prediction actioning is supported")
	ErrPredictionAlreadyActioned               = errors.New("prediction has already been actioned")
	ErrUnkownActionType                        = errors.New("unknown action type")
)

type Daemon struct {
	store            statestorage.StateStorage
	market           market.IMarket
	predImageBuilder imagebuilder.PredictionImageBuilder
	enableTweeting   bool
	enableReplying   bool

	errs []error
}

func NewDaemon(market market.IMarket, store statestorage.StateStorage, imgBuilder imagebuilder.PredictionImageBuilder, enableTweeting, enableReplying bool) *Daemon {
	return &Daemon{store: store, market: market, predImageBuilder: imgBuilder, enableTweeting: enableTweeting, enableReplying: enableReplying}
}

func (r *Daemon) BlockinglyRunEvery(dur time.Duration) {
	log.Info().Msgf("Daemon scheduler started and will run again every: %v", dur)
	for {
		r.Run(int(time.Now().Unix()))
		time.Sleep(dur)
	}
}

func (r *Daemon) Run(nowTs int) []error {
	r.errs = []error{}
	var (
		predictionsScanner = NewEvolvablePredictionsScanner(r.store)
		prediction         types.Prediction
	)

	for predictionsScanner.Scan(&prediction) {
		r.MaybeActionPredictionCreated(prediction, nowTs)
		if ok := r.EvolvePrediction(&prediction, r.market, nowTs); !ok {
			continue
		}
		r.MaybeActionPredictionFinal(prediction, nowTs)
		r.StoreEvolvedPrediction(prediction)
	}
	r.AddErrs(nil, predictionsScanner.Error)

	log.Info().Msgf("Daemon.Run: finished with cache hit ratio of %.2f\n", r.market.(market.Market).CalculateCacheHitRatio())
	if len(r.errs) > 0 {
		log.Info().Errs("errs", r.errs).Msg("Daemon.Run: finished with errors")
	}

	return r.errs
}

func (r *Daemon) MaybeActionPredictionCreated(prediction types.Prediction, nowTs int) {
	if prediction.State.Status != types.UNSTARTED {
		return
	}
	err := r.ActionPrediction(prediction, actionTypePredictionCreated, nowTs)
	r.AddErrs(&prediction, err)
}

func (r *Daemon) EvolvePrediction(prediction *types.Prediction, m market.IMarket, nowTs int) bool {
	predRunner, errs := NewPredRunner(prediction, r.market, nowTs)
	r.AddErrs(prediction, errs...)
	if len(errs) > 0 {
		return false
	}
	errs = predRunner.Run(false)
	r.AddErrs(predRunner.prediction, errs...)

	return true
}

func (r *Daemon) MaybeActionPredictionFinal(prediction types.Prediction, nowTs int) {
	if !prediction.Evaluate().IsFinal() {
		return
	}

	err := r.store.LogPredictionStateValueChange(types.PredictionStateValueChange{
		PredictionUUID: prediction.UUID,
		StateValue:     prediction.State.Value.String(),
		CreatedAt:      types.ISO8601(time.Now().Format(time.RFC3339)),
	})
	r.AddErrs(&prediction, err)

	description := printer.NewPredictionPrettyPrinter(prediction).Default()
	log.Info().Msgf("Prediction just finished: [%v] with value [%v]!\n", description, prediction.State.Value)

	if prediction.State.Value == types.CORRECT || prediction.State.Value == types.INCORRECT {
		err := r.ActionPrediction(prediction, actionTypeBecameFinal, nowTs)
		r.AddErrs(&prediction, err)
	}
}

func (r *Daemon) StoreEvolvedPrediction(prediction types.Prediction) {
	_, err := r.store.UpsertPredictions([]*types.Prediction{&prediction})
	r.AddErrs(&prediction, err)
}

func (r *Daemon) AddErrs(prediction *types.Prediction, errs ...error) {
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
