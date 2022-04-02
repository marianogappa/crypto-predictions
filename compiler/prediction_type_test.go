package compiler

import (
	"testing"
	"time"

	"github.com/marianogappa/predictions/metadatafetcher"
	mfTypes "github.com/marianogappa/predictions/metadatafetcher/types"
	"github.com/marianogappa/predictions/types"
	"github.com/stretchr/testify/require"
)

func TestPredictionType(t *testing.T) {
	tss := []struct {
		name, pred string
		expected   types.PredictionType
	}{
		{
			name: "Basic PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE",
			pred: `{
				"reporter": "admin",
				"postUrl": "https://twitter.com/CryptoCapo_/status/1491357566974054400",
				"postedAt": "2022-02-09T10:25:26.000Z",
				"given": {
					"main": {
						"condition": "COIN:BINANCE:ADA-USDT <= 0.845",
						"toDuration": "2d"
					}
				},
				"predict": {
					"predict": "main"
				}
			}`,
			expected: types.PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE,
		},
		{
			name: "Basic PREDICTION_TYPE_COIN_WILL_RANGE",
			pred: `{
				"reporter": "admin",
				"postUrl": "https://twitter.com/CryptoCapo_/status/1491357566974054400",
				"postedAt": "2022-02-09T10:25:26.000Z",
				"given": {
					"main": {
						"condition": "COIN:BINANCE:ADA-USDT BETWEEN 0.845 AND 0.1",
						"toDuration": "2d"
					}
				},
				"predict": {
					"predict": "main"
				}
			}`,
			expected: types.PREDICTION_TYPE_COIN_WILL_RANGE,
		},
		{
			name: "Basic PREDICTION_TYPE_THE_FLIPPENING",
			pred: `{
				"reporter": "admin",
				"postUrl": "https://twitter.com/CryptoCapo_/status/1491357566974054400",
				"postedAt": "2022-02-09T10:25:26.000Z",
				"given": {
					"main": {
						"condition": "MARKETCAP:MESSARI:ETH > MARKETCAP:MESSARI:BTC",
						"toDuration": "2d"
					}
				},
				"predict": {
					"predict": "main"
				}
			}`,
			expected: types.PREDICTION_TYPE_THE_FLIPPENING,
		},
		{
			name: "Basic PREDICTION_TYPE_COIN_WILL_REACH_BEFORE_IT_REACHES",
			pred: `{
				"postUrl": "https://twitter.com/Trader_XO/status/1503690856125145092",
				"reporter": "admin",
				"given": {
				  "a": {
					"condition": "COIN:BINANCE:BTC-USDT >= 47000",
					"toDuration": "eoy",
					"errorMarginRatio": 0.03
				  },
				  "b": {
					"condition": "COIN:BINANCE:BTC-USDT <= 30000",
					"toDuration": "eoy",
					"errorMarginRatio": 0.03
				  }
				},
				"predict": {
				  "predict": "a and (not b)"
				}
			  }`,
			expected: types.PREDICTION_TYPE_COIN_WILL_REACH_BEFORE_IT_REACHES,
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			pc := NewPredictionCompiler(metadatafetcher.NewMetadataFetcher(), nil)
			pc.metadataFetcher.Fetchers = []metadatafetcher.SpecificFetcher{
				newTestMetadataFetcher(mfTypes.PostMetadata{
					Author:        types.Account{Handle: "CryptoCapo_"},
					PostCreatedAt: tpToISO("2020-01-02 00:00:00"),
				}, nil),
			}
			pc.timeNow = func() time.Time { return time.Now() }
			pred, _, err := pc.Compile([]byte(ts.pred))
			require.Nil(t, err)

			actual := CalculatePredictionType(pred)
			require.Equal(t, ts.expected, actual)
		})
	}
}
