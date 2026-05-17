export type LogLevel = 'debug' | 'info' | 'warn' | 'error';

export type LogRecord = {
	level: LogLevel;
	event: string;
	service?: string;
	module?: string;
	request_id?: string;
	route?: string;
	method?: string;
	status?: number;
	duration_ms?: number;
	error?: string;
	metadata?: Record<string, unknown>;
};

const REDACT_KEY = /(authorization|cookie|set-cookie|password|passwd|token|secret|api[_-]?key|database[_-]?url|connection[_-]?string|jwt)/i;
const REDACT_VALUE = /\b(Bearer\s+[A-Za-z0-9._~+\/-]+=*|(?:access_token|refresh_token|token|password|secret|cookie|authorization)\s*[:=]\s*[^\s,;}&]+)/i;
const MAX_STRING_LENGTH = 1000;

export function redactForStructuredLog(value: unknown): unknown {
	if (Array.isArray(value)) return value.map((item) => redactForStructuredLog(item));
	if (value && typeof value === 'object') {
		const out: Record<string, unknown> = {};
		for (const [key, entry] of Object.entries(value as Record<string, unknown>)) {
			out[key] = REDACT_KEY.test(key) ? '[REDACTED]' : redactForStructuredLog(entry);
		}
		return out;
	}
	if (typeof value === 'string') {
		if (REDACT_VALUE.test(value)) return '[REDACTED]';
		if (value.length > MAX_STRING_LENGTH) return `${value.slice(0, MAX_STRING_LENGTH)}…[truncated]`;
	}
	return value;
}

export function log(record: LogRecord): void {
	const safe = redactForStructuredLog({
		time: new Date().toISOString(),
		service: 'web-admin',
		...record
	});
	const line = JSON.stringify(safe);
	if (record.level === 'error') console.error(line);
	else if (record.level === 'warn') console.warn(line);
	else console.log(line);
}
