package web

import (
	"fmt"
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/go-go-goja/engine"
)

type ExpressRegistrar struct{ host *Host }

func NewExpressRegistrar(host *Host) *ExpressRegistrar { return &ExpressRegistrar{host: host} }
func (r *ExpressRegistrar) ID() string                 { return "express-http" }

func (r *ExpressRegistrar) RegisterRuntimeModules(ctx *engine.RuntimeModuleContext, reg *require.Registry) error {
	if r.host == nil {
		return fmt.Errorf("express registrar requires host")
	}
	r.host.SetRuntime(ctx.Owner)
	reg.RegisterNativeModule("express", r.loader)
	return nil
}

func (r *ExpressRegistrar) loader(vm *goja.Runtime, moduleObj *goja.Object) {
	exports := moduleObj.Get("exports").(*goja.Object)
	_ = exports.Set("app", func() goja.Value { return r.appObject(vm) })
}

func (r *ExpressRegistrar) appObject(vm *goja.Runtime) goja.Value {
	obj := vm.NewObject()
	for _, method := range []string{"get", "post", "put", "patch", "delete", "all"} {
		method := method
		_ = obj.Set(method, func(pattern string, handler goja.Value) error {
			fn, ok := goja.AssertFunction(handler)
			if !ok {
				return fmt.Errorf("app.%s(%q) requires a function handler", method, pattern)
			}
			r.host.Register(strings.ToUpper(method), pattern, fn)
			return nil
		})
	}
	_ = obj.Set("static", func(prefix, dir string) error {
		if prefix == "" || dir == "" {
			return fmt.Errorf("app.static requires prefix and directory")
		}
		r.host.RegisterStatic(prefix, dir)
		return nil
	})
	return obj
}
