package kanbanddsl_test

import (
	"context"
	"strings"
	"testing"

	"github.com/go-go-golems/go-go-goja/modules/uidsl"
	"github.com/go-go-golems/go-go-goja/pkg/engine"
	"github.com/go-go-golems/goja-site/pkg/kanbanddsl"
)

func newRuntime(t *testing.T) *engine.Runtime {
	t.Helper()
	factory, err := engine.NewRuntimeFactoryBuilder().
		WithModules(kanbanddsl.NewRegistrar(), uidsl.NewRegistrar()).
		Build()
	if err != nil {
		t.Fatalf("build factory: %v", err)
	}
	rt, err := factory.NewRuntime(engine.WithStartupContext(context.Background()))
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	t.Cleanup(func() { _ = rt.Close(context.Background()) })
	return rt
}

func TestBuilderRendersBoardAndClientScriptURL(t *testing.T) {
	rt := newRuntime(t)
	value, err := rt.VM.RunString(`
		const kanban = require("kanban.dsl");
		const ui = require("ui.dsl");
		const board = kanban.board("test")
		  .title("Test Board")
		  .columns(cols => cols
		    .column("todo").title("To Do").done()
		    .column("done").title("Done").terminal(true).done())
		  .data(data => data
		    .cards(ctx => [{ id: 1, title: "One", status: "todo", position: 10 }, { id: 2, title: "Two", status: "done", position: 20 }])
		    .id(card => String(card.id))
		    .column(card => card.status)
		    .position(card => card.position)
		    .searchText(card => card.title))
		  .features(features => features.search().dragDrop())
		  .render(render => render.card(card => ui.fragment(ui.h3(card.title))))
		  .actions(actions => actions.cardMoved(event => ({ ok: true, refresh: true, moved: event.cardId })))
		  .build();
		ui.render(board.render({ query: {} }));
	`)
	if err != nil {
		t.Fatalf("run JS: %v", err)
	}
	html := value.String()
	for _, want := range []string{`data-kb-board-id="test"`, `data-kb-card-id="1"`, `draggable="true"`, `role="listitem"`, `data-kb-card-actions`, `aria-haspopup="menu"`, `To Do`, `Done`} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered html missing %q:\n%s", want, html)
		}
	}
}

func TestBuilderAggregatesValidationErrors(t *testing.T) {
	rt := newRuntime(t)
	_, err := rt.VM.RunString(`
		const kanban = require("kanban.dsl");
		kanban.board("broken")
		  .features(features => features.dragDrop())
		  .build();
	`)
	if err == nil {
		t.Fatalf("expected build error")
	}
	message := err.Error()
	for _, want := range []string{"at least one column", "data.cards", "data.id", "data.column", "actions.cardMoved"} {
		if !strings.Contains(message, want) {
			t.Fatalf("error missing %q:\n%s", want, message)
		}
	}
}

func TestDispatchNormalizesCardMovedEvent(t *testing.T) {
	rt := newRuntime(t)
	value, err := rt.VM.RunString(`
		const kanban = require("kanban.dsl");
		let seen;
		const board = kanban.board("dispatch")
		  .columns(cols => cols.column("todo").title("To Do").done().column("done").title("Done").done())
		  .data(data => data.cards(() => []).id(card => String(card.id)).column(card => card.status))
		  .features(features => features.dragDrop())
		  .actions(actions => actions.cardMoved(event => { seen = event; return { ok: true, refresh: false }; }))
		  .build();
		board.dispatch("cardMoved", { cardId: "7", fromColumnId: "todo", fromIndex: 1, toColumnId: "done", toIndex: 0 });
		seen.boardId + ':' + seen.action + ':' + seen.cardId + ':' + seen.from.columnId + ':' + seen.to.columnId + ':' + seen.to.index;
	`)
	if err != nil {
		t.Fatalf("run JS: %v", err)
	}
	if got, want := value.String(), "dispatch:cardMoved:7:todo:done:0"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestClientScriptIsServedByModule(t *testing.T) {
	rt := newRuntime(t)
	value, err := rt.VM.RunString(`require("kanban.dsl").clientScript().includes("data-kb-board-id") ? "ok" : "missing"`)
	if err != nil {
		t.Fatalf("run JS: %v", err)
	}
	if got := value.String(); got != "ok" {
		t.Fatalf("got %q", got)
	}
}
