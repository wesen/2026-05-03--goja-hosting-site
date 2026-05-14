package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/goja-site/pkg/app"
	"github.com/go-go-golems/goja-site/pkg/observability"
)

type serveCommand struct{ *cmds.CommandDescription }

var _ cmds.BareCommand = (*serveCommand)(nil)

type serveSettings struct {
	Addr            string   `glazed:"addr"`
	DBPath          string   `glazed:"db"`
	ScriptDirs      []string `glazed:"scripts"`
	Dev             bool     `glazed:"dev"`
	DBPolicy        string   `glazed:"db-policy"`
	ReadOnly        bool     `glazed:"readonly"`
	AllowWrites     bool     `glazed:"allow-writes"`
	MetricsAddr     string   `glazed:"metrics-addr"`
	MetricsPath     string   `glazed:"metrics-path"`
	Pprof           bool     `glazed:"pprof"`
	OtelEnabled     bool     `glazed:"otel-enabled"`
	OtelEndpoint    string   `glazed:"otel-endpoint"`
	OtelSampleRatio float64  `glazed:"otel-sample-ratio"`
	ServiceName     string   `glazed:"service-name"`
}

func newServeCommand() (*serveCommand, error) {
	desc := cmds.NewCommandDescription(
		"serve",
		cmds.WithShort("Serve a trusted JavaScript website"),
		cmds.WithLong(`Serve a small website written in JavaScript.

The server creates one go-go-goja runtime, exposes a preconfigured SQLite database
module, exposes the fs module, registers an Express-style HTTP module, registers
an HTML UI DSL module, loads all scripts from --scripts, and dispatches HTTP
requests to the routes registered by those scripts.

Example:
  goja-site serve --db examples/kanban/kanban.db --scripts examples/kanban/scripts --addr :8080
  goja-site serve --db data.sqlite --scripts examples/db-browser/generic-browser/scripts --db-policy simple --readonly`),
		cmds.WithFlags(
			fields.New("addr", fields.TypeString, fields.WithDefault(":8080"), fields.WithHelp("HTTP bind address")),
			fields.New("db", fields.TypeString, fields.WithDefault("./app.db"), fields.WithHelp("SQLite database path")),
			fields.New("scripts", fields.TypeStringList, fields.WithHelp("Directory containing JavaScript files to load (repeatable; defaults to ./scripts)")),
			fields.New("db-policy", fields.TypeChoice, fields.WithDefault(string(app.DBPolicyGuarded)), fields.WithChoices(string(app.DBPolicyGuarded), string(app.DBPolicySimple)), fields.WithHelp("Database policy: guarded exposes db.guard; simple uses a read/write gate")),
			fields.New("readonly", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Disable writes for --db-policy simple")),
			fields.New("allow-writes", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Allow writes for --db-policy simple when --readonly is false")),
			fields.New("dev", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Show detailed development errors in HTTP responses")),
			fields.New("metrics-addr", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Private diagnostics bind address for Prometheus metrics (disabled when empty)")),
			fields.New("metrics-path", fields.TypeString, fields.WithDefault("/metrics"), fields.WithHelp("Prometheus metrics path on the diagnostics listener")),
			fields.New("pprof", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Expose pprof handlers on the private diagnostics listener (requires --metrics-addr)")),
			fields.New("otel-enabled", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Enable OpenTelemetry tracing with OTLP HTTP export")),
			fields.New("otel-endpoint", fields.TypeString, fields.WithDefault("http://127.0.0.1:4318/v1/traces"), fields.WithHelp("OpenTelemetry OTLP HTTP traces endpoint")),
			fields.New("otel-sample-ratio", fields.TypeFloat, fields.WithDefault(0.01), fields.WithHelp("OpenTelemetry trace sample ratio between 0 and 1")),
			fields.New("service-name", fields.TypeString, fields.WithDefault("goja-site"), fields.WithHelp("OpenTelemetry service.name resource attribute")),
		),
	)
	return &serveCommand{CommandDescription: desc}, nil
}

func (c *serveCommand) Run(ctx context.Context, vals *values.Values) error {
	settings := serveSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, &settings); err != nil {
		return err
	}
	if settings.Addr == "" {
		settings.Addr = ":8080"
	}
	if settings.DBPath == "" {
		settings.DBPath = "./app.db"
	}
	if len(settings.ScriptDirs) == 0 {
		settings.ScriptDirs = []string{"./scripts"}
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var obs *observability.Observability
	if settings.MetricsAddr != "" || settings.Pprof || settings.OtelEnabled {
		obs = observability.New()
	}
	tracing, err := observability.InitTracing(ctx, observability.TracingConfig{Enabled: settings.OtelEnabled, ServiceName: settings.ServiceName, Endpoint: settings.OtelEndpoint, SampleRatio: settings.OtelSampleRatio})
	if err != nil {
		return err
	}
	defer func() { _ = tracing.Shutdown(context.Background()) }()
	if obs != nil {
		obs.EnableTracing(tracing)
	}
	diagnostics, err := observability.StartDiagnostics(ctx, observability.Config{MetricsAddr: settings.MetricsAddr, MetricsPath: settings.MetricsPath, EnablePprof: settings.Pprof}, obs)
	if err != nil {
		return err
	}
	defer func() { _ = diagnostics.Close(context.Background()) }()

	srv, err := app.NewServer(app.Config{Addr: settings.Addr, DBPath: settings.DBPath, ScriptDirs: settings.ScriptDirs, Dev: settings.Dev, DBPolicy: app.DBPolicy(settings.DBPolicy), ReadOnly: settings.ReadOnly, AllowWrites: settings.AllowWrites, SiteName: "default", Observability: obs})
	if err != nil {
		return err
	}
	defer func() { _ = srv.Close(context.Background()) }()
	if obs != nil && obs.Multi != nil {
		obs.Multi.SetHostsConfigured("single", 1)
		obs.Multi.SetSiteUp("default", true)
	}

	fmt.Printf("goja-site serving addr=%s db=%s scripts=%v dbPolicy=%s readonly=%v allowWrites=%v metricsAddr=%s pprof=%v otel=%v\n", settings.Addr, settings.DBPath, settings.ScriptDirs, settings.DBPolicy, settings.ReadOnly, settings.AllowWrites, settings.MetricsAddr, settings.Pprof, settings.OtelEnabled)
	return srv.Run(ctx)
}
