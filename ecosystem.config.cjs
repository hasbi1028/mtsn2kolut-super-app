/**
 * Deployment topology:
 *   VPS1 — pusaka-frontend (SvelteKit + SQLite)
 *   VPS2..N — pusaka-worker (Playwright, no DB, connects via HTTP API)
 *
 * Required env vars per VPS:
 *   Frontend : DB_PATH, WORKER_TOKEN
 *   Worker   : FRONTEND_URL, WORKER_TOKEN, WORKER_ID, WORKER_CONCURRENCY
 */
module.exports = {
  apps: [
    // ── Frontend (VPS1) ──────────────────────────────────────────────────────
    {
      name: 'pusaka-frontend',
      cwd: './frontend',
      script: 'build/index.js',
      interpreter: 'node',
      env_file: './frontend/.env',
      env: {
        HOST: '0.0.0.0',
        PORT: '8021',
        // Sesuaikan ORIGIN dengan URL akses app (IP/domain:port) — wajib agar form login tidak ditolak CSRF
        ORIGIN: 'http://localhost:8021',
        DB_PATH: 'data/pusaka.sqlite',
        NODE_ENV: 'production',
        // WORKER_TOKEN: 'ganti-dengan-secret-yang-kuat'
        // SESSION_SECRET: 'ganti-dengan-random-string-panjang'
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

    // ── Worker (VPS2 / VPS3 / dst) ───────────────────────────────────────────
    // Copy blok ini ke masing-masing VPS worker, sesuaikan WORKER_ID.
    // Tidak perlu DB_PATH — worker tidak akses SQLite langsung.
    {
      name: 'pusaka-worker',
      cwd: './worker',
      script: 'node_modules/.bin/tsx',
      args: 'src/index.ts',
      interpreter: 'node',
      env_file: './worker/.env',
      env: {
        FRONTEND_URL:       'http://localhost:8021',   // ganti IP/domain VPS1
        WORKER_TOKEN:       '',                         // harus sama dengan frontend
        WORKER_ID:          'worker-vps2',              // unik per VPS
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
