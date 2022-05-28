package printer

import (
	"fmt"
	"time"

	"github.com/marianogappa/predictions/types"
)

type PredictionPrettyPrinter struct {
	prediction types.Prediction
}

func NewPredictionPrettyPrinter(p types.Prediction) PredictionPrettyPrinter {
	return PredictionPrettyPrinter{prediction: p}
}

func (p PredictionPrettyPrinter) Default() string {
	switch p.prediction.Type {
	case types.PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE:
		// TODO: this if is necessary because there's a bug where some predictions are miscategorised
		if p.prediction.Predict.Predict.Literal != nil && len(p.prediction.Predict.Predict.Literal.Operands) == 2 {
			return p.predictionTypeCoinOperatorFloatDeadline()
		}
	// case types.PREDICTION_TYPE_COIN_WILL_RANGE:
	// 	p.predictionTypeCoinWillRange()
	case types.PREDICTION_TYPE_COIN_WILL_REACH_BEFORE_IT_REACHES:
		p.predictionTypeCoinWillReachBeforeItReaches()
	case types.PREDICTION_TYPE_THE_FLIPPENING:
		p.predictionTypeTheFlippening()
	}

	if p.prediction.PrePredict.Predict != nil {
		return fmt.Sprintf("%v predicts that, given that %v, then %v", p.prediction.PostAuthor, printPrePredict(p.prediction.PrePredict), printPredict(p.prediction.Predict))
	}
	return fmt.Sprintf("%v predicts that %v", p.prediction.PostAuthor, printPredict(p.prediction.Predict))
}

func (p PredictionPrettyPrinter) predictionTypeCoinOperatorFloatDeadline() string {
	cond := p.prediction.Predict.Predict.Literal
	coin := parseOperand(cond.Operands[0])
	number := parseOperand(cond.Operands[1])

	temporalPart := fmt.Sprintf("by %v", cond.ToTs)
	if cond.ToDuration != "" {
		temporalPart = parseDuration(cond.ToDuration, time.Unix(int64(cond.FromTs), 0))
	}

	operator := ""
	switch cond.Operator {
	case ">":
		operator = "will hit"
	case ">=":
		operator = "will exceed"
	case "<":
		operator = "will fall to"
	case "<=":
		operator = "will go below"
	}

	return fmt.Sprintf("%v %v %v %v", coin, operator, number, temporalPart)
}

// func (p PredictionPrettyPrinter) predictionTypeCoinWillRange() string {
// 	// TODO
// 	return ""
// }

func (p PredictionPrettyPrinter) predictionTypeCoinWillReachBeforeItReaches() string {
	coin := parseOperand(p.prediction.Predict.Predict.Operands[0].Literal.Operands[0])
	willReach := parseNumber(p.prediction.Predict.Predict.Operands[0].Literal.Operands[1].Number)
	beforeIfReaches := parseNumber(p.prediction.Predict.Predict.Operands[1].Operands[0].Literal.Operands[1].Number)
	return fmt.Sprintf("%v will reach %v before it reaches %v", coin, willReach, beforeIfReaches)
}

func (p PredictionPrettyPrinter) predictionTypeTheFlippening() string {
	coin1 := parseOperand(p.prediction.Predict.Predict.Literal.Operands[0])
	coin2 := parseOperand(p.prediction.Predict.Predict.Literal.Operands[1])

	cond := p.prediction.Predict.Predict.Literal
	temporalPart := fmt.Sprintf("by %v", cond.ToTs)
	if cond.ToDuration != "" {
		temporalPart = parseDuration(cond.ToDuration, time.Unix(int64(cond.FromTs), 0))
	}
	return fmt.Sprintf("%v will flip %v (in marketcap) %v", coin1, coin2, temporalPart)
}
