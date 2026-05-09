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

type serveCommand struct{ *cmds.CommandDescription }

var _ cmds.BareCommand = (*serveCommand)(nil)

type serveSettings struct {
	Addr        string   `glazed:"addr"`
	DBPath      string   `glazed:"db"`
	ScriptDirs  []string `glazed:"scripts"`
	Dev         bool     `glazed:"dev"`
	DBPolicy    string   `glazed:"db-policy"`
	ReadOnly    bool     `glazed:"readonly"`
	AllowWrites bool     `glazed:"allow-writes"`
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

	srv, err := app.NewServer(app.Config{Addr: settings.Addr, DBPath: settings.DBPath, ScriptDirs: settings.ScriptDirs, Dev: settings.Dev, DBPolicy: app.DBPolicy(settings.DBPolicy), ReadOnly: settings.ReadOnly, AllowWrites: settings.AllowWrites})
	if err != nil {
		return err
	}
	defer func() { _ = srv.Close(context.Background()) }()

	fmt.Printf("goja-site serving addr=%s db=%s scripts=%v dbPolicy=%s readonly=%v allowWrites=%v\n", settings.Addr, settings.DBPath, settings.ScriptDirs, settings.DBPolicy, settings.ReadOnly, settings.AllowWrites)
	return srv.Run(ctx)
}
