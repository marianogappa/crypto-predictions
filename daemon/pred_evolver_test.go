package daemon

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/marianogappa/crypto-candles/candles/common"
	"github.com/marianogappa/crypto-candles/candles/iterator"
	"github.com/marianogappa/predictions/core"
	"github.com/stretchr/testify/require"
)

func TestNewPredRunner(t *testing.T) {
	var (
		trueCond      = &core.Condition{State: core.ConditionState{Value: core.TRUE}}
		falseCond     = &core.Condition{State: core.ConditionState{Value: core.FALSE}}
		undecidedCond = &core.Condition{
			FromTs:   tInt("2022-02-27 15:20:00"),
			ToTs:     tInt("2022-03-27 15:20:00"),
			Operands: []core.Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
			State:    core.ConditionState{Value: core.UNDECIDED, LastTs: 0, LastTicks: map[string]common.Tick{}},
		}
		literalTrueBoolExpr      = &core.BoolExpr{Operator: core.LITERAL, Operands: nil, Literal: trueCond}
		literalFalseBoolExpr     = &core.BoolExpr{Operator: core.LITERAL, Operands: nil, Literal: falseCond}
		literalUndecidedBoolExpr = &core.BoolExpr{Operator: core.LITERAL, Operands: nil, Literal: undecidedCond}
	)

	tss := []struct {
		name        string
		prediction  core.Prediction
		nowTs       int
		isError     bool
		marketCalls []marketCall
	}{
		{
			name: "Correct prediction makes no calls",
			prediction: newPredictionWith(
				core.PrePredict{},
				core.Predict{
					Predict: *literalTrueBoolExpr,
				}),
			nowTs:       tInt("2022-02-27 15:20:00"),
			isError:     true,
			marketCalls: nil,
		},
		{
			name: "Incorrect prediction makes no calls",
			prediction: newPredictionWith(
				core.PrePredict{},
				core.Predict{
					Predict: *literalFalseBoolExpr,
				}),
			nowTs:       tInt("2022-02-27 15:20:00"),
			isError:     true,
			marketCalls: nil,
		},
		{
			name: "False pre-prediction makes no calls",
			prediction: newPredictionWith(
				core.PrePredict{
					Predict: literalFalseBoolExpr,
				},
				core.Predict{}),
			nowTs:       tInt("2022-02-27 15:20:00"),
			isError:     true,
			marketCalls: nil,
		},
		{
			name: "Wrong pre-prediction makes no calls",
			prediction: newPredictionWith(
				core.PrePredict{
					WrongIf: literalTrueBoolExpr,
				},
				core.Predict{}),
			nowTs:       tInt("2022-02-27 15:20:00"),
			isError:     true,
			marketCalls: nil,
		},
		{
			name: "Annulled pre-prediction makes no calls",
			prediction: newPredictionWith(
				core.PrePredict{
					AnnulledIf: literalTrueBoolExpr,
				},
				core.Predict{}),
			nowTs:       tInt("2022-02-27 15:20:00"),
			isError:     true,
			marketCalls: nil,
		},
		{
			name: "Annulled pre-prediction makes no calls even if all undecided boolexprs everywhere",
			prediction: newPredictionWith(
				core.PrePredict{
					AnnulledIf: literalTrueBoolExpr,
					WrongIf:    literalUndecidedBoolExpr,
					Predict:    literalUndecidedBoolExpr,
				},
				core.Predict{
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
				core.PrePredict{},
				core.Predict{
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
				core.PrePredict{
					AnnulledIf: literalFalseBoolExpr,
					WrongIf:    literalTrueBoolExpr,
					Predict:    literalUndecidedBoolExpr,
				},
				core.Predict{
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
				core.PrePredict{},
				core.Predict{
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
				core.PrePredict{},
				core.Predict{
					Predict: *literalUndecidedBoolExpr,
				}),
			nowTs:       tInt("2022-02-27 15:20:00"),
			isError:     false,
			marketCalls: []marketCall{{marketSource: marketSource("COIN:BINANCE:BTC-USDT"), tm: tp("2022-02-27 15:20:00")}},
		},
		{
			name: "Undecided prediction should make a call with start time in the future, in the next minute",
			prediction: newPredictionWith(
				core.PrePredict{},
				core.Predict{
					Predict: core.BoolExpr{Operator: core.LITERAL, Operands: nil, Literal: &core.Condition{
						FromTs:   tInt("2022-02-27 15:20:00"),
						ToTs:     tInt("2022-03-27 15:20:00"),
						Operands: []core.Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
						State: core.ConditionState{
							Value:     core.UNDECIDED,
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
				core.PrePredict{},
				core.Predict{
					Predict: core.BoolExpr{Operator: core.LITERAL, Operands: nil, Literal: &core.Condition{
						FromTs:   tInt("2022-02-27 15:20:00"),
						ToTs:     tInt("2022-03-27 15:20:00"),
						Operands: []core.Operand{operand("COIN:BINANCE:BTC-USDT"), operand("60000")},
						State: core.ConditionState{
							Value:     core.UNDECIDED,
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

func mapOperand(v string) (core.Operand, error) {
	v = strings.ToUpper(v)
	f, err := strconv.ParseFloat(v, 64)
	if err == nil {
		return core.Operand{Type: core.NUMBER, Number: core.JSONFloat64(f), Str: v}, nil
	}
	strVariable := `(COIN|MARKETCAP):([A-Z]+):([A-Z]+)(-([A-Z]+))?`
	rxVariable := regexp.MustCompile(fmt.Sprintf("^%v$", strVariable))
	matches := rxVariable.FindStringSubmatch(v)

	operandType, _ := core.OperandTypeFromString(matches[1])

	return core.Operand{
		Type:       operandType,
		Provider:   matches[2],
		BaseAsset:  matches[3],
		QuoteAsset: matches[5],
		Str:        v,
	}, nil
}

func operand(s string) core.Operand {
	op, _ := mapOperand(s)
	return op
}

func marketSource(s string) common.MarketSource {
	return operand(s).ToMarketSource()
}

func newPredictionWith(prePredict core.PrePredict, predict core.Predict) core.Prediction {
	return core.Prediction{
		UUID:       "ed47db4d-cc0b-4c3c-af18-e6fcbff82338",
		Version:    "1.0.0",
		CreatedAt:  core.ISO8601("2022-02-27 15:14:00"),
		PostAuthor: "JohnDoe",
		PostText:   "Test prediction!",
		PostedAt:   core.ISO8601("2022-02-27 15:14:00"),
		PostURL:    "https://twitter.com/trader1sz/status/1494458312238247950",
		Given:      map[string]*core.Condition{},
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

func (m *testMarket) Iterator(marketSource common.MarketSource, tm time.Time, candlestickInterval time.Duration) (iterator.Iterator, error) {
	m.calls = append(m.calls, marketCall{marketSource, tm, false})
	return testIterator{m}, nil
}

type testIterator struct{ mkt *testMarket }

func (i testIterator) NextTick() (common.Tick, error) {
	return common.Tick{}, nil
}

func (i testIterator) Next() (common.Candlestick, error) {
	return common.Candlestick{}, nil
}

// Not using the Scanner interface
func (i testIterator) Scan(*common.Candlestick) bool { return false }
func (i testIterator) Error() error                  { return nil }

func (i testIterator) SetStartFromNext(b bool)         { i.mkt.calls[len(i.mkt.calls)-1].startFromNext = b }
func (i testIterator) SetTimeNowFunc(func() time.Time) {}
