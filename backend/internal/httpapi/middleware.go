package httpapi

import (
	"net/http"
	"slices"
	"strings"
)

// recoverPanic runs next and converts a panic before response commitment to JSON.
// Once a response has started, it aborts the connection rather than corrupting it.
func (app *api) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := &responseWriter{ResponseWriter: w}
		defer func() {
			if value := recover(); value != nil {
				if value == http.ErrAbortHandler {
					panic(value)
				}
				app.logger.Error("request panic", "panic", value, "method", r.Method, "path", r.URL.Path)
				if writer.committed {
					panic(http.ErrAbortHandler)
				}
				w.Header().Set("Connection", "close")
				app.errorResponse(w, r, http.StatusInternalServerError, "INTERNAL_ERROR")
			}
		}()
		next.ServeHTTP(writer, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	committed bool
}

// WriteHeader forwards status and records whether final headers were committed.
func (w *responseWriter) WriteHeader(status int) {
	if status >= 200 {
		w.committed = true
	}
	w.ResponseWriter.WriteHeader(status)
}

// Write commits the response and forwards data to the underlying writer.
func (w *responseWriter) Write(data []byte) (int, error) {
	w.committed = true
	return w.ResponseWriter.Write(data)
}

// Unwrap exposes the underlying writer to http.ResponseController.
func (w *responseWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

// enableCORS wraps next with exact matching of origins. Only calculator POST
// preflights requesting Content-Type are granted; other requests reach the router.
func (app *api) enableCORS(next http.Handler, origins []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")
		w.Header().Add("Vary", "Access-Control-Request-Headers")
		origin := r.Header.Get("Origin")
		if origin != "" && slices.Contains(origins, origin) {
			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				if r.URL.Path == "/calculate" && r.Header.Get("Access-Control-Request-Method") == http.MethodPost && allowedHeaders(r.Header.Get("Access-Control-Request-Headers")) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Methods", "POST")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
					w.WriteHeader(http.StatusNoContent)
					return
				}
			} else {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// allowedHeaders reports whether comma-separated headers request only Content-Type.
func allowedHeaders(headers string) bool {
	if headers == "" {
		return true
	}
	for header := range strings.SplitSeq(headers, ",") {
		if !strings.EqualFold(strings.TrimSpace(header), "Content-Type") {
			return false
		}
	}
	return true
}
