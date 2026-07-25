import type { PageServerLoad } from './$types.js';
import { env } from '$env/dynamic/private';

const API_BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');

export const load: PageServerLoad = async ({ fetch, locals }) => {
	const accessToken = locals.accessToken as string | undefined;
	const headers: Record<string, string> = { 'Content-Type': 'application/json' };
	if (accessToken) headers['Authorization'] = `Bearer ${accessToken}`;

	const [muridRes, rombelRes] = await Promise.all([
		fetch(`${API_BASE}/api/kesiswaan/murid`, { headers }),
		fetch(`${API_BASE}/api/academic/rombels`, { headers }),
	]);

	let murid: any[] = [];
	if (muridRes.ok) {
		const payload = await muridRes.json();
		murid = payload.data ?? [];
	}

	let rombels: any[] = [];
	if (rombelRes.ok) {
		const payload = await rombelRes.json();
		rombels = payload.data ?? [];
	}

	return { murid, rombels };
};
