// The common package contains the input and output types of the signal checking function.
package common

import (
	"time"

	"github.com/marianogappa/predictions/types"
)

const (
	BINANCE              = "binance"
	FTX                  = "ftx"
	COINBASE             = "coinbase"
	HUOBI                = "huobi"
	KRAKEN               = "kraken"
	KUCOIN               = "kucoin"
	BINANCE_USDM_FUTURES = "binanceusdmfutures"
)

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

func CandlesticksToTicks(cs []types.Candlestick) []types.Tick {
	ts := make([]types.Tick, len(cs))
	for i := 0; i < len(cs); i++ {
		ts[i] = types.Tick{Timestamp: cs[i].Timestamp, Value: cs[i].ClosePrice}
	}
	return ts
}

func PatchCandlestickHoles(cs []types.Candlestick, startTimeTs, durSecs int) []types.Candlestick {
	startTimeTs = NormalizeTimestamp(time.Unix(int64(startTimeTs), 0), time.Duration(durSecs)*time.Second, "TODO_PROVIDER", false)
	lastTs := startTimeTs - durSecs
	for len(cs) > 0 && cs[0].Timestamp < lastTs+durSecs {
		cs = cs[1:]
	}
	if len(cs) == 0 {
		return cs
	}

	fixedCss := []types.Candlestick{}
	for _, candlestick := range cs {
		if candlestick.Timestamp == lastTs+durSecs {
			fixedCss = append(fixedCss, candlestick)
			lastTs = candlestick.Timestamp
			continue
		}
		for candlestick.Timestamp >= lastTs+durSecs {
			clonedCandlestick := candlestick
			clonedCandlestick.Timestamp = lastTs + durSecs
			fixedCss = append(fixedCss, clonedCandlestick)
			lastTs += durSecs
		}
	}
	return fixedCss
}

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
