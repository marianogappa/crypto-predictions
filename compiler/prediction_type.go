package compiler

import (
	"github.com/marianogappa/predictions/types"
)

// CalculatePredictionType infers the prediction type by looking at its structure.
func CalculatePredictionType(pred types.Prediction) types.PredictionType {
	for predictionType, is := range predictionTypes {
		if is(pred) {
			return predictionType
		}
	}
	return types.PREDICTION_TYPE_UNSUPPORTED
}

var (
	predictionTypes = map[types.PredictionType]func(types.Prediction) bool{
		types.PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE: func(pred types.Prediction) bool {
			return pred.PrePredict.Predict == nil && pred.PrePredict.AnnulledIf == nil &&
				pred.PrePredict.WrongIf == nil && pred.Predict.AnnulledIf == nil && pred.Predict.WrongIf == nil &&
				pred.Predict.Predict.Operator == types.LITERAL && len(pred.Predict.Predict.Literal.Operands) == 2 &&
				pred.Predict.Predict.Literal.Operands[0].Type == types.COIN &&
				pred.Predict.Predict.Literal.Operands[1].Type == types.NUMBER
		},
		types.PREDICTION_TYPE_COIN_WILL_RANGE: func(pred types.Prediction) bool {
			return pred.PrePredict.Predict == nil && pred.PrePredict.AnnulledIf == nil &&
				pred.PrePredict.WrongIf == nil && pred.Predict.AnnulledIf == nil && pred.Predict.WrongIf == nil &&
				pred.Predict.Predict.Operator == types.LITERAL && len(pred.Predict.Predict.Literal.Operands) == 3 &&
				pred.Predict.Predict.Literal.Operator == "BETWEEN" &&
				pred.Predict.Predict.Literal.Operands[0].Type == types.COIN &&
				pred.Predict.Predict.Literal.Operands[1].Type == types.NUMBER &&
				pred.Predict.Predict.Literal.Operands[2].Type == types.NUMBER
		},
		types.PREDICTION_TYPE_COIN_WILL_REACH_BEFORE_IT_REACHES: func(pred types.Prediction) bool {
			return pred.PrePredict.Predict == nil &&
				pred.PrePredict.AnnulledIf == nil &&
				pred.PrePredict.WrongIf == nil &&
				pred.Predict.AnnulledIf == nil &&
				pred.Predict.WrongIf == nil &&
				pred.Predict.Predict.Operator == types.AND &&
				len(pred.Predict.Predict.Operands) == 2 &&
				pred.Predict.Predict.Operands[0].Operator == types.LITERAL &&
				len(pred.Predict.Predict.Operands[0].Literal.Operands) == 2 &&
				pred.Predict.Predict.Operands[0].Literal.Operands[0].Type == types.COIN &&
				pred.Predict.Predict.Operands[0].Literal.Operands[1].Type == types.NUMBER &&
				pred.Predict.Predict.Operands[1].Operator == types.NOT &&
				len(pred.Predict.Predict.Operands[1].Operands) == 1 &&
				pred.Predict.Predict.Operands[1].Operands[0].Operator == types.LITERAL &&
				len(pred.Predict.Predict.Operands[1].Operands[0].Literal.Operands) == 2 &&
				pred.Predict.Predict.Operands[1].Operands[0].Literal.Operands[0].Type == types.COIN &&
				pred.Predict.Predict.Operands[1].Operands[0].Literal.Operands[1].Type == types.NUMBER &&
				pred.Predict.Predict.Operands[0].Literal.Operands[0] == pred.Predict.Predict.Operands[1].Operands[0].Literal.Operands[0] &&
				pred.Predict.Predict.Operands[0].Literal.Operator != pred.Predict.Predict.Operands[1].Operands[0].Literal.Operator
		},
		types.PREDICTION_TYPE_THE_FLIPPENING: func(pred types.Prediction) bool {
			return pred.PrePredict.Predict == nil && pred.PrePredict.AnnulledIf == nil &&
				pred.PrePredict.WrongIf == nil && pred.Predict.AnnulledIf == nil && pred.Predict.WrongIf == nil &&
				pred.Predict.Predict.Operator == types.LITERAL && len(pred.Predict.Predict.Literal.Operands) == 2 &&
				pred.Predict.Predict.Literal.Operands[0].Type == types.MARKETCAP &&
				pred.Predict.Predict.Literal.Operands[1].Type == types.MARKETCAP &&
				pred.Predict.Predict.Literal.Operator == ">"
		},
	}
)
