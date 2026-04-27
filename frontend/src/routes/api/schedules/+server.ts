import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiGet, apiPut, handleRouteError } from '$lib/server/api';

interface GoSchedule {
	id: string; label: string; run_type: string;
	run_time: string; is_enabled: boolean;
}

export const GET: RequestHandler = async () => {
	try {
		const items = await apiGet<GoSchedule[]>('/api/schedules');
		return json({ items });
	} catch (e) {
		return handleRouteError(e, 'schedules GET');
	}
};

export const PUT: RequestHandler = async ({ request }) => {
	try {
		const { schedules } = await request.json() as { schedules?: GoSchedule[] };
		if (!Array.isArray(schedules)) return json({ error: 'schedules harus array' }, { status: 400 });

		await Promise.all(schedules.map((s) =>
			apiPut(`/api/schedules/${s.id}`, {
				label:      s.label,
				run_time:   s.run_time,
				run_type:   s.run_type,
				is_enabled: s.is_enabled,
			})
		));
		return json({ ok: true });
	} catch (e) {
		return handleRouteError(e, 'schedules PUT');
	}
};
