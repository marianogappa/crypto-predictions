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
	"github.com/marianogappa/predictions/market/kraken"
	"github.com/marianogappa/predictions/market/kucoin"
	"github.com/marianogappa/predictions/market/messari"
	"github.com/marianogappa/predictions/market/tickiterator"
	"github.com/marianogappa/predictions/types"
)

type IMarket interface {
	GetTickIterator(operand types.Operand, initialISO8601 types.ISO8601, startFromNext bool) (types.TickIterator, error)
}

type Market struct {
	cache       *cache.MemoryCache
	timeNowFunc func() time.Time
}

var (
	exchanges = map[string]common.Exchange{
		common.BINANCE:              binance.NewBinance(),
		common.FTX:                  ftx.NewFTX(),
		common.COINBASE:             coinbase.NewCoinbase(),
		common.KRAKEN:               kraken.NewKraken(),
		common.KUCOIN:               kucoin.NewKucoin(),
		common.BINANCE_USDM_FUTURES: binanceusdmfutures.NewBinanceUSDMFutures(),
	}
	supportedVariableProviders = map[string]struct{}{}
	errEmptyBaseAsset          = errors.New("base asset must be supplied in order to create Tick Iterator")
)

func NewMarket() Market {
	for exchangeName := range exchanges {
		supportedVariableProviders[strings.ToUpper(exchangeName)] = struct{}{}
	}

	return Market{cache: cache.NewMemoryCache(10000, 1000), timeNowFunc: time.Now}
}

func (m Market) GetTickIterator(operand types.Operand, initialISO8601 types.ISO8601, startFromNext bool) (types.TickIterator, error) {
	switch operand.Type {
	case types.COIN:
		return m.getCoinTickIterator(operand, initialISO8601, startFromNext)
	case types.MARKETCAP:
		if operand.Provider != "MESSARI" {
			return nil, fmt.Errorf("only supported provider for MARKETCAP is 'MESSARI', got %v", operand.Provider)
		}
		return m.getMarketcapTickIterator(operand, initialISO8601, startFromNext)
	default:
		return nil, fmt.Errorf("invalid operand type %v", operand.Type)
	}
}

func (m Market) getCoinTickIterator(operand types.Operand, initialISO8601 types.ISO8601, startFromNext bool) (types.TickIterator, error) {
	if _, ok := supportedVariableProviders[operand.Provider]; !ok {
		return nil, fmt.Errorf("the '%v' provider is not supported for %v:%v-%v", operand.Provider, operand.Provider, operand.BaseAsset, operand.QuoteAsset)
	}
	exchange := exchanges[strings.ToLower(operand.Provider)]
	return tickiterator.NewTickIterator(operand, initialISO8601, m.cache, exchange, m.timeNowFunc, startFromNext)
}

func (m Market) getMarketcapTickIterator(operand types.Operand, initialISO8601 types.ISO8601, startFromNext bool) (types.TickIterator, error) {
	if operand.BaseAsset == "" {
		return nil, errEmptyBaseAsset
	}
	mess := messari.NewMessari()
	return tickiterator.NewTickIterator(operand, initialISO8601, m.cache, mess, m.timeNowFunc, startFromNext)
}

func (m Market) CalculateCacheHitRatio() float64 {
	if m.cache.CacheRequests == 0 {
		return 0
	}
	return float64(m.cache.CacheMisses) / float64(m.cache.CacheRequests) * 100
}
