import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const DELETE = async (event: RequestEvent) => {
	try {
		const studentId = event.params.studentId;
		await proxy(event).del(`/api/academic/rombels/${event.params.id}/students/${studentId}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'rombel student DELETE');
	}
};
