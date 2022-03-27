package daemon

import (
	"errors"
	"fmt"
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
	Predictions []*types.Prediction
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

	// log.Printf("Daemon.Run: %v active prediction runners\n", len(predRunners))
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
			r.store.LogPredictionStateValueChange(types.PredictionStateValueChange{
				PredictionUUID: prediction.UUID,
				StateValue:     prediction.State.Value.String(),
				CreatedAt:      types.ISO8601(time.Now().Format(time.RFC3339)),
			})
			description := printer.NewPredictionPrettyPrinter(*prediction).Default()
			log.Printf("Prediction just finished: [%v] with value [%v]!\n", description, prediction.State.Value)
		}
	}

	log.Printf("Daemon.Run: finished with cache hit ratio of %.2f\n", r.market.(market.Market).CalculateCacheHitRatio())

	// Upsert state with changed predictions
	_, err = r.store.UpsertPredictions(result.Predictions)
	if err != nil {
		result.Errors = append(result.Errors, err)
		return result
	}
	return result
}
