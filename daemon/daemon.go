package daemon

import (
	"errors"
	"log"
	"time"

	"github.com/marianogappa/predictions/market"
	"github.com/marianogappa/predictions/printer"
	"github.com/marianogappa/predictions/statestorage"
	"github.com/marianogappa/predictions/types"
)

type Daemon struct {
	store  statestorage.StateStorage
	market market.IMarket
}

type DaemonResult struct {
	Errors      []error
	Predictions map[string]types.Prediction
}

func NewDaemon(market market.IMarket, store statestorage.StateStorage) *Daemon {
	return &Daemon{store: store, market: market}
}

func (r *Daemon) BlockinglyRunEvery(dur time.Duration) DaemonResult {
	for {
		result := r.Run(int(time.Now().Unix()))
		if len(result.Errors) > 0 {
			log.Println("Daemon run finished with errors:")
			for i, err := range result.Errors {
				log.Printf("%v) %v", i+1, err.Error())
			}
			log.Println()
		}
		time.Sleep(dur)
	}
}

func (r *Daemon) Run(nowTs int) DaemonResult {
	var result = DaemonResult{Predictions: map[string]types.Prediction{}, Errors: []error{}}

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
		log.Printf("Daemon.Run: %v active prediction runners\n", len(activePredRunners))
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

	for _, inactivePrediction := range result.Predictions {
		if inactivePrediction.Evaluate().IsFinal() {
			description := printer.NewPredictionPrettyPrinter(inactivePrediction).Default()
			log.Printf("Prediction just finished: [%v] with value [%v]!\n", description, inactivePrediction.State.Value)
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
