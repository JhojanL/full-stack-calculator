// Package httpapi exposes the calculator's HTTP contract.
package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/JhojanL/full-stack-calculator/backend/internal/config"
	"github.com/julienschmidt/httprouter"
)

type api struct {
	logger  *slog.Logger
	config  config.Config
	version string
}

// New constructs an independent router, logging transport failures to logger.
// A nil logger uses slog.Default. cfg must contain validated runtime settings;
// version is the build identifier reported by the healthcheck. Each router owns
// its limiter state and starts no background goroutines.
func New(logger *slog.Logger, cfg config.Config, version string) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	app := &api{logger: logger, config: cfg, version: version}
	router := httprouter.New()
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false
	router.Handler(http.MethodPost, "/calculate", app.rateLimit(http.HandlerFunc(app.calculate)))
	router.HandlerFunc(http.MethodGet, "/healthcheck", app.healthcheck)
	return app.recoverPanic(app.enableCORS(router, append([]string(nil), cfg.TrustedOrigins...)))
}
