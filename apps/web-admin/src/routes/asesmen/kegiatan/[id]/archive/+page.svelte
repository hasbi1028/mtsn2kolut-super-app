<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import ArchiveIcon from '@lucide/svelte/icons/archive';
	import CheckCircle2Icon from '@lucide/svelte/icons/check-circle-2';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import ShieldCheckIcon from '@lucide/svelte/icons/shield-check';
	import ClipboardCheckIcon from '@lucide/svelte/icons/clipboard-check';
	import BadgeAlertIcon from '@lucide/svelte/icons/badge-alert';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import MicroActionTable from '$lib/components/ops/MicroActionTable.svelte';
	import { clientApiPath, readClientApiData } from '$lib/client/api';
	import { listApprovals, assessmentApprovalLabels, type AssessmentApprovalRecord } from '$lib/asesmen/approval-client';

	type EventInfo = {
		id: string;
		title: string;
		exam_type: string;
		scope: string;
		academic_year_name: string;
		status: string;
	};
	type EventSession = {
		id: string;
		title: string;
		status?: string;
		participant_count?: number;
		room_count?: number;
	};
	type EventOverview = {
		member_count?: number;
		package_count?: number;
		session_count?: number;
		room_count?: number;
		card_count?: number;
		result_count?: number;
		submitted_count?: number;
	};
	type ArchiveData = {
		info: EventInfo;
		overview: EventOverview | null;
		sessions: EventSession[];
		approvals: AssessmentApprovalRecord[];
	};
	type ArchiveApprovalType = 'package_ready' | 'participants_rooms_ready' | 'tokens_cards_ready' | 'results_verified' | 'final_archive';
	type ArchiveChecklistRow = {
		id: string;
		title: string;
		helper: string;
		href: string;
		status: string;
		coverage: string;
		ready: boolean;
		icon: typeof ArchiveIcon;
	};
	type ApprovalRow = {
		id: string;
		title: string;
		status: string;
		actor: string;
		ready: boolean;
	};

	const eventId = page.params.id ?? '';
	const archiveApprovalTypes: ArchiveApprovalType[] = [
		'package_ready',
		'participants_rooms_ready',
		'tokens_cards_ready',
		'results_verified',
		'final_archive'
	];
	const checklistColumns = [
		{ key: 'document', label: 'Dokumen', class: 'min-w-[18rem]' },
		{ key: 'status', label: 'Status', class: 'min-w-[12rem]' },
		{ key: 'coverage', label: 'Cakupan', class: 'min-w-[14rem]' }
	];
	const approvalColumns = [
		{ key: 'approval', label: 'Pengesahan', class: 'min-w-[18rem]' },
		{ key: 'status', label: 'Status', class: 'min-w-[12rem]' },
		{ key: 'actor', label: 'Aktor/Waktu', class: 'min-w-[14rem]' }
	];
	const sessionColumns = [
		{ key: 'session', label: 'Sesi', class: 'min-w-[16rem]' },
		{ key: 'status', label: 'Status', class: 'min-w-[10rem]' },
		{ key: 'coverage', label: 'Cakupan', class: 'min-w-[14rem]' }
	];

	let archivePromise = $state<Promise<ArchiveData> | null>(null);

	async function optionalApiData<T>(path: string, fallback: T): Promise<T> {
		try {
			return await fetch(path).then((response) => readClientApiData<T>(response));
		} catch {
			return fallback;
		}
	}

	function parseArrayPayload<T>(payload: unknown, key: string): T[] {
		if (Array.isArray(payload)) return payload as T[];
		if (typeof payload === 'object' && payload !== null && Array.isArray((payload as Record<string, unknown>)[key])) {
			return (payload as Record<string, unknown>)[key] as T[];
		}
		return [];
	}

	async function fetchArchive(): Promise<ArchiveData> {
		const [info, overview, sessionPayload, approvals] = await Promise.all([
			fetch(clientApiPath`/api/asesmen/events/${eventId}`).then((response) =>
				readClientApiData<EventInfo>(response, 'Gagal memuat kegiatan asesmen')
			),
			optionalApiData<EventOverview | null>(clientApiPath`/api/asesmen/events/${eventId}/overview`, null),
			optionalApiData<unknown>(clientApiPath`/api/asesmen/events/${eventId}/sessions`, []),
			listApprovals('event', eventId).catch(() => [])
		]);
		return {
			info,
			overview,
			sessions: parseArrayPayload<EventSession>(sessionPayload, 'sessions'),
			approvals
		};
	}

	function loadArchive() {
		archivePromise = fetchArchive();
	}

	function retryArchive(reset?: () => void) {
		reset?.();
		loadArchive();
	}

	function archiveErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat arsip kegiatan asesmen';
	}

	function handleArchiveRenderError(error: unknown) {
		console.error('Arsip kegiatan asesmen belum dapat ditampilkan', error);
	}

	function approved(approvals: AssessmentApprovalRecord[], approvalType: ArchiveApprovalType | string) {
		return approvals.find((row) => row.approval_type === approvalType && row.status === 'approved');
	}

	function formatDate(value: string | null | undefined) {
		if (!value) return '—';
		return `${new Date(value).toLocaleString('id-ID', {
			timeZone: 'Asia/Makassar',
			day: '2-digit',
			month: 'short',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		})} WITA`;
	}

	function count(value: unknown) {
		const n = Number(value ?? 0);
		return Number.isFinite(n) ? n : 0;
	}

	function statusBadgeClass(ready: boolean) {
		return ready
			? 'border-success/20 bg-success/10 text-success'
			: 'border-warning/30 bg-warning/10 text-warning';
	}

	onMount(() => loadArchive());
