package httpapi

import "net/http"

type failure struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// errorResponse writes the contractual message for code to w with status,
// using r to identify transport failures in logs.
func (app *api) errorResponse(w http.ResponseWriter, r *http.Request, status int, code string) {
	message := "Check the expression."
	switch code {
	case "INVALID_REQUEST":
		message = "Request must be a JSON object containing only a string expression."
	case "UNSUPPORTED_MEDIA_TYPE":
		message = "Content-Type must be application/json."
	case "EMPTY_EXPRESSION":
		message = "Enter a calculation."
	case "UNMATCHED_PARENTHESES":
		message = "Check the parentheses."
	case "DIVISION_BY_ZERO":
		message = "Cannot divide by zero."
	case "NEGATIVE_SQUARE_ROOT":
		message = "Square root requires a nonnegative value."
	case "UNSUPPORTED_POWER":
		message = "This power cannot be calculated."
	case "NUMERIC_OUT_OF_RANGE":
		message = "The result is outside the supported range."
	case "INTERNAL_ERROR":
		message = "Could not calculate. Try again."
	}
	app.writeJSON(w, r, status, envelope{"error": failure{code, message}})
}
