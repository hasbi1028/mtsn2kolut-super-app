import { redirect } from '@sveltejs/kit';
import type { Handle } from '@sveltejs/kit';
import { initScheduler } from '$lib/server/scheduler';
import { verifySession } from '$lib/server/auth';

initScheduler();

const PUBLIC_PREFIXES = ['/login', '/api/worker/'];

export const handle: Handle = async ({ event, resolve }) => {
	const sid = event.cookies.get('sid');
	if (verifySession(sid)) {
		event.locals.user = { id: 'admin' };
	}

	const isPublic = PUBLIC_PREFIXES.some((p) => event.url.pathname.startsWith(p));
	if (!isPublic && !event.locals.user) {
		const from = encodeURIComponent(event.url.pathname + event.url.search);
		throw redirect(302, `/login?from=${from}`);
	}

	return resolve(event);
};
