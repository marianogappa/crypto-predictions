package coinbase

import (
	"time"

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

func (c *Coinbase) RequestTicks(operand types.Operand, startTimeTs int) ([]types.Tick, error) {
	startTimeTm := time.Unix(int64(startTimeTs), 0)
	startTimeISO8601 := startTimeTm.Format(time.RFC3339)
	endTimeISO8601 := startTimeTm.Add(299 * 60 * time.Second).Format(time.RFC3339)

	res, err := c.getKlines(operand.BaseAsset, operand.QuoteAsset, startTimeISO8601, endTimeISO8601)
	if err != nil {
		if res.coinbaseErrorMessage == "NotFound" {
			return nil, types.ErrInvalidMarketPair
		}
		return nil, err
	}

	// Reverse slice, because Coinbase returns candlesticks in descending order
	for i, j := 0, len(res.candlesticks)-1; i < j; i, j = i+1, j-1 {
		res.candlesticks[i], res.candlesticks[j] = res.candlesticks[j], res.candlesticks[i]
	}

	return common.PatchTickHoles(common.CandlesticksToTicks(res.candlesticks), startTimeTs, 60), nil
}

func (c *Coinbase) GetPatience() time.Duration { return 1 * time.Minute }

func (c *Coinbase) SetDebug(debug bool) {
	c.debug = debug
}
