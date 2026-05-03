package kanbanddsl

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/dop251/goja"
	"github.com/go-go-golems/goja-site/pkg/uidsl"
)

func (b *Board) JSObject() *goja.Object {
	obj := b.vm.NewObject()
	_ = obj.Set("id", b.cfg.ID)
	_ = obj.Set("render", func(call goja.FunctionCall) goja.Value {
		n, err := b.Render(valueOrUndefined(call, 0))
		if err != nil {
			panic(b.vm.NewGoError(err))
		}
		return b.vm.ToValue(n)
	})
	_ = obj.Set("mount", func(app goja.Value, prefix string) goja.Value {
		if err := b.Mount(app, prefix); err != nil {
			panic(b.vm.NewGoError(err))
		}
		return obj
	})
	_ = obj.Set("dispatch", func(action string, event goja.Value) goja.Value {
		result, err := b.Dispatch(action, event)
		if err != nil {
			panic(b.vm.NewGoError(err))
		}
		return result
	})
	_ = obj.Set("clientScriptURL", func() string {
		prefix := b.mounted
		if prefix == "" {
			prefix = "/_kanban"
		}
		return cleanJoin(prefix, "client.js")
	})
	return obj
}

func (b *Board) Render(ctx goja.Value) (uidsl.Node, error) {
	if ctx == nil || goja.IsUndefined(ctx) || goja.IsNull(ctx) {
		ctx = b.vm.ToValue(map[string]any{})
	}
	cards, err := b.loadCards(ctx)
	if err != nil {
		return nil, err
	}
	byColumn := map[string][]renderedCard{}
	for _, c := range cards {
		byColumn[c.ColumnID] = append(byColumn[c.ColumnID], c)
	}
	for id := range byColumn {
		sort.SliceStable(byColumn[id], func(i, j int) bool {
			if byColumn[id][i].Position == byColumn[id][j].Position {
				return byColumn[id][i].Index < byColumn[id][j].Index
			}
			return byColumn[id][i].Position < byColumn[id][j].Position
		})
	}

	children := []uidsl.Node{}
	if b.mounted != "" {
		children = append(children, &uidsl.Element{Tag: "script", Attrs: map[string]any{"src": cleanJoin(b.mounted, "client.js"), "defer": true}})
	}
	if b.cfg.Render.Toolbar != nil {
		v, err := b.cfg.Render.Toolbar(goja.Undefined(), ctx)
		if err != nil {
			return nil, err
		}
		n, err := uidsl.Normalize(b.vm, v)
		if err != nil {
			return nil, err
		}
		children = append(children, n)
	} else if b.cfg.Features.Search.Enabled {
		children = append(children, defaultToolbar())
	}

	columnNodes := []uidsl.Node{}
	for _, col := range b.cfg.Columns {
		columnNodes = append(columnNodes, b.renderColumn(ctx, col, byColumn[col.ID]))
	}
	boardAttrs := mergeAttrs(b.cfg.Attrs, map[string]any{
		"id":                  "kanban-" + b.cfg.ID,
		"class":               classList("kb-board", "kanban-board", b.cfg.ClassName, themeClass(b.cfg.Theme)),
		"data-kb-board-id":    b.cfg.ID,
		"data-kb-action-base": b.actionBase(),
	})
	children = append(children, &uidsl.Element{Tag: "div", Attrs: boardAttrs, Children: columnNodes})
	root := &uidsl.Element{Tag: "section", Attrs: map[string]any{"class": "kb-root", "data-kb-root": b.cfg.ID}, Children: children}
	if b.cfg.Render.BoardShell != nil {
		v, err := b.cfg.Render.BoardShell(goja.Undefined(), b.vm.ToValue(root), ctx)
		if err != nil {
			return nil, err
		}
		return uidsl.Normalize(b.vm, v)
	}
	return root, nil
}

