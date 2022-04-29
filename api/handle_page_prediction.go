package api

import (
	"context"

	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/types"
	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/usecase"
)

type apiResGetPagesPrediction struct {
	Prediction                    UUID                    `json:"prediction"`
	Summary                       PredictionSummary       `json:"summary"`
	Top10AccountsByFollowerCount  []URL                   `json:"top10AccountsByFollowerCount"`
	AccountsByURL                 mapOfCompilerAccounts   `json:"accountsByURL"`
	Latest10Predictions           []UUID                  `json:"latest10Predictions"`
	Latest5PredictionsSameAccount []UUID                  `json:"latest5PredictionsSameAccount"`
	Latest5PredictionsSameCoin    []UUID                  `json:"latest5PredictionsSameCoin"`
	PredictionsByUUID             mapOfCompilerPrediction `json:"predictionsByUUID"`
	_                             struct{}                `query:"_" additionalProperties:"false"`
}

type apiReqGetPagesPrediction struct {
	URL string   `path:"url" format:"uri" description:"e.g. https://twitter.com/VLoveIt2Hack/status/1465354862372298763"`
	_   struct{} `query:"_" additionalProperties:"false"`
}

func (a *API) getPagesPrediction(url string) apiResponse[apiResGetPagesPrediction] {
	preds, err := a.store.GetPredictions(
		types.APIFilters{URLs: []string{url}},
		[]string{},
		0, 0,
	)
	if err != nil {
		return failWith(ErrStorageErrorRetrievingPredictions, err, apiResGetPagesPrediction{})
	}
	if len(preds) != 1 {
		return failWith(ErrPredictionNotFound, err, apiResGetPagesPrediction{})
	}
	pred := preds[0]

	ps := compiler.NewPredictionSerializer()
	compilerPred, err := ps.PreSerializeForAPI(&pred)
	if err != nil {
		return failWith(ErrFailedToSerializePredictions, err, apiResGetPagesPrediction{})
	}
	predictionsByUUID := map[UUID]compiler.Prediction{}
	predictionsByUUID[UUID(compilerPred.UUID)] = compilerPred

	summary, err := a.BuildPredictionMarketSummary(pred)
	if err != nil {
		return failWith(ErrStorageErrorRetrievingPredictions, err, apiResGetPagesPrediction{})
	}

	// Latest Predictions
	latest10Predictions, err := a.store.GetPredictions(types.APIFilters{}, []string{types.CREATED_AT_DESC.String()}, 10, 0)
	if err != nil {
		return failWith(ErrStorageErrorRetrievingPredictions, err, apiResGetPagesPrediction{})
	}
	latestPredictionUUIDs := []UUID{}
	for _, prediction := range latest10Predictions {
		predictionUUID := prediction.UUID
		latestPredictionUUIDs = append(latestPredictionUUIDs, UUID(predictionUUID))
		compilerPr, err := ps.PreSerializeForAPI(&prediction)
		if err != nil {
			return failWith(ErrFailedToSerializePredictions, err, apiResGetPagesPrediction{})
		}
		predictionsByUUID[UUID(predictionUUID)] = compilerPr
	}

	// Latest Predictions by same author URL
	latestPredictionSameAuthorURL := []UUID{}
	if pred.PostAuthorURL != "" {
		latest5PredictionsSameAuthor, err := a.store.GetPredictions(types.APIFilters{AuthorURLs: []string{pred.PostAuthorURL}}, []string{types.CREATED_AT_DESC.String()}, 5, 0)
		if err != nil {
			return failWith(ErrStorageErrorRetrievingPredictions, err, apiResGetPagesPrediction{})
		}
		for _, prediction := range latest5PredictionsSameAuthor {
			predictionUUID := prediction.UUID
			latestPredictionSameAuthorURL = append(latestPredictionSameAuthorURL, UUID(predictionUUID))
			compilerPr, err := ps.PreSerializeForAPI(&prediction)
			if err != nil {
				return failWith(ErrFailedToSerializePredictions, err, apiResGetPagesPrediction{})
			}
			predictionsByUUID[UUID(predictionUUID)] = compilerPr
		}
	}

	// Latest Predictions by same coin
	latestPredictionSameCoinUUID := []UUID{}
	latest5PredictionsSameCoin, err := a.store.GetPredictions(types.APIFilters{Tags: []string{pred.CalculateMainCoin().Str}}, []string{types.CREATED_AT_DESC.String()}, 5, 0)
	if err != nil {
		return failWith(ErrStorageErrorRetrievingPredictions, err, apiResGetPagesPrediction{})
	}
	for _, prediction := range latest5PredictionsSameCoin {
		predictionUUID := prediction.UUID
		latestPredictionSameCoinUUID = append(latestPredictionSameCoinUUID, UUID(predictionUUID))
		compilerPr, err := ps.PreSerializeForAPI(&prediction)
		if err != nil {
			return failWith(ErrFailedToSerializePredictions, err, apiResGetPagesPrediction{})
		}
		predictionsByUUID[UUID(predictionUUID)] = compilerPr
	}

	accountURLSet := map[URL]struct{}{}

	// Get the top 10 Accounts by follower count
	topAccounts, err := a.store.GetAccounts(types.APIAccountFilters{}, []string{types.ACCOUNT_FOLLOWER_COUNT_DESC.String()}, 10, 0)
	if err != nil {
		return failWith(ErrStorageErrorRetrievingAccounts, err, apiResGetPagesPrediction{})
	}
	top10AccountURLsByFollowerCount := []URL{}
	for _, account := range topAccounts {
		accountURL := account.URL.String()
		top10AccountURLsByFollowerCount = append(top10AccountURLsByFollowerCount, URL(accountURL))
		accountURLSet[URL(accountURL)] = struct{}{}
	}

	// Also gather all account urls from all predictions into the set
	for _, prediction := range predictionsByUUID {
		if prediction.PostAuthorURL == "" {
			continue
		}
		accountURLSet[URL(prediction.PostAuthorURL)] = struct{}{}
	}

	// Make the set a slice
	accountURLs := []string{}
	for accountURL := range accountURLSet {
		accountURLs = append(accountURLs, string(accountURL))
	}

	// Get all accounts from the slice
	allAccounts, err := a.store.GetAccounts(types.APIAccountFilters{URLs: accountURLs}, []string{}, 0, 0)
	if err != nil {
		return failWith(ErrStorageErrorRetrievingAccounts, err, apiResGetPagesPrediction{})
	}

	accountsByURL := map[URL]compiler.Account{}
	accountSerializer := compiler.NewAccountSerializer()
	for _, account := range allAccounts {
		accountURL := account.URL.String()
		compilerAcc, err := accountSerializer.PreSerialize(&account)
		if err != nil {
			return failWith(ErrFailedToSerializePredictions, err, apiResGetPagesPrediction{})
		}
		accountsByURL[URL(accountURL)] = compilerAcc
	}

	return apiResponse[apiResGetPagesPrediction]{Status: 200, Data: apiResGetPagesPrediction{
		Prediction:                    UUID(compilerPred.UUID),
		Summary:                       summary,
		Top10AccountsByFollowerCount:  top10AccountURLsByFollowerCount,
		Latest10Predictions:           latestPredictionUUIDs,
		Latest5PredictionsSameAccount: latestPredictionSameAuthorURL,
		Latest5PredictionsSameCoin:    latestPredictionSameCoinUUID,
		AccountsByURL:                 accountsByURL,
		PredictionsByUUID:             mapOfCompilerPrediction(predictionsByUUID),
	}}
}

