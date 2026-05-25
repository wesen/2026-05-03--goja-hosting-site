#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"

PORT="${PORT:-19220}"
BASE_URL="http://127.0.0.1:${PORT}"
DB_PATH="${DB_PATH:-$(mktemp -t goja-site-kanban-a11y.XXXXXX.sqlite)}"
LOG_PATH="${LOG_PATH:-$(mktemp -t goja-site-kanban-a11y.XXXXXX.log)}"
PLAYWRIGHT_VERSION="${PLAYWRIGHT_VERSION:-1.59.1}"
PLAYWRIGHT_TMP="${PLAYWRIGHT_TMP:-$(mktemp -d -t goja-site-kanban-a11y.XXXXXX)}"
KEEP_DB="${KEEP_DB:-0}"

cleanup() {
  local status=$?
  if [[ -n "${SERVER_PID:-}" ]] && kill -0 "$SERVER_PID" 2>/dev/null; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
  if [[ "$KEEP_DB" != "1" ]]; then
    rm -f "$DB_PATH"
  fi
  rm -rf "$PLAYWRIGHT_TMP"
  if [[ $status -ne 0 ]]; then
    echo "goja-site log: $LOG_PATH" >&2
    [[ -f "$LOG_PATH" ]] && tail -200 "$LOG_PATH" >&2 || true
  fi
}
trap cleanup EXIT

rm -f "$DB_PATH"

go run ./cmd/goja-site serve \
  --db "$DB_PATH" \
  --scripts bench/scripts/kanban-board \
  --db-policy simple \
  --allow-writes \
  --addr "127.0.0.1:${PORT}" \
  --dev >"$LOG_PATH" 2>&1 &
SERVER_PID=$!

for _ in $(seq 1 100); do
  if curl -fsS "$BASE_URL/" >/dev/null 2>&1; then
    break
  fi
  if ! kill -0 "$SERVER_PID" 2>/dev/null; then
    echo "goja-site exited before becoming ready" >&2
    exit 1
  fi
  sleep 0.2
done
curl -fsS "$BASE_URL/" >/dev/null

cat >"$PLAYWRIGHT_TMP/package.json" <<EOF_PKG
{"private":true,"dependencies":{"playwright":"${PLAYWRIGHT_VERSION}"}}
EOF_PKG
npm --prefix "$PLAYWRIGHT_TMP" install --silent --no-audit --fund=false >/dev/null

cat >"$PLAYWRIGHT_TMP/kanban-a11y.js" <<'EOF_JS'
const { chromium } = require('playwright');
const assert = require('node:assert/strict');

const baseURL = process.env.BASE_URL;

(async () => {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();
  const consoleMessages = [];
  page.on('console', msg => consoleMessages.push({ type: msg.type(), text: msg.text() }));
  page.on('pageerror', err => consoleMessages.push({ type: 'pageerror', text: err.stack || err.message }));

  await page.goto(baseURL + '/', { waitUntil: 'networkidle' });
  await page.getByRole('heading', { name: 'Kanban benchmark' }).waitFor({ state: 'visible' });

  const card = page.locator('[data-kb-card-id="1"]').first();
  await card.waitFor({ state: 'visible' });
  assert.equal(await card.getAttribute('role'), 'listitem');
  assert.equal(await card.getAttribute('tabindex'), '0');
  assert.match(await card.getAttribute('aria-label'), /Card 1, To Do, position \d+ of \d+/);

  await card.focus();
  await page.keyboard.press('Enter');
  const menu = page.locator('[data-kb-action-menu]');
  await menu.waitFor({ state: 'visible' });
  assert.equal(await menu.getAttribute('role'), 'menu');
  await page.keyboard.press('ArrowDown');
  await page.keyboard.press('Escape');
  await menu.waitFor({ state: 'detached' });

  const trigger = card.locator('[data-kb-card-actions]').first();
  await trigger.click();
  await menu.waitFor({ state: 'visible' });
  await Promise.all([
    page.waitForResponse(resp => resp.url().includes('/_kanban/bench/action/cardMoved') && resp.status() === 200),
    page.getByRole('menuitem', { name: 'Move to Done' }).click(),
  ]);

  const moved = page.locator('[data-kb-card-id="1"]').first();
  await moved.waitFor({ state: 'visible' });
  await page.waitForFunction(() => {
    const el = document.querySelector('[data-kb-card-id="1"]');
    return el && el.dataset.kbCardColumn === 'done';
  });
  await page.waitForFunction(() => document.activeElement && document.activeElement.dataset.kbCardId === '1');
  const liveText = await page.locator('[data-kb-live-region]').textContent();
  assert.match(liveText || '', /Moved card 1/);

  const badMessages = consoleMessages.filter(m => ['error', 'warning', 'pageerror'].includes(m.type));
  assert.deepEqual(badMessages, [], 'Browser console should not contain warnings/errors');
  await browser.close();
})().catch(err => {
  console.error(err);
  process.exit(1);
});
EOF_JS

BASE_URL="$BASE_URL" node "$PLAYWRIGHT_TMP/kanban-a11y.js"
echo "kanban accessibility playwright smoke passed: $BASE_URL"
