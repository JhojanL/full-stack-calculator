package calculator

import (
	"context"
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// TestEvaluate verifies all operations, grammar precedence, and float64 display with t.
func TestEvaluate(t *testing.T) {
	tests := []struct{ expression, want string }{
		{"2 + 3 × 4", "14"}, {"(2 + 3) × 4", "20"}, {"2^3^2", "512"},
		{"5 − 8", "-3"}, {"1 ÷ 3", "0.333"}, {"−2 ÷ 3", "-0.667"},
		{"√2", "1.414"}, {"√(9 + 5)", "3.742"}, {"√9+5", "8"},
		{"200 × 10%", "20"}, {"200 + 10%", "200.1"}, {"200*(1+10%)", "220"},
		{"-2^2", "-4"}, {"(-2)^2", "4"}, {"2^-2", "0.25"}, {"2^-2^2", "0.063"},
		{"100^50%", "10"}, {"√9%", "0.03"}, {"(20+30)%", "0.5"},
		{"8/4/2", "1"}, {"8-4-2", "2"}, {"2--3", "5"}, {"2*-3", "-6"},
		{"(-2)^3", "-8"}, {"(-2)^-3", "-0.125"}, {"(-8)^(6/2)", "-512"},
		{"9^0.5", "3"}, {"27^(1/3)", "3"}, {"2^√2", "2.665"},
		{"0^2", "0"}, {"4^0", "1"}, {"√-0", "0"}, {"√(-0)", "0"},
		{"01. + 0.", "1"}, {" \t1\r\n+ 2 ", "3"},
		{"1.2345", "1.235"}, {"-1.2345", "-1.235"}, {"0.0005", "0.001"},
		{"-0.0005", "-0.001"}, {"-0.0004", "0"}, {"2.500", "2.5"},
		{"1/3*3", "1"}, {"0.333*3", "0.999"}, {"1/0.0001", "10000"},
		{"10^-400", "0"}, {"9007199254740992+1-9007199254740992", "0"},
		{"0.1+0.2", "0.3"}, {"2^53", "9007199254740992"},
	}
	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			got, err := Evaluate(t.Context(), tt.expression)
			if err != nil || got != tt.want {
				t.Fatalf("Evaluate(%q) = %q, %v; want %q", tt.expression, got, err, tt.want)
			}
		})
	}
}

// TestDomainErrors checks evaluation ordering and unrounded domain decisions with t.
func TestDomainErrors(t *testing.T) {
	tests := []struct {
		expression string
		want       error
	}{
		{"1/0", DivisionByZero}, {"1/-0", DivisionByZero}, {"1/(2-2)", DivisionByZero},
		{"√-1", NegativeSquareRoot}, {"√(2-3)", NegativeSquareRoot}, {"√-0.0001", NegativeSquareRoot},
		{"0^0", UnsupportedPower}, {"0^-1", UnsupportedPower}, {"(-2)^0.5", UnsupportedPower},
		{"(-2)^2.0001", UnsupportedPower}, {"2^1024", NumericOutOfRange},
		{"10^308*10", NumericOutOfRange}, {"(10^308*10)*0", NumericOutOfRange},
		{"√-1+1/0", NegativeSquareRoot}, {"1/0+√-1", DivisionByZero},
		{"0^(1/0)", DivisionByZero}, {"1/0+", InvalidExpression},
		{"1/0+(", UnmatchedParentheses}, {"1/0+sin(1)", UnsupportedOperation},
		{"1/(10^-400)", DivisionByZero}, {"1" + strings.Repeat("0", 309), NumericOutOfRange},
	}
	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			_, err := Evaluate(t.Context(), tt.expression)
			if err != tt.want {
				t.Fatalf("Evaluate(%q): got %v, want %v", tt.expression, err, tt.want)
			}
		})
	}
}

// TestMaximumFloat verifies that formatting a finite maximum does not overflow with t.
func TestMaximumFloat(t *testing.T) {
	want := strconv.FormatFloat(math.MaxFloat64, 'f', 0, 64)
	got, err := Evaluate(t.Context(), want)
	if err != nil || got != want {
		t.Fatalf("maximum float: got %q, %v; want %q", got, err, want)
	}
}

// TestCancellation verifies that canceled ctx stops evaluation using t.
func TestCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := Evaluate(ctx, "1+2"); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want cancellation", err)
	}
}

// TestConcurrentEvaluation exercises independent requests through t's race run.
func TestConcurrentEvaluation(t *testing.T) {
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			got, err := Evaluate(t.Context(), "(2+3)*√9")
			if err != nil || got != "15" {
				t.Errorf("got %q, %v; want 15", got, err)
			}
		})
	}
	wg.Wait()
}

// FuzzEvaluate checks that arbitrary expressions cannot panic and every success
// has the contract's decimal format. f supplies seeds and mutation input.
func FuzzEvaluate(f *testing.F) {
	for _, seed := range []string{"2^3^2", "√-1", "1..2", "()", "1/0", "", "NaN", "2 3"} {
		f.Add(seed)
	}
	decimal := regexp.MustCompile(`^(?:0|-?[1-9][0-9]*|-?(?:0|[1-9][0-9]*)\.[0-9]{0,2}[1-9])$`)
	f.Fuzz(func(t *testing.T, expression string) {
		if len(expression) > 8192 {
			return
		}
		got, err := Evaluate(t.Context(), expression)
		if err == nil && !decimal.MatchString(got) {
			t.Fatalf("expression %q returned invalid result %q", expression, got)
		}
	})
}

// TestNestingLimit exercises inclusive depth boundaries and mixed recursion with t.
func TestNestingLimit(t *testing.T) {
	for _, tt := range []struct {
		name, expression string
		want             error
	}{
		{"128 groups", strings.Repeat("(", 128) + "1" + strings.Repeat(")", 128), nil},
		{"129 groups", strings.Repeat("(", 129) + "1" + strings.Repeat(")", 129), InvalidExpression},
		{"128 powers", strings.Repeat("1^", 128) + "1", nil},
		{"129 powers", strings.Repeat("1^", 129) + "1", InvalidExpression},
		{"mixed depth", strings.Repeat("(", 64) + strings.Repeat("1^", 65) + "1" + strings.Repeat(")", 64), InvalidExpression},
		{"root nesting", strings.Repeat("√(", 129) + "1" + strings.Repeat(")", 129), InvalidExpression},
		{"flat expression", strings.Repeat("1+", 5000) + "1", nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(t.Context(), tt.expression)
			if err != tt.want {
				t.Fatalf("got %v, want %v", err, tt.want)
			}
		})
	}
}
