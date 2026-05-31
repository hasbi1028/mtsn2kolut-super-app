export type PackageReadinessFilter =
	| 'all'
	| 'empty'
	| 'ready'
	| 'kurang_pg'
	| 'kurang_essay'
	| 'metadata_gap'
	| 'locked'
	| 'used';

export type PackageReadinessStatus = {
	status?: string;
	target_pg_count?: number;
	target_essay_count?: number;
	missing_pg_count?: number;
	missing_essay_count?: number;
	question_count?: number;
	pg_count?: number;
	essay_count?: number;
	total_points?: number;
	published_count?: number;
	unpublished_count?: number;
	metadata_gap_count?: number;
	session_count?: number;
	locked?: boolean;
	ready?: boolean;
};

export type PackageReadinessListItem = {
	id: string;
	event_id?: string;
	subject_id: string;
	subject_code?: string;
	subject_name: string;
	title: string;
	description?: string;
	duration_minutes: number;
	question_count: number;
	session_count: number;
	locked?: boolean;
	snapshot_version?: number;
	readiness?: PackageReadinessStatus;
};

export function packageReadiness(item: PackageReadinessListItem): PackageReadinessStatus {
	return {
		question_count: Number(item.question_count ?? 0),
		session_count: Number(item.session_count ?? 0),
		locked: Boolean(item.locked),
		ready: Number(item.question_count ?? 0) > 0 && !item.locked,
		...(item.readiness ?? {})
	};
}

export function packageReadinessBadge(item: PackageReadinessListItem): { label: string; tone: 'slate' | 'emerald' | 'amber' | 'orange' | 'sky' } {
	const readiness = packageReadiness(item);
	if (readiness.locked || item.locked) return { label: 'Terkunci', tone: 'slate' };
	if (Number(readiness.question_count ?? item.question_count ?? 0) === 0 || readiness.status === 'kosong') return { label: 'Kosong', tone: 'amber' };
	if (Number(readiness.missing_pg_count ?? 0) > 0) return { label: 'Kurang PG', tone: 'orange' };
	if (Number(readiness.missing_essay_count ?? 0) > 0) return { label: 'Kurang Essay', tone: 'orange' };
	if (Number(readiness.metadata_gap_count ?? 0) > 0) return { label: 'Metadata Gap', tone: 'orange' };
	if (Number(readiness.unpublished_count ?? 0) > 0) return { label: 'Belum Terbit', tone: 'orange' };
	if (readiness.ready) return { label: 'Siap dikunci', tone: 'emerald' };
	if (Number(readiness.session_count ?? item.session_count ?? 0) > 0) return { label: 'Dipakai', tone: 'sky' };
	return { label: 'Perlu dilengkapi', tone: 'orange' };
}

export function filterPackagesByReadiness(items: PackageReadinessListItem[], filter: PackageReadinessFilter): PackageReadinessListItem[] {
	if (filter === 'all') return items;
	return items.filter((item) => {
		const readiness = packageReadiness(item);
		switch (filter) {
			case 'empty':
				return Number(readiness.question_count ?? item.question_count ?? 0) === 0;
			case 'ready':
				return Boolean(readiness.ready) && !readiness.locked && Number(readiness.metadata_gap_count ?? 0) === 0;
			case 'kurang_pg':
				return Number(readiness.missing_pg_count ?? 0) > 0;
			case 'kurang_essay':
				return Number(readiness.missing_essay_count ?? 0) > 0;
			case 'metadata_gap':
				return Number(readiness.metadata_gap_count ?? 0) > 0;
			case 'locked':
				return Boolean(readiness.locked || item.locked);
			case 'used':
				return Number(readiness.session_count ?? item.session_count ?? 0) > 0;
			default:
				return true;
		}
	});
}

function csvCell(value: unknown): string {
	const text = String(value ?? '');
	if (!/[",\n\r]/.test(text)) return text;
	return `"${text.replaceAll('"', '""')}"`;
}

export function buildPackageReadinessCsv(items: PackageReadinessListItem[]): string {
	const header = ['No', 'Mapel', 'Judul Paket', 'Status', 'PG', 'Essay', 'Gap Metadata', 'Belum Terbit', 'Poin', 'Pemakaian'];
	const rows = items.map((item, index) => {
		const readiness = packageReadiness(item);
		const pg = `${Number(readiness.pg_count ?? 0)}/${Number(readiness.target_pg_count ?? 20)}`;
		const essay = `${Number(readiness.essay_count ?? 0)}/${Number(readiness.target_essay_count ?? 5)}`;
		return [
			index + 1,
			item.subject_name || item.subject_code || '-',
			item.title,
			packageReadinessBadge(item).label,
			pg,
			essay,
			Number(readiness.metadata_gap_count ?? 0),
			Number(readiness.unpublished_count ?? 0),
			Number(readiness.total_points ?? item.question_count ?? 0),
			`${Number(readiness.session_count ?? item.session_count ?? 0)} sesi`
		];
	});
	return [header, ...rows].map((row) => row.map(csvCell).join(',')).join('\n');
}
