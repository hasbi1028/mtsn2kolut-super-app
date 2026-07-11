#!/usr/bin/env bash
set -a
# shellcheck source=.env
source "$(dirname "$0")/.env"
set +a
export PORT="${PORT:-8022}"
export HOST="${HOST:-0.0.0.0}"
export NODE_ENV=production
exec node "$(dirname "$0")/build/index.js"
