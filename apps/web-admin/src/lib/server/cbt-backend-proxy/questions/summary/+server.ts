import { cbtBackendPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { ApiError, apiPathWithQuery, handleRouteError, proxy } from '$lib/server/api';

type QuestionListPayload = {
	items?: Array<Record<string, unknown>>;
	meta?: { total?: number };
};

type SummaryCounts = {
	all: number;
	total: number;
	unpublished: number;
	draft: number;
	review: number;
	rejected: number;
	approved: number;
	published: number;
	package_usage: number;
};

function text(value: unknown): string {
	return typeof value === 'string' ? value.trim() : '';
}

function numberValue(value: unknown): number {
	return typeof value === 'number' && Number.isFinite(value) ? value : 0;
}

function shouldFallbackToList(error: unknown): boolean {
	return error instanceof ApiError
		&& error.status === 400
		&& /invalid id|id tidak valid/i.test(`${error.message} ${error.upstreamMessage}`);
}

function aggregateListFallback(payload: QuestionListPayload) {
	const items = payload.items ?? [];
	const counts: SummaryCounts = {
		all: payload.meta?.total ?? items.length,
		total: payload.meta?.total ?? items.length,
		unpublished: 0,
		draft: 0,
		review: 0,
		rejected: 0,
		approved: 0,
		published: 0,
		package_usage: 0
	};
	const subjects = new Map<string, { subject_id: string; subject_name: string; total: number; published: number }>();
	const cognitive = new Map<string, { cognitive_level: string; total: number }>();
	const recent = items.slice(0, 8).map((item) => ({
		id: text(item.id),
		code: text(item.code),
		stem: text(item.stem_html) || text(item.question_text),
		workflow_status: text(item.workflow_status),
		status: text(item.status),
		updated_at: text(item.updated_at) || text(item.created_at),
		author_username: text(item.author_username)
	}));

	for (const item of items) {
		const workflow = text(item.workflow_status);
		const status = text(item.status);
		if (status === 'draft') counts.unpublished += 1;
		if (workflow === 'submitted' || workflow === 'review') counts.review += 1;
		else if (workflow === 'revision_needed' || workflow === 'rejected' || workflow === 'revision') counts.rejected += 1;
		else if (workflow === 'approved') counts.approved += 1;
		else if (!workflow || workflow === 'draft') counts.draft += 1;
		if (status === 'published') counts.published += 1;
		const packageCount = numberValue(item.package_count) || numberValue((item.usage as Record<string, unknown> | undefined)?.package_count);
		if (packageCount > 0) counts.package_usage += 1;

		const subjectId = text(item.subject_id) || text((item.subject as Record<string, unknown> | undefined)?.id) || text(item.subject_name) || 'tanpa-mapel';
		const subjectName = text(item.subject_name) || text((item.subject as Record<string, unknown> | undefined)?.name) || 'Tanpa mapel';
		const subject = subjects.get(subjectId) ?? { subject_id: subjectId, subject_name: subjectName, total: 0, published: 0 };
		subject.total += 1;
		if (status === 'published' || workflow === 'approved') subject.published += 1;
		subjects.set(subjectId, subject);

		const level = text(item.cognitive_level) || 'Belum diisi';
		const bucket = cognitive.get(level) ?? { cognitive_level: level, total: 0 };
		bucket.total += 1;
		cognitive.set(level, bucket);
	}

	return {
		counts,
		by_subject: Array.from(subjects.values()),
		by_cognitive_level: Array.from(cognitive.values()),
		recent
	};
}

async function fallbackSummaryFromList(event: RequestEvent) {
	const params = new URLSearchParams({ limit: '500', page: '1' });
	const payload = await proxy(event).get<QuestionListPayload>(apiPathWithQuery(cbtBackendPath('/questions'), params));
	return aggregateListFallback(payload);
}

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get(cbtBackendPath('/questions/summary'));
		return json(data);
	} catch (e) {
		if (shouldFallbackToList(e)) {
			try {
				return json(await fallbackSummaryFromList(event));
			} catch (fallbackError) {
				return handleRouteError(fallbackError, 'cbt/questions/summary fallback GET');
			}
		}
		return handleRouteError(e, 'cbt/questions/summary GET');
	}
};
