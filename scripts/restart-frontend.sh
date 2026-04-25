#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

"$SCRIPT_DIR/stop-frontend.sh"
sleep 1
"$SCRIPT_DIR/start-frontend.sh"
