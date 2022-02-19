package types

import (
	"fmt"

	"github.com/marianogappa/signal-checker/common"
)

type Condition struct {
	Name     string
	Operator string
	Operands []Operand
	FromTs   int // won't work for dynamic
	ToTs     int // won't work for dynamic
	Assumed  []string
	State    ConditionState
}

func (c *Condition) Run(ticks map[string]*common.Tick) error {
	// log.Printf("Condition.RunTick: with name %v status is %v and value is %v, ticks are %v\n", c.Name, c.State.Status, c.State.Value, ticks)
	if c.State.Status == FINISHED {
		// log.Println("status == finished")
		return nil
	}
	var timestamp int
	for _, tick := range ticks {
		// N.B. we only need to check one timestamp, because they must match.
		timestamp = tick.Timestamp
		break
	}
	if timestamp < c.FromTs {
		// log.Println("timestamp < fromTs")
		return nil
	}
	if timestamp > c.ToTs {
		// log.Println("timestamp > toTs")
		c.State.Status = FINISHED
		c.State.Value = FALSE
		return nil
	}
	if c.State.Status == UNSTARTED {
		c.State.Status = STARTED
	}
	c.State.LastTs = timestamp

	operandValues := []float64{}
	for _, operand := range c.Operands {
		if operand.Type == NUMBER {
			operandValues = append(operandValues, float64(operand.Number))
		} else {
			tick, ok := ticks[operand.Str]
			if !ok {
				return fmt.Errorf("internal error: ticker for operand %v was not supplied", operand.Str)
			}
			operandValues = append(operandValues, float64(tick.Price))
		}
	}

	switch c.Operator {
	case "==":
		if operandValues[0] == operandValues[1] {
			c.State.Status = FINISHED
			c.State.Value = TRUE
		}
	case ">=":
		if operandValues[0] >= operandValues[1] {
			c.State.Status = FINISHED
			c.State.Value = TRUE
		}
	case "<=":
		if operandValues[0] <= operandValues[1] {
			c.State.Status = FINISHED
			c.State.Value = TRUE
		}
	case ">":
		if operandValues[0] > operandValues[1] {
			c.State.Status = FINISHED
			c.State.Value = TRUE
		}
	case "<":
		if operandValues[0] < operandValues[1] {
			c.State.Status = FINISHED
			c.State.Value = TRUE
		}
	case "BETWEEN":
		if operandValues[0] >= operandValues[1] && operandValues[0] <= operandValues[2] {
			c.State.Status = FINISHED
			c.State.Value = TRUE
		}
	}
	// log.Println("status", c.State.Status, "value", c.State.Value, "timestmap", t.Timestamp, "fromTs", c.FromTs, "toTs", c.ToTs)
	return nil
}

func (c Condition) Evaluate() ConditionStateValue {
	return c.State.Value
}
