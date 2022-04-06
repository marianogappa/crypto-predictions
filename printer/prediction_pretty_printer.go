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

	temporalPart := fmt.Sprintf("by %v ", cond.ToTs)
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
