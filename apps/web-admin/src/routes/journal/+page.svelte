<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { toast } from '$lib/components/ui/sonner';
	import { confirmAction } from '$lib/confirm-dialog';
	import { clientApiPath, clientApiPathWithQuery, readClientApiData, readClientJson } from '$lib/client/api';

	type Assignment = {
		id: string;
		class_name: string;
		class_code: string;
		subject_name: string;
		subject_code: string;
		teacher_name: string;
	};

	type Session = {
		id: string;
		assignment_id: string;
		tanggal: string;
		pertemuan_ke: number;
		materi: string;
		kegiatan: string;
		catatan: string;
		guru_hadir: boolean;
		class_name: string;
		subject_name: string;
		teacher_name: string;
	};

	type AttendanceSummary = {
		student_id: string;
		nis: string;
		nisn: string;
		nama: string;
		total_pertemuan: number;
		hadir: number;
		sakit: number;
		izin: number;
		alpha: number;
	};

	type Overview = {
		assignments: Assignment[];
		sessions: Session[];
		summary: AttendanceSummary[];
	};

	type CreateSessionResponse = { session?: { id?: string }; id?: string };

	let overviewPromise = $state<Promise<Overview> | null>(null);
	let overviewRequestId = 0;
	let assignments = $state<Assignment[]>([]);
	let assignmentId = $state('');
	let activeTab = $state<'sessions' | 'rekap'>('sessions');

	let createOpen = $state(false);
	let createBusy = $state(false);
	let deleteBusy = $state<Record<string, boolean>>({});

	let newTanggal = $state('');
	let newMateri = $state('');
	let newKegiatan = $state('');
	let newCatatan = $state('');
	let newGuruHadir = $state(true);

	function isRecord(value: unknown): value is Record<string, unknown> {
		return typeof value === 'object' && value !== null;
	}

	function applyOverview(data: Overview) {
		assignments = data.assignments ?? [];
	}

	async function fetchOverview(selectedAssignmentId: string): Promise<Overview> {
		const params = new URLSearchParams();
		if (selectedAssignmentId) params.set('assignment_id', selectedAssignmentId);
		const res = await fetch(clientApiPathWithQuery('/api/journal', params));
		return readClientApiData<Overview>(res, 'Gagal memuat data jurnal');
	}

	function loadOverview(selectedAssignmentId: string) {
		const requestId = ++overviewRequestId;
		overviewPromise = fetchOverview(selectedAssignmentId)
			.then((data) => {
				if (requestId === overviewRequestId) {
					applyOverview(data);
					return data;
				}
				return { assignments, sessions: [], summary: [] };
			})
			.catch((error: unknown) => {
				if (requestId === overviewRequestId) throw error;
				return { assignments, sessions: [], summary: [] };
			});
	}

	async function refreshOverview() {
		const requestId = ++overviewRequestId;
		const data = await fetchOverview(assignmentId);
		if (requestId === overviewRequestId) {
			applyOverview(data);
			overviewPromise = Promise.resolve(data);
		}
	}

	function retryOverview(reset?: () => void) {
		reset?.();
		loadOverview(assignmentId);
	}

	function overviewErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat data jurnal';
	}

	function handleOverviewRenderError(error: unknown) {
		console.error('Journal overview render failed', error);
	}

	function mutationErrorMessage(error: unknown, fallback: string) {
		if (error instanceof Error && error.message.trim() && !error.message.toLowerCase().includes('fetch')) return error.message;
		return fallback;
	}

	async function refreshOverviewAfterMutation() {
		try {
			await refreshOverview();
		} catch (error) {
			toast.error(overviewErrorMessage(error));
		}
	}

	onMount(() => {
		loadOverview(assignmentId);
	});

	async function createSession() {
		if (!assignmentId) { toast.error('Pilih kelas-mapel terlebih dahulu'); return; }
		if (!newTanggal) { toast.error('Tanggal wajib diisi'); return; }
		createBusy = true;
		try {
			const res = await fetch('/api/journal/sessions', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					assignment_id: assignmentId,
					tanggal: newTanggal,
					materi: newMateri,
					kegiatan: newKegiatan,
					catatan: newCatatan,
					guru_hadir: newGuruHadir
				})
			});
			const payloadBody = await readClientApiData<CreateSessionResponse>(res, res.status === 409 ? 'Pertemuan pada tanggal ini sudah ada' : 'Gagal membuat pertemuan');
			const session = isRecord(payloadBody) && isRecord(payloadBody.session) ? payloadBody.session : null;
			const sessionId = typeof session?.id === 'string' ? session.id : typeof payloadBody.id === 'string' ? payloadBody.id : '';
			if (!sessionId) throw new Error('Respons sesi baru tidak lengkap');
			createOpen = false;
			newTanggal = '';
			newMateri = '';
			newKegiatan = '';
			newCatatan = '';
			newGuruHadir = true;
			toast.success('Sesi berhasil dibuat');
			goto(clientApiPath`/journal/${sessionId}`);
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Terjadi kesalahan jaringan'));
		} finally {
			createBusy = false;
		}
	}

	async function deleteSession(id: string) {
		if (!(await confirmAction({
			title: 'Hapus Pertemuan Jurnal',
			message: 'Hapus pertemuan ini? Semua data kehadiran akan ikut terhapus.',
			confirmLabel: 'Hapus Pertemuan',
			tone: 'danger'
		}))) return;
		deleteBusy = { ...deleteBusy, [id]: true };
		try {
			const res = await fetch(clientApiPath`/api/journal/sessions/${id}`, { method: 'DELETE' });
			await readClientJson<unknown>(res);
			toast.success('Sesi dihapus');
			await refreshOverviewAfterMutation();
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Terjadi kesalahan jaringan'));
		} finally {
			deleteBusy = { ...deleteBusy, [id]: false };
		}
	}

	function formatTanggal(raw: string) {
		if (!raw) return '-';
		const d = new Date(raw);
		return d.toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' });
	}
