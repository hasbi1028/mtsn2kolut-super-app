import { sanitizeAnalyticsMetadata } from './internal-analytics';

export const PUBLIC_ANALYTICS_RUNTIME_STATE = 'deferred_fail_closed' as const;

const PUBLIC_ANALYTICS_EVENT_GROUPS = {
	'public.page_view': 'public',
	'public.cta_click': 'public',
	'public.download': 'public',
	'public.search': 'public',
	'public.form_start': 'public',
	'public.form_submit': 'public'
} as const;

type Fetcher = typeof fetch;

export type PublicAnalyticsPayload = {
	event_name: string;
	event_group: 'public';
	source_surface: 'public_website';
	route_group: string;
	module: 'public_site';
	metadata: Record<string, unknown>;
};

type PublicAnalyticsOptions = {
	pathname?: string;
	metadata?: Record<string, unknown>;
};

export function buildPublicAnalyticsPayload(eventName: string, options: PublicAnalyticsOptions = {}): PublicAnalyticsPayload {
	const eventGroup = PUBLIC_ANALYTICS_EVENT_GROUPS[eventName as keyof typeof PUBLIC_ANALYTICS_EVENT_GROUPS];
	if (!eventGroup) {
		throw new Error('public analytics event is not allowlisted');
	}
	return {
		event_name: eventName,
		event_group: eventGroup,
		source_surface: 'public_website',
		route_group: publicRouteGroup(options.pathname),
		module: 'public_site',
		metadata: sanitizeAnalyticsMetadata(options.metadata)
	};
}

export async function trackPublicAnalyticsEvent(
	eventName: string,
	options: PublicAnalyticsOptions = {},
	_fetcher: Fetcher = fetch
): Promise<{ sent: false; reason: 'public_collector_deferred' }> {
	buildPublicAnalyticsPayload(eventName, options);
	return { sent: false, reason: 'public_collector_deferred' };
}

function publicRouteGroup(pathname: string | undefined) {
	const clean = (pathname ?? '').split(/[?#]/, 1)[0] ?? '';
	const first = clean.split('/').filter(Boolean)[0] ?? 'home';
	return first.trim().toLowerCase().replace(/-/g, '_').replace(/[^a-z0-9_]/g, '') || 'home';
}
