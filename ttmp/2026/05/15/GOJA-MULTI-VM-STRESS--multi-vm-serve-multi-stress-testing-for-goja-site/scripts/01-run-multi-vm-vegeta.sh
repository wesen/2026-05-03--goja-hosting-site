#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"

VM_COUNT="1"
SCENARIO="null"
DISTRIBUTION="even-hot"
RATE="100/s"
DURATION="10s"
WARMUP_DURATION="3s"
PORT="18800"
METRICS_PORT="19800"
OUT_DIR=""
BINARY=""
CAPTURE_PPROF="0"
PPROF_SECONDS="10"
VEGETA_BIN="${VEGETA_BIN:-vegeta}"

usage() {
  cat <<'EOF'
Usage: 01-run-multi-vm-vegeta.sh [options]

Options:
  --vm-count N           Number of serve-multi sites/Goja VMs (default: 1)
  --scenario NAME        null | render | db-read | db-write | kanban-fragment | kanban-action (default: null)
  --distribution NAME    even-hot | one-hot | skewed (default: even-hot)
  --rate RATE            Total offered Vegeta rate across target file (default: 100/s)
  --duration DURATION    Measured duration (default: 10s)
  --warmup-duration D    Warmup duration (default: 3s)
  --port PORT            App port (default: 18800)
  --metrics-port PORT    Diagnostics port (default: 19800)
  --out-dir DIR          Output directory (default: bench/results/multi-vm-<stamp>)
  --binary PATH          Existing goja-site binary to reuse
  --pprof                Capture pprof artifacts from diagnostics listener
  --pprof-seconds N      CPU profile seconds for --pprof (default: 10)
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --vm-count) VM_COUNT="$2"; shift 2 ;;
    --scenario) SCENARIO="$2"; shift 2 ;;
    --distribution) DISTRIBUTION="$2"; shift 2 ;;
    --rate) RATE="$2"; shift 2 ;;
    --duration) DURATION="$2"; shift 2 ;;
    --warmup-duration) WARMUP_DURATION="$2"; shift 2 ;;
    --port) PORT="$2"; shift 2 ;;
    --metrics-port) METRICS_PORT="$2"; shift 2 ;;
    --out-dir) OUT_DIR="$2"; shift 2 ;;
    --binary) BINARY="$2"; shift 2 ;;
    --pprof) CAPTURE_PPROF="1"; shift ;;
    --pprof-seconds) PPROF_SECONDS="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "unknown option: $1" >&2; usage >&2; exit 2 ;;
  esac
done

if ! [[ "$VM_COUNT" =~ ^[0-9]+$ ]] || [[ "$VM_COUNT" -lt 1 ]]; then
  echo "--vm-count must be a positive integer" >&2
  exit 2
fi
if ! command -v "$VEGETA_BIN" >/dev/null 2>&1; then
  echo "vegeta not found; install github.com/tsenart/vegeta/v12" >&2
  exit 127
fi

STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
if [[ -z "$OUT_DIR" ]]; then
  safe_rate="$(echo "$RATE" | tr '/:' '__')"
  OUT_DIR="bench/results/multi-vm-${STAMP}-${SCENARIO}-${VM_COUNT}vm-${DISTRIBUTION}-${safe_rate}"
fi
mkdir -p "$OUT_DIR"
OUT_DIR="$(cd "$OUT_DIR" && pwd)"
TMP_DIR="$(mktemp -d -t goja-site-multi-vm.XXXXXX)"
APP_ADDR="127.0.0.1:${PORT}"
BASE_URL="http://${APP_ADDR}"
METRICS_ADDR="127.0.0.1:${METRICS_PORT}"
METRICS_URL="http://${METRICS_ADDR}/metrics"
LOG_PATH="$OUT_DIR/server.log"
CONFIG_PATH="$OUT_DIR/sites.yaml"
TARGETS_PATH="$OUT_DIR/targets.txt"
PPROF_ARGS=()
if [[ "$CAPTURE_PPROF" == "1" ]]; then
  PPROF_ARGS=(--pprof)
fi

cleanup() {
  local status=$?
  if [[ -n "${SERVER_PID:-}" ]] && kill -0 "$SERVER_PID" 2>/dev/null; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
  rm -rf "$TMP_DIR"
  if [[ $status -ne 0 ]]; then
    echo "multi-vm output dir: $OUT_DIR" >&2
    [[ -f "$LOG_PATH" ]] && tail -200 "$LOG_PATH" >&2 || true
  fi
}
trap cleanup EXIT

