package daemon

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/marianogappa/predictions/market/common"
	"github.com/marianogappa/predictions/types"
	"github.com/stretchr/testify/require"
)

func TestNewPredRunner(t *testing.T) {
	var (
		trueCond      = &types.Condition{State: types.ConditionState{Value: types.TRUE}}
		falseCond     = &types.Condition{State: types.ConditionState{Value: types.FALSE}}
		undecidedCond = &types.Condition{
			FromTs:   tInt("2022-02-27 15:20:00"),
			ToTs:     tInt("2022-03-27 15:20:00"),
			Operands: []types.Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
			State:    types.ConditionState{Value: types.UNDECIDED, LastTs: 0, LastTicks: map[string]common.Tick{}},
		}
		literalTrueBoolExpr      = &types.BoolExpr{Operator: types.LITERAL, Operands: nil, Literal: trueCond}
		literalFalseBoolExpr     = &types.BoolExpr{Operator: types.LITERAL, Operands: nil, Literal: falseCond}
		literalUndecidedBoolExpr = &types.BoolExpr{Operator: types.LITERAL, Operands: nil, Literal: undecidedCond}
	)

	tss := []struct {
		name        string
		prediction  types.Prediction
		nowTs       int
		isError     bool
		marketCalls []marketCall
	}{
		{
			name: "Correct prediction makes no calls",
			prediction: newPredictionWith(
				types.PrePredict{},
				types.Predict{
					Predict: *literalTrueBoolExpr,
				}),
			nowTs:       tInt("2022-02-27 15:20:00"),
			isError:     true,
			marketCalls: nil,
		},
		{
			name: "Incorrect prediction makes no calls",
			prediction: newPredictionWith(
				types.PrePredict{},
				types.Predict{
					Predict: *literalFalseBoolExpr,
				}),
			nowTs:       tInt("2022-02-27 15:20:00"),
			isError:     true,
			marketCalls: nil,
		},
		{
			name: "False pre-prediction makes no calls",
			prediction: newPredictionWith(
				types.PrePredict{
					Predict: literalFalseBoolExpr,
				},
				types.Predict{}),
			nowTs:       tInt("2022-02-27 15:20:00"),
			isError:     true,
			marketCalls: nil,
		},
		{
			name: "Wrong pre-prediction makes no calls",
			prediction: newPredictionWith(
				types.PrePredict{
					WrongIf: literalTrueBoolExpr,
				},
				types.Predict{}),
			nowTs:       tInt("2022-02-27 15:20:00"),
			isError:     true,
			marketCalls: nil,
		},
		{
			name: "Annulled pre-prediction makes no calls",
			prediction: newPredictionWith(
				types.PrePredict{
					AnnulledIf: literalTrueBoolExpr,
				},
				types.Predict{}),
			nowTs:       tInt("2022-02-27 15:20:00"),
			isError:     true,
			marketCalls: nil,
		},
		{
			name: "Annulled pre-prediction makes no calls even if all undecided boolexprs everywhere",
			prediction: newPredictionWith(
				types.PrePredict{
					AnnulledIf: literalTrueBoolExpr,
					WrongIf:    literalUndecidedBoolExpr,
					Predict:    literalUndecidedBoolExpr,
				},
				types.Predict{
					AnnulledIf: literalUndecidedBoolExpr,
					WrongIf:    literalUndecidedBoolExpr,
					Predict:    *literalUndecidedBoolExpr,
				}),
			nowTs:       tInt("2022-02-27 15:20:00"),
			isError:     true,
			marketCalls: nil,
		},
		{
			name: "Annulled prediction makes no calls even if all undecided boolexprs everywhere",
			prediction: newPredictionWith(
				types.PrePredict{},
				types.Predict{
					AnnulledIf: literalTrueBoolExpr,
					WrongIf:    literalUndecidedBoolExpr,
					Predict:    *literalUndecidedBoolExpr,
				}),
			nowTs:       tInt("2022-02-27 15:20:00"),
			isError:     true,
			marketCalls: nil,
		},
		{
			name: "Wrong pre-prediction makes no calls even if all undecided boolexprs everywhere",
			prediction: newPredictionWith(
				types.PrePredict{
					AnnulledIf: literalFalseBoolExpr,
					WrongIf:    literalTrueBoolExpr,
					Predict:    literalUndecidedBoolExpr,
				},
				types.Predict{
					AnnulledIf: literalUndecidedBoolExpr,
					WrongIf:    literalUndecidedBoolExpr,
					Predict:    *literalUndecidedBoolExpr,
				}),
			nowTs:       tInt("2022-02-27 15:20:00"),
			isError:     true,
			marketCalls: nil,
		},
		{
			name: "Wrong prediction makes no calls even if all undecided boolexprs everywhere",
			prediction: newPredictionWith(
				types.PrePredict{},
				types.Predict{
					AnnulledIf: literalFalseBoolExpr,
					WrongIf:    literalTrueBoolExpr,
					Predict:    *literalUndecidedBoolExpr,
				}),
			nowTs:       tInt("2022-02-27 15:20:00"),
			isError:     true,
			marketCalls: nil,
		},
		{
			name: "Undecided prediction should make a call",
			prediction: newPredictionWith(
				types.PrePredict{},
				types.Predict{
					Predict: *literalUndecidedBoolExpr,
				}),
			nowTs:       tInt("2022-02-27 15:20:00"),
			isError:     false,
			marketCalls: []marketCall{{marketSource: marketSource("COIN:BINANCE:BTC-USDT"), tm: tp("2022-02-27 15:20:00")}},
		},
		{
			name: "Undecided prediction should make a call with start time in the future, in the next minute",
			prediction: newPredictionWith(
				types.PrePredict{},
				types.Predict{
					Predict: types.BoolExpr{Operator: types.LITERAL, Operands: nil, Literal: &types.Condition{
						FromTs:   tInt("2022-02-27 15:20:00"),
						ToTs:     tInt("2022-03-27 15:20:00"),
						Operands: []types.Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
						State: types.ConditionState{
							Value:     types.UNDECIDED,
							LastTs:    tInt("2022-02-27 16:20:00"),
							LastTicks: map[string]common.Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: tInt("2022-02-27 16:20:00"), Value: 60000}},
						},
					}},
				}),
			nowTs:       tInt("2022-02-28 16:20:00"),
			isError:     false,
			marketCalls: []marketCall{{marketSource: marketSource("COIN:BINANCE:BTC-USDT"), tm: tp("2022-02-27 16:20:00"), startFromNext: true}},
		},
		{
			name: "Undecided prediction should make a call in the future, which is fine because it should return ErrNoNewTicksYet",
			prediction: newPredictionWith(
				types.PrePredict{},
				types.Predict{
					Predict: types.BoolExpr{Operator: types.LITERAL, Operands: nil, Literal: &types.Condition{
						FromTs:   tInt("2022-02-27 15:20:00"),
						ToTs:     tInt("2022-03-27 15:20:00"),
						Operands: []types.Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
						State: types.ConditionState{
							Value:     types.UNDECIDED,
							LastTs:    tInt("2022-02-27 16:20:00"),
							LastTicks: map[string]common.Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: tInt("2022-02-27 16:20:00"), Value: 60000}},
						},
					}},
				}),
			nowTs:       tInt("2022-02-27 15:00:00"), // Earlier than when the tick iterator starts
			isError:     false,
			marketCalls: []marketCall{{marketSource: marketSource("COIN:BINANCE:BTC-USDT"), tm: tp("2022-02-27 16:20:00"), startFromNext: true}},
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			tm := &testMarket{}
			_, errs := NewPredEvolver(&ts.prediction, tm, ts.nowTs)
			if len(errs) > 0 && !ts.isError {
				t.Logf("should not have errored but these errors happened: %v", errs)
				t.FailNow()

			}
			if len(errs) == 0 && ts.isError {
				t.Log("should have errored but no errors happened")
				t.FailNow()

			}
			require.Equal(t, tm.calls, ts.marketCalls)
		})
	}
}

