export type CommandRoomStatus = 'setup' | 'ready' | 'running' | 'attention' | 'critical' | 'finishing' | 'final';
export type CommandIssueSeverity = 'critical' | 'warning' | 'info';

export type CommandRoomStatusInput = {
	participant_count?: number;
	joined_count?: number;
	submitted_count?: number;
	no_show_count?: number;
	suspicious_count?: number;
	incident_event_count?: number;
	missing_seat_count?: number;
	force_submit_count?: number;
	reset_access_count?: number;
	offline_count?: number;
	proctor_count?: number;
	handover_locked?: boolean;
	allow_web_fallback?: boolean;
	minutes_remaining?: number | null;
	session_status?: string | null;
};

export type CommandCenterIssue = {
	id: string;
	severity: CommandIssueSeverity;
	title: string;
	description?: string;
	actionLabel?: string;
	actionTab?: string;
	actionHref?: string;
	roomId?: string;
};

const severityRank: Record<CommandIssueSeverity, number> = {
	critical: 0,
	warning: 1,
	info: 2
};

export function maskToken(token: string | null | undefined) {
	if (!token) return '••••';
	const trimmed = String(token).trim();
	if (trimmed.length < 10) return '••••';
	return `${trimmed.slice(0, 4)}••••${trimmed.slice(-4)}`;
}

export function deriveCommandRoomStatus(input: CommandRoomStatusInput): CommandRoomStatus {
	if (input.handover_locked) return 'final';
	if ((input.offline_count ?? 0) > 0 && (input.minutes_remaining ?? 99) <= 10) return 'critical';
	if ((input.proctor_count ?? 1) === 0 && input.session_status === 'active') return 'critical';
	if ((input.missing_seat_count ?? 0) > 0) return 'attention';
	if ((input.suspicious_count ?? 0) > 0 || (input.incident_event_count ?? 0) > 0) return 'attention';
	if ((input.force_submit_count ?? 0) > 0 || (input.reset_access_count ?? 0) > 0) return 'attention';
	if (input.allow_web_fallback) return 'attention';
	if ((input.minutes_remaining ?? 99) <= 10 && (input.submitted_count ?? 0) < (input.participant_count ?? 0)) return 'finishing';
	if (input.session_status === 'active') return 'running';
	if ((input.participant_count ?? 0) === 0 || (input.missing_seat_count ?? 0) > 0) return 'setup';
	return 'ready';
}

export function commandRoomStatusLabel(status: CommandRoomStatus) {
	const labels: Record<CommandRoomStatus, string> = {
		setup: 'Setup Belum Lengkap',
		ready: 'Siap',
		running: 'Berjalan Normal',
		attention: 'Perlu Atensi',
		critical: 'Kritis',
		finishing: 'Mendekati Akhir',
		final: 'Final'
	};
	return labels[status];
}

export function commandRoomStatusClass(status: CommandRoomStatus) {
	if (status === 'critical') return 'border-destructive/40 bg-destructive/10 text-destructive';
	if (status === 'attention' || status === 'finishing') return 'border-warning/40 bg-warning/10 text-warning';
	if (status === 'final') return 'border-muted bg-muted text-muted-foreground';
	if (status === 'running' || status === 'ready') return 'border-success/30 bg-success/10 text-success';
	return 'border-border bg-muted/50 text-muted-foreground';
}

export function prioritizeCommandCenterIssues<T extends CommandCenterIssue>(issues: T[]): T[] {
	return [...issues].sort((a, b) => {
		const severity = severityRank[a.severity] - severityRank[b.severity];
		if (severity !== 0) return severity;
		return a.title.localeCompare(b.title, 'id');
	});
}
