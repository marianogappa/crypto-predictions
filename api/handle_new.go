package api

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/daemon"
	"github.com/marianogappa/predictions/types"
)

type newBody struct {
	Store      bool            `json:"store"`
	Prediction json.RawMessage `json:"prediction"`
}

func (a *API) handleNew(newBody newBody) (response, error) {
	pc := compiler.NewPredictionCompiler(&a.mFetcher, a.NowFunc)
	pred, account, err := pc.Compile(newBody.Prediction)
	if err != nil {
		return response{}, fmt.Errorf("%w: %v", ErrFailedToCompilePrediction, err)
	}

	// If the state is empty, run one tick to see if the prediction is decided at start time. If so, it's invalid.
	if pred.State == (types.PredictionState{}) {
		predRunner, errs := daemon.NewPredRunner(&pred, a.mkt, int(a.NowFunc().Unix()))
		if len(errs) == 0 {
			predRunnerErrs := predRunner.Run(true)
			for _, err := range predRunnerErrs {
				if errors.Is(err, types.ErrInvalidMarketPair) {
					return response{}, types.ErrInvalidMarketPair
				}
			}
			if pred.Evaluate().IsFinal() {
				return response{}, types.ErrPredictionFinishedAtStartTime
			}
		}
	}

	if newBody.Store {
		// N.B. as per interface, UpsertPredictions may add UUIDs in-place on predictions
		_, err = a.store.UpsertPredictions([]*types.Prediction{&pred})
		if err != nil {
			return response{}, fmt.Errorf("%w: %v", ErrStorageErrorStoringPrediction, err)
		}

		if account != nil {
			_, err := a.store.UpsertAccounts([]*types.Account{account})
			if err != nil {
				return response{}, fmt.Errorf("%w: %v", ErrStorageErrorStoringPrediction, err)
			}
		}
	}

	bs, err := compiler.NewPredictionSerializer().Serialize(&pred)
	if err != nil {
		return response{}, fmt.Errorf("%w: %v", ErrFailedToSerializePredictions, err)
	}
	raw := json.RawMessage(bs)

	return response{pred: &raw, stored: pBool(newBody.Store)}, nil
}

func pBool(b bool) *bool {
	return &b
}
