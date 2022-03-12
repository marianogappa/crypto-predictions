package kraken

import (
	"github.com/marianogappa/predictions/market/common"
	"github.com/marianogappa/predictions/types"
)

type Kraken struct {
	apiURL string
	debug  bool
}

func NewKraken() *Kraken {
	return &Kraken{apiURL: "https://api.kraken.com/0/"}
}

func (k *Kraken) overrideAPIURL(apiURL string) {
	k.apiURL = apiURL
}

func (b *Kraken) SetDebug(debug bool) {
	b.debug = debug
}

func (k Kraken) BuildCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 types.ISO8601) *common.CandlestickIterator {
	return common.NewCandlestickIterator(k.newCandlestickIterator(baseAsset, quoteAsset, initialISO8601).next)
}
