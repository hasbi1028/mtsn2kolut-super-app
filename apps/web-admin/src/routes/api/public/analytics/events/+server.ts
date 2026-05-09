import { json } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import type { RequestHandler } from './$types';
import { readRequestJson } from '$lib/server/api';
import { publicAnalyticsRateAllowed, sanitizePublicAnalyticsCollectorPayload } from '$lib/server/public-analytics';

const BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');
const PUBLIC_ANALYTICS_BODY_LIMIT = 4 * 1024;

export const POST: RequestHandler = async (event) => {
	if (!publicAnalyticsRateAllowed(event.getClientAddress())) {
		return json({ error: 'rate limited' }, { status: 429 });
	}
	const internalKey = env.INTERNAL_API_KEY?.trim();
	if (!internalKey) {
		return json({ error: 'collector unavailable' }, { status: 503 });
	}

	let body: unknown;
	try {
		body = await readRequestJson<unknown>(event.request, PUBLIC_ANALYTICS_BODY_LIMIT);
	} catch {
		return json({ error: 'invalid payload' }, { status: 400 });
	}
	const payload = sanitizePublicAnalyticsCollectorPayload(body);
	if (!payload) {
		return json({ error: 'invalid payload' }, { status: 400 });
	}

	try {
		const response = await event.fetch(`${BASE}/api/internal-analytics/public-events`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				'X-Internal-Key': internalKey
			},
			body: JSON.stringify(payload)
		});
		if (!response.ok) {
			return json({ error: 'collector unavailable' }, { status: response.status >= 500 ? 503 : response.status });
		}
		return new Response(null, { status: 204 });
	} catch {
		return json({ error: 'collector unavailable' }, { status: 503 });
	}
};
