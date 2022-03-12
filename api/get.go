package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/types"
)

type getBody struct {
	Filters  types.APIFilters `json:"filters"`
	OrderBys []string         `json:"orderBys"`
}

func (a *API) getHandler(w http.ResponseWriter, r *http.Request) {
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		respond(w, nil, nil, nil, fmt.Errorf("%w: %v", ErrInvalidRequestBody, err))
		return
	}
	defer r.Body.Close()

	var params getBody
	err = json.Unmarshal(bs, &params)
	if err != nil {
		respond(w, nil, nil, nil, fmt.Errorf("%w: %v", ErrInvalidRequestJSON, err))
		return
	}

	preds, err := a.store.GetPredictions(
		params.Filters,
		params.OrderBys,
	)
	if err != nil {
		respond(w, nil, nil, nil, fmt.Errorf("%w: %v", ErrStorageErrorRetrievingPredictions, err))
		return
	}

	ps := compiler.NewPredictionSerializer()
	raws := []json.RawMessage{}
	for _, pred := range preds {
		bs, err := ps.Serialize(&pred)
		if err != nil {
			respond(w, nil, nil, nil, fmt.Errorf("%w: %v", ErrFailedToSerializePredictions, err))
			return
		}
		raw := json.RawMessage(bs)
		raws = append(raws, raw)
	}

	respond(w, nil, &raws, nil, nil)
}
