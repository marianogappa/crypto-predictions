package types

import (
	"github.com/marianogappa/signal-checker/common"
)

type Prediction struct {
	UUID       string
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
	Reporter   string
}

func (p *Prediction) Evaluate() PredictionStateValue {
	// TODO: only calculate if not in final state? Why not?
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
	// TODO: only calculate if not in final state? Why not?
	prePredictValue := p.PrePredict.Evaluate()
	if prePredictValue == ONGOING_PRE_PREDICTION || prePredictValue == INCORRECT || prePredictValue == ANNULLED {
		return prePredictValue
	}
	predictValue := p.Predict.Evaluate()
	return predictValue
}

func (p *Prediction) ClearState() {
	p.State = PredictionState{}
	p.PrePredict.ClearState()
	p.Predict.ClearState()
}
