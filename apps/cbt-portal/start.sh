#!/usr/bin/env bash
set -a
if [ -f "$(dirname "$0")/.env" ]; then
  # shellcheck source=.env
  source "$(dirname "$0")/.env"
fi
set +a
export HOST="${HOST:-0.0.0.0}"
export PORT="${PORT:-8031}"
export NODE_ENV="${NODE_ENV:-production}"
exec node "$(dirname "$0")/build/index.js"
