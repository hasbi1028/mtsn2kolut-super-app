#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
PID_FILE="$ROOT_DIR/logs/worker.pid"
LOG_OUT="$ROOT_DIR/logs/worker-out.log"
LOG_ERR="$ROOT_DIR/logs/worker-error.log"
ENV_FILE="$ROOT_DIR/worker/.env"

mkdir -p "$ROOT_DIR/logs"

if [ -f "$PID_FILE" ]; then
  PID=$(cat "$PID_FILE")
  if kill -0 "$PID" 2>/dev/null; then
    echo "Worker already running (PID $PID)"
    exit 0
  fi
  rm -f "$PID_FILE"
fi

[ -f "$ENV_FILE" ] && set -a && source "$ENV_FILE" && set +a

export FRONTEND_URL="${FRONTEND_URL:-http://localhost:8021}"
export WORKER_TOKEN="${WORKER_TOKEN:-}"
export WORKER_ID="${WORKER_ID:-worker-$(hostname)-$$}"
export WORKER_CONCURRENCY="${WORKER_CONCURRENCY:-5}"
export HEADLESS="${HEADLESS:-true}"
export POLL_MS="${POLL_MS:-8000}"
export SCRAPE_RETRIES="${SCRAPE_RETRIES:-3}"
export SCRAPE_RETRY_MS="${SCRAPE_RETRY_MS:-5000}"
export ACTION_TIMEOUT="${ACTION_TIMEOUT:-20000}"
export WORKER_LOG_PATH="${WORKER_LOG_PATH:-$ROOT_DIR/logs/worker.log}"
export SCREENSHOT_DIR="${SCREENSHOT_DIR:-$ROOT_DIR/logs/screenshots}"

mkdir -p "$SCREENSHOT_DIR"

TSX="$ROOT_DIR/worker/node_modules/.bin/tsx"
if [ ! -f "$TSX" ]; then
  TSX="$(command -v tsx 2>/dev/null || true)"
fi
if [ -z "$TSX" ]; then
  echo "Error: tsx not found. Run: cd worker && npm install" >&2
  exit 1
fi

cd "$ROOT_DIR/worker"
nohup "$TSX" src/index.ts >> "$LOG_OUT" 2>> "$LOG_ERR" &
echo $! > "$PID_FILE"
echo "Worker started (PID $(cat "$PID_FILE"))"
