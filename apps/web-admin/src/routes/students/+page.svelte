<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { confirmAction } from '$lib/confirm-dialog';
	import { clientApiPathWithQuery, readClientApiData, readClientJson } from '$lib/client/api';

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

	const studentFormSteps: Array<{ id: StudentFormStep; label: string; description: string }> = [
		{ id: 'identity', label: 'Identitas', description: 'NIS, NISN, nama, dan gender' },
		{ id: 'class', label: 'Kelas', description: 'Kelas aktif dan lifecycle' },
		{ id: 'guardian', label: 'Wali', description: 'Kontak orang tua/wali' },
		{ id: 'status', label: 'Status', description: 'Aktif/nonaktif dan review akhir' }
	];

	const studentFormStepIndex = $derived(studentFormSteps.findIndex((step) => step.id === studentFormStep));

	let filtered = $derived(
		search.trim()
			? students.filter(s =>
				s.nama.toLowerCase().includes(search.toLowerCase()) ||
				s.nis.includes(search) ||
				(s.nisn ?? '').includes(search)
			)
			: students
	);

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

	function resetForm() {
		formNis = ''; formNisn = ''; formNama = ''; formGender = 'L';
		formParentName = ''; formParentPhone = ''; formClassId = ''; formActive = true; formStatus = 'active';
		editId = null;
		studentFormStep = 'identity';
		showForm = false;
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
		window.scrollTo({ top: 0, behavior: 'smooth' });
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
		if (status === 'prospective') return 'bg-sky-100 text-sky-700 border-sky-200';
		if (status === 'alumni') return 'bg-sky-100 text-sky-700 border-sky-200';
		if (status === 'mutated') return 'bg-amber-100 text-amber-700 border-amber-200';
		return 'bg-emerald-100 text-emerald-700 border-emerald-200';
	}

	function parentSummary(student: Student) {
		if (student.linked_parent_count > 0) {
			return student.linked_parent_count > 1
				? `${student.linked_parent_names} (${student.linked_parent_count} relasi)`
				: student.linked_parent_names;
		}
		return student.parent_name || 'Wali belum diisi';
	}

	onMount(() => {
		void load();
	});
</script>

