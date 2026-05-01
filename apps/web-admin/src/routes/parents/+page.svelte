<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { toast } from '$lib/components/ui/sonner';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import SuccessPanel from '$lib/components/SuccessPanel.svelte';
	import { confirmAction } from '$lib/confirm-dialog';
	import { readClientApiData, readClientJson } from '$lib/client/api';

	type Parent = {
		id: string;
		nama: string;
		phone: string;
		address: string;
		created_at: string;
	};

	type Student = {
		id: string;
		nama: string;
		nis: string;
		class_name: string;
	};

	type ParentsOverview = {
		parents: Parent[];
		students: Student[];
	};

	let parents = $state<Parent[]>([]);
	let students = $state<Student[]>([]);
	let parentsPromise = $state<Promise<ParentsOverview> | null>(null);
	let parentsRequestId = 0;
	let success = $state('');
	let showForm = $state(false);
	let showLinkDialog = $state(false);
	let selectedParent = $state<Parent | null>(null);
	let linkedStudents = $state<Student[]>([]);
	let linkedStudentsPromise = $state<Promise<Student[]> | null>(null);
	let linkedStudentsRequestId = 0;

	let fNama = $state('');
	let fPhone = $state('');
	let fAddress = $state('');
	let fSelectedStudentId = $state('');
	let fBusy = $state(false);
	let linkBusy = $state(false);
	let unlinkBusyId = $state<string | null>(null);

	async function ensureMutationOk(response: Response, fallbackMessage: string) {
		try {
			await readClientJson<unknown>(response);
		} catch (error) {
			throw new Error(mutationErrorMessage(error, fallbackMessage));
		}
	}

	function mutationErrorMessage(error: unknown, fallbackMessage: string) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return fallbackMessage;
	}

	function applyOverview(overview: ParentsOverview) {
		parents = overview.parents;
		students = overview.students;
		return overview;
	}

	async function fetchParentsOverview(): Promise<ParentsOverview> {
		const [parentsRes, studentsRes] = await Promise.all([
			fetch('/api/parents'),
			fetch('/api/students')
		]);
		const [nextParents, nextStudents] = await Promise.all([
			readClientApiData<Parent[]>(parentsRes, 'Gagal memuat data orang tua.'),
			readClientApiData<Student[]>(studentsRes, 'Gagal memuat daftar siswa.')
		]);
		return {
			parents: nextParents ?? [],
			students: nextStudents ?? []
		};
	}

	function load() {
		const requestId = ++parentsRequestId;
		parents = [];
		students = [];
		parentsPromise = fetchParentsOverview()
			.then((overview) => {
				if (requestId === parentsRequestId) return applyOverview(overview);
				return { parents, students };
			})
			.catch((error: unknown) => {
				if (requestId === parentsRequestId) throw error;
				return { parents, students };
			});
	}

	async function refreshOverview() {
		const requestId = ++parentsRequestId;
		const overview = await fetchParentsOverview();
		if (requestId === parentsRequestId) {
			applyOverview(overview);
			parentsPromise = Promise.resolve(overview);
		}
	}

	function retryOverview(reset?: () => void) {
		reset?.();
		load();
	}

	async function fetchLinkedStudents(parentId: string) {
		const res = await fetch(`/api/parents/${parentId}/children`);
		return readClientApiData<Student[]>(res, 'Gagal memuat daftar anak.');
	}

	function loadLinkedStudents(parentId: string) {
		const requestId = ++linkedStudentsRequestId;
		linkedStudents = [];
		linkedStudentsPromise = fetchLinkedStudents(parentId)
			.then((nextStudents) => {
				if (requestId === linkedStudentsRequestId && selectedParent?.id === parentId) {
					linkedStudents = nextStudents ?? [];
				}
				return linkedStudents;
			})
			.catch((error: unknown) => {
				if (requestId === linkedStudentsRequestId && selectedParent?.id === parentId) throw error;
				return linkedStudents;
			});
	}

	async function refreshLinkedStudents() {
		if (!selectedParent) return;
		const requestId = ++linkedStudentsRequestId;
		const parentId = selectedParent.id;
		const nextStudents = await fetchLinkedStudents(parentId);
		if (requestId === linkedStudentsRequestId && selectedParent?.id === parentId) {
			linkedStudents = nextStudents ?? [];
			linkedStudentsPromise = Promise.resolve(linkedStudents);
		}
	}

	function retryLinkedStudents(reset?: () => void) {
		if (!selectedParent) return;
		reset?.();
		loadLinkedStudents(selectedParent.id);
	}

	function overviewErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat data orang tua dan daftar siswa. Coba lagi untuk melanjutkan pengelolaan relasi keluarga.';
	}

	function linkedStudentsErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat daftar anak yang tertaut.';
	}

	function handleOverviewRenderError(error: unknown) {
		console.error('Parents overview render failed', error);
	}

	function handleLinkedStudentsRenderError(error: unknown) {
		console.error('Linked students render failed', error);
	}

	async function saveParent() {
		if (!fNama) return;
		fBusy = true;
		try {
			success = '';
			const res = await fetch('/api/parents', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ nama: fNama, phone: fPhone, address: fAddress })
			});
			await ensureMutationOk(res, 'Gagal menyimpan data orang tua.');
			fNama = ''; fPhone = ''; fAddress = '';
			showForm = false;
			toast.success('Data orang tua berhasil disimpan');
			success = 'Profil orang tua berhasil ditambahkan. Selanjutnya kamu bisa membuka relasi anak untuk mulai menautkan siswa.';
			await refreshOverview();
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Gagal menyimpan data'));
		} finally { fBusy = false; }
	}

	function openLinkDialog(parent: Parent) {
		selectedParent = parent;
		fSelectedStudentId = '';
		showLinkDialog = true;
		loadLinkedStudents(parent.id);
	}

	async function linkStudent() {
		if (!selectedParent || !fSelectedStudentId) return;
		linkBusy = true;
		try {
			success = '';
			const res = await fetch(`/api/parents/${selectedParent.id}/link`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ student_id: fSelectedStudentId })
			});
			await ensureMutationOk(res, 'Gagal menautkan siswa.');
			toast.success('Siswa berhasil ditautkan');
			fSelectedStudentId = '';
			success = `Relasi keluarga berhasil diperbarui. ${selectedParent.nama} sekarang terhubung dengan siswa yang dipilih.`;
			await refreshLinkedStudents();
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Gagal menautkan siswa'));
		} finally { linkBusy = false; }
	}

	async function unlinkStudent(studentId: string, studentName: string) {
		if (!selectedParent) return;
		if (!(await confirmAction({
			title: 'Lepas Tautan Orang Tua',
			message: `Lepas tautan ${studentName} dari ${selectedParent.nama}?`,
			confirmLabel: 'Lepas Tautan',
			tone: 'warning'
		}))) return;
		unlinkBusyId = studentId;
		try {
			success = '';
			const res = await fetch(`/api/parents/${selectedParent.id}/unlink`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ student_id: studentId })
			});
			await ensureMutationOk(res, 'Gagal melepas tautan siswa.');
			toast.success('Tautan siswa berhasil dilepas');
			success = `Tautan ${studentName} berhasil dilepas dari ${selectedParent.nama}.`;
			await refreshLinkedStudents();
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Gagal melepas tautan siswa'));
		} finally { unlinkBusyId = null; }
	}

	onMount(() => {
		void load();
	});
