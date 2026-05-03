package dbguard

import (
	"fmt"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/go-go-goja/engine"
)

type Registrar struct{ guard *Guard }

func NewRegistrar(guard *Guard) *Registrar { return &Registrar{guard: guard} }
func (r *Registrar) ID() string            { return "db-guard" }

func (r *Registrar) RegisterRuntimeModules(ctx *engine.RuntimeModuleContext, reg *require.Registry) error {
	if r.guard == nil {
		return fmt.Errorf("db.guard registrar requires guard")
	}
	reg.RegisterNativeModule("db.guard", r.loader)
	return nil
}

func (r *Registrar) loader(vm *goja.Runtime, moduleObj *goja.Object) {
	r.guard.SetRuntime(vm)
	exports := moduleObj.Get("exports").(*goja.Object)
	_ = exports.Set("configure", func(v goja.Value) error {
		r.guard.Configure(optionsFromValue(v))
		return nil
	})
	_ = exports.Set("onLimitExceeded", func(fn goja.Value) error {
		call, ok := goja.AssertFunction(fn)
		if !ok {
			return fmt.Errorf("db.guard.onLimitExceeded requires a function")
		}
		r.guard.SetCallback(call)
		return nil
	})
	_ = exports.Set("stats", func() goja.Value {
		stats, err := r.guard.Stats()
		if err != nil {
			panic(vm.NewGoError(err))
		}
		return vm.ToValue(stats.Map())
	})
	_ = exports.Set("checkNow", func(call goja.FunctionCall) goja.Value {
		reason := "manual"
		if len(call.Arguments) > 0 && !goja.IsUndefined(call.Arguments[0]) && !goja.IsNull(call.Arguments[0]) {
			reason = call.Arguments[0].String()
		}
		result, err := r.guard.CheckNow(reason)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		return vm.ToValue(result.Map())
	})
	_ = exports.Set("isOverLimit", func() bool { return r.guard.IsOverLimit() })
	_ = exports.Set("lastResult", func() goja.Value { return vm.ToValue(r.guard.LastResult().Map()) })
}

func optionsFromValue(v goja.Value) Options {
	opts := Options{IncludeWAL: true}
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return opts
	}
	m, ok := v.Export().(map[string]any)
	if !ok {
		return opts
	}
	opts.MaxBytes = int64From(m["maxBytes"])
	opts.SoftMaxBytes = int64From(m["softMaxBytes"])
	opts.HardMaxBytes = int64From(m["hardMaxBytes"])
	if cooldown := int64From(m["cooldownMs"]); cooldown > 0 {
		opts.Cooldown = time.Duration(cooldown) * time.Millisecond
	}
	opts.CheckEveryWrites = int64From(m["checkEveryWrites"])
	if include, ok := m["includeWal"]; ok {
		opts.IncludeWAL = truthy(include)
	}
	if fail, ok := m["failWritesOverHardLimit"]; ok {
		opts.FailOverHardLimit = truthy(fail)
	}
	return opts
}

func int64From(v any) int64 {
	switch x := v.(type) {
	case int:
		return int64(x)
	case int64:
		return x
	case int32:
		return int64(x)
	case float64:
		return int64(x)
	case float32:
		return int64(x)
	case uint64:
		return int64(x)
	case uint:
		return int64(x)
	case string:
		var n int64
		_, _ = fmt.Sscan(x, &n)
		return n
	default:
		return 0
	}
}

func truthy(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case string:
		return x != "" && x != "false"
	case nil:
		return false
	default:
		return true
	}
}
