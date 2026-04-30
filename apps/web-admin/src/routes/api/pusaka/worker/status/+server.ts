import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { requireAuthHeaders, handleRouteError } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const accessToken = event.locals.accessToken ?? event.cookies.get('access_token');
		const res = await fetch(`${event.url.origin}/health`, {
			headers: requireAuthHeaders(accessToken),
		});
		if (!res.ok) throw new Error(`health check HTTP ${res.status}`);
		const data = await res.json();
		const workersRaw: Record<string, string> = data.workers ?? {};
		const active_workers = Object.values(workersRaw)
			.map((v) => { try { return JSON.parse(v); } catch { return null; } })
			.filter(Boolean);
		return json({
			data: {
				active_workers,
				total: active_workers.length,
				queue: data.queue ?? {},
				last_checked: new Date().toISOString(),
			},
		});
	} catch (e) {
		return handleRouteError(e, 'pusaka/worker/status');
	}
};
