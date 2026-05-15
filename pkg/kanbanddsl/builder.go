package kanbanddsl

import (
	"fmt"
	"strings"

	"github.com/dop251/goja"
)

type BoardBuilder struct {
	runtime *Runtime
	vm      *goja.Runtime
	cfg     BoardConfig
	errors  []string
	built   bool
}

type ColumnListBuilder struct {
	board   *BoardBuilder
	pending *ColumnBuilder
	obj     *goja.Object
}

type ColumnBuilder struct {
	parent *ColumnListBuilder
	spec   ColumnSpec
	done   bool
	obj    *goja.Object
}

type DataBuilder struct {
	board *BoardBuilder
	obj   *goja.Object
}
type FeatureBuilder struct {
	board *BoardBuilder
	obj   *goja.Object
}
type RenderBuilder struct {
	board *BoardBuilder
	obj   *goja.Object
}
type ActionBuilder struct {
	board *BoardBuilder
	obj   *goja.Object
}

func newBoardBuilder(rt *Runtime, vm *goja.Runtime, id string) *BoardBuilder {
	return &BoardBuilder{runtime: rt, vm: vm, cfg: BoardConfig{ID: strings.TrimSpace(id), Actions: ActionSpec{Custom: map[string]goja.Callable{}}}}
}

func (b *BoardBuilder) JSObject() *goja.Object {
	obj := b.vm.NewObject()
	_ = obj.Set("title", func(title string) goja.Value { b.cfg.Title = strings.TrimSpace(title); return obj })
	_ = obj.Set("description", func(text string) goja.Value { b.cfg.Description = text; return obj })
	_ = obj.Set("theme", func(theme string) goja.Value { b.cfg.Theme = strings.TrimSpace(theme); return obj })
	_ = obj.Set("className", func(className string) goja.Value { b.cfg.ClassName = strings.TrimSpace(className); return obj })
	_ = obj.Set("attrs", func(v goja.Value) goja.Value { b.cfg.Attrs = exportMap(v); return obj })
	_ = obj.Set("columns", func(fn goja.Value) goja.Value {
		b.runSubBuilder("columns", fn, newColumnListBuilder(b).JSObject())
		return obj
	})
	_ = obj.Set("data", func(fn goja.Value) goja.Value {
		b.runSubBuilder("data", fn, newDataBuilder(b).JSObject())
		return obj
	})
	_ = obj.Set("features", func(fn goja.Value) goja.Value {
		b.runSubBuilder("features", fn, newFeatureBuilder(b).JSObject())
		return obj
	})
	_ = obj.Set("render", func(fn goja.Value) goja.Value {
		b.runSubBuilder("render", fn, newRenderBuilder(b).JSObject())
		return obj
	})
	_ = obj.Set("actions", func(fn goja.Value) goja.Value {
		b.runSubBuilder("actions", fn, newActionBuilder(b).JSObject())
		return obj
	})
	_ = obj.Set("build", func() goja.Value {
		board, err := b.Build()
		if err != nil {
			panic(b.vm.NewGoError(err))
		}
		return board.JSObject()
	})
	_ = obj.Set("mount", func(app goja.Value, prefix string) goja.Value {
		board, err := b.Build()
		if err != nil {
			panic(b.vm.NewGoError(err))
		}
		if err := board.Mount(app, prefix); err != nil {
			panic(b.vm.NewGoError(err))
		}
		return board.JSObject()
	})
	return obj
}

func (b *BoardBuilder) runSubBuilder(name string, fn goja.Value, sub *goja.Object) {
	call, ok := goja.AssertFunction(fn)
	if !ok {
		b.errors = append(b.errors, fmt.Sprintf("%s(fn) requires a function", name))
		return
	}
	if _, err := call(goja.Undefined(), sub); err != nil {
		panic(err)
	}
}

