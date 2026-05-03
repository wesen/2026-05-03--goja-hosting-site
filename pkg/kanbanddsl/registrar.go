package kanbanddsl

import (
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/go-go-goja/engine"
)

type Registrar struct{}

func NewRegistrar() *Registrar  { return &Registrar{} }
func (r *Registrar) ID() string { return "kanban-dsl" }
func (r *Registrar) RegisterRuntimeModules(ctx *engine.RuntimeModuleContext, reg *require.Registry) error {
	reg.RegisterNativeModule("kanban.dsl", Loader)
	return nil
}
