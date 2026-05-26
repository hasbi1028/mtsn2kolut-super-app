<script lang="ts">
	import PrinterIcon from '@lucide/svelte/icons/printer';
	import QrCodeIcon from '@lucide/svelte/icons/qr-code';
	import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import UsersIcon from '@lucide/svelte/icons/users';
	import ShieldCheckIcon from '@lucide/svelte/icons/shield-check';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { clientApiPath, readClientApiData } from '$lib/client/api';
	import { summarizeDocumentPrintStatus, type DocumentPrintOverview, type SupervisorCardLike } from '$lib/asesmen/document-print-readiness';
	import type { ExamCardLike } from '$lib/asesmen/exam-card-print';

	type EventInfo = { id: string; title: string; exam_type?: string; academic_year_name?: string; status?: string; scope?: string };
	type HubData = { info: EventInfo | null; overview: DocumentPrintOverview | null; examCards: ExamCardLike[]; supervisorCards: SupervisorCardLike[] };

	const eventId = page.params.id ?? '';
	let loading = $state(true);
	let loadError = $state<unknown>(null);
	let data = $state<HubData>({ info: null, overview: null, examCards: [], supervisorCards: [] });
	let issueBusy = $state<'participant' | 'supervisor' | null>(null);
	let readiness = $derived(summarizeDocumentPrintStatus({ examCards: data.examCards, supervisorCards: data.supervisorCards, overview: data.overview }));

	async function fetchEventInfo() {
		try {
			const response = await fetch(clientApiPath`/api/asesmen/events/${eventId}`);
			return await readClientApiData<EventInfo>(response, 'Gagal memuat kegiatan');
		} catch {
			return null;
		}
	}

	async function fetchOverview() {
		try {
			const response = await fetch(clientApiPath`/api/asesmen/events/${eventId}/overview`);
			return await readClientApiData<DocumentPrintOverview>(response, 'Gagal memuat ringkasan kegiatan');
		} catch {
			return null;
		}
	}

	async function fetchExamCards() {
		const response = await fetch(clientApiPath`/api/asesmen/events/${eventId}/exam-access-cards`);
		const payload = await readClientApiData<unknown>(response, 'Gagal memuat kartu peserta');
		if (Array.isArray(payload)) return payload as ExamCardLike[];
		if (payload && typeof payload === 'object' && Array.isArray((payload as { cards?: unknown[] }).cards)) return (payload as { cards: ExamCardLike[] }).cards;
		return [];
	}

	async function fetchSupervisorCards() {
		const response = await fetch(clientApiPath`/api/asesmen/events/${eventId}/supervisor-access-cards`);
		const payload = await readClientApiData<unknown>(response, 'Gagal memuat lembar pengawas ruang');
		if (Array.isArray(payload)) return payload as SupervisorCardLike[];
		if (payload && typeof payload === 'object' && Array.isArray((payload as { cards?: unknown[] }).cards)) return (payload as { cards: SupervisorCardLike[] }).cards;
		return [];
	}

	async function loadHub() {
		loading = true;
		loadError = null;
		try {
			const [info, overview, examCards, supervisorCards] = await Promise.all([fetchEventInfo(), fetchOverview(), fetchExamCards(), fetchSupervisorCards()]);
			data = { info, overview, examCards, supervisorCards };
		} catch (error) {
			loadError = error;
		} finally {
			loading = false;
		}
	}

	async function issueCards(kind: 'participant' | 'supervisor', regenerate = false) {
		const label = kind === 'participant' ? 'kartu peserta' : 'lembar pengawas ruang';
		if (regenerate && !confirm(`Reset ulang semua QR+PIN ${label} kegiatan ini? PIN lama tidak berlaku.`)) return;
		issueBusy = kind;
		try {
			const endpoint = kind === 'participant' ? clientApiPath`/api/asesmen/events/${eventId}/exam-access-cards/issue` : clientApiPath`/api/asesmen/events/${eventId}/supervisor-access-cards/issue`;
			const response = await fetch(endpoint, {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ regenerate, expires_hours: 0 })
			});
			const payload = await readClientApiData<unknown>(response, `Gagal menerbitkan ${label}`);
			const issuedCards = Array.isArray(payload)
				? payload
				: payload && typeof payload === 'object' && Array.isArray((payload as { cards?: unknown[] }).cards)
					? (payload as { cards: unknown[] }).cards
					: null;
			if (issuedCards) {
				data = kind === 'participant'
					? { ...data, examCards: issuedCards as ExamCardLike[] }
					: { ...data, supervisorCards: issuedCards as SupervisorCardLike[] };
			} else {
				await loadHub();
			}
		} finally {
			issueBusy = null;
		}
	}

	function statusClass(state: string) {
		if (state === 'ready') return 'border-success/20 bg-success/10 text-success';
		if (state === 'warning') return 'border-warning/30 bg-warning/10 text-warning';
		return 'border-muted bg-muted text-muted-foreground';
	}

	function errorMessage(error: unknown) {
		return error instanceof Error && error.message.trim() ? error.message : 'Dokumen kegiatan belum dapat dimuat.';
	}

	onMount(() => {
		void loadHub();
	});