</script>

<svelte:head><title>Manajemen Orang Tua — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6 p-6">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-slate-800">Manajemen Orang Tua</h1>
			<p class="text-sm text-slate-500 mt-1">Kelola data wali murid dan relasi dengan siswa</p>
		</div>
		<Button onclick={() => (showForm = !showForm)}>
			{showForm ? 'Batal' : '+ Tambah Orang Tua'}
		</Button>
	</div>

	{#if success}
		<SuccessPanel title="Relasi Berhasil Diperbarui" message={success} />
	{/if}

	{#if showForm}
		<Card.Root>
			<Card.Header class="pb-2"><Card.Title class="text-base">Data Orang Tua Baru</Card.Title></Card.Header>
			<Card.Content class="space-y-4">
				<div class="grid gap-3 sm:grid-cols-3">
					<div>
						<label for="p-nama" class="text-xs text-slate-500 mb-1 block">Nama Lengkap</label>
						<Input id="p-nama" bind:value={fNama} placeholder="Nama Orang Tua" />
					</div>
					<div>
						<label for="p-phone" class="text-xs text-slate-500 mb-1 block">No. HP / WhatsApp</label>
						<Input id="p-phone" bind:value={fPhone} placeholder="08xxx" />
					</div>
					<div>
						<label for="p-addr" class="text-xs text-slate-500 mb-1 block">Alamat</label>
						<Input id="p-addr" bind:value={fAddress} placeholder="Alamat lengkap" />
					</div>
				</div>
				<div class="flex gap-2">
					<LoadingButton onclick={() => void saveParent()} loading={fBusy} disabled={fBusy || !fNama}>
						Simpan Data
					</LoadingButton>
					<Button variant="outline" onclick={() => (showForm = false)}>Batal</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	<AsyncContent promise={parentsPromise} onerror={handleOverviewRenderError}>
		{#snippet pending()}
			<div class="grid gap-3 md:grid-cols-3">
				{#each ['Data Orang Tua', 'Siswa Tersedia', 'Relasi Keluarga'] as label (label)}
					<div class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-4">
						<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-slate-500">{label}</p>
						<Skeleton class="mt-3 h-8 w-16" />
						<Skeleton class="mt-2 h-4 w-44" />
					</div>
				{/each}
			</div>

			<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
				<Card.Content class="p-0">
					<div class="space-y-3 p-6">
						{#each Array.from({ length: 5 }) as _, index (`parent-skeleton-${index}`)}
							<div class="grid gap-3 md:grid-cols-[1.4fr_1fr_1.4fr_auto] md:items-center">
								<Skeleton class="h-5 w-40" />
								<Skeleton class="h-5 w-28" />
								<Skeleton class="h-5 w-full max-w-xs" />
								<Skeleton class="h-9 w-28 justify-self-end" />
							</div>
						{/each}
					</div>
				</Card.Content>
			</Card.Root>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Relasi Orang Tua Belum Tersaji"
				message={overviewErrorMessage(error)}
				onRetry={() => retryOverview(reset)}
			/>
		{/snippet}

		{#snippet children(value)}
			{@const overview = value as ParentsOverview}
			<div class="grid gap-3 md:grid-cols-3">
				<div class="rounded-2xl border border-emerald-100 bg-emerald-50 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-emerald-700">Data Orang Tua</p>
					<p class="mt-2 text-2xl font-semibold text-slate-900">{overview.parents.length}</p>
					<p class="text-sm text-slate-600">profil wali murid yang sudah tercatat</p>
				</div>
				<div class="rounded-2xl border border-sky-100 bg-sky-50 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-sky-700">Siswa Tersedia</p>
					<p class="mt-2 text-2xl font-semibold text-slate-900">{overview.students.length}</p>
					<p class="text-sm text-slate-600">daftar siswa yang bisa ditautkan ke akun orang tua</p>
				</div>
				<div class="rounded-2xl border border-amber-100 bg-amber-50 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-amber-700">Relasi Keluarga</p>
					<p class="mt-2 text-2xl font-semibold text-slate-900">{linkedStudents.length}</p>
					<p class="text-sm text-slate-600">anak yang sedang tampil pada panel relasi aktif</p>
				</div>
			</div>

			<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
				<Card.Content class="p-0">
					<Table.Root>
						<Table.Header>
							<Table.Row class="bg-slate-50">
								<Table.Head>Nama Orang Tua</Table.Head>
								<Table.Head>No. HP</Table.Head>
								<Table.Head>Alamat</Table.Head>
								<Table.Head class="text-right">Aksi</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each overview.parents as p (p.id)}
								<Table.Row>
									<Table.Cell class="font-medium">{p.nama}</Table.Cell>
									<Table.Cell class="text-sm">{p.phone || '—'}</Table.Cell>
									<Table.Cell class="text-sm text-slate-600">{p.address || '—'}</Table.Cell>
									<Table.Cell class="text-right">
										<Button variant="outline" size="sm" onclick={() => openLinkDialog(p)}>
											Lihat Anak
										</Button>
									</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={4} class="p-4">
										<EmptyStatePanel
											compact
											title="Belum ada data orang tua"
											description="Tambahkan wali murid pertama agar relasi keluarga, portal orang tua, dan komunikasi sekolah bisa mulai dibangun."
										/>
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			</Card.Root>
		{/snippet}
	</AsyncContent>
</div>

<!-- Link Student Dialog -->
<Dialog.Root bind:open={showLinkDialog}>
	<Dialog.Content>
		<div class="sm:max-w-[500px]">
		<Dialog.Header>
			<Dialog.Title>Anak dari {selectedParent?.nama}</Dialog.Title>
			<Dialog.Description>Tautkan data siswa dengan orang tua ini.</Dialog.Description>
		</Dialog.Header>
		<div class="space-y-4 py-4">
			<div class="flex gap-2">
				<select class="flex-1 rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fSelectedStudentId}>
					<option value="">-- Pilih Siswa untuk Ditautkan --</option>
					{#each students as s (s.id)}
						<option value={s.id}>{s.nama} ({s.nis})</option>
					{/each}
				</select>
				<LoadingButton size="sm" onclick={() => void linkStudent()} loading={linkBusy} loadingLabel="Menautkan..." disabled={linkBusy || !fSelectedStudentId}>
					Tautkan
				</LoadingButton>
			</div>

			<AsyncContent promise={linkedStudentsPromise} onerror={handleLinkedStudentsRenderError}>
				{#snippet pending()}
					<div class="rounded-md border border-slate-200 p-4">
						<div class="space-y-3">
							{#each Array.from({ length: 3 }) as _, index (`linked-student-skeleton-${index}`)}
								<div class="grid grid-cols-[1fr_96px_56px] items-center gap-3">
									<Skeleton class="h-5 w-36" />
									<Skeleton class="h-5 w-20" />
									<Skeleton class="h-8 w-14 justify-self-end" />
								</div>
							{/each}
						</div>
					</div>
				{/snippet}

				{#snippet failed(error, reset)}
					<RecoveryPanel
						compact
						title="Daftar Anak Belum Tersaji"
						message={linkedStudentsErrorMessage(error)}
						onRetry={() => retryLinkedStudents(reset)}
					/>
				{/snippet}

				{#snippet children(value)}
					{@const currentLinkedStudents = value as Student[]}
					<div class="rounded-md border border-slate-200">
						<Table.Root>
							<Table.Header>
								<Table.Row class="bg-slate-50">
									<Table.Head class="h-9">Nama Siswa</Table.Head>
									<Table.Head class="h-9">Kelas</Table.Head>
									<Table.Head class="h-9 text-right">Aksi</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each currentLinkedStudents as c (c.id)}
									<Table.Row>
										<Table.Cell class="py-2">{c.nama}</Table.Cell>
										<Table.Cell class="py-2 text-xs">{c.class_name || '—'}</Table.Cell>
										<Table.Cell class="py-2 text-right">
											<LoadingButton
												size="sm"
												variant="ghost"
												class="text-red-600 hover:bg-red-50 hover:text-red-700"
												onclick={() => unlinkStudent(c.id, c.nama)}
												loading={unlinkBusyId === c.id}
												loadingLabel="Melepas..."
												disabled={unlinkBusyId !== null && unlinkBusyId !== c.id}
											>
												Lepas
											</LoadingButton>
										</Table.Cell>
									</Table.Row>
								{:else}
									<Table.Row>
										<Table.Cell colspan={3} class="p-4">
											<EmptyStatePanel
												compact
												title="Belum ada siswa yang ditautkan"
												description="Pilih siswa dari dropdown di atas untuk mulai membangun relasi orang tua dan anak."
											/>
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</div>
				{/snippet}
			</AsyncContent>
		</div>
		</div>
	</Dialog.Content>
</Dialog.Root>
