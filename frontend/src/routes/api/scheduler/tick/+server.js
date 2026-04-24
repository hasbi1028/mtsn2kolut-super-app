import { json } from '@sveltejs/kit';
import { schedulerTick } from '$lib/server/scheduler';

export function POST() {
  const enqueued = schedulerTick();
  return json({ enqueued });
}
