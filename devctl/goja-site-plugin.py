#!/usr/bin/env python3
"""devctl plugin for the goja-site local development server.

Protocol rule: stdout is NDJSON only. Human logs go to stderr.
"""

from __future__ import annotations

import json
import os
import shutil
import sys
from pathlib import Path
from typing import Any

PLUGIN_NAME = "goja-site"
DEFAULT_ADDR = "127.0.0.1:60131"
DEFAULT_DB = "tmp/form-generator.db"
DEFAULT_SCRIPTS = "sites/forms/scripts"
DEFAULT_METRICS_ADDR = "127.0.0.1:60132"
DEFAULT_BASE_URL = "http://127.0.0.1:60131"


def emit(obj: dict[str, Any]) -> None:
    sys.stdout.write(json.dumps(obj, separators=(",", ":")) + "\n")
    sys.stdout.flush()


def log(message: str) -> None:
    sys.stderr.write(f"[{PLUGIN_NAME}] {message}\n")
    sys.stderr.flush()


def response_ok(request_id: str, output: dict[str, Any]) -> None:
    emit({"type": "response", "request_id": request_id, "ok": True, "output": output})


def response_error(request_id: str, code: str, message: str) -> None:
    emit({
        "type": "response",
        "request_id": request_id,
        "ok": False,
        "error": {"code": code, "message": message},
    })


def repo_root(ctx: dict[str, Any]) -> Path:
    return Path(str(ctx.get("repo_root") or os.getcwd())).resolve()


def merged_env(config: dict[str, Any] | None = None) -> dict[str, str]:
    env = {
        "GOJA_SITE_ADDR": DEFAULT_ADDR,
        "GOJA_SITE_DB": DEFAULT_DB,
        "GOJA_SITE_SCRIPTS": DEFAULT_SCRIPTS,
        "GOJA_SITE_METRICS_ADDR": DEFAULT_METRICS_ADDR,
        "GOJA_SITE_BASE_URL": DEFAULT_BASE_URL,
    }
    if config:
        config_env = config.get("env")
        if isinstance(config_env, dict):
            env.update({str(k): str(v) for k, v in config_env.items()})
    env.update({key: os.environ[key] for key in env.keys() if key in os.environ})
    return env


def handle_config_mutate(request_id: str) -> None:
    response_ok(request_id, {
        "config_patch": {
            "set": {
                "env.GOJA_SITE_ADDR": DEFAULT_ADDR,
                "env.GOJA_SITE_DB": DEFAULT_DB,
                "env.GOJA_SITE_SCRIPTS": DEFAULT_SCRIPTS,
                "env.GOJA_SITE_METRICS_ADDR": DEFAULT_METRICS_ADDR,
                "env.GOJA_SITE_BASE_URL": DEFAULT_BASE_URL,
                "services.goja-site.port": "60131",
                "services.goja-site.url": DEFAULT_BASE_URL,
                "services.goja-site.db": DEFAULT_DB,
                "services.goja-site.scripts": DEFAULT_SCRIPTS,
                "services.goja-site.metrics_url": "http://127.0.0.1:60132/metrics",
            },
            "unset": [],
        }
    })


def handle_validate(request_id: str, ctx: dict[str, Any], input_obj: dict[str, Any]) -> None:
    root = repo_root(ctx)
    config = input_obj.get("config") if isinstance(input_obj.get("config"), dict) else {}
    env = merged_env(config)
    errors: list[dict[str, str]] = []
    warnings: list[dict[str, str]] = []

    if shutil.which("go") is None:
        errors.append({"code": "E_MISSING_TOOL", "message": "missing required tool: go"})

    for rel in ("go.mod", "cmd/goja-site/main.go"):
        if not (root / rel).exists():
            errors.append({"code": "E_MISSING_FILE", "message": f"missing expected file: {rel}"})

    scripts_dir = root / env["GOJA_SITE_SCRIPTS"]
    if not scripts_dir.is_dir():
        errors.append({"code": "E_MISSING_SITE_DIR", "message": f"missing form generator scripts directory: {env['GOJA_SITE_SCRIPTS']}"})
    elif not list(scripts_dir.glob("*.js")):
        warnings.append({"code": "W_NO_SCRIPTS", "message": f"no JavaScript files found in {env['GOJA_SITE_SCRIPTS']}"})

    response_ok(request_id, {"valid": len(errors) == 0, "errors": errors, "warnings": warnings})


def handle_launch_plan(request_id: str, ctx: dict[str, Any], input_obj: dict[str, Any]) -> None:
    config = input_obj.get("config") if isinstance(input_obj.get("config"), dict) else {}
    env = merged_env(config)
    root = repo_root(ctx)
    db_path = env["GOJA_SITE_DB"]
    scripts_dir = env["GOJA_SITE_SCRIPTS"]
    addr = env["GOJA_SITE_ADDR"]
    metrics_addr = env["GOJA_SITE_METRICS_ADDR"]
    base_url = env["GOJA_SITE_BASE_URL"].rstrip("/")

    command = (
        "mkdir -p tmp .devctl/logs && "
        "exec go run ./cmd/goja-site serve "
        f"--addr {addr} "
        f"--db {db_path} "
        f"--scripts {scripts_dir} "
        "--dev "
        f"--metrics-addr {metrics_addr} "
        "--service-name goja-site-devctl"
    )

    if ctx.get("dry_run"):
        log("dry-run: returning launch plan without starting services")

    response_ok(request_id, {
        "services": [
            {
                "name": "goja-site",
                "cwd": ".",
                "command": ["bash", "--noprofile", "--norc", "-lc", command],
                "env": {
                    "GOJA_SITE_ADDR": addr,
                    "GOJA_SITE_DB": db_path,
                    "GOJA_SITE_SCRIPTS": scripts_dir,
                    "GOJA_SITE_METRICS_ADDR": metrics_addr,
                    "GOJA_SITE_BASE_URL": base_url,
                },
                "health": {"type": "http", "url": base_url + "/", "timeout_ms": 60000},
            }
        ],
        "notes": [
            f"Local multi-site server: {base_url}",
            "Form generator home: " + base_url + "/",
            f"Metrics: http://{metrics_addr}/metrics",
            f"Scripts: {root / scripts_dir}",
            f"SQLite DB: {root / db_path}",
        ],
    })


emit({
    "type": "handshake",
    "protocol_version": "v2",
    "plugin_name": PLUGIN_NAME,
    "capabilities": {"ops": ["config.mutate", "validate.run", "launch.plan"]},
})

for raw_line in sys.stdin:
    line = raw_line.strip()
    if not line:
        continue
    try:
        req = json.loads(line)
        request_id = str(req.get("request_id") or "")
        op = str(req.get("op") or "")
        ctx = req.get("ctx") if isinstance(req.get("ctx"), dict) else {}
        input_obj = req.get("input") if isinstance(req.get("input"), dict) else {}

        if op == "config.mutate":
            handle_config_mutate(request_id)
        elif op == "validate.run":
            handle_validate(request_id, ctx, input_obj)
        elif op == "launch.plan":
            handle_launch_plan(request_id, ctx, input_obj)
        else:
            response_error(request_id, "E_UNSUPPORTED", f"unsupported op: {op}")
    except Exception as exc:  # noqa: BLE001 - plugin boundary should convert all errors to protocol errors.
        response_error("", "E_PROTOCOL", str(exc))
