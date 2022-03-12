package binanceusdmfutures

import (
	"github.com/marianogappa/predictions/market/common"
	"github.com/marianogappa/predictions/types"
)

type BinanceUSDMFutures struct {
	apiURL string
	debug  bool
}

func NewBinanceUSDMFutures() *BinanceUSDMFutures {
	return &BinanceUSDMFutures{apiURL: "https://fapi.binance.com/fapi/v1/"}
}

func (b *BinanceUSDMFutures) overrideAPIURL(url string) {
	b.apiURL = url
}

func (b *BinanceUSDMFutures) SetDebug(debug bool) {
	b.debug = debug
}

func (b BinanceUSDMFutures) BuildCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 types.ISO8601) *common.CandlestickIterator {
	return common.NewCandlestickIterator(b.newCandlestickIterator(baseAsset, quoteAsset, initialISO8601).next)
}

const ERR_INVALID_SYMBOL = -1121
