package printer

import (
	"fmt"

	"github.com/marianogappa/predictions/types"
)

func printPredict(p types.Predict) string {
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

func printPrePredict(p types.PrePredict) string {
	if p.PredictIf == nil {
		return ""
	}
	if p.WrongIf != nil && p.AnnulledIf != nil {
		return fmt.Sprintf("%v, being wrong if %v, unless %v in which case all bets are off", printBoolExpr(p.PredictIf, 0), printBoolExpr(p.WrongIf, 0), printBoolExpr(p.AnnulledIf, 0))
	}
	if p.WrongIf != nil {
		return fmt.Sprintf("%v, being wrong if %v", printBoolExpr(p.PredictIf, 0), printBoolExpr(p.WrongIf, 0))
	}
	if p.AnnulledIf != nil {
		return fmt.Sprintf("%v, unless %v in which case all bets are off", printBoolExpr(p.PredictIf, 0), printBoolExpr(p.AnnulledIf, 0))
	}
	// Unreachable
	return "???"
}
