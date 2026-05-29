export type AsesmenWorkflowSessionLike = {
	status?: string;
	session_status?: string;
	scheduled_start?: string;
	event_id?: string;
	unassigned_participant_count?: number;
};

export type AsesmenWorkflowReadiness = {
	label: string;
	tone: 'neutral' | 'warning' | 'success';
};

export function latestEventDocumentHubHref(sessions: AsesmenWorkflowSessionLike[]) {
	const eventId = sessions.find((session) => session.event_id?.trim())?.event_id?.trim();
	return eventId ? `/asesmen/kegiatan/${encodeURIComponent(eventId)}/cetak` : '/asesmen/kegiatan';
}

export function countRunningSessions(sessions: AsesmenWorkflowSessionLike[]) {
	return sessions.filter((session) => (session.status ?? session.session_status) === 'active').length;
}

export function countUnassignedParticipants(sessions: AsesmenWorkflowSessionLike[]) {
	return sessions.reduce((sum, session) => sum + (session.unassigned_participant_count ?? 0), 0);
}

export function isSessionToday(value?: string, now = new Date()) {
	if (!value) return false;
	const date = new Date(value);
	if (Number.isNaN(date.getTime())) return false;
	return date.toDateString() === now.toDateString();
}

export function summarizeWorkflowReadiness(sessions: AsesmenWorkflowSessionLike[], now = new Date()): AsesmenWorkflowReadiness {
	if (sessions.length === 0) return { label: 'Belum ada sesi', tone: 'neutral' };
	if (countUnassignedParticipants(sessions) > 0) return { label: 'Perlu penempatan peserta', tone: 'warning' };
	if (countRunningSessions(sessions) > 0) return { label: 'Sedang berjalan', tone: 'success' };
	if (sessions.some((session) => isSessionToday(session.scheduled_start, now))) return { label: 'Siap hari ini', tone: 'neutral' };
	return { label: 'Terkendali', tone: 'neutral' };
}

export function workflowReadinessClass(tone: AsesmenWorkflowReadiness['tone']) {
	if (tone === 'warning') return 'bg-warning/15 text-warning';
	if (tone === 'success') return 'bg-success/15 text-success';
	return 'bg-muted text-muted-foreground';
}

export type AsesmenWorkflowUserLike = {
	role?: string | null;
	roles?: string[] | null;
	permissions?: string[] | null;
};

export type AsesmenWorkflowAccess = {
	canOpenPreparation: boolean;
	canOpenExecution: boolean;
	canOpenResults: boolean;
	canLoadDashboardStats: boolean;
};

export type AsesmenWorkflowPhaseCard = {
	code: string;
	title: string;
	description: string;
	href: string;
	cta: string;
	tone?: 'default' | 'primary';
};

export function normalizeAsesmenUserAccess(user?: AsesmenWorkflowUserLike | null) {
	const roles = user?.roles ?? (user?.role ? [user.role] : []);
	const permissions = (user?.permissions ?? []).map((permission) => permission.trim()).filter(Boolean);
	return { roles, permissions };
}

export function deriveAsesmenWorkflowAccess(user?: AsesmenWorkflowUserLike | null): AsesmenWorkflowAccess {
	const { roles, permissions } = normalizeAsesmenUserAccess(user);
	const isAdmin = roles.includes('admin');
	const canOpenPreparation =
		isAdmin
		|| permissions.includes('asesmen.operator')
		|| permissions.includes('asesmen.event_manage')
		|| permissions.includes('asesmen.package_manage')
		|| permissions.includes('asesmen.session_manage')
		|| permissions.includes('asesmen.participant_manage');
	const canOpenExecution = canOpenPreparation || permissions.includes('asesmen.proctor');
	const canOpenResults = isAdmin || permissions.includes('asesmen.result_read');
	const canLoadDashboardStats = canOpenPreparation || permissions.includes('asesmen.read');
	return { canOpenPreparation, canOpenExecution, canOpenResults, canLoadDashboardStats };
}

export type AsesmenWorkflowPhasePath = '/asesmen/persiapan' | '/asesmen/pelaksanaan' | '/asesmen/hasil';

export function buildAsesmenWorkflowPhaseCards(
	access: AsesmenWorkflowAccess,
	documentHubHref: string,
	resolveHref: (href: AsesmenWorkflowPhasePath) => string = (href) => href
): AsesmenWorkflowPhaseCard[] {
	return [
		{
			code: '7.1',
			title: 'Persiapan',
			description: 'Kegiatan, paket, sesi, ruang, peserta, dan pengawas sebelum pelaksanaan.',
			href: resolveHref('/asesmen/persiapan'),
			cta: 'Buka persiapan',
			show: access.canOpenPreparation
		},
		{
			code: '7.2',
			title: 'Pelaksanaan',
			description: 'Sesi panitia, Ruang Saya, panel ruang, dan bantuan portal saat ujian berjalan.',
			href: resolveHref('/asesmen/pelaksanaan'),
			cta: 'Buka pelaksanaan',
			tone: 'primary' as const,
			show: access.canOpenExecution
		},
		{
			code: '7.3',
			title: 'Hasil',
			description: 'Rekap nilai, status submit, koreksi uraian, analisis butir, dan sinkronisasi.',
			href: resolveHref('/asesmen/hasil'),
			cta: 'Buka hasil',
			show: access.canOpenResults
		},
		{
			code: '7.4',
			title: 'Arsip',
			description: 'Berita acara, rekap pelaksanaan, dan tindak lanjut sesi setelah ujian.',
			href: documentHubHref,
			cta: 'Buka arsip',
			show: access.canOpenPreparation || access.canOpenResults
		}
	].filter((item) => item.show).map(({ show: _show, ...item }) => item);
}
