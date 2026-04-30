import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

interface GoStats { queued: number; running: number; success: number; failed: number }

export const GET = async (event: RequestEvent) => {
	try {
		const s = await proxy(event).get<GoStats>('/api/pusaka/jobs/stats');
		return json({
			queued:    s.queued,
			running:   s.running,
			success:   s.success,
			failed:    s.failed,
			retry_due: 0,
			total:     s.queued + s.running + s.success + s.failed,
		});
	} catch (e) {
		return handleRouteError(e, 'queue/stats');
	}
};
