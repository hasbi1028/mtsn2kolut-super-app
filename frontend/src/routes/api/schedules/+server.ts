import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { db } from '$lib/server/db';
import { schedules } from '$lib/server/schema';
import { eq, asc } from 'drizzle-orm';
import type { Schedule } from '$lib/server/schema';

export const GET: RequestHandler = () => {
	const items = db.select().from(schedules).orderBy(asc(schedules.run_time)).all();
	return json({ items });
};

export const PUT: RequestHandler = async ({ request }) => {
	const { schedules: rows } = await request.json() as { schedules?: Schedule[] };
	if (!Array.isArray(rows)) return json({ error: 'schedules harus array' }, { status: 400 });

	db.transaction((tx) => {
		for (const s of rows) {
			tx.update(schedules)
				.set({ label: s.label, run_time: s.run_time, run_type: s.run_type, is_enabled: s.is_enabled, updated_at: new Date().toISOString() })
				.where(eq(schedules.id, s.id))
				.run();
		}
	});

	return json({ ok: true });
};
