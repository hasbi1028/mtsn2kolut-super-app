<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { toast } from '$lib/components/ui/sonner';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import SuccessPanel from '$lib/components/SuccessPanel.svelte';

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

	let parents = $state<Parent[]>([]);
	let students = $state<Student[]>([]);
	let loading = $state(true);
	let error = $state('');
	let success = $state('');
	let showForm = $state(false);
	let showLinkDialog = $state(false);
	let selectedParent = $state<Parent | null>(null);
	let children = $state<Student[]>([]);

	let fNama = $state('');
	let fPhone = $state('');
	let fAddress = $state('');
	let fSelectedStudentId = $state('');
	let fBusy = $state(false);

	async function load() {
		try {
			error = '';
			const [pRes, sRes] = await Promise.all([
				fetch('/api/parents'),
				fetch('/api/students')
			]);
			parents = await pRes.json();
			students = await sRes.json();
		} catch {
			error = 'Gagal memuat data orang tua dan daftar siswa. Coba lagi untuk melanjutkan pengelolaan relasi keluarga.';
		} finally {
			loading = false;
		}
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
			if (!res.ok) throw new Error();
			fNama = ''; fPhone = ''; fAddress = '';
			showForm = false;
			toast.success('Data orang tua berhasil disimpan');
			success = 'Profil orang tua berhasil ditambahkan. Selanjutnya kamu bisa membuka relasi anak untuk mulai menautkan siswa.';
			await load();
		} catch {
			toast.error('Gagal menyimpan data');
		} finally { fBusy = false; }
	}

	async function openLinkDialog(parent: Parent) {
		selectedParent = parent;
		showLinkDialog = true;
		const res = await fetch(`/api/parents/${parent.id}/children`);
		children = await res.json();
	}

	async function linkStudent() {
		if (!selectedParent || !fSelectedStudentId) return;
		fBusy = true;
		try {
			success = '';
			const res = await fetch(`/api/parents/${selectedParent.id}/link`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ student_id: fSelectedStudentId })
			});
			if (!res.ok) throw new Error();
			toast.success('Siswa berhasil ditautkan');
			fSelectedStudentId = '';
			success = `Relasi keluarga berhasil diperbarui. ${selectedParent.nama} sekarang terhubung dengan siswa yang dipilih.`;
			// Refresh children list
			const cRes = await fetch(`/api/parents/${selectedParent.id}/children`);
			children = await cRes.json();
		} catch {
			toast.error('Gagal menautkan siswa');
		} finally { fBusy = false; }
	}

	async function unlinkStudent(studentId: string, studentName: string) {
		if (!selectedParent) return;
		if (!confirm(`Lepas tautan ${studentName} dari ${selectedParent.nama}?`)) return;
		fBusy = true;
		try {
			success = '';
			const res = await fetch(`/api/parents/${selectedParent.id}/unlink`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ student_id: studentId })
			});
			if (!res.ok) throw new Error();
			toast.success('Tautan siswa berhasil dilepas');
			success = `Tautan ${studentName} berhasil dilepas dari ${selectedParent.nama}.`;
			const cRes = await fetch(`/api/parents/${selectedParent.id}/children`);
			children = await cRes.json();
		} catch {
			toast.error('Gagal melepas tautan siswa');
		} finally { fBusy = false; }
	}

	onMount(load);
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

	<div class="grid gap-3 md:grid-cols-3">
		<div class="rounded-2xl border border-emerald-100 bg-emerald-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-emerald-700">Data Orang Tua</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{parents.length}</p>
			<p class="text-sm text-slate-600">profil wali murid yang sudah tercatat</p>
		</div>
		<div class="rounded-2xl border border-sky-100 bg-sky-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-sky-700">Siswa Tersedia</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{students.length}</p>
			<p class="text-sm text-slate-600">daftar siswa yang bisa ditautkan ke akun orang tua</p>
		</div>
		<div class="rounded-2xl border border-amber-100 bg-amber-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-amber-700">Relasi Keluarga</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{children.length}</p>
			<p class="text-sm text-slate-600">anak yang sedang tampil pada panel relasi aktif</p>
		</div>
	</div>

	{#if error}
		<RecoveryPanel title="Relasi Orang Tua Belum Tersaji" message={error} onRetry={load} />
	{/if}

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
					<LoadingButton onclick={saveParent} loading={fBusy} disabled={fBusy || !fNama}>
						Simpan Data
					</LoadingButton>
					<Button variant="outline" onclick={() => (showForm = false)}>Batal</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
		<Card.Content class="p-0">
			{#if loading && parents.length === 0}
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
			{:else}
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
						{#each parents as p (p.id)}
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
			{/if}
		</Card.Content>
	</Card.Root>
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
					{#each students as s}
						<option value={s.id}>{s.nama} ({s.nis})</option>
					{/each}
				</select>
				<LoadingButton size="sm" onclick={linkStudent} loading={fBusy} disabled={fBusy || !fSelectedStudentId}>
					Tautkan
				</LoadingButton>
			</div>

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
						{#each children as c}
							<Table.Row>
								<Table.Cell class="py-2">{c.nama}</Table.Cell>
								<Table.Cell class="py-2 text-xs">{c.class_name || '—'}</Table.Cell>
								<Table.Cell class="py-2 text-right">
									<Button size="sm" variant="ghost" class="text-red-600 hover:bg-red-50 hover:text-red-700" onclick={() => unlinkStudent(c.id, c.nama)}>
										Lepas
									</Button>
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
		</div>
		</div>
	</Dialog.Content>
</Dialog.Root>
