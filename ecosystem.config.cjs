/**
 * Compatibility PM2 entrypoint.
 *
 * Production single source of truth lives in deploy/pm2/*.config.cjs.
 * This root file only aggregates those deploy configs for older local/operator
 * commands that still call `pm2 start ecosystem.config.cjs`.
 *
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
 * Each VPS should run its unit-specific deploy/pm2 config.
 * Use `node scripts/validate-pm2-configs.mjs` before applying a PM2 config change.
 */
const backend = require('./deploy/pm2/backend.config.cjs');
const web = require('./deploy/pm2/web.config.cjs');
const worker = require('./deploy/pm2/worker.config.cjs');
const cbtPortal = require('./deploy/pm2/cbt-portal.config.cjs');

module.exports = {
  apps: [
    ...backend.apps,
    ...web.apps,
    ...cbtPortal.apps,
    ...worker.apps
  ]
};
