package calculator

import "testing"

// TestSyntaxErrors verifies lexical, parenthesis, and grammar precedence with t.
func TestSyntaxErrors(t *testing.T) {
	tests := []struct {
		expression string
		want       error
	}{
		{"", EmptyExpression}, {" \t\r\n", EmptyExpression},
		{".5", InvalidOperand}, {"1..2", InvalidOperand}, {"1e3", InvalidOperand},
		{"1E-3", InvalidOperand}, {"NaN", NumericOutOfRange}, {"-Infinity", NumericOutOfRange},
		{"sin(1)", UnsupportedOperation}, {"2 ** 3", UnsupportedOperation}, {"3!", UnsupportedOperation},
		{"2\u00a03", UnsupportedOperation}, {"２", UnsupportedOperation},
		{"(1", UnmatchedParentheses}, {")1(", UnmatchedParentheses},
		{"(1 + .5", InvalidOperand}, {"(1 + sin(2)", UnsupportedOperation},
		{"2 +", InvalidExpression}, {"2(3)", InvalidExpression}, {"2 3", InvalidExpression},
		{"10%%", InvalidExpression}, {"+2", InvalidExpression}, {"--2", InvalidExpression},
		{"√√4", InvalidExpression}, {"√-(4)", InvalidExpression}, {"()", InvalidExpression},
	}
	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			tokens, err := lex(tt.expression)
			if err == nil {
				_, err = parse(tokens)
			}
			if err != tt.want {
				t.Fatalf("expression %q: got %v, want %v", tt.expression, err, tt.want)
			}
		})
	}
}
