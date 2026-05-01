import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPathWithQuery, handleRouteError, proxy, readProxyJson, toWITA } from '$lib/server/api';

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

export const GET = async (event: RequestEvent) => {
	try {
		const limit  = Math.min(500, Math.max(1, Number(event.url.searchParams.get('limit') ?? 50)));
		const status = event.url.searchParams.get('status') ?? '';
		const params = new URLSearchParams({ per_page: String(limit), page: '1' });
		if (status) params.set('status', status);

		const raw = await proxy(event).fetch(apiPathWithQuery('/api/pusaka/jobs', params));
		const res = await readProxyJson<GoJobsResponse>(raw);
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
		return handleRouteError(e, 'pusaka/jobs GET');
	}
};
