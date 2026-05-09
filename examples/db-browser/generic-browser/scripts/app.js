const db = require("db");
const express = require("express");
const ui = require("ui.dsl");

const app = express.app();
app.get("/favicon.ico", (req, res) => res.status(204).end());

const retroCSS = `
:root{--ink:#050505;--paper:#f5f5f5;--panel:#fff;--line:#050505;--shade:#d8d8d8;--shade2:#b9b9b9;--accent:#1d4d9f;--muted:#666;--green:#dfe8df;--red:#ead9d4;--blue:#dfe4ef}*{box-sizing:border-box}body{margin:0;background:#f7f7f7;color:var(--ink);font-family:"Chicago","Geneva","Monaco",system-ui,sans-serif;font-size:13px;background-image:radial-gradient(#ddd 1px,transparent 1px);background-size:5px 5px}.system-menubar{height:28px;background:#fff;border-bottom:2px solid #000;display:flex;align-items:center;justify-content:space-between;padding:0 14px;font-weight:900;font-size:15px}.system-menubar__menus{display:flex;gap:24px}.system-menubar__apple{font-size:18px}.system-menubar__right{display:flex;gap:14px}.desktop{max-width:1220px;margin:12px auto 24px;padding:0 14px}.window{background:var(--panel);border:2px solid var(--line);box-shadow:0 0 0 1px #777;margin-bottom:18px}.titlebar{height:24px;border-bottom:2px solid var(--line);display:grid;grid-template-columns:1fr auto 1fr;align-items:center;background:repeating-linear-gradient(0deg,#fff 0 2px,#bbb 2px 3px,#fff 3px 5px);font-weight:900;text-align:center}.titlebar:before{content:"□";justify-self:start;margin-left:8px;background:#fff}.titlebar:after{content:"↗";justify-self:end;margin-right:8px;background:#fff;border:1px solid #000;padding:0 2px}.titlebar span{background:#fff;padding:0 10px}.hero{display:flex;align-items:center;justify-content:space-between;gap:20px;border-bottom:2px solid var(--line);padding:26px 24px}.hero__left{display:flex;align-items:center;gap:18px}.db-icon{font-size:38px;line-height:1}.hero h1{font-size:28px;margin:0 0 8px}.content{padding:14px}.top-tabs{display:flex;gap:3px;margin:12px 0 0}.top-tabs span{border:2px solid #000;border-bottom:0;border-radius:4px 4px 0 0;background:#f8f8f8;padding:6px 12px;box-shadow:2px 0 0 #aaa}.top-tabs .active{font-weight:900;background:#fff}.toolbar,.ui-table-filters{display:flex;flex-wrap:wrap;gap:10px;align-items:center;border:2px solid var(--line);background:#eee;padding:8px;margin:10px 0}.ui-table-filters label{font-weight:900}.card-grid{display:grid;grid-template-columns:repeat(3,1fr);gap:14px;margin-bottom:10px}.card{border:2px solid var(--line);background:#fff;padding:12px;min-height:72px}.card small{text-transform:uppercase}.card strong{font-size:26px;display:block;margin-top:8px}input,select,button{font:inherit;border:2px solid #111;background:#fff;padding:3px 7px;box-shadow:2px 2px 0 #bbb}button{font-weight:900;background:#f8f8f8}table.ui-table{width:100%;border-collapse:separate;border-spacing:0;background:#fff;border:2px solid #111}th,td{padding:7px 8px;border-right:1px solid #111;border-bottom:1px solid #111;text-align:left;vertical-align:top}th{font-weight:900;background:#fff}.align-right{text-align:right}.ui-badge,.ui-tag{display:inline-block;border:1px solid #111;border-radius:2px;padding:1px 6px;background:#fff;font-size:12px;font-weight:900}.ui-badge--info{background:var(--blue)}.ui-badge--success{background:var(--green)}.ui-badge--warning{background:#eee4d0}.ui-badge--danger{background:var(--red)}.ui-badge--muted{background:#eee}.ui-table-pagination{margin-top:8px;display:inline-block;border:2px solid #111;background:#fff;padding:5px 8px;box-shadow:3px 3px 0 #aaa}a{color:#0000ee;font-weight:900}.muted{color:#333}.danger{border-left-color:#000}.ui-codeblock{border:2px solid #111;background:#fff;box-shadow:3px 3px 0 #aaa;margin:10px 0}.ui-codeblock__caption{display:flex;justify-content:space-between;align-items:center;border-bottom:2px solid #111;background:#eee;padding:4px 8px;font-weight:900}.ui-codeblock__copy{font-size:12px;padding:2px 7px}.ui-codeblock__pre,.ui-codeblock.ui-codeblock--wrap,.ui-codeblock.ui-codeblock--nowrap{margin:0;padding:9px 12px;overflow:auto}.ui-codeblock--wrap code{white-space:pre-wrap}.ui-codeblock--nowrap code{white-space:pre}.ui-codeblock code{font-family:"Monaco","Courier New",monospace;font-size:12px;line-height:1.35}.ui-codeblock__token--keyword{font-weight:900;color:#000}.ui-codeblock__token--string{color:#244f2c}.ui-codeblock__token--number{color:#363c91}.ui-codeblock__token--comment{color:#777;font-style:italic}.ui-codeblock__token--key{color:#0000aa;font-weight:900}.ui-codeblock__token--literal{color:#7b3f00;font-weight:900}.ui-tabs{border:2px solid #111;background:#fff;margin-top:14px}.ui-tabs__tablist{display:flex;flex-wrap:wrap;gap:3px;border-bottom:2px solid #111;background:#eee;padding:6px 6px 0}.ui-tabs__radio{position:absolute;left:-9999px}.ui-tabs__tab{border:2px solid #111;border-bottom:0;border-radius:4px 4px 0 0;background:#f8f8f8;padding:5px 10px;font-weight:900}.ui-tabs__tab--disabled{opacity:.45}.ui-tabs__panels{padding:10px}.ui-tabs__panel{display:none}.ui-tabs__panel--active{display:block}.section-title{border:2px solid #111;background:#fff;padding:6px 8px;font-size:16px;margin:12px 0 0}
`;

