import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { makassarDateKey, normalizeSesiCbtRow, summarizeSessionRows, type SesiCbtRow } from '$lib/asesmen/sesi-cbt';
import { ApiError, apiPath, apiPathWithQuery, handleRouteError, proxy } from '$lib/server/api';

type ExamOption = {
	id?: string;
	title?: string;
	nama?: string;
};

function asArray(value: unknown): unknown[] {
	if (Array.isArray(value)) return value;
	if (value && typeof value === 'object') {
		const record = value as Record<string, unknown>;
		if (Array.isArray(record.items)) return record.items;
		if (Array.isArray(record.data)) return record.data;
	}
	return [];
}

function queryNumber(params: URLSearchParams, name: string, fallback: number) {
	const parsed = Number(params.get(name) ?? fallback);
	return Number.isFinite(parsed) && parsed >= 0 ? parsed : fallback;
}

function matchesFilters(row: SesiCbtRow, params: URLSearchParams) {
	const status = params.get('status')?.trim();
	const date = params.get('date')?.trim();
	const subjectId = params.get('subject_id')?.trim();
	const includeArchived = params.get('include_archived') === '1' || params.get('include_archived') === 'true';
	if (status && row.status !== status) return false;
	if (!includeArchived && row.status === 'cancelled') return false;
	if (date && makassarDateKey(row.scheduled_start) !== date) return false;
	if (subjectId && ![row.subject_name, row.title].some((value) => String(value ?? '').toLowerCase().includes(subjectId.toLowerCase()))) return false;
	return true;
}

async function sessionsForExam(event: RequestEvent, exam: ExamOption): Promise<SesiCbtRow[]> {
	if (!exam.id) return [];
	try {
		const data = await proxy(event).get(apiPath`/api/cbt/events/${exam.id}/sessions`);
		return asArray(data).map((item) => normalizeSesiCbtRow(item, exam));
	} catch {
		return [];
	}
}

async function fallbackSessions(event: RequestEvent): Promise<SesiCbtRow[]> {
	const examParams = new URLSearchParams({ limit: event.url.searchParams.get('exam_limit') ?? '50' });
	const exams = asArray(await proxy(event).get(apiPathWithQuery('/api/asesmen/exams', examParams))) as ExamOption[];
	const batches = await Promise.all(exams.map((exam) => sessionsForExam(event, exam)));
	return batches.flat();
}

export const GET = async (event: RequestEvent) => {
	try {
		const params = event.url.searchParams;
		const examId = params.get('exam_id')?.trim();
		let rows: SesiCbtRow[];

		if (examId) {
			const data = await proxy(event).get(apiPath`/api/cbt/events/${examId}/sessions`);
			rows = asArray(data).map((item) => normalizeSesiCbtRow(item, { id: examId }));
		} else {
			try {
				const data = await proxy(event).get(apiPathWithQuery('/api/asesmen/sessions', params));
				rows = asArray(data).map((item) => normalizeSesiCbtRow(item));
			} catch (error) {
				if (!(error instanceof ApiError) || (error.status !== 404 && error.status !== 405)) throw error;
				rows = await fallbackSessions(event);
			}
		}

		const filtered = rows.filter((row) => matchesFilters(row, params));
		const offset = queryNumber(params, 'offset', 0);
		const limit = queryNumber(params, 'limit', 100);
		const items = filtered
			.sort((a, b) => String(a.scheduled_start ?? '').localeCompare(String(b.scheduled_start ?? '')))
			.slice(offset, limit > 0 ? offset + limit : undefined);

		return json({
			items,
			summary: summarizeSessionRows(filtered),
			generated_at: new Date().toISOString()
		});
	} catch (e) {
		return handleRouteError(e, 'asesmen/sessions GET');
	}
};
