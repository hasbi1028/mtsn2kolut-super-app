import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { createSessionCookie } from '$lib/server/auth';
import { apiLogin, ApiError } from '$lib/server/api';

export const load: PageServerLoad = async ({ locals, url }) => {
	if (locals.user) throw redirect(302, url.searchParams.get('from') ?? '/');
	return {};
};

export const actions: Actions = {
	default: async ({ request, cookies, url }) => {
		const data     = await request.formData();
		const username = String(data.get('username') ?? '').trim();
		const password = String(data.get('password') ?? '');

		try {
			await apiLogin(username, password);
		} catch (e) {
			if (e instanceof ApiError && e.status === 401) {
				return fail(401, { error: 'Username atau password salah' });
			}
			return fail(500, { error: 'Server error, coba lagi' });
		}

		cookies.set('sid', createSessionCookie(), {
			path:     '/',
			httpOnly: true,
			sameSite: 'lax',
			secure:   process.env.NODE_ENV === 'production',
			maxAge:   12 * 60 * 60,
		});

		throw redirect(302, url.searchParams.get('from') ?? '/');
	},
};
