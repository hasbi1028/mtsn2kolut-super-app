import { clientApiPath, clientApiPathWithQuery, readClientApiData } from '$lib/client/api';

export type AssessmentApprovalEntityType = 'event' | 'session' | 'package' | 'result';
export type AssessmentApprovalType =
	| 'package_ready'
	| 'participants_rooms_ready'
	| 'tokens_cards_ready'
	| 'results_verified'
	| 'final_archive'
	| 'session_minutes'
	| 'room_handover';

export type AssessmentApprovalRecord = {
	id: string;
	entity_type: AssessmentApprovalEntityType | string;
	entity_id: string;
	approval_type: AssessmentApprovalType | string;
	status: 'approved' | 'revoked' | string;
	approved_by?: string | null;
	approved_by_display_name?: string;
	approved_by_username?: string;
	approved_at?: string | null;
	revoked_by?: string | null;
	revoked_by_display_name?: string;
	revoked_by_username?: string;
	revoked_at?: string | null;
	notes?: string;
	created_at?: string | null;
	updated_at?: string | null;
};

export type CreateAssessmentApprovalPayload = {
	entity_type: AssessmentApprovalEntityType;
	entity_id: string;
	approval_type: AssessmentApprovalType;
	notes?: string;
};

export const assessmentApprovalLabels: Record<AssessmentApprovalType, string> = {
	package_ready: 'Sahkan Paket Soal',
	participants_rooms_ready: 'Sahkan Peserta & Ruang',
	tokens_cards_ready: 'Sahkan Token & Kartu',
	results_verified: 'Sahkan Hasil',
	final_archive: 'Finalkan & Arsipkan',
	session_minutes: 'Sahkan Berita Acara',
	room_handover: 'Sahkan Serah Terima Ruang',
};

export async function listApprovals(entityType: AssessmentApprovalEntityType, entityId: string): Promise<AssessmentApprovalRecord[]> {
	const params = new URLSearchParams({ entity_type: entityType, entity_id: entityId });
	const response = await fetch(clientApiPathWithQuery('/api/asesmen/approvals', params));
	const rows = await readClientApiData<AssessmentApprovalRecord[]>(response, 'Gagal memuat pengesahan SOP');
	return Array.isArray(rows) ? rows : [];
}

export async function createApproval(payload: CreateAssessmentApprovalPayload): Promise<AssessmentApprovalRecord> {
	const response = await fetch('/api/asesmen/approvals', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(payload),
	});
	return readClientApiData<AssessmentApprovalRecord>(response, 'Gagal menyimpan pengesahan SOP');
}

export async function revokeApproval(id: string, reason: string): Promise<AssessmentApprovalRecord> {
	const response = await fetch(clientApiPath`/api/asesmen/approvals/${id}/revoke`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ notes: reason }),
	});
	return readClientApiData<AssessmentApprovalRecord>(response, 'Gagal mencabut pengesahan SOP');
}
