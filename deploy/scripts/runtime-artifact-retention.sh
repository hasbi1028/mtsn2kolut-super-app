#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
MODE="dry-run"
LOG_DIR="${MTSN2KOLUT_LOG_DIR:-/home/servermtsn2kolut/logs/mtsn2kolut-super-app}"
SCREENSHOT_DIR="${MTSN2KOLUT_SCREENSHOT_DIR:-${LOG_DIR}/screenshots}"
BINARY_BACKUP_DIR="${MTSN2KOLUT_BINARY_BACKUP_DIR:-/home/servermtsn2kolut/backups/mtsn2kolut-super-app/binaries}"
LOG_RETENTION_DAYS="${MTSN2KOLUT_LOG_RETENTION_DAYS:-14}"
SCREENSHOT_RETENTION_DAYS="${MTSN2KOLUT_SCREENSHOT_RETENTION_DAYS:-14}"
BINARY_RETENTION_DAYS="${MTSN2KOLUT_BINARY_RETENTION_DAYS:-30}"

usage() {
  cat <<'USAGE'
Usage: deploy/scripts/runtime-artifact-retention.sh [options]

Dry-run by default. This script only audits/deletes old runtime artifacts from
operator-owned paths outside the repository checkout. It refuses repository
paths so production files are not moved or removed accidentally.

Options:
  --apply                       Delete matched files. Default is dry-run.
  --log-dir DIR                 PM2/application log directory.
  --screenshot-dir DIR          Worker screenshot directory.
  --binary-backup-dir DIR       Binary backup directory.
  --log-days N                  Retention for rotated logs. Default: 14.
  --screenshot-days N           Retention for worker screenshots. Default: 14.
  --binary-days N               Retention for binary backups. Default: 30.
  -h, --help                    Show help.

Recommended paths:
  logs       /home/servermtsn2kolut/logs/mtsn2kolut-super-app
  screenshots /home/servermtsn2kolut/logs/mtsn2kolut-super-app/screenshots
  binaries   /home/servermtsn2kolut/backups/mtsn2kolut-super-app/binaries
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --apply) MODE="apply"; shift ;;
    --log-dir) LOG_DIR="${2:-}"; shift 2 ;;
    --screenshot-dir) SCREENSHOT_DIR="${2:-}"; shift 2 ;;
    --binary-backup-dir) BINARY_BACKUP_DIR="${2:-}"; shift 2 ;;
    --log-days) LOG_RETENTION_DAYS="${2:-}"; shift 2 ;;
    --screenshot-days) SCREENSHOT_RETENTION_DAYS="${2:-}"; shift 2 ;;
    --binary-days) BINARY_RETENTION_DAYS="${2:-}"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage >&2; exit 2 ;;
  esac
done

absolute_path() {
  realpath -m "$1"
}

require_days() {
  local label="$1"
  local value="$2"
  if [[ ! "$value" =~ ^[0-9]+$ || "$value" == "0" ]]; then
    printf 'ERROR: %s must be a positive integer, got %s\n' "$label" "$value" >&2
    exit 1
  fi
}

refuse_repo_path() {
  local label="$1"
  local raw="$2"
  local resolved
  resolved="$(absolute_path "$raw")"
  if [[ "$resolved" == "/" || "$resolved" == "$ROOT_DIR" || "$resolved" == "$ROOT_DIR"/* ]]; then
    printf 'ERROR: refusing %s inside repository checkout: %s\n' "$label" "$resolved" >&2
    exit 1
  fi
}

run_find() {
  local label="$1"
  local dir="$2"
  local days="$3"
  shift 3

  local resolved
  resolved="$(absolute_path "$dir")"
  if [[ ! -d "$resolved" ]]; then
    printf '[skip] %s directory does not exist: %s\n' "$label" "$resolved"
    return
  fi

  printf '[scan] %s older than %s days in %s\n' "$label" "$days" "$resolved"
  if [[ "$MODE" == "apply" ]]; then
    find "$resolved" "$@" -mtime +"$days" -print -delete
  else
    find "$resolved" "$@" -mtime +"$days" -print
  fi
}

require_days "--log-days" "$LOG_RETENTION_DAYS"
require_days "--screenshot-days" "$SCREENSHOT_RETENTION_DAYS"
require_days "--binary-days" "$BINARY_RETENTION_DAYS"
refuse_repo_path "--log-dir" "$LOG_DIR"
refuse_repo_path "--screenshot-dir" "$SCREENSHOT_DIR"
refuse_repo_path "--binary-backup-dir" "$BINARY_BACKUP_DIR"

printf 'Runtime artifact retention mode: %s\n' "$MODE"
printf '  log dir           : %s\n' "$(absolute_path "$LOG_DIR")"
printf '  screenshot dir    : %s\n' "$(absolute_path "$SCREENSHOT_DIR")"
printf '  binary backup dir : %s\n' "$(absolute_path "$BINARY_BACKUP_DIR")"

run_find "rotated logs" "$LOG_DIR" "$LOG_RETENTION_DAYS" \
  -type f \( -name '*.log.*' -o -name '*.log-*.gz' -o -name '*.gz' -o -name '*.old' \)

run_find "worker screenshots" "$SCREENSHOT_DIR" "$SCREENSHOT_RETENTION_DAYS" \
  -type f \( -name '*.png' -o -name '*.jpg' -o -name '*.jpeg' -o -name '*.webp' -o -name '*.txt' -o -name '*.json' -o -name '*.zip' \)

run_find "binary backups" "$BINARY_BACKUP_DIR" "$BINARY_RETENTION_DAYS" \
  -type f \( -name 'api.backup-*' -o -name '*.backup-*' -o -name '*.bak' -o -name '*.tar.gz' -o -name '*.zip' \)

if [[ "$MODE" != "apply" ]]; then
  echo "Dry-run only; no files deleted. Re-run with --apply after reviewing the list."
fi
