package market

import (
	"github.com/marianogappa/predictions/types"
	"github.com/marianogappa/signal-checker/common"
)

type TickFromCandleIterator struct {
	f            func() (common.Candlestick, error)
	ticks        []types.Tick
	lastTick     types.Tick
	isOutOfTicks bool
}

func newTickFromCandleIterator(f func() (common.Candlestick, error)) *TickFromCandleIterator {
	return &TickFromCandleIterator{f: f}
}

// N.B. When next() hits an error, it returns the previous (last) tick.
// This is because when ErrOutOfCandlesticks happens, we want the previous
// candlestick to calculate things for the finished_dataset event.
func (it *TickFromCandleIterator) Next() (types.Tick, error) {
	if len(it.ticks) > 0 {
		c := it.ticks[0]
		it.ticks = it.ticks[1:]
		it.lastTick = c
		return c, nil
	}
	candlestick, err := it.f()
	if err != nil {
		if err == common.ErrOutOfCandlesticks {
			it.isOutOfTicks = true
		}
		return it.lastTick, err
	}
	it.ticks = append(it.ticks, commonTicksToTypeTicks(candlestick.ToTicks())...)
	next, err := it.Next()
	if err == types.ErrOutOfTicks {
		it.isOutOfTicks = true
	}
	return next, err
}

func (it *TickFromCandleIterator) IsOutOfTicks() bool {
	return it.isOutOfTicks
}

func commonTicksToTypeTicks(ts []common.Tick) []types.Tick {
	res := []types.Tick{}
	for _, t := range ts {
		res = append(res, types.Tick{
			Timestamp: t.Timestamp,
			Value:     t.Price,
		})
	}
	return res
}
