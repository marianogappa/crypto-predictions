package kucoin

import (
	"github.com/marianogappa/predictions/types"
)

type kucoinCandlestickIterator struct {
	kucoin                Kucoin
	baseAsset, quoteAsset string
	candlesticks          []types.Candlestick
	requestFromSecs       int
}

func (k Kucoin) newCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 types.ISO8601) *kucoinCandlestickIterator {
	// N.B. already validated
	initial, _ := initialISO8601.Time()
	return &kucoinCandlestickIterator{
		kucoin:          k,
		baseAsset:       baseAsset,
		quoteAsset:      quoteAsset,
		requestFromSecs: int(initial.Unix()),
	}
}

func (it *kucoinCandlestickIterator) next() (types.Candlestick, error) {
	if len(it.candlesticks) > 0 {
		// N.B. KuCoin returns data in descending order
		c := it.candlesticks[len(it.candlesticks)-1]
		it.candlesticks = it.candlesticks[:len(it.candlesticks)-1]
		return c, nil
	}
	klinesResult, err := it.kucoin.getKlines(it.baseAsset, it.quoteAsset, it.requestFromSecs)
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
	for len(it.candlesticks) > 0 && it.candlesticks[len(it.candlesticks)-1].Timestamp < it.requestFromSecs {
		it.candlesticks = it.candlesticks[:len(it.candlesticks)-1]
	}
	if len(it.candlesticks) > 0 {
		it.requestFromSecs = it.candlesticks[0].Timestamp + 60
	}

	return it.next()
}
