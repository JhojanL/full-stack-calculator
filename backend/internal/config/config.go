// Package config parses and validates the service's command-line settings.
package config

import (
	"flag"
	"fmt"
	"io"
	"math"
	"net/url"
	"strings"
)

// Config contains listener, environment, rate-limiting, and browser-origin settings.
// An empty TrustedOrigins list disables cross-origin grants.
// DisplayVersion requests version output instead of server startup.
type Config struct {
	DisplayVersion bool
	Port           int
	Environment    string
	Limiter        Limiter
	TrustedOrigins []string
}

// Limiter configures per-IP token buckets. RPS is the positive refill rate in
// requests per second, Burst is the positive capacity, and Enabled controls use.
type Limiter struct {
	Enabled bool
	RPS     float64
	Burst   int
}

// Parse reads CLI args (excluding the executable name) without process-global flags.
// Empty args select defaults; output receives flag help/errors (nil uses stderr).
// On error, the returned Config may be partial and must not be used to start the API.
func Parse(args []string, output io.Writer) (Config, error) {
	cfg := Config{}
	flags := flag.NewFlagSet("calculator", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.BoolVar(&cfg.DisplayVersion, "version", false, "Display version and exit")
	flags.IntVar(&cfg.Port, "port", 4000, "API server port (1-65535)")
	flags.StringVar(&cfg.Environment, "env", "development", "Environment (development|staging|production)")
	flags.BoolVar(&cfg.Limiter.Enabled, "limiter-enabled", true, "Enable per-IP calculation rate limiting")
	flags.Float64Var(&cfg.Limiter.RPS, "limiter-rps", 2, "Rate limiter requests per second")
	flags.IntVar(&cfg.Limiter.Burst, "limiter-burst", 4, "Rate limiter burst capacity")
	var origins string
	flags.StringVar(&origins, "cors-trusted-origins", "", "Trusted CORS origins (space separated)")
	if err := flags.Parse(args); err != nil {
		return cfg, err
	}
	if flags.NArg() != 0 {
		return cfg, fmt.Errorf("unexpected positional arguments")
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return cfg, fmt.Errorf("port must be between 1 and 65535")
	}
	switch cfg.Environment {
	case "development", "staging", "production":
	default:
		return cfg, fmt.Errorf("environment must be development, staging, or production")
	}
	if cfg.Limiter.RPS <= 0 || math.IsNaN(cfg.Limiter.RPS) || math.IsInf(cfg.Limiter.RPS, 0) {
		return cfg, fmt.Errorf("limiter-rps must be finite and positive")
	}
	if cfg.Limiter.Burst < 1 {
		return cfg, fmt.Errorf("limiter-burst must be positive")
	}
	cfg.TrustedOrigins = strings.Fields(origins)
	for _, origin := range cfg.TrustedOrigins {
		u, err := url.Parse(origin)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.Hostname() == "" || strings.Contains(u.Host, "*") || u.User != nil || u.Path != "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
			return cfg, fmt.Errorf("invalid CORS origin %q: use an http(s) origin without a path", origin)
		}
	}
	return cfg, nil
}
