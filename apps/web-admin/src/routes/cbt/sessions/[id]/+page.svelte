<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';

	type SessionInfo = {
		id: string; title: string; package_title: string; duration_minutes: number;
		class_name: string; class_code: string;
		scheduled_start: string; scheduled_end: string; status: string;
	};
	type ResultRow = {
		participant_id: string; student_id: string;
		nis: string; nama: string; gender: string;
		submitted_at: string | null; score: string | null;
		total_answers: number; correct_answers: number;
	};

	const sessionId = page.params.id;

	let session = $state<SessionInfo | null>(null);
	let results = $state<ResultRow[]>([]);
	let loading = $state(true);
	let error = $state('');
	let toast = $state({ msg: '', ok: true });
	let scoreBusy = $state(false);

	const statusLabel: Record<string, string> = {
		draft: 'Draft', scheduled: 'Terjadwal', active: 'Berlangsung',
		finished: 'Selesai', cancelled: 'Dibatalkan',
	};

	function statusClass(s: string) {
		if (s === 'active') return 'bg-emerald-100 text-emerald-700 border-emerald-200';
		if (s === 'finished') return 'bg-slate-100 text-slate-500 border-slate-200';
		if (s === 'cancelled') return 'bg-red-100 text-red-700 border-red-200';
		if (s === 'scheduled') return 'bg-blue-100 text-blue-700 border-blue-200';
		return 'bg-amber-100 text-amber-700 border-amber-200';
	}

	function fmtDt(iso: string | null) {
		if (!iso) return '—';
		return new Date(iso).toLocaleString('id-ID', {
			timeZone: 'Asia/Makassar', year: 'numeric', month: 'short',
			day: 'numeric', hour: '2-digit', minute: '2-digit',
		}) + ' WITA';
	}

	function fmtScore(score: string | null) {
		if (score === null || score === undefined || score === '') return '—';
		const n = parseFloat(score);
		return isNaN(n) ? '—' : n.toFixed(1);
	}

	function scoreClass(score: string | null) {
		if (!score) return 'text-slate-400';
		const n = parseFloat(score);
		if (n >= 75) return 'text-emerald-600 font-semibold';
		if (n >= 60) return 'text-amber-600 font-semibold';
		return 'text-red-600 font-semibold';
	}

	let stats = $derived({
		total: results.length,
		submitted: results.filter(r => r.submitted_at).length,
		avgScore: results.length > 0
			? results.reduce((sum, r) => sum + (r.score ? parseFloat(r.score) : 0), 0) / results.length
			: 0,
		passing: results.filter(r => r.score && parseFloat(r.score) >= 75).length,
	});

	async function load() {
		try {
			const res = await fetch(`/api/cbt/sessions/${sessionId}/results`);
			const json = await res.json();
			if (json.error) { error = json.error; return; }
			const d = json.data ?? json;
			session = d.session ?? null;
			results = d.results ?? [];
		} catch {
			error = 'Gagal memuat hasil ujian';
		} finally {
			loading = false;
		}
	}

	function showToast(msg: string, ok = true) {
		toast = { msg, ok };
		setTimeout(() => (toast = { msg: '', ok: true }), 4000);
	}

	async function triggerScoring() {
		if (!confirm('Hitung ulang skor semua peserta berdasarkan jawaban yang masuk?')) return;
		scoreBusy = true;
		try {
			const res = await fetch(`/api/cbt/sessions/${sessionId}/score`, { method: 'POST' });
			if (!res.ok) { const j = await res.json(); showToast(j.error ?? 'Gagal', false); return; }
			showToast('Penilaian selesai — skor diperbarui');
			await load();
		} finally { scoreBusy = false; }
	}

	function exportCSV() {
		if (!session || results.length === 0) return;
		const header = 'NIS,Nama,L/P,Jawaban Masuk,Benar,Skor,Waktu Submit';
		const rows = results.map(r =>
			[r.nis, `"${r.nama}"`, r.gender, r.total_answers, r.correct_answers,
			 fmtScore(r.score), r.submitted_at ? fmtDt(r.submitted_at) : ''].join(',')
		);
		const csv = [header, ...rows].join('\n');
		const blob = new Blob([csv], { type: 'text/csv' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `hasil_${session.title.replace(/\s+/g, '_')}.csv`;
		a.click();
		URL.revokeObjectURL(url);
	}

	onMount(load);
</script>

<svelte:head>
	<title>{session?.title ?? 'Hasil Ujian'} — MTSN 2 Kolut</title>
</svelte:head>

<div class="space-y-6 p-6 max-w-5xl mx-auto">
	<!-- Breadcrumb -->
	<div class="flex items-center gap-2 text-sm text-slate-500">
		<a href="/cbt/sessions" class="hover:text-slate-700">Sesi Ujian</a>
		<span>/</span>
		<span class="text-slate-700 font-medium truncate max-w-xs">{session?.title ?? '...'}</span>
	</div>

	{#if toast.msg}
		<div class="rounded-md px-4 py-3 text-sm {toast.ok ? 'bg-emerald-50 border border-emerald-200 text-emerald-800' : 'bg-red-50 border border-red-200 text-red-800'}">
			{toast.msg}
		</div>
	{/if}
	{#if error}
		<div class="rounded-md bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-800">{error}</div>
	{/if}

	{#if loading}
		<p class="text-sm text-slate-500">Memuat data...</p>
	{:else if session}
		<!-- Session header -->
		<div class="flex items-start justify-between gap-4 flex-wrap">
			<div>
				<h1 class="text-2xl font-semibold text-slate-800">{session.title}</h1>
				<div class="flex flex-wrap gap-3 mt-2 text-sm text-slate-500">
					<span>{session.package_title}</span>
					<span>·</span>
					<span>Kelas {session.class_code}</span>
					<span>·</span>
					<span>{session.duration_minutes} menit</span>
					<span>·</span>
					<span>{fmtDt(session.scheduled_start)}</span>
				</div>
			</div>
			<div class="flex items-center gap-2">
				<Badge class={statusClass(session.status)}>{statusLabel[session.status] ?? session.status}</Badge>
				{#if session.status === 'finished' || session.status === 'active'}
					<Button size="sm" variant="outline" disabled={scoreBusy} onclick={triggerScoring}>
						{scoreBusy ? 'Menghitung...' : '⟳ Hitung Skor'}
					</Button>
				{/if}
				{#if results.length > 0}
					<Button size="sm" variant="outline" onclick={exportCSV}>
						↓ Ekspor CSV
					</Button>
				{/if}
			</div>
		</div>

		<!-- Stats cards -->
		<div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
			<Card.Root>
				<Card.Content class="pt-4 pb-3 px-4">
					<p class="text-xs text-slate-500 mb-1">Total Peserta</p>
					<p class="text-2xl font-bold text-slate-800">{stats.total}</p>
				</Card.Content>
			</Card.Root>
			<Card.Root>
				<Card.Content class="pt-4 pb-3 px-4">
					<p class="text-xs text-slate-500 mb-1">Sudah Submit</p>
					<p class="text-2xl font-bold text-slate-800">{stats.submitted}</p>
				</Card.Content>
			</Card.Root>
			<Card.Root>
				<Card.Content class="pt-4 pb-3 px-4">
					<p class="text-xs text-slate-500 mb-1">Rata-rata Skor</p>
					<p class="text-2xl font-bold {stats.avgScore >= 75 ? 'text-emerald-600' : stats.avgScore >= 60 ? 'text-amber-600' : 'text-red-600'}">
						{stats.total > 0 ? stats.avgScore.toFixed(1) : '—'}
					</p>
				</Card.Content>
			</Card.Root>
			<Card.Root>
				<Card.Content class="pt-4 pb-3 px-4">
					<p class="text-xs text-slate-500 mb-1">Lulus (≥75)</p>
					<p class="text-2xl font-bold text-slate-800">{stats.passing}
						{#if stats.submitted > 0}
							<span class="text-sm font-normal text-slate-400">/ {stats.submitted}</span>
						{/if}
					</p>
				</Card.Content>
			</Card.Root>
		</div>

		<!-- Results table -->
		<Card.Root>
			<Card.Header class="pb-2">
				<Card.Title class="text-base">Daftar Nilai ({results.length} peserta)</Card.Title>
			</Card.Header>
			<Card.Content class="p-0">
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head class="w-8">#</Table.Head>
							<Table.Head>NIS</Table.Head>
							<Table.Head>Nama</Table.Head>
							<Table.Head>L/P</Table.Head>
							<Table.Head class="text-center">Jawaban</Table.Head>
							<Table.Head class="text-center">Benar</Table.Head>
							<Table.Head class="text-center">Skor</Table.Head>
							<Table.Head>Waktu Submit</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each results as r, i}
							<Table.Row>
								<Table.Cell class="text-slate-400 text-xs">{i + 1}</Table.Cell>
								<Table.Cell class="font-mono text-sm">{r.nis}</Table.Cell>
								<Table.Cell class="font-medium">{r.nama}</Table.Cell>
								<Table.Cell>
									<Badge variant="outline" class="text-xs">{r.gender}</Badge>
								</Table.Cell>
								<Table.Cell class="text-center text-sm">{r.total_answers}</Table.Cell>
								<Table.Cell class="text-center text-sm">{r.correct_answers}</Table.Cell>
								<Table.Cell class="text-center">
									<span class={scoreClass(r.score)}>{fmtScore(r.score)}</span>
								</Table.Cell>
								<Table.Cell class="text-slate-500 text-xs whitespace-nowrap">
									{#if r.submitted_at}
										{fmtDt(r.submitted_at)}
									{:else}
										<span class="text-slate-300">Belum submit</span>
									{/if}
								</Table.Cell>
							</Table.Row>
						{:else}
							<Table.Row>
								<Table.Cell colspan={8} class="text-center text-slate-400 py-10">
									{#if session.status === 'draft' || session.status === 'scheduled'}
										Sesi belum dimulai — peserta belum mengerjakan ujian
									{:else}
										Belum ada data hasil ujian
									{/if}
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</Card.Content>
		</Card.Root>
	{/if}
</div>