</script>

<svelte:head>
	<title>Dokumen & Cetak Kegiatan</title>
</svelte:head>

{#if loading}
	<div class="mx-auto max-w-7xl space-y-5 p-6">
		<Skeleton class="h-8 w-72" />
		<Skeleton class="h-24 rounded-xl" />
		<div class="grid gap-4 md:grid-cols-2"><Skeleton class="h-52 rounded-xl" /><Skeleton class="h-52 rounded-xl" /></div>
	</div>
{:else if loadError}
	<div class="mx-auto max-w-7xl p-6">
		<RecoveryPanel title="Dokumen & Cetak Belum Tersaji" message={errorMessage(loadError)} onRetry={loadHub} />
	</div>
{:else}
	<div class="mx-auto max-w-7xl space-y-6 p-6">
		<div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
			<div>
				<p class="text-sm font-medium text-primary">Asesmen · Kegiatan</p>
				<h1 class="text-2xl font-semibold text-foreground">Dokumen & Cetak</h1>
				<p class="text-sm text-muted-foreground">{data.info?.title ?? 'Kegiatan Ujian'}{data.info?.academic_year_name ? ` · ${data.info.academic_year_name}` : ''}</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<a href={resolve(`/asesmen/kegiatan/${eventId}`)} class="inline-flex items-center rounded-md border border-border bg-card px-3 py-2 text-sm font-semibold text-foreground hover:bg-muted">Kembali ke Kegiatan</a>
				<Button variant="outline" onclick={loadHub}><RefreshCwIcon class="mr-2 size-4" />Refresh Status</Button>
			</div>
		</div>

		<div class="grid gap-3 md:grid-cols-4">
			<div class="rounded-xl border border-primary/20 bg-primary/10 p-4">
				<p class="text-xs font-medium uppercase tracking-wide text-primary">Peserta/Kartu</p>
				<p class="mt-1 text-2xl font-semibold text-primary">{data.examCards.length || data.overview?.card_count || data.overview?.token_count || 0}</p>
			</div>
			<div class="rounded-xl border border-border bg-card p-4">
				<p class="text-xs font-medium uppercase tracking-wide text-muted-foreground">Sesi</p>
				<p class="mt-1 text-2xl font-semibold text-foreground">{data.overview?.session_count ?? 0}</p>
			</div>
			<div class="rounded-xl border border-border bg-card p-4">
				<p class="text-xs font-medium uppercase tracking-wide text-muted-foreground">Ruang</p>
				<p class="mt-1 text-2xl font-semibold text-foreground">{data.overview?.room_count ?? data.supervisorCards.length}</p>
			</div>
			<div class="rounded-xl border border-success/20 bg-success/10 p-4">
				<p class="text-xs font-medium uppercase tracking-wide text-success">Lembar Pengawas</p>
				<p class="mt-1 text-2xl font-semibold text-success">{data.supervisorCards.length}</p>
			</div>
		</div>

		{#if readiness.blockers.length > 0}
			<div class="rounded-xl border border-warning/30 bg-warning/10 p-4 text-sm text-warning" role="alert">
				<p class="font-semibold">Ada dokumen yang perlu dicek sebelum cetak final.</p>
				<ul class="mt-2 list-disc space-y-1 pl-5">
					{#each readiness.blockers.slice(0, 5) as blocker}<li>{blocker}</li>{/each}
				</ul>
			</div>
		{:else}
			<div class="rounded-xl border border-success/20 bg-success/10 p-4 text-sm text-success">
				Dokumen utama sudah siap. Gunakan halaman cetak masing-masing untuk pratinjau dan cetak massal.
			</div>
		{/if}

		<div class="grid gap-4 lg:grid-cols-2">
			<section class="rounded-2xl border bg-card p-5 shadow-sm">
				<div class="flex items-start justify-between gap-3">
					<div class="flex gap-3">
						<div class="grid size-11 shrink-0 place-items-center rounded-xl bg-primary/10 text-primary"><UsersIcon class="size-5" /></div>
						<div>
							<h2 class="text-lg font-semibold text-foreground">Kartu Peserta Ujian</h2>
							<p class="text-sm text-muted-foreground">QR/PIN, ruang, nomor meja, jadwal, dan identitas peserta. Default cetak: per ruang lalu nomor meja.</p>
						</div>
					</div>
					<Badge variant="outline" class={statusClass(readiness.participantCards.state)}>{readiness.participantCards.label}</Badge>
				</div>
				<div class="mt-4 grid gap-2 text-sm text-muted-foreground sm:grid-cols-3">
					<p><span class="font-semibold text-foreground">{readiness.participantCards.count}</span><br />kartu termuat</p>
					<p><span class="font-semibold text-foreground">Filter</span><br />kelas/sesi/ruang</p>
					<p><span class="font-semibold text-foreground">Layout</span><br />A4 8, A4 4, 1 halaman</p>
				</div>
				<div class="mt-5 flex flex-wrap gap-2">
					<a href={resolve(`/asesmen/kegiatan/${eventId}/exam-cards`)} class="inline-flex items-center rounded-md bg-primary px-3 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90"><PrinterIcon class="mr-2 size-4" />Pratinjau & Cetak Massal</a>
					<Button variant="outline" onclick={() => issueCards('participant', false)} disabled={issueBusy !== null}><QrCodeIcon class="mr-2 size-4" />Terbitkan QR+PIN</Button>
					<Button variant="outline" onclick={() => issueCards('participant', true)} disabled={issueBusy !== null}>Reset QR+PIN</Button>
				</div>
			</section>

			<section class="rounded-2xl border bg-card p-5 shadow-sm">
				<div class="flex items-start justify-between gap-3">
					<div class="flex gap-3">
						<div class="grid size-11 shrink-0 place-items-center rounded-xl bg-success/10 text-success"><ShieldCheckIcon class="size-5" /></div>
						<div>
							<h2 class="text-lg font-semibold text-foreground">Lembar Pengawas Ruang</h2>
							<p class="text-sm text-muted-foreground">Satu QR+PIN per ruang, bukan akun personal. Cocok untuk pengawas pengganti pada hari-H.</p>
						</div>
					</div>
					<Badge variant="outline" class={statusClass(readiness.supervisorCards.state)}>{readiness.supervisorCards.label}</Badge>
				</div>
				<div class="mt-4 grid gap-2 text-sm text-muted-foreground sm:grid-cols-3">
					<p><span class="font-semibold text-foreground">{readiness.supervisorCards.count}</span><br />lembar termuat</p>
					<p><span class="font-semibold text-foreground">Portal</span><br />pengawasan ruang</p>
					<p><span class="font-semibold text-foreground">PIN</span><br />akses ruang</p>
				</div>
				<div class="mt-5 flex flex-wrap gap-2">
					<a href={resolve(`/asesmen/kegiatan/${eventId}/pengawas-cards`)} class="inline-flex items-center rounded-md bg-primary px-3 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90"><PrinterIcon class="mr-2 size-4" />Pratinjau & Cetak Semua</a>
					<Button variant="outline" onclick={() => issueCards('supervisor', false)} disabled={issueBusy !== null}><QrCodeIcon class="mr-2 size-4" />Terbitkan QR+PIN Ruang</Button>
					<Button variant="outline" onclick={() => issueCards('supervisor', true)} disabled={issueBusy !== null}>Reset QR+PIN</Button>
				</div>
			</section>
		</div>

		<section class="rounded-2xl border border-dashed bg-muted/30 p-5">
			<div class="flex gap-3">
				<div class="grid size-11 shrink-0 place-items-center rounded-xl bg-muted text-muted-foreground"><FileTextIcon class="size-5" /></div>
				<div>
					<h2 class="font-semibold text-foreground">Dokumen lanjutan</h2>
					<p class="text-sm text-muted-foreground">Daftar hadir per ruang, denah/nomor meja, dan paket cetak ruang bisa ditautkan ke hub ini setelah template final disetujui. Route yang sudah ada tetap hidup dan tidak dihapus.</p>
				</div>
			</div>
		</section>
	</div>
{/if}
