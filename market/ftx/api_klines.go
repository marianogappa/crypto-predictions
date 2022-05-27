package ftx

import (
	"encoding/json"
	"errors"
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

func (f FTX) getKlines(baseAsset string, quoteAsset string, startTimeSecs int, intervalMinutes int) (klinesResult, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%vmarkets/%v/%v/candles", f.apiURL, strings.ToUpper(baseAsset), strings.ToUpper(quoteAsset)), nil)
	q := req.URL.Query()

	resolution := intervalMinutes * 60

	validResolutions := map[int]bool{
		15:    true,
		60:    true,
		300:   true,
		900:   true,
		3600:  true,
		14400: true,
		86400: true,
		// All multiples of 86400 up to 30*86400 are actually valid
		// https://docs.ftx.com/#get-historical-prices
		86400 * 7: true,
	}
	if isValid := validResolutions[resolution]; !isValid {
		return klinesResult{}, errors.New("unsupported resolution")
	}

	q.Add("resolution", fmt.Sprintf("%v", resolution))
	q.Add("start_time", fmt.Sprintf("%v", startTimeSecs))

	// N.B.: if you don't supply end_time, or if you supply a very large range, FTX silently ignores this and
	// instead gives you recent data.
	q.Add("end_time", fmt.Sprintf("%v", startTimeSecs+1000*intervalMinutes*60))

	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return klinesResult{err: err}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return klinesResult{httpStatus: 404, err: types.ErrInvalidMarketPair}, types.ErrInvalidMarketPair
	}

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

	if f.debug {
		log.Info().Str("exchange", "FTX").Int("candlestick_count", len(maybeResponse.Result)).Msg("Candlestick request successful!")
	}

	return klinesResult{
		candlesticks: maybeResponse.toCandlesticks(),
		httpStatus:   200,
	}, nil
}

// FTX uses the strategy of having candlesticks on multiples of an hour or a day, and truncating the requested
// millisecond timestamps to the closest mutiple in the future. To test this, use the following snippet:
//
// curl -s "https://ftx.com/api/markets/BTC/USDT/candles?resolution=60&start_time="$(date -j -f "%Y-%m-%d %H:%M:%S" "2020-04-07 00:00:00" "+%s")"&end_time="$(date -j -f "%Y-%m-%d %H:%M:%S" "2020-04-07 00:03:00" "+%s")"" | jq '.result | .[] | .startTime'
//
// Two important caveats for FTX:
//
// 1) if end_time is not specified, start_time is ignored silently and recent data is returned.
// 2) if the range between start_time & end_time is too broad, start_time will be pushed upwards until the range spans 1500 candlesticks.
//
// On the 15 resolution, candlesticks exist at: 00, 15, 30, 45
// On the 60 resolution, candlesticks exist at every minute
// On the 300 resolution, candlesticks exist at: 00, 05, 10 ...
// On the 900 resolution, candlesticks exist at: 00, 15, 30 & 45
// On the 3600 resolution, candlesticks exist at every hour at :00
// On the 14400 resolution, candlesticks exist at: 00:00, 04:00, 08:00 ...
//
// From the 86400 resolution and onwards, FTX is a unique case:
// - it first truncates the date to the beginning of the supplied start_time's day
// - then it returns candlesticks at multiples of the truncated date, starting at that date rather than a prescribed one
