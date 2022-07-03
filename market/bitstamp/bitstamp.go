package bitstamp

import (
	"fmt"
	"sync"
	"time"

	"github.com/marianogappa/predictions/market/common"
	"github.com/marianogappa/predictions/types"
)

// Bitstamp struct enables requesting candlesticks from Bitstamp
type Bitstamp struct {
	apiURL    string
	debug     bool
	lock      sync.Mutex
	requester common.RequesterWithRetry
}

// NewBitstamp is the constructor for Bitstamp
func NewBitstamp() *Bitstamp {
	e := &Bitstamp{
		apiURL: "https://www.bitstamp.net/api/v2/",
	}

	e.requester = common.NewRequesterWithRetry(
		e.requestCandlesticks,
		common.RetryStrategy{Attempts: 3, FirstSleepTime: 1 * time.Second, SleepTimeMultiplier: 2.0},
		&e.debug,
	)

	return e
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
func (e *Bitstamp) RequestCandlesticks(operand types.Operand, startTimeTs int, intervalMinutes int) ([]types.Candlestick, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	candlesticks, err := e.requestCandlesticks(operand.BaseAsset, operand.QuoteAsset, startTimeTs, intervalMinutes)
	if err != nil {
		return nil, err
	}

	fmt.Println(candlesticks)
	return common.PatchCandlestickHoles(candlesticks, startTimeTs, 60*intervalMinutes), nil
}

// GetPatience returns the delay that this exchange usually takes in order for it to return candlesticks.
//
// Some exchanges may return results for unfinished candles (e.g. the current minute) and some may not, so callers
// should not request unfinished candles. This patience should be taken into account in addition to unfinished candles.
func (e *Bitstamp) GetPatience() time.Duration { return 0 * time.Second }

// SetDebug sets exchange-wide debug logging. It's useful to know how many times requests are being sent to exchanges.
func (e *Bitstamp) SetDebug(debug bool) {
	e.debug = debug
}