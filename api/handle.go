package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type response struct {
	pred    *json.RawMessage
	preds   *[]json.RawMessage
	summary *json.RawMessage
	stored  *bool

	predictionUUID                *string
	accountURL                    *string
	latest5PredictionsSameAccount *[]string
	latest10Predictions           *[]string
	latest5PredictionsSameCoin    *[]string
	top10AccountsByFollowerCount  *[]string
	predictionsByUUID             *map[string]json.RawMessage
	accountsByURL                 *map[string]json.RawMessage
}

func buildHandler[P any](do func(p P) (response, error)) func(w http.ResponseWriter, r *http.Request) {
	f := func(w http.ResponseWriter, r *http.Request) {
		bs, err := io.ReadAll(r.Body)
		if err != nil {
			respond(w, response{}, fmt.Errorf("%w: %v", ErrInvalidRequestBody, err))
			return
		}
		defer r.Body.Close()

		var params P
		if len(bs) > 0 {
			err = json.Unmarshal(bs, &params)
			if err != nil {
				respond(w, response{}, fmt.Errorf("%w: %v", ErrInvalidRequestJSON, err))
				return
			}
		}

		resp, err := do(params)
		if err != nil {
			respond(w, response{}, err)
			return
		}

		respond(w, resp, nil)
	}

	return f
}

type APIResponse struct {
	Status            int                `json:"status"`
	Message           string             `json:"message,omitempty"`
	InternalMessage   string             `json:"internalMessage,omitempty"`
	ErrorCode         string             `json:"errorCode,omitempty"`
	Prediction        *json.RawMessage   `json:"prediction,omitempty"`
	Predictions       *[]json.RawMessage `json:"predictions,omitempty"`
	PredictionSummary *json.RawMessage   `json:"predictionSummary,omitempty"`

	// Prediction page fields
	PredictionUUID                *string                     `json:"predictionUUID,omitempty"`
	AccountURL                    *string                     `json:"accountURL,omitempty"`
	Latest5PredictionsSameAccount *[]string                   `json:"latest5PredictionsSameAccount,omitempty"`
	Latest10Predictions           *[]string                   `json:"latest10Predictions,omitempty"`
	Latest5PredictionsSameCoin    *[]string                   `json:"latest5PredictionsSameCoin,omitempty"`
	Top10AccountsByFollowerCount  *[]string                   `json:"top10AccountsByFollowerCount,omitempty"`
	PredictionsByUUID             *map[string]json.RawMessage `json:"predictionsByUUID,omitempty"`
	AccountsByURL                 *map[string]json.RawMessage `json:"accountsByURL,omitempty"`

	Stored *bool `json:"stored,omitempty"`
}

func respond(w http.ResponseWriter, resp response, err error) {
	if err == nil {
		doRespond(w, APIResponse{
			Message:           "",
			Prediction:        resp.pred,
			Predictions:       resp.preds,
			PredictionSummary: resp.summary,
			Stored:            resp.stored,

			PredictionUUID:                resp.predictionUUID,
			AccountURL:                    resp.accountURL,
			Latest5PredictionsSameAccount: resp.latest5PredictionsSameAccount,
			Latest10Predictions:           resp.latest10Predictions,
			Latest5PredictionsSameCoin:    resp.latest5PredictionsSameCoin,
			Top10AccountsByFollowerCount:  resp.top10AccountsByFollowerCount,
			PredictionsByUUID:             resp.predictionsByUUID,
			AccountsByURL:                 resp.accountsByURL,

			Status: 200,
		})
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
