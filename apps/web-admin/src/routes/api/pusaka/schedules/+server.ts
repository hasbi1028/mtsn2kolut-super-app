import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

interface GoSchedule {
	id: string; label: string; run_type: string;
	run_time: string; is_enabled: boolean;
}

export const GET = async (event: RequestEvent) => {
	try {
		const items = await proxy(event).get<GoSchedule[]>('/api/pusaka/schedules');
		return json({ items });
	} catch (e) {
		return handleRouteError(e, 'pusaka/schedules GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await event.request.json();
		const sched = await proxy(event).post<GoSchedule>('/api/pusaka/schedules', body);
		return json(sched, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'pusaka/schedules POST');
	}
};

export const PUT = async (event: RequestEvent) => {
	try {
		const { schedules } = await event.request.json() as { schedules?: GoSchedule[] };
		if (!Array.isArray(schedules)) return json({ error: 'schedules harus array' }, { status: 400 });
		const p = proxy(event);
		await Promise.all(schedules.map((s) =>
			p.put(`/api/pusaka/schedules/${s.id}`, {
				label: s.label,
				run_time: s.run_time,
				is_enabled: s.is_enabled,
			})
		));
		return json({ ok: true });
	} catch (e) {
		return handleRouteError(e, 'pusaka/schedules PUT');
	}
};
