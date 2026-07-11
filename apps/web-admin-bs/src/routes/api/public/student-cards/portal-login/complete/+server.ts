import { dev } from '$app/environment';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, readRequestJson } from '$lib/server/api';

type PortalLoginResponse = {
	access_token?: string;
	refresh_token?: string;
	must_change_password?: boolean;
	message?: string;
	student_id?: string;
	ok?: boolean;
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).post<PortalLoginResponse>('/api/public/student-cards/portal-login/complete', body);
		if (data.access_token && data.refresh_token) {
			event.cookies.set('access_token', data.access_token, {
				path: '/', httpOnly: true, sameSite: 'lax', secure: !dev, maxAge: 60 * 60,
			});
			event.cookies.set('refresh_token', data.refresh_token, {
				path: '/', httpOnly: true, sameSite: 'lax', secure: !dev, maxAge: 7 * 24 * 60 * 60,
			});
		}
		return json({ ok: data.ok === true, message: data.message, student_id: data.student_id, must_change_password: data.must_change_password });
	} catch (e) {
		return handleRouteError(e, 'public student card portal-login complete');
	}
};
