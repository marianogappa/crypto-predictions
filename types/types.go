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
	Status    ConditionStatus
	LastTs    int
	LastTicks map[string]Tick
	Value     ConditionStateValue
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
	case "ANNULLED":
		return ANNULLED, nil
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
	case ANNULLED:
		return "ANNULLED"
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
	ONGOING_PRE_PREDICTION PredictionStateValue = iota
	ONGOING_PREDICTION
	CORRECT
	INCORRECT
	ANNULLED
)

type PredictionState struct {
	Status ConditionStatus
	LastTs int
	Value  PredictionStateValue
	// add state to provide evidence of alleged condition result
}
