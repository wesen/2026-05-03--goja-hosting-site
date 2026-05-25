package observability

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type DiagnosticsServer struct {
	srv *http.Server
}

func StartDiagnostics(ctx context.Context, cfg Config, obs *Observability) (*DiagnosticsServer, error) {
	if cfg.MetricsAddr == "" {
		if cfg.EnablePprof {
			return nil, fmt.Errorf("--pprof requires --metrics-addr so diagnostics stay on a private listener")
		}
		return nil, nil
	}
	if obs == nil || obs.Registry == nil {
		return nil, fmt.Errorf("observability registry is required when metrics address is configured")
	}

	mux := http.NewServeMux()
	mux.Handle(cfg.normalizedMetricsPath(), promhttp.HandlerFor(obs.Registry, promhttp.HandlerOpts{}))
	if cfg.EnablePprof {
		MountPprof(mux, "/debug/pprof")
	}

	srv := &http.Server{Addr: cfg.MetricsAddr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	ds := &DiagnosticsServer{srv: srv}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	errCh := make(chan error, 1)
	go func() {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return nil, err
		}
		return ds, nil
	case <-time.After(50 * time.Millisecond):
		return ds, nil
	}
}

func (d *DiagnosticsServer) Close(ctx context.Context) error {
	if d == nil || d.srv == nil {
		return nil
	}
	return d.srv.Shutdown(ctx)
}

func MountPprof(mux *http.ServeMux, prefix string) {
	mux.HandleFunc(prefix+"/", pprof.Index)
	mux.HandleFunc(prefix+"/cmdline", pprof.Cmdline)
	mux.HandleFunc(prefix+"/profile", pprof.Profile)
	mux.HandleFunc(prefix+"/symbol", pprof.Symbol)
	mux.HandleFunc(prefix+"/trace", pprof.Trace)
	mux.Handle(prefix+"/allocs", pprof.Handler("allocs"))
	mux.Handle(prefix+"/block", pprof.Handler("block"))
	mux.Handle(prefix+"/goroutine", pprof.Handler("goroutine"))
	mux.Handle(prefix+"/heap", pprof.Handler("heap"))
	mux.Handle(prefix+"/mutex", pprof.Handler("mutex"))
	mux.Handle(prefix+"/threadcreate", pprof.Handler("threadcreate"))
}
