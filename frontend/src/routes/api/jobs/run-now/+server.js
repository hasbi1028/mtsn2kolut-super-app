import { json } from '@sveltejs/kit';
import { db } from '$lib/server/db';
import { enqueueOne } from '$lib/server/queue';
import { logInfo, logWarn } from '$lib/server/logger';

export async function POST({ request }) {
  const { employee_id, run_type = 'morning' } = await request.json();
  if (!employee_id) return json({ error: 'employee_id wajib' }, { status: 400 });

  const employee = db.prepare('SELECT id, nip, nama FROM employees WHERE id = ? AND is_active = 1').get(employee_id);
  if (!employee) return json({ error: 'Pegawai tidak ditemukan / nonaktif' }, { status: 404 });

  const result = enqueueOne(employee_id, run_type, 3);
  if (result.inserted === 0) {
    logWarn('run-now skipped duplicate job', { employee_id, run_type });
    return json({ inserted: 0, skipped: 1, reason: 'job queued/running already exists' }, { status: 200 });
  }

  logInfo('run-now enqueued', { employee_id, nip: employee.nip, nama: employee.nama, run_type });
  return json({ inserted: 1, skipped: 0 }, { status: 201 });
}
