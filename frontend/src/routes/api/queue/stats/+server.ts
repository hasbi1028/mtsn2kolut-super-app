import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiGet, handleRouteError } from '$lib/server/api';

interface GoStats { queued: number; running: number; success: number; failed: number }

export const GET: RequestHandler = async () => {
	try {
		const s = await apiGet<GoStats>('/api/jobs/stats');
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
