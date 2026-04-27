import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPost } from '$lib/server/api';

export const POST: RequestHandler = async ({ request }) => {
	const payload = await request.json().catch(() => ({})) as Record<string, unknown>;
	const run_type    = String(payload?.run_type ?? 'morning');
	const max_attempts = Math.max(1, Math.min(Number(payload?.max_attempts ?? 3), 10));

	const result = await apiPost<{ inserted: number; skipped: number }>(
		'/api/jobs/run-all',
		{ run_type, max_attempts }
	);
	return json({ run_type, ...result }, { status: 201 });
};
