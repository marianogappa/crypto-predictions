package kraken

import (
	"time"

	"github.com/marianogappa/predictions/market/common"
	"github.com/marianogappa/predictions/types"
)

// NOTE: this exchange has to be removed because it does not provide data properly
// https://stackoverflow.com/questions/48508150/kraken-api-ohlc-request-doesnt-honor-the-since-parameter
type Kraken struct {
	apiURL string
	debug  bool
}

func NewKraken() *Kraken {
	return &Kraken{apiURL: "https://api.kraken.com/0/"}
}

func (k *Kraken) overrideAPIURL(apiURL string) {
	k.apiURL = apiURL
}

func (k *Kraken) RequestTicks(operand types.Operand, startTimeTs int) ([]types.Tick, error) {
	res, err := k.getKlines(operand.BaseAsset, operand.QuoteAsset, startTimeTs)
	if err != nil {
		return nil, err
	}
	return common.PatchTickHoles(common.CandlesticksToTicks(res.candlesticks), startTimeTs, 60), nil
}

func (k *Kraken) GetPatience() time.Duration { return 0 * time.Second }

func (k *Kraken) SetDebug(debug bool) {
	k.debug = debug
}
