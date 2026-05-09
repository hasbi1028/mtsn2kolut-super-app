import { describe, expect, it } from 'vitest';
import {
	publicAnalyticsRateAllowed,
	resetPublicAnalyticsRateBucketsForTest,
	sanitizePublicAnalyticsCollectorPayload
} from './public-analytics';

describe('public analytics collector guard', () => {
	it('sanitizes public payloads to event and metadata allowlists', () => {
		const payload = sanitizePublicAnalyticsCollectorPayload({
			event_name: 'public.search',
			event_group: 'public',
			source_surface: 'public_website',
			route_group: 'berita?token=secret',
			result: 'success',
			metadata: {
				page_key: 'Berita',
				search_bucket: 'info-umum',
				search_length_bucket: '11_20',
				raw_user_agent: 'Mozilla/5.0',
				query: 'nama calon siswa',
				token: 'secret',
				NISN: '1234567890',
				full_url: 'https://example.test/berita?token=secret',
				unknown_safe: 'ignored'
			}
		});

		expect(payload).toEqual({
			event_name: 'public.search',
			route_group: 'home',
			result: 'success',
			metadata: {
				page_key: 'berita',
				search_bucket: 'info_umum',
				search_length_bucket: '11_20'
			}
		});
		expect(JSON.stringify(payload)).not.toContain('Mozilla');
		expect(JSON.stringify(payload)).not.toContain('nama calon siswa');
		expect(JSON.stringify(payload)).not.toContain('1234567890');
		expect(JSON.stringify(payload)).not.toContain('token=secret');
	});

	it('rejects non-public or malformed collector events', () => {
		expect(sanitizePublicAnalyticsCollectorPayload({ event_name: 'dashboard.view' })).toBeNull();
		expect(sanitizePublicAnalyticsCollectorPayload(null)).toBeNull();
		expect(sanitizePublicAnalyticsCollectorPayload({ event_name: 'public.page_view', metadata: [] })).toEqual({
			event_name: 'public.page_view',
			route_group: 'home',
			metadata: {}
		});
	});

	it('rate limits by hashed client bucket without storing raw keys in payload', () => {
		resetPublicAnalyticsRateBucketsForTest();
		for (let i = 0; i < 60; i += 1) {
			expect(publicAnalyticsRateAllowed('203.0.113.10', 1000)).toBe(true);
		}
		expect(publicAnalyticsRateAllowed('203.0.113.10', 1000)).toBe(false);
		expect(publicAnalyticsRateAllowed('203.0.113.10', 61_001)).toBe(true);
	});
});
