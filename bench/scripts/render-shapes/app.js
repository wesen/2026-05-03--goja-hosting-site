const express = require("express");
const ui = require("ui.dsl");
const app = express.app();

function flatPage(n) {
  const children = [];
  for (let i = 0; i < n; i++) {
    children.push(ui.div({ class: "flat-node", "data-index": String(i) }, "node-" + i));
  }
  return ui.html(
    ui.head(ui.title("goja render flat bench")),
    ui.body(ui.main(ui.h1("Flat render benchmark"), ui.section({ class: "flat-root" }, children)))
  );
}

function attrPage(n) {
  const children = [];
  for (let i = 0; i < n; i++) {
    children.push(ui.div({
      class: "attr-node state-" + (i % 5),
      id: "attr-node-" + i,
      "data-index": String(i),
      "data-kind": "benchmark",
      "data-group": String(i % 17),
      "data-label": "attribute heavy node " + i,
      "aria-label": "Attribute heavy node " + i,
      role: "listitem",
      tabindex: "0"
    }, ui.span({ class: "label" }, "node-" + i)));
  }
  return ui.html(
    ui.head(ui.title("goja render attrs bench")),
    ui.body(ui.main(ui.h1("Attribute-heavy render benchmark"), ui.section({ class: "attr-root", role: "list" }, children)))
  );
}

app.get("/", (req, res) => {
  res.html(flatPage(Number(req.query.n || 1000)));
});

app.get("/flat", (req, res) => {
  res.html(flatPage(Number(req.query.n || 1000)));
});

app.get("/attrs", (req, res) => {
  res.html(attrPage(Number(req.query.n || 1000)));
});

app.get("/healthz", (req, res) => {
  res.type("text/plain").send("ok");
});