func (b *BoardBuilder) Build() (*Board, error) {
	if b.built {
		return nil, fmt.Errorf("kanban.board(%q): builder has already been built", b.cfg.ID)
	}
	b.built = true
	errs := append([]string{}, b.errors...)
	if b.cfg.ID == "" {
		errs = append(errs, "board ID is required")
	}
	if b.cfg.ID != "" {
		if _, exists := b.runtime.boards[b.cfg.ID]; exists {
			errs = append(errs, fmt.Sprintf("board ID %q is already registered", b.cfg.ID))
		}
	}
	if len(b.cfg.Columns) == 0 {
		errs = append(errs, "at least one column is required")
	}
	seen := map[string]bool{}
	for i, c := range b.cfg.Columns {
		if strings.TrimSpace(c.ID) == "" {
			errs = append(errs, fmt.Sprintf("column #%d has an empty ID", i+1))
		}
		if strings.TrimSpace(c.Title) == "" {
			errs = append(errs, fmt.Sprintf("column %q must have a title", c.ID))
		}
		if c.ID != "" && seen[c.ID] {
			errs = append(errs, fmt.Sprintf("duplicate column ID %q", c.ID))
		}
		seen[c.ID] = true
		if c.Limit < 0 {
			errs = append(errs, fmt.Sprintf("column %q limit must be positive", c.ID))
		}
	}
	if b.cfg.Data.Cards == nil {
		errs = append(errs, "data.cards(fn) is required")
	}
	if b.cfg.Data.ID == nil {
		errs = append(errs, "data.id(fn) is required")
	}
	if b.cfg.Data.Column == nil {
		errs = append(errs, "data.column(fn) is required")
	}
	if b.cfg.Features.DragDrop && !b.cfg.Features.ReadOnly && b.cfg.Actions.CardMoved == nil {
		errs = append(errs, "features.dragDrop() requires actions.cardMoved(fn) unless readOnly() is enabled")
	}
	if b.cfg.Features.CreateCard && b.cfg.Actions.CardCreated == nil {
		errs = append(errs, "features.createCard() requires actions.cardCreated(fn)")
	}
	if b.cfg.Features.CardMenu && b.cfg.Actions.CardMenuAction == nil {
		errs = append(errs, "features.cardMenu() requires actions.cardMenuAction(fn)")
	}
	if b.cfg.Features.Search.Enabled && b.cfg.Features.Search.Mode != "client" && b.cfg.Features.Search.Mode != "server" {
		errs = append(errs, "features.search({mode}) must use mode \"client\" or \"server\"")
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("kanban.board(%q) is invalid:\n  - %s", b.cfg.ID, strings.Join(errs, "\n  - "))
	}
	board := &Board{runtime: b.runtime, vm: b.vm, cfg: b.cfg}
	b.runtime.boards[b.cfg.ID] = board
	return board, nil
}

func newColumnListBuilder(b *BoardBuilder) *ColumnListBuilder { return &ColumnListBuilder{board: b} }
func (c *ColumnListBuilder) JSObject() *goja.Object {
	if c.obj != nil {
		return c.obj
	}
	obj := c.board.vm.NewObject()
	c.obj = obj
	_ = obj.Set("column", func(id string) goja.Value {
		if c.pending != nil && !c.pending.done {
			c.board.errors = append(c.board.errors, fmt.Sprintf("column %q was started but not finalized with .done()", c.pending.spec.ID))
		}
		cb := &ColumnBuilder{parent: c, spec: ColumnSpec{ID: strings.TrimSpace(id), Attrs: map[string]any{}}}
		c.pending = cb
		return cb.JSObject()
	})
	return obj
}

func (c *ColumnBuilder) JSObject() *goja.Object {
	if c.obj != nil {
		return c.obj
	}
	obj := c.parent.board.vm.NewObject()
	c.obj = obj
	_ = obj.Set("title", func(title string) goja.Value { c.spec.Title = strings.TrimSpace(title); return obj })
	_ = obj.Set("description", func(text string) goja.Value { c.spec.Description = text; return obj })
	_ = obj.Set("limit", func(n int) goja.Value { c.spec.Limit = n; return obj })
	_ = obj.Set("terminal", func(v bool) goja.Value { c.spec.Terminal = v; return obj })
	_ = obj.Set("className", func(className string) goja.Value { c.spec.ClassName = strings.TrimSpace(className); return obj })
	_ = obj.Set("attrs", func(v goja.Value) goja.Value { c.spec.Attrs = exportMap(v); return obj })
	_ = obj.Set("done", func() goja.Value {
		if !c.done {
			c.done = true
			c.parent.board.cfg.Columns = append(c.parent.board.cfg.Columns, c.spec)
		}
		return c.parent.JSObject()
	})
	return obj
}

func newDataBuilder(b *BoardBuilder) *DataBuilder { return &DataBuilder{board: b} }
func (d *DataBuilder) JSObject() *goja.Object {
	if d.obj != nil {
		return d.obj
	}
	obj := d.board.vm.NewObject()
	d.obj = obj
	_ = obj.Set("cards", func(fn goja.Value) goja.Value { d.board.cfg.Data.Cards = d.mustFunction("data.cards", fn); return obj })
	_ = obj.Set("id", func(fn goja.Value) goja.Value { d.board.cfg.Data.ID = d.mustFunction("data.id", fn); return obj })
	_ = obj.Set("column", func(fn goja.Value) goja.Value {
		d.board.cfg.Data.Column = d.mustFunction("data.column", fn)
		return obj
	})
	_ = obj.Set("position", func(fn goja.Value) goja.Value {
		d.board.cfg.Data.Position = d.mustFunction("data.position", fn)
		return obj
	})
	_ = obj.Set("searchText", func(fn goja.Value) goja.Value {
		d.board.cfg.Data.SearchText = d.mustFunction("data.searchText", fn)
		return obj
	})
	return obj
}
func (d *DataBuilder) mustFunction(name string, v goja.Value) goja.Callable {
	return mustFunction(d.board, name, v)
}

