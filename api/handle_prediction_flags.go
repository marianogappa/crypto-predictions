package api

import (
	"context"
	"fmt"

	"github.com/marianogappa/predictions/types"
	"github.com/swaggest/usecase"
)

type apiReqUuidPath struct {
	UUID string   `path:"uuid" required:"true" format:"uuid"`
	_    struct{} `query:"_" additionalProperties:"false"`
}
type apiResStored struct {
	Stored bool     `json:"stored" required:"true"`
	_      struct{} `query:"_" additionalProperties:"false"`
}

func (a *API) predictionStorageActionWithUUID(uuid string, fn func(string) error) apiResponse[apiResStored] {
	ps, err := a.store.GetPredictions(types.APIFilters{UUIDs: []string{uuid}}, nil, 0, 0)
	if err != nil {
		return failWith(ErrPredictionNotFound, err, apiResStored{})
	}
	if len(ps) == 0 {
		return failWith(ErrPredictionNotFound, ErrPredictionNotFound, apiResStored{})
	}

	if len(ps) != 1 {
		return failWith(ErrFailedToCompilePrediction, fmt.Errorf("%w: expected to find exactly one prediction but found %v", ErrFailedToCompilePrediction, len(ps)), apiResStored{})
	}
	pred := ps[0]
	if err := fn(pred.UUID); err != nil {
		return failWith(ErrFailedToCompilePrediction, err, apiResStored{})
	}
	return apiResponse[apiResStored]{Status: 200, Data: apiResStored{Stored: true}}
}

func (a *API) apiPredictionStorageActionWithUUID(fn func(string) error, title string) usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, input apiReqUuidPath, output *apiResponse[apiResStored]) error {
		out := a.predictionStorageActionWithUUID(input.UUID, fn)
		*output = out
		return nil
	})
	u.SetTags("Prediction")
	u.SetTitle(title)

	return u
}

func (a *API) apiPredictionRefetchAccount() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, input apiReqUuidPath, output *apiResponse[apiResStored]) error {
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
	ps, err := a.store.GetPredictions(types.APIFilters{UUIDs: []string{uuid}}, nil, 0, 0)
	if err != nil {
		return failWith(ErrPredictionNotFound, err, apiResStored{})
	}
	if len(ps) == 0 {
		return failWith(ErrPredictionNotFound, ErrPredictionNotFound, apiResStored{})
	}

	if len(ps) != 1 {
		return failWith(ErrFailedToCompilePrediction, fmt.Errorf("%w: expected to find exactly one prediction but found %v", ErrFailedToCompilePrediction, len(ps)), apiResStored{})
	}
	pred := ps[0]

	metadata, err := a.mFetcher.Fetch(pred.PostUrl)
	if err != nil {
		return failWith(ErrFailedToCompilePrediction, fmt.Errorf("%w: error fetching metadata for url: %v", ErrFailedToCompilePrediction, pred.PostUrl), apiResStored{})
	}

	if _, err := a.store.UpsertAccounts([]*types.Account{&metadata.Author}); err != nil {
		return failWith(ErrStorageErrorStoringAccount, fmt.Errorf("%w: error storing account: %v", ErrStorageErrorStoringAccount, err), apiResStored{})
	}

	return apiResponse[apiResStored]{Status: 200, Data: apiResStored{Stored: true}}
}
