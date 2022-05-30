package binanceusdmfutures

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

type errorResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (r errorResponse) toError() error {
	if r.Code == 0 && r.Msg == "" {
		return nil
	}
	if r.Code == eRRINVALIDSYMBOL {
		return types.ErrInvalidMarketPair
	}
	return fmt.Errorf("binance returned error code! Code: %v, Message: %v", r.Code, r.Msg)
}

// [
//   [
//     1499040000000,      // Open time
//     "0.01634790",       // Open
//     "0.80000000",       // High
//     "0.01575800",       // Low
//     "0.01577100",       // Close
//     "148976.11427815",  // Volume
//     1499644799999,      // Close time
//     "2434.19055334",    // Quote asset volume
//     308,                // Number of trades
//     "1756.87402397",    // Taker buy base asset volume
//     "28.46694368",      // Taker buy quote asset volume
//     "17928899.62484339" // Ignore.
//   ]
// ]
type successfulResponse struct {
	ResponseCandlesticks [][]interface{}
}

func interfaceToFloatRoundInt(i interface{}) (int, bool) {
	f, ok := i.(float64)
	if !ok {
		return 0, false
	}
	return int(f), true
}

func (r successfulResponse) toCandlesticks() ([]types.Candlestick, error) {
	candlesticks := make([]types.Candlestick, len(r.ResponseCandlesticks))
	for i := 0; i < len(r.ResponseCandlesticks); i++ {
		raw := r.ResponseCandlesticks[i]
		candlestick := binanceCandlestick{}
		if len(raw) != 12 {
			return candlesticks, fmt.Errorf("candlestick %v has len != 12! Invalid syntax from Binance", i)
		}
		rawOpenTime, ok := interfaceToFloatRoundInt(raw[0])
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v has non-int open time! Invalid syntax from Binance", i)
		}
		candlestick.openAt = time.Unix(0, int64(rawOpenTime)*int64(time.Millisecond))

		rawOpen, ok := raw[1].(string)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v has non-string open! Invalid syntax from Binance", i)
		}
		openPrice, err := strconv.ParseFloat(rawOpen, 64)
		if err != nil {
			return candlesticks, fmt.Errorf("candlestick %v had open = %v! Invalid syntax from Binance", i, openPrice)
		}
		candlestick.openPrice = openPrice

		rawHigh, ok := raw[2].(string)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v has non-string high! Invalid syntax from Binance", i)
		}
		highPrice, err := strconv.ParseFloat(rawHigh, 64)
		if err != nil {
			return candlesticks, fmt.Errorf("candlestick %v had high = %v! Invalid syntax from Binance", i, highPrice)
		}
		candlestick.highPrice = highPrice

		rawLow, ok := raw[3].(string)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v has non-string low! Invalid syntax from Binance", i)
		}
		lowPrice, err := strconv.ParseFloat(rawLow, 64)
		if err != nil {
			return candlesticks, fmt.Errorf("candlestick %v had low = %v! Invalid syntax from Binance", i, lowPrice)
		}
		candlestick.lowPrice = lowPrice

		rawClose, ok := raw[4].(string)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v has non-string close! Invalid syntax from Binance", i)
		}
		closePrice, err := strconv.ParseFloat(rawClose, 64)
		if err != nil {
			return candlesticks, fmt.Errorf("candlestick %v had close = %v! Invalid syntax from Binance", i, closePrice)
		}
		candlestick.closePrice = closePrice

		rawVolume, ok := raw[5].(string)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v has non-string volume! Invalid syntax from Binance", i)
		}
		volume, err := strconv.ParseFloat(rawVolume, 64)
		if err != nil {
			return candlesticks, fmt.Errorf("candlestick %v had volume = %v! Invalid syntax from Binance", i, volume)
		}
		candlestick.volume = volume

		rawCloseTime, ok := interfaceToFloatRoundInt(raw[6])
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v has non-int close time! Invalid syntax from Binance", i)
		}
		candlestick.closeAt = time.Unix(0, int64(rawCloseTime)*int64(time.Millisecond))

		rawQuoteAssetVolume, ok := raw[7].(string)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v has non-string quote asset volume! Invalid syntax from Binance", i)
		}
		quoteAssetVolume, err := strconv.ParseFloat(rawQuoteAssetVolume, 64)
		if err != nil {
			return candlesticks, fmt.Errorf("candlestick %v had quote asset volume = %v! Invalid syntax from Binance", i, quoteAssetVolume)
		}
		candlestick.quoteAssetVolume = quoteAssetVolume

		rawNumberOfTrades, ok := interfaceToFloatRoundInt(raw[8])
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v has non-int number of trades! Invalid syntax from Binance", i)
		}
		candlestick.tradeCount = rawNumberOfTrades

		rawTakerBaseAssetVolume, ok := raw[9].(string)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v has non-string taker base asset volume! Invalid syntax from Binance", i)
		}
		takerBaseAssetVolume, err := strconv.ParseFloat(rawTakerBaseAssetVolume, 64)
		if err != nil {
			return candlesticks, fmt.Errorf("candlestick %v had taker base asset volume = %v! Invalid syntax from Binance", i, takerBaseAssetVolume)
		}
		candlestick.takerBuyBaseAssetVolume = takerBaseAssetVolume

		rawTakerQuoteAssetVolume, ok := raw[10].(string)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v has non-string taker quote asset volume! Invalid syntax from Binance", i)
		}
		takerBuyQuoteAssetVolume, err := strconv.ParseFloat(rawTakerQuoteAssetVolume, 64)
		if err != nil {
			return candlesticks, fmt.Errorf("candlestick %v had taker quote asset volume = %v! Invalid syntax from Binance", i, takerBuyQuoteAssetVolume)
		}
		candlestick.takerBuyQuoteAssetVolume = takerBuyQuoteAssetVolume

		candlesticks[i] = candlestick.toCandlestick()
	}

	return candlesticks, nil
}

