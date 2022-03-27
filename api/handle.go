package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type response struct {
	pred   *json.RawMessage
	preds  *[]json.RawMessage
	stored *bool
}

func buildHandler[P any](do func(p P) (response, error)) func(w http.ResponseWriter, r *http.Request) {
	f := func(w http.ResponseWriter, r *http.Request) {
		bs, err := io.ReadAll(r.Body)
		if err != nil {
			respond(w, nil, nil, nil, fmt.Errorf("%w: %v", ErrInvalidRequestBody, err))
			return
		}
		defer r.Body.Close()

		var params P
		err = json.Unmarshal(bs, &params)
		if err != nil {
			respond(w, nil, nil, nil, fmt.Errorf("%w: %v", ErrInvalidRequestJSON, err))
			return
		}

		resp, err := do(params)
		if err != nil {
			respond(w, nil, nil, nil, err)
			return
		}

		respond(w, resp.pred, resp.preds, resp.stored, nil)
	}

	return f
}

type APIResponse struct {
	Status          int                `json:"status"`
	Message         string             `json:"message,omitempty"`
	InternalMessage string             `json:"internalMessage,omitempty"`
	ErrorCode       string             `json:"errorCode,omitempty"`
	Prediction      *json.RawMessage   `json:"prediction,omitempty"`
	Predictions     *[]json.RawMessage `json:"predictions,omitempty"`
	Stored          *bool              `json:"stored,omitempty"`
}

func respond(w http.ResponseWriter, pred *json.RawMessage, preds *[]json.RawMessage, stored *bool, err error) {
	if err == nil {
		doRespond(w, APIResponse{Message: "", Prediction: pred, Predictions: preds, Stored: stored, Status: 200})
		return
	}

	r := APIResponse{Message: "Unknown internal error.", Status: 500, InternalMessage: err.Error()}
	for maybeErr, maybeResp := range errToResponse {
		if errors.Is(err, maybeErr) {
			r = maybeResp
			r.InternalMessage = err.Error()
		}
	}
	doRespond(w, r)
}

func doRespond(w http.ResponseWriter, r APIResponse) {
	w.WriteHeader(r.Status)
	enc := json.NewEncoder(w)
	enc.Encode(r)
}
