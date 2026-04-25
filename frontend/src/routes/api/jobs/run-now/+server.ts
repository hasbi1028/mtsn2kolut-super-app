import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { eq, and } from 'drizzle-orm';
import { db } from '$lib/server/db';
import { employees } from '$lib/server/schema';
import { enqueueOne } from '$lib/server/queue';
import { logInfo, logWarn } from '$lib/server/logger';
import type { RunType } from '$lib/server/schema';

export const POST: RequestHandler = async ({ request }) => {
	const { employee_id, run_type = 'morning' } = await request.json() as { employee_id?: string; run_type?: RunType };
	if (!employee_id) return json({ error: 'employee_id wajib' }, { status: 400 });

	const emp = db.select({ id: employees.id, nip: employees.nip, nama: employees.nama })
		.from(employees)
		.where(and(eq(employees.id, employee_id), eq(employees.is_active, 1)))
		.get();
	if (!emp) return json({ error: 'Pegawai tidak ditemukan / nonaktif' }, { status: 404 });

	const result = enqueueOne(employee_id, run_type, 3);
	if (result.inserted === 0) {
		logWarn('run-now skipped duplicate job', { employee_id, run_type });
		return json({ inserted: 0, skipped: 1, reason: 'job queued/running already exists' });
	}

	logInfo('run-now enqueued', { employee_id, nip: emp.nip, nama: emp.nama, run_type });
	return json({ inserted: 1, skipped: 0 }, { status: 201 });
};
