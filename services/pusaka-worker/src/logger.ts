import fs from 'fs';
import crypto from 'crypto';
import path from 'path';

import { LOG_PATH } from './config.js';

export type LogLevel = 'INFO' | 'WARN' | 'ERROR';

const SENSITIVE_KEYS = new Set([
  'password',
  'pusaka_password',
  'token',
  'cookie',
  'cookies',
  'session',
  'session_state',
  'storage_state',
]);

function hashValue(value: unknown): string {
  return crypto.createHash('sha256').update(String(value)).digest('hex').slice(0, 12);
}

function sanitizeContext(value: unknown): unknown {
  if (Array.isArray(value)) {
    return value.map((item) => sanitizeContext(item));
  }
  if (!value || typeof value !== 'object') {
    return value;
  }

  const safe: Record<string, unknown> = {};
  for (const [key, entry] of Object.entries(value)) {
    const normalizedKey = key.toLowerCase();
    if (normalizedKey === 'username' || normalizedKey === 'pusaka_username') {
      safe.username_hash = hashValue(entry);
      continue;
    }
    if (SENSITIVE_KEYS.has(normalizedKey)) {
      safe[key] = '[REDACTED]';
      continue;
    }
    safe[key] = sanitizeContext(entry);
  }
  return safe;
}

export function ensureLogDirectories(): void {
  fs.mkdirSync(path.dirname(LOG_PATH), { recursive: true });
}

export function log(
  level: LogLevel,
  message: string,
  context: Record<string, unknown> = {},
): void {
  const line = `${new Date().toISOString()} [${level}] ${message} ${JSON.stringify(sanitizeContext(context))}`;
  try {
    fs.appendFileSync(LOG_PATH, `${line}\n`);
  } catch {
    // best-effort
  }

  if (level === 'ERROR') {
    console.error(line);
    return;
  }
  if (level === 'WARN') {
    console.warn(line);
    return;
  }
  console.log(line);
}
