type SearchLike = URLSearchParams | string | null | undefined;

type VersionLike = {
	version_number?: number | null;
	is_latest_version?: boolean | null;
	created_at?: string | null;
	updated_at?: string | null;
	version_note?: string | null;
	review_notes?: string | null;
};

const LIST_CONTEXT_KEYS = [
	'event_id',
	'subject_id',
	'workflow_status',
	'status',
	'q',
	'page',
	'revision_source',
] as const;

function paramsFromSearch(search: SearchLike): URLSearchParams {
	if (search instanceof URLSearchParams) return new URLSearchParams(search);
	return new URLSearchParams((search ?? '').replace(/^\?/, ''));
}

export function detailRouteQuestionId(pathId: string | null | undefined): string {
	return (pathId ?? '').trim();
}

export function composerRouteQuestionId(pathId: string | null | undefined, search: SearchLike): string {
	const fromPath = detailRouteQuestionId(pathId);
	if (fromPath) return fromPath;
	return (paramsFromSearch(search).get('question_id') ?? '').trim();
}

export function bankSoalListContextParams(search: SearchLike, overrides: Record<string, string | number | null | undefined> = {}): URLSearchParams {
	const source = paramsFromSearch(search);
	const params = new URLSearchParams();
	for (const key of LIST_CONTEXT_KEYS) {
		const value = source.get(key);
		if (value) params.set(key, value);
	}
	for (const [key, value] of Object.entries(overrides)) {
		if (!LIST_CONTEXT_KEYS.includes(key as (typeof LIST_CONTEXT_KEYS)[number])) continue;
		const normalized = String(value ?? '').trim();
		if (normalized) params.set(key, normalized);
		else params.delete(key);
	}
	return params;
}

export function bankSoalListHref(search: SearchLike, overrides: Record<string, string | number | null | undefined> = {}): '/bank-soal' | `/bank-soal?${string}` {
	const params = bankSoalListContextParams(search, overrides);
	const query = params.toString();
	return query ? `/bank-soal?${query}` : '/bank-soal';
}

export function bankSoalQuestionDetailHref(id: string, search: SearchLike = ''): string {
	const safeId = encodeURIComponent(detailRouteQuestionId(id));
	const query = bankSoalListContextParams(search).toString();
	return query ? `/bank-soal/soal/${safeId}?${query}` : `/bank-soal/soal/${safeId}`;
}

export function questionVersionLabel(item: VersionLike | null | undefined): string {
	const version = item?.version_number && item.version_number > 0 ? item.version_number : 1;
	return `v${version}`;
}

export function composerDateTimeLabel(value: string | null | undefined): string {
	if (!value) return '-';
	const date = new Date(value);
	if (Number.isNaN(date.getTime())) return '-';
	return date.toLocaleString('id-ID', {
		day: '2-digit',
		month: 'short',
		year: 'numeric',
		hour: '2-digit',
		minute: '2-digit',
	});
}

export function composerTimeLabel(value: string | null | undefined): string {
	if (!value) return '';
	const date = new Date(value);
	if (Number.isNaN(date.getTime())) return '';
	return date.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });
}

export function compactVersionNote(item: VersionLike | null | undefined, maxLength = 120): string {
	const note = (item?.version_note || item?.review_notes || '').replace(/\s+/g, ' ').trim();
	if (!note) return 'Belum ada catatan versi.';
	return note.length > maxLength ? `${note.slice(0, Math.max(0, maxLength - 3))}...` : note;
}
