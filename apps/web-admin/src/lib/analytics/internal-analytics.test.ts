import { describe, expect, it, vi } from 'vitest';

import {
	buildInternalAnalyticsPayload,
	trackInternalAnalyticsEvent,
	trackInternalPageView
} from './internal-analytics';

describe('internal analytics client helper', () => {
	it('builds allowlisted web-admin payloads without raw query or sensitive metadata', () => {
		const payload = buildInternalAnalyticsPayload('dashboard.view', {
			pathname: '/?token=secret&nisn=123',
			role: 'admin',
			metadata: {
				page_key: 'dashboard',
				route_group: 'dashboard',
				raw_user_agent: 'Mozilla/5.0 must not be sent',
				nested: { NIK: '7400000000000001', safe_label: 'ok' },
				search: '?utm_source=secret'
			}
		});

		expect(payload).toMatchObject({
			event_name: 'dashboard.view',
			event_group: 'dashboard',
			source_surface: 'web_admin',
			route_group: 'dashboard',
			module: 'dashboard',
			metadata: {
				page_key: 'dashboard',
				route_group: 'dashboard',
				nested: { safe_label: 'ok' }
			}
		});
		expect(JSON.stringify(payload)).not.toContain('token=secret');
		expect(JSON.stringify(payload)).not.toContain('nisn');
		expect(JSON.stringify(payload)).not.toContain('raw_user_agent');
		expect(JSON.stringify(payload)).not.toContain('Mozilla');
		expect(JSON.stringify(payload)).not.toContain('7400000000000001');
		expect(JSON.stringify(payload)).not.toContain('utm_source');
	});

	it('rejects non-allowlisted internal analytics event names before fetch', async () => {
		const fetcher = vi.fn();

		await expect(trackInternalAnalyticsEvent('dashboard.every_click', {}, fetcher)).resolves.toBe(false);

		expect(fetcher).not.toHaveBeenCalled();
	});

	it('posts to the BFF route and fails silent for UI callers', async () => {
		const fetcher = vi.fn()
			.mockRejectedValueOnce(new Error('network down'))
			.mockResolvedValueOnce(new Response('{}', { status: 201 }));

		await expect(trackInternalAnalyticsEvent('bank_soal.list_view', {
			pathname: '/bank-soal?question_id=secret',
			role: 'guru',
			metadata: { page_key: 'bank_soal' }
		}, fetcher)).resolves.toBe(false);
		await expect(trackInternalPageView('bank_soal.list_view', '/bank-soal?x=secret', fetcher)).resolves.toBe(true);

		expect(fetcher).toHaveBeenCalledTimes(2);
		expect(fetcher).toHaveBeenLastCalledWith('/api/internal-analytics/events', expect.objectContaining({
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			keepalive: true
		}));
		const body = JSON.parse(fetcher.mock.calls[1][1].body as string) as Record<string, unknown>;
		expect(body).toMatchObject({
			event_name: 'bank_soal.list_view',
			event_group: 'bank_soal',
			source_surface: 'web_admin',
			route_group: 'bank_soal',
			module: 'bank_soal'
		});
		expect(JSON.stringify(body)).not.toContain('question_id');
		expect(JSON.stringify(body)).not.toContain('x=secret');
	});
});
