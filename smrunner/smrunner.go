package smrunner

import (
	"errors"
	"log"
	"time"

	"github.com/marianogappa/predictions/market"
	"github.com/marianogappa/predictions/statestorage"
	"github.com/marianogappa/predictions/types"
)

type SMRunner struct {
	store  statestorage.StateStorage
	market market.IMarket
}

type SMRunnerResult struct {
	Errors      []error
	Predictions map[string]types.Prediction
}

func NewSMRunner(market market.IMarket, store statestorage.StateStorage) *SMRunner {
	return &SMRunner{store: store, market: market}
}

func (r *SMRunner) BlockinglyRunEvery(dur time.Duration) SMRunnerResult {
	for {
		result := r.Run(int(time.Now().Unix()))
		if len(result.Errors) > 0 {
			log.Println("State Machine runner finished with errors:")
			for i, err := range result.Errors {
				log.Printf("%v) %v", i+1, err.Error())
			}
			log.Println()
		}
		time.Sleep(dur)
	}
}

func (r *SMRunner) Run(nowTs int) SMRunnerResult {
	var result = SMRunnerResult{Predictions: map[string]types.Prediction{}, Errors: []error{}}

	// Get ongoing predictions from storage
	predictions, err := r.store.GetPredictions(
		types.APIFilters{
			PredictionStateValues: []string{
				types.ONGOING_PRE_PREDICTION.String(),
				types.ONGOING_PREDICTION.String(),
			},
		},
		[]string{types.CREATED_AT_DESC.String()},
	)
	if err != nil {
		result.Errors = append(result.Errors, err)
		return result
	}

	// Create prediction runners from all ongoing predictions
	activePredRunners := map[string]*PredRunner{}
	for _, prediction := range predictions {
		pred := prediction
		predRunner, errs := NewPredRunner(&pred, r.market, nowTs)
		for _, err := range errs {
			if !errors.Is(err, errPredictionAtFinalStateAtCreation) {
				result.Errors = append(result.Errors, err)
			}
		}
		if len(errs) > 0 {
			continue
		}
		activePredRunners[prediction.UUID] = predRunner
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
				result.Predictions[pk] = *predRunner.prediction
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
