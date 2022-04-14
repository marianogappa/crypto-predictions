package messari

import (
	"fmt"
	"strings"
	"time"

	"github.com/marianogappa/predictions/market/common"
	"github.com/marianogappa/predictions/types"
)

type Messari struct {
	apiURL      string
	debug       bool
	apiKey      string
	timeNowFunc func() time.Time
}

func NewMessari() *Messari {
	return &Messari{apiURL: "https://data.messari.io/api/v1/", apiKey: "1ec22c58-744e-4453-93c6-ad73e2641054", timeNowFunc: time.Now}
}

func (m *Messari) SetDebug(debug bool) {
	m.debug = debug
}

func (m *Messari) RequestCandlesticks(operand types.Operand, startTimeTs int) ([]types.Candlestick, error) {
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

func (m *Messari) GetPatience() time.Duration { return 0 * time.Second }
