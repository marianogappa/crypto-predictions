package messari

import (
	"fmt"
	"strings"
	"time"

	"github.com/marianogappa/predictions/market/common"
	"github.com/marianogappa/predictions/types"
)

// Messari struct enables requesting datapoints from Messari
type Messari struct {
	apiURL      string
	debug       bool
	apiKey      string
	timeNowFunc func() time.Time
}

// NewMessari is the constructor for Messari
func NewMessari() *Messari {
	return &Messari{apiURL: "https://data.messari.io/api/v1/", apiKey: "1ec22c58-744e-4453-93c6-ad73e2641054", timeNowFunc: time.Now}
}

// RequestCandlesticks requests candlesticks for the given market pair, of candlestick interval "intervalMinutes",
// starting at "startTimeTs".
//
// Since this provider (Messari) is used for getting ticks rather than candlesticks and they are always daily,
// the intervalMinutes parameter is ignored.
//
// Candlesticks will start at the next multiple of "startTimeTs" as defined by
// time.Truncate(intervalMinutes * time.Minute)), except in some documented exceptions.
// This is enforced by the exchange.
//
// Some exchanges return candlesticks with gaps, but this method will patch the gaps by cloning the candlestick
// received right before the gap as many times as gaps, or the first candlestick if the gaps is at the start.
func (m *Messari) RequestCandlesticks(operand types.Operand, startTimeTs int, _ignoredIntervalMinutes int) ([]types.Candlestick, error) {
	res, err := m.getMetrics(operand.BaseAsset, "mcap.out", startTimeTs*1000)
	if err != nil {
		if res.messariErrorCode == 404 && strings.HasPrefix(res.messariErrorMessage, "Asset with key = ") && strings.HasSuffix(res.messariErrorMessage, " not found.") {
			return nil, fmt.Errorf("%w: %v", types.ErrInvalidMarketPair, res.messariErrorMessage)
		}
		return nil, err
	}

	patchedTicks := common.PatchCandlestickHoles(res.candlesticks, startTimeTs, 60*24*24)

	// Messari sometimes returns no error but no data for some symbols (e.g. happened with FTM)
	if len(patchedTicks) == 0 {
		y, mo, d := time.Unix(int64(startTimeTs), 0).Date()
		nextDay := time.Date(y, mo, d+1, 0, 0, 0, 0, time.UTC)
		if nextDay.Before(m.timeNowFunc()) {
			return nil, fmt.Errorf("%w: Messari did not fail but returned no data even though it was supposed to", types.ErrInvalidMarketPair)
		}
	}

	return patchedTicks, nil
}

// GetPatience returns the delay that this exchange usually takes in order for it to return candlesticks.
//
// Some exchanges may return results for unfinished candles (e.g. the current minute) and some may not, so callers
// should not request unfinished candles. This patience should be taken into account in addition to unfinished candles.
func (m *Messari) GetPatience() time.Duration { return 0 * time.Second }

// SetDebug sets exchange-wide debug logging. It's useful to know how many times requests are being sent to exchanges.
func (m *Messari) SetDebug(debug bool) {
	m.debug = debug
}
