package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSiteScript(t *testing.T, dir, body string) string {
	t.Helper()
	scripts := filepath.Join(dir, "scripts")
	if err := os.MkdirAll(scripts, 0o755); err != nil {
		t.Fatalf("mkdir scripts: %v", err)
	}
	if err := os.WriteFile(filepath.Join(scripts, "app.js"), []byte(body), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}
	return scripts
}

func TestMultiConfigNormalizeRejectsDuplicates(t *testing.T) {
	cfg := MultiConfig{DataDir: t.TempDir(), BaseDomain: "kanban.example.test", Sites: []SiteConfig{
		{Name: "trail", ScriptDirs: []string{"./a"}},
		{Name: "trail", ScriptDirs: []string{"./b"}},
	}}
	if err := cfg.Normalize(); err == nil || !strings.Contains(err.Error(), "duplicate site name") {
		t.Fatalf("expected duplicate site name error, got %v", err)
	}

	cfg = MultiConfig{DataDir: t.TempDir(), Sites: []SiteConfig{
		{Name: "trail", Host: "same.example.test", ScriptDirs: []string{"./a"}},
		{Name: "crm", Host: "same.example.test", ScriptDirs: []string{"./b"}},
	}}
	if err := cfg.Normalize(); err == nil || !strings.Contains(err.Error(), "duplicate site host") {
		t.Fatalf("expected duplicate site host error, got %v", err)
	}
}

func TestMultiServerRoutesByHostAndIsolatesDBs(t *testing.T) {
	root := t.TempDir()
	script := `
		const db = require("database");
		const express = require("express");
		const app = express.app();
		db.exec("CREATE TABLE IF NOT EXISTS visits(id INTEGER PRIMARY KEY AUTOINCREMENT)");
		app.get("/", (req, res) => {
		  db.exec("INSERT INTO visits DEFAULT VALUES");
		  const rows = db.query("SELECT COUNT(*) AS count FROM visits");
		  res.type("text/plain").send(String(rows[0].count));
		});
	`
	trailScripts := writeSiteScript(t, filepath.Join(root, "trail"), script)
	crmScripts := writeSiteScript(t, filepath.Join(root, "crm"), script)
	cfg := MultiConfig{Addr: ":0", DataDir: filepath.Join(root, "data"), Sites: []SiteConfig{
		{Name: "trail", Host: "trail.kanban.example.test", ScriptDirs: []string{trailScripts}},
		{Name: "crm", Host: "crm.kanban.example.test", ScriptDirs: []string{crmScripts}},
	}}
	srv, err := NewMultiServer(cfg)
	if err != nil {
		t.Fatalf("new multi server: %v", err)
	}
	defer srv.Close(context.Background())

	request := func(host string) string {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, "http://"+host+"/", nil)
		req.Host = host
		rr := httptest.NewRecorder()
		srv.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("%s status = %d body=%s", host, rr.Code, rr.Body.String())
		}
		return strings.TrimSpace(rr.Body.String())
	}

	if got := request("trail.kanban.example.test"); got != "1" {
		t.Fatalf("trail first count = %q", got)
	}
	if got := request("trail.kanban.example.test"); got != "2" {
		t.Fatalf("trail second count = %q", got)
	}
	if got := request("crm.kanban.example.test"); got != "1" {
		t.Fatalf("crm first count = %q", got)
	}

	req := httptest.NewRequest(http.MethodGet, "http://unknown.example.test/", nil)
	req.Host = "unknown.example.test"
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("unknown status = %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "http://127.0.0.1/healthz", nil)
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || strings.TrimSpace(rr.Body.String()) != "ok" {
		t.Fatalf("healthz = %d %q", rr.Code, rr.Body.String())
	}
}
