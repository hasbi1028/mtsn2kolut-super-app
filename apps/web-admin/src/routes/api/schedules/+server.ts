import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

interface GoSchedule {
	id: string; label: string; run_type: string;
	run_time: string; is_enabled: boolean;
}

export const GET = async (event: RequestEvent) => {
	try {
		const items = await proxy(event).get<GoSchedule[]>('/api/schedules');
		return json({ items });
	} catch (e) {
		return handleRouteError(e, 'schedules GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await event.request.json();
		const sched = await proxy(event).post<GoSchedule>('/api/schedules', body);
		return json(sched, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'schedules POST');
	}
};

export const PUT = async (event: RequestEvent) => {
	try {
		const { schedules } = await event.request.json() as { schedules?: GoSchedule[] };
		if (!Array.isArray(schedules)) return json({ error: 'schedules harus array' }, { status: 400 });

		const p = proxy(event);
		await Promise.all(schedules.map((s) =>
			p.put(`/api/schedules/${s.id}`, {
				label:      s.label,
				run_time:   s.run_time,
				is_enabled: s.is_enabled,
			})
		));
		return json({ ok: true });
	} catch (e) {
		return handleRouteError(e, 'schedules PUT');
	}
};
