package compiler

import (
	"fmt"

	"github.com/marianogappa/predictions/compiler/boolunmarshal"
	"github.com/marianogappa/predictions/types"
)

func parseBoolExpr(s string, def map[string]*types.Condition) (*types.BoolExpr, error) {
	n, err := boolunmarshal.NewExprParser(s).Parse()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", types.ErrBoolExprSyntaxError, err)
	}
	return nodeToBoolExpr(n, def)
}

func nodeToBoolExpr(n boolunmarshal.Node, def map[string]*types.Condition) (*types.BoolExpr, error) {
	switch n.TT {
	case boolunmarshal.UNKNOWN, boolunmarshal.EOF:
		return nil, fmt.Errorf("%w: attempted to parse an invalid/unresolved node to a bool expression", types.ErrBoolExprSyntaxError)
	case boolunmarshal.INDENTIFIER:
		if _, ok := def[n.Token]; !ok {
			return nil, fmt.Errorf("%w: unknown identifier '%v'...maybe you forgot to add it to the define clause", types.ErrBoolExprSyntaxError, n.Token)
		}
		return &types.BoolExpr{
			Operator: types.LITERAL,
			Literal:  def[n.Token],
		}, nil
	case boolunmarshal.AND:
		if len(n.Nodes) == 0 {
			return nil, fmt.Errorf("%w: AND clause with zero operands", types.ErrBoolExprSyntaxError)
		}
		e := types.BoolExpr{
			Operator: types.AND,
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
			return nil, fmt.Errorf("%w: OR clause with zero operands", types.ErrBoolExprSyntaxError)
		}
		e := types.BoolExpr{
			Operator: types.OR,
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
			return nil, fmt.Errorf("%w: NOT clause must have exactly one operand", types.ErrBoolExprSyntaxError)
		}
		operand, err := nodeToBoolExpr(n.Nodes[0], def)
		if err != nil {
			return nil, err
		}
		return &types.BoolExpr{
			Operator: types.NOT,
			Operands: []*types.BoolExpr{operand},
		}, nil
	}
	return nil, fmt.Errorf("%w: unknown token %v", types.ErrBoolExprSyntaxError, n.Token)
}
