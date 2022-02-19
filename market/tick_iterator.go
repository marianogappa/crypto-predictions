package market

import (
	"github.com/marianogappa/signal-checker/common"
)

func buildTickIterator(f func() (common.Candlestick, error)) func() (common.Tick, error) {
	return newTickIterator(f).Next
}

type TickIterator struct {
	f            func() (common.Candlestick, error)
	ticks        []common.Tick
	lastTick     common.Tick
	IsOutOfTicks bool
}

func newTickIterator(f func() (common.Candlestick, error)) *TickIterator {
	return &TickIterator{f: f}
}

// N.B. When next() hits an error, it returns the previous (last) tick.
// This is because when ErrOutOfCandlesticks happens, we want the previous
// candlestick to calculate things for the finished_dataset event.
func (it *TickIterator) Next() (common.Tick, error) {
	if len(it.ticks) > 0 {
		c := it.ticks[0]
		it.ticks = it.ticks[1:]
		it.lastTick = c
		return c, nil
	}
	candlestick, err := it.f()
	if err != nil {
		if err == common.ErrOutOfCandlesticks {
			it.IsOutOfTicks = true
		}
		return it.lastTick, err
	}
	it.ticks = append(it.ticks, candlestick.ToTicks()...)
	next, err := it.Next()
	if err == common.ErrOutOfCandlesticks {
		it.IsOutOfTicks = true
	}
	return next, err
}
