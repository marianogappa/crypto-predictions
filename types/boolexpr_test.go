package types

import (
	"reflect"
	"testing"

	"github.com/marianogappa/crypto-candles/candles/common"
)

func TestBoolExprEvaluate(t *testing.T) {
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
		expr     *BoolExpr
		expected ConditionStateValue
	}{
		{
			name:     "Nil expression evaluates to TRUE",
			expr:     nil,
			expected: TRUE,
		},
		{
			name:     "Literal true cond should evaluate to true",
			expr:     literalTrueBoolExpr,
			expected: TRUE,
		},
		{
			name:     "Literal false cond should evaluate to false",
			expr:     literalFalseBoolExpr,
			expected: FALSE,
		},
		{
			name:     "Literal undecided cond should evaluate to undecided",
			expr:     literalUndecidedBoolExpr,
			expected: UNDECIDED,
		},
		{
			name: "NOT with true cond should evaluate to false",
			expr: &BoolExpr{
				Operator: NOT,
				Operands: []*BoolExpr{literalTrueBoolExpr},
			},
			expected: FALSE,
		},
		{
			name: "NOT with false cond should evaluate to true",
			expr: &BoolExpr{
				Operator: NOT,
				Operands: []*BoolExpr{literalFalseBoolExpr},
			},
			expected: TRUE,
		},
		{
			name: "NOT with undecided cond should evaluate to undecided",
			expr: &BoolExpr{
				Operator: NOT,
				Operands: []*BoolExpr{literalUndecidedBoolExpr},
			},
			expected: UNDECIDED,
		},
		{
			name: "AND with zero operands evaluates to TRUE",
			expr: &BoolExpr{
				Operator: AND,
				Operands: []*BoolExpr{},
			},
			expected: TRUE,
		},
		{
			name: "OR with zero operands evaluates to TRUE",
			expr: &BoolExpr{
				Operator: OR,
				Operands: []*BoolExpr{},
			},
			expected: TRUE,
		},
		{
			name: "AND with one true operand evaluates to TRUE",
			expr: &BoolExpr{
				Operator: AND,
				Operands: []*BoolExpr{literalTrueBoolExpr},
			},
			expected: TRUE,
		},
		{
			name: "AND with one false operand evaluates to FALSE",
			expr: &BoolExpr{
				Operator: AND,
				Operands: []*BoolExpr{literalFalseBoolExpr},
			},
			expected: FALSE,
		},
		{
			name: "OR with one true operand evaluates to TRUE",
			expr: &BoolExpr{
				Operator: OR,
				Operands: []*BoolExpr{literalTrueBoolExpr},
			},
			expected: TRUE,
		},
		{
			name: "OR Literal with one false operand evaluates to FALSE",
			expr: &BoolExpr{
				Operator: OR,
				Operands: []*BoolExpr{literalFalseBoolExpr},
			},
			expected: FALSE,
		},
		{
			name: "AND with two true operands evaluates to TRUE",
			expr: &BoolExpr{
				Operator: AND,
				Operands: []*BoolExpr{literalTrueBoolExpr, literalTrueBoolExpr},
			},
			expected: TRUE,
		},
		{
			name: "AND with two operands, left false evaluates to FALSE",
			expr: &BoolExpr{
				Operator: AND,
				Operands: []*BoolExpr{literalFalseBoolExpr, literalTrueBoolExpr},
			},
			expected: FALSE,
		},
		{
			name: "AND with two operands, right false evaluates to FALSE",
			expr: &BoolExpr{
				Operator: AND,
				Operands: []*BoolExpr{literalTrueBoolExpr, literalFalseBoolExpr},
			},
			expected: FALSE,
		},
		{
			name: "OR with two true operands evaluates to TRUE",
			expr: &BoolExpr{
				Operator: OR,
				Operands: []*BoolExpr{literalTrueBoolExpr, literalTrueBoolExpr},
			},
			expected: TRUE,
		},
		{
			name: "OR with two operands, left false evaluates to TRUE",
			expr: &BoolExpr{
				Operator: OR,
				Operands: []*BoolExpr{literalFalseBoolExpr, literalTrueBoolExpr},
			},
			expected: TRUE,
		},
		{
			name: "OR with two operands, right false evaluates to TRUE",
			expr: &BoolExpr{
				Operator: OR,
				Operands: []*BoolExpr{literalTrueBoolExpr, literalFalseBoolExpr},
			},
			expected: TRUE,
		},
		{
			name: "AND with three true operands evaluates to TRUE",
			expr: &BoolExpr{
				Operator: AND,
				Operands: []*BoolExpr{literalTrueBoolExpr, literalTrueBoolExpr, literalTrueBoolExpr},
			},
			expected: TRUE,
		},
		{
			name: "AND with three operands, one false evaluates to FALSE",
			expr: &BoolExpr{
				Operator: AND,
				Operands: []*BoolExpr{literalTrueBoolExpr, literalFalseBoolExpr, literalTrueBoolExpr},
			},
			expected: FALSE,
		},
		{
			name: "OR with three false operands evaluates to FALSE",
			expr: &BoolExpr{
				Operator: OR,
				Operands: []*BoolExpr{literalFalseBoolExpr, literalFalseBoolExpr, literalFalseBoolExpr},
			},
			expected: FALSE,
		},
		{
			name: "OR with three operands, one true evaluates to TRUE",
			expr: &BoolExpr{
				Operator: OR,
				Operands: []*BoolExpr{literalFalseBoolExpr, literalTrueBoolExpr, literalFalseBoolExpr},
			},
			expected: TRUE,
		},
		{
			name: "NOT(NOT(TRUE)) == TRUE",
			expr: &BoolExpr{
				Operator: NOT,
				Operands: []*BoolExpr{
					{Operator: NOT, Operands: []*BoolExpr{literalTrueBoolExpr}},
				},
			},
			expected: TRUE,
		},
		{
			name: "NOT(NOT(FALSE)) == FALSE",
			expr: &BoolExpr{
				Operator: NOT,
				Operands: []*BoolExpr{
					{Operator: NOT, Operands: []*BoolExpr{literalFalseBoolExpr}},
				},
			},
			expected: FALSE,
		},
		{
			name: "NOT(NOT(UNDECIDED)) == UNDECIDED",
			expr: &BoolExpr{
				Operator: NOT,
				Operands: []*BoolExpr{
					{Operator: NOT, Operands: []*BoolExpr{literalUndecidedBoolExpr}},
				},
			},
			expected: UNDECIDED,
		},
		{
			name: "AND(NOT(FALSE), TRUE) == TRUE",
			expr: &BoolExpr{
				Operator: AND,
				Operands: []*BoolExpr{
					{Operator: NOT, Operands: []*BoolExpr{literalFalseBoolExpr}},
					literalTrueBoolExpr,
				},
			},
			expected: TRUE,
		},
		{
			name: "OR(NOT(TRUE), FALSE) == FALSE",
			expr: &BoolExpr{
				Operator: OR,
				Operands: []*BoolExpr{
					{Operator: NOT, Operands: []*BoolExpr{literalTrueBoolExpr}},
					literalFalseBoolExpr,
				},
			},
			expected: FALSE,
		},
		{
			name: "OR(NOT(TRUE), UNDECIDED) == UNDECIDED",
			expr: &BoolExpr{
				Operator: OR,
				Operands: []*BoolExpr{
					{Operator: NOT, Operands: []*BoolExpr{literalTrueBoolExpr}},
					literalUndecidedBoolExpr,
				},
			},
			expected: UNDECIDED,
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

func TestBoolExprUndecidedConditions(t *testing.T) {
	var (
		trueCond       = &Condition{State: ConditionState{Value: TRUE}}
		falseCond      = &Condition{State: ConditionState{Value: FALSE}}
		undecidedCond1 = &Condition{State: ConditionState{Value: UNDECIDED}}
		undecidedCond2 = &Condition{State: ConditionState{Value: UNDECIDED}}
		// undecidedCon3            = &Condition{State: ConditionState{Value: UNDECIDED}}
		literalTrueBoolExpr       = &BoolExpr{Operator: LITERAL, Operands: nil, Literal: trueCond}
		literalFalseBoolExpr      = &BoolExpr{Operator: LITERAL, Operands: nil, Literal: falseCond}
		literalUndecidedBoolExpr1 = &BoolExpr{Operator: LITERAL, Operands: nil, Literal: undecidedCond1}
		literalUndecidedBoolExpr2 = &BoolExpr{Operator: LITERAL, Operands: nil, Literal: undecidedCond2}
	)

	tss := []struct {
		name     string
		expr     *BoolExpr
		expected []*Condition
	}{
		{
			name:     "Nil expression evaluates to empty slice",
			expr:     nil,
			expected: []*Condition{},
		},
		{
			name:     "Literal true expression evaluates to empty slice",
			expr:     literalTrueBoolExpr,
			expected: []*Condition{},
		},
		{
			name:     "Literal false expression evaluates to empty slice",
			expr:     literalFalseBoolExpr,
			expected: []*Condition{},
		},
		{
			name:     "Literal undecided expression evaluates to one undecided condition",
			expr:     literalUndecidedBoolExpr1,
			expected: []*Condition{undecidedCond1},
		},
		{
			name: "NOT(UNDECIDED) expression evaluates to one undecided condition",
			expr: &BoolExpr{
				Operator: NOT, Operands: []*BoolExpr{literalUndecidedBoolExpr1},
			},
			expected: []*Condition{undecidedCond1},
		},
		{
			name: "AND(UNDECIDED, UNDECIDED) expression evaluates to two undecided conditions",
			expr: &BoolExpr{
				Operator: AND, Operands: []*BoolExpr{literalUndecidedBoolExpr1, literalUndecidedBoolExpr2},
			},
			expected: []*Condition{undecidedCond1, undecidedCond2},
		},
		{
			name: "AND(TRUE, UNDECIDED) expression evaluates to one undecided condition",
			expr: &BoolExpr{
				Operator: AND, Operands: []*BoolExpr{literalTrueBoolExpr, literalUndecidedBoolExpr1},
			},
			expected: []*Condition{undecidedCond1},
		},
		{
			name: "OR(UNDECIDED, FALSE) expression evaluates to one undecided condition",
			expr: &BoolExpr{
				Operator: OR, Operands: []*BoolExpr{literalUndecidedBoolExpr1, literalFalseBoolExpr},
			},
			expected: []*Condition{undecidedCond1},
		},
		{
			name: "OR(NOT(UNDECIDED), UNDECIDED) expression evaluates to two undecided conditions",
			expr: &BoolExpr{
				Operator: AND, Operands: []*BoolExpr{{
					Operator: NOT, Operands: []*BoolExpr{literalUndecidedBoolExpr1},
				}, literalUndecidedBoolExpr2},
			},
			expected: []*Condition{undecidedCond1, undecidedCond2},
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

func TestBoolExprClearState(t *testing.T) {
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

	c := &BoolExpr{
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
	c.ClearState()
	if !reflect.DeepEqual(*c, expected) {
		t.Errorf("expected state to be %v but was %v", expected, *c)
	}
}
