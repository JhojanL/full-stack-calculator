package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/JhojanL/full-stack-calculator/backend/internal/config"
)

// serve listens on cfg.Port using handler and reports lifecycle events to logger.
// Canceling ctx starts a five-second drain, followed by forced connection closure.
func serve(ctx context.Context, cfg config.Config, handler http.Handler, logger *slog.Logger) error {
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       time.Minute,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}
	return runServer(ctx, server, logger)
}

// runServer runs server until a listener error or cancellation of ctx. It waits
// for shutdown to finish and logs startup and shutdown through logger.
func runServer(ctx context.Context, server *http.Server, logger *slog.Logger) error {
	stopped := make(chan error, 1)
	go func() { stopped <- server.ListenAndServe() }()
	logger.Info("starting server", "addr", server.Addr)
	select {
	case err := <-stopped:
		return err
	case <-ctx.Done():
		drain, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(drain); err != nil {
			_ = server.Close()
			<-stopped
			return err
		}
		if err := <-stopped; !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		logger.Info("shutdown complete")
		return nil
	}
}
