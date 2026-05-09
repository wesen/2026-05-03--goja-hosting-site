const db = require("db");
const express = require("express");
const fs = require("fs");
const path = require("path");
const yaml = require("yaml");
const ui = require("ui.dsl");

const app = express.app();
app.get("/favicon.ico", (req, res) => res.status(204).end());

const specPath = path.join("examples", "db-browser", "yaml-dashboard", "dashboard.yaml");
const retroCSS = `
:root{--ink:#121212;--paper:#f3efe4;--panel:#fffdf5;--line:#121212;--shadow:#989287;--accent:#758d86;--accent2:#8e7a8f}body{margin:0;background:var(--paper);color:var(--ink);font-family:"Chicago","Geneva",system-ui,sans-serif;font-size:14px;background-image:radial-gradient(rgba(0,0,0,.08) 1px,transparent 1px);background-size:6px 6px}.desktop{max-width:1080px;margin:22px auto;padding:0 16px}.window{background:var(--panel);border:2px solid var(--line);box-shadow:6px 6px 0 var(--shadow)}.titlebar{padding:5px 8px;border-bottom:2px solid var(--line);font-weight:900;background:repeating-linear-gradient(0deg,#fffdf5 0 2px,#ddd8cc 2px 4px)}.content{padding:16px}.metric-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(160px,1fr));gap:10px;margin:12px 0}.metric{border:2px solid var(--line);background:#fffdf8;padding:12px;box-shadow:3px 3px 0 #c5beb2;border-top:10px solid var(--accent)}.metric:nth-child(even){border-top-color:var(--accent2)}.metric strong{display:block;font-size:28px}.ui-table-filters{display:flex;gap:8px;flex-wrap:wrap;align-items:end;border:2px solid var(--line);background:#e5dfd4;padding:9px;margin:12px 0}.ui-table-filters label{display:flex;flex-direction:column;font-size:12px;font-weight:900}input,button{font:inherit;border:2px solid var(--line);background:#fffdf8;padding:4px 6px;box-shadow:2px 2px 0 var(--shadow)}button{background:#ddd8cc;font-weight:900}table.ui-table{width:100%;border:2px solid var(--line);border-collapse:separate;border-spacing:0;background:#fffdf8}th,td{border-right:1px solid var(--line);border-bottom:1px solid var(--line);padding:7px 8px;text-align:left}th{background:#ddd8cc}.align-right{text-align:right}.ui-badge,.ui-tag{display:inline-block;border:1px solid var(--line);padding:1px 5px;background:#d8e2dc;font-size:12px;font-weight:800}a{color:#172f44;font-weight:800}.muted{color:#504c46}.ui-table-pagination{margin-top:8px;border:2px solid var(--line);display:inline-block;padding:5px 8px;background:#fffdf8;box-shadow:3px 3px 0 var(--shadow)}
`;

function page(title, ...children) {
  return ui.page({ title }, ui.style(ui.raw(retroCSS)), ui.main({ class: "desktop" }, ui.section({ class: "window" }, ui.div({ class: "titlebar" }, title), ui.div({ class: "content" }, children))));
}

function loadSpec() {
  return yaml.parse(fs.readFileSync(specPath, "utf8"));
}

function metricRows(spec) {
  return (spec.metrics || []).map(metric => {
    const row = db.query(metric.sql)[0] || { value: null };
    return { metric: metric.label, value: row.value, group: metric.group || "default", tone: metric.tone || "neutral" };
  });
}

app.get("/", (req, res) => {
  const spec = loadSpec();
  const rows = metricRows(spec);
  res.html(page(spec.title || "YAML Dashboard",
    ui.h1(spec.title || "YAML Dashboard"),
    ui.p({ class: "muted" }, "Metrics are declared in dashboard.yaml and rendered through the yaml module."),
    ui.div({ class: "metric-grid" }, rows.map(row => ui.div({ class: "metric" }, ui.small(row.group), ui.strong(row.value), ui.span(row.metric)))),
    ui.table.fromRows("metrics", rows)
      .columns(c => c
        .text("metric").label("Metric").sortable().filterable()
        .text("group").label("Group").sortable().filterable()
        .badge("tone").label("Tone").filterable()
        .text("value").label("Value").align("right").sortable()
      )
      .features(f => f.filters().pagination({ size: 10 }).sorting())
      .render({ query: req.query })
  ));
});
