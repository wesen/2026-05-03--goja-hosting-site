const express = require("express");

const app = express.app();

Herald.board.mount(app, "/_kanban");

function wantsPanel(req) {
  return String((req.headers && req.headers["X-Herald-Panel"]) || "") === "1";
}

function renderPanel(req, res, storyId) {
  const story = Herald.workflow.activeStory(req.session, { story: storyId });
  res.html(Herald.views.dossier(story));
}

app.get("/assets/herald.js", (req, res) => {
  res.type("application/javascript; charset=utf-8").send(Herald.clientScript);
});

app.get("/", (req, res) => {
  res.html(Herald.views.page(req));
});

app.get("/favicon.ico", (req, res) => {
  res.status(204).end();
});

app.post("/stories", (req, res) => {
  const body = req.body || {};
  Herald.workflow.seedIfEmpty(req.session);
  const story = Herald.repo.insertStory(Herald.util.sessionId(req.session), {
    title: String(body.title || "Untitled dispatch"),
    dek: Herald.util.deskLabel(String(body.desk || "city")),
    description: String(body.description || "Newly filed from the city room."),
    status: "pitches",
    desk: String(body.desk || "city"),
    authorKey: String(body.authorKey || "blackwell"),
    priority: "Normal",
    dueDate: "May 27, 1887",
    tags: "Assignment",
    sourceNotes: "Filed from web form.",
    wordTarget: 800,
    wordCount: 0,
    position: 999,
  });
  ["Assign reporter", "Confirm sources", "Draft first paragraph"].forEach(
    (label, index) =>
      Herald.repo.insertChecklist(
        Herald.util.sessionId(req.session),
        story.id,
        { label, done: false },
        index,
      ),
  );
  res.redirect("/?story=" + story.id);
});

app.post("/stories/:id/checklist/:itemId/toggle", (req, res) => {
  Herald.workflow.toggleChecklist(
    req.session,
    req.params.id,
    req.params.itemId,
  );
  if (wantsPanel(req)) {
    renderPanel(req, res, req.params.id);
    return;
  }
  res.redirect("/?story=" + req.params.id);
});

app.get("/stories/:id/panel", (req, res) => {
  renderPanel(req, res, req.params.id);
});

app.get("/stories/:id", (req, res) => {
  const story = Herald.workflow.activeStory(req.session, {
    story: req.params.id,
  });
  res.json(story || {});
});

app.get("/api/stories", (req, res) => {
  res.json(Herald.workflow.listStories(req.session, req.query || {}));
});

app.get("/api/stories/:id", (req, res) => {
  const story = Herald.workflow.activeStory(req.session, {
    story: req.params.id,
  });
  res.json(story || {});
});

app.get("/api/assignments", (req, res) => {
  res.json(
    Herald.workflow.listStories(req.session, req.query || {}).map((story) => ({
      id: story.id,
      title: story.title,
      author: story.author_name,
      status: story.workflow_status,
      checklist: story.checklist,
    })),
  );
});

app.get("/api/desks", (req, res) => {
  res.json(Herald.workflow.deskSummary(req.session));
});

app.get("/api/metrics", (req, res) => {
  res.json(Herald.workflow.metrics(req.session));
});
