<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { toast } from '$lib/components/ui/sonner';
	import { onMount } from 'svelte';
	import { clientApiPath, readClientApiData, readClientJson } from '$lib/client/api';

	type SessionDetail = {
		id: string;
		assignment_id: string;
		tanggal: string;
		pertemuan_ke: number;
		materi: string;
		kegiatan: string;
		catatan: string;
		guru_hadir: boolean;
		class_name: string;
		class_code: string;
		subject_name: string;
		teacher_name: string;
		teacher_employee_id: string;
	};

	type Attendance = {
		id: string;
		session_id: string;
		student_id: string;
		status: string;
		catatan: string;
		nis: string;
		nisn: string;
		nama: string;
		gender: string;
	};

	type JournalDetail = {
		session: SessionDetail;
		attendances: Attendance[];
	};

	const sessionId = $derived($page.params.id ?? '');

	let detailPromise = $state<Promise<JournalDetail> | null>(null);
	let detailRequestId = 0;
	let detail = $state<JournalDetail | null>(null);

	let editMode = $state(false);
	let editBusy = $state(false);
	let saveBusy = $state(false);

	let editMateri = $state('');
	let editKegiatan = $state('');
	let editCatatan = $state('');
	let editGuruHadir = $state(true);

	// attendance state: student_id -> {status, catatan}
	let attendanceState = $state<Record<string, { status: string; catatan: string }>>({});

	function isRecord(value: unknown): value is Record<string, unknown> {
		return typeof value === 'object' && value !== null;
	}

	function applyDetail(data: JournalDetail) {
		detail = data;
		editMateri = data.session.materi;
		editKegiatan = data.session.kegiatan;
		editCatatan = data.session.catatan;
		editGuruHadir = data.session.guru_hadir;
		const map: Record<string, { status: string; catatan: string }> = {};
		for (const a of data.attendances) {
			map[a.student_id] = { status: a.status, catatan: a.catatan };
		}
		attendanceState = map;
	}

	async function fetchDetail(): Promise<JournalDetail> {
		const res = await fetch(clientApiPath`/api/journal/sessions/${sessionId}`);
		return readClientApiData<JournalDetail>(res, 'Gagal memuat detail sesi');
	}

	function loadDetail() {
		const requestId = ++detailRequestId;
		detailPromise = fetchDetail()
			.then((data) => {
				if (requestId === detailRequestId) {
					applyDetail(data);
					return data;
				}
				return detail ?? data;
			})
			.catch((error: unknown) => {
				if (requestId === detailRequestId) throw error;
				if (detail) return detail;
				throw error;
			});
	}

	function retryDetail(reset?: () => void) {
		reset?.();
		loadDetail();
	}

	function detailErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat detail sesi';
	}

	function handleDetailRenderError(error: unknown) {
		console.error('Journal detail render failed', error);
	}

	function mutationErrorMessage(error: unknown, fallback: string) {
		if (error instanceof Error && error.message.trim() && !error.message.toLowerCase().includes('fetch')) return error.message;
		return fallback;
	}

	onMount(() => {
		loadDetail();
	});

	async function saveSession() {
		if (!detail) return;
		editBusy = true;
		try {
			const res = await fetch(clientApiPath`/api/journal/sessions/${sessionId}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					materi: editMateri,
					kegiatan: editKegiatan,
					catatan: editCatatan,
					guru_hadir: editGuruHadir
				})
			});
			const payload = await readClientApiData<Partial<SessionDetail> | { session?: Partial<SessionDetail> }>(res, 'Gagal menyimpan');
			const payloadRecord = isRecord(payload) ? payload as Record<string, unknown> : null;
			const savedSession = payloadRecord && isRecord(payloadRecord.session) ? payloadRecord.session as Partial<SessionDetail> : payload as Partial<SessionDetail>;
			detail = { ...detail, session: { ...detail.session, ...savedSession } };
			detailPromise = Promise.resolve(detail);
			editMode = false;
			toast.success('Sesi berhasil diperbarui');
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Terjadi kesalahan jaringan'));
		} finally {
			editBusy = false;
		}
	}

	async function saveAttendances() {
		saveBusy = true;
		try {
			const entries = Object.entries(attendanceState).map(([student_id, v]) => ({
				student_id,
				status: v.status,
				catatan: v.catatan
			}));
			const res = await fetch(clientApiPath`/api/journal/sessions/${sessionId}/attendances`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ entries })
			});
			await readClientJson<unknown>(res);
			toast.success('Kehadiran berhasil disimpan');
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Terjadi kesalahan jaringan'));
		} finally {
			saveBusy = false;
		}
	}

	function setStatus(studentId: string, status: string) {
		attendanceState = {
			...attendanceState,
			[studentId]: { ...attendanceState[studentId], status }
		};
	}

	function formatTanggal(raw: string) {
		if (!raw) return '-';
		const d = new Date(raw);
		return d.toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' });
	}

	const statusColors: Record<string, string> = {
		hadir: 'bg-green-600 text-white hover:bg-green-700',
		sakit: 'bg-yellow-500 text-white hover:bg-yellow-600',
		izin: 'bg-sky-500 text-white hover:bg-sky-600',
		alpha: 'bg-red-600 text-white hover:bg-red-700'
	};

	const statusInactive = 'bg-gray-100 text-gray-600 hover:bg-gray-200';
</script>

<div class="container mx-auto max-w-5xl space-y-6 p-6">
	<!-- Back button -->
	<Button variant="ghost" size="sm" onclick={() => goto(resolve('/journal'))}>
		← Kembali ke Jurnal
	</Button>

	<AsyncContent promise={detailPromise} onerror={handleDetailRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<Skeleton class="h-32 w-full" />
				<Skeleton class="h-64 w-full" />
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Detail Jurnal Belum Tersaji"
				message={detailErrorMessage(error)}
				onRetry={() => retryDetail(reset)}
			/>
		{/snippet}

		{#snippet children(value)}
			{@const currentDetail = detail ?? (value as JournalDetail)}
		<!-- Session Header Card -->
		<Card.Root>
			<Card.Header>
				<div class="flex items-start justify-between">
					<div class="space-y-1">
						<div class="flex items-center gap-2">
							<Badge variant="outline" class="text-base font-semibold">Pertemuan ke-{currentDetail.session.pertemuan_ke}</Badge>
							{#if currentDetail.session.guru_hadir}
								<Badge class="bg-green-100 text-green-800 hover:bg-green-100">Guru Hadir</Badge>
							{:else}
								<Badge class="bg-red-100 text-red-800 hover:bg-red-100">Guru Tidak Hadir</Badge>
							{/if}
						</div>
						<p class="text-lg font-medium text-gray-800">{formatTanggal(currentDetail.session.tanggal)}</p>
						<p class="text-sm text-gray-500">
							{currentDetail.session.class_name} ({currentDetail.session.class_code}) &nbsp;·&nbsp;
							{currentDetail.session.subject_name} &nbsp;·&nbsp;
							{currentDetail.session.teacher_name}
						</p>
					</div>
					<Button variant="outline" size="sm" onclick={() => (editMode = !editMode)}>
						{editMode ? 'Batal' : 'Edit'}
					</Button>
				</div>
			</Card.Header>
			<Card.Content class="space-y-4">
				{#if editMode}
					<div class="space-y-3">
						<div class="space-y-1">
							<label for="edit-materi" class="text-sm font-medium text-gray-700">Materi</label>
							<Textarea id="edit-materi" bind:value={editMateri} rows={3} />
						</div>
						<div class="space-y-1">
							<label for="edit-kegiatan" class="text-sm font-medium text-gray-700">Kegiatan</label>
							<Textarea id="edit-kegiatan" bind:value={editKegiatan} rows={3} />
						</div>
						<div class="space-y-1">
							<label for="edit-catatan" class="text-sm font-medium text-gray-700">Catatan</label>
							<Textarea id="edit-catatan" bind:value={editCatatan} rows={2} />
						</div>
						<div class="flex items-center gap-3">
							<input
								id="edit-guru-hadir"
								type="checkbox"
								bind:checked={editGuruHadir}
								class="h-4 w-4 rounded border-gray-300 text-green-600"
							/>
							<label for="edit-guru-hadir" class="text-sm font-medium text-gray-700">Guru hadir mengajar</label>
						</div>
						<div class="flex justify-end">
							<LoadingButton loading={editBusy} onclick={() => void saveSession()}>Simpan Perubahan</LoadingButton>
						</div>
					</div>
				{:else}
					<div class="grid gap-3 text-sm">
						<div>
							<p class="font-medium text-gray-700">Materi</p>
							<p class="text-gray-600 whitespace-pre-wrap">{currentDetail.session.materi || '–'}</p>
						</div>
						<div>
							<p class="font-medium text-gray-700">Kegiatan</p>
							<p class="text-gray-600 whitespace-pre-wrap">{currentDetail.session.kegiatan || '–'}</p>
						</div>
						{#if currentDetail.session.catatan}
							<div>
								<p class="font-medium text-gray-700">Catatan</p>
								<p class="text-gray-600 whitespace-pre-wrap">{currentDetail.session.catatan}</p>
							</div>
						{/if}
					</div>
				{/if}
			</Card.Content>
		</Card.Root>

		<!-- Attendance Section -->
		<Card.Root>
			<Card.Header>
				<Card.Title class="text-base">Kehadiran Siswa</Card.Title>
				<Card.Description>
					{currentDetail.attendances.length} siswa · Klik status untuk mengubah
				</Card.Description>
			</Card.Header>
			<Card.Content class="p-0">
				{#if currentDetail.attendances.length === 0}
					<div class="p-8 text-center text-sm text-gray-500">Tidak ada siswa yang terdaftar.</div>
				{:else}
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head class="w-12">No</Table.Head>
								<Table.Head>Nama Siswa</Table.Head>
								<Table.Head class="w-24">NIS</Table.Head>
								<Table.Head class="w-52">Status</Table.Head>
								<Table.Head>Catatan</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each currentDetail.attendances as att, i (att.student_id)}
								{@const cur = attendanceState[att.student_id] ?? { status: att.status, catatan: att.catatan }}
								<Table.Row>
									<Table.Cell class="text-gray-500">{i + 1}</Table.Cell>
									<Table.Cell class="font-medium">{att.nama}</Table.Cell>
									<Table.Cell class="text-sm text-gray-500">{att.nis}</Table.Cell>
									<Table.Cell>
										<div class="flex gap-1">
											{#each ['hadir', 'sakit', 'izin', 'alpha'] as s (s)}
												<button
													class="rounded px-2 py-0.5 text-xs font-medium transition-colors {cur.status === s ? statusColors[s] : statusInactive}"
													onclick={() => setStatus(att.student_id, s)}
												>
													{s.charAt(0).toUpperCase()}
												</button>
											{/each}
										</div>
									</Table.Cell>
									<Table.Cell>
										<input
											type="text"
											value={cur.catatan}
											oninput={(e) => {
												attendanceState = {
													...attendanceState,
													[att.student_id]: { ...cur, catatan: (e.target as HTMLInputElement).value }
												};
											}}
											placeholder="Catatan (opsional)"
											class="w-full rounded border border-gray-200 px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-green-600"
										/>
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				{/if}
			</Card.Content>
			{#if currentDetail.attendances.length > 0}
				<div class="flex justify-end border-t border-gray-100 p-4">
					<LoadingButton loading={saveBusy} onclick={() => void saveAttendances()}>Simpan Kehadiran</LoadingButton>
				</div>
			{/if}
		</Card.Root>
		{/snippet}
	</AsyncContent>
</div>
