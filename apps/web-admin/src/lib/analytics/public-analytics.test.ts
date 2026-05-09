import { describe, expect, it, vi } from 'vitest';

import {
	PUBLIC_ANALYTICS_RUNTIME_STATE,
	buildPublicAnalyticsPayload,
	trackPublicAnalyticsEvent
} from './public-analytics';

describe('public analytics fail-closed scaffold', () => {
	it('keeps public website runtime deferred until an unauthenticated collector policy exists', async () => {
		const fetcher = vi.fn();

		await expect(trackPublicAnalyticsEvent('public.page_view', {
			pathname: '/berita?utm_source=campaign&token=secret'
		}, fetcher)).resolves.toEqual({ sent: false, reason: 'public_collector_deferred' });

		expect(PUBLIC_ANALYTICS_RUNTIME_STATE).toBe('deferred_fail_closed');
		expect(fetcher).not.toHaveBeenCalled();
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
});
