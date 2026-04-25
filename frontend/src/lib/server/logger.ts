import fs from 'fs';
import path from 'path';

const LOG_PATH = process.env.APP_LOG_PATH ?? path.resolve('../logs/app.log');
fs.mkdirSync(path.dirname(LOG_PATH), { recursive: true });

type LogContext = Record<string, unknown>;

export function logInfo(message: string, context: LogContext = {}) {
	writeLog('INFO', message, context);
}

export function logWarn(message: string, context: LogContext = {}) {
	writeLog('WARN', message, context);
}

export function logError(message: string, context: LogContext = {}) {
	writeLog('ERROR', message, context);
}

function writeLog(level: 'INFO' | 'WARN' | 'ERROR', message: string, context: LogContext) {
	const line = `${new Date().toISOString()} [${level}] ${message} ${JSON.stringify(context)}\n`;
	try { fs.appendFileSync(LOG_PATH, line); } catch { /* ignore */ }
	if (level === 'ERROR') console.error(line.trim());
	else if (level === 'WARN')  console.warn(line.trim());
	else                        console.log(line.trim());
}
