package printer

import (
	"fmt"
	"time"

	"github.com/marianogappa/predictions/types"
)

// PredictionPrettyPrinter builds a human-readable description of a prediction, e.g.:
// "Bitcoin will be below $17k in 2 weeks"
type PredictionPrettyPrinter struct {
	prediction types.Prediction
}

// NewPredictionPrettyPrinter constructs a PredictionPrettyPrinter.
func NewPredictionPrettyPrinter(p types.Prediction) PredictionPrettyPrinter {
	return PredictionPrettyPrinter{prediction: p}
}

// String returns a human-readable description of a prediction, e.g.:
// "Bitcoin will be below $17k in 2 weeks"
func (p PredictionPrettyPrinter) String() string {
	switch p.prediction.Type {
	case types.PredictionTypeCoinOperatorFloatDeadline:
		// TODO: this if is necessary because there's a bug where some predictions are miscategorised
		if p.prediction.Predict.Predict.Literal != nil && len(p.prediction.Predict.Predict.Literal.Operands) == 2 {
			return p.predictionTypeCoinOperatorFloatDeadline()
		}
	// case types.PREDICTION_TYPE_COIN_WILL_RANGE:
	// 	p.predictionTypeCoinWillRange()
	case types.PredictionTypeCoinWillReachBeforeItReaches:
		p.predictionTypeCoinWillReachBeforeItReaches()
	}

	if p.prediction.PrePredict.Predict != nil {
		return fmt.Sprintf("%v predicts that, given that %v, then %v", p.prediction.PostAuthor, printPrePredict(p.prediction.PrePredict), printPredict(p.prediction.Predict))
	}
	return fmt.Sprintf("%v predicts that %v", p.prediction.PostAuthor, printPredict(p.prediction.Predict))
}

func (p PredictionPrettyPrinter) predictionTypeCoinOperatorFloatDeadline() string {
	cond := p.prediction.Predict.Predict.Literal
	coin, useDollarSign := parseOperand(cond.Operands[0], false)
	number, _ := parseOperand(cond.Operands[1], useDollarSign)

	humanToTs := time.Unix(int64(cond.ToTs), 0).Format("Jan 2, 2006")
	temporalPart := fmt.Sprintf("by %v", humanToTs)
	if cond.ToDuration != "" {
		temporalPart = parseDuration(cond.ToDuration, time.Unix(int64(cond.FromTs), 0))
	}

	operator := ""
	switch cond.Operator {
	case ">", ">=":
		operator = "will exceed"
	case "<", "<=":
		operator = "will be below"
	}

	return fmt.Sprintf("%v %v %v %v", coin, operator, number, temporalPart)
}

// func (p PredictionPrettyPrinter) predictionTypeCoinWillRange() string {
// 	// TODO
// 	return ""
// }

func (p PredictionPrettyPrinter) predictionTypeCoinWillReachBeforeItReaches() string {
	coin := legacyParseOperand(p.prediction.Predict.Predict.Operands[0].Literal.Operands[0])
	willReach := parseNumber(p.prediction.Predict.Predict.Operands[0].Literal.Operands[1].Number, false)
	beforeIfReaches := parseNumber(p.prediction.Predict.Predict.Operands[1].Operands[0].Literal.Operands[1].Number, false)
	return fmt.Sprintf("%v will reach %v before it reaches %v", coin, willReach, beforeIfReaches)
}

func (p PredictionPrettyPrinter) predictionTypeTheFlippening() string {
	coin1 := legacyParseOperand(p.prediction.Predict.Predict.Literal.Operands[0])
	coin2 := legacyParseOperand(p.prediction.Predict.Predict.Literal.Operands[1])

	cond := p.prediction.Predict.Predict.Literal
	temporalPart := fmt.Sprintf("by %v", cond.ToTs)
	if cond.ToDuration != "" {
		temporalPart = parseDuration(cond.ToDuration, time.Unix(int64(cond.FromTs), 0))
	}
	return fmt.Sprintf("%v will flip %v (in marketcap) %v", coin1, coin2, temporalPart)
}
