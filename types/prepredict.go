package types

// PrePredict represents the initial, non-main prediction step.
type PrePredict struct {
	WrongIf                           *BoolExpr
	AnnulledIf                        *BoolExpr
	Predict                           *BoolExpr
	AnnulledIfPredictIsFalse          bool
	IgnoreUndecidedIfPredictIsDefined bool
}

// UndecidedConditions is the list of conditions within this Prediction step that haven't reached a final status.
func (p PrePredict) UndecidedConditions() []*Condition {
	conds := []*Condition{}
	conds = append(conds, p.WrongIf.UndecidedConditions()...)
	conds = append(conds, p.AnnulledIf.UndecidedConditions()...)
	conds = append(conds, p.Predict.UndecidedConditions()...)
	return conds
}

// Evaluate is a non-mutable method that evaluates this Prediction step's current value.
func (p PrePredict) Evaluate() PredictionStateValue {
	if p.WrongIf == nil && p.AnnulledIf == nil && p.Predict == nil {
		return ONGOINGPREDICTION
	}
	var (
		wrongIfValue    = FALSE
		annulledIfValue = FALSE
		predictValue    = TRUE
	)
	if p.AnnulledIf != nil {
		annulledIfValue = p.AnnulledIf.Evaluate()
	}
	if annulledIfValue == TRUE {
		return ANNULLED
	}
	if p.WrongIf != nil {
		wrongIfValue = p.WrongIf.Evaluate()
	}
	if p.Predict != nil {
		predictValue = p.Predict.Evaluate()
	}
	if p.AnnulledIfPredictIsFalse && p.IgnoreUndecidedIfPredictIsDefined && predictValue == FALSE && annulledIfValue == UNDECIDED && wrongIfValue == UNDECIDED {
		return ANNULLED
	}
	if p.AnnulledIfPredictIsFalse && predictValue == FALSE && wrongIfValue == FALSE {
		return ANNULLED
	}
	if p.IgnoreUndecidedIfPredictIsDefined && predictValue != UNDECIDED && wrongIfValue == TRUE {
		return INCORRECT
	}
	if p.IgnoreUndecidedIfPredictIsDefined && predictValue == TRUE && wrongIfValue == UNDECIDED {
		return ONGOINGPREDICTION
	}
	if p.IgnoreUndecidedIfPredictIsDefined && predictValue == FALSE {
		return INCORRECT
	}
	if annulledIfValue == FALSE && (predictValue == FALSE || wrongIfValue == TRUE) {
		return INCORRECT
	}
	if wrongIfValue == UNDECIDED || annulledIfValue == UNDECIDED || predictValue == UNDECIDED {
		return ONGOINGPREPREDICTION
	}
	return ONGOINGPREDICTION
}

// ClearState removes all state arising from evolving this prediction's step.
func (p *PrePredict) ClearState() {
	if p.AnnulledIf != nil {
		p.AnnulledIf.ClearState()
	}
	if p.WrongIf != nil {
		p.WrongIf.ClearState()
	}
	if p.Predict != nil {
		p.Predict.ClearState()
	}
}
