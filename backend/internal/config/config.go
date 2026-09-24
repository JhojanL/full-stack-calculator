// Package config parses and validates the service's command-line settings.
package config

import (
	"flag"
	"fmt"
	"io"
	"net/url"
	"strings"
)

// Config contains the listener port and explicitly trusted browser origins.
// An empty TrustedOrigins list disables cross-origin grants.
type Config struct {
	Port           int
	TrustedOrigins []string
}

// Parse reads args without process-global flags and writes flag help/errors to output.
func Parse(args []string, output io.Writer) (Config, error) {
	cfg := Config{}
	flags := flag.NewFlagSet("calculator", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.IntVar(&cfg.Port, "port", 4000, "API server port (1-65535)")
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
	cfg.TrustedOrigins = strings.Fields(origins)
	for _, origin := range cfg.TrustedOrigins {
		u, err := url.Parse(origin)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.Hostname() == "" || strings.Contains(u.Host, "*") || u.User != nil || u.Path != "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
			return cfg, fmt.Errorf("invalid CORS origin %q: use an http(s) origin without a path", origin)
		}
	}
	return cfg, nil
}
