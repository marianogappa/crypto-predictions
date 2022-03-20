// The cache package implements an in-memory LRU cache layer between crypto exchanges and the TickIterators.
//
// It solves this problem: if there are 1000 predictions about BTC/USDT that need the current value of the market
// pair right now, (1) it would take 1000*(network request against exchange) to get the same value 1000 times, and
// (2) the exchange would rate-limit the IP making the request.
//
// The package exposes a MemoryCache struct instantiated via NewMemoryCache.
//
// Usage:
//
// ```
//
// cache := cache.NewMemoryCache(10000, 1000)
//
// opBTCUSDT := types.Operand{
// 	Type:       types.COIN,
// 	Provider:   "BINANCE",
// 	BaseAsset:  "BTC",
// 	QuoteAsset: "USDT",
// }
//
// startISO8601 := types.ISO8601("2022-03-20T12:22:00Z")
// startTs, err := startISO8601.Seconds()
//
// err := cache.Put(operand, []types.Tick{{Timestamp: startTs, Value: 1234}, {Timestamp: startTs+60, Value: 2345}, ...})
//
// ticks, err := cache.Get(operand, startISO8601)
//
// ```
//
// Internally, it is composed of two caches: one for marketPairs (e.g. BTC/USDT) whose granularity is one tick per
// minute, and assets (e.g. BTC) whose granularity is one tick per day.
//
// Each minutely-cache entry is an [1440]float64 keyed by "%provider%-%baseAsset%-%quoteAsset%-%year%-%month%-%day%",
// holding the values for an entire day.
//
// Each daily-cache entry is an [366]float64 keyed by "%provider%-%baseAsset%--%year%", holding the values for an
// entire year (it contains up to 366 values on leap years, or 365 otherwise).
package cache

import (
	"errors"
	"fmt"

	lru "github.com/hashicorp/golang-lru"
	"github.com/marianogappa/predictions/types"
)

// MemoryCache implements the in-memory LRU cache layer that this package exposes.
type MemoryCache struct {
	minutelyValueCache *lru.Cache
	dailyValueCache    *lru.Cache

	CacheMisses   int
	CacheRequests int
}

var (
	ErrTimestampMustHaveZeroInSecondsPart = errors.New("timestamp must have zero in seconds part")
	ErrTimestampMustHaveZeroInTimePart    = errors.New("timestamp must have zero in time part")
	ErrReceivedTickWithZeroValue          = errors.New("received tick with zero value")
	ErrReceivedNonSubsequentTick          = errors.New("received non-subsequent tick")
	ErrInvalidISO8601                     = errors.New("invalid ISO8601")
	ErrCacheMiss                          = errors.New("cache miss")
)

// NewMemoryCache instantiates the in-memory LRU cache layer that this package exposes.
//
// Internally, it is composed of two caches: one for marketPairs (e.g. BTC/USDT) whose granularity is one tick per
// minute, and assets (e.g. BTC) whose granularity is one tick per day.
//
// The minutelySize and dailySize parameters configure the maximum amount of _days_ and _years_ of a given
// (provider, marketPair/asset) tuple these caches can store before evicting.
func NewMemoryCache(minutelySize int, dailySize int) *MemoryCache {
	if minutelySize <= 0 {
		minutelySize = 1
	}
	if dailySize <= 0 {
		dailySize = 1
	}
	minutelyValueCache, _ := lru.New(minutelySize)
	dailyValueCache, _ := lru.New(dailySize)
	return &MemoryCache{minutelyValueCache: minutelyValueCache, dailyValueCache: dailyValueCache}
}

// Put pushes a slice of ticks from the given (provider, marketPair/asset) into the cache. May evict older entries.
//
// * Fails with ErrReceivedTickWithZeroValue if a tick with a Value of 0 is supplied.
//
// * Fails with ErrReceivedNonSubsequentTick if supplied ticks are not sorted ascendingly.
//
// * Fails with ErrReceivedNonSubsequentTick if supplied ticks are not exactly 60 seconds apart for marketPairs, or
//   a day apart for assets.
//
// * Fails with ErrTimestampMustHaveZeroInSecondsPart if ticks' timestamps for marketPairs don't have 0 for the seconds
//   component.
//
// * Fails with ErrTimestampMustHaveZeroInTimePart if ticks' timestamps for assets don't have 0 for the complete date
//   component.
func (c *MemoryCache) Put(operand types.Operand, ticks []types.Tick) error {
	if len(ticks) == 0 {
		return nil
	}

	if isMinutely(operand) {
		return c.putMinutely(operand, ticks)
	}
	return c.putDaily(operand, ticks)
}

// Get retrieves ticks for the given (provider, marketPair/asset) starting at the supplied date time. If the supplied
// date time has non-zero seconds/date components (respectively for marketPair/asset), it will be normalized to the
// immediately next zero date time.
//
// It will retrieve all subsequent ticks starting _exactly_ at the normalized date time up to the end of the day/year
// respectively. If there's no entry for exactly that date time, it will fail with ErrCacheMiss. It will stop at the
// first gap, rather than return gaps.
//
// Even if there is an entry for the last minute of the day (or day of the year) and an entry for the first minute of
// the next day (or day of next year), a Get's response won't span multiple days (or years).
//
// * Fails with ErrInvalidISO8601 if the supplied date time is invalid (note that the type wraps string, so it does
//   not prevent invalid strings to be supplied).
//
// * Fails with ErrCacheMiss if there are no values available in the cache. Client must handle this error, as it's
//   completely normal to have cache misses.
//
// * Note it does not fail with non-sensical values for operand components, as this is completely foreign to the
//   cache's logic.
func (c *MemoryCache) Get(operand types.Operand, initialISO8601 types.ISO8601) ([]types.Tick, error) {
	tm, err := initialISO8601.Time()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidISO8601, initialISO8601)
	}
	c.CacheRequests++
	startingTimestamp := calculateNormalizedStartingTimestamp(operand, tm)

	if isMinutely(operand) {
		return c.getMinutely(operand, startingTimestamp)
	}
	return c.getDaily(operand, startingTimestamp)
}
