package core

// BoolExpr represents a boolean expression in a prediction step.
type BoolExpr struct {
	Operator BoolOperator
	Operands []*BoolExpr
	Literal  *Condition
}

// UndecidedConditions calculates a list of conditions that haven't reached a final value.
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

// Evaluate returns the evaluated value of a boolean expression: one of TRUE|FALSE|UNDECIDED.
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
			result = result.and(operand.Evaluate())
		}
		return result
	case OR:
		if len(p.Operands) == 0 {
			return TRUE
		}
		result := p.Operands[0].Evaluate()
		for _, operand := range p.Operands[1:] {
			result = result.or(operand.Evaluate())
		}
		return result
	case NOT:
		return p.Operands[0].Evaluate().not()
	default:
		if p.Literal == nil {
			return TRUE
		}
		return p.Literal.Evaluate()
	}
}

// ClearState removes all state arising from evolving the boolean expression.
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
