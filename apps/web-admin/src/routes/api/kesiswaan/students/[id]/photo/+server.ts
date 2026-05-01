import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const POST: RequestHandler = async (event) => {
	try {
		const form = await event.request.formData();
		const response = await proxy(event).fetch(`/api/kesiswaan/students/${event.params.id}/photo`, {
			method: 'POST',
			body: form,
		});
		const payload = await response.json();
		if (!response.ok) {
			return json(payload, { status: response.status });
		}
		return json(payload.data ?? payload);
	} catch (e) {
		return handleRouteError(e, 'kesiswaan/students/[id]/photo POST');
	}
};
