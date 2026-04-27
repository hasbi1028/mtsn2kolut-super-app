FRONTEND_DIR := frontend
WORKER_DIR   := worker
LOGS_DIR     := logs

# ── Install ────────────────────────────────────────────────────────────────────

.PHONY: install install-frontend install-worker

install: install-frontend install-worker

install-frontend:
	cd $(FRONTEND_DIR) && npm install

install-worker:
	cd $(WORKER_DIR) && npm install

# ── Dev ────────────────────────────────────────────────────────────────────────

.PHONY: dev dev-frontend dev-worker

dev-frontend:
	cd $(FRONTEND_DIR) && npm run dev

dev-worker:
	cd $(WORKER_DIR) && npm run dev

# ── Type check ────────────────────────────────────────────────────────────────

.PHONY: check check-frontend

check-frontend:
	cd $(FRONTEND_DIR) && npm run check

check: check-frontend

# ── Build ─────────────────────────────────────────────────────────────────────

.PHONY: build build-frontend build-worker

build-frontend:
	cd $(FRONTEND_DIR) && npm run build

build-worker:
	@echo "worker: no build step (tsx runs TypeScript directly)"

build: build-frontend build-worker

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

zip: zip-frontend zip-worker

# ── Start (production, manual) ────────────────────────────────────────────────

.PHONY: start-frontend start-worker logs

start-frontend:
	mkdir -p $(LOGS_DIR)
	cd $(FRONTEND_DIR) && node build/index.js

start-worker:
	mkdir -p $(LOGS_DIR)
	cd $(WORKER_DIR) && npm run start

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

.PHONY: db-studio db-generate db-migrate

db-studio:
	cd $(FRONTEND_DIR) && npm run db:studio

db-generate:
	cd $(FRONTEND_DIR) && npm run db:generate

db-migrate:
	cd $(FRONTEND_DIR) && npm run db:migrate

# ── Clean ─────────────────────────────────────────────────────────────────────

.PHONY: clean clean-build clean-zip

clean-build:
	rm -rf $(FRONTEND_DIR)/build $(WORKER_DIR)/build

clean-zip:
	rm -f dist-frontend.zip dist-worker.zip

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
	@echo "  db-studio          drizzle-kit studio"
	@echo "  db-generate        drizzle-kit generate"
	@echo "  db-migrate         drizzle-kit migrate"
	@echo "  clean              hapus build output + zip"
	@echo ""

.DEFAULT_GOAL := help
