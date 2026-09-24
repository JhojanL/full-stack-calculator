// Package httpapi exposes the calculator's HTTP contract.
package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strings"

	"github.com/JhojanL/full-stack-calculator/backend/internal/calculator"
	"github.com/julienschmidt/httprouter"
)

type api struct{ logger *slog.Logger }

const maxBodyBytes = 64 * 1024

type failure struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// New constructs an independent router, logging transport failures to logger.
// A nil logger uses slog.Default. origins contains exact trusted CORS origins;
// an empty list grants no cross-origin access.
func New(logger *slog.Logger, origins []string) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	app := &api{logger: logger}
	router := httprouter.New()
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false
	router.HandlerFunc(http.MethodPost, "/calculate", app.calculate)
	return app.recoverPanic(app.enableCORS(router, append([]string(nil), origins...)))
}

// calculate validates request r, evaluates its expression, and writes JSON to w.
func (app *api) calculate(w http.ResponseWriter, r *http.Request) {
	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" || (params["charset"] != "" && !strings.EqualFold(params["charset"], "utf-8")) {
		app.errorResponse(w, r, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE")
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err != nil {
		status := http.StatusBadRequest
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			status = http.StatusRequestEntityTooLarge
		}
		app.errorResponse(w, r, status, "INVALID_REQUEST")
		return
	}
	expression, err := readExpression(bytes.NewReader(body))
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "INVALID_REQUEST")
		return
	}
	result, err := calculator.Evaluate(r.Context(), expression)
	if err != nil {
		var problem calculator.Error
		if errors.As(err, &problem) {
			app.errorResponse(w, r, http.StatusUnprocessableEntity, string(problem))
			return
		}
		app.logger.Error("calculation failed", "error", err)
		app.errorResponse(w, r, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}
	app.writeJSON(w, r, http.StatusOK, struct {
		Result string `json:"result"`
	}{result})
}

// readExpression consumes body as exactly one object containing one string member.
// Token decoding deliberately rejects duplicate keys, nulls, and case aliases.
func readExpression(body io.Reader) (string, error) {
	invalid := errors.New("invalid calculation request")
	dec := json.NewDecoder(body)
	opening, err := dec.Token()
	if err != nil || opening != json.Delim('{') {
		return "", invalid
	}
	key, err := dec.Token()
	if err != nil || key != "expression" {
		return "", invalid
	}
	value, err := dec.Token()
	expression, ok := value.(string)
	if err != nil || !ok {
		return "", invalid
	}
	closing, err := dec.Token()
	if err != nil || closing != json.Delim('}') {
		return "", invalid
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return "", invalid
	}
	return expression, nil
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
	app.writeJSON(w, r, status, struct {
		Error failure `json:"error"`
	}{failure{code, message}})
}

// writeJSON serializes value before committing status to w. Request r supplies
// log context; failed writes are logged without appending a second response.
func (app *api) writeJSON(w http.ResponseWriter, r *http.Request, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(data, '\n')); err != nil {
		app.logger.Error("response write failed", "method", r.Method, "path", r.URL.Path, "error", err)
	}
}
