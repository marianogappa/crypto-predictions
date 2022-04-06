package types

type Prediction struct {
	UUID          string
	Version       string
	CreatedAt     ISO8601
	PostAuthor    string
	PostAuthorURL string
	PostText      string
	PostedAt      ISO8601
	PostUrl       string
	Given         map[string]*Condition
	PrePredict    PrePredict
	Predict       Predict
	State         PredictionState
	Reporter      string
	Type          PredictionType
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

func (p *Prediction) UndecidedConditions() []*Condition {
	var undecidedConditions []*Condition
	undecidedConditions = append(undecidedConditions, p.PrePredict.UndecidedConditions()...)
	undecidedConditions = append(undecidedConditions, p.Predict.UndecidedConditions()...)
	return undecidedConditions
}

// ActionableUndecidedConditions are undecided conditions that should be evolved now, as opposed to conditions that
// are undecided, but need to wait for other conditions to be decided first.
func (p *Prediction) ActionableUndecidedConditions() []*Condition {
	switch p.Evaluate() {
	case ONGOING_PRE_PREDICTION:
		return p.PrePredict.UndecidedConditions()
	case ONGOING_PREDICTION:
		return p.Predict.UndecidedConditions()
	}
	return []*Condition{}
}

func (p *Prediction) CalculateTags() []string {
	tags := map[string]struct{}{}

	for _, cond := range p.Given {
		for _, operand := range cond.NonNumberOperands() {
			tags[operand.Str] = struct{}{}
		}
	}

	res := []string{}
	for tag := range tags {
		res = append(res, tag)
	}

	return res
}

func (p *Prediction) CalculateMainCoin() Operand {
	switch p.Type {
	case PREDICTION_TYPE_COIN_OPERATOR_FLOAT_DEADLINE, PREDICTION_TYPE_COIN_WILL_REACH_BEFORE_IT_REACHES, PREDICTION_TYPE_COIN_WILL_RANGE, PREDICTION_TYPE_THE_FLIPPENING:
		return p.Predict.Predict.Literal.Operands[0]
	default:
		// In unsupported cases, return the first available operand (Note: non-deterministic due to map).
		for _, cond := range p.Given {
			for _, op := range cond.NonNumberOperands() {
				return op
			}
		}
		return Operand{}
	}
}
