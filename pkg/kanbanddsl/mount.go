package kanbanddsl

import (
	"fmt"
	"strings"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/modules/uidsl"
)

func (b *Board) Mount(app goja.Value, prefix string) error {
	if app == nil || goja.IsUndefined(app) || goja.IsNull(app) {
		return fmt.Errorf("board.mount(app, prefix) requires an express app object")
	}
	prefix = normalizePrefix(prefix)
	b.mounted = prefix
	if b.runtime.boards[b.cfg.ID] == nil {
		b.runtime.boards[b.cfg.ID] = b
	}
	if b.runtime.clientPrefixes == nil {
		b.runtime.clientPrefixes = map[string]bool{}
	}
	if !b.runtime.clientPrefixes[prefix] {
		if err := callAppMethod(b.vm, app, "get", cleanJoin(prefix, "client.js"), func(req, res goja.Value) goja.Value {
			resObj := res.ToObject(b.vm)
			callMethod(b.vm, resObj, "type", b.vm.ToValue("application/javascript; charset=utf-8"))
			callMethod(b.vm, resObj, "send", b.vm.ToValue(ClientScript()))
			return goja.Undefined()
		}); err != nil {
			return err
		}
		b.runtime.clientPrefixes[prefix] = true
	}
	if err := callAppMethod(b.vm, app, "get", cleanJoin(cleanJoin(prefix, b.cfg.ID), "fragment"), func(req, res goja.Value) goja.Value {
		reqObj := req.ToObject(b.vm)
		node, err := b.Render(b.vm.ToValue(map[string]any{"query": reqObj.Get("query").Export(), "session": reqObj.Get("session").Export()}))
		if err != nil {
			panic(b.vm.NewGoError(err))
		}
		callMethod(b.vm, res.ToObject(b.vm), "html", b.vm.ToValue(node))
		return goja.Undefined()
	}); err != nil {
		return err
	}
	if err := callAppMethod(b.vm, app, "post", cleanJoin(cleanJoin(cleanJoin(prefix, b.cfg.ID), "action"), ":action"), func(req, res goja.Value) goja.Value {
		reqObj := req.ToObject(b.vm)
		params := reqObj.Get("params").ToObject(b.vm)
		action := params.Get("action").String()
		body := reqObj.Get("body")
		if missingValue(body) {
			body = b.vm.ToValue(map[string]any{})
		}
		bodyObj := body.ToObject(b.vm)
		_ = bodyObj.Set("session", reqObj.Get("session"))
		result, err := b.Dispatch(action, bodyObj)
		if err != nil {
			panic(b.vm.NewGoError(err))
		}
		out := map[string]any{"ok": true}
		if exported, ok := result.Export().(map[string]any); ok {
			for k, v := range exported {
				out[k] = v
			}
		} else {
			out["data"] = result.Export()
		}
		if okValue, exists := out["ok"]; exists && !truthyGo(okValue) {
			callMethod(b.vm, res.ToObject(b.vm), "status", b.vm.ToValue(400))
			callMethod(b.vm, res.ToObject(b.vm), "json", b.vm.ToValue(out))
			return goja.Undefined()
		}
		if shouldRefresh(out["refresh"]) {
			node, err := b.Render(b.vm.ToValue(map[string]any{"query": reqObj.Get("query").Export(), "session": reqObj.Get("session").Export()}))
			if err != nil {
				panic(b.vm.NewGoError(err))
			}
			html, err := uidsl.RenderAny(b.vm, b.vm.ToValue(node))
			if err != nil {
				panic(b.vm.NewGoError(err))
			}
			out["html"] = html
		}
		callMethod(b.vm, res.ToObject(b.vm), "json", b.vm.ToValue(out))
		return goja.Undefined()
	}); err != nil {
		return err
	}
	return nil
}

func callAppMethod(vm *goja.Runtime, app goja.Value, method string, args ...any) error {
	obj := app.ToObject(vm)
	fn, ok := goja.AssertFunction(obj.Get(method))
	if !ok {
		return fmt.Errorf("app.%s is not a function", method)
	}
	values := make([]goja.Value, 0, len(args))
	for _, arg := range args {
		values = append(values, vm.ToValue(arg))
	}
	_, err := fn(obj, values...)
	return err
}

func callMethod(vm *goja.Runtime, obj *goja.Object, method string, args ...goja.Value) goja.Value {
	fn, ok := goja.AssertFunction(obj.Get(method))
	if !ok {
		panic(vm.NewGoError(fmt.Errorf("response.%s is not a function", method)))
	}
	v, err := fn(obj, args...)
	if err != nil {
		panic(err)
	}
	return v
}

func normalizePrefix(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return "/_kanban"
	}
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if len(prefix) > 1 {
		prefix = strings.TrimRight(prefix, "/")
	}
	return prefix
}

func truthyGo(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case nil:
		return false
	case string:
		return x != "" && x != "false"
	default:
		return true
	}
}

func shouldRefresh(v any) bool {
	if v == nil {
		return true
	}
	return truthyGo(v)
}
