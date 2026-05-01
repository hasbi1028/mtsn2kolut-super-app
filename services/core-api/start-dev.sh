#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  . ./.env
  set +a
fi

DB_HOST="${POSTGRES_HOST:-localhost}"
DB_PORT="${POSTGRES_PORT:-5432}"
CONTAINER_ENGINE=""
COMPOSE_CMD=()

db_ready() {
  (echo >"/dev/tcp/${DB_HOST}/${DB_PORT}") >/dev/null 2>&1
}

is_local_db_host() {
  [[ "$DB_HOST" == "localhost" || "$DB_HOST" == "127.0.0.1" || "$DB_HOST" == "::1" ]]
}

detect_container_engine() {
  if [[ -n "${CONTAINER_ENGINE:-}" ]]; then
    return 0
  fi

  if command -v podman >/dev/null 2>&1; then
    CONTAINER_ENGINE="podman"
    if podman compose version >/dev/null 2>&1; then
      COMPOSE_CMD=(podman compose)
      return 0
    fi
    if command -v podman-compose >/dev/null 2>&1; then
      COMPOSE_CMD=(podman-compose)
      return 0
    fi
  fi

  if command -v docker >/dev/null 2>&1; then
    if docker info >/dev/null 2>&1; then
      CONTAINER_ENGINE="docker"
      COMPOSE_CMD=(docker compose)
      return 0
    fi
  fi

  return 1
}

maybe_start_container_db() {
  detect_container_engine || return 1

  if "$CONTAINER_ENGINE" ps -a --filter "name=^pusaka-postgres$" --format '{{.Names}}' | grep -qx 'pusaka-postgres'; then
    echo "PostgreSQL belum aktif di ${DB_HOST}:${DB_PORT}. Menjalankan ${CONTAINER_ENGINE} start pusaka-postgres..."
    "$CONTAINER_ENGINE" start pusaka-postgres >/dev/null
    return 0
  fi

  echo "PostgreSQL belum aktif di ${DB_HOST}:${DB_PORT}. Menjalankan ${COMPOSE_CMD[*]} up -d..."
  "${COMPOSE_CMD[@]}" up -d >/dev/null
}

if ! db_ready && is_local_db_host; then
  maybe_start_container_db || true

  for _ in $(seq 1 15); do
    if db_ready; then
      break
    fi
    sleep 1
  done
fi

if ! db_ready; then
  cat <<EOF
PostgreSQL tidak aktif di ${DB_HOST}:${DB_PORT}.

Pilihan:
1. Jika ini host lokal, jalankan database dev container:
   cd services/core-api && podman compose up -d
2. Jika ini host remote/VPS, pastikan firewall dan kredensial PostgreSQL memang mengizinkan koneksi dari mesin ini.

Setelah database aktif, jalankan lagi:
  make dev-backend
EOF
  exit 1
fi

exec go run ./cmd/api
