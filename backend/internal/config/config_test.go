package config

import (
	"flag"
	"io"
	"testing"
)

// TestParse checks valid settings and startup failures with t.
func TestParse(t *testing.T) {
	cfg, err := Parse(nil, io.Discard)
	if err != nil || cfg.Port != 4000 || len(cfg.TrustedOrigins) != 0 {
		t.Fatalf("defaults: %+v, %v", cfg, err)
	}
	cfg, err = Parse([]string{"-port=8080", "-cors-trusted-origins=https://example.com http://localhost:5173"}, io.Discard)
	if err != nil || cfg.Port != 8080 || len(cfg.TrustedOrigins) != 2 {
		t.Fatalf("configured: %+v, %v", cfg, err)
	}
	for _, args := range [][]string{
		{"-port=0"}, {"-port=65536"}, {"-port=no"}, {"unexpected"}, {"-unknown"},
		{"-cors-trusted-origins=*"}, {"-cors-trusted-origins=https://example.com/"},
		{"-cors-trusted-origins=https://user@example.com"}, {"-cors-trusted-origins=https://example.com?q=x"},
	} {
		if _, err := Parse(args, io.Discard); err == nil {
			t.Errorf("args %q: expected error", args)
		}
	}
	if _, err := Parse([]string{"-help"}, io.Discard); err != flag.ErrHelp {
		t.Errorf("help: %v", err)
	}
}
