const path = require('path');

const rootDir = __dirname;

/**
 * Deployment topology:
 *   VPS-Backend  — mtsn2kolut-core-api     (Go Chi API, port 8080) + PostgreSQL
 *   VPS-Frontend — mtsn2kolut-web-admin    (SvelteKit,  port 8021)
 *   VPS-Worker   — mtsn2kolut-pusaka-worker (Playwright, no HTTP port, pull-based)
 *
 * Required env vars:
 *   Backend  : DATABASE_URL, JWT_SECRET, ADMIN_PASSWORD, WORKER_API_KEY, INTERNAL_API_KEY, PORT
 *   Frontend : API_BASE_URL, INTERNAL_API_KEY, SESSION_SECRET, ORIGIN
 *   Worker   : BACKEND_URL, WORKER_API_KEY, WORKER_ID, WORKER_CONCURRENCY
 *
 * Each VPS has its own ecosystem.config.cjs — copy the relevant app block only.
 */
module.exports = {
  apps: [
    // ── Backend (VPS-Backend) ────────────────────────────────────────────────
    // Run: mtsn2kolut-core-api/bin/api  (go build -o bin/api ./cmd/api/)
    {
      name: 'mtsn2kolut-core-api',
      cwd: './services/core-api',
      script: 'bin/api',
      interpreter: 'none',
      env_file: path.join(rootDir, 'services/core-api/.env'),
      env: {
        PORT:             '8080',
        NODE_ENV:         'production',
      },
      instances: 1,
      exec_mode: 'fork',
      autorestart: true,
      watch: false,
      max_memory_restart: '256M',
      error_file: '../logs/backend-error.log',
      out_file:   '../logs/backend-out.log',
      log_date_format: 'YYYY-MM-DD HH:mm:ss Z'
    },

    // ── Frontend (VPS-Frontend) ──────────────────────────────────────────────
    {
      name: 'mtsn2kolut-web-admin',
      cwd: './apps/web-admin',
      script: 'start.sh',
      interpreter: 'bash',
      env: {
        HOST:             '0.0.0.0',
        PORT:             '8021',
        NODE_ENV:         'production',
      },
      instances: 1,
      exec_mode: 'fork',
      autorestart: true,
      watch: false,
      max_memory_restart: '300M',
      error_file: '../logs/frontend-error.log',
      out_file:   '../logs/frontend-out.log',
      log_date_format: 'YYYY-MM-DD HH:mm:ss Z'
    },

    // ── Worker (VPS-Worker / multiple VPS) ───────────────────────────────────
    // Copy blok ini ke setiap VPS worker, sesuaikan WORKER_ID.
    // Worker langsung memanggil Go API — tidak butuh akses ke SvelteKit.
    {
      name: 'mtsn2kolut-pusaka-worker',
      cwd: './services/pusaka-worker',
      script: 'node_modules/.bin/tsx',
      args: 'src/index.ts',
      interpreter: 'node',
      env_file: path.join(rootDir, 'services/pusaka-worker/.env'),
      // Nilai sensitif (BACKEND_URL, WORKER_API_KEY) diambil dari worker/.env
      // Jangan override di sini agar tidak menimpa nilai dari env_file
      env: {
        WORKER_ID:          'worker-vps1',               // unik per VPS
        WORKER_CONCURRENCY: '5',
        HEADLESS:           'true',
        POLL_MS:            '8000',
        SCRAPE_RETRIES:     '3',
        SCRAPE_RETRY_MS:    '5000',
        ACTION_TIMEOUT:     '20000',
        WORKER_LOG_PATH:    '../logs/worker.log',
        SCREENSHOT_DIR:     '../logs/screenshots'
      },
      instances: 1,
      exec_mode: 'fork',
      autorestart: true,
      watch: false,
      max_memory_restart: '600M',
      error_file: '../logs/worker-error.log',
      out_file:   '../logs/worker-out.log',
      log_date_format: 'YYYY-MM-DD HH:mm:ss Z'
    }
  ]
};
