#!/usr/bin/env bash
set -euo pipefail

# Guarded runner for the official UTS Genap 2025/2026 production-like seed.
# This script intentionally defaults to DRY RUN and requires explicit confirmation
# before it can apply data changes.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SEED_BASE="$ROOT_DIR/scripts/seed-uts-genap-2025-2026.sql"
SEED_SESSIONS="$ROOT_DIR/scripts/seed-uts-genap-2025-2026-sessions.sql"
MODE="dry-run"
EXPECTED_DB_NAME="${EXPECTED_DB_NAME:-}"
CONFIRM_VALUE="SETUJU_SEED_UTS_GENAP_2025_2026"

usage() {
  cat <<USAGE
Usage: $0 [--dry-run|--apply]

Environment required:
  DATABASE_URL                 PostgreSQL connection string (not printed)

Apply mode additionally requires:
  CONFIRM_SEED_UTS_GENAP=$CONFIRM_VALUE
  CONFIRM_BACKUP_CREATED=1
  EXPECTED_DB_NAME=<database name returned by current_database()>

Examples:
  DATABASE_URL='[REDACTED]' $0 --dry-run
  EXPECTED_DB_NAME=mtsn2kolut CONFIRM_BACKUP_CREATED=1 CONFIRM_SEED_UTS_GENAP=$CONFIRM_VALUE DATABASE_URL='[REDACTED]' $0 --apply
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run) MODE="dry-run" ;;
    --apply) MODE="apply" ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown argument: $1" >&2; usage >&2; exit 2 ;;
  esac
  shift
done

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "DATABASE_URL is required but will not be printed." >&2
  exit 2
fi

for file in "$SEED_BASE" "$SEED_SESSIONS"; do
  if [[ ! -f "$file" ]]; then
    echo "Missing seed file: $file" >&2
    exit 2
  fi
done

DB_NAME="$(psql "$DATABASE_URL" -Atc "SELECT current_database()" 2>/dev/null)"
ACADEMIC_YEAR_COUNT="$(psql "$DATABASE_URL" -Atc "SELECT COUNT(*) FROM academic_years WHERE name = '2025/2026'" 2>/dev/null)"
CLASS_COUNT="$(psql "$DATABASE_URL" -Atc "SELECT COUNT(*) FROM school_classes sc JOIN academic_years ay ON ay.id = sc.academic_year_id WHERE ay.name = '2025/2026' AND sc.is_active = TRUE AND sc.level IN ('VII','VIII')" 2>/dev/null)"
STUDENT_COUNT="$(psql "$DATABASE_URL" -Atc "SELECT COUNT(*) FROM students st JOIN school_classes sc ON sc.id = st.class_id JOIN academic_years ay ON ay.id = sc.academic_year_id WHERE ay.name = '2025/2026' AND sc.level IN ('VII','VIII') AND st.is_active = TRUE" 2>/dev/null)"
EVENT_COUNT="$(psql "$DATABASE_URL" -Atc "SELECT COUNT(*) FROM cbt_exam_events e JOIN academic_years ay ON ay.id = e.academic_year_id WHERE ay.name = '2025/2026' AND e.title = 'UTS Genap'" 2>/dev/null)"

cat <<SUMMARY
Seed preflight summary:
  database: $DB_NAME
  academic_year_2025_2026_rows: $ACADEMIC_YEAR_COUNT
  active_target_classes_VII_VIII: $CLASS_COUNT
  active_target_students_VII_VIII: $STUDENT_COUNT
  existing_UTS_Genap_events: $EVENT_COUNT
  mode: $MODE
SUMMARY

if [[ "$MODE" == "dry-run" ]]; then
  echo "DRY RUN only. No SQL seed executed. Re-run with --apply plus confirmation env vars to apply."
  exit 0
fi

if [[ "${CONFIRM_SEED_UTS_GENAP:-}" != "$CONFIRM_VALUE" ]]; then
  echo "Refusing apply: set CONFIRM_SEED_UTS_GENAP=$CONFIRM_VALUE" >&2
  exit 2
fi

if [[ "${CONFIRM_BACKUP_CREATED:-}" != "1" ]]; then
  echo "Refusing apply: create/verify backup first, then set CONFIRM_BACKUP_CREATED=1" >&2
  exit 2
fi

if [[ -z "$EXPECTED_DB_NAME" || "$DB_NAME" != "$EXPECTED_DB_NAME" ]]; then
  echo "Refusing apply: EXPECTED_DB_NAME must be set and match current_database()." >&2
  echo "Current database was detected but connection string is not printed." >&2
  exit 2
fi

if [[ "$ACADEMIC_YEAR_COUNT" != "1" ]]; then
  echo "Refusing apply: academic year 2025/2026 must exist exactly once." >&2
  exit 2
fi

if [[ "$CLASS_COUNT" -le 0 || "$STUDENT_COUNT" -le 0 ]]; then
  echo "Refusing apply: target classes/students preview is empty." >&2
  exit 2
fi

BACKUP_NOTE="Create and verify a PostgreSQL backup before running this apply step."
echo "$BACKUP_NOTE"
echo "Applying base seed..."
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -v confirm_uts_genap_seed=1 -f "$SEED_BASE"
echo "Applying sessions/rooms/participants seed..."
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -v confirm_uts_genap_seed=1 -f "$SEED_SESSIONS"
echo "Seed apply completed. Verify event/package/session/participant counts before publishing schedules."
