package core

import (
	"reflect"
	"testing"

	"github.com/marianogappa/crypto-candles/candles/common"
)

func TestPrePredictEvaluate(t *testing.T) {
	var (
		trueCond                 = &Condition{State: ConditionState{Value: TRUE}}
		falseCond                = &Condition{State: ConditionState{Value: FALSE}}
		undecidedCond            = &Condition{State: ConditionState{Value: UNDECIDED}}
		literalTrueBoolExpr      = &BoolExpr{Operator: LITERAL, Operands: nil, Literal: trueCond}
		literalFalseBoolExpr     = &BoolExpr{Operator: LITERAL, Operands: nil, Literal: falseCond}
		literalUndecidedBoolExpr = &BoolExpr{Operator: LITERAL, Operands: nil, Literal: undecidedCond}
	)

	tss := []struct {
		name     string
		expr     PrePredict
		expected PredictionStateValue
	}{
		{
			name:     "Empty PrePredict evaluates to ONGOING_PREDICTION",
			expr:     PrePredict{},
			expected: ONGOINGPREDICTION,
		},
		{
			name: "Only Predict evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				Predict: literalUndecidedBoolExpr,
			},
			expected: ONGOINGPREPREDICTION,
		},
		{
			name: "Only Predict evaluates to ONGOING_PREDICTION",
			expr: PrePredict{
				Predict: literalTrueBoolExpr,
			},
			expected: ONGOINGPREDICTION,
		},
		{
			name: "Only Predict evaluates to INCORRECT",
			expr: PrePredict{
				Predict: literalFalseBoolExpr,
			},
			expected: INCORRECT,
		},
		{
			name: "Only WrongIf evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				WrongIf: literalUndecidedBoolExpr,
			},
			expected: ONGOINGPREPREDICTION,
		},
		{
			name: "Only WrongIf evaluates to INCORRECT",
			expr: PrePredict{
				WrongIf: literalTrueBoolExpr,
			},
			expected: INCORRECT,
		},
		{
			name: "Only WrongIf evaluates to ONGOING_PREDICTION",
			expr: PrePredict{
				WrongIf: literalFalseBoolExpr,
			},
			expected: ONGOINGPREDICTION,
		},
		{
			name: "Only AnnulledIf evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				AnnulledIf: literalUndecidedBoolExpr,
			},
			expected: ONGOINGPREPREDICTION,
		},
		{
			name: "Only AnnulledIf evaluates to ANNULLED",
			expr: PrePredict{
				AnnulledIf: literalTrueBoolExpr,
			},
			expected: ANNULLED,
		},
		{
			name: "Only AnnulledIf evaluates to ONGOING_PREDICTION",
			expr: PrePredict{
				AnnulledIf: literalFalseBoolExpr,
			},
			expected: ONGOINGPREDICTION,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=TRUE) evaluates to ANNULLED",
			expr: PrePredict{
				WrongIf:    literalTrueBoolExpr,
				AnnulledIf: literalTrueBoolExpr,
			},
			expected: ANNULLED,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=FALSE) evaluates to INCORRECT",
			expr: PrePredict{
				WrongIf:    literalTrueBoolExpr,
				AnnulledIf: literalFalseBoolExpr,
			},
			expected: INCORRECT,
		},
		{
			name: "(WrongIf=FALSE,AnnulledIf=TRUE) evaluates to ANNULLED",
			expr: PrePredict{
				WrongIf:    literalFalseBoolExpr,
				AnnulledIf: literalTrueBoolExpr,
			},
			expected: ANNULLED,
		},
		{
			name: "(WrongIf=FALSE,AnnulledIf=FALSE) evaluates to ONGOING_PREDICTION",
			expr: PrePredict{
				WrongIf:    literalFalseBoolExpr,
				AnnulledIf: literalFalseBoolExpr,
			},
			expected: ONGOINGPREDICTION,
		},
		{
			name: "(WrongIf=UNDECIDED,AnnulledIf=FALSE) evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				WrongIf:    literalUndecidedBoolExpr,
				AnnulledIf: literalFalseBoolExpr,
			},
			expected: ONGOINGPREPREDICTION,
		},
		{
			name: "(WrongIf=UNDECIDED,AnnulledIf=TRUE) evaluates to ANNULLED",
			expr: PrePredict{
				WrongIf:    literalUndecidedBoolExpr,
				AnnulledIf: literalTrueBoolExpr,
			},
			expected: ANNULLED,
		},
		{
			name: "(WrongIf=FALSE,AnnulledIf=UNDECIDED) evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				WrongIf:    literalFalseBoolExpr,
				AnnulledIf: literalUndecidedBoolExpr,
			},
			expected: ONGOINGPREPREDICTION,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=UNDECIDED) evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				WrongIf:    literalTrueBoolExpr,
				AnnulledIf: literalUndecidedBoolExpr,
			},
			expected: ONGOINGPREPREDICTION,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=TRUE,Predict=UNDECIDED) evaluates to ANNULLED",
			expr: PrePredict{
				WrongIf:    literalTrueBoolExpr,
				AnnulledIf: literalTrueBoolExpr,
				Predict:    literalUndecidedBoolExpr,
			},
			expected: ANNULLED,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=FALSE,Predict=UNDECIDED) evaluates to INCORRECT",
			expr: PrePredict{
				WrongIf:    literalTrueBoolExpr,
				AnnulledIf: literalFalseBoolExpr,
				Predict:    literalUndecidedBoolExpr,
			},
			expected: INCORRECT,
		},
		{
			name: "(WrongIf=FALSE,AnnulledIf=TRUE,Predict=UNDECIDED) evaluates to ANNULLED",
			expr: PrePredict{
				WrongIf:    literalFalseBoolExpr,
				AnnulledIf: literalTrueBoolExpr,
				Predict:    literalUndecidedBoolExpr,
			},
			expected: ANNULLED,
		},
		{
			name: "(WrongIf=FALSE,AnnulledIf=FALSE,Predict=UNDECIDED) evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				WrongIf:    literalFalseBoolExpr,
				AnnulledIf: literalFalseBoolExpr,
				Predict:    literalUndecidedBoolExpr,
			},
			expected: ONGOINGPREPREDICTION,
		},
		{
			name: "(WrongIf=UNDECIDED,AnnulledIf=FALSE,Predict=UNDECIDED) evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				WrongIf:    literalUndecidedBoolExpr,
				AnnulledIf: literalFalseBoolExpr,
				Predict:    literalUndecidedBoolExpr,
			},
			expected: ONGOINGPREPREDICTION,
		},
		{
			name: "(WrongIf=UNDECIDED,AnnulledIf=TRUE,Predict=UNDECIDED) evaluates to ANNULLED",
			expr: PrePredict{
				WrongIf:    literalUndecidedBoolExpr,
				AnnulledIf: literalTrueBoolExpr,
				Predict:    literalUndecidedBoolExpr,
			},
			expected: ANNULLED,
		},
		{
			name: "(WrongIf=FALSE,AnnulledIf=UNDECIDED,Predict=UNDECIDED) evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				WrongIf:    literalFalseBoolExpr,
				AnnulledIf: literalUndecidedBoolExpr,
				Predict:    literalUndecidedBoolExpr,
			},
			expected: ONGOINGPREPREDICTION,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=UNDECIDED,Predict=UNDECIDED) evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				WrongIf:    literalTrueBoolExpr,
				AnnulledIf: literalUndecidedBoolExpr,
				Predict:    literalUndecidedBoolExpr,
			},
			expected: ONGOINGPREPREDICTION,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=TRUE,Predict=FALSE) evaluates to ANNULLED",
			expr: PrePredict{
				WrongIf:    literalTrueBoolExpr,
				AnnulledIf: literalTrueBoolExpr,
				Predict:    literalFalseBoolExpr,
			},
			expected: ANNULLED,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=FALSE,Predict=FALSE) evaluates to INCORRECT",
			expr: PrePredict{
				WrongIf:    literalTrueBoolExpr,
				AnnulledIf: literalFalseBoolExpr,
				Predict:    literalFalseBoolExpr,
			},
			expected: INCORRECT,
		},
		{
			name: "(WrongIf=FALSE,AnnulledIf=TRUE,Predict=FALSE) evaluates to ANNULLED",
			expr: PrePredict{
				WrongIf:    literalFalseBoolExpr,
				AnnulledIf: literalTrueBoolExpr,
				Predict:    literalFalseBoolExpr,
			},
			expected: ANNULLED,
		},
		{
			name: "(WrongIf=FALSE,AnnulledIf=FALSE,Predict=FALSE) evaluates to INCORRECT",
			expr: PrePredict{
				WrongIf:    literalFalseBoolExpr,
				AnnulledIf: literalFalseBoolExpr,
				Predict:    literalFalseBoolExpr,
			},
			expected: INCORRECT,
		},
		{
			name: "(WrongIf=UNDECIDED,AnnulledIf=FALSE,Predict=FALSE) evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				WrongIf:    literalUndecidedBoolExpr,
				AnnulledIf: literalFalseBoolExpr,
				Predict:    literalFalseBoolExpr,
			},
			expected: INCORRECT,
		},
		{
			name: "(WrongIf=UNDECIDED,AnnulledIf=TRUE,Predict=FALSE) evaluates to ANNULLED",
			expr: PrePredict{
				WrongIf:    literalUndecidedBoolExpr,
				AnnulledIf: literalTrueBoolExpr,
				Predict:    literalFalseBoolExpr,
			},
			expected: ANNULLED,
		},
		{
			name: "(WrongIf=FALSE,AnnulledIf=UNDECIDED,Predict=FALSE) evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				WrongIf:    literalFalseBoolExpr,
				AnnulledIf: literalUndecidedBoolExpr,
				Predict:    literalFalseBoolExpr,
			},
			expected: ONGOINGPREPREDICTION,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=UNDECIDED,Predict=FALSE) evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				WrongIf:    literalTrueBoolExpr,
				AnnulledIf: literalUndecidedBoolExpr,
				Predict:    literalFalseBoolExpr,
			},
			expected: ONGOINGPREPREDICTION,
		},
		// AnnulledIfPredictIsFalse flag
		{
			name: "(WrongIf=FALSE,AnnulledIf=FALSE,Predict=FALSE w/Annulled flag) evaluates to ANULLED",
			expr: PrePredict{
				WrongIf:                  literalFalseBoolExpr,
				AnnulledIf:               literalFalseBoolExpr,
				Predict:                  literalFalseBoolExpr,
				AnnulledIfPredictIsFalse: true,
			},
			expected: ANNULLED,
		},
		{
			name: "(WrongIf=FALSE,AnnulledIf=FALSE,Predict=TRUE w/Annulled flag) evaluates to ONGOING_PREDICTION",
			expr: PrePredict{
				WrongIf:                  literalFalseBoolExpr,
				AnnulledIf:               literalFalseBoolExpr,
				Predict:                  literalTrueBoolExpr,
				AnnulledIfPredictIsFalse: true,
			},
			expected: ONGOINGPREDICTION,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=FALSE,Predict=TRUE w/Annulled flag) evaluates to INCORRECT",
			expr: PrePredict{
				WrongIf:                  literalTrueBoolExpr,
				AnnulledIf:               literalFalseBoolExpr,
				Predict:                  literalTrueBoolExpr,
				AnnulledIfPredictIsFalse: true,
			},
			expected: INCORRECT,
		},
		{
			name: "(WrongIf=UNDECIDED,AnnulledIf=UNDECIDED,Predict=FALSE w/Annulled flag) evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				WrongIf:                  literalUndecidedBoolExpr,
				AnnulledIf:               literalUndecidedBoolExpr,
				Predict:                  literalFalseBoolExpr,
				AnnulledIfPredictIsFalse: true,
			},
			expected: ONGOINGPREPREDICTION,
		},
		// IgnoreUndecidedIfPredictIsDefined flag
		{
			name: "(WrongIf=UNDECIDED,AnnulledIf=UNDECIDED,Predict=FALSE w/IgnoreUndecided flag) evaluates to INCORRECT",
			expr: PrePredict{
				WrongIf:                           literalUndecidedBoolExpr,
				AnnulledIf:                        literalUndecidedBoolExpr,
				Predict:                           literalFalseBoolExpr,
				IgnoreUndecidedIfPredictIsDefined: true,
			},
			expected: INCORRECT,
		},
		{
			name: "(WrongIf=UNDECIDED,AnnulledIf=UNDECIDED,Predict=TRUE w/IgnoreUndecided flag) evaluates to ONGOING_PREDICTION",
			expr: PrePredict{
				WrongIf:                           literalUndecidedBoolExpr,
				AnnulledIf:                        literalUndecidedBoolExpr,
				Predict:                           literalTrueBoolExpr,
				IgnoreUndecidedIfPredictIsDefined: true,
			},
			expected: ONGOINGPREDICTION,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=UNDECIDED,Predict=TRUE w/IgnoreUndecided flag) evaluates to INCORRECT",
			expr: PrePredict{
				WrongIf:                           literalTrueBoolExpr,
				AnnulledIf:                        literalUndecidedBoolExpr,
				Predict:                           literalTrueBoolExpr,
				IgnoreUndecidedIfPredictIsDefined: true,
			},
			expected: INCORRECT,
		},
		{
			name: "(WrongIf=UNDECIDED,AnnulledIf=TRUE,Predict=TRUE w/IgnoreUndecided flag) evaluates to ANNULLED",
			expr: PrePredict{
				WrongIf:                           literalUndecidedBoolExpr,
				AnnulledIf:                        literalTrueBoolExpr,
				Predict:                           literalTrueBoolExpr,
				IgnoreUndecidedIfPredictIsDefined: true,
			},
			expected: ANNULLED,
		},
		// AnnulledIfPredictIsFalse flag && IgnoreUndecidedIfPredictIsDefined flag
		{
			name: "(WrongIf=UNDECIDED,AnnulledIf=UNDECIDED,Predict=FALSE w/IgnoreUndecided & Annulled flag) evaluates to ANNULLED",
			expr: PrePredict{
				WrongIf:                           literalUndecidedBoolExpr,
				AnnulledIf:                        literalUndecidedBoolExpr,
				Predict:                           literalFalseBoolExpr,
				IgnoreUndecidedIfPredictIsDefined: true,
				AnnulledIfPredictIsFalse:          true,
			},
			expected: ANNULLED,
		},
		{
			name: "(WrongIf=UNDECIDED,AnnulledIf=UNDECIDED,Predict=TRUE w/IgnoreUndecided & Annulled flag) evaluates to ONGOING_PREDICTION",
			expr: PrePredict{
				WrongIf:                           literalUndecidedBoolExpr,
				AnnulledIf:                        literalUndecidedBoolExpr,
				Predict:                           literalTrueBoolExpr,
				IgnoreUndecidedIfPredictIsDefined: true,
				AnnulledIfPredictIsFalse:          true,
			},
			expected: ONGOINGPREDICTION,
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			actual := ts.expr.Evaluate()
			if !reflect.DeepEqual(actual, ts.expected) {
				t.Logf("expected %v but got %v", ts.expected, actual)
				t.FailNow()
			}
		})
	}
}

