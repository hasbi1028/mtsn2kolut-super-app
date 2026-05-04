import { dev } from '$app/environment';
import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { ApiError, apiLoginWithFetch } from '$lib/server/api';
import { safeSameOriginRedirectPath } from '$lib/server/redirects';

export const load: PageServerLoad = async ({ locals, url }) => {
	if (locals.user) throw redirect(302, safeSameOriginRedirectPath(url.searchParams.get('from')));
	return {};
};

export const actions: Actions = {
	default: async ({ request, cookies, url, getClientAddress, fetch }) => {
		const data = await request.formData();
		const username = String(data.get('username') ?? '').trim();
		const password = String(data.get('password') ?? '');

		try {
			const pair = await apiLoginWithFetch(fetch, username, password, {
				userAgent: request.headers.get('user-agent') ?? '',
				ipAddress: getClientAddress(),
			});
			cookies.set('access_token', pair.access_token, {
				path: '/', httpOnly: true, sameSite: 'lax',
				secure: !dev,
				maxAge: 60 * 60,
			});
			cookies.set('refresh_token', pair.refresh_token, {
				path: '/', httpOnly: true, sameSite: 'lax',
				secure: !dev,
				maxAge: 7 * 24 * 60 * 60,
			});
		} catch (e) {
			if (e instanceof ApiError && e.status === 401) {
				return fail(401, { error: 'Username atau password salah' });
			}
			return fail(500, { error: 'Server error, coba lagi' });
		}


		throw redirect(302, safeSameOriginRedirectPath(url.searchParams.get('from')));
	},
};
