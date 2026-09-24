package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"
)

// TestRunServerCancellation checks shutdown even if cancellation precedes startup with t.
func TestRunServerCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	server := &http.Server{Addr: "127.0.0.1:0", ReadHeaderTimeout: time.Second}
	if err := runServer(ctx, server, slog.New(slog.NewTextHandler(io.Discard, nil))); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

// TestRunServerListenError verifies startup failures return without waiting for a signal with t.
func TestRunServerListenError(t *testing.T) {
	server := &http.Server{Addr: "127.0.0.1:invalid", ReadHeaderTimeout: time.Second}
	if err := runServer(t.Context(), server, slog.New(slog.NewTextHandler(io.Discard, nil))); err == nil {
		t.Fatal("expected listener error")
	}
}
