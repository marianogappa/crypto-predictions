package types

import (
	"errors"
	"fmt"
)

type ConditionStateValue int

const (
	UNDECIDED ConditionStateValue = iota
	TRUE
	FALSE
)

var (
	ErrUnknownConditionStateValue = errors.New("invalid state: unknown condition state value")
)

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

func (a ConditionStateValue) And(b ConditionStateValue) ConditionStateValue {
	if a == FALSE || b == FALSE {
		return FALSE
	}
	if a == TRUE && b == TRUE {
		return TRUE
	}
	return UNDECIDED
}

func (a ConditionStateValue) Or(b ConditionStateValue) ConditionStateValue {
	if a == TRUE || b == TRUE {
		return TRUE
	}
	if a == FALSE && b == FALSE {
		return FALSE
	}
	return UNDECIDED
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
