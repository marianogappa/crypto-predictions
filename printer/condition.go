package printer

import (
	"fmt"
	"strings"
	"time"

	"github.com/marianogappa/predictions/types"
)

func printCondition(c types.Condition) string {
	fromTs := formatTs(c.FromTs)
	toTs := formatTs(c.ToTs)
	suffix := ""
	if len(c.Assumed) > 0 {
		suffix = fmt.Sprintf("(%v assumed from prediction text)", strings.Join(c.Assumed, ", "))
	}
	if c.Operator == "BETWEEN" {
		return fmt.Sprintf("%v BETWEEN %v AND %v from %v to %v %v", c.Operands[0].Str, c.Operands[1].Str, c.Operands[2].Str, fromTs, toTs, suffix)
	}
	return fmt.Sprintf("%v %v %v from %v to %v %v", c.Operands[0].Str, c.Operator, c.Operands[1].Str, fromTs, toTs, suffix)
}

func formatTs(ts int) string {
	return time.Unix(int64(ts), 0).Format(time.RFC3339)
}
