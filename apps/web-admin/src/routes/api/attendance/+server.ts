import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import { toWITA, handleRouteError } from '$lib/server/api';

const BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');
const INTERNAL_KEY = env.INTERNAL_API_KEY ?? '';

interface GoAttendance {
	id: string; employee_id: string; tanggal: string;
	jam_masuk: string; jam_pulang: string; source_job_id: string;
	updated_at: string; created_at: string;
	employee_nip: string; employee_nama: string;
}
interface GoAttResponse { data: GoAttendance[]; meta: { total: number } }

export const GET: RequestHandler = async ({ url }) => {
	try {
		const limit = Math.min(500, Math.max(1, Number(url.searchParams.get('limit') ?? 50)));
		const date  = url.searchParams.get('date') ?? '';

		const isDate = /^\d{4}-\d{2}-\d{2}$/.test(date);
		const endpoint = isDate
			? `${BASE}/api/attendance/by-date/${date}`
			: `${BASE}/api/attendance?per_page=${limit}&page=1`;

		const raw = await fetch(endpoint, { headers: { 'X-Internal-Key': INTERNAL_KEY } });
		if (!raw.ok) {
			const err = await raw.json().catch(() => ({})) as { error?: string };
			throw Object.assign(new Error(err.error ?? `HTTP ${raw.status}`), { status: raw.status });
		}
		const res = await raw.json() as GoAttResponse | { data: GoAttendance[] };

		const rows = (res.data ?? []) as GoAttendance[];
		const items = rows.map((a) => ({
			...a,
			updated_at_wita: toWITA(a.updated_at),
		}));

		return json({ items });
	} catch (e) {
		return handleRouteError(e, 'attendance GET');
	}
};
