import { describe, expect, it, vi } from 'vitest';

import {
	PUBLIC_ANALYTICS_RUNTIME_STATE,
	buildPublicAnalyticsPayload,
	trackPublicAnalyticsEvent
} from './public-analytics';

describe('public analytics first-party helper', () => {
	it('posts sanitized public website payloads to the first-party BFF collector', async () => {
		const fetcher = vi.fn().mockResolvedValue(new Response(null, { status: 204 }));

		await expect(trackPublicAnalyticsEvent('public.page_view', {
			pathname: '/berita?utm_source=campaign&token=secret',
			metadata: { page_key: 'berita', raw_user_agent: 'Mozilla/5.0', unknown_safe: 'ignored' }
		}, fetcher)).resolves.toBe(true);

		expect(PUBLIC_ANALYTICS_RUNTIME_STATE).toBe('active_first_party');
		expect(fetcher).toHaveBeenCalledWith('/api/public/analytics/events', expect.objectContaining({
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			keepalive: true
		}));
		const body = JSON.parse(fetcher.mock.calls[0][1].body as string) as Record<string, unknown>;
		expect(body).toMatchObject({
			event_name: 'public.page_view',
			event_group: 'public',
			source_surface: 'public_website',
			route_group: 'berita',
			module: 'public_site',
			metadata: { page_key: 'berita' }
		});
		expect(JSON.stringify(body)).not.toContain('utm_source');
		expect(JSON.stringify(body)).not.toContain('Mozilla');
		expect(JSON.stringify(body)).not.toContain('unknown_safe');
	});

	it('can build sanitized public payload previews without sending them', () => {
		const payload = buildPublicAnalyticsPayload('public.page_view', {
			pathname: '/pengumuman?nisn=123&utm_source=secret',
			metadata: {
				page_key: 'pengumuman',
				raw_user_agent: 'Mozilla/5.0',
				form_value: 'isi form',
				token: 'secret'
			}
		});

		expect(payload).toMatchObject({
			event_name: 'public.page_view',
			event_group: 'public',
			source_surface: 'public_website',
			route_group: 'pengumuman',
			module: 'public_site',
			metadata: { page_key: 'pengumuman' }
		});
		expect(JSON.stringify(payload)).not.toContain('nisn');
		expect(JSON.stringify(payload)).not.toContain('utm_source');
		expect(JSON.stringify(payload)).not.toContain('Mozilla');
		expect(JSON.stringify(payload)).not.toContain('isi form');
		expect(JSON.stringify(payload)).not.toContain('secret');
	});

	it('fails silent before sending non-allowlisted public events', async () => {
		const fetcher = vi.fn();

		await expect(trackPublicAnalyticsEvent('dashboard.view', {}, fetcher)).resolves.toBe(false);

		expect(fetcher).not.toHaveBeenCalled();
	});
});
