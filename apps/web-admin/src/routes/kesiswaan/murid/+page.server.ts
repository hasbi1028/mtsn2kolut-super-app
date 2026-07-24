import type { PageServerLoad } from './$types.js';
import { env } from '$env/dynamic/private';

const API_BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');

export const load: PageServerLoad = async ({ fetch, locals }) => {
	const accessToken = locals.accessToken as string | undefined;
	const headers: Record<string, string> = { 'Content-Type': 'application/json' };
	if (accessToken) headers['Authorization'] = `Bearer ${accessToken}`;

	const res = await fetch(`${API_BASE}/api/kesiswaan/murid`, { headers });

	let murid: any[] = [];
	if (res.ok) {
		const payload = await res.json();
		murid = payload.data ?? [];
	}

	return { murid };
};
