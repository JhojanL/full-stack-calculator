package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/JhojanL/full-stack-calculator/backend/internal/config"
	"github.com/JhojanL/full-stack-calculator/backend/internal/httpapi"
	"github.com/JhojanL/full-stack-calculator/backend/internal/vcs"
)

var version = vcs.Version()

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg, err := config.Parse(os.Args[1:], os.Stderr)
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	if cfg.DisplayVersion {
		fmt.Printf("Version:\t%s\n", version)
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := serve(ctx, cfg, httpapi.New(logger, cfg, version), logger); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
