package kucoin

import (
	"time"

	"github.com/marianogappa/predictions/market/common"
	"github.com/marianogappa/predictions/types"
)

// Kucoin struct enables requesting candlesticks from Kucoin
type Kucoin struct {
	apiURL string
	debug  bool
}

// NewKucoin is the constructor for Kucoin
func NewKucoin() *Kucoin {
	return &Kucoin{apiURL: "https://api.kucoin.com/api/v1/"}
}

func (k *Kucoin) overrideAPIURL(apiURL string) {
	k.apiURL = apiURL
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

	return common.PatchCandlestickHoles(res.candlesticks, startTimeTs, 60*intervalMinutes), nil
}

// GetPatience returns the delay that this exchange usually takes in order for it to return candlesticks.
//
// Some exchanges may return results for unfinished candles (e.g. the current minute) and some may not, so callers
// should not request unfinished candles. This patience should be taken into account in addition to unfinished candles.
func (k *Kucoin) GetPatience() time.Duration { return 0 * time.Second }

// SetDebug sets exchange-wide debug logging. It's useful to know how many times requests are being sent to exchanges.
func (k *Kucoin) SetDebug(debug bool) {
	k.debug = debug
}
