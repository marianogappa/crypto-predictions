package api

import (
	"context"

	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/types"
	"github.com/swaggest/usecase"
)

type apiResGetPredictions struct {
	Predictions []compiler.Prediction `json:"predictions"`
}

type apiReqGetPredictions struct {
	Tags                  []string `json:"tags" query:"tags" description:"hello!"`
	AuthorHandles         []string `json:"authorHandles" query:"authorHandles" description:"e.g. Datadash"`
	AuthorURLs            []string `json:"authorURLs" query:"authorURLs" description:"e.g. https://twitter.com/CryptoCapo_"`
	UUIDs                 []string `json:"uuids" query:"uuids" description:"These are the prediction's primary keys (although they are also unique by url)"`
	URLs                  []string `json:"urls" query:"urls" description:"e.g. https://twitter.com/VLoveIt2Hack/status/1465354862372298763"`
	PredictionStateValues []string `json:"predictionStateValues" query:"predictionStateValues" description:"hello!"`
	PredictionStateStatus []string `json:"predictionStateStatus" query:"predictionStateStatus" description:"hello!"`
	Deleted               *bool    `json:"deleted" query:"deleted" description:"hello!"`
	Paused                *bool    `json:"paused" query:"paused" description:"hello!"`
	Hidden                *bool    `json:"hidden" query:"hidden" description:"hello!"`
	OrderBys              []string `json:"orderBys" query:"orderBys" description:"Order in which predictions are returned. Defaults to CREATED_AT_DESC." enum:"CREATED_AT_DESC,CREATED_AT_ASC"`
	_                     bool     `additionalProperties:"false"`
}

func (a *API) getPredictions(req apiReqGetPredictions) apiResponse[apiResGetPredictions] {
	filters := types.APIFilters{
		Tags:                  req.Tags,
		AuthorHandles:         req.AuthorHandles,
		AuthorURLs:            req.AuthorURLs,
		UUIDs:                 req.UUIDs,
		URLs:                  req.URLs,
		PredictionStateValues: req.PredictionStateValues,
		PredictionStateStatus: req.PredictionStateStatus,
		Deleted:               req.Deleted,
		Paused:                req.Paused,
		Hidden:                req.Hidden,
	}

	preds, err := a.store.GetPredictions(
		filters,
		req.OrderBys,
		0, 0,
	)
	if err != nil {
		return failWith(ErrStorageErrorRetrievingPredictions, err, apiResGetPredictions{})
	}

	ps := compiler.NewPredictionSerializer()
	res := []compiler.Prediction{}
	for _, pred := range preds {
		compPred, err := ps.PreSerialize(&pred)
		if err != nil {
			if err != nil {
				return failWith(ErrFailedToSerializePredictions, err, apiResGetPredictions{})
			}
		}
		res = append(res, compPred)
	}

	return apiResponse[apiResGetPredictions]{Status: 200, Data: apiResGetPredictions{Predictions: res}}
}

func (a *API) apiGetPredictions() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, input apiReqGetPredictions, output *apiResponse[apiResGetPredictions]) error {
		out := a.getPredictions(input)
		*output = out
		return nil
	})
	u.SetTags("Prediction")
	u.SetDescription("")
	u.SetTitle("Main API call for getting predictions based on filters and ordering.")
	return u
}