func (a *API) apiGetPagesPrediction() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, input apiReqGetPagesPrediction, output *apiResponse[apiResGetPagesPrediction]) error {
		out := a.getPagesPrediction(input.URL)
		*output = out
		return nil
	})
	u.SetTags("Pages")
	u.SetTitle("Main API call to retrieve all info for the website page that shows a prediction.")
	u.SetDescription(`
This endpoint returns not only the prediction associated to the specified URL, but also a lot of other content that could be useful to be shown in the prediction page, e.g. latest predictions, top accounts (an "account" represents a social media account, which for now is either Twitter or Facebook).

Since there are potentially a lot of duplicate predictions & accounts in the response, there are two top-level objects: <i>predictionsByUUID</i> and <i>accountsByURL</i>, and everywhere else only UUIDs and URLs are used.

The top-level <i>prediction</i> property has the UUID of the prediction associated to the specified URL.

The following top-level objects contain an ordered array of prediction UUIDs or account URLs:

- top10AccountsByFollowerCount
- latest10Predictions
- latest5PredictionsSameAccount
- latest5PredictionsSameCoin

The <i>summary</i> top-level property contains extra information about the prediction associated to the specified URL that is not captured on the prediction itself, necessary to plot a candlestick chart. Note that there are different types of predictions, which all have candlesticks, but depending on the type the extra information will be different, so implementations will need to switch based on the <i>predictionType</i>.

Currently available prediction types (and their properties) are:

- PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE: e.g. "Bitcoin >= 45k within 10 days", will provide <i>coin</i> (e.g. <i>BINANCE:COIN:BTC-USDT</i>), <i>operator</i> (e.g. <i>>=</i>), <i>goal</i> (e.g. <i>45000</i>), <i>deadline</i> (e.g. <i>2006-01-02T15:04:05Z07:00</i>).
- PREDICTION_TYPE_COIN_WILL_RANGE: e.g. "Bitcoin will range between 30k and 40k for 10 days", will provide <i>coin</i> (e.g. <i>BINANCE:COIN:BTC-USDT</i>), <i>rangeLow</i> (e.g. <i>30000</i>), <i>rangeHigh</i> (e.g. <i>40000</i>), <i>deadline</i> (e.g. <i>2006-01-02T15:04:05Z07:00</i>).
- PREDICTION_TYPE_COIN_WILL_REACH_BEFORE_IT_REACHES: e.g. "Bitcoin will reach 50k before it reaches 30k", will provide <i>coin</i> (e.g. <i>BINANCE:COIN:BTC-USDT</i>), <i>willReach</i> (e.g. <i>50000</i>), <i>beforeItReaches</i> (e.g. <i>30000</i>), <i>deadline</i> (e.g. <i>2006-01-02T15:04:05Z07:00</i>).
- PREDICTION_TYPE_THE_FLIPPENING: e.g. "Ethereum's Marketcap will flip Bitcoin's Marketcap by end of year", will provide <i>coin</i> (e.g. <i>MESSARI:MARKETCAP:BTC</i>), <i>otherCoin</i> (e.g. <i>MESSARI:MARKETCAP:ETH</i>), <i>deadline</i> (e.g. <i>2006-01-02T15:04:05Z07:00</i>).

To learn about the prediction's schema, please review the <i>GET /predictions</i> documentation.

Gotchas for accounts:

- Youtube has Channel names, but there are no handles, so the handle field is empty. Both Twitter & Youtube do have names though. Accounts are unique by URL, which is always there.
- Both Youtube & Twitter have a verified feature, but Youtube doesn't expose it in the API, so for now this field is elided.
- The <i>created_at</> field for accounts is the time the social media account was created; not when we inserted it into our system.
	`)
	return u
}

