import 'dotenv/config';

import {
  BACKEND_URL,
  CONFIG_SYNC_MS,
  HTTP_PORT,
  POLL_MS,
  WORKER_HEARTBEAT_MS,
  WORKER_ID,
  createRuntimeConfig,
} from './config.js';
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
supervisor.start();
