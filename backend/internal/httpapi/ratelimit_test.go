package httpapi

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/JhojanL/full-stack-calculator/backend/internal/config"
)

// limiterConfig returns the CLI defaults and reports parsing failures through t.
func limiterConfig(t *testing.T) config.Config {
	t.Helper()
	cfg, err := config.Parse(nil, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	cfg.TrustedOrigins = []string{"http://localhost:5173"}
	return cfg
}

// calculateFrom sends one calculation through handler from peer, adding forwarded
// as an untrusted proxy header. It returns the recorded response for assertions.
func calculateFrom(handler http.Handler, peer, forwarded string) *httptest.ResponseRecorder {
	request := httptest.NewRequest("POST", "/calculate", strings.NewReader(`{"expression":"2+3"}`))
	request.RemoteAddr = peer
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("X-Forwarded-For", forwarded)
	request.Header.Set("X-Real-IP", forwarded)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

// TestRateLimit verifies burst, refill, peer identity, and operational exemptions
// with t. synctest supplies deterministic time without real sleeps.
func TestRateLimit(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		handler := New(nil, limiterConfig(t), "build-test")
		for i := range 5 {
			response := calculateFrom(handler, fmt.Sprintf("192.0.2.1:%d", 1000+i), fmt.Sprintf("198.51.100.%d", i))
			want := 200
			if i == 4 {
				want = 429
			}
			if response.Code != want {
				t.Fatalf("request %d: status %d, want %d", i+1, response.Code, want)
			}
			if want == 429 {
				assertJSON(t, response.Body.String(), `{"error":{"code":"RATE_LIMIT_EXCEEDED","message":"Rate limit exceeded. Try again shortly."}}`)
				if response.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
					t.Error("429 lost CORS grant")
				}
			}
		}
		for _, peer := range []string{"192.0.2.2:1000", "[2001:db8::1]:1000"} {
			if response := calculateFrom(handler, peer, ""); response.Code != 200 {
				t.Errorf("independent peer %s: status %d", peer, response.Code)
			}
		}
		for range 6 {
			request := httptest.NewRequest("GET", "/healthcheck", nil)
			request.RemoteAddr = "192.0.2.1:1000"
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != 200 {
				t.Fatalf("healthcheck status = %d", response.Code)
			}
			assertJSON(t, response.Body.String(), `{"status":"available","system_info":{"environment":"development","version":"build-test"}}`)
		}
		preflight := httptest.NewRequest("OPTIONS", "/calculate", nil)
		preflight.RemoteAddr = "192.0.2.1:1000"
		preflight.Header.Set("Origin", "http://localhost:5173")
		preflight.Header.Set("Access-Control-Request-Method", "POST")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, preflight)
		if response.Code != 204 {
			t.Fatalf("preflight status = %d", response.Code)
		}
		time.Sleep(500 * time.Millisecond)
		if got := calculateFrom(handler, "192.0.2.1:1000", "").Code; got != 200 {
			t.Fatalf("refill status = %d", got)
		}
		if got := calculateFrom(handler, "192.0.2.1:1000", "").Code; got != 429 {
			t.Fatalf("only one token should refill: %d", got)
		}
	})
}

// TestRateLimitDisabled verifies the configured bypass with t.
func TestRateLimitDisabled(t *testing.T) {
	cfg := limiterConfig(t)
	cfg.Limiter.Enabled = false
	handler := New(nil, cfg, "test")
	for range 20 {
		if got := calculateFrom(handler, "192.0.2.1:1000", "").Code; got != 200 {
			t.Fatalf("disabled limiter returned %d", got)
		}
	}
}

// TestRateLimitConcurrent checks that one IP cannot exceed its burst through
// concurrent requests; t records results after all request goroutines finish.
func TestRateLimitConcurrent(t *testing.T) {
	cfg := limiterConfig(t)
	cfg.Limiter.RPS = 0.001
	handler := New(nil, cfg, "test")
	var wg sync.WaitGroup
	var accepted atomic.Int32
	for range 30 {
		wg.Go(func() {
			code := calculateFrom(handler, "192.0.2.1:1000", "").Code
			if code == 200 {
				accepted.Add(1)
			} else if code != 429 {
				t.Errorf("unexpected status %d", code)
			}
		})
	}
	wg.Wait()
	if got := accepted.Load(); got != 4 {
		t.Fatalf("accepted %d requests, want 4", got)
	}
}

// TestRateLimitCleanup checks bounded client retention, eviction, and preservation
// of depleted slow-refill buckets with t and deterministic time.
func TestRateLimitCleanup(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		handler := New(nil, limiterConfig(t), "test")
		for i := range maxLimiterClients {
			peer := fmt.Sprintf("10.0.%d.%d:1000", i/256, i%256)
			if got := calculateFrom(handler, peer, "").Code; got != 200 {
				t.Fatalf("client %d: status %d", i, got)
			}
		}
		if got := calculateFrom(handler, "192.0.2.1:1000", "").Code; got != 429 {
			t.Fatalf("full client map: got %d", got)
		}
		time.Sleep(3*time.Minute + time.Second)
		if got := calculateFrom(handler, "192.0.2.1:1000", "").Code; got != 200 {
			t.Fatalf("idle cleanup: got %d", got)
		}
	})
	synctest.Test(t, func(t *testing.T) {
		cfg := limiterConfig(t)
		cfg.Limiter.RPS = 0.001
		cfg.Limiter.Burst = 1
		handler := New(nil, cfg, "test")
		if got := calculateFrom(handler, "192.0.2.1:1000", "").Code; got != 200 {
			t.Fatal(got)
		}
		time.Sleep(4 * time.Minute)
		if got := calculateFrom(handler, "192.0.2.1:1000", "").Code; got != 429 {
			t.Fatalf("cleanup reset a depleted bucket: %d", got)
		}
	})
}
