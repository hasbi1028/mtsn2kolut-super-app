import os from 'os';
import path from 'path';
import process from 'process';

import type { RuntimeConfig } from './types.js';

export const BACKEND_URL = (
  process.env.BACKEND_URL ?? 'http://localhost:8080'
).replace(/\/$/, '');
export const WORKER_API_KEY = process.env.WORKER_API_KEY ?? '';
export const WORKER_ID =
  process.env.WORKER_ID ?? `worker-${os.hostname()}-${process.pid}`;
export const DEFAULT_MAX_CONCURRENT = Math.max(
  1,
  Number(process.env.WORKER_CONCURRENCY ?? 5),
);
export const DEFAULT_HEADLESS = ['1', 'true'].includes(
  process.env.HEADLESS ?? 'true',
);
export const POLL_MS = Number(process.env.POLL_MS ?? 8000);
export const CONFIG_SYNC_MS = Math.max(
  5000,
  Number(process.env.CONFIG_SYNC_MS ?? 30000),
);
export const SCRAPE_RETRIES = Math.max(
  1,
  Number(process.env.SCRAPE_RETRIES ?? 3),
);
export const SCRAPE_RETRY_MS = Math.max(
  3000,
  Number(process.env.SCRAPE_RETRY_MS ?? 5000),
);
export const ACTION_TIMEOUT = Math.max(
  5000,
  Number(process.env.ACTION_TIMEOUT ?? 20000),
);
export const LOG_PATH =
  process.env.WORKER_LOG_PATH ?? path.resolve('../logs/worker.log');
export const SCREENSHOT_DIR =
  process.env.SCREENSHOT_DIR ?? path.resolve('../logs/screenshots');

export const BASE_URL = 'https://pusaka-v3.kemenag.go.id';
export const BASE_LAT = -3.2163111;
export const BASE_LNG = 121.0428659;

export const BROWSER_ARGS = [
  '--no-sandbox',
  '--disable-dev-shm-usage',
  '--disable-gpu',
  '--disable-extensions',
  '--disable-background-networking',
];

export const USER_AGENT =
  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36';

export function createRuntimeConfig(): RuntimeConfig {
  return {
    maxConcurrent: DEFAULT_MAX_CONCURRENT,
    headless: DEFAULT_HEADLESS,
  };
}
