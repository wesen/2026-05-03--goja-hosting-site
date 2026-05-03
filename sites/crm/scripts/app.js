const db = require("database");
const express = require("express");
const ui = require("ui.dsl");
const kanban = require("kanban.dsl");

const app = express.app();
const columns = [
  ["lead", "Lead"],
  ["qualified", "Qualified"],
  ["demo", "Demo"],
  ["proposal", "Proposal"],
  ["won", "Won"],
];
const doneColumn = "won";
function validStatus(status) {
  return columns.some(([v]) => v === status) ? status : "lead";
}
function sessionId(session) {
  if (typeof session === "string") return session;
  return String(session?.id || "default");
}
function ignoreDuplicateColumn(fn) {
  try {
    fn();
  } catch (e) {}
}
function money(n) {
  n = Number(n || 0);
  return "$" + Math.round(n).toLocaleString();
}
function probability(card) {
  return Math.max(0, Math.min(100, Number(card.probability || 0)));
}

function migrate() {
  db.exec(`CREATE TABLE IF NOT EXISTS crm_deals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL DEFAULT 'default',
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'lead',
    position INTEGER NOT NULL DEFAULT 0,
    tag TEXT NOT NULL DEFAULT 'Lead',
    value TEXT NOT NULL DEFAULT '',
    amount INTEGER NOT NULL DEFAULT 0,
    probability INTEGER NOT NULL DEFAULT 10,
    contact TEXT NOT NULL DEFAULT '',
    next_step TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT '',
    done INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
  )`);
  ignoreDuplicateColumn(() =>
    db.exec(
      "ALTER TABLE crm_deals ADD COLUMN amount INTEGER NOT NULL DEFAULT 0",
    ),
  );
  ignoreDuplicateColumn(() =>
    db.exec(
      "ALTER TABLE crm_deals ADD COLUMN probability INTEGER NOT NULL DEFAULT 10",
    ),
  );
  ignoreDuplicateColumn(() =>
    db.exec(
      "ALTER TABLE crm_deals ADD COLUMN contact TEXT NOT NULL DEFAULT ''",
    ),
  );
  ignoreDuplicateColumn(() =>
    db.exec(
      "ALTER TABLE crm_deals ADD COLUMN next_step TEXT NOT NULL DEFAULT ''",
    ),
  );
  ignoreDuplicateColumn(() =>
    db.exec("ALTER TABLE crm_deals ADD COLUMN source TEXT NOT NULL DEFAULT ''"),
  );
}

