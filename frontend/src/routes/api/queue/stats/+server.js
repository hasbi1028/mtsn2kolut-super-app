import { json } from '@sveltejs/kit';
import { getQueueStats } from '$lib/server/queue';

export function GET() {
  return json(getQueueStats());
}
