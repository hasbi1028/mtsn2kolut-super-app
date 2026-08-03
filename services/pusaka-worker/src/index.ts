import 'dotenv/config';

import {
  BACKEND_URL,
  BROWSER_CHECK_MS,
  CONFIG_SYNC_MS,
  HTTP_PORT,
  POLL_MS,
  WORKER_HEARTBEAT_MS,
  WORKER_ID,
  createRuntimeConfig,
} from './config.js';
import { ensureBrowserInstalled, startBrowserWatchdog } from './browser-guard.js';
import { startHealthServer } from './health-server.js';
import { ensureLogDirectories, log } from './logger.js';
import { pruneOldScreenshots } from './pusaka-runner.js';
import { WorkerSupervisor } from './worker-supervisor.js';

const runtimeConfig = createRuntimeConfig();

ensureLogDirectories();
pruneOldScreenshots();
log('INFO', 'worker starting', {
  WORKER_ID,
  BACKEND_URL,
  HTTP_PORT,
  maxConcurrent: runtimeConfig.maxConcurrent,
  headless: runtimeConfig.headless,
  POLL_MS,
  CONFIG_SYNC_MS,
  WORKER_HEARTBEAT_MS,
});

const supervisor = new WorkerSupervisor(runtimeConfig);
startHealthServer(() => supervisor.getStatus());

// Pastikan Chromium ada sebelum mulai memproses job (DomCloud sering
// menghapus browser cache; watchdog akan install ulang otomatis).
const browserReady = await ensureBrowserInstalled();
if (!browserReady) {
  log('ERROR', 'chromium unavailable at startup — watchdog akan coba install ulang', {});
}
startBrowserWatchdog(BROWSER_CHECK_MS);

supervisor.start();
