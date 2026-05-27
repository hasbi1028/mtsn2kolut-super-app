<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { toast } from '$lib/components/ui/sonner';
	import { clientApiPath, clientApiPathWithQuery, readClientApiData } from '$lib/client/api';
	import { csvRow } from '$lib/csv';
	import { classifyProctorEvent, proctorEventDetail, proctorEventLabel, proctorEvidenceCategoryLabel } from '$lib/cbt/proctor-evidence';

	type ProctoringRow = {
		participant_id: string;
		nis: string;
		nama: string;
		room_id?: string | null;
		room_name: string;
		risk_level?: string;
		locked_at?: string | null;
		last_heartbeat: string | null;
		violation_count?: number;
		risk_score?: number;
	};
	type ProctoringEvent = {
		id: string;
		participant_id: string;
		nis: string;
		nama: string;
		room_id?: string | null;
		room_name: string;
		event_type: string;
		event_data: unknown;
		created_at: string;
	};

	const sessionId = page.params.id ?? '';
	let loading = $state(true);
	let loadError = $state('');
	let participants = $state<ProctoringRow[]>([]);
	let events = $state<ProctoringEvent[]>([]);
	let roomFilter = $state('all');
	let participantFilter = $state('');
	let eventTypeFilter = $state('all');
	let riskFilter = $state('all');
	let reviewFilter = $state<'all' | 'reviewed' | 'unreviewed'>('all');

	let roomOptions = $derived.by(() => Array.from(new Map(participants.map((p) => [p.room_id ?? p.room_name, p.room_name || 'Tanpa ruang'])).entries()).map(([id, name]) => ({ id, name })));
	let eventTypeOptions = $derived.by(() => Array.from(new Set(events.map((event) => event.event_type))).sort());
	let reviewedParticipantIds = $derived.by(() => new Set(events.filter(isReviewEvent).map((event) => event.participant_id)));
	let filteredEvents = $derived.by(() => events.filter((event) => {
		if (roomFilter !== 'all' && (event.room_id ?? event.room_name) !== roomFilter) return false;
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
	let incidentCount = $derived(filteredEvents.filter(isIncidentEvent).length);
	let reportParticipants = $derived.by(() => participants.filter((participant) => {
		if (roomFilter === 'all') return true;
		return (participant.room_id ?? participant.room_name) === roomFilter;
	}));
	let submittedParticipantIds = $derived.by(() => new Set(events.filter(isSubmitEvent).map((event) => event.participant_id)));
	let forceSubmittedParticipantIds = $derived.by(() => new Set(events.filter((event) => event.event_type === 'proctor_force_submit').map((event) => event.participant_id)));
	let attendanceSummary = $derived.by(() => {
		const total = reportParticipants.length;
		const submitted = reportParticipants.filter((participant) => submittedParticipantIds.has(participant.participant_id) || forceSubmittedParticipantIds.has(participant.participant_id)).length;
		const locked = reportParticipants.filter((participant) => participant.locked_at || riskBadge(participant) === 'locked').length;
		const reviewed = reportParticipants.filter((participant) => reviewedParticipantIds.has(participant.participant_id)).length;
		const highRisk = reportParticipants.filter((participant) => riskBadge(participant) === 'high').length;
		const warning = reportParticipants.filter((participant) => riskBadge(participant) === 'warning').length;
		return { total, submitted, notSubmitted: Math.max(total - submitted, 0), locked, reviewed, highRisk, warning };
	});
	let technicalEvents = $derived(filteredEvents.filter(isTechnicalIncident));
	let suspectedMisconductEvents = $derived(filteredEvents.filter(isSuspectedMisconduct));
	let proctorActionEvents = $derived(filteredEvents.filter(isProctorActionEvent));
	let accessActionEvents = $derived(filteredEvents.filter(isAccessActionEvent));

	onMount(() => {
		void loadReport();
	});

	async function loadReport() {
		loading = true;
		loadError = '';
		try {
		const [participantRes, eventRes] = await Promise.all([
			fetch(clientApiPath`/api/asesmen/sessions/${sessionId}/proctoring`),
			fetch(clientApiPathWithQuery(`/api/asesmen/sessions/${sessionId}/proctoring/events`, new URLSearchParams({ limit: '500' })))
		]);
		participants = await readClientApiData<ProctoringRow[]>(participantRes, 'Gagal memuat peserta pengawasan');
		events = await readClientApiData<ProctoringEvent[]>(eventRes, 'Gagal memuat riwayat pengawasan');
		} catch (error) {
			loadError = error instanceof Error ? error.message : 'Gagal memuat rekap';
		} finally {
			loading = false;
		}
	}

	function isIncidentEvent(event: ProctoringEvent) {
		return ['anti_cheat_violation', 'app_switch', 'screenshot_attempt', 'proctor_force_submit', 'proctor_reset_access', 'proctor_unlock', 'proctor_acknowledge', 'proctor_incident_action', 'participant_command'].includes(event.event_type);
	}

	function isReviewEvent(event: ProctoringEvent) {
		return ['proctor_acknowledge', 'proctor_incident_action'].includes(event.event_type);
	}

	function isSubmitEvent(event: ProctoringEvent) {
		return ['submit', 'manual_submit', 'auto_submit'].includes(event.event_type);
	}

	function isTechnicalIncident(event: ProctoringEvent) {
		const category = classifyProctorEvent(event);
		return ['stale_connection', 'submit_guard', 'heartbeat'].includes(category ?? '') || ['heartbeat_failed', 'offline_short', 'offline_mass', 'web_connection_degraded', 'pending_sync', 'submit_held_pending_sync'].includes(event.event_type);
	}

	function isSuspectedMisconduct(event: ProctoringEvent) {
		const category = classifyProctorEvent(event);
		return ['anti_cheat', 'app_background_resume', 'device_mismatch'].includes(category ?? '');
	}

	function isProctorActionEvent(event: ProctoringEvent) {
		return ['proctor_acknowledge', 'proctor_incident_action', 'participant_command', 'participant_command_ack'].includes(event.event_type);
	}

	function isAccessActionEvent(event: ProctoringEvent) {
		return ['proctor_reset_access', 'proctor_unlock', 'proctor_force_submit'].includes(event.event_type);
	}

	function countEvents(types: string[]) {
		return filteredEvents.filter((event) => types.includes(event.event_type)).length;
	}

	function fmtDate(value?: string | null) {
		if (!value) return '-';
		return new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value));
	}

	function riskBadge(row?: ProctoringRow) {
		return row?.risk_level ?? 'normal';
	}

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
			rows.push([
				event.created_at,
				event.room_name,
				event.nis,
				event.nama,
				proctorEventLabel(event),
				proctorEvidenceCategoryLabel(classifyProctorEvent(event)),
				riskLabel(riskBadge(row)),
				reviewedParticipantIds.has(event.participant_id) ? 'ya' : 'belum',
				proctorEventDetail(event)
			]);
		}
		const blob = new Blob([rows.map((row) => csvRow(row)).join('\n')], { type: 'text/csv;charset=utf-8' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `rekap_insiden_cbt_${sessionId}.csv`;
		a.click();
		URL.revokeObjectURL(url);
		toast.success('Berkas rekap insiden dibuat');
	}
</script>

<svelte:head>
	<title>Rekap Insiden Ujian Digital</title>
</svelte:head>

<div class="mx-auto max-w-7xl space-y-6 p-4 md:p-6 print:max-w-none print:p-0">
	<div class="flex flex-wrap items-start justify-between gap-3 print:hidden">
		<div>
			<p class="text-sm text-muted-foreground">Panel Pengawasan</p>
			<h1 class="text-2xl font-semibold tracking-tight">Rekap Insiden & Berita Acara Ujian Digital</h1>
			<p class="text-sm text-muted-foreground">Saring, unduh CSV, dan cetak berita acara pengawasan per sesi.</p>
		</div>
		<div class="flex flex-wrap gap-2">
			<Button variant="outline" href={resolve(`/asesmen/sesi/${sessionId}/proctoring`)}>Kembali</Button>
			<Button variant="outline" onclick={exportCsv}>Unduh CSV</Button>
			<Button onclick={() => window.print()}>Cetak Berita Acara</Button>
		</div>
	</div>

	{#if loading}
		<Card.Root><Card.Content class="p-6 text-sm text-muted-foreground">Memuat rekap insiden...</Card.Content></Card.Root>
	{:else if loadError}
		<Card.Root><Card.Content class="space-y-3 p-6"><p class="text-sm text-destructive">{loadError}</p><Button variant="outline" onclick={() => void loadReport()}>Coba lagi</Button></Card.Content></Card.Root>
	{:else}

		<div class="grid gap-4 md:grid-cols-4 print:hidden">
			<Card.Root><Card.Content class="p-4"><p class="text-xs text-muted-foreground">Peserta</p><p class="text-2xl font-semibold">{participants.length}</p></Card.Content></Card.Root>
			<Card.Root><Card.Content class="p-4"><p class="text-xs text-muted-foreground">Kejadian</p><p class="text-2xl font-semibold">{events.length}</p></Card.Content></Card.Root>
			<Card.Root><Card.Content class="p-4"><p class="text-xs text-muted-foreground">Insiden</p><p class="text-2xl font-semibold">{incidentCount}</p></Card.Content></Card.Root>
			<Card.Root><Card.Content class="p-4"><p class="text-xs text-muted-foreground">Diperiksa</p><p class="text-2xl font-semibold">{reviewedParticipantIds.size}</p></Card.Content></Card.Root>
		</div>

		<Card.Root class="print:hidden">
			<Card.Header><Card.Title>Filter</Card.Title></Card.Header>
			<Card.Content class="grid gap-3 md:grid-cols-5">
				<select class="rounded-md border bg-background px-3 py-2 text-sm" bind:value={roomFilter}>
					<option value="all">Semua ruang</option>
					{#each roomOptions as room}<option value={room.id}>{room.name}</option>{/each}
				</select>
				<input class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Cari nama/NIS" bind:value={participantFilter} />
				<select class="rounded-md border bg-background px-3 py-2 text-sm" bind:value={eventTypeFilter}>
					<option value="all">Semua kejadian</option>
					{#each eventTypeOptions as type}<option value={type}>{eventTypeLabel(type)}</option>{/each}
				</select>
				<select class="rounded-md border bg-background px-3 py-2 text-sm" bind:value={riskFilter}>
					<option value="all">Semua risiko</option><option value="normal">Normal</option><option value="warning">Perlu perhatian</option><option value="high">Bahaya</option>
				</select>
				<select class="rounded-md border bg-background px-3 py-2 text-sm" bind:value={reviewFilter}>
					<option value="all">Semua status</option><option value="reviewed">Sudah diperiksa</option><option value="unreviewed">Belum diperiksa</option>
				</select>
			</Card.Content>
		</Card.Root>

		<Card.Root class="print:border-0 print:shadow-none">
			<Card.Header class="border-b print:break-after-avoid print:text-center">
				<Card.Title>Berita Acara Pengawasan Ujian Digital</Card.Title>
				<Card.Description>Dokumen rekap resmi per sesi · Dicetak: {fmtDate(new Date().toISOString())}</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-6 print:space-y-4 print:p-0 print:text-[11px]">
				<section class="rounded-lg border p-4 print:break-inside-avoid print:rounded-none print:border-black">
					<h2 class="mb-3 text-sm font-semibold uppercase tracking-wide">A. Identitas Sesi dan Lingkup Berita Acara</h2>
					<div class="grid gap-2 text-sm md:grid-cols-3 print:grid-cols-3">
						<div><span class="text-muted-foreground print:text-black">Sesi:</span> {sessionId}</div>
						<div><span class="text-muted-foreground print:text-black">Ruang:</span> {roomFilter === 'all' ? 'Semua ruang' : roomOptions.find((r) => r.id === roomFilter)?.name}</div>
						<div><span class="text-muted-foreground print:text-black">Jumlah kejadian terfilter:</span> {filteredEvents.length}</div>
					</div>
				</section>

				<section class="rounded-lg border p-4 print:break-inside-avoid print:rounded-none print:border-black">
					<h2 class="mb-3 text-sm font-semibold uppercase tracking-wide">B. Daftar Hadir dan Status Peserta</h2>
					<div class="grid gap-3 text-sm sm:grid-cols-2 lg:grid-cols-4 print:grid-cols-4">
						<div><p class="text-muted-foreground print:text-black">Peserta dalam lingkup</p><p class="text-xl font-semibold">{attendanceSummary.total}</p></div>
						<div><p class="text-muted-foreground print:text-black">Sudah kirim jawaban</p><p class="text-xl font-semibold">{attendanceSummary.submitted}</p></div>
						<div><p class="text-muted-foreground print:text-black">Belum tercatat kirim</p><p class="text-xl font-semibold">{attendanceSummary.notSubmitted}</p></div>
						<div><p class="text-muted-foreground print:text-black">Terkunci/dikunci</p><p class="text-xl font-semibold">{attendanceSummary.locked}</p></div>
						<div><p class="text-muted-foreground print:text-black">Risiko tinggi</p><p class="text-xl font-semibold">{attendanceSummary.highRisk}</p></div>
						<div><p class="text-muted-foreground print:text-black">Perlu perhatian</p><p class="text-xl font-semibold">{attendanceSummary.warning}</p></div>
						<div><p class="text-muted-foreground print:text-black">Sudah diperiksa pengawas</p><p class="text-xl font-semibold">{attendanceSummary.reviewed}</p></div>
						<div><p class="text-muted-foreground print:text-black">Koneksi terakhir tercatat</p><p class="text-xl font-semibold">{reportParticipants.filter((participant) => participant.last_heartbeat).length}</p></div>
					</div>
					<p class="mt-3 text-xs text-muted-foreground print:text-black">Status kirim disusun dari kejadian kirim jawaban dan paksa kirim yang tersedia pada log pengawasan; validasi akhir tetap mengacu arsip hasil ujian.</p>
				</section>

				<section class="grid gap-4 lg:grid-cols-2 print:grid-cols-2 print:break-inside-avoid">
					<div class="rounded-lg border p-4 print:rounded-none print:border-black">
						<h2 class="mb-3 text-sm font-semibold uppercase tracking-wide">C. Insiden Teknis</h2>
						<p class="text-sm text-muted-foreground print:text-black">Koneksi/sinkronisasi/submit tertahan: {technicalEvents.length}</p>
						<ul class="mt-3 space-y-1 text-sm">
							<li>Gangguan koneksi: {technicalEvents.filter((event) => classifyProctorEvent(event) === 'stale_connection').length}</li>
							<li>Submit/sinkronisasi tertahan: {technicalEvents.filter((event) => classifyProctorEvent(event) === 'submit_guard').length}</li>
							<li>Heartbeat/koneksi pulih: {technicalEvents.filter((event) => classifyProctorEvent(event) === 'heartbeat').length}</li>
						</ul>
					</div>
					<div class="rounded-lg border p-4 print:rounded-none print:border-black">
						<h2 class="mb-3 text-sm font-semibold uppercase tracking-wide">D. Dugaan Pelanggaran Tata Tertib</h2>
						<p class="text-sm text-muted-foreground print:text-black">Kejadian aplikasi/perangkat mencurigakan: {suspectedMisconductEvents.length}</p>
						<ul class="mt-3 space-y-1 text-sm">
							<li>Pelanggaran anti-cheat/tangkapan layar: {suspectedMisconductEvents.filter((event) => classifyProctorEvent(event) === 'anti_cheat').length}</li>
							<li>Aplikasi ditinggalkan/fokus berpindah: {suspectedMisconductEvents.filter((event) => classifyProctorEvent(event) === 'app_background_resume').length}</li>
							<li>Perangkat atau token tidak sesuai: {suspectedMisconductEvents.filter((event) => classifyProctorEvent(event) === 'device_mismatch').length}</li>
						</ul>
					</div>
				</section>

				<section class="grid gap-4 lg:grid-cols-2 print:grid-cols-2 print:break-inside-avoid">
					<div class="rounded-lg border p-4 print:rounded-none print:border-black">
						<h2 class="mb-3 text-sm font-semibold uppercase tracking-wide">E. Ringkasan Tindakan Pengawas</h2>
						<ul class="space-y-1 text-sm">
							<li>Ditandai sudah diperiksa/tindak lanjut: {proctorActionEvents.length}</li>
							<li>Peserta unik diperiksa: {reviewedParticipantIds.size}</li>
							<li>Instruksi ke aplikasi siswa: {countEvents(['participant_command', 'participant_command_ack'])}</li>
						</ul>
					</div>
					<div class="rounded-lg border p-4 print:rounded-none print:border-black">
						<h2 class="mb-3 text-sm font-semibold uppercase tracking-wide">F. Reset, Buka Kunci, dan Paksa Kirim</h2>
						<ul class="space-y-1 text-sm">
							<li>Reset akses: {countEvents(['proctor_reset_access'])}</li>
							<li>Buka kunci peserta: {countEvents(['proctor_unlock'])}</li>
							<li>Paksa kirim jawaban: {countEvents(['proctor_force_submit'])}</li>
							<li>Total tindakan akses: {accessActionEvents.length}</li>
						</ul>
					</div>
				</section>

				<section class="rounded-lg border p-4 print:rounded-none print:border-black">
					<h2 class="mb-3 text-sm font-semibold uppercase tracking-wide">G. Rincian Kejadian dan Tindakan</h2>
					<Table.Root>
						<Table.Header><Table.Row><Table.Head>Waktu</Table.Head><Table.Head>Ruang</Table.Head><Table.Head>Peserta</Table.Head><Table.Head>Kategori</Table.Head><Table.Head>Status</Table.Head><Table.Head>Detail/Tindakan</Table.Head></Table.Row></Table.Header>
						<Table.Body>
							{#each filteredEvents as event}
								{@const row = participants.find((p) => p.participant_id === event.participant_id)}
								<Table.Row class="print:break-inside-avoid">
									<Table.Cell>{fmtDate(event.created_at)}</Table.Cell>
									<Table.Cell>{event.room_name}</Table.Cell>
									<Table.Cell><div class="font-medium">{event.nama}</div><div class="text-xs text-muted-foreground print:text-black">{event.nis}</div></Table.Cell>
									<Table.Cell><Badge variant="outline">{proctorEvidenceCategoryLabel(classifyProctorEvent(event))}</Badge></Table.Cell>
									<Table.Cell><Badge variant="outline">{reviewedParticipantIds.has(event.participant_id) ? 'sudah diperiksa' : riskLabel(riskBadge(row))}</Badge></Table.Cell>
									<Table.Cell class="max-w-md text-xs">{proctorEventDetail(event)}</Table.Cell>
								</Table.Row>
							{/each}
							{#if filteredEvents.length === 0}
								<Table.Row><Table.Cell colspan={6} class="text-center text-muted-foreground print:text-black">Tidak ada kejadian pada filter ini.</Table.Cell></Table.Row>
							{/if}
						</Table.Body>
					</Table.Root>
				</section>

				<section class="grid gap-4 lg:grid-cols-2 print:grid-cols-2 print:break-inside-avoid">
					<div class="rounded-lg border p-4 print:rounded-none print:border-black">
						<h2 class="mb-3 text-sm font-semibold uppercase tracking-wide">H. Verifikasi Hasil dan Arsip Final</h2>
						<p class="text-sm">Operator memverifikasi status submit, sinkronisasi jawaban, dan hasil akhir melalui arsip sistem setelah sesi ditutup.</p>
						<p class="mt-2 text-sm">Tautan arsip/panel sesi: <a class="underline" href={resolve(`/asesmen/sesi/${sessionId}/proctoring`)}>Pengawasan sesi</a></p>
						<p class="mt-2 text-xs text-muted-foreground print:text-black">Lampiran CSV dan cetak halaman ini dapat disimpan sebagai arsip pendukung berita acara.</p>
					</div>
					<div class="rounded-lg border p-4 print:rounded-none print:border-black">
						<h2 class="mb-3 text-sm font-semibold uppercase tracking-wide">I. Serah Terima dan Catatan</h2>
						<div class="min-h-24 rounded border border-dashed p-3 text-sm text-muted-foreground print:border-black print:text-black">Catatan manual pengawas/operator: ................................................................................................................................................................................................................</div>
					</div>
				</section>

				<section class="grid gap-8 pt-4 text-center text-sm md:grid-cols-3 print:grid-cols-3 print:break-inside-avoid">
					<div><p>Pengawas Ruang</p><div class="h-20"></div><p class="border-t pt-2">Nama & Tanda Tangan</p></div>
					<div><p>Operator/Admin Ujian</p><div class="h-20"></div><p class="border-t pt-2">Nama & Tanda Tangan</p></div>
					<div><p>Ketua Panitia</p><div class="h-20"></div><p class="border-t pt-2">Nama & Tanda Tangan</p></div>
				</section>
			</Card.Content>
		</Card.Root>
	{/if}
</div>
