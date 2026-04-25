#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
PID_FILE="$ROOT_DIR/logs/worker.pid"

if [ ! -f "$PID_FILE" ]; then
  echo "Worker not running (no PID file)"
  exit 0
fi

PID=$(cat "$PID_FILE")
if kill -0 "$PID" 2>/dev/null; then
  kill "$PID"
  echo "Worker stopped (PID $PID)"
else
  echo "Worker not running (stale PID $PID)"
fi

rm -f "$PID_FILE"
