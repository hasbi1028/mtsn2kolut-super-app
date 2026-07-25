// BFF proxy for public website analytics events.
// The Go API collector sanitizes and persists first-party analytics.
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';

export const POST = async (event: RequestEvent): Promise<Response> => {
	try {
		const body = await event.request.json();
		const goResp = await event.fetch('http://localhost:8080/api/internal-analytics/public-events', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(body),
		});
		return new Response(null, { status: goResp.ok ? 204 : goResp.status });
	} catch {
		return new Response(null, { status: 204 }); // silent fail — analytics is non-critical
	}
};
