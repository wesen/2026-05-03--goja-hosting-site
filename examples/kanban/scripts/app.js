const db = require("database");
const express = require("express");
const ui = require("ui.dsl");

const app = express.app();

function migrate() {
  db.exec(`CREATE TABLE IF NOT EXISTS cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'todo',
    position INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
  )`);
}

function seedIfEmpty() {
  const rows = db.query("SELECT COUNT(*) AS count FROM cards");
  if (rows.length && Number(rows[0].count || 0) === 0) {
    db.exec("INSERT INTO cards(title, description, status, position) VALUES (?, ?, ?, ?)", "Write server", "Implement Express-style routing", "todo", 1);
    db.exec("INSERT INTO cards(title, description, status, position) VALUES (?, ?, ?, ?)", "Build UI DSL", "Render safe HTML from JavaScript", "doing", 2);
    db.exec("INSERT INTO cards(title, description, status, position) VALUES (?, ?, ?, ?)", "Test in browser", "Use Playwright for smoke tests", "done", 3);
  }
}

function stylesheet() {
  return `
    :root { color-scheme: light; font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }
    body { margin: 0; background: #f6f7fb; color: #172033; }
    header { padding: 28px 36px; background: linear-gradient(135deg, #1f4ed8, #8b5cf6); color: white; box-shadow: 0 12px 30px rgba(31, 78, 216, .22); }
    header h1 { margin: 0; font-size: 32px; letter-spacing: -0.03em; }
    header p { margin: 8px 0 0; opacity: .88; }
    .shell { padding: 28px 36px 40px; }
    .new-card { display: grid; grid-template-columns: minmax(220px, 1fr) minmax(260px, 2fr) 170px auto; gap: 10px; margin-bottom: 24px; background: white; padding: 16px; border-radius: 18px; box-shadow: 0 8px 30px rgba(15, 23, 42, .08); }
    input, select, button { border-radius: 12px; border: 1px solid #d8deea; padding: 10px 12px; font: inherit; }
    button { cursor: pointer; border: 0; background: #1f4ed8; color: white; font-weight: 700; }
    button.secondary { background: #edf2ff; color: #1f4ed8; }
    .board { display: grid; grid-template-columns: repeat(3, minmax(240px, 1fr)); gap: 18px; align-items: start; }
    .column { background: #e9edf7; border-radius: 22px; padding: 14px; min-height: 320px; }
    .column h2 { margin: 4px 6px 14px; text-transform: uppercase; font-size: 13px; letter-spacing: .12em; color: #526179; }
    .card { background: white; border-radius: 18px; padding: 14px; margin-bottom: 12px; box-shadow: 0 6px 20px rgba(15, 23, 42, .08); border: 1px solid rgba(15, 23, 42, .04); }
    .card h3 { margin: 0 0 6px; font-size: 17px; }
    .card p { margin: 0 0 12px; color: #64748b; line-height: 1.35; }
    .move { display: flex; gap: 8px; }
    .move select { min-width: 0; flex: 1; }
    .empty { color: #7b879a; font-style: italic; padding: 12px 8px; }
    @media (max-width: 860px) { .board, .new-card { grid-template-columns: 1fr; } header, .shell { padding-left: 18px; padding-right: 18px; } }
  `;
}

function cardView(card) {
  return ui.article({ class: "card", "data-card-id": card.id },
    ui.h3(card.title),
    card.description ? ui.p(card.description) : null,
    ui.form({ class: "move", method: "post", action: `/cards/${card.id}/move` },
      ui.select({ name: "status", "aria-label": "Move card status" },
        ["todo", "doing", "done"].map(status =>
          ui.option({ value: status, selected: status === card.status }, status)
        )
      ),
      ui.button({ class: "secondary", type: "submit" }, "Move")
    )
  );
}

function columnView(cards, status, label) {
  const filtered = cards.filter(card => card.status === status);
  return ui.section({ class: "column", "data-status": status },
    ui.h2(`${label} (${filtered.length})`),
    filtered.length ? filtered.map(cardView) : ui.div({ class: "empty" }, "No cards yet")
  );
}

function boardPage() {
  const cards = db.query("SELECT * FROM cards ORDER BY position, id");
  return ui.page({ title: "Goja Kanban" },
    ui.link({ rel: "stylesheet", href: "/style.css" }),
    ui.header(
      ui.h1("Goja Kanban"),
      ui.p("A tiny website written entirely in JavaScript, backed by SQLite, and rendered by ui.dsl.")
    ),
    ui.main({ class: "shell" },
      ui.form({ class: "new-card", method: "post", action: "/cards" },
        ui.input({ name: "title", placeholder: "Card title", required: true }),
        ui.input({ name: "description", placeholder: "Description" }),
        ui.select({ name: "status" },
          ui.option({ value: "todo" }, "todo"),
          ui.option({ value: "doing" }, "doing"),
          ui.option({ value: "done" }, "done")
        ),
        ui.button({ type: "submit" }, "Add card")
      ),
      ui.div({ class: "board" },
        columnView(cards, "todo", "To do"),
        columnView(cards, "doing", "Doing"),
        columnView(cards, "done", "Done")
      )
    )
  );
}

migrate();
seedIfEmpty();

app.get("/", (req, res) => res.html(boardPage()));

app.get("/style.css", (req, res) => {
  res.type("text/css; charset=utf-8").send(stylesheet());
});

app.get("/api/cards", (req, res) => {
  res.json(db.query("SELECT * FROM cards ORDER BY position, id"));
});

app.post("/cards", (req, res) => {
  const body = req.body || {};
  const title = String(body.title || "").trim();
  if (!title) {
    return res.status(400).send("title is required");
  }
  db.exec("INSERT INTO cards(title, description, status, position) VALUES (?, ?, ?, ?)",
    title,
    String(body.description || ""),
    String(body.status || "todo"),
    Date.now()
  );
  res.redirect("/");
});

app.post("/cards/:id/move", (req, res) => {
  const body = req.body || {};
  const status = String(body.status || "todo");
  db.exec("UPDATE cards SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", status, req.params.id);
  res.redirect("/");
});
