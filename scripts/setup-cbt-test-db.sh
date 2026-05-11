#!/usr/bin/env bash
set -euo pipefail

# Create/update a dedicated PostgreSQL database for CBT seed testing.
# Production DATABASE_URL remains untouched. This script only targets TEST_DATABASE_URL
# (or a derived *_test database when TEST_DATABASE_URL is not provided).
#
# Usage:
#   scripts/setup-cbt-test-db.sh
#   scripts/setup-cbt-test-db.sh --reset
#   TEST_DATABASE_URL=postgresql://pusaka:pusaka_dev@localhost:5432/pusaka_cbt_test?sslmode=disable scripts/setup-cbt-test-db.sh --reset
#
# Requirements:
#   - PostgreSQL client tools: psql, createdb/dropdb optional but psql is enough
#   - current DB role must be allowed to create/drop TEST database when --reset or DB missing

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CORE_ENV="${CORE_ENV:-${ROOT_DIR}/services/core-api/.env}"
DB_SCRIPTS_DIR="${ROOT_DIR}/services/core-api/db/scripts"
RESET=false
SEED=true
APPLY_MIGRATIONS=true

usage() {
  cat <<'EOF'
Usage: scripts/setup-cbt-test-db.sh [options]

Options:
  --reset          Drop and recreate the test database before migrating/seeding.
  --no-seed        Only create/migrate the test database; skip CBT seed data.
  --no-migrate     Skip migrations; only run seed/create checks.
  -h, --help       Show this help.

Environment:
  DATABASE_URL       Production/main DB URL. Loaded from services/core-api/.env if unset.
  TEST_DATABASE_URL  Dedicated test DB URL. If unset, derived from DATABASE_URL by suffixing DB name with _test.
  ALLOW_NON_TEST_DB=true  Override safety check requiring test DB name to contain "test".

Examples:
  scripts/setup-cbt-test-db.sh --reset
  TEST_DATABASE_URL=postgresql://pusaka:pusaka_dev@localhost:5432/pusaka_cbt_test?sslmode=disable scripts/setup-cbt-test-db.sh --reset

Important:
  This script never rewrites services/core-api/.env or apps/web-admin/.env.
  To run services against test DB, pass DATABASE_URL="$TEST_DATABASE_URL" in that process only.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --reset) RESET=true ;;
    --no-seed) SEED=false ;;
    --no-migrate) APPLY_MIGRATIONS=false ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage; exit 2 ;;
  esac
  shift
done

if [[ -f "${CORE_ENV}" ]]; then
  set -a
  # shellcheck source=/dev/null
  source "${CORE_ENV}"
  set +a
fi

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "DATABASE_URL is not set and ${CORE_ENV} could not provide it." >&2
  exit 1
fi

URL_INFO_JSON="$(python3 - <<'PY'
import json, os, sys
from urllib.parse import urlsplit, urlunsplit

prod = os.environ['DATABASE_URL']
test = os.environ.get('TEST_DATABASE_URL')

def normalize(url):
    parts = urlsplit(url)
    if parts.scheme not in ('postgresql', 'postgres'):
        raise SystemExit(f'Unsupported DB URL scheme: {parts.scheme}')
    db = parts.path.lstrip('/')
    if not db:
        raise SystemExit('DATABASE_URL must include database name')
    return parts, db

prod_parts, prod_db = normalize(prod)
if not test:
    test_db = prod_db if prod_db.endswith('_test') else f'{prod_db}_test'
    test_parts = prod_parts._replace(path='/' + test_db)
    test = urlunsplit(test_parts)
else:
    test_parts, test_db = normalize(test)

# Maintenance URL uses postgres database on the same server as TEST_DATABASE_URL.
maintenance = urlunsplit(test_parts._replace(path='/postgres'))
print(json.dumps({
    'prod': prod,
    'prod_db': prod_db,
    'test': test,
    'test_db': test_db,
    'maintenance': maintenance,
}))
PY
)"

PROD_DB="$(URL_INFO_JSON="${URL_INFO_JSON}" python3 - <<'PY'
import json, os
print(json.loads(os.environ['URL_INFO_JSON'])['prod_db'])
PY
)"
TEST_DB="$(URL_INFO_JSON="${URL_INFO_JSON}" python3 - <<'PY'
import json, os
print(json.loads(os.environ['URL_INFO_JSON'])['test_db'])
PY
)"
TEST_DATABASE_URL_EFFECTIVE="$(URL_INFO_JSON="${URL_INFO_JSON}" python3 - <<'PY'
import json, os
print(json.loads(os.environ['URL_INFO_JSON'])['test'])
PY
)"
MAINTENANCE_URL="$(URL_INFO_JSON="${URL_INFO_JSON}" python3 - <<'PY'
import json, os
print(json.loads(os.environ['URL_INFO_JSON'])['maintenance'])
PY
)"

