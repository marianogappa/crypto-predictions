package smrunner

import (
	"log"

	"github.com/marianogappa/predictions/market"
	"github.com/marianogappa/predictions/statestorage"
	"github.com/marianogappa/predictions/types"
)

type SMRunner struct {
	store  statestorage.StateStorage
	market market.Market
}

type SMRunnerResult struct {
	Errors      []error
	Predictions map[string]types.Prediction
}

func NewSMRunner(store statestorage.StateStorage, market market.Market) *SMRunner {
	return &SMRunner{store: store, market: market}
}

func (r *SMRunner) Run(nowTs int) SMRunnerResult {
	var result = SMRunnerResult{Predictions: map[string]types.Prediction{}, Errors: []error{}}

	// Get ongoing predictions from storage
	predictions, err := r.store.GetPredictions([]types.PredictionStateValue{types.ONGOING_PRE_PREDICTION, types.ONGOING_PREDICTION})
	if err != nil {
		result.Errors = append(result.Errors, err)
		return result
	}

	// Create prediction runners from all ongoing predictions
	activePredRunners := map[string]*predRunner{}
	for pk, prediction := range predictions {
		predRunner, errs := newPredRunner(prediction, r.market, nowTs)
		if len(errs) > 0 {
			result.Errors = append(result.Errors, errs...)
			continue
		}
		activePredRunners[pk] = predRunner
	}

	// Continuously run prediction runners until there aren't any active ones
	for len(activePredRunners) > 0 {
		log.Printf("SMRunner.Run: %v active prediction runners\n", len(activePredRunners))
		for pk, predRunner := range activePredRunners {
			errs := predRunner.Run()
			if len(errs) > 0 {
				result.Errors = append(result.Errors, errs...)
			}
			if predRunner.isInactive {
				result.Predictions[pk] = predRunner.prediction
				delete(activePredRunners, pk)
			}
		}
	}

	// Upsert state with changed predictions
	err = r.store.UpsertPredictions(result.Predictions)
	if err != nil {
		result.Errors = append(result.Errors, err)
		return result
	}
	return result
}
