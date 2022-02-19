package market

import (
	"fmt"
	"strings"

	"github.com/marianogappa/predictions/types"
	"github.com/marianogappa/signal-checker/binance"
	"github.com/marianogappa/signal-checker/binanceusdmfutures"
	"github.com/marianogappa/signal-checker/coinbase"
	"github.com/marianogappa/signal-checker/common"
	"github.com/marianogappa/signal-checker/ftx"
	"github.com/marianogappa/signal-checker/kraken"
	"github.com/marianogappa/signal-checker/kucoin"
)

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
)

func NewMarket() Market {
	for exchangeName := range exchanges {
		supportedVariableProviders[strings.ToUpper(exchangeName)] = struct{}{}
	}

	return Market{}
}

func (m Market) GetTickIterator(operand types.Operand, initialISO8601 common.ISO8601) (*TickIterator, error) {
	if operand.Type == types.MARKETCAP {
		return nil, fmt.Errorf("the 'MARKETCAP' operand type is not supported yet")
	}
	if _, ok := supportedVariableProviders[operand.Provider]; !ok {
		return nil, fmt.Errorf("the '%v' provider is not supported for %v:%v-%v", operand.Provider, operand.Provider, operand.BaseAsset, operand.QuoteAsset)
	}
	exchange := exchanges[strings.ToLower(operand.Provider)]
	return newTickIterator(exchange.BuildCandlestickIterator(operand.BaseAsset, operand.QuoteAsset, initialISO8601).Next), nil
}
