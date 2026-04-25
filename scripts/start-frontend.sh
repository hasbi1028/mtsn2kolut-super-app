#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
PID_FILE="$ROOT_DIR/logs/frontend.pid"
LOG_OUT="$ROOT_DIR/logs/frontend-out.log"
LOG_ERR="$ROOT_DIR/logs/frontend-error.log"
ENV_FILE="$ROOT_DIR/frontend/.env"

mkdir -p "$ROOT_DIR/logs"

if [ -f "$PID_FILE" ]; then
  PID=$(cat "$PID_FILE")
  if kill -0 "$PID" 2>/dev/null; then
    echo "Frontend already running (PID $PID)"
    exit 0
  fi
  rm -f "$PID_FILE"
fi

[ -f "$ENV_FILE" ] && set -a && source "$ENV_FILE" && set +a

export HOST="${HOST:-0.0.0.0}"
export PORT="${PORT:-8021}"
export NODE_ENV="${NODE_ENV:-production}"
export DB_PATH="${DB_PATH:-$ROOT_DIR/data/pusaka.sqlite}"

mkdir -p "$(dirname "${DB_PATH:-}")" 2>/dev/null || true

cd "$ROOT_DIR/frontend"
nohup node build/index.js >> "$LOG_OUT" 2>> "$LOG_ERR" &
echo $! > "$PID_FILE"
echo "Frontend started (PID $(cat "$PID_FILE"))"
