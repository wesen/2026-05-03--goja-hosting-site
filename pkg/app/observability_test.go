package app

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/goja-site/pkg/observability"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestServerHTTPMetrics(t *testing.T) {
	root := t.TempDir()
	scripts := writeSiteScript(t, root, `
		const express = require("express");
		const app = express.app();
		app.get("/", (req, res) => res.type("text/plain").send("ok"));
	`)
	obs := observability.New()
	srv, err := NewServer(Config{DBPath: filepath.Join(root, "app.db"), ScriptDirs: []string{scripts}, DBPolicy: DBPolicySimple, ReadOnly: true, SiteName: "bench", Observability: obs})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	defer func() { _ = srv.Close(context.Background()) }()

	if got := getServerBody(t, srv, "/"); got != "ok" {
		t.Fatalf("body = %q, want ok", got)
	}

	value := gatherCounter(t, obs.Registry, "goja_site_http_requests_total", map[string]string{
		"site":         "bench",
		"method":       http.MethodGet,
		"route":        "/",
		"status_class": "2xx",
	})
	if value != 1 {
		t.Fatalf("http request counter = %v, want 1", value)
	}
}

func TestServerDBMetrics(t *testing.T) {
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
	obs := observability.New()
	srv, err := NewServer(Config{DBPath: filepath.Join(root, "app.db"), ScriptDirs: []string{scripts}, DBPolicy: DBPolicyGuarded, SiteName: "dbsite", Observability: obs})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	defer func() { _ = srv.Close(context.Background()) }()

	if got := getServerBody(t, srv, "/"); got != "1" {
		t.Fatalf("body = %q, want 1", got)
	}

	if got := gatherCounter(t, obs.Registry, "goja_site_db_operations_total", map[string]string{"site": "dbsite", "db_policy": "guarded", "operation": "exec", "sql_kind": "insert"}); got != 1 {
		t.Fatalf("insert exec counter = %v, want 1", got)
	}
	if got := gatherCounter(t, obs.Registry, "goja_site_db_operations_total", map[string]string{"site": "dbsite", "db_policy": "guarded", "operation": "query", "sql_kind": "select"}); got != 1 {
		t.Fatalf("select query counter = %v, want 1", got)
	}
	if got := gatherCounter(t, obs.Registry, "goja_site_db_guard_checks_total", map[string]string{"site": "dbsite", "phase": "after_exec", "result": "skipped_no_limit"}); got < 1 {
		t.Fatalf("guard skipped_no_limit counter = %v, want >= 1", got)
	}
}

func TestServerKanbanMetrics(t *testing.T) {
	root := t.TempDir()
	scripts := writeSiteScript(t, root, `
		const express = require("express");
		const ui = require("ui.dsl");
		const kanban = require("kanban.dsl");
		const app = express.app();
		let cards = [{ id: 1, title: "One", status: "todo", position: 10 }];
		const board = kanban.board("metrics-board")
		  .columns(cols => cols.column("todo").title("To Do").done().column("done").title("Done").done())
		  .data(data => data.cards(() => cards).id(card => String(card.id)).column(card => card.status).position(card => card.position).searchText(card => card.title))
		  .render(render => render.card(card => ui.h3(card.title)))
		  .actions(actions => actions.cardMoved(event => { cards[0].status = event.to.columnId; return { ok: true, refresh: true }; }))
		  .build();
		board.mount(app, "/_kanban");
		app.get("/", (req, res) => res.html(board.render({ query: req.query })));
	`)
	obs := observability.New()
	srv, err := NewServer(Config{DBPath: filepath.Join(root, "app.db"), ScriptDirs: []string{scripts}, DBPolicy: DBPolicySimple, ReadOnly: true, SiteName: "kanban-site", Observability: obs})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	defer func() { _ = srv.Close(context.Background()) }()

	req := httptest.NewRequest(http.MethodGet, "http://example.test/_kanban/metrics-board/fragment", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("fragment status = %d body=%s", rr.Code, rr.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "http://example.test/_kanban/metrics-board/action/cardMoved", bytes.NewBufferString(`{"cardId":"1","to":{"columnId":"done","index":0}}`))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("action status = %d body=%s", rr.Code, rr.Body.String())
	}

	if got := gatherHistogramCount(t, obs.Registry, "goja_site_kanban_fragment_duration_seconds", map[string]string{"site": "kanban-site", "board": "metrics-board"}); got != 1 {
		t.Fatalf("fragment histogram count = %v, want 1", got)
	}
	if got := gatherHistogramCount(t, obs.Registry, "goja_site_kanban_action_duration_seconds", map[string]string{"site": "kanban-site", "board": "metrics-board", "action": "cardMoved", "refresh": "true"}); got != 1 {
		t.Fatalf("action histogram count = %v, want 1", got)
	}
	if got := gatherHistogramCount(t, obs.Registry, "goja_site_kanban_dispatch_duration_seconds", map[string]string{"site": "kanban-site", "board": "metrics-board", "action": "cardMoved"}); got != 1 {
		t.Fatalf("dispatch histogram count = %v, want 1", got)
	}
	if got := gatherHistogramCount(t, obs.Registry, "goja_site_kanban_render_duration_seconds", map[string]string{"site": "kanban-site", "board": "metrics-board", "reason": "action_refresh"}); got != 1 {
		t.Fatalf("action refresh render histogram count = %v, want 1", got)
	}
}

