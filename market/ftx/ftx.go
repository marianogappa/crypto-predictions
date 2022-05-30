package ftx

import (
	"time"

	"github.com/marianogappa/predictions/market/common"
	"github.com/marianogappa/predictions/types"
)

// FTX struct enables requesting candlesticks from FTX
type FTX struct {
	apiURL string
	debug  bool
}

// NewFTX is the constructor for FTX
func NewFTX() *FTX {
	return &FTX{apiURL: "https://ftx.com/api/"}
}

// func (f *FTX) overrideAPIURL(apiURL string) {
// 	f.apiURL = apiURL
// }

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
func (f *FTX) RequestCandlesticks(operand types.Operand, startTimeTs int, intervalMinutes int) ([]types.Candlestick, error) {
	res, err := f.getKlines(operand.BaseAsset, operand.QuoteAsset, startTimeTs, intervalMinutes)
	if err != nil {
		return nil, err
	}
	return common.PatchCandlestickHoles(res.candlesticks, startTimeTs, 60*intervalMinutes), nil
}

// GetPatience returns the delay that this exchange usually takes in order for it to return candlesticks.
//
// Some exchanges may return results for unfinished candles (e.g. the current minute) and some may not, so callers
// should not request unfinished candles. This patience should be taken into account in addition to unfinished candles.
func (f *FTX) GetPatience() time.Duration { return 0 * time.Second }

// SetDebug sets exchange-wide debug logging. It's useful to know how many times requests are being sent to exchanges.
func (f *FTX) SetDebug(debug bool) {
	f.debug = debug
}
