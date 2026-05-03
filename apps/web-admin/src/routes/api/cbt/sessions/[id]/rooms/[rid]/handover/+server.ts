import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

type RoomHandoverPayload = {
	attendance_checked?: boolean;
	all_submitted_checked?: boolean;
	device_issue_checked?: boolean;
	room_clean_checked?: boolean;
	token_returned_checked?: boolean;
	assets_returned_checked?: boolean;
	incident_notes?: string;
	operator_notes?: string;
	handover_notes?: string;
};

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const rid = requiredRouteParam(event.params.rid, 'rid');
		const data = await proxy(event).get(apiPath`/api/cbt/sessions/${id}/rooms/${rid}/handover`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id]/rooms/[rid]/handover GET');
	}
};

export const PUT = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const rid = requiredRouteParam(event.params.rid, 'rid');
		const body = await readRequestJson<RoomHandoverPayload>(event.request);
		const data = await proxy(event).put(apiPath`/api/cbt/sessions/${id}/rooms/${rid}/handover`, {
			attendance_checked: body.attendance_checked === true,
			all_submitted_checked: body.all_submitted_checked === true,
			device_issue_checked: body.device_issue_checked === true,
			room_clean_checked: body.room_clean_checked === true,
			token_returned_checked: body.token_returned_checked === true,
			assets_returned_checked: body.assets_returned_checked === true,
			incident_notes: body.incident_notes ?? '',
			operator_notes: body.operator_notes ?? '',
			handover_notes: body.handover_notes ?? '',
		});
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id]/rooms/[rid]/handover PUT');
	}
};