func newFeatureBuilder(b *BoardBuilder) *FeatureBuilder { return &FeatureBuilder{board: b} }
func (f *FeatureBuilder) JSObject() *goja.Object {
	if f.obj != nil {
		return f.obj
	}
	obj := f.board.vm.NewObject()
	f.obj = obj
	_ = obj.Set("search", func(call goja.FunctionCall) goja.Value {
		f.board.cfg.Features.Search.Enabled = true
		f.board.cfg.Features.Search.Mode = "client"
		if len(call.Arguments) > 0 {
			m := exportMap(call.Arguments[0])
			if mode, ok := m["mode"]; ok {
				f.board.cfg.Features.Search.Mode = fmt.Sprint(mode)
			}
		}
		return obj
	})
	_ = obj.Set("dragDrop", func() goja.Value { f.board.cfg.Features.DragDrop = true; return obj })
	_ = obj.Set("createCard", func() goja.Value { f.board.cfg.Features.CreateCard = true; return obj })
	_ = obj.Set("cardMenu", func() goja.Value { f.board.cfg.Features.CardMenu = true; return obj })
	_ = obj.Set("readOnly", func() goja.Value { f.board.cfg.Features.ReadOnly = true; return obj })
	return obj
}

func newRenderBuilder(b *BoardBuilder) *RenderBuilder { return &RenderBuilder{board: b} }
func (r *RenderBuilder) JSObject() *goja.Object {
	if r.obj != nil {
		return r.obj
	}
	obj := r.board.vm.NewObject()
	r.obj = obj
	_ = obj.Set("card", func(fn goja.Value) goja.Value {
		r.board.cfg.Render.Card = mustFunction(r.board, "render.card", fn)
		return obj
	})
	_ = obj.Set("columnHeader", func(fn goja.Value) goja.Value {
		r.board.cfg.Render.ColumnHeader = mustFunction(r.board, "render.columnHeader", fn)
		return obj
	})
	_ = obj.Set("toolbar", func(fn goja.Value) goja.Value {
		r.board.cfg.Render.Toolbar = mustFunction(r.board, "render.toolbar", fn)
		return obj
	})
	_ = obj.Set("emptyColumn", func(fn goja.Value) goja.Value {
		r.board.cfg.Render.EmptyColumn = mustFunction(r.board, "render.emptyColumn", fn)
		return obj
	})
	_ = obj.Set("boardShell", func(fn goja.Value) goja.Value {
		r.board.cfg.Render.BoardShell = mustFunction(r.board, "render.boardShell", fn)
		return obj
	})
	return obj
}

func newActionBuilder(b *BoardBuilder) *ActionBuilder { return &ActionBuilder{board: b} }
func (a *ActionBuilder) JSObject() *goja.Object {
	if a.obj != nil {
		return a.obj
	}
	obj := a.board.vm.NewObject()
	a.obj = obj
	_ = obj.Set("cardMoved", func(fn goja.Value) goja.Value {
		a.board.cfg.Actions.CardMoved = mustFunction(a.board, "actions.cardMoved", fn)
		return obj
	})
	_ = obj.Set("cardCreated", func(fn goja.Value) goja.Value {
		a.board.cfg.Actions.CardCreated = mustFunction(a.board, "actions.cardCreated", fn)
		return obj
	})
	_ = obj.Set("cardUpdated", func(fn goja.Value) goja.Value {
		a.board.cfg.Actions.CardUpdated = mustFunction(a.board, "actions.cardUpdated", fn)
		return obj
	})
	_ = obj.Set("cardDeleted", func(fn goja.Value) goja.Value {
		a.board.cfg.Actions.CardDeleted = mustFunction(a.board, "actions.cardDeleted", fn)
		return obj
	})
	_ = obj.Set("cardClicked", func(fn goja.Value) goja.Value {
		a.board.cfg.Actions.CardClicked = mustFunction(a.board, "actions.cardClicked", fn)
		return obj
	})
	_ = obj.Set("cardMenuAction", func(fn goja.Value) goja.Value {
		a.board.cfg.Actions.CardMenuAction = mustFunction(a.board, "actions.cardMenuAction", fn)
		return obj
	})
	_ = obj.Set("custom", func(name string, fn goja.Value) goja.Value {
		if a.board.cfg.Actions.Custom == nil {
			a.board.cfg.Actions.Custom = map[string]goja.Callable{}
		}
		a.board.cfg.Actions.Custom[name] = mustFunction(a.board, "actions.custom("+name+")", fn)
		return obj
	})
	return obj
}

func mustFunction(b *BoardBuilder, name string, v goja.Value) goja.Callable {
	fn, ok := goja.AssertFunction(v)
	if !ok {
		b.errors = append(b.errors, name+" requires a function")
		return nil
	}
	return fn
}

func exportMap(v goja.Value) map[string]any {
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return map[string]any{}
	}
	if m, ok := v.Export().(map[string]any); ok {
		return m
	}
	return map[string]any{}
}
