package api

import (
	"context"

	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/types"
	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/usecase"
)

type apiResGetPredictions struct {
	Predictions []compilerPrediction `json:"predictions"`
	_           struct{}             `query:"_" additionalProperties:"false"`
}

type apiReqGetPredictions struct {
	Tags                  []string `json:"tags" query:"tags" description:"Currently, predictions are tagged with the coins they are for, e.g. BINANCE:COIN:BTC-USDT"`
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
	_                     struct{} `query:"_" additionalProperties:"false"`
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
	res := []compilerPrediction{}
	for _, pred := range preds {
		compPred, err := ps.PreSerialize(&pred)
		if err != nil {
			if err != nil {
				return failWith(ErrFailedToSerializePredictions, err, apiResGetPredictions{})
			}
		}
		res = append(res, compilerPrediction(compPred))
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
	u.SetDescription(`Returns an ordered array of predictions based on the configured filters and ordering. If no query string is provided, all predictions are returned.

All date times throughout the API (except for candlestick timestamps) are ISO8601 (or RFC3339) strings.

The prediction schema contains the following properties:

- uuid: uniquely identifies the prediction. For now, predictions are also unique by <i>postUrl</i>. It's auto-generated.
- version: for now, it's always <i>1.0.0</i>. It's auto-generated.
- createdAt: datetime in which the prediction was reported into this system, not to be confused with the datetime the author of the social network post actually posted it (which is <i>postedAt</i>). It's auto-generated.
- reporter: username of the user who reported this prediction. For now, always set it to <i>admin</i>.
- postAuthor: username or handle of the social network user who posted this prediction. It's auto-generated.
- postAuthorURL: url to the profile of the social network user who posted this prediction. It's auto-generated.
- postedAt: datetime the author of the social network post posted it. It's auto-generated.
- postUrl: url to the social network post in which this prediction was made.
- given: the map of conditions declared in this prediction.
- prePredict: sometimes, a prediction has two steps, in which case the <i>prePredict</i> is step one and <i>predict</i> is step two.
- predict: the actual prediction, stated as a boolean algebra of conditions described in <i>given</i>, e.g. "(a and b and (not c)) or d".
- state: the current state of the prediction, which includes whether it's UNSTARTED/STARTED/FINISHED, ONGOING_PRE_PREDICTION/ONGOING_PREDICTION/CORRECT/INCORRECT/ANNULLED and the last candlestick timestamp analyzed for it.
- type: an enum value identifying the "type of prediction". There are currently 4 types of predictions. You can read more about them on the <i>GET /pages/prediction</i> documentation.
- predictionText: a human-readable text describing the prediction.`)
	u.SetTitle("Main API call for getting predictions based on filters and ordering.")
	return u
}

type compilerPrediction compiler.Prediction

func (compilerPrediction) PrepareJSONSchema(schema *jsonschema.Schema) error {
	schema.WithExamples(compilerPrediction{
		UUID:          "1f4bf3e4-2e5d-49d0-9bc0-d46f338a5d1b",
		Version:       "1.0.0",
		CreatedAt:     "2022-03-25T13:33:54Z",
		Reporter:      "admin",
		PostAuthor:    "rovercrc",
		PostAuthorURL: "",
		PostedAt:      "2022-03-24T15:02:48.000Z",
		PostUrl:       "https://twitter.com/rovercrc/status/1507010047737405444",
		Given: map[string]compiler.Condition{
			"a": {
				Condition:   "COIN:BINANCE:BTC-USDT >= 29000",
				FromISO8601: "2022-03-24T15:02:48.000Z",
				ToISO8601:   "2023-01-01T00:00:00.000Z",
				ToDuration:  "eoy",
				State: compiler.ConditionState{
					Status: "STARTED",
					LastTs: 1651162692,
					LastTicks: map[string]types.Tick{
						"COIN:BINANCE:BTC-USDT": {
							Timestamp: 1651162692,
							Value:     41000,
						},
					},
					Value: "ONGOING_PREDICTION",
				},
				ErrorMarginRatio: 0.03,
			},
		},
		Predict: compiler.Predict{Predict: "a"},
		PredictionState: compiler.PredictionState{
			Status: "STARTED",
			LastTs: 1651162692,
			Value:  "ONGOING_PREDICTION",
		},
		Type:           "PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE",
		PredictionText: "COIN:BINANCE:BTC-USDT will hit 29000 by end of year",
	})

	return nil
}
