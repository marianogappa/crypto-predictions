package binance

import (
	"github.com/marianogappa/predictions/market/common"
	"github.com/marianogappa/predictions/types"
)

type Binance struct {
	apiURL string
	debug  bool
}

func NewBinance() *Binance {
	return &Binance{apiURL: "https://api.binance.com/api/v3/"}
}

func (b *Binance) overrideAPIURL(url string) {
	b.apiURL = url
}

func (b *Binance) SetDebug(debug bool) {
	b.debug = debug
}

func (b Binance) BuildCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 types.ISO8601) *common.CandlestickIterator {
	return common.NewCandlestickIterator(b.newCandlestickIterator(baseAsset, quoteAsset, initialISO8601).next)
}

const ERR_INVALID_SYMBOL = -1121
