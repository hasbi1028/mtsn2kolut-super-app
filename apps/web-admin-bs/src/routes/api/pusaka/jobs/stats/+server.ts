import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

interface GoStats {
	queued: number;
	running: number;
	success: number;
	failed: number;
}

export const GET = async (event: RequestEvent) => {
	try {
		const stats = await proxy(event).get<GoStats>('/api/pusaka/jobs/stats');
		return json({
			queued: stats.queued,
			running: stats.running,
			success: stats.success,
			failed: stats.failed,
			retry_due: 0,
			total: stats.queued + stats.running + stats.success + stats.failed
		});
	} catch (e) {
		return handleRouteError(e, 'pusaka/jobs/stats');
	}
};
