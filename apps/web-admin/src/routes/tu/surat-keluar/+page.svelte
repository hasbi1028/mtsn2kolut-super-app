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

	type OutgoingLettersOverview = {
		letters: OutgoingLetter[];
		classifications: Classification[];
	};

	let classifications = $state<Classification[]>([]);
	let lettersPromise = $state<Promise<OutgoingLettersOverview> | null>(null);
	let lettersRequestId = 0;
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

	function applyOverview(overview: OutgoingLettersOverview) {
		classifications = overview.classifications ?? [];
	}

	async function fetchOverview(): Promise<OutgoingLettersOverview> {
		const [lettersRes, classificationsRes] = await Promise.all([
			fetch(`/api/tu/surat/outgoing?search=${encodeURIComponent(search)}`),
			classifications.length === 0 ? fetch('/api/tu/surat/klasifikasi') : Promise.resolve(null)
		]);
		const [nextLetters, nextClassifications] = await Promise.all([
			readClientApiData<OutgoingLetter[]>(lettersRes, 'Gagal memuat surat keluar'),
			classificationsRes
				? readClientApiData<Classification[]>(classificationsRes, 'Gagal memuat kode klasifikasi')
				: Promise.resolve(classifications)
		]);
		return {
			letters: nextLetters ?? [],
			classifications: nextClassifications ?? []
		};
	}

	function loadLetters() {
		const requestId = ++lettersRequestId;
		const emptyOverview: OutgoingLettersOverview = { letters: [], classifications };
		applyOverview(emptyOverview);
		lettersPromise = fetchOverview()
			.then((overview) => {
				if (requestId === lettersRequestId) {
					applyOverview(overview);
					return overview;
				}
				return { letters: [], classifications };
			})
			.catch((error: unknown) => {
				if (requestId === lettersRequestId) throw error;
				return { letters: [], classifications };
			});
	}

	async function refreshLetters() {
		const requestId = ++lettersRequestId;
		const overview = await fetchOverview();
		if (requestId === lettersRequestId) {
			applyOverview(overview);
			lettersPromise = Promise.resolve(overview);
		}
	}

	function retryLetters(reset?: () => void) {
		reset?.();
		loadLetters();
	}

	function lettersErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat surat keluar';
	}

	function handleLettersRenderError(error: unknown) {
		console.error('TU outgoing letters render failed', error);
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

	async function updatePreview() {
		if (!newKlasifikasi || !newTglSurat || newManualNomor.trim()) {
			nomorPreview = newManualNomor.trim() ? `Manual: ${newManualNomor}` : '';
			return;
		}
		try {
			const res = await fetch(
				`/api/tu/surat/outgoing/preview-number?classification_code=${encodeURIComponent(newKlasifikasi)}&tanggal=${encodeURIComponent(newTglSurat)}`
			);
			const payload = await readClientApiData<{ preview?: string }>(res, 'Gagal memuat pratinjau nomor');
			nomorPreview = typeof payload.preview === 'string' ? payload.preview : '';
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
			await readClientJson<unknown>(res);
			createOpen = false;
			newKlasifikasi = ''; newTglSurat = ''; newTujuan = ''; newPerihal = ''; newSifat = 'biasa'; newCatatan = ''; newManualNomor = ''; nomorPreview = '';
			toast.success('Surat keluar berhasil dicatat');
			await refreshLettersAfterMutation();
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Terjadi kesalahan jaringan'));
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
			await readClientJson<unknown>(res);
			editOpen = false;
			toast.success('Surat keluar diperbarui');
			await refreshLettersAfterMutation();
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Terjadi kesalahan jaringan'));
		} finally {
			editBusy = false;
		}
	}

	async function deleteLetter(id: string) {
		if (!(await confirmAction({
			title: 'Hapus Surat Keluar',
			message: 'Hapus surat keluar ini?',
			confirmLabel: 'Hapus Surat',
			tone: 'danger'
		}))) return;
		deleteBusy = { ...deleteBusy, [id]: true };
		try {
			const res = await fetch(`/api/tu/surat/outgoing/${id}`, { method: 'DELETE' });
			await readClientJson<unknown>(res);
			toast.success('Surat dihapus');
			await refreshLettersAfterMutation();
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Terjadi kesalahan jaringan'));
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
		biasa: 'bg-muted text-foreground hover:bg-muted',
		penting: 'bg-accent text-accent-foreground hover:bg-accent',
		segera: 'bg-warning/15 text-warning hover:bg-warning/15',
		rahasia: 'bg-destructive/15 text-destructive hover:bg-destructive/15'
	};
</script>

<div class="container mx-auto max-w-7xl space-y-6 p-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-foreground">Surat Keluar</h1>
			<p class="text-sm text-muted-foreground">Pencatatan surat keluar dengan penomoran otomatis Kemenag</p>
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
							title="Surat Keluar Belum Tersaji"
							message={lettersErrorMessage(error)}
							onRetry={() => retryLetters(reset)}
						/>
					</div>
				{/snippet}

				{#snippet children(value)}
					{@const currentLetters = (value as OutgoingLettersOverview).letters}
					{#if currentLetters.length === 0}
						<div class="p-8 text-center text-sm text-muted-foreground">
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
								{#each currentLetters as letter, i (letter.id)}
									<Table.Row>
										<Table.Cell class="text-muted-foreground">{i + 1}</Table.Cell>
										<Table.Cell>
											<p class="font-mono text-xs font-medium text-foreground">{letter.nomor_surat}</p>
											<p class="text-xs text-muted-foreground">{letter.classification_name || letter.classification_code}</p>
										</Table.Cell>
										<Table.Cell>
											<p class="text-sm font-medium text-foreground">{letter.tujuan}</p>
											<p class="max-w-xs truncate text-xs text-muted-foreground">{letter.perihal}</p>
										</Table.Cell>
										<Table.Cell class="text-sm text-muted-foreground">{formatDate(letter.tanggal_surat)}</Table.Cell>
										<Table.Cell>
											<Badge class={sifatColors[letter.sifat] ?? 'bg-muted text-foreground hover:bg-muted'}>
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
				{/snippet}
			</AsyncContent>
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
				<label for="new-klasifikasi" class="text-sm font-medium text-foreground">Kode Klasifikasi <span class="text-destructive">*</span></label>
				<select
					id="new-klasifikasi"
					bind:value={newKlasifikasi}
					onchange={updatePreview}
					class="w-full rounded-md border border-border bg-card px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
				>
					<option value="">-- Pilih kode klasifikasi --</option>
					{#each classifications as c (c.code)}
						<option value={c.code}>{c.code} — {c.name}</option>
					{/each}
				</select>
			</div>
			<div class="space-y-1">
				<label for="new-tgl-keluar" class="text-sm font-medium text-foreground">Tanggal Surat <span class="text-destructive">*</span></label>
				<Input id="new-tgl-keluar" type="date" bind:value={newTglSurat} oninput={updatePreview} />
			</div>
			{#if nomorPreview}
				<div class="rounded-md bg-success/10 p-3">
					<p class="text-xs text-muted-foreground">Nomor surat yang akan digenerate:</p>
					<p class="font-mono text-sm font-medium text-success">{nomorPreview}</p>
				</div>
			{/if}
			<div class="space-y-1">
				<label for="new-manual-nomor" class="text-sm font-medium text-foreground">Nomor Manual (opsional)</label>
				<Input id="new-manual-nomor" bind:value={newManualNomor} oninput={updatePreview} placeholder="Isi jika ingin pakai nomor kustom" />
				<p class="text-xs text-muted-foreground">Kosongkan untuk menggunakan nomor otomatis</p>
			</div>
			<div class="space-y-1">
				<label for="new-tujuan" class="text-sm font-medium text-foreground">Tujuan <span class="text-destructive">*</span></label>
				<Input id="new-tujuan" bind:value={newTujuan} placeholder="Instansi atau nama tujuan" />
			</div>
			<div class="space-y-1">
				<label for="new-perihal-k" class="text-sm font-medium text-foreground">Perihal <span class="text-destructive">*</span></label>
				<Input id="new-perihal-k" bind:value={newPerihal} placeholder="Perihal / isi singkat surat" />
			</div>
			<div class="space-y-1">
				<label for="new-sifat-k" class="text-sm font-medium text-foreground">Sifat</label>
				<select
					id="new-sifat-k"
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
				<label for="new-catatan-k" class="text-sm font-medium text-foreground">Catatan</label>
				<Textarea id="new-catatan-k" bind:value={newCatatan} placeholder="Catatan tambahan (opsional)" rows={2} />
			</div>
		</div>
		<Dialog.Footer>
			<Button variant="outline" onclick={() => (createOpen = false)}>Batal</Button>
			<LoadingButton loading={createBusy} onclick={() => void createLetter()}>Simpan</LoadingButton>
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
				<label for="edit-tgl-surat" class="text-sm font-medium text-foreground">Tanggal Surat</label>
				<Input id="edit-tgl-surat" type="date" bind:value={editTglSurat} />
			</div>
			<div class="space-y-1">
				<label for="edit-tujuan" class="text-sm font-medium text-foreground">Tujuan <span class="text-destructive">*</span></label>
				<Input id="edit-tujuan" bind:value={editTujuan} />
			</div>
			<div class="space-y-1">
				<label for="edit-perihal" class="text-sm font-medium text-foreground">Perihal <span class="text-destructive">*</span></label>
				<Input id="edit-perihal" bind:value={editPerihal} />
			</div>
			<div class="space-y-1">
				<label for="edit-sifat" class="text-sm font-medium text-foreground">Sifat</label>
				<select
					id="edit-sifat"
					bind:value={editSifat}
					class="w-full rounded-md border border-border bg-card px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
				>
					<option value="biasa">Biasa</option>
					<option value="penting">Penting</option>
					<option value="segera">Segera</option>
					<option value="rahasia">Rahasia</option>
				</select>
			</div>
			<div class="space-y-1">
				<label for="edit-catatan" class="text-sm font-medium text-foreground">Catatan</label>
				<Textarea id="edit-catatan" bind:value={editCatatan} rows={2} />
			</div>
		</div>
		<Dialog.Footer>
			<Button variant="outline" onclick={() => (editOpen = false)}>Batal</Button>
			<LoadingButton loading={editBusy} onclick={() => void saveEdit()}>Simpan</LoadingButton>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
