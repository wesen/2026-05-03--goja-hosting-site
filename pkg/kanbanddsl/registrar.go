package kanbanddsl

import (
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/go-go-goja/pkg/engine"
)

type Registrar struct{ observer Observer }

func NewRegistrar(observers ...Observer) *Registrar {
	var observer Observer
	if len(observers) > 0 {
		observer = observers[0]
	}
	return &Registrar{observer: observer}
}
func (r *Registrar) ID() string { return "kanban-dsl" }
func (r *Registrar) RegisterRuntimeModule(ctx *engine.RuntimeModuleRegistrationContext, reg *require.Registry) error {
	reg.RegisterNativeModule("kanban.dsl", loaderWithObserver(r.observer))
	return nil
}
