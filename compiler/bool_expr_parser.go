package compiler

import (
	"fmt"

	"github.com/marianogappa/predictions/compiler/boolunmarshal"
	"github.com/marianogappa/predictions/core"
)

func parseBoolExpr(s string, def map[string]*core.Condition) (*core.BoolExpr, error) {
	n, err := boolunmarshal.NewExprParser(s).Parse()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", core.ErrBoolExprSyntaxError, err)
	}
	return nodeToBoolExpr(n, def)
}

func nodeToBoolExpr(n boolunmarshal.Node, def map[string]*core.Condition) (*core.BoolExpr, error) {
	switch n.TT {
	case boolunmarshal.UNKNOWN, boolunmarshal.EOF:
		return nil, fmt.Errorf("%w: attempted to parse an invalid/unresolved node to a bool expression", core.ErrBoolExprSyntaxError)
	case boolunmarshal.INDENTIFIER:
		if _, ok := def[n.Token]; !ok {
			return nil, fmt.Errorf("%w: unknown identifier '%v'...maybe you forgot to add it to the define clause", core.ErrBoolExprSyntaxError, n.Token)
		}
		return &core.BoolExpr{
			Operator: core.LITERAL,
			Literal:  def[n.Token],
		}, nil
	case boolunmarshal.AND:
		if len(n.Nodes) == 0 {
			return nil, fmt.Errorf("%w: AND clause with zero operands", core.ErrBoolExprSyntaxError)
		}
		e := core.BoolExpr{
			Operator: core.AND,
		}
		for _, node := range n.Nodes {
			operand, err := nodeToBoolExpr(node, def)
			if err != nil {
				return nil, err
			}
			e.Operands = append(e.Operands, operand)
		}
		return &e, nil
	case boolunmarshal.OR:
		if len(n.Nodes) == 0 {
			return nil, fmt.Errorf("%w: OR clause with zero operands", core.ErrBoolExprSyntaxError)
		}
		e := core.BoolExpr{
			Operator: core.OR,
		}
		for _, node := range n.Nodes {
			operand, err := nodeToBoolExpr(node, def)
			if err != nil {
				return nil, err
			}
			e.Operands = append(e.Operands, operand)
		}
		return &e, nil
	case boolunmarshal.NOT:
		if len(n.Nodes) != 1 {
			return nil, fmt.Errorf("%w: NOT clause must have exactly one operand", core.ErrBoolExprSyntaxError)
		}
		operand, err := nodeToBoolExpr(n.Nodes[0], def)
		if err != nil {
			return nil, err
		}
		return &core.BoolExpr{
			Operator: core.NOT,
			Operands: []*core.BoolExpr{operand},
		}, nil
	}
	return nil, fmt.Errorf("%w: unknown token %v", core.ErrBoolExprSyntaxError, n.Token)
}
