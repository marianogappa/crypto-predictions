// The common package contains the input and output types of the signal checking function.
package common

import "github.com/marianogappa/predictions/types"

const (
	BINANCE              = "binance"
	FTX                  = "ftx"
	COINBASE             = "coinbase"
	HUOBI                = "huobi"
	KRAKEN               = "kraken"
	KUCOIN               = "kucoin"
	BINANCE_USDM_FUTURES = "binanceusdmfutures"

	// Used for testing
	FAKE = "fake"
)

type Exchange interface {
	BuildCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 types.ISO8601) *CandlestickIterator
	SetDebug(debug bool)
}
