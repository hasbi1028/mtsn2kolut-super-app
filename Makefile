FRONTEND_DIR := frontend
WORKER_DIR   := worker
BACKEND_DIR  := backend
LOGS_DIR     := logs

# ── Install ────────────────────────────────────────────────────────────────────

.PHONY: install install-frontend install-worker install-backend

install: install-frontend install-worker install-backend

install-frontend:
	cd $(FRONTEND_DIR) && npm install

install-worker:
	cd $(WORKER_DIR) && npm install

install-backend:
	cd $(BACKEND_DIR) && go mod download

# ── Dev ────────────────────────────────────────────────────────────────────────

.PHONY: dev dev-frontend dev-worker dev-backend

dev-frontend:
	cd $(FRONTEND_DIR) && npm run dev

dev-worker:
	cd $(WORKER_DIR) && npm run dev

dev-backend:
	cd $(BACKEND_DIR) && go run ./cmd/api/

# ── Type check ────────────────────────────────────────────────────────────────

.PHONY: check check-frontend

check-frontend:
	cd $(FRONTEND_DIR) && npm run check

check: check-frontend

# ── Build ─────────────────────────────────────────────────────────────────────

.PHONY: build build-frontend build-worker build-backend

build-frontend:
	cd $(FRONTEND_DIR) && npm run build

build-worker:
	@echo "worker: no build step (tsx runs TypeScript directly)"

build-backend:
	cd $(BACKEND_DIR) && go build -o bin/api ./cmd/api/

build: build-frontend build-worker build-backend

# ── Zip (untuk upload ke VPS) ─────────────────────────────────────────────────

.PHONY: zip zip-frontend zip-worker

zip-frontend: build-frontend
	rm -f dist-frontend.zip
	cd $(FRONTEND_DIR) && zip -r ../dist-frontend.zip . \
		--exclude "node_modules/*" \
		--exclude ".env" \
		--exclude "data/*" \
		--exclude "*.zip"
	@echo "dist-frontend.zip siap"

zip-worker:
	rm -f dist-worker.zip
	cd $(WORKER_DIR) && zip -r ../dist-worker.zip . \
		--exclude "node_modules/*" \
		--exclude ".env" \
		--exclude "*.zip"
	@echo "dist-worker.zip siap"

zip-backend: build-backend
	rm -f dist-backend.zip
	cd $(BACKEND_DIR) && zip -r ../dist-backend.zip bin/ .env.example \
		--exclude "*.zip" \
		--exclude "postgres_data/*"
	@echo "dist-backend.zip siap"

zip: zip-frontend zip-worker zip-backend

# ── Start (production, manual) ────────────────────────────────────────────────

.PHONY: start-frontend start-worker start-backend logs

start-frontend:
	mkdir -p $(LOGS_DIR)
	cd $(FRONTEND_DIR) && node build/index.js

start-worker:
	mkdir -p $(LOGS_DIR)
	cd $(WORKER_DIR) && npm run start

start-backend:
	$(BACKEND_DIR)/bin/api

# ── PM2 (via ecosystem.config.cjs) ───────────────────────────────────────────

.PHONY: pm2-start pm2-stop pm2-restart pm2-logs pm2-status

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

# ── DB helpers ────────────────────────────────────────────────────────────────

.PHONY: db-studio db-generate db-migrate db-sqlc db-schema

db-studio:
	cd $(FRONTEND_DIR) && npm run db:studio

db-generate:
	cd $(FRONTEND_DIR) && npm run db:generate

db-migrate:
	cd $(FRONTEND_DIR) && npm run db:migrate

db-sqlc:
	cd $(BACKEND_DIR)/db && sqlc generate

db-schema:
	psql "$${DATABASE_URL}" -f $(BACKEND_DIR)/db/migrations/001_initial_schema.sql

# ── Clean ─────────────────────────────────────────────────────────────────────

.PHONY: clean clean-build clean-zip

clean-build:
	rm -rf $(FRONTEND_DIR)/build $(WORKER_DIR)/build $(BACKEND_DIR)/bin

clean-zip:
	rm -f dist-frontend.zip dist-worker.zip dist-backend.zip

clean: clean-build clean-zip

# ── Help ──────────────────────────────────────────────────────────────────────

.PHONY: help

help:
	@echo ""
	@echo "  install            npm install di frontend + worker"
	@echo "  dev-frontend       vite dev (hot-reload)"
	@echo "  dev-worker         tsx watch src/index.ts"
	@echo "  check              svelte-check + TypeScript"
	@echo "  build              build frontend (adapter-node)"
	@echo "  zip                build + buat dist-frontend.zip & dist-worker.zip"
	@echo "  zip-frontend       zip frontend saja"
	@echo "  zip-worker         zip worker saja"
	@echo "  start-frontend     jalankan frontend (node build/index.js)"
	@echo "  start-worker       jalankan worker (tsx src/index.ts)"
	@echo "  pm2-start/stop/restart/logs/status"
	@echo "  dev-backend        go run ./cmd/api/"
	@echo "  build-backend      go build -o bin/api"
	@echo "  start-backend      jalankan backend binary"
	@echo "  zip-backend        build + buat dist-backend.zip"
	@echo "  db-sqlc            sqlc generate (Go typed queries)"
	@echo "  db-schema          apply 001_initial_schema.sql ke PostgreSQL"
	@echo "  db-studio          drizzle-kit studio (SQLite, legacy)"
	@echo "  db-generate        drizzle-kit generate (SQLite, legacy)"
	@echo "  db-migrate         drizzle-kit migrate (SQLite, legacy)"
	@echo "  clean              hapus build output + zip"
	@echo ""

.DEFAULT_GOAL := help
