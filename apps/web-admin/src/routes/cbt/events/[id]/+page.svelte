<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { clientApiPath, readClientApiData } from '$lib/client/api';
	import { csvRow } from '$lib/csv';

	type EventInfo = {
		id: string; title: string; exam_type: string; scope: string;
		target_levels?: string[];
		academic_year_name: string; status: string;
	};
	type ResultRow = {
		participant_id: string; session_id: string; session_title: string;
		nis: string; student_nama: string; gender: string;
		class_code: string; score: string | null; submitted_at: string | null;
	};
	type EventSession = {
		id: string; title: string; status?: string; participant_count?: number; room_count?: number;
		missing_seat_count?: number; rooms_without_proctor?: number; unassigned_participant_count?: number;
	};
	type EventPackage = { id: string; title: string; question_count?: number; is_active?: boolean };
	type EventOverview = {
		member_count?: number; target_count?: number; review_count?: number; question_count?: number; published_question_count?: number;
		package_count?: number; session_count?: number; room_count?: number; token_count?: number; card_count?: number; result_count?: number;
	};
	type EventCommandDetail = {
		info: EventInfo;
		results: ResultRow[];
		overview: EventOverview | null;
		sessions: EventSession[];
		packages: EventPackage[];
	};
	type ChecklistHref =
		| `/cbt/events/${string}/members`
		| `/cbt/soal?event_id=${string}`
		| `/cbt/soal/review?event_id=${string}`
		| `/cbt/packages?event_id=${string}`
		| `/cbt/sessions?event_id=${string}`
		| `/cbt/sessions?event_id=${string}&readiness=needs_rooms`
		| `/cbt/events/${string}/exam-cards`
		| `/cbt/events/${string}#hasil`;
	type ChecklistItem = {
		label: string; helper: string; count: number | null; href: ChecklistHref; tone: 'success' | 'warning' | 'info'; action: string;
	};

	const eventId = page.params.id ?? '';
	let info = $state<EventInfo | null>(null);
	let results = $state<ResultRow[]>([]);
	let detailPromise = $state<Promise<EventCommandDetail> | null>(null);
	let detailRequestId = 0;

	const statusLabel: Record<string, string> = { draft: 'Draft', active: 'Aktif', finished: 'Selesai' };
	const scopeLabel: Record<string, string> = { class: 'Per Kelas', grade: 'Per Tingkat', school: 'Seluruh Sekolah' };

	function isRecord(value: unknown): value is Record<string, unknown> {
		return typeof value === 'object' && value !== null;
	}

	function parseArrayPayload<T>(payload: unknown, key: string): T[] {
		if (Array.isArray(payload)) return payload as T[];
		if (isRecord(payload) && Array.isArray(payload[key])) return payload[key] as T[];
		return [];
	}

	async function optionalApiData<T>(path: string, fallback: T): Promise<T> {
		try {
			return await fetch(path).then((response) => readClientApiData<T>(response));
		} catch {
			return fallback;
		}
	}

	async function fetchDetail(): Promise<EventCommandDetail> {
		const [nextInfo, nextResults, overviewPayload, sessionPayload, packagePayload] = await Promise.all([
			fetch(clientApiPath`/api/cbt/events/${eventId}`).then((response) => readClientApiData<EventInfo>(response, 'Gagal memuat kegiatan ujian')),
			fetch(clientApiPath`/api/cbt/events/${eventId}/results`).then((response) => readClientApiData<ResultRow[]>(response, 'Gagal memuat rekap nilai kegiatan')),
			optionalApiData<unknown>(clientApiPath`/api/cbt/events/${eventId}/overview`, null),
			optionalApiData<unknown>(clientApiPath`/api/cbt/events/${eventId}/sessions`, []),
			optionalApiData<unknown>(clientApiPath`/api/cbt/events/${eventId}/packages`, []),
		]);
		return {
			info: nextInfo,
			results: Array.isArray(nextResults) ? nextResults : [],
			overview: isRecord(overviewPayload) ? overviewPayload as EventOverview : null,
			sessions: parseArrayPayload<EventSession>(sessionPayload, 'sessions'),
			packages: parseArrayPayload<EventPackage>(packagePayload, 'packages'),
		};
	}

	function applyDetail(detail: EventCommandDetail) {
		info = detail.info;
		results = detail.results;
	}

	function loadInitial() {
		const requestId = ++detailRequestId;
		info = null;
		results = [];
		detailPromise = fetchDetail().then((detail) => {
			if (requestId !== detailRequestId) {
				if (!info) throw new Error('Permintaan dashboard kegiatan dibatalkan');
				return { info, results, overview: null, sessions: [], packages: [] };
			}
			applyDetail(detail);
			return detail;
		}).catch((error: unknown) => {
			if (requestId === detailRequestId || !info) throw error;
			return { info, results, overview: null, sessions: [], packages: [] };
		});
	}

	function retryDetail(reset?: () => void) {
		reset?.();
		loadInitial();
	}

	function detailErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat pusat kendali kegiatan';
	}

	function handleDetailRenderError(error: unknown) {
		console.error('CBT event command center render failed', error);
	}

	function fmtScore(score: string | null) {
		if (!score) return '-';
		const n = parseFloat(score);
		return isNaN(n) ? '-' : n.toFixed(1);
	}

	function countFrom(value: number | undefined, fallback: number | null) {
		return typeof value === 'number' ? value : fallback;
	}

	function checklistTone(count: number | null): ChecklistItem['tone'] {
		if (count === null) return 'info';
		return count > 0 ? 'success' : 'warning';
	}

	function checklistClass(tone: ChecklistItem['tone']) {
		if (tone === 'success') return 'border-emerald-200 bg-emerald-50/60';
		if (tone === 'warning') return 'border-amber-200 bg-amber-50/70';
		return 'border-slate-200 bg-white';
	}

	function statusClass(status: string) {
		if (status === 'active') return 'bg-emerald-100 text-emerald-700 border-emerald-200';
		if (status === 'finished') return 'bg-slate-100 text-slate-600 border-slate-200';
		return 'bg-amber-100 text-amber-700 border-amber-200';
	}

	function buildChecklist(detail: EventCommandDetail): ChecklistItem[] {
		const overview = detail.overview;
		const sessionFallback = detail.sessions.length > 0 ? detail.sessions.length : null;
		const packageFallback = detail.packages.length > 0 ? detail.packages.length : null;
		const roomIssues = detail.sessions.filter((session) => (session.room_count ?? 0) === 0 || (session.missing_seat_count ?? 0) > 0 || (session.rooms_without_proctor ?? 0) > 0 || (session.unassigned_participant_count ?? 0) > 0).length;
		const items: Array<Omit<ChecklistItem, 'tone'>> = [
			{ label: 'Penugasan', helper: 'Guru pembuat soal dan reviewer kegiatan', count: countFrom(overview?.member_count, null), href: `/cbt/events/${eventId}/members`, action: 'Atur penugasan' },
			{ label: 'Bank Soal/Target', helper: 'Soal terbit dan target per mapel', count: countFrom(overview?.published_question_count ?? overview?.question_count, null), href: `/cbt/soal?event_id=${eventId}`, action: 'Buka bank soal' },
			{ label: 'Review', helper: 'Antrean soal kegiatan yang perlu keputusan', count: countFrom(overview?.review_count, null), href: `/cbt/soal/review?event_id=${eventId}`, action: 'Review soal' },
			{ label: 'Paket', helper: 'Paket siap dipakai sesi ujian', count: countFrom(overview?.package_count, packageFallback), href: `/cbt/packages?event_id=${eventId}`, action: 'Kelola paket' },
			{ label: 'Sesi/Jadwal', helper: 'Sesi, status, dan jadwal operasional', count: countFrom(overview?.session_count, sessionFallback), href: `/cbt/sessions?event_id=${eventId}`, action: 'Kelola sesi' },
			{ label: 'Ruang/Pengawas/Kursi', helper: roomIssues > 0 ? `${roomIssues} sesi masih perlu dirapikan` : 'Cek ruang, pengawas, kapasitas, dan nomor meja', count: countFrom(overview?.room_count, detail.sessions.length > 0 ? detail.sessions.reduce((sum, session) => sum + (session.room_count ?? 0), 0) : null), href: `/cbt/sessions?event_id=${eventId}&readiness=needs_rooms`, action: 'Cek ruang' },
			{ label: 'Token/Kartu', helper: 'Token peserta dan kartu ujian siap cetak', count: countFrom(overview?.token_count ?? overview?.card_count, null), href: `/cbt/events/${eventId}/exam-cards`, action: 'Cetak kartu' },
			{ label: 'Hasil', helper: 'Rekap nilai gabungan tetap tersedia di bagian bawah', count: countFrom(overview?.result_count, detail.results.length), href: `/cbt/events/${eventId}#hasil`, action: 'Lihat hasil' },
		];
		return items.map((item) => ({ ...item, tone: checklistTone(item.count) }));
	}

	function exportCSV() {
		if (!info || results.length === 0) return;
		const header = csvRow(['NIS', 'Nama', 'Kelas', 'Sesi', 'Skor', 'Waktu Submit']);
		const rows = results.map(r => csvRow([r.nis, r.student_nama, r.class_code, r.session_title, fmtScore(r.score), r.submitted_at ?? '']));
		const csv = [header, ...rows].join('\n');
		const blob = new Blob([csv], { type: 'text/csv' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `rekap_${info.title.replace(/\s+/g, '_')}.csv`;
		a.click();
		URL.revokeObjectURL(url);
	}

	onMount(() => {
		void loadInitial();
	});
</script>

<svelte:head><title>Pusat Kendali Event — {info?.title ?? 'Kegiatan Ujian'}</title></svelte:head>

<div class="space-y-6 p-6">
	<div class="flex items-center gap-2 text-sm text-slate-500">
		<a href={resolve('/cbt/events')} class="hover:text-slate-700">Kegiatan Ujian</a>
		<span>/</span>
		<span class="text-slate-700 font-medium truncate max-w-xs">{info?.title ?? 'Pusat Kendali'}</span>
	</div>

	<AsyncContent promise={detailPromise} onerror={handleDetailRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<Skeleton class="h-32 w-full" />
				<div class="grid gap-3 md:grid-cols-4">
					{#each Array.from({ length: 8 }) as _, index (`event-command-skeleton-${index}`)}
						<Skeleton class="h-32 w-full" />
					{/each}
				</div>
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel title="Pusat Kendali Belum Tersaji" message={detailErrorMessage(error)} onRetry={() => retryDetail(reset)} />
		{/snippet}

		{#snippet children(value)}
			{@const detail = value as EventCommandDetail}
			{@const currentInfo = detail.info}
			{@const currentResults = detail.results}
			{@const checklist = buildChecklist(detail)}
			<section class="rounded-2xl border border-green-200 bg-gradient-to-br from-green-50 to-white p-5 shadow-sm">
				<div class="flex flex-wrap items-start justify-between gap-4">
					<div class="max-w-3xl">
						<p class="text-xs font-bold uppercase tracking-[0.18em] text-green-700">Event Command Center</p>
						<h1 class="mt-1 text-2xl font-semibold text-slate-900">{currentInfo.title}</h1>
						<p class="mt-2 text-sm text-slate-600">{currentInfo.academic_year_name} · <span class="capitalize">{currentInfo.exam_type}</span> · {scopeLabel[currentInfo.scope] ?? currentInfo.scope}</p>
						<p class="mt-2 text-sm text-slate-500">Kelola alur CBT dari penugasan hingga hasil tanpa kehilangan konteks kegiatan.</p>
					</div>
					<div class="flex flex-wrap items-center gap-2">
						<Badge class={statusClass(currentInfo.status)}>{statusLabel[currentInfo.status] ?? currentInfo.status}</Badge>
						{#each currentInfo.target_levels ?? [] as level (level)}
							<Badge variant="outline" class="bg-white">Tingkat {level}</Badge>
						{/each}
						{#if !currentInfo.target_levels?.length}
							<Badge variant="outline" class="bg-white">Target mengikuti cakupan</Badge>
						{/if}
					</div>
				</div>
			</section>

			<section class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
				{#each checklist as item (item.label)}
					<a href={resolve(item.href)} class={`block rounded-xl border p-4 shadow-sm transition hover:-translate-y-0.5 hover:shadow-md ${checklistClass(item.tone)}`}>
						<div class="flex items-start justify-between gap-3">
							<div>
								<p class="text-sm font-semibold text-slate-900">{item.label}</p>
								<p class="mt-1 text-xs text-slate-500">{item.helper}</p>
							</div>
							<Badge variant={item.tone === 'warning' ? 'secondary' : 'outline'} class="bg-white">{item.count ?? 'Cek'}</Badge>
						</div>
						<p class="mt-4 text-sm font-semibold text-green-800">{item.action}</p>
					</a>
				{/each}
			</section>

			<section class="grid gap-3 lg:grid-cols-2">
				<Card.Root>
					<Card.Header class="pb-2"><Card.Title class="text-base">Jalur Operasi Cepat</Card.Title><Card.Description>Semua tautan membawa event_id agar operator tetap berada dalam konteks kegiatan ini.</Card.Description></Card.Header>
					<Card.Content class="flex flex-wrap gap-2">
						<a href={resolve(`/cbt/events/${eventId}/members`)} class="rounded-md border border-green-200 bg-green-50 px-3 py-2 text-sm font-semibold text-green-800 hover:bg-green-100">Penugasan</a>
						<a href={resolve(`/cbt/soal?event_id=${eventId}`)} class="rounded-md border border-input bg-background px-3 py-2 text-sm font-medium text-slate-700 hover:bg-muted">Bank Soal</a>
						<a href={resolve(`/cbt/soal/review?event_id=${eventId}`)} class="rounded-md border border-input bg-background px-3 py-2 text-sm font-medium text-slate-700 hover:bg-muted">Review</a>
						<a href={resolve(`/cbt/packages?event_id=${eventId}`)} class="rounded-md border border-input bg-background px-3 py-2 text-sm font-medium text-slate-700 hover:bg-muted">Paket</a>
						<a href={resolve(`/cbt/sessions?event_id=${eventId}`)} class="rounded-md border border-input bg-background px-3 py-2 text-sm font-medium text-slate-700 hover:bg-muted">Sesi</a>
						<a href={resolve(`/cbt/events/${eventId}/exam-cards`)} class="rounded-md border border-input bg-background px-3 py-2 text-sm font-medium text-slate-700 hover:bg-muted">Kartu Ujian</a>
					</Card.Content>
				</Card.Root>
				<Card.Root>
					<Card.Header class="pb-2"><Card.Title class="text-base">Ringkasan Sumber Data</Card.Title><Card.Description>Endpoint overview/sessions/packages dipakai bila backend sudah tersedia; halaman tetap berjalan dengan data event dan hasil lama.</Card.Description></Card.Header>
					<Card.Content class="flex flex-wrap gap-2">
						<Badge variant="outline" class="bg-white">{detail.overview ? 'Overview tersedia' : 'Overview fallback'}</Badge>
						<Badge variant="outline" class="bg-white">{detail.packages.length} paket terdeteksi</Badge>
						<Badge variant="outline" class="bg-white">{detail.sessions.length} sesi terdeteksi</Badge>
						<Badge variant="outline" class="bg-white">{currentResults.length} baris hasil</Badge>
					</Card.Content>
				</Card.Root>
			</section>

			<Card.Root id="hasil">
				<Card.Header class="pb-2">
					<div class="flex flex-wrap items-start justify-between gap-3">
						<div><Card.Title class="text-base">Hasil dan Rekap Nilai Gabungan</Card.Title><Card.Description>Nilai dari seluruh sesi yang terhubung ke kegiatan ini.</Card.Description></div>
						<LoadingButton variant="outline" onclick={exportCSV} disabled={currentResults.length === 0} label="Ekspor CSV" />
					</div>
				</Card.Header>
				<Card.Content class="p-0 overflow-x-auto">
					<Table.Root>
						<Table.Header><Table.Row class="bg-slate-50"><Table.Head>NIS</Table.Head><Table.Head>Nama Siswa</Table.Head><Table.Head>Kelas</Table.Head><Table.Head>Sesi Ujian</Table.Head><Table.Head class="text-center">Skor</Table.Head><Table.Head>Status</Table.Head></Table.Row></Table.Header>
						<Table.Body>
							{#each currentResults as r (r.participant_id)}
								<Table.Row><Table.Cell class="font-mono text-sm">{r.nis}</Table.Cell><Table.Cell class="font-medium">{r.student_nama}</Table.Cell><Table.Cell><Badge variant="secondary" class="text-xs">{r.class_code || '-'}</Badge></Table.Cell><Table.Cell class="text-sm text-slate-600">{r.session_title}</Table.Cell><Table.Cell class="text-center font-bold text-green-700">{fmtScore(r.score)}</Table.Cell><Table.Cell>{#if r.submitted_at}<Badge variant="outline" class="bg-green-50 text-green-700 border-green-200">Selesai</Badge>{:else}<Badge variant="outline" class="text-slate-400 border-slate-200">Belum</Badge>{/if}</Table.Cell></Table.Row>
							{:else}
								<Table.Row><Table.Cell colspan={6} class="py-12 text-center text-slate-400">Belum ada data nilai untuk kegiatan ini.</Table.Cell></Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			</Card.Root>
		{/snippet}
	</AsyncContent>
</div>
