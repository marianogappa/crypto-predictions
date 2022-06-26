package request

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/rs/zerolog/log"
)

var (

	// ErrAPIClientFailedToMarshalRequestBody means: failed to marshal request body
	ErrAPIClientFailedToMarshalRequestBody = errors.New("failed to marshal request body")

	// ErrAPIClientFailedToCreateRequest means: failed to create request
	ErrAPIClientFailedToCreateRequest = errors.New("failed to create request")

	// ErrAPIClientFailedToExecuteRequest means: failed to execute request
	ErrAPIClientFailedToExecuteRequest = errors.New("failed to execute request")

	// ErrAPIClientBrokenResponseBody means: broken response body
	ErrAPIClientBrokenResponseBody = errors.New("broken response body")

	// ErrAPIClientInvalidResponseJSON means: invalid response json
	ErrAPIClientInvalidResponseJSON = errors.New("invalid response json")

	// ErrAPIClientFailedToParseResponse means: failed to parse response
	ErrAPIClientFailedToParseResponse = errors.New("failed to parse response")
)

// Request is the data for a generic function to make an HTTP request and JSON parse its response.
type Request[A, B any] struct {
	// Required
	BaseURL       string
	Path          string
	ParseResponse func(A) (B, error)
	ParseError    func(error) B

	// Optional
	BasicAuthUser string
	BasicAuthPass string
	QueryString   map[string][]string
	Headers       map[string]string
	Body          any
	HTTPMethod    string        // Defaults to "GET"
	Timeout       time.Duration // Defaults to 10 * time Second
}

// MakeRequest is a generic function to make an HTTP request and JSON parse its response.
func MakeRequest[A, B any](reqData Request[A, B], debug bool) B {
	var buf *bytes.Buffer
	if reqData.Body != nil {
		bs, err := json.Marshal(reqData.Body)
		if err != nil {
			return reqData.ParseError(fmt.Errorf("%w: %v", ErrAPIClientFailedToMarshalRequestBody, err))
		}
		buf = bytes.NewBuffer(bs)
	}

	httpMethod := "GET"
	if reqData.HTTPMethod != "" {
		httpMethod = reqData.HTTPMethod
	}

	// N.B. this is very interesting behaviour so remember this! Sending a *bytes.Buffer whose value is nil causes
	// a panic, but sending an untyped nil works fine. That's very confusing, but if you look at the implementation
	// it makes sense.
	var (
		req *http.Request
		err error
	)
	if buf != nil {
		req, err = http.NewRequest(httpMethod, fmt.Sprintf("%v/%v", reqData.BaseURL, reqData.Path), buf)
	} else {
		req, err = http.NewRequest(httpMethod, fmt.Sprintf("%v/%v", reqData.BaseURL, reqData.Path), nil)
	}
	if err != nil {
		return reqData.ParseError(fmt.Errorf("%w: %v", ErrAPIClientFailedToCreateRequest, err))
	}

	if len(reqData.QueryString) > 0 {
		values := req.URL.Query()
		for k, vs := range reqData.QueryString {
			for _, v := range vs {
				values.Add(k, v)
			}
		}
		req.URL.RawQuery = values.Encode()
	}

	if reqData.Headers == nil {
		reqData.Headers = map[string]string{}
	}
	if reqData.BasicAuthUser != "" || reqData.BasicAuthPass != "" {
		reqData.Headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", reqData.BasicAuthUser, reqData.BasicAuthPass)))
	}
	for k, v := range reqData.Headers {
		req.Header.Add(k, v)
	}

	timeout := 10 * time.Second
	if reqData.Timeout != 0 {
		timeout = reqData.Timeout
	}
	client := &http.Client{Timeout: timeout}

	if debug {
		res, _ := httputil.DumpRequest(req, true)
		log.Info().Msgf("APIClient.MakeRequest: making API request: %v\n", string(res))
	}

	resp, err := client.Do(req)
	if err != nil {
		return reqData.ParseError(fmt.Errorf("%w: %v", ErrAPIClientFailedToExecuteRequest, err))
	}
	defer resp.Body.Close()

	byts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return reqData.ParseError(fmt.Errorf("%w: %v", ErrAPIClientBrokenResponseBody, err))
	}

	var rp A
	err = json.Unmarshal(byts, &rp)
	if err != nil {
		return reqData.ParseError(fmt.Errorf("%w: %v on [%v]", ErrAPIClientInvalidResponseJSON, err, string(byts)))
	}

	res, err := reqData.ParseResponse(rp)
	if err != nil {
		if debug {
			log.Info().Msgf("APIClient.MakeRequest: error parsing response from api: %v\n", err)
		}
		return reqData.ParseError(fmt.Errorf("%w: %v", ErrAPIClientFailedToParseResponse, err))
	}

	if debug {
		log.Info().Msg("APIClient.MakeRequest: successfully parsed response from api")
	}

	return res
}
