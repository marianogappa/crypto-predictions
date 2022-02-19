package types

import (
	"fmt"

	"github.com/marianogappa/signal-checker/common"
)

type Operand struct {
	Type       OperandType
	Provider   string
	QuoteAsset string
	BaseAsset  string
	Number     common.JsonFloat64
	Str        string
}

type OperandType int

const (
	NUMBER OperandType = iota
	COIN
	MARKETCAP
)

func OperandTypeFromString(s string) (OperandType, error) {
	switch s {
	case "NUMBER", "":
		return NUMBER, nil
	case "COIN":
		return COIN, nil
	case "MARKETCAP":
		return MARKETCAP, nil
	default:
		return 0, fmt.Errorf("unknown value for OperandType: %v", s)
	}
}
func (v OperandType) String() string {
	switch v {
	case NUMBER:
		return "NUMBER"
	case COIN:
		return "COIN"
	case MARKETCAP:
		return "MARKETCAP"
	default:
		return ""
	}
}

type ConditionState struct {
	Status ConditionStatus
	LastTs int
	Value  ConditionStateValue
	// add state to provide evidence of alleged condition result
}

func (a ConditionStateValue) And(b ConditionStateValue) ConditionStateValue {
	if a == FALSE || b == FALSE {
		return FALSE
	}
	if a != UNDECIDED || b != UNDECIDED {
		return UNDECIDED
	}
	return TRUE
}

func (a ConditionStateValue) Or(b ConditionStateValue) ConditionStateValue {
	if a == TRUE || b == TRUE {
		return TRUE
	}
	if a != UNDECIDED || b != UNDECIDED {
		return UNDECIDED
	}
	return FALSE
}

func (a ConditionStateValue) Not() ConditionStateValue {
	switch a {
	case UNDECIDED:
		return UNDECIDED
	case TRUE:
		return FALSE
	default: // FALSE
		return TRUE
	}
}

type BoolOperator int

func BoolOperatorFromString(s string) (BoolOperator, error) {
	switch s {
	case "LITERAL":
		return LITERAL, nil
	case "AND":
		return AND, nil
	case "OR":
		return OR, nil
	case "NOT":
		return NOT, nil
	default:
		return 0, fmt.Errorf("unknown value for BoolOperator: %v", s)
	}
}

type ConditionStatus int

func ConditionStatusFromString(s string) (ConditionStatus, error) {
	switch s {
	case "UNSTARTED", "":
		return UNSTARTED, nil
	case "STARTED":
		return STARTED, nil
	case "FINISHED":
		return FINISHED, nil
	default:
		return 0, fmt.Errorf("unknown value for ConditionStatus: %v", s)
	}
}
func (v ConditionStatus) String() string {
	switch v {
	case UNSTARTED:
		return "UNSTARTED"
	case STARTED:
		return "STARTED"
	case FINISHED:
		return "FINISHED"
	default:
		return ""
	}
}

type ConditionStateValue int

func ConditionStateValueFromString(s string) (ConditionStateValue, error) {
	switch s {
	case "UNDECIDED", "":
		return UNDECIDED, nil
	case "TRUE":
		return TRUE, nil
	case "FALSE":
		return FALSE, nil
	default:
		return 0, fmt.Errorf("unknown value for ConditionStateValue: %v", s)
	}
}
func (v ConditionStateValue) String() string {
	switch v {
	case UNDECIDED:
		return "UNDECIDED"
	case TRUE:
		return "TRUE"
	case FALSE:
		return "FALSE"
	default:
		return ""
	}
}

type PredictionStateValue int

func PredictionStateValueFromString(s string) (PredictionStateValue, error) {
	switch s {
	case "ONGOING_PRE_PREDICTION", "":
		return ONGOING_PRE_PREDICTION, nil
	case "ONGOING_PREDICTION":
		return ONGOING_PREDICTION, nil
	case "CORRECT":
		return CORRECT, nil
	case "INCORRECT":
		return INCORRECT, nil
	default:
		return 0, fmt.Errorf("unknown value for PredictionStateValue: %v", s)
	}
}
func (v PredictionStateValue) String() string {
	switch v {
	case ONGOING_PRE_PREDICTION:
		return "ONGOING_PRE_PREDICTION"
	case ONGOING_PREDICTION:
		return "ONGOING_PREDICTION"
	case CORRECT:
		return "CORRECT"
	case INCORRECT:
		return "INCORRECT"
	default:
		return ""
	}
}

const (
	LITERAL BoolOperator = iota
	AND
	OR
	NOT
)

const (
	UNSTARTED ConditionStatus = iota
	STARTED
	FINISHED
)

