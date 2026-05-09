export type InternalAnalyticsSummaryGroup = {
	event_group: string;
	count: number;
};

export type InternalAnalyticsSummaryEvent = {
	event_name: string;
	event_group: string;
	count: number;
};

export type InternalAnalyticsSummary = {
	days: number;
	total_count: number;
	groups: InternalAnalyticsSummaryGroup[];
	top_events: InternalAnalyticsSummaryEvent[];
};

export type InternalAnalyticsDailyItem = {
	aggregate_date: string;
	event_group: string;
	event_name: string;
	source_surface: string;
	role?: string;
	result?: string;
	count: number;
};

export type InternalAnalyticsDailyResult = {
	days: number;
	event_group?: string;
	items: InternalAnalyticsDailyItem[];
};

export type InternalAnalyticsOverview = {
	summary: InternalAnalyticsSummary;
	daily: InternalAnalyticsDailyResult;
};

const DEFAULT_DAYS = 30;
const DEFAULT_EXPORT_FILENAME = 'internal-analytics-aggregate.csv';

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null;
}

function asString(value: unknown): string {
	return typeof value === 'string' ? value.trim() : '';
}

function asCount(value: unknown): number {
	if (typeof value === 'number' && Number.isFinite(value) && value > 0) return Math.floor(value);
	if (typeof value === 'string') {
		const parsed = Number(value);
		if (Number.isFinite(parsed) && parsed > 0) return Math.floor(parsed);
	}
	return 0;
}

function asDays(value: unknown, fallbackDays = DEFAULT_DAYS): number {
	const parsed = asCount(value);
	if (parsed > 0) return parsed;
	return fallbackDays > 0 ? fallbackDays : DEFAULT_DAYS;
}

function asArray(value: unknown): unknown[] {
	return Array.isArray(value) ? value : [];
}

export function normalizeInternalAnalyticsSummary(value: unknown, fallbackDays = DEFAULT_DAYS): InternalAnalyticsSummary {
	const source = isRecord(value) ? value : {};
	const days = asDays(source.days, fallbackDays);
	const groups = asArray(source.groups)
		.map((item) => {
			if (!isRecord(item)) return null;
			const eventGroup = asString(item.event_group);
			if (!eventGroup) return null;
			return { event_group: eventGroup, count: asCount(item.count) };
		})
		.filter((item): item is InternalAnalyticsSummaryGroup => item !== null);
	const topEvents = asArray(source.top_events)
		.map((item) => {
			if (!isRecord(item)) return null;
			const eventName = asString(item.event_name);
			const eventGroup = asString(item.event_group);
			if (!eventName || !eventGroup) return null;
			return { event_name: eventName, event_group: eventGroup, count: asCount(item.count) };
		})
		.filter((item): item is InternalAnalyticsSummaryEvent => item !== null);
	const totalCount = asCount(source.total_count) || groups.reduce((sum, item) => sum + item.count, 0);
	return { days, total_count: totalCount, groups, top_events: topEvents };
}

export function normalizeInternalAnalyticsDailyResult(
	value: unknown,
	fallbackDays = DEFAULT_DAYS,
	fallbackEventGroup = ''
): InternalAnalyticsDailyResult {
	const source = isRecord(value) ? value : {};
	const eventGroup = asString(source.event_group) || fallbackEventGroup;
	const items = asArray(source.items)
		.map((item) => {
			if (!isRecord(item)) return null;
			const aggregateDate = asString(item.aggregate_date);
			const itemEventGroup = asString(item.event_group);
			const eventName = asString(item.event_name);
			const sourceSurface = asString(item.source_surface);
			if (!aggregateDate || !itemEventGroup || !eventName || !sourceSurface) return null;
			const role = asString(item.role);
			const result = asString(item.result);
			return {
				aggregate_date: aggregateDate,
				event_group: itemEventGroup,
				event_name: eventName,
				source_surface: sourceSurface,
				...(role ? { role } : {}),
				...(result ? { result } : {}),
				count: asCount(item.count)
			};
		})
		.filter((item): item is InternalAnalyticsDailyItem => item !== null);
	return { days: asDays(source.days, fallbackDays), ...(eventGroup ? { event_group: eventGroup } : {}), items };
}

export function normalizeInternalAnalyticsOverview(
	summary: unknown,
	daily: unknown,
	fallbackDays = DEFAULT_DAYS,
	fallbackEventGroup = ''
): InternalAnalyticsOverview {
	return {
		summary: normalizeInternalAnalyticsSummary(summary, fallbackDays),
		daily: normalizeInternalAnalyticsDailyResult(daily, fallbackDays, fallbackEventGroup)
	};
}

export function hasInternalAnalyticsData(overview: InternalAnalyticsOverview): boolean {
	return overview.summary.total_count > 0 || overview.summary.groups.length > 0 || overview.summary.top_events.length > 0 || overview.daily.items.length > 0;
}

export function internalAnalyticsExportFilename(disposition: string | null): string {
	const encoded = /filename\*=UTF-8''([^;]+)/i.exec(disposition ?? '')?.[1];
	let raw = /filename="?([^";]+)"?/i.exec(disposition ?? '')?.[1];
	if (encoded) {
		try {
			raw = decodeURIComponent(encoded);
		} catch {
			raw = encoded;
		}
	}
	const sanitized = (raw ?? DEFAULT_EXPORT_FILENAME)
		.trim()
		.replace(/[\\/:*?"<>|\r\n]/g, '-');
	return sanitized || DEFAULT_EXPORT_FILENAME;
}
