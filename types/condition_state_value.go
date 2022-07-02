package types

import (
	"errors"
	"fmt"
)

// ConditionStateValue is the value of a condition: one of TRUE|FALSE|UNDECIDED.
type ConditionStateValue int

const (
	// UNDECIDED is one of the possible values of a condition.
	UNDECIDED ConditionStateValue = iota
	// TRUE is one of the possible values of a condition.
	TRUE
	// FALSE is one of the possible values of a condition.
	FALSE
)

var (
	// ErrUnknownConditionStateValue means: invalid state: unknown condition state value
	ErrUnknownConditionStateValue = errors.New("invalid state: unknown condition state value")
)

// ConditionStateValueFromString constructs a ConditionStateValue from a string.
func ConditionStateValueFromString(s string) (ConditionStateValue, error) {
	switch s {
	case "UNDECIDED", "":
		return UNDECIDED, nil
	case "TRUE":
		return TRUE, nil
	case "FALSE":
		return FALSE, nil
	default:
		return 0, fmt.Errorf("%w: %v", ErrUnknownConditionStateValue, s)
	}
}
func (a ConditionStateValue) String() string {
	switch a {
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

func (a ConditionStateValue) and(b ConditionStateValue) ConditionStateValue {
	if a == FALSE || b == FALSE {
		return FALSE
	}
	if a == TRUE && b == TRUE {
		return TRUE
	}
	return UNDECIDED
}

func (a ConditionStateValue) or(b ConditionStateValue) ConditionStateValue {
	if a == TRUE || b == TRUE {
		return TRUE
	}
	if a == FALSE && b == FALSE {
		return FALSE
	}
	return UNDECIDED
}

func (a ConditionStateValue) not() ConditionStateValue {
	switch a {
	case UNDECIDED:
		return UNDECIDED
	case TRUE:
		return FALSE
	default: // FALSE
		return TRUE
	}
}
