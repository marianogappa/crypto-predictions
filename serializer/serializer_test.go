package serializer

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/marianogappa/predictions/core"
	"github.com/stretchr/testify/require"
)

func TestSerialize(t *testing.T) {
	tss := []struct {
		name     string
		pred     core.Prediction
		err      error
		expected string
	}{
		{
			name: "Happy Case",
			pred: core.Prediction{
				Version:    "1.0.0",
				CreatedAt:  tpToISO("2020-01-03 00:00:00"),
				Reporter:   "admin",
				PostAuthor: "CryptoCapo_",
				PostText:   "",
				PostedAt:   tpToISO("2020-01-02 00:00:00"),
				PostURL:    "https://twitter.com/CryptoCapo_/status/1491357566974054400",
				Given: map[string]*core.Condition{
					"main": {
						Name:     "main",
						Operator: "<=",
						Operands: []core.Operand{
							{Type: core.COIN, Provider: "BINANCE", BaseAsset: "ADA", QuoteAsset: "USDT", Str: "COIN:BINANCE:ADA-USDT"},
							{Type: core.NUMBER, Number: core.JSONFloat64(0.845), Str: "0.845"},
						},
						FromTs:           int(tp("2020-01-02 00:00:00").Unix()),
						ToTs:             int(tp("2020-01-04 00:00:00").Unix()),
						ToDuration:       "2d",
						ErrorMarginRatio: 0.03,
					},
				},
				PrePredict: core.PrePredict{
					Predict: &core.BoolExpr{
						Operator: core.LITERAL,
						Operands: nil,
						Literal: &core.Condition{
							Name:     "main",
							Operator: "<=",
							Operands: []core.Operand{
								{Type: core.COIN, Provider: "BINANCE", BaseAsset: "ADA", QuoteAsset: "USDT", Str: "COIN:BINANCE:ADA-USDT"},
								{Type: core.NUMBER, Number: core.JSONFloat64(0.845), Str: "0.845"},
							},
							FromTs:           int(tp("2020-01-02 00:00:00").Unix()),
							ToTs:             int(tp("2020-01-04 00:00:00").Unix()),
							ToDuration:       "2d",
							ErrorMarginRatio: 0.03,
						},
					},
					AnnulledIfPredictIsFalse:          true,
					IgnoreUndecidedIfPredictIsDefined: true,
				},
				Predict: core.Predict{
					Predict: core.BoolExpr{
						Operator: core.LITERAL,
						Operands: nil,
						Literal: &core.Condition{
							Name:     "main",
							Operator: "<=",
							Operands: []core.Operand{
								{Type: core.COIN, Provider: "BINANCE", BaseAsset: "ADA", QuoteAsset: "USDT", Str: "COIN:BINANCE:ADA-USDT"},
								{Type: core.NUMBER, Number: core.JSONFloat64(0.845), Str: "0.845"},
							},
							FromTs:           int(tp("2020-01-02 00:00:00").Unix()),
							ToTs:             int(tp("2020-01-04 00:00:00").Unix()),
							ToDuration:       "2d",
							ErrorMarginRatio: 0.03,
						},
					},
					IgnoreUndecidedIfPredictIsDefined: true,
				},
				Type: core.PredictionTypeCoinOperatorFloatDeadline,
			},
			err: nil,
			expected: `{
				"uuid": "",
				"version": "1.0.0",
				"createdAt": "2020-01-03T00:00:00Z",
				"reporter": "admin",
				"postAuthor": "CryptoCapo_",
				"postedAt": "2020-01-02T00:00:00Z",
				"postUrl": "https://twitter.com/CryptoCapo_/status/1491357566974054400",
				"given":
				{
					"main":
					{
						"condition": "COIN:BINANCE:ADA-USDT <= 0.845",
						"fromISO8601": "2020-01-02T00:00:00Z",
						"toISO8601": "2020-01-04T00:00:00Z",
						"toDuration": "2d",
						"assumed": null,
						"state":
						{
							"status": "UNSTARTED",
							"lastTs": 0,
							"lastTicks": null,
							"value": "UNDECIDED"
						},
						"errorMarginRatio": 0.03
					}
				},
				"prePredict":
				{
					"predict": "main",
					"annulledIfPredictIsFalse": true,
					"ignoreUndecidedIfPredictIsDefined": true
				},
				"predict":
				{
					"predict": "main",
					"ignoreUndecidedIfPredictIsDefined": true
				},
				"state":
				{
					"status": "UNSTARTED",
					"lastTs": 0,
					"value": "ONGOING_PRE_PREDICTION"
				},
				"type": "PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE",
				"summary": {}
			}`,
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			var raw json.RawMessage
			err := json.Unmarshal([]byte(ts.expected), &raw)
			if err != nil {
				t.Fatalf("invalid JSON in test expectation, fix this! err: %v", err)
			}
			expected, _ := json.Marshal(raw)
			ps := NewPredictionSerializer(nil)
			actual, actualErr := ps.Serialize(&ts.pred)

			if actualErr != nil && ts.err == nil {
				t.Logf("expected no error but had '%v'", actualErr)
				t.FailNow()
			}
			if actualErr == nil && ts.err != nil {
				t.Logf("expected error '%v' but had no error", actualErr)
				t.FailNow()
			}
			if ts.err != nil && actualErr != nil && !errors.Is(actualErr, ts.err) {
				t.Logf("expected error '%v' but had error '%v'", ts.err, actualErr)
				t.FailNow()
			}
			if ts.err == nil {
				require.Equal(t, string(expected), string(actual))
			}
		})
	}
}

func tpToISO(s string) core.ISO8601 {
	t, _ := time.Parse("2006-01-02 15:04:05", s)
	return core.ISO8601(t.Format(time.RFC3339))
}

func tp(s string) time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", s)
	return t
}