function seedIfEmpty(session) {
  const sid = sessionId(session);
  const rows = db.query(
    "SELECT COUNT(*) AS count FROM crm_deals WHERE session_id = ?",
    sid,
  );
  if (rows.length && Number(rows[0].count || 0) === 0) {
    const cards = [
      [
        "Northstar Labs",
        "Wants a hosted internal planning board for research ops.",
        "lead",
        10,
        "SaaS",
        "$8k",
        8000,
        15,
        "Priya Shah",
        "Send discovery notes",
        "Inbound",
        0,
      ],
      [
        "Cedar Studio",
        "Needs editorial workflow and lightweight approval tracking.",
        "qualified",
        10,
        "Agency",
        "$12k",
        12000,
        35,
        "Jon Bell",
        "Book workflow mapping call",
        "Referral",
        0,
      ],
      [
        "Atlas Research",
        "Security review for self-hosted Goja apps and SQLite data retention.",
        "demo",
        10,
        "Research",
        "$24k",
        24000,
        55,
        "Dr. Kwan",
        "Run architecture demo",
        "Conference",
        0,
      ],
      [
        "Trail Supply Co",
        "Proposal sent for sales and inventory board prototype.",
        "proposal",
        10,
        "Retail",
        "$18k",
        18000,
        70,
        "Sam Rivera",
        "Follow up on pilot budget",
        "Outbound",
        0,
      ],
      [
        "Blue Lake Guides",
        "Signed pilot for trip planning Kanban with public checklist pages.",
        "won",
        10,
        "Outdoor",
        "$6k",
        6000,
        100,
        "Casey Moore",
        "Schedule onboarding",
        "Partner",
        1,
      ],
    ];
    cards.forEach((c) =>
      db.exec(
        "INSERT INTO crm_deals(session_id,title,description,status,position,tag,value,amount,probability,contact,next_step,source,done) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
        sid,
        ...c,
      ),
    );
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
    where.push(
      "(lower(title) LIKE ? OR lower(description) LIKE ? OR lower(tag) LIKE ? OR lower(contact) LIKE ? OR lower(source) LIKE ? OR lower(next_step) LIKE ?)",
    );
    args.push(q, q, q, q, q, q);
  }
  if (filters.status) {
    where.push("status = ?");
    args.push(validStatus(filters.status));
  }
  if (filters.tag) {
    where.push("tag = ?");
    args.push(filters.tag);
  }
  return db.query(
    `SELECT * FROM crm_deals WHERE ${where.join(" AND ")} ORDER BY position,id`,
    ...args,
  );
}
function listColumn(session, status) {
  return db.query(
    "SELECT * FROM crm_deals WHERE session_id = ? AND status = ? ORDER BY position,id",
    sessionId(session),
    status,
  );
}
function normalizeColumn(session, status) {
  const sid = sessionId(session);
  listColumn(sid, status).forEach((card, index) =>
    db.exec(
      "UPDATE crm_deals SET position = ? WHERE session_id = ? AND id = ?",
      (index + 1) * 10,
      sid,
      card.id,
    ),
  );
}
function moveCard({ session, id, toStatus, toIndex }) {
  const sid = sessionId(session);
  id = Number(id);
  toStatus = validStatus(String(toStatus || "lead"));
  toIndex = Number.isFinite(Number(toIndex)) ? Number(toIndex) : 0;
  const existing = db.query(
    "SELECT * FROM crm_deals WHERE session_id = ? AND id = ?",
    sid,
    id,
  )[0];
  if (!existing) throw new Error("card " + id + " not found");
  const fromStatus = existing.status;
  const done = toStatus === doneColumn ? 1 : 0;
  const destination = db.query(
    "SELECT * FROM crm_deals WHERE session_id = ? AND status = ? AND id != ? ORDER BY position,id",
    sid,
    toStatus,
    id,
  );
  destination.splice(Math.max(0, Math.min(toIndex, destination.length)), 0, {
    ...existing,
    status: toStatus,
    done,
  });
  db.exec(
    "UPDATE crm_deals SET status = ?, done = ?, updated_at = CURRENT_TIMESTAMP WHERE session_id = ? AND id = ?",
    toStatus,
    done,
    sid,
    id,
  );
  destination.forEach((card, index) =>
    db.exec(
      "UPDATE crm_deals SET position = ? WHERE session_id = ? AND id = ?",
      (index + 1) * 10,
      sid,
      card.id,
    ),
  );
  if (fromStatus !== toStatus) normalizeColumn(sid, fromStatus);
  return db.query(
    "SELECT * FROM crm_deals WHERE session_id = ? AND id = ?",
    sid,
    id,
  )[0];
}
function nextPosition(session, status) {
  const rows = db.query(
    "SELECT COALESCE(MAX(position),0) AS max_position FROM crm_deals WHERE session_id = ? AND status = ?",
    sessionId(session),
    status,
  );
  return Number(rows[0]?.max_position || 0) + 10;
}
function searchText(card) {
  return `${card.title || ""} ${card.description || ""} ${card.tag || ""} ${card.contact || ""} ${card.source || ""} ${card.next_step || ""}`.toLowerCase();
}
function summary(session) {
  const rows = listCards(session, {});
  const pipeline = rows.reduce((a, c) => a + Number(c.amount || 0), 0);
  const weighted = rows.reduce(
    (a, c) => a + (Number(c.amount || 0) * probability(c)) / 100,
    0,
  );
  return {
    deals: rows.length,
    pipeline,
    weighted: Math.round(weighted),
    won: rows.filter((c) => c.status === "won").length,
  };
}

