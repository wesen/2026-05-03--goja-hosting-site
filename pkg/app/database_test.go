package app

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestSimplePolicyReadOnlyRejectsWrites(t *testing.T) {
	root := t.TempDir()
	dbPath := filepath.Join(root, "app.db")
	seedSQLite(t, dbPath, `CREATE TABLE items(id INTEGER PRIMARY KEY, name TEXT); INSERT INTO items(name) VALUES ('seed');`)
	scripts := writeSiteScript(t, root, `
		const db = require("database");
		const express = require("express");
		const app = express.app();
		app.get("/", (req, res) => {
		  let writeError = "";
		  try {
		    db.exec("INSERT INTO items(name) VALUES ('blocked')");
		  } catch (err) {
		    writeError = String(err && err.message ? err.message : err);
		  }
		  const rows = db.query("SELECT COUNT(*) AS count FROM items");
		  res.type("text/plain").send(writeError + "|count=" + rows[0].count);
		});
	`)

	srv, err := NewServer(Config{DBPath: dbPath, ScriptDirs: []string{scripts}, DBPolicy: DBPolicySimple, ReadOnly: true})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	defer func() { _ = srv.Close(context.Background()) }()

	body := getServerBody(t, srv, "/")
	if !strings.Contains(body, "database writes are disabled") {
		t.Fatalf("expected write rejection, got %q", body)
	}
	if !strings.Contains(body, "count=1") {
		t.Fatalf("expected original row count, got %q", body)
	}
}

func TestSimplePolicyReadOnlyQueryRejectsMutatingPragmaAndWith(t *testing.T) {
	root := t.TempDir()
	dbPath := filepath.Join(root, "app.db")
	seedSQLite(t, dbPath, `CREATE TABLE items(id INTEGER PRIMARY KEY, name TEXT); INSERT INTO items(name) VALUES ('seed');`)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	defer func() { _ = db.Close() }()

	readOnly := &simpleDB{db: db}
	for _, query := range []string{
		"PRAGMA user_version=123",
		"WITH doomed AS (SELECT id FROM items) DELETE FROM items WHERE id IN (SELECT id FROM doomed) RETURNING id",
	} {
		if rows, err := readOnly.Query(query); err == nil {
			_ = rows.Close()
			t.Fatalf("Query(%q) unexpectedly succeeded", query)
		} else if !strings.Contains(err.Error(), "database writes are disabled") {
			t.Fatalf("Query(%q) error = %v, want writes disabled", query, err)
		}
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM items").Scan(&count); err != nil {
		t.Fatalf("count items: %v", err)
	}
	if count != 1 {
		t.Fatalf("item count after rejected queries = %d, want 1", count)
	}
}

func TestSimplePolicyReadOnlyQueryAllowsReadPragma(t *testing.T) {
	root := t.TempDir()
	dbPath := filepath.Join(root, "app.db")
	seedSQLite(t, dbPath, `CREATE TABLE items(id INTEGER PRIMARY KEY, name TEXT);`)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	defer func() { _ = db.Close() }()

	rows, err := (&simpleDB{db: db}).Query("PRAGMA table_info(items)")
	if err != nil {
		t.Fatalf("read pragma rejected: %v", err)
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		t.Fatalf("expected table_info rows")
	}
}

func TestSimplePolicyAllowWrites(t *testing.T) {
	root := t.TempDir()
	scripts := writeSiteScript(t, root, `
		const db = require("database");
		const express = require("express");
		const app = express.app();
		db.exec("CREATE TABLE IF NOT EXISTS visits(id INTEGER PRIMARY KEY AUTOINCREMENT)");
		app.get("/", (req, res) => {
		  db.exec("INSERT INTO visits DEFAULT VALUES");
		  const rows = db.query("SELECT COUNT(*) AS count FROM visits");
		  res.type("text/plain").send(String(rows[0].count));
		});
	`)

	srv, err := NewServer(Config{DBPath: filepath.Join(root, "app.db"), ScriptDirs: []string{scripts}, DBPolicy: DBPolicySimple, AllowWrites: true})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	defer func() { _ = srv.Close(context.Background()) }()

	if got := getServerBody(t, srv, "/"); got != "1" {
		t.Fatalf("first visit = %q, want 1", got)
	}
	if got := getServerBody(t, srv, "/"); got != "2" {
		t.Fatalf("second visit = %q, want 2", got)
	}
}

func TestGuardedPolicyRegistersDBGuard(t *testing.T) {
	root := t.TempDir()
	scripts := writeSiteScript(t, root, `
		const guard = require("db.guard");
		const express = require("express");
		const app = express.app();
		app.get("/", (req, res) => {
		  res.type("text/plain").send(typeof guard.configure + ":" + typeof guard.stats);
		});
	`)

	srv, err := NewServer(Config{DBPath: filepath.Join(root, "app.db"), ScriptDirs: []string{scripts}, DBPolicy: DBPolicyGuarded})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	defer func() { _ = srv.Close(context.Background()) }()

	if got := getServerBody(t, srv, "/"); got != "function:function" {
		t.Fatalf("db.guard exports = %q, want function:function", got)
	}
}

func TestNormalizeDBPolicyConfig(t *testing.T) {
	cfg := Config{}
	if err := normalizeDBPolicyConfig(&cfg); err != nil {
		t.Fatalf("normalize default: %v", err)
	}
	if cfg.DBPolicy != DBPolicyGuarded {
		t.Fatalf("default policy = %q, want %q", cfg.DBPolicy, DBPolicyGuarded)
	}

	cfg = Config{DBPolicy: DBPolicySimple}
	if err := normalizeDBPolicyConfig(&cfg); err != nil {
		t.Fatalf("normalize simple: %v", err)
	}
	if !cfg.ReadOnly {
		t.Fatalf("simple policy without allow-writes should normalize to read-only")
	}

	cfg = Config{DBPolicy: DBPolicySimple, AllowWrites: true}
	if err := normalizeDBPolicyConfig(&cfg); err != nil {
		t.Fatalf("normalize simple allow writes: %v", err)
	}
	if cfg.ReadOnly {
		t.Fatalf("simple policy with allow-writes should not force read-only")
	}

	cfg = Config{DBPolicy: DBPolicy("nope")}
	if err := normalizeDBPolicyConfig(&cfg); err == nil || !strings.Contains(err.Error(), "unsupported database policy") {
		t.Fatalf("expected unsupported policy error, got %v", err)
	}
}

func getServerBody(t *testing.T, srv *Server, path string) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "http://example.test"+path, nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET %s status = %d body=%s", path, rr.Code, rr.Body.String())
	}
	return strings.TrimSpace(rr.Body.String())
}

func seedSQLite(t *testing.T, path, statements string) {
	t.Helper()
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open sqlite seed db: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.Exec(statements); err != nil {
		t.Fatalf("seed sqlite db: %v", err)
	}
}
