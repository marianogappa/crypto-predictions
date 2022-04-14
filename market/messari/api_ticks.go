package messari

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/marianogappa/predictions/types"
	"github.com/rs/zerolog/log"
)

type responseStatus struct {
	Elapsed      int    `json:"elapsed"`
	Timestamp    string `json:"timestamp"`
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

type responseData struct {
	Values [][]interface{} `json:"values"`
}

// [
//       [
//         1577836800000,
//         130044373322.33379
//       ]
// ]
type response struct {
	Status responseStatus `json:"status"`
	Data   responseData   `json:"data"`
}

func interfaceToFloatRoundInt(i interface{}) (int, bool) {
	f, ok := i.(float64)
	if !ok {
		return 0, false
	}
	return int(f), true
}

func interfaceToFloat(i interface{}) (float64, bool) {
	f, ok := i.(float64)
	if !ok {
		return 0, false
	}
	return f, true
}

func (r response) toCandlesticks() ([]types.Candlestick, error) {
	candlesticks := make([]types.Candlestick, len(r.Data.Values))
	for i := 0; i < len(r.Data.Values); i++ {
		raw := r.Data.Values[i]
		candlestick := types.Candlestick{}
		if len(raw) != 2 {
			return candlesticks, fmt.Errorf("candlestick %v has len != 2! Invalid syntax from messari", i)
		}
		rawTimestampMillis, ok := interfaceToFloatRoundInt(raw[0])
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v has non-int timestamp! Invalid syntax from messari", i)
		}
		candlestick.Timestamp = rawTimestampMillis / 1000
		price, ok := interfaceToFloat(raw[1])
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v has non-float price! Invalid syntax from messari", i)
		}
		candlestick.OpenPrice = types.JsonFloat64(price)
		candlestick.HighestPrice = types.JsonFloat64(price)
		candlestick.LowestPrice = types.JsonFloat64(price)
		candlestick.ClosePrice = types.JsonFloat64(price)
		candlesticks[i] = candlestick
	}

	return candlesticks, nil
}

type metricsResult struct {
	candlesticks        []types.Candlestick
	err                 error
	messariErrorCode    int
	messariErrorMessage string
	httpStatus          int
}

func (b Messari) getMetrics(asset, metricID string, startTimeMillis int) (metricsResult, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%vassets/%v/metrics/%v/time-series", b.apiURL, asset, metricID), nil)

	after := time.Unix(int64(startTimeMillis/1000), 0).Format("2006-01-02")

	if b.debug {
		log.Info().Msgf("Running time-series request against Messari API for asset %v after %v...\n", asset, after)
	}

	q := req.URL.Query()
	q.Add("after", after)
	q.Add("interval", "1d")
	q.Add("order", "ascending")
	q.Add("startTime", fmt.Sprintf("%v", startTimeMillis))

	req.Header.Add("x-messari-api-key", b.apiKey)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return metricsResult{err: err, httpStatus: 500}, err
	}
	defer resp.Body.Close()

	// N.B. commenting this out, because 400 returns valid JSON with error description, which we need!
	// if resp.StatusCode != http.StatusOK {
	// 	err := fmt.Errorf("messari returned %v status code", resp.StatusCode)
	// 	return metricsResult{httpStatus: 500, err: err}, err
	// }

	byts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err := fmt.Errorf("messari returned broken body response! Was: %v", string(byts))
		return metricsResult{err: err, httpStatus: 500}, err
	}
	log.Info().Msgf("Messari API response: %v", string(byts))

	res := response{}
	err = json.Unmarshal(byts, &res)
	if err != nil {
		err := fmt.Errorf("messari returned invalid JSON response! Was: %v", string(byts))
		return metricsResult{err: err, httpStatus: 500}, err
	}

	if res.Status.ErrorCode != 0 {
		err := fmt.Errorf("error from Messari API: %v", res.Status.ErrorMessage)
		return metricsResult{
			httpStatus:          500,
			messariErrorCode:    res.Status.ErrorCode,
			messariErrorMessage: res.Status.ErrorMessage,
			err:                 err,
		}, err
	}

	candlesticks, err := res.toCandlesticks()
	if err != nil {
		return metricsResult{
			httpStatus: resp.StatusCode,
			err:        err,
		}, err
	}

	if len(candlesticks) == 0 {
		return metricsResult{
			httpStatus: 200,
			err:        types.ErrOutOfTicks,
		}, types.ErrOutOfTicks
	}

	if b.debug {
		log.Info().Msgf("messari tick request successful! Candlestick count: %v\n", len(candlesticks))
	}

	return metricsResult{
		candlesticks: candlesticks,
		httpStatus:   200,
	}, nil
}
