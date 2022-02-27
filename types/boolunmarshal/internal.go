package boolunmarshal

import (
	"fmt"
	"strings"
)

type TokenType int

func (t TokenType) String() string {
	switch t {
	case EOF:
		return "EOF"
	case INDENTIFIER:
		return "INDENTIFIER"
	case AND:
		return "AND"
	case OR:
		return "OR"
	case NOT:
		return "NOT"
	default:
		return "UNKNOWN"
	}
}

const (
	UNKNOWN TokenType = iota
	EOF
	INDENTIFIER
	AND
	OR
	NOT
)

var reservedTokens = map[TokenType]string{AND: "and", OR: "or", NOT: "not"}

type Node struct {
	TT    TokenType
	Token string
	Nodes []Node
}

func (n Node) isExpression() bool {
	if n.TT == INDENTIFIER || (n.TT == NOT && len(n.Nodes) == 1 && n.Nodes[0].isExpression()) {
		return true
	}
	if (n.TT != OR && n.TT != AND) || len(n.Nodes) < 2 {
		return false
	}
	for _, node := range n.Nodes {
		if !node.isExpression() {
			return false
		}
	}
	return true
}

type BoolExprParser struct {
	s string
	i int
}

func NewExprParser(s string) *BoolExprParser {
	return &BoolExprParser{s: s, i: 0}
}

func (p *BoolExprParser) error(message string, args ...interface{}) error {
	if len(p.s) == 0 || p.i == len(p.s) {
		return fmt.Errorf(fmt.Sprintf("parsing '%v': %v", p.s, message), args...)
	}
	return fmt.Errorf(fmt.Sprintf("parsing '%v[%v]%v': %v", p.s[:p.i], string(p.s[p.i]), p.s[p.i:], message), args...)
}

func (p *BoolExprParser) Parse() (Node, error) {
	p.s = strings.ToLower(p.s)
	p.i = 0

	node, err := p.pop()
	if err != nil {
		return Node{}, err
	}
	if node.isExpression() {
		return p.popAndOrChain(node)
	}
	switch node.TT {
	case NOT:
		node, err := p.pop()
		if err != nil {
			return Node{}, err
		}
		if !node.isExpression() {
			return Node{}, p.error("expected EXPRESSION after NOT but found %v", node.Token)
		}
		eofNode, err := p.pop()
		if err != nil {
			return Node{}, err
		}
		if eofNode.TT != EOF {
			return Node{}, p.error("expected EOF but found %v", node.Token)
		}
		return Node{TT: NOT, Token: "not", Nodes: []Node{node}}, nil
	case EOF:
		return Node{}, p.error("reached EOF when beginning to parse")
	case AND, OR:
		return Node{}, p.error("cannot start expression with AND or NOT")
	default:
		return Node{}, p.error("unknown token %v when beginning to parse", node.Token)
	}
}

func (p *BoolExprParser) popAndOrChain(node Node) (Node, error) {
	nodes := []Node{node}
	operator := UNKNOWN
	isOperatorNext := true
	for {
		nextNode, err := p.pop()
		if err != nil {
			return Node{}, err
		}
		if isOperatorNext {
			if nextNode.TT == EOF {
				if operator == UNKNOWN {
					return node, nil
				}
				return Node{TT: operator, Token: strings.ToLower(operator.String()), Nodes: nodes}, nil
			} else {
				if nextNode.isExpression() || (nextNode.TT != AND && nextNode.TT != OR) {
					return Node{}, p.error("expected literal AND/OR literal but found '%v'", nextNode.Token)
				}
				if operator != UNKNOWN && nextNode.TT != operator {
					return Node{}, p.error("expected literal %v but found literal %v", operator, nextNode.TT)
				}
				operator = nextNode.TT
			}
		} else {
			if !nextNode.isExpression() {
				return Node{}, p.error("expected EXPRESSION but found '%v'", node.Token)
			}
			nodes = append(nodes, nextNode)
		}
		isOperatorNext = !isOperatorNext
	}
}

func (p *BoolExprParser) pop() (Node, error) {
	p.popSpaces()
	if p.i == len(p.s) {
		return Node{TT: EOF, Token: ""}, nil
	}
	node, _ := p.popExpression()
	for reservedTT, reservedToken := range reservedTokens {
		if node.TT == INDENTIFIER && node.Token == reservedToken {
			return Node{TT: reservedTT, Token: reservedToken}, nil
		}
	}
	return node, nil
}

func (p *BoolExprParser) popExpression() (Node, error) {
	p.popSpaces()
	if p.i >= len(p.s) {
		return Node{TT: EOF, Token: ""}, nil
	}
	switch {
	case p.s[p.i] == '(':
		raw := []byte{}
		nestLevel := -1
		for p.i < len(p.s) && (p.s[p.i] != ')' || nestLevel > 0) {
			if p.s[p.i] == '(' {
				nestLevel++
			}
			if p.s[p.i] == ')' {
				nestLevel--
			}
			raw = append(raw, p.s[p.i])
			p.i++
		}
		if p.i == len(p.s) {
			return Node{TT: UNKNOWN, Token: string(raw)}, nil
		}
		n, err := NewExprParser(string(raw[1:])).Parse()
		if err != nil {
			return Node{}, err
		}
		p.i++
		return n, nil
	case (p.s[p.i] >= 'a' && p.s[p.i] <= 'z') || p.s[p.i] == '-' || p.s[p.i] == '_':
		identifier := []byte{}
		for p.i < len(p.s) && ((p.s[p.i] >= 'a' && p.s[p.i] <= 'z') || p.s[p.i] == '-' || p.s[p.i] == '_') {
			identifier = append(identifier, p.s[p.i])
			p.i++
		}
		if p.i < len(p.s) && p.s[p.i] != ' ' && p.s[p.i] != ')' {
			identifier = append(identifier, p.s[p.i])
			return Node{TT: UNKNOWN, Token: string(identifier)}, nil
		}
		return Node{TT: INDENTIFIER, Token: string(identifier)}, nil
	default:
		return Node{TT: UNKNOWN, Token: string(p.s[p.i])}, nil
	}
}

func (p *BoolExprParser) popSpaces() {
	for p.i < len(p.s) && p.s[p.i] == ' ' {
		p.i++
	}
}
