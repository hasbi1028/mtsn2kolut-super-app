<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import EntityDrawer from '$lib/components/EntityDrawer.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { confirmAction } from '$lib/confirm-dialog';
	import { clientApiPathWithQuery, readClientApiData, readClientJson } from '$lib/client/api';
	import {
		generateStudentAccounts,
		previewStudentAccounts,
		type StudentAccountGenerationCandidate,
		type StudentAccountGenerationResult
	} from '$lib/client/account-generation';

	type Student = {
		id: string; nis: string; nisn: string; nama: string; gender: string;
		parent_name: string; parent_phone: string;
		class_id: string; class_name: string; class_code: string;
		linked_parent_names: string; linked_parent_count: number;
		is_active: boolean; status: string; created_at: string;
	};
	type SchoolClass = { id: string; name: string; code: string; level: string; };

	type StudentsOverview = {
		students: Student[];
		classes: SchoolClass[];
	};

	type StudentFormStep = 'identity' | 'class' | 'guardian' | 'status';

	let students = $state<Student[]>([]);
	let classes = $state<SchoolClass[]>([]);
	let studentsPromise = $state<Promise<StudentsOverview> | null>(null);
	let studentsRequestId = 0;
	let search = $state('');
	let selectedStudentIds = $state<Set<string>>(new Set());
	let visibleSelectionCheckbox = $state<HTMLInputElement | null>(null);
	let wasStudentDrawerOpen = false;

	let formNis = $state('');
	let formNisn = $state('');
	let formNama = $state('');
	let formGender = $state('L');
	let formParentName = $state('');
	let formParentPhone = $state('');
	let formClassId = $state('');
	let formActive = $state(true);
	let formStatus = $state('active');
	let formBusy = $state(false);
	let showForm = $state(false);
	let editId = $state<string | null>(null);
	let studentFormStep = $state<StudentFormStep>('identity');
	let deleteBusyId = $state('');
	let lifecycleBusyKey = $state('');
	let studentAccountBusy = $state<'preview' | 'generate' | ''>('');
	let studentAccountPreview = $state<StudentAccountGenerationResult | null>(null);
	let studentAccountResult = $state<StudentAccountGenerationResult | null>(null);

	const studentFormSteps: Array<{ id: StudentFormStep; label: string; description: string }> = [
		{ id: 'identity', label: 'Identitas', description: 'NIS, NISN, nama, dan gender' },
		{ id: 'class', label: 'Kelas', description: 'Kelas aktif dan lifecycle' },
		{ id: 'guardian', label: 'Wali', description: 'Kontak orang tua/wali' },
		{ id: 'status', label: 'Status', description: 'Aktif/nonaktif dan review akhir' }
	];

	const studentFormStepIndex = $derived(studentFormSteps.findIndex((step) => step.id === studentFormStep));
	const userRoles = $derived(page.data.user?.roles ?? (page.data.user?.role ? [page.data.user.role] : []));
	const userPermissions = $derived(page.data.user?.permissions ?? []);
	const canManageStudentAccounts = $derived(userRoles.includes('admin') || userPermissions.includes('student_accounts.manage'));
	const studentAccountSummary = $derived(studentAccountResult ?? studentAccountPreview);
	const editingStudent = $derived(editId ? (students.find((student) => student.id === editId) ?? null) : null);
	const studentFormDirty = $derived(
		editId
			? Boolean(editingStudent) && (
				formNis !== (editingStudent?.nis ?? '') ||
				formNisn !== (editingStudent?.nisn ?? '') ||
				formNama !== (editingStudent?.nama ?? '') ||
				formGender !== (editingStudent?.gender ?? 'L') ||
				formParentName !== (editingStudent?.parent_name ?? '') ||
				formParentPhone !== (editingStudent?.parent_phone ?? '') ||
				formClassId !== (editingStudent?.class_id ?? '') ||
				formActive !== (editingStudent?.is_active ?? true) ||
				formStatus !== (editingStudent?.status ?? 'active')
			)
			: Boolean(formNis || formNisn || formNama || formParentName || formParentPhone || formClassId || formGender !== 'L' || !formActive || formStatus !== 'active')
	);

	let filtered = $derived(
		search.trim()
			? students.filter(s =>
				s.nama.toLowerCase().includes(search.toLowerCase()) ||
				s.nis.includes(search) ||
				(s.nisn ?? '').includes(search)
			)
			: students
	);
	const filteredStudentIds = $derived(filtered.map((student) => student.id));
	const selectedCount = $derived(selectedStudentIds.size);
	const filteredSelectedCount = $derived(filteredStudentIds.filter((id) => selectedStudentIds.has(id)).length);
	const allFilteredStudentsSelected = $derived(filteredStudentIds.length > 0 && filteredSelectedCount === filteredStudentIds.length);

	function isRecord(value: unknown): value is Record<string, unknown> {
		return typeof value === 'object' && value !== null;
	}

	function extractClasses(payload: unknown) {
		if (!isRecord(payload)) return [];
		const maybeClasses = payload.classes;
		return Array.isArray(maybeClasses) ? (maybeClasses as SchoolClass[]) : [];
	}

	function applyOverview(overview: StudentsOverview) {
		students = overview.students;
		classes = overview.classes;
		return overview;
	}

	async function fetchOverview(): Promise<StudentsOverview> {
		const [studentsRes, academicRes] = await Promise.all([
			fetch('/api/students'),
			fetch('/api/academic'),
		]);
		const [nextStudents, academic] = await Promise.all([
			readClientApiData<Student[]>(studentsRes, 'Gagal memuat data siswa.'),
			readClientApiData<unknown>(academicRes, 'Gagal memuat data akademik.'),
		]);
		return {
			students: nextStudents ?? [],
			classes: extractClasses(academic),
		};
	}

	function load() {
		const requestId = ++studentsRequestId;
		students = [];
		classes = [];
		studentsPromise = fetchOverview()
			.then((overview) => {
				if (requestId === studentsRequestId) return applyOverview(overview);
				return { students, classes };
			})
			.catch((error: unknown) => {
				if (requestId === studentsRequestId) throw error;
				return { students, classes };
			});
	}

	async function refreshOverview() {
		const requestId = ++studentsRequestId;
		const overview = await fetchOverview();
		if (requestId === studentsRequestId) {
			applyOverview(overview);
			studentsPromise = Promise.resolve(overview);
		}
	}

	function retryOverview(reset?: () => void) {
		reset?.();
		load();
	}

	function overviewErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat data siswa';
	}

	function handleOverviewRenderError(error: unknown) {
		console.error('Students overview render failed', error);
	}

	function showToast(msg: string) {
		toast.success(msg);
	}

	function showError(msg: string) {
		toast.error(msg);
	}

	function mutationErrorMessage(error: unknown, fallback: string) {
		if (error instanceof Error && error.message.trim() && !error.message.toLowerCase().includes('fetch')) return error.message;
		return fallback;
	}

	async function refreshOverviewAfterMutation() {
		try {
			await refreshOverview();
		} catch (error) {
			studentsPromise = Promise.resolve({ students, classes });
			showError(overviewErrorMessage(error));
		}
	}

	function clearFormFields() {
		formNis = ''; formNisn = ''; formNama = ''; formGender = 'L';
		formParentName = ''; formParentPhone = ''; formClassId = ''; formActive = true; formStatus = 'active';
		editId = null;
		studentFormStep = 'identity';
	}

	function resetForm() {
		clearFormFields();
		showForm = false;
	}

	function openCreate() {
		clearFormFields();
		showForm = true;
	}

	function openEdit(s: Student) {
		formNis = s.nis;
		formNisn = s.nisn;
		formNama = s.nama;
		formGender = s.gender;
		formParentName = s.parent_name;
		formParentPhone = s.parent_phone;
		formClassId = s.class_id || '';
		formActive = s.is_active;
		formStatus = s.status || 'active';
		editId = s.id;
		studentFormStep = 'identity';
		showForm = true;
	}

	function clearSelection() {
		selectedStudentIds = new Set();
	}

	function updateStudentSelection(studentId: string, selected: boolean) {
		const next = new Set(selectedStudentIds);
		if (selected) next.add(studentId);
		else next.delete(studentId);
		selectedStudentIds = next;
	}

	function handleStudentSelectionChange(event: Event, studentId: string) {
		const input = event.currentTarget as HTMLInputElement | null;
		updateStudentSelection(studentId, Boolean(input?.checked));
	}

	function handleFilteredSelectionChange(event: Event) {
		const input = event.currentTarget as HTMLInputElement | null;
		const next = new Set(selectedStudentIds);
		for (const id of filteredStudentIds) {
			if (input?.checked) next.add(id);
			else next.delete(id);
		}
		selectedStudentIds = next;
	}

	function nextStudentFormStep() {
		const next = studentFormSteps[studentFormStepIndex + 1];
		if (next) studentFormStep = next.id;
	}

	function previousStudentFormStep() {
		const previous = studentFormSteps[studentFormStepIndex - 1];
		if (previous) studentFormStep = previous.id;
	}

	function lifecycleKey(studentId: string, status: string) {
		return `${studentId}:${status}`;
	}

	function studentMutationPath(id: string) {
		return clientApiPathWithQuery('/api/students', new URLSearchParams({ id }));
	}

	async function saveStudent() {
		if (!formNis || !formNama || !formGender) return;
		formBusy = true;
		try {
			const method = editId ? 'PUT' : 'POST';
			const path = editId ? studentMutationPath(editId) : '/api/students';
			const res = await fetch(path, {
				method,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					nis: formNis, nisn: formNisn, nama: formNama, gender: formGender,
					parent_name: formParentName, parent_phone: formParentPhone,
					class_id: formClassId, is_active: formActive, status: formStatus,
				}),
			});
			await readClientJson<unknown>(res);
			showToast(editId ? 'Data siswa diperbarui' : 'Siswa berhasil ditambahkan');
			resetForm();
			await refreshOverviewAfterMutation();
		} catch (error) {
			showError(mutationErrorMessage(error, 'Gagal menyimpan data siswa. Periksa koneksi lalu coba lagi.'));
		} finally { formBusy = false; }
	}

	async function deleteStudent(id: string, nama: string) {
		if (!(await confirmAction({
			title: 'Hapus Data Siswa',
			message: `Hapus siswa "${nama}"? Data terkait siswa ini dapat memengaruhi kelas, nilai, dan CBT.`,
			confirmLabel: 'Hapus Siswa',
			tone: 'danger'
		}))) return;
		deleteBusyId = id;
		try {
			const res = await fetch(studentMutationPath(id), { method: 'DELETE' });
			await readClientJson<unknown>(res);
			showToast('Siswa dihapus');
			await refreshOverviewAfterMutation();
		} catch (error) {
			showError(mutationErrorMessage(error, 'Gagal menghapus siswa. Periksa koneksi lalu coba lagi.'));
		} finally {
			deleteBusyId = '';
		}
	}

	async function updateLifecycle(student: Student, status: string) {
		const labels: Record<string, string> = {
			prospective: 'Calon Siswa',
			active: 'Aktif',
			alumni: 'Alumni',
			mutated: 'Mutasi',
		};
		if (!(await confirmAction({
			title: 'Ubah Lifecycle Siswa',
			message: `Ubah status ${student.nama} menjadi ${labels[status] ?? status}?`,
			confirmLabel: 'Ubah Status',
			tone: 'warning'
		}))) return;
		const busyKey = lifecycleKey(student.id, status);
		lifecycleBusyKey = busyKey;
		try {
			const res = await fetch(studentMutationPath(student.id), {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ status }),
			});
			await readClientJson<unknown>(res);
			showToast('Lifecycle siswa diperbarui');
			await refreshOverviewAfterMutation();
		} catch (error) {
			showError(mutationErrorMessage(error, 'Gagal memperbarui lifecycle siswa. Periksa koneksi lalu coba lagi.'));
		} finally {
			lifecycleBusyKey = '';
		}
	}

	function lifecycleBadgeClass(status: string) {
		if (status === 'prospective') return 'bg-accent text-accent-foreground border-accent';
		if (status === 'alumni') return 'bg-accent text-accent-foreground border-accent';
		if (status === 'mutated') return 'bg-warning/15 text-warning border-warning/30';
		return 'bg-primary/15 text-primary border-primary/20';
	}

	function parentSummary(student: Student) {
		if (student.linked_parent_count > 0) {
			return student.linked_parent_count > 1
				? `${student.linked_parent_names} (${student.linked_parent_count} relasi)`
				: student.linked_parent_names;
		}
		return student.parent_name || 'Wali belum diisi';
	}

	function studentAccountCandidate(student: Student): StudentAccountGenerationCandidate | undefined {
		return studentAccountSummary?.candidates.find((candidate) => candidate.student_id === student.id);
	}

	function accountStatusLabel(candidate: StudentAccountGenerationCandidate | undefined) {
		if (!candidate) return 'Belum dicek';
		if (candidate.status === 'created') return 'Akun dibuat';
		if (candidate.status === 'ready') return 'Siap dibuat';
		if (candidate.status === 'skipped') return 'Dilewati';
		return 'Gagal';
	}

	function accountStatusClass(candidate: StudentAccountGenerationCandidate | undefined) {
		if (!candidate) return 'bg-muted text-muted-foreground border-border';
		if (candidate.status === 'created') return 'bg-primary/15 text-primary border-primary/20';
		if (candidate.status === 'ready') return 'bg-accent text-accent-foreground border-accent';
		if (candidate.status === 'failed') return 'bg-destructive/10 text-destructive border-destructive/20';
		return 'bg-muted text-muted-foreground border-border';
	}

	async function previewAccounts() {
		if (!canManageStudentAccounts) return;
		studentAccountBusy = 'preview';
		try {
			studentAccountPreview = await previewStudentAccounts();
			studentAccountResult = null;
			showToast('Preview akun siswa siap ditinjau');
		} catch (error) {
			showError(mutationErrorMessage(error, 'Gagal memuat preview akun siswa.'));
		} finally {
			studentAccountBusy = '';
		}
	}

	async function generateAccounts() {
		if (!canManageStudentAccounts) return;
		if (!(await confirmAction({
			title: 'Generate Akun Siswa',
			message: 'Password awal akan memakai NISN siswa dan hanya tampil sekali. Simpan/unduh hasil generate sekarang.',
			confirmLabel: 'Generate Akun',
			tone: 'warning'
		}))) return;
		studentAccountBusy = 'generate';
		try {
			studentAccountResult = await generateStudentAccounts();
			studentAccountPreview = studentAccountResult;
			showToast('Generate akun siswa selesai');
		} catch (error) {
			showError(mutationErrorMessage(error, 'Gagal generate akun siswa.'));
		} finally {
			studentAccountBusy = '';
		}
	}

	$effect(() => {
		if (visibleSelectionCheckbox) {
			visibleSelectionCheckbox.indeterminate = filteredSelectedCount > 0 && !allFilteredStudentsSelected;
		}
	});

	$effect(() => {
		const activeIds = new Set(students.map((student) => student.id));
		const next = new Set([...selectedStudentIds].filter((id) => activeIds.has(id)));
		if (next.size !== selectedStudentIds.size) selectedStudentIds = next;
	});

	$effect(() => {
		if (wasStudentDrawerOpen && !showForm) clearFormFields();
		wasStudentDrawerOpen = showForm;
	});

	onMount(() => {
		void load();
	});
