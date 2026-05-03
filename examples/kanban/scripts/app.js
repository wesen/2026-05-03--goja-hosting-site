const db = require("database");
const express = require("express");
const ui = require("ui.dsl");
const kanban = require("kanban.dsl");

const app = express.app();
app.static("/assets", "examples/kanban/assets");

const columns = [
  ["todo", "To Do"],
  ["progress", "In Progress"],
  ["done", "Done"],
  ["someday", "Someday"],
];

function columnLabel(status) {
  const found = columns.find(([value]) => value === status);
  return found ? found[1] : status;
}

function validStatus(status) {
  return columns.some(([value]) => value === status) ? status : "todo";
}

function ignoreDuplicateColumn(fn) {
  try { fn(); } catch (e) { /* older demo dbs may already have the column */ }
}

function migrate() {
  db.exec(`CREATE TABLE IF NOT EXISTS cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'todo',
    position INTEGER NOT NULL DEFAULT 0,
    tag TEXT NOT NULL DEFAULT 'Planning',
    due_date TEXT NOT NULL DEFAULT '',
    done INTEGER NOT NULL DEFAULT 0,
    image TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
  )`);
  ignoreDuplicateColumn(() => db.exec("ALTER TABLE cards ADD COLUMN tag TEXT NOT NULL DEFAULT 'Planning'"));
  ignoreDuplicateColumn(() => db.exec("ALTER TABLE cards ADD COLUMN due_date TEXT NOT NULL DEFAULT ''"));
  ignoreDuplicateColumn(() => db.exec("ALTER TABLE cards ADD COLUMN done INTEGER NOT NULL DEFAULT 0"));
  ignoreDuplicateColumn(() => db.exec("ALTER TABLE cards ADD COLUMN image TEXT NOT NULL DEFAULT ''"));
}

