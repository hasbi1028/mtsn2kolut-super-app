export const INTERNAL_ANALYTICS_EVENT_GROUPS = {
	'dashboard.view': 'dashboard',
	'dashboard.refresh': 'dashboard',
	'dashboard.widget_view': 'dashboard',
	'dashboard.filter_change': 'dashboard',
	'dashboard.export': 'dashboard',
	'bank_soal.list_view': 'bank_soal',
	'bank_soal.editor_open': 'bank_soal',
	'bank_soal.draft_save': 'bank_soal',
	'bank_soal.question_create': 'bank_soal',
	'bank_soal.question_update': 'bank_soal',
	'bank_soal.import_start': 'bank_soal',
	'bank_soal.import_complete': 'bank_soal',
	'bank_soal.export': 'bank_soal',
	'bank_soal.review_view': 'bank_soal',
	'bank_soal.review_decision': 'bank_soal',
	'bank_soal.readiness_check': 'bank_soal',
	'bank_soal.asset_upload': 'bank_soal',
	'asesmen.hub_view': 'asesmen',
	'asesmen.package_create': 'asesmen',
	'asesmen.package_update': 'asesmen',
	'asesmen.event_create': 'asesmen',
	'asesmen.session_create': 'asesmen',
	'asesmen.session_update': 'asesmen',
	'asesmen.proctoring_view': 'asesmen',
	'asesmen.participant_action': 'asesmen',
	'asesmen.result_view': 'asesmen',
	'asesmen.result_export': 'asesmen',
	'asesmen.non_test_sync': 'asesmen',
	'pusaka.dashboard_view': 'pusaka',
	'pusaka.manual_run': 'pusaka',
	'pusaka.scheduler_tick': 'pusaka',
	'pusaka.queue_cancel': 'pusaka',
	'pusaka.employee_scope_update': 'pusaka',
	'pusaka.settings_update': 'pusaka',
	'users.list_view': 'users',
	'users.create': 'users',
	'users.update': 'users',
	'users.reset_password': 'users',
	'users.suspend_change': 'users',
	'rbac.roles_view': 'rbac',
	'rbac.permission_update': 'rbac',
	'rbac.permission_denied': 'rbac',
	'security.settings_view': 'security',
	'security.settings_update': 'security',
	'security.analytics_view': 'security',
	'security.analytics_filter': 'security',
	'security.forbidden': 'security',
	'security.export_requested': 'security'
} as const;

const FORBIDDEN_KEYS = new Set([
	'password',
	'passphrase',
	'pin',
	'otp',
	'token',
	'access_token',
	'refresh_token',
	'exam_token',
	'csrf_token',
	'session_token',
	'cookie',
	'set_cookie',
	'authorization',
	'auth_header',
	'bearer',
	'jwt',
	'secret',
	'api_key',
	'apikey',
	'client_secret',
	'nik',
	'nip',
	'nisn',
	'nisn_full',
	'nip_full',
	'device_fingerprint',
	'fingerprint',
	'device_fingerprint_hash',
	'raw_ip',
	'ip_address_raw',
	'remote_addr',
	'x_forwarded_for_raw',
	'raw_user_agent',
	'user_agent_raw',
	'ua_raw',
	'user_agent',
	'query_string',
	'raw_query',
	'url',
	'full_url',
	'search',
	'query',
	'query_string',
	'url',
	'full_url',
	'request_body',
	'response_body',
	'raw_payload',
	'form_value',
	'form_values',
	'sql',
	'stack_trace'
]);

type AnalyticsMetadataValue = string | number | boolean | null | AnalyticsMetadataValue[] | { [key: string]: AnalyticsMetadataValue };
type AnalyticsMetadata = Record<string, AnalyticsMetadataValue>;
type Fetcher = typeof fetch;

export type InternalAnalyticsPayload = {
	event_name: string;
	event_group: string;
	source_surface: 'web_admin';
	permission_code?: string;
	route_group: string;
	module: string;
	result?: string;
	status_code_class?: string;
	duration_bucket?: string;
	metadata: AnalyticsMetadata;
};

