package binance

import (
	"time"

	"github.com/marianogappa/predictions/market/common"
	"github.com/marianogappa/predictions/types"
)

// Binance struct enables requesting candlesticks from Binance
type Binance struct {
	apiURL string
	debug  bool
}

// NewBinance is the constructor for Binance
func NewBinance() *Binance {
	return &Binance{apiURL: "https://api.binance.com/api/v3/"}
}

func (b *Binance) overrideAPIURL(url string) {
	b.apiURL = url
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
func (b *Binance) RequestCandlesticks(operand types.Operand, startTimeTs int, intervalMinutes int) ([]types.Candlestick, error) {
	res, err := b.getKlines(operand.BaseAsset, operand.QuoteAsset, startTimeTs*1000, intervalMinutes)
	if err != nil {
		return nil, err
	}

	return common.PatchCandlestickHoles(res.candlesticks, startTimeTs, 60*intervalMinutes), nil
}

// GetPatience returns the delay that this exchange usually takes in order for it to return candlesticks.
//
// Some exchanges may return results for unfinished candles (e.g. the current minute) and some may not, so callers
// should not request unfinished candles. This patience should be taken into account in addition to unfinished candles.
func (b *Binance) GetPatience() time.Duration { return 0 * time.Minute }

// SetDebug sets exchange-wide debug logging. It's useful to know how many times requests are being sent to exchanges.
func (b *Binance) SetDebug(debug bool) {
	b.debug = debug
}

const eRRINVALIDSYMBOL = -1121