func TestPrePredictUndecidedConditions(t *testing.T) {
	var (
		trueCond       = &Condition{State: ConditionState{Value: TRUE}}
		falseCond      = &Condition{State: ConditionState{Value: FALSE}}
		undecidedCond1 = &Condition{State: ConditionState{Value: UNDECIDED}}
		undecidedCond2 = &Condition{State: ConditionState{Value: UNDECIDED}}
		undecidedCond3 = &Condition{State: ConditionState{Value: UNDECIDED}}

		literalTrueBoolExpr       = &BoolExpr{Operator: LITERAL, Operands: nil, Literal: trueCond}
		literalFalseBoolExpr      = &BoolExpr{Operator: LITERAL, Operands: nil, Literal: falseCond}
		literalUndecidedBoolExpr1 = &BoolExpr{Operator: LITERAL, Operands: nil, Literal: undecidedCond1}
		literalUndecidedBoolExpr2 = &BoolExpr{Operator: LITERAL, Operands: nil, Literal: undecidedCond2}
		literalUndecidedBoolExpr3 = &BoolExpr{Operator: LITERAL, Operands: nil, Literal: undecidedCond3}
	)

	tss := []struct {
		name     string
		expr     PrePredict
		expected []*Condition
	}{
		{
			name: "UndecidedConditions extracts literal undecided conditions from all sections",
			expr: PrePredict{
				WrongIf:    literalUndecidedBoolExpr1,
				AnnulledIf: literalUndecidedBoolExpr2,
				Predict:    literalUndecidedBoolExpr3,
			},
			expected: []*Condition{undecidedCond1, undecidedCond2, undecidedCond3},
		},
		{
			name: "UndecidedConditions extracts literal undecided conditions from an AND",
			expr: PrePredict{
				WrongIf:    &BoolExpr{Operator: AND, Operands: []*BoolExpr{literalUndecidedBoolExpr1, literalUndecidedBoolExpr2}},
				AnnulledIf: literalTrueBoolExpr,
				Predict:    literalFalseBoolExpr,
			},
			expected: []*Condition{undecidedCond1, undecidedCond2},
		},
		{
			name: "UndecidedConditions extracts literal undecided conditions from an OR",
			expr: PrePredict{
				WrongIf:    &BoolExpr{Operator: OR, Operands: []*BoolExpr{literalUndecidedBoolExpr1, literalUndecidedBoolExpr2}},
				AnnulledIf: literalTrueBoolExpr,
				Predict:    literalUndecidedBoolExpr3,
			},
			expected: []*Condition{undecidedCond1, undecidedCond2, undecidedCond3},
		},
		{
			name: "UndecidedConditions extracts literal undecided conditions from a NOT",
			expr: PrePredict{
				WrongIf:    &BoolExpr{Operator: NOT, Operands: []*BoolExpr{literalUndecidedBoolExpr1}},
				AnnulledIf: literalTrueBoolExpr,
				Predict:    literalUndecidedBoolExpr3,
			},
			expected: []*Condition{undecidedCond1, undecidedCond3},
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			actual := ts.expr.UndecidedConditions()
			if !reflect.DeepEqual(actual, ts.expected) {
				t.Logf("expected %v but got %v", ts.expected, actual)
				t.FailNow()
			}
		})
	}
}

