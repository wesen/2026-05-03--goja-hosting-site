const db = require("database");
const express = require("express");
const ui = require("ui.dsl");
const kanban = require("kanban.dsl");

const app = express.app();

const columns = [
  ["pitch", "Pitch"],
  ["outline", "Outline"],
  ["draft", "Draft"],
  ["copyedit", "Copy Edit"],
  ["published", "Published"],
];
const doneColumn = "published";

function columnLabel(status) {
  const found = columns.find(([value]) => value === status);
  return found ? found[1] : status;
}
function validStatus(status) { return columns.some(([value]) => value === status) ? status : "pitch"; }
function ignoreDuplicateColumn(fn) { try { fn(); } catch (e) {} }
function sessionId(session) { if (typeof session === "string") return session; return String(session?.id || "default"); }

function migrate() {
  db.exec(`CREATE TABLE IF NOT EXISTS editorial_cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL DEFAULT 'default',
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pitch',
    position INTEGER NOT NULL DEFAULT 0,
    tag TEXT NOT NULL DEFAULT 'Essay',
    owner TEXT NOT NULL DEFAULT '',
    format TEXT NOT NULL DEFAULT 'Feature',
    channel TEXT NOT NULL DEFAULT 'Web',
    priority TEXT NOT NULL DEFAULT 'Normal',
    deadline TEXT NOT NULL DEFAULT '',
    checklist_done INTEGER NOT NULL DEFAULT 0,
    checklist_total INTEGER NOT NULL DEFAULT 4,
    done INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
  )`);
  ignoreDuplicateColumn(() => db.exec("ALTER TABLE editorial_cards ADD COLUMN format TEXT NOT NULL DEFAULT 'Feature'"));
  ignoreDuplicateColumn(() => db.exec("ALTER TABLE editorial_cards ADD COLUMN channel TEXT NOT NULL DEFAULT 'Web'"));
  ignoreDuplicateColumn(() => db.exec("ALTER TABLE editorial_cards ADD COLUMN priority TEXT NOT NULL DEFAULT 'Normal'"));
  ignoreDuplicateColumn(() => db.exec("ALTER TABLE editorial_cards ADD COLUMN deadline TEXT NOT NULL DEFAULT ''"));
  ignoreDuplicateColumn(() => db.exec("ALTER TABLE editorial_cards ADD COLUMN checklist_done INTEGER NOT NULL DEFAULT 0"));
  ignoreDuplicateColumn(() => db.exec("ALTER TABLE editorial_cards ADD COLUMN checklist_total INTEGER NOT NULL DEFAULT 4"));
}

