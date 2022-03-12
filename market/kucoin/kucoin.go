package kucoin

import (
	"github.com/marianogappa/predictions/market/common"
	"github.com/marianogappa/predictions/types"
)

type Kucoin struct {
	apiURL string
	debug  bool
}

func NewKucoin() *Kucoin {
	return &Kucoin{apiURL: "https://api.kucoin.com/api/v1/"}
}

func (k *Kucoin) overrideAPIURL(apiURL string) {
	k.apiURL = apiURL
}

func (b *Kucoin) SetDebug(debug bool) {
	b.debug = debug
}

func (k Kucoin) BuildCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 types.ISO8601) *common.CandlestickIterator {
	return common.NewCandlestickIterator(k.newCandlestickIterator(baseAsset, quoteAsset, initialISO8601).next)
}
