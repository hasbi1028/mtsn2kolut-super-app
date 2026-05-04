import os from 'os';
import path from 'path';
import process from 'process';

import type { RuntimeConfig } from './types.js';

export const RUNTIME_CONFIG_LIMITS = {
  maxConcurrent: { min: 1, max: 50 },
} as const;

const NODE_ENV = (process.env.NODE_ENV ?? '').toLowerCase();
const WORKER_ENV = (process.env.WORKER_ENV ?? '').toLowerCase();
const APP_ENV = (process.env.APP_ENV ?? '').toLowerCase();

function parsePositiveNumber(
  name: string,
  defaultValue: number,
  options: { min: number; max: number; integer?: boolean },
): number {
  const raw = process.env[name];
  const parsed = raw === undefined || raw === '' ? defaultValue : Number(raw);
  if (!Number.isFinite(parsed) || parsed < options.min || parsed > options.max) {
    throw new Error(
      `Invalid worker config: ${name} must be a finite number between ${options.min} and ${options.max}`,
    );
  }
  return options.integer ? Math.floor(parsed) : parsed;
}

function parseBoolean(name: string, defaultValue: boolean): boolean {
  const raw = process.env[name];
  if (raw === undefined || raw === '') {
    return defaultValue;
  }
  if (['1', 'true', 'yes'].includes(raw.toLowerCase())) {
    return true;
  }
  if (['0', 'false', 'no'].includes(raw.toLowerCase())) {
    return false;
  }
  throw new Error(`Invalid worker config: ${name} must be true or false`);
}

function parseBackendUrl(raw: string): URL {
  let url: URL;
  try {
    url = new URL(raw);
  } catch {
    throw new Error('Invalid worker config: BACKEND_URL must be a valid http(s) URL');
  }

  if (url.protocol !== 'http:' && url.protocol !== 'https:') {
    throw new Error('Invalid worker config: BACKEND_URL must use http or https');
  }
  if (url.username || url.password) {
    throw new Error('Invalid worker config: BACKEND_URL must not include credentials');
  }
  if (url.search || url.hash) {
    throw new Error('Invalid worker config: BACKEND_URL must not include query or fragment');
  }
  return url;
}

function normalizeBackendUrl(raw: string): string {
  const url = parseBackendUrl(raw);
  const pathname = url.pathname === '/' ? '' : url.pathname.replace(/\/+$/, '');
  return `${url.origin}${pathname}`;
}

function isLocalOrTestEnv(backendUrl: string): boolean {
  if (['test', 'local', 'development'].includes(NODE_ENV)) {
    return true;
  }
  if (['test', 'local', 'development'].includes(WORKER_ENV)) {
    return true;
  }
  if (['test', 'local', 'development'].includes(APP_ENV)) {
    return true;
  }
  const hostname = parseBackendUrl(backendUrl).hostname.toLowerCase();
  return hostname === 'localhost' || hostname === '127.0.0.1' || hostname === '[::1]' || hostname === '::1';
}

export const BACKEND_URL = normalizeBackendUrl(process.env.BACKEND_URL ?? 'http://localhost:8080');
export const WORKER_API_KEY = process.env.WORKER_API_KEY ?? '';
if (!WORKER_API_KEY && !isLocalOrTestEnv(BACKEND_URL)) {
  throw new Error('Invalid worker config: WORKER_API_KEY is required');
}
export const WORKER_ID =
  process.env.WORKER_ID ?? `worker-${os.hostname()}-${process.pid}`;
export const DEFAULT_MAX_CONCURRENT = parsePositiveNumber('WORKER_CONCURRENCY', 5, {
  min: RUNTIME_CONFIG_LIMITS.maxConcurrent.min,
  max: RUNTIME_CONFIG_LIMITS.maxConcurrent.max,
  integer: true,
});
export const DEFAULT_HEADLESS = parseBoolean('HEADLESS', true);
export const POLL_MS = parsePositiveNumber('POLL_MS', 8000, {
  min: 500,
  max: 60000,
  integer: true,
});
export const CONFIG_SYNC_MS = parsePositiveNumber('CONFIG_SYNC_MS', 30000, {
  min: 5000,
  max: 300000,
  integer: true,
});
export const SCRAPE_RETRIES = parsePositiveNumber('SCRAPE_RETRIES', 3, {
  min: 1,
  max: 10,
  integer: true,
});
export const SCRAPE_RETRY_MS = parsePositiveNumber('SCRAPE_RETRY_MS', 5000, {
  min: 1000,
  max: 60000,
  integer: true,
});
export const ACTION_TIMEOUT = parsePositiveNumber('ACTION_TIMEOUT', 20000, {
  min: 5000,
  max: 120000,
  integer: true,
});
export const WORKER_API_TIMEOUT_MS = parsePositiveNumber('WORKER_API_TIMEOUT_MS', 10000, {
  min: 1000,
  max: 60000,
  integer: true,
});
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

function normalizeRuntimeBoolean(value: unknown): boolean | undefined {
  if (typeof value === 'boolean') {
    return value;
  }
  if (typeof value !== 'string') {
    return undefined;
  }

  const normalized = value.toLowerCase();
  if (['1', 'true', 'yes'].includes(normalized)) {
    return true;
  }
  if (['0', 'false', 'no'].includes(normalized)) {
    return false;
  }
  return undefined;
}

export function normalizeRuntimeConfigPatch(input: unknown): Partial<RuntimeConfig> {
  const data = input && typeof input === 'object' ? input as Record<string, unknown> : {};
  const next: Partial<RuntimeConfig> = {};

  if (data.max_concurrent !== undefined || data.maxConcurrent !== undefined) {
    const rawMaxConcurrent = data.max_concurrent ?? data.maxConcurrent;
    const parsed = typeof rawMaxConcurrent === 'number'
      ? rawMaxConcurrent
      : Number(rawMaxConcurrent);
    if (
      Number.isFinite(parsed) &&
      parsed >= RUNTIME_CONFIG_LIMITS.maxConcurrent.min &&
      parsed <= RUNTIME_CONFIG_LIMITS.maxConcurrent.max
    ) {
      next.maxConcurrent = Math.floor(parsed);
    }
  }

  if (data.headless !== undefined) {
    const parsedHeadless = normalizeRuntimeBoolean(data.headless);
    if (parsedHeadless !== undefined) {
      next.headless = parsedHeadless;
    }
  }

  return next;
}
