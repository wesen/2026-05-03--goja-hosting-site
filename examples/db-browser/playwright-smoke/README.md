# Playwright Smoke DB App

A tiny seeded SQLite app for browser smoke testing `goja-site serve` with Playwright.

Seed database:

- `data/app.db`
- tables: `customers`, `orders`

Run manually:

```bash
go run ./cmd/goja-site serve \
  --db examples/db-browser/playwright-smoke/data/app.db \
  --scripts examples/db-browser/playwright-smoke/scripts \
  --db-policy simple \
  --readonly \
  --addr :19090 \
  --dev
```

Then open <http://127.0.0.1:19090/>.
