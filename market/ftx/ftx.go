package ftx

import (
	"github.com/marianogappa/predictions/market/common"
	"github.com/marianogappa/predictions/types"
)

type FTX struct {
	apiURL string
	debug  bool
}

func NewFTX() *FTX {
	return &FTX{apiURL: "https://ftx.com/api/"}
}

func (f *FTX) overrideAPIURL(apiURL string) {
	f.apiURL = apiURL
}

func (b *FTX) SetDebug(debug bool) {
	b.debug = debug
}

func (f FTX) BuildCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 types.ISO8601) *common.CandlestickIterator {
	return common.NewCandlestickIterator(f.newCandlestickIterator(baseAsset, quoteAsset, initialISO8601).next)
}
