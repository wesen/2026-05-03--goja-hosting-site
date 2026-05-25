package kanbanddsl_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/engine"
	expressmod "github.com/go-go-golems/go-go-goja/modules/express"
	"github.com/go-go-golems/go-go-goja/modules/uidsl"
	"github.com/go-go-golems/go-go-goja/pkg/gojahttp"
	"github.com/go-go-golems/goja-site/pkg/kanbanddsl"
)

func TestMountedBoardServesClientScriptAndActionEndpoint(t *testing.T) {
	host := gojahttp.NewHost(gojahttp.HostOptions{Dev: true, Renderer: uidsl.RenderAny})
	factory, err := engine.NewBuilder().
		WithRuntimeModuleRegistrars(expressmod.NewRegistrar(host), uidsl.NewRegistrar(), kanbanddsl.NewRegistrar()).
		Build()
	if err != nil {
		t.Fatalf("build factory: %v", err)
	}
	rt, err := factory.NewRuntime(context.Background())
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	t.Cleanup(func() { _ = rt.Close(context.Background()) })
	host.SetRuntime(rt.Owner)

	_, err = rt.Owner.Call(context.Background(), "load-test-script", func(ctx context.Context, vm *goja.Runtime) (any, error) {
		_, err := vm.RunString(`
			const express = require("express");
			const ui = require("ui.dsl");
			const kanban = require("kanban.dsl");
			const app = express.app();
			let cards = [{ id: 1, title: "One", status: "todo", position: 10 }];
			const board = kanban.board("mounted")
			  .columns(cols => cols.column("todo").title("To Do").done().column("done").title("Done").done())
			  .data(data => data.cards(() => cards).id(card => String(card.id)).column(card => card.status).position(card => card.position).searchText(card => card.title))
			  .features(features => features.search().dragDrop())
			  .render(render => render.card(card => ui.h3(card.title)))
			  .actions(actions => actions.cardMoved(event => { cards[0].status = event.to.columnId; return { ok: true, refresh: true, sessionId: event.session.id }; }))
			  .build();
			board.mount(app, "/_kanban");
			app.get("/", (req, res) => res.html(ui.page({ title: "Mounted" }, board.render({ query: req.query }))));
		`)
		return nil, err
	})
	if err != nil {
		t.Fatalf("load script: %v", err)
	}

	server := httptest.NewServer(host)
	defer server.Close()

	page := getString(t, server.URL+"/")
	if !strings.Contains(page, `src="/_kanban/client.js"`) || !strings.Contains(page, `data-kb-board-id="mounted"`) {
		t.Fatalf("page did not include mounted board/client script:\n%s", page)
	}

	client := getString(t, server.URL+"/_kanban/client.js")
	if !strings.Contains(client, "postAction") || !strings.Contains(client, "data-kb-card-actions") || !strings.Contains(client, "data-kb-live-region") {
		t.Fatalf("client script missing expected runtime markers")
	}

	pageResp, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("get page for session cookie: %v", err)
	}
	pageResp.Body.Close()
	cookies := pageResp.Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected session cookie from page request")
	}
	postReq, err := http.NewRequest(http.MethodPost, server.URL+"/_kanban/mounted/action/cardMoved", bytes.NewBufferString(`{"cardId":"1","to":{"columnId":"done","index":0}}`))
	if err != nil {
		t.Fatalf("new post request: %v", err)
	}
	postReq.Header.Set("Content-Type", "application/json")
	postReq.AddCookie(cookies[0])
	resp, err := http.DefaultClient.Do(postReq)
	if err != nil {
		t.Fatalf("post action: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("action status %d: %s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), `"html"`) || !strings.Contains(string(body), `data-kb-card-column=\"done\"`) {
		t.Fatalf("action response missing refreshed HTML: %s", body)
	}
	if !strings.Contains(string(body), cookies[0].Value) {
		t.Fatalf("action response missing callback session id %q: %s", cookies[0].Value, body)
	}
}

func getString(t *testing.T, url string) string {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("get %s: %v", url, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get %s status %d: %s", url, resp.StatusCode, body)
	}
	return string(body)
}
