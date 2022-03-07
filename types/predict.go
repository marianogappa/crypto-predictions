package types

type Predict struct {
	WrongIf                           *BoolExpr
	AnnulledIf                        *BoolExpr
	Predict                           BoolExpr
	IgnoreUndecidedIfPredictIsDefined bool
}

func (p Predict) UndecidedConditions() []*Condition {
	conds := []*Condition{}
	conds = append(conds, p.WrongIf.UndecidedConditions()...)
	conds = append(conds, p.AnnulledIf.UndecidedConditions()...)
	conds = append(conds, p.Predict.UndecidedConditions()...)
	return conds
}

func (p Predict) Evaluate() PredictionStateValue {
	var (
		wrongIfValue    = FALSE
		annulledIfValue = FALSE
		predictValue    = p.Predict.Evaluate()
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
	if annulledIfValue == FALSE && (predictValue == FALSE || wrongIfValue == TRUE) {
		return INCORRECT
	}
	if p.IgnoreUndecidedIfPredictIsDefined && predictValue != UNDECIDED {
		switch {
		case annulledIfValue == FALSE && predictValue == TRUE && wrongIfValue == TRUE:
			return INCORRECT
		case annulledIfValue == FALSE && predictValue == TRUE && wrongIfValue == FALSE:
			return CORRECT
		case annulledIfValue == FALSE && predictValue == TRUE && wrongIfValue == UNDECIDED:
			return CORRECT
		case annulledIfValue == FALSE && predictValue == FALSE && wrongIfValue == TRUE:
			return INCORRECT
		case annulledIfValue == FALSE && predictValue == FALSE && wrongIfValue == FALSE:
			return INCORRECT
		case annulledIfValue == FALSE && predictValue == FALSE && wrongIfValue == UNDECIDED:
			return INCORRECT
		case annulledIfValue == UNDECIDED && predictValue == TRUE && wrongIfValue == TRUE:
			return INCORRECT
		case annulledIfValue == UNDECIDED && predictValue == TRUE && wrongIfValue == FALSE:
			return CORRECT
		case annulledIfValue == UNDECIDED && predictValue == TRUE && wrongIfValue == UNDECIDED:
			return CORRECT
		case annulledIfValue == UNDECIDED && predictValue == FALSE && wrongIfValue == TRUE:
			return INCORRECT
		case annulledIfValue == UNDECIDED && predictValue == FALSE && wrongIfValue == FALSE:
			return INCORRECT
		case annulledIfValue == UNDECIDED && predictValue == FALSE && wrongIfValue == UNDECIDED:
			return INCORRECT
		}
	}
	if wrongIfValue == UNDECIDED || annulledIfValue == UNDECIDED || predictValue == UNDECIDED {
		return ONGOING_PREDICTION
	}
	return CORRECT
}

func (p *Predict) ClearState() {
	if p.AnnulledIf != nil {
		p.AnnulledIf.ClearState()
	}
	if p.WrongIf != nil {
		p.WrongIf.ClearState()
	}
	p.Predict.ClearState()
}
