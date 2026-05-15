PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS benchmark_matrices (
  matrix_id TEXT PRIMARY KEY,
  created_at_utc TEXT NOT NULL,
  imported_at_utc TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
  out_root TEXT NOT NULL,
  repo_commit TEXT NOT NULL,
  git_dirty INTEGER NOT NULL DEFAULT 0,
  duration TEXT NOT NULL,
  warmup_duration TEXT NOT NULL,
  scenarios TEXT NOT NULL,
  rates TEXT NOT NULL,
  repeat_count INTEGER NOT NULL,
  command_line TEXT NOT NULL,
  source_summary_json TEXT NOT NULL,
  source_summary_md TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS benchmark_runs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  matrix_id TEXT NOT NULL REFERENCES benchmark_matrices(matrix_id) ON DELETE CASCADE,
  scenario TEXT NOT NULL,
  rate_target TEXT NOT NULL,
  run_number INTEGER NOT NULL,
  out_dir TEXT NOT NULL,
  created_at_utc TEXT,
  duration TEXT NOT NULL,
  warmup_duration TEXT NOT NULL,
  requests INTEGER NOT NULL,
  throughput REAL NOT NULL,
  success_ratio REAL NOT NULL,
  p50_ms REAL NOT NULL,
  p95_ms REAL NOT NULL,
  p99_ms REAL NOT NULL,
  max_ms REAL NOT NULL,
  bytes_in_total INTEGER NOT NULL DEFAULT 0,
  bytes_out_total INTEGER NOT NULL DEFAULT 0,
  status_codes_json TEXT NOT NULL,
  errors_json TEXT NOT NULL,
  metadata_json TEXT NOT NULL,
  vegeta_json TEXT NOT NULL,
  metrics_delta_text TEXT NOT NULL,
  UNIQUE(matrix_id, scenario, rate_target, run_number)
);

CREATE INDEX IF NOT EXISTS idx_benchmark_runs_matrix_scenario_rate
  ON benchmark_runs(matrix_id, scenario, rate_target);

CREATE TABLE IF NOT EXISTS benchmark_metric_deltas (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id INTEGER NOT NULL REFERENCES benchmark_runs(id) ON DELETE CASCADE,
  matrix_id TEXT NOT NULL,
  scenario TEXT NOT NULL,
  rate_target TEXT NOT NULL,
  metric_name TEXT NOT NULL,
  labels_json TEXT NOT NULL,
  delta_value REAL NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_metric_deltas_matrix_name
  ON benchmark_metric_deltas(matrix_id, metric_name);

CREATE INDEX IF NOT EXISTS idx_metric_deltas_run
  ON benchmark_metric_deltas(run_id);
