package types

import (
	"errors"
	"fmt"
)

type Condition struct {
	Name             string
	Operator         string
	Operands         []Operand
	FromTs           int // won't work for dynamic
	ToTs             int // won't work for dynamic
	ToDuration       string
	Assumed          []string
	State            ConditionState
	ErrorMarginRatio float64
}

var (
	conditionOpFuncs = map[string]func(ops []float64, errRatio float64) bool{
		">=": func(ops []float64, errRatio float64) bool { return ops[0] >= ops[1]*(1.0-errRatio) },
		"<=": func(ops []float64, errRatio float64) bool { return ops[0] <= ops[1]*(1.0+errRatio) },
		">":  func(ops []float64, errRatio float64) bool { return ops[0] > ops[1]*(1.0-errRatio) },
		"<":  func(ops []float64, errRatio float64) bool { return ops[0] < ops[1]*(1.0+errRatio) },
		"BETWEEN": func(ops []float64, errRatio float64) bool {
			return ops[0] >= ops[1]*(1.0-errRatio/2) && ops[0] <= ops[2]*(1.0+errRatio/2)
		},
	}
	errAtLeastOneTickRequired            = errors.New("internal error: at least one tick must be supplied")
	errInvalidTickSupplied               = errors.New("internal error: invalid tick supplied (empty timestamp)")
	errMismatchingTickTimestampsSupplied = errors.New("internal error: ticks with mismatching timestamps were supplied")
	errOlderTickTimestampSupplied        = errors.New("internal error: ticks with older timestamps than previous were supplied")
)

// Run evolves the state of a condition by analyzing the next tuple of ticks from its non-literal operands.
// In order to run, it depends on the caller (i.e. PredRunner) to call all associated TickIterators and supply the
// next ticks.
func (c *Condition) Run(ticks map[string]Tick) error {
	// State should not change after it's in a final status.
	if c.State.Status == FINISHED {
		return nil
	}

	// Conditions with no non-literal operands are not supported, because their value would be resolved already.
	if len(ticks) == 0 {
		return errAtLeastOneTickRequired
	}

	// Before evolving the state, the timestamp of this run must be resolved. There must be exactly one timestamp,
	// even if more than one non-literal operands exist. If there's two or more, they must match and be newer than or
	// equal to the previous state's tick. It can be equal, because candlesticks have an low/high tick, but their
	// timestamps match, and there's no way to know which one happened first (tick iterators must send low first).
	var timestamp int
	for _, tick := range ticks {
		// If there's no timestamp in the tick, it's invalid.
		if tick.Timestamp == 0 {
			return errInvalidTickSupplied
		}
		if timestamp == 0 {
			timestamp = tick.Timestamp
		} else if timestamp != tick.Timestamp {
			// If this is not the first analyzed tick, its timestamp must match the first one.
			return errMismatchingTickTimestampsSupplied
		}
	}

	// If the last timestamp in the state is newer than the current one, there's a problem with the supplied ticks!
	if timestamp < c.State.LastTs {
		return errOlderTickTimestampSupplied
	}

	// If the supplied ticks are older than when the condition begins, then ignore these ticks.
	if timestamp < c.FromTs {
		return nil
	}

	// Considering we already know this condition is not in a final state, and if the supplied ticks are newer than
	// the finish timestamp of this condition, then finish the condition with a FALSE value.
	if timestamp > c.ToTs {
		c.State.Status = FINISHED
		c.State.Value = FALSE
		return nil
	}

	// Resolve all operands to numbers. All non-literal operands must have associated ticks, or else this condition
	// cannot be evolved.
	operandValues := []float64{}
	for _, operand := range c.Operands {
		if operand.Type == NUMBER {
			operandValues = append(operandValues, float64(operand.Number))
		} else {
			tick, ok := ticks[operand.Str]
			if !ok {
				return fmt.Errorf("internal error: ticker for operand %v was not supplied", operand.Str)
			}
			operandValues = append(operandValues, float64(tick.Value))
		}
	}

	// Since no more errors are possible at this point, state is ready to be updated, and condition can officially be
	// started if it wasn't already. The point of having an "UNSTARTED" state is only to answer if this condition has
	// not been evaluated so far.
	c.State.LastTs = timestamp
	c.State.LastTicks = ticks
	c.State.Status = STARTED

	if opFunc, ok := conditionOpFuncs[c.Operator]; ok {
		// Finally, run the actual condition expression!
		if opFunc(operandValues, c.ErrorMarginRatio) {
			c.State.Status = FINISHED
			c.State.Value = TRUE
		} else if timestamp >= c.ToTs {
			// If we're evolving with the very last ticks, and considering it didn't evolve to TRUE, then it must
			// evolve to FALSE.
			c.State.Status = FINISHED
			c.State.Value = FALSE
		}
	}
	return nil
}

func (c Condition) Evaluate() ConditionStateValue {
	return c.State.Value
}

func (c *Condition) ClearState() {
	c.State = ConditionState{}
}
