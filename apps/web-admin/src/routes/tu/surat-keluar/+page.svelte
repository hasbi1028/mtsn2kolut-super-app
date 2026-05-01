<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import { toast } from '$lib/components/ui/sonner';
	import { onMount } from 'svelte';

	type Classification = {
		code: string;
		name: string;
	};

	type OutgoingLetter = {
		id: string;
		nomor_surat: string;
		classification_code: string;
		classification_name: string;
		tanggal_surat: string;
		tujuan: string;
		perihal: string;
		sifat: string;
		issued_by_name: string;
	};

	let loading = $state(true);
	let error = $state('');
	let letters = $state<OutgoingLetter[]>([]);
	let classifications = $state<Classification[]>([]);
	let search = $state('');

	// Create dialog
	let createOpen = $state(false);
	let createBusy = $state(false);
	let newKlasifikasi = $state('');
	let newTglSurat = $state('');
	let newTujuan = $state('');
	let newPerihal = $state('');
	let newSifat = $state('biasa');
	let newCatatan = $state('');
	let newManualNomor = $state('');
	let nomorPreview = $state('');

	// Edit dialog
	let editOpen = $state(false);
	let editBusy = $state(false);
	let editId = $state('');
	let editTglSurat = $state('');
	let editTujuan = $state('');
	let editPerihal = $state('');
	let editSifat = $state('biasa');
	let editCatatan = $state('');

	let deleteBusy = $state<Record<string, boolean>>({});

	async function loadLetters() {
		loading = true;
		error = '';
		try {
			const [letRes, klasRes] = await Promise.all([
				fetch(`/api/tu/surat/outgoing?search=${encodeURIComponent(search)}`),
				classifications.length === 0 ? fetch('/api/tu/surat/klasifikasi') : null
			]);
			if (!letRes.ok) throw new Error((await letRes.json()).error ?? `HTTP ${letRes.status}`);
			letters = (await letRes.json()) as OutgoingLetter[];
			if (klasRes && klasRes.ok) {
				classifications = (await klasRes.json()) as Classification[];
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'Gagal memuat surat keluar';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadLetters();
	});

	async function updatePreview() {
		if (!newKlasifikasi || !newTglSurat || newManualNomor.trim()) {
			nomorPreview = newManualNomor.trim() ? `Manual: ${newManualNomor}` : '';
			return;
		}
		try {
			const res = await fetch(
				`/api/tu/surat/outgoing/preview-number?classification_code=${encodeURIComponent(newKlasifikasi)}&tanggal=${encodeURIComponent(newTglSurat)}`
			);
			if (res.ok) {
				const json = await res.json() as { preview: string };
				nomorPreview = json.preview;
			}
		} catch {
			nomorPreview = '';
		}
	}

	async function createLetter() {
		if (!newKlasifikasi) { toast.error('Pilih kode klasifikasi'); return; }
		if (!newTglSurat) { toast.error('Tanggal surat wajib diisi'); return; }
		if (!newTujuan.trim()) { toast.error('Tujuan surat wajib diisi'); return; }
		if (!newPerihal.trim()) { toast.error('Perihal wajib diisi'); return; }
		createBusy = true;
		try {
			const res = await fetch('/api/tu/surat/outgoing', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					classification_code: newKlasifikasi,
					tanggal_surat: newTglSurat,
					tujuan: newTujuan,
					perihal: newPerihal,
					sifat: newSifat,
					catatan: newCatatan,
					manual_nomor: newManualNomor
				})
			});
			const json = await res.json();
			if (!res.ok) { toast.error(json.error ?? 'Gagal mencatat surat'); return; }
			createOpen = false;
			newKlasifikasi = ''; newTglSurat = ''; newTujuan = ''; newPerihal = ''; newSifat = 'biasa'; newCatatan = ''; newManualNomor = ''; nomorPreview = '';
			toast.success('Surat keluar berhasil dicatat');
			await loadLetters();
		} catch {
			toast.error('Terjadi kesalahan jaringan');
		} finally {
			createBusy = false;
		}
	}

	function openEdit(letter: OutgoingLetter) {
		editId = letter.id;
		editTglSurat = letter.tanggal_surat.split('T')[0];
		editTujuan = letter.tujuan;
		editPerihal = letter.perihal;
		editSifat = letter.sifat;
		editCatatan = '';
		editOpen = true;
	}

	async function saveEdit() {
		if (!editTujuan.trim()) { toast.error('Tujuan wajib diisi'); return; }
		if (!editPerihal.trim()) { toast.error('Perihal wajib diisi'); return; }
		editBusy = true;
		try {
			const res = await fetch(`/api/tu/surat/outgoing/${editId}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					tanggal_surat: editTglSurat,
					tujuan: editTujuan,
					perihal: editPerihal,
					sifat: editSifat,
					catatan: editCatatan
				})
			});
			const json = await res.json();
			if (!res.ok) { toast.error(json.error ?? 'Gagal menyimpan'); return; }
			editOpen = false;
			toast.success('Surat keluar diperbarui');
			await loadLetters();
		} catch {
			toast.error('Terjadi kesalahan jaringan');
		} finally {
			editBusy = false;
		}
	}

	async function deleteLetter(id: string) {
		if (!confirm('Hapus surat keluar ini?')) return;
		deleteBusy = { ...deleteBusy, [id]: true };
		try {
			const res = await fetch(`/api/tu/surat/outgoing/${id}`, { method: 'DELETE' });
			if (!res.ok && res.status !== 204) { toast.error((await res.json()).error ?? 'Gagal menghapus'); return; }
			toast.success('Surat dihapus');
			await loadLetters();
		} catch {
			toast.error('Terjadi kesalahan jaringan');
		} finally {
			deleteBusy = { ...deleteBusy, [id]: false };
		}
	}

	function formatDate(raw: string) {
		if (!raw) return '-';
		const d = new Date(raw);
		return d.toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' });
	}

	const sifatColors: Record<string, string> = {
		biasa: 'bg-gray-100 text-gray-700 hover:bg-gray-100',
		penting: 'bg-blue-100 text-blue-800 hover:bg-blue-100',
		segera: 'bg-orange-100 text-orange-800 hover:bg-orange-100',
		rahasia: 'bg-red-100 text-red-800 hover:bg-red-100'
	};
</script>

<div class="container mx-auto max-w-7xl space-y-6 p-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Surat Keluar</h1>
			<p class="text-sm text-gray-500">Pencatatan surat keluar dengan penomoran otomatis Kemenag</p>
		</div>
		<Button onclick={() => (createOpen = true)}>+ Catat Surat Keluar</Button>
	</div>

	<!-- Search -->
	<Card.Root>
		<Card.Content class="pt-4">
			<Input
				placeholder="Cari nomor, perihal, atau tujuan..."
				bind:value={search}
				class="w-72"
				oninput={() => loadLetters()}
			/>
		</Card.Content>
	</Card.Root>

	{#if error}
		<div class="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</div>
	{/if}

	<Card.Root>
		<Card.Content class="p-0">
			{#if loading}
				<div class="space-y-2 p-4">
					{#each [1, 2, 3] as _}
						<Skeleton class="h-12 w-full" />
					{/each}
				</div>
			{:else if letters.length === 0}
				<div class="p-8 text-center text-sm text-gray-500">
					{search ? 'Tidak ada surat yang sesuai pencarian.' : 'Belum ada surat keluar yang dicatat.'}
				</div>
			{:else}
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head class="w-12">No</Table.Head>
							<Table.Head>No. Surat</Table.Head>
							<Table.Head>Tujuan / Perihal</Table.Head>
							<Table.Head class="w-28">Tanggal</Table.Head>
							<Table.Head class="w-20">Sifat</Table.Head>
							<Table.Head class="w-32">Aksi</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each letters as letter, i}
							<Table.Row>
								<Table.Cell class="text-gray-500">{i + 1}</Table.Cell>
								<Table.Cell>
									<p class="font-mono text-xs font-medium text-gray-800">{letter.nomor_surat}</p>
									<p class="text-xs text-gray-400">{letter.classification_name || letter.classification_code}</p>
								</Table.Cell>
								<Table.Cell>
									<p class="text-sm font-medium text-gray-800">{letter.tujuan}</p>
									<p class="max-w-xs truncate text-xs text-gray-500">{letter.perihal}</p>
								</Table.Cell>
								<Table.Cell class="text-sm text-gray-600">{formatDate(letter.tanggal_surat)}</Table.Cell>
								<Table.Cell>
									<Badge class={sifatColors[letter.sifat] ?? 'bg-gray-100 text-gray-700 hover:bg-gray-100'}>
										{letter.sifat}
									</Badge>
								</Table.Cell>
								<Table.Cell>
									<div class="flex gap-1">
										<Button variant="outline" size="sm" onclick={() => openEdit(letter)}>Edit</Button>
										<LoadingButton
											variant="destructive"
											size="sm"
											loading={deleteBusy[letter.id] ?? false}
											onclick={() => deleteLetter(letter.id)}
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
</div>

<!-- Create Surat Keluar Dialog -->
<Dialog.Root bind:open={createOpen}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>Catat Surat Keluar</Dialog.Title>
			<Dialog.Description>Nomor surat akan digenerate otomatis berdasarkan klasifikasi dan tanggal.</Dialog.Description>
		</Dialog.Header>
		<div class="space-y-3 py-2">
			<div class="space-y-1">
				<label for="new-klasifikasi" class="text-sm font-medium text-gray-700">Kode Klasifikasi <span class="text-red-500">*</span></label>
				<select
					id="new-klasifikasi"
					bind:value={newKlasifikasi}
					onchange={updatePreview}
					class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-green-600"
				>
					<option value="">-- Pilih kode klasifikasi --</option>
					{#each classifications as c}
						<option value={c.code}>{c.code} — {c.name}</option>
					{/each}
				</select>
			</div>
			<div class="space-y-1">
				<label for="new-tgl-keluar" class="text-sm font-medium text-gray-700">Tanggal Surat <span class="text-red-500">*</span></label>
				<Input id="new-tgl-keluar" type="date" bind:value={newTglSurat} oninput={updatePreview} />
			</div>
			{#if nomorPreview}
				<div class="rounded-md bg-green-50 p-3">
					<p class="text-xs text-gray-500">Nomor surat yang akan digenerate:</p>
					<p class="font-mono text-sm font-medium text-green-800">{nomorPreview}</p>
				</div>
			{/if}
			<div class="space-y-1">
				<label for="new-manual-nomor" class="text-sm font-medium text-gray-700">Nomor Manual (opsional)</label>
				<Input id="new-manual-nomor" bind:value={newManualNomor} oninput={updatePreview} placeholder="Isi jika ingin pakai nomor kustom" />
				<p class="text-xs text-gray-400">Kosongkan untuk menggunakan nomor otomatis</p>
			</div>
			<div class="space-y-1">
				<label for="new-tujuan" class="text-sm font-medium text-gray-700">Tujuan <span class="text-red-500">*</span></label>
				<Input id="new-tujuan" bind:value={newTujuan} placeholder="Instansi atau nama tujuan" />
			</div>
			<div class="space-y-1">
				<label for="new-perihal-k" class="text-sm font-medium text-gray-700">Perihal <span class="text-red-500">*</span></label>
				<Input id="new-perihal-k" bind:value={newPerihal} placeholder="Perihal / isi singkat surat" />
			</div>
			<div class="space-y-1">
				<label for="new-sifat-k" class="text-sm font-medium text-gray-700">Sifat</label>
				<select
					id="new-sifat-k"
					bind:value={newSifat}
					class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-green-600"
				>
					<option value="biasa">Biasa</option>
					<option value="penting">Penting</option>
					<option value="segera">Segera</option>
					<option value="rahasia">Rahasia</option>
				</select>
			</div>
			<div class="space-y-1">
				<label for="new-catatan-k" class="text-sm font-medium text-gray-700">Catatan</label>
				<Textarea id="new-catatan-k" bind:value={newCatatan} placeholder="Catatan tambahan (opsional)" rows={2} />
			</div>
		</div>
		<Dialog.Footer>
			<Button variant="outline" onclick={() => (createOpen = false)}>Batal</Button>
			<LoadingButton loading={createBusy} onclick={createLetter}>Simpan</LoadingButton>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<!-- Edit Surat Keluar Dialog -->
<Dialog.Root bind:open={editOpen}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>Edit Surat Keluar</Dialog.Title>
			<Dialog.Description>Nomor surat tidak dapat diubah setelah dibuat.</Dialog.Description>
		</Dialog.Header>
		<div class="space-y-3 py-2">
			<div class="space-y-1">
				<label for="edit-tgl-surat" class="text-sm font-medium text-gray-700">Tanggal Surat</label>
				<Input id="edit-tgl-surat" type="date" bind:value={editTglSurat} />
			</div>
			<div class="space-y-1">
				<label for="edit-tujuan" class="text-sm font-medium text-gray-700">Tujuan <span class="text-red-500">*</span></label>
				<Input id="edit-tujuan" bind:value={editTujuan} />
			</div>
			<div class="space-y-1">
				<label for="edit-perihal" class="text-sm font-medium text-gray-700">Perihal <span class="text-red-500">*</span></label>
				<Input id="edit-perihal" bind:value={editPerihal} />
			</div>
			<div class="space-y-1">
				<label for="edit-sifat" class="text-sm font-medium text-gray-700">Sifat</label>
				<select
					id="edit-sifat"
					bind:value={editSifat}
					class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-green-600"
				>
					<option value="biasa">Biasa</option>
					<option value="penting">Penting</option>
					<option value="segera">Segera</option>
					<option value="rahasia">Rahasia</option>
				</select>
			</div>
			<div class="space-y-1">
				<label for="edit-catatan" class="text-sm font-medium text-gray-700">Catatan</label>
				<Textarea id="edit-catatan" bind:value={editCatatan} rows={2} />
			</div>
		</div>
		<Dialog.Footer>
			<Button variant="outline" onclick={() => (editOpen = false)}>Batal</Button>
			<LoadingButton loading={editBusy} onclick={saveEdit}>Simpan</LoadingButton>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