export type InternalAnalyticsOptions = {
	pathname?: string;
	role?: string;
	permissionCode?: string;
	routeGroup?: string;
	module?: string;
	result?: string;
	statusCodeClass?: string;
	durationBucket?: string;
	metadata?: Record<string, unknown>;
};

export function buildInternalAnalyticsPayload(eventName: string, options: InternalAnalyticsOptions = {}): InternalAnalyticsPayload {
	const eventGroup = INTERNAL_ANALYTICS_EVENT_GROUPS[eventName as keyof typeof INTERNAL_ANALYTICS_EVENT_GROUPS];
	if (!eventGroup) {
		throw new Error('internal analytics event is not allowlisted');
	}
	const routeGroup = safeSegment(options.routeGroup) || routeGroupFromPath(options.pathname) || eventGroup;
	const moduleName = safeSegment(options.module) || routeGroup;
	const metadata = sanitizeAnalyticsMetadata({
		...options.metadata,
		role: options.role
	});
	return {
		event_name: eventName,
		event_group: eventGroup,
		source_surface: 'web_admin',
		...(options.permissionCode ? { permission_code: options.permissionCode } : {}),
		route_group: routeGroup,
		module: moduleName,
		...(options.result ? { result: options.result } : {}),
		...(options.statusCodeClass ? { status_code_class: options.statusCodeClass } : {}),
		...(options.durationBucket ? { duration_bucket: options.durationBucket } : {}),
		metadata
	};
}

export async function trackInternalAnalyticsEvent(
	eventName: string,
	options: InternalAnalyticsOptions = {},
	fetcher: Fetcher = fetch
): Promise<boolean> {
	try {
		const payload = buildInternalAnalyticsPayload(eventName, options);
		const response = await fetcher('/api/internal-analytics/events', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload),
			keepalive: true
		});
		return response.ok;
	} catch {
		return false;
	}
}

export function trackInternalPageView(eventName: string, pathname?: string, fetcher: Fetcher = fetch) {
	return trackInternalAnalyticsEvent(eventName, { pathname }, fetcher);
}

export function sanitizeAnalyticsMetadata(input: Record<string, unknown> | undefined): AnalyticsMetadata {
	const output: AnalyticsMetadata = {};
	if (!input) return output;
	for (const [key, value] of Object.entries(input)) {
		const normalizedKey = normalizeAnalyticsKey(key);
		if (!normalizedKey || FORBIDDEN_KEYS.has(normalizedKey)) continue;
		const sanitized = sanitizeAnalyticsValue(value);
		if (sanitized !== undefined) {
			output[normalizedKey] = sanitized;
		}
	}
	return output;
}

function sanitizeAnalyticsValue(value: unknown): AnalyticsMetadataValue | undefined {
	if (value === null || typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean') {
		if (typeof value === 'string' && looksLikeRawQuery(value)) return undefined;
		return value;
	}
	if (Array.isArray(value)) {
		const items = value
			.map((item) => sanitizeAnalyticsValue(item))
			.filter((item): item is AnalyticsMetadataValue => item !== undefined);
		return items;
	}
	if (typeof value === 'object') {
		const nested = sanitizeAnalyticsMetadata(value as Record<string, unknown>);
		return Object.keys(nested).length > 0 ? nested : undefined;
	}
	return undefined;
}

function normalizeAnalyticsKey(key: string) {
	return key.trim()
		.replace(/([a-z0-9])([A-Z])/g, '$1_$2')
		.toLowerCase()
		.replace(/[-.\s]+/g, '_')
		.replace(/^_+|_+$/g, '');
}

function routeGroupFromPath(pathname: string | undefined) {
	const clean = (pathname ?? '').split(/[?#]/, 1)[0] ?? '';
	const first = clean.split('/').filter(Boolean)[0] ?? 'dashboard';
	return safeSegment(first);
}

function safeSegment(value: string | undefined) {
	return (value ?? '').trim().toLowerCase().replace(/-/g, '_').replace(/[^a-z0-9_]/g, '');
}

function looksLikeRawQuery(value: string) {
	return value.startsWith('?') || /[?&](token|access_token|refresh_token|nisn|nip|nik|utm_|question_id)=/i.test(value);
}
