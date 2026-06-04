<script lang="ts">
	import { onMount } from 'svelte';
	import { makassarDateKey, normalizeSesiCbtRow, sessionIsActiveWindow, sessionMonitorHref, sessionStatusLabel, sessionStatusTone, summarizeSessionRows, type SesiCbtRow } from '$lib/asesmen/sesi-cbt';

	let rows = $state<SesiCbtRow[]>([]);
	let loading = $state(true);
	let error = $state('');
	let includeArchived = $state(false);
	let quickFilter = $state<'today' | 'active' | 'all'>('today');
	let searchFilter = $state('');

	const todayKey = $derived(makassarDateKey(new Date()));
	const visibleRows = $derived(rows.filter((row) => {
		const keyword = searchFilter.trim().toLowerCase();
		if (!includeArchived && row.status === 'cancelled') return false;
		if (quickFilter === 'today' && makassarDateKey(row.scheduled_start) !== todayKey) return false;
		if (quickFilter === 'active' && row.status !== 'active' && !sessionIsActiveWindow(row)) return false;
		if (keyword) {
			const text = [row.exam_title, row.title, row.subject_name, row.package_title, row.class_code, row.class_name, row.room_labels.join(' ')]
				.filter(Boolean)
				.join(' ')
				.toLowerCase();
			if (!text.includes(keyword)) return false;
		}
		return true;
	}));
	const summary = $derived(summarizeSessionRows(visibleRows));

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
	<section class="mt-3 rounded-3xl border border-white/10 bg-white/10 p-3">
		<div class="grid grid-cols-3 gap-2 text-xs font-bold">
			<button type="button" class={`rounded-2xl px-3 py-2 ${quickFilter === 'today' ? 'bg-emerald-400 text-slate-950' : 'bg-white/10 text-slate-200'}`} onclick={() => (quickFilter = 'today')}>Hari ini</button>
			<button type="button" class={`rounded-2xl px-3 py-2 ${quickFilter === 'active' ? 'bg-emerald-400 text-slate-950' : 'bg-white/10 text-slate-200'}`} onclick={() => (quickFilter = 'active')}>Berlangsung/Aktif</button>
			<button type="button" class={`rounded-2xl px-3 py-2 ${quickFilter === 'all' ? 'bg-emerald-400 text-slate-950' : 'bg-white/10 text-slate-200'}`} onclick={() => (quickFilter = 'all')}>Semua</button>
		</div>
		<input class="mt-3 w-full rounded-2xl border border-white/10 bg-slate-950/50 px-3 py-3 text-sm text-slate-100 placeholder:text-slate-500" placeholder="Cari sesi, mapel, rombel, ruang" bind:value={searchFilter} />
	</section>

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
				{@const monitorHref = sessionMonitorHref(row)}
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
					{#if monitorHref}
						<a class="mt-3 block rounded-2xl bg-emerald-400 px-4 py-3 text-center text-sm font-black text-slate-950" href={monitorHref}>Monitor</a>
					{/if}
				</article>
			{/each}
		{/if}
	</main>
</div>
