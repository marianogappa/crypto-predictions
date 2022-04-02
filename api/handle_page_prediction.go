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
	bs, err := ps.Serialize(&pred)
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

	return response{pred: &raw, summary: &rawSummary}, nil
}
