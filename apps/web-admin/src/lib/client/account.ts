export type AccountIdentity = {
	id: string;
	username: string;
	display_name?: string | null;
	roles: string[];
	profile_type?: string | null;
	profile_nama?: string | null;
	photo_url?: string | null;
	avatar_url?: string | null;
	contact?: AccountContact | null;
	employee_id?: string | null;
	student_id?: string | null;
	parent_id?: string | null;
	is_active?: boolean;
	last_login_at?: string | null;
	created_at?: string | null;
};

export type AccountContact = {
	phone?: string | null;
	email?: string | null;
	address?: string | null;
	editable_fields?: string[];
};

export type AuthSession = {
	id: string;
	device_label?: string | null;
	ip_address?: string | null;
	user_agent?: string | null;
	last_used_at?: string | null;
	created_at?: string | null;
	expires_at?: string | null;
};

export type AccountChangeRequestStatus = 'pending' | 'approved' | 'rejected' | 'cancelled';

export type AccountChangeRequest = {
	id: string;
	requester_user_id?: string | null;
	requester_username?: string | null;
	requester_display_name?: string | null;
	profile_type: string;
	profile_nama?: string | null;
	target_employee_id?: string | null;
	target_student_id?: string | null;
	target_parent_id?: string | null;
	field_key: string;
	field_label?: string | null;
	current_value: string;
	requested_value: string;
	reason: string;
	status: AccountChangeRequestStatus | string;
	reviewer_user_id?: string | null;
	reviewer_username?: string | null;
	review_note?: string | null;
	reviewed_at?: string | null;
	created_at?: string | null;
	updated_at?: string | null;
};

export type AccountChangeHistoryItem = {
	action: string;
	field_key: string;
	status: string;
	created_at?: string | null;
	reviewer_username?: string | null;
	review_note?: string | null;
};

export type AccountTimelineItem = {
	label: string;
	at?: string | null;
	actor?: string | null;
	note?: string | null;
	status?: string | null;
};

export type OfficialChangeFieldOption = {
	profile_type: string;
	field_key: string;
	label: string;
	value_type?: string | null;
	self_requestable?: boolean;
	reviewer_permission?: string | null;
	is_active?: boolean;
};

export type SidebarPreferences = {
	pinned_items?: string[];
	recent_items?: string[];
};

const roleLabels: Record<string, string> = {
	admin: 'Administrator',
	guru: 'Guru',
	staf: 'Staf',
	kesiswaan: 'Kesiswaan',
	siswa: 'Siswa',
	ortu: 'Orang Tua'
};

const profileTypeLabels: Record<string, string> = {
	employee: 'Pegawai',
	student: 'Siswa',
	parent: 'Orang Tua'
};

const officialFieldLabels: Record<string, string> = {
	nama: 'Nama resmi',
	tanggal_lahir: 'Tanggal lahir',
	parent_name: 'Nama orang tua/wali'
};

const profileHistoryFieldLabels: Record<string, string> = {
	contact: 'Kontak pribadi',
	avatar: 'Foto profil',
	...officialFieldLabels
};

const changeRequestStatusLabels: Record<string, string> = {
	pending: 'Menunggu',
	approved: 'Disetujui',
	rejected: 'Ditolak',
	cancelled: 'Dibatalkan'
};

const profileHistoryStatusLabels: Record<string, string> = {
	...changeRequestStatusLabels,
	completed: 'Selesai'
};

const profileHistoryActionLabels: Record<string, string> = {
	contact_update: 'Kontak pribadi diperbarui',
	avatar_update: 'Foto profil diperbarui',
	avatar_delete: 'Foto profil dihapus',
	change_request_created: 'Permintaan perubahan diajukan',
	change_request_cancelled: 'Permintaan perubahan dibatalkan',
	change_request_approved: 'Permintaan perubahan disetujui',
	change_request_rejected: 'Permintaan perubahan ditolak'
};

export function accountDisplayName(account: AccountIdentity | null | undefined) {
	const displayName = account?.display_name?.trim();
	if (displayName) return displayName;
	const profileName = account?.profile_nama?.trim();
	if (profileName) return profileName;
	return account?.username?.trim() || 'Akun';
}

export function roleLabel(role: string) {
	const normalized = role.trim();
	return roleLabels[normalized] ?? normalized;
}

export function profileTypeLabel(type: string | null | undefined) {
	const normalized = type?.trim() ?? '';
	return profileTypeLabels[normalized] ?? 'Tidak tertaut';
}

export function linkedProfileLabel(account: AccountIdentity | null | undefined) {
	if (!account) return 'Tidak tertaut';
	const profileName = account.profile_nama?.trim();
	const profileType = profileTypeLabel(account.profile_type);
	if (profileName) return `${profileType}: ${profileName}`;
	if (account.employee_id || account.student_id || account.parent_id) return `${profileType}: tertaut`;
	return 'Tidak tertaut';
}

export function accountAvatarUrl(account: AccountIdentity | null | undefined) {
	return account?.avatar_url?.trim() || account?.photo_url?.trim() || '';
}

export function accountInitials(account: AccountIdentity | null | undefined) {
	const label = accountDisplayName(account);
	const parts = label.split(/\s+/).filter(Boolean);
	const initials = parts.slice(0, 2).map((part) => part.charAt(0).toUpperCase()).join('');
	return initials || 'A';
}

