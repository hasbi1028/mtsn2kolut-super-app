import { sanitizeAnalyticsMetadata } from './internal-analytics';

export const PUBLIC_ANALYTICS_RUNTIME_STATE = 'active_first_party' as const;

const PUBLIC_ANALYTICS_EVENT_GROUPS = {
	'public.page_view': 'public',
	'public.cta_click': 'public',
	'public.download': 'public',
	'public.search': 'public',
	'public.form_start': 'public',
	'public.form_submit': 'public'
} as const;

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

type Fetcher = typeof fetch;

export type PublicAnalyticsPayload = {
	event_name: string;
	event_group: 'public';
	source_surface: 'public_website';
	route_group: string;
	module: 'public_site';
	result?: string;
	metadata: Record<string, unknown>;
};

type PublicAnalyticsOptions = {
	pathname?: string;
	result?: string;
	metadata?: Record<string, unknown>;
};

export function buildPublicAnalyticsPayload(eventName: string, options: PublicAnalyticsOptions = {}): PublicAnalyticsPayload {
	const eventGroup = PUBLIC_ANALYTICS_EVENT_GROUPS[eventName as keyof typeof PUBLIC_ANALYTICS_EVENT_GROUPS];
	if (!eventGroup) {
		throw new Error('public analytics event is not allowlisted');
	}
	const payload: PublicAnalyticsPayload = {
		event_name: eventName,
		event_group: eventGroup,
		source_surface: 'public_website',
		route_group: publicRouteGroup(options.pathname),
		module: 'public_site',
		metadata: sanitizePublicAnalyticsMetadata(options.metadata)
	};
	const result = publicResult(options.result ?? asString(payload.metadata.result));
	if (result) payload.result = result;
	return payload;
}

export async function trackPublicAnalyticsEvent(
	eventName: string,
	options: PublicAnalyticsOptions = {},
	fetcher: Fetcher = fetch
): Promise<boolean> {
	try {
		const payload = buildPublicAnalyticsPayload(eventName, options);
		const response = await fetcher('/api/public/analytics/events', {
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

export function trackPublicPageView(pathname?: string, metadata?: Record<string, unknown>, fetcher: Fetcher = fetch) {
	return trackPublicAnalyticsEvent('public.page_view', { pathname, metadata }, fetcher);
}

function sanitizePublicAnalyticsMetadata(input: Record<string, unknown> | undefined) {
	const sanitized = sanitizeAnalyticsMetadata(input);
	return Object.fromEntries(Object.entries(sanitized).filter(([key]) => PUBLIC_METADATA_KEYS.has(key)));
}

function asString(value: unknown) {
	return typeof value === 'string' ? value.trim() : '';
}

function publicResult(value: string) {
	const normalized = value.trim().toLowerCase().replace(/[^a-z_]/g, '');
	return ['success', 'failed', 'error', 'validation_failed', 'started'].includes(normalized) ? normalized : '';
}

function publicRouteGroup(pathname: string | undefined) {
	const clean = (pathname ?? '').split(/[?#]/, 1)[0] ?? '';
	const first = clean.split('/').filter(Boolean)[0] ?? 'home';
	return first.trim().toLowerCase().replace(/-/g, '_').replace(/[^a-z0-9_]/g, '') || 'home';
}
