package api

import (
	"encoding/json"
	"fmt"

	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/types"
)

type bodyPagePrediction struct {
	URL string `json:"url"`
}

func (a *API) handleBodyPagePrediction(params bodyPagePrediction) (response, error) {
	url := params.URL

	preds, err := a.store.GetPredictions(
		types.APIFilters{URLs: []string{url}},
		[]string{},
		0, 0,
	)
	if err != nil {
		return response{}, fmt.Errorf("%w: %v", ErrStorageErrorRetrievingPredictions, err)
	}
	if len(preds) != 1 {
		return response{}, fmt.Errorf("%w: %v", ErrPredictionNotFound, err)
	}
	pred := preds[0]

	ps := compiler.NewPredictionSerializer()
	bs, err := ps.SerializeForAPI(&pred)
	if err != nil {
		return response{}, fmt.Errorf("%w: %v", ErrFailedToSerializePredictions, err)
	}
	raw := json.RawMessage(bs)

	summary, err := a.BuildPredictionMarketSummary(pred)
	if err != nil {
		return response{}, fmt.Errorf("%w: %v", ErrStorageErrorRetrievingPredictions, err)
	}

	summaryBs, _ := json.Marshal(summary)
	rawSummary := json.RawMessage(summaryBs)

	// Latest Predictions
	latest10Predictions, err := a.store.GetPredictions(types.APIFilters{}, []string{types.CREATED_AT_DESC.String()}, 10, 0)
	if err != nil {
		return response{}, fmt.Errorf("%w: %v", ErrStorageErrorRetrievingPredictions, err)
	}
	latestPredictionUUIDs := []string{}
	predictionsByUUID := map[string]json.RawMessage{}
	for _, prediction := range latest10Predictions {
		predictionUUID := prediction.UUID
		latestPredictionUUIDs = append(latestPredictionUUIDs, predictionUUID)
		bs, err := ps.SerializeForAPI(&prediction)
		if err != nil {
			return response{}, fmt.Errorf("%w: %v", ErrFailedToSerializePredictions, err)
		}
		predictionsByUUID[predictionUUID] = json.RawMessage(bs)
	}

	// Latest Predictions by same author URL
	latestPredictionSameAuthorURL := []string{}
	if pred.PostAuthorURL != "" {
		latest5PredictionsSameAuthor, err := a.store.GetPredictions(types.APIFilters{AuthorURLs: []string{pred.PostAuthorURL}}, []string{types.CREATED_AT_DESC.String()}, 5, 0)
		if err != nil {
			return response{}, fmt.Errorf("%w: %v", ErrStorageErrorRetrievingPredictions, err)
		}
		for _, prediction := range latest5PredictionsSameAuthor {
			predictionUUID := prediction.UUID
			latestPredictionSameAuthorURL = append(latestPredictionSameAuthorURL, predictionUUID)
			bs, err := ps.SerializeForAPI(&prediction)
			if err != nil {
				return response{}, fmt.Errorf("%w: %v", ErrFailedToSerializePredictions, err)
			}
			predictionsByUUID[predictionUUID] = json.RawMessage(bs)
		}
	}

	// Latest Predictions by same coin
	latestPredictionSameCoinUUID := []string{}
	latest5PredictionsSameCoin, err := a.store.GetPredictions(types.APIFilters{Tags: []string{pred.CalculateMainCoin().Str}}, []string{types.CREATED_AT_DESC.String()}, 5, 0)
	if err != nil {
		return response{}, fmt.Errorf("%w: %v", ErrStorageErrorRetrievingPredictions, err)
	}
	for _, prediction := range latest5PredictionsSameCoin {
		predictionUUID := prediction.UUID
		latestPredictionSameCoinUUID = append(latestPredictionSameCoinUUID, predictionUUID)
		bs, err := ps.SerializeForAPI(&prediction)
		if err != nil {
			return response{}, fmt.Errorf("%w: %v", ErrFailedToSerializePredictions, err)
		}
		predictionsByUUID[predictionUUID] = json.RawMessage(bs)
	}

	// Accounts
	accounts, err := a.store.GetAccounts(types.APIAccountFilters{}, []string{types.ACCOUNT_FOLLOWER_COUNT_DESC.String()}, 10, 0)
	if err != nil {
		return response{}, fmt.Errorf("%w: %v", ErrStorageErrorRetrievingAccounts, err)
	}
	accountURLs := []string{}
	accountsByURL := map[string]json.RawMessage{}
	accountSerializer := compiler.NewAccountSerializer()
	for _, account := range accounts {
		accountURL := account.URL.String()
		accountURLs = append(accountURLs, accountURL)
		bs, err := accountSerializer.Serialize(&account)
		if err != nil {
			return response{}, fmt.Errorf("%w: %v", ErrFailedToSerializePredictions, err)
		}
		accountsByURL[accountURL] = json.RawMessage(bs)
	}

	return response{
		pred:                          &raw,
		summary:                       &rawSummary,
		top10AccountsByFollowerCount:  &accountURLs,
		accountsByURL:                 &accountsByURL,
		latest10Predictions:           &latestPredictionUUIDs,
		latest5PredictionsSameAccount: &latestPredictionSameAuthorURL,
		latest5PredictionsSameCoin:    &latestPredictionSameCoinUUID,
		predictionsByUUID:             &predictionsByUUID,
	}, nil
}
