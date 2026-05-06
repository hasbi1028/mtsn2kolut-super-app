export type AccountIdentity = {
	id: string;
	username: string;
	display_name?: string | null;
	roles: string[];
	profile_type?: string | null;
	profile_nama?: string | null;
	employee_id?: string | null;
	student_id?: string | null;
	parent_id?: string | null;
	is_active?: boolean;
	last_login_at?: string | null;
	created_at?: string | null;
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