</script>

<div class="container mx-auto max-w-6xl space-y-6 p-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-foreground">Jurnal Kelas</h1>
			<p class="text-sm text-muted-foreground">Catatan pertemuan dan kehadiran siswa per mata pelajaran</p>
		</div>
		<Button onclick={() => (createOpen = true)} disabled={!assignmentId}>
			+ Tambah Pertemuan
		</Button>
	</div>

	<!-- Assignment selector -->
	<Card.Root>
		<Card.Content class="pt-4">
			<div class="flex items-center gap-4">
				<label for="assignment-select" class="w-32 shrink-0 text-sm font-medium text-foreground">Kelas – Mapel</label>
				{#if !overviewPromise && assignments.length === 0}
					<Skeleton class="h-9 w-72" />
				{:else}
					<select
						id="assignment-select"
						value={assignmentId}
						onchange={(event) => {
							assignmentId = (event.currentTarget as HTMLSelectElement).value;
							loadOverview(assignmentId);
						}}
						class="w-72 rounded-md border border-border bg-card px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
					>
						<option value="">-- Pilih kelas & mata pelajaran --</option>
						{#each assignments as a (a.id)}
							<option value={a.id}>{a.class_name} – {a.subject_name} ({a.teacher_name})</option>
						{/each}
					</select>
				{/if}
			</div>
		</Card.Content>
	</Card.Root>

	<AsyncContent promise={overviewPromise} onerror={handleOverviewRenderError}>
		{#snippet pending()}
			{#if assignmentId}
				<Card.Root>
					<Card.Content class="space-y-2 p-4">
						{#each [1, 2, 3] as row (row)}
							<Skeleton class="h-12 w-full" />
						{/each}
					</Card.Content>
				</Card.Root>
			{/if}
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Jurnal Belum Tersaji"
				message={overviewErrorMessage(error)}
				onRetry={() => retryOverview(reset)}
			/>
		{/snippet}

		{#snippet children(value)}
			{@const currentOverview = value as Overview}
			{#if assignmentId}
				<div class="flex gap-1 border-b border-border">
					<button
						class="px-4 py-2 text-sm font-medium transition-colors {activeTab === 'sessions'
							? 'border-b-2 border-success text-success'
							: 'text-muted-foreground hover:text-foreground'}"
						onclick={() => (activeTab = 'sessions')}
					>
						Daftar Pertemuan
					</button>
					<button
						class="px-4 py-2 text-sm font-medium transition-colors {activeTab === 'rekap'
							? 'border-b-2 border-success text-success'
							: 'text-muted-foreground hover:text-foreground'}"
						onclick={() => (activeTab = 'rekap')}
					>
						Rekap Kehadiran
					</button>
				</div>

				{#if activeTab === 'sessions'}
					<Card.Root>
						<Card.Content class="p-0">
							{#if currentOverview.sessions.length === 0}
								<div class="p-8 text-center text-sm text-muted-foreground">
									Belum ada pertemuan. Klik "Tambah Pertemuan" untuk mulai.
								</div>
							{:else}
								<Table.Root>
									<Table.Header>
										<Table.Row>
											<Table.Head class="w-12">No</Table.Head>
											<Table.Head class="w-20">Ke-</Table.Head>
											<Table.Head>Tanggal</Table.Head>
											<Table.Head>Materi</Table.Head>
											<Table.Head class="w-28">Guru</Table.Head>
											<Table.Head class="w-32">Aksi</Table.Head>
										</Table.Row>
									</Table.Header>
									<Table.Body>
										{#each currentOverview.sessions as s, i (s.id)}
											<Table.Row>
												<Table.Cell class="text-muted-foreground">{i + 1}</Table.Cell>
												<Table.Cell>
													<Badge variant="outline">P-{s.pertemuan_ke}</Badge>
												</Table.Cell>
												<Table.Cell class="text-sm">{formatTanggal(s.tanggal)}</Table.Cell>
												<Table.Cell class="max-w-xs truncate text-sm text-foreground">
													{#if s.materi}
														{s.materi}
													{:else}
														<span class="italic text-muted-foreground">–</span>
													{/if}
												</Table.Cell>
												<Table.Cell>
													{#if s.guru_hadir}
														<Badge class="bg-success/15 text-success hover:bg-success/15">Hadir</Badge>
													{:else}
														<Badge class="bg-destructive/15 text-destructive hover:bg-destructive/15">Tidak Hadir</Badge>
													{/if}
												</Table.Cell>
												<Table.Cell>
													<div class="flex gap-1">
														<Button variant="outline" size="sm" onclick={() => goto(clientApiPath`/journal/${s.id}`)}>
															Lihat
														</Button>
														<LoadingButton
															variant="destructive"
															size="sm"
															loading={deleteBusy[s.id] ?? false}
															onclick={() => deleteSession(s.id)}
														>
															Hapus
														</LoadingButton>
													</div>
												</Table.Cell>
											</Table.Row>
										{/each}
									</Table.Body>
								</Table.Root>
							{/if}
						</Card.Content>
					</Card.Root>
				{/if}

				{#if activeTab === 'rekap'}
					<Card.Root>
						<Card.Content class="p-0">
							{#if currentOverview.summary.length === 0}
								<div class="p-8 text-center text-sm text-muted-foreground">
									Belum ada data rekap kehadiran.
								</div>
							{:else}
								<Table.Root>
									<Table.Header>
										<Table.Row>
											<Table.Head class="w-12">No</Table.Head>
											<Table.Head>Nama Siswa</Table.Head>
											<Table.Head class="w-24">NIS</Table.Head>
											<Table.Head class="w-16 text-center">H</Table.Head>
											<Table.Head class="w-16 text-center">S</Table.Head>
											<Table.Head class="w-16 text-center">I</Table.Head>
											<Table.Head class="w-16 text-center">A</Table.Head>
											<Table.Head class="w-20 text-center">Total</Table.Head>
										</Table.Row>
									</Table.Header>
									<Table.Body>
										{#each currentOverview.summary as st, i (st.student_id)}
											<Table.Row>
												<Table.Cell class="text-muted-foreground">{i + 1}</Table.Cell>
												<Table.Cell class="font-medium">{st.nama}</Table.Cell>
												<Table.Cell class="text-sm text-muted-foreground">{st.nis}</Table.Cell>
												<Table.Cell class="text-center">
													<Badge class="bg-success/15 text-success hover:bg-success/15">{st.hadir}</Badge>
												</Table.Cell>
												<Table.Cell class="text-center">
													<Badge class="bg-warning/15 text-warning hover:bg-warning/15">{st.sakit}</Badge>
												</Table.Cell>
												<Table.Cell class="text-center">
													<Badge class="bg-accent text-accent-foreground hover:bg-accent">{st.izin}</Badge>
												</Table.Cell>
												<Table.Cell class="text-center">
													<Badge class="bg-destructive/15 text-destructive hover:bg-destructive/15">{st.alpha}</Badge>
												</Table.Cell>
												<Table.Cell class="text-center text-sm text-muted-foreground">{st.total_pertemuan}</Table.Cell>
											</Table.Row>
										{/each}
									</Table.Body>
								</Table.Root>
							{/if}
						</Card.Content>
					</Card.Root>
				{/if}
			{/if}
		{/snippet}
	</AsyncContent>
</div>

<!-- Create Session Dialog -->
<Dialog.Root bind:open={createOpen}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>Tambah Pertemuan</Dialog.Title>
			<Dialog.Description>Isi detail pertemuan baru untuk kelas ini.</Dialog.Description>
		</Dialog.Header>
		<div class="space-y-4 py-2">
			<div class="space-y-1">
				<label for="new-tanggal" class="text-sm font-medium text-foreground">
					Tanggal <span class="text-destructive">*</span>
				</label>
				<Input id="new-tanggal" type="date" bind:value={newTanggal} />
			</div>
			<div class="space-y-1">
				<label for="new-materi" class="text-sm font-medium text-foreground">Materi</label>
				<Textarea id="new-materi" bind:value={newMateri} placeholder="Topik atau materi yang diajarkan" rows={2} />
			</div>
			<div class="space-y-1">
				<label for="new-kegiatan" class="text-sm font-medium text-foreground">Kegiatan</label>
				<Textarea id="new-kegiatan" bind:value={newKegiatan} placeholder="Aktivitas pembelajaran" rows={2} />
			</div>
			<div class="space-y-1">
				<label for="new-catatan" class="text-sm font-medium text-foreground">Catatan</label>
				<Textarea id="new-catatan" bind:value={newCatatan} placeholder="Catatan tambahan (opsional)" rows={2} />
			</div>
			<div class="flex items-center gap-3">
				<input
					id="new-guru-hadir"
					type="checkbox"
					bind:checked={newGuruHadir}
					class="h-4 w-4 rounded border-border text-success"
				/>
				<label for="new-guru-hadir" class="text-sm font-medium text-foreground">Guru hadir mengajar</label>
			</div>
		</div>
		<Dialog.Footer>
			<Button variant="outline" onclick={() => (createOpen = false)}>Batal</Button>
			<LoadingButton loading={createBusy} onclick={() => void createSession()}>Simpan & Buka</LoadingButton>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
