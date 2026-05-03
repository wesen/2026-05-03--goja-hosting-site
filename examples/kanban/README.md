# Goja Kanban Example

Run the example website:

```bash
go run ./cmd/goja-site serve \
  --db examples/kanban/kanban.db \
  --scripts examples/kanban/scripts \
  --addr :8080 \
  --dev
```

Open <http://localhost:8080/>. The entire page is registered from JavaScript, persisted in SQLite through `require("database")`, routed through `require("express")`, and rendered as HTML through `require("ui.dsl")`.