</script>

<svelte:head>
	<title>Arsip Kegiatan Asesmen</title>
</svelte:head>

<AsyncContent promise={archivePromise} onerror={handleArchiveRenderError}>
	{#snippet pending()}
		<div class="mx-auto max-w-7xl space-y-6 p-6">
			<Skeleton class="h-10 w-72" />
			<Skeleton class="h-20 rounded-xl" />
			<Skeleton class="h-56 rounded-xl" />
		</div>
	{/snippet}

	{#snippet failed(error, reset)}
		<div class="mx-auto max-w-7xl p-6">
			<RecoveryPanel
				title="Arsip Kegiatan Belum Tersaji"
				message={archiveErrorMessage(error)}
				onRetry={() => retryArchive(reset)}
			/>
		</div>
	{/snippet}

	{#snippet children(value)}
		{@const data = value as ArchiveData}
		{@const finalApproval = approved(data.approvals, 'final_archive')}
		{@const checklistRows: ArchiveChecklistRow[] = [
			{
				id: 'cards',
				title: 'Kartu Ujian & Token Distribusi',
				helper: 'Distribusi terbatas sebelum ujian melalui Kartu Ujian dan Dokumen & Cetak.',
				href: resolve(`/asesmen/kegiatan/${eventId}/cetak`),
				status: count(data.overview?.card_count) > 0 ? 'Dokumen tersedia' : 'Perlu dicek',
				coverage: `${count(data.overview?.card_count)} kartu · ${count(data.overview?.room_count)} ruang`,
				ready: count(data.overview?.card_count) > 0,
				icon: ClipboardCheckIcon
			},
			{
				id: 'minutes',
				title: 'Daftar Hadir & BA Sesi',
				helper: 'Cetak BA dari masing-masing sesi setelah pengawasan dan pengiriman jawaban selesai.',
				href: resolve(`/asesmen/sesi?event_id=${eventId}`),
				status: data.sessions.length > 0 ? 'Sesi siap dibuka' : 'Belum ada sesi',
				coverage: `${data.sessions.length} sesi · ${count(data.overview?.submitted_count)} kiriman`,
				ready: data.sessions.length > 0,
				icon: FileTextIcon
			},
			{
				id: 'results',
				title: 'Rekap Hasil & Verifikasi',
				helper: 'Pastikan nilai, kiriman peserta, dan analisis butir sudah diperiksa sebelum final arsip.',
				href: resolve(`/asesmen/kegiatan/${eventId}#hasil`),
				status: count(data.overview?.result_count) > 0 ? 'Rekap tersedia' : 'Belum ada rekap',
				coverage: `${count(data.overview?.result_count)} hasil · ${count(data.overview?.submitted_count)} kiriman`,
				ready: count(data.overview?.result_count) > 0,
				icon: ArchiveIcon
			},
			{
				id: 'incidents',
				title: 'Rekap Insiden Pengawasan',
				helper: 'Buka panel pengawasan saat perlu meninjau catatan ruang bermasalah atau kejadian penting.',
				href: resolve('/asesmen/ruang-saya'),
				status: 'Siap dibuka',
				coverage: 'Panel pengawasan · kejadian ruang',
				ready: true,
				icon: BadgeAlertIcon
			},
			{
				id: 'approvals',
				title: 'Riwayat Pengesahan',
				helper: 'Paket, peserta/ruang, token, hasil, dan final arsip dicatat sebagai pengesahan formal.',
				href: resolve(`/asesmen/kegiatan/${eventId}`),
				status: data.approvals.some((row) => row.status === 'approved') ? 'Ada catatan' : 'Belum ada pengesahan',
				coverage: `${data.approvals.filter((row) => row.status === 'approved').length} pengesahan aktif`,
				ready: data.approvals.some((row) => row.status === 'approved'),
				icon: ShieldCheckIcon
			}
		]}
		{@const approvalRows: ApprovalRow[] = archiveApprovalTypes.map((approvalType) => {
			const record = approved(data.approvals, approvalType);
			return {
				id: approvalType,
				title: assessmentApprovalLabels[approvalType] ?? approvalType,
				status: record ? 'Disahkan' : 'Belum',
				actor: record
					? `${record.approved_by_display_name || record.approved_by_username || 'Admin'} · ${formatDate(record.approved_at)}`
					: 'Belum ada catatan pengesahan aktif.',
				ready: Boolean(record)
			};
		})}
		<div class="mx-auto max-w-7xl space-y-6 p-6">
			<section class="space-y-4">
				<div class="space-y-2">
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-primary">Asesmen · Arsip</p>
					<h1 class="text-2xl font-semibold tracking-tight text-foreground">Arsip & Penutupan Kegiatan</h1>
					<p class="max-w-3xl text-sm leading-6 text-muted-foreground">
						{data.info.title} · {data.info.academic_year_name}. Halaman ini dirapikan menjadi checklist
						penutupan: dokumen, BA sesi, pengesahan, dan jejak hasil akhir.
					</p>
				</div>
				<div class="flex flex-wrap gap-2">
					<Button variant="outline" href={resolve(`/asesmen/kegiatan/${eventId}`)}>Kembali ke Kegiatan</Button>
					<Button variant="outline" href={resolve(`/asesmen/kegiatan/${eventId}/cetak`)}>Dokumen & Cetak</Button>
					<Button variant="outline" href={resolve(`/asesmen/kegiatan/${eventId}/exam-cards`)}>Kartu Ujian</Button>
					<Button variant="outline" href={resolve(`/asesmen/kegiatan/${eventId}/pengawas-cards`)}>Lembar Pengawas</Button>
					{#if finalApproval}
						<Badge variant="outline" class="border-success/20 bg-success/10 text-success">Final arsip sudah disahkan</Badge>
					{:else}
						<Badge variant="outline" class="border-warning/30 bg-warning/10 text-warning">Final arsip belum disahkan</Badge>
					{/if}
				</div>
			</section>

			<div class="grid gap-3 md:grid-cols-4">
				<div class="rounded-xl border border-border bg-card px-4 py-3">
					<p class="text-[11px] font-semibold uppercase tracking-[0.16em] text-muted-foreground">Paket</p>
					<p class="mt-1 text-2xl font-semibold text-foreground">{count(data.overview?.package_count)}</p>
				</div>
				<div class="rounded-xl border border-border bg-card px-4 py-3">
					<p class="text-[11px] font-semibold uppercase tracking-[0.16em] text-muted-foreground">Sesi</p>
					<p class="mt-1 text-2xl font-semibold text-foreground">{data.sessions.length || count(data.overview?.session_count)}</p>
				</div>
				<div class="rounded-xl border border-border bg-card px-4 py-3">
					<p class="text-[11px] font-semibold uppercase tracking-[0.16em] text-muted-foreground">Ruang</p>
					<p class="mt-1 text-2xl font-semibold text-foreground">{count(data.overview?.room_count)}</p>
				</div>
				<div class="rounded-xl border border-primary/20 bg-primary/10 px-4 py-3">
					<p class="text-[11px] font-semibold uppercase tracking-[0.16em] text-primary">Hasil</p>
					<p class="mt-1 text-2xl font-semibold text-primary">{count(data.overview?.result_count)}</p>
				</div>
			</div>

			<MicroActionTable
				title="Checklist penutupan"
				description="Buka hanya dokumen yang dibutuhkan untuk menutup kegiatan. Tidak perlu lagi menyisir banyak halaman."
				columns={checklistColumns}
				rows={checklistRows}
				rowKey={(row) => (row as ArchiveChecklistRow).id}
				tableClass="min-w-[920px]"
			>
				{#snippet cell(row, column)}
					{@const item = row as ArchiveChecklistRow}
					{#if column.key === 'document'}
						<div class="flex items-start gap-3">
							<div class={`grid size-10 shrink-0 place-items-center rounded-xl ${item.ready ? 'bg-primary/10 text-primary' : 'bg-warning/10 text-warning'}`}>
								<item.icon class="size-4" />
							</div>
							<div class="space-y-1">
								<p class="font-semibold text-foreground">{item.title}</p>
								<p class="text-xs leading-5 text-muted-foreground">{item.helper}</p>
							</div>
						</div>
					{:else if column.key === 'status'}
						<Badge variant="outline" class={statusBadgeClass(item.ready)}>{item.status}</Badge>
					{:else}
						<p class="text-xs leading-5 text-muted-foreground">{item.coverage}</p>
					{/if}
				{/snippet}
				{#snippet actions(row)}
					{@const item = row as ArchiveChecklistRow}
					<a href={item.href} class="inline-flex items-center rounded-md bg-primary px-3 py-2 text-xs font-semibold text-primary-foreground hover:bg-primary/90">Buka</a>
				{/snippet}
			</MicroActionTable>

			<MicroActionTable
				title="Status pengesahan"
				description="Pantau pengesahan formal sebelum menutup arsip."
				columns={approvalColumns}
				rows={approvalRows}
				rowKey={(row) => (row as ApprovalRow).id}
				tableClass="min-w-[920px]"
			>
				{#snippet cell(row, column)}
					{@const item = row as ApprovalRow}
					{#if column.key === 'approval'}
						<p class="font-semibold text-foreground">{item.title}</p>
					{:else if column.key === 'status'}
						<Badge variant="outline" class={statusBadgeClass(item.ready)}>{item.status}</Badge>
					{:else}
						<p class="text-xs leading-5 text-muted-foreground">{item.actor}</p>
					{/if}
				{/snippet}
			</MicroActionTable>

			{#if finalApproval}
				<div class="rounded-xl border border-success/20 bg-success/10 p-4 text-sm text-success">
					<CheckCircle2Icon class="mb-2 size-5" />
					Kegiatan sudah difinalkan untuk arsip. Gunakan checklist di atas bila perlu membuka ulang
					dokumen pendukung.
				</div>
			{/if}

			<MicroActionTable
				title="Sesi & BA"
				description="Jalur cepat untuk membuka BA per sesi tanpa tenggelam di halaman detail lain."
				columns={sessionColumns}
				rows={data.sessions}
				rowKey={(row) => (row as EventSession).id}
				tableClass="min-w-[920px]"
				emptyTitle="Belum ada sesi untuk kegiatan ini."
				emptyDescription="Buat sesi terlebih dahulu sebelum menutup arsip."
			>
				{#snippet cell(row, column)}
					{@const session = row as EventSession}
					{#if column.key === 'session'}
						<p class="font-semibold text-foreground">{session.title}</p>
					{:else if column.key === 'status'}
						<Badge variant="outline" class="border-border bg-card text-muted-foreground">{session.status ?? '—'}</Badge>
					{:else}
						<p class="text-xs leading-5 text-muted-foreground">{session.participant_count ?? 0} peserta · {session.room_count ?? 0} ruang</p>
					{/if}
				{/snippet}
				{#snippet actions(row)}
					{@const session = row as EventSession}
					<a href={resolve(`/asesmen/sesi/${session.id}/minutes`)} class="inline-flex items-center rounded-md border border-border bg-card px-3 py-2 text-xs font-semibold text-foreground hover:bg-muted">BA Sesi</a>
				{/snippet}
			</MicroActionTable>
		</div>
	{/snippet}
</AsyncContent>
