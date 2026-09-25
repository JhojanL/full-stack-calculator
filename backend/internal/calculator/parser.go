package calculator

import (
	"math"
	"strconv"
	"strings"
)

type token struct {
	kind  rune
	value float64
}

// node uses 'n' for literals and '~' for unary minus; unary operands use left.
type node struct {
	op          rune
	value       float64
	left, right *node
}

// lex normalizes aliases in expression while preserving number boundaries.
// Lexical errors precede parenthesis and grammar errors.
func lex(expression string) ([]token, error) {
	runes := []rune(expression)
	var tokens []token
	for i := 0; i < len(runes); {
		c := runes[i]
		if strings.ContainsRune(" \t\r\n", c) {
			i++
			continue
		}
		if c >= '0' && c <= '9' || c == '.' {
			start := i
			dots := 0
			for i < len(runes) && (runes[i] >= '0' && runes[i] <= '9' || runes[i] == '.') {
				if runes[i] == '.' {
					dots++
				}
				i++
			}
			if c == '.' || dots > 1 || i < len(runes) && (runes[i] == 'e' || runes[i] == 'E') {
				return nil, InvalidOperand
			}
			value, err := strconv.ParseFloat(string(runes[start:i]), 64)
			if err != nil || math.IsInf(value, 0) || math.IsNaN(value) {
				return nil, NumericOutOfRange
			}
			tokens = append(tokens, token{kind: 'n', value: value})
			continue
		}
		if c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' {
			start := i
			for i < len(runes) && (runes[i] >= 'A' && runes[i] <= 'Z' || runes[i] >= 'a' && runes[i] <= 'z') {
				i++
			}
			word := string(runes[start:i])
			if word == "NaN" || word == "Infinity" {
				return nil, NumericOutOfRange
			}
			return nil, UnsupportedOperation
		}
		switch c {
		case '−':
			c = '-'
		case '×':
			c = '*'
		case '÷':
			c = '/'
		}
		if !strings.ContainsRune("+-*/^%()√", c) {
			return nil, UnsupportedOperation
		}
		if c == '*' && i+1 < len(runes) && runes[i+1] == '*' {
			return nil, UnsupportedOperation
		}
		tokens = append(tokens, token{kind: c})
		i++
	}
	if len(tokens) == 0 {
		return nil, EmptyExpression
	}
	depth := 0
	for _, t := range tokens {
		if t.kind == '(' {
			depth++
		}
		if t.kind == ')' {
			depth--
		}
		if depth < 0 {
			return nil, UnmatchedParentheses
		}
	}
	if depth != 0 {
		return nil, UnmatchedParentheses
	}
	return append(tokens, token{}), nil
}

type parser struct {
	tokens []token
	pos    int
	depth  int
}

const maxNestingDepth = 128

// nested runs next one level deeper for a parenthesized expression or a
// right-hand power operand. These are the grammar's recursive nesting edges.
func (p *parser) nested(next func() (*node, error)) (*node, error) {
	if p.depth >= maxNestingDepth {
		return nil, InvalidExpression
	}
	p.depth++
	defer func() { p.depth-- }()
	return next()
}

// parse builds a complete syntax tree from the already lexed tokens, including EOF.
func parse(tokens []token) (*node, error) {
	p := parser{tokens: tokens}
	n, err := p.sum()
	if err == nil && p.peek() != 0 {
		err = InvalidExpression
	}
	return n, err
}

func (p *parser) peek() rune { return p.tokens[p.pos].kind }

// take consumes kind when it is the next token.
func (p *parser) take(kind rune) bool {
	if p.peek() != kind {
		return false
	}
	p.pos++
	return true
}

// sum folds addition and subtraction left to right.
func (p *parser) sum() (*node, error) {
	n, err := p.product()
	for err == nil && (p.peek() == '+' || p.peek() == '-') {
		op := p.peek()
		p.pos++
		var right *node
		right, err = p.product()
		n = &node{op: op, left: n, right: right}
	}
	return n, err
}

// product folds multiplication and division left to right.
func (p *parser) product() (*node, error) {
	n, err := p.unary()
	for err == nil && (p.peek() == '*' || p.peek() == '/') {
		op := p.peek()
		p.pos++
		var right *node
		right, err = p.unary()
		n = &node{op: op, left: n, right: right}
	}
	return n, err
}

// unary permits one leading minus, applied after any power expression.
func (p *parser) unary() (*node, error) {
	negative := p.take('-')
	n, err := p.power()
	if negative {
		n = &node{op: '~', left: n}
	}
	return n, err
}

// power applies one postfix percent before a right-associative power.
// Parsing the exponent through unary also permits negative exponents.
func (p *parser) power() (*node, error) {
	n, err := p.primary()
	if err != nil {
		return nil, err
	}
	if p.take('%') {
		n = &node{op: '%', left: n}
	}
	if p.take('^') {
		right, err := p.nested(p.unary)
		return &node{op: '^', left: n, right: right}, err
	}
	return n, nil
}

// primary accepts a group, root, or number. An ungrouped root consumes only
// a possibly negative number, leaving following operators to the outer parser.
func (p *parser) primary() (*node, error) {
	if p.take('(') {
		n, err := p.nested(p.sum)
		if err != nil {
			return nil, err
		}
		if !p.take(')') {
			return nil, InvalidExpression
		}
		return n, nil
	}
	if p.take('√') {
		if p.peek() == '(' {
			n, err := p.primary()
			return &node{op: '√', left: n}, err
		}
		negative := p.take('-')
		n, err := p.number()
		if negative {
			n = &node{op: '~', left: n}
		}
		return &node{op: '√', left: n}, err
	}
	return p.number()
}

func (p *parser) number() (*node, error) {
	if p.peek() != 'n' {
		return nil, InvalidExpression
	}
	n := &node{op: 'n', value: p.tokens[p.pos].value}
	p.pos++
	return n, nil
}