func (b *Board) renderColumn(ctx goja.Value, col ColumnSpec, cards []renderedCard) uidsl.Node {
	header := uidsl.Node(&uidsl.Element{Tag: "div", Attrs: map[string]any{"class": "kb-column-header column-header"}, Children: []uidsl.Node{
		&uidsl.Element{Tag: "h2", Children: []uidsl.Node{&uidsl.Text{Value: col.Title}}},
		&uidsl.Element{Tag: "span", Attrs: map[string]any{"class": "kb-count count", "data-kb-count": col.ID}, Children: []uidsl.Node{&uidsl.Text{Value: strconv.Itoa(len(cards))}}},
	}})
	if b.cfg.Render.ColumnHeader != nil {
		v, err := b.cfg.Render.ColumnHeader(goja.Undefined(), b.vm.ToValue(col), ctx)
		if err == nil {
			if n, err := uidsl.Normalize(b.vm, v); err == nil {
				header = n
			}
		}
	}
	listChildren := []uidsl.Node{}
	if len(cards) == 0 {
		if b.cfg.Render.EmptyColumn != nil {
			v, err := b.cfg.Render.EmptyColumn(goja.Undefined(), b.vm.ToValue(col), ctx)
			if err == nil {
				if n, err := uidsl.Normalize(b.vm, v); err == nil {
					listChildren = append(listChildren, n)
				}
			}
		} else {
			listChildren = append(listChildren, &uidsl.Element{Tag: "div", Attrs: map[string]any{"class": "kb-empty empty"}, Children: []uidsl.Node{&uidsl.Text{Value: "No visible cards"}}})
		}
	}
	for i, card := range cards {
		listChildren = append(listChildren, b.renderCard(ctx, card, i, len(cards)))
	}
	listChildren = append(listChildren, &uidsl.Element{Tag: "div", Attrs: map[string]any{"class": "kb-drop-sentinel", "data-kb-drop-sentinel": true}})
	attrs := mergeAttrs(col.Attrs, map[string]any{
		"class":                classList("kb-column", "column", col.ClassName),
		"data-kb-column-id":    col.ID,
		"data-kb-column-title": col.Title,
		"data-kb-limit":        positiveOrEmpty(col.Limit),
	})
	return &uidsl.Element{Tag: "section", Attrs: attrs, Children: []uidsl.Node{header, &uidsl.Element{Tag: "div", Attrs: map[string]any{"class": "kb-card-list card-list", "data-kb-drop-column": col.ID}, Children: listChildren}}}
}

func (b *Board) renderCard(ctx goja.Value, card renderedCard, index, columnCount int) uidsl.Node {
	body := uidsl.Node(&uidsl.Element{Tag: "h3", Children: []uidsl.Node{&uidsl.Text{Value: card.ID}}})
	if b.cfg.Render.Card != nil {
		cardCtx := b.vm.ToValue(map[string]any{"boardId": b.cfg.ID, "cardId": card.ID, "columnId": card.ColumnID, "index": index})
		v, err := b.cfg.Render.Card(goja.Undefined(), card.Value, cardCtx)
		if err == nil {
			if n, err := uidsl.Normalize(b.vm, v); err == nil {
				body = n
			}
		}
	} else {
		obj := card.Value.ToObject(b.vm)
		title := obj.Get("title").String()
		if title == "" || title == "undefined" {
			title = card.ID
		}
		body = &uidsl.Element{Tag: "h3", Children: []uidsl.Node{&uidsl.Text{Value: title}}}
	}
	children := []uidsl.Node{body}
	if b.cfg.Features.PreciseMove && !b.cfg.Features.ReadOnly {
		children = append(children, b.preciseMoveForm(card, index, columnCount))
	}
	attrs := map[string]any{
		"class":               "kb-card kanban-card",
		"data-kb-card-id":     card.ID,
		"data-kb-card-column": card.ColumnID,
		"data-kb-card-index":  index,
		"data-kb-search-text": card.SearchText,
	}
	if b.cfg.Features.DragDrop && !b.cfg.Features.ReadOnly {
		attrs["draggable"] = true
	}
	return &uidsl.Element{Tag: "article", Attrs: attrs, Children: children}
}

