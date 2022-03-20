package tickiterator

import (
	"fmt"
	"log"
	"time"

	"github.com/marianogappa/predictions/market/cache"
	"github.com/marianogappa/predictions/market/common"
	"github.com/marianogappa/predictions/types"
)

type TickIteratorImpl struct {
	ticks        []types.Tick
	lastTs       int
	tickCache    *cache.MemoryCache
	operand      types.Operand
	tickProvider common.TickProvider
	timeNowFunc  func() time.Time
}

func NewTickIterator(operand types.Operand, startISO8601 types.ISO8601, tickCache *cache.MemoryCache, tickProvider common.TickProvider, timeNowFunc func() time.Time, startFromNext bool) (*TickIteratorImpl, error) {
	startTm, err := startISO8601.Time()
	if err != nil {
		return nil, cache.ErrInvalidISO8601
	}
	startTs := calculateNormalizedStartingTimestamp(operand, startTm, startFromNext)

	return &TickIteratorImpl{
		operand:      operand,
		tickCache:    tickCache,
		tickProvider: tickProvider,
		ticks:        []types.Tick{},
		timeNowFunc:  timeNowFunc,
		lastTs:       startTs - tickDurationSecs(operand),
	}, nil
}

func (t *TickIteratorImpl) Next() (types.Tick, error) {
	// If the ticks buffer is empty, try to get ticks from the cache.
	if len(t.ticks) == 0 && t.tickCache != nil {
		ticks, err := t.tickCache.Get(t.operand, t.nextISO8601())
		if err == nil {
			t.ticks = ticks
		}
	}

	// If the ticks buffer isn't empty (cache hit), use it.
	if len(t.ticks) > 0 {
		tick := t.ticks[0]
		t.ticks = t.ticks[1:]
		t.lastTs = tick.Timestamp
		return tick, nil
	}

	// If we reach here, before asking the exchange, let's see if it's too early to have new values.
	if t.nextTime().After(t.timeNowFunc().Add(-t.tickProvider.GetPatience())) {
		return types.Tick{}, types.ErrNoNewTicksYet
	}

	// If we reach here, the buffer was empty and the cache was empty too. Last chance: try the exchange.
	ticks, err := t.tickProvider.RequestTicks(t.operand, t.nextTs())
	if err != nil {
		return types.Tick{}, err
	}

	// If the exchange returned early ticks, prune them.
	ticks = t.pruneOlderTicks(ticks)
	if len(ticks) == 0 {
		return types.Tick{}, types.ErrExchangeReturnedNoTicks
	}

	// The first retrieved tick from the exchange must be exactly the required one.
	nextTs := t.nextTs()
	if ticks[0].Timestamp != nextTs {
		expected := time.Unix(int64(nextTs), 0).Format(time.RFC3339)
		actual := time.Unix(int64(ticks[0].Timestamp), 0).Format(time.RFC3339)
		return types.Tick{}, fmt.Errorf("%w: expected %v but got %v", types.ErrExchangeReturnedOutOfSyncTick, expected, actual)
	}

	// Put in the cache for future uses.
	if t.tickCache != nil {
		if err := t.tickCache.Put(t.operand, ticks); err != nil {
			log.Printf("TickIteratorImpl.Next: ignoring error putting into cache: %v\n", err)
		}
	}

	// Also put in the buffer, except for the first tick.
	tick := ticks[0]
	t.ticks = ticks[1:]
	t.lastTs = tick.Timestamp

	// Return the first tick from exchange request.
	return tick, nil
}

func (t *TickIteratorImpl) nextISO8601() types.ISO8601 {
	return types.ISO8601(t.nextTime().Format(time.RFC3339))
}

func (t *TickIteratorImpl) nextTime() time.Time {
	return time.Unix(int64(t.nextTs()), 0)
}

func (t *TickIteratorImpl) nextTs() int {
	return t.lastTs + t.tickDurationSecs()
}

func (t *TickIteratorImpl) tickDurationSecs() int {
	return tickDurationSecs(t.operand)
}

func tickDurationSecs(op types.Operand) int {
	if isMinutely(op) {
		return 60
	}
	// MARKETCAP: 60*60*24
	return 86400
}

func (t *TickIteratorImpl) pruneOlderTicks(ticks []types.Tick) []types.Tick {
	nextTs := t.nextTs()
	for _, tick := range ticks {
		if tick.Timestamp < nextTs {
			ticks = ticks[1:]
		}
	}
	return ticks
}

func calculateNormalizedStartingTimestamp(operand types.Operand, tm time.Time, startFromNext bool) int {
	if isMinutely(operand) {
		if tm.Second() == 0 {
			return int(tm.Unix()) + b2i(startFromNext)*tickDurationSecs(operand)
		}
		year, month, day := tm.Date()
		return int(time.Date(year, month, day, tm.Hour(), tm.Minute()+1+b2i(startFromNext), 0, 0, time.UTC).Unix())
	}

	if tm.Second() == 0 && tm.Minute() == 0 && tm.Hour() == 0 {
		return int(tm.Unix()) + b2i(startFromNext)*tickDurationSecs(operand)
	}
	year, month, day := tm.Date()
	return int(time.Date(year, month, day+1+b2i(startFromNext), 0, 0, 0, 0, time.UTC).Unix())
}

func isMinutely(op types.Operand) bool {
	return op.Type == types.COIN
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}
