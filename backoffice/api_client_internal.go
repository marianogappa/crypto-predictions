package backoffice

import (
	"encoding/json"
	"errors"

	"github.com/marianogappa/crypto-candles/candles/common"
	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/request"
	"github.com/marianogappa/predictions/types"
	"github.com/rs/zerolog/log"
)

var (
	errorResponses = map[error]parsedResponse{
		request.ErrAPIClientFailedToMarshalRequestBody: {
			ErrorCode:    "ErrAPIClientFailedToMarshalRequestBody",
			Status:       500,
			ErrorMessage: "Client API failed to marshal request body",
		},
		request.ErrAPIClientFailedToCreateRequest: {
			ErrorCode:    "ErrAPIClientFailedToCreateRequest",
			Status:       500,
			ErrorMessage: "Client API failed to create request",
		},
		request.ErrAPIClientFailedToExecuteRequest: {
			ErrorCode:    "ErrAPIClientFailedToExecuteRequest",
			Status:       500,
			ErrorMessage: "Client API failed to execute request",
		},
		request.ErrAPIClientBrokenResponseBody: {
			ErrorCode:    "ErrAPIClientBrokenResponseBody",
			Status:       500,
			ErrorMessage: "API returned broken response body",
		},
		request.ErrAPIClientInvalidResponseJSON: {
			ErrorCode:    "ErrAPIClientInvalidResponseJSON",
			Status:       500,
			ErrorMessage: "API returned invalid response json",
		},
		request.ErrAPIClientFailedToParseResponse: {
			ErrorCode:    "ErrAPIClientFailedToParseResponse",
			Status:       500,
			ErrorMessage: "Client API failed to parse response",
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
			resp.InternalErrorMessage = err.Error()
			return resp
		}
	}

	return parsedResponse{
		Status:               500,
		ErrorMessage:         "Unknown error.",
		InternalErrorMessage: err.Error(),
		ErrorCode:            "ErrUnknown",
	}
}

type response struct {
	Status               int          `json:"status"`
	ErrorMessage         string       `json:"errorMessage,omitempty"`
	InternalErrorMessage string       `json:"internalErrorMessage,omitempty"`
	Data                 responseData `json:"data"`
	ErrorCode            string       `json:"errorCode,omitempty"`
}

type responseData struct {
	Prediction        *json.RawMessage   `json:"prediction,omitempty"`
	Predictions       *[]json.RawMessage `json:"predictions,omitempty"`
	PredictionSummary *json.RawMessage   `json:"predictionSummary,omitempty"`
	Stored            *bool              `json:"stored,omitempty"`
	Base64Image       *string            `json:"base64Image,omitempty"`
}

func (r response) parse() parsedResponse {
	var (
		pred        *types.Prediction
		preds       *[]types.Prediction
		predSummary *predictionSummary
		stored      = r.Data.Stored
		pc          = compiler.NewPredictionCompiler(nil, nil)
		base64Image *string
	)
	if r.Data.Prediction != nil {
		p, _, _ := pc.Compile(*r.Data.Prediction)
		pred = &p
	}
	if r.Data.Predictions != nil {
		preds = &[]types.Prediction{}
		for _, rawPred := range *r.Data.Predictions {
			p, _, _ := pc.Compile(rawPred)
			(*preds) = append((*preds), p)
		}
	}
	if r.Data.PredictionSummary != nil {
		predSummary = &predictionSummary{}
		err := json.Unmarshal(*r.Data.PredictionSummary, predSummary)
		if err != nil {
			log.Info().Msgf("Ignoring error unmarshalling prediction summary: %v\n", err)
		}
	}
	if r.Data.Base64Image != nil {
		base64Image = r.Data.Base64Image
	}

	return parsedResponse{
		Status:               r.Status,
		ErrorMessage:         r.ErrorMessage,
		InternalErrorMessage: r.InternalErrorMessage,
		ErrorCode:            r.ErrorCode,
		Prediction:           pred,
		Predictions:          preds,
		PredictionSummary:    predSummary,
		Stored:               stored,
		Base64Image:          base64Image,
	}
}

type predictionSummary struct {
	TickMap  map[string][]common.Tick `json:"tickMap"`
	Coin     string                   `json:"coin"`
	Goal     types.JSONFloat64        `json:"goal"`
	Operator string                   `json:"operator"`
	Deadline types.ISO8601            `json:"deadline"`
}
type parsedResponse struct {
	Status               int
	ErrorMessage         string
	InternalErrorMessage string
	ErrorCode            string
	Prediction           *types.Prediction
	Predictions          *[]types.Prediction
	PredictionSummary    *predictionSummary
	Stored               *bool
	Base64Image          *string
}

func (r parsedResponse) String() string {
	bs, _ := json.Marshal(r)
	return string(bs)
}
