package ftx

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/marianogappa/predictions/types"
)

//[
//	{
//		"startTime":"2021-07-05T18:20:00+00:00",
//		"time":1625509200000.0,
//		"open":33831.0,
//		"high":33837.0,
//		"low":33810.0,
//		"close":33837.0,
//		"volume":11679.9302
//	}
//]
type responseCandlestick struct {
	StartTime string  `json:"startTime"`
	Time      float64 `json:"time"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}

type response struct {
	Success bool                  `json:"success"`
	Error   string                `json:"error"`
	Result  []responseCandlestick `json:"result"`
}

func (r response) toCandlesticks() []types.Candlestick {
	candlesticks := make([]types.Candlestick, len(r.Result))
	for i := 0; i < len(r.Result); i++ {
		raw := r.Result[i]
		candlestick := types.Candlestick{
			Timestamp:    int(raw.Time) / 1000,
			OpenPrice:    types.JsonFloat64(raw.Open),
			ClosePrice:   types.JsonFloat64(raw.Close),
			LowestPrice:  types.JsonFloat64(raw.Low),
			HighestPrice: types.JsonFloat64(raw.High),
			Volume:       types.JsonFloat64(raw.Volume),
		}
		candlesticks[i] = candlestick
	}

	return candlesticks
}

type klinesResult struct {
	candlesticks    []types.Candlestick
	err             error
	ftxErrorMessage string
	httpStatus      int
}

func (f FTX) getKlines(baseAsset string, quoteAsset string, startTimeSecs int) (klinesResult, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%vmarkets/%v/%v/candles", f.apiURL, strings.ToUpper(baseAsset), strings.ToUpper(quoteAsset)), nil)
	q := req.URL.Query()
	q.Add("resolution", "60")
	q.Add("start_time", fmt.Sprintf("%v", startTimeSecs))

	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return klinesResult{err: err}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		byts, _ := ioutil.ReadAll(resp.Body)
		err := fmt.Errorf("ftx returned %v status code with payload [%v]", resp.StatusCode, string(byts))
		return klinesResult{httpStatus: 500, err: err}, err
	}

	byts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err := fmt.Errorf("ftx returned broken body response! Was: %v", string(byts))
		return klinesResult{err: err, httpStatus: 500}, err
	}

	maybeResponse := response{}
	err = json.Unmarshal(byts, &maybeResponse)
	if err != nil {
		err := fmt.Errorf("ftx returned invalid JSON response! Was: %v", string(byts))
		return klinesResult{err: err, httpStatus: 500}, err
	}

	if !maybeResponse.Success {
		return klinesResult{
			httpStatus:      500,
			ftxErrorMessage: maybeResponse.Error,
			err:             fmt.Errorf("FTX returned error: %v", maybeResponse.Error),
		}, err
	}

	log.Info().Msgf("FTX candlestick request successful! Candlestick count: %v\n", len(maybeResponse.Result))

	return klinesResult{
		candlesticks: maybeResponse.toCandlesticks(),
		httpStatus:   200,
	}, nil
}
