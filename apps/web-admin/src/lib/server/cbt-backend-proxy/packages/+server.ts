import { cbtApiPath, cbtBackendPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

function queryPath(path: string, params: URLSearchParams) {
	const query = params.toString();
	return query ? `${path}?${query}` : path;
}

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get(queryPath(cbtBackendPath('/packages'), event.url.searchParams));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/packages GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const { subject_id, title, description, duration_minutes, randomize_questions, is_active, question_ids, question_weights, event_id } = body;
		if (!subject_id || !title || !duration_minutes) {
			return json({ error: 'subject_id, title, duration_minutes wajib diisi' }, { status: 400 });
		}
		const data = await proxy(event).post(cbtBackendPath('/packages'), {
			subject_id, title, description: description ?? '',
			duration_minutes: Number(duration_minutes),
			randomize_questions: randomize_questions ?? false,
			is_active: is_active ?? true,
			question_ids: question_ids ?? [],
			question_weights: question_weights ?? {},
			...(typeof event_id === 'string' && event_id.trim() ? { event_id } : {}),
		});
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'cbt/packages POST');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.url.searchParams.get('id') ?? undefined, 'id');
		await proxy(event).del(cbtApiPath`/packages/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'cbt/packages DELETE');
	}
};
