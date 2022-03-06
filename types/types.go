package types

import (
	"errors"
	"fmt"

	"github.com/marianogappa/signal-checker/common"
)

var (
	ErrUnknownOperandType                 = errors.New("unknown value for operandType")
	ErrUnknownBoolOperator                = errors.New("unknown value for BoolOperator")
	ErrUnknownConditionStatus             = errors.New("invalid state: unknown value for ConditionStatus")
	ErrUnknownPredictionStateValue        = errors.New("invalid state: unknown value for PredictionStateValue")
	ErrInvalidOperand                     = errors.New("invalid operand")
	ErrEmptyQuoteAsset                    = errors.New("quote asset cannot be empty")
	ErrNonEmptyQuoteAssetOnNonCoin        = errors.New("quote asset must be empty for non-coin operand types")
	ErrEqualBaseQuoteAssets               = errors.New("base asset cannot be equal to quote asset")
	ErrInvalidDuration                    = errors.New("invalid duration")
	ErrInvalidFromISO8601                 = errors.New("invalid FromISO8601")
	ErrInvalidToISO8601                   = errors.New("invalid ToISO8601")
	ErrOneOfToISO8601ToDurationRequired   = errors.New("one of ToISO8601 or ToDuration is required")
	ErrInvalidConditionSyntax             = errors.New("invalid condition syntax")
	ErrUnknownConditionOperator           = errors.New("unknown condition operator: supported are [>|<|>=|<=|BETWEEN...AND]")
	ErrErrorMarginRatioAbove30            = errors.New("error margin ratio above 30%% is not allowed")
	ErrInvalidJSON                        = errors.New("invalid JSON")
	ErrEmptyPostURL                       = errors.New("postUrl cannot be empty")
	ErrEmptyPostAuthor                    = errors.New("postAuthor cannot be empty")
	ErrEmptyPostedAt                      = errors.New("postedAt cannot be empty")
	ErrInvalidPostedAt                    = errors.New("postedAt must be a valid ISO8601 timestamp")
	ErrMissingRequiredPrePredictPredictIf = errors.New("pre-predict clause must have predictIf if it has either wrongIf or annuledIf. Otherwise, add them directly on predict clause")
	ErrBoolExprSyntaxError                = errors.New("syntax error in bool expression")
	ErrPredictionFinishedAtStartTime      = errors.New("prediction is finished at start time")
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
		return 0, fmt.Errorf("%w: %v", ErrUnknownOperandType, s)
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
		return 0, fmt.Errorf("%w: %v", ErrUnknownBoolOperator, s)
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
		return 0, fmt.Errorf("%w: %v", ErrUnknownConditionStatus, s)
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
		return 0, fmt.Errorf("%w: %v", ErrUnknownPredictionStateValue, s)
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

func (v PredictionStateValue) IsFinal() bool {
	return v != ONGOING_PRE_PREDICTION && v != ONGOING_PREDICTION
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
