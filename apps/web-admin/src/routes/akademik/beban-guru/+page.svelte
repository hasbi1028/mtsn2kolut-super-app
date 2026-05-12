<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { readClientApiData } from '$lib/client/api';

	type WorkloadAssignment = { class_code: string; class_name: string; subject_name: string; expected_weekly_hours: number };
	type WorkloadRow = { teacher_employee_id: string; teacher_name: string; teacher_nip: string; total_weekly_hours: number; assignment_count: number; status: 'kurang' | 'cukup' | 'lebih' | string; status_label: string; assignments: WorkloadAssignment[] };
	type WorkloadPayload = { active_academic_year_name: string; items: WorkloadRow[] };

	let payload = $state<WorkloadPayload | null>(null);
	let loading = $state(true);
	let statusFilter = $state('all');
	let query = $state('');

	const rows = $derived.by(() => {
		const q = query.trim().toLowerCase();
		return (payload?.items ?? []).filter((item) => {
			if (statusFilter !== 'all' && item.status !== statusFilter) return false;
			if (!q) return true;
			return [item.teacher_name, item.teacher_nip, item.status_label].some((value) => (value ?? '').toLowerCase().includes(q));
		});
	});
	const summary = $derived.by(() => ({ total: payload?.items.length ?? 0, kurang: (payload?.items ?? []).filter((item) => item.status === 'kurang').length, cukup: (payload?.items ?? []).filter((item) => item.status === 'cukup').length, lebih: (payload?.items ?? []).filter((item) => item.status === 'lebih').length }));

	async function load() {
		loading = true;
		try {
			const response = await fetch('/api/academic/teacher-workload');
			payload = await readClientApiData<WorkloadPayload>(response, 'Gagal memuat beban mengajar');
		} finally {
			loading = false;
		}
	}

	function downloadCsv() {
		const lines = [['Guru', 'NIP', 'Jumlah JP', 'Jumlah Penugasan', 'Status'], ...rows.map((item) => [item.teacher_name, item.teacher_nip, String(item.total_weekly_hours), String(item.assignment_count), item.status_label])];
		const csv = lines.map((line) => line.map((cell) => `"${String(cell).replaceAll('"', '""')}"`).join(',')).join('\n');
		const url = URL.createObjectURL(new Blob([csv], { type: 'text/csv;charset=utf-8' }));
		const a = document.createElement('a');
		a.href = url;
		a.download = 'beban-guru.csv';
		a.click();
		URL.revokeObjectURL(url);
	}

	onMount(load);
</script>

<svelte:head><title>Beban Guru — Akademik</title></svelte:head>

<div class="space-y-6 p-4 md:p-6">
	<div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
		<div><p class="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">Akademik</p><h1 class="text-2xl font-semibold text-foreground">Beban Guru</h1><p class="text-sm text-muted-foreground">Pantau kecukupan JP mengajar berdasarkan penugasan guru mapel.</p></div>
		<div class="flex gap-2"><Badge variant="outline">{payload?.active_academic_year_name || 'Tahun ajaran aktif'}</Badge><Button variant="outline" onclick={downloadCsv}>Unduh data</Button></div>
	</div>
	<div class="grid gap-3 md:grid-cols-4">
		<Card.Root><Card.Content class="p-4"><p class="text-xs text-muted-foreground">Total guru</p><p class="text-2xl font-semibold">{summary.total}</p></Card.Content></Card.Root>
		<Card.Root><Card.Content class="p-4"><p class="text-xs text-muted-foreground">Kurang 24 JP</p><p class="text-2xl font-semibold">{summary.kurang}</p></Card.Content></Card.Root>
		<Card.Root><Card.Content class="p-4"><p class="text-xs text-muted-foreground">Cukup</p><p class="text-2xl font-semibold">{summary.cukup}</p></Card.Content></Card.Root>
		<Card.Root><Card.Content class="p-4"><p class="text-xs text-muted-foreground">Perlu peninjauan</p><p class="text-2xl font-semibold">{summary.lebih}</p></Card.Content></Card.Root>
	</div>
	<Card.Root>
		<Card.Header><Card.Title class="text-base">Daftar Beban Mengajar</Card.Title><Card.Description>Gunakan filter untuk menemukan guru yang perlu penyesuaian penugasan.</Card.Description></Card.Header>
		<Card.Content class="space-y-4">
			<div class="grid gap-3 md:grid-cols-3"><Input bind:value={query} placeholder="Cari guru atau NIP" /><select bind:value={statusFilter} class="h-10 rounded-md border border-input bg-background px-3 text-sm text-foreground"><option value="all">Semua status</option><option value="kurang">Kurang 24 JP</option><option value="cukup">Cukup</option><option value="lebih">Perlu peninjauan</option></select><Button variant="outline" onclick={load}>Muat ulang</Button></div>
			{#if loading}<p class="text-sm text-muted-foreground">Memuat beban guru...</p>{:else if rows.length === 0}<p class="text-sm text-muted-foreground">Data belum tersedia.</p>{:else}
				<Table.Root><Table.Header><Table.Row><Table.Head>Guru</Table.Head><Table.Head>JP</Table.Head><Table.Head>Penugasan</Table.Head><Table.Head>Status</Table.Head><Table.Head>Rincian</Table.Head></Table.Row></Table.Header><Table.Body>{#each rows as item}<Table.Row><Table.Cell><div class="font-medium">{item.teacher_name}</div><div class="text-xs text-muted-foreground">{item.teacher_nip || '-'}</div></Table.Cell><Table.Cell>{item.total_weekly_hours}</Table.Cell><Table.Cell>{item.assignment_count}</Table.Cell><Table.Cell><Badge variant={item.status === 'cukup' ? 'default' : 'outline'}>{item.status_label}</Badge></Table.Cell><Table.Cell class="text-xs text-muted-foreground">{item.assignments.slice(0, 3).map((a) => `${a.class_code} · ${a.subject_name} (${a.expected_weekly_hours} JP)`).join(', ')}{item.assignments.length > 3 ? ' ...' : ''}</Table.Cell></Table.Row>{/each}</Table.Body></Table.Root>
			{/if}
		</Card.Content>
	</Card.Root>
</div>
