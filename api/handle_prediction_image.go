package api

import (
	"context"
	"errors"

	"github.com/marianogappa/predictions/types"
	"github.com/swaggest/usecase"
)

type apiResGetPredictionImage struct {
	Base64Image string   `json:"base64Image"`
	_           struct{} `query:"_" additionalProperties:"false"`
}

type apiReqGetPredictionImage struct {
	UUID string   `path:"uuid" format:"uuid" description:"e.g. 48e9a5db-29f6-48fc-b5af-cb97631bf05d" example:"97accda2-87bf-4417-9785-f18513e8421c"`
	_    struct{} `query:"_" additionalProperties:"false"`
}

func (a *API) getPredictionImage(uuid string) apiResponse[apiResGetPredictionImage] {
	preds, err := a.store.GetPredictions(
		types.APIFilters{UUIDs: []string{uuid}},
		[]string{},
		0, 0,
	)
	if err != nil {
		return failWith(ErrStorageErrorRetrievingPredictions, err, apiResGetPredictionImage{})
	}
	if len(preds) != 1 {
		return failWith(ErrPredictionNotFound, err, apiResGetPredictionImage{})
	}
	pred := preds[0]
	if pred.PostAuthorURL == "" {
		return failWith(types.ErrStorageErrorRetrievingAccounts, errors.New("prediction has no postAuthorURL"), apiResGetPredictionImage{})
	}

	accounts, err := a.store.GetAccounts(types.APIAccountFilters{URLs: []string{pred.PostAuthorURL}}, []string{}, 0, 0)
	if err != nil {
		return failWith(types.ErrStorageErrorRetrievingAccounts, err, apiResGetPredictionImage{})
	}
	if len(accounts) != 1 {
		return failWith(types.ErrStorageErrorRetrievingAccounts, err, apiResGetPredictionImage{})
	}
	account := accounts[0]

	base64Image, err := a.imageBuilder.BuildImageBase64(pred, account)
	if err != nil {
		return failWith(ErrStorageErrorRetrievingPredictions, err, apiResGetPredictionImage{})
	}

	return apiResponse[apiResGetPredictionImage]{Status: 200, Data: apiResGetPredictionImage{
		Base64Image: base64Image,
	}}
}

func (a *API) apiGetPredictionImage() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, input apiReqGetPredictionImage, output *apiResponse[apiResGetPredictionImage]) error {
		out := a.getPredictionImage(input.UUID)
		*output = out
		return nil
	})
	u.SetTags("Prediction")
	u.SetTitle("Build and return the base64 encoding of an image of the prediction status, which is optimised for Twitter.")

	return u
}
