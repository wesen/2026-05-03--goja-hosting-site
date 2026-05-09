package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

type MultiServer struct {
	cfg     MultiConfig
	sites   map[string]*Server
	httpSrv *http.Server
}

func NewMultiServer(cfg MultiConfig) (*MultiServer, error) {
	if err := cfg.Normalize(); err != nil {
		return nil, err
	}
	m := &MultiServer{cfg: cfg, sites: map[string]*Server{}}
	for _, site := range cfg.Sites {
		srv, err := NewServer(Config{DBPath: site.DBPath, ScriptDirs: site.ScriptDirs, Dev: cfg.Dev})
		if err != nil {
			_ = m.Close(context.Background())
			return nil, fmt.Errorf("create site %s (%s): %w", site.Name, site.Host, err)
		}
		m.sites[site.Host] = srv
	}
	return m, nil
}

func (m *MultiServer) Run(ctx context.Context) error {
	m.httpSrv = &http.Server{Addr: m.cfg.Addr, Handler: m, ReadHeaderTimeout: 5 * time.Second}
	errCh := make(chan error, 1)
	go func() {
		err := m.httpSrv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = m.httpSrv.Shutdown(shutdownCtx)
		return nil
	case err := <-errCh:
		return err
	}
}

func (m *MultiServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
		return
	}
	host := normalizeHost(r.Host)
	site := m.sites[host]
	if site == nil {
		http.Error(w, "unknown goja-site host", http.StatusNotFound)
		return
	}
	site.ServeHTTP(w, r)
}

func (m *MultiServer) Close(ctx context.Context) error {
	var errs []error
	if m.httpSrv != nil {
		if err := m.httpSrv.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	hosts := make([]string, 0, len(m.sites))
	for host := range m.sites {
		hosts = append(hosts, host)
	}
	sort.Strings(hosts)
	for _, host := range hosts {
		if err := m.sites[host].Close(ctx); err != nil {
			errs = append(errs, fmt.Errorf("close %s: %w", host, err))
		}
	}
	return errors.Join(errs...)
}

func (m *MultiServer) SiteHosts() []string {
	hosts := make([]string, 0, len(m.sites))
	for host := range m.sites {
		hosts = append(hosts, host)
	}
	sort.Strings(hosts)
	return hosts
}

func (m *MultiServer) SiteSummary() string {
	return strings.Join(m.SiteHosts(), ", ")
}
