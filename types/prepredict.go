package types

type PrePredict struct {
	WrongIf    *BoolExpr
	AnnulledIf *BoolExpr
	PredictIf  *BoolExpr
}

func (p PrePredict) UndecidedConditions() []*Condition {
	conds := []*Condition{}
	conds = append(conds, p.WrongIf.UndecidedConditions()...)
	conds = append(conds, p.AnnulledIf.UndecidedConditions()...)
	conds = append(conds, p.PredictIf.UndecidedConditions()...)
	return conds
}

func (p PrePredict) Evaluate() PredictionStateValue {
	if p.WrongIf == nil && p.AnnulledIf == nil && p.PredictIf == nil {
		return ONGOING_PREDICTION
	}
	var (
		wrongIfValue    = FALSE
		annulledIfValue = FALSE
		predictIfValue  = TRUE
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
	if p.PredictIf != nil {
		predictIfValue = p.PredictIf.Evaluate()
	}
	if annulledIfValue == FALSE && (predictIfValue == FALSE || wrongIfValue == TRUE) {
		return INCORRECT
	}
	if wrongIfValue == UNDECIDED || annulledIfValue == UNDECIDED || predictIfValue == UNDECIDED {
		return ONGOING_PRE_PREDICTION
	}
	return ONGOING_PREDICTION
}

func (p *PrePredict) ClearState() {
	if p.AnnulledIf != nil {
		p.AnnulledIf.ClearState()
	}
	if p.WrongIf != nil {
		p.WrongIf.ClearState()
	}
	if p.PredictIf != nil {
		p.PredictIf.ClearState()
	}
}