func TestMultiServerMetrics(t *testing.T) {
	root := t.TempDir()
	script := `
		const express = require("express");
		const app = express.app();
		app.get("/", (req, res) => res.type("text/plain").send("ok"));
	`
	trailScripts := writeSiteScript(t, filepath.Join(root, "trail"), script)
	obs := observability.New()
	srv, err := NewMultiServer(MultiConfig{DataDir: filepath.Join(root, "data"), Observability: obs, Sites: []SiteConfig{
		{Name: "trail", Host: "trail.example.test", ScriptDirs: []string{trailScripts}, DBPolicy: DBPolicySimple, ReadOnly: true},
	}})
	if err != nil {
		t.Fatalf("NewMultiServer() error = %v", err)
	}
	defer func() { _ = srv.Close(context.Background()) }()

	req := httptest.NewRequest(http.MethodGet, "http://trail.example.test/", nil)
	req.Host = "trail.example.test"
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("known host status = %d body=%s", rr.Code, rr.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "http://unknown.example.test/", nil)
	req.Host = "unknown.example.test"
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("unknown host status = %d", rr.Code)
	}

	if got := gatherGauge(t, obs.Registry, "goja_site_hosts_configured", map[string]string{"mode": "multi"}); got != 1 {
		t.Fatalf("hosts configured = %v, want 1", got)
	}
	if got := gatherGauge(t, obs.Registry, "goja_site_site_up", map[string]string{"site": "trail"}); got != 1 {
		t.Fatalf("site up = %v, want 1", got)
	}
	if got := gatherCounter(t, obs.Registry, "goja_site_unknown_host_requests_total", map[string]string{"host_class": "unknown"}); got != 1 {
		t.Fatalf("unknown host counter = %v, want 1", got)
	}
	if got := gatherCounter(t, obs.Registry, "goja_site_http_requests_total", map[string]string{"site": "trail", "method": http.MethodGet, "route": "/", "status_class": "2xx"}); got != 1 {
		t.Fatalf("site http counter = %v, want 1", got)
	}
}

func gatherCounter(t *testing.T, registry *prometheus.Registry, name string, labels map[string]string) float64 {
	t.Helper()
	metric := gatherMetric(t, registry, name, labels)
	if metric.GetCounter() == nil {
		t.Fatalf("metric %s with labels %v is not a counter", name, labels)
	}
	return metric.GetCounter().GetValue()
}

func gatherHistogramCount(t *testing.T, registry *prometheus.Registry, name string, labels map[string]string) uint64 {
	t.Helper()
	metric := gatherMetric(t, registry, name, labels)
	if metric.GetHistogram() == nil {
		t.Fatalf("metric %s with labels %v is not a histogram", name, labels)
	}
	return metric.GetHistogram().GetSampleCount()
}

func gatherGauge(t *testing.T, registry *prometheus.Registry, name string, labels map[string]string) float64 {
	t.Helper()
	metric := gatherMetric(t, registry, name, labels)
	if metric.GetGauge() == nil {
		t.Fatalf("metric %s with labels %v is not a gauge", name, labels)
	}
	return metric.GetGauge().GetValue()
}

func gatherMetric(t *testing.T, registry *prometheus.Registry, name string, labels map[string]string) *dto.Metric {
	t.Helper()
	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}
	for _, family := range families {
		if family.GetName() != name {
			continue
		}
		for _, metric := range family.GetMetric() {
			if metricHasLabels(metric, labels) {
				return metric
			}
		}
	}
	t.Fatalf("metric %s with labels %v not found", name, labels)
	return nil
}

func metricHasLabels(metric *dto.Metric, want map[string]string) bool {
	got := map[string]string{}
	for _, pair := range metric.GetLabel() {
		got[pair.GetName()] = pair.GetValue()
	}
	for key, value := range want {
		if got[key] != value {
			return false
		}
	}
	return true
}
