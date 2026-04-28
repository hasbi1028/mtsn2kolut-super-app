<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';

	type Student = {
		id: string; nis: string; nisn: string; nama: string; gender: string;
		parent_name: string; parent_phone: string;
		class_id: string; class_name: string; class_code: string;
		is_active: boolean; created_at: string;
	};
	type SchoolClass = { id: string; name: string; code: string; level: string; };

	let students = $state<Student[]>([]);
	let classes = $state<SchoolClass[]>([]);
	let loading = $state(true);
	let error = $state('');
	let toast = $state('');
	let search = $state('');

	let formNis = $state('');
	let formNisn = $state('');
	let formNama = $state('');
	let formGender = $state('L');
	let formParentName = $state('');
	let formParentPhone = $state('');
	let formClassId = $state('');
	let formActive = $state(true);
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
		toast = msg;
		setTimeout(() => (toast = ''), 3000);
	}

	function resetForm() {
		formNis = ''; formNisn = ''; formNama = ''; formGender = 'L';
		formParentName = ''; formParentPhone = ''; formClassId = ''; formActive = true;
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
					class_id: formClassId, is_active: formActive,
				}),
			});
			if (!res.ok) { const j = await res.json(); showToast(j.error ?? 'Gagal'); return; }
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

	{#if toast}
		<div class="rounded-md bg-emerald-50 border border-emerald-200 px-4 py-3 text-sm text-emerald-800">{toast}</div>
	{/if}

	{#if error}
		<div class="rounded-md bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-800">{error}</div>
	{/if}

	{#if showForm}
		<Card.Root>
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
					<Button disabled={formBusy || !formNis || !formNama} onclick={saveStudent}>
						{formBusy ? 'Menyimpan...' : (editId ? 'Perbarui Siswa' : 'Simpan Siswa')}
					</Button>
					<Button variant="outline" onclick={resetForm}>Batal</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	{#if loading}
		<p class="text-sm text-slate-500">Memuat data...</p>
	{:else}
		<Card.Root>
			<Card.Header class="pb-3">
				<div class="flex flex-col sm:flex-row sm:items-center gap-3">
					<Card.Title class="text-base shrink-0">Daftar Siswa ({students.length} total)</Card.Title>
					<Input placeholder="Cari nama, NIS, NISN..." bind:value={search} class="w-full sm:max-w-xs sm:ml-auto" />
				</div>
			</Card.Header>
			<Card.Content class="p-0 overflow-x-auto">
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>NIS</Table.Head>
							<Table.Head>Nama</Table.Head>
							<Table.Head>L/P</Table.Head>
							<Table.Head>Kelas</Table.Head>
							<Table.Head>Wali</Table.Head>
							<Table.Head>Status</Table.Head>
							<Table.Head class="text-right">Aksi</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each filtered as s}
							<Table.Row>
								<Table.Cell class="font-mono text-sm">{s.nis}</Table.Cell>
								<Table.Cell class="font-medium">{s.nama}</Table.Cell>
								<Table.Cell>
									<Badge variant="outline" class="text-xs">
										{s.gender === 'L' ? 'L' : 'P'}
									</Badge>
								</Table.Cell>
								<Table.Cell class="text-slate-500">{s.class_code || '—'}</Table.Cell>
								<Table.Cell class="text-slate-500 text-sm">{s.parent_name || '—'}</Table.Cell>
								<Table.Cell>
									{#if s.is_active}
										<Badge class="bg-emerald-100 text-emerald-700 border-emerald-200">Aktif</Badge>
									{:else}
										<Badge variant="secondary">Tidak Aktif</Badge>
									{/if}
								</Table.Cell>
								<Table.Cell class="text-right">
									<div class="flex gap-2 justify-end">
										<Button variant="outline" size="sm" onclick={() => openEdit(s)}>Edit</Button>
										<Button variant="destructive" size="sm" onclick={() => deleteStudent(s.id, s.nama)}>Hapus</Button>
									</div>
								</Table.Cell>
							</Table.Row>
						{:else}
							<Table.Row>
								<Table.Cell colspan={7} class="text-center text-slate-400 py-8">
									{search ? 'Tidak ada hasil pencarian' : 'Belum ada data siswa'}
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</Card.Content>
		</Card.Root>
	{/if}
</div>
