<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { readClientApiData, readClientJson } from '$lib/client/api';
	import { displayName } from '$lib/utils/display-name';

	interface LoanRow {
		id: string;
		book_id: string;
		book_kode: string;
		book_judul: string;
		member_type: string;
		member_nama: string;
		member_nip_nis: string;
		dipinjam_at: string;
		jatuh_tempo: string;
		dikembalikan_at: string | null;
		denda_per_hari: number;
		denda_total: number;
		denda_lunas: boolean;
		status: string;
		is_overdue: boolean;
	}

	interface Book { id: string; kode: string; judul: string; tersedia: number; }
	interface Student { id: string; nis: string; nama: string; display_name?: string; }
	interface Employee { id: string; nip: string; nama: string; display_name?: string; }

	interface LoansOverview {
		loans: LoanRow[];
		books: Book[];
		students: Student[];
		employees: Employee[];
	}

	let loans = $state<LoanRow[]>([]);
	let books = $state<Book[]>([]);
	let students = $state<Student[]>([]);
	let employees = $state<Employee[]>([]);
	let loansPromise = $state<Promise<LoansOverview> | null>(null);
	let loansRequestId = 0;
	let busy = $state(false);

	let tabStatus = $state<'active' | 'returned' | ''>('active');

	// Loan form state
	let showLoanDialog = $state(false);
	let loanFormMode = $state<'beginner' | 'advance'>('beginner');
	let fMemberType = $state<'student' | 'employee'>('student');
	let fMemberSearch = $state('');
	let fMemberId = $state('');
	let fMemberLabel = $state('');
	let fBookSearch = $state('');
	let fBookId = $state('');
	let fBookLabel = $state('');
	let fDueDays = $state(7);

	// Return dialog
	let returnLoan = $state<LoanRow | null>(null);
	let showReturnDialog = $state(false);

	// Denda lunas confirm
	let lunasLoanId = $state<string | null>(null);
	let showLunasDialog = $state(false);

	const filteredStudents = $derived(
		students.filter((s) => {
			const q = fMemberSearch.toLowerCase();
			return !q || s.nama.toLowerCase().includes(q) || s.nis.includes(q);
		}).slice(0, 8)
	);

	const filteredEmployees = $derived(
		employees.filter((e) => {
			const q = fMemberSearch.toLowerCase();
			return !q || e.nama.toLowerCase().includes(q) || e.nip.includes(q);
		}).slice(0, 8)
	);

	const availableBooks = $derived(
		books.filter((b) => {
			const q = fBookSearch.toLowerCase();
			const matchSearch = !q || b.judul.toLowerCase().includes(q) || b.kode.toLowerCase().includes(q);
			return matchSearch && b.tersedia > 0;
		}).slice(0, 8)
	);

	const estimatedDenda = $derived.by(() => {
		if (!returnLoan) return 0;
		const due = new Date(returnLoan.jatuh_tempo);
		const now = new Date();
		if (now <= due) return 0;
		const days = Math.ceil((now.getTime() - due.getTime()) / 86400000);
		return days * (returnLoan.denda_per_hari ?? 500);
	});

	function formatDate(iso: string | null | undefined) {
		if (!iso) return '-';
		return new Date(iso).toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' });
	}

	function formatRupiah(n: number) {
		return `Rp ${n.toLocaleString('id-ID')}`;
	}

	function filterLoans(loanRows: LoanRow[]) {
		return loanRows.filter((l) => !tabStatus || l.status === tabStatus);
	}

	function applyOverview(overview: LoansOverview) {
		loans = overview.loans ?? [];
		books = overview.books ?? [];
		students = overview.students ?? [];
		employees = overview.employees ?? [];
	}

	async function fetchOverview(): Promise<LoansOverview> {
		const [lRes, bRes, sRes, eRes] = await Promise.all([
			fetch('/api/library/loans'),
			fetch('/api/library/books'),
			fetch('/api/students'),
			fetch('/api/employees'),
		]);
		const [nextLoans, nextBooks, nextStudents, nextEmployees] = await Promise.all([
			readClientApiData<LoanRow[]>(lRes, 'Gagal memuat data peminjaman.'),
			readClientApiData<Book[]>(bRes, 'Gagal memuat stok buku.'),
			readClientApiData<Student[]>(sRes, 'Gagal memuat data siswa.'),
			readClientApiData<Employee[]>(eRes, 'Gagal memuat data pegawai.'),
		]);
		return {
			loans: nextLoans,
			books: nextBooks,
			students: nextStudents,
			employees: nextEmployees,
		};
	}

	function load() {
		const requestId = ++loansRequestId;
		const emptyOverview: LoansOverview = { loans: [], books: [], students: [], employees: [] };
		applyOverview(emptyOverview);
		loansPromise = fetchOverview()
			.then((overview) => {
				if (requestId === loansRequestId) {
					applyOverview(overview);
					return overview;
				}
				return { loans, books, students, employees };
			})
			.catch((error: unknown) => {
				if (requestId === loansRequestId) throw error;
				return { loans, books, students, employees };
			});
	}

	async function refreshOverview() {
		const requestId = ++loansRequestId;
		const overview = await fetchOverview();
		if (requestId === loansRequestId) {
			applyOverview(overview);
			loansPromise = Promise.resolve(overview);
		}
	}

	function retryLoans(reset?: () => void) {
		reset?.();
		load();
	}

	function loansErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat data peminjaman, anggota, atau stok buku. Coba lagi untuk memulihkan tampilan operasional.';
	}

	function mutationErrorMessage(error: unknown, fallback: string) {
		if (error instanceof Error && error.message.trim() && !error.message.toLowerCase().includes('fetch')) return error.message;
		return fallback;
	}

	async function refreshLoansAfterMutation() {
		try {
			await refreshOverview();
		} catch (error) {
			loansPromise = Promise.resolve({ loans, books, students, employees });
			toast.error(loansErrorMessage(error));
		}
	}

	function handleLoansRenderError(error: unknown) {
		console.error('Library loans render failed', error);
	}

	function selectMember(id: string, nama: string, nip_nis: string) {
		fMemberId = id;
		fMemberLabel = `${nama} (${nip_nis})`;
		fMemberSearch = nama;
	}

	function memberName(member: Pick<LoanRow, 'member_nama'> | Student | Employee) {
		return displayName(
			{
				display_name: 'display_name' in member ? member.display_name : '',
				nama: 'nama' in member ? member.nama : member.member_nama
			},
			'Anggota'
		);
	}

	function selectBook(id: string, kode: string, judul: string) {
		fBookId = id;
		fBookLabel = `[${kode}] ${judul}`;
		fBookSearch = `[${kode}] ${judul}`;
	}

	function resetLoanForm() {
		loanFormMode = 'beginner';
		fMemberType = 'student'; fMemberSearch = ''; fMemberId = ''; fMemberLabel = '';
		fBookSearch = ''; fBookId = ''; fBookLabel = ''; fDueDays = 7;
	}

	async function submitLoan() {
		if (!fMemberId) { toast.error('Pilih anggota terlebih dahulu'); return; }
		if (!fBookId) { toast.error('Pilih buku terlebih dahulu'); return; }
		busy = true;
		try {
			const res = await fetch('/api/library/loans', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ book_id: fBookId, member_type: fMemberType, member_id: fMemberId, due_days: fDueDays }),
			});
			await readClientJson<unknown>(res);
			toast.success('Buku berhasil dipinjamkan');
			showLoanDialog = false;
			resetLoanForm();
			await refreshLoansAfterMutation();
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Gagal meminjamkan buku. Periksa koneksi lalu coba lagi.'));
		} finally {
			busy = false;
		}
	}

	async function submitReturn() {
		if (!returnLoan) return;
		busy = true;
		try {
			const res = await fetch(`/api/library/loans/${returnLoan.id}/return`, { method: 'POST' });
			await readClientJson<unknown>(res);
			toast.success('Buku berhasil dikembalikan');
			returnLoan = null;
			showReturnDialog = false;
			await refreshLoansAfterMutation();
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Gagal mengembalikan buku. Periksa koneksi lalu coba lagi.'));
		} finally {
			busy = false;
		}
	}

	async function submitLunas() {
		if (!lunasLoanId) return;
		busy = true;
		try {
			const res = await fetch(`/api/library/loans/${lunasLoanId}/lunas`, { method: 'POST' });
			await readClientJson<unknown>(res);
			toast.success('Denda ditandai lunas');
			lunasLoanId = null;
			showLunasDialog = false;
			await refreshLoansAfterMutation();
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Gagal memperbarui status denda. Periksa koneksi lalu coba lagi.'));
		} finally {
			busy = false;
		}
	}

	onMount(() => {
		void load();
	});
