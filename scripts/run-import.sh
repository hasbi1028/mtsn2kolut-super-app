#!/bin/bash
# Run this from the project root
set -e
cd "$(dirname "$0")/../services/core-api"
source .env
PGPASSWORD="${POSTGRES_PASSWORD:-$POSTGRESS_PASSWORD}" psql -h 127.0.0.1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -p "${POSTGRES_PORT:-5432}" -f ../../scripts/import-all.sql
