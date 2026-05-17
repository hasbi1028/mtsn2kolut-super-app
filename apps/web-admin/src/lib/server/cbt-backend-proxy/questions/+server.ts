import { cbtApiPath, cbtBackendPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import {
	ApiError,
	apiPath,
	apiPathWithQuery,
	handleRouteError,
	proxy,
	readRequestJson,
	requiredRouteParam
} from '$lib/server/api';
import { log } from '$lib/server/logger';
import { getOrCreateRequestId } from '$lib/server/request-id';

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get(apiPathWithQuery(cbtBackendPath('/questions'), event.url.searchParams));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/questions GET');
	}
};

export const POST = async (event: RequestEvent) => {
	const request_id = getOrCreateRequestId(event.request);
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const { subject_id } = body;
		if (!subject_id) {
			log({
				level: 'warn',
				event: 'bank_soal_questions_proxy_validation_failed',
				module: 'bank-soal',
				request_id,
				route: 'POST /api/bank-soal/questions',
				method: 'POST',
				status: 400,
				metadata: questionPayloadSummary(body)
			});
			return json({ error: 'subject_id wajib diisi' }, { status: 400, headers: { 'x-request-id': request_id } });
		}
		const data = await proxy(event).post(cbtBackendPath('/questions'), body);
		log({
			level: 'info',
			event: 'bank_soal_questions_proxy_success',
			module: 'bank-soal',
			request_id,
			route: 'POST /api/bank-soal/questions',
			method: 'POST',
			status: 201,
			metadata: questionPayloadSummary(body)
		});
		return json(data, { status: 201, headers: { 'x-request-id': request_id } });
	} catch (e) {
		if (e instanceof ApiError) {
			const apiError = e as ApiError;
			log({
				level: apiError.status >= 500 ? 'error' : 'warn',
				event: 'bank_soal_questions_upstream_error',
				module: 'bank-soal',
				request_id,
				route: 'POST /api/bank-soal/questions',
				method: 'POST',
				status: apiError.status,
				error: apiError.upstreamMessage
			});
		}
		const response = handleRouteError(e, 'cbt/questions POST');
		response.headers.set('x-request-id', request_id);
		return response;
	}
};

function questionPayloadSummary(body: Record<string, unknown>): Record<string, unknown> {
	const options = Array.isArray(body.options) ? body.options : [];
	return {
		has_subject_id: Boolean(body.subject_id),
		has_event_id: Boolean(body.event_id),
		subject_id: typeof body.subject_id === 'string' ? body.subject_id : undefined,
		event_id: typeof body.event_id === 'string' ? body.event_id : undefined,
		question_type: body.question_type,
		target_level: body.target_level,
		grade_level: body.grade_level,
		difficulty: body.difficulty,
		workflow_status: body.workflow_status,
		authoring_mode: body.authoring_mode,
		options_count: options.length,
		stem_length: typeof body.stem_html === 'string' ? body.stem_html.length : 0,
		question_text_length: typeof body.question_text === 'string' ? body.question_text.length : 0
	};
}

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.url.searchParams.get('id') ?? undefined, 'id');
		await proxy(event).del(cbtApiPath`/questions/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'cbt/questions DELETE');
	}
};
