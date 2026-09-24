package httpapi

import (
	"errors"
	"mime"
	"net/http"
	"strings"

	"github.com/JhojanL/full-stack-calculator/backend/internal/calculator"
)

// calculate validates request r, evaluates its expression, and writes JSON to w.
func (app *api) calculate(w http.ResponseWriter, r *http.Request) {
	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" || (params["charset"] != "" && !strings.EqualFold(params["charset"], "utf-8")) {
		app.errorResponse(w, r, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE")
		return
	}
	expression, err := readExpression(w, r)
	if err != nil {
		status := http.StatusBadRequest
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			status = http.StatusRequestEntityTooLarge
		}
		app.errorResponse(w, r, status, "INVALID_REQUEST")
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
	app.writeJSON(w, r, http.StatusOK, envelope{"result": result})
}