</script>

<svelte:head><title>Data Siswa — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-foreground">Data Siswa</h1>
			<p class="text-sm text-muted-foreground mt-1">Kelola daftar siswa aktif madrasah</p>
		</div>
		<Button onclick={() => { if (showForm) resetForm(); else openCreate(); }}>
			{showForm ? 'Tutup' : '+ Tambah Siswa'}
		</Button>
	</div>

	<AsyncContent promise={studentsPromise} onerror={handleOverviewRenderError}>
		{#snippet pending()}
			<div class="grid gap-3 md:grid-cols-3">
				{#each ['Total Siswa', 'Siswa Aktif', 'Relasi Ortu'] as label (label)}
					<div class="rounded-2xl border border-border bg-muted/50 px-4 py-4">
						<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-muted-foreground">{label}</p>
						<Skeleton class="mt-3 h-8 w-16" />
						<Skeleton class="mt-2 h-4 w-48" />
					</div>
				{/each}
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel message={overviewErrorMessage(error)} onRetry={() => retryOverview(reset)} />
		{/snippet}

		{#snippet children(value)}
			{@const overview = value as StudentsOverview}
			<div class="grid gap-3 md:grid-cols-3">
				<div class="rounded-2xl border border-primary/20 bg-primary/10 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-primary">Total Siswa</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{overview.students.length}</p>
					<p class="text-sm text-muted-foreground">seluruh entitas siswa yang sudah tersimpan</p>
				</div>
				<div class="rounded-2xl border border-accent bg-accent/60 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-accent-foreground">Siswa Aktif</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{overview.students.filter((item) => item.status === 'active').length}</p>
					<p class="text-sm text-muted-foreground">siap dipakai untuk kelas, nilai, dan CBT</p>
				</div>
				<div class="rounded-2xl border border-warning/30 bg-warning/10 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-warning">Relasi Ortu</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{overview.students.filter((item) => item.linked_parent_count > 0).length}</p>
					<p class="text-sm text-muted-foreground">siswa yang sudah terhubung ke akun orang tua</p>
				</div>
			</div>
		{/snippet}
	</AsyncContent>

	<Card.Root class="overflow-hidden border-border shadow-sm">
		<Card.Header class="pb-3">
			<div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
				<div>
					<Card.Title class="text-base">Akun Siswa</Card.Title>
					<p class="mt-1 text-sm text-muted-foreground">Password awal memakai NISN siswa dan hanya tampil sekali. Simpan/unduh hasil generate sekarang.</p>
					<p class="mt-1 text-sm text-muted-foreground">Akun wajib mengganti password saat login pertama.</p>
					<p class="mt-1 text-sm text-muted-foreground">Data resmi tetap dikunci dan perubahan melalui approval.</p>
				</div>
				<div class="flex flex-wrap gap-2">
					<LoadingButton
						variant="outline"
						loading={studentAccountBusy === 'preview'}
						loadingLabel="Memuat..."
						disabled={!canManageStudentAccounts || studentAccountBusy !== ''}
						onclick={() => void previewAccounts()}
						label="Preview akun siswa"
					/>
					<LoadingButton
						loading={studentAccountBusy === 'generate'}
						loadingLabel="Generate..."
						disabled={!canManageStudentAccounts || studentAccountBusy !== ''}
						onclick={() => void generateAccounts()}
						label="Generate akun siswa"
					/>
				</div>
			</div>
			{#if !canManageStudentAccounts}
				<p class="mt-3 rounded-md border border-border bg-muted/60 px-3 py-2 text-sm text-muted-foreground">
					Aksi akun siswa memerlukan permission student_accounts.manage.
				</p>
			{/if}
		</Card.Header>
		{#if studentAccountSummary}
			<Card.Content class="border-t border-border p-0">
				<div class="grid gap-3 border-b border-border p-4 sm:grid-cols-4">
					<div>
						<p class="text-xs text-muted-foreground">Total</p>
						<p class="text-xl font-semibold">{studentAccountSummary.total}</p>
					</div>
					<div>
						<p class="text-xs text-muted-foreground">Siap</p>
						<p class="text-xl font-semibold">{studentAccountSummary.ready}</p>
					</div>
					<div>
						<p class="text-xs text-muted-foreground">Dibuat</p>
						<p class="text-xl font-semibold">{studentAccountSummary.created}</p>
					</div>
					<div>
						<p class="text-xs text-muted-foreground">Dilewati/Gagal</p>
						<p class="text-xl font-semibold">{studentAccountSummary.skipped + studentAccountSummary.failed}</p>
					</div>
				</div>
				<div class="max-h-80 overflow-auto">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>Nama</Table.Head>
								<Table.Head>Username</Table.Head>
								<Table.Head>Password awal</Table.Head>
								<Table.Head>Status akun</Table.Head>
								<Table.Head>Keterangan</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each studentAccountSummary.candidates as candidate (candidate.student_id)}
								<Table.Row>
									<Table.Cell class="font-medium">{candidate.nama}</Table.Cell>
									<Table.Cell class="font-mono text-sm">{candidate.generated_username || '—'}</Table.Cell>
									<Table.Cell class="font-mono text-sm">{candidate.temporary_password || '—'}</Table.Cell>
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

	<EntityDrawer
		bind:open={showForm}
		title={editId ? 'Edit Data Siswa' : 'Tambah Siswa Baru'}
		subtitle={editId ? 'Perbarui identitas, kelas, wali, dan status siswa.' : 'Isi data pokok siswa untuk kelas, nilai, portal orang tua, dan CBT.'}
		hasUnsavedChanges={studentFormDirty}
		closeLabel="Tutup drawer siswa"
	>
				<div class="mb-4 grid gap-2 md:grid-cols-4">
					{#each studentFormSteps as step, index (step.id)}
						<button
							type="button"
							class={`rounded-2xl border px-3 py-3 text-left transition-colors ${studentFormStep === step.id
								? 'border-primary/20 bg-primary/10 text-primary'
								: 'border-border bg-muted/50 text-muted-foreground hover:bg-card'}`}
							onclick={() => (studentFormStep = step.id)}
						>
							<span class="text-[10px] font-semibold uppercase tracking-[0.18em]">Langkah {index + 1}</span>
							<span class="mt-1 block text-sm font-semibold">{step.label}</span>
							<span class="mt-1 block text-xs leading-5">{step.description}</span>
						</button>
					{/each}
				</div>

				<div class="rounded-2xl border border-border bg-card p-4">
					{#if studentFormStep === 'identity'}
						<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
							<div>
								<label for="s-nis" class="text-xs text-muted-foreground mb-1 block">NIS <span class="text-destructive">*</span></label>
								<Input id="s-nis" placeholder="Masukkan NIS siswa" bind:value={formNis} />
							</div>
							<div>
								<label for="s-nisn" class="text-xs text-muted-foreground mb-1 block">NISN</label>
								<Input id="s-nisn" placeholder="Isi jika sudah tersedia" bind:value={formNisn} />
							</div>
							<div class="lg:col-span-2">
								<label for="s-nama" class="text-xs text-muted-foreground mb-1 block">Nama Lengkap <span class="text-destructive">*</span></label>
								<Input id="s-nama" placeholder="Masukkan nama lengkap siswa" bind:value={formNama} />
							</div>
							<div>
								<label for="s-gender" class="text-xs text-muted-foreground mb-1 block">Jenis Kelamin <span class="text-destructive">*</span></label>
								<select id="s-gender" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={formGender}>
									<option value="L">Laki-laki</option>
									<option value="P">Perempuan</option>
								</select>
							</div>
						</div>
					{:else if studentFormStep === 'class'}
						<div class="grid gap-3 sm:grid-cols-2">
							<div>
								<label for="s-class" class="text-xs text-muted-foreground mb-1 block">Kelas</label>
								<select id="s-class" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={formClassId}>
									<option value="">-- Belum ada kelas --</option>
									{#each classes as c (c.id)}
										<option value={c.id}>{c.code} — {c.name}</option>
									{/each}
								</select>
							</div>
							<div>
								<label for="s-status" class="text-xs text-muted-foreground mb-1 block">Lifecycle Siswa</label>
								<select id="s-status" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={formStatus}>
									<option value="prospective">Calon Siswa</option>
									<option value="active">Aktif</option>
									<option value="alumni">Alumni</option>
									<option value="mutated">Mutasi</option>
								</select>
							</div>
						</div>
					{:else if studentFormStep === 'guardian'}
						<div class="grid gap-3 sm:grid-cols-2">
							<div>
								<label for="s-wali" class="text-xs text-muted-foreground mb-1 block">Nama Wali</label>
								<Input id="s-wali" placeholder="Nama orang tua atau wali utama" bind:value={formParentName} />
							</div>
							<div>
								<label for="s-hp" class="text-xs text-muted-foreground mb-1 block">HP Wali</label>
								<Input id="s-hp" placeholder="Nomor WhatsApp yang aktif" bind:value={formParentPhone} />
							</div>
						</div>
					{:else}
						<div class="grid gap-3 lg:grid-cols-[1fr_1.2fr]">
							<div>
								<label for="s-active" class="text-xs text-muted-foreground mb-1 block">Status Aktif</label>
								<select id="s-active" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={formActive}>
									<option value={true}>Aktif</option>
									<option value={false}>Nonaktif</option>
								</select>
							</div>
							<div class="rounded-xl border border-primary/20 bg-primary/10 px-4 py-3 text-sm text-primary">
								<p class="font-semibold">{formNama || 'Nama siswa belum diisi'}</p>
								<p class="mt-1">NIS {formNis || '-'} · {formClassId ? 'Kelas dipilih' : 'Belum ada kelas'} · {formStatus}</p>
							</div>
						</div>
					{/if}
				</div>

		{#snippet actions()}
				<div class="flex flex-wrap gap-2">
					{#if studentFormStepIndex > 0}
						<Button variant="outline" onclick={previousStudentFormStep}>Kembali</Button>
					{/if}
					{#if studentFormStepIndex < studentFormSteps.length - 1}
						<Button variant="outline" onclick={nextStudentFormStep}>Lanjut</Button>
					{/if}
					<LoadingButton
						loading={formBusy}
						loadingLabel="Menyimpan..."
						disabled={!formNis || !formNama}
						onclick={() => void saveStudent()}
						label={editId ? 'Perbarui Siswa' : 'Simpan Siswa'}
					/>
					<Button variant="outline" onclick={resetForm}>Batal</Button>
				</div>
		{/snippet}
	</EntityDrawer>

	<AsyncContent promise={studentsPromise} onerror={handleOverviewRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<div class="flex flex-wrap items-start justify-between gap-4">
					<div class="space-y-2">
						<Skeleton class="h-8 w-48" />
						<Skeleton class="h-4 w-72" />
					</div>
					<Skeleton class="h-9 w-36" />
				</div>
				<Skeleton class="h-12 w-full" />
				<Skeleton class="h-14 w-full" />
				<Skeleton class="h-14 w-full" />
				<Skeleton class="h-14 w-full" />
			</div>
		{/snippet}

		{#snippet children(value)}
			{@const overview = value as StudentsOverview}
		<Card.Root>
				<Card.Header class="pb-3">
					<div class="flex flex-col sm:flex-row sm:items-center gap-3">
						<Card.Title class="text-base shrink-0">Daftar Siswa ({overview.students.length} total)</Card.Title>
						<Input placeholder="Cari siswa berdasarkan nama, NIS, atau NISN..." bind:value={search} class="w-full sm:max-w-xs sm:ml-auto" />
					</div>
					{#if selectedCount > 0}
						<div class="mt-3 flex flex-wrap items-center justify-between gap-3 rounded-md border border-primary/20 bg-primary/10 px-3 py-2">
							<div>
								<p class="text-sm font-medium text-primary">{selectedCount} siswa dipilih</p>
							</div>
							<Button variant="outline" size="sm" class="bg-card" onclick={clearSelection}>Bersihkan pilihan</Button>
						</div>
					{/if}
				</Card.Header>
				<Card.Content class="p-0">
					<div class="hidden overflow-x-auto lg:block">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head class="w-12">
									<input
										bind:this={visibleSelectionCheckbox}
										type="checkbox"
										checked={allFilteredStudentsSelected}
										disabled={filteredStudentIds.length === 0}
										aria-label="Pilih semua siswa pada hasil filter"
										onchange={handleFilteredSelectionChange}
										class="rounded accent-green-700"
									/>
								</Table.Head>
								<Table.Head>NIS</Table.Head>
								<Table.Head>Nama</Table.Head>
							<Table.Head>L/P</Table.Head>
							<Table.Head>Kelas</Table.Head>
							<Table.Head>Wali</Table.Head>
							<Table.Head>Status</Table.Head>
							<Table.Head>Lifecycle</Table.Head>
							<Table.Head>Status akun</Table.Head>
							<Table.Head>Username</Table.Head>
							<Table.Head>Role</Table.Head>
							<Table.Head class="text-right">Aksi</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each filtered as s (s.id)}
							{@const activeKey = lifecycleKey(s.id, 'active')}
							{@const alumniKey = lifecycleKey(s.id, 'alumni')}
								{@const mutatedKey = lifecycleKey(s.id, 'mutated')}
								{@const accountCandidate = studentAccountCandidate(s)}
								<Table.Row>
									<Table.Cell>
										<input
											type="checkbox"
											checked={selectedStudentIds.has(s.id)}
											aria-label={`Pilih ${s.nama}`}
											onchange={(event) => handleStudentSelectionChange(event, s.id)}
											class="rounded accent-green-700"
										/>
									</Table.Cell>
									<Table.Cell class="font-mono text-sm">{s.nis}</Table.Cell>
									<Table.Cell class="font-medium">{s.nama}</Table.Cell>
								<Table.Cell>
									<Badge variant="outline" class="text-xs">
										{s.gender === 'L' ? 'L' : 'P'}
									</Badge>
								</Table.Cell>
								<Table.Cell class="text-muted-foreground">{s.class_code || '—'}</Table.Cell>
								<Table.Cell class="text-muted-foreground text-sm">
									<div class="max-w-56">
										<p class="truncate">{parentSummary(s)}</p>
										{#if s.linked_parent_count > 0}
											<p class="mt-1 text-[11px] text-primary">Tautan akun orang tua aktif</p>
										{/if}
									</div>
								</Table.Cell>
								<Table.Cell>
									{#if s.is_active}
										<Badge class="bg-primary/15 text-primary border-primary/20">Aktif</Badge>
									{:else}
										<Badge variant="secondary">Nonaktif</Badge>
									{/if}
								</Table.Cell>
								<Table.Cell>
									<Badge class={lifecycleBadgeClass(s.status)}>{s.status}</Badge>
								</Table.Cell>
								<Table.Cell>
									<Badge class={accountStatusClass(accountCandidate)}>{accountStatusLabel(accountCandidate)}</Badge>
								</Table.Cell>
								<Table.Cell class="font-mono text-sm">{accountCandidate?.generated_username || '—'}</Table.Cell>
								<Table.Cell class="text-sm">{accountCandidate?.role || '—'}</Table.Cell>
								<Table.Cell class="text-right">
									<div class="flex gap-2 justify-end">
										<LoadingButton
											variant="outline"
											size="sm"
											onclick={() => updateLifecycle(s, 'active')}
											loading={lifecycleBusyKey === activeKey}
											loadingLabel="Memproses..."
											disabled={deleteBusyId !== '' || (lifecycleBusyKey !== '' && lifecycleBusyKey !== activeKey)}
										>Aktif</LoadingButton>
										<LoadingButton
											variant="outline"
											size="sm"
											onclick={() => updateLifecycle(s, 'alumni')}
											loading={lifecycleBusyKey === alumniKey}
											loadingLabel="Memproses..."
											disabled={deleteBusyId !== '' || (lifecycleBusyKey !== '' && lifecycleBusyKey !== alumniKey)}
										>Alumni</LoadingButton>
										<LoadingButton
											variant="outline"
											size="sm"
											onclick={() => updateLifecycle(s, 'mutated')}
											loading={lifecycleBusyKey === mutatedKey}
											loadingLabel="Memproses..."
											disabled={deleteBusyId !== '' || (lifecycleBusyKey !== '' && lifecycleBusyKey !== mutatedKey)}
										>Mutasi</LoadingButton>
										<Button variant="outline" size="sm" onclick={() => openEdit(s)}>Edit</Button>
										<LoadingButton
											variant="destructive"
											size="sm"
											onclick={() => deleteStudent(s.id, s.nama)}
											loading={deleteBusyId === s.id}
											loadingLabel="Menghapus..."
											disabled={lifecycleBusyKey !== '' || (deleteBusyId !== '' && deleteBusyId !== s.id)}
										>Hapus</LoadingButton>
									</div>
								</Table.Cell>
							</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={12} class="p-4">
											<EmptyStatePanel
											compact
											eyebrow={search ? 'Filter Tidak Menemukan Hasil' : 'Mulai Data Pokok'}
											title={search ? 'Tidak ada siswa yang cocok' : 'Belum ada data siswa'}
											description={search
												? 'Coba ganti kata kunci pencarian, atau kosongkan filter untuk melihat seluruh daftar siswa.'
												: 'Tambahkan siswa pertama agar modul kelas, orang tua, nilai, dan CBT bisa mulai terhubung.'}
										>
											{#if search}
												<Button variant="outline" size="sm" onclick={() => (search = '')}>Reset pencarian</Button>
											{:else}
												<Button size="sm" onclick={openCreate}>Tambah siswa pertama</Button>
											{/if}
										</EmptyStatePanel>
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
				</div>

				<div class="grid gap-3 p-4 lg:hidden">
					{#each filtered as s (s.id)}
						{@const activeKey = lifecycleKey(s.id, 'active')}
						{@const alumniKey = lifecycleKey(s.id, 'alumni')}
						{@const mutatedKey = lifecycleKey(s.id, 'mutated')}
						{@const accountCandidate = studentAccountCandidate(s)}
						<div class="rounded-2xl border border-border bg-card p-4 shadow-sm">
							<div class="flex items-start justify-between gap-3">
								<div class="flex min-w-0 items-start gap-3">
									<input
										type="checkbox"
										checked={selectedStudentIds.has(s.id)}
										aria-label={`Pilih ${s.nama}`}
										onchange={(event) => handleStudentSelectionChange(event, s.id)}
										class="mt-1 rounded accent-green-700"
									/>
									<div class="min-w-0">
										<p class="text-sm font-semibold text-foreground">{s.nama}</p>
										<p class="mt-1 font-mono text-xs text-muted-foreground">NIS {s.nis}{s.nisn ? ` • NISN ${s.nisn}` : ''}</p>
									</div>
								</div>
								{#if s.is_active}
									<Badge class="bg-primary/15 text-primary border-primary/20">Aktif</Badge>
								{:else}
									<Badge variant="secondary">Nonaktif</Badge>
								{/if}
							</div>
							<div class="mt-3 flex flex-wrap items-center gap-2">
								<Badge variant="outline" class="text-xs">{s.gender === 'L' ? 'Laki-laki' : 'Perempuan'}</Badge>
								<Badge variant="outline" class="text-xs">{s.class_code || 'Belum ada kelas'}</Badge>
								<Badge class={lifecycleBadgeClass(s.status)}>{s.status}</Badge>
								<Badge class={accountStatusClass(accountCandidate)}>{accountStatusLabel(accountCandidate)}</Badge>
							</div>
							<p class="mt-2 font-mono text-xs text-muted-foreground">
								Username {accountCandidate?.generated_username || '—'} · Role {accountCandidate?.role || '—'}
							</p>
							<p class="mt-3 text-sm text-muted-foreground">{parentSummary(s)}</p>
							{#if s.linked_parent_count > 0}
								<p class="mt-1 text-xs text-primary">Relasi orang tua terhubung ke akun portal</p>
							{/if}
							<div class="mt-4 grid grid-cols-2 gap-2">
								<LoadingButton
									variant="outline"
									size="sm"
									onclick={() => updateLifecycle(s, 'active')}
									loading={lifecycleBusyKey === activeKey}
									loadingLabel="Memproses..."
									disabled={deleteBusyId !== '' || (lifecycleBusyKey !== '' && lifecycleBusyKey !== activeKey)}
								>Aktif</LoadingButton>
								<LoadingButton
									variant="outline"
									size="sm"
									onclick={() => updateLifecycle(s, 'alumni')}
									loading={lifecycleBusyKey === alumniKey}
									loadingLabel="Memproses..."
									disabled={deleteBusyId !== '' || (lifecycleBusyKey !== '' && lifecycleBusyKey !== alumniKey)}
								>Alumni</LoadingButton>
								<LoadingButton
									variant="outline"
									size="sm"
									onclick={() => updateLifecycle(s, 'mutated')}
									loading={lifecycleBusyKey === mutatedKey}
									loadingLabel="Memproses..."
									disabled={deleteBusyId !== '' || (lifecycleBusyKey !== '' && lifecycleBusyKey !== mutatedKey)}
								>Mutasi</LoadingButton>
								<Button variant="outline" size="sm" onclick={() => openEdit(s)}>Edit</Button>
								<LoadingButton
									variant="destructive"
									size="sm"
									onclick={() => deleteStudent(s.id, s.nama)}
									loading={deleteBusyId === s.id}
									loadingLabel="Menghapus..."
									disabled={lifecycleBusyKey !== '' || (deleteBusyId !== '' && deleteBusyId !== s.id)}
								>Hapus</LoadingButton>
							</div>
						</div>
					{:else}
							<EmptyStatePanel
								eyebrow={search ? 'Filter Tidak Menemukan Hasil' : 'Mulai Data Pokok'}
								title={search ? 'Tidak ada siswa yang cocok' : 'Belum ada data siswa'}
								description={search
									? 'Ubah kata kunci pencarian atau reset filter untuk melihat kembali seluruh daftar siswa.'
									: 'Tambahkan siswa pertama dari form di atas agar data akademik dan portal orang tua bisa mulai berjalan.'}
							>
								{#if search}
									<Button variant="outline" size="sm" onclick={() => (search = '')}>Reset pencarian</Button>
								{:else}
									<Button size="sm" onclick={openCreate}>Tambah siswa pertama</Button>
								{/if}
							</EmptyStatePanel>
					{/each}
				</div>
			</Card.Content>
		</Card.Root>
		{/snippet}
	</AsyncContent>
</div>
