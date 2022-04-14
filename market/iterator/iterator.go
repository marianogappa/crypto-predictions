package iterator

import (
	"fmt"
	"time"

	"github.com/marianogappa/predictions/market/cache"
	"github.com/marianogappa/predictions/market/common"
	"github.com/marianogappa/predictions/types"
	"github.com/rs/zerolog/log"
)

type IteratorImpl struct {
	candlesticks        []types.Candlestick
	lastTs              int
	candlestickCache    *cache.MemoryCache
	operand             types.Operand
	candlestickProvider common.CandlestickProvider
	timeNowFunc         func() time.Time
}

func NewIterator(operand types.Operand, startISO8601 types.ISO8601, candlestickCache *cache.MemoryCache, candlestickProvider common.CandlestickProvider, timeNowFunc func() time.Time, startFromNext bool) (*IteratorImpl, error) {
	startTm, err := startISO8601.Time()
	if err != nil {
		return nil, cache.ErrInvalidISO8601
	}
	startTs := calculateNormalizedStartingTimestamp(operand, startTm, startFromNext)

	return &IteratorImpl{
		operand:             operand,
		candlestickCache:    candlestickCache,
		candlestickProvider: candlestickProvider,
		candlesticks:        []types.Candlestick{},
		timeNowFunc:         timeNowFunc,
		lastTs:              startTs - candlestickDurationSecs(operand),
	}, nil
}

func (t *IteratorImpl) NextTick() (types.Tick, error) {
	cs, err := t.NextCandlestick()
	if err != nil {
		return types.Tick{}, err
	}
	return types.Tick{Timestamp: cs.Timestamp, Value: cs.ClosePrice}, nil
}

func (t *IteratorImpl) NextCandlestick() (types.Candlestick, error) {
	// If the candlesticks buffer is empty, try to get candlesticks from the cache.
	if len(t.candlesticks) == 0 && t.candlestickCache != nil {
		ticks, err := t.candlestickCache.Get(t.operand, t.nextISO8601())
		if err == nil {
			t.candlesticks = ticks
		}
	}

	// If the ticks buffer isn't empty (cache hit), use it.
	if len(t.candlesticks) > 0 {
		candlestick := t.candlesticks[0]
		t.candlesticks = t.candlesticks[1:]
		t.lastTs = candlestick.Timestamp
		return candlestick, nil
	}

	// If we reach here, before asking the exchange, let's see if it's too early to have new values.
	if t.nextTime().After(t.timeNowFunc().Add(-t.candlestickProvider.GetPatience() - time.Duration(t.candlestickDurationSecs())*time.Second)) {
		return types.Candlestick{}, types.ErrNoNewTicksYet
	}

	// If we reach here, the buffer was empty and the cache was empty too. Last chance: try the exchange.
	candlesticks, err := t.candlestickProvider.RequestCandlesticks(t.operand, t.nextTs())
	if err != nil {
		return types.Candlestick{}, err
	}

	// If the exchange returned early candlesticks, prune them.
	candlesticks = t.pruneOlderCandlesticks(candlesticks)
	if len(candlesticks) == 0 {
		return types.Candlestick{}, types.ErrExchangeReturnedNoTicks
	}

	// The first retrieved candlestick from the exchange must be exactly the required one.
	nextTs := t.nextTs()
	if candlesticks[0].Timestamp != nextTs {
		expected := time.Unix(int64(nextTs), 0).Format(time.RFC3339)
		actual := time.Unix(int64(candlesticks[0].Timestamp), 0).Format(time.RFC3339)
		return types.Candlestick{}, fmt.Errorf("%w: expected %v but got %v", types.ErrExchangeReturnedOutOfSyncTick, expected, actual)
	}

	// Put in the cache for future uses.
	if t.candlestickCache != nil {
		if err := t.candlestickCache.Put(t.operand, candlesticks); err != nil {
			log.Info().Msgf("IteratorImpl.Next: ignoring error putting into cache: %v\n", err)
		}
	}

	// Also put in the buffer, except for the first candlestick.
	candlestick := candlesticks[0]
	t.candlesticks = candlesticks[1:]
	t.lastTs = candlestick.Timestamp

	// Return the first candlestick from exchange request.
	return candlestick, nil
}

func (t *IteratorImpl) nextISO8601() types.ISO8601 {
	return types.ISO8601(t.nextTime().Format(time.RFC3339))
}

func (t *IteratorImpl) nextTime() time.Time {
	return time.Unix(int64(t.nextTs()), 0)
}

func (t *IteratorImpl) nextTs() int {
	return t.lastTs + t.candlestickDurationSecs()
}

func (t *IteratorImpl) candlestickDurationSecs() int {
	return candlestickDurationSecs(t.operand)
}

func candlestickDurationSecs(op types.Operand) int {
	if isMinutely(op) {
		return 60
	}
	// MARKETCAP: 60*60*24
	return 86400
}

func (t *IteratorImpl) pruneOlderCandlesticks(candlesticks []types.Candlestick) []types.Candlestick {
	nextTs := t.nextTs()
	for _, tick := range candlesticks {
		if tick.Timestamp < nextTs {
			candlesticks = candlesticks[1:]
		}
	}
	return candlesticks
}

func calculateNormalizedStartingTimestamp(operand types.Operand, tm time.Time, startFromNext bool) int {
	if isMinutely(operand) {
		if tm.Second() == 0 {
			return int(tm.Unix()) + b2i(startFromNext)*candlestickDurationSecs(operand)
		}
		year, month, day := tm.Date()
		return int(time.Date(year, month, day, tm.Hour(), tm.Minute()+1+b2i(startFromNext), 0, 0, time.UTC).Unix())
	}

	if tm.Second() == 0 && tm.Minute() == 0 && tm.Hour() == 0 {
		return int(tm.Unix()) + b2i(startFromNext)*candlestickDurationSecs(operand)
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
