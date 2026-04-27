import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiGet } from '$lib/server/api';

interface GoStats { queued: number; running: number; success: number; failed: number }

export const GET: RequestHandler = async () => {
	const s = await apiGet<GoStats>('/api/jobs/stats');
	return json({
		queued:    s.queued,
		running:   s.running,
		success:   s.success,
		failed:    s.failed,
		retry_due: 0,
		total:     s.queued + s.running + s.success + s.failed,
	});
};
