package messari

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/marianogappa/predictions/market/common"
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
		candlestick.OpenPrice = types.JSONFloat64(price)
		candlestick.HighestPrice = types.JSONFloat64(price)
		candlestick.LowestPrice = types.JSONFloat64(price)
		candlestick.ClosePrice = types.JSONFloat64(price)
		candlesticks[i] = candlestick
	}

	return candlesticks, nil
}

func (e *Messari) requestCandlesticks(asset, metricID string, startTimeMillis int, _unused int) ([]types.Candlestick, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%vassets/%v/metrics/%v/time-series", e.apiURL, asset, metricID), nil)

	after := time.Unix(int64(startTimeMillis/1000), 0).Format("2006-01-02")

	if e.debug {
		log.Info().Msgf("Running time-series request against Messari API for asset %v after %v...\n", asset, after)
	}

	q := req.URL.Query()
	q.Add("after", after)
	q.Add("interval", "1d")
	q.Add("order", "ascending")
	q.Add("startTime", fmt.Sprintf("%v", startTimeMillis))

	req.Header.Add("x-messari-api-key", e.apiKey)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return nil, common.CandleReqError{IsNotRetryable: true, Err: common.ErrExecutingRequest}
	}
	defer resp.Body.Close()

	byts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, common.CandleReqError{IsNotRetryable: false, IsExchangeSide: true, Err: common.ErrBrokenBodyResponse}
	}

	if e.debug {
		log.Info().Msgf("Messari API response: %v", string(byts))
	}

	res := response{}
	err = json.Unmarshal(byts, &res)
	if err != nil {
		return nil, common.CandleReqError{IsNotRetryable: false, IsExchangeSide: true, Err: common.ErrInvalidJSONResponse}
	}

	if res.Status.ErrorCode == 404 && strings.HasPrefix(res.Status.ErrorMessage, "Asset with key = ") && strings.HasSuffix(res.Status.ErrorMessage, " not found.") {
		return nil, common.CandleReqError{IsNotRetryable: true, IsExchangeSide: true, Err: types.ErrInvalidMarketPair}
	}

	if res.Status.ErrorCode != 0 {
		return nil, common.CandleReqError{
			IsNotRetryable: false,
			IsExchangeSide: true,
			Err:            fmt.Errorf("error from Messari API: %v", res.Status.ErrorMessage),
			Code:           res.Status.ErrorCode,
		}
	}

	candlesticks, err := res.toCandlesticks()
	if err != nil {
		return nil, common.CandleReqError{IsNotRetryable: false, IsExchangeSide: true, Err: err}
	}

	if len(candlesticks) == 0 {
		return nil, common.CandleReqError{IsNotRetryable: true, IsExchangeSide: true, Err: types.ErrOutOfCandlesticks}
	}

	if e.debug {
		log.Info().Str("exchange", "Messari").Str("asset", fmt.Sprintf("%v", asset)).Int("candlestick_count", len(candlesticks)).Msg("Candlestick request successful!")
	}

	return candlesticks, nil
}
