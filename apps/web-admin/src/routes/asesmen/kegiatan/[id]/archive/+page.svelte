<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import ArchiveIcon from '@lucide/svelte/icons/archive';
	import CheckCircle2Icon from '@lucide/svelte/icons/check-circle-2';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import ShieldCheckIcon from '@lucide/svelte/icons/shield-check';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
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

	const eventId = page.params.id ?? '';
	const archiveApprovalTypes = ['package_ready', 'participants_rooms_ready', 'tokens_cards_ready', 'results_verified', 'final_archive'] as const;
	type ArchiveApprovalType = (typeof archiveApprovalTypes)[number];
	type ArchiveChecklistItem = { label: string; helper: string; href: string; ready: boolean };
	const checklistColumns = [
		{ key: 'item', label: 'Dokumen', class: 'min-w-[16rem]' },
		{ key: 'status', label: 'Status' },
	];
	const approvalColumns = [
		{ key: 'approval', label: 'Pengesahan', class: 'min-w-[16rem]' },
		{ key: 'status', label: 'Status' },
		{ key: 'actor', label: 'Aktor/Waktu' },
	];
	const sessionColumns = [
		{ key: 'session', label: 'Sesi', class: 'min-w-[14rem]' },
		{ key: 'status', label: 'Status' },
		{ key: 'participants', label: 'Peserta' },
		{ key: 'rooms', label: 'Ruang' },
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
			fetch(clientApiPath`/api/asesmen/events/${eventId}`).then((response) => readClientApiData<EventInfo>(response, 'Gagal memuat kegiatan asesmen')),
			optionalApiData<EventOverview | null>(clientApiPath`/api/asesmen/events/${eventId}/overview`, null),
			optionalApiData<unknown>(clientApiPath`/api/asesmen/events/${eventId}/sessions`, []),
			listApprovals('event', eventId).catch(() => []),
		]);
		return { info, overview, sessions: parseArrayPayload<EventSession>(sessionPayload, 'sessions'), approvals };
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
		return `${new Date(value).toLocaleString('id-ID', { timeZone: 'Asia/Makassar', day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' })} WITA`;
	}

	function count(value: unknown) {
		const n = Number(value ?? 0);
		return Number.isFinite(n) ? n : 0;
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
			<div class="grid gap-4 md:grid-cols-3">
				{#each Array.from({ length: 6 }) as _, index (`archive-skeleton-${index}`)}
					<Skeleton class="h-32 rounded-2xl" />
				{/each}
			</div>
		</div>
	{/snippet}

	{#snippet failed(error, reset)}
		<div class="mx-auto max-w-7xl p-6">
			<RecoveryPanel title="Arsip Kegiatan Belum Tersaji" message={archiveErrorMessage(error)} onRetry={() => retryArchive(reset)} />
		</div>
	{/snippet}

	{#snippet children(value)}
		{@const data = value as ArchiveData}
		{@const finalApproval = approved(data.approvals, 'final_archive')}
		<div class="mx-auto max-w-7xl space-y-6 p-6">
			<div class="flex flex-col gap-4 rounded-3xl bg-gradient-to-br from-slate-950 via-slate-900 to-emerald-900 p-6 text-white shadow-xl lg:flex-row lg:items-center lg:justify-between">
				<div class="space-y-3">
					<Badge class="w-fit bg-white/15 text-white hover:bg-white/15">Arsip Digital CBT</Badge>
					<div>
						<h1 class="text-3xl font-black tracking-tight">Arsip Kegiatan Asesmen</h1>
						<p class="mt-2 max-w-3xl text-sm text-white/75">{data.info.title} · {data.info.academic_year_name}. Gunakan halaman ini untuk menutup kegiatan, mencetak dokumen operasional, dan memastikan pengesahan formal tercatat.</p>
					</div>
					<p class="rounded-2xl border border-amber-200/30 bg-amber-200/10 px-4 py-3 text-sm text-amber-50">Dokumen arsip final tidak menampilkan token ujian mentah secara default. Token lengkap hanya untuk distribusi terbatas sebelum ujian melalui kartu ujian/berkas pengawas.</p>
				</div>
				<div class="flex flex-wrap gap-2">
					<Button variant="secondary" href={resolve(`/asesmen/kegiatan/${eventId}`)}>Kembali ke Kegiatan</Button>
					<Button variant="secondary" href={resolve(`/asesmen/kegiatan/${eventId}/exam-cards`)}>Kartu Ujian</Button>
				</div>
			</div>

			<div class="grid gap-4 md:grid-cols-4">
				<Card.Root><Card.Content class="p-5"><p class="text-xs font-bold uppercase text-muted-foreground">Paket</p><p class="mt-2 text-3xl font-black">{count(data.overview?.package_count)}</p></Card.Content></Card.Root>
				<Card.Root><Card.Content class="p-5"><p class="text-xs font-bold uppercase text-muted-foreground">Sesi</p><p class="mt-2 text-3xl font-black">{data.sessions.length || count(data.overview?.session_count)}</p></Card.Content></Card.Root>
				<Card.Root><Card.Content class="p-5"><p class="text-xs font-bold uppercase text-muted-foreground">Ruang</p><p class="mt-2 text-3xl font-black">{count(data.overview?.room_count)}</p></Card.Content></Card.Root>
				<Card.Root><Card.Content class="p-5"><p class="text-xs font-bold uppercase text-muted-foreground">Hasil</p><p class="mt-2 text-3xl font-black">{count(data.overview?.result_count)}</p></Card.Content></Card.Root>
			</div>

			<div class="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
				<Card.Root>
					<Card.Header><Card.Title class="flex items-center gap-2"><ArchiveIcon class="h-5 w-5" /> Checklist Arsip</Card.Title></Card.Header>
					<Card.Content class="space-y-3">
						{@const checklist: ArchiveChecklistItem[] = [
							{ label: 'Kartu ujian/token distribusi', helper: 'Untuk distribusi terbatas sebelum ujian.', href: `/asesmen/kegiatan/${eventId}/exam-cards`, ready: count(data.overview?.card_count) > 0 },
							{ label: 'Daftar hadir dan BA sesi', helper: 'Cetak dari masing-masing sesi/ruang.', href: `/asesmen/sesi?event_id=${eventId}`, ready: data.sessions.length > 0 },
							{ label: 'Rekap hasil', helper: 'Pastikan nilai dan kiriman peserta sudah diperiksa.', href: `/asesmen/kegiatan/${eventId}#hasil`, ready: count(data.overview?.result_count) > 0 },
							{ label: 'Rekap insiden pengawasan', helper: 'Buka panel pengawasan/rekap untuk ruang yang punya kejadian perhatian.', href: '/asesmen/pengawasan', ready: true },
							{ label: 'Riwayat pengesahan', helper: 'Pengesahan paket, peserta/ruang, token/kartu, hasil, dan final arsip.', href: `/asesmen/kegiatan/${eventId}`, ready: data.approvals.some((row) => row.status === 'approved') },
						]}
						{#each checklist as item}
							<a class="flex items-start justify-between gap-4 rounded-2xl border p-4 transition hover:border-primary/40 hover:bg-primary/5" href={item.href}>
								<div>
									<p class="font-bold">{item.label}</p>
									<p class="mt-1 text-sm text-muted-foreground">{item.helper}</p>
								</div>
								<Badge variant={item.ready ? 'default' : 'secondary'}>{item.ready ? 'Ada' : 'Perlu cek'}</Badge>
							</a>
						{/each}
					</Card.Content>
				</Card.Root>

				<Card.Root>
					<Card.Header><Card.Title class="flex items-center gap-2"><ShieldCheckIcon class="h-5 w-5" /> Status Pengesahan</Card.Title></Card.Header>
					<Card.Content class="space-y-3">
						{#each archiveApprovalTypes as approvalType}
							{@const record = approved(data.approvals, approvalType)}
							<div class="rounded-2xl border p-4">
								<div class="flex items-center justify-between gap-3">
									<p class="font-bold">{assessmentApprovalLabels[approvalType] ?? approvalType}</p>
									<Badge variant={record ? 'default' : 'secondary'}>{record ? 'Disahkan' : 'Belum'}</Badge>
								</div>
								<p class="mt-1 text-xs text-muted-foreground">{record ? `${record.approved_by_display_name || record.approved_by_username || 'Admin'} · ${formatDate(record.approved_at)}` : 'Belum ada catatan pengesahan aktif.'}</p>
							</div>
						{/each}
						{#if finalApproval}
							<div class="rounded-2xl border border-emerald-200 bg-emerald-50 p-4 text-emerald-900"><CheckCircle2Icon class="mb-2 h-5 w-5" /><p class="font-bold">Kegiatan sudah difinalkan untuk arsip.</p></div>
						{/if}
					</Card.Content>
				</Card.Root>
			</div>

			<Card.Root>
				<Card.Header><Card.Title class="flex items-center gap-2"><FileTextIcon class="h-5 w-5" /> Sesi dan Dokumen BA</Card.Title></Card.Header>
				<Card.Content>
					{#if data.sessions.length === 0}
						<div class="rounded-2xl border border-dashed p-6 text-sm text-muted-foreground">Belum ada sesi untuk kegiatan ini. Buat sesi terlebih dahulu sebelum menutup arsip.</div>
					{:else}
						<div class="overflow-x-auto">
							<Table.Root>
								<Table.Header><Table.Row><Table.Head>Sesi</Table.Head><Table.Head>Status</Table.Head><Table.Head>Peserta</Table.Head><Table.Head>Ruang</Table.Head><Table.Head class="text-right">Dokumen</Table.Head></Table.Row></Table.Header>
								<Table.Body>
									{#each data.sessions as session}
										<Table.Row>
											<Table.Cell class="font-semibold">{session.title}</Table.Cell>
											<Table.Cell><Badge variant="outline">{session.status ?? '—'}</Badge></Table.Cell>
											<Table.Cell>{session.participant_count ?? 0}</Table.Cell>
											<Table.Cell>{session.room_count ?? 0}</Table.Cell>
											<Table.Cell class="text-right"><Button size="sm" variant="outline" href={resolve(`/asesmen/sesi/${session.id}/minutes`)}>BA Sesi</Button></Table.Cell>
										</Table.Row>
									{/each}
								</Table.Body>
							</Table.Root>
						</div>
					{/if}
				</Card.Content>
			</Card.Root>
		</div>
	{/snippet}
</AsyncContent>
