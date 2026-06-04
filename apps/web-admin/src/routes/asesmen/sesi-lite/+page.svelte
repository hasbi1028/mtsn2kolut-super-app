<script lang="ts">
	import { onMount } from 'svelte';
	import { normalizeSesiCbtRow, sessionStatusLabel, sessionStatusTone, summarizeSessionRows, type SesiCbtRow } from '$lib/asesmen/sesi-cbt';

	let rows = $state<SesiCbtRow[]>([]);
	let loading = $state(true);
	let error = $state('');
	let includeArchived = $state(false);

	const summary = $derived(summarizeSessionRows(rows.filter((row) => includeArchived || row.status !== 'cancelled')));
	const visibleRows = $derived(rows.filter((row) => includeArchived || row.status !== 'cancelled'));

	function unwrap(payload: unknown): { items?: unknown[]; error?: string } {
		if (payload && typeof payload === 'object' && 'data' in payload) return (payload as { data: { items?: unknown[]; error?: string } }).data ?? {};
		return (payload ?? {}) as { items?: unknown[]; error?: string };
	}

	async function loadSessions() {
		loading = true;
		error = '';
		try {
			const params = new URLSearchParams({ limit: '200' });
			if (includeArchived) params.set('include_archived', '1');
			const response = await fetch(`/api/asesmen/sessions?${params}`);
			const payload = unwrap(await response.json().catch(() => ({})));
			if (!response.ok) throw new Error(payload.error ?? `HTTP ${response.status}`);
			rows = Array.isArray(payload.items) ? payload.items.map((item) => normalizeSesiCbtRow(item)) : [];
		} catch (err) {
			error = err instanceof Error ? err.message : 'Sesi operasional belum dapat dimuat.';
			rows = [];
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
		void loadSessions();
	});
</script>

<svelte:head><title>Sesi Operasional · Mode Cepat HP</title></svelte:head>

<div class="min-h-screen bg-slate-950 p-3 text-slate-100 sm:p-4">
	<header class="rounded-3xl border border-white/10 bg-white/10 p-4 shadow-xl backdrop-blur">
		<div class="flex items-start justify-between gap-3">
			<div>
				<p class="text-xs font-semibold tracking-[0.2em] text-emerald-300 uppercase">Mode cepat HP</p>
				<h1 class="mt-1 text-2xl font-black">Sesi Operasional</h1>
				<p class="mt-1 text-sm text-slate-300">Halaman ringan untuk operator/pengawas saat ujian berlangsung.</p>
			</div>
			<button class="rounded-xl bg-emerald-400 px-3 py-2 text-xs font-bold text-slate-950 disabled:opacity-60" type="button" onclick={() => void loadSessions()} disabled={loading}>{loading ? '...' : 'Refresh'}</button>
		</div>
		<div class="mt-4 grid grid-cols-5 gap-2 text-center text-xs">
			<div class="rounded-2xl bg-white/10 p-2"><strong class="block text-lg">{summary.total}</strong><span>Total</span></div>
			<div class="rounded-2xl bg-white/10 p-2"><strong class="block text-lg">{summary.active}</strong><span>Aktif</span></div>
			<div class="rounded-2xl bg-white/10 p-2"><strong class="block text-lg">{summary.running}</strong><span>Jalan</span></div>
			<div class="rounded-2xl bg-white/10 p-2"><strong class="block text-lg">{summary.finished}</strong><span>Selesai</span></div>
			<div class="rounded-2xl bg-white/10 p-2"><strong class="block text-lg">{summary.incident}</strong><span>Insiden</span></div>
		</div>
	</header>

	<div class="mt-3 flex items-center justify-between gap-2">
		<a class="rounded-xl border border-white/10 bg-white/10 px-3 py-2 text-xs font-bold text-slate-100" href="/asesmen/sesi">Sesi CBT Lengkap</a>
		<label class="flex items-center gap-2 text-xs text-slate-300"><input type="checkbox" bind:checked={includeArchived} onchange={() => void loadSessions()} /> Tampilkan batal</label>
	</div>

	{#if error}<p class="mt-3 rounded-2xl border border-red-400/30 bg-red-400/10 p-3 text-sm text-red-100">{error}</p>{/if}

	<main class="mt-3 space-y-3">
		{#if loading && visibleRows.length === 0}
			<p class="rounded-2xl border border-white/10 bg-white/10 p-4 text-sm text-slate-300">Memuat sesi…</p>
		{:else if visibleRows.length === 0}
			<div class="rounded-2xl border border-dashed border-white/20 bg-white/10 p-5 text-center">
				<p class="font-bold">Belum ada sesi operasional</p>
				<p class="mt-1 text-sm text-slate-300">Jadwalkan sesi dari halaman Sesi CBT lengkap.</p>
			</div>
		{:else}
			{#each visibleRows as row (row.id)}
				<article class="rounded-3xl border border-white/10 bg-white/10 p-4 shadow-lg">
					<div class="flex items-start justify-between gap-3">
						<div class="min-w-0">
							<p class="truncate text-lg font-black">{row.subject_name || row.title}</p>
							<p class="mt-1 text-sm text-slate-300">{formatDateTime(row.scheduled_start)}</p>
						</div>
						<span class={`rounded-full border px-2.5 py-1 text-[11px] font-bold ${sessionStatusTone(row.status)}`}>{sessionStatusLabel(row.status)}</span>
					</div>
					<p class="mt-2 text-sm text-slate-300">{row.exam_title}</p>
					<p class="text-xs text-slate-400">{row.package_title || 'Paket'} · {row.class_code || row.class_name || 'Rombel'} · {row.room_labels.join(', ') || `${row.room_count} ruang`}</p>
					<div class="mt-3 grid grid-cols-4 gap-2 text-center text-xs">
						<div class="rounded-2xl bg-slate-950/40 p-2"><strong class="block text-base">{row.participant_count}</strong><span>Peserta</span></div>
						<div class="rounded-2xl bg-slate-950/40 p-2"><strong class="block text-base">{row.working_count}</strong><span>Kerja</span></div>
						<div class="rounded-2xl bg-slate-950/40 p-2"><strong class="block text-base">{row.finished_count}</strong><span>Selesai</span></div>
						<div class="rounded-2xl bg-slate-950/40 p-2"><strong class="block text-base">{row.incident_count}</strong><span>Insiden</span></div>
					</div>
					<a class="mt-3 block rounded-2xl bg-emerald-400 px-4 py-3 text-center text-sm font-black text-slate-950" href="/asesmen/pelaksanaan">Buka monitor pengawas</a>
				</article>
			{/each}
		{/if}
	</main>
</div>
