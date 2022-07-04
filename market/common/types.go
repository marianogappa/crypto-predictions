// Package common contains shared interfaces and code across the market super-package.
package common

import (
	"errors"
	"time"

	"github.com/marianogappa/predictions/types"
)

const (
	// BINANCE is an enumesque string value representing the BINANCE exchange
	BINANCE = "binance"
	// FTX is an enumesque string value representing the FTX exchange
	FTX = "ftx"
	// COINBASE is an enumesque string value representing the COINBASE exchange
	COINBASE = "coinbase"
	// HUOBI is an enumesque string value representing the HUOBI exchange
	HUOBI = "huobi"
	// KRAKEN is an enumesque string value representing the KRAKEN exchange
	KRAKEN = "kraken"
	// KUCOIN is an enumesque string value representing the KUCOIN exchange
	KUCOIN = "kucoin"
	// BINANCEUSDMFUTURES is an enumesque string value representing the BINANCEUSDMFUTURES exchange
	BINANCEUSDMFUTURES = "binanceusdmfutures"
	// BITSTAMP is an enumesque string value representing the BITSTAMP exchange
	BITSTAMP = "bitstamp"
)

var (
	// ErrUnsupportedCandlestickInterval means: unsupported candlestick interval
	ErrUnsupportedCandlestickInterval = errors.New("unsupported candlestick interval")

	// ErrExecutingRequest means: error executing client.Do() http request method
	ErrExecutingRequest = errors.New("error executing client.Do() http request method")

	// ErrBrokenBodyResponse means: exchange returned broken body response
	ErrBrokenBodyResponse = errors.New("exchange returned broken body response")

	// ErrInvalidJSONResponse means: exchange returned invalid JSON response
	ErrInvalidJSONResponse = errors.New("exchange returned invalid JSON response")
)

// Exchange is the interface for a crypto exchange.
type Exchange interface {
	CandlestickProvider
	SetDebug(debug bool)
}

// CandlestickProvider wraps a crypto exchanges' API method to retrieve historical candlesticks behind a common
// interface.
type CandlestickProvider interface {
	// RequestCandlesticks requests candlesticks for a given marketPair/asset at a given starting time.
	//
	// Since this is an HTTP request against one of many different exchanges, there's a myriad of things that can go
	// wrong (e.g. Internet out, AWS outage affects exchange, exchange does not honor its API), so implementations do
	// a best-effort of grouping known errors into wrapped errors, but clients must be prepared to handle unknown
	// errors.
	//
	// Resulting candlesticks will start from the given startTimeTs rounded to the next minute or day (respectively for
	// marketPair/asset).
	//
	// Some exchanges return results with gaps. In this case, implementations will fill gaps with the next known value.
	//
	// * Fails with ErrInvalidMarketPair if the operand's marketPair / asset does not exist at the exchange. In some
	//   cases, an exchange may not have data for a marketPair / asset and still not explicitly return an error.
	RequestCandlesticks(operand types.Operand, startTimeTs, intervalMinutes int) ([]types.Candlestick, error)

	// GetPatience documents the recommended latency a client should observe for requesting the latest candlesticks
	// for a given market pair. Clients may ignore it, but are more likely to have to deal with empty results, errors
	// and rate limiting.
	GetPatience() time.Duration
}

// CandlesticksToTicks takes a candlestick slice and turns it into a slice of ticks, using their close prices.
func CandlesticksToTicks(cs []types.Candlestick) []types.Tick {
	ts := make([]types.Tick, len(cs))
	for i := 0; i < len(cs); i++ {
		ts[i] = types.Tick{Timestamp: cs[i].Timestamp, Value: cs[i].ClosePrice}
	}
	return ts
}

// PatchCandlestickHoles takes a slice of candlesticks and it patches any holes in it, either at the beginning or within
// any pair of candlesticks whose difference in seconds doesn't match the supplied "durSecs", by cloning the latest
// available candlestick "on the left", or the first candlestick (i.e. "on the right") if it's at the beginning.
func PatchCandlestickHoles(cs []types.Candlestick, startTimeTs, durSecs int) []types.Candlestick {
	startTimeTs = NormalizeTimestamp(time.Unix(int64(startTimeTs), 0), time.Duration(durSecs)*time.Second, "TODO_PROVIDER", false)
	lastTs := startTimeTs - durSecs
	for len(cs) > 0 && cs[0].Timestamp < lastTs+durSecs {
		cs = cs[1:]
	}
	if len(cs) == 0 {
		return cs
	}

	fixedCSS := []types.Candlestick{}
	for _, candlestick := range cs {
		if candlestick.Timestamp == lastTs+durSecs {
			fixedCSS = append(fixedCSS, candlestick)
			lastTs = candlestick.Timestamp
			continue
		}
		for candlestick.Timestamp >= lastTs+durSecs {
			clonedCandlestick := candlestick
			clonedCandlestick.Timestamp = lastTs + durSecs
			fixedCSS = append(fixedCSS, clonedCandlestick)
			lastTs += durSecs
		}
	}
	return fixedCSS
}

// NormalizeTimestamp takes a time and a candlestick interval, and normalizes the timestamp by returning the immediately
// next multiple of that time as defined by .Truncate(candlestickInterval), unless the time already satisfies it.
//
// It also optionally returns the next time (i.e. it appends a candlestick interval to it).
//
// TODO: this function only currently supports 1m, 5m, 15m, 1h & 1d intervals. Using other intervals will
// result in silently incorrect behaviour due to exchanges behaving differently. Please review api_klines files for
// documented differences in behaviour.
func NormalizeTimestamp(rawTm time.Time, candlestickInterval time.Duration, provider string, startFromNext bool) int {
	rawTm = rawTm.UTC()
	tm := rawTm.Truncate(candlestickInterval).UTC()
	if tm != rawTm {
		tm = tm.Add(candlestickInterval)
	}
	return int(tm.Add(candlestickInterval * time.Duration(b2i(startFromNext))).Unix())
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// CandleReqError is an error arising from a call to requestCandlesticks
type CandleReqError struct {
	Code           int
	Err            error
	IsNotRetryable bool
	IsExchangeSide bool
	RetryAfter     time.Duration
}

func (e CandleReqError) Error() string { return e.Err.Error() }
