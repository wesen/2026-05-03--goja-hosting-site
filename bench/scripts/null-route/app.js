const express = require("express");
const app = express.app();

app.get("/", (req, res) => {
  res.type("text/plain").send("ok");
});

app.get("/health", (req, res) => {
  res.json({ ok: true, scenario: "null-route" });
});
