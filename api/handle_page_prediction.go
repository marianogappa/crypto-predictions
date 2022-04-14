package api

import (
	"context"

	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/types"
	"github.com/swaggest/usecase"
)

type apiResGetPagesPrediction struct {
	Prediction                    compiler.Prediction            `json:"prediction"`
	Summary                       PredictionSummary              `json:"summary"`
	Top10AccountsByFollowerCount  []string                       `json:"top10AccountsByFollowerCount"`
	AccountsByURL                 map[string]compiler.Account    `json:"accountsByURL"`
	Latest10Predictions           []string                       `json:"latest10Predictions"`
	Latest5PredictionsSameAccount []string                       `json:"latest5PredictionsSameAccount"`
	Latest5PredictionsSameCoin    []string                       `json:"latest5PredictionsSameCoin"`
	PredictionsByUUID             map[string]compiler.Prediction `json:"predictionsByUUID"`
}

type apiReqGetPagesPrediction struct {
	URL string `path:"url" format:"uri" description:"e.g. https://twitter.com/VLoveIt2Hack/status/1465354862372298763"`
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

	summary, err := a.BuildPredictionMarketSummary(pred)
	if err != nil {
		return failWith(ErrStorageErrorRetrievingPredictions, err, apiResGetPagesPrediction{})
	}

	// Latest Predictions
	latest10Predictions, err := a.store.GetPredictions(types.APIFilters{}, []string{types.CREATED_AT_DESC.String()}, 10, 0)
	if err != nil {
		return failWith(ErrStorageErrorRetrievingPredictions, err, apiResGetPagesPrediction{})
	}
	latestPredictionUUIDs := []string{}
	predictionsByUUID := map[string]compiler.Prediction{}
	for _, prediction := range latest10Predictions {
		predictionUUID := prediction.UUID
		latestPredictionUUIDs = append(latestPredictionUUIDs, predictionUUID)
		compilerPr, err := ps.PreSerializeForAPI(&prediction)
		if err != nil {
			return failWith(ErrFailedToSerializePredictions, err, apiResGetPagesPrediction{})
		}
		predictionsByUUID[predictionUUID] = compilerPr
	}

	// Latest Predictions by same author URL
	latestPredictionSameAuthorURL := []string{}
	if pred.PostAuthorURL != "" {
		latest5PredictionsSameAuthor, err := a.store.GetPredictions(types.APIFilters{AuthorURLs: []string{pred.PostAuthorURL}}, []string{types.CREATED_AT_DESC.String()}, 5, 0)
		if err != nil {
			return failWith(ErrStorageErrorRetrievingPredictions, err, apiResGetPagesPrediction{})
		}
		for _, prediction := range latest5PredictionsSameAuthor {
			predictionUUID := prediction.UUID
			latestPredictionSameAuthorURL = append(latestPredictionSameAuthorURL, predictionUUID)
			compilerPr, err := ps.PreSerializeForAPI(&prediction)
			if err != nil {
				return failWith(ErrFailedToSerializePredictions, err, apiResGetPagesPrediction{})
			}
			predictionsByUUID[predictionUUID] = compilerPr
		}
	}

	// Latest Predictions by same coin
	latestPredictionSameCoinUUID := []string{}
	latest5PredictionsSameCoin, err := a.store.GetPredictions(types.APIFilters{Tags: []string{pred.CalculateMainCoin().Str}}, []string{types.CREATED_AT_DESC.String()}, 5, 0)
	if err != nil {
		return failWith(ErrStorageErrorRetrievingPredictions, err, apiResGetPagesPrediction{})
	}
	for _, prediction := range latest5PredictionsSameCoin {
		predictionUUID := prediction.UUID
		latestPredictionSameCoinUUID = append(latestPredictionSameCoinUUID, predictionUUID)
		compilerPr, err := ps.PreSerializeForAPI(&prediction)
		if err != nil {
			return failWith(ErrFailedToSerializePredictions, err, apiResGetPagesPrediction{})
		}
		predictionsByUUID[predictionUUID] = compilerPr
	}

	accountURLSet := map[string]struct{}{}

	// Get the top 10 Accounts by follower count
	topAccounts, err := a.store.GetAccounts(types.APIAccountFilters{}, []string{types.ACCOUNT_FOLLOWER_COUNT_DESC.String()}, 10, 0)
	if err != nil {
		return failWith(ErrStorageErrorRetrievingAccounts, err, apiResGetPagesPrediction{})
	}
	top10AccountURLsByFollowerCount := []string{}
	for _, account := range topAccounts {
		accountURL := account.URL.String()
		top10AccountURLsByFollowerCount = append(top10AccountURLsByFollowerCount, accountURL)
		accountURLSet[accountURL] = struct{}{}
	}

	// Also gather all account urls from all predictions into the set
	for _, prediction := range predictionsByUUID {
		if prediction.PostAuthorURL == "" {
			continue
		}
		accountURLSet[prediction.PostAuthorURL] = struct{}{}
	}

	// Make the set a slice
	accountURLs := []string{}
	for accountURL := range accountURLSet {
		accountURLs = append(accountURLs, accountURL)
	}

	// Get all accounts from the slice
	allAccounts, err := a.store.GetAccounts(types.APIAccountFilters{URLs: accountURLs}, []string{}, 0, 0)
	if err != nil {
		return failWith(ErrStorageErrorRetrievingAccounts, err, apiResGetPagesPrediction{})
	}

	accountsByURL := map[string]compiler.Account{}
	accountSerializer := compiler.NewAccountSerializer()
	for _, account := range allAccounts {
		accountURL := account.URL.String()
		compilerAcc, err := accountSerializer.PreSerialize(&account)
		if err != nil {
			return failWith(ErrFailedToSerializePredictions, err, apiResGetPagesPrediction{})
		}
		accountsByURL[accountURL] = compilerAcc
	}

	return apiResponse[apiResGetPagesPrediction]{Status: 200, Data: apiResGetPagesPrediction{
		Prediction:                    compilerPred,
		Summary:                       summary,
		Top10AccountsByFollowerCount:  top10AccountURLsByFollowerCount,
		AccountsByURL:                 accountsByURL,
		Latest10Predictions:           latestPredictionUUIDs,
		Latest5PredictionsSameAccount: latestPredictionSameAuthorURL,
		Latest5PredictionsSameCoin:    latestPredictionSameCoinUUID,
		PredictionsByUUID:             predictionsByUUID,
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

	return u
}
