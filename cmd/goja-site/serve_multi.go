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

type serveMultiCommand struct{ *cmds.CommandDescription }

var _ cmds.BareCommand = (*serveMultiCommand)(nil)

type serveMultiSettings struct {
	Config          string  `glazed:"config"`
	MetricsAddr     string  `glazed:"metrics-addr"`
	MetricsPath     string  `glazed:"metrics-path"`
	Pprof           bool    `glazed:"pprof"`
	OtelEnabled     bool    `glazed:"otel-enabled"`
	OtelEndpoint    string  `glazed:"otel-endpoint"`
	OtelSampleRatio float64 `glazed:"otel-sample-ratio"`
	ServiceName     string  `glazed:"service-name"`
}

func newServeMultiCommand() (*serveMultiCommand, error) {
	desc := cmds.NewCommandDescription(
		"serve-multi",
		cmds.WithShort("Serve multiple trusted JavaScript websites by Host header"),
		cmds.WithLong(`Serve multiple small websites from one process.

The config file declares one or more sites. Each site gets its own scripts
directory, SQLite database, Goja runtime, Express-style route host, ui.dsl,
kanban.dsl, and db.guard instance. The outer HTTP server dispatches requests by
Host header.

Example:
  goja-site serve-multi --config deploy/sites.local.yaml`),
		cmds.WithFlags(
			fields.New("config", fields.TypeString, fields.WithDefault("./deploy/sites.yaml"), fields.WithHelp("YAML or JSON multi-site config path")),
			fields.New("metrics-addr", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Private diagnostics bind address for Prometheus metrics (disabled when empty)")),
			fields.New("metrics-path", fields.TypeString, fields.WithDefault("/metrics"), fields.WithHelp("Prometheus metrics path on the diagnostics listener")),
			fields.New("pprof", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Expose pprof handlers on the private diagnostics listener (requires --metrics-addr)")),
			fields.New("otel-enabled", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Enable OpenTelemetry tracing with OTLP HTTP export")),
			fields.New("otel-endpoint", fields.TypeString, fields.WithDefault("http://127.0.0.1:4318/v1/traces"), fields.WithHelp("OpenTelemetry OTLP HTTP traces endpoint")),
			fields.New("otel-sample-ratio", fields.TypeFloat, fields.WithDefault(0.01), fields.WithHelp("OpenTelemetry trace sample ratio between 0 and 1")),
			fields.New("service-name", fields.TypeString, fields.WithDefault("goja-site"), fields.WithHelp("OpenTelemetry service.name resource attribute")),
		),
	)
	return &serveMultiCommand{CommandDescription: desc}, nil
}

func (c *serveMultiCommand) Run(ctx context.Context, vals *values.Values) error {
	settings := serveMultiSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, &settings); err != nil {
		return err
	}
	if settings.Config == "" {
		settings.Config = "./deploy/sites.yaml"
	}
	cfg, err := app.LoadMultiConfig(settings.Config)
	if err != nil {
		return err
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
	cfg.Observability = obs

	srv, err := app.NewMultiServer(cfg)
	if err != nil {
		return err
	}
	defer func() { _ = srv.Close(context.Background()) }()

	fmt.Printf("goja-site serving multi addr=%s config=%s hosts=%s metricsAddr=%s pprof=%v otel=%v\n", cfg.Addr, settings.Config, srv.SiteSummary(), settings.MetricsAddr, settings.Pprof, settings.OtelEnabled)
	return srv.Run(ctx)
}
