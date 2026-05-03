package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/engine"
	databasemod "github.com/go-go-golems/go-go-goja/modules/database"
	"github.com/go-go-golems/goja-site/pkg/dbguard"
	"github.com/go-go-golems/goja-site/pkg/kanbanddsl"
	"github.com/go-go-golems/goja-site/pkg/uidsl"
	"github.com/go-go-golems/goja-site/pkg/web"
	_ "github.com/mattn/go-sqlite3"
)

// Server owns the database, goja runtime, route host, and HTTP server.
type Server struct {
	cfg     Config
	db      *sql.DB
	runtime *engine.Runtime
	host    *web.Host
	httpSrv *http.Server
}

func NewServer(cfg Config) (*Server, error) {
	if cfg.Addr == "" {
		cfg.Addr = ":8080"
	}
	if cfg.DBPath == "" {
		cfg.DBPath = "./app.db"
	}
	if cfg.ScriptsDir == "" {
		cfg.ScriptsDir = "./scripts"
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

	host := web.NewHost(web.HostOptions{Dev: cfg.Dev, Renderer: uidsl.RenderAny})
	guard := dbguard.New(db, cfg.DBPath)
	meteredDB := dbguard.NewMeteredDB(db, guard)
	databaseModule := databasemod.New(
		databasemod.WithPreconfiguredDB(meteredDB),
		databasemod.WithConfigureEnabled(false),
	)
	dbAliasModule := databasemod.New(
		databasemod.WithName("db"),
		databasemod.WithPreconfiguredDB(meteredDB),
		databasemod.WithConfigureEnabled(false),
	)

	factory, err := engine.NewBuilder().
		WithModules(
			engine.NativeModuleSpec{ModuleID: "database:app", ModuleName: databaseModule.Name(), Loader: databaseModule.Loader},
			engine.NativeModuleSpec{ModuleID: "database:db-alias", ModuleName: dbAliasModule.Name(), Loader: dbAliasModule.Loader},
		).
		UseModuleMiddleware(engine.MiddlewareOnly("fs", "path", "time", "timer")).
		WithRuntimeModuleRegistrars(web.NewExpressRegistrar(host), uidsl.NewRegistrar(), kanbanddsl.NewRegistrar(), dbguard.NewRegistrar(guard)).
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
	return s.host
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.host.ServeHTTP(w, r)
}

func (s *Server) Run(ctx context.Context) error {
	s.httpSrv = &http.Server{Addr: s.cfg.Addr, Handler: s.host, ReadHeaderTimeout: 5 * time.Second}
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
	files, err := scriptFiles(s.cfg.ScriptsDir)
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

func scriptFiles(dir string) ([]string, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("stat scripts directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("scripts path %s is not a directory", dir)
	}
	var files []string
	if err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".js") {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk scripts directory: %w", err)
	}
	sort.Strings(files)
	return files, nil
}
