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

type predictionPageBody struct {
	URL string `json:"url"`
}

func (c APIClient) PredictionPage(body getBody) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HttpMethod:    "POST",
		BaseUrl:       c.apiURL,
		Path:          "prediction",
		QueryString:   nil,
		Body:          predictionPageBody{URL: body.Filters.URLs[0]},
		ParseResponse: parseResponse,
		ParseError:    parseError,
	}

	return request.MakeRequest(reqData)
}

type uuidBody struct {
	UUID string `json:"uuid"`
}

func (c APIClient) PausePrediction(uuid string) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HttpMethod:    "POST",
		BaseUrl:       c.apiURL,
		Path:          "predictionPause",
		QueryString:   nil,
		Body:          uuidBody{uuid},
		ParseResponse: parseResponse,
		ParseError:    parseError,
	}

	return request.MakeRequest(reqData)
}

func (c APIClient) UnpausePrediction(uuid string) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HttpMethod:    "POST",
		BaseUrl:       c.apiURL,
		Path:          "predictionUnpause",
		QueryString:   nil,
		Body:          uuidBody{uuid},
		ParseResponse: parseResponse,
		ParseError:    parseError,
	}

	return request.MakeRequest(reqData)
}

func (c APIClient) HidePrediction(uuid string) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HttpMethod:    "POST",
		BaseUrl:       c.apiURL,
		Path:          "predictionHide",
		QueryString:   nil,
		Body:          uuidBody{uuid},
		ParseResponse: parseResponse,
		ParseError:    parseError,
	}

	return request.MakeRequest(reqData)
}

func (c APIClient) UnhidePrediction(uuid string) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HttpMethod:    "POST",
		BaseUrl:       c.apiURL,
		Path:          "predictionUnhide",
		QueryString:   nil,
		Body:          uuidBody{uuid},
		ParseResponse: parseResponse,
		ParseError:    parseError,
	}

	return request.MakeRequest(reqData)
}

func (c APIClient) DeletePrediction(uuid string) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HttpMethod:    "POST",
		BaseUrl:       c.apiURL,
		Path:          "predictionDelete",
		QueryString:   nil,
		Body:          uuidBody{uuid},
		ParseResponse: parseResponse,
		ParseError:    parseError,
	}

	return request.MakeRequest(reqData)
}

func (c APIClient) UndeletePrediction(uuid string) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HttpMethod:    "POST",
		BaseUrl:       c.apiURL,
		Path:          "predictionUndelete",
		QueryString:   nil,
		Body:          uuidBody{uuid},
		ParseResponse: parseResponse,
		ParseError:    parseError,
	}

	return request.MakeRequest(reqData)
}

func (c APIClient) RefreshAccount(uuid string) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HttpMethod:    "POST",
		BaseUrl:       c.apiURL,
		Path:          "predictionRefreshAccount",
		QueryString:   nil,
		Body:          uuidBody{uuid},
		ParseResponse: parseResponse,
		ParseError:    parseError,
	}

	return request.MakeRequest(reqData)
}
