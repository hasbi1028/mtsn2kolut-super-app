import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { schedulerTick } from '$lib/server/scheduler';
import { handleRouteError } from '$lib/server/api';

export const POST: RequestHandler = async () => {
	try {
		return json({ enqueued: await schedulerTick() });
	} catch (e) {
		return handleRouteError(e, 'scheduler/tick');
	}
};
