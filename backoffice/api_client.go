package backoffice

import (
	"encoding/json"

	"github.com/marianogappa/predictions/request"
	"github.com/marianogappa/predictions/types"
)

type APIClient struct {
	apiURL string
}

func NewAPIClient(apiURL string) APIClient {
	return APIClient{apiURL}
}

type newBody struct {
	Prediction json.RawMessage `json:"prediction"`
	Store      bool            `json:"store"`
}

func (c APIClient) New(pred []byte, store bool) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HttpMethod:    "POST",
		BaseUrl:       c.apiURL,
		Path:          "new",
		QueryString:   nil,
		Body:          newBody{pred, store},
		ParseResponse: parseResponse,
		ParseError:    parseError,
	}

	return request.MakeRequest(reqData)
}

type getBody struct {
	Filters  types.APIFilters `json:"filters"`
	OrderBys []string         `json:"orderBys"`
}

func (c APIClient) Get(body getBody) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HttpMethod:    "POST",
		BaseUrl:       c.apiURL,
		Path:          "get",
		QueryString:   nil,
		Body:          body,
		ParseResponse: parseResponse,
		ParseError:    parseError,
	}

	return request.MakeRequest(reqData)
}
