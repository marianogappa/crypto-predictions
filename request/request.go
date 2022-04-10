package request

import (
	"bytes"
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
	ErrAPIClientFailedToMarshalRequestBody = errors.New("failed to marshal request body")
	ErrAPIClientFailedToCreateRequest      = errors.New("failed to create request")
	ErrAPIClientFailedToExecuteRequest     = errors.New("failed to execute request")
	ErrAPIClientBrokenResponseBody         = errors.New("broken response body")
	ErrAPIClientInvalidResponseJSON        = errors.New("invalid response json")
	ErrAPIClientFailedToParseResponse      = errors.New("failed to parse response")
)

type Request[A, B any] struct {
	// Required
	BaseUrl       string
	Path          string
	ParseResponse func(A) (B, error)
	ParseError    func(error) B

	// Optional
	QueryString map[string][]string
	Headers     map[string]string
	Body        any
	HttpMethod  string        // Defaults to "GET"
	Timeout     time.Duration // Defaults to 10 * time Second
}

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
	if reqData.HttpMethod != "" {
		httpMethod = reqData.HttpMethod
	}

	// N.B. this is very interesting behaviour so remember this! Sending a *bytes.Buffer whose value is nil causes
	// a panic, but sending an untyped nil works fine. That's very confusing, but if you look at the implementation
	// it makes sense.
	var (
		req *http.Request
		err error
	)
	if buf != nil {
		req, err = http.NewRequest(httpMethod, fmt.Sprintf("%v/%v", reqData.BaseUrl, reqData.Path), buf)
	} else {
		req, err = http.NewRequest(httpMethod, fmt.Sprintf("%v/%v", reqData.BaseUrl, reqData.Path), nil)
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
		return reqData.ParseError(fmt.Errorf("%w: %v", ErrAPIClientFailedToParseResponse, err))
	}

	if debug {
		log.Info().Msgf("APIClient.MakeRequest: response from API: %+v, error: %v\n", res, err)
	}

	return res
}
