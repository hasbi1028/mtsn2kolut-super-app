import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, ApiError } from '$lib/server/api';
import { env } from '$env/dynamic/private';

const BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');

export const POST = async (event: RequestEvent) => {
	try {
		const accessToken = event.locals.accessToken ?? event.cookies.get('access_token');
		if (!accessToken) throw new ApiError(401, 'unauthorized');

		const formData = await event.request.formData();
		const res = await fetch(`${BASE}/api/website/media`, {
			method: 'POST',
			headers: { Authorization: `Bearer ${accessToken}` },
			body: formData,
		});

		const data = await res.json().catch(() => ({}));
		if (!res.ok) {
			return json({ error: (data as { error?: string }).error ?? 'Gagal upload gambar.' }, { status: res.status });
		}
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'website/media POST');
	}
};
