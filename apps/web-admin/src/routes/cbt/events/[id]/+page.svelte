<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';

	type EventInfo = {
		id: string; title: string; exam_type: string; scope: string;
		academic_year_name: string; status: string;
	};
	type ResultRow = {
		participant_id: string; session_id: string; session_title: string;
		nis: string; student_nama: string; gender: string;
		class_code: string; score: string | null; submitted_at: string | null;
	};

	const eventId = page.params.id;
	let info = $state<EventInfo | null>(null);
	let results = $state<ResultRow[]>([]);
	let loading = $state(true);
	let error = $state('');

	async function load() {
		try {
			const [iRes, rRes] = await Promise.all([
				fetch(`/api/cbt/events/${eventId}`),
				fetch(`/api/cbt/events/${eventId}/results`),
			]);
			const iJson = await iRes.json();
			const rJson = await rRes.json();
			info = iJson;
			results = rJson.data ?? rJson ?? [];
		} catch {
			error = 'Gagal memuat rekap nilai event';
		} finally {
			loading = false;
		}
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

	onMount(load);
</script>

<svelte:head><title>Rekap Nilai — {info?.title ?? 'Event'}</title></svelte:head>

<div class="space-y-6 p-6">
	<div class="flex items-center gap-2 text-sm text-slate-500">
		<a href="/cbt/events" class="hover:text-slate-700">Kegiatan Ujian</a>
		<span>/</span>
		<span class="text-slate-700 font-medium truncate max-w-xs">{info?.title ?? '...'}</span>
	</div>

	{#if loading}
		<p class="text-sm text-slate-500">Memuat rekap nilai...</p>
	{:else if info}
		<div class="flex items-start justify-between gap-4 flex-wrap">
			<div>
				<h1 class="text-2xl font-semibold text-slate-800">{info.title}</h1>
				<p class="text-sm text-slate-500 mt-1">
					Tahun Ajaran: {info.academic_year_name} · Tipe: <span class="capitalize">{info.exam_type}</span>
				</p>
			</div>
			<Button variant="outline" onclick={exportCSV} disabled={results.length === 0}>
				↓ Export CSV (Rekap Semua Sesi)
			</Button>
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
						{#each results as r}
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
	{/if}
</div>
