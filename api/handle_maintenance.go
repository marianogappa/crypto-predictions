package api

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/metadatafetcher"
	"github.com/marianogappa/predictions/types"
	"github.com/swaggest/usecase"
)

type apiResMaintenance struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	_       struct{} `query:"_" additionalProperties:"false"`
}

type apiReqMaintenance struct {
	Action string   `path:"action"`
	_      struct{} `query:"_" additionalProperties:"false"`
}

func (a *API) maintenance(req apiReqMaintenance) apiResponse[apiResMaintenance] {
	switch req.Action {
	case "ensureAllPredictionsHavePostAuthorURL":
		return a.ensureAllPredictionsHavePostAuthorURL(req)
	default:
		return apiResponse[apiResMaintenance]{Status: 400, Data: apiResMaintenance{Success: false, Message: "action does not exist"}}
	}
}

func (a *API) ensureAllPredictionsHavePostAuthorURL(req apiReqMaintenance) apiResponse[apiResMaintenance] {
	preds, err := a.store.GetPredictions(
		types.APIFilters{},
		[]string{},
		0, 0,
	)
	if err != nil {
		return failWith(ErrStorageErrorRetrievingPredictions, err, apiResMaintenance{})
	}

	// Fetch and upsert all accounts
	predsToUpdate := []*types.Prediction{}
	for _, pred := range preds {
		if pred.PostAuthorURL != "" {
			continue
		}
		// Re-compile prediction, this time with metadatafetcher, which will create an account and add additional fields to prediction
		ps := compiler.NewPredictionSerializer()
		serialised, err := ps.Serialize(&pred)
		if err != nil {
			return failWith(ErrFailedToSerializePredictions, fmt.Errorf("%w: error serializing prediction: %v", ErrFailedToSerializePredictions, err), apiResMaintenance{})
		}

		metadataFetcher := metadatafetcher.NewMetadataFetcher()

		pc := compiler.NewPredictionCompiler(metadataFetcher, time.Now)
		newPred, _, err := pc.Compile(serialised)
		if err != nil {
			return failWith(ErrFailedToCompilePrediction, fmt.Errorf("%w: error compiling prediction: %v", ErrFailedToSerializePredictions, err), apiResMaintenance{})
		}

		if newPred.PostAuthorURL == "" {
			failWith(ErrFailedToCompilePrediction, fmt.Errorf("%w: metadata fetcher could not resolve postAuthorURL from postURL: %v", ErrFailedToCompilePrediction, pred.PostUrl), apiResMaintenance{})
		}

		log.Println(newPred.PostAuthorURL)

		predsToUpdate = append(predsToUpdate, &newPred)
	}

	_, err = a.store.UpsertPredictions(predsToUpdate)
	if err != nil {
		failWith(ErrStorageErrorStoringPrediction, fmt.Errorf("%w: failed to upsert predictions: %v", ErrStorageErrorStoringPrediction, err), apiResMaintenance{})
	}

	msg := fmt.Sprintf("Upserted %v new postAuthorURLs!", len(predsToUpdate))

	return apiResponse[apiResMaintenance]{Status: 200, Data: apiResMaintenance{Success: true, Message: msg}}
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
