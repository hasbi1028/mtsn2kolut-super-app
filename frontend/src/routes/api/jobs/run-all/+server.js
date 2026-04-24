import { json } from '@sveltejs/kit';
import { enqueueAllActive } from '$lib/server/queue';
import { logInfo } from '$lib/server/logger';

export async function POST({ request }) {
  const payload = await request.json().catch(() => ({}));
  const run_type = payload?.run_type === 'afternoon' ? 'afternoon' : 'morning';
  const max_attempts = Number(payload?.max_attempts || 3);

  const result = enqueueAllActive(run_type, Math.max(1, Math.min(max_attempts, 10)));
  logInfo('run-all enqueued', { run_type, ...result });

  return json({ run_type, ...result }, { status: 201 });
}