function seedIfEmpty() {
  const rows = db.query("SELECT COUNT(*) AS count FROM cards");
  if (rows.length && Number(rows[0].count || 0) === 0) {
    const cards = [
      ["Research campsites", "Look into options near Sahale and Colonial Creek.", "todo", 10, "Planning", "May 12", 0, ""],
      ["Book permit", "Wilderness permit for 2 nights.", "todo", 20, "Logistics", "May 13", 0, ""],
      ["Check weather", "Keep an eye on the forecast and pack accordingly.", "todo", 30, "Planning", "May 14", 0, ""],
      ["Map the route", "Confirm trailhead, mileage, and elevation gain.", "progress", 10, "Planning", "May 15", 0, "/assets/trail-map.png"],
      ["Gear check", "Make a list and check everything twice.", "progress", 20, "Logistics", "May 16", 0, ""],
      ["Review trail conditions", "All clear on WTA. Snow mostly melted out.", "done", 10, "Research", "May 9", 1, ""],
      ["Plan meals", "Simple, light, and good food.", "done", 20, "Planning", "May 10", 1, ""],
      ["Invite crew", "Jordan and Casey are in. Let’s go.", "done", 30, "Logistics", "May 11", 1, ""],
      ["Side trip: Blue Lake", "Look into adding Blue Lake as a day hike.", "someday", 10, "Ideas", "", 0, ""],
      ["New camera?", "Start saving for something lighter.", "someday", 20, "Gear", "", 0, ""],
    ];
    cards.forEach(c => db.exec("INSERT INTO cards(title, description, status, position, tag, due_date, done, image) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", ...c));
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

function listCards(filters) {
  filters = normalizeFilters(filters);
  const where = [];
  const args = [];

  if (filters.search) {
    const q = "%" + filters.search.toLowerCase() + "%";
    where.push("(lower(title) LIKE ? OR lower(description) LIKE ? OR lower(tag) LIKE ?)");
    args.push(q, q, q);
  }
  if (filters.status) {
    where.push("status = ?");
    args.push(validStatus(filters.status));
  }
  if (filters.tag) {
    where.push("tag = ?");
    args.push(filters.tag);
  }

  const sql = "SELECT * FROM cards" + (where.length ? " WHERE " + where.join(" AND ") : "") + " ORDER BY position, id";
  return db.query(sql, ...args);
}

function listColumn(status) {
  return db.query("SELECT * FROM cards WHERE status = ? ORDER BY position, id", status);
}

function normalizeColumn(status) {
  listColumn(status).forEach((card, index) => {
    db.exec("UPDATE cards SET position = ? WHERE id = ?", (index + 1) * 10, card.id);
  });
}

function moveCard({ id, toStatus, toIndex }) {
  id = Number(id);
  toStatus = validStatus(String(toStatus || "todo"));
  toIndex = Number.isFinite(Number(toIndex)) ? Number(toIndex) : 0;

  const existing = db.query("SELECT * FROM cards WHERE id = ?", id)[0];
  if (!existing) throw new Error("card " + id + " not found");

  const fromStatus = existing.status;
  const done = toStatus === "done" ? 1 : 0;
  const destination = db.query("SELECT * FROM cards WHERE status = ? AND id != ? ORDER BY position, id", toStatus, id);
  const clamped = Math.max(0, Math.min(toIndex, destination.length));
  destination.splice(clamped, 0, { ...existing, status: toStatus, done });

  db.exec("UPDATE cards SET status = ?, done = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", toStatus, done, id);
  destination.forEach((card, index) => {
    db.exec("UPDATE cards SET position = ? WHERE id = ?", (index + 1) * 10, card.id);
  });
  if (fromStatus !== toStatus) normalizeColumn(fromStatus);

  return db.query("SELECT * FROM cards WHERE id = ?", id)[0];
}

function nextPosition(status) {
  const rows = db.query("SELECT COALESCE(MAX(position), 0) AS max_position FROM cards WHERE status = ?", status);
  return Number(rows[0]?.max_position || 0) + 10;
}

function stylesheet() {
  return `
    :root { --ink: #0d0d0d; --paper: #fbfbfa; --soft: #f2f2ef; --muted: #666; --line: #111; font-family: "Courier New", "IBM Plex Mono", ui-monospace, monospace; }
    * { box-sizing: border-box; }
    body { margin: 0; background: var(--paper); color: var(--ink); font: 16px/1.45 "Courier New", ui-monospace, monospace; }
    [hidden] { display: none !important; }
    .page { max-width: 1500px; margin: 0 auto; padding: 34px 36px 24px; }
    .hero { border-bottom: 2px solid var(--line); padding-bottom: 26px; margin-bottom: 34px; display: flex; justify-content: space-between; gap: 24px; align-items: end; }
    .brand { display: flex; gap: 22px; align-items: center; }
    .mascot { width: 68px; height: 68px; border: 3px solid var(--line); box-shadow: 4px 4px 0 var(--line); display: grid; place-items: center; font-size: 32px; transform: rotate(-1deg); background: white; }
    h1, h2, h3, p { margin: 0; }
    .brand h1 { font-size: 31px; line-height: 1; letter-spacing: -1px; }
    .brand p, .subtitle { margin-top: 8px; }
    .title-row { display: flex; justify-content: space-between; align-items: end; gap: 24px; margin-bottom: 22px; }
    .title-row h2 { font-size: 30px; line-height: 1.1; }
    .toolbar { display: flex; gap: 14px; align-items: center; flex-wrap: wrap; }
    button, input, select { font: inherit; }
    .toolbar button, .new-card button, .icon-button, .search-form button { border: 2px solid var(--line); background: white; color: var(--ink); min-height: 42px; padding: 8px 18px; box-shadow: 3px 3px 0 var(--line); font-weight: 700; cursor: pointer; }
    .toolbar .primary { background: var(--ink); color: white; box-shadow: none; }
    .icon-button { width: 50px; padding-inline: 0; }
    .search-form { display: grid; grid-template-columns: 1fr auto auto; gap: 10px; align-items: center; margin-bottom: 14px; }
    .search-form input, .search-form select { border: 2px solid var(--line); background: white; padding: 9px 10px; min-height: 42px; }
    .new-card { display: grid; grid-template-columns: minmax(180px, 1fr) minmax(260px, 2fr) 150px 130px 120px; gap: 10px; margin-bottom: 26px; border: 2px solid var(--line); padding: 12px; background: white; box-shadow: 4px 4px 0 var(--line); }
    .new-card input, .new-card select { border: 2px solid var(--line); padding: 8px; background: white; }
    .board { display: grid; grid-template-columns: repeat(4, minmax(230px, 1fr)); gap: 14px; align-items: start; }
    .column { border: 2px solid var(--line); background: #f7f7f4; min-height: 638px; }
    .column.drag-over { outline: 4px double var(--line); outline-offset: 4px; }
    .column-header { display: flex; justify-content: space-between; align-items: center; border-bottom: 2px solid var(--line); padding: 9px 14px; background-image: radial-gradient(var(--line) .55px, transparent .55px); background-size: 4px 4px; }
    .column-header h2 { font-size: 20px; background: var(--paper); padding-right: 6px; }
    .count { border: 2px solid var(--line); background: var(--paper); padding: 0 8px; font-weight: 700; box-shadow: 2px 2px 0 var(--line); }
    .card-list { padding: 12px 9px 8px; min-height: 520px; }
    .kanban-card { border: 2px solid var(--line); background: white; padding: 12px; margin-bottom: 12px; box-shadow: 3px 3px 0 var(--line); cursor: grab; }
    .kanban-card.dragging { opacity: .45; }
    .card-top { display: grid; grid-template-columns: 20px 1fr 34px; gap: 8px; align-items: start; margin-bottom: 7px; }
    .check { width: 17px; height: 17px; border: 2px solid var(--line); display: inline-grid; place-items: center; font-size: 14px; line-height: 1; background:white; padding:0; box-shadow:none; }
    .kanban-card h3 { font-size: 16px; font-weight: 700; }
    .card-menu { border: 0; background: transparent; font-weight: 900; padding: 0; cursor: pointer; letter-spacing: 2px; }
    .desc { white-space: pre-line; margin-bottom: 16px; }
    .card-image { width: 100%; display: block; border: 0; margin: 6px 0 16px; image-rendering: pixelated; filter: grayscale(1) contrast(1.2); }
    .move-form { display: grid; grid-template-columns: 1fr .75fr auto; gap: 6px; margin-top: 12px; }
    .move-form select, .move-form button { border: 2px solid var(--line); background: white; padding: 3px 5px; font-size: 13px; min-width: 0; }
    .move-form button { font-weight: 700; cursor: pointer; box-shadow: 1px 1px 0 var(--line); }
    .card-meta { display: flex; justify-content: space-between; align-items: end; gap: 8px; margin-top: 14px; }
    .tag { border: 2px solid var(--line); padding: 2px 7px; background: var(--paper); box-shadow: 1px 1px 0 var(--line); font-size: 14px; }
    time { font-size: 14px; white-space: nowrap; }
    .add-card { border: 0; background: transparent; padding: 0 3px 12px; cursor: pointer; text-align: left; }
    .empty { color: var(--muted); padding: 6px 3px 12px; }
    .footer { border-top: 2px solid var(--line); margin-top: 48px; padding-top: 20px; display: flex; justify-content: space-between; }
    @media (max-width: 1100px) { .board { grid-template-columns: repeat(2, minmax(230px, 1fr)); } .new-card { grid-template-columns: 1fr 1fr; } }
    @media (max-width: 680px) { .page { padding: 20px 14px; } .hero, .title-row { display: block; } .toolbar { margin-top: 20px; } .board, .new-card, .search-form { grid-template-columns: 1fr; } }
  `;
}

function toolbarButton(icon, label, klass) {
  return ui.button({ class: klass || "" }, ui.span({ "aria-hidden": "true" }, icon), " ", label);
}

function searchText(card) {
  return `${card.title || ""} ${card.description || ""} ${card.tag || ""} ${columnLabel(card.status)}`.toLowerCase();
}

migrate();
seedIfEmpty();

const board = kanban.board("trail-notes")
  .title("Trail Notes: Cascade Loop")
  .theme("field-notes")
  .className("board")
  .columns(cols => cols
    .column("todo").title("To Do").done()
    .column("progress").title("In Progress").done()
    .column("done").title("Done").terminal(true).done()
    .column("someday").title("Someday").done()
  )
  .data(data => data
    .cards(ctx => listCards(ctx.query || {}))
    .id(card => String(card.id))
    .column(card => card.status)
    .position(card => Number(card.position || 0))
    .searchText(card => searchText(card))
  )
  .features(features => features
    .search({ mode: "client" })
    .preciseMove()
    .dragDrop()
  )
  .render(render => render
    .toolbar(ctx => ui.form({ class: "search-form", method: "get", action: "/" },
      ui.input({ id: "search", name: "search", placeholder: "Search notes...", value: String(ctx.query?.search || ""), "data-kb-search": true, autocomplete: "off" }),
      ui.select({ name: "status", "aria-label": "Filter status" },
        ui.option({ value: "", selected: String(ctx.query?.status || "") === "" }, "All columns"),
        columns.map(([value, label]) => ui.option({ value, selected: value === String(ctx.query?.status || "") }, label))
      ),
      ui.button({ type: "submit" }, "Search")
    ))
    .card((card, ctx) => ui.fragment(
      ui.div({ class: "card-top" },
        ui.span({ class: "check", "aria-hidden": "true" }, Number(card.done) ? "✓" : ""),
        ui.h3(card.title),
        ui.button({ class: "card-menu", "aria-label": "Card menu" }, "...")
      ),
      ui.p({ class: "desc" }, card.description),
      card.image ? ui.img({ class: "card-image", src: card.image, alt: "Trail map sketch" }) : null,
      ui.div({ class: "card-meta" },
        ui.span({ class: "tag" }, card.tag || "Planning"),
        card.due_date ? ui.time({ datetime: card.due_date }, card.due_date) : ui.span("")
      )
    ))
    .emptyColumn(column => ui.div({ class: "empty" }, "No visible cards"))
  )
  .actions(actions => actions
    .cardMoved(event => {
      const moved = moveCard({
        id: event.cardId,
        toStatus: event.to.columnId,
        toIndex: event.to.index,
      });
      return { ok: true, refresh: true, card: moved, toast: "Moved card" };
    })
  )
  .build();

board.mount(app, "/_kanban");

function boardPage(query) {
  const filters = normalizeFilters(query);
  return ui.page({ title: "Trail Notes: Cascade Loop" },
    ui.link({ rel: "stylesheet", href: "/style.css" }),
    ui.main({ class: "page" },
      ui.header({ class: "hero" },
        ui.div({ class: "brand" }, ui.div({ class: "mascot" }, "☻"), ui.div(ui.h1("Field Notes"), ui.p("Observations from the trail."))),
        ui.div({ class: "footer-note" }, "Made with goja-site")
      ),
      ui.div({ class: "title-row" },
        ui.div(ui.h2("Trail Notes: Cascade Loop"), ui.p({ class: "subtitle" }, "Planning and notes for our weekend in the mountains.")),
        ui.div({ class: "toolbar" },
          toolbarButton("≡", "Filter"),
          toolbarButton("↕", "Sort"),
          toolbarButton("+", "New Card", "primary"),
          ui.button({ class: "icon-button", "aria-label": "More options" }, "...")
        )
      ),
      ui.form({ class: "new-card", method: "post", action: "/cards" },
        ui.input({ name: "title", placeholder: "Card title", required: true }),
        ui.input({ name: "description", placeholder: "Description" }),
        ui.select({ name: "status" }, columns.map(([value, label]) => ui.option({ value }, label))),
        ui.input({ name: "tag", placeholder: "Tag", value: "Planning" }),
        ui.button({ type: "submit" }, "Add card")
      ),
      board.render({ query: filters }),
      ui.footer({ class: "footer" }, ui.span("© 2025 Field Notes"), ui.span("Made with Microscape ◱"))
    )
  );
}

app.get("/", (req, res) => res.html(boardPage(req.query)));

app.get("/style.css", (req, res) => {
  res.type("text/css; charset=utf-8").send(stylesheet());
});

app.get("/favicon.ico", (req, res) => {
  res.status(204).end();
});

app.get("/api/cards", (req, res) => {
  res.json(listCards(req.query));
});

app.post("/cards", (req, res) => {
  const body = req.body || {};
  const title = String(body.title || "").trim();
  if (!title) return res.status(400).send("title is required");
  const status = validStatus(String(body.status || "todo"));
  db.exec("INSERT INTO cards(title, description, status, position, tag, due_date, done, image) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
    title,
    String(body.description || ""),
    status,
    nextPosition(status),
    String(body.tag || "Planning"),
    "",
    status === "done" ? 1 : 0,
    ""
  );
  res.redirect("/");
});
