package market

import (
	"errors"
	"fmt"
	"strings"

	"github.com/marianogappa/predictions/market/messari"
	"github.com/marianogappa/predictions/types"
	"github.com/marianogappa/signal-checker/binance"
	"github.com/marianogappa/signal-checker/binanceusdmfutures"
	"github.com/marianogappa/signal-checker/coinbase"
	"github.com/marianogappa/signal-checker/common"
	"github.com/marianogappa/signal-checker/ftx"
	"github.com/marianogappa/signal-checker/kraken"
	"github.com/marianogappa/signal-checker/kucoin"
)

type IMarket interface {
	GetTickIterator(operand types.Operand, initialISO8601 common.ISO8601) (types.TickIterator, error)
}

type Market struct{}

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

	return Market{}
}

func (m Market) GetTickIterator(operand types.Operand, initialISO8601 common.ISO8601) (types.TickIterator, error) {
	switch operand.Type {
	case types.COIN:
		return m.getCoinTickIterator(operand, initialISO8601)
	case types.MARKETCAP:
		if operand.Provider != "MESSARI" {
			return nil, fmt.Errorf("only supported provider for MARKETCAP is 'MESSARI', got %v", operand.Provider)
		}
		return m.getMarketcapTickIterator(operand, initialISO8601)
	default:
		return nil, fmt.Errorf("invalid operand type %v", operand.Type)
	}
}

func (m Market) getCoinTickIterator(operand types.Operand, initialISO8601 common.ISO8601) (types.TickIterator, error) {
	if _, ok := supportedVariableProviders[operand.Provider]; !ok {
		return nil, fmt.Errorf("the '%v' provider is not supported for %v:%v-%v", operand.Provider, operand.Provider, operand.BaseAsset, operand.QuoteAsset)
	}
	exchange := exchanges[strings.ToLower(operand.Provider)]
	return newTickFromCandleIterator(exchange.BuildCandlestickIterator(operand.BaseAsset, operand.QuoteAsset, initialISO8601).Next), nil
}

func (m Market) getMarketcapTickIterator(operand types.Operand, initialISO8601 common.ISO8601) (types.TickIterator, error) {
	if operand.BaseAsset == "" {
		return nil, errEmptyBaseAsset
	}
	return messari.NewMessari().BuildTickIterator(operand.BaseAsset, "mcap.out", initialISO8601), nil
}
