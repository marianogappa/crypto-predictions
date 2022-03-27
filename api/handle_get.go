package api

import (
	"encoding/json"
	"fmt"

	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/types"
)

type getBody struct {
	Filters  types.APIFilters `json:"filters"`
	OrderBys []string         `json:"orderBys"`
}

func (a *API) handleGet(params getBody) (response, error) {
	preds, err := a.store.GetPredictions(
		params.Filters,
		params.OrderBys,
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
