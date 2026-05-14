const express = require("express");
const ui = require("ui.dsl");
const app = express.app();

function page(n) {
  const rows = [];
  for (let i = 0; i < n; i++) {
    rows.push(ui.li({ class: "bench-row" }, "row-" + i));
  }
  return ui.html(
    ui.head(ui.title("goja render bench")),
    ui.body(
      ui.main(
        ui.h1("goja render bench"),
        ui.p("Server-rendered UI DSL benchmark page."),
        ui.ul(rows)
      )
    )
  );
}

app.get("/", (req, res) => {
  res.html(page(Number(req.query.n || 100)));
});

app.get("/render", (req, res) => {
  res.html(page(Number(req.query.n || 100)));
});
