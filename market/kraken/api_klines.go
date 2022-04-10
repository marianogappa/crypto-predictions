package kraken

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/marianogappa/predictions/types"
)

//[1625623260,"34221.6","34221.6","34215.7","34215.7","34215.7","0.25998804",7]
type responseCandlestick = [][]interface{}

type response struct {
	Error  []string               `json:"error"`
	Result map[string]interface{} `json:"result"`
}

func (r response) findDataKey() (string, error) {
	// N.B. if you ever change this logic, please note that BTC is aliased to XBT on Kraken (and who knows what else is)
	// so don't try to find "${baseAsset}${quoteAsset}" here, cause it might not exist.
	for key := range r.Result {
		if key != "last" {
			return key, nil
		}
	}
	return "", errors.New("no data key found")
}

type krakenCandlestick struct {
	timestamp int
	open      float64
	high      float64
	low       float64
	close     float64
	vwap      float64
	volume    float64
	count     int
}

func interfaceToFloatRoundInt(i interface{}) (int, bool) {
	f, ok := i.(float64)
	if !ok {
		return 0, false
	}
	return int(f), true
}

func (r response) getNextSince() (int, error) {
	nextSince, ok := interfaceToFloatRoundInt(r.Result["last"])
	if !ok {
		return nextSince, fmt.Errorf("'next since' was not valid: [%v]! Invalid syntax from Kraken", r.Result["last"])
	}
	return nextSince, nil
}

func (r response) toCandlesticks() ([]types.Candlestick, error) {
	dataKey, err := r.findDataKey()
	if err != nil {
		return []types.Candlestick{}, err
	}
	rawData := r.Result[dataKey]
	rawDataOuterArr, ok := rawData.([]interface{})
	if !ok {
		return []types.Candlestick{}, fmt.Errorf("data key [%v] did not contain an array of datapoints", dataKey)
	}

	candlesticks := make([]types.Candlestick, len(rawDataOuterArr))
	for i := 0; i < len(rawDataOuterArr); i++ {
		raw, ok := rawDataOuterArr[i].([]interface{})
		if !ok {
			return []types.Candlestick{}, fmt.Errorf("candlestick [%v] did not contain an array of data fields, instead: [%v]", i, rawDataOuterArr[i])
		}
		kCandlestick := krakenCandlestick{}

		rawOpenTime, ok := interfaceToFloatRoundInt(raw[0])
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v has non-int open time! Invalid syntax from Kraken", i)
		}
		kCandlestick.timestamp = rawOpenTime

		rawOpen, ok := raw[1].(string)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v has non-string open! Invalid syntax from Kraken", i)
		}
		openPrice, err := strconv.ParseFloat(rawOpen, 64)
		if err != nil {
			return candlesticks, fmt.Errorf("candlestick %v had open = %v! Invalid syntax from Kraken", i, openPrice)
		}
		kCandlestick.open = openPrice
		rawHigh, ok := raw[2].(string)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v has non-string high! Invalid syntax from Kraken", i)
		}
		highPrice, err := strconv.ParseFloat(rawHigh, 64)
		if err != nil {
			return candlesticks, fmt.Errorf("candlestick %v had high = %v! Invalid syntax from Kraken", i, highPrice)
		}
		kCandlestick.high = highPrice

		rawLow, ok := raw[3].(string)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v has non-string low! Invalid syntax from Kraken", i)
		}
		lowPrice, err := strconv.ParseFloat(rawLow, 64)
		if err != nil {
			return candlesticks, fmt.Errorf("candlestick %v had low = %v! Invalid syntax from Kraken", i, lowPrice)
		}
		kCandlestick.low = lowPrice

		rawClose, ok := raw[4].(string)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v has non-string close! Invalid syntax from Kraken", i)
		}
		closePrice, err := strconv.ParseFloat(rawClose, 64)
		if err != nil {
			return candlesticks, fmt.Errorf("candlestick %v had close = %v! Invalid syntax from Kraken", i, closePrice)
		}
		kCandlestick.close = closePrice

		rawVWap, ok := raw[5].(string)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v has non-string vwap! Invalid syntax from Kraken", i)
		}
		vwap, err := strconv.ParseFloat(rawVWap, 64)
		if err != nil {
			return candlesticks, fmt.Errorf("candlestick %v had vwap = %v! Invalid syntax from Kraken", i, vwap)
		}
		kCandlestick.vwap = vwap

		rawVolume, ok := raw[6].(string)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v has non-string volume! Invalid syntax from Kraken", i)
		}
		volume, err := strconv.ParseFloat(rawVolume, 64)
		if err != nil {
			return candlesticks, fmt.Errorf("candlestick %v had volume = %v! Invalid syntax from Kraken", i, volume)
		}
		kCandlestick.volume = volume

		rawCount, ok := interfaceToFloatRoundInt(raw[7])
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v has non-int trade count! Invalid syntax from Kraken", i)
		}
		kCandlestick.count = rawCount

		candlestick := types.Candlestick{
			Timestamp:      kCandlestick.timestamp,
			OpenPrice:      types.JsonFloat64(kCandlestick.open),
			ClosePrice:     types.JsonFloat64(kCandlestick.close),
			LowestPrice:    types.JsonFloat64(kCandlestick.low),
			HighestPrice:   types.JsonFloat64(kCandlestick.high),
			Volume:         types.JsonFloat64(kCandlestick.volume),
			NumberOfTrades: kCandlestick.count,
		}
		candlesticks[i] = candlestick
	}

	return candlesticks, nil
}

