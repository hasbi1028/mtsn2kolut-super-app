import { json } from '@sveltejs/kit';
import { db } from '$lib/server/db';

export function POST() {
  const result = db.prepare(
    `UPDATE jobs
     SET status = 'failed',
         error_message = 'Dibatalkan manual',
         next_retry_at = NULL,
         updated_at = CURRENT_TIMESTAMP
     WHERE status IN ('queued', 'running')`
  ).run();

  return json({ ok: true, cancelled: result.changes });
}
