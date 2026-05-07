WEB_DIR        := apps/web-admin
WORKER_DIR     := services/pusaka-worker
BACKEND_DIR    := services/core-api
DB_SCRIPTS_DIR := $(BACKEND_DIR)/db/scripts
LOGS_DIR       := logs
TOOLS_BIN      := .tools/bin
SQLC_VERSION   := v1.30.0
SQLC_BIN       := $(TOOLS_BIN)/sqlc

# ── Install ──────────────────────────────────────────────────────────────────

.PHONY: install install-web install-worker install-backend install-db-scripts install-bun-dev

install: install-web install-worker install-backend install-db-scripts

install-web:
	cd $(WEB_DIR) && npm install

install-worker:
	cd $(WORKER_DIR) && npm install

install-backend:
	cd $(BACKEND_DIR) && go mod download

install-db-scripts:
	cd $(DB_SCRIPTS_DIR) && npm install

install-bun-dev:
	bun install --no-save

# ── Dev ──────────────────────────────────────────────────────────────────────

.PHONY: dev-web dev-worker dev-backend dev-web-bun dev-worker-bun

dev-web:
	cd $(WEB_DIR) && npm run dev

dev-worker:
	cd $(WORKER_DIR) && npm run dev

dev-web-bun:
	cd $(WEB_DIR) && bun run dev

dev-worker-bun:
	cd $(WORKER_DIR) && bun run dev

dev-backend:
	cd $(BACKEND_DIR) && ./start-dev.sh

# ── Check / Test ─────────────────────────────────────────────────────────────

.PHONY: check check-web check-web-bun check-worker check-worker-bun check-mobile test-web test-web-bun test-worker test-worker-bun test-mobile test-backend coverage-backend-unit coverage-backend-unit-88 vet-backend audit-web lint-backend lint ci-check
.PHONY: ops-health ops-health-backend ops-health-frontend ops-health-worker ops-backup

check-web:
	cd $(WEB_DIR) && npm run check

check-web-bun:
	cd $(WEB_DIR) && bun run check

test-web:
	cd $(WEB_DIR) && npm run test:unit

test-web-bun:
	cd $(WEB_DIR) && bun run test:unit

check-worker:
	cd $(WORKER_DIR) && ./node_modules/.bin/tsc --noEmit

check-worker-bun:
	cd $(WORKER_DIR) && bun run check

check-mobile:
	cd apps/mobile && flutter analyze

test-worker:
	cd $(WORKER_DIR) && npm run test

test-worker-bun:
	cd $(WORKER_DIR) && bun run test

test-mobile:
	cd apps/mobile && flutter test

test-backend:
	./$(BACKEND_DIR)/run-go-test.sh

coverage-backend-unit:
	./$(BACKEND_DIR)/run-go-unit-coverage.sh

coverage-backend-unit-88:
	COVERAGE_THRESHOLD=88 ./$(BACKEND_DIR)/run-go-unit-coverage.sh

vet-backend:
	./$(BACKEND_DIR)/run-go-vet.sh

audit-web:
	cd $(WEB_DIR) && npm audit --audit-level=high

lint-backend:
	cd $(BACKEND_DIR) && (golangci-lint run ./... || true)

lint: lint-backend

ci-check: check-web test-web check-worker test-worker test-backend test-mobile

check: check-web test-web check-worker test-worker test-backend test-mobile vet-backend audit-web

# ── Build ────────────────────────────────────────────────────────────────────

.PHONY: build build-web build-worker build-backend mobile-release-apk

build-web:
	cd $(WEB_DIR) && npm run build

build-worker:
	cd $(WORKER_DIR) && npm run build

build-backend:
	cd $(BACKEND_DIR) && go build -o bin/api ./cmd/api

build: build-web build-worker build-backend

mobile-release-apk:
	@if [ -z "$${API_BASE_URL}" ]; then \
		printf 'API_BASE_URL is required, for example: make mobile-release-apk API_BASE_URL=https://api.sekolah.example\n' >&2; \
		exit 1; \
	fi
	@case "$${API_BASE_URL}" in https://*) ;; *) printf 'API_BASE_URL must use https:// for release APK builds\n' >&2; exit 1 ;; esac
	cd apps/mobile && flutter build apk --release --dart-define=API_BASE_URL="$${API_BASE_URL}"

# ── Start (production, manual) ───────────────────────────────────────────────

.PHONY: start-web start-worker start-backend

start-web:
	mkdir -p $(LOGS_DIR)
	cd $(WEB_DIR) && ./start.sh

start-worker:
	mkdir -p $(LOGS_DIR)
	cd $(WORKER_DIR) && npm run start

start-backend:
	$(BACKEND_DIR)/bin/api

# ── PM2 (via ecosystem.config.cjs) ──────────────────────────────────────────

.PHONY: pm2-start pm2-stop pm2-restart pm2-logs pm2-status
.PHONY: pm2-start-backend pm2-start-web pm2-start-worker
.PHONY: pm2-restart-backend pm2-restart-web pm2-restart-worker
.PHONY: pm2-stop-backend pm2-stop-web pm2-stop-worker

pm2-start:
	pm2 start ecosystem.config.cjs

pm2-stop:
	pm2 stop ecosystem.config.cjs

pm2-restart:
	pm2 restart ecosystem.config.cjs

pm2-logs:
	pm2 logs

pm2-status:
	pm2 status

pm2-start-backend:
	pm2 start deploy/pm2/backend.config.cjs

pm2-start-web:
	pm2 start deploy/pm2/web.config.cjs

