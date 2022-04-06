package backoffice

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/request"
	"github.com/marianogappa/predictions/types"
)

var (
	errorResponses = map[error]parsedResponse{
		request.ErrAPIClientFailedToMarshalRequestBody: {
			ErrorCode: "ErrAPIClientFailedToMarshalRequestBody",
			Status:    500,
			Message:   "Client API failed to marshal request body",
		},
		request.ErrAPIClientFailedToCreateRequest: {
			ErrorCode: "ErrAPIClientFailedToCreateRequest",
			Status:    500,
			Message:   "Client API failed to create request",
		},
		request.ErrAPIClientFailedToExecuteRequest: {
			ErrorCode: "ErrAPIClientFailedToExecuteRequest",
			Status:    500,
			Message:   "Client API failed to execute request",
		},
		request.ErrAPIClientBrokenResponseBody: {
			ErrorCode: "ErrAPIClientBrokenResponseBody",
			Status:    500,
			Message:   "API returned broken response body",
		},
		request.ErrAPIClientInvalidResponseJSON: {
			ErrorCode: "ErrAPIClientInvalidResponseJSON",
			Status:    500,
			Message:   "API returned invalid response json",
		},
		request.ErrAPIClientFailedToParseResponse: {
			ErrorCode: "ErrAPIClientFailedToParseResponse",
			Status:    500,
			Message:   "Client API failed to parse response",
		},
	}
)

func parseResponse(rawResp response) (parsedResponse, error) {
	return rawResp.parse(), nil
}

func parseError(err error) parsedResponse {
	for errType, errResp := range errorResponses {
		if errors.Is(err, errType) {
			resp := errResp
			resp.InternalMessage = err.Error()
			return resp
		}
	}

	return parsedResponse{
		Status:          500,
		Message:         "Unknown error.",
		InternalMessage: err.Error(),
		ErrorCode:       "ErrUnknown",
	}
}

type response struct {
	Status            int                `json:"status"`
	Message           string             `json:"message,omitempty"`
	InternalMessage   string             `json:"internalMessage,omitempty"`
	ErrorCode         string             `json:"errorCode,omitempty"`
	Prediction        *json.RawMessage   `json:"prediction,omitempty"`
	Predictions       *[]json.RawMessage `json:"predictions,omitempty"`
	PredictionSummary *json.RawMessage   `json:"predictionSummary,omitempty"`
	Stored            *bool              `json:"stored,omitempty"`
}

func (r response) parse() parsedResponse {
	var (
		pred        *types.Prediction
		preds       *[]types.Prediction
		predSummary *predictionSummary
		stored      = r.Stored
		pc          = compiler.NewPredictionCompiler(nil, nil)
	)
	if r.Prediction != nil {
		p, _, _ := pc.Compile(*r.Prediction)
		pred = &p
	}
	if r.Predictions != nil {
		preds = &[]types.Prediction{}
		for _, rawPred := range *r.Predictions {
			p, _, _ := pc.Compile(rawPred)
			(*preds) = append((*preds), p)
		}
	}
	if r.PredictionSummary != nil {
		predSummary = &predictionSummary{}
		err := json.Unmarshal(*r.PredictionSummary, predSummary)
		if err != nil {
			log.Printf("Ignoring error unmarshalling prediction summary: %v\n", err)
		}
	}

	return parsedResponse{
		Status:            r.Status,
		Message:           r.Message,
		InternalMessage:   r.InternalMessage,
		ErrorCode:         r.ErrorCode,
		Prediction:        pred,
		Predictions:       preds,
		PredictionSummary: predSummary,
		Stored:            stored,
	}
}

type predictionSummary struct {
	TickMap  map[string][]types.Tick `json:"tickMap"`
	Coin     string                  `json:"coin"`
	Goal     types.JsonFloat64       `json:"goal"`
	Operator string                  `json:"operator"`
	Deadline types.ISO8601           `json:"deadline"`
}
type parsedResponse struct {
	Status            int
	Message           string
	InternalMessage   string
	ErrorCode         string
	Prediction        *types.Prediction
	Predictions       *[]types.Prediction
	PredictionSummary *predictionSummary
	Stored            *bool
}
