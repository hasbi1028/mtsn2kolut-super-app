import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { enqueueAllActive } from '$lib/server/queue';
import { logInfo } from '$lib/server/logger';
import type { RunType } from '$lib/server/schema';

export const POST: RequestHandler = async ({ request }) => {
	const payload = await request.json().catch(() => ({})) as Record<string, unknown>;
	const run_type: RunType     = payload?.run_type === 'afternoon' ? 'afternoon'
		: payload?.run_type === 'checkin' ? 'checkin'
		: payload?.run_type === 'checkout' ? 'checkout'
		: 'morning';
	const max_attempts = Math.max(1, Math.min(Number(payload?.max_attempts ?? 3), 10));

	const result = enqueueAllActive(run_type, max_attempts);
	logInfo('run-all enqueued', { run_type, ...result });

	return json({ run_type, ...result }, { status: 201 });
};
