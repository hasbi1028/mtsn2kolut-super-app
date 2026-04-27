WEB_DIR        := apps/web-admin
WORKER_DIR     := services/pusaka-worker
BACKEND_DIR    := services/core-api
DB_SCRIPTS_DIR := $(BACKEND_DIR)/db/scripts
LOGS_DIR       := logs

# ── Install ──────────────────────────────────────────────────────────────────

.PHONY: install install-web install-worker install-backend install-db-scripts

install: install-web install-worker install-backend install-db-scripts

install-web:
	cd $(WEB_DIR) && npm install

install-worker:
	cd $(WORKER_DIR) && npm install

install-backend:
	cd $(BACKEND_DIR) && go mod download

install-db-scripts:
	cd $(DB_SCRIPTS_DIR) && npm install

# ── Dev ──────────────────────────────────────────────────────────────────────

.PHONY: dev-web dev-worker dev-backend

dev-web:
	cd $(WEB_DIR) && npm run dev

dev-worker:
	cd $(WORKER_DIR) && npm run dev

dev-backend:
	cd $(BACKEND_DIR) && go run ./cmd/api

# ── Check / Test ─────────────────────────────────────────────────────────────

.PHONY: check check-web check-worker test-backend

check-web:
	cd $(WEB_DIR) && npm run check

check-worker:
	cd $(WORKER_DIR) && ./node_modules/.bin/tsc --noEmit

test-backend:
	cd $(BACKEND_DIR) && go test ./...

check: check-web check-worker test-backend

# ── Build ────────────────────────────────────────────────────────────────────

.PHONY: build build-web build-worker build-backend

build-web:
	cd $(WEB_DIR) && npm run build

build-worker:
	@echo "worker: no build step (tsx runs TypeScript directly)"

build-backend:
	cd $(BACKEND_DIR) && go build -o bin/api ./cmd/api

build: build-web build-worker build-backend

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

.PHONY: db-sqlc db-migrate db-schema

db-sqlc:
	cd $(BACKEND_DIR)/db && sqlc generate

db-migrate:
	cd $(DB_SCRIPTS_DIR) && npm run migrate:pg

db-schema:
	psql "$${DATABASE_URL}" -f $(BACKEND_DIR)/db/migrations/001_initial_schema.sql

# ── Clean ────────────────────────────────────────────────────────────────────

.PHONY: clean clean-build

clean-build:
	rm -rf $(WEB_DIR)/build $(BACKEND_DIR)/bin

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
	@echo ""
	@echo "Dev:"
	@echo "  dev-backend            go run ./cmd/api"
	@echo "  dev-web                npm run dev"
	@echo "  dev-worker             npm run dev"
	@echo ""
	@echo "Verify:"
	@echo "  check                  web check + worker typecheck + backend tests"
	@echo "  db-sqlc                regenerate sqlc code"
	@echo "  db-migrate             apply PostgreSQL migrations"
	@echo ""
	@echo "Deploy:"
	@echo "  pm2-start-backend      start backend-only PM2 config"
	@echo "  pm2-start-web          start web-only PM2 config"
	@echo "  pm2-start-worker       start worker-only PM2 config"
	@echo "  deploy/DEPLOY.md       step-by-step deploy contract for 3 VPS"
	@echo ""
	@echo "Build:"
	@echo "  build                  build web + backend"
	@echo "  start-web              start SvelteKit production server"
	@echo "  start-worker           start Playwright worker"
	@echo "  start-backend          start compiled Go API"
	@echo ""

.DEFAULT_GOAL := help
