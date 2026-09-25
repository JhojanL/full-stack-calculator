package config

import (
	"flag"
	"io"
	"testing"
)

// TestParse checks valid settings and startup failures with t.
func TestParse(t *testing.T) {
	cfg, err := Parse(nil, io.Discard)
	if err != nil || cfg.DisplayVersion || cfg.Port != 4000 || len(cfg.TrustedOrigins) != 0 || cfg.Environment != "development" || !cfg.Limiter.Enabled || cfg.Limiter.RPS != 2 || cfg.Limiter.Burst != 4 {
		t.Fatalf("defaults: %+v, %v", cfg, err)
	}
	cfg, err = Parse([]string{"-port=8080", "-cors-trusted-origins=https://example.com http://localhost:5173"}, io.Discard)
	if err != nil || cfg.Port != 8080 || len(cfg.TrustedOrigins) != 2 {
		t.Fatalf("configured: %+v, %v", cfg, err)
	}
	for _, args := range [][]string{
		{"-env=invalid"}, {"-limiter-rps=0"}, {"-limiter-rps=-1"}, {"-limiter-rps=NaN"}, {"-limiter-rps=+Inf"}, {"-limiter-burst=0"}, {"-limiter-burst=-1"},
		{"-port=0"}, {"-port=65536"}, {"-port=no"}, {"unexpected"}, {"-unknown"},
		{"-cors-trusted-origins=*"}, {"-cors-trusted-origins=https://example.com/"},
		{"-cors-trusted-origins=https://user@example.com"}, {"-cors-trusted-origins=https://example.com?q=x"},
	} {
		if _, err := Parse(args, io.Discard); err == nil {
			t.Errorf("args %q: expected error", args)
		}
	}
	cfg, err = Parse([]string{"-env=production", "-limiter-enabled=false", "-limiter-rps=5", "-limiter-burst=10"}, io.Discard)
	if err != nil || cfg.Environment != "production" || cfg.Limiter.Enabled || cfg.Limiter.RPS != 5 || cfg.Limiter.Burst != 10 {
		t.Fatalf("limiter config: %+v, %v", cfg, err)
	}
	if _, err := Parse([]string{"-help"}, io.Discard); err != flag.ErrHelp {
		t.Errorf("help: %v", err)
	}
}

// TestParseVersion checks version flag forms and malformed values with t.
func TestParseVersion(t *testing.T) {
	for _, arg := range []string{"-version", "-version=true", "-version=false"} {
		cfg, err := Parse([]string{arg}, io.Discard)
		if err != nil || cfg.DisplayVersion != (arg != "-version=false") {
			t.Errorf("%s: config %+v, error %v", arg, cfg, err)
		}
	}
	if _, err := Parse([]string{"-version=invalid"}, io.Discard); err == nil {
		t.Error("expected an error for malformed version flag")
	}
}
