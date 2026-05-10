import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

type RevealTokenRequest = {
	room_token?: string;
};

export const POST = async (event: RequestEvent) => {
	try {
		const participantID = requiredRouteParam(event.params.participant_id, 'participant_id');
		const body = await readRequestJson<RevealTokenRequest>(event.request, 4 * 1024);
		const data = await proxy(event).post(
			apiPath`/api/portal/student/cbt/${participantID}/reveal-token`,
			{ room_token: body.room_token ?? '' }
		);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'portal/siswa/cbt/reveal-token POST');
	}
};