if [[ -z "$BINARY" ]]; then
  BINARY="$TMP_DIR/goja-site"
  go build -o "$BINARY" ./cmd/goja-site
fi

script_dir_for_scenario() {
  case "$SCENARIO" in
    null) echo "bench/scripts/null-route" ;;
    render) echo "bench/scripts/render-route" ;;
    db-read|db-write) echo "bench/scripts/db-read-write" ;;
    kanban-fragment|kanban-action) echo "bench/scripts/kanban-board" ;;
    *) echo "unsupported scenario: $SCENARIO" >&2; exit 2 ;;
  esac
}

write_config() {
  local script_dir="$1"
  cat >"$CONFIG_PATH" <<EOF_CFG
addr: "${APP_ADDR}"
dataDir: "${TMP_DIR}/multi-data"
baseDomain: "multi-vm.bench.test"
dev: false
sites:
EOF_CFG
  for i in $(seq 1 "$VM_COUNT"); do
    name="site-$(printf '%03d' "$i")"
    cat >>"$CONFIG_PATH" <<EOF_SITE
  - name: ${name}
    host: ${name}.multi-vm.bench.test
    dbPolicy: simple
    allowWrites: true
    scripts:
      - ${script_dir}
EOF_SITE
  done
}

emit_target_for_site() {
  local i="$1"
  local host="site-$(printf '%03d' "$i").multi-vm.bench.test"
  case "$SCENARIO" in
    null)
      printf 'GET %s/\nHost: %s\n\n' "$BASE_URL" "$host" >>"$TARGETS_PATH"
      ;;
    render)
      printf 'GET %s/render?n=100\nHost: %s\n\n' "$BASE_URL" "$host" >>"$TARGETS_PATH"
      ;;
    db-read)
      printf 'GET %s/read?n=10\nHost: %s\n\n' "$BASE_URL" "$host" >>"$TARGETS_PATH"
      ;;
    db-write)
      local payload="$OUT_DIR/write-one.json"
      echo '{"n":1}' >"$payload"
      printf 'POST %s/write\nHost: %s\nContent-Type: application/json\n@%s\n\n' "$BASE_URL" "$host" "$payload" >>"$TARGETS_PATH"
      ;;
    kanban-fragment)
      printf 'GET %s/_kanban/bench/fragment\nHost: %s\n\n' "$BASE_URL" "$host" >>"$TARGETS_PATH"
      ;;
    kanban-action)
      local payload="$OUT_DIR/kanban-move.json"
      echo '{"cardId":"1","to":{"columnId":"done","index":0}}' >"$payload"
      printf 'POST %s/_kanban/bench/action/cardMoved\nHost: %s\nContent-Type: application/json\nCookie: goja_session=bench-session-%s\n@%s\n\n' "$BASE_URL" "$host" "$i" "$payload" >>"$TARGETS_PATH"
      ;;
  esac
}

write_targets() {
  : >"$TARGETS_PATH"
  case "$DISTRIBUTION" in
    even-hot)
      for i in $(seq 1 "$VM_COUNT"); do emit_target_for_site "$i"; done
      ;;
    one-hot)
      emit_target_for_site 1
      ;;
    skewed)
      for _ in $(seq 1 9); do emit_target_for_site 1; done
      if [[ "$VM_COUNT" -gt 1 ]]; then
        for i in $(seq 2 "$VM_COUNT"); do emit_target_for_site "$i"; done
      fi
      ;;
    *) echo "unsupported distribution: $DISTRIBUTION" >&2; exit 2 ;;
  esac
}

wait_ready() {
  for _ in $(seq 1 200); do
    if ! kill -0 "$SERVER_PID" 2>/dev/null; then
      echo "server exited before readiness" >&2
      return 1
    fi
    if curl -fsS --max-time 1 "$BASE_URL/readyz" >/dev/null 2>&1 && curl -fsS --max-time 1 "$METRICS_URL" >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.1
  done
  echo "server did not become ready" >&2
  return 1
}

