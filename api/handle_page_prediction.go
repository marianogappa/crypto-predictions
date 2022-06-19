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
		return failWith(ErrPredictionNotFound, ErrPredictionNotFound, apiResGetPagesPrediction{})
	}
	pred := preds[0]

	ps := compiler.NewPredictionSerializer(&a.mkt)
	mainCompilerPred, err := ps.PreSerializeForAPI(&pred, true)
	if err != nil {
		return failWith(ErrFailedToSerializePredictions, err, apiResGetPagesPrediction{})
	}
	predictionsByUUID := map[UUID]compiler.Prediction{}
	predictionsByUUID[UUID(mainCompilerPred.UUID)] = mainCompilerPred

	// Latest Predictions
	latest10Predictions, err := a.store.GetPredictions(types.APIFilters{}, []string{types.CREATED_AT_DESC.String()}, 10, 0)
	if err != nil {
		return failWith(ErrStorageErrorRetrievingPredictions, err, apiResGetPagesPrediction{})
	}
	latestPredictionUUIDs := []UUID{}
	for _, prediction := range latest10Predictions {
		predictionUUID := prediction.UUID
		if predictionUUID == mainCompilerPred.UUID {
			continue
		}
		latestPredictionUUIDs = append(latestPredictionUUIDs, UUID(predictionUUID))
		compilerPr, err := ps.PreSerializeForAPI(&prediction, false)
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
			if predictionUUID == mainCompilerPred.UUID {
				continue
			}
			latestPredictionSameAuthorURL = append(latestPredictionSameAuthorURL, UUID(predictionUUID))
			compilerPr, err := ps.PreSerializeForAPI(&prediction, false)
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
		if predictionUUID == mainCompilerPred.UUID {
			continue
		}
		latestPredictionSameCoinUUID = append(latestPredictionSameCoinUUID, UUID(predictionUUID))
		compilerPr, err := ps.PreSerializeForAPI(&prediction, false)
		if err != nil {
			return failWith(ErrFailedToSerializePredictions, err, apiResGetPagesPrediction{})
		}
		predictionsByUUID[UUID(predictionUUID)] = compilerPr
	}

	accountURLSet := map[URL]struct{}{}

	// Get the top 10 Accounts by follower count
	topAccounts, err := a.store.GetAccounts(types.APIAccountFilters{}, []string{types.ACCOUNT_FOLLOWER_COUNT_DESC.String()}, 10, 0)
	if err != nil {
		return failWith(types.ErrStorageErrorRetrievingAccounts, err, apiResGetPagesPrediction{})
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
		return failWith(types.ErrStorageErrorRetrievingAccounts, err, apiResGetPagesPrediction{})
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
		Prediction:                    UUID(mainCompilerPred.UUID),
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

To learn about the prediction's schema, please review the <i>GET /predictions</i> documentation. This notably contains info about the "summary" property, which is necessary to make a candlestick chart for the main prediction.

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

// UUID type exists only to provide examples on the Swagger docs
type UUID string

// PrepareJSONSchema provides examples on the Swagger docs
func (UUID) PrepareJSONSchema(schema *jsonschema.Schema) error {
	schema.WithFormat("uuid").WithExamples("c6ea7af5-0a29-48ae-a6cc-271545b3a53c", "1129a8c0-0189-491c-be4e-4a0aec5eeb23")

	return nil
}

// URL type exists only to provide examples on the Swagger docs
type URL string

// PrepareJSONSchema provides examples on the Swagger docs
func (URL) PrepareJSONSchema(schema *jsonschema.Schema) error {
	schema.WithFormat("uri").WithExamples("https://twitter.com/Sheldon_Sniper")

	return nil
}
