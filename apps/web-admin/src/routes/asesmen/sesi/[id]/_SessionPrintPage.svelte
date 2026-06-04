<script lang="ts">
	import { onMount } from 'svelte';
	import { normalizeSesiCbtRow, sessionStatusLabel, type SesiCbtRow } from '$lib/asesmen/sesi-cbt';

	type PrintKind = 'absen' | 'ba' | 'kartu';
	type Props = { id: string; kind: PrintKind };
	type DetailPayload = { data?: unknown; items?: unknown[]; error?: string };

	let { id, kind }: Props = $props();
	let row = $state<SesiCbtRow | null>(null);
	let error = $state('');
	let loading = $state(true);

	const title = $derived(kind === 'absen' ? 'Daftar Hadir Sesi CBT' : kind === 'ba' ? 'Berita Acara Sesi CBT' : 'Kartu Peserta Sesi CBT');

	function unwrap(payload: unknown): unknown {
		if (payload && typeof payload === 'object' && 'data' in payload) return (payload as DetailPayload).data;
		return payload;
	}

	async function loadSession() {
		loading = true;
		error = '';
		try {
			const detail = await fetch(`/api/asesmen/sessions/${encodeURIComponent(id)}`);
			const detailPayload = await detail.json().catch(() => ({}));
			if (detail.ok) {
				row = normalizeSesiCbtRow(unwrap(detailPayload));
				return;
			}
			const list = await fetch('/api/asesmen/sessions?limit=500&include_archived=1');
			const listPayload = await list.json().catch(() => ({}));
			if (!list.ok) throw new Error((listPayload as DetailPayload).error ?? `HTTP ${list.status}`);
			const data = unwrap(listPayload) as DetailPayload;
			const item = Array.isArray(data?.items) ? data.items.find((candidate) => normalizeSesiCbtRow(candidate).id === id) : undefined;
			row = normalizeSesiCbtRow(item ?? { id, title: 'Sesi CBT' });
		} catch (err) {
			error = err instanceof Error ? err.message : 'Detail sesi belum dapat dimuat.';
			row = normalizeSesiCbtRow({ id, title: 'Sesi CBT' });
		} finally {
			loading = false;
		}
	}

	function formatDateTime(value?: string) {
		if (!value) return 'Belum dijadwalkan';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value;
		return new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium', timeStyle: 'short', timeZone: 'Asia/Makassar' }).format(date) + ' WITA';
	}

	onMount(() => {
		void loadSession();
	});
</script>

<svelte:head><title>{title} · MTsN 2 Kolaka Utara</title></svelte:head>

<div class="min-h-screen bg-white p-6 text-slate-950 print:p-0">
	<div class="mx-auto max-w-4xl border border-slate-300 p-6 print:border-0">
		<header class="border-b border-slate-300 pb-4 text-center">
			<p class="text-sm font-semibold uppercase">MTs Negeri 2 Kolaka Utara</p>
			<h1 class="mt-2 text-2xl font-bold">{title}</h1>
		</header>

		{#if loading}
			<p class="mt-6 text-sm text-slate-600">Memuat detail sesi...</p>
		{:else}
			{#if error}<p class="mt-4 rounded border border-amber-300 bg-amber-50 p-2 text-sm text-amber-800">{error}</p>{/if}
			{#if row}
				<section class="mt-5 grid gap-2 text-sm sm:grid-cols-2">
					<p><strong>Sesi:</strong> {row.subject_name || row.title}</p>
					<p><strong>Status:</strong> {sessionStatusLabel(row.status)}</p>
					<p><strong>Kegiatan:</strong> {row.exam_title}</p>
					<p><strong>Jadwal:</strong> {formatDateTime(row.scheduled_start)}</p>
					<p><strong>Paket:</strong> {row.package_title || '-'}</p>
					<p><strong>Rombel:</strong> {row.class_code || row.class_name || '-'}</p>
					<p><strong>Ruang:</strong> {row.room_labels.join(', ') || `${row.room_count} ruang`}</p>
					<p><strong>Peserta:</strong> {row.participant_count}</p>
				</section>

				{#if kind === 'absen'}
					<table class="mt-6 w-full border-collapse text-sm">
						<thead><tr><th class="border p-2 text-left">No</th><th class="border p-2 text-left">Nama Peserta</th><th class="border p-2 text-left">Ruang/Kursi</th><th class="border p-2 text-left">Tanda Tangan</th></tr></thead>
						<tbody>{#each Array.from({ length: Math.max(10, row.participant_count || 10) }) as _, index}<tr><td class="border p-2">{index + 1}</td><td class="border p-2"></td><td class="border p-2"></td><td class="border p-2"></td></tr>{/each}</tbody>
					</table>
				{:else if kind === 'ba'}
					<div class="mt-6 space-y-4 text-sm leading-7">
						<p>Pada hari ini, pelaksanaan sesi CBT di atas dinyatakan berlangsung sesuai jadwal operasional sekolah.</p>
						<div class="grid grid-cols-2 gap-8 pt-8 text-center"><p>Pengawas Ruang<br /><br /><br />(........................)</p><p>Operator<br /><br /><br />(........................)</p></div>
					</div>
				{:else}
					<div class="mt-6 grid gap-3 sm:grid-cols-2">
						{#each Array.from({ length: Math.max(4, Math.min(row.participant_count || 4, 12)) }) as _, index}
							<div class="border border-slate-300 p-3 text-sm">
								<p class="font-bold">Kartu Peserta #{index + 1}</p>
								<p class="mt-2">Sesi: {row.subject_name || row.title}</p>
								<p>Jadwal: {formatDateTime(row.scheduled_start)}</p>
								<p>Ruang/Kursi: __________________</p>
							</div>
						{/each}
					</div>
				{/if}
			{/if}
		{/if}

		<div class="mt-6 flex justify-end print:hidden">
			<button type="button" class="rounded-md border px-3 py-2 text-sm font-semibold" onclick={() => window.print()}>Cetak</button>
		</div>
	</div>
</div>