capture_metrics_delta() {
  python3 - "$OUT_DIR/metrics-before.prom" "$OUT_DIR/metrics-after.prom" "$OUT_DIR/metrics-delta.txt" <<'PY'
import re, sys
metric_re = re.compile(r'^([a-zA-Z_:][a-zA-Z0-9_:]*(?:\{[^}]*\})?)\s+([-+0-9.eE]+)$')
def load(path):
    d = {}
    for line in open(path, encoding='utf-8'):
        if line.startswith('#'):
            continue
        m = metric_re.match(line.strip())
        if not m:
            continue
        try:
            d[m.group(1)] = float(m.group(2))
        except ValueError:
            pass
    return d
before, after = load(sys.argv[1]), load(sys.argv[2])
with open(sys.argv[3], 'w', encoding='utf-8') as f:
    for key in sorted(after):
        delta = after[key] - before.get(key, 0.0)
        if delta and (key.startswith('goja_site_') or key.startswith('go_') or key.startswith('process_')):
            f.write(f'{key} {delta}\n')
PY
}

script_dir="$(script_dir_for_scenario)"
write_config "$script_dir"
write_targets

"$BINARY" serve-multi --config "$CONFIG_PATH" --metrics-addr "$METRICS_ADDR" "${PPROF_ARGS[@]}" >"$LOG_PATH" 2>&1 &
SERVER_PID=$!
wait_ready

if [[ "$WARMUP_DURATION" != "0s" && "$WARMUP_DURATION" != "0" ]]; then
  "$VEGETA_BIN" attack -duration="$WARMUP_DURATION" -rate="$RATE" -targets="$TARGETS_PATH" >/dev/null
fi

curl -fsS "$METRICS_URL" >"$OUT_DIR/metrics-before.prom"
CPU_PPROF_PID=""
if [[ "$CAPTURE_PPROF" == "1" ]]; then
  # Start CPU profiling before the measured attack so samples overlap the load,
  # not the post-run idle drain period.
  curl -fsS "http://${METRICS_ADDR}/debug/pprof/profile?seconds=${PPROF_SECONDS}" -o "$OUT_DIR/cpu.pprof" &
  CPU_PPROF_PID=$!
  sleep 0.2
fi
"$VEGETA_BIN" attack -duration="$DURATION" -rate="$RATE" -targets="$TARGETS_PATH" -output="$OUT_DIR/vegeta.bin"
if [[ -n "$CPU_PPROF_PID" ]]; then
  wait "$CPU_PPROF_PID"
fi
"$VEGETA_BIN" report "$OUT_DIR/vegeta.bin" | tee "$OUT_DIR/vegeta.txt"
"$VEGETA_BIN" report -type=json "$OUT_DIR/vegeta.bin" >"$OUT_DIR/vegeta.json"
curl -fsS "$METRICS_URL" >"$OUT_DIR/metrics-after.prom"
capture_metrics_delta

if [[ "$CAPTURE_PPROF" == "1" ]]; then
  curl -fsS "http://${METRICS_ADDR}/debug/pprof/heap" -o "$OUT_DIR/heap.pprof"
  curl -fsS "http://${METRICS_ADDR}/debug/pprof/allocs" -o "$OUT_DIR/allocs.pprof"
  curl -fsS "http://${METRICS_ADDR}/debug/pprof/goroutine?debug=2" -o "$OUT_DIR/goroutine.txt"
fi

git_commit="$(git rev-parse HEAD 2>/dev/null || true)"
git_dirty="false"
if ! git diff --quiet || ! git diff --cached --quiet || [[ -n "$(git status --porcelain --untracked-files=no)" ]]; then
  git_dirty="true"
fi
cat >"$OUT_DIR/metadata.json" <<EOF_META
{
  "created_at_utc": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "git_commit": "${git_commit}",
  "git_dirty": ${git_dirty},
  "vm_count": ${VM_COUNT},
  "scenario": "${SCENARIO}",
  "distribution": "${DISTRIBUTION}",
  "rate": "${RATE}",
  "duration": "${DURATION}",
  "warmup_duration": "${WARMUP_DURATION}",
  "app_addr": "${APP_ADDR}",
  "metrics_addr": "${METRICS_ADDR}",
  "config_path": "${CONFIG_PATH}",
  "targets_path": "${TARGETS_PATH}"
}
EOF_META

echo "multi-vm benchmark complete: $OUT_DIR"
