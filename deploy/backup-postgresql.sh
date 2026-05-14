#!/usr/bin/env bash
# Backup PostgreSQL database for mtsn2kolut-super-app
# Database: pusaka (system PostgreSQL on 127.0.0.1:5432)
# Output: /home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql/

set -Eeuo pipefail
umask 077

APP_DIR="/home/servermtsn2kolut/mtsn2kolut-super-app"
ENV_FILE="$APP_DIR/services/core-api/.env"
BACKUP_DIR="/home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql"
LOG_DIR="/home/servermtsn2kolut/backups/mtsn2kolut-super-app/logs"
RETENTION_DAYS="30"
LOCK_FILE="/tmp/mtsn2kolut-postgresql-backup.lock"
LOG_FILE="$LOG_DIR/postgresql-backup.log"

mkdir -p "$BACKUP_DIR" "$LOG_DIR"
chmod 700 "$BACKUP_DIR" "$LOG_DIR"
touch "$LOG_FILE"
chmod 600 "$LOG_FILE"

log() {
  printf '[%s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S %Z')" "$*"
}

# Prevent overlapping backups
exec 9>"$LOCK_FILE"
chmod 600 "$LOCK_FILE" || true
if ! flock -n 9; then
  log "Backup already running, exiting."
  exit 0
fi

if [[ ! -f "$ENV_FILE" ]]; then
  log "ERROR: env file not found: $ENV_FILE"
  exit 1
fi

# Load database config from core-api .env without exposing secrets in logs
set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

DB_DSN="${DATABASE_URL:-}"
DB_NAME="${POSTGRES_DB:-pusaka}"
DB_USER="${POSTGRES_USER:-pusaka}"
DB_HOST="127.0.0.1"
DB_PORT="${POSTGRES_PORT:-5432}"
DB_PASSWORD="${POSTGRES_PASSWORD:-}"

if [[ -n "$DB_DSN" ]]; then
  mapfile -t DB_URL_PARTS < <(python3 - <<'PY'
import os
from urllib.parse import unquote, urlparse
url = urlparse(os.environ['DATABASE_URL'])
print(unquote(url.path.lstrip('/') or ''))
print(unquote(url.username or ''))
print(unquote(url.hostname or ''))
print(str(url.port or ''))
print(unquote(url.password or ''))
PY
  )
  DB_NAME="${DB_URL_PARTS[0]:-$DB_NAME}"
  DB_USER="${DB_URL_PARTS[1]:-$DB_USER}"
  DB_HOST="${DB_URL_PARTS[2]:-$DB_HOST}"
  DB_PORT="${DB_URL_PARTS[3]:-$DB_PORT}"
  DB_PASSWORD="${DB_URL_PARTS[4]:-$DB_PASSWORD}"
elif [[ -z "$DB_PASSWORD" ]]; then
  log "ERROR: DATABASE_URL or POSTGRES_PASSWORD must be set in $ENV_FILE"
  exit 1
fi

TIMESTAMP="$(date '+%Y%m%d_%H%M%S')"
BACKUP_FILE="$BACKUP_DIR/${DB_NAME}_${TIMESTAMP}.dump"
SHA_FILE="$BACKUP_FILE.sha256"
LATEST_LINK="$BACKUP_DIR/latest.dump"

{
  log "Starting PostgreSQL backup: database=$DB_NAME host=$DB_HOST port=$DB_PORT user=$DB_USER"
  export PGPASSWORD="$DB_PASSWORD"
  trap 'unset PGPASSWORD' EXIT
  pg_dump \
    --host="$DB_HOST" \
    --port="$DB_PORT" \
    --username="$DB_USER" \
    --dbname="$DB_NAME" \
    --format=custom \
    --verbose \
    --file="$BACKUP_FILE"
  unset PGPASSWORD
  trap - EXIT
  chmod 600 "$BACKUP_FILE"

  sha256sum "$BACKUP_FILE" > "$SHA_FILE"
  chmod 600 "$SHA_FILE"
  ln -sfn "$BACKUP_FILE" "$LATEST_LINK"
  (cd "$BACKUP_DIR" && sha256sum latest.dump > latest.dump.sha256)
  chmod 600 "$BACKUP_DIR/latest.dump.sha256"
  chmod 700 "$BACKUP_DIR" "$LOG_DIR"

  BACKUP_SIZE="$(du -h "$BACKUP_FILE" | awk '{print $1}')"
  log "Backup completed: $BACKUP_FILE ($BACKUP_SIZE)"
  log "Checksum: $SHA_FILE"

  # Remove old backups and checksums
  find "$BACKUP_DIR" -type f -name "${DB_NAME}_*.dump" -mtime +"$RETENTION_DAYS" -delete
  find "$BACKUP_DIR" -type f -name "${DB_NAME}_*.dump.sha256" -mtime +"$RETENTION_DAYS" -delete

  log "Retention cleanup done: keep last $RETENTION_DAYS days"
} >> "$LOG_FILE" 2>&1
