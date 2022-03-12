package kucoin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/marianogappa/predictions/types"
)

type response struct {
	Code string     `json:"code"`
	Msg  string     `json:"msg"`
	Data [][]string `json:"data"`
}

type kucoinCandlestick struct {
	Time     int     // Start time of the candle cycle
	Open     float64 // Opening price
	Close    float64 // Closing price
	High     float64 // Highest price
	Low      float64 // Lowest price
	Volume   float64 // Transaction volume
	Turnover float64 // Transaction amount
}

func responseToCandlesticks(data [][]string) ([]types.Candlestick, error) {
	candlesticks := make([]types.Candlestick, len(data))
	for i := 0; i < len(data); i++ {
		raw := data[i]
		candlestick := kucoinCandlestick{}
		if len(raw) != 7 {
			return candlesticks, fmt.Errorf("candlestick %v has len != 7! Invalid syntax from Kucoin", i)
		}
		rawOpenTime, err := strconv.Atoi(raw[0])
		if err != nil {
			return candlesticks, fmt.Errorf("candlestick %v has non-int open time! Err was %v. Invalid syntax from Kucoin", i, err)
		}
		candlestick.Time = rawOpenTime

		rawOpen, err := strconv.ParseFloat(raw[1], 64)
		if err != nil {
			return candlesticks, fmt.Errorf("candlestick %v has non-float open! Err was %v. Invalid syntax from Kucoin", i, err)
		}
		candlestick.Open = rawOpen

		rawClose, err := strconv.ParseFloat(raw[2], 64)
		if err != nil {
			return candlesticks, fmt.Errorf("candlestick %v has non-float close! Err was %v. Invalid syntax from Kucoin", i, err)
		}
		candlestick.Close = rawClose

		rawHigh, err := strconv.ParseFloat(raw[3], 64)
		if err != nil {
			return candlesticks, fmt.Errorf("candlestick %v has non-float high! Err was %v. Invalid syntax from Kucoin", i, err)
		}
		candlestick.High = rawHigh

		rawLow, err := strconv.ParseFloat(raw[4], 64)
		if err != nil {
			return candlesticks, fmt.Errorf("candlestick %v has non-float low! Err was %v. Invalid syntax from Kucoin", i, err)
		}
		candlestick.Low = rawLow

		rawVolume, err := strconv.ParseFloat(raw[5], 64)
		if err != nil {
			return candlesticks, fmt.Errorf("candlestick %v has non-float volume! Err was %v. Invalid syntax from Kucoin", i, err)
		}
		candlestick.Volume = rawVolume

		rawTurnover, err := strconv.ParseFloat(raw[6], 64)
		if err != nil {
			return candlesticks, fmt.Errorf("candlestick %v has non-float turnover! Err was %v. Invalid syntax from Kucoin", i, err)
		}
		candlestick.Turnover = rawTurnover

		candlesticks[i] = types.Candlestick{
			Timestamp:    candlestick.Time,
			OpenPrice:    types.JsonFloat64(candlestick.Open),
			ClosePrice:   types.JsonFloat64(candlestick.Close),
			LowestPrice:  types.JsonFloat64(candlestick.Low),
			HighestPrice: types.JsonFloat64(candlestick.High),
			Volume:       types.JsonFloat64(candlestick.Volume),
		}
	}

	return candlesticks, nil
}

type klinesResult struct {
	candlesticks       []types.Candlestick
	err                error
	kucoinErrorCode    string
	kucoinErrorMessage string
	httpStatus         int
}

func (k Kucoin) getKlines(baseAsset string, quoteAsset string, startTimeSecs int) (klinesResult, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%vmarket/candles", k.apiURL), nil)
	symbol := fmt.Sprintf("%v-%v", strings.ToUpper(baseAsset), strings.ToUpper(quoteAsset))

	q := req.URL.Query()
	q.Add("symbol", symbol)
	q.Add("type", "1min")
	q.Add("startAt", fmt.Sprintf("%v", startTimeSecs))

	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return klinesResult{err: err}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("kucoin returned %v status code", resp.StatusCode)
		return klinesResult{httpStatus: resp.StatusCode, err: err}, err
	}

	byts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err := fmt.Errorf("kucoin returned broken body response! Was: %v", string(byts))
		return klinesResult{err: err, httpStatus: resp.StatusCode}, err
	}

	maybeResponse := response{}
	err = json.Unmarshal(byts, &maybeResponse)
	if err == nil && (maybeResponse.Code != "200000" || maybeResponse.Msg != "") {
		err := fmt.Errorf("kucoin returned error code! Code: %v, Message: %v", maybeResponse.Code, maybeResponse.Msg)
		return klinesResult{
			kucoinErrorCode:    maybeResponse.Code,
			kucoinErrorMessage: maybeResponse.Msg,
			httpStatus:         500,
		}, err
	}
	if err != nil {
		err := fmt.Errorf("kucoin returned invalid JSON response! Was: %v", string(byts))
		return klinesResult{err: err, httpStatus: 500}, err
	}

	candlesticks, err := responseToCandlesticks(maybeResponse.Data)
	if err != nil {
		return klinesResult{
			httpStatus: 500,
			err:        err,
		}, err
	}

	return klinesResult{
		candlesticks: candlesticks,
		httpStatus:   200,
	}, nil
}
