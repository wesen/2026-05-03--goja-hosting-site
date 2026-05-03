package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/engine"
	"github.com/go-go-golems/goja-site/pkg/uidsl"
)

func TestSessionCookieIssuedAndReused(t *testing.T) {
	host := NewHost(HostOptions{Dev: true, Renderer: uidsl.RenderAny})
	factory, err := engine.NewBuilder().WithRuntimeModuleRegistrars(NewExpressRegistrar(host), uidsl.NewRegistrar()).Build()
	if err != nil {
		t.Fatalf("build factory: %v", err)
	}
	rt, err := factory.NewRuntime(context.Background())
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	defer rt.Close(context.Background())
	host.SetRuntime(rt.Owner)

	_, err = rt.Owner.Call(context.Background(), "load-test", func(_ context.Context, vm *goja.Runtime) (any, error) {
		_, err := vm.RunString(`
			const express = require("express");
			const app = express.app();
			app.get("/session", (req, res) => res.json({ id: req.session.id, isNew: req.session.isNew }));
		`)
		return nil, err
	})
	if err != nil {
		t.Fatalf("load script: %v", err)
	}

	first := httptest.NewRecorder()
	host.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/session", nil))
	if first.Code != http.StatusOK {
		t.Fatalf("first status=%d body=%s", first.Code, first.Body.String())
	}
	cookies := first.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != defaultSessionCookieName {
		t.Fatalf("expected %s cookie, got %#v", defaultSessionCookieName, cookies)
	}
	if !validSessionID(cookies[0].Value) {
		t.Fatalf("invalid session id %q", cookies[0].Value)
	}
	if !strings.Contains(first.Body.String(), `"isNew":true`) || !strings.Contains(first.Body.String(), cookies[0].Value) {
		t.Fatalf("first response missing new session: %s", first.Body.String())
	}

	secondReq := httptest.NewRequest(http.MethodGet, "/session", nil)
	secondReq.AddCookie(cookies[0])
	second := httptest.NewRecorder()
	host.ServeHTTP(second, secondReq)
	if second.Code != http.StatusOK {
		t.Fatalf("second status=%d body=%s", second.Code, second.Body.String())
	}
	if strings.Contains(second.Header().Get("Set-Cookie"), defaultSessionCookieName) {
		t.Fatalf("did not expect replacement session cookie: %s", second.Header().Get("Set-Cookie"))
	}
	if !strings.Contains(second.Body.String(), `"isNew":false`) || !strings.Contains(second.Body.String(), cookies[0].Value) {
		t.Fatalf("second response missing reused session: %s", second.Body.String())
	}
}
