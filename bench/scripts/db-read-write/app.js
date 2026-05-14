const express = require("express");
const db = require("database");
const app = express.app();

db.exec("CREATE TABLE IF NOT EXISTS bench_items(id INTEGER PRIMARY KEY AUTOINCREMENT, k TEXT NOT NULL, v TEXT NOT NULL, created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)");
db.exec("CREATE INDEX IF NOT EXISTS idx_bench_items_k ON bench_items(k)");

for (let i = 0; i < 100; i++) {
  db.exec("INSERT INTO bench_items(k, v) VALUES (?, ?)", "seed", "value-" + i);
}

app.get("/", (req, res) => {
  const rows = db.query("SELECT COUNT(*) AS count FROM bench_items");
  res.json({ ok: true, count: rows[0].count });
});

app.get("/read", (req, res) => {
  const n = Number(req.query.n || 10);
  let total = 0;
  for (let i = 0; i < n; i++) {
    total += db.query("SELECT COUNT(*) AS count FROM bench_items WHERE k = ?", "seed")[0].count;
  }
  res.json({ ok: true, queries: n, total });
});

app.post("/write", (req, res) => {
  const n = Number((req.body && req.body.n) || 1);
  for (let i = 0; i < n; i++) {
    db.exec("INSERT INTO bench_items(k, v) VALUES (?, ?)", "hot", "value-" + Date.now() + "-" + i);
  }
  const rows = db.query("SELECT COUNT(*) AS count FROM bench_items");
  res.json({ ok: true, inserted: n, count: rows[0].count });
});
