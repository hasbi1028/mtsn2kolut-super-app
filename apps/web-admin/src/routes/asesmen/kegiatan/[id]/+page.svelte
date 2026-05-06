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
		| `/asesmen/kegiatan/${string}/members`
		| `/bank-soal/tambah?event_id=${string}`
		| '/bank-soal/tambah'
		| '/bank-soal/verifikasi'
		| `/asesmen/paket?event_id=${string}`
		| `/asesmen/sesi?event_id=${string}`
		| `/asesmen/sesi?event_id=${string}&readiness=not_ready`
		| `/asesmen/sesi?event_id=${string}&readiness=needs_rooms`
		| `/asesmen/sesi?event_id=${string}&readiness=needs_proctors`
		| `/asesmen/kegiatan/${string}/exam-cards`
		| `/asesmen/kegiatan/${string}#hasil`;
	type ChecklistItem = {
		label: string; helper: string; count: number | null; href: ChecklistHref; tone: 'success' | 'warning' | 'info'; action: string;
	};
	type EventSection = 'ringkasan' | 'persiapan' | 'operasional' | 'hasil';
	type ReadinessGroup = {
		id: EventSection;
		title: string;
		description: string;
		items: ChecklistItem[];
	};
	type NextAction = ChecklistItem & { priority: string };

	const eventId = page.params.id ?? '';
	let info = $state<EventInfo | null>(null);
	let results = $state<ResultRow[]>([]);
	let detailPromise = $state<Promise<EventCommandDetail> | null>(null);
	let activeSection = $state<EventSection>('ringkasan');
	let hasilSectionElement = $state<HTMLElement | null>(null);
	let hasilFocusRequest = $state(0);
	let detailRequestId = 0;
	let handledHasilFocusRequest = 0;
	const sectionTabs: Array<{ id: EventSection; label: string }> = [
		{ id: 'ringkasan', label: 'Ringkasan' },
		{ id: 'persiapan', label: 'Persiapan' },
		{ id: 'operasional', label: 'Operasional' },
		{ id: 'hasil', label: 'Hasil' },
	];

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
			fetch(clientApiPath`/api/asesmen/events/${eventId}`).then((response) => readClientApiData<EventInfo>(response, 'Gagal memuat kegiatan ujian')),
			fetch(clientApiPath`/api/asesmen/events/${eventId}/results`).then((response) => readClientApiData<ResultRow[]>(response, 'Gagal memuat rekap nilai kegiatan')),
			optionalApiData<unknown>(clientApiPath`/api/asesmen/events/${eventId}/overview`, null),
			optionalApiData<unknown>(clientApiPath`/api/asesmen/events/${eventId}/sessions`, []),
			optionalApiData<unknown>(clientApiPath`/api/asesmen/events/${eventId}/packages`, []),
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
		return 'Gagal memuat wizard kesiapan kegiatan';
	}

	function handleDetailRenderError(error: unknown) {
		console.error('CBT event readiness wizard render failed', error);
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

	function phaseBadgeClass(tone: ChecklistItem['tone']) {
		if (tone === 'success') return 'border-emerald-200 bg-emerald-50 text-emerald-700';
		if (tone === 'warning') return 'border-amber-200 bg-amber-50 text-amber-700';
		return 'border-slate-200 bg-white text-slate-600';
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
		const proctorIssues = detail.sessions.filter((session) => (session.rooms_without_proctor ?? 0) > 0).length;
		const items: Array<Omit<ChecklistItem, 'tone'>> = [
			{ label: 'Penugasan', helper: 'Guru pembuat soal dan reviewer kegiatan', count: countFrom(overview?.member_count, null), href: `/asesmen/kegiatan/${eventId}/members`, action: 'Atur penugasan' },
			{ label: 'Kebutuhan Soal', helper: 'Target kebutuhan event; Bank Soal tetap repositori mandiri sebelum dipakai paket', count: countFrom(overview?.published_question_count ?? overview?.question_count, null), href: `/bank-soal/tambah?event_id=${eventId}`, action: 'Cek target kebutuhan' },
			{ label: 'Review Repositori', helper: 'Antrean review dari Bank Soal sebelum soal diterbitkan dan masuk paket', count: countFrom(overview?.review_count, null), href: '/bank-soal/verifikasi', action: 'Review repositori' },
			{ label: 'Paket Event', helper: 'Prioritas persiapan: paket yang tertaut event agar sesi ujian bisa memakai paket yang tepat', count: countFrom(overview?.package_count, packageFallback), href: `/asesmen/paket?event_id=${eventId}`, action: 'Kelola paket event' },
			{ label: 'Sesi/Jadwal', helper: 'Sesi, status, dan jadwal operasional', count: countFrom(overview?.session_count, sessionFallback), href: `/asesmen/sesi?event_id=${eventId}`, action: 'Kelola sesi' },
			{ label: 'Ruang/Pengawas/Kursi', helper: roomIssues > 0 ? `${roomIssues} sesi masih perlu dirapikan${proctorIssues > 0 ? `, ${proctorIssues} butuh pengawas` : ''}` : 'Cek ruang, pengawas, kapasitas, dan nomor meja', count: countFrom(overview?.room_count, detail.sessions.length > 0 ? detail.sessions.reduce((sum, session) => sum + (session.room_count ?? 0), 0) : null), href: `/asesmen/sesi?event_id=${eventId}&readiness=not_ready`, action: 'Cek ruang' },
			{ label: 'Token/Kartu', helper: 'Token peserta dan kartu ujian siap cetak', count: countFrom(overview?.token_count ?? overview?.card_count, null), href: `/asesmen/kegiatan/${eventId}/exam-cards`, action: 'Cetak kartu' },
			{ label: 'Hasil', helper: 'Rekap nilai gabungan tersedia di tab Hasil', count: countFrom(overview?.result_count, detail.results.length), href: `/asesmen/kegiatan/${eventId}#hasil`, action: 'Buka tab hasil' },
		];
		return items.map((item) => ({ ...item, tone: item.label === 'Ruang/Pengawas/Kursi' && roomIssues > 0 ? 'warning' : checklistTone(item.count) }));
	}

	function readinessStatusLabel(item: ChecklistItem) {
		if (item.tone === 'success') return item.count === null ? 'Tersedia' : `${item.count} siap`;
		if (item.tone === 'warning') return item.count === null ? 'Perlu dicek' : `${item.count} perlu dilengkapi`;
		return item.count === null ? 'Cek data' : `${item.count} terbaca`;
	}

	function readinessDotClass(tone: ChecklistItem['tone']) {
		if (tone === 'success') return 'bg-emerald-500';
		if (tone === 'warning') return 'bg-amber-500';
		return 'bg-slate-300';
	}

	function buildReadinessGroups(checklist: ChecklistItem[]): ReadinessGroup[] {
		const byLabel = new Map(checklist.map((item) => [item.label, item]));
		return [
			{
				id: 'persiapan',
				title: 'Paket Soal',
				description: 'Tim, kebutuhan soal, review, dan paket yang akan dipakai sesi.',
				items: ['Penugasan', 'Kebutuhan Soal', 'Review Repositori', 'Paket Event'].map((label) => byLabel.get(label)).filter((item): item is ChecklistItem => Boolean(item)),
			},
			{
				id: 'operasional',
				title: 'Kegiatan & Sesi',
				description: 'Jadwal, ruang, pengawas, kursi, token, dan kartu ujian.',
				items: ['Sesi/Jadwal', 'Token/Kartu'].map((label) => byLabel.get(label)).filter((item): item is ChecklistItem => Boolean(item)),
			},
			{
				id: 'operasional',
				title: 'Monitoring',
				description: 'Kesiapan ruang dan pengawasan saat ujian berlangsung.',
				items: ['Ruang/Pengawas/Kursi'].map((label) => byLabel.get(label)).filter((item): item is ChecklistItem => Boolean(item)),
			},
			{
				id: 'hasil',
				title: 'Hasil',
				description: 'Rekap nilai gabungan dan ekspor saat data sudah masuk.',
				items: ['Hasil'].map((label) => byLabel.get(label)).filter((item): item is ChecklistItem => Boolean(item)),
			},
		];
	}

	function buildNextActions(detail: EventCommandDetail, checklist: ChecklistItem[]): NextAction[] {
		const needsAttention = checklist.filter((item) => item.tone === 'warning');
		const preferredLabels = detail.info.status === 'finished'
			? ['Hasil', 'Token/Kartu', 'Sesi/Jadwal']
			: ['Paket Event', 'Sesi/Jadwal', 'Ruang/Pengawas/Kursi', 'Token/Kartu', 'Hasil'];
		const preferred = preferredLabels.map((label) => checklist.find((item) => item.label === label)).filter((item): item is ChecklistItem => Boolean(item));
		const ordered = [...needsAttention, ...preferred, ...checklist].filter((item, index, source) => source.findIndex((candidate) => candidate.label === item.label) === index);
		return ordered.slice(0, 3).map((item, index) => ({ ...item, priority: index === 0 ? 'Utama' : `Langkah ${index + 1}` }));
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

	function activateHasilHash() {
		if (window.location.hash !== '#hasil') return;
		activeSection = 'hasil';
		hasilFocusRequest += 1;
	}

	function handleHashChange() {
		activateHasilHash();
	}

	$effect(() => {
		if (hasilFocusRequest === handledHasilFocusRequest || activeSection !== 'hasil' || !hasilSectionElement) return;
		handledHasilFocusRequest = hasilFocusRequest;
		hasilSectionElement.focus({ preventScroll: true });
		hasilSectionElement.scrollIntoView({ block: 'start', behavior: 'smooth' });
	});

	onMount(() => {
		activateHasilHash();
		void loadInitial();
	});
</script>

<svelte:head><title>Wizard Kesiapan Event — {info?.title ?? 'Kegiatan Ujian'}</title></svelte:head>

<svelte:window onhashchange={handleHashChange} />

<div class="space-y-6 p-6">
	<div class="flex items-center gap-2 text-sm text-slate-500">
		<a href={resolve('/asesmen/kegiatan')} class="hover:text-slate-700">Kegiatan & Sesi CBT</a>
		<span>/</span>
		<span class="text-slate-700 font-medium truncate max-w-xs">{info?.title ?? 'Pusat Kendali'}</span>
	</div>

	<AsyncContent promise={detailPromise} onerror={handleDetailRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<Skeleton class="h-32 w-full" />
				<div class="grid gap-3 lg:grid-cols-[0.82fr_1.18fr]">
					<Skeleton class="h-56 w-full" />
					<Skeleton class="h-56 w-full" />
				</div>
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel title="Wizard Kesiapan Belum Tersaji" message={detailErrorMessage(error)} onRetry={() => retryDetail(reset)} />
		{/snippet}

		{#snippet children(value)}
			{@const detail = value as EventCommandDetail}
			{@const currentInfo = detail.info}
			{@const currentResults = detail.results}
			{@const checklist = buildChecklist(detail)}
			{@const readinessGroups = buildReadinessGroups(checklist)}
			{@const nextActions = buildNextActions(detail, checklist)}
			<section class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
				<div class="flex flex-wrap items-start justify-between gap-4">
					<div class="max-w-3xl p-5">
						<p class="text-xs font-bold uppercase tracking-[0.18em] text-emerald-700">Wizard Kesiapan CBT</p>
						<h1 class="mt-1 text-2xl font-semibold text-slate-900">{currentInfo.title}</h1>
						<p class="mt-2 text-sm text-slate-600">{currentInfo.academic_year_name} · <span class="capitalize">{currentInfo.exam_type}</span> · {scopeLabel[currentInfo.scope] ?? currentInfo.scope}</p>
						<p class="mt-2 text-sm text-slate-500">Ikuti langkah kesiapan dari paket soal, sesi, monitoring, sampai hasil tanpa membuka banyak kartu modul yang setara.</p>
					</div>
					<div class="flex flex-wrap items-center gap-2 p-5 lg:justify-end">
						<Badge class={statusClass(currentInfo.status)}>{statusLabel[currentInfo.status] ?? currentInfo.status}</Badge>
						{#each currentInfo.target_levels ?? [] as level (level)}
							<Badge variant="outline" class="bg-white">Tingkat {level}</Badge>
						{/each}
						{#if !currentInfo.target_levels?.length}
							<Badge variant="outline" class="bg-white">Target mengikuti cakupan</Badge>
						{/if}
					</div>
				</div>
				<nav class="flex gap-1 overflow-x-auto border-t border-slate-200 bg-slate-50 px-3 py-2" aria-label="Bagian pusat kegiatan">
					{#each sectionTabs as tab (tab.id)}
						<button
							type="button"
							class={`rounded-full px-3 py-1.5 text-sm font-semibold transition ${activeSection === tab.id ? 'bg-white text-emerald-800 shadow-sm ring-1 ring-emerald-200' : 'text-slate-600 hover:bg-white hover:text-slate-900'}`}
							aria-current={activeSection === tab.id ? 'page' : undefined}
							aria-pressed={activeSection === tab.id}
							onclick={() => activeSection = tab.id}
						>
							{tab.label}
						</button>
					{/each}
				</nav>
			</section>

			<section class="grid gap-4 lg:grid-cols-[0.82fr_1.18fr]" aria-label="Wizard kesiapan kegiatan">
				<Card.Root class="border-emerald-200 bg-emerald-50/40 shadow-sm">
					<Card.Header class="pb-2">
						<Card.Title class="text-base">Langkah berikutnya</Card.Title>
						<Card.Description>Rekomendasi ringkas dari data kesiapan yang tersedia saat ini.</Card.Description>
					</Card.Header>
					<Card.Content class="space-y-3">
						{#each nextActions as action (action.label)}
							<a href={resolve(action.href)} class="block rounded-xl border border-white bg-white p-4 shadow-sm transition hover:border-emerald-200 hover:shadow-md">
								<div class="flex items-start justify-between gap-3">
									<div class="min-w-0">
										<p class="text-xs font-bold uppercase tracking-[0.16em] text-emerald-700">{action.priority}</p>
										<p class="mt-1 text-sm font-semibold text-slate-900">{action.action}</p>
										<p class="mt-1 text-xs leading-5 text-slate-500">{action.helper}</p>
									</div>
									<Badge variant={action.tone === 'warning' ? 'secondary' : 'outline'} class={phaseBadgeClass(action.tone)}>{readinessStatusLabel(action)}</Badge>
								</div>
							</a>
						{/each}
					</Card.Content>
				</Card.Root>

				<Card.Root class="border-slate-200 shadow-sm">
					<Card.Header class="pb-2">
						<Card.Title class="text-base">Kesiapan kegiatan</Card.Title>
						<Card.Description>Satu permukaan utama untuk membaca progres. Buka tab di bawah untuk rincian kerja atau hasil.</Card.Description>
					</Card.Header>
					<Card.Content class="space-y-4">
						{#each readinessGroups as group (group.title)}
							<div class="rounded-xl border border-slate-200 bg-white p-4">
								<div class="flex flex-wrap items-start justify-between gap-3">
									<div>
										<p class="text-sm font-semibold text-slate-900">{group.title}</p>
										<p class="mt-1 text-xs text-slate-500">{group.description}</p>
									</div>
									<button type="button" class="text-xs font-semibold text-emerald-800 hover:text-emerald-900" onclick={() => activeSection = group.id}>Buka tab</button>
								</div>
								<div class="mt-3 divide-y divide-slate-100">
									{#each group.items as item (item.label)}
										<a href={resolve(item.href)} class="flex items-center justify-between gap-3 py-2 text-sm hover:text-emerald-800">
											<span class="flex min-w-0 items-center gap-2">
												<span class={`size-2 rounded-full ${readinessDotClass(item.tone)}`}></span>
												<span class="truncate font-medium text-slate-800">{item.label}</span>
											</span>
											<span class="shrink-0 text-xs font-semibold text-slate-500">{readinessStatusLabel(item)}</span>
										</a>
									{/each}
								</div>
							</div>
						{/each}
					</Card.Content>
				</Card.Root>
			</section>

			{#if activeSection === 'ringkasan'}
				<section>
					<Card.Root>
						<Card.Header class="pb-2"><Card.Title class="text-base">Kelengkapan data</Card.Title><Card.Description>Angka praktis dari paket, sesi, dan hasil yang sudah terbaca.</Card.Description></Card.Header>
						<Card.Content class="grid gap-2 sm:grid-cols-2 lg:grid-cols-4">
							<div class="rounded-xl bg-slate-50 p-3"><p class="text-xs text-slate-500">Paket</p><p class="text-lg font-semibold text-slate-900">{detail.packages.length}</p></div>
							<div class="rounded-xl bg-slate-50 p-3"><p class="text-xs text-slate-500">Sesi</p><p class="text-lg font-semibold text-slate-900">{detail.sessions.length}</p></div>
							<div class="rounded-xl bg-slate-50 p-3"><p class="text-xs text-slate-500">Baris hasil</p><p class="text-lg font-semibold text-slate-900">{currentResults.length}</p></div>
							<div class="rounded-xl bg-slate-50 p-3"><p class="text-xs text-slate-500">Ringkasan kesiapan</p><p class="text-sm font-semibold text-slate-900">{detail.overview ? 'Lengkap dari sistem' : 'Sebagian data tersedia'}</p></div>
						</Card.Content>
					</Card.Root>
				</section>
			{/if}

			{#if activeSection === 'persiapan' || activeSection === 'operasional'}
				{@const activeGroups = readinessGroups.filter((item) => item.id === activeSection)}
				{#if activeGroups.length > 0}
					<section class="grid gap-3 md:grid-cols-2">
						{#each activeGroups as group (group.title)}
							{#each group.items as item (item.label)}
							<a href={resolve(item.href)} class={`block rounded-xl border p-4 shadow-sm transition hover:-translate-y-0.5 hover:shadow-md ${checklistClass(item.tone)}`}>
								<div class="flex items-start justify-between gap-3">
									<div>
										<p class="text-xs font-semibold uppercase tracking-[0.14em] text-slate-500">{group.title}</p>
										<p class="mt-1 text-sm font-semibold text-slate-900">{item.label}</p>
										<p class="mt-1 text-xs text-slate-500">{item.helper}</p>
									</div>
									<Badge variant={item.tone === 'warning' ? 'secondary' : 'outline'} class="bg-white">{item.count ?? 'Cek'}</Badge>
								</div>
								<p class="mt-4 text-sm font-semibold text-green-800">{item.action}</p>
							</a>
							{/each}
						{/each}
					</section>
				{/if}
			{/if}

			{#if activeSection === 'hasil'}
			<Card.Root id="hasil" bind:ref={hasilSectionElement} tabindex={-1}>
				<Card.Header class="pb-2">
					<div class="flex flex-wrap items-start justify-between gap-3">
						<div><Card.Title class="text-base">Hasil & Analisis</Card.Title><Card.Description>Rekap nilai gabungan dari seluruh sesi dalam kegiatan ini; gunakan ekspor untuk analisis lanjutan.</Card.Description></div>
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
			{/if}
		{/snippet}
	</AsyncContent>
</div>
