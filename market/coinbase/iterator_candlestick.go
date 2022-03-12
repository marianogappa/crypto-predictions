package coinbase

import (
	"time"

	"github.com/marianogappa/predictions/types"
)

type coinbaseCandlestickIterator struct {
	coinbase              Coinbase
	baseAsset, quoteAsset string
	candlesticks          []types.Candlestick
	requestFromTime       time.Time
	initialSeconds        int
}

func (c Coinbase) newCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 types.ISO8601) *coinbaseCandlestickIterator {
	// N.B. already validated
	initial, _ := initialISO8601.Time()
	initialSeconds := int(initial.Unix())
	return &coinbaseCandlestickIterator{
		coinbase:        c,
		baseAsset:       baseAsset,
		quoteAsset:      quoteAsset,
		requestFromTime: initial,
		initialSeconds:  initialSeconds,
	}
}

func (it *coinbaseCandlestickIterator) next() (types.Candlestick, error) {
	if len(it.candlesticks) > 0 {
		// N.B. Coinbase returns data in descending order
		c := it.candlesticks[len(it.candlesticks)-1]
		it.candlesticks = it.candlesticks[:len(it.candlesticks)-1]
		return c, nil
	}
	if it.requestFromTime.After(time.Now().Add(-1 * time.Minute)) {
		return types.Candlestick{}, types.ErrOutOfCandlesticks
	}
	startTimeISO8601 := it.requestFromTime.Format(time.RFC3339)
	endTimeISO8601 := it.requestFromTime.Add(299 * 60 * time.Second).Format(time.RFC3339)

	klinesResult, err := it.coinbase.getKlines(it.baseAsset, it.quoteAsset, startTimeISO8601, endTimeISO8601)
	if err != nil {
		return types.Candlestick{}, err
	}
	it.candlesticks = klinesResult.candlesticks
	if len(it.candlesticks) == 0 {
		return types.Candlestick{}, types.ErrOutOfCandlesticks
	}
	// Some exchanges return earlier candlesticks to the requested time. Prune them.
	// Note that this may remove all items, but this does not necessarily mean we are out of candlesticks.
	// In this case we just need to fetch again.
	for len(it.candlesticks) > 0 && it.candlesticks[len(it.candlesticks)-1].Timestamp < it.initialSeconds {
		it.candlesticks = it.candlesticks[:len(it.candlesticks)-1]
	}
	if len(it.candlesticks) > 0 {
		it.requestFromTime = it.requestFromTime.Add(299 * 60 * time.Second)
	}
	return it.next()
}
