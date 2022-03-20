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
	TickProvider
	SetDebug(debug bool)
}

// TickProvider wraps a crypto exchanges' API method to retrieve historical candlesticks behind a common interface.
type TickProvider interface {
	// RequestTicks requests ticks for a given marketPair/asset at a given starting time.
	//
	// Since this is an HTTP request against one of many different exchanges, there's a myriad of things that can go
	// wrong (e.g. Internet out, AWS outage affects exchange, exchange does not honor its API), so implementations do
	// a best-effort of grouping known errors into wrapped errors, but clients must be prepared to handle unknown
	// errors.
	//
	// Resulting ticks will start from the given startTimeTs rounded to the next minute or day (respectively for
	// marketPair/asset).
	//
	// Some exchanges return results with gaps. In this case, implementations will fill gaps with the next known value.
	//
	// * Fails with ErrInvalidMarketPair if the operand's marketPair / asset does not exist at the exchange. In some
	//   cases, an exchange may not have data for a marketPair / asset and still not explicitly return an error.
	RequestTicks(operand types.Operand, startTimeTs int) ([]types.Tick, error)

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

func PatchTickHoles(ts []types.Tick, startTimeTs, durSecs int) []types.Tick {
	startTimeTs = calculateNormalizedStartingTimestamp(startTimeTs, durSecs)
	lastTs := startTimeTs - durSecs
	for len(ts) > 0 && ts[0].Timestamp < lastTs+durSecs {
		ts = ts[1:]
	}
	if len(ts) == 0 {
		return ts
	}

	fixedTs := []types.Tick{}
	for _, tick := range ts {
		if tick.Timestamp == lastTs+durSecs {
			fixedTs = append(fixedTs, tick)
			lastTs = tick.Timestamp
			continue
		}
		for tick.Timestamp >= lastTs+durSecs {
			fixedTs = append(fixedTs, types.Tick{Timestamp: lastTs + durSecs, Value: tick.Value})
			lastTs += durSecs
		}
	}
	return fixedTs
}

func calculateNormalizedStartingTimestamp(startTimeTs, durSecs int) int {
	tm := time.Unix(int64(startTimeTs), 0)
	if isMinutely(durSecs) {
		if tm.Second() == 0 {
			return int(tm.Unix())
		}
		year, month, day := tm.Date()

		return int(time.Date(year, month, day, tm.Hour(), tm.Minute()+1, 0, 0, time.UTC).Unix())
	}

	if tm.Second() == 0 && tm.Minute() == 0 && tm.Hour() == 0 {
		return int(tm.Unix())
	}
	year, month, day := tm.Date()
	return int(time.Date(year, month, day+1, 0, 0, 0, 0, time.UTC).Unix())
}

func isMinutely(durSecs int) bool {
	return durSecs == 60
}
