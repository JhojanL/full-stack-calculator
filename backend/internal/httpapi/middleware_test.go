package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestCORS verifies exact origin matching and constrained preflights using t.
func TestCORS(t *testing.T) {
	handler := New(nil, []string{"http://localhost:5173"})
	tests := []struct {
		name, method, path, origin, requestedMethod, requestedHeaders string
		grant                                                         bool
	}{
		{"trusted actual", "POST", "/calculate", "http://localhost:5173", "", "", true},
		{"untrusted actual", "POST", "/calculate", "http://localhost:5173.evil.test", "", "", false},
		{"trusted preflight", "OPTIONS", "/calculate", "http://localhost:5173", "POST", "content-type", true},
		{"header casing", "OPTIONS", "/calculate", "http://localhost:5173", "POST", "Content-Type", true},
		{"no headers", "OPTIONS", "/calculate", "http://localhost:5173", "POST", "", true},
		{"untrusted preflight", "OPTIONS", "/calculate", "https://evil.test", "POST", "Content-Type", false},
		{"wrong method", "OPTIONS", "/calculate", "http://localhost:5173", "DELETE", "Content-Type", false},
		{"wrong header", "OPTIONS", "/calculate", "http://localhost:5173", "POST", "Content-Type, Authorization", false},
		{"wrong path", "OPTIONS", "/missing", "http://localhost:5173", "POST", "Content-Type", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(`{"expression":"1+2"}`))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Origin", tt.origin)
			request.Header.Set("Access-Control-Request-Method", tt.requestedMethod)
			request.Header.Set("Access-Control-Request-Headers", tt.requestedHeaders)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			got := response.Header().Get("Access-Control-Allow-Origin")
			if tt.grant && got != tt.origin || !tt.grant && got != "" {
				t.Fatalf("grant = %q, want grant %t", got, tt.grant)
			}
			if tt.grant && tt.method == "OPTIONS" {
				if response.Code != http.StatusNoContent {
					t.Errorf("preflight status %d, want 204", response.Code)
				}
				if response.Header().Get("Access-Control-Allow-Methods") != "POST" {
					t.Error("preflight must advertise POST")
				}
				if response.Header().Get("Access-Control-Allow-Headers") != "Content-Type" {
					t.Error("preflight must advertise Content-Type")
				}
			}
			vary := strings.Join(response.Header().Values("Vary"), ",")
			for _, header := range []string{"Origin", "Access-Control-Request-Method", "Access-Control-Request-Headers"} {
				if !strings.Contains(vary, header) {
					t.Errorf("missing Vary: %s", header)
				}
			}
		})
	}
}

// TestRecovery verifies generic errors and behavior after commitment with t.
func TestRecovery(t *testing.T) {
	app := &api{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	response := httptest.NewRecorder()
	handler := app.recoverPanic(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("private details") }))
	handler.ServeHTTP(response, httptest.NewRequest("POST", "/calculate", nil))
	if response.Code != 500 || response.Header().Get("Connection") != "close" {
		t.Fatalf("unexpected recovery response: %v", response.Result())
	}
	assertJSON(t, response.Body.String(), `{"error":{"code":"INTERNAL_ERROR","message":"Could not calculate. Try again."}}`)

	t.Run("committed response", func(t *testing.T) {
		response := httptest.NewRecorder()
		handler := app.recoverPanic(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			_, _ = w.Write([]byte("partial"))
			panic("failure after write")
		}))
		defer func() {
			if got := recover(); got != http.ErrAbortHandler {
				t.Errorf("panic = %v, want ErrAbortHandler", got)
			}
			if response.Body.String() != "partial" {
				t.Errorf("response was corrupted: %s", response.Body)
			}
		}()
		handler.ServeHTTP(response, httptest.NewRequest("POST", "/calculate", nil))
	})
	t.Run("abort sentinel", func(t *testing.T) {
		defer func() {
			if got := recover(); got != http.ErrAbortHandler {
				t.Errorf("got %v, want abort sentinel", got)
			}
		}()
		app.recoverPanic(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic(http.ErrAbortHandler) })).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	})
}

// TestCanceledRequest checks unexpected service failure mapping using t.
func TestCanceledRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	request := httptest.NewRequest("POST", "/calculate", strings.NewReader(`{"expression":"1+2"}`)).WithContext(ctx)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	New(slog.New(slog.NewTextHandler(io.Discard, nil)), nil).ServeHTTP(response, request)
	if response.Code != 500 {
		t.Fatalf("status = %d, want 500", response.Code)
	}
	assertJSON(t, response.Body.String(), `{"error":{"code":"INTERNAL_ERROR","message":"Could not calculate. Try again."}}`)
}
