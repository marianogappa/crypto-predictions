package types

import (
	"errors"
	"fmt"
	"time"
)

// Condition represents a boolean condition that looks like "BTC/USDT >= 45000 within 3 weeks".
//
// A Prediction is composed by one or more Conditions.
//
// Each Condition is assigned a free-form string name (i.e. a variable name) within the Given block,
// represented here by the Name property.
//
// Conditions have timestamps for the moment they should start to be evaluated, and the deadline for the condition to
// be met, represented by FromTs and ToTs. Often, when people make predictions, they expect them to become true after
// a period of time from the moment they make the prediction, in which case ToDuration represents that period, but
// ToTs still computes the timestamp as a result of adding the period to FromTs. Review compiler.parseDuration for an
// exhaustive definition of valid values of ToDuration.
//
// Often, people consider predictions that _almost_ became true to be true nonetheless. Because of this reason,
// ErrorMarginRatio allows a Condition to become CORRECT when the literal number in the boolean condition evaluated is
// not yet satisfying the condition, but it's ErrorMarginRatio% away from satisfying it.
//
// Conditions are evolved by calling Condition.Run and supplying market Ticks, which are the closing prices of market
// candlesticks at 1 minute intervals. The Condition's State contains the results of that evolution, that is, whether
// the Condition has finished evolving and reached a final state, what were the latest ticks supplied, etc.
//
// Conditions start as UNDECIDED and can only change as a result of invoking Condition.Run. Subsequent invocations of
// Condition.Run will make the Value become TRUE if the boolean condition becomes true, or FALSE if the supplied Tick's
// timestamp exceeds the Condition's ToTs without the boolean condition becoming true.
type Condition struct {
	Name             string
	Operator         string
	Operands         []Operand
	FromTs           int // TODO: won't work for dynamic
	ToTs             int // TODO: won't work for dynamic
	ToDuration       string
	Assumed          []string // unused for now
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
	received := time.Unix(int64(timestamp), 0).Format(time.RFC3339)
	last := time.Unix(int64(c.State.LastTs), 0).Format(time.RFC3339)
	if timestamp <= c.State.LastTs {
		return fmt.Errorf("%w: for cond %v received %v but last is %v", errOlderTickTimestampSupplied, c.Name, received, last)
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

// Evaluate is a non-mutating function that returns the Value of a Condition, that is, if the Condition has reached
// a final value (i.e. TRUE, FALSE) or not (i.e. UNDECIDED).
func (c Condition) Evaluate() ConditionStateValue {
	return c.State.Value
}

// ClearState is a mutating function that empties the State of a Condition, allowing it to start from scratch.
// It is meant to be used as a Back Office operation, when the admin identifies a problem with the evolving of a
// prediction and wants to restart it from scratch and let it evolve again.
func (c *Condition) ClearState() {
	c.State = ConditionState{}
}

// NonNumberOperands returns the slice of Operands of a Condition that are either COINs or MARKETCAPs, but not NUMBERs.
func (c *Condition) NonNumberOperands() []Operand {
	ops := []Operand{}
	for _, op := range c.Operands {
		if op.Type == NUMBER {
			continue
		}
		ops = append(ops, op)
	}
	return ops
}

// Clone returns a deep copy of Condition that does not share any memory with the original struct.
func (c Condition) Clone() Condition {
	clonedOperands := make([]Operand, len(c.Operands))
	copy(clonedOperands, c.Operands)

	return Condition{
		Name:             c.Name,
		Operator:         c.Operator,
		Operands:         clonedOperands,
		FromTs:           c.FromTs,
		ToTs:             c.ToTs,
		ToDuration:       c.ToDuration,
		Assumed:          c.Assumed,
		State:            c.State.Clone(),
		ErrorMarginRatio: c.ErrorMarginRatio,
	}
}
