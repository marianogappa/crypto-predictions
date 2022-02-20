package types

import (
	"github.com/marianogappa/signal-checker/common"
)

type Prediction struct {
	Version    string
	CreatedAt  common.ISO8601
	PostAuthor string
	PostText   string
	PostedAt   common.ISO8601
	PostUrl    string
	Given      map[string]*Condition
	PrePredict PrePredict
	Predict    Predict
	State      PredictionState
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
	if prePredictValue == ONGOING_PRE_PREDICTION || prePredictValue == INCORRECT {
		return prePredictValue
	}
	predictValue := p.Predict.Evaluate()
	return predictValue
}
