#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

check() {
  local name="$1"
  local pid_file="$ROOT_DIR/logs/$2.pid"
  if [ -f "$pid_file" ]; then
    PID=$(cat "$pid_file")
    if kill -0 "$PID" 2>/dev/null; then
      echo "$name: running (PID $PID)"
    else
      echo "$name: stopped (stale PID $PID)"
    fi
  else
    echo "$name: stopped"
  fi
}

check "Frontend" "frontend"
check "Worker  " "worker"
