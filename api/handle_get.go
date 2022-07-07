package api

import (
	"context"
	"strconv"

	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/market/common"
	"github.com/marianogappa/predictions/serializer"
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
	PredictionStateValues []string `json:"predictionStateValues" query:"predictionStateValues" description:""`
	PredictionStateStatus []string `json:"predictionStateStatus" query:"predictionStateStatus" description:""`
	Deleted               *bool    `json:"deleted" query:"deleted" description:""`
	Paused                *bool    `json:"paused" query:"paused" description:""`
	Hidden                *bool    `json:"hidden" query:"hidden" description:""`
	IncludeUIUnsupported  bool     `json:"includeUIUnsupported" query:"includeUIUnsupported" description:"by default, prediction types that have no UI support are not returned"`
	OrderBys              []string `json:"orderBys" query:"orderBys" description:"Order in which predictions are returned. Defaults to CREATED_AT_DESC."`
	Limit                 string   `json:"limit" query:"limit" description:"How many predictions to return"`
	Offset                string   `json:"offset" query:"offset" description:"From which prediction to return 'limit' predictions"`
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
		IncludeUIUnsupported:  req.IncludeUIUnsupported,
	}

	limit := 0
	offset := 0
	rLimit, err := strconv.Atoi(req.Limit)
	if err == nil && rLimit > 0 {
		limit = rLimit
		rOffset, err := strconv.Atoi(req.Offset)
		if err == nil && rOffset > 0 {
			offset = rOffset
		}
	}

	preds, err := a.store.GetPredictions(
		filters,
		req.OrderBys,
		limit, offset,
	)
	if err != nil {
		return failWith(ErrStorageErrorRetrievingPredictions, err, apiResGetPredictions{})
	}

	ps := serializer.NewPredictionSerializer(nil)
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
- type: an enum value identifying the "type of prediction". There are currently 4 types of predictions. You can read more about them below.
- predictionText: a human-readable text describing the prediction.
- summary: contains extra information about the prediction, necessary to plot a candlestick chart. Note that there are different types of predictions, which all have candlesticks, but depending on the type the extra information will be different, so implementations will need to switch based on the <i>predictionType</i>.

Currently available prediction types (and their properties) are:

- PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE: e.g. "Bitcoin >= 45k within 10 days", will provide <i>coin</i> (e.g. <i>BINANCE:COIN:BTC-USDT</i>), <i>operator</i> (e.g. <i>>=</i>), <i>goal</i> (e.g. <i>45000</i>), <i>deadline</i> (e.g. <i>2006-01-02T15:04:05Z07:00</i>).
- PREDICTION_TYPE_COIN_WILL_RANGE: e.g. "Bitcoin will range between 30k and 40k for 10 days", will provide <i>coin</i> (e.g. <i>BINANCE:COIN:BTC-USDT</i>), <i>rangeLow</i> (e.g. <i>30000</i>), <i>rangeHigh</i> (e.g. <i>40000</i>), <i>deadline</i> (e.g. <i>2006-01-02T15:04:05Z07:00</i>).
- PREDICTION_TYPE_COIN_WILL_REACH_BEFORE_IT_REACHES: e.g. "Bitcoin will reach 50k before it reaches 30k", will provide <i>coin</i> (e.g. <i>BINANCE:COIN:BTC-USDT</i>), <i>willReach</i> (e.g. <i>50000</i>), <i>beforeItReaches</i> (e.g. <i>30000</i>), <i>deadline</i> (e.g. <i>2006-01-02T15:04:05Z07:00</i>).
- PREDICTION_TYPE_THE_FLIPPENING: e.g. "Ethereum's Marketcap will flip Bitcoin's Marketcap by end of year", will provide <i>coin</i> (e.g. <i>MESSARI:MARKETCAP:BTC</i>), <i>otherCoin</i> (e.g. <i>MESSARI:MARKETCAP:ETH</i>), <i>deadline</i> (e.g. <i>2006-01-02T15:04:05Z07:00</i>).`)
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
		PostURL:       "https://twitter.com/rovercrc/status/1507010047737405444",
		Given: map[string]compiler.Condition{
			"a": {
				Condition:   "COIN:BINANCE:BTC-USDT >= 29000",
				FromISO8601: "2022-03-24T15:02:48.000Z",
				ToISO8601:   "2023-01-01T00:00:00.000Z",
				ToDuration:  "eoy",
				State: compiler.ConditionState{
					Status: "STARTED",
					LastTs: 1651162692,
					LastTicks: map[string]common.Tick{
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