func TestPrePredictClearState(t *testing.T) {
	expected := BoolExpr{
		Operator: AND,
		Operands: []*BoolExpr{
			{
				Operator: LITERAL,
				Literal: &Condition{
					State: ConditionState{
						Status:    UNSTARTED,
						LastTs:    0,
						LastTicks: nil,
						Value:     UNDECIDED,
					},
				},
			},
			{
				Operator: LITERAL,
				Literal: &Condition{
					State: ConditionState{
						Status:    UNSTARTED,
						LastTs:    0,
						LastTicks: nil,
						Value:     UNDECIDED,
					},
				},
			},
		},
		Literal: &Condition{
			State: ConditionState{
				Status:    UNSTARTED,
				LastTs:    0,
				LastTicks: nil,
				Value:     UNDECIDED,
			},
		},
	}

	be := BoolExpr{
		Operator: AND,
		Operands: []*BoolExpr{
			{
				Operator: LITERAL,
				Literal: &Condition{
					State: ConditionState{
						Status:    FINISHED,
						LastTs:    12345,
						LastTicks: map[string]common.Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: 12345, Value: 1000}},
						Value:     TRUE,
					},
				},
			},
			{
				Operator: LITERAL,
				Literal: &Condition{
					State: ConditionState{
						Status:    FINISHED,
						LastTs:    12345,
						LastTicks: map[string]common.Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: 12345, Value: 1000}},
						Value:     TRUE,
					},
				},
			},
		},
		Literal: &Condition{
			State: ConditionState{
				Status:    FINISHED,
				LastTs:    12345,
				LastTicks: map[string]common.Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: 12345, Value: 1000}},
				Value:     TRUE,
			},
		},
	}
	annulledIf := be
	wrongIf := be
	predictIf := be

	pp := &PrePredict{
		AnnulledIf: &annulledIf,
		WrongIf:    &wrongIf,
		Predict:    &predictIf,
	}

	pp.ClearState()
	if !reflect.DeepEqual(*pp.AnnulledIf, expected) {
		t.Errorf("expected AnnulledIf state to be %v but was %v", expected, pp.AnnulledIf)
	}
	if !reflect.DeepEqual(*pp.WrongIf, expected) {
		t.Errorf("expected WrongIf state to be %v but was %v", expected, pp.WrongIf)
	}
	if !reflect.DeepEqual(*pp.Predict, expected) {
		t.Errorf("expected Predict state to be %v but was %v", expected, pp.Predict)
	}
}
