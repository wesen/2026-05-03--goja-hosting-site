package kanbanddsl

import "github.com/dop251/goja"

func Loader(vm *goja.Runtime, moduleObj *goja.Object) {
	rt := &Runtime{vm: vm, boards: map[string]*Board{}, clientPrefixes: map[string]bool{}}
	exports := moduleObj.Get("exports").(*goja.Object)
	_ = exports.Set("board", func(id string) goja.Value {
		return newBoardBuilder(rt, vm, id).JSObject()
	})
	_ = exports.Set("clientScript", func() string { return ClientScript() })
}
