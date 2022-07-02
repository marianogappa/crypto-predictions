package types

import (
	"reflect"
	"testing"
)

func TestPredictEvaluate(t *testing.T) {
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
		expr     Predict
		expected PredictionStateValue
	}{
		{
			name: "Only Predict evaluates to ONGOING_PREDICTION",
			expr: Predict{
				Predict: *literalUndecidedBoolExpr,
			},
			expected: ONGOINGPREDICTION,
		},
		{
			name: "Only Predict evaluates to CORRECT",
			expr: Predict{
				Predict: *literalTrueBoolExpr,
			},
			expected: CORRECT,
		},
		{
			name: "Only Predict evaluates to INCORRECT",
			expr: Predict{
				Predict: *literalFalseBoolExpr,
			},
			expected: INCORRECT,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=TRUE,Predict=UNDECIDED) evaluates to ANNULLED",
			expr: Predict{
				WrongIf:    literalTrueBoolExpr,
				AnnulledIf: literalTrueBoolExpr,
				Predict:    *literalUndecidedBoolExpr,
			},
			expected: ANNULLED,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=FALSE,Predict=UNDECIDED) evaluates to INCORRECT",
			expr: Predict{
				WrongIf:    literalTrueBoolExpr,
				AnnulledIf: literalFalseBoolExpr,
				Predict:    *literalUndecidedBoolExpr,
			},
			expected: INCORRECT,
		},
		{
			name: "(WrongIf=FALSE,AnnulledIf=TRUE,Predict=UNDECIDED) evaluates to ANNULLED",
			expr: Predict{
				WrongIf:    literalFalseBoolExpr,
				AnnulledIf: literalTrueBoolExpr,
				Predict:    *literalUndecidedBoolExpr,
			},
			expected: ANNULLED,
		},
		{
			name: "(WrongIf=FALSE,AnnulledIf=FALSE,Predict=UNDECIDED) evaluates to ONGOING_PREDICTION",
			expr: Predict{
				WrongIf:    literalFalseBoolExpr,
				AnnulledIf: literalFalseBoolExpr,
				Predict:    *literalUndecidedBoolExpr,
			},
			expected: ONGOINGPREDICTION,
		},
		{
			name: "(WrongIf=UNDECIDED,AnnulledIf=FALSE,Predict=UNDECIDED) evaluates to ONGOING_PREDICTION",
			expr: Predict{
				WrongIf:    literalUndecidedBoolExpr,
				AnnulledIf: literalFalseBoolExpr,
				Predict:    *literalUndecidedBoolExpr,
			},
			expected: ONGOINGPREDICTION,
		},
		{
			name: "(WrongIf=UNDECIDED,AnnulledIf=TRUE,Predict=UNDECIDED) evaluates to ANNULLED",
			expr: Predict{
				WrongIf:    literalUndecidedBoolExpr,
				AnnulledIf: literalTrueBoolExpr,
				Predict:    *literalUndecidedBoolExpr,
			},
			expected: ANNULLED,
		},
		{
			name: "(WrongIf=FALSE,AnnulledIf=UNDECIDED,Predict=UNDECIDED) evaluates to ONGOING_PREDICTION",
			expr: Predict{
				WrongIf:    literalFalseBoolExpr,
				AnnulledIf: literalUndecidedBoolExpr,
				Predict:    *literalUndecidedBoolExpr,
			},
			expected: ONGOINGPREDICTION,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=UNDECIDED,Predict=UNDECIDED) evaluates to ONGOING_PREDICTION",
			expr: Predict{
				WrongIf:    literalTrueBoolExpr,
				AnnulledIf: literalUndecidedBoolExpr,
				Predict:    *literalUndecidedBoolExpr,
			},
			expected: ONGOINGPREDICTION,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=TRUE,Predict=FALSE) evaluates to ANNULLED",
			expr: Predict{
				WrongIf:    literalTrueBoolExpr,
				AnnulledIf: literalTrueBoolExpr,
				Predict:    *literalFalseBoolExpr,
			},
			expected: ANNULLED,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=FALSE,Predict=FALSE) evaluates to INCORRECT",
			expr: Predict{
				WrongIf:    literalTrueBoolExpr,
				AnnulledIf: literalFalseBoolExpr,
				Predict:    *literalFalseBoolExpr,
			},
			expected: INCORRECT,
		},
		{
			name: "(WrongIf=FALSE,AnnulledIf=TRUE,Predict=FALSE) evaluates to ANNULLED",
			expr: Predict{
				WrongIf:    literalFalseBoolExpr,
				AnnulledIf: literalTrueBoolExpr,
				Predict:    *literalFalseBoolExpr,
			},
			expected: ANNULLED,
		},
		{
			name: "(WrongIf=FALSE,AnnulledIf=FALSE,Predict=FALSE) evaluates to INCORRECT",
			expr: Predict{
				WrongIf:    literalFalseBoolExpr,
				AnnulledIf: literalFalseBoolExpr,
				Predict:    *literalFalseBoolExpr,
			},
			expected: INCORRECT,
		},
		{
			name: "(WrongIf=UNDECIDED,AnnulledIf=FALSE,Predict=FALSE) evaluates to INCORRECT",
			expr: Predict{
				WrongIf:    literalUndecidedBoolExpr,
				AnnulledIf: literalFalseBoolExpr,
				Predict:    *literalFalseBoolExpr,
			},
			expected: INCORRECT,
		},
		{
			name: "(WrongIf=UNDECIDED,AnnulledIf=TRUE,Predict=FALSE) evaluates to ANNULLED",
			expr: Predict{
				WrongIf:    literalUndecidedBoolExpr,
				AnnulledIf: literalTrueBoolExpr,
				Predict:    *literalFalseBoolExpr,
			},
			expected: ANNULLED,
		},
		{
			name: "(WrongIf=FALSE,AnnulledIf=UNDECIDED,Predict=FALSE) evaluates to ONGOING_PREDICTION",
			expr: Predict{
				WrongIf:    literalFalseBoolExpr,
				AnnulledIf: literalUndecidedBoolExpr,
				Predict:    *literalFalseBoolExpr,
			},
			expected: ONGOINGPREDICTION,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=UNDECIDED,Predict=FALSE) evaluates to ONGOING_PREDICTION",
			expr: Predict{
				WrongIf:    literalTrueBoolExpr,
				AnnulledIf: literalUndecidedBoolExpr,
				Predict:    *literalFalseBoolExpr,
			},
			expected: ONGOINGPREDICTION,
		},
		// IgnoreUndecidedIfPredictIsDefined flag
		{
			name: "(WrongIf=FALSE,AnnulledIf=UNDECIDED,Predict=FALSE) w/IgnoreUndecided flag evaluates to INCORRECT",
			expr: Predict{
				WrongIf:                           literalFalseBoolExpr,
				AnnulledIf:                        literalUndecidedBoolExpr,
				Predict:                           *literalFalseBoolExpr,
				IgnoreUndecidedIfPredictIsDefined: true,
			},
			expected: INCORRECT,
		},
		{
			name: "(WrongIf=FALSE,AnnulledIf=UNDECIDED,Predict=TRUE) w/IgnoreUndecided flag evaluates to CORRECT",
			expr: Predict{
				WrongIf:                           literalFalseBoolExpr,
				AnnulledIf:                        literalUndecidedBoolExpr,
				Predict:                           *literalTrueBoolExpr,
				IgnoreUndecidedIfPredictIsDefined: true,
			},
			expected: CORRECT,
		},
		{
			name: "(WrongIf=UNDECIDED,AnnulledIf=UNDECIDED,Predict=FALSE) w/IgnoreUndecided flag evaluates to INCORRECT",
			expr: Predict{
				WrongIf:                           literalUndecidedBoolExpr,
				AnnulledIf:                        literalUndecidedBoolExpr,
				Predict:                           *literalFalseBoolExpr,
				IgnoreUndecidedIfPredictIsDefined: true,
			},
			expected: INCORRECT,
		},
		{
			name: "(WrongIf=UNDECIDED,AnnulledIf=UNDECIDED,Predict=TRUE) w/IgnoreUndecided flag evaluates to CORRECT",
			expr: Predict{
				WrongIf:                           literalUndecidedBoolExpr,
				AnnulledIf:                        literalUndecidedBoolExpr,
				Predict:                           *literalTrueBoolExpr,
				IgnoreUndecidedIfPredictIsDefined: true,
			},
			expected: CORRECT,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=UNDECIDED,Predict=TRUE) w/IgnoreUndecided flag evaluates to INCORRECT",
			expr: Predict{
				WrongIf:                           literalTrueBoolExpr,
				AnnulledIf:                        literalUndecidedBoolExpr,
				Predict:                           *literalTrueBoolExpr,
				IgnoreUndecidedIfPredictIsDefined: true,
			},
			expected: INCORRECT,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=UNDECIDED,Predict=FALSE) w/IgnoreUndecided flag evaluates to INCORRECT",
			expr: Predict{
				WrongIf:                           literalTrueBoolExpr,
				AnnulledIf:                        literalUndecidedBoolExpr,
				Predict:                           *literalFalseBoolExpr,
				IgnoreUndecidedIfPredictIsDefined: true,
			},
			expected: INCORRECT,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=UNDECIDED,Predict=TRUE) w/IgnoreUndecided flag evaluates to INCORRECT",
			expr: Predict{
				WrongIf:                           literalTrueBoolExpr,
				AnnulledIf:                        literalUndecidedBoolExpr,
				Predict:                           *literalTrueBoolExpr,
				IgnoreUndecidedIfPredictIsDefined: true,
			},
			expected: INCORRECT,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=FALSE,Predict=TRUE) w/IgnoreUndecided flag evaluates to INCORRECT",
			expr: Predict{
				WrongIf:                           literalTrueBoolExpr,
				AnnulledIf:                        literalFalseBoolExpr,
				Predict:                           *literalTrueBoolExpr,
				IgnoreUndecidedIfPredictIsDefined: true,
			},
			expected: INCORRECT,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=FALSE,Predict=FALSE) w/IgnoreUndecided flag evaluates to INCORRECT",
			expr: Predict{
				WrongIf:                           literalTrueBoolExpr,
				AnnulledIf:                        literalFalseBoolExpr,
				Predict:                           *literalFalseBoolExpr,
				IgnoreUndecidedIfPredictIsDefined: true,
			},
			expected: INCORRECT,
		},
		{
			name: "(WrongIf=TRUE,AnnulledIf=FALSE,Predict=TRUE) w/IgnoreUndecided flag evaluates to INCORRECT",
			expr: Predict{
				WrongIf:                           literalTrueBoolExpr,
				AnnulledIf:                        literalFalseBoolExpr,
				Predict:                           *literalTrueBoolExpr,
				IgnoreUndecidedIfPredictIsDefined: true,
			},
			expected: INCORRECT,
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

func TestPredictUndecidedConditions(t *testing.T) {
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
		expr     Predict
		expected []*Condition
	}{
		{
			name: "UndecidedConditions extracts literal undecided conditions from all sections",
			expr: Predict{
				WrongIf:    literalUndecidedBoolExpr1,
				AnnulledIf: literalUndecidedBoolExpr2,
				Predict:    *literalUndecidedBoolExpr3,
			},
			expected: []*Condition{undecidedCond1, undecidedCond2, undecidedCond3},
		},
		{
			name: "UndecidedConditions extracts literal undecided conditions from an AND",
			expr: Predict{
				WrongIf:    &BoolExpr{Operator: AND, Operands: []*BoolExpr{literalUndecidedBoolExpr1, literalUndecidedBoolExpr2}},
				AnnulledIf: literalTrueBoolExpr,
				Predict:    *literalFalseBoolExpr,
			},
			expected: []*Condition{undecidedCond1, undecidedCond2},
		},
		{
			name: "UndecidedConditions extracts literal undecided conditions from an OR",
			expr: Predict{
				WrongIf:    &BoolExpr{Operator: OR, Operands: []*BoolExpr{literalUndecidedBoolExpr1, literalUndecidedBoolExpr2}},
				AnnulledIf: literalTrueBoolExpr,
				Predict:    *literalUndecidedBoolExpr3,
			},
			expected: []*Condition{undecidedCond1, undecidedCond2, undecidedCond3},
		},
		{
			name: "UndecidedConditions extracts literal undecided conditions from a NOT",
			expr: Predict{
				WrongIf:    &BoolExpr{Operator: NOT, Operands: []*BoolExpr{literalUndecidedBoolExpr1}},
				AnnulledIf: literalTrueBoolExpr,
				Predict:    *literalUndecidedBoolExpr3,
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

func TestPredictClearState(t *testing.T) {
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

	pp := &Predict{
		AnnulledIf: &annulledIf,
		WrongIf:    &wrongIf,
		Predict:    predictIf,
	}

	pp.ClearState()
	if !reflect.DeepEqual(*pp.AnnulledIf, expected) {
		t.Errorf("expected AnnulledIf state to be %v but was %v", expected, pp.AnnulledIf)
	}
	if !reflect.DeepEqual(*pp.WrongIf, expected) {
		t.Errorf("expected WrongIf state to be %v but was %v", expected, pp.WrongIf)
	}
	if !reflect.DeepEqual(pp.Predict, expected) {
		t.Errorf("expected Predict state to be %v but was %v", expected, pp.Predict)
	}
}
