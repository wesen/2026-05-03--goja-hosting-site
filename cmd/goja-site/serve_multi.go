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
)

type serveMultiCommand struct{ *cmds.CommandDescription }

var _ cmds.BareCommand = (*serveMultiCommand)(nil)

type serveMultiSettings struct {
	Config string `glazed:"config"`
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

	srv, err := app.NewMultiServer(cfg)
	if err != nil {
		return err
	}
	defer func() { _ = srv.Close(context.Background()) }()

	fmt.Printf("goja-site serving multi addr=%s config=%s hosts=%s\n", cfg.Addr, settings.Config, srv.SiteSummary())
	return srv.Run(ctx)
}
