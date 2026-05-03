const express = require("express");
const ui = require("ui.dsl");
const app = express.app();

app.get("/", (req, res) => {
  res.html(ui.html(
    ui.head(ui.title("CRM Kanban")),
    ui.body(
      ui.h1("CRM Kanban"),
      ui.p("Placeholder site for the multi-site goja-site deployment."),
      ui.p("Replace this script with a CRM pipeline Kanban board.")
    )
  ));
});
