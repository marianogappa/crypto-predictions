package backoffice

import (
	"fmt"

	"github.com/marianogappa/predictions/request"
	"github.com/marianogappa/predictions/types"
)

type APIClient struct {
	apiURL string
	debug  bool
}

func NewAPIClient(apiURL string) *APIClient {
	return &APIClient{apiURL: apiURL}
}

func (c *APIClient) SetDebug(b bool) {
	c.debug = b
}

type newBody struct {
	Prediction string `json:"prediction"`
	Store      bool   `json:"store"`
}

func (c APIClient) New(pred []byte, store bool) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HttpMethod:    "POST",
		BaseUrl:       c.apiURL,
		Path:          "predictions",
		QueryString:   nil,
		Body:          newBody{string(pred), store},
		ParseResponse: parseResponse,
		ParseError:    parseError,
	}

	return request.MakeRequest(reqData, c.debug)
}

type getBody struct {
	Filters  types.APIFilters `json:"filters"`
	OrderBys []string         `json:"orderBys"`
}

func (c APIClient) Get(body getBody) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HttpMethod:    "GET",
		BaseUrl:       c.apiURL,
		Path:          "predictions",
		QueryString:   body.Filters.ToQueryStringWithOrderBy(body.OrderBys),
		Body:          nil,
		ParseResponse: parseResponse,
		ParseError:    parseError,
	}

	return request.MakeRequest(reqData, c.debug)
}

type predictionPageBody struct {
	URL string `json:"url"`
}

func (c APIClient) PredictionPage(body getBody) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HttpMethod:    "POST",
		BaseUrl:       c.apiURL,
		Path:          fmt.Sprintf("pages/prediction/%v", body.Filters.URLs[0]),
		QueryString:   nil,
		Body:          nil,
		ParseResponse: parseResponse,
		ParseError:    parseError,
	}

	return request.MakeRequest(reqData, c.debug)
}

type uuidBody struct {
	UUID string `json:"uuid"`
}

func (c APIClient) PausePrediction(uuid string) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HttpMethod:    "POST",
		BaseUrl:       c.apiURL,
		Path:          fmt.Sprintf("predictions/%v/pause", uuid),
		QueryString:   nil,
		Body:          nil,
		ParseResponse: parseResponse,
		ParseError:    parseError,
	}

	return request.MakeRequest(reqData, c.debug)
}

func (c APIClient) UnpausePrediction(uuid string) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HttpMethod:    "POST",
		BaseUrl:       c.apiURL,
		Path:          fmt.Sprintf("predictions/%v/unpause", uuid),
		QueryString:   nil,
		Body:          nil,
		ParseResponse: parseResponse,
		ParseError:    parseError,
	}

	return request.MakeRequest(reqData, c.debug)
}

func (c APIClient) HidePrediction(uuid string) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HttpMethod:    "POST",
		BaseUrl:       c.apiURL,
		Path:          fmt.Sprintf("predictions/%v/hide", uuid),
		QueryString:   nil,
		Body:          nil,
		ParseResponse: parseResponse,
		ParseError:    parseError,
	}

	return request.MakeRequest(reqData, c.debug)
}

func (c APIClient) UnhidePrediction(uuid string) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HttpMethod:    "POST",
		BaseUrl:       c.apiURL,
		Path:          fmt.Sprintf("predictions/%v/unhide", uuid),
		QueryString:   nil,
		Body:          nil,
		ParseResponse: parseResponse,
		ParseError:    parseError,
	}

	return request.MakeRequest(reqData, c.debug)
}

func (c APIClient) DeletePrediction(uuid string) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HttpMethod:    "POST",
		BaseUrl:       c.apiURL,
		Path:          fmt.Sprintf("predictions/%v/delete", uuid),
		QueryString:   nil,
		Body:          nil,
		ParseResponse: parseResponse,
		ParseError:    parseError,
	}

	return request.MakeRequest(reqData, c.debug)
}

func (c APIClient) UndeletePrediction(uuid string) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HttpMethod:    "POST",
		BaseUrl:       c.apiURL,
		Path:          fmt.Sprintf("predictions/%v/undelete", uuid),
		QueryString:   nil,
		Body:          nil,
		ParseResponse: parseResponse,
		ParseError:    parseError,
	}

	return request.MakeRequest(reqData, c.debug)
}

func (c APIClient) RefreshAccount(uuid string) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HttpMethod:    "POST",
		BaseUrl:       c.apiURL,
		Path:          fmt.Sprintf("predictions/%v/refetchAccount", uuid),
		QueryString:   nil,
		Body:          nil,
		ParseResponse: parseResponse,
		ParseError:    parseError,
	}

	return request.MakeRequest(reqData, c.debug)
}