pm2-start-worker:
	pm2 start deploy/pm2/worker.config.cjs

pm2-restart-backend:
	pm2 restart deploy/pm2/backend.config.cjs

pm2-restart-web:
	pm2 restart deploy/pm2/web.config.cjs

pm2-restart-worker:
	pm2 restart deploy/pm2/worker.config.cjs

pm2-stop-backend:
	pm2 stop deploy/pm2/backend.config.cjs

pm2-stop-web:
	pm2 stop deploy/pm2/web.config.cjs

pm2-stop-worker:
	pm2 stop deploy/pm2/worker.config.cjs

# ── DB helpers ───────────────────────────────────────────────────────────────

.PHONY: db-sqlc db-sqlc-install db-migrate db-migrate-local db-schema

db-sqlc: db-sqlc-install
	cd $(BACKEND_DIR)/db && ../../../$(SQLC_BIN) generate -f sqlc.yaml

db-sqlc-install:
	@if [ ! -x "$(SQLC_BIN)" ]; then \
		printf 'Installing sqlc %s into %s\n' "$(SQLC_VERSION)" "$(TOOLS_BIN)"; \
		mkdir -p "$(TOOLS_BIN)"; \
		GOBIN="$(CURDIR)/$(TOOLS_BIN)" go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION); \
	fi

db-migrate:
	cd $(DB_SCRIPTS_DIR) && \
		if [ -n "$${DATABASE_URL}" ]; then \
			npm run migrate:pg; \
		elif [ -f ../../.env ]; then \
			set -a; . ../../.env; set +a; npm run migrate:pg; \
		else \
			npm run migrate:pg; \
		fi

db-migrate-local:
	cd $(DB_SCRIPTS_DIR) && ALLOW_LOCAL_DATABASE_URL=true npm run migrate:pg

db-schema:
	psql "$${DATABASE_URL}" -f $(BACKEND_DIR)/db/migrations/001_initial_schema.sql

# ── Ops Hardening ────────────────────────────────────────────────────────────

ops-health:
	./deploy/scripts/health-check.sh all

ops-health-backend:
	./deploy/scripts/health-check.sh backend

ops-health-frontend:
	./deploy/scripts/health-check.sh frontend

ops-health-worker:
	./deploy/scripts/health-check.sh worker

ops-backup:
	./deploy/backup-postgresql.sh

# ── Clean ────────────────────────────────────────────────────────────────────

.PHONY: clean clean-build

clean-build:
	rm -rf $(WEB_DIR)/build $(WORKER_DIR)/dist $(BACKEND_DIR)/bin

clean: clean-build

# ── Help ─────────────────────────────────────────────────────────────────────

.PHONY: help

help:
	@echo ""
	@echo "Monorepo layout:"
	@echo "  apps/web-admin         SvelteKit admin app"
	@echo "  services/core-api      Go Chi API + sqlc + PostgreSQL"
	@echo "  services/pusaka-worker Playwright worker"
	@echo ""
	@echo "Install:"
	@echo "  install                install semua dependency"
	@echo "  install-web            npm install apps/web-admin"
	@echo "  install-worker         npm install services/pusaka-worker"
	@echo "  install-backend        go mod download services/core-api"
	@echo "  install-db-scripts     npm install services/core-api/db/scripts"
	@echo "  install-bun-dev        optional local Bun install without replacing npm lock"
	@echo ""
	@echo "Dev:"
	@echo "  dev-backend            go run ./cmd/api"
	@echo "  dev-web                npm run dev"
	@echo "  dev-worker             npm run dev"
	@echo "  dev-web-bun            optional Bun dev runner for web-admin only"
	@echo "  dev-worker-bun         optional Bun script runner; worker production runtime remains Node/dist"
	@echo ""
	@echo "Verify:"
	@echo "  check                  web check + web unit + worker typecheck + worker tests + backend tests + mobile tests + vet + npm audit"
	@echo "  ci-check               web check + web unit + worker typecheck + worker tests + backend tests + mobile tests"
	@echo "  coverage-backend-unit  Go unit coverage excluding cmd/api and generated sqlc"
	@echo "  coverage-backend-unit-88  same coverage scope with 88% threshold"
	@echo "  lint                   golangci-lint run (falls back if not installed)"
	@echo "  check-mobile           flutter analyze apps/mobile"
	@echo "  check-web-bun          optional Bun runner for Svelte check"
	@echo "  test-web-bun           optional Bun runner for Vitest"
	@echo "  test-mobile            flutter test apps/mobile"
	@echo "  db-sqlc                install repo-local sqlc if needed, then regenerate sqlc code"
	@echo "  db-migrate             apply PostgreSQL migrations using DATABASE_URL"
	@echo "  db-migrate-local       apply migrations to explicit local dev database"
	@echo "  ops-health             health check backend + frontend + worker"
	@echo "  ops-backup             run PostgreSQL backup script"
	@echo ""
	@echo "Deploy:"
	@echo "  pm2-start-backend      start backend-only PM2 config"
	@echo "  pm2-start-web          start web-only PM2 config"
	@echo "  pm2-start-worker       start worker-only PM2 config"
	@echo "  deploy/DEPLOY.md       step-by-step deploy contract for 3 VPS"
	@echo ""
	@echo "Build:"
	@echo "  build                  build web + worker + backend"
	@echo "  mobile-release-apk     build Flutter APK release; requires API_BASE_URL=https://..."
	@echo "  start-web              start SvelteKit production server"
	@echo "  start-worker           start Playwright worker"
	@echo "  start-backend          start compiled Go API"
	@echo ""

.DEFAULT_GOAL := help