function stylesheet() {
  return `
  :root{--ink:#07111f;--paper:#eef7ff;--line:#0f172a;--green:#16a34a;--blue:#2563eb;--muted:#64748b;font-family:'IBM Plex Mono','Courier New',monospace}*{box-sizing:border-box}body{margin:0;background:radial-gradient(circle at top left,#dbeafe,#f8fafc 42%,#dcfce7);color:var(--ink)}.page{max-width:1540px;margin:0 auto;padding:30px}.hero{display:flex;justify-content:space-between;gap:24px;align-items:center;margin-bottom:18px}.logo{font-size:46px;font-weight:900;letter-spacing:-3px}.subtitle{color:var(--muted);margin-top:6px}.dash{display:grid;grid-template-columns:repeat(4,1fr);gap:10px}.tile{background:#0f172a;color:white;border-radius:16px;padding:14px;box-shadow:0 8px 0 #93c5fd}.tile strong{display:block;font-size:25px}.new-card{display:grid;grid-template-columns:1.1fr 1.7fr .8fr .8fr .8fr .8fr auto;gap:8px;background:white;border:3px solid var(--line);border-radius:18px;padding:12px;margin:18px 0}.search-form{display:grid;grid-template-columns:1fr auto auto;gap:8px;margin:12px 0}input,select,button{font:inherit;border:2px solid var(--line);border-radius:10px;padding:8px 10px;background:white}button{font-weight:900;background:#bbf7d0;cursor:pointer}.board{display:grid;grid-template-columns:repeat(5,minmax(220px,1fr));gap:12px}.column{background:rgba(255,255,255,.72);border:2px solid #1e293b;border-radius:18px;min-height:620px;overflow:hidden}.column-header{background:#0f172a;color:white;padding:11px 12px}.count{float:right;background:#22c55e;color:#06120b;border-radius:999px;padding:0 8px}.card-list{min-height:520px;padding:10px}.kanban-card{background:white;border:2px solid #0f172a;border-radius:16px;padding:12px;margin-bottom:12px;box-shadow:5px 6px 0 #bfdbfe}.deal-top{display:flex;justify-content:space-between;gap:8px;align-items:start}.deal-top h3{font-size:17px;margin:0}.amount{font-size:21px;font-weight:900;color:#15803d;white-space:nowrap}.desc{color:#334155;margin:9px 0}.prob{height:12px;background:#e2e8f0;border-radius:99px;overflow:hidden;border:1px solid #0f172a}.probbar{height:100%;background:linear-gradient(90deg,#22c55e,#2563eb)}.grid{display:grid;grid-template-columns:1fr 1fr;gap:6px;margin-top:10px;font-size:12px}.chip{background:#eff6ff;border:1px solid #1e3a8a;border-radius:999px;padding:3px 7px}.next{grid-column:1/-1;background:#f0fdf4;border-color:#166534}.move-form{display:grid;grid-template-columns:1fr .55fr auto;gap:5px;margin-top:10px}.move-form select,.move-form button{font-size:12px;padding:4px}.empty{padding:14px;color:var(--muted)}@media(max-width:1100px){.board{grid-template-columns:repeat(2,1fr)}.dash,.new-card{grid-template-columns:1fr 1fr}}@media(max-width:700px){.board,.dash,.new-card,.search-form{grid-template-columns:1fr}.hero{display:block}}`;
}

migrate();
const board = kanban
  .board("crm-pipeline")
  .title("CRM Pipeline")
  .className("board")
  .columns((cols) =>
    cols
      .column("lead")
      .title("Lead")
      .done()
      .column("qualified")
      .title("Qualified")
      .done()
      .column("demo")
      .title("Demo")
      .done()
      .column("proposal")
      .title("Proposal")
      .done()
      .column("won")
      .title("Won")
      .terminal(true)
      .done(),
  )
  .data((data) =>
    data
      .cards((ctx) => listCards(ctx.session, ctx.query || {}))
      .id((card) => String(card.id))
      .column((card) => card.status)
      .position((card) => Number(card.position || 0))
      .searchText((card) => searchText(card)),
  )
  .features((features) =>
    features.search({ mode: "client" }).preciseMove().dragDrop(),
  )
  .render((render) =>
    render
      .toolbar((ctx) =>
        ui.form(
          { class: "search-form", method: "get", action: "/" },
          ui.input({
            id: "search",
            name: "search",
            placeholder: "Search contact, source, next step...",
            value: String(ctx.query?.search || ""),
            "data-kb-search": true,
            autocomplete: "off",
          }),
          ui.select(
            { name: "status" },
            ui.option({ value: "" }, "All stages"),
            columns.map(([value, label]) =>
              ui.option(
                { value, selected: value === String(ctx.query?.status || "") },
                label,
              ),
            ),
          ),
          ui.button({ type: "submit" }, "Search"),
        ),
      )
      .card((card) =>
        ui.fragment(
          ui.div(
            { class: "deal-top" },
            ui.h3(card.title),
            ui.span(
              { class: "amount" },
              money(card.amount || String(card.value).replace(/[^0-9]/g, "")),
            ),
          ),
          ui.p({ class: "desc" }, card.description),
          ui.div(
            { class: "prob", title: probability(card) + "% probability" },
            ui.div({ class: "probbar", style: `width:${probability(card)}%` }),
          ),
          ui.div(
            { class: "grid" },
            ui.span({ class: "chip" }, "☎ " + (card.contact || "Unknown")),
            ui.span({ class: "chip" }, card.source || "Source"),
            ui.span({ class: "chip" }, card.tag || "Lead"),
            ui.span({ class: "chip" }, probability(card) + "%"),
            ui.span(
              { class: "chip next" },
              "Next: " + (card.next_step || "Follow up"),
            ),
          ),
        ),
      )
      .emptyColumn((column) =>
        ui.div({ class: "empty" }, "No deals in this stage."),
      ),
  )
  .actions((actions) =>
    actions.cardMoved((event) => ({
      ok: true,
      refresh: true,
      card: moveCard({
        session: event.session,
        id: event.cardId,
        toStatus: event.to.columnId,
        toIndex: event.to.index,
      }),
      toast: "Deal moved",
    })),
  )
  .build();
