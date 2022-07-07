package api

import (
	"context"
	"fmt"

	"github.com/marianogappa/predictions/core"
	"github.com/marianogappa/predictions/statestorage"
	"github.com/swaggest/usecase"
)

type apiReqUUIDPath struct {
	UUID string   `path:"uuid" required:"true" format:"uuid"`
	_    struct{} `query:"_" additionalProperties:"false"`
}
type apiResStored struct {
	Stored bool     `json:"stored" required:"true"`
	_      struct{} `query:"_" additionalProperties:"false"`
}

func (a *API) predictionStorageActionWithUUID(uuid string, fn func(string) error) apiResponse[apiResStored] {
	pred, errResp := getPredictionByUUID(uuid, a.store, apiResStored{})
	if errResp != nil {
		return *errResp
	}

	if err := fn(pred.UUID); err != nil {
		return failWith(ErrFailedToCompilePrediction, err, apiResStored{})
	}
	return apiResponse[apiResStored]{Status: 200, Data: apiResStored{Stored: true}}
}

func (a *API) apiPredictionStorageActionWithUUID(fn func(string) error, title string) usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, input apiReqUUIDPath, output *apiResponse[apiResStored]) error {
		out := a.predictionStorageActionWithUUID(input.UUID, fn)
		*output = out
		return nil
	})
	u.SetTags("Prediction")
	u.SetTitle(title)

	return u
}

func (a *API) apiPredictionRefetchAccount() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, input apiReqUUIDPath, output *apiResponse[apiResStored]) error {
		out := a.predictionRefetchAccount(input.UUID)
		*output = out
		return nil
	})
	u.SetTags("Prediction")
	u.SetTitle("Refetch a social media Account's metadata (e.g. name, thumbnails, isVerified, followerCount).")
	u.SetDescription("All predictions are made by an Account. The account is fetched when the prediction is created. After some time it may get outdated. Use this call to refetch it.")

	return u
}

func (a *API) predictionRefetchAccount(uuid string) apiResponse[apiResStored] {
	pred, errResp := getPredictionByUUID(uuid, a.store, apiResStored{})
	if errResp != nil {
		return *errResp
	}

	metadata, err := a.mFetcher.Fetch(pred.PostURL)
	if err != nil {
		return failWith(ErrFailedToCompilePrediction, fmt.Errorf("%w: error fetching metadata for url: %v", ErrFailedToCompilePrediction, pred.PostURL), apiResStored{})
	}

	if _, err := a.store.UpsertAccounts([]*core.Account{&metadata.Author}); err != nil {
		return failWith(ErrStorageErrorStoringAccount, fmt.Errorf("%w: error storing account: %v", ErrStorageErrorStoringAccount, err), apiResStored{})
	}

	return apiResponse[apiResStored]{Status: 200, Data: apiResStored{Stored: true}}
}

func (a *API) apiPredictionClearState() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, input apiReqUUIDPath, output *apiResponse[apiResStored]) error {
		out := a.predictionClearState(input.UUID)
		*output = out
		return nil
	})
	u.SetTags("Prediction")
	u.SetTitle("Start evolving the prediction from scratch, as if created just now.")

	return u
}

func (a *API) predictionClearState(uuid string) apiResponse[apiResStored] {
	pred, errResp := getPredictionByUUID(uuid, a.store, apiResStored{})
	if errResp != nil {
		return *errResp
	}

	pred.ClearState()
	if _, err := a.store.UpsertPredictions([]*core.Prediction{&pred}); err != nil {
		return failWith(ErrStorageErrorStoringPrediction, fmt.Errorf("%w: error storing prediction: %v", ErrStorageErrorStoringPrediction, err), apiResStored{})
	}
	return apiResponse[apiResStored]{Status: 200, Data: apiResStored{Stored: true}}
}

func getPredictionByUUIDOrURL[D any](uuid, url string, store statestorage.StateStorage, zero D) (core.Prediction, *apiResponse[D]) {
	if uuid != "" {
		return getPredictionByFilter(core.APIFilters{UUIDs: []string{uuid}, URLs: []string{}}, store, zero)
	}
	return getPredictionByFilter(core.APIFilters{UUIDs: []string{}, URLs: []string{url}}, store, zero)
}

func getPredictionByUUID[D any](uuid string, store statestorage.StateStorage, zero D) (core.Prediction, *apiResponse[D]) {
	return getPredictionByFilter(core.APIFilters{UUIDs: []string{uuid}}, store, zero)
}

func getPredictionByFilter[D any](filter core.APIFilters, store statestorage.StateStorage, zero D) (core.Prediction, *apiResponse[D]) {
	ps, err := store.GetPredictions(filter, nil, 0, 0)
	if err != nil {
		errResp := failWith(ErrPredictionNotFound, err, zero)
		return core.Prediction{}, &errResp
	}
	if len(ps) == 0 {
		errResp := failWith(ErrPredictionNotFound, ErrPredictionNotFound, zero)
		return core.Prediction{}, &errResp
	}
	if len(ps) != 1 {
		errResp := failWith(ErrFailedToCompilePrediction, fmt.Errorf("%w: expected to find exactly one prediction but found %v", ErrFailedToCompilePrediction, len(ps)), zero)
		return core.Prediction{}, &errResp
	}
	return ps[0], nil
}
