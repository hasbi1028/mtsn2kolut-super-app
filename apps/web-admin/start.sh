#!/usr/bin/env bash
set -a
# shellcheck source=.env
source "$(dirname "$0")/.env"
set +a
exec node "$(dirname "$0")/build/index.js"
