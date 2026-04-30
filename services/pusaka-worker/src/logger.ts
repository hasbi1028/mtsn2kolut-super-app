import fs from 'fs';
import path from 'path';

import { LOG_PATH } from './config.js';

export type LogLevel = 'INFO' | 'WARN' | 'ERROR';

export function ensureLogDirectories(): void {
  fs.mkdirSync(path.dirname(LOG_PATH), { recursive: true });
}

export function log(
  level: LogLevel,
  message: string,
  context: Record<string, unknown> = {},
): void {
  const line = `${new Date().toISOString()} [${level}] ${message} ${JSON.stringify(context)}`;
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
