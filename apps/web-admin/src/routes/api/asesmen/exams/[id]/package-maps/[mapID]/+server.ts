import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, requiredRouteParam } from '$lib/server/api';

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const mapID = requiredRouteParam(event.params.mapID, 'mapID');
		const data = await proxy(event).del(apiPath`/api/asesmen/exams/${id}/package-maps/${mapID}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'asesmen/exams/[id]/package-maps/[mapID] DELETE');
	}
};
