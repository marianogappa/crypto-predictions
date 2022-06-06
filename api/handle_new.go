package api

import (
	"context"
	"errors"

	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/daemon"
	"github.com/marianogappa/predictions/types"
	"github.com/rs/zerolog/log"
	"github.com/swaggest/usecase"
)

type apiReqPostPrediction struct {
	Prediction string   `json:"prediction" formValue:"prediction" required:"true"`
	Store      bool     `json:"store" formValue:"store" required:"true"`
	_          struct{} `query:"_" additionalProperties:"false"`
}

type apiResPostPrediction struct {
	Prediction *compiler.Prediction `json:"prediction"`
	Stored     bool                 `json:"stored"`
	_          struct{}             `query:"_" additionalProperties:"false"`
}

func (a *API) apiPostPrediction() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, input apiReqPostPrediction, output *apiResponse[apiResPostPrediction]) error {
		out := a.postPrediction(input)
		*output = out
		return nil
	})
	u.SetTags("Prediction")
	u.SetTitle("Main API call to create a prediction.")

	errs := []error{}
	for _, errContent := range errToResponse {
		errs = append(errs, errContent)
	}

	u.SetExpectedErrors(errs...)

	return u
}

func (a *API) postPrediction(req apiReqPostPrediction) apiResponse[apiResPostPrediction] {
	if a.debug {
		log.Info().Msgf("API.postPrediction: with request: %+v", req)
	}

	pc := compiler.NewPredictionCompiler(&a.mFetcher, a.NowFunc)
	pred, account, err := pc.Compile([]byte(req.Prediction))
	if err != nil {
		return failWith(ErrFailedToCompilePrediction, err, apiResPostPrediction{})
	}

	// If the state is empty, run one tick to see if the prediction is decided at start time. If so, it's invalid.
	if pred.State == (types.PredictionState{}) {
		predRunner, errs := daemon.NewPredEvolver(&pred, a.mkt, int(a.NowFunc().Unix()))
		if len(errs) == 0 {
			predRunnerErrs := predRunner.Run(true)
			for _, err := range predRunnerErrs {
				if errors.Is(err, types.ErrInvalidMarketPair) {
					return failWith(types.ErrInvalidMarketPair, err, apiResPostPrediction{})
				}
			}
			if pred.Evaluate().IsFinal() {
				return failWith(types.ErrPredictionFinishedAtStartTime, types.ErrPredictionFinishedAtStartTime, apiResPostPrediction{})
			}
			// The evaluation will set the initial state for the prediction, but we want the Daemon to pick it up
			// as UNSTARTED so that it will post the initial tweet, so let's clear the state.
			(&pred).ClearState()
		}
	}

	if req.Store {
		// N.B. as per interface, UpsertPredictions may add UUIDs in-place on predictions
		_, err = a.store.UpsertPredictions([]*types.Prediction{&pred})
		if err != nil {
			return failWith(ErrStorageErrorStoringPrediction, err, apiResPostPrediction{})
		}

		if account != nil {
			_, err := a.store.UpsertAccounts([]*types.Account{account})
			if err != nil {
				return failWith(ErrStorageErrorStoringPrediction, err, apiResPostPrediction{})
			}
		}
	}

	res, err := compiler.NewPredictionSerializer(nil).PreSerialize(&pred)
	if err != nil {
		return failWith(ErrFailedToSerializePredictions, err, apiResPostPrediction{})
	}

	return apiResponse[apiResPostPrediction]{Status: 200, Data: apiResPostPrediction{Prediction: &res, Stored: req.Store}}
}
