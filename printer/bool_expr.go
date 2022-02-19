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
	switch e.Operator {
	case types.AND, types.OR:
		operands := []string{}
		for _, operand := range e.Operands {
			s := printBoolExpr(operand, nestLevel+1)
			if s == "" {
				continue
			}
			operands = append(operands, s)
		}
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
		var prefix, postfix string
		if nestLevel > 0 {
			prefix = "("
			postfix = ")"
		}

		return fmt.Sprintf("%v%v%v", prefix, strings.Join(operands, connector), postfix)
	default:
		return printCondition(*e.Literal)
	}
}
