import os from 'os';
import path from 'path';
import process from 'process';

import type { RuntimeConfig } from './types.js';

export const RUNTIME_CONFIG_LIMITS = {
  maxConcurrent: { min: 1, max: 50 },
  geoBaseLat: { min: -90, max: 90 },
  geoBaseLng: { min: -180, max: 180 },
  geoRadiusMeters: { min: 1, max: 200 },
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

function isExplicitLocalOrTestEnv(): boolean {
  if (['test', 'local', 'development'].includes(NODE_ENV)) {
    return true;
  }
  if (['test', 'local', 'development'].includes(WORKER_ENV)) {
    return true;
  }
  if (['test', 'local', 'development'].includes(APP_ENV)) {
    return true;
  }
  return false;
}

const RAW_BACKEND_URL = process.env.BACKEND_URL;
if (!RAW_BACKEND_URL && !isExplicitLocalOrTestEnv()) {
  throw new Error('Invalid worker config: BACKEND_URL is required outside local/test/development');
}
export const BACKEND_URL = normalizeBackendUrl(RAW_BACKEND_URL ?? 'http://localhost:8080');
export const WORKER_API_KEY = process.env.WORKER_API_KEY ?? '';
if (!WORKER_API_KEY && !isExplicitLocalOrTestEnv()) {
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
export const WORKER_HEARTBEAT_MS = parsePositiveNumber('WORKER_HEARTBEAT_MS', 30000, {
  min: 5000,
  max: 60000,
  integer: true,
});
// Port HTTP untuk health/metrics server (DomCloud & PM2). Prioritas:
// WORKER_PORT → PORT (konvensi platform PaaS) → 8091.
export const HTTP_PORT = parsePositiveNumber(
  'WORKER_PORT',
  parsePositiveNumber('PORT', 8091, { min: 1, max: 65535, integer: true }),
  { min: 1, max: 65535, integer: true },
);
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
    geo: {
      baseLat: BASE_LAT,
      baseLng: BASE_LNG,
      defaultRadiusMeters: 50,
      checkinRadiusMeters: 55,
      checkoutRadiusMeters: 28,
    },
  };
}

function normalizeRuntimeNumber(
  data: Record<string, unknown>,
  snakeKey: string,
  camelKey: string,
  options: { min: number; max: number; integer?: boolean },
): number | undefined {
  if (data[snakeKey] === undefined && data[camelKey] === undefined) {
    return undefined;
  }
  const raw = data[snakeKey] ?? data[camelKey];
  const parsed = typeof raw === 'number' ? raw : Number(raw);
  if (!Number.isFinite(parsed) || parsed < options.min || parsed > options.max) {
    return undefined;
  }
  return options.integer ? Math.floor(parsed) : parsed;
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
    const parsed = normalizeRuntimeNumber(
      data,
      'max_concurrent',
      'maxConcurrent',
      { ...RUNTIME_CONFIG_LIMITS.maxConcurrent, integer: true },
    );
    if (parsed !== undefined) {
      next.maxConcurrent = parsed;
    }
  }

  if (data.headless !== undefined) {
    const parsedHeadless = normalizeRuntimeBoolean(data.headless);
    if (parsedHeadless !== undefined) {
      next.headless = parsedHeadless;
    }
  }

  const geo: Partial<RuntimeConfig['geo']> = {};
  const geoInput = data.geo && typeof data.geo === 'object'
    ? data.geo as Record<string, unknown>
    : {};
  const geoData = {
    ...geoInput,
    baseLat: data.geoBaseLat ?? geoInput.baseLat,
    baseLng: data.geoBaseLng ?? geoInput.baseLng,
    defaultRadiusMeters: data.geoDefaultRadiusMeters ?? geoInput.defaultRadiusMeters,
    checkinRadiusMeters: data.geoCheckinRadiusMeters ?? geoInput.checkinRadiusMeters,
    checkoutRadiusMeters: data.geoCheckoutRadiusMeters ?? geoInput.checkoutRadiusMeters,
    ...data,
  };
  const baseLat = normalizeRuntimeNumber(
    geoData,
    'pusaka_geo_base_lat',
    'baseLat',
    RUNTIME_CONFIG_LIMITS.geoBaseLat,
  );
  if (baseLat !== undefined) geo.baseLat = baseLat;

  const baseLng = normalizeRuntimeNumber(
    geoData,
    'pusaka_geo_base_lng',
    'baseLng',
    RUNTIME_CONFIG_LIMITS.geoBaseLng,
  );
  if (baseLng !== undefined) geo.baseLng = baseLng;

  const defaultRadiusMeters = normalizeRuntimeNumber(
    geoData,
    'pusaka_geo_default_radius_m',
    'defaultRadiusMeters',
    { ...RUNTIME_CONFIG_LIMITS.geoRadiusMeters, integer: true },
  );
  if (defaultRadiusMeters !== undefined) geo.defaultRadiusMeters = defaultRadiusMeters;

  const checkinRadiusMeters = normalizeRuntimeNumber(
    geoData,
    'pusaka_geo_checkin_radius_m',
    'checkinRadiusMeters',
    { ...RUNTIME_CONFIG_LIMITS.geoRadiusMeters, integer: true },
  );
  if (checkinRadiusMeters !== undefined) geo.checkinRadiusMeters = checkinRadiusMeters;

  const checkoutRadiusMeters = normalizeRuntimeNumber(
    geoData,
    'pusaka_geo_checkout_radius_m',
    'checkoutRadiusMeters',
    { ...RUNTIME_CONFIG_LIMITS.geoRadiusMeters, integer: true },
  );
  if (checkoutRadiusMeters !== undefined) geo.checkoutRadiusMeters = checkoutRadiusMeters;

  if (Object.keys(geo).length > 0) {
    next.geo = geo as RuntimeConfig['geo'];
  }

  return next;
}
