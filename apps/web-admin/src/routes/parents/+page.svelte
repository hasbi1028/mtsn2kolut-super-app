<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
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
	import {
		generateParentAccounts,
		previewParentAccounts,
		type ParentAccountGenerationCandidate,
		type ParentAccountGenerationResult
	} from '$lib/client/account-generation';

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
	let parentAccountBusy = $state<'preview' | 'generate' | ''>('');
	let parentAccountPreview = $state<ParentAccountGenerationResult | null>(null);
	let parentAccountResult = $state<ParentAccountGenerationResult | null>(null);

	const userRoles = $derived(page.data.user?.roles ?? (page.data.user?.role ? [page.data.user.role] : []));
	const userPermissions = $derived(page.data.user?.permissions ?? []);
	const canManageParentAccounts = $derived(userRoles.includes('admin') || userPermissions.includes('parent_accounts.manage'));
	const parentAccountSummary = $derived(parentAccountResult ?? parentAccountPreview);

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

	function parentAccountCandidate(parent: Parent): ParentAccountGenerationCandidate | undefined {
		return parentAccountSummary?.candidates.find((candidate) => candidate.parent_id === parent.id);
	}

	function accountStatusLabel(candidate: ParentAccountGenerationCandidate | undefined) {
		if (!candidate) return 'Belum dicek';
		if (candidate.status === 'created') return 'Akun dibuat';
		if (candidate.status === 'ready') return 'Siap dibuat';
		if (candidate.status === 'skipped') return 'Dilewati';
		return 'Gagal';
	}

	function accountStatusClass(candidate: ParentAccountGenerationCandidate | undefined) {
		if (!candidate) return 'bg-muted text-muted-foreground border-border';
		if (candidate.status === 'created') return 'bg-primary/15 text-primary border-primary/20';
		if (candidate.status === 'ready') return 'bg-accent text-accent-foreground border-accent';
		if (candidate.status === 'failed') return 'bg-destructive/10 text-destructive border-destructive/20';
		return 'bg-muted text-muted-foreground border-border';
	}

	async function previewAccounts() {
		if (!canManageParentAccounts) return;
		parentAccountBusy = 'preview';
		try {
			parentAccountPreview = await previewParentAccounts();
			parentAccountResult = null;
			toast.success('Preview akun orang tua siap ditinjau');
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Gagal memuat preview akun orang tua.'));
		} finally {
			parentAccountBusy = '';
		}
	}

	async function generateAccounts() {
		if (!canManageParentAccounts) return;
		if (!(await confirmAction({
			title: 'Generate Akun Orang Tua',
			message: 'Password hanya tampil sekali. Simpan/unduh hasil generate sekarang.',
			confirmLabel: 'Generate Akun',
			tone: 'warning'
		}))) return;
		parentAccountBusy = 'generate';
		try {
			parentAccountResult = await generateParentAccounts();
			parentAccountPreview = parentAccountResult;
			toast.success('Generate akun orang tua selesai');
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Gagal generate akun orang tua.'));
		} finally {
			parentAccountBusy = '';
		}
	}

	onMount(() => {
		void load();
	});
</script>

