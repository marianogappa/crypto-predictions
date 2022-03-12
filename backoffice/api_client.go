package backoffice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/types"
)

type response struct {
	Status          int                `json:"status"`
	Message         string             `json:"message,omitempty"`
	InternalMessage string             `json:"internalMessage,omitempty"`
	ErrorCode       string             `json:"errorCode,omitempty"`
	Prediction      *json.RawMessage   `json:"prediction,omitempty"`
	Predictions     *[]json.RawMessage `json:"predictions,omitempty"`
	Stored          *bool              `json:"stored,omitempty"`
}

func (r response) parse() parsedResponse {
	var (
		pred   *types.Prediction
		preds  *[]types.Prediction
		stored *bool
		pc     = compiler.NewPredictionCompiler(nil, nil)
	)
	if r.Prediction != nil {
		p, _ := pc.Compile(*r.Prediction)
		pred = &p
	}
	if r.Predictions != nil {
		preds = &[]types.Prediction{}
		for _, rawPred := range *r.Predictions {
			p, _ := pc.Compile(rawPred)
			(*preds) = append((*preds), p)
		}
	}

	return parsedResponse{
		Status:          r.Status,
		Message:         r.Message,
		InternalMessage: r.InternalMessage,
		ErrorCode:       r.ErrorCode,
		Prediction:      pred,
		Predictions:     preds,
		Stored:          stored,
	}
}

type parsedResponse struct {
	Status          int
	Message         string
	InternalMessage string
	ErrorCode       string
	Prediction      *types.Prediction
	Predictions     *[]types.Prediction
	Stored          *bool
}

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
	body := newBody{pred, store}
	bs, err := json.Marshal(body)
	if err != nil {
		log.Println(err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%v/new", c.apiURL), bytes.NewBuffer(bs))
	if err != nil {
		return response{
			Status:          500,
			Message:         "Error creating request against API.",
			InternalMessage: err.Error(),
			ErrorCode:       "ErrCreatingRequestAgainstAPI",
		}.parse()
	}

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return response{
			Status:          500,
			Message:         "Error executing request against API.",
			InternalMessage: err.Error(),
			ErrorCode:       "ErrExecutingRequestAgainstAPI",
		}.parse()
	}
	defer resp.Body.Close()

	byts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err := fmt.Errorf("API returned broken body response! Was: %v", string(byts))
		return response{
			Status:          500,
			Message:         "API returned broken body response.",
			InternalMessage: err.Error(),
			ErrorCode:       "ErrAPIReturnedBrokenBodyResponse",
		}.parse()
	}

	res := response{}
	if err := json.Unmarshal(byts, &res); err != nil {
		err2 := fmt.Errorf("API returned invalid JSON response! Response was: %v. Error is: %v", string(byts), err)
		return response{
			Status:          500,
			Message:         "API returned invalid JSON.",
			InternalMessage: err2.Error(),
			ErrorCode:       "ErrAPIReturnedInvalidJSON",
		}.parse()
	}

	return res.parse()
}

type getFilters struct {
	UUIDs          []string
	authors        []string
	rawStatuses    []string
	rawStateValues []string
}

type getBody struct {
	Filters  types.APIFilters `json:"filters"`
	OrderBys []string         `json:"orderBys"`
}

func (c APIClient) Get(body getBody) parsedResponse {
	bs, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", fmt.Sprintf("%v/get", c.apiURL), bytes.NewBuffer(bs))
	if err != nil {
		return response{
			Status:          500,
			Message:         "Error creating request against API.",
			InternalMessage: err.Error(),
			ErrorCode:       "ErrCreatingRequestAgainstAPI",
		}.parse()
	}

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return response{
			Status:          500,
			Message:         "Error executing request against API.",
			InternalMessage: err.Error(),
			ErrorCode:       "ErrExecutingRequestAgainstAPI",
		}.parse()
	}
	defer resp.Body.Close()

	byts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err := fmt.Errorf("API returned broken body response! Was: %v", string(byts))
		return response{
			Status:          500,
			Message:         "API returned broken body response.",
			InternalMessage: err.Error(),
			ErrorCode:       "ErrAPIReturnedBrokenBodyResponse",
		}.parse()
	}

	res := response{}
	if err := json.Unmarshal(byts, &res); err != nil {
		err2 := fmt.Errorf("API returned invalid JSON response! Response was: %v. Error is: %v", string(byts), err)
		return response{
			Status:          500,
			Message:         "API returned invalid JSON.",
			InternalMessage: err2.Error(),
			ErrorCode:       "ErrAPIReturnedInvalidJSON",
		}.parse()
	}

	return res.parse()
}
