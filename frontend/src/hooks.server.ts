import type { Handle } from '@sveltejs/kit';
import { initScheduler } from '$lib/server/scheduler';

initScheduler();

export const handle: Handle = async ({ event, resolve }) => {
	return resolve(event);
};