</script>

<svelte:head><title>Peminjaman — Perpustakaan</title></svelte:head>

<div class="space-y-4">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-foreground">Peminjaman Buku</h1>
			<p class="text-sm text-muted-foreground">Kelola sirkulasi buku, pengembalian, dan status denda anggota perpustakaan.</p>
		</div>
		<Button onclick={() => { resetLoanForm(); showLoanDialog = true; }} size="sm">+ Pinjam Buku</Button>
	</div>

	<AsyncContent promise={loansPromise} onerror={handleLoansRenderError}>
		{#snippet pending()}
			<div class="grid gap-3 md:grid-cols-3">
				{#each ['Pinjaman Aktif', 'Terlambat', 'Buku Siap Pinjam'] as label (label)}
					<div class="rounded-2xl border border-border bg-muted/50 px-4 py-4">
						<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-muted-foreground">{label}</p>
						<Skeleton class="mt-3 h-8 w-16" />
						<Skeleton class="mt-2 h-4 w-52" />
					</div>
				{/each}
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Sirkulasi Belum Tersaji"
				message={loansErrorMessage(error)}
				onRetry={() => retryLoans(reset)}
			/>
		{/snippet}

		{#snippet children(value)}
			{@const overview = value as LoansOverview}
			<div class="grid gap-3 md:grid-cols-3">
				<div class="rounded-2xl border border-primary/20 bg-primary/10 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-primary">Pinjaman Aktif</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{overview.loans.filter((item) => item.status === 'active').length}</p>
					<p class="text-sm text-muted-foreground">transaksi yang masih berjalan saat ini</p>
				</div>
				<div class="rounded-2xl border border-warning/30 bg-warning/10 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-warning">Terlambat</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{overview.loans.filter((item) => item.is_overdue && item.status === 'active').length}</p>
					<p class="text-sm text-muted-foreground">pinjaman aktif yang melewati jatuh tempo</p>
				</div>
				<div class="rounded-2xl border border-accent bg-accent/60 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-accent-foreground">Buku Siap Pinjam</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{overview.books.filter((item) => item.tersedia > 0).length}</p>
					<p class="text-sm text-muted-foreground">judul yang masih punya stok untuk dipinjam</p>
				</div>
			</div>
		{/snippet}
	</AsyncContent>

	<Card.Root class="border-border shadow-sm">
		<Card.Content class="p-2">
			<div class="flex flex-wrap gap-1">
				{#each [['active', 'Aktif'], ['', 'Semua'], ['returned', 'Dikembalikan']] as [val, label] (val)}
					<button
						class="rounded-full px-4 py-2 text-sm font-medium transition-colors {tabStatus === val ? 'bg-primary text-primary-foreground shadow-sm' : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
						onclick={() => (tabStatus = val as 'active' | 'returned' | '')}
					>{label}</button>
				{/each}
			</div>
		</Card.Content>
	</Card.Root>

	<Card.Root class="overflow-hidden border-border shadow-sm">
		<Card.Content class="p-0">
			<AsyncContent promise={loansPromise} onerror={handleLoansRenderError}>
				{#snippet pending()}
					<div class="space-y-3 p-6">
						{#each Array.from({ length: 6 }) as _, index (`loan-skeleton-${index}`)}
							<div class="grid gap-3 md:grid-cols-[1.2fr_1fr_0.8fr_0.8fr_0.7fr_0.7fr_auto] md:items-center">
								<Skeleton class="h-5 w-40" />
								<Skeleton class="h-5 w-32" />
								<Skeleton class="h-5 w-24" />
								<Skeleton class="h-5 w-24" />
								<Skeleton class="h-6 w-20" />
								<Skeleton class="h-5 w-20" />
								<Skeleton class="h-9 w-28 justify-self-end" />
							</div>
						{/each}
					</div>
				{/snippet}

				{#snippet failed(error, reset)}
					<div class="p-4">
						<RecoveryPanel
							compact
							title="Sirkulasi Belum Tersaji"
							message={loansErrorMessage(error)}
							onRetry={() => retryLoans(reset)}
						/>
					</div>
				{/snippet}

				{#snippet children(value)}
					{@const currentLoans = filterLoans((value as LoansOverview).loans)}
					{#if currentLoans.length === 0}
						<div class="p-4">
							<EmptyStatePanel
								title={tabStatus === 'returned' ? 'Belum ada riwayat pengembalian' : tabStatus === 'active' ? 'Belum ada pinjaman aktif' : 'Belum ada data pinjaman'}
								description={tabStatus === 'returned'
									? 'Riwayat pengembalian akan muncul di sini setelah buku mulai diproses dan dikembalikan.'
									: tabStatus === 'active'
										? 'Belum ada transaksi pinjam yang sedang berjalan. Gunakan tombol pinjam untuk membuat transaksi pertama.'
										: 'Belum ada transaksi sirkulasi yang tercatat pada perpustakaan ini.'}
								compact
							/>
						</div>
					{:else}
						<Table.Root>
							<Table.Header>
								<Table.Row class="bg-muted/50 text-xs">
									<Table.Head>Buku</Table.Head>
									<Table.Head>Anggota</Table.Head>
									<Table.Head>Dipinjam</Table.Head>
									<Table.Head>Jatuh Tempo</Table.Head>
									<Table.Head>Status</Table.Head>
									<Table.Head>Denda</Table.Head>
									<Table.Head class="text-right">Aksi</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each currentLoans as loan (loan.id)}
									<Table.Row class="text-sm">
										<Table.Cell>
											<p class="font-medium max-w-[140px] truncate" title={loan.book_judul}>{loan.book_judul}</p>
											<p class="text-xs text-muted-foreground font-mono">{loan.book_kode}</p>
										</Table.Cell>
										<Table.Cell>
											<p class="max-w-[120px] truncate" title={memberName(loan)}>{memberName(loan)}</p>
											<p class="text-xs text-muted-foreground">{loan.member_nip_nis}</p>
										</Table.Cell>
										<Table.Cell class="text-xs text-muted-foreground">{formatDate(loan.dipinjam_at)}</Table.Cell>
										<Table.Cell>
											<span class={loan.is_overdue ? 'text-destructive font-semibold' : 'text-foreground'}>
												{formatDate(loan.jatuh_tempo)}
											</span>
										</Table.Cell>
										<Table.Cell>
											{#if loan.status === 'returned'}
												<Badge variant="secondary" class="text-xs">Dikembalikan</Badge>
											{:else if loan.is_overdue}
												<Badge variant="destructive" class="text-xs">Terlambat</Badge>
											{:else}
												<Badge class="bg-accent text-accent-foreground text-xs hover:bg-accent">Aktif</Badge>
											{/if}
										</Table.Cell>
										<Table.Cell class="text-xs">
											{#if loan.denda_total > 0}
												<span class={loan.denda_lunas ? 'text-muted-foreground line-through' : 'text-warning font-medium'}>
													{formatRupiah(loan.denda_total)}
												</span>
												{#if loan.denda_lunas}
													<span class="ml-1 text-success">✓</span>
												{/if}
											{:else}
												<span class="text-muted-foreground">-</span>
											{/if}
										</Table.Cell>
										<Table.Cell class="text-right">
											{#if loan.status === 'active'}
												<Button size="sm" variant="outline" onclick={() => { returnLoan = loan; showReturnDialog = true; }}>Kembalikan</Button>
											{:else if loan.denda_total > 0 && !loan.denda_lunas}
												<Button size="sm" variant="outline" onclick={() => { lunasLoanId = loan.id; showLunasDialog = true; }}>
													Tandai Lunas
												</Button>
											{/if}
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

<!-- Loan form dialog -->
<Dialog.Root bind:open={showLoanDialog}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>Pinjam Buku</Dialog.Title>
		</Dialog.Header>
		<!-- Mode toggle -->
		<div class="flex items-center gap-2 border-b border-border pb-3">
			<span class="text-xs text-muted-foreground">Mode:</span>
			{#each [['beginner', 'Cepat'], ['advance', 'Lengkap']] as [val, label] (val)}
				<button
					class="rounded-full border px-3 py-1 text-xs transition-colors {loanFormMode === val ? 'bg-success text-background border-success' : 'border-border text-muted-foreground hover:border-border'}"
					onclick={() => (loanFormMode = val as 'beginner' | 'advance')}
				>{label}</button>
			{/each}
			{#if loanFormMode === 'beginner'}
				<span class="text-xs text-muted-foreground">— 7 hari, langsung pinjam</span>
			{:else}
				<span class="text-xs text-muted-foreground">— atur durasi & catatan</span>
			{/if}
		</div>

		<div class="space-y-4 py-2">
			<!-- Member type -->
			<div class="space-y-2">
				<p class="text-sm font-medium">Jenis Anggota</p>
				<div class="flex gap-2">
					{#each [['student', 'Siswa'], ['employee', 'Pegawai']] as [val, label] (val)}
						<button
							class="rounded-full border px-3 py-1 text-xs transition-colors {fMemberType === val ? 'bg-success text-background border-success' : 'bg-card text-muted-foreground border-border'}"
							onclick={() => { fMemberType = val as 'student' | 'employee'; fMemberId = ''; fMemberLabel = ''; fMemberSearch = ''; }}
						>{label}</button>
					{/each}
				</div>
			</div>

			<!-- Member search -->
			<div class="space-y-1">
				<label for="m-search" class="block text-sm font-medium">Cari Anggota <span class="text-destructive">*</span></label>
				<Input id="m-search" bind:value={fMemberSearch} placeholder={fMemberType === 'student' ? 'Nama atau NIS siswa…' : 'Nama atau NIP pegawai…'} />
				{#if fMemberSearch && !fMemberId}
					<div class="mt-1 rounded-md border border-border bg-card shadow-sm">
						{#each (fMemberType === 'student' ? filteredStudents : filteredEmployees) as m (m.id)}
							<button
								class="flex w-full items-center gap-2 px-3 py-2 text-sm hover:bg-muted/50 text-left"
								onclick={() => selectMember(m.id, memberName(m), (m as Student).nis ?? (m as Employee).nip)}
							>
								<span class="font-medium">{memberName(m)}</span>
								<span class="text-xs text-muted-foreground">{(m as Student).nis ?? (m as Employee).nip}</span>
							</button>
						{:else}
							<p class="px-3 py-2 text-sm text-muted-foreground">Tidak ditemukan</p>
						{/each}
					</div>
				{/if}
				{#if fMemberId}
					<p class="text-xs text-success">✓ {fMemberLabel}</p>
				{/if}
			</div>

			<!-- Book search -->
			<div class="space-y-1">
				<label for="b-search" class="block text-sm font-medium">Cari Buku <span class="text-destructive">*</span></label>
				<Input id="b-search" bind:value={fBookSearch} placeholder="Judul atau kode buku (hanya yang tersedia)…" />
				{#if fBookSearch && !fBookId}
					<div class="mt-1 rounded-md border border-border bg-card shadow-sm">
						{#each availableBooks as b (b.id)}
							<button
								class="flex w-full items-center gap-2 px-3 py-2 text-sm hover:bg-muted/50 text-left"
								onclick={() => selectBook(b.id, b.kode, b.judul)}
							>
								<span class="font-mono text-xs text-muted-foreground">{b.kode}</span>
								<span class="font-medium">{b.judul}</span>
								<span class="ml-auto text-xs text-success">{b.tersedia} tersedia</span>
							</button>
						{:else}
							<p class="px-3 py-2 text-sm text-muted-foreground">Tidak ditemukan atau stok habis</p>
						{/each}
					</div>
				{/if}
				{#if fBookId}
					<p class="text-xs text-success">✓ {fBookLabel}</p>
				{/if}
			</div>

			{#if loanFormMode === 'advance'}
				<!-- Duration (advance only) -->
				<div class="space-y-1">
					<label for="due-days" class="block text-sm font-medium">Durasi Pinjam (hari)</label>
					<Input id="due-days" type="number" min="1" max="30" bind:value={fDueDays} />
					<p class="text-xs text-muted-foreground">Maksimum 30 hari. Default: 7 hari.</p>
				</div>
			{:else}
				<p class="text-xs text-muted-foreground bg-muted/50 rounded px-3 py-2">Durasi pinjam: <strong>7 hari</strong>. Pilih mode Lengkap untuk mengubah.</p>
			{/if}
		</div>
		<Dialog.Footer>
			<Button variant="outline" onclick={() => (showLoanDialog = false)}>Batal</Button>
			<LoadingButton onclick={() => void submitLoan()} loading={busy} disabled={busy || !fMemberId || !fBookId}>
				Pinjamkan
			</LoadingButton>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<!-- Return confirm dialog -->
<Dialog.Root bind:open={showReturnDialog}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>Konfirmasi Pengembalian</Dialog.Title>
		</Dialog.Header>
		{#if returnLoan}
			<div class="space-y-2 text-sm text-foreground">
				<p><span class="font-medium">Buku:</span> {returnLoan.book_judul}</p>
				<p><span class="font-medium">Anggota:</span> {memberName(returnLoan)}</p>
				<p><span class="font-medium">Jatuh Tempo:</span> {formatDate(returnLoan.jatuh_tempo)}</p>
				{#if estimatedDenda > 0}
					<div class="rounded-md bg-warning/10 border border-warning/30 p-3">
						<p class="font-medium text-warning">Denda Keterlambatan</p>
						<p class="text-warning text-lg font-bold">{formatRupiah(estimatedDenda)}</p>
					</div>
				{:else}
					<p class="text-success">Pengembalian tepat waktu — tidak ada denda.</p>
				{/if}
			</div>
		{/if}
		<Dialog.Footer>
			<Button variant="outline" onclick={() => { returnLoan = null; showReturnDialog = false; }}>Batal</Button>
			<LoadingButton onclick={() => void submitReturn()} loading={busy} disabled={busy}>Konfirmasi Kembalikan</LoadingButton>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<!-- Denda lunas confirm -->
<Dialog.Root bind:open={showLunasDialog}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>Tandai Denda Lunas</Dialog.Title>
			<Dialog.Description>Tandai bahwa denda untuk pinjaman ini sudah dilunasi oleh anggota.</Dialog.Description>
		</Dialog.Header>
		<Dialog.Footer>
			<Button variant="outline" onclick={() => { lunasLoanId = null; showLunasDialog = false; }}>Batal</Button>
			<LoadingButton onclick={() => void submitLunas()} loading={busy} disabled={busy}>Tandai Lunas</LoadingButton>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
