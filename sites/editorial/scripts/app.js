const db = require("database");
const express = require("express");
const ui = require("ui.dsl");
const kanban = require("kanban.dsl");

const app = express.app();

const columns = [
  ["pitch", "Pitch"],
  ["draft", "Draft"],
  ["edit", "Edit"],
  ["scheduled", "Scheduled"],
  ["published", "Published"]
];
const doneColumn = "published";
const tableName = "editorial_cards";

function columnLabel(status) {
  const found = columns.find(([value]) => value === status);
  return found ? found[1] : status;
}

function validStatus(status) {
  return columns.some(([value]) => value === status) ? status : columns[0][0];
}

function migrate() {
  db.exec(`CREATE TABLE IF NOT EXISTS editorial_cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL DEFAULT 'default',
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pitch',
    position INTEGER NOT NULL DEFAULT 0,
    tag TEXT NOT NULL DEFAULT 'Article',
    owner TEXT NOT NULL DEFAULT '',
    done INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
  )`);
}

function sessionId(session) {
  if (typeof session === "string") return session;
  return String(session?.id || "default");
}

function seedIfEmpty(session) {
  const sid = sessionId(session);
  const rows = db.query(`SELECT COUNT(*) AS count FROM editorial_cards WHERE session_id = ?`, sid);
  if (rows.length && Number(rows[0].count || 0) === 0) {
    const cards = [
      ["Kanban DSL launch notes", "Explain how server-rendered boards and Goja callbacks fit together.", "pitch", 10, "Architecture", "Manuel", 0],
      ["SQLite guard implementation", "Turn the db.guard work into an operator-facing article.", "draft", 10, "Backend", "Ada", 0],
      ["K3s deployment runbook", "Document the Argo CD and local-path PVC rollout lessons.", "edit", 10, "Ops", "Lin", 0],
      ["Session handling guide", "Review opaque cookie sessions and DB scoping examples.", "scheduled", 10, "Security", "Mira", 0],
      ["Field Notes board demo", "Publish the first product screenshot walkthrough.", "published", 10, "Demo", "Casey", 1]
    ];
    cards.forEach(c => db.exec(`INSERT INTO editorial_cards(session_id, title, description, status, position, tag, owner, done) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, sid, ...c));
  }
}

function normalizeFilters(query) {
  query = query || {};
  return {
    search: String(query.search || "").trim(),
    status: String(query.status || "").trim(),
    tag: String(query.tag || "").trim(),
  };
}

function listCards(session, filters) {
  const sid = sessionId(session);
  seedIfEmpty(sid);
  filters = normalizeFilters(filters);
  const where = ["session_id = ?"];
  const args = [sid];
  if (filters.search) {
    const q = "%" + filters.search.toLowerCase() + "%";
    where.push(`(lower(title) LIKE ? OR lower(description) LIKE ? OR lower(tag) LIKE ? OR lower(owner) LIKE ?)`);
    args.push(q, q, q, q);
  }
  if (filters.status) { where.push("status = ?"); args.push(validStatus(filters.status)); }
  if (filters.tag) { where.push("tag = ?"); args.push(filters.tag); }
  return db.query(`SELECT * FROM editorial_cards WHERE ${where.join(" AND ")} ORDER BY position, id`, ...args);
}

function listColumn(session, status) {
  return db.query(`SELECT * FROM editorial_cards WHERE session_id = ? AND status = ? ORDER BY position, id`, sessionId(session), status);
}

function normalizeColumn(session, status) {
  const sid = sessionId(session);
  listColumn(sid, status).forEach((card, index) => {
    db.exec(`UPDATE editorial_cards SET position = ? WHERE session_id = ? AND id = ?`, (index + 1) * 10, sid, card.id);
  });
}

function moveCard({ session, id, toStatus, toIndex }) {
  const sid = sessionId(session);
  id = Number(id);
  toStatus = validStatus(String(toStatus || columns[0][0]));
  toIndex = Number.isFinite(Number(toIndex)) ? Number(toIndex) : 0;
  const existing = db.query(`SELECT * FROM editorial_cards WHERE session_id = ? AND id = ?`, sid, id)[0];
  if (!existing) throw new Error("card " + id + " not found");
  const fromStatus = existing.status;
  const done = toStatus === doneColumn ? 1 : 0;
  const destination = db.query(`SELECT * FROM editorial_cards WHERE session_id = ? AND status = ? AND id != ? ORDER BY position, id`, sid, toStatus, id);
  const clamped = Math.max(0, Math.min(toIndex, destination.length));
  destination.splice(clamped, 0, { ...existing, status: toStatus, done });
  db.exec(`UPDATE editorial_cards SET status = ?, done = ?, updated_at = CURRENT_TIMESTAMP WHERE session_id = ? AND id = ?`, toStatus, done, sid, id);
  destination.forEach((card, index) => {
    db.exec(`UPDATE editorial_cards SET position = ? WHERE session_id = ? AND id = ?`, (index + 1) * 10, sid, card.id);
  });
  if (fromStatus !== toStatus) normalizeColumn(sid, fromStatus);
  return db.query(`SELECT * FROM editorial_cards WHERE session_id = ? AND id = ?`, sid, id)[0];
}

function nextPosition(session, status) {
  const rows = db.query(`SELECT COALESCE(MAX(position), 0) AS max_position FROM editorial_cards WHERE session_id = ? AND status = ?`, sessionId(session), status);
  return Number(rows[0]?.max_position || 0) + 10;
}

function searchText(card) { return `${card.title || ""} ${card.description || ""} ${card.tag || ""} ${card.owner || ""} ${columnLabel(card.status)}`.toLowerCase(); }

function stylesheet() { return `
  :root { --ink:#17202a; --paper:#fbfaf7; --panel:#ffffff; --muted:#6b7280; --line:#1f2937; --accent:#334155; font-family: Inter, ui-sans-serif, system-ui, sans-serif; }
  * { box-sizing:border-box; } body { margin:0; background:linear-gradient(135deg,#f8fafc,#fefce8); color:var(--ink); }
  .page { max-width:1500px; margin:0 auto; padding:32px; } .hero { display:flex; justify-content:space-between; gap:24px; align-items:end; border-bottom:3px solid var(--line); padding-bottom:22px; margin-bottom:24px; }
  .brand { display:flex; gap:16px; align-items:center; } .badge { width:58px; height:58px; border:3px solid var(--line); display:grid; place-items:center; font-size:28px; font-weight:900; background:white; box-shadow:5px 5px 0 var(--line); }
  h1,h2,h3,p { margin:0; } h1 { font-size:34px; letter-spacing:-1px; } .subtitle { color:var(--muted); margin-top:8px; }
  .new-card,.search-form { display:grid; gap:10px; margin:18px 0; } .new-card { grid-template-columns:minmax(180px,1fr) minmax(260px,2fr) 150px 150px 110px; background:white; border:2px solid var(--line); padding:12px; box-shadow:4px 4px 0 var(--line); }
  .search-form { grid-template-columns:1fr auto auto; } input,select,button { font:inherit; border:2px solid var(--line); padding:9px 10px; background:white; } button { font-weight:800; cursor:pointer; box-shadow:2px 2px 0 var(--line); }
  .board { display:grid; grid-template-columns:repeat(5, minmax(210px, 1fr)); gap:12px; align-items:start; } .column { background:#f8fafc; border:2px solid var(--line); min-height:560px; }
  .column-header { display:flex; justify-content:space-between; align-items:center; padding:10px 12px; border-bottom:2px solid var(--line); background:#e2e8f0; } .column-header h2 { font-size:18px; } .count,.tag { border:2px solid var(--line); background:white; padding:2px 8px; font-weight:800; }
  .card-list { min-height:480px; padding:10px; } .kanban-card { background:white; border:2px solid var(--line); padding:12px; margin-bottom:10px; box-shadow:3px 3px 0 var(--line); } .kanban-card h3 { font-size:16px; margin-bottom:6px; }
  .desc { color:#374151; margin:8px 0 12px; } .card-meta { display:flex; justify-content:space-between; gap:8px; align-items:center; } .extra { color:var(--muted); font-weight:800; }
  .move-form { display:grid; grid-template-columns:1fr .75fr auto; gap:6px; margin-top:12px; } .move-form select,.move-form button { padding:4px; font-size:13px; } .empty { color:var(--muted); padding:10px; }
  @media (max-width:1100px) { .board { grid-template-columns:repeat(2,minmax(220px,1fr)); } .new-card { grid-template-columns:1fr 1fr; } } @media (max-width:700px) { .page { padding:18px; } .hero { display:block; } .board,.new-card,.search-form { grid-template-columns:1fr; } }
`; }

migrate();

const board = kanban.board("editorial-pipeline")
  .title("Editorial Pipeline")
  .className("board")
  .columns(cols => cols
    .column("pitch").title("Pitch").done()
    .column("draft").title("Draft").done()
    .column("edit").title("Edit").done()
    .column("scheduled").title("Scheduled").done()
    .column("published").title("Published").terminal(true).done()
  )
  .data(data => data
    .cards(ctx => listCards(ctx.session, ctx.query || {}))
    .id(card => String(card.id))
    .column(card => card.status)
    .position(card => Number(card.position || 0))
    .searchText(card => searchText(card))
  )
  .features(features => features.search({ mode: "client" }).preciseMove().dragDrop())
  .render(render => render
    .toolbar(ctx => ui.form({ class: "search-form", method: "get", action: "/" },
      ui.input({ id: "search", name: "search", placeholder: "Search editorial pipeline...", value: String(ctx.query?.search || ""), "data-kb-search": true, autocomplete: "off" }),
      ui.select({ name: "status", "aria-label": "Filter status" },
        ui.option({ value: "", selected: String(ctx.query?.status || "") === "" }, "All columns"),
        columns.map(([value, label]) => ui.option({ value, selected: value === String(ctx.query?.status || "") }, label))
      ),
      ui.button({ type: "submit" }, "Search")
    ))
    .card((card, ctx) => ui.fragment(
      ui.h3(card.title),
      ui.p({ class: "desc" }, card.description),
      ui.div({ class: "card-meta" }, ui.span({ class: "tag" }, card.tag || "Article"), ui.span({ class: "extra" }, card.owner || "Owner"))
    ))
    .emptyColumn(column => ui.div({ class: "empty" }, "Nothing here yet"))
  )
  .actions(actions => actions.cardMoved(event => {
    const moved = moveCard({ session: event.session, id: event.cardId, toStatus: event.to.columnId, toIndex: event.to.index });
    return { ok: true, refresh: true, card: moved, toast: "Moved" };
  }))
  .build();

board.mount(app, "/_kanban");

function boardPage(req) {
  const filters = normalizeFilters(req.query);
  return ui.page({ title: "Editorial Pipeline" },
    ui.style(stylesheet()),
    ui.main({ class: "page" },
      ui.header({ class: "hero" }, ui.div({ class: "brand" }, ui.div({ class: "badge" }, "✎"), ui.div(ui.h1("Editorial Desk"), ui.p({ class: "subtitle" }, "Plan, draft, edit, and publish long-form pieces."))), ui.div("goja-site multi tenant")),
      ui.form({ class: "new-card", method: "post", action: "/cards" },
        ui.input({ name: "title", placeholder: "Title", required: true }),
        ui.input({ name: "description", placeholder: "Description" }),
        ui.select({ name: "status" }, columns.map(([value, label]) => ui.option({ value }, label))),
        ui.input({ name: "tag", placeholder: "Tag", value: "Article" }),
        ui.input({ name: "owner", placeholder: "Owner" }),
        ui.button({ type: "submit" }, "Add")
      ),
      board.render({ query: filters, session: req.session })
    )
  );
}

app.get("/", (req, res) => res.html(boardPage(req)));
app.get("/favicon.ico", (req, res) => res.status(204).end());
app.get("/api/cards", (req, res) => res.json(listCards(req.session, req.query)));
app.post("/cards", (req, res) => {
  const body = req.body || {};
  const title = String(body.title || "").trim();
  if (!title) return res.status(400).send("title is required");
  const status = validStatus(String(body.status || columns[0][0]));
  db.exec(`INSERT INTO editorial_cards(session_id, title, description, status, position, tag, owner, done) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
    sessionId(req.session), title, String(body.description || ""), status, nextPosition(req.session, status), String(body.tag || "Article"), String(body.owner || ""), status === doneColumn ? 1 : 0);
  res.redirect("/");
});
