import 'dotenv/config';

import {
  BACKEND_URL,
  CONFIG_SYNC_MS,
  POLL_MS,
  WORKER_ID,
  createRuntimeConfig,
} from './config.js';
import { ensureLogDirectories, log } from './logger.js';
import { pruneOldScreenshots } from './pusaka-runner.js';
import { WorkerSupervisor } from './worker-supervisor.js';

const runtimeConfig = createRuntimeConfig();

ensureLogDirectories();
pruneOldScreenshots();
log('INFO', 'worker starting', {
  WORKER_ID,
  BACKEND_URL,
  maxConcurrent: runtimeConfig.maxConcurrent,
  headless: runtimeConfig.headless,
  POLL_MS,
  CONFIG_SYNC_MS,
});

new WorkerSupervisor(runtimeConfig).start();
