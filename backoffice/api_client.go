package backoffice

import (
	"fmt"

	"github.com/marianogappa/predictions/request"
	"github.com/marianogappa/predictions/types"
)

type apiClient struct {
	apiURL        string
	debug         bool
	basicAuthUser string
	basicAuthPass string
}

func newAPIClient(apiURL, basicAuthUser, basicAuthPass string) *apiClient {
	return &apiClient{apiURL: apiURL, basicAuthUser: basicAuthUser, basicAuthPass: basicAuthPass}
}

func (c *apiClient) setDebug(b bool) {
	c.debug = b
}

type newBody struct {
	Prediction string `json:"prediction"`
	Store      bool   `json:"store"`
}

func (c apiClient) new(pred []byte, store bool) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HTTPMethod:    "POST",
		BaseURL:       c.apiURL,
		Path:          "predictions",
		QueryString:   nil,
		Body:          newBody{string(pred), store},
		ParseResponse: parseResponse,
		ParseError:    parseError,
		BasicAuthUser: c.basicAuthUser,
		BasicAuthPass: c.basicAuthPass,
	}

	return request.MakeRequest(reqData, c.debug)
}

type getBody struct {
	Filters  types.APIFilters `json:"filters"`
	OrderBys []string         `json:"orderBys"`
}

func (c apiClient) get(body getBody) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HTTPMethod:    "GET",
		BaseURL:       c.apiURL,
		Path:          "predictions",
		QueryString:   body.Filters.ToQueryStringWithOrderBy(body.OrderBys),
		Body:          nil,
		ParseResponse: parseResponse,
		ParseError:    parseError,
		BasicAuthUser: c.basicAuthUser,
		BasicAuthPass: c.basicAuthPass,
	}

	return request.MakeRequest(reqData, c.debug)
}

func (c apiClient) predictionPage(body getBody) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HTTPMethod:    "POST",
		BaseURL:       c.apiURL,
		Path:          fmt.Sprintf("pages/prediction/%v", body.Filters.URLs[0]),
		QueryString:   nil,
		Body:          nil,
		ParseResponse: parseResponse,
		ParseError:    parseError,
	}

	return request.MakeRequest(reqData, c.debug)
}

type predictionImageBody struct {
	UUID string `json:"uuid"`
}

func (c apiClient) predictionImage(body predictionImageBody) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HTTPMethod:    "GET",
		BaseURL:       c.apiURL,
		Path:          fmt.Sprintf("predictions/%v/image", body.UUID),
		QueryString:   nil,
		Body:          nil,
		ParseResponse: parseResponse,
		ParseError:    parseError,
		BasicAuthUser: c.basicAuthUser,
		BasicAuthPass: c.basicAuthPass,
	}

	return request.MakeRequest(reqData, c.debug)
}

func (c apiClient) pausePrediction(uuid string) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HTTPMethod:    "POST",
		BaseURL:       c.apiURL,
		Path:          fmt.Sprintf("predictions/%v/pause", uuid),
		QueryString:   nil,
		Body:          nil,
		ParseResponse: parseResponse,
		ParseError:    parseError,
		BasicAuthUser: c.basicAuthUser,
		BasicAuthPass: c.basicAuthPass,
	}

	return request.MakeRequest(reqData, c.debug)
}

func (c apiClient) unpausePrediction(uuid string) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HTTPMethod:    "POST",
		BaseURL:       c.apiURL,
		Path:          fmt.Sprintf("predictions/%v/unpause", uuid),
		QueryString:   nil,
		Body:          nil,
		ParseResponse: parseResponse,
		ParseError:    parseError,
		BasicAuthUser: c.basicAuthUser,
		BasicAuthPass: c.basicAuthPass,
	}

	return request.MakeRequest(reqData, c.debug)
}

func (c apiClient) hidePrediction(uuid string) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HTTPMethod:    "POST",
		BaseURL:       c.apiURL,
		Path:          fmt.Sprintf("predictions/%v/hide", uuid),
		QueryString:   nil,
		Body:          nil,
		ParseResponse: parseResponse,
		ParseError:    parseError,
		BasicAuthUser: c.basicAuthUser,
		BasicAuthPass: c.basicAuthPass,
	}

	return request.MakeRequest(reqData, c.debug)
}

func (c apiClient) unhidePrediction(uuid string) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HTTPMethod:    "POST",
		BaseURL:       c.apiURL,
		Path:          fmt.Sprintf("predictions/%v/unhide", uuid),
		QueryString:   nil,
		Body:          nil,
		ParseResponse: parseResponse,
		ParseError:    parseError,
		BasicAuthUser: c.basicAuthUser,
		BasicAuthPass: c.basicAuthPass,
	}

	return request.MakeRequest(reqData, c.debug)
}

func (c apiClient) deletePrediction(uuid string) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HTTPMethod:    "POST",
		BaseURL:       c.apiURL,
		Path:          fmt.Sprintf("predictions/%v/delete", uuid),
		QueryString:   nil,
		Body:          nil,
		ParseResponse: parseResponse,
		ParseError:    parseError,
		BasicAuthUser: c.basicAuthUser,
		BasicAuthPass: c.basicAuthPass,
	}

	return request.MakeRequest(reqData, c.debug)
}

func (c apiClient) undeletePrediction(uuid string) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HTTPMethod:    "POST",
		BaseURL:       c.apiURL,
		Path:          fmt.Sprintf("predictions/%v/undelete", uuid),
		QueryString:   nil,
		Body:          nil,
		ParseResponse: parseResponse,
		ParseError:    parseError,
		BasicAuthUser: c.basicAuthUser,
		BasicAuthPass: c.basicAuthPass,
	}

	return request.MakeRequest(reqData, c.debug)
}

func (c apiClient) refreshAccount(uuid string) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HTTPMethod:    "POST",
		BaseURL:       c.apiURL,
		Path:          fmt.Sprintf("predictions/%v/refetchAccount", uuid),
		QueryString:   nil,
		Body:          nil,
		ParseResponse: parseResponse,
		ParseError:    parseError,
		BasicAuthUser: c.basicAuthUser,
		BasicAuthPass: c.basicAuthPass,
	}

	return request.MakeRequest(reqData, c.debug)
}

func (c apiClient) clearState(uuid string) parsedResponse {
	reqData := request.Request[response, parsedResponse]{
		HTTPMethod:    "POST",
		BaseURL:       c.apiURL,
		Path:          fmt.Sprintf("predictions/%v/clearState", uuid),
		QueryString:   nil,
		Body:          nil,
		ParseResponse: parseResponse,
		ParseError:    parseError,
		BasicAuthUser: c.basicAuthUser,
		BasicAuthPass: c.basicAuthPass,
	}

	return request.MakeRequest(reqData, c.debug)
}
