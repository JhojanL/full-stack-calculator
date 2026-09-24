// Package calculator parses and evaluates the expression language in docs/openapi.yaml.
package calculator

// Error is a stable calculator failure code, independent of HTTP transport.
type Error string

// Error codes distinguish syntax, domain, and numerical failures.
const (
	EmptyExpression      Error = "EMPTY_EXPRESSION"
	InvalidExpression    Error = "INVALID_EXPRESSION"
	InvalidOperand       Error = "INVALID_OPERAND"
	UnmatchedParentheses Error = "UNMATCHED_PARENTHESES"
	UnsupportedOperation Error = "UNSUPPORTED_OPERATION"
	DivisionByZero       Error = "DIVISION_BY_ZERO"
	NegativeSquareRoot   Error = "NEGATIVE_SQUARE_ROOT"
	UnsupportedPower     Error = "UNSUPPORTED_POWER"
	NumericOutOfRange    Error = "NUMERIC_OUT_OF_RANGE"
)

// Error returns the machine-readable failure code.
func (e Error) Error() string { return string(e) }
