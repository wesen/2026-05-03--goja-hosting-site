const ui = require("ui.dsl");
const kanban = require("kanban.dsl");

globalThis.Herald = globalThis.Herald || {};

Herald.views = {
  page(req) {
    const active = Herald.workflow.activeStory(req.session, req.query || {});
    return ui.page(
      { title: "The Daily Herald" },
      ui.style(Herald.styles),
      ui.script({ src: "/assets/herald.js", defer: true }),
      ui.main(
        { class: "page" },
        this.sidebar(req),
        ui.section({ class: "main" }, this.topbar(), this.workflow(req)),
        this.dossier(active),
      ),
    );
  },

  sidebar(req) {
    const desks = Herald.workflow.deskSummary(req.session);
    return ui.section(
      { class: "sidebar" },
      ui.div({ class: "building", "aria-hidden": "true" }, "▥"),
      ui.nav(
        { class: "nav-group" },
        this.navItem("⌂", "Dashboard"),
        this.navItem("▣", "Editorial Workflow", true),
        this.navItem("♧", "Assignments"),
        this.navItem("□", "Calendar"),
        this.navItem("▤", "Archive"),
        this.navItem("♙", "People"),
      ),
      ui.div(
        ui.div({ class: "nav-title" }, "Desks"),
        ui.nav(
          { class: "nav-group" },
          desks.map((desk) => this.navItem("☞", `${desk.nav} · ${desk.total}`)),
        ),
      ),
      ui.div(
        { class: "editor-footer" },
        ui.div(
          { class: "profile" },
          ui.span({ class: "avatar" }, "EB"),
          ui.div(ui.div("E. Blackwell"), ui.small("Editor in Chief")),
        ),
        ui.small("EST. 1887"),
      ),
    );
  },

  navItem(icon, label, active) {
    return ui.a(
      { class: active ? "nav-item active" : "nav-item", href: "#" },
      ui.span({ "aria-hidden": "true" }, icon),
      ui.span(label),
    );
  },

  topbar() {
    return ui.header(
      { class: "topbar" },
      ui.div(
        { class: "masthead" },
        ui.h1("The Daily Herald"),
        ui.p("Editorial Room⌄"),
      ),
      ui.div(
        { class: "global-actions" },
        ui.label(
          { class: "search-wrap" },
          ui.input({
            class: "search",
            placeholder: "Search stories, people...",
          }),
        ),
        ui.button({ class: "icon-button", "aria-label": "Notifications" }, "♧"),
        ui.span({ class: "avatar" }, "EB"),
      ),
    );
  },

  workflow(req) {
    return ui.div(
      { class: "content" },
      ui.div(
        { class: "workflow-head" },
        ui.div(
          ui.h2("Editorial Workflow"),
          ui.div(
            { class: "tabs" },
            ui.span({ class: "tab active" }, "▣", "Board"),
            ui.span({ class: "tab" }, "▦", "Table"),
          ),
        ),
        ui.div(
          { class: "board-tools" },
          ui.span("Filter⌄"),
          ui.span("Sort: Priority⌄"),
          ui.a({ class: "new-story", href: "#new-story" }, "+ New Story"),
        ),
      ),
      this.newStoryForm(),
      Herald.board.render({ session: req.session, query: req.query }),
    );
  },

  newStoryForm() {
    return ui.form(
      {
        id: "new-story",
        class: "new-story-form",
        method: "post",
        action: "/stories",
      },
      ui.input({ name: "title", placeholder: "Headline", required: true }),
      ui.input({ name: "description", placeholder: "Assignment brief" }),
      ui.select(
        { name: "desk" },
        Herald.desks.map((desk) => ui.option({ value: desk.id }, desk.label)),
      ),
      ui.select(
        { name: "authorKey" },
        Herald.staffSeed.map((staff) =>
          ui.option({ value: staff.key }, staff.name),
        ),
      ),
      ui.button({ type: "submit" }, "File story"),
    );
  },

  storyCard(story) {
    return ui.a(
      {
        class: "story-card",
        href: "/?story=" + story.id,
        "data-herald-story-link": "true",
        "data-herald-panel-url": "/stories/" + story.id + "/panel",
      },
      ui.h3(story.title),
      ui.div(
        { class: "story-meta" },
        ui.div(story.dek || story.desk_label),
        ui.div("By " + (story.author_name || story.author_key)),
        ui.div(
          story.workflow_status === "ready"
            ? story.due_date
            : "Due " + story.due_date,
        ),
      ),
      ui.div(
        { class: "card-foot" },
        ui.span({ class: "progress" }, `Checklist ${story.progress_label}`),
        story.workflow_status === "ready"
          ? ui.span({ class: "ready-mark" }, "✓")
          : ui.span({ class: "avatar" }, story.author_initials || "??"),
      ),
    );
  },

  dossier(story) {
    if (!story)
      return ui.section({ class: "dossier" }, ui.p("No story selected."));
    return ui.section(
      { class: "dossier" },
      ui.a({ class: "close", href: "/" }, "×"),
      ui.div(
        { class: "status-label" },
        Herald.util.statusLabel(story.workflow_status),
      ),
      ui.h2(story.title),
      ui.div({ class: "desk-line" }, story.dek || story.desk_label),
      ui.div(
        { class: "author-row" },
        ui.span({ class: "avatar" }, story.author_initials || "??"),
        ui.div(
          ui.div("By " + (story.author_name || story.author_key)),
          ui.div(story.author_role || "Correspondent"),
        ),
      ),
      ui.div({ class: "fact-row" }, "□", ui.span("Due " + story.due_date)),
      ui.section(
        { class: "dossier-section" },
        ui.p({ class: "section-title" }, "Description"),
        ui.p({ class: "description" }, story.description),
      ),
      ui.section(
        { class: "dossier-section" },
        ui.p({ class: "section-title" }, "Tags"),
        ui.div(
          { class: "tags" },
          Herald.util
            .splitTags(story.tags)
            .map((tag) => ui.span({ class: "tag" }, tag)),
          ui.span({ class: "tag" }, "+"),
        ),
      ),
      ui.section(
        { class: "dossier-section" },
        ui.p({ class: "section-title" }, "Checklist"),
        ui.div(
          { class: "checklist" },
          story.checklist.map((item) => this.checkItem(story, item)),
          ui.span({ class: "check-item" }, "+", "Add item"),
        ),
      ),
      ui.section(
        { class: "dossier-section" },
        ui.p({ class: "section-title" }, "Assignment"),
        ui.div(
          { class: "assignment-grid" },
          ui.div("Priority", ui.strong(story.priority)),
          ui.div(
            "Words",
            ui.strong(`${story.word_count}/${story.word_target}`),
          ),
          ui.div("Desk", ui.strong(story.desk_label)),
          ui.div("Sources", ui.strong(story.source_notes)),
        ),
      ),
      ui.a(
        { class: "open-full", href: "/stories/" + story.id },
        "Open full view ↗",
      ),
    );
  },

  checkItem(story, item) {
    return ui.form(
      {
        method: "post",
        action: `/stories/${story.id}/checklist/${item.id}/toggle`,
        "data-herald-panel-form": "true",
      },
      ui.button(
        { class: "check-item", type: "submit" },
        ui.span(
          { class: Number(item.done || 0) ? "box done" : "box" },
          Number(item.done || 0) ? "✓" : "",
        ),
        ui.span(item.label),
      ),
    );
  },
};

Herald.board = kanban
  .board("daily-herald")
  .title("Daily Herald Editorial Workflow")
  .className("board")
  .columns((columns) =>
    columns
      .column("pitches")
      .title("Pitches")
      .done()
      .column("reporting")
      .title("Reporting")
      .done()
      .column("writing")
      .title("Writing")
      .done()
      .column("editing")
      .title("Editing")
      .done()
      .column("ready")
      .title("Ready for Print")
      .terminal(true)
      .done(),
  )
  .data((data) =>
    data
      .cards((ctx) => Herald.workflow.listStories(ctx.session))
      .id((story) => String(story.id))
      .column((story) => story.workflow_status)
      .position((story) => Number(story.position || 0))
      .searchText((story) => Herald.util.storySearchText(story)),
  )
  .features((features) => features.search({ mode: "client" }).dragDrop())
  .render((render) =>
    render
      .card((story) => Herald.views.storyCard(story))
      .emptyColumn(() => ui.div({ class: "empty" }, "No copy on this desk.")),
  )
  .actions((actions) =>
    actions.cardMoved((event) => Herald.workflow.moveStory(event)),
  )
  .build();
