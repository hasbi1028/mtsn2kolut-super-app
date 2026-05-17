import fs from 'fs';
import crypto from 'crypto';
import path from 'path';

import { LOG_PATH } from './config.js';

export type LogLevel = 'INFO' | 'WARN' | 'ERROR';

const SENSITIVE_KEY_PATTERN = /(authorization|cookie|set-cookie|password|passwd|token|secret|api[_-]?key|database[_-]?url|connection[_-]?string|jwt)/i;
const SENSITIVE_VALUE_PATTERN = /\b(Bearer\s+[A-Za-z0-9._~+\/-]+=*|(?:access_token|refresh_token|token|password|secret|cookie|authorization)\s*[:=]\s*[^\s,;}&]+)/i;

function hashValue(value: unknown): string {
  return crypto.createHash('sha256').update(String(value)).digest('hex').slice(0, 12);
}

function sanitizeContext(value: unknown): unknown {
  if (Array.isArray(value)) {
    return value.map((item) => sanitizeContext(item));
  }
  if (!value || typeof value !== 'object') {
    if (typeof value === 'string') {
      if (SENSITIVE_VALUE_PATTERN.test(value)) return '[REDACTED]';
      if (value.length > 1000) return `${value.slice(0, 1000)}…[truncated]`;
    }
    return value;
  }

  const safe: Record<string, unknown> = {};
  for (const [key, entry] of Object.entries(value)) {
    const normalizedKey = key.toLowerCase();
    if (normalizedKey === 'username' || normalizedKey === 'pusaka_username') {
      safe.username_hash = hashValue(entry);
      continue;
    }
    if (SENSITIVE_KEY_PATTERN.test(normalizedKey)) {
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
  const record = {
    time: new Date().toISOString(),
    service: 'pusaka-worker',
    level: level.toLowerCase(),
    event: message,
    ...sanitizeContext(context) as Record<string, unknown>,
  };
  const line = JSON.stringify(record);
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

export const __loggerInternalsForTest = { sanitizeContext };
