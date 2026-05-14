package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/engine"
	expressmod "github.com/go-go-golems/go-go-goja/modules/express"
	"github.com/go-go-golems/go-go-goja/modules/uidsl"
	"github.com/go-go-golems/go-go-goja/pkg/gojahttp"
	"github.com/go-go-golems/goja-site/pkg/kanbanddsl"
	"github.com/go-go-golems/goja-site/pkg/observability"
	_ "github.com/mattn/go-sqlite3"
)

// Server owns the database, goja runtime, route host, and HTTP server.
type Server struct {
	cfg     Config
	db      *sql.DB
	runtime *engine.Runtime
	host    *gojahttp.Host
	httpSrv *http.Server
}

func NewServer(cfg Config) (*Server, error) {
	if cfg.Addr == "" {
		cfg.Addr = ":8080"
	}
	if cfg.DBPath == "" {
		cfg.DBPath = "./app.db"
	}
	if len(cfg.ScriptDirs) == 0 {
		cfg.ScriptDirs = []string{"./scripts"}
	}
	if err := normalizeDBPolicyConfig(&cfg); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0o755); err != nil && filepath.Dir(cfg.DBPath) != "." {
		return nil, fmt.Errorf("create db directory: %w", err)
	}
	db, err := sql.Open("sqlite3", cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite database: %w", err)
	}

	host := gojahttp.NewHost(gojahttp.HostOptions{Dev: cfg.Dev, Renderer: uidsl.RenderAny})
	dbRuntime, err := buildDatabaseRuntimeConfig(cfg, db)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	var kanbanObserver kanbanddsl.Observer
	if cfg.Observability != nil && cfg.Observability.Kanban != nil {
		kanbanObserver = observability.NewKanbanObserver(cfg.SiteName, cfg.Observability.Kanban)
	}
	registrars := []engine.RuntimeModuleRegistrar{expressmod.NewRegistrar(host), uidsl.NewRegistrar(), kanbanddsl.NewRegistrar(kanbanObserver)}
	registrars = append(registrars, dbRuntime.registrars...)

	factory, err := engine.NewBuilder().
		WithModules(dbRuntime.moduleSpecs...).
		UseModuleMiddleware(engine.MiddlewareOnly("fs", "path", "time", "timer", "yaml")).
		WithRuntimeModuleRegistrars(registrars...).
		Build()
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("build goja factory: %w", err)
	}

	rt, err := factory.NewRuntime(context.Background())
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create goja runtime: %w", err)
	}
	host.SetRuntime(rt.Owner)

	s := &Server{cfg: cfg, db: db, runtime: rt, host: host}
	if err := s.LoadScripts(context.Background()); err != nil {
		_ = s.Close(context.Background())
		return nil, err
	}
	return s, nil
}

func (s *Server) Handler() http.Handler {
	handler := http.Handler(s.host)
	if s.cfg.Observability != nil && s.cfg.Observability.HTTP != nil {
		handler = s.cfg.Observability.HTTP.Wrap(s.cfg.SiteName, handler)
	}
	if s.cfg.Observability != nil {
		handler = s.cfg.Observability.WrapTrace(s.cfg.SiteName, handler)
	}
	return handler
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Handler().ServeHTTP(w, r)
}

func (s *Server) Run(ctx context.Context) error {
	s.httpSrv = &http.Server{Addr: s.cfg.Addr, Handler: s.Handler(), ReadHeaderTimeout: 5 * time.Second}
	errCh := make(chan error, 1)
	go func() {
		err := s.httpSrv.ListenAndServe()
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
		_ = s.httpSrv.Shutdown(shutdownCtx)
		return nil
	case err := <-errCh:
		return err
	}
}

func (s *Server) Close(ctx context.Context) error {
	var errs []error
	if s.httpSrv != nil {
		if err := s.httpSrv.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	if s.runtime != nil {
		if err := s.runtime.Close(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (s *Server) LoadScripts(ctx context.Context) error {
	files, err := scriptFiles(s.cfg.ScriptDirs)
	if err != nil {
		return err
	}
	for _, file := range files {
		file := file
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read script %s: %w", file, err)
		}
		_, err = s.runtime.Owner.Call(ctx, "load-script", func(_ context.Context, vm *goja.Runtime) (any, error) {
			_, err := vm.RunScript(file, string(data))
			return nil, err
		})
		if err != nil {
			return fmt.Errorf("execute script %s: %w", file, err)
		}
	}
	return nil
}
