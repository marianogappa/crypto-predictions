package printer

import (
	"fmt"

	"github.com/marianogappa/predictions/types"
)

type PredictionPrettyPrinter struct {
	prediction types.Prediction
}

func NewPredictionPrettyPrinter(p types.Prediction) PredictionPrettyPrinter {
	return PredictionPrettyPrinter{prediction: p}
}

func (p PredictionPrettyPrinter) Default() string {
	if p.prediction.PrePredict.PredictIf != nil {
		return fmt.Sprintf("%v predicts that, given that %v, then %v", p.prediction.AuthorHandle, printPrePredict(p.prediction.PrePredict), printPredict(p.prediction.Predict))
	}
	return fmt.Sprintf("%v predicts that %v", p.prediction.AuthorHandle, printPredict(p.prediction.Predict))
}
