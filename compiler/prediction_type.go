package compiler

import "github.com/marianogappa/predictions/core"

// CalculatePredictionType infers the prediction type by looking at its structure.
func CalculatePredictionType(pred core.Prediction) core.PredictionType {
	for predictionType, is := range predictionTypes {
		if is(pred) {
			return predictionType
		}
	}
	return core.PredictionTypeUnsupported
}

var (
	predictionTypes = map[core.PredictionType]func(core.Prediction) bool{
		core.PredictionTypeCoinOperatorFloatDeadline: func(pred core.Prediction) bool {
			return pred.PrePredict.Predict == nil && pred.PrePredict.AnnulledIf == nil &&
				pred.PrePredict.WrongIf == nil && pred.Predict.AnnulledIf == nil && pred.Predict.WrongIf == nil &&
				pred.Predict.Predict.Operator == core.LITERAL && len(pred.Predict.Predict.Literal.Operands) == 2 &&
				pred.Predict.Predict.Literal.Operands[0].Type == core.COIN &&
				pred.Predict.Predict.Literal.Operands[1].Type == core.NUMBER
		},
		core.PredictionTypeCoinWillReachInvalidatedIfItReaches: func(pred core.Prediction) bool {
			return pred.PrePredict.Predict == nil && pred.PrePredict.AnnulledIf == nil &&
				pred.PrePredict.WrongIf == nil && pred.Predict.AnnulledIf != nil && pred.Predict.WrongIf == nil &&

				// Predict section
				pred.Predict.Predict.Operator == core.LITERAL && len(pred.Predict.Predict.Literal.Operands) == 2 &&
				pred.Predict.Predict.Literal.Operands[0].Type == core.COIN &&
				pred.Predict.Predict.Literal.Operands[1].Type == core.NUMBER &&

				// AnnulledIf section
				pred.Predict.AnnulledIf.Operator == core.LITERAL && len(pred.Predict.AnnulledIf.Literal.Operands) == 2 &&
				pred.Predict.AnnulledIf.Literal.Operands[0].Type == core.COIN &&
				pred.Predict.AnnulledIf.Literal.Operands[1].Type == core.NUMBER &&

				// Operators are opposite
				operatorsAreOpposite(pred.Predict.Predict.Literal.Operator, pred.Predict.AnnulledIf.Literal.Operator)
		},
		core.PredictionTypeCoinWillRange: func(pred core.Prediction) bool {
			return pred.PrePredict.Predict == nil && pred.PrePredict.AnnulledIf == nil &&
				pred.PrePredict.WrongIf == nil && pred.Predict.AnnulledIf == nil && pred.Predict.WrongIf == nil &&
				pred.Predict.Predict.Operator == core.LITERAL && len(pred.Predict.Predict.Literal.Operands) == 3 &&
				pred.Predict.Predict.Literal.Operator == "BETWEEN" &&
				pred.Predict.Predict.Literal.Operands[0].Type == core.COIN &&
				pred.Predict.Predict.Literal.Operands[1].Type == core.NUMBER &&
				pred.Predict.Predict.Literal.Operands[2].Type == core.NUMBER
		},
		core.PredictionTypeCoinWillReachBeforeItReaches: func(pred core.Prediction) bool {
			return pred.PrePredict.Predict == nil &&
				pred.PrePredict.AnnulledIf == nil &&
				pred.PrePredict.WrongIf == nil &&
				pred.Predict.AnnulledIf == nil &&
				pred.Predict.WrongIf == nil &&
				pred.Predict.Predict.Operator == core.AND &&
				len(pred.Predict.Predict.Operands) == 2 &&
				pred.Predict.Predict.Operands[0].Operator == core.LITERAL &&
				len(pred.Predict.Predict.Operands[0].Literal.Operands) == 2 &&
				pred.Predict.Predict.Operands[0].Literal.Operands[0].Type == core.COIN &&
				pred.Predict.Predict.Operands[0].Literal.Operands[1].Type == core.NUMBER &&
				pred.Predict.Predict.Operands[1].Operator == core.NOT &&
				len(pred.Predict.Predict.Operands[1].Operands) == 1 &&
				pred.Predict.Predict.Operands[1].Operands[0].Operator == core.LITERAL &&
				len(pred.Predict.Predict.Operands[1].Operands[0].Literal.Operands) == 2 &&
				pred.Predict.Predict.Operands[1].Operands[0].Literal.Operands[0].Type == core.COIN &&
				pred.Predict.Predict.Operands[1].Operands[0].Literal.Operands[1].Type == core.NUMBER &&
				pred.Predict.Predict.Operands[0].Literal.Operands[0] == pred.Predict.Predict.Operands[1].Operands[0].Literal.Operands[0] &&
				pred.Predict.Predict.Operands[0].Literal.Operator != pred.Predict.Predict.Operands[1].Operands[0].Literal.Operator
		},
	}
)

func operatorsAreOpposite(op1, op2 string) bool {
	return (op1 == ">=" && op2 == "<=") || (op1 == "<=" && op2 == ">=") || (op1 == ">" && op2 == "<") || (op1 == "<" && op2 == ">")
}
