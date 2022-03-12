package compiler

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/marianogappa/predictions/types"
)

func TestSerialize(t *testing.T) {
	tss := []struct {
		name     string
		pred     types.Prediction
		err      error
		expected string
	}{
		{
			name: "Happy Case",
			pred: types.Prediction{
				Version:    "1.0.0",
				CreatedAt:  tpToISO("2020-01-03 00:00:00"),
				Reporter:   "admin",
				PostAuthor: "CryptoCapo_",
				PostText:   "",
				PostedAt:   tpToISO("2020-01-02 00:00:00"),
				PostUrl:    "https://twitter.com/CryptoCapo_/status/1491357566974054400",
				Given: map[string]*types.Condition{
					"main": {
						Name:     "main",
						Operator: "<=",
						Operands: []types.Operand{
							{Type: types.COIN, Provider: "BINANCE", BaseAsset: "ADA", QuoteAsset: "USDT", Str: "COIN:BINANCE:ADA-USDT"},
							{Type: types.NUMBER, Number: types.JsonFloat64(0.845), Str: "0.845"},
						},
						FromTs:           int(tp("2020-01-02 00:00:00").Unix()),
						ToTs:             int(tp("2020-01-04 00:00:00").Unix()),
						ToDuration:       "2d",
						ErrorMarginRatio: 0.03,
					},
				},
				PrePredict: types.PrePredict{
					Predict: &types.BoolExpr{
						Operator: types.LITERAL,
						Operands: nil,
						Literal: &types.Condition{
							Name:     "main",
							Operator: "<=",
							Operands: []types.Operand{
								{Type: types.COIN, Provider: "BINANCE", BaseAsset: "ADA", QuoteAsset: "USDT", Str: "COIN:BINANCE:ADA-USDT"},
								{Type: types.NUMBER, Number: types.JsonFloat64(0.845), Str: "0.845"},
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
				Predict: types.Predict{
					Predict: types.BoolExpr{
						Operator: types.LITERAL,
						Operands: nil,
						Literal: &types.Condition{
							Name:     "main",
							Operator: "<=",
							Operands: []types.Operand{
								{Type: types.COIN, Provider: "BINANCE", BaseAsset: "ADA", QuoteAsset: "USDT", Str: "COIN:BINANCE:ADA-USDT"},
								{Type: types.NUMBER, Number: types.JsonFloat64(0.845), Str: "0.845"},
							},
							FromTs:           int(tp("2020-01-02 00:00:00").Unix()),
							ToTs:             int(tp("2020-01-04 00:00:00").Unix()),
							ToDuration:       "2d",
							ErrorMarginRatio: 0.03,
						},
					},
					IgnoreUndecidedIfPredictIsDefined: true,
				},
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
				}
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
			ps := NewPredictionSerializer()
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
			if ts.err == nil && string(actual) != string(expected) {
				t.Logf("expected %v but got %v", string(expected), string(actual))
				t.FailNow()
			}
		})
	}
}
