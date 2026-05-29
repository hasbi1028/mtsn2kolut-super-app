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
