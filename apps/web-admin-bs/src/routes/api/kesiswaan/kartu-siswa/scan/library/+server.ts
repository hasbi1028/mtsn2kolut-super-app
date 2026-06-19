import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, readRequestJson } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).post('/api/student-id-cards/scan/library', body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'kesiswaan/kartu-siswa scan library');
	}
};
