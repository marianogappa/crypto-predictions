package market

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/marianogappa/predictions/market/binance"
	"github.com/marianogappa/predictions/market/binanceusdmfutures"
	"github.com/marianogappa/predictions/market/cache"
	"github.com/marianogappa/predictions/market/coinbase"
	"github.com/marianogappa/predictions/market/common"
	"github.com/marianogappa/predictions/market/ftx"
	"github.com/marianogappa/predictions/market/iterator"
	"github.com/marianogappa/predictions/market/kucoin"
	"github.com/marianogappa/predictions/market/messari"
	"github.com/marianogappa/predictions/types"
)

// IMarket only exists so that tests can use a test iterator.
type IMarket interface {
	GetIterator(operand types.Operand, initialISO8601 types.ISO8601, startFromNext bool, intervalMinutes int) (types.Iterator, error)
}

// Market struct implements the crypto market.
type Market struct {
	cache                      *cache.MemoryCache
	timeNowFunc                func() time.Time
	debug                      bool
	exchanges                  map[string]common.Exchange
	supportedVariableProviders map[string]struct{}
}

var (
	errEmptyBaseAsset = errors.New("base asset must be supplied in order to create Tick Iterator")
)

// NewMarket constructs a market.
func NewMarket(cacheSizes map[time.Duration]int) Market {
	exchanges := map[string]common.Exchange{
		common.BINANCE:            binance.NewBinance(),
		common.FTX:                ftx.NewFTX(),
		common.COINBASE:           coinbase.NewCoinbase(),
		common.KUCOIN:             kucoin.NewKucoin(),
		common.BINANCEUSDMFUTURES: binanceusdmfutures.NewBinanceUSDMFutures(),
	}
	supportedVariableProviders := map[string]struct{}{}
	for exchangeName := range exchanges {
		supportedVariableProviders[strings.ToUpper(exchangeName)] = struct{}{}
	}
	cache := cache.NewMemoryCache(
		map[time.Duration]int{
			time.Minute:    10000,
			1 * time.Hour:  1000,
			24 * time.Hour: 1000,
		},
	)

	return Market{cache: cache, timeNowFunc: time.Now, exchanges: exchanges, supportedVariableProviders: supportedVariableProviders}
}

// SetDebug sets debug logging across all exchanges and the Market struct itself. Useful to know how many times an
// exchange is being requested.
func (m *Market) SetDebug(debug bool) {
	m.debug = debug
	for _, exchange := range m.exchanges {
		exchange.SetDebug(debug)
	}
}

// GetIterator returns a market iterator for a given operand at a given time and for a given candlestick interval.
func (m Market) GetIterator(operand types.Operand, initialISO8601 types.ISO8601, startFromNext bool, intervalMinutes int) (types.Iterator, error) {
	switch operand.Type {
	case types.COIN:
		return m.getCoinIterator(operand, initialISO8601, startFromNext, intervalMinutes)
	case types.MARKETCAP:
		if operand.Provider != "MESSARI" {
			return nil, fmt.Errorf("only supported provider for MARKETCAP is 'MESSARI', got %v", operand.Provider)
		}
		return m.getMarketcapIterator(operand, initialISO8601, startFromNext)
	default:
		return nil, fmt.Errorf("invalid operand type %v", operand.Type)
	}
}

func (m Market) getCoinIterator(operand types.Operand, initialISO8601 types.ISO8601, startFromNext bool, intervalMinutes int) (types.Iterator, error) {
	if _, ok := m.supportedVariableProviders[operand.Provider]; !ok {
		return nil, fmt.Errorf("the '%v' provider is not supported for %v:%v-%v", operand.Provider, operand.Provider, operand.BaseAsset, operand.QuoteAsset)
	}
	exchange := m.exchanges[strings.ToLower(operand.Provider)]
	return iterator.NewIterator(operand, initialISO8601, m.cache, exchange, m.timeNowFunc, startFromNext, intervalMinutes)
}

func (m Market) getMarketcapIterator(operand types.Operand, initialISO8601 types.ISO8601, startFromNext bool) (types.Iterator, error) {
	if operand.BaseAsset == "" {
		return nil, errEmptyBaseAsset
	}
	mess := messari.NewMessari()
	return iterator.NewIterator(operand, initialISO8601, m.cache, mess, m.timeNowFunc, startFromNext, 60*24)
}

// CalculateCacheHitRatio returns the hit ratio of the cache of the market. Used to see if the cache is useful.
func (m Market) CalculateCacheHitRatio() float64 {
	if m.cache.CacheRequests == 0 {
		return 0
	}
	return float64(m.cache.CacheMisses) / float64(m.cache.CacheRequests) * 100
}
