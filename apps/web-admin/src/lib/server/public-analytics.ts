import { createHash } from 'node:crypto';

const PUBLIC_ANALYTICS_EVENTS = new Set([
	'public.page_view',
	'public.cta_click',
	'public.download',
	'public.search',
	'public.form_start',
	'public.form_submit'
]);

const PUBLIC_METADATA_KEYS = new Set([
	'page_key',
	'page_kind',
	'cta_key',
	'cta_group',
	'link_kind',
	'file_kind',
	'download_kind',
	'search_bucket',
	'search_length_bucket',
	'form_key',
	'form_step',
	'result',
	'device_class',
	'source_component'
]);

const FORBIDDEN_VALUE_PATTERN = /:\/\/|[?&=]|@|\b\d{8,}\b/i;

export type PublicAnalyticsCollectorPayload = {
	event_name: string;
	route_group: string;
	result?: string;
	metadata: Record<string, string | number | boolean>;
};

type RateEntry = {
	count: number;
	resetAt: number;
};

const RATE_WINDOW_MS = 60_000;
const RATE_LIMIT = 60;
const rateBuckets = new Map<string, RateEntry>();

export function publicAnalyticsRateAllowed(clientAddress: string, now = Date.now()): boolean {
	const key = hashClientAddress(clientAddress);
	const current = rateBuckets.get(key);
	if (!current || current.resetAt <= now) {
		rateBuckets.set(key, { count: 1, resetAt: now + RATE_WINDOW_MS });
		return true;
	}
	if (current.count >= RATE_LIMIT) return false;
	current.count += 1;
	return true;
}

export function sanitizePublicAnalyticsCollectorPayload(input: unknown): PublicAnalyticsCollectorPayload | null {
	if (!isRecord(input)) return null;
	const eventName = safeEventName(input.event_name);
	if (!eventName) return null;
	const metadata = sanitizePublicMetadata(isRecord(input.metadata) ? input.metadata : {});
	const result = safeResult(input.result, metadata);
	return {
		event_name: eventName,
		route_group: safeToken(input.route_group, 'home'),
		...(result ? { result } : {}),
		metadata
	};
}

export function resetPublicAnalyticsRateBucketsForTest() {
	rateBuckets.clear();
}

function hashClientAddress(value: string): string {
	return createHash('sha256').update(value || 'unknown').digest('hex').slice(0, 32);
}

function sanitizePublicMetadata(input: Record<string, unknown>): Record<string, string | number | boolean> {
	const output: Record<string, string | number | boolean> = {};
	for (const [key, value] of Object.entries(input)) {
		const normalizedKey = normalizeKey(key);
		if (!PUBLIC_METADATA_KEYS.has(normalizedKey)) continue;
		const sanitized = sanitizePublicValue(value);
		if (sanitized !== undefined) output[normalizedKey] = sanitized;
	}
	return output;
}

function sanitizePublicValue(value: unknown): string | number | boolean | undefined {
	if (typeof value === 'boolean') return value;
	if (typeof value === 'number' && Number.isFinite(value) && value >= 0 && value <= 10000) return Math.floor(value);
	if (typeof value === 'string') {
		const token = safeToken(value, '');
		return token || undefined;
	}
	return undefined;
}

function safeEventName(value: unknown): string {
	if (typeof value !== 'string') return '';
	const trimmed = value.trim();
	return PUBLIC_ANALYTICS_EVENTS.has(trimmed) ? trimmed : '';
}

function safeResult(value: unknown, metadata: Record<string, string | number | boolean>): string {
	const raw = typeof value === 'string' ? value : typeof metadata.result === 'string' ? metadata.result : '';
	const token = safeToken(raw, '');
	return ['success', 'failed', 'error', 'validation_failed', 'started'].includes(token) ? token : '';
}

function safeToken(value: unknown, fallback: string): string {
	if (typeof value !== 'string') return fallback;
	const normalized = value.trim().toLowerCase();
	if (!normalized || FORBIDDEN_VALUE_PATTERN.test(normalized)) return fallback;
	const safe = normalized
		.replace(/[-/\s]+/g, '_')
		.replace(/[^a-z0-9_]/g, '')
		.replace(/^_+|_+$/g, '')
		.slice(0, 64);
	return safe || fallback;
}

function normalizeKey(key: string): string {
	return key.trim()
		.replace(/([a-z0-9])([A-Z])/g, '$1_$2')
		.toLowerCase()
		.replace(/[-.\s]+/g, '_')
		.replace(/^_+|_+$/g, '');
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null && !Array.isArray(value);
}
