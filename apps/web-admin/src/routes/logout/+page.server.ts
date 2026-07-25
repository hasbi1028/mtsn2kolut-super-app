import { redirect } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';

const API_BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');

async function doLogout(event: RequestEvent) {
	const { fetch, cookies } = event;
	const refreshToken = cookies.get('refresh_token') || '';
	try {
		await fetch(`${API_BASE}/api/auth/logout`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ refresh_token: refreshToken }),
		});
	} catch {
		// Proceed with local cleanup even if backend call fails
	}

	// Clear all auth cookies
	cookies.delete('access_token', { path: '/' });
	cookies.delete('refresh_token', { path: '/' });
	cookies.delete('session_id', { path: '/' });
}

export async function load(event: RequestEvent) {
	await doLogout(event);
	redirect(302, '/');
}

export const actions = {
	default: async (event: RequestEvent) => {
		await doLogout(event);
		redirect(302, '/');
	},
};