<svelte:head><title>Data Siswa — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-slate-800">Data Siswa</h1>
			<p class="text-sm text-slate-500 mt-1">Kelola daftar siswa aktif madrasah</p>
		</div>
		<Button onclick={() => { if (showForm) resetForm(); else showForm = true; }}>
			{showForm ? 'Batal' : '+ Tambah Siswa'}
		</Button>
	</div>

	<AsyncContent promise={studentsPromise} onerror={handleOverviewRenderError}>
		{#snippet pending()}
			<div class="grid gap-3 md:grid-cols-3">
				{#each ['Total Siswa', 'Siswa Aktif', 'Relasi Ortu'] as label (label)}
					<div class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-4">
						<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-slate-500">{label}</p>
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
				<div class="rounded-2xl border border-emerald-100 bg-emerald-50 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-emerald-700">Total Siswa</p>
					<p class="mt-2 text-2xl font-semibold text-slate-900">{overview.students.length}</p>
					<p class="text-sm text-slate-600">seluruh entitas siswa yang sudah tersimpan</p>
				</div>
				<div class="rounded-2xl border border-sky-100 bg-sky-50 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-sky-700">Siswa Aktif</p>
					<p class="mt-2 text-2xl font-semibold text-slate-900">{overview.students.filter((item) => item.status === 'active').length}</p>
					<p class="text-sm text-slate-600">siap dipakai untuk kelas, nilai, dan CBT</p>
				</div>
				<div class="rounded-2xl border border-amber-100 bg-amber-50 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-amber-700">Relasi Ortu</p>
					<p class="mt-2 text-2xl font-semibold text-slate-900">{overview.students.filter((item) => item.linked_parent_count > 0).length}</p>
					<p class="text-sm text-slate-600">siswa yang sudah terhubung ke akun orang tua</p>
				</div>
			</div>
		{/snippet}
	</AsyncContent>

	{#if showForm}
		<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
			<Card.Header class="pb-2">
				<Card.Title class="text-base">{editId ? 'Edit Data Siswa' : 'Tambah Siswa Baru'}</Card.Title>
			</Card.Header>
			<Card.Content>
				<div class="mb-4 grid gap-2 md:grid-cols-4">
					{#each studentFormSteps as step, index (step.id)}
						<button
							type="button"
							class={`rounded-2xl border px-3 py-3 text-left transition-colors ${studentFormStep === step.id
								? 'border-emerald-200 bg-emerald-50 text-emerald-900'
								: 'border-slate-200 bg-slate-50 text-slate-600 hover:bg-white'}`}
							onclick={() => (studentFormStep = step.id)}
						>
							<span class="text-[10px] font-semibold uppercase tracking-[0.18em]">Langkah {index + 1}</span>
							<span class="mt-1 block text-sm font-semibold">{step.label}</span>
							<span class="mt-1 block text-xs leading-5">{step.description}</span>
						</button>
					{/each}
				</div>

				<div class="rounded-2xl border border-slate-200 bg-white p-4">
					{#if studentFormStep === 'identity'}
						<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
							<div>
								<label for="s-nis" class="text-xs text-slate-500 mb-1 block">NIS <span class="text-red-500">*</span></label>
								<Input id="s-nis" placeholder="Masukkan NIS siswa" bind:value={formNis} />
							</div>
							<div>
								<label for="s-nisn" class="text-xs text-slate-500 mb-1 block">NISN</label>
								<Input id="s-nisn" placeholder="Isi jika sudah tersedia" bind:value={formNisn} />
							</div>
							<div class="lg:col-span-2">
								<label for="s-nama" class="text-xs text-slate-500 mb-1 block">Nama Lengkap <span class="text-red-500">*</span></label>
								<Input id="s-nama" placeholder="Masukkan nama lengkap siswa" bind:value={formNama} />
							</div>
							<div>
								<label for="s-gender" class="text-xs text-slate-500 mb-1 block">Jenis Kelamin <span class="text-red-500">*</span></label>
								<select id="s-gender" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={formGender}>
									<option value="L">Laki-laki</option>
									<option value="P">Perempuan</option>
								</select>
							</div>
						</div>
					{:else if studentFormStep === 'class'}
						<div class="grid gap-3 sm:grid-cols-2">
							<div>
								<label for="s-class" class="text-xs text-slate-500 mb-1 block">Kelas</label>
								<select id="s-class" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={formClassId}>
									<option value="">-- Belum ada kelas --</option>
									{#each classes as c (c.id)}
										<option value={c.id}>{c.code} — {c.name}</option>
									{/each}
								</select>
							</div>
							<div>
								<label for="s-status" class="text-xs text-slate-500 mb-1 block">Lifecycle Siswa</label>
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
								<label for="s-wali" class="text-xs text-slate-500 mb-1 block">Nama Wali</label>
								<Input id="s-wali" placeholder="Nama orang tua atau wali utama" bind:value={formParentName} />
							</div>
							<div>
								<label for="s-hp" class="text-xs text-slate-500 mb-1 block">HP Wali</label>
								<Input id="s-hp" placeholder="Nomor WhatsApp yang aktif" bind:value={formParentPhone} />
							</div>
						</div>
					{:else}
						<div class="grid gap-3 lg:grid-cols-[1fr_1.2fr]">
							<div>
								<label for="s-active" class="text-xs text-slate-500 mb-1 block">Status Aktif</label>
								<select id="s-active" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={formActive}>
									<option value={true}>Aktif</option>
									<option value={false}>Nonaktif</option>
								</select>
							</div>
							<div class="rounded-xl border border-emerald-100 bg-emerald-50 px-4 py-3 text-sm text-emerald-900">
								<p class="font-semibold">{formNama || 'Nama siswa belum diisi'}</p>
								<p class="mt-1">NIS {formNis || '-'} · {formClassId ? 'Kelas dipilih' : 'Belum ada kelas'} · {formStatus}</p>
							</div>
						</div>
					{/if}
				</div>

				<div class="mt-4 flex flex-wrap gap-2">
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
			</Card.Content>
		</Card.Root>
	{/if}

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
			</Card.Header>
			<Card.Content class="p-0">
				<div class="hidden overflow-x-auto lg:block">
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>NIS</Table.Head>
							<Table.Head>Nama</Table.Head>
							<Table.Head>L/P</Table.Head>
							<Table.Head>Kelas</Table.Head>
							<Table.Head>Wali</Table.Head>
							<Table.Head>Status</Table.Head>
							<Table.Head>Lifecycle</Table.Head>
							<Table.Head class="text-right">Aksi</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each filtered as s (s.id)}
							{@const activeKey = lifecycleKey(s.id, 'active')}
							{@const alumniKey = lifecycleKey(s.id, 'alumni')}
							{@const mutatedKey = lifecycleKey(s.id, 'mutated')}
							<Table.Row>
								<Table.Cell class="font-mono text-sm">{s.nis}</Table.Cell>
								<Table.Cell class="font-medium">{s.nama}</Table.Cell>
								<Table.Cell>
									<Badge variant="outline" class="text-xs">
										{s.gender === 'L' ? 'L' : 'P'}
									</Badge>
								</Table.Cell>
								<Table.Cell class="text-slate-500">{s.class_code || '—'}</Table.Cell>
								<Table.Cell class="text-slate-500 text-sm">
									<div class="max-w-56">
										<p class="truncate">{parentSummary(s)}</p>
										{#if s.linked_parent_count > 0}
											<p class="mt-1 text-[11px] text-emerald-700">Tautan akun orang tua aktif</p>
										{/if}
									</div>
								</Table.Cell>
								<Table.Cell>
									{#if s.is_active}
										<Badge class="bg-emerald-100 text-emerald-700 border-emerald-200">Aktif</Badge>
									{:else}
										<Badge variant="secondary">Nonaktif</Badge>
									{/if}
								</Table.Cell>
								<Table.Cell>
									<Badge class={lifecycleBadgeClass(s.status)}>{s.status}</Badge>
								</Table.Cell>
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
								<Table.Cell colspan={8} class="p-4">
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
												<Button size="sm" onclick={() => (showForm = true)}>Tambah siswa pertama</Button>
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
						<div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
							<div class="flex items-start justify-between gap-3">
								<div class="min-w-0">
									<p class="text-sm font-semibold text-slate-900">{s.nama}</p>
									<p class="mt-1 font-mono text-xs text-slate-500">NIS {s.nis}{s.nisn ? ` • NISN ${s.nisn}` : ''}</p>
								</div>
								{#if s.is_active}
									<Badge class="bg-emerald-100 text-emerald-700 border-emerald-200">Aktif</Badge>
								{:else}
									<Badge variant="secondary">Nonaktif</Badge>
								{/if}
							</div>
							<div class="mt-3 flex flex-wrap items-center gap-2">
								<Badge variant="outline" class="text-xs">{s.gender === 'L' ? 'Laki-laki' : 'Perempuan'}</Badge>
								<Badge variant="outline" class="text-xs">{s.class_code || 'Belum ada kelas'}</Badge>
								<Badge class={lifecycleBadgeClass(s.status)}>{s.status}</Badge>
							</div>
							<p class="mt-3 text-sm text-slate-600">{parentSummary(s)}</p>
							{#if s.linked_parent_count > 0}
								<p class="mt-1 text-xs text-emerald-700">Relasi orang tua terhubung ke akun portal</p>
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
									<Button size="sm" onclick={() => (showForm = true)}>Tambah siswa pertama</Button>
								{/if}
							</EmptyStatePanel>
					{/each}
				</div>
			</Card.Content>
		</Card.Root>
		{/snippet}
	</AsyncContent>
</div>
