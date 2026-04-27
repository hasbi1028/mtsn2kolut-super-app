import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { toWITA, handleRouteError } from '$lib/server/api';

interface GoJob {
	id: string; employee_id: string;
	employee_nama: string; employee_nip: string;
	run_type: string; status: string; error_message: string;
	claimed_by: string; claimed_at: string;
	attempts: number; max_attempts: number;
	next_retry_at: string; created_at: string; updated_at: string;
}

interface GoJobsResponse {
	data: GoJob[];
	meta: { total: number; page: number; per_page: number };
}

export const GET: RequestHandler = async ({ url }) => {
	try {
		const limit  = Math.min(500, Math.max(1, Number(url.searchParams.get('limit') ?? 50)));
		const status = url.searchParams.get('status') ?? '';

		const params = new URLSearchParams({ per_page: String(limit), page: '1' });
		if (status) params.set('status', status);

		const raw = await fetch(
			`${process.env.API_BASE_URL ?? 'http://localhost:8080'}/api/jobs?${params}`,
			{ headers: { 'X-Internal-Key': process.env.INTERNAL_API_KEY ?? '' } }
		);
		if (!raw.ok) {
			const err = await raw.json().catch(() => ({})) as { error?: string };
			throw Object.assign(new Error(err.error ?? `HTTP ${raw.status}`), { status: raw.status });
		}
		const res = await raw.json() as GoJobsResponse;

		const items = (res.data ?? []).map((j) => ({
			id: j.id, run_type: j.run_type, status: j.status,
			error_message: j.error_message, attempts: j.attempts,
			max_attempts: j.max_attempts, next_retry_at: j.next_retry_at,
			claimed_by: j.claimed_by, created_at: j.created_at,
			created_at_wita: toWITA(j.created_at),
			nip: j.employee_nip, nama: j.employee_nama,
		}));

		return json({ items });
	} catch (e) {
		return handleRouteError(e, 'jobs GET');
	}
};
