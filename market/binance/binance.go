package binance

import (
	"time"

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

func (b *Binance) RequestCandlesticks(operand types.Operand, startTimeTs int, intervalMinutes int) ([]types.Candlestick, error) {
	res, err := b.getKlines(operand.BaseAsset, operand.QuoteAsset, startTimeTs*1000, intervalMinutes)
	if err != nil {
		return nil, err
	}

	return common.PatchCandlestickHoles(res.candlesticks, startTimeTs, 60*intervalMinutes), nil
}

func (b *Binance) GetPatience() time.Duration { return 0 * time.Minute }

func (b *Binance) SetDebug(debug bool) {
	b.debug = debug
}

const ERR_INVALID_SYMBOL = -1121
