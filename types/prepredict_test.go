package types

import (
	"reflect"
	"testing"
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
			expected: ONGOING_PREDICTION,
		},
		{
			name: "Only PredictIf evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				PredictIf: literalUndecidedBoolExpr,
			},
			expected: ONGOING_PRE_PREDICTION,
		},
		{
			name: "Only PredictIf evaluates to ONGOING_PREDICTION",
			expr: PrePredict{
				PredictIf: literalTrueBoolExpr,
			},
			expected: ONGOING_PREDICTION,
		},
		{
			name: "Only PredictIf evaluates to INCORRECT",
			expr: PrePredict{
				PredictIf: literalFalseBoolExpr,
			},
			expected: INCORRECT,
		},
		{
			name: "Only WrongIf evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				WrongIf: literalUndecidedBoolExpr,
			},
			expected: ONGOING_PRE_PREDICTION,
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
			expected: ONGOING_PREDICTION,
		},
		{
			name: "Only AnnulledIf evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				AnnulledIf: literalUndecidedBoolExpr,
			},
			expected: ONGOING_PRE_PREDICTION,
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
			expected: ONGOING_PREDICTION,
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
			expected: ONGOING_PREDICTION,
		},
		{
			name: "(WrongIf=UNDECIDED,AnnulledIf=FALSE) evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				WrongIf:    literalUndecidedBoolExpr,
				AnnulledIf: literalFalseBoolExpr,
			},
			expected: ONGOING_PRE_PREDICTION,
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
			expected: ONGOING_PRE_PREDICTION,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=UNDECIDED) evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				WrongIf:    literalTrueBoolExpr,
				AnnulledIf: literalUndecidedBoolExpr,
			},
			expected: ONGOING_PRE_PREDICTION,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=TRUE,PredictIf=UNDECIDED) evaluates to ANNULLED",
			expr: PrePredict{
				WrongIf:    literalTrueBoolExpr,
				AnnulledIf: literalTrueBoolExpr,
				PredictIf:  literalUndecidedBoolExpr,
			},
			expected: ANNULLED,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=FALSE,PredictIf=UNDECIDED) evaluates to INCORRECT",
			expr: PrePredict{
				WrongIf:    literalTrueBoolExpr,
				AnnulledIf: literalFalseBoolExpr,
				PredictIf:  literalUndecidedBoolExpr,
			},
			expected: INCORRECT,
		},
		{
			name: "(WrongIf=FALSE,AnnulledIf=TRUE,PredictIf=UNDECIDED) evaluates to ANNULLED",
			expr: PrePredict{
				WrongIf:    literalFalseBoolExpr,
				AnnulledIf: literalTrueBoolExpr,
				PredictIf:  literalUndecidedBoolExpr,
			},
			expected: ANNULLED,
		},
		{
			name: "(WrongIf=FALSE,AnnulledIf=FALSE,PredictIf=UNDECIDED) evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				WrongIf:    literalFalseBoolExpr,
				AnnulledIf: literalFalseBoolExpr,
				PredictIf:  literalUndecidedBoolExpr,
			},
			expected: ONGOING_PRE_PREDICTION,
		},
		{
			name: "(WrongIf=UNDECIDED,AnnulledIf=FALSE,PredictIf=UNDECIDED) evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				WrongIf:    literalUndecidedBoolExpr,
				AnnulledIf: literalFalseBoolExpr,
				PredictIf:  literalUndecidedBoolExpr,
			},
			expected: ONGOING_PRE_PREDICTION,
		},
		{
			name: "(WrongIf=UNDECIDED,AnnulledIf=TRUE,PredictIf=UNDECIDED) evaluates to ANNULLED",
			expr: PrePredict{
				WrongIf:    literalUndecidedBoolExpr,
				AnnulledIf: literalTrueBoolExpr,
				PredictIf:  literalUndecidedBoolExpr,
			},
			expected: ANNULLED,
		},
		{
			name: "(WrongIf=FALSE,AnnulledIf=UNDECIDED,PredictIf=UNDECIDED) evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				WrongIf:    literalFalseBoolExpr,
				AnnulledIf: literalUndecidedBoolExpr,
				PredictIf:  literalUndecidedBoolExpr,
			},
			expected: ONGOING_PRE_PREDICTION,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=UNDECIDED,PredictIf=UNDECIDED) evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				WrongIf:    literalTrueBoolExpr,
				AnnulledIf: literalUndecidedBoolExpr,
				PredictIf:  literalUndecidedBoolExpr,
			},
			expected: ONGOING_PRE_PREDICTION,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=TRUE,PredictIf=FALSE) evaluates to ANNULLED",
			expr: PrePredict{
				WrongIf:    literalTrueBoolExpr,
				AnnulledIf: literalTrueBoolExpr,
				PredictIf:  literalFalseBoolExpr,
			},
			expected: ANNULLED,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=FALSE,PredictIf=FALSE) evaluates to INCORRECT",
			expr: PrePredict{
				WrongIf:    literalTrueBoolExpr,
				AnnulledIf: literalFalseBoolExpr,
				PredictIf:  literalFalseBoolExpr,
			},
			expected: INCORRECT,
		},
		{
			name: "(WrongIf=FALSE,AnnulledIf=TRUE,PredictIf=FALSE) evaluates to ANNULLED",
			expr: PrePredict{
				WrongIf:    literalFalseBoolExpr,
				AnnulledIf: literalTrueBoolExpr,
				PredictIf:  literalFalseBoolExpr,
			},
			expected: ANNULLED,
		},
		{
			name: "(WrongIf=FALSE,AnnulledIf=FALSE,PredictIf=FALSE) evaluates to INCORRECT",
			expr: PrePredict{
				WrongIf:    literalFalseBoolExpr,
				AnnulledIf: literalFalseBoolExpr,
				PredictIf:  literalFalseBoolExpr,
			},
			expected: INCORRECT,
		},
		{
			name: "(WrongIf=UNDECIDED,AnnulledIf=FALSE,PredictIf=FALSE) evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				WrongIf:    literalUndecidedBoolExpr,
				AnnulledIf: literalFalseBoolExpr,
				PredictIf:  literalFalseBoolExpr,
			},
			expected: INCORRECT,
		},
		{
			name: "(WrongIf=UNDECIDED,AnnulledIf=TRUE,PredictIf=FALSE) evaluates to ANNULLED",
			expr: PrePredict{
				WrongIf:    literalUndecidedBoolExpr,
				AnnulledIf: literalTrueBoolExpr,
				PredictIf:  literalFalseBoolExpr,
			},
			expected: ANNULLED,
		},
		{
			name: "(WrongIf=FALSE,AnnulledIf=UNDECIDED,PredictIf=FALSE) evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				WrongIf:    literalFalseBoolExpr,
				AnnulledIf: literalUndecidedBoolExpr,
				PredictIf:  literalFalseBoolExpr,
			},
			expected: ONGOING_PRE_PREDICTION,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=UNDECIDED,PredictIf=FALSE) evaluates to ONGOING_PRE_PREDICTION",
			expr: PrePredict{
				WrongIf:    literalTrueBoolExpr,
				AnnulledIf: literalUndecidedBoolExpr,
				PredictIf:  literalFalseBoolExpr,
			},
			expected: ONGOING_PRE_PREDICTION,
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
				PredictIf:  literalUndecidedBoolExpr3,
			},
			expected: []*Condition{undecidedCond1, undecidedCond2, undecidedCond3},
		},
		{
			name: "UndecidedConditions extracts literal undecided conditions from an AND",
			expr: PrePredict{
				WrongIf:    &BoolExpr{Operator: AND, Operands: []*BoolExpr{literalUndecidedBoolExpr1, literalUndecidedBoolExpr2}},
				AnnulledIf: literalTrueBoolExpr,
				PredictIf:  literalFalseBoolExpr,
			},
			expected: []*Condition{undecidedCond1, undecidedCond2},
		},
		{
			name: "UndecidedConditions extracts literal undecided conditions from an OR",
			expr: PrePredict{
				WrongIf:    &BoolExpr{Operator: OR, Operands: []*BoolExpr{literalUndecidedBoolExpr1, literalUndecidedBoolExpr2}},
				AnnulledIf: literalTrueBoolExpr,
				PredictIf:  literalUndecidedBoolExpr3,
			},
			expected: []*Condition{undecidedCond1, undecidedCond2, undecidedCond3},
		},
		{
			name: "UndecidedConditions extracts literal undecided conditions from a NOT",
			expr: PrePredict{
				WrongIf:    &BoolExpr{Operator: NOT, Operands: []*BoolExpr{literalUndecidedBoolExpr1}},
				AnnulledIf: literalTrueBoolExpr,
				PredictIf:  literalUndecidedBoolExpr3,
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
						LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: 12345, Value: 1000}},
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
						LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: 12345, Value: 1000}},
						Value:     TRUE,
					},
				},
			},
		},
		Literal: &Condition{
			State: ConditionState{
				Status:    FINISHED,
				LastTs:    12345,
				LastTicks: map[string]Tick{"COIN:BINANCE:BTC-USDT": {Timestamp: 12345, Value: 1000}},
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
		PredictIf:  &predictIf,
	}

	pp.ClearState()
	if !reflect.DeepEqual(*pp.AnnulledIf, expected) {
		t.Errorf("expected AnnulledIf state to be %v but was %v", expected, pp.AnnulledIf)
	}
	if !reflect.DeepEqual(*pp.WrongIf, expected) {
		t.Errorf("expected WrongIf state to be %v but was %v", expected, pp.WrongIf)
	}
	if !reflect.DeepEqual(*pp.PredictIf, expected) {
		t.Errorf("expected PredictIf state to be %v but was %v", expected, pp.PredictIf)
	}
}