export function editableContactFields(contact: AccountContact | null | undefined) {
	const fields = contact?.editable_fields;
	if (!Array.isArray(fields)) return [];
	const allowed = new Set(['phone', 'email', 'address']);
	return fields.filter((field) => allowed.has(field));
}

export function contactFieldEditable(contact: AccountContact | null | undefined, field: string) {
	return editableContactFields(contact).includes(field);
}

export function hasEditableContact(contact: AccountContact | null | undefined) {
	return editableContactFields(contact).length > 0;
}

export function officialFieldLabel(fieldKey: string | null | undefined) {
	const normalized = fieldKey?.trim() ?? '';
	return officialFieldLabels[normalized] ?? (normalized || 'Field resmi');
}

export function changeRequestStatusLabel(status: string | null | undefined) {
	const normalized = status?.trim() ?? '';
	return changeRequestStatusLabels[normalized] ?? (normalized || 'Tidak diketahui');
}

export function profileHistoryActionLabel(action: string | null | undefined) {
	const normalized = action?.trim() ?? '';
	return profileHistoryActionLabels[normalized] ?? (normalized || 'Aktivitas profil');
}

export function profileHistoryFieldLabel(fieldKey: string | null | undefined) {
	const normalized = fieldKey?.trim() ?? '';
	return profileHistoryFieldLabels[normalized] ?? (normalized || 'Profil');
}

export function profileHistoryStatusLabel(status: string | null | undefined) {
	const normalized = status?.trim() ?? '';
	return profileHistoryStatusLabels[normalized] ?? (normalized || 'Tidak diketahui');
}

export function changeRequestCanCancel(request: Pick<AccountChangeRequest, 'status'>) {
	return request.status === 'pending';
}

export function officialChangeFieldOptions(fields: OfficialChangeFieldOption[] | null | undefined): OfficialChangeFieldOption[] {
	if (!Array.isArray(fields)) return [];
	return fields
		.filter((field) =>
			typeof field.profile_type === 'string'
			&& field.profile_type.trim() !== ''
			&& typeof field.field_key === 'string'
			&& field.field_key.trim() !== ''
			&& field.self_requestable !== false
			&& field.is_active !== false
		)
		.map((field) => ({
			...field,
			profile_type: field.profile_type.trim(),
			field_key: field.field_key.trim(),
			label: field.label?.trim() || officialFieldLabel(field.field_key),
			value_type: field.value_type?.trim() || 'text'
		}));
}

export function normalizeChangeRequests(value: AccountChangeRequest[] | null | undefined) {
	return Array.isArray(value) ? value : [];
}

function isAccountChangeHistoryItem(value: unknown): value is AccountChangeHistoryItem {
	if (typeof value !== 'object' || value === null) return false;
	const item = value as Partial<AccountChangeHistoryItem>;
	return typeof item.action === 'string' && item.action.trim() !== ''
		&& typeof item.field_key === 'string'
		&& typeof item.status === 'string' && item.status.trim() !== '';
}

export function normalizeAccountChangeHistory(value: AccountChangeHistoryItem[] | null | undefined) {
	return Array.isArray(value) ? value.filter(isAccountChangeHistoryItem) : [];
}

export function changeRequestTimelineItems(request: AccountChangeRequest): AccountTimelineItem[] {
	const items: AccountTimelineItem[] = [{
		label: 'Diajukan',
		at: request.created_at,
		actor: request.requester_display_name || request.requester_username || null,
		status: 'pending'
	}];

	if (request.status === 'cancelled') {
		items.push({
			label: 'Dibatalkan',
			at: request.updated_at,
			status: 'cancelled'
		});
		return items;
	}

	if (request.reviewed_at || request.reviewer_username || request.review_note) {
		items.push({
			label: `Review ${changeRequestStatusLabel(request.status)}`,
			at: request.reviewed_at || request.updated_at,
			actor: request.reviewer_username || null,
			note: request.review_note || null,
			status: request.status
		});
	}

	return items;
}

export function formatAccountDateTime(value: string | null | undefined, emptyLabel = 'Belum tercatat') {
	if (!value) return emptyLabel;
	const date = new Date(value);
	if (Number.isNaN(date.getTime())) return emptyLabel;
	return date.toLocaleString('id-ID', {
		dateStyle: 'medium',
		timeStyle: 'short',
		timeZone: 'Asia/Makassar'
	});
}

export function sessionTitle(session: AuthSession) {
	const label = session.device_label?.trim();
	if (label) return label;
	const suffix = session.id.trim().slice(0, 8);
	return suffix ? `Sesi ${suffix}` : 'Sesi aktif';
}

export function normalizeAccountSessions(value: AuthSession[] | null | undefined) {
	return Array.isArray(value) ? value : [];
}

export function isCurrentSession(session: AuthSession, currentSessionId: string | null | undefined) {
	return session.id === (currentSessionId ?? '');
}

export function preferenceItemCount(items: string[] | null | undefined) {
	return Array.isArray(items) ? items.filter((item) => item.trim()).length : 0;
}

export function accountErrorMessage(error: unknown, fallbackMessage: string) {
	if (error instanceof Error && error.message.trim()) return error.message;
	if (typeof error === 'string' && error.trim()) return error;
	return fallbackMessage;
}