mask_url() {
  python3 - "$1" <<'PY'
import re, sys
print(re.sub(r':([^:@/]+)@', ':***@', sys.argv[1]))
PY
}

if [[ "${TEST_DATABASE_URL_EFFECTIVE}" == "${DATABASE_URL}" || "${TEST_DB}" == "${PROD_DB}" ]]; then
  echo "Refusing: TEST database resolves to the same DB as production (${TEST_DB})." >&2
  exit 1
fi

if [[ "${ALLOW_NON_TEST_DB:-false}" != "true" && "${TEST_DB,,}" != *test* ]]; then
  echo "Refusing: test DB name must contain 'test'. Got: ${TEST_DB}" >&2
  echo "Set ALLOW_NON_TEST_DB=true only if you are absolutely sure." >&2
  exit 1
fi

if ! command -v psql >/dev/null 2>&1; then
  echo "psql not found." >&2
  exit 1
fi

printf 'Production DB : %s\n' "${PROD_DB}"
printf 'Test DB       : %s\n' "${TEST_DB}"
printf 'Test URL      : %s\n' "$(mask_url "${TEST_DATABASE_URL_EFFECTIVE}")"

EXISTS="$(psql "${MAINTENANCE_URL}" -tAc "SELECT 1 FROM pg_database WHERE datname = '${TEST_DB//\'/\'\'}'" || true)"
CAN_CREATE_DB="$(psql "${MAINTENANCE_URL}" -tAc "SELECT CASE WHEN rolsuper OR rolcreatedb THEN '1' ELSE '0' END FROM pg_roles WHERE rolname = current_user" || true)"
APP_DB_USER="$(psql "${MAINTENANCE_URL}" -tAc "SELECT current_user" || true)"
if { [[ "${RESET}" == "true" ]] || [[ "${EXISTS}" != "1" ]]; } && [[ "${CAN_CREATE_DB}" != "1" ]]; then
  cat >&2 <<EOF
Current database user (${APP_DB_USER}) cannot create databases, so I cannot prepare ${TEST_DB} automatically.
Ask a PostgreSQL admin to run once:

  ALTER ROLE ${APP_DB_USER} CREATEDB;

or create the DB manually as postgres:

  CREATE DATABASE ${TEST_DB} OWNER ${APP_DB_USER};

Then rerun without --reset, or set TEST_DATABASE_URL to an already-created test database.
EOF
  exit 1
fi

if [[ "${RESET}" == "true" ]]; then
  echo "Resetting test database ${TEST_DB}..."
  psql "${MAINTENANCE_URL}" --set=ON_ERROR_STOP=1 --set=db="${TEST_DB}" <<'SQL'
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = :'db'
  AND pid <> pg_backend_pid();
SELECT format('DROP DATABASE IF EXISTS %I', :'db')\gexec
SQL
fi

EXISTS="$(psql "${MAINTENANCE_URL}" -tAc "SELECT 1 FROM pg_database WHERE datname = '${TEST_DB//\'/\'\'}'" || true)"
if [[ "${EXISTS}" != "1" ]]; then
  echo "Creating test database ${TEST_DB}..."
  psql "${MAINTENANCE_URL}" --set=ON_ERROR_STOP=1 --set=db="${TEST_DB}" <<'SQL'
SELECT format('CREATE DATABASE %I', :'db')\gexec
SQL
else
  echo "Test database already exists."
fi

if [[ "${APPLY_MIGRATIONS}" == "true" ]]; then
  echo "Applying migrations to ${TEST_DB}..."
  (cd "${ROOT_DIR}" && DATABASE_URL="${TEST_DATABASE_URL_EFFECTIVE}" node "${DB_SCRIPTS_DIR}/apply_migrations.js")
fi

if [[ "${SEED}" == "true" ]]; then
  echo "Seeding CBT operational/mobile data into ${TEST_DB}..."
  (cd "${ROOT_DIR}" && DATABASE_URL="${TEST_DATABASE_URL_EFFECTIVE}" node "${DB_SCRIPTS_DIR}/seed_mobile_cbt.js")
fi

cat <<EOF

Done.
Use this database for test processes only:
  export TEST_DATABASE_URL='${TEST_DATABASE_URL_EFFECTIVE}'
  DATABASE_URL='${TEST_DATABASE_URL_EFFECTIVE}' ./services/core-api/bin/api

Production/main .env was not modified.
EOF
