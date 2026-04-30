<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';

	type Student = {
		id: string; nis: string; nisn: string; nama: string; gender: string;
		parent_name: string; parent_phone: string;
		class_id: string; class_name: string; class_code: string;
		linked_parent_names: string; linked_parent_count: number;
		is_active: boolean; status: string; created_at: string;
	};
	type SchoolClass = { id: string; name: string; code: string; level: string; };

	let students = $state<Student[]>([]);
	let classes = $state<SchoolClass[]>([]);
	let loading = $state(true);
	let error = $state('');
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

	let filtered = $derived(
		search.trim()
			? students.filter(s =>
				s.nama.toLowerCase().includes(search.toLowerCase()) ||
				s.nis.includes(search) ||
				(s.nisn ?? '').includes(search)
			)
			: students
	);

	async function load() {
		try {
			const [sRes, aRes] = await Promise.all([
				fetch('/api/students'),
				fetch('/api/academic'),
			]);
			const sJson = await sRes.json();
			const aJson = await aRes.json();
			if (sJson.error) { error = sJson.error; return; }
			students = sJson.data ?? sJson ?? [];
			classes = (aJson.data ?? aJson)?.classes ?? [];
		} catch {
			error = 'Gagal memuat data siswa';
		} finally {
			loading = false;
		}
	}

	function showToast(msg: string) {
		toast.success(msg);
	}

	function showError(msg: string) {
		toast.error(msg);
	}

	function resetForm() {
		formNis = ''; formNisn = ''; formNama = ''; formGender = 'L';
		formParentName = ''; formParentPhone = ''; formClassId = ''; formActive = true; formStatus = 'active';
		editId = null;
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
		showForm = true;
		window.scrollTo({ top: 0, behavior: 'smooth' });
	}

	async function saveStudent() {
		if (!formNis || !formNama || !formGender) return;
		formBusy = true;
		try {
			const method = editId ? 'PUT' : 'POST';
			const path = editId ? `/api/students/${editId}` : '/api/students';
			const res = await fetch(path, {
				method,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					nis: formNis, nisn: formNisn, nama: formNama, gender: formGender,
					parent_name: formParentName, parent_phone: formParentPhone,
					class_id: formClassId, is_active: formActive, status: formStatus,
				}),
			});
			if (!res.ok) { const j = await res.json(); showError(j.error ?? 'Gagal'); return; }
			showToast(editId ? 'Data siswa diperbarui' : 'Siswa berhasil ditambahkan');
			resetForm();
			await load();
		} finally { formBusy = false; }
	}

	async function deleteStudent(id: string, nama: string) {
		if (!confirm(`Hapus siswa "${nama}"?`)) return;
		await fetch(`/api/students?id=${id}`, { method: 'DELETE' });
		showToast('Siswa dihapus');
		await load();
	}

	async function updateLifecycle(student: Student, status: string) {
		const labels: Record<string, string> = {
			prospective: 'Calon Siswa',
			active: 'Aktif',
			alumni: 'Alumni',
			mutated: 'Mutasi',
		};
		if (!confirm(`Ubah status ${student.nama} menjadi ${labels[status] ?? status}?`)) return;
		const res = await fetch(`/api/students?id=${student.id}`, {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ status }),
		});
		if (!res.ok) {
			const payload = await res.json().catch(() => ({}));
			showError(payload.error ?? 'Gagal memperbarui lifecycle siswa');
			return;
		}
		showToast('Lifecycle siswa diperbarui');
		await load();
	}

	function lifecycleBadgeClass(status: string) {
		if (status === 'prospective') return 'bg-sky-100 text-sky-700 border-sky-200';
		if (status === 'alumni') return 'bg-violet-100 text-violet-700 border-violet-200';
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

	onMount(load);
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

	<div class="grid gap-3 md:grid-cols-3">
		<div class="rounded-2xl border border-emerald-100 bg-emerald-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-emerald-700">Total Siswa</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{students.length}</p>
			<p class="text-sm text-slate-600">seluruh entitas siswa yang sudah tersimpan</p>
		</div>
		<div class="rounded-2xl border border-sky-100 bg-sky-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-sky-700">Siswa Aktif</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{students.filter((item) => item.status === 'active').length}</p>
			<p class="text-sm text-slate-600">siap dipakai untuk kelas, nilai, dan CBT</p>
		</div>
		<div class="rounded-2xl border border-amber-100 bg-amber-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-amber-700">Relasi Ortu</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{students.filter((item) => item.linked_parent_count > 0).length}</p>
			<p class="text-sm text-slate-600">siswa yang sudah terhubung ke akun orang tua</p>
		</div>
	</div>

	{#if error}
		<div class="rounded-md bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-800">{error}</div>
	{/if}

	{#if showForm}
		<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
			<Card.Header class="pb-2">
				<Card.Title class="text-base">{editId ? 'Edit Data Siswa' : 'Tambah Siswa Baru'}</Card.Title>
			</Card.Header>
			<Card.Content>
				<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
					<div>
						<label for="s-nis" class="text-xs text-slate-500 mb-1 block">NIS <span class="text-red-500">*</span></label>
						<Input id="s-nis" placeholder="Nomor Induk Siswa" bind:value={formNis} />
					</div>
					<div>
						<label for="s-nisn" class="text-xs text-slate-500 mb-1 block">NISN</label>
						<Input id="s-nisn" placeholder="Nomor Induk Nasional" bind:value={formNisn} />
					</div>
					<div>
						<label for="s-nama" class="text-xs text-slate-500 mb-1 block">Nama Lengkap <span class="text-red-500">*</span></label>
						<Input id="s-nama" placeholder="Nama siswa" bind:value={formNama} />
					</div>
					<div>
						<label for="s-gender" class="text-xs text-slate-500 mb-1 block">Jenis Kelamin <span class="text-red-500">*</span></label>
						<select id="s-gender" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={formGender}>
							<option value="L">Laki-laki</option>
							<option value="P">Perempuan</option>
						</select>
					</div>
					<div>
						<label for="s-class" class="text-xs text-slate-500 mb-1 block">Kelas</label>
						<select id="s-class" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={formClassId}>
							<option value="">-- Belum ada kelas --</option>
							{#each classes as c}
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
					<div>
						<label for="s-active" class="text-xs text-slate-500 mb-1 block">Status Aktif</label>
						<select id="s-active" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={formActive}>
							<option value={true}>Aktif</option>
							<option value={false}>Tidak Aktif</option>
						</select>
					</div>
					<div>
						<label for="s-wali" class="text-xs text-slate-500 mb-1 block">Nama Wali</label>
						<Input id="s-wali" placeholder="Nama orang tua/wali" bind:value={formParentName} />
					</div>
					<div>
						<label for="s-hp" class="text-xs text-slate-500 mb-1 block">HP Wali</label>
						<Input id="s-hp" placeholder="No. HP orang tua" bind:value={formParentPhone} />
					</div>
				</div>
				<div class="mt-4 flex gap-2">
					<LoadingButton
						loading={formBusy}
						loadingLabel="Menyimpan..."
						disabled={!formNis || !formNama}
						onclick={saveStudent}
						label={editId ? 'Perbarui Siswa' : 'Simpan Siswa'}
					/>
					<Button variant="outline" onclick={resetForm}>Batal</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	{#if loading}
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
	{:else}
		<Card.Root>
			<Card.Header class="pb-3">
				<div class="flex flex-col sm:flex-row sm:items-center gap-3">
					<Card.Title class="text-base shrink-0">Daftar Siswa ({students.length} total)</Card.Title>
					<Input placeholder="Cari nama, NIS, NISN..." bind:value={search} class="w-full sm:max-w-xs sm:ml-auto" />
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
										<Badge variant="secondary">Tidak Aktif</Badge>
									{/if}
								</Table.Cell>
								<Table.Cell>
									<Badge class={lifecycleBadgeClass(s.status)}>{s.status}</Badge>
								</Table.Cell>
								<Table.Cell class="text-right">
									<div class="flex gap-2 justify-end">
										<Button variant="outline" size="sm" onclick={() => updateLifecycle(s, 'active')}>Aktif</Button>
										<Button variant="outline" size="sm" onclick={() => updateLifecycle(s, 'alumni')}>Alumni</Button>
										<Button variant="outline" size="sm" onclick={() => updateLifecycle(s, 'mutated')}>Mutasi</Button>
										<Button variant="outline" size="sm" onclick={() => openEdit(s)}>Edit</Button>
										<Button variant="destructive" size="sm" onclick={() => deleteStudent(s.id, s.nama)}>Hapus</Button>
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
										{#snippet children()}
											{#if search}
												<Button variant="outline" size="sm" onclick={() => (search = '')}>Reset pencarian</Button>
											{:else}
												<Button size="sm" onclick={() => (showForm = true)}>Tambah siswa pertama</Button>
											{/if}
										{/snippet}
									</EmptyStatePanel>
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
				</div>

				<div class="grid gap-3 p-4 lg:hidden">
					{#each filtered as s (s.id)}
						<div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
							<div class="flex items-start justify-between gap-3">
								<div class="min-w-0">
									<p class="text-sm font-semibold text-slate-900">{s.nama}</p>
									<p class="mt-1 font-mono text-xs text-slate-500">NIS {s.nis}{s.nisn ? ` • NISN ${s.nisn}` : ''}</p>
								</div>
								{#if s.is_active}
									<Badge class="bg-emerald-100 text-emerald-700 border-emerald-200">Aktif</Badge>
								{:else}
									<Badge variant="secondary">Tidak Aktif</Badge>
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
								<Button variant="outline" size="sm" onclick={() => updateLifecycle(s, 'active')}>Aktif</Button>
								<Button variant="outline" size="sm" onclick={() => updateLifecycle(s, 'alumni')}>Alumni</Button>
								<Button variant="outline" size="sm" onclick={() => updateLifecycle(s, 'mutated')}>Mutasi</Button>
								<Button variant="outline" size="sm" onclick={() => openEdit(s)}>Edit</Button>
								<Button variant="destructive" size="sm" onclick={() => deleteStudent(s.id, s.nama)}>Hapus</Button>
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
							{#snippet children()}
								{#if search}
									<Button variant="outline" size="sm" onclick={() => (search = '')}>Reset pencarian</Button>
								{:else}
									<Button size="sm" onclick={() => (showForm = true)}>Tambah siswa pertama</Button>
								{/if}
							{/snippet}
						</EmptyStatePanel>
					{/each}
				</div>
			</Card.Content>
		</Card.Root>
	{/if}
</div>