// Examples for Swagger docs

type mapOfCompilerAccounts map[URL]compiler.Account

func (mapOfCompilerAccounts) PrepareJSONSchema(schema *jsonschema.Schema) error {
	schema.WithExamples(map[URL]compiler.Account{
		URL("https://twitter.com/Sheldon_Sniper"): {
			URL:           "https://twitter.com/Sheldon_Sniper",
			AccountType:   "TWITTER",
			Handle:        "Sheldon_Sniper",
			FollowerCount: 387033,
			Thumbnails:    []string{"https://pbs.twimg.com/profile_images/1480879644618510336/2iXc8iDk_normal.jpg", "https://pbs.twimg.com/profile_images/1480879644618510336/2iXc8iDk_400x400.jpg"},
			Name:          "Sheldon The Sniper",
			Description:   "",
			CreatedAt:     "2021-03-24T07:51:06Z",
		},
	})

	return nil
}

type mapOfCompilerPrediction map[UUID]compiler.Prediction

func (mapOfCompilerPrediction) PrepareJSONSchema(schema *jsonschema.Schema) error {
	schema.WithExamples(map[string]compiler.Prediction{
		"1f4bf3e4-2e5d-49d0-9bc0-d46f338a5d1b": {
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
		},
	})

	return nil
}

type UUID string

func (UUID) PrepareJSONSchema(schema *jsonschema.Schema) error {
	schema.WithFormat("uuid").WithExamples("c6ea7af5-0a29-48ae-a6cc-271545b3a53c", "1129a8c0-0189-491c-be4e-4a0aec5eeb23")

	return nil
}

type URL string

func (URL) PrepareJSONSchema(schema *jsonschema.Schema) error {
	schema.WithFormat("uri").WithExamples("https://twitter.com/Sheldon_Sniper")

	return nil
}
