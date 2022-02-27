package types

type BoolExpr struct {
	Operator BoolOperator
	Operands []*BoolExpr
	Literal  *Condition
}

func (p *BoolExpr) UndecidedConditions() []*Condition {
	conds := []*Condition{}
	if p == nil || (p.Operator == LITERAL && p.Literal == nil) {
		return conds
	}
	switch p.Operator {
	case AND, OR, NOT:
		for _, operand := range p.Operands {
			conds = append(conds, operand.UndecidedConditions()...)
		}
	default:
		if (*p).Literal.Evaluate() == UNDECIDED {
			conds = append(conds, p.Literal)
		}
	}
	return conds
}

func (p *BoolExpr) Evaluate() ConditionStateValue {
	if p == nil {
		return TRUE
	}
	switch p.Operator {
	case AND:
		if len(p.Operands) == 0 {
			return TRUE
		}
		result := p.Operands[0].Evaluate()
		for _, operand := range p.Operands[1:] {
			result = result.And(operand.Evaluate())
		}
		return result
	case OR:
		if len(p.Operands) == 0 {
			return TRUE
		}
		result := p.Operands[0].Evaluate()
		for _, operand := range p.Operands[1:] {
			result = result.Or(operand.Evaluate())
		}
		return result
	case NOT:
		return p.Operands[0].Evaluate().Not()
	default:
		if p.Literal == nil {
			return TRUE
		}
		return p.Literal.Evaluate()
	}
}

func (p *BoolExpr) ClearState() {
	if p.Literal != nil {
		p.Literal.ClearState()
	}
	for _, be := range p.Operands {
		if be != nil {
			be.ClearState()
		}
	}
}
