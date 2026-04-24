import fs from 'fs';
import path from 'path';

const LOG_PATH = process.env.APP_LOG_PATH || path.resolve('../logs/app.log');
fs.mkdirSync(path.dirname(LOG_PATH), { recursive: true });

export function logInfo(message, context = {}) {
  writeLog('INFO', message, context);
}

export function logWarn(message, context = {}) {
  writeLog('WARN', message, context);
}

export function logError(message, context = {}) {
  writeLog('ERROR', message, context);
}

function writeLog(level, message, context) {
  const line = `${new Date().toISOString()} [${level}] ${message} ${JSON.stringify(context)}\n`;
  try {
    fs.appendFileSync(LOG_PATH, line);
  } catch {
    // ignore file logging failure; keep console output
  }
  if (level === 'ERROR') console.error(line.trim());
  else if (level === 'WARN') console.warn(line.trim());
  else console.log(line.trim());
}
