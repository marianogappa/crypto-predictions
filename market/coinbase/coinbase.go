package coinbase

import (
	"time"

	"github.com/marianogappa/predictions/market/common"
	"github.com/marianogappa/predictions/types"
)

// Coinbase struct enables requesting candlesticks from Coinbase
type Coinbase struct {
	apiURL string
	debug  bool
}

// NewCoinbase is the constructor for Coinbase
func NewCoinbase() *Coinbase {
	return &Coinbase{apiURL: "https://api.pro.coinbase.com/"}
}

func (c *Coinbase) overrideAPIURL(apiURL string) {
	c.apiURL = apiURL
}

// RequestCandlesticks requests candlesticks for the given market pair, of candlestick interval "intervalMinutes",
// starting at "startTimeTs".
//
// The supplied "intervalMinutes" may not be supported by this exchange.
//
// Candlesticks will start at the next multiple of "startTimeTs" as defined by
// time.Truncate(intervalMinutes * time.Minute)), except in some documented exceptions.
// This is enforced by the exchange.
//
// Some exchanges return candlesticks with gaps, but this method will patch the gaps by cloning the candlestick
// received right before the gap as many times as gaps, or the first candlestick if the gaps is at the start.
//
// Most of the usage of this method is with 1 minute intervals, the interval used to follow predictions.
func (c *Coinbase) RequestCandlesticks(operand types.Operand, startTimeTs int, intervalMinutes int) ([]types.Candlestick, error) {
	startTimeTm := time.Unix(int64(startTimeTs), 0)
	startTimeISO8601 := startTimeTm.Format(time.RFC3339)
	endTimeISO8601 := startTimeTm.Add(299 * 60 * time.Second).Format(time.RFC3339)

	res, err := c.getKlines(operand.BaseAsset, operand.QuoteAsset, startTimeISO8601, endTimeISO8601, intervalMinutes)
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

	return common.PatchCandlestickHoles(res.candlesticks, startTimeTs, 60*intervalMinutes), nil
}

// GetPatience returns the delay that this exchange usually takes in order for it to return candlesticks.
//
// Some exchanges may return results for unfinished candles (e.g. the current minute) and some may not, so callers
// should not request unfinished candles. This patience should be taken into account in addition to unfinished candles.
func (c *Coinbase) GetPatience() time.Duration { return 1 * time.Minute }

// SetDebug sets exchange-wide debug logging. It's useful to know how many times requests are being sent to exchanges.
func (c *Coinbase) SetDebug(debug bool) {
	c.debug = debug
}
