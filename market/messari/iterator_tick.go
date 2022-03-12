package messari

import (
	"github.com/marianogappa/predictions/types"
)

type messariTickIterator struct {
	messari           Messari
	asset, metricID   string
	ticks             []types.Tick
	requestFromMillis int
	initialSeconds    int
	isOutOfTicks      bool
}

func (m Messari) newTickIterator(asset, metricID string, initialISO8601 types.ISO8601) *messariTickIterator {
	// N.B. already validated
	initial, _ := initialISO8601.Time()
	initialSeconds := int(initial.Unix())
	return &messariTickIterator{
		messari:           m,
		asset:             asset,
		metricID:          metricID,
		requestFromMillis: initialSeconds * 1000,
		initialSeconds:    initialSeconds,
	}
}

func (it *messariTickIterator) Next() (types.Tick, error) {
	if len(it.ticks) > 0 {
		c := it.ticks[0]
		it.ticks = it.ticks[1:]
		return c, nil
	}
	result, err := it.messari.getMetrics(it.asset, it.metricID, it.requestFromMillis)
	if err != nil {
		if err == types.ErrOutOfTicks {
			it.isOutOfTicks = true
		}
		return types.Tick{}, err
	}
	it.ticks = result.ticks
	// Some exchanges return earlier ticks to the requested time. Prune them.
	// Note that this may remove all items, but this does not necessarily mean we are out of ticks.
	// In this case we just need to fetch again.
	for len(it.ticks) > 0 && it.ticks[0].Timestamp < it.initialSeconds {
		it.ticks = it.ticks[1:]
	}
	if len(it.ticks) > 0 {
		it.requestFromMillis = (it.ticks[len(it.ticks)-1].Timestamp + 60*60*24) * 1000
	}
	next, err := it.Next()
	if err == types.ErrOutOfTicks {
		it.isOutOfTicks = true
	}
	return next, err
}

func (it *messariTickIterator) IsOutOfTicks() bool {
	return it.isOutOfTicks
}