func (b *Board) preciseMoveForm(card renderedCard, index, columnCount int) uidsl.Node {
	statusOptions := []uidsl.Node{}
	for _, col := range b.cfg.Columns {
		statusOptions = append(statusOptions, &uidsl.Element{Tag: "option", Attrs: map[string]any{"value": col.ID, "selected": col.ID == card.ColumnID}, Children: []uidsl.Node{&uidsl.Text{Value: col.Title}}})
	}
	positionOptions := []uidsl.Node{}
	count := columnCount
	if count < 1 {
		count = 1
	}
	for i := 0; i < count; i++ {
		positionOptions = append(positionOptions, &uidsl.Element{Tag: "option", Attrs: map[string]any{"value": i, "selected": i == index}, Children: []uidsl.Node{&uidsl.Text{Value: fmt.Sprintf("#%d", i+1)}}})
	}
	return &uidsl.Element{Tag: "form", Attrs: map[string]any{"class": "kb-move-form move-form", "method": "post", "action": cleanJoin(b.actionBase(), "cardMoved"), "data-kb-move-form": true}, Children: []uidsl.Node{
		&uidsl.Element{Tag: "input", Attrs: map[string]any{"type": "hidden", "name": "cardId", "value": card.ID}},
		&uidsl.Element{Tag: "input", Attrs: map[string]any{"type": "hidden", "name": "fromColumnId", "value": card.ColumnID}},
		&uidsl.Element{Tag: "input", Attrs: map[string]any{"type": "hidden", "name": "fromIndex", "value": index}},
		&uidsl.Element{Tag: "select", Attrs: map[string]any{"name": "toColumnId", "aria-label": "Destination column"}, Children: statusOptions},
		&uidsl.Element{Tag: "select", Attrs: map[string]any{"name": "toIndex", "aria-label": "Destination position"}, Children: positionOptions},
		&uidsl.Element{Tag: "button", Attrs: map[string]any{"type": "submit"}, Children: []uidsl.Node{&uidsl.Text{Value: "Move"}}},
	}}
}

func (b *Board) loadCards(ctx goja.Value) ([]renderedCard, error) {
	v, err := b.cfg.Data.Cards(goja.Undefined(), ctx)
	if err != nil {
		return nil, err
	}
	values := arrayValues(b.vm, v)
	out := make([]renderedCard, 0, len(values))
	for i, card := range values {
		id, err := callString(b.cfg.Data.ID, card)
		if err != nil {
			return nil, fmt.Errorf("card #%d id: %w", i, err)
		}
		col, err := callString(b.cfg.Data.Column, card)
		if err != nil {
			return nil, fmt.Errorf("card %q column: %w", id, err)
		}
		pos := float64(i)
		if b.cfg.Data.Position != nil {
			pv, err := b.cfg.Data.Position(goja.Undefined(), card)
			if err != nil {
				return nil, err
			}
			pos = pv.ToFloat()
		}
		search := ""
		if b.cfg.Data.SearchText != nil {
			sv, err := b.cfg.Data.SearchText(goja.Undefined(), card)
			if err != nil {
				return nil, err
			}
			search = strings.ToLower(sv.String())
		}
		out = append(out, renderedCard{Value: card, ID: id, ColumnID: col, Position: pos, SearchText: search, Index: i})
	}
	return out, nil
}

func callString(fn goja.Callable, arg goja.Value) (string, error) {
	v, err := fn(goja.Undefined(), arg)
	if err != nil {
		return "", err
	}
	return v.String(), nil
}

func arrayValues(vm *goja.Runtime, v goja.Value) []goja.Value {
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return nil
	}
	obj := v.ToObject(vm)
	length := int(obj.Get("length").ToInteger())
	if length < 0 {
		length = 0
	}
	out := make([]goja.Value, 0, length)
	for i := 0; i < length; i++ {
		out = append(out, obj.Get(strconv.Itoa(i)))
	}
	return out
}

func valueOrUndefined(call goja.FunctionCall, index int) goja.Value {
	if len(call.Arguments) <= index {
		return goja.Undefined()
	}
	return call.Arguments[index]
}

func defaultToolbar() uidsl.Node {
	return &uidsl.Element{Tag: "div", Attrs: map[string]any{"class": "kb-toolbar"}, Children: []uidsl.Node{
		&uidsl.Element{Tag: "input", Attrs: map[string]any{"type": "search", "name": "search", "placeholder": "Search cards...", "data-kb-search": true, "autocomplete": "off"}},
	}}
}

func mergeAttrs(base map[string]any, extra map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extra {
		if v != nil && fmt.Sprint(v) != "" {
			out[k] = v
		}
	}
	return out
}

func classList(parts ...string) string {
	out := []string{}
	for _, p := range parts {
		for _, x := range strings.Fields(p) {
			if x != "" {
				out = append(out, x)
			}
		}
	}
	return strings.Join(out, " ")
}
func themeClass(theme string) string {
	if theme == "" {
		return ""
	}
	return "kb-theme-" + theme
}
func positiveOrEmpty(n int) any {
	if n > 0 {
		return n
	}
	return nil
}
func cleanJoin(prefix, suffix string) string {
	return strings.TrimRight(prefix, "/") + "/" + strings.TrimLeft(suffix, "/")
}
func (b *Board) actionBase() string {
	prefix := b.mounted
	if prefix == "" {
		prefix = "/_kanban"
	}
	return cleanJoin(cleanJoin(prefix, b.cfg.ID), "action")
}
