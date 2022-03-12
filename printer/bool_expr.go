package printer

import (
	"fmt"
	"strings"

	"github.com/marianogappa/predictions/types"
)

func printBoolExpr(e *types.BoolExpr, nestLevel int) string {
	if e == nil {
		return ""
	}
	var prefix, postfix string
	if nestLevel > 0 {
		prefix = "("
		postfix = ")"
	}
	operands := []string{}
	for _, operand := range e.Operands {
		s := printBoolExpr(operand, nestLevel+1)
		if s == "" {
			continue
		}
		operands = append(operands, s)
	}
	switch e.Operator {
	case types.AND, types.OR:
		if len(operands) == 0 {
			return ""
		}
		if len(operands) == 1 {
			return operands[0]
		}

		connector := " and "
		if e.Operator == types.OR {
			connector = " or "
		}
		return fmt.Sprintf("%v%v%v", prefix, strings.Join(operands, connector), postfix)
	case types.NOT:
		return fmt.Sprintf("%vNOT %v%v", prefix, operands[0], postfix)
	default:
		return printCondition(*e.Literal, true)
	}
}
