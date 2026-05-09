const db = require("db");
const express = require("express");
const ui = require("ui.dsl");

const app = express.app();

app.get("/favicon.ico", (req, res) => res.status(204).end());

const retroCSS = `
:root { --ink:#151515; --paper:#f4f1e7; --panel:#fffdf4; --line:#151515; --shadow:#9b988e; --accent:#6f8f8a; --accent-2:#9b8378; --accent-3:#7c7f9c; }
* { box-sizing: border-box; }
body { margin:0; color:var(--ink); background:var(--paper); font-family:"Chicago", "Geneva", "Monaco", system-ui, sans-serif; font-size:14px; background-image:linear-gradient(45deg, rgba(0,0,0,.035) 25%, transparent 25%), linear-gradient(-45deg, rgba(0,0,0,.035) 25%, transparent 25%); background-size:8px 8px; }
a { color:#172f44; font-weight:700; text-decoration:underline; }
.macos-desktop { max-width:1120px; margin:24px auto; padding:0 16px; }
.window { background:var(--panel); border:2px solid var(--line); box-shadow:6px 6px 0 var(--shadow); margin-bottom:20px; }
.titlebar { display:flex; align-items:center; gap:8px; border-bottom:2px solid var(--line); padding:5px 8px; font-weight:800; background:repeating-linear-gradient(0deg, #fffdf4 0 2px, #dedacf 2px 4px); }
.closebox { width:13px; height:13px; border:2px solid var(--line); background:#fffdf4; box-shadow:inset 2px 2px 0 #d8d3c8; }
.content { padding:16px; }
.hero { display:grid; grid-template-columns:1.4fr .9fr; gap:14px; }
.panel { border:2px solid var(--line); padding:12px; background:#fffdf7; }
.panel--accent { border-left:12px solid var(--accent); }
.stat-grid { display:grid; grid-template-columns:repeat(3, 1fr); gap:10px; margin:12px 0; }
.stat { border:2px solid var(--line); background:#f8f6ee; padding:10px; box-shadow:3px 3px 0 #c8c2b6; }
.stat strong { display:block; font-size:24px; }
.ui-table-filters { display:flex; flex-wrap:wrap; gap:8px; align-items:end; border:2px solid var(--line); background:#ebe7dc; padding:10px; margin:12px 0; }
.ui-table-filters label { font-size:12px; font-weight:800; display:flex; flex-direction:column; gap:3px; }
input, button { font:inherit; border:2px solid var(--line); background:#fffdf7; color:var(--ink); padding:4px 6px; box-shadow:2px 2px 0 var(--shadow); }
button { background:#dedacf; font-weight:800; cursor:pointer; }
table.ui-table { width:100%; border-collapse:separate; border-spacing:0; background:#fffdf7; border:2px solid var(--line); }
th, td { border-right:1px solid var(--line); border-bottom:1px solid var(--line); padding:7px 8px; text-align:left; }
th { background:#dedacf; font-weight:900; }
tr:nth-child(even) td { background:#f0ede3; }
.align-right { text-align:right; }
.ui-badge, .ui-tag { display:inline-block; border:1px solid var(--line); padding:1px 5px; background:#e3e8e4; font-size:12px; font-weight:800; }
.ui-badge--vip, .ui-tag--vip { background:#d7e3df; }
.ui-badge--pending { background:#eadfcb; }
.ui-badge--shipped, .ui-badge--paid { background:#dfe5d1; }
.ui-table-pagination { margin-top:8px; border:2px solid var(--line); display:inline-block; padding:5px 8px; background:#fffdf7; box-shadow:3px 3px 0 var(--shadow); }
.small { font-size:12px; }
`;

function page(title, ...children) {
  return ui.page(
    { title },
    ui.style(ui.raw(retroCSS)),
    ui.main({ class: "macos-desktop" },
      ui.section({ class: "window" },
        ui.div({ class: "titlebar" }, ui.span({ class: "closebox" }), ui.span(title)),
        ui.div({ class: "content" }, children)
      )
    )
  );
}

function customerRows() {
  return db.query(`
    SELECT
      customers.id,
      customers.name,
      customers.segment,
      customers.email,
      COUNT(orders.id) AS order_count,
      COALESCE(SUM(orders.total_cents), 0) AS total_cents,
      GROUP_CONCAT(DISTINCT orders.status) AS tags
    FROM customers
    LEFT JOIN orders ON orders.customer_id = customers.id
    GROUP BY customers.id
    ORDER BY customers.id
  `) || [];
}

function orderRows(customerId) {
  return db.query(`
    SELECT id, status, total_cents, created_at
    FROM orders
    WHERE customer_id = ?
    ORDER BY created_at DESC
  `, customerId) || [];
}

app.get("/", (req, res) => {
  const rows = customerRows();
  const orderCount = rows.reduce((sum, row) => sum + Number(row.order_count || 0), 0);
  const revenue = rows.reduce((sum, row) => sum + Number(row.total_cents || 0), 0);
  res.html(page("Playwright Smoke DB",
    ui.div({ class: "hero" },
      ui.div({ class: "panel panel--accent" },
        ui.h1("Playwright Smoke DB"),
        ui.p("A tiny seeded SQLite app rendered through Goja, Express, db, and ui.dsl."),
        ui.p({ class: "small" }, "Try ?q=alice or ?filter.segment=vip to exercise functional filters.")
      ),
      ui.div({ class: "panel" },
        ui.h2("Ledger"),
        ui.div({ class: "stat-grid" },
          ui.div({ class: "stat" }, ui.small("Customers"), ui.strong(rows.length)),
          ui.div({ class: "stat" }, ui.small("Orders"), ui.strong(orderCount)),
          ui.div({ class: "stat" }, ui.small("Revenue"), ui.strong("$" + Math.floor(revenue / 100) + "." + String(revenue % 100).padStart(2, "0")))
        )
      )
    ),
    ui.table.fromRows("customers", rows)
      .columns(c => c
        .text("id").label("ID").sortable()
        .text("name").label("Customer").sortable().filterable().link(row => "/customers/" + row.id)
        .badge("segment").label("Segment").filterable()
        .text("email").label("Email").filterable()
        .text("order_count").label("Orders").align("right").sortable()
        .money("total_cents").label("Total").align("right").sortable()
        .tags("tags").label("Order states").filterable()
      )
      .features(f => f.filters().pagination({ size: 10 }).sorting().columnPicker())
      .render({ query: req.query })
  ));
});

app.get("/customers/:id", (req, res) => {
  const customer = db.query("SELECT * FROM customers WHERE id = ?", req.params.id)[0];
  if (!customer) return res.status(404).send("not found");
  const rows = orderRows(customer.id);
  res.html(page(customer.name,
    ui.p(ui.a({ href: "/" }, "← Customers")),
    ui.div({ class: "panel panel--accent" },
      ui.h1(customer.name),
      ui.p("Segment: " + customer.segment),
      ui.p("Email: " + customer.email)
    ),
    ui.table.fromRows("orders", rows)
      .columns(c => c
        .date("created_at").label("Date").sortable()
        .text("id").label("Order").sortable()
        .badge("status").label("Status").filterable()
        .money("total_cents").label("Total").align("right").sortable()
      )
      .features(f => f.filters().sorting())
      .render({ query: req.query })
  ));
});
