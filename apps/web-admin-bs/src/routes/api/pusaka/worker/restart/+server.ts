import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { hasAnyRole } from '$lib/server/route-access';

export const POST: RequestHandler = async (event) => {
	if (!hasAnyRole(event.locals.user, ['admin'])) {
		return json({ error: 'forbidden: admin role required' }, { status: 403 });
	}

	return json(
		{
			error:
				'Restart worker tidak dijalankan dari web-admin. Gunakan runbook deploy/PM2 pada VPS worker.'
		},
		{ status: 410 }
	);
};
