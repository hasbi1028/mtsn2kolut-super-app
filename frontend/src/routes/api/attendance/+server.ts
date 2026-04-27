import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { toWITA } from '$lib/server/api';

const BASE = (process.env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');
const INTERNAL_KEY = process.env.INTERNAL_API_KEY ?? '';

interface GoAttendance {
	id: string; employee_id: string; tanggal: string;
	jam_masuk: string; jam_pulang: string; source_job_id: string;
	updated_at: string; created_at: string;
	employee_nip: string; employee_nama: string;
}
interface GoAttResponse { data: GoAttendance[]; meta: { total: number } }

export const GET: RequestHandler = async ({ url }) => {
	const limit = Math.min(500, Math.max(1, Number(url.searchParams.get('limit') ?? 50)));
	const date  = url.searchParams.get('date') ?? '';

	const isDate = /^\d{4}-\d{2}-\d{2}$/.test(date);
	const endpoint = isDate
		? `${BASE}/api/attendance/by-date/${date}`
		: `${BASE}/api/attendance?per_page=${limit}&page=1`;

	const raw = await fetch(endpoint, { headers: { 'X-Internal-Key': INTERNAL_KEY } });
	const res = await raw.json() as GoAttResponse | { data: GoAttendance[] };

	const rows = (res.data ?? []) as GoAttendance[];
	const items = rows.map(({ employee_nip, employee_nama, ...a }) => ({
		...a,
		nip:             employee_nip,
		nama:            employee_nama,
		updated_at_wita: toWITA(a.updated_at),
	}));

	return json({ items });
};
