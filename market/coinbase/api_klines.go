package coinbase

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

type successResponse = [][]interface{}

type errorResponse struct {
	Message string `json:"message"`
}

func coinbaseToCandlesticks(response successResponse) ([]types.Candlestick, error) {
	candlesticks := make([]types.Candlestick, len(response))
	for i := 0; i < len(response); i++ {
		raw := response[i]
		timestampFloat, ok := raw[0].(float64)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v had timestampMillis = %v! Invalid syntax from Coinbase", i, timestampFloat)
		}
		timestamp := int(timestampFloat)
		lowestPrice, ok := raw[1].(float64)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v had lowestPrice = %v! Invalid syntax from Coinbase", i, lowestPrice)
		}
		highestPrice, ok := raw[2].(float64)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v had highestPrice = %v! Invalid syntax from Coinbase", i, highestPrice)
		}
		openPrice, ok := raw[3].(float64)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v had openPrice = %v! Invalid syntax from Coinbase", i, openPrice)
		}
		closePrice, ok := raw[4].(float64)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v had closePrice = %v! Invalid syntax from Coinbase", i, closePrice)
		}
		volume, ok := raw[5].(float64)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v had volume = %v! Invalid syntax from Coinbase", i, volume)
		}

		candlestick := types.Candlestick{
			Timestamp:    timestamp,
			LowestPrice:  types.JsonFloat64(lowestPrice),
			HighestPrice: types.JsonFloat64(highestPrice),
			OpenPrice:    types.JsonFloat64(openPrice),
			ClosePrice:   types.JsonFloat64(closePrice),
			Volume:       types.JsonFloat64(volume),
		}
		candlesticks[i] = candlestick
	}

	return candlesticks, nil
}

type klinesResult struct {
	candlesticks         []types.Candlestick
	err                  error
	coinbaseErrorMessage string
	httpStatus           int
}

func (c Coinbase) getKlines(baseAsset string, quoteAsset string, startTimeISO8601, endTimeISO8601 string, intervalMinutes int) (klinesResult, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%vproducts/%v-%v/candles", c.apiURL, strings.ToUpper(baseAsset), strings.ToUpper(quoteAsset)), nil)

	q := req.URL.Query()

	granularity := intervalMinutes * 60

	validGranularities := map[int]bool{
		60:    true,
		300:   true,
		900:   true,
		3600:  true,
		21600: true,
		86400: true,
	}
	if isValid := validGranularities[granularity]; !isValid {
		return klinesResult{}, errors.New("unsupported resolution")
	}

	q.Add("granularity", fmt.Sprintf("%v", granularity))
	q.Add("start", fmt.Sprintf("%v", startTimeISO8601))
	q.Add("end", fmt.Sprintf("%v", endTimeISO8601))

	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return klinesResult{err: err}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		byts, _ := ioutil.ReadAll(resp.Body)
		err := fmt.Errorf("coinbase returned %v status code with payload [%v]", resp.StatusCode, string(byts))
		return klinesResult{httpStatus: 500, err: err}, err
	}

	byts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err := fmt.Errorf("coinbase returned broken body response! Was: %v", string(byts))
		return klinesResult{err: err, httpStatus: 500}, err
	}

	maybeErrorResponse := errorResponse{}
	err = json.Unmarshal(byts, &maybeErrorResponse)
	if err == nil && (maybeErrorResponse.Message != "") {
		err := fmt.Errorf("coinbase returned error code! Message: %v", maybeErrorResponse.Message)
		return klinesResult{
			coinbaseErrorMessage: maybeErrorResponse.Message,
			httpStatus:           500,
		}, err
	}

	maybeResponse := successResponse{}
	err = json.Unmarshal(byts, &maybeResponse)
	if err != nil {
		err := fmt.Errorf("coinbase returned invalid JSON response! Was: %v", string(byts))
		return klinesResult{err: err, httpStatus: 500}, err
	}

	candlesticks, err := coinbaseToCandlesticks(maybeResponse)
	if err != nil {
		return klinesResult{
			httpStatus: 500,
			err:        fmt.Errorf("error unmarshalling successful JSON response from Coinbase: %v", err),
		}, err
	}

	if c.debug {
		log.Info().Msgf("Coinbase candlestick request successful! Candlestick count: %v\n", len(candlesticks))
	}

	return klinesResult{
		candlesticks: candlesticks,
		httpStatus:   200,
	}, nil
}

// Coinbase uses the strategy of having candlesticks on multiples of an hour or a day, and truncating the requested
// millisecond timestamps to the closest mutiple in the future. To test this, use the following snippet:
//
// curl -s "https://api.pro.coinbase.com/products/BTC-USD/candles?granularity=60&start=2022-01-16T10:45:24Z&end=2022-01-16T10:59:24Z" | jq '.[] | .[0] | todate'
//
// Note that if `end` - `start` / `granularity` > 300, rather than failing silently, the following error will be
// returned (which is great):
//
// {"message":"granularity too small for the requested time range. Count of aggregations requested exceeds 300"}
//
//
// On the 60 resolution, candlesticks exist at every minute
// On the 300 resolution, candlesticks exist at: 00, 05, 10 ...
// On the 900 resolution, candlesticks exist at: 00, 15, 30 & 45
// On the 3600 resolution, candlesticks exist at every hour
// On the 21600 resolution, candlesticks exist at: 00:00, 06:00, 12:00 & 18:00
// On the 86400 resolution, candlesticks exist at every day at 00:00:00
