import { cbtApiPath, cbtBackendPath } from '$lib/server/cbt-backend-paths';
import type { RequestEvent } from '@sveltejs/kit';
import { json } from '@sveltejs/kit';
import { apiPathWithQuery, handleRouteError, jsonProxyResponse, proxy } from '$lib/server/api';

type AssetPayload = { url?: string; [key: string]: unknown };

function normalizeBankSoalAssetUrl(url: string) {
	return url.replace(/^\/api\/cbt\/assets\//, '/api/bank-soal/assets/');
}

function normalizeAssetPayload(data: unknown) {
	if (!data || typeof data !== 'object') return data;
	const payload = data as AssetPayload;
	if (typeof payload.url !== 'string') return data;
	return { ...payload, url: normalizeBankSoalAssetUrl(payload.url) };
}

export const GET = async (event: RequestEvent) => {
	try {
		const questionId = event.url.searchParams.get('question_id');
		if (!questionId) return json({ error: 'question_id wajib diisi' }, { status: 400 });
		const res = await proxy(event).fetch(apiPathWithQuery(cbtBackendPath('/assets'), new URLSearchParams({ question_id: questionId })));
		return await jsonProxyResponse<{ data?: unknown }, unknown>(res, {
			map: (data) => data.data ?? data
		});
	} catch (e) {
		return handleRouteError(e, 'cbt/assets GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const form = await event.request.formData();
		const res = await proxy(event).fetch(cbtBackendPath('/assets'), {
			method: 'POST',
			body: form,
		});
		return await jsonProxyResponse<{ data?: unknown }, unknown>(res, {
			status: 201,
			map: (data) => normalizeAssetPayload(data.data ?? data)
		});
	} catch (e) {
		return handleRouteError(e, 'cbt/assets POST');
	}
};
