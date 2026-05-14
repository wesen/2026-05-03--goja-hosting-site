package observability

import "strings"

// Config controls the private diagnostics listener used for Prometheus metrics
// and optional pprof handlers.
type Config struct {
	MetricsAddr string
	MetricsPath string
	EnablePprof bool
}

func (c Config) normalizedMetricsPath() string {
	path := strings.TrimSpace(c.MetricsPath)
	if path == "" {
		return "/metrics"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}
