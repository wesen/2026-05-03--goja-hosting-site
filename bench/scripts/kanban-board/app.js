const express = require("express");
const ui = require("ui.dsl");
const kanban = require("kanban.dsl");
const app = express.app();

const statuses = ["todo", "doing", "done"];
let cards = [];
for (let i = 1; i <= 120; i++) {
  cards.push({
    id: String(i),
    title: "Card " + i,
    status: statuses[(i - 1) % statuses.length],
    position: i,
    size: i % 5 === 0 ? "large" : "normal",
  });
}

function findCard(id) {
  for (let i = 0; i < cards.length; i++) {
    if (String(cards[i].id) === String(id)) return cards[i];
  }
  return null;
}

const board = kanban.board("bench")
  .columns(cols => cols
    .column("todo").title("To Do").done()
    .column("doing").title("Doing").done()
    .column("done").title("Done").done())
  .data(data => data
    .cards(() => cards)
    .id(card => String(card.id))
    .column(card => card.status)
    .position(card => card.position)
    .searchText(card => card.title + " " + card.status + " " + card.size))
  .features(features => features.search().dragDrop())
  .render(render => render
    .card(card => ui.div(
      { class: "bench-card-body" },
      ui.h3(card.title),
      ui.p("status=" + card.status + " size=" + card.size)
    )))
  .actions(actions => actions.cardMoved(event => {
    const card = findCard(event.cardId || event.id || "1");
    if (!card) return { ok: false, error: "card not found", refresh: false };
    const nextStatus = event.to && event.to.columnId ? String(event.to.columnId) : "done";
    if (statuses.indexOf(nextStatus) < 0) return { ok: false, error: "invalid column", refresh: false };
    card.status = nextStatus;
    card.position = Number(event.to && event.to.index != null ? event.to.index : 0);
    return { ok: true, refresh: true, moved: card.id, status: card.status };
  }))
  .build();

board.mount(app, "/_kanban");

app.get("/", (req, res) => {
  res.html(ui.page({ title: "Kanban benchmark" },
    ui.main(
      ui.h1("Kanban benchmark"),
      ui.p("Synthetic 120-card board for goja-site load testing."),
      board.render({ query: req.query, session: req.session })
    )
  ));
});

app.get("/healthz", (req, res) => {
  res.type("text/plain").send("ok");
});
