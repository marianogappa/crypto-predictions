package types

import (
	"log"

	"github.com/marianogappa/signal-checker/common"
)

type Prediction struct {
	Version      string
	CreatedAt    common.ISO8601
	AuthorHandle string
	Post         string
	Define       map[string]*Condition
	PrePredict   PrePredict
	Predict      Predict
	State        PredictionState
}

func (p *Prediction) Evaluate() PredictionStateValue {
	value := p.calculateValue()
	p.State.Value = value
	switch p.State.Value {
	case ONGOING_PRE_PREDICTION, ONGOING_PREDICTION:
		p.State.Status = STARTED
	case CORRECT, INCORRECT, ANNULLED:
		p.State.Status = FINISHED
	}
	return value
}

func (p Prediction) calculateValue() PredictionStateValue {
	prePredictValue := p.PrePredict.Evaluate()
	log.Printf("Prediction.calculateValue: for %v, prePredictValue = %s\n", p.Post, prePredictValue)
	if prePredictValue == ONGOING_PRE_PREDICTION || prePredictValue == INCORRECT {
		return prePredictValue
	}
	predictValue := p.Predict.Evaluate()
	log.Printf("Prediction.calculateValue: for %v, predictValue = %s\n", p.Post, predictValue)
	return predictValue
}
