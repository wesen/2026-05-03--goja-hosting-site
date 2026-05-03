package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/pkg/runtimeowner"
)

type HostOptions struct {
	Dev      bool
	Renderer Renderer
}

type Host struct {
	registry *Registry
	dev      bool
	renderer Renderer
	owner    runtimeowner.Runner
}

func NewHost(opts HostOptions) *Host {
	return &Host{registry: NewRegistry(), dev: opts.Dev, renderer: opts.Renderer}
}

func (h *Host) SetRuntime(owner runtimeowner.Runner) { h.owner = owner }
func (h *Host) Register(method, pattern string, handler goja.Callable) {
	h.registry.Add(method, pattern, handler)
}

func (h *Host) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.owner == nil {
		http.Error(w, "runtime not initialized", http.StatusInternalServerError)
		return
	}
	route, params, ok := h.registry.Match(r.Method, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	req, err := NewRequestDTO(r, params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res := NewResponse(w, h.renderer)
	_, err = h.owner.Call(r.Context(), "http-handler", func(ctx context.Context, vm *goja.Runtime) (any, error) {
		result, err := route.Handler(goja.Undefined(), vm.ToValue(req.Map()), res.JSObject(vm))
		if err != nil {
			return nil, err
		}
		if !res.Sent() && !goja.IsUndefined(result) && !goja.IsNull(result) {
			if _, ok := result.Export().(string); ok {
				return nil, res.Send(vm, result)
			}
			return nil, res.HTML(vm, result)
		}
		if !res.Sent() {
			return nil, res.End()
		}
		return nil, nil
	})
	if err != nil && !res.Sent() {
		if h.dev {
			http.Error(w, fmt.Sprintf("JavaScript handler error: %v", err), http.StatusInternalServerError)
		} else {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	}
}
