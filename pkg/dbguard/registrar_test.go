package dbguard_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-go-golems/go-go-goja/engine"
	databasemod "github.com/go-go-golems/go-go-goja/modules/database"
	"github.com/go-go-golems/goja-site/pkg/dbguard"
	_ "github.com/mattn/go-sqlite3"
)

func TestDBGuardCallbackRunsThroughDatabaseModule(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.db")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	guard := dbguard.New(db, path)
	metered := dbguard.NewMeteredDB(db, guard)
	databaseModule := databasemod.New(databasemod.WithPreconfiguredDB(metered), databasemod.WithConfigureEnabled(false))
	factory, err := engine.NewBuilder().
		WithModules(engine.NativeModuleSpec{ModuleID: "database:test", ModuleName: databaseModule.Name(), Loader: databaseModule.Loader}).
		WithRuntimeModuleRegistrars(dbguard.NewRegistrar(guard)).
		Build()
	if err != nil {
		t.Fatalf("build factory: %v", err)
	}
	rt, err := factory.NewRuntime(context.Background())
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	defer rt.Close(context.Background())

	value, err := rt.VM.RunString(`
		const db = require("database");
		const guard = require("db.guard");
		let calls = 0;
		guard.configure({ maxBytes: 1, checkEveryWrites: 1, cooldownMs: 1 });
		guard.onLimitExceeded(event => {
		  calls++;
		  db.exec("DELETE FROM cards WHERE id < 0"); // should not recurse
		  return { ok: true, calls, overByBytes: event.overByBytes };
		});
		db.exec("CREATE TABLE cards(id INTEGER PRIMARY KEY, title TEXT)");
		db.exec("INSERT INTO cards(title) VALUES (?)", "one");
		const last = guard.lastResult();
		JSON.stringify({ calls, callbackCalled: last.callbackCalled, stillOverLimit: last.stillOverLimit, hasStats: guard.stats().totalBytes > 0 });
	`)
	if err != nil {
		t.Fatalf("run JS: %v", err)
	}
	got := value.String()
	if got == "" || got == "null" {
		t.Fatalf("empty result")
	}
	if want := `"callbackCalled":true`; !strings.Contains(got, want) {
		t.Fatalf("result missing %s: %s", want, got)
	}
	if want := `"hasStats":true`; !strings.Contains(got, want) {
		t.Fatalf("result missing %s: %s", want, got)
	}
}

func TestDBGuardHardLimitSurfacesThroughDatabaseModule(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.db")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec("CREATE TABLE cards(id INTEGER PRIMARY KEY, title TEXT)"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	guard := dbguard.New(db, path)
	metered := dbguard.NewMeteredDB(db, guard)
	databaseModule := databasemod.New(databasemod.WithPreconfiguredDB(metered), databasemod.WithConfigureEnabled(false))
	factory, err := engine.NewBuilder().
		WithModules(engine.NativeModuleSpec{ModuleID: "database:test", ModuleName: databaseModule.Name(), Loader: databaseModule.Loader}).
		WithRuntimeModuleRegistrars(dbguard.NewRegistrar(guard)).
		Build()
	if err != nil {
		t.Fatalf("build factory: %v", err)
	}
	rt, err := factory.NewRuntime(context.Background())
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	defer rt.Close(context.Background())

	value, err := rt.VM.RunString(`
		const db = require("database");
		const guard = require("db.guard");
		guard.configure({ maxBytes: 1, hardMaxBytes: 1, failWritesOverHardLimit: true, checkEveryWrites: 1 });
		try {
		  db.exec("INSERT INTO cards(title) VALUES ('blocked')");
		  "missing error";
		} catch (e) {
		  String(e).includes("sqlite hard limit exceeded") ? "blocked" : String(e);
		}
	`)
	if err != nil {
		t.Fatalf("run JS: %v", err)
	}
	if got := value.String(); got != "blocked" {
		t.Fatalf("got %q", got)
	}
}
