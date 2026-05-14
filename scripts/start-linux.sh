#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
LOG_DIR="${LOG_DIR:-$ROOT_DIR/logs}"
mkdir -p "$LOG_DIR"

export KIRO_API_KEY="${KIRO_API_KEY:?KIRO_API_KEY is required}"
export KIRO_CLI_PATH="${KIRO_CLI_PATH:-/usr/local/bin/kiro-cli}"
export LISTEN_ADDR="${LISTEN_ADDR:-:8080}"

exec "$ROOT_DIR/bin/kiro-proxy" >> "$LOG_DIR/proxy.log" 2>&1