function seedIfEmpty(session) {
  const sid = sessionId(session);
  const rows = db.query("SELECT COUNT(*) AS count FROM editorial_cards WHERE session_id = ?", sid);
  if (rows.length && Number(rows[0].count || 0) === 0) {
    const cards = [
      ["Interview: Goja as a CMS", "A Q&A about letting trusted JavaScript own product workflows without building a bespoke SPA.", "pitch", 10, "Interview", "Mira", "Interview", "Newsletter", "High", "May 7", 1, 5, 0],
      ["Kanban DSL launch essay", "Narrative launch article with diagrams, callback timeline, and production screenshots.", "outline", 10, "Architecture", "Manuel", "Feature", "Web", "High", "May 9", 3, 6, 0],
      ["SQLite guard field report", "Explain soft cleanup callbacks, hard-limit rejection, and the operational design trade-offs.", "draft", 10, "Backend", "Ada", "Deep Dive", "Docs", "Normal", "May 11", 4, 7, 0],
      ["K3s deployment incident note", "Short runbook update about local-path WaitForFirstConsumer and Argo sync waves.", "copyedit", 10, "Ops", "Lin", "Runbook", "Internal", "Urgent", "May 6", 5, 5, 0],
      ["Field Notes board walkthrough", "Screenshot-led post showing drag/drop, server callbacks, and session-isolated data.", "published", 10, "Demo", "Casey", "Tutorial", "Web", "Normal", "May 3", 6, 6, 1],
    ];
    cards.forEach(c => db.exec(`INSERT INTO editorial_cards(session_id, title, description, status, position, tag, owner, format, channel, priority, deadline, checklist_done, checklist_total, done)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, sid, ...c));
  }
}

function normalizeFilters(query) {
  query = query || {};
  return { search: String(query.search || "").trim(), status: String(query.status || "").trim(), tag: String(query.tag || "").trim() };
}
function listCards(session, filters) {
  const sid = sessionId(session); seedIfEmpty(sid); filters = normalizeFilters(filters);
  const where = ["session_id = ?"]; const args = [sid];
  if (filters.search) { const q = "%" + filters.search.toLowerCase() + "%"; where.push("(lower(title) LIKE ? OR lower(description) LIKE ? OR lower(tag) LIKE ? OR lower(owner) LIKE ? OR lower(format) LIKE ? OR lower(channel) LIKE ?)"); args.push(q, q, q, q, q, q); }
  if (filters.status) { where.push("status = ?"); args.push(validStatus(filters.status)); }
  if (filters.tag) { where.push("tag = ?"); args.push(filters.tag); }
  return db.query(`SELECT * FROM editorial_cards WHERE ${where.join(" AND ")} ORDER BY position, id`, ...args);
}
function listColumn(session, status) { return db.query("SELECT * FROM editorial_cards WHERE session_id = ? AND status = ? ORDER BY position, id", sessionId(session), status); }
function normalizeColumn(session, status) { const sid = sessionId(session); listColumn(sid, status).forEach((card, index) => db.exec("UPDATE editorial_cards SET position = ? WHERE session_id = ? AND id = ?", (index + 1) * 10, sid, card.id)); }
function moveCard({ session, id, toStatus, toIndex }) {
  const sid = sessionId(session); id = Number(id); toStatus = validStatus(String(toStatus || "pitch")); toIndex = Number.isFinite(Number(toIndex)) ? Number(toIndex) : 0;
  const existing = db.query("SELECT * FROM editorial_cards WHERE session_id = ? AND id = ?", sid, id)[0]; if (!existing) throw new Error("card " + id + " not found");
  const fromStatus = existing.status; const done = toStatus === doneColumn ? 1 : 0;
  const destination = db.query("SELECT * FROM editorial_cards WHERE session_id = ? AND status = ? AND id != ? ORDER BY position, id", sid, toStatus, id);
  destination.splice(Math.max(0, Math.min(toIndex, destination.length)), 0, { ...existing, status: toStatus, done });
  db.exec("UPDATE editorial_cards SET status = ?, done = ?, updated_at = CURRENT_TIMESTAMP WHERE session_id = ? AND id = ?", toStatus, done, sid, id);
  destination.forEach((card, index) => db.exec("UPDATE editorial_cards SET position = ? WHERE session_id = ? AND id = ?", (index + 1) * 10, sid, card.id));
  if (fromStatus !== toStatus) normalizeColumn(sid, fromStatus);
  return db.query("SELECT * FROM editorial_cards WHERE session_id = ? AND id = ?", sid, id)[0];
}
function nextPosition(session, status) { const rows = db.query("SELECT COALESCE(MAX(position), 0) AS max_position FROM editorial_cards WHERE session_id = ? AND status = ?", sessionId(session), status); return Number(rows[0]?.max_position || 0) + 10; }
function searchText(card) { return `${card.title || ""} ${card.description || ""} ${card.tag || ""} ${card.owner || ""} ${card.format || ""} ${card.channel || ""}`.toLowerCase(); }
function progress(card) { const total = Math.max(1, Number(card.checklist_total || 1)); return Math.round((Number(card.checklist_done || 0) / total) * 100); }
function summary(session) { const rows = listCards(session, {}); return { total: rows.length, published: rows.filter(c => c.status === "published").length, urgent: rows.filter(c => String(c.priority).toLowerCase() === "urgent").length, avgProgress: rows.length ? Math.round(rows.reduce((a, c) => a + progress(c), 0) / rows.length) : 0 }; }

function stylesheet() { return `
  :root{--ink:#25140d;--paper:#fff7ed;--line:#7c2d12;--muted:#8a5b45;--accent:#ea580c;font-family:Georgia,'Times New Roman',serif}*{box-sizing:border-box}body{margin:0;background:#fff7ed;color:var(--ink)}.page{max-width:1540px;margin:0 auto;padding:34px}.hero{display:grid;grid-template-columns:1.3fr .7fr;gap:24px;align-items:end;border-bottom:5px double var(--line);padding-bottom:22px;margin-bottom:20px}.masthead{font-size:48px;line-height:.95;letter-spacing:-2px}.subtitle{color:var(--muted);margin-top:10px}.metrics{display:grid;grid-template-columns:repeat(4,1fr);gap:10px}.metric{border:2px solid var(--line);background:#ffedd5;padding:10px;text-align:center}.metric strong{display:block;font-size:28px}.new-card{display:grid;grid-template-columns:1.2fr 2fr .8fr .8fr .8fr .8fr auto;gap:8px;background:#431407;color:#fff;padding:12px;margin:18px 0}.new-card input,.new-card select,.new-card button,.search-form input,.search-form select,.search-form button{font:inherit;border:2px solid var(--line);padding:8px;background:#fffaf0}.new-card button,.search-form button{font-weight:900;background:#fed7aa}.search-form{display:grid;grid-template-columns:1fr auto auto;gap:8px;margin:12px 0}.board{display:grid;grid-template-columns:repeat(5,minmax(220px,1fr));gap:14px}.column{background:#fffbeb;border-left:4px solid var(--line);min-height:620px}.column-header{padding:12px;border-bottom:2px solid var(--line);text-transform:uppercase;letter-spacing:.08em}.count{float:right;border:1px solid var(--line);padding:0 6px;background:white}.card-list{min-height:520px;padding:10px}.kanban-card{background:white;border:1px solid #9a3412;padding:0;margin-bottom:14px;box-shadow:0 8px 0 #fed7aa}.story-top{background:#7c2d12;color:white;padding:8px 10px;display:flex;justify-content:space-between;gap:8px;font-size:12px;text-transform:uppercase}.story-body{padding:12px}.story-body h3{font-size:21px;line-height:1.05;margin:0 0 8px}.desc{color:#5b3426;margin:0 0 12px}.progress{height:9px;border:1px solid var(--line);background:#ffedd5}.bar{height:100%;background:#ea580c}.card-meta{display:grid;grid-template-columns:1fr auto;gap:8px;margin-top:10px;font-size:13px}.pill{border:1px solid var(--line);padding:2px 6px;background:#fff7ed}.urgent{background:#fee2e2}.move-form{display:grid;grid-template-columns:1fr .55fr auto;gap:5px;margin:10px 12px 12px}.move-form select,.move-form button{font-size:12px;padding:4px}.empty{padding:12px;color:var(--muted);font-style:italic}@media(max-width:1100px){.board{grid-template-columns:repeat(2,1fr)}.new-card,.hero{grid-template-columns:1fr}}`; }

migrate();
const board = kanban.board("editorial-pipeline").title("Editorial Pipeline").className("board")
  .columns(cols => cols.column("pitch").title("Pitch").done().column("outline").title("Outline").done().column("draft").title("Draft").done().column("copyedit").title("Copy Edit").done().column("published").title("Published").terminal(true).done())
  .data(data => data.cards(ctx => listCards(ctx.session, ctx.query || {})).id(card => String(card.id)).column(card => card.status).position(card => Number(card.position || 0)).searchText(card => searchText(card)))
  .features(features => features.search({ mode: "client" }).preciseMove().dragDrop())
  .render(render => render
    .toolbar(ctx => ui.form({ class: "search-form", method: "get", action: "/" }, ui.input({ id: "search", name: "search", placeholder: "Search by owner, channel, format...", value: String(ctx.query?.search || ""), "data-kb-search": true, autocomplete: "off" }), ui.select({ name: "status" }, ui.option({ value: "" }, "All desks"), columns.map(([value, label]) => ui.option({ value, selected: value === String(ctx.query?.status || "") }, label))), ui.button({ type: "submit" }, "Search")))
    .card(card => ui.fragment(ui.div({ class: "story-top" }, ui.span(card.format || "Feature"), ui.span(card.channel || "Web")), ui.div({ class: "story-body" }, ui.h3(card.title), ui.p({ class: "desc" }, card.description), ui.div({ class: "progress", title: progress(card) + "% complete" }, ui.div({ class: "bar", style: `width:${progress(card)}%` })), ui.div({ class: "card-meta" }, ui.span({ class: "pill" }, "✍ " + (card.owner || "Unassigned")), ui.span({ class: "pill " + (card.priority === "Urgent" ? "urgent" : "") }, card.priority || "Normal"), ui.span({ class: "pill" }, card.tag || "Essay"), ui.span({ class: "pill" }, card.deadline || "No date")))))
    .emptyColumn(column => ui.div({ class: "empty" }, "No copy on this desk.")))
  .actions(actions => actions.cardMoved(event => ({ ok: true, refresh: true, card: moveCard({ session: event.session, id: event.cardId, toStatus: event.to.columnId, toIndex: event.to.index }), toast: "Story moved" }))).build();
board.mount(app, "/_kanban");

function boardPage(req) { const s = summary(req.session); return ui.page({ title: "Editorial Desk" }, ui.style(stylesheet()), ui.main({ class: "page" }, ui.header({ class: "hero" }, ui.div(ui.h1({ class: "masthead" }, "Editorial Desk"), ui.p({ class: "subtitle" }, "A newsroom-style board with deadlines, owners, channels, formats, and checklist progress.")), ui.div({ class: "metrics" }, ui.div({ class: "metric" }, ui.strong(s.total), ui.span("stories")), ui.div({ class: "metric" }, ui.strong(s.published), ui.span("published")), ui.div({ class: "metric" }, ui.strong(s.urgent), ui.span("urgent")), ui.div({ class: "metric" }, ui.strong(s.avgProgress + "%"), ui.span("avg done")))), ui.form({ class: "new-card", method: "post", action: "/cards" }, ui.input({ name: "title", placeholder: "Headline", required: true }), ui.input({ name: "description", placeholder: "Nut graf" }), ui.select({ name: "status" }, columns.map(([value, label]) => ui.option({ value }, label))), ui.input({ name: "owner", placeholder: "Owner" }), ui.input({ name: "format", placeholder: "Format", value: "Feature" }), ui.input({ name: "deadline", placeholder: "Deadline" }), ui.button({ type: "submit" }, "File")), board.render({ query: normalizeFilters(req.query), session: req.session }))); }
app.get("/", (req, res) => res.html(boardPage(req)));
app.get("/favicon.ico", (req, res) => res.status(204).end());
app.get("/api/summary", (req, res) => res.json(summary(req.session)));
app.get("/api/cards", (req, res) => res.json(listCards(req.session, req.query)));
app.post("/cards", (req, res) => { const body = req.body || {}; const title = String(body.title || "").trim(); if (!title) return res.status(400).send("title is required"); const status = validStatus(String(body.status || "pitch")); db.exec("INSERT INTO editorial_cards(session_id, title, description, status, position, tag, owner, format, channel, priority, deadline, checklist_done, checklist_total, done) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", sessionId(req.session), title, String(body.description || ""), status, nextPosition(req.session, status), String(body.tag || "Essay"), String(body.owner || ""), String(body.format || "Feature"), String(body.channel || "Web"), String(body.priority || "Normal"), String(body.deadline || ""), 0, 4, status === doneColumn ? 1 : 0); res.redirect("/"); });
