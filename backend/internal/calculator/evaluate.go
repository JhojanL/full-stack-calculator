package calculator

import (
	"context"
	"math"
	"strconv"
	"strings"
)

// Evaluate parses the complete expression and evaluates it using float64.
// ctx must be non-nil; cancellation interrupts tree evaluation. Results are
// rounded only for display, to at most three places with halfway values away
// from zero. Binary floating-point precision and underflow apply throughout.
func Evaluate(ctx context.Context, expression string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	tokens, err := lex(expression)
	if err != nil {
		return "", err
	}
	tree, err := parse(tokens)
	if err != nil {
		return "", err
	}
	value, err := evaluate(ctx, tree)
	if err != nil {
		return "", err
	}
	return format(value), nil
}

// evaluate visits n's left subtree before its right subtree, checking ctx at
// each node and checking domain errors on unrounded float64 operands.
func evaluate(ctx context.Context, n *node) (float64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if n.op == 'n' {
		return n.value, nil
	}
	left, err := evaluate(ctx, n.left)
	if err != nil {
		return 0, err
	}
	var right float64
	if n.right != nil {
		right, err = evaluate(ctx, n.right)
		if err != nil {
			return 0, err
		}
	}
	var result float64
	switch n.op {
	case '~':
		result = -left
	case '+':
		result = left + right
	case '-':
		result = left - right
	case '*':
		result = left * right
	case '/':
		if right == 0 {
			return 0, DivisionByZero
		}
		result = left / right
	case '%':
		result = left / 100
	case '√':
		if left < 0 {
			return 0, NegativeSquareRoot
		}
		result = math.Sqrt(left)
	case '^':
		if left == 0 && right <= 0 || left < 0 && math.Trunc(right) != right {
			return 0, UnsupportedPower
		}
		result = math.Pow(left, right)
	default:
		panic("unknown calculator node")
	}
	if math.IsNaN(result) || math.IsInf(result, 0) {
		return 0, NumericOutOfRange
	}
	return result, nil
}

// format rounds finite value for display and emits ordinary decimal notation.
// At magnitudes >= 2^52, float64 has no fractional digits, so scaling is both
// unnecessary and potentially overflowing.
func format(value float64) string {
	if math.Abs(value) >= 1<<52 {
		return strconv.FormatFloat(value, 'f', 0, 64)
	}
	value = math.Round(value*1000) / 1000
	if value == 0 {
		return "0"
	}
	return strings.TrimRight(strings.TrimRight(strconv.FormatFloat(value, 'f', 3, 64), "0"), ".")
}
