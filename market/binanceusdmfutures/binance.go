package binanceusdmfutures

import (
	"time"

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

func (b *BinanceUSDMFutures) RequestCandlesticks(operand types.Operand, startTimeTs int, intervalMinutes int) ([]types.Candlestick, error) {
	res, err := b.getKlines(operand.BaseAsset, operand.QuoteAsset, startTimeTs*1000, intervalMinutes)
	if err != nil {
		return nil, err
	}
	return common.PatchCandlestickHoles(res.candlesticks, startTimeTs, 60*intervalMinutes), nil
}

func (b *BinanceUSDMFutures) GetPatience() time.Duration { return 0 * time.Minute }

func (b *BinanceUSDMFutures) SetDebug(debug bool) {
	b.debug = debug
}

const ERR_INVALID_SYMBOL = -1121
