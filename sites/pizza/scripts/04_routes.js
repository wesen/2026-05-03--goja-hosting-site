const express = require("express");

const app = express.app();

Pizza.boards.kitchen.mount(app, "/_kanban");
Pizza.boards.delivery.mount(app, "/_kanban");

app.get("/", (req, res) => {
  res.html(Pizza.views.page(req));
});

app.get("/favicon.ico", (req, res) => {
  res.status(204).end();
});

app.post("/orders", (req, res) => {
  Pizza.store.createOrder(req.session, req.body || {});
  res.redirect("/");
});

app.post("/orders/:id/pay", (req, res) => {
  try {
    Pizza.store.payOrder(req.session, req.params.id, (req.body || {}).tip || 0);
    res.redirect("/");
  } catch (error) {
    res.status(400).send(String(error));
  }
});

app.get("/api/tally", (req, res) => {
  res.json(Pizza.store.tally(req.session));
});

app.get("/api/orders", (req, res) => {
  res.json(Pizza.store.listOrders(req.session));
});

app.get("/api/tasks", (req, res) => {
  res.json(Pizza.store.listTasks(req.session));
});
