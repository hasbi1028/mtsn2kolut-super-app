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

.PHONY: zip zip-frontend zip-worker zip-backend

zip-frontend: build-frontend
	rm -f dist-frontend.zip
	cd $(FRONTEND_DIR) && zip -r ../dist-frontend.zip . \
		--exclude "node_modules/*" \
		--exclude ".env" \
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

.PHONY: db-sqlc db-schema

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
	@echo "Arsitektur: 3 komponen terpisah"
	@echo "  VPS-Backend  : Go Chi API  (port 8080)"
	@echo "  VPS-Frontend : SvelteKit   (port 8021)"
	@echo "  VPS-Worker   : Playwright  (no HTTP port, pull jobs)"
	@echo ""
	@echo "── Install ──────────────────────────────────────────────────────────"
	@echo "  install            npm install frontend + worker, go mod download"
	@echo "  install-frontend   npm install frontend saja"
	@echo "  install-worker     npm install worker saja"
	@echo "  install-backend    go mod download"
	@echo ""
	@echo "── Dev (jalankan 3 terminal terpisah) ───────────────────────────────"
	@echo "  dev-backend        go run ./cmd/api/        (butuh .env di backend/)"
	@echo "  dev-frontend       vite dev hot-reload       (butuh .env di frontend/)"
	@echo "  dev-worker         tsx watch src/index.ts   (butuh .env di worker/)"
	@echo ""
	@echo "── Build ────────────────────────────────────────────────────────────"
	@echo "  build              build semua (frontend + backend)"
	@echo "  build-frontend     npm run build (adapter-node)"
	@echo "  build-backend      go build -o bin/api"
	@echo ""
	@echo "── Deploy (zip untuk upload ke VPS) ─────────────────────────────────"
	@echo "  zip                zip semua: frontend + worker + backend"
	@echo "  zip-frontend       build + zip dist-frontend.zip"
	@echo "  zip-worker         zip dist-worker.zip (no build)"
	@echo "  zip-backend        build + zip dist-backend.zip"
	@echo ""
	@echo "── PM2 ──────────────────────────────────────────────────────────────"
	@echo "  pm2-start          pm2 start ecosystem.config.cjs"
	@echo "  pm2-stop/restart   pm2 stop/restart ecosystem.config.cjs"
	@echo "  pm2-logs           pm2 logs"
	@echo "  pm2-status         pm2 status"
	@echo ""
	@echo "── Database ─────────────────────────────────────────────────────────"
	@echo "  db-sqlc            sqlc generate (regenerate Go typed queries)"
	@echo "  db-schema          apply 001_initial_schema.sql ke PostgreSQL"
	@echo ""
	@echo "── Clean ────────────────────────────────────────────────────────────"
	@echo "  clean              hapus build output + zip"
	@echo ""

.DEFAULT_GOAL := help
