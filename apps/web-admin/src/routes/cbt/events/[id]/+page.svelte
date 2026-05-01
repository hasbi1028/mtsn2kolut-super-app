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
	type EventResultsDetail = {
		info: EventInfo;
		results: ResultRow[];
	};

	const eventId = page.params.id ?? '';
	let info = $state<EventInfo | null>(null);
	let results = $state<ResultRow[]>([]);
	let detailPromise = $state<Promise<EventResultsDetail> | null>(null);
	let detailRequestId = 0;

	async function fetchDetail(): Promise<EventResultsDetail> {
		const [nextInfo, nextResults] = await Promise.all([
			fetch(clientApiPath`/api/cbt/events/${eventId}`).then((response) => readClientApiData<EventInfo>(response, 'Gagal memuat kegiatan ujian')),
			fetch(clientApiPath`/api/cbt/events/${eventId}/results`).then((response) => readClientApiData<ResultRow[]>(response, 'Gagal memuat rekap nilai kegiatan')),
		]);
		return {
			info: nextInfo,
			results: Array.isArray(nextResults) ? nextResults : [],
		};
	}

	function applyDetail(detail: EventResultsDetail) {
		info = detail.info;
		results = detail.results;
	}

	function loadInitial() {
		const requestId = ++detailRequestId;
		info = null;
		results = [];
		detailPromise = fetchDetail().then((detail) => {
			if (requestId !== detailRequestId) {
				if (!info) throw new Error('Permintaan rekap nilai dibatalkan');
				return { info, results };
			}
			applyDetail(detail);
			return detail;
		}).catch((error: unknown) => {
			if (requestId === detailRequestId || !info) throw error;
			return { info, results };
		});
	}

	function retryDetail(reset?: () => void) {
		reset?.();
		loadInitial();
	}

	function detailErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat rekap nilai kegiatan';
	}

	function handleDetailRenderError(error: unknown) {
		console.error('CBT event detail render failed', error);
	}

	function fmtScore(score: string | null) {
		if (!score) return '—';
		const n = parseFloat(score);
		return isNaN(n) ? '—' : n.toFixed(1);
	}

	function exportCSV() {
		if (!info || results.length === 0) return;
		const header = 'NIS,Nama,Kelas,Sesi,Skor,Waktu Submit';
		const rows = results.map(r =>
			[r.nis, `"${r.student_nama}"`, r.class_code, `"${r.session_title}"`, fmtScore(r.score), r.submitted_at ?? ''].join(',')
		);
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

<svelte:head><title>Rekap Nilai — {info?.title ?? 'Kegiatan Ujian'}</title></svelte:head>

<div class="space-y-6 p-6">
	<div class="flex items-center gap-2 text-sm text-slate-500">
		<a href={resolve('/cbt/events')} class="hover:text-slate-700">Kegiatan Ujian</a>
		<span>/</span>
		<span class="text-slate-700 font-medium truncate max-w-xs">{info?.title ?? '...'}</span>
	</div>

	<AsyncContent promise={detailPromise} onerror={handleDetailRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<div class="space-y-2">
					<Skeleton class="h-4 w-48" />
					<Skeleton class="h-8 w-72" />
					<Skeleton class="h-4 w-64" />
				</div>
				<Card.Root>
					<Card.Content class="space-y-3 p-6">
						{#each Array.from({ length: 6 }) as _, index (`event-result-skeleton-${index}`)}
							<div class="grid gap-3 lg:grid-cols-[0.8fr_1.4fr_0.6fr_1.2fr_0.5fr_0.6fr] lg:items-center">
								<Skeleton class="h-5 w-20" />
								<Skeleton class="h-5 w-full max-w-xs" />
								<Skeleton class="h-6 w-16" />
								<Skeleton class="h-5 w-36" />
								<Skeleton class="h-5 w-14" />
								<Skeleton class="h-6 w-16" />
							</div>
						{/each}
					</Card.Content>
				</Card.Root>
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Rekap Nilai Belum Tersaji"
				message={detailErrorMessage(error)}
				onRetry={() => retryDetail(reset)}
			/>
		{/snippet}

		{#snippet children(value)}
			{@const detail = value as EventResultsDetail}
			{@const currentInfo = detail.info}
			{@const currentResults = detail.results}
		<div class="flex items-start justify-between gap-4 flex-wrap">
			<div>
				<h1 class="text-2xl font-semibold text-slate-800">{currentInfo.title}</h1>
				<p class="text-sm text-slate-500 mt-1">
					Tahun Ajaran: {currentInfo.academic_year_name} · Tipe: <span class="capitalize">{currentInfo.exam_type}</span>
				</p>
				{#if currentInfo.target_levels?.length}
					<p class="mt-2 text-sm text-slate-600">Tingkat yang diikutkan: <span class="font-medium">{currentInfo.target_levels.join(', ')}</span></p>
				{/if}
			</div>
			<div class="flex flex-wrap gap-2">
				<a href={resolve(`/cbt/events/${eventId}/exam-cards`)} class="inline-flex items-center rounded-md border border-input bg-background px-3 py-2 text-sm font-medium text-slate-700 hover:bg-muted">
					Kartu Ujian
				</a>
				<LoadingButton variant="outline" onclick={exportCSV} disabled={currentResults.length === 0} label="↓ Ekspor CSV (Rekap Semua Sesi)" />
			</div>
		</div>

		<Card.Root>
			<Card.Header class="pb-2">
				<Card.Title class="text-base">Daftar Nilai Gabungan</Card.Title>
				<Card.Description>Nilai dari seluruh sesi yang terhubung ke kegiatan ini</Card.Description>
			</Card.Header>
			<Card.Content class="p-0 overflow-x-auto">
				<Table.Root>
					<Table.Header>
						<Table.Row class="bg-slate-50">
							<Table.Head>NIS</Table.Head>
							<Table.Head>Nama Siswa</Table.Head>
							<Table.Head>Kelas</Table.Head>
							<Table.Head>Sesi Ujian</Table.Head>
							<Table.Head class="text-center">Skor</Table.Head>
							<Table.Head>Status</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each currentResults as r (r.participant_id)}
							<Table.Row>
								<Table.Cell class="font-mono text-sm">{r.nis}</Table.Cell>
								<Table.Cell class="font-medium">{r.student_nama}</Table.Cell>
								<Table.Cell><Badge variant="secondary" class="text-xs">{r.class_code || '—'}</Badge></Table.Cell>
								<Table.Cell class="text-sm text-slate-600">{r.session_title}</Table.Cell>
								<Table.Cell class="text-center font-bold text-green-700">{fmtScore(r.score)}</Table.Cell>
								<Table.Cell>
									{#if r.submitted_at}
										<Badge variant="outline" class="bg-green-50 text-green-700 border-green-200">Selesai</Badge>
									{:else}
										<Badge variant="outline" class="text-slate-400 border-slate-200">Belum</Badge>
									{/if}
								</Table.Cell>
							</Table.Row>
						{:else}
							<Table.Row>
								<Table.Cell colspan={6} class="py-12 text-center text-slate-400">
									Belum ada data nilai untuk kegiatan ini.
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</Card.Content>
		</Card.Root>
		{/snippet}
	</AsyncContent>
</div>
