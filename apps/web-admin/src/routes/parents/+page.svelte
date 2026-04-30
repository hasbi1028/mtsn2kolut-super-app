<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';
	import * as Dialog from '$lib/components/ui/dialog';

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
			const [pRes, sRes] = await Promise.all([
				fetch('/api/parents'),
				fetch('/api/students')
			]);
			parents = await pRes.json();
			students = await sRes.json();
		} catch {
			toast.error('Gagal memuat data');
		} finally {
			loading = false;
		}
	}

	async function saveParent() {
		if (!fNama) return;
		fBusy = true;
		try {
			const res = await fetch('/api/parents', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ nama: fNama, phone: fPhone, address: fAddress })
			});
			if (!res.ok) throw new Error();
			fNama = ''; fPhone = ''; fAddress = '';
			showForm = false;
			toast.success('Data orang tua berhasil disimpan');
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
			const res = await fetch(`/api/parents/${selectedParent.id}/link`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ student_id: fSelectedStudentId })
			});
			if (!res.ok) throw new Error();
			toast.success('Siswa berhasil ditautkan');
			fSelectedStudentId = '';
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
			const res = await fetch(`/api/parents/${selectedParent.id}/unlink`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ student_id: studentId })
			});
			if (!res.ok) throw new Error();
			toast.success('Tautan siswa berhasil dilepas');
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
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold text-slate-800">Manajemen Orang Tua</h1>
			<p class="text-sm text-slate-500 mt-1">Kelola data wali murid dan relasi dengan siswa</p>
		</div>
		<Button onclick={() => (showForm = !showForm)}>
			{showForm ? 'Batal' : '+ Tambah Orang Tua'}
		</Button>
	</div>

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
					<Button onclick={saveParent} disabled={fBusy || !fNama}>
						{fBusy ? 'Menyimpan...' : 'Simpan Data'}
					</Button>
					<Button variant="outline" onclick={() => (showForm = false)}>Batal</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

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
					{/each}
				</Table.Body>
			</Table.Root>
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
				<Button size="sm" onclick={linkStudent} disabled={fBusy || !fSelectedStudentId}>
					Tautkan
				</Button>
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
								<Table.Cell colspan={3} class="text-center py-4 text-slate-400 text-sm italic">
									Belum ada siswa yang ditautkan
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
