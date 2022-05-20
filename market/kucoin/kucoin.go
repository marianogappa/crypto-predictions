package kucoin

import (
	"time"

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

func (k *Kucoin) RequestCandlesticks(operand types.Operand, startTimeTs int, intervalMinutes int) ([]types.Candlestick, error) {
	res, err := k.getKlines(operand.BaseAsset, operand.QuoteAsset, startTimeTs, intervalMinutes)
	if err != nil {
		if res.kucoinErrorCode == "400100" && res.kucoinErrorMessage == "This pair is not provided at present" {
			return nil, types.ErrInvalidMarketPair
		}
		return nil, err
	}

	// Reverse slice, because Kucoin returns candlesticks in descending order
	for i, j := 0, len(res.candlesticks)-1; i < j; i, j = i+1, j-1 {
		res.candlesticks[i], res.candlesticks[j] = res.candlesticks[j], res.candlesticks[i]
	}

	return common.PatchCandlestickHoles(res.candlesticks, startTimeTs, 60), nil
}

func (k *Kucoin) GetPatience() time.Duration { return 0 * time.Second }

func (k *Kucoin) SetDebug(debug bool) {
	k.debug = debug
}
