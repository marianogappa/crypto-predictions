package api

import (
	"context"
	"fmt"

	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/types"
	"github.com/swaggest/usecase"
)

type apiResMaintenance struct {
	Success bool     `json:"success"`
	_       struct{} `query:"_" additionalProperties:"false"`
}

type apiReqMaintenance struct {
	Action string   `path:"action"`
	_      struct{} `query:"_" additionalProperties:"false"`
}

func (a *API) maintenance(req apiReqMaintenance) apiResponse[apiResMaintenance] {
	preds, err := a.store.GetPredictions(
		types.APIFilters{},
		[]string{},
		0, 0,
	)
	if err != nil {
		return failWith(ErrStorageErrorRetrievingPredictions, err, apiResMaintenance{})
	}

	// Fetch and upsert all accounts
	for i, pred := range preds {
		// Re-compile prediction, this time with metadatafetcher, which will create an account and add additional fields to prediction
		ps := compiler.NewPredictionSerializer()
		serialised, err := ps.Serialize(&pred)
		if err != nil {
			return failWith(ErrFailedToSerializePredictions, fmt.Errorf("%w: error serializing prediction: %v", ErrFailedToSerializePredictions, err), apiResMaintenance{})
		}

		pc := compiler.NewPredictionCompiler(nil, nil)
		newPred, _, err := pc.Compile(serialised)
		if err != nil {
			return failWith(ErrFailedToCompilePrediction, fmt.Errorf("%w: error compiling prediction: %v", ErrFailedToSerializePredictions, err), apiResMaintenance{})
		}

		preds[i] = newPred
	}

	ps := []*types.Prediction{}
	for i := range preds {
		ps = append(ps, &preds[i])
	}
	a.store.UpsertPredictions(ps)

	return apiResponse[apiResMaintenance]{Status: 200, Data: apiResMaintenance{Success: true}}
}

func (a *API) apiMaintenance() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, input apiReqMaintenance, output *apiResponse[apiResMaintenance]) error {
		out := a.maintenance(input)
		*output = out
		return nil
	})
	u.SetTags("Maintenance")
	u.SetDescription("")
	u.SetTitle("Perform maintenance operations.")
	return u
}