board.mount(app, "/_kanban");
function boardPage(req) {
  const s = summary(req.session);
  return ui.page(
    { title: "CRM Pipeline" },
    ui.style(stylesheet()),
    ui.main(
      { class: "page" },
      ui.header(
        { class: "hero" },
        ui.div(
          ui.h1({ class: "logo" }, "Sales Room"),
          ui.p(
            { class: "subtitle" },
            "A revenue board with deal values, weighted pipeline, contacts, sources, probabilities, and next steps.",
          ),
        ),
        ui.div(
          { class: "dash" },
          ui.div({ class: "tile" }, ui.strong(s.deals), ui.span("deals")),
          ui.div(
            { class: "tile" },
            ui.strong(money(s.pipeline)),
            ui.span("pipeline"),
          ),
          ui.div(
            { class: "tile" },
            ui.strong(money(s.weighted)),
            ui.span("weighted"),
          ),
          ui.div({ class: "tile" }, ui.strong(s.won), ui.span("won")),
        ),
      ),
      ui.form(
        { class: "new-card", method: "post", action: "/cards" },
        ui.input({ name: "title", placeholder: "Account", required: true }),
        ui.input({ name: "description", placeholder: "Opportunity notes" }),
        ui.select(
          { name: "status" },
          columns.map(([value, label]) => ui.option({ value }, label)),
        ),
        ui.input({ name: "amount", placeholder: "Amount" }),
        ui.input({ name: "contact", placeholder: "Contact" }),
        ui.input({ name: "next_step", placeholder: "Next step" }),
        ui.button({ type: "submit" }, "Add"),
      ),
      board.render({
        query: normalizeFilters(req.query),
        session: req.session,
      }),
    ),
  );
}
app.get("/", (req, res) => res.html(boardPage(req)));
app.get("/favicon.ico", (req, res) => res.status(204).end());
app.get("/api/summary", (req, res) => res.json(summary(req.session)));
app.get("/api/cards", (req, res) =>
  res.json(listCards(req.session, req.query)),
);
app.post("/cards", (req, res) => {
  const body = req.body || {};
  const title = String(body.title || "").trim();
  if (!title) return res.status(400).send("title is required");
  const status = validStatus(String(body.status || "lead"));
  const amount = Number(String(body.amount || "0").replace(/[^0-9]/g, ""));
  db.exec(
    "INSERT INTO crm_deals(session_id,title,description,status,position,tag,value,amount,probability,contact,next_step,source,done) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
    sessionId(req.session),
    title,
    String(body.description || ""),
    status,
    nextPosition(req.session, status),
    String(body.tag || "Lead"),
    money(amount),
    amount,
    Number(body.probability || 20),
    String(body.contact || ""),
    String(body.next_step || "Follow up"),
    String(body.source || "Manual"),
    status === doneColumn ? 1 : 0,
  );
  res.redirect("/");
});
