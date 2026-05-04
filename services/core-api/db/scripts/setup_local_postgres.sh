#!/usr/bin/env bash
set -euo pipefail

# Setup PostgreSQL OS lokal untuk development MTsN 2 Kolaka Utara.
# Default ini selaras dengan services/core-api/.env lokal:
# DATABASE_URL=postgresql://pusaka:pusaka_dev@localhost:5432/pusaka?sslmode=disable

APP_DB="${APP_DB:-pusaka}"
APP_USER="${APP_USER:-pusaka}"
APP_PASSWORD="${APP_PASSWORD:-pusaka_dev}"
APP_HOST="${APP_HOST:-localhost}"
APP_PORT="${APP_PORT:-5432}"

if ! command -v psql >/dev/null 2>&1; then
  echo "psql tidak ditemukan. Install PostgreSQL client/server terlebih dahulu." >&2
  exit 1
fi

if ! command -v sudo >/dev/null 2>&1; then
  echo "sudo tidak ditemukan. Jalankan setup role/database PostgreSQL secara manual." >&2
  exit 1
fi

echo "Menyiapkan role '${APP_USER}' dan database '${APP_DB}' di PostgreSQL lokal..."

sudo -u postgres psql \
  --set=ON_ERROR_STOP=1 \
  --set=app_user="${APP_USER}" \
  --set=app_password="${APP_PASSWORD}" \
  --set=app_db="${APP_DB}" <<'SQL'
SELECT format('CREATE ROLE %I LOGIN PASSWORD %L', :'app_user', :'app_password')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'app_user')\gexec

SELECT format('ALTER ROLE %I WITH LOGIN PASSWORD %L', :'app_user', :'app_password')\gexec

SELECT format('CREATE DATABASE %I OWNER %I', :'app_db', :'app_user')
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = :'app_db')\gexec

SELECT format('ALTER DATABASE %I OWNER TO %I', :'app_db', :'app_user')\gexec
SELECT format('GRANT ALL PRIVILEGES ON DATABASE %I TO %I', :'app_db', :'app_user')\gexec
SQL

sudo -u postgres psql \
  --set=ON_ERROR_STOP=1 \
  --set=app_user="${APP_USER}" \
  --set=app_db="${APP_DB}" <<'SQL'
\connect :app_db
SELECT format('GRANT USAGE, CREATE ON SCHEMA public TO %I', :'app_user')\gexec
SELECT format('ALTER SCHEMA public OWNER TO %I', :'app_user')\gexec
SQL

DATABASE_URL="postgresql://${APP_USER}:${APP_PASSWORD}@${APP_HOST}:${APP_PORT}/${APP_DB}?sslmode=disable"

echo "Menguji koneksi aplikasi..."
psql "${DATABASE_URL}" -c "SELECT current_database(), current_user;"

echo "Setup PostgreSQL lokal selesai."
echo "DATABASE_URL=${DATABASE_URL}"
echo "Lanjutkan dengan: make db-migrate"
