import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const payload = await event.request.json().catch(() => ({})) as Record<string, unknown>;
		const run_type    = String(payload?.run_type ?? 'morning');
		const max_attempts = Math.max(1, Math.min(Number(payload?.max_attempts ?? 3), 10));

		const result = await proxy(event).post<{ inserted: number; skipped: number }>(
			'/api/jobs/run-all',
			{ run_type, max_attempts }
		);
		return json({ run_type, ...result }, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'jobs/run-all');
	}
};
