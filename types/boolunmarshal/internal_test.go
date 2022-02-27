package boolunmarshal

import (
	"errors"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	var anyError = errors.New("any error for now...")

	tss := []struct {
		name     string
		s        string
		err      error
		expected Node
	}{
		{
			name:     "Empty string cannot be parsed",
			s:        "",
			err:      anyError,
			expected: Node{},
		},
		{
			name:     "Cannot start with AND",
			s:        "and a",
			err:      anyError,
			expected: Node{},
		},
		{
			name:     "Cannot start with OR",
			s:        "or a",
			err:      anyError,
			expected: Node{},
		},
		{
			name:     "'a' works: just an identifier",
			s:        "a",
			err:      nil,
			expected: Node{TT: INDENTIFIER, Token: "a"},
		},
		{
			name:     "'a and b' works: basic AND",
			s:        "a and b",
			err:      nil,
			expected: Node{TT: AND, Token: "and", Nodes: []Node{{TT: INDENTIFIER, Token: "a"}, {TT: INDENTIFIER, Token: "b"}}},
		},
		{
			name:     "'a or b' works: basic OR",
			s:        "a or b",
			err:      nil,
			expected: Node{TT: OR, Token: "or", Nodes: []Node{{TT: INDENTIFIER, Token: "a"}, {TT: INDENTIFIER, Token: "b"}}},
		},
		{
			name:     "'cats and bears' works: multiple letter variables",
			s:        "cats and bears",
			err:      nil,
			expected: Node{TT: AND, Token: "and", Nodes: []Node{{TT: INDENTIFIER, Token: "cats"}, {TT: INDENTIFIER, Token: "bears"}}},
		},
		{
			name:     "'andy and orwell' works: variables that start with reserved works are ok",
			s:        "andy and orwell",
			err:      nil,
			expected: Node{TT: AND, Token: "and", Nodes: []Node{{TT: INDENTIFIER, Token: "andy"}, {TT: INDENTIFIER, Token: "orwell"}}},
		},
		{
			name:     "'a and b and c' works: multiple AND",
			s:        "a and b and c",
			err:      nil,
			expected: Node{TT: AND, Token: "and", Nodes: []Node{{TT: INDENTIFIER, Token: "a"}, {TT: INDENTIFIER, Token: "b"}, {TT: INDENTIFIER, Token: "c"}}},
		},
		{
			name:     "'a or b or c' works: multiple OR",
			s:        "a or b or c",
			err:      nil,
			expected: Node{TT: OR, Token: "or", Nodes: []Node{{TT: INDENTIFIER, Token: "a"}, {TT: INDENTIFIER, Token: "b"}, {TT: INDENTIFIER, Token: "c"}}},
		},
		{
			name:     "'a Or b oR C' works: case insensitive",
			s:        "a or b or c",
			err:      nil,
			expected: Node{TT: OR, Token: "or", Nodes: []Node{{TT: INDENTIFIER, Token: "a"}, {TT: INDENTIFIER, Token: "b"}, {TT: INDENTIFIER, Token: "c"}}},
		},
		{
			name:     "'(not a) and b and (c)' works",
			s:        "(not a) and b and (c)",
			err:      nil,
			expected: Node{TT: AND, Token: "and", Nodes: []Node{{TT: NOT, Token: "not", Nodes: []Node{{TT: INDENTIFIER, Token: "a"}}}, {TT: INDENTIFIER, Token: "b"}, {TT: INDENTIFIER, Token: "c"}}},
		},
		{
			name:     "'(a)' works: parens resolved to 'a'",
			s:        "(a)",
			err:      nil,
			expected: Node{TT: INDENTIFIER, Token: "a"},
		},
		{
			name:     "'((((((a))))))' works: parens resolved to 'a'",
			s:        "((((((a))))))",
			err:      nil,
			expected: Node{TT: INDENTIFIER, Token: "a"},
		},
		{
			name: "'a and b or c' fails: cannot mix operators without parens",
			s:    "a and b or c",
			err:  anyError,
		},
		{
			name: "'a or b and c' fails: cannot mix operators without parens",
			s:    "a or b and c",
			err:  anyError,
		},
		{
			name: "'not a or b' fails: cannot mix operators without parens",
			s:    "not a or b",
			err:  anyError,
		},
		{
			name: "'not a or b' fails: cannot mix operators without parens",
			s:    "not a and b",
			err:  anyError,
		},
		{
			name:     "'(not a) and b' works",
			s:        "(not a) and b",
			err:      nil,
			expected: Node{TT: AND, Token: "and", Nodes: []Node{{TT: NOT, Token: "not", Nodes: []Node{{TT: INDENTIFIER, Token: "a"}}}, {TT: INDENTIFIER, Token: "b"}}},
		},
		{
			name:     "'not (a and b)' works",
			s:        "not (a and b)",
			err:      nil,
			expected: Node{TT: NOT, Token: "not", Nodes: []Node{{TT: AND, Token: "and", Nodes: []Node{{TT: INDENTIFIER, Token: "a"}, {TT: INDENTIFIER, Token: "b"}}}}},
		},
		{
			name:     "Incomplete AND",
			s:        "a and",
			err:      anyError,
			expected: Node{},
		},
		{
			name:     "Incomplete OR",
			s:        "a or",
			err:      anyError,
			expected: Node{},
		},
		{
			name:     "Incomplete NOT",
			s:        "not",
			err:      anyError,
			expected: Node{},
		},
		{
			name:     "Incomplete Parens #1",
			s:        "(a and b",
			err:      anyError,
			expected: Node{},
		},
		{
			name:     "Incomplete Parens #2",
			s:        "(not a",
			err:      anyError,
			expected: Node{},
		},
		{
			name:     "Incomplete Parens #3",
			s:        "((a and b)",
			err:      anyError,
			expected: Node{},
		},
		{
			name:     "Only parens",
			s:        "()",
			err:      anyError,
			expected: Node{},
		},
		{
			name:     "'not ((a and b))' works",
			s:        "not ((a and b))",
			err:      nil,
			expected: Node{TT: NOT, Token: "not", Nodes: []Node{{TT: AND, Token: "and", Nodes: []Node{{TT: INDENTIFIER, Token: "a"}, {TT: INDENTIFIER, Token: "b"}}}}},
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			actualNode, actualErr := NewExprParser(ts.s).Parse()

			if actualErr != nil && ts.err == nil {
				t.Logf("expected no error but had '%v'", actualErr)
				t.FailNow()
			}
			if actualErr == nil && ts.err != nil {
				t.Logf("expected error '%v' but had no error", actualErr)
				t.FailNow()
			}
			if !reflect.DeepEqual(actualNode, ts.expected) {
				t.Logf("expected %v but got %v", ts.expected, actualNode)
				t.FailNow()
			}
		})
	}
}
