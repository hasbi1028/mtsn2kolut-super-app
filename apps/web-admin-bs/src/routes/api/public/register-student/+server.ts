import type { RequestEvent } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import { handleRouteError, jsonProxyResponse, readRequestJson } from '$lib/server/api';
import { optionalStringField, requireStringField, validationErrorResponse, ValidationError } from '$lib/server/validation';

const BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');

export const POST = async (event: RequestEvent) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const gender = requireStringField(body, 'gender');
		if (!['L', 'P'].includes(gender)) {
			throw new ValidationError('gender tidak valid');
		}
		const payload = {
			nis: requireStringField(body, 'nis'),
			nama: requireStringField(body, 'nama'),
			gender,
			parent_name: optionalStringField(body, 'parent_name'),
			parent_phone: optionalStringField(body, 'parent_phone'),
		};
		const clientAddress = event.getClientAddress();
		const headers: Record<string, string> = {
			'Content-Type': 'application/json',
			'X-Client-IP': clientAddress,
			'X-Forwarded-For': clientAddress
		};
		const userAgent = event.request.headers.get('user-agent');
		if (userAgent) headers['X-Client-User-Agent'] = userAgent;
		const res = await event.fetch(`${BASE}/api/public/register-student`, {
			method: 'POST',
			headers,
			body: JSON.stringify(payload),
		});
		return await jsonProxyResponse<Record<string, unknown>>(res, { status: 201 });
	} catch (e) {
		const validation = validationErrorResponse(e);
		if (validation) return validation;
		return handleRouteError(e, 'public/register-student POST');
	}
};
