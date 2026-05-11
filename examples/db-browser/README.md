# db-browser examples for goja-site

These examples were migrated from the retired `db-browser` shell. They now run
through the canonical `goja-site serve` command using the simple, read-only
SQLite policy:

```bash
go run ./cmd/goja-site serve \
  --db /path/to/app.sqlite \
  --scripts examples/db-browser/generic-browser/scripts \
  --db-policy simple \
  --readonly \
  --dev
```

Use `--allow-writes` and omit `--readonly` only for trusted scripts that should
mutate the database.
