package ftx

import (
	"time"

	"github.com/marianogappa/predictions/market/common"
	"github.com/marianogappa/predictions/types"
)

type FTX struct {
	apiURL string
	debug  bool
}

func NewFTX() *FTX {
	return &FTX{apiURL: "https://ftx.com/api/"}
}

func (f *FTX) overrideAPIURL(apiURL string) {
	f.apiURL = apiURL
}

func (f *FTX) RequestCandlesticks(operand types.Operand, startTimeTs int) ([]types.Candlestick, error) {
	res, err := f.getKlines(operand.BaseAsset, operand.QuoteAsset, startTimeTs)
	if err != nil {
		return nil, err
	}
	return common.PatchCandlestickHoles(res.candlesticks, startTimeTs, 60), nil
}

func (f *FTX) GetPatience() time.Duration { return 0 * time.Second }

func (f *FTX) SetDebug(debug bool) {
	f.debug = debug
}
