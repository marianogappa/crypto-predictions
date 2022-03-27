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

	ps := compiler.NewPredictionSerializer()
	raws := []json.RawMessage{}
	for _, pred := range preds {
		bs, err := ps.Serialize(&pred)
		if err != nil {
			return response{}, fmt.Errorf("%w: %v", ErrFailedToSerializePredictions, err)
		}
		raw := json.RawMessage(bs)
		raws = append(raws, raw)
	}

	return response{preds: &raws}, nil
}
