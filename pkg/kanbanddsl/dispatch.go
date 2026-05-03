package kanbanddsl

import (
	"fmt"

	"github.com/dop251/goja"
)

func (b *Board) Dispatch(action string, event goja.Value) (goja.Value, error) {
	fn, ok := b.action(action)
	if !ok {
		return b.vm.ToValue(map[string]any{"ok": false, "error": fmt.Sprintf("unknown kanban action %q", action)}), nil
	}
	if event == nil || goja.IsUndefined(event) || goja.IsNull(event) {
		event = b.vm.ToValue(map[string]any{})
	}
	normalized := b.normalizeEvent(action, event)
	result, err := fn(goja.Undefined(), normalized)
	if err != nil {
		return nil, err
	}
	if goja.IsUndefined(result) || goja.IsNull(result) {
		return b.vm.ToValue(map[string]any{"ok": true}), nil
	}
	return result, nil
}

func (b *Board) action(action string) (goja.Callable, bool) {
	switch action {
	case "cardMoved":
		return b.cfg.Actions.CardMoved, b.cfg.Actions.CardMoved != nil
	case "cardCreated":
		return b.cfg.Actions.CardCreated, b.cfg.Actions.CardCreated != nil
	case "cardUpdated":
		return b.cfg.Actions.CardUpdated, b.cfg.Actions.CardUpdated != nil
	case "cardDeleted":
		return b.cfg.Actions.CardDeleted, b.cfg.Actions.CardDeleted != nil
	case "cardClicked":
		return b.cfg.Actions.CardClicked, b.cfg.Actions.CardClicked != nil
	case "cardMenuAction":
		return b.cfg.Actions.CardMenuAction, b.cfg.Actions.CardMenuAction != nil
	default:
		if b.cfg.Actions.Custom != nil {
			fn, ok := b.cfg.Actions.Custom[action]
			return fn, ok
		}
		return nil, false
	}
}

func (b *Board) normalizeEvent(action string, event goja.Value) goja.Value {
	obj := event.ToObject(b.vm)
	_ = obj.Set("boardId", b.cfg.ID)
	_ = obj.Set("action", action)
	if action == "cardMoved" {
		// Form posts use flat fields; drag/drop posts already send from/to objects.
		if missingValue(obj.Get("cardId")) || obj.Get("cardId").String() == "" {
			_ = obj.Set("cardId", obj.Get("id"))
		}
		if missingValue(obj.Get("from")) {
			_ = obj.Set("from", map[string]any{
				"columnId": obj.Get("fromColumnId").String(),
				"index":    obj.Get("fromIndex").ToInteger(),
			})
		}
		if missingValue(obj.Get("to")) {
			_ = obj.Set("to", map[string]any{
				"columnId": firstString(obj.Get("toColumnId"), obj.Get("toStatus"), obj.Get("status")),
				"index":    firstInt(obj.Get("toIndex"), obj.Get("index")),
			})
		}
	}
	return obj
}

func firstString(values ...goja.Value) string {
	for _, v := range values {
		if !missingValue(v) && v.String() != "" && v.String() != "undefined" {
			return v.String()
		}
	}
	return ""
}

func firstInt(values ...goja.Value) int64 {
	for _, v := range values {
		if !missingValue(v) {
			return v.ToInteger()
		}
	}
	return 0
}

func missingValue(v goja.Value) bool {
	return v == nil || goja.IsUndefined(v) || goja.IsNull(v)
}
