package market

import (
	"github.com/marianogappa/predictions/types"
)

type TickFromCandleIterator struct {
	f            func() (types.Candlestick, error)
	ticks        []types.Tick
	lastTick     types.Tick
	isOutOfTicks bool
}

func newTickFromCandleIterator(f func() (types.Candlestick, error)) *TickFromCandleIterator {
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
		if err == types.ErrOutOfCandlesticks {
			it.isOutOfTicks = true
			err = types.ErrOutOfTicks
		}
		return it.lastTick, err
	}
	it.ticks = append(it.ticks, candlestick.ToTicks()...)
	next, err := it.Next()
	if err == types.ErrOutOfTicks {
		it.isOutOfTicks = true
	}
	return next, err
}

func (it *TickFromCandleIterator) IsOutOfTicks() bool {
	return it.isOutOfTicks
}
