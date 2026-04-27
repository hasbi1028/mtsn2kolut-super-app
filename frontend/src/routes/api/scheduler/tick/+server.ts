import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { schedulerTick } from '$lib/server/scheduler';

export const POST: RequestHandler = async () => json({ enqueued: await schedulerTick() });