<svelte:head><title>Manajemen Orang Tua — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6 p-6">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-foreground">Manajemen Orang Tua</h1>
			<p class="text-sm text-muted-foreground mt-1">Kelola data wali murid dan relasi dengan siswa</p>
		</div>
		<Button onclick={() => (showForm = !showForm)}>
			{showForm ? 'Batal' : '+ Tambah Orang Tua'}
		</Button>
	</div>

	{#if success}
		<SuccessPanel title="Relasi Berhasil Diperbarui" message={success} />
	{/if}

	<Card.Root class="overflow-hidden border-border shadow-sm">
		<Card.Header class="pb-3">
			<div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
				<div>
					<Card.Title class="text-base">Akun Orang Tua/Wali</Card.Title>
					<p class="mt-1 text-sm text-muted-foreground">Password hanya tampil sekali. Simpan/unduh hasil generate sekarang.</p>
					<p class="mt-1 text-sm text-muted-foreground">Akun wajib mengganti password saat login pertama.</p>
					<p class="mt-1 text-sm text-muted-foreground">Data resmi tetap dikunci dan perubahan melalui approval.</p>
				</div>
				<div class="flex flex-wrap gap-2">
					<LoadingButton
						variant="outline"
						loading={parentAccountBusy === 'preview'}
						loadingLabel="Memuat..."
						disabled={!canManageParentAccounts || parentAccountBusy !== ''}
						onclick={() => void previewAccounts()}
						label="Preview akun orang tua"
					/>
					<LoadingButton
						loading={parentAccountBusy === 'generate'}
						loadingLabel="Generate..."
						disabled={!canManageParentAccounts || parentAccountBusy !== ''}
						onclick={() => void generateAccounts()}
						label="Generate akun orang tua"
					/>
				</div>
			</div>
			{#if !canManageParentAccounts}
				<p class="mt-3 rounded-md border border-border bg-muted/60 px-3 py-2 text-sm text-muted-foreground">
					Aksi akun orang tua memerlukan permission parent_accounts.manage.
				</p>
			{/if}
		</Card.Header>
		{#if parentAccountSummary}
			<Card.Content class="border-t border-border p-0">
				<div class="grid gap-3 border-b border-border p-4 sm:grid-cols-4">
					<div>
						<p class="text-xs text-muted-foreground">Total</p>
						<p class="text-xl font-semibold">{parentAccountSummary.total}</p>
					</div>
					<div>
						<p class="text-xs text-muted-foreground">Siap</p>
						<p class="text-xl font-semibold">{parentAccountSummary.ready}</p>
					</div>
					<div>
						<p class="text-xs text-muted-foreground">Dibuat</p>
						<p class="text-xl font-semibold">{parentAccountSummary.created}</p>
					</div>
					<div>
						<p class="text-xs text-muted-foreground">Dilewati/Gagal</p>
						<p class="text-xl font-semibold">{parentAccountSummary.skipped + parentAccountSummary.failed}</p>
					</div>
				</div>
				<div class="max-h-80 overflow-auto">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>Nama Orang Tua</Table.Head>
								<Table.Head>Username</Table.Head>
								<Table.Head>Password awal</Table.Head>
								<Table.Head>Anak terhubung</Table.Head>
								<Table.Head>Status akun</Table.Head>
								<Table.Head>Keterangan</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each parentAccountSummary.candidates as candidate (candidate.parent_id)}
								<Table.Row>
									<Table.Cell class="font-medium">{candidate.nama}</Table.Cell>
									<Table.Cell class="font-mono text-sm">{candidate.generated_username || '—'}</Table.Cell>
									<Table.Cell class="font-mono text-sm">{candidate.temporary_password || '—'}</Table.Cell>
									<Table.Cell>{candidate.child_count}</Table.Cell>
									<Table.Cell><Badge class={accountStatusClass(candidate)}>{accountStatusLabel(candidate)}</Badge></Table.Cell>
									<Table.Cell class="text-sm text-muted-foreground">{candidate.reason || '—'}</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</div>
			</Card.Content>
		{/if}
	</Card.Root>

	{#if showForm}
		<Card.Root>
			<Card.Header class="pb-2"><Card.Title class="text-base">Data Orang Tua Baru</Card.Title></Card.Header>
			<Card.Content class="space-y-4">
				<div class="grid gap-3 sm:grid-cols-3">
					<div>
						<label for="p-nama" class="text-xs text-muted-foreground mb-1 block">Nama Lengkap</label>
						<Input id="p-nama" bind:value={fNama} placeholder="Nama Orang Tua" />
					</div>
					<div>
						<label for="p-phone" class="text-xs text-muted-foreground mb-1 block">No. HP / WhatsApp</label>
						<Input id="p-phone" bind:value={fPhone} placeholder="08xxx" />
					</div>
					<div>
						<label for="p-addr" class="text-xs text-muted-foreground mb-1 block">Alamat</label>
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
					<div class="rounded-2xl border border-border bg-muted/50 px-4 py-4">
						<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-muted-foreground">{label}</p>
						<Skeleton class="mt-3 h-8 w-16" />
						<Skeleton class="mt-2 h-4 w-44" />
					</div>
				{/each}
			</div>

			<Card.Root class="overflow-hidden border-border shadow-sm">
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
				<div class="rounded-2xl border border-primary/20 bg-primary/10 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-primary">Data Orang Tua</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{overview.parents.length}</p>
					<p class="text-sm text-muted-foreground">profil wali murid yang sudah tercatat</p>
				</div>
				<div class="rounded-2xl border border-accent bg-accent/60 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-accent-foreground">Siswa Tersedia</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{overview.students.length}</p>
					<p class="text-sm text-muted-foreground">daftar siswa yang bisa ditautkan ke akun orang tua</p>
				</div>
				<div class="rounded-2xl border border-warning/30 bg-warning/10 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-warning">Relasi Keluarga</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{linkedStudents.length}</p>
					<p class="text-sm text-muted-foreground">anak yang sedang tampil pada panel relasi aktif</p>
				</div>
			</div>

			<Card.Root class="overflow-hidden border-border shadow-sm">
				<Card.Content class="p-0">
					<Table.Root>
						<Table.Header>
							<Table.Row class="bg-muted/50">
								<Table.Head>Nama Orang Tua</Table.Head>
								<Table.Head>No. HP</Table.Head>
								<Table.Head>Alamat</Table.Head>
								<Table.Head>Status akun</Table.Head>
								<Table.Head>Username</Table.Head>
								<Table.Head>Anak terhubung</Table.Head>
								<Table.Head class="text-right">Aksi</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each overview.parents as p (p.id)}
								{@const accountCandidate = parentAccountCandidate(p)}
								<Table.Row>
									<Table.Cell class="font-medium">{p.nama}</Table.Cell>
									<Table.Cell class="text-sm">{p.phone || '—'}</Table.Cell>
									<Table.Cell class="text-sm text-muted-foreground">{p.address || '—'}</Table.Cell>
									<Table.Cell><Badge class={accountStatusClass(accountCandidate)}>{accountStatusLabel(accountCandidate)}</Badge></Table.Cell>
									<Table.Cell class="font-mono text-sm">{accountCandidate?.generated_username || '—'}</Table.Cell>
									<Table.Cell>{accountCandidate?.child_count ?? '—'}</Table.Cell>
									<Table.Cell class="text-right">
										<Button variant="outline" size="sm" onclick={() => openLinkDialog(p)}>
											Lihat Anak
										</Button>
									</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={7} class="p-4">
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
					<div class="rounded-md border border-border p-4">
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
					<div class="rounded-md border border-border">
						<Table.Root>
							<Table.Header>
								<Table.Row class="bg-muted/50">
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
												class="text-destructive hover:bg-destructive/10 hover:text-destructive"
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
