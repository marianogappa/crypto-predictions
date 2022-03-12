package common

import (
	"time"

	"github.com/marianogappa/predictions/types"
)

type CandlestickIterator struct {
	SavedCandlesticks []types.Candlestick
	next              func() (types.Candlestick, error)
	calmDuration      time.Duration
}

func NewCandlestickIterator(next func() (types.Candlestick, error)) *CandlestickIterator {
	return &CandlestickIterator{next: next, SavedCandlesticks: nil, calmDuration: 1 * time.Second}
}

func (ci *CandlestickIterator) SaveCandlesticks() {
	ci.SavedCandlesticks = []types.Candlestick{}
}

func (ci *CandlestickIterator) Next() (types.Candlestick, error) {
	cs, err := ci.next()
	if ci.SavedCandlesticks != nil && err == nil {
		ci.SavedCandlesticks = append(ci.SavedCandlesticks, cs)
	}
	return cs, err
}

func (ci *CandlestickIterator) GetPriceAt(at types.ISO8601) (types.JsonFloat64, error) {
	rateLimitAttempts := 5
	atTimestamp, err := at.Seconds()
	if err != nil {
		return types.JsonFloat64(0.0), err
	}
	for {
		candlestick, err := ci.next()
		if err == types.ErrRateLimit && rateLimitAttempts > 0 {
			time.Sleep(ci.calmDuration)
			rateLimitAttempts--
			continue
		}
		if err != nil {
			return types.JsonFloat64(0.0), err
		}
		if candlestick.Timestamp < atTimestamp {
			continue
		}
		return candlestick.OpenPrice, nil
	}
}
