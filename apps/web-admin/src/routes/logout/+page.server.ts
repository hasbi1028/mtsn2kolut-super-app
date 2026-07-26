import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	// Server-side: clear cookies and redirect
	const refreshToken = cookies.get('refresh_token') ?? '';
	if (refreshToken) {
		try {
			await fetch('/api/auth/logout', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ refresh_token: refreshToken }),
			});
		} catch { /* best effort */ }
	}
	cookies.delete('access_token', { path: '/' });
	cookies.delete('refresh_token', { path: '/' });
	throw redirect(302, '/login');
};
