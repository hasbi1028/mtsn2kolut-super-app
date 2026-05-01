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

	type IncomingLetter = {
		id: string;
		nomor_surat: string;
		nomor_agenda: string;
		tanggal_surat: string;
		tanggal_terima: string;
		asal: string;
		perihal: string;
		sifat: string;
		status: string;
		received_by_name: string;
		disposisi_count: number;
	};

	type DispositionRow = {
		id: string;
		incoming_letter_id: string;
		assignee_name: string;
		instruksi: string;
		catatan_tindak_lanjut: string;
		status: string;
		disposed_by_name: string;
		disposed_at: string;
		nomor_agenda: string;
		letter_perihal: string;
	};

	let loading = $state(true);
	let error = $state('');
	let letters = $state<IncomingLetter[]>([]);
	let search = $state('');
	let filterStatus = $state('');

	// Create dialog
	let createOpen = $state(false);
	let createBusy = $state(false);
	let newNomor = $state('');
	let newTglSurat = $state('');
	let newTglTerima = $state('');
	let newAsal = $state('');
	let newPerihal = $state('');
	let newSifat = $state('biasa');
	let newCatatan = $state('');

	// Status update
	let statusBusy = $state<Record<string, boolean>>({});

	// Delete
	let deleteBusy = $state<Record<string, boolean>>({});

	// Disposisi dialog
	let disposisiOpen = $state(false);
	let disposisiLetter = $state<IncomingLetter | null>(null);
	let disposisiList = $state<DispositionRow[]>([]);
	let disposisiLoading = $state(false);
	let newDisposisiAssignee = $state('');
	let newDisposisiInstruksi = $state('');
	let disposisiBusy = $state(false);

	async function loadLetters() {
		loading = true;
		error = '';
		try {
			const params = new URLSearchParams();
			if (search) params.set('search', search);
			if (filterStatus) params.set('status', filterStatus);
			const res = await fetch(`/api/tu/surat/incoming?${params}`);
			if (!res.ok) throw new Error((await res.json()).error ?? `HTTP ${res.status}`);
			letters = (await res.json()) as IncomingLetter[];
		} catch (e) {
			error = e instanceof Error ? e.message : 'Gagal memuat surat masuk';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadLetters();
	});

	async function createLetter() {
		if (!newNomor.trim()) { toast.error('Nomor surat wajib diisi'); return; }
		if (!newTglSurat) { toast.error('Tanggal surat wajib diisi'); return; }
		if (!newTglTerima) { toast.error('Tanggal terima wajib diisi'); return; }
		if (!newAsal.trim()) { toast.error('Asal surat wajib diisi'); return; }
		if (!newPerihal.trim()) { toast.error('Perihal wajib diisi'); return; }
		createBusy = true;
		try {
			const res = await fetch('/api/tu/surat/incoming', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					nomor_surat: newNomor,
					tanggal_surat: newTglSurat,
					tanggal_terima: newTglTerima,
					asal: newAsal,
					perihal: newPerihal,
					sifat: newSifat,
					catatan: newCatatan
				})
			});
			const json = await res.json();
			if (!res.ok) { toast.error(json.error ?? 'Gagal mencatat surat'); return; }
			createOpen = false;
			newNomor = ''; newTglSurat = ''; newTglTerima = ''; newAsal = ''; newPerihal = ''; newSifat = 'biasa'; newCatatan = '';
			toast.success('Surat masuk berhasil dicatat');
			await loadLetters();
		} catch {
			toast.error('Terjadi kesalahan jaringan');
		} finally {
			createBusy = false;
		}
	}

	async function updateStatus(letter: IncomingLetter, newStatus: string) {
		if (letter.status === newStatus) return;
		statusBusy = { ...statusBusy, [letter.id]: true };
		try {
			const res = await fetch(`/api/tu/surat/incoming/${letter.id}/status`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ status: newStatus })
			});
			if (!res.ok) { toast.error((await res.json()).error ?? 'Gagal mengubah status'); return; }
			await loadLetters();
			toast.success('Status diperbarui');
		} catch {
			toast.error('Terjadi kesalahan jaringan');
		} finally {
			statusBusy = { ...statusBusy, [letter.id]: false };
		}
	}

	async function deleteLetter(id: string) {
		if (!confirm('Hapus surat masuk ini? Semua disposisi akan ikut terhapus.')) return;
		deleteBusy = { ...deleteBusy, [id]: true };
		try {
			const res = await fetch(`/api/tu/surat/incoming/${id}`, { method: 'DELETE' });
			if (!res.ok && res.status !== 204) { toast.error((await res.json()).error ?? 'Gagal menghapus'); return; }
			toast.success('Surat dihapus');
			await loadLetters();
		} catch {
			toast.error('Terjadi kesalahan jaringan');
		} finally {
			deleteBusy = { ...deleteBusy, [id]: false };
		}
	}

	async function openDisposisi(letter: IncomingLetter) {
		disposisiLetter = letter;
		disposisiOpen = true;
		disposisiLoading = true;
		try {
			const res = await fetch(`/api/tu/surat/disposisi?incoming_letter_id=${letter.id}`);
			if (!res.ok) throw new Error((await res.json()).error ?? 'Gagal');
			disposisiList = (await res.json()) as DispositionRow[];
		} catch {
			disposisiList = [];
		} finally {
			disposisiLoading = false;
		}
	}

	async function createDisposisi() {
		if (!disposisiLetter) return;
		if (!newDisposisiAssignee.trim()) { toast.error('Nama/ID pegawai penerima wajib diisi'); return; }
		disposisiBusy = true;
		try {
			const res = await fetch('/api/tu/surat/disposisi', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					incoming_letter_id: disposisiLetter.id,
					assignee_employee_id: newDisposisiAssignee,
					instruksi: newDisposisiInstruksi
				})
			});
			const json = await res.json();
			if (!res.ok) { toast.error(json.error ?? 'Gagal mendisposisi'); return; }
			newDisposisiAssignee = '';
			newDisposisiInstruksi = '';
			toast.success('Disposisi berhasil dibuat');
			await openDisposisi(disposisiLetter);
			await loadLetters();
		} catch {
			toast.error('Terjadi kesalahan jaringan');
		} finally {
			disposisiBusy = false;
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

	const statusColors: Record<string, string> = {
		baru: 'bg-sky-100 text-sky-800 hover:bg-sky-100',
		didisposisi: 'bg-yellow-100 text-yellow-800 hover:bg-yellow-100',
		selesai: 'bg-green-100 text-green-800 hover:bg-green-100',
		arsip: 'bg-gray-100 text-gray-600 hover:bg-gray-100'
	};

	const disposisiStatusColors: Record<string, string> = {
		terkirim: 'bg-sky-100 text-sky-800 hover:bg-sky-100',
		dibaca: 'bg-purple-100 text-purple-800 hover:bg-purple-100',
		ditindaklanjuti: 'bg-yellow-100 text-yellow-800 hover:bg-yellow-100',
		selesai: 'bg-green-100 text-green-800 hover:bg-green-100'
	};
</script>

<div class="container mx-auto max-w-7xl space-y-6 p-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Surat Masuk</h1>
			<p class="text-sm text-gray-500">Pencatatan dan pengelolaan surat masuk</p>
		</div>
		<Button onclick={() => (createOpen = true)}>+ Catat Surat Masuk</Button>
	</div>

	<!-- Filters -->
	<Card.Root>
		<Card.Content class="pt-4">
			<div class="flex flex-wrap items-center gap-3">
				<Input
					placeholder="Cari nomor, perihal, atau asal..."
					bind:value={search}
					class="w-72"
					oninput={() => loadLetters()}
				/>
				<select
					bind:value={filterStatus}
					onchange={() => loadLetters()}
					class="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-green-600"
				>
					<option value="">Semua Status</option>
					<option value="baru">Baru</option>
					<option value="didisposisi">Didisposisi</option>
					<option value="selesai">Selesai</option>
					<option value="arsip">Arsip</option>
				</select>
			</div>
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
					{search || filterStatus ? 'Tidak ada surat yang sesuai filter.' : 'Belum ada surat masuk yang dicatat.'}
				</div>
			{:else}
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head class="w-12">No</Table.Head>
							<Table.Head class="w-32">No. Agenda</Table.Head>
							<Table.Head>No. Surat</Table.Head>
							<Table.Head>Asal / Perihal</Table.Head>
							<Table.Head class="w-28">Tgl Terima</Table.Head>
							<Table.Head class="w-20">Sifat</Table.Head>
							<Table.Head class="w-28">Status</Table.Head>
							<Table.Head class="w-40">Aksi</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each letters as letter, i}
							<Table.Row>
								<Table.Cell class="text-gray-500">{i + 1}</Table.Cell>
								<Table.Cell class="font-mono text-xs font-medium text-gray-700">{letter.nomor_agenda}</Table.Cell>
								<Table.Cell class="max-w-[180px] truncate text-xs text-gray-600">{letter.nomor_surat}</Table.Cell>
								<Table.Cell>
									<p class="text-sm font-medium text-gray-800">{letter.asal}</p>
									<p class="max-w-xs truncate text-xs text-gray-500">{letter.perihal}</p>
								</Table.Cell>
								<Table.Cell class="text-sm text-gray-600">{formatDate(letter.tanggal_terima)}</Table.Cell>
								<Table.Cell>
									<Badge class={sifatColors[letter.sifat] ?? 'bg-gray-100 text-gray-700 hover:bg-gray-100'}>
										{letter.sifat}
									</Badge>
								</Table.Cell>
								<Table.Cell>
									<select
										value={letter.status}
										onchange={(e) => updateStatus(letter, (e.target as HTMLSelectElement).value)}
										disabled={statusBusy[letter.id]}
										class="rounded border border-gray-200 px-2 py-1 text-xs focus:outline-none focus:ring-1 focus:ring-green-600"
									>
										<option value="baru">Baru</option>
										<option value="didisposisi">Didisposisi</option>
										<option value="selesai">Selesai</option>
										<option value="arsip">Arsip</option>
									</select>
								</Table.Cell>
								<Table.Cell>
									<div class="flex gap-1">
										<Button
											variant="outline"
											size="sm"
											onclick={() => openDisposisi(letter)}
										>
											Disposisi {#if letter.disposisi_count > 0}({letter.disposisi_count}){/if}
										</Button>
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

<!-- Create Surat Masuk Dialog -->
<Dialog.Root bind:open={createOpen}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>Catat Surat Masuk</Dialog.Title>
			<Dialog.Description>Isi detail surat masuk yang diterima. Nomor agenda akan digenerate otomatis.</Dialog.Description>
		</Dialog.Header>
		<div class="space-y-3 py-2">
			<div class="space-y-1">
				<label for="new-nomor" class="text-sm font-medium text-gray-700">No. Surat <span class="text-red-500">*</span></label>
				<Input id="new-nomor" bind:value={newNomor} placeholder="Nomor surat dari pengirim" />
			</div>
			<div class="grid grid-cols-2 gap-3">
				<div class="space-y-1">
					<label for="new-tgl-surat" class="text-sm font-medium text-gray-700">Tanggal Surat <span class="text-red-500">*</span></label>
					<Input id="new-tgl-surat" type="date" bind:value={newTglSurat} />
				</div>
				<div class="space-y-1">
					<label for="new-tgl-terima" class="text-sm font-medium text-gray-700">Tanggal Terima <span class="text-red-500">*</span></label>
					<Input id="new-tgl-terima" type="date" bind:value={newTglTerima} />
				</div>
			</div>
			<div class="space-y-1">
				<label for="new-asal" class="text-sm font-medium text-gray-700">Asal / Pengirim <span class="text-red-500">*</span></label>
				<Input id="new-asal" bind:value={newAsal} placeholder="Instansi atau nama pengirim" />
			</div>
			<div class="space-y-1">
				<label for="new-perihal" class="text-sm font-medium text-gray-700">Perihal <span class="text-red-500">*</span></label>
				<Input id="new-perihal" bind:value={newPerihal} placeholder="Perihal / isi singkat surat" />
			</div>
			<div class="space-y-1">
				<label for="new-sifat" class="text-sm font-medium text-gray-700">Sifat</label>
				<select
					id="new-sifat"
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
				<label for="new-catatan" class="text-sm font-medium text-gray-700">Catatan</label>
				<Textarea id="new-catatan" bind:value={newCatatan} placeholder="Catatan tambahan (opsional)" rows={2} />
			</div>
		</div>
		<Dialog.Footer>
			<Button variant="outline" onclick={() => (createOpen = false)}>Batal</Button>
			<LoadingButton loading={createBusy} onclick={createLetter}>Simpan</LoadingButton>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<!-- Disposisi Dialog -->
<Dialog.Root bind:open={disposisiOpen}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>Disposisi Surat</Dialog.Title>
			<Dialog.Description>
				{#if disposisiLetter}
					<span class="font-mono">{disposisiLetter.nomor_agenda}</span> — {disposisiLetter.perihal}
				{/if}
			</Dialog.Description>
		</Dialog.Header>
		<div class="space-y-4 py-2">
			<!-- Existing dispositions -->
			{#if disposisiLoading}
				<Skeleton class="h-16 w-full" />
			{:else if disposisiList.length > 0}
				<div class="space-y-2">
					<p class="text-sm font-medium text-gray-700">Disposisi sebelumnya</p>
					{#each disposisiList as d}
						<div class="rounded-md border border-gray-100 bg-gray-50 p-3 text-sm">
							<div class="flex items-center justify-between">
								<span class="font-medium text-gray-800">{d.assignee_name || '–'}</span>
								<Badge class={disposisiStatusColors[d.status] ?? ''} >{d.status}</Badge>
							</div>
							{#if d.instruksi}
								<p class="mt-1 text-gray-600">{d.instruksi}</p>
							{/if}
							{#if d.catatan_tindak_lanjut}
								<p class="mt-1 text-gray-500 italic">Tindak lanjut: {d.catatan_tindak_lanjut}</p>
							{/if}
						</div>
					{/each}
				</div>
			{/if}

			<!-- New disposition form -->
			<div class="space-y-2 border-t border-gray-100 pt-3">
				<p class="text-sm font-medium text-gray-700">Buat Disposisi Baru</p>
				<div class="space-y-1">
					<label for="disp-assignee" class="text-xs text-gray-600">ID Pegawai Penerima <span class="text-red-500">*</span></label>
					<Input id="disp-assignee" bind:value={newDisposisiAssignee} placeholder="UUID pegawai" />
				</div>
				<div class="space-y-1">
					<label for="disp-instruksi" class="text-xs text-gray-600">Instruksi</label>
					<Textarea id="disp-instruksi" bind:value={newDisposisiInstruksi} placeholder="Instruksi disposisi (opsional)" rows={2} />
				</div>
			</div>
		</div>
		<Dialog.Footer>
			<Button variant="outline" onclick={() => (disposisiOpen = false)}>Tutup</Button>
			<LoadingButton loading={disposisiBusy} onclick={createDisposisi}>Disposisi</LoadingButton>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
