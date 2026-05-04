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
		const res = await event.fetch(`${BASE}/api/public/register-student`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload),
		});
		return await jsonProxyResponse<Record<string, unknown>>(res, { status: 201 });
	} catch (e) {
		const validation = validationErrorResponse(e);
		if (validation) return validation;
		return handleRouteError(e, 'public/register-student POST');
	}
};
