package printer

import (
	"fmt"

	"github.com/marianogappa/predictions/core"
)

func printPredict(p core.Predict) string {
	if p.WrongIf != nil && p.AnnulledIf != nil {
		return fmt.Sprintf("%v, being wrong if %v, unless %v in which case all bets are off", printBoolExpr(&p.Predict, 0), printBoolExpr(p.WrongIf, 0), printBoolExpr(p.AnnulledIf, 0))
	}
	if p.WrongIf != nil {
		return fmt.Sprintf("%v, being wrong if %v", printBoolExpr(&p.Predict, 0), printBoolExpr(p.WrongIf, 0))
	}
	if p.AnnulledIf != nil {
		return fmt.Sprintf("%v, unless %v in which case all bets are off", printBoolExpr(&p.Predict, 0), printBoolExpr(p.AnnulledIf, 0))
	}
	return printBoolExpr(&p.Predict, 0)
}

func printPrePredict(p core.PrePredict) string {
	if p.Predict == nil {
		return ""
	}
	if p.WrongIf != nil && p.AnnulledIf != nil {
		return fmt.Sprintf("%v, being wrong if %v, unless %v in which case all bets are off", printBoolExpr(p.Predict, 0), printBoolExpr(p.WrongIf, 0), printBoolExpr(p.AnnulledIf, 0))
	}
	if p.WrongIf != nil {
		return fmt.Sprintf("%v, being wrong if %v", printBoolExpr(p.Predict, 0), printBoolExpr(p.WrongIf, 0))
	}
	if p.AnnulledIf != nil {
		return fmt.Sprintf("%v, unless %v in which case all bets are off", printBoolExpr(p.Predict, 0), printBoolExpr(p.AnnulledIf, 0))
	}
	// Unreachable
	return "???"
}
