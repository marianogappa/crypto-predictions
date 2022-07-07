package core

// Prediction is the struct that represents a prediction after it being compiled.
type Prediction struct {
	UUID          string
	Version       string
	CreatedAt     ISO8601
	PostAuthor    string
	PostAuthorURL string
	PostText      string
	PostedAt      ISO8601
	PostURL       string
	Given         map[string]*Condition
	PrePredict    PrePredict
	Predict       Predict
	State         PredictionState
	Reporter      string
	Type          PredictionType

	// These fields are stored outside of the DB blob field and are filled by StateStorage. They are not read by
	// compilation nor serialised.
	Deleted bool
	Hidden  bool
	Paused  bool
}

// Evaluate is a stateful method that evaluates & stores the value of this prediction, also potentially changing its
// status.
func (p *Prediction) Evaluate() PredictionStateValue {
	// TODO: only calculate if not in final state? Why not?
	value := p.calculateValue()
	p.State.Value = value
	switch p.State.Value {
	case ONGOINGPREPREDICTION, ONGOINGPREDICTION:
		p.State.Status = STARTED
	case CORRECT, INCORRECT, ANNULLED:
		p.State.Status = FINISHED
	}
	for _, cond := range p.Given {
		if cond.State.LastTs > p.State.LastTs {
			p.State.LastTs = cond.State.LastTs
		}
	}
	return value
}

func (p Prediction) calculateValue() PredictionStateValue {
	// TODO: only calculate if not in final state? Why not?
	prePredictValue := p.PrePredict.Evaluate()
	if prePredictValue == ONGOINGPREPREDICTION || prePredictValue == INCORRECT || prePredictValue == ANNULLED {
		return prePredictValue
	}
	predictValue := p.Predict.Evaluate()
	return predictValue
}

// ClearState removes all state from a Prediction, effectively making it able to "start evolving from scratch".
func (p *Prediction) ClearState() {
	p.State = PredictionState{}
	p.PrePredict.ClearState()
	p.Predict.ClearState()
}

// UndecidedConditions are all conditions in this Prediction that are not TRUE/FALSE yet.
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
	case ONGOINGPREPREDICTION:
		return p.PrePredict.UndecidedConditions()
	case ONGOINGPREDICTION:
		return p.Predict.UndecidedConditions()
	}
	return []*Condition{}
}

// CalculateTags (for now) lists the set of all Operands in this Prediction.
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

// CalculateMainCoin returns the main Operand of this Prediction.
func (p *Prediction) CalculateMainCoin() Operand {
	switch p.Type {
	case PredictionTypeCoinOperatorFloatDeadline, PredictionTypeCoinWillReachBeforeItReaches, PredictionTypeCoinWillRange:
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
