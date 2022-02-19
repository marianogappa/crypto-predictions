package types

import (
	"errors"
	"fmt"

	"github.com/marianogappa/predictions/types/boolunmarshal"
)

func parseBoolExpr(s string, def map[string]*Condition) (*BoolExpr, error) {
	n, err := boolunmarshal.NewExprParser(s).Parse()
	if err != nil {
		return nil, err
	}
	return nodeToBoolExpr(n, def)
}

func nodeToBoolExpr(n boolunmarshal.Node, def map[string]*Condition) (*BoolExpr, error) {
	switch n.TT {
	case boolunmarshal.UNKNOWN, boolunmarshal.EOF:
		return nil, errors.New("attempted to parse an invalid/unresolved node to a bool expression")
	case boolunmarshal.INDENTIFIER:
		if _, ok := def[n.Token]; !ok {
			return nil, fmt.Errorf("unknown identifier '%v'...maybe you forgot to add it to the define clause", n.Token)
		}
		return &BoolExpr{
			Operator: LITERAL,
			Literal:  def[n.Token],
		}, nil
	case boolunmarshal.AND:
		if len(n.Nodes) == 0 {
			return nil, fmt.Errorf("AND clause with zero operands")
		}
		e := BoolExpr{
			Operator: AND,
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
			return nil, fmt.Errorf("OR clause with zero operands")
		}
		e := BoolExpr{
			Operator: OR,
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
			return nil, fmt.Errorf("NOT clause must have exactly one operand")
		}
		operand, err := nodeToBoolExpr(n.Nodes[0], def)
		if err != nil {
			return nil, err
		}
		return &BoolExpr{
			Operator: NOT,
			Operands: []*BoolExpr{operand},
		}, nil
	}
	return nil, fmt.Errorf("unknown token %v", n.Token)
}