func mapOperand(v string) (types.Operand, error) {
	v = strings.ToUpper(v)
	f, err := strconv.ParseFloat(v, 64)
	if err == nil {
		return types.Operand{Type: types.NUMBER, Number: types.JSONFloat64(f), Str: v}, nil
	}
	strVariable := `(COIN|MARKETCAP):([A-Z]+):([A-Z]+)(-([A-Z]+))?`
	rxVariable := regexp.MustCompile(fmt.Sprintf("^%v$", strVariable))
	matches := rxVariable.FindStringSubmatch(v)

	operandType, _ := types.OperandTypeFromString(matches[1])

	return types.Operand{
		Type:       operandType,
		Provider:   matches[2],
		BaseAsset:  matches[3],
		QuoteAsset: matches[5],
		Str:        v,
	}, nil
}

func operand(s string) types.Operand {
	op, _ := mapOperand(s)
	return op
}

func marketSource(s string) common.MarketSource {
	return operand(s).ToMarketSource()
}

func newPredictionWith(prePredict types.PrePredict, predict types.Predict) types.Prediction {
	return types.Prediction{
		UUID:       "ed47db4d-cc0b-4c3c-af18-e6fcbff82338",
		Version:    "1.0.0",
		CreatedAt:  types.ISO8601("2022-02-27 15:14:00"),
		PostAuthor: "JohnDoe",
		PostText:   "Test prediction!",
		PostedAt:   types.ISO8601("2022-02-27 15:14:00"),
		PostURL:    "https://twitter.com/trader1sz/status/1494458312238247950",
		Given:      map[string]*types.Condition{},
		PrePredict: prePredict,
		Predict:    predict,
	}
}

func tp(s string) time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", s)
	t.UTC()
	return t
}

func tInt(s string) int {
	return int(tp(s).Unix())
}

type marketCall struct {
	marketSource  common.MarketSource
	tm            time.Time
	startFromNext bool
}

type testMarket struct {
	calls []marketCall
}

func (m *testMarket) GetIterator(marketSource common.MarketSource, tm time.Time, startFromNext bool, intervalMinutes int) (common.Iterator, error) {
	m.calls = append(m.calls, marketCall{marketSource, tm, startFromNext})
	return testIterator{}, nil
}

type testIterator struct{}

func (i testIterator) NextTick() (common.Tick, error) {
	return common.Tick{}, nil
}

func (i testIterator) NextCandlestick() (common.Candlestick, error) {
	return common.Candlestick{}, nil
}
func (i testIterator) IsOutOfTicks() bool {
	return true
}