type binanceCandlestick struct {
	openAt                   time.Time
	closeAt                  time.Time
	openPrice                float64
	closePrice               float64
	lowPrice                 float64
	highPrice                float64
	volume                   float64
	quoteAssetVolume         float64
	tradeCount               int
	takerBuyBaseAssetVolume  float64
	takerBuyQuoteAssetVolume float64
}

func (c binanceCandlestick) toCandlestick() types.Candlestick {
	return types.Candlestick{
		Timestamp:      int(c.openAt.Unix()),
		OpenPrice:      types.JsonFloat64(c.openPrice),
		ClosePrice:     types.JsonFloat64(c.closePrice),
		LowestPrice:    types.JsonFloat64(c.lowPrice),
		HighestPrice:   types.JsonFloat64(c.highPrice),
		Volume:         types.JsonFloat64(c.volume),
		NumberOfTrades: c.tradeCount,
	}
}

type klinesResult struct {
	candlesticks        []types.Candlestick
	err                 error
	binanceErrorCode    int
	binanceErrorMessage string
	httpStatus          int
}

func (b BinanceUSDMFutures) getKlines(baseAsset string, quoteAsset string, startTimeMillis int, intervalMinutes int) (klinesResult, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%vklines", b.apiURL), nil)
	symbol := fmt.Sprintf("%v%v", strings.ToUpper(baseAsset), strings.ToUpper(quoteAsset))

	q := req.URL.Query()
	q.Add("symbol", symbol)

	switch intervalMinutes {
	case 1:
		q.Add("interval", "1m")
	case 3:
		q.Add("interval", "3m")
	case 5:
		q.Add("interval", "5m")
	case 15:
		q.Add("interval", "15m")
	case 30:
		q.Add("interval", "30m")
	case 1 * 60:
		q.Add("interval", "1h")
	case 2 * 60:
		q.Add("interval", "2h")
	case 4 * 60:
		q.Add("interval", "4h")
	case 6 * 60:
		q.Add("interval", "6h")
	case 8 * 60:
		q.Add("interval", "8h")
	case 12 * 60:
		q.Add("interval", "12h")
	case 1 * 60 * 24:
		q.Add("interval", "1d")
	case 3 * 60 * 24:
		q.Add("interval", "3d")
	case 7 * 60 * 24:
		q.Add("interval", "1w")
	// TODO This one is problematic because cannot patch holes or do other calculations (because months can have 28, 29, 30 & 31 days)
	case 30 * 60 * 24:
		q.Add("interval", "1M")
	default:
		return klinesResult{}, errors.New("unsupported interval minutes")
	}

	q.Add("limit", "1000")
	q.Add("startTime", fmt.Sprintf("%v", startTimeMillis))

	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return klinesResult{err: err, httpStatus: 500}, err
	}
	defer resp.Body.Close()

	// N.B. commenting this out, because 400 returns valid JSON with error description, which we need!
	// if resp.StatusCode != http.StatusOK {
	// 	err := fmt.Errorf("binance returned %v status code", resp.StatusCode)
	// 	return klinesResult{httpStatus: 500, err: err}, err
	// }

	byts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err := fmt.Errorf("binance returned broken body response! Was: %v", string(byts))
		return klinesResult{err: err, httpStatus: 500}, err
	}

	maybeErrorResponse := errorResponse{}
	err = json.Unmarshal(byts, &maybeErrorResponse)
	errResp := maybeErrorResponse.toError()
	if err == nil && errResp != nil {
		return klinesResult{
			binanceErrorCode:    maybeErrorResponse.Code,
			binanceErrorMessage: maybeErrorResponse.Msg,
			httpStatus:          500,
			err:                 errResp,
		}, errResp
	}

	maybeResponse := successfulResponse{}
	err = json.Unmarshal(byts, &maybeResponse.ResponseCandlesticks)
	if err != nil {
		err := fmt.Errorf("binance returned invalid JSON response! Was: %v", string(byts))
		return klinesResult{err: err, httpStatus: 500}, err
	}

	candlesticks, err := maybeResponse.toCandlesticks()
	if err != nil {
		return klinesResult{
			httpStatus: resp.StatusCode,
			err:        err,
		}, err
	}

	if len(candlesticks) == 0 {
		return klinesResult{
			httpStatus: 200,
			err:        types.ErrOutOfCandlesticks,
		}, types.ErrOutOfCandlesticks
	}

	if b.debug {
		log.Info().Str("exchange", "BinanceUDSMFutures").Int("candlestick_count", len(candlesticks)).Msg("Candlestick request successful!")
	}

	return klinesResult{
		candlesticks: candlesticks,
		httpStatus:   200,
	}, nil
}
