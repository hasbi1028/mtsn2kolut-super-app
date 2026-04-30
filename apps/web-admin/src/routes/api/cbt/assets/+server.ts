import { env } from '$env/dynamic/private';
import type { RequestEvent } from '@sveltejs/kit';
import { json } from '@sveltejs/kit';
import { handleRouteError, requireAuthHeaders } from '$lib/server/api';

const BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');

export const GET = async (event: RequestEvent) => {
	try {
		const questionId = event.url.searchParams.get('question_id');
		if (!questionId) return json({ error: 'question_id wajib diisi' }, { status: 400 });
		const accessToken = event.locals.accessToken ?? event.cookies.get('access_token');
		const res = await fetch(`${BASE}/api/cbt/assets?question_id=${encodeURIComponent(questionId)}`, {
			headers: requireAuthHeaders(accessToken),
		});
		const data = await res.json().catch(() => ({}));
		if (!res.ok) return json({ error: data.error ?? `HTTP ${res.status}` }, { status: res.status });
		return json(data.data ?? data);
	} catch (e) {
		return handleRouteError(e, 'cbt/assets GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const form = await event.request.formData();
		const accessToken = event.locals.accessToken ?? event.cookies.get('access_token');
		const res = await fetch(`${BASE}/api/cbt/assets`, {
			method: 'POST',
			headers: requireAuthHeaders(accessToken),
			body: form,
		});
		const data = await res.json().catch(() => ({}));
		if (!res.ok) return json({ error: data.error ?? `HTTP ${res.status}` }, { status: res.status });
		return json(data.data ?? data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'cbt/assets POST');
	}
};
