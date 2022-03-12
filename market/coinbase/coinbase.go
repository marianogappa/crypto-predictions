package coinbase

import (
	"github.com/marianogappa/predictions/market/common"
	"github.com/marianogappa/predictions/types"
)

type Coinbase struct {
	apiURL string
	debug  bool
}

func NewCoinbase() *Coinbase {
	return &Coinbase{apiURL: "https://api.pro.coinbase.com/"}
}

func (c *Coinbase) overrideAPIURL(apiURL string) {
	c.apiURL = apiURL
}

func (b *Coinbase) SetDebug(debug bool) {
	b.debug = debug
}

func (c Coinbase) BuildCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 types.ISO8601) *common.CandlestickIterator {
	return common.NewCandlestickIterator(c.newCandlestickIterator(baseAsset, quoteAsset, initialISO8601).next)
}