function page(title, ...children) {
  return ui.page({ title }, ui.style(ui.raw(retroCSS)),
    ui.div({ class: "system-menubar" },
      ui.div({ class: "system-menubar__menus" }, ui.span({ class: "system-menubar__apple" }, "●"), ui.span("File"), ui.span("Edit"), ui.span("View"), ui.span("Window"), ui.span("Help")),
      ui.div({ class: "system-menubar__right" }, ui.span("10:42 AM"), ui.span("?"), ui.span("▣"))
    ),
    ui.main({ class: "desktop" }, ui.section({ class: "window" },
      ui.div({ class: "titlebar" }, ui.span(title)),
      ui.div({ class: "hero" },
        ui.div({ class: "hero__left" }, ui.div({ class: "db-icon" }, "▦"), ui.div(ui.h1("SQLite Trace Browser"), ui.p("Prototype pages for inspecting/debugging SQLite artifacts."))),
        ui.badge("Schema", { tone: "muted" })
      ),
      ui.div({ class: "content" },
        ui.div({ class: "top-tabs" }, ui.span("Overview"), ui.span("Conversation"), ui.span("Correlations"), ui.span("Delivery"), ui.span("Reasoning"), ui.span("Tool Calls"), ui.span("Entities"), ui.span({ class: "active" }, "Schema")),
        children
      )
    ))
  );
}

function quoteIdent(name) {
  if (!/^[A-Za-z_][A-Za-z0-9_]*$/.test(name)) throw new Error("unsafe table name: " + name);
  return '"' + name.replace(/"/g, '""') + '"';
}

function tableList() {
  const rows = db.query(`
    SELECT name, type, sql
    FROM sqlite_schema
    WHERE type IN ('table', 'view')
      AND name NOT LIKE 'sqlite_%'
    ORDER BY name
  `) || [];
  return rows.map(row => ({ name: row.name, type: row.type, column_count: columnsFor(row.name).length, sql: row.sql || "" }));
}

function schemaFor(name) {
  return db.query(`SELECT name, type, sql FROM sqlite_schema WHERE name = ?`, name)[0];
}

function columnsFor(name) {
  return db.query(`PRAGMA table_info(${quoteIdent(name)})`) || [];
}

function rowsFor(name) {
  return db.query(`SELECT * FROM ${quoteIdent(name)} LIMIT 100`) || [];
}

app.get("/", (req, res) => {
  const tables = tableList();
  res.html(page("Generic SQLite Browser",
    ui.h1("Generic SQLite Browser"),
    ui.p({ class: "muted" }, "A monochrome, System-1-ish explorer for any SQLite file."),
    ui.div({ class: "card-grid" },
      ui.div({ class: "card" }, ui.small("Relations"), ui.strong(tables.length)),
      ui.div({ class: "card" }, ui.small("Database"), ui.strong("SQLite")),
      ui.div({ class: "card danger" }, ui.small("Mode"), ui.strong("Read-only UI"))
    ),
    ui.table.fromRows("tables", tables)
      .columns(c => c
        .text("name").label("Table").sortable().filterable().link(row => "/tables/" + encodeURIComponent(row.name))
        .badge("type").label("Type").filterable()
        .text("column_count").label("Columns").align("right").sortable()
      )
      .features(f => f.filters().pagination({ size: 25 }).sorting().columnPicker())
      .render({ query: req.query })
  ));
});

app.get("/tables/:name", (req, res) => {
  const name = req.params.name;
  const schema = schemaFor(name);
  if (!schema) return res.status(404).send("not found");
  const cols = columnsFor(name);
  const rows = rowsFor(name);
  res.html(page("Table " + name,
    ui.p(ui.a({ href: "/" }, "← Tables")),
    ui.h1("Table: " + name),
    ui.p(ui.badge(schema.type, { tone: schema.type === "view" ? "info" : "muted", title: "SQLite schema type" })),
    ui.tabs("table-detail-tabs", [
      {
        id: "columns",
        label: "Columns",
        content: ui.table.fromRows("columns", cols)
          .columns(c => c.text("cid").label("#").text("name").label("Name").filterable().text("type").label("Type").filterable().text("notnull").label("Required"))
          .features(f => f.filters().sorting())
          .render({ query: req.query })
      },
      {
        id: "rows",
        label: "Rows",
        content: ui.table.fromRows("rows", rows)
          .features(f => f.filters().pagination({ size: 25 }).sorting())
          .render({ query: req.query })
      },
      {
        id: "sql",
        label: "SQL",
        content: ui.sql(schema.sql || "-- no schema SQL available", { title: name, lineNumbers: true, copy: true, wrap: false, maxHeight: "420px" })
      },
      {
        id: "json",
        label: "Debug JSON",
        content: ui.jsonBlock({ schema, columns: cols, sampleRows: rows.slice(0, 3) }, { title: "debug payload", lineNumbers: true, maxHeight: "420px" })
      }
    ], { selected: req.query.tab || "columns" })
  ));
});
