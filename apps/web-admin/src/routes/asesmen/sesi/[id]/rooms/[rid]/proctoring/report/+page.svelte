<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { toast } from '$lib/components/ui/sonner';
	import { clientApiPath, readClientApiData } from '$lib/client/api';
	import { csvRow } from '$lib/csv';
	import { classifyProctorEvent, proctorEventLabel, proctorEvidenceCategoryLabel } from '$lib/cbt/proctor-evidence';

	type RoomDashboard = { id: string; session_id: string; room_name: string; session_title: string; session_status: string; scheduled_start: string; scheduled_end: string; package_title: string; participant_count: number; submitted_count: number; online_count: number; suspicious_count: number };
	type RoomProctor = { id: string; nama: string; nip: string; role: string };
	type ProctoringRow = { participant_id: string; nis: string; nama: string; room_name: string; risk_level?: string; locked_at?: string | null; last_heartbeat: string | null; violation_count?: number; risk_score?: number };
	type ProctoringEvent = { id: string; participant_id: string; nis: string; nama: string; room_name: string; event_type: string; event_data: unknown; created_at: string };
	type DashboardPayload = { room: RoomDashboard; proctors: RoomProctor[]; participants: ProctoringRow[]; events: ProctoringEvent[] };

	const sessionId = page.params.id ?? '';
	const roomId = page.params.rid ?? '';
	let loading = $state(true);
	let loadError = $state('');
	let room = $state<RoomDashboard | null>(null);
	let proctors = $state<RoomProctor[]>([]);
	let participants = $state<ProctoringRow[]>([]);
	let events = $state<ProctoringEvent[]>([]);
	let participantFilter = $state('');
	let eventTypeFilter = $state('all');
	let riskFilter = $state('all');
	let reviewFilter = $state<'all' | 'reviewed' | 'unreviewed'>('all');

	let eventTypeOptions = $derived.by(() => Array.from(new Set(events.map((event) => event.event_type))).sort());
	let reviewedParticipantIds = $derived.by(() => new Set(events.filter(isReviewEvent).map((event) => event.participant_id)));
	let filteredEvents = $derived.by(() => events.filter((event) => {
		if (participantFilter.trim()) {
			const q = participantFilter.trim().toLowerCase();
			if (!`${event.nama} ${event.nis}`.toLowerCase().includes(q)) return false;
		}
		if (eventTypeFilter !== 'all' && event.event_type !== eventTypeFilter) return false;
		if (riskFilter !== 'all') {
			const row = participants.find((p) => p.participant_id === event.participant_id);
			if ((row?.risk_level ?? 'normal') !== riskFilter) return false;
		}
		if (reviewFilter === 'reviewed' && !reviewedParticipantIds.has(event.participant_id)) return false;
		if (reviewFilter === 'unreviewed' && reviewedParticipantIds.has(event.participant_id)) return false;
		return true;
	}));

	onMount(() => {
		void loadReport();
	});

	async function loadReport() {
		loading = true;
		loadError = '';
		try {
		const res = await fetch(clientApiPath`/api/asesmen/sessions/${sessionId}/rooms/${roomId}/proctoring`);
		const payload = await readClientApiData<DashboardPayload>(res, 'Gagal memuat rekap ruang');
		room = payload.room;
		proctors = payload.proctors ?? [];
		participants = payload.participants ?? [];
		events = payload.events ?? [];
		} catch (error) {
			loadError = error instanceof Error ? error.message : 'Gagal memuat rekap';
		} finally {
			loading = false;
		}
	}

	function isReviewEvent(event: ProctoringEvent) {
		return ['proctor_acknowledge', 'proctor_incident_action'].includes(event.event_type);
	}
	function eventDetail(event: ProctoringEvent) {
		const data = event.event_data && typeof event.event_data === 'object' ? event.event_data as Record<string, unknown> : {};
		const parts = [proctorEventLabel(event)];
		for (const key of ['action', 'status', 'reason', 'notes', 'message', 'command_type', 'actor']) {
			const value = data[key];
			if (typeof value === 'string' && value.trim()) parts.push(`${key}=${value}`);
		}
		return parts.join(' · ');
	}
	function fmtDate(value?: string | null) {
		if (!value) return '-';
		return new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value));
	}
	function riskBadge(row?: ProctoringRow) { return row?.risk_level ?? 'normal'; }
	function riskLabel(value: string) {
		const labels: Record<string, string> = {
			normal: 'Normal',
			warning: 'Perlu perhatian',
			high: 'Bahaya',
			locked: 'Terkunci',
		};
		return labels[value] ?? value;
	}
	function eventTypeLabel(type: string) {
		return proctorEventLabel({ event_type: type, event_data: null } as ProctoringEvent);
	}
	function exportCsv() {
		const rows = [['tanggal', 'ruang', 'nis', 'nama', 'kejadian', 'kategori', 'risiko', 'sudah_diperiksa', 'detail']];
		for (const event of filteredEvents) {
			const row = participants.find((p) => p.participant_id === event.participant_id);
			rows.push([event.created_at, event.room_name, event.nis, event.nama, proctorEventLabel(event), proctorEvidenceCategoryLabel(classifyProctorEvent(event)), riskLabel(riskBadge(row)), reviewedParticipantIds.has(event.participant_id) ? 'ya' : 'belum', eventDetail(event)]);
		}
		const blob = new Blob([rows.map((row) => csvRow(row)).join('\n')], { type: 'text/csv;charset=utf-8' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `rekap_insiden_${room?.room_name ?? roomId}.csv`;
		a.click();
		URL.revokeObjectURL(url);
		toast.success('CSV ruang dibuat');
	}
</script>

<svelte:head><title>Berita Acara Ruang CBT</title></svelte:head>

<div class="mx-auto max-w-7xl space-y-6 p-4 md:p-6 print:max-w-none print:p-0">
	<div class="flex flex-wrap items-start justify-between gap-3 print:hidden">
		<div>
			<p class="text-sm text-muted-foreground">Ruang Pengawasan</p>
			<h1 class="text-2xl font-semibold tracking-tight">Rekap Insiden Ruang</h1>
			<p class="text-sm text-muted-foreground">Berita acara per ruang/lab dengan unduhan CSV.</p>
		</div>
		<div class="flex flex-wrap gap-2">
			<Button variant="outline" href={resolve(`/asesmen/sesi/${sessionId}/rooms/${roomId}/proctoring`)}>Kembali</Button>
			<Button variant="outline" onclick={exportCsv}>Unduh CSV</Button>
			<Button onclick={() => window.print()}>Cetak</Button>
		</div>
	</div>

	{#if loading}
		<Card.Root><Card.Content class="p-6 text-sm text-muted-foreground">Memuat berita acara ruang...</Card.Content></Card.Root>
	{:else if loadError}
		<Card.Root><Card.Content class="space-y-3 p-6"><p class="text-sm text-destructive">{loadError}</p><Button variant="outline" onclick={() => void loadReport()}>Coba lagi</Button></Card.Content></Card.Root>
	{:else}

		<Card.Root class="print:hidden">
			<Card.Header><Card.Title>Filter</Card.Title></Card.Header>
			<Card.Content class="grid gap-3 md:grid-cols-4">
				<input class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Cari nama/NIS" bind:value={participantFilter} />
				<select class="rounded-md border bg-background px-3 py-2 text-sm" bind:value={eventTypeFilter}><option value="all">Semua kejadian</option>{#each eventTypeOptions as type}<option value={type}>{eventTypeLabel(type)}</option>{/each}</select>
				<select class="rounded-md border bg-background px-3 py-2 text-sm" bind:value={riskFilter}><option value="all">Semua risiko</option><option value="normal">Normal</option><option value="warning">Perlu perhatian</option><option value="high">Bahaya</option></select>
				<select class="rounded-md border bg-background px-3 py-2 text-sm" bind:value={reviewFilter}><option value="all">Semua status</option><option value="reviewed">Sudah diperiksa</option><option value="unreviewed">Belum diperiksa</option></select>
			</Card.Content>
		</Card.Root>

		<Card.Root>
			<Card.Header class="print:text-center"><Card.Title>Berita Acara Pengawasan CBT Ruang</Card.Title><Card.Description>{room?.session_title} · {room?.room_name} · dicetak {fmtDate(new Date().toISOString())}</Card.Description></Card.Header>
			<Card.Content class="space-y-4">
				<div class="grid gap-2 text-sm md:grid-cols-3">
					<div><span class="text-muted-foreground">Paket:</span> {room?.package_title}</div>
					<div><span class="text-muted-foreground">Jadwal:</span> {fmtDate(room?.scheduled_start)} – {fmtDate(room?.scheduled_end)}</div>
					<div><span class="text-muted-foreground">Pengawas:</span> {proctors.map((p) => p.nama).join(', ') || '-'}</div>
					<div><span class="text-muted-foreground">Peserta:</span> {participants.length}</div>
					<div><span class="text-muted-foreground">Kirim:</span> {room?.submitted_count ?? 0}</div>
					<div><span class="text-muted-foreground">Kejadian tersaring:</span> {filteredEvents.length}</div>
				</div>
				<Table.Root>
					<Table.Header><Table.Row><Table.Head>Waktu</Table.Head><Table.Head>Peserta</Table.Head><Table.Head>Kejadian</Table.Head><Table.Head>Status</Table.Head><Table.Head>Detail/Tindakan</Table.Head></Table.Row></Table.Header>
					<Table.Body>{#each filteredEvents as event}{@const row = participants.find((p) => p.participant_id === event.participant_id)}<Table.Row><Table.Cell>{fmtDate(event.created_at)}</Table.Cell><Table.Cell><div class="font-medium">{event.nama}</div><div class="text-xs text-muted-foreground">{event.nis}</div></Table.Cell><Table.Cell><Badge variant="outline">{proctorEventLabel(event)}</Badge></Table.Cell><Table.Cell><Badge variant="outline">{reviewedParticipantIds.has(event.participant_id) ? 'sudah diperiksa' : riskLabel(riskBadge(row))}</Badge></Table.Cell><Table.Cell class="max-w-md text-xs">{eventDetail(event)}</Table.Cell></Table.Row>{/each}</Table.Body>
				</Table.Root>
				<div class="grid gap-8 pt-8 text-center text-sm md:grid-cols-3 print:grid-cols-3"><div><p>Pengawas Ruang</p><div class="h-16"></div><p class="border-t pt-2">Nama & Tanda Tangan</p></div><div><p>Operator/Admin Ujian</p><div class="h-16"></div><p class="border-t pt-2">Nama & Tanda Tangan</p></div><div><p>Ketua Panitia</p><div class="h-16"></div><p class="border-t pt-2">Nama & Tanda Tangan</p></div></div>
			</Card.Content>
		</Card.Root>
	{/if}
</div>
