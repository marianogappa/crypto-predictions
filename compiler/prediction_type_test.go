package compiler

import (
	"testing"
	"time"

	"github.com/marianogappa/predictions/core"
	"github.com/marianogappa/predictions/metadatafetcher"
	mfTypes "github.com/marianogappa/predictions/metadatafetcher/types"
	"github.com/stretchr/testify/require"
)

func TestPredictionType(t *testing.T) {
	tss := []struct {
		name, pred string
		expected   core.PredictionType
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
			expected: core.PredictionTypeCoinOperatorFloatDeadline,
		},
		{
			name: "Basic PREDICTION_TYPE_COIN_WILL_REACH_INVALIDATED_IF_IT_REACHES",
			pred: `{
				"postUrl": "https://www.youtube.com/watch?v=JWVrWmuSHic&t=88",
				"reporter": "admin",
				"given": {
				  "main": {
					"condition": "COIN:BINANCE:BTC-USDT >= 100000",
					"toDuration": "eoy",
					"errorMarginRatio": 0.03
				  },
				  "a": {
					"condition": "COIN:BINANCE:BTC-USDT <= 30000",
					"toDuration": "eoy"
				  }
				},
				"predict": {
				  "predict": "main",
				  "annulledIf": "a",
				  "ignoreUndecidedIfPredictIsDefined": true
				}
			  }`,
			expected: core.PredictionTypeCoinWillReachInvalidatedIfItReaches,
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
			expected: core.PredictionTypeCoinWillRange,
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
			expected: core.PredictionTypeCoinWillReachBeforeItReaches,
		},
		{
			name: "Basic PREDICTION_TYPE_UNSUPPORTED",
			pred: `{
				"uuid": "dbe0e928-1aeb-4d68-bebf-a0b5d6703531",
				"version": "1.0.0",
				"createdAt": "2022-03-20T14:12:01Z",
				"reporter": "admin",
				"postAuthor": "CryptoCapo_",
				"postedAt": "2022-03-19T19:59:19.000Z",
				"postUrl": "https://twitter.com/CryptoCapo_/status/1505272725832708098",
				"given":
				{
					"a":
					{
						"condition": "COIN:BINANCE:BTC-USDT <= 23000",
						"fromISO8601": "2022-03-19T19:59:19Z",
						"toISO8601": "2022-06-17T20:59:19+01:00",
						"toDuration": "3m",
						"assumed": null,
						"state":
						{
							"status": "STARTED",
							"lastTs": 1648903740,
							"lastTicks":
							{
								"COIN:BINANCE:BTC-USDT":
								{
									"t": 1648903740,
									"v": 46800.010000000002
								}
							},
							"value": "UNDECIDED"
						},
						"errorMarginRatio": 0.03
					},
					"b":
					{
						"condition": "COIN:BINANCE:ETH-USDT <= 1300",
						"fromISO8601": "2022-03-19T19:59:19Z",
						"toISO8601": "2022-06-17T20:59:19+01:00",
						"toDuration": "3m",
						"assumed": null,
						"state":
						{
							"status": "STARTED",
							"lastTs": 1648903800,
							"lastTicks":
							{
								"COIN:BINANCE:ETH-USDT":
								{
									"t": 1648903800,
									"v": 3519.19
								}
							},
							"value": "UNDECIDED"
						},
						"errorMarginRatio": 0.03
					},
					"c":
					{
						"condition": "COIN:BINANCE:ADA-USDT <= 0.45",
						"fromISO8601": "2022-03-19T19:59:19Z",
						"toISO8601": "2022-06-17T20:59:19+01:00",
						"toDuration": "3m",
						"assumed": null,
						"state":
						{
							"status": "STARTED",
							"lastTs": 1648903800,
							"lastTicks":
							{
								"COIN:BINANCE:ADA-USDT":
								{
									"t": 1648903800,
									"v": 1.189
								}
							},
							"value": "UNDECIDED"
						},
						"errorMarginRatio": 0.03
					},
					"d":
					{
						"condition": "COIN:BINANCE:LUNA-USDT <= 45",
						"fromISO8601": "2022-03-19T19:59:19Z",
						"toISO8601": "2022-06-17T20:59:19+01:00",
						"toDuration": "3m",
						"assumed": null,
						"state":
						{
							"status": "STARTED",
							"lastTs": 1648903740,
							"lastTicks":
							{
								"COIN:BINANCE:LUNA-USDT":
								{
									"t": 1648903740,
									"v": 112.2
								}
							},
							"value": "UNDECIDED"
						},
						"errorMarginRatio": 0.03
					}
				},
				"prePredict":
				{},
				"predict":
				{
					"predict": "a and b and c and d"
				},
				"state":
				{
					"status": "STARTED",
					"lastTs": 0,
					"value": "ONGOING_PREDICTION"
				}
			}`,
			expected: core.PredictionTypeUnsupported,
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			pc := NewPredictionCompiler(metadatafetcher.NewMetadataFetcher(), nil)
			pc.metadataFetcher.Fetchers = []metadatafetcher.SpecificFetcher{
				newTestMetadataFetcher(mfTypes.PostMetadata{
					Author:        core.Account{Handle: "CryptoCapo_"},
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