const (
	UNDECIDED ConditionStateValue = iota
	TRUE
	FALSE
)

const (
	ONGOING_PRE_PREDICTION PredictionStateValue = iota
	ONGOING_PREDICTION
	CORRECT
	INCORRECT
	ANNULLED
)

type BoolExpr struct {
	Operator BoolOperator
	Operands []*BoolExpr
	Literal  *Condition
}

func (p *BoolExpr) UndecidedConditions() []*Condition {
	conds := []*Condition{}
	if p == nil || (p.Operator == LITERAL && p.Literal == nil) {
		return conds
	}
	switch p.Operator {
	case AND, OR:
		for _, operand := range p.Operands {
			conds = append(conds, operand.UndecidedConditions()...)
		}
	default:
		if (*p).Literal.Evaluate() == UNDECIDED {
			conds = append(conds, p.Literal)
		}
	}
	return conds
}

func (p *BoolExpr) Evaluate() ConditionStateValue {
	if p == nil {
		return TRUE
	}
	switch p.Operator {
	case AND:
		if len(p.Operands) == 0 {
			return TRUE
		}
		result := p.Operands[0].Evaluate()
		for _, operand := range p.Operands[1:] {
			result = result.And(operand.Evaluate())
		}
		return result
	case OR:
		if len(p.Operands) == 0 {
			return TRUE
		}
		result := p.Operands[0].Evaluate()
		for _, operand := range p.Operands[1:] {
			result = result.Or(operand.Evaluate())
		}
		return result
	case NOT:
		return p.Literal.Evaluate().Not()
	default:
		return p.Literal.Evaluate()
	}
}

type PrePredict struct {
	WrongIf    *BoolExpr
	AnnulledIf *BoolExpr
	PredictIf  *BoolExpr
}

func (p PrePredict) UndecidedConditions() []*Condition {
	conds := []*Condition{}
	conds = append(conds, p.WrongIf.UndecidedConditions()...)
	conds = append(conds, p.AnnulledIf.UndecidedConditions()...)
	conds = append(conds, p.PredictIf.UndecidedConditions()...)
	return conds
}

func (p PrePredict) Evaluate() PredictionStateValue {
	if p.WrongIf == nil && p.AnnulledIf == nil && p.PredictIf == nil {
		return CORRECT
	}
	var (
		wrongIfValue    = FALSE
		annulledIfValue = FALSE
		predictIfValue  = TRUE
	)
	if p.WrongIf != nil {
		wrongIfValue = p.WrongIf.Evaluate()
	}
	if wrongIfValue == TRUE {
		return INCORRECT
	}
	if p.AnnulledIf != nil {
		annulledIfValue = p.AnnulledIf.Evaluate()
	}
	if annulledIfValue == TRUE {
		return ANNULLED
	}
	if p.PredictIf != nil {
		predictIfValue = p.PredictIf.Evaluate()
	}
	if wrongIfValue == UNDECIDED || annulledIfValue == UNDECIDED || predictIfValue == UNDECIDED {
		return ONGOING_PRE_PREDICTION
	}
	if predictIfValue == FALSE {
		return INCORRECT
	}
	return ONGOING_PREDICTION
}

type Predict struct {
	WrongIf    *BoolExpr
	AnnulledIf *BoolExpr
	Predict    BoolExpr
}

func (p Predict) UndecidedConditions() []*Condition {
	conds := []*Condition{}
	conds = append(conds, p.WrongIf.UndecidedConditions()...)
	conds = append(conds, p.AnnulledIf.UndecidedConditions()...)
	conds = append(conds, p.Predict.UndecidedConditions()...)
	return conds
}

func (p Predict) Evaluate() PredictionStateValue {
	var (
		wrongIfValue    = FALSE
		annulledIfValue = FALSE
		predictValue    = p.Predict.Evaluate()
	)
	if p.WrongIf != nil {
		wrongIfValue = p.WrongIf.Evaluate()
	}
	if wrongIfValue == TRUE {
		return INCORRECT
	}
	if p.AnnulledIf != nil {
		annulledIfValue = p.AnnulledIf.Evaluate()
	}
	if annulledIfValue == TRUE {
		return ANNULLED
	}
	if wrongIfValue == UNDECIDED || annulledIfValue == UNDECIDED || predictValue == UNDECIDED {
		return ONGOING_PREDICTION
	}
	if predictValue == FALSE {
		return INCORRECT
	}
	return CORRECT
}

type PredictionState struct {
	Status ConditionStatus
	LastTs int
	Value  PredictionStateValue
	// add state to provide evidence of alleged condition result
}
