#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PORT="${PORT:-19111}"
BASE_URL="http://127.0.0.1:${PORT}"
DB_PATH="${DB_PATH:-$(mktemp -t goja-site-kanban-smoke.XXXXXX.sqlite)}"
LOG_PATH="${LOG_PATH:-$(mktemp -t goja-site-kanban-smoke.XXXXXX.log)}"
KEEP_DB="${KEEP_DB:-0}"
PLAYWRIGHT_VERSION="${PLAYWRIGHT_VERSION:-1.59.1}"
PLAYWRIGHT_TMP="${PLAYWRIGHT_TMP:-$(mktemp -d -t goja-site-playwright.XXXXXX)}"

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
    if [[ -f "$LOG_PATH" ]]; then
      tail -200 "$LOG_PATH" >&2 || true
    fi
  fi
}
trap cleanup EXIT

cd "$ROOT"
rm -f "$DB_PATH"

go run ./cmd/goja-site serve \
  --db "$DB_PATH" \
  --scripts examples/kanban/scripts \
  --db-policy guarded \
  --addr ":${PORT}" \
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

cat >"$PLAYWRIGHT_TMP/kanban-smoke.js" <<'EOF_JS'
const { chromium } = require('playwright');
const assert = require('node:assert/strict');

const baseURL = process.env.BASE_URL;

async function visibleText(locator) {
  return (await locator.textContent()) || '';
}

(async () => {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();
  const consoleMessages = [];
  page.on('console', msg => consoleMessages.push({ type: msg.type(), text: msg.text() }));
  page.on('pageerror', err => consoleMessages.push({ type: 'pageerror', text: err.stack || err.message }));

  await page.goto(baseURL + '/', { waitUntil: 'networkidle' });
  await assertTitle(page, 'Trail Notes: Cascade Loop');
  await assertVisible(page, 'Field Notes');
  await assertVisible(page, 'Research campsites');
  await assertVisible(page, 'Map the route');
  await assertColumnCount(page, 'To Do', '3');
  await assertColumnCount(page, 'In Progress', '2');
  await assertColumnCount(page, 'Done', '3');
  await assertColumnCount(page, 'Someday', '2');

  const statusSelect = page.getByLabel('Filter status');
  await assert.equal(await statusSelect.inputValue(), '', 'All columns filter should submit an empty value');

  await page.getByRole('textbox', { name: 'Card title' }).fill('Playwright E2E card');
  await page.getByRole('textbox', { name: 'Description' }).fill('Created by browser automation');
  await page.locator('form.new-card select[name="status"]').selectOption('progress');
  await page.getByRole('textbox', { name: 'Tag' }).fill('Testing');
  await Promise.all([
    page.waitForURL(baseURL + '/', { waitUntil: 'networkidle' }),
    page.getByRole('button', { name: 'Add card' }).click(),
  ]);
  await assertVisible(page, 'Playwright E2E card');
  await assertColumnCount(page, 'In Progress', '3');

  await page.getByRole('textbox', { name: 'Search notes...' }).fill('Playwright');
  await Promise.all([
    page.waitForURL(/search=Playwright/, { waitUntil: 'networkidle' }),
    page.getByRole('button', { name: 'Search' }).click(),
  ]);
  assert.equal(new URL(page.url()).searchParams.get('status'), '', 'Search form should not submit the All columns label');
  await assertVisible(page, 'Playwright E2E card');
  await assertColumnCount(page, 'In Progress', '1');
  await assertColumnCount(page, 'To Do', '0');
  await assertColumnCount(page, 'Done', '0');

  const card = page.locator('article', { hasText: 'Playwright E2E card' });
  await card.focus();
  await page.keyboard.press('Enter');
  await page.getByRole('menuitem', { name: 'Move to Done' }).click();
  await page.waitForTimeout(500);
  await page.goto(baseURL + '/', { waitUntil: 'networkidle' });
  await assertVisible(page, 'Playwright E2E card');
  await assertColumnCount(page, 'Done', '4');

  const badMessages = consoleMessages.filter(m => ['error', 'warning', 'pageerror'].includes(m.type));
  assert.deepEqual(badMessages, [], 'Browser console should not contain warnings/errors');
  await browser.close();
})().catch(err => {
  console.error(err);
  process.exit(1);
});

async function assertTitle(page, expected) {
  assert.equal(await page.title(), expected);
}

async function assertVisible(page, text) {
  await page.getByText(text, { exact: false }).first().waitFor({ state: 'visible' });
}

async function assertColumnCount(page, columnTitle, expected) {
  const column = page.locator('.column', { has: page.getByRole('heading', { name: columnTitle }) });
  await column.waitFor({ state: 'visible' });
  const count = (await visibleText(column.locator('.count').first())).trim();
  assert.equal(count, expected, `${columnTitle} count`);
}
EOF_JS

BASE_URL="$BASE_URL" node "$PLAYWRIGHT_TMP/kanban-smoke.js"
echo "kanban playwright smoke passed: $BASE_URL"
