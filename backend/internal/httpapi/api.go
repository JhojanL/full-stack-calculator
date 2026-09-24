// Package httpapi exposes the calculator's HTTP contract.
package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type api struct{ logger *slog.Logger }

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
