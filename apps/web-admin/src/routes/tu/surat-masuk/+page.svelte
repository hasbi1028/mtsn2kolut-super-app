<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { toast } from '$lib/components/ui/sonner';
	import { onMount } from 'svelte';
	import { confirmAction } from '$lib/confirm-dialog';
	import { readClientApiData, readClientJson } from '$lib/client/api';

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

	let lettersPromise = $state<Promise<IncomingLetter[]> | null>(null);
	let lettersRequestId = 0;
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
	let disposisiPromise = $state<Promise<DispositionRow[]> | null>(null);
	let disposisiRequestId = 0;
	let newDisposisiAssignee = $state('');
	let newDisposisiInstruksi = $state('');
	let disposisiBusy = $state(false);

	async function fetchLetters(): Promise<IncomingLetter[]> {
		const params = new URLSearchParams();
		if (search) params.set('search', search);
		if (filterStatus) params.set('status', filterStatus);
		const res = await fetch(`/api/tu/surat/incoming?${params}`);
		return readClientApiData<IncomingLetter[]>(res, 'Gagal memuat surat masuk');
	}

	function loadLetters() {
		const requestId = ++lettersRequestId;
		lettersPromise = fetchLetters()
			.then((rows) => (requestId === lettersRequestId ? (rows ?? []) : []))
			.catch((error: unknown) => {
				if (requestId === lettersRequestId) throw error;
				return [];
			});
	}

	async function refreshLetters() {
		const requestId = ++lettersRequestId;
		const rows = await fetchLetters();
		if (requestId === lettersRequestId) {
			lettersPromise = Promise.resolve(rows ?? []);
		}
	}

	function retryLetters(reset?: () => void) {
		reset?.();
		loadLetters();
	}

	function lettersErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat surat masuk';
	}

	function disposisiErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat disposisi surat. Periksa koneksi lalu coba lagi.';
	}

	function handleLettersRenderError(error: unknown) {
		console.error('TU incoming letters render failed', error);
	}

	function handleDisposisiRenderError(error: unknown) {
		console.error('TU incoming letter dispositions render failed', error);
	}

	function mutationErrorMessage(error: unknown, fallback: string) {
		if (error instanceof Error && error.message.trim() && !error.message.toLowerCase().includes('fetch')) return error.message;
		return fallback;
	}

	async function refreshLettersAfterMutation() {
		try {
			await refreshLetters();
		} catch (error) {
			toast.error(lettersErrorMessage(error));
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
			await readClientJson<unknown>(res);
			createOpen = false;
			newNomor = ''; newTglSurat = ''; newTglTerima = ''; newAsal = ''; newPerihal = ''; newSifat = 'biasa'; newCatatan = '';
			toast.success('Surat masuk berhasil dicatat');
			await refreshLettersAfterMutation();
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Terjadi kesalahan jaringan'));
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
			await readClientJson<unknown>(res);
			await refreshLettersAfterMutation();
			toast.success('Status diperbarui');
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Terjadi kesalahan jaringan'));
		} finally {
			statusBusy = { ...statusBusy, [letter.id]: false };
		}
	}

	async function deleteLetter(id: string) {
		if (!(await confirmAction({
			title: 'Hapus Surat Masuk',
			message: 'Hapus surat masuk ini? Semua disposisi akan ikut terhapus.',
			confirmLabel: 'Hapus Surat',
			tone: 'danger'
		}))) return;
		deleteBusy = { ...deleteBusy, [id]: true };
		try {
			const res = await fetch(`/api/tu/surat/incoming/${id}`, { method: 'DELETE' });
			await readClientJson<unknown>(res);
			toast.success('Surat dihapus');
			await refreshLettersAfterMutation();
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Terjadi kesalahan jaringan'));
		} finally {
			deleteBusy = { ...deleteBusy, [id]: false };
		}
	}

	async function fetchDisposisi(letterId: string): Promise<DispositionRow[]> {
		const res = await fetch(`/api/tu/surat/disposisi?incoming_letter_id=${letterId}`);
		return readClientApiData<DispositionRow[]>(res, 'Gagal memuat disposisi');
	}

	function setDisposisiPromise(letter: IncomingLetter) {
		const requestId = ++disposisiRequestId;
		disposisiList = [];
		disposisiPromise = fetchDisposisi(letter.id)
			.then((rows) => {
				if (requestId === disposisiRequestId && disposisiLetter?.id === letter.id) {
					disposisiList = rows ?? [];
					return disposisiList;
				}
				return disposisiList;
			})
			.catch((error: unknown) => {
				if (requestId === disposisiRequestId && disposisiLetter?.id === letter.id) throw error;
				return disposisiList;
			});
	}

	function openDisposisi(letter: IncomingLetter) {
		disposisiLetter = letter;
		disposisiOpen = true;
		setDisposisiPromise(letter);
	}

	function retryDisposisi(reset?: () => void) {
		if (!disposisiLetter) return;
		reset?.();
		setDisposisiPromise(disposisiLetter);
	}

	async function refreshDisposisiList() {
		if (!disposisiLetter) return;
		const requestId = ++disposisiRequestId;
		const letterId = disposisiLetter.id;
		const rows = await fetchDisposisi(disposisiLetter.id);
		if (requestId === disposisiRequestId && disposisiLetter?.id === letterId) {
			disposisiList = rows ?? [];
			disposisiPromise = Promise.resolve(disposisiList);
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
			await readClientJson<unknown>(res);
			newDisposisiAssignee = '';
			newDisposisiInstruksi = '';
			toast.success('Disposisi berhasil dibuat');
			await refreshDisposisiList();
			await refreshLettersAfterMutation();
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Terjadi kesalahan jaringan'));
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
		biasa: 'bg-muted text-foreground hover:bg-muted',
		penting: 'bg-accent text-accent-foreground hover:bg-accent',
		segera: 'bg-warning/15 text-warning hover:bg-warning/15',
		rahasia: 'bg-destructive/15 text-destructive hover:bg-destructive/15'
	};

	const disposisiStatusColors: Record<string, string> = {
		terkirim: 'bg-accent text-accent-foreground hover:bg-accent',
		dibaca: 'bg-accent text-accent-foreground hover:bg-accent',
		ditindaklanjuti: 'bg-warning/15 text-warning hover:bg-warning/15',
		selesai: 'bg-success/15 text-success hover:bg-success/15'
	};
</script>

<div class="container mx-auto max-w-7xl space-y-6 p-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-foreground">Surat Masuk</h1>
			<p class="text-sm text-muted-foreground">Pencatatan dan pengelolaan surat masuk</p>
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
					class="rounded-md border border-border bg-card px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
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

	<Card.Root>
		<Card.Content class="p-0">
			<AsyncContent promise={lettersPromise} onerror={handleLettersRenderError}>
				{#snippet pending()}
					<div class="space-y-2 p-4">
						{#each [1, 2, 3] as row (row)}
							<Skeleton class="h-12 w-full" />
						{/each}
					</div>
				{/snippet}

				{#snippet failed(error, reset)}
					<div class="p-4">
						<RecoveryPanel
							compact
							title="Surat Masuk Belum Tersaji"
							message={lettersErrorMessage(error)}
							onRetry={() => retryLetters(reset)}
						/>
					</div>
				{/snippet}

				{#snippet children(value)}
					{@const currentLetters = value as IncomingLetter[]}
					{#if currentLetters.length === 0}
						<div class="p-8 text-center text-sm text-muted-foreground">
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
								{#each currentLetters as letter, i (letter.id)}
									<Table.Row>
										<Table.Cell class="text-muted-foreground">{i + 1}</Table.Cell>
										<Table.Cell class="font-mono text-xs font-medium text-foreground">{letter.nomor_agenda}</Table.Cell>
										<Table.Cell class="max-w-[180px] truncate text-xs text-muted-foreground">{letter.nomor_surat}</Table.Cell>
										<Table.Cell>
											<p class="text-sm font-medium text-foreground">{letter.asal}</p>
											<p class="max-w-xs truncate text-xs text-muted-foreground">{letter.perihal}</p>
										</Table.Cell>
										<Table.Cell class="text-sm text-muted-foreground">{formatDate(letter.tanggal_terima)}</Table.Cell>
										<Table.Cell>
											<Badge class={sifatColors[letter.sifat] ?? 'bg-muted text-foreground hover:bg-muted'}>
												{letter.sifat}
											</Badge>
										</Table.Cell>
										<Table.Cell>
											<select
												value={letter.status}
												onchange={(e) => updateStatus(letter, (e.target as HTMLSelectElement).value)}
												disabled={statusBusy[letter.id]}
												class="rounded border border-border px-2 py-1 text-xs focus:outline-none focus:ring-1 focus:ring-ring"
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
				{/snippet}
			</AsyncContent>
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
				<label for="new-nomor" class="text-sm font-medium text-foreground">No. Surat <span class="text-destructive">*</span></label>
				<Input id="new-nomor" bind:value={newNomor} placeholder="Nomor surat dari pengirim" />
			</div>
			<div class="grid grid-cols-2 gap-3">
				<div class="space-y-1">
					<label for="new-tgl-surat" class="text-sm font-medium text-foreground">Tanggal Surat <span class="text-destructive">*</span></label>
					<Input id="new-tgl-surat" type="date" bind:value={newTglSurat} />
				</div>
				<div class="space-y-1">
					<label for="new-tgl-terima" class="text-sm font-medium text-foreground">Tanggal Terima <span class="text-destructive">*</span></label>
					<Input id="new-tgl-terima" type="date" bind:value={newTglTerima} />
				</div>
			</div>
			<div class="space-y-1">
				<label for="new-asal" class="text-sm font-medium text-foreground">Asal / Pengirim <span class="text-destructive">*</span></label>
				<Input id="new-asal" bind:value={newAsal} placeholder="Instansi atau nama pengirim" />
			</div>
			<div class="space-y-1">
				<label for="new-perihal" class="text-sm font-medium text-foreground">Perihal <span class="text-destructive">*</span></label>
				<Input id="new-perihal" bind:value={newPerihal} placeholder="Perihal / isi singkat surat" />
			</div>
			<div class="space-y-1">
				<label for="new-sifat" class="text-sm font-medium text-foreground">Sifat</label>
				<select
					id="new-sifat"
					bind:value={newSifat}
					class="w-full rounded-md border border-border bg-card px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
				>
					<option value="biasa">Biasa</option>
					<option value="penting">Penting</option>
					<option value="segera">Segera</option>
					<option value="rahasia">Rahasia</option>
				</select>
			</div>
			<div class="space-y-1">
				<label for="new-catatan" class="text-sm font-medium text-foreground">Catatan</label>
				<Textarea id="new-catatan" bind:value={newCatatan} placeholder="Catatan tambahan (opsional)" rows={2} />
			</div>
		</div>
		<Dialog.Footer>
			<Button variant="outline" onclick={() => (createOpen = false)}>Batal</Button>
			<LoadingButton loading={createBusy} onclick={() => void createLetter()}>Simpan</LoadingButton>
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
			<AsyncContent promise={disposisiPromise} onerror={handleDisposisiRenderError}>
				{#snippet pending()}
					<div class="space-y-2">
						<Skeleton class="h-16 w-full" />
						<Skeleton class="h-16 w-full" />
					</div>
				{/snippet}
				{#snippet failed(error, reset)}
					<RecoveryPanel compact title="Disposisi Belum Tersaji" message={disposisiErrorMessage(error)} onRetry={() => retryDisposisi(reset)} />
				{/snippet}
				{#snippet children(rows)}
					{@const currentDisposisi = rows as DispositionRow[]}
					{#if currentDisposisi.length > 0}
						<div class="space-y-2">
							<p class="text-sm font-medium text-foreground">Disposisi sebelumnya</p>
							{#each currentDisposisi as d (d.id)}
								<div class="rounded-md border border-border bg-muted/50 p-3 text-sm">
									<div class="flex items-center justify-between">
										<span class="font-medium text-foreground">{d.assignee_name || '–'}</span>
										<Badge class={disposisiStatusColors[d.status] ?? ''} >{d.status}</Badge>
									</div>
									{#if d.instruksi}
										<p class="mt-1 text-muted-foreground">{d.instruksi}</p>
									{/if}
									{#if d.catatan_tindak_lanjut}
										<p class="mt-1 text-muted-foreground italic">Tindak lanjut: {d.catatan_tindak_lanjut}</p>
									{/if}
								</div>
							{/each}
						</div>
					{/if}
				{/snippet}
			</AsyncContent>

			<!-- New disposition form -->
			<div class="space-y-2 border-t border-border pt-3">
				<p class="text-sm font-medium text-foreground">Buat Disposisi Baru</p>
				<div class="space-y-1">
					<label for="disp-assignee" class="text-xs text-muted-foreground">ID Pegawai Penerima <span class="text-destructive">*</span></label>
					<Input id="disp-assignee" bind:value={newDisposisiAssignee} placeholder="ID pegawai" />
				</div>
				<div class="space-y-1">
					<label for="disp-instruksi" class="text-xs text-muted-foreground">Instruksi</label>
					<Textarea id="disp-instruksi" bind:value={newDisposisiInstruksi} placeholder="Instruksi disposisi (opsional)" rows={2} />
				</div>
			</div>
		</div>
		<Dialog.Footer>
			<Button variant="outline" onclick={() => (disposisiOpen = false)}>Tutup</Button>
			<LoadingButton loading={disposisiBusy} onclick={() => void createDisposisi()}>Disposisi</LoadingButton>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
