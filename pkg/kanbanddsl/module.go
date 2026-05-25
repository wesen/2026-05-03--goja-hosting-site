package kanbanddsl

import "github.com/dop251/goja"

func Loader(vm *goja.Runtime, moduleObj *goja.Object) {
	loaderWithObserver(nil)(vm, moduleObj)
}

func loaderWithObserver(observer Observer) func(*goja.Runtime, *goja.Object) {
	return func(vm *goja.Runtime, moduleObj *goja.Object) {
		rt := &Runtime{vm: vm, boards: map[string]*Board{}, clientPrefixes: map[string]bool{}, observer: observer}
		exports := moduleObj.Get("exports").(*goja.Object)
		_ = exports.Set("board", func(id string) goja.Value {
			return newBoardBuilder(rt, vm, id).JSObject()
		})
		_ = exports.Set("clientScript", func() string { return ClientScript() })
	}
}
