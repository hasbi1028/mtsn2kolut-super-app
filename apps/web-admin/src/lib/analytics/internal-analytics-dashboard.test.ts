import { describe, expect, it } from 'vitest';
import {
	hasInternalAnalyticsData,
	internalAnalyticsExportFilename,
	normalizeInternalAnalyticsDailyResult,
	normalizeInternalAnalyticsOverview,
	normalizeInternalAnalyticsSummary
} from './internal-analytics-dashboard';

describe('internal analytics dashboard helpers', () => {
	it('normalizes null and sparse aggregate payloads into safe empty arrays', () => {
		const summary = normalizeInternalAnalyticsSummary(null, 14);
		const daily = normalizeInternalAnalyticsDailyResult({ days: '14', items: null }, 30, 'dashboard');
		const overview = normalizeInternalAnalyticsOverview({ total_count: 0 }, undefined, 14, 'dashboard');

		expect(summary).toEqual({ days: 14, total_count: 0, groups: [], top_events: [] });
		expect(daily).toEqual({ days: 14, event_group: 'dashboard', items: [] });
		expect(overview.summary.groups).toEqual([]);
		expect(overview.daily.items).toEqual([]);
		expect(hasInternalAnalyticsData(overview)).toBe(false);
	});

	it('keeps only complete aggregate rows and derives totals when needed', () => {
		const summary = normalizeInternalAnalyticsSummary({
			days: 7,
			groups: [{ event_group: 'dashboard', count: '3' }, { event_group: '', count: 2 }],
			top_events: [{ event_name: 'dashboard.view', event_group: 'dashboard', count: 3 }]
		});
		const daily = normalizeInternalAnalyticsDailyResult({
			days: 7,
			items: [
				{ aggregate_date: '2026-05-09', event_group: 'dashboard', event_name: 'dashboard.view', source_surface: 'web_admin', count: '3' },
				{ aggregate_date: '', event_group: 'dashboard', event_name: 'dashboard.view', source_surface: 'web_admin', count: 2 }
			]
		});

		expect(summary.total_count).toBe(3);
		expect(summary.groups).toEqual([{ event_group: 'dashboard', count: 3 }]);
		expect(daily.items).toEqual([
			{ aggregate_date: '2026-05-09', event_group: 'dashboard', event_name: 'dashboard.view', source_surface: 'web_admin', count: 3 }
		]);
		expect(hasInternalAnalyticsData({ summary, daily })).toBe(true);
	});

	it('extracts and sanitizes aggregate CSV filenames from content disposition', () => {
		expect(internalAnalyticsExportFilename('attachment; filename="internal-analytics-aggregate.csv"')).toBe('internal-analytics-aggregate.csv');
		expect(internalAnalyticsExportFilename("attachment; filename*=UTF-8''internal-analytics-aggregate%202026.csv")).toBe('internal-analytics-aggregate 2026.csv');
		expect(internalAnalyticsExportFilename('attachment; filename="../raw.csv"')).toBe('..-raw.csv');
		expect(internalAnalyticsExportFilename(null)).toBe('internal-analytics-aggregate.csv');
	});
});
