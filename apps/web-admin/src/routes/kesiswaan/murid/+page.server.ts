import type { PageServerLoad } from './$types.js';
import { env } from '$env/dynamic/private';

const API_BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');

export const load: PageServerLoad = async ({ fetch, locals, url }) => {
	const accessToken = locals.accessToken as string | undefined;
	const headers: Record<string, string> = { 'Content-Type': 'application/json' };
	if (accessToken) headers['Authorization'] = `Bearer ${accessToken}`;

	const page = url.searchParams.get('page') || '1';
	const perPage = url.searchParams.get('per_page') || '25';

	const [muridRes, rombelRes] = await Promise.all([
		fetch(`${API_BASE}/api/kesiswaan/murid?page=${page}&per_page=${perPage}`, { headers }),
		fetch(`${API_BASE}/api/academic/rombels`, { headers }),
	]);

	let murid: any[] = [];
	let total = 0;
	let pages = 0;
	if (muridRes.ok) {
		const payload = await muridRes.json();
		murid = payload.data ?? [];
		total = payload.total ?? 0;
		pages = payload.pages ?? 0;
	}

	let rombels: any[] = [];
	if (rombelRes.ok) {
		const payload = await rombelRes.json();
		rombels = payload.data ?? [];
	}

	return { murid, rombels, total, pages, page: parseInt(page), perPage: parseInt(perPage) };
};
