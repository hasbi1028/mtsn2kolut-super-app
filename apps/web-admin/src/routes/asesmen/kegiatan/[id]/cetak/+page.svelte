<script lang="ts">
	import PrinterIcon from '@lucide/svelte/icons/printer';
	import QrCodeIcon from '@lucide/svelte/icons/qr-code';
	import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import UsersIcon from '@lucide/svelte/icons/users';
	import ShieldCheckIcon from '@lucide/svelte/icons/shield-check';
	import ClipboardCheckIcon from '@lucide/svelte/icons/clipboard-check';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import MicroActionTable from '$lib/components/ops/MicroActionTable.svelte';
	import { clientApiPath, readClientApiData } from '$lib/client/api';
	import {
		summarizeDocumentPrintStatus,
		type DocumentPrintOverview,
		type SupervisorCardLike
	} from '$lib/asesmen/document-print-readiness';
	import type { ExamCardLike } from '$lib/asesmen/exam-card-print';

	type EventInfo = {
		id: string;
		title: string;
		exam_type?: string;
		academic_year_name?: string;
		status?: string;
		scope?: string;
	};
	type HubData = {
		info: EventInfo | null;
		overview: DocumentPrintOverview | null;
		examCards: ExamCardLike[];
		supervisorCards: SupervisorCardLike[];
	};
	type DocumentRow = {
		id: 'participant' | 'supervisor' | 'archive';
		title: string;
		helper: string;
		countLabel: string;
		statusLabel: string;
		state: 'ready' | 'warning' | 'empty';
		href: string;
		primaryAction: string;
		secondaryAction?: string;
		resetAction?: string;
		icon: typeof UsersIcon;
	};

	const eventId = page.params.id ?? '';
	const rowColumns = [
		{ key: 'document', label: 'Dokumen', class: 'min-w-[18rem]' },
		{ key: 'status', label: 'Status', class: 'min-w-[12rem]' },
		{ key: 'coverage', label: 'Cakupan', class: 'min-w-[14rem]' }
	];

	let loading = $state(true);
	let loadError = $state<unknown>(null);
	let data = $state<HubData>({ info: null, overview: null, examCards: [], supervisorCards: [] });
	let issueBusy = $state<'participant' | 'supervisor' | null>(null);
	let readiness = $derived(
		summarizeDocumentPrintStatus({
			examCards: data.examCards,
			supervisorCards: data.supervisorCards,
			overview: data.overview
		})
	);
	let rows = $derived<DocumentRow[]>([
		{
			id: 'participant',
			title: 'Kartu Peserta Ujian',
			helper:
				'QR/PIN peserta, ruang, nomor meja, dan jadwal. Gunakan saat distribusi kartu sebelum ujian.',
			countLabel: `${readiness.participantCards.count || data.overview?.card_count || data.overview?.token_count || 0} kartu`,
			statusLabel: readiness.participantCards.label,
			state: readiness.participantCards.state,
			href: resolve(`/asesmen/kegiatan/${eventId}/exam-cards`),
			primaryAction: 'Pratinjau & Cetak Massal',
			secondaryAction: 'Terbitkan QR+PIN',
			resetAction: 'Reset QR+PIN',
			icon: UsersIcon
		},
		{
			id: 'supervisor',
			title: 'Lembar Pengawas Ruang',
			helper:
				'Satu QR+PIN per ruang, bukan akun personal. Tetap bisa dipakai pengawas pengganti pada hari-H.',
			countLabel: `${readiness.supervisorCards.count || data.overview?.room_count || 0} lembar`,
			statusLabel: readiness.supervisorCards.label,
			state: readiness.supervisorCards.state,
			href: resolve(`/asesmen/kegiatan/${eventId}/pengawas-cards`),
			primaryAction: 'Pratinjau & Cetak Semua',
			secondaryAction: 'Terbitkan QR+PIN Ruang',
			resetAction: 'Reset QR+PIN',
			icon: ShieldCheckIcon
		},
		{
			id: 'archive',
			title: 'Checklist Arsip & Dokumen Lanjutan',
			helper:
				'Pusat tindak lanjut untuk arsip final, daftar hadir, berita acara, dan rekap hasil setelah cetak operasional selesai.',
			countLabel: `${data.overview?.session_count ?? 0} sesi · ${data.overview?.room_count ?? 0} ruang`,
			statusLabel: readiness.blockers.length > 0 ? 'Masih ada yang perlu dicek' : 'Siap lanjut ke arsip',
			state: readiness.blockers.length > 0 ? 'warning' : 'ready',
			href: resolve(`/asesmen/kegiatan/${eventId}/archive`),
			primaryAction: 'Buka Checklist Arsip',
			icon: ClipboardCheckIcon
		}
	]);

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
		if (payload && typeof payload === 'object' && Array.isArray((payload as { cards?: unknown[] }).cards)) {
			return (payload as { cards: ExamCardLike[] }).cards;
		}
		return [];
	}

	async function fetchSupervisorCards() {
		const response = await fetch(clientApiPath`/api/asesmen/events/${eventId}/supervisor-access-cards`);
		const payload = await readClientApiData<unknown>(response, 'Gagal memuat lembar pengawas ruang');
		if (Array.isArray(payload)) return payload as SupervisorCardLike[];
		if (payload && typeof payload === 'object' && Array.isArray((payload as { cards?: unknown[] }).cards)) {
			return (payload as { cards: SupervisorCardLike[] }).cards;
		}
		return [];
	}

	async function loadHub() {
		loading = true;
		loadError = null;
		try {
			const [info, overview, examCards, supervisorCards] = await Promise.all([
				fetchEventInfo(),
				fetchOverview(),
				fetchExamCards(),
				fetchSupervisorCards()
			]);
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
			const endpoint =
				kind === 'participant'
					? clientApiPath`/api/asesmen/events/${eventId}/exam-access-cards/issue`
					: clientApiPath`/api/asesmen/events/${eventId}/supervisor-access-cards/issue`;
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
				data =
					kind === 'participant'
						? { ...data, examCards: issuedCards as ExamCardLike[] }
						: { ...data, supervisorCards: issuedCards as SupervisorCardLike[] };
			} else {
				await loadHub();
			}
		} finally {
			issueBusy = null;
		}
	}

	function statusClass(state: 'ready' | 'warning' | 'empty') {
		if (state === 'ready') return 'border-success/20 bg-success/10 text-success';
		if (state === 'warning') return 'border-warning/30 bg-warning/10 text-warning';
		return 'border-muted bg-muted text-muted-foreground';
	}

	function errorMessage(error: unknown) {
		return error instanceof Error && error.message.trim()
			? error.message
			: 'Dokumen kegiatan belum dapat dimuat.';
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
		<Skeleton class="h-20 rounded-xl" />
		<Skeleton class="h-56 rounded-xl" />
	</div>
{:else if loadError}
	<div class="mx-auto max-w-7xl p-6">
		<RecoveryPanel
			title="Dokumen & Cetak Belum Tersaji"
			message={errorMessage(loadError)}
			onRetry={loadHub}
		/>
	</div>
{:else}
	<div class="mx-auto max-w-7xl space-y-6 p-6">
		<div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
			<div>
				<p class="text-sm font-medium text-primary">Asesmen · Kegiatan</p>
				<h1 class="text-2xl font-semibold text-foreground">Dokumen & Cetak</h1>
				<p class="text-sm text-muted-foreground">
					{data.info?.title ?? 'Kegiatan Ujian'}
					{data.info?.academic_year_name ? ` · ${data.info.academic_year_name}` : ''}
				</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<a
					href={resolve(`/asesmen/kegiatan/${eventId}`)}
					class="inline-flex items-center rounded-md border border-border bg-card px-3 py-2 text-sm font-semibold text-foreground hover:bg-muted"
				>
					Kembali ke Kegiatan
				</a>
				<Button variant="outline" onclick={loadHub}
					><RefreshCwIcon class="mr-2 size-4" />Refresh Status</Button
				>
			</div>
		</div>

		<div class="grid gap-3 md:grid-cols-4">
			<div class="rounded-xl border border-primary/20 bg-primary/10 px-4 py-3">
				<p class="text-[11px] font-semibold uppercase tracking-[0.16em] text-primary">Peserta/Kartu</p>
				<p class="mt-1 text-2xl font-semibold text-primary">
					{data.examCards.length || data.overview?.card_count || data.overview?.token_count || 0}
				</p>
			</div>
			<div class="rounded-xl border border-border bg-card px-4 py-3">
				<p class="text-[11px] font-semibold uppercase tracking-[0.16em] text-muted-foreground">Sesi</p>
				<p class="mt-1 text-2xl font-semibold text-foreground">{data.overview?.session_count ?? 0}</p>
			</div>
			<div class="rounded-xl border border-border bg-card px-4 py-3">
				<p class="text-[11px] font-semibold uppercase tracking-[0.16em] text-muted-foreground">Ruang</p>
				<p class="mt-1 text-2xl font-semibold text-foreground">
					{data.overview?.room_count ?? data.supervisorCards.length}
				</p>
			</div>
			<div class="rounded-xl border border-success/20 bg-success/10 px-4 py-3">
				<p class="text-[11px] font-semibold uppercase tracking-[0.16em] text-success">Lembar Pengawas</p>
				<p class="mt-1 text-2xl font-semibold text-success">{data.supervisorCards.length}</p>
			</div>
		</div>

		{#if readiness.blockers.length > 0}
			<div class="rounded-xl border border-warning/30 bg-warning/10 p-4 text-sm text-warning" role="alert">
				<p class="font-semibold">Ada dokumen yang perlu dicek sebelum cetak final.</p>
				<ul class="mt-2 list-disc space-y-1 pl-5">
					{#each readiness.blockers.slice(0, 5) as blocker}
						<li>{blocker}</li>
					{/each}
				</ul>
			</div>
		{:else}
			<div class="rounded-xl border border-success/20 bg-success/10 p-4 text-sm text-success">
				Dokumen operasional utama sudah siap. Pilih baris dokumen di bawah untuk pratinjau,
				cetak, atau lanjut ke arsip final.
			</div>
		{/if}

		<MicroActionTable
			title="Pusat dokumen kegiatan"
			description="Semua jalur cetak utama dipusatkan di sini agar operator tidak perlu mencari tombol di banyak halaman."
			columns={rowColumns}
			rows={rows}
			rowKey={(row) => (row as DocumentRow).id}
			tableClass="min-w-[900px]"
			emptyTitle="Belum ada dokumen yang bisa ditampilkan."
		>
			{#snippet cell(row, column)}
				{@const item = row as DocumentRow}
				{#if column.key === 'document'}
					<div class="flex items-start gap-3">
						<div class={`grid size-10 shrink-0 place-items-center rounded-xl ${item.id === 'participant' ? 'bg-primary/10 text-primary' : item.id === 'supervisor' ? 'bg-success/10 text-success' : 'bg-muted text-muted-foreground'}`}>
							<item.icon class="size-4" />
						</div>
						<div class="space-y-1">
							<div class="font-semibold text-foreground">{item.title}</div>
							<p class="text-xs leading-5 text-muted-foreground">{item.helper}</p>
						</div>
					</div>
				{:else if column.key === 'status'}
					<div class="space-y-2">
						<Badge variant="outline" class={statusClass(item.state)}>{item.statusLabel}</Badge>
						<p class="text-xs text-muted-foreground">
							{item.id === 'archive'
								? 'Lanjutkan ke arsip setelah dokumen operasional siap.'
								: item.state === 'ready'
									? 'Bisa dicetak sekarang.'
									: 'Perlu cek data sebelum cetak final.'}
						</p>
					</div>
				{:else}
					<div class="space-y-1 text-xs text-muted-foreground">
						<p class="font-semibold text-foreground">{item.countLabel}</p>
						<p>
							{item.id === 'participant'
								? 'Urutan default: ruang → meja → nama.'
								: item.id === 'supervisor'
									? 'Akses ruang fleksibel untuk pengawas pengganti.'
									: 'Daftar hadir, BA, hasil, dan pengesahan.'}
						</p>
					</div>
				{/if}
			{/snippet}
			{#snippet actions(row)}
				{@const item = row as DocumentRow}
				<a
					href={item.href}
					class="inline-flex items-center rounded-md bg-primary px-3 py-2 text-xs font-semibold text-primary-foreground hover:bg-primary/90"
				>
					<PrinterIcon class="mr-1.5 size-3.5" />{item.primaryAction}
				</a>
				{#if item.id === 'participant'}
					<Button
						variant="outline"
						size="sm"
						onclick={() => issueCards('participant', false)}
						disabled={issueBusy !== null}
					>
						<QrCodeIcon class="mr-1.5 size-3.5" />{item.secondaryAction}
					</Button>
					<Button
						variant="outline"
						size="sm"
						onclick={() => issueCards('participant', true)}
						disabled={issueBusy !== null}
					>
						{item.resetAction}
					</Button>
				{:else if item.id === 'supervisor'}
					<Button
						variant="outline"
						size="sm"
						onclick={() => issueCards('supervisor', false)}
						disabled={issueBusy !== null}
					>
						<QrCodeIcon class="mr-1.5 size-3.5" />{item.secondaryAction}
					</Button>
					<Button
						variant="outline"
						size="sm"
						onclick={() => issueCards('supervisor', true)}
						disabled={issueBusy !== null}
					>
						{item.resetAction}
					</Button>
				{:else}
					<a
						href={resolve(`/asesmen/sesi?event_id=${eventId}`)}
						class="inline-flex items-center rounded-md border border-border bg-card px-3 py-2 text-xs font-semibold text-foreground hover:bg-muted"
					>
						<FileTextIcon class="mr-1.5 size-3.5" />Buka Sesi & BA
					</a>
				{/if}
			{/snippet}
			{#snippet mobile(row)}
				{@const item = row as DocumentRow}
				<div class="space-y-3">
					<div class="flex items-start justify-between gap-3">
						<div class="space-y-1">
							<p class="font-semibold text-foreground">{item.title}</p>
							<p class="text-xs leading-5 text-muted-foreground">{item.helper}</p>
						</div>
						<Badge variant="outline" class={statusClass(item.state)}>{item.statusLabel}</Badge>
					</div>
					<div class="rounded-lg bg-muted/50 px-3 py-2 text-xs text-muted-foreground">
						<p class="font-semibold text-foreground">{item.countLabel}</p>
					</div>
					<div class="flex flex-wrap gap-2">
						<a
							href={item.href}
							class="inline-flex items-center rounded-md bg-primary px-3 py-2 text-xs font-semibold text-primary-foreground hover:bg-primary/90"
						>
							<PrinterIcon class="mr-1.5 size-3.5" />{item.primaryAction}
						</a>
						{#if item.id === 'participant'}
							<Button
								variant="outline"
								size="sm"
								onclick={() => issueCards('participant', false)}
								disabled={issueBusy !== null}
							>
								<QrCodeIcon class="mr-1.5 size-3.5" />{item.secondaryAction}
							</Button>
						{:else if item.id === 'supervisor'}
							<Button
								variant="outline"
								size="sm"
								onclick={() => issueCards('supervisor', false)}
								disabled={issueBusy !== null}
							>
								<QrCodeIcon class="mr-1.5 size-3.5" />{item.secondaryAction}
							</Button>
						{:else}
							<a
								href={resolve(`/asesmen/sesi?event_id=${eventId}`)}
								class="inline-flex items-center rounded-md border border-border bg-card px-3 py-2 text-xs font-semibold text-foreground hover:bg-muted"
							>
								<FileTextIcon class="mr-1.5 size-3.5" />Buka Sesi & BA
							</a>
						{/if}
					</div>
				</div>
			{/snippet}
		</MicroActionTable>
	</div>
{/if}
