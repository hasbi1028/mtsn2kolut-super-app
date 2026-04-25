import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { getQueueStats } from '$lib/server/queue';

export const GET: RequestHandler = () => json(getQueueStats());
