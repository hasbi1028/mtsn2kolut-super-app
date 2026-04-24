import { json } from '@sveltejs/kit';
import { db } from '$lib/server/db';

export async function POST({ request }) {
  const { employee_id } = await request.json();
  if (!employee_id) return json({ error: 'employee_id wajib' }, { status: 400 });

  const result = db.prepare(
    `UPDATE jobs
     SET status = 'failed',
         error_message = 'Dibatalkan manual',
         next_retry_at = NULL,
         updated_at = CURRENT_TIMESTAMP
     WHERE employee_id = ? AND status IN ('queued', 'running')`
  ).run(employee_id);

  return json({ ok: true, cancelled: result.changes });
}