type klinesResult struct {
	candlesticks       []types.Candlestick
	err                error
	krakenErrorMessage string
	httpStatus         int
	nextSince          int
}

func (k Kraken) getKlines(baseAsset string, quoteAsset string, startTimeSecs int) (klinesResult, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%vpublic/OHLC", k.apiURL), nil)
	pair := fmt.Sprintf("%v%v", strings.ToUpper(baseAsset), strings.ToUpper(quoteAsset))

	q := req.URL.Query()
	q.Add("pair", pair)
	q.Add("interval", "1")
	q.Add("since", fmt.Sprintf("%v", startTimeSecs))

	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return klinesResult{err: err}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		byts, _ := ioutil.ReadAll(resp.Body)
		err := fmt.Errorf("kraken returned %v status code with payload [%v]", resp.StatusCode, string(byts))
		return klinesResult{httpStatus: 500, err: err}, err
	}

	byts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err := fmt.Errorf("kraken returned broken body response! Was: %v", string(byts))
		return klinesResult{err: err, httpStatus: 500}, err
	}

	maybeResponse := response{}
	err = json.Unmarshal(byts, &maybeResponse)
	if err != nil {
		err := fmt.Errorf("kraken returned invalid JSON response! Was: %v", string(byts))
		return klinesResult{err: err, httpStatus: 500}, err
	}

	if len(maybeResponse.Error) > 0 {
		return klinesResult{
			httpStatus:         500,
			krakenErrorMessage: fmt.Sprintf("%v", maybeResponse.Error),
			err:                fmt.Errorf("kraken returned errors: %v", maybeResponse.Error),
		}, fmt.Errorf("kraken returned errors: %v", maybeResponse.Error)
	}

	candlesticks, err := maybeResponse.toCandlesticks()
	if err != nil {
		wrappedErr := fmt.Errorf("error unmarshalling candlesticks from successful response data from Kraken: %v", err)
		return klinesResult{
			httpStatus: 500,
			err:        wrappedErr,
		}, wrappedErr
	}

	if len(candlesticks) > 0 {
		// N.B. https://docs.kraken.com/rest/#operation/getOHLCData
		// the last entry in the OHLC array is for the current, not-yet-committed frame and will always be present,
		// regardless of the value of since.
		candlesticks = candlesticks[:len(candlesticks)-1]
	}

	nextSince, err := maybeResponse.getNextSince()
	if err != nil {
		wrappedErr := fmt.Errorf("error unmarshalling nextSince from successful response data from Kraken: %v", err)
		return klinesResult{
			httpStatus: 500,
			err:        wrappedErr,
		}, wrappedErr
	}

	if k.debug {
		log.Info().Msgf("Kraken candlestick request successful! Candlestick count: %v\n", len(candlesticks))
	}

	return klinesResult{
		candlesticks: candlesticks,
		nextSince:    nextSince,
		httpStatus:   200,
	}, nil
}
