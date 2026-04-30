<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';

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
	interface Student { id: string; nis: string; nama: string; }
	interface Employee { id: string; nip: string; nama: string; }

	let loans = $state<LoanRow[]>([]);
	let books = $state<Book[]>([]);
	let students = $state<Student[]>([]);
	let employees = $state<Employee[]>([]);
	let loading = $state(true);
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

	const filteredLoans = $derived(
		loans.filter((l) => !tabStatus || l.status === tabStatus)
	);

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

	async function load() {
		loading = true;
		try {
			const [lRes, bRes, sRes, eRes] = await Promise.all([
				fetch('/api/library/loans'),
				fetch('/api/library/books'),
				fetch('/api/students'),
				fetch('/api/employees'),
			]);
			const lj = await lRes.json(); loans = lj.data ?? lj ?? [];
			const bj = await bRes.json(); books = bj.data ?? bj ?? [];
			const sj = await sRes.json(); students = sj.data ?? sj ?? [];
			const ej = await eRes.json(); employees = ej.data ?? ej ?? [];
		} finally {
			loading = false;
		}
	}

	function selectMember(id: string, nama: string, nip_nis: string) {
		fMemberId = id;
		fMemberLabel = `${nama} (${nip_nis})`;
		fMemberSearch = nama;
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
			const j = await res.json();
			if (!res.ok) { toast.error(j.error ?? 'Gagal meminjamkan buku'); return; }
			toast.success('Buku berhasil dipinjamkan');
			showLoanDialog = false;
			resetLoanForm();
			await load();
		} finally {
			busy = false;
		}
	}

	async function submitReturn() {
		if (!returnLoan) return;
		busy = true;
		try {
			const res = await fetch(`/api/library/loans/${returnLoan.id}/return`, { method: 'POST' });
			const j = await res.json();
			if (!res.ok) { toast.error(j.error ?? 'Gagal mengembalikan'); return; }
			toast.success('Buku berhasil dikembalikan');
			returnLoan = null;
			showReturnDialog = false;
			await load();
		} finally {
			busy = false;
		}
	}

	async function submitLunas() {
		if (!lunasLoanId) return;
		busy = true;
		try {
			const res = await fetch(`/api/library/loans/${lunasLoanId}/lunas`, { method: 'POST' });
			const j = await res.json();
			if (!res.ok) { toast.error(j.error ?? 'Gagal'); return; }
			toast.success('Denda ditandai lunas');
			lunasLoanId = null;
			showLunasDialog = false;
			await load();
		} finally {
			busy = false;
		}
	}

	onMount(load);
</script>

<svelte:head><title>Peminjaman — Perpustakaan</title></svelte:head>

<div class="space-y-4">
	<div class="flex flex-wrap items-center justify-between gap-3">
		<div>
			<h1 class="text-lg font-semibold text-slate-800">Peminjaman Buku</h1>
			<p class="text-sm text-slate-500">{loans.filter(l => l.status === 'active').length} pinjaman aktif</p>
		</div>
		<Button onclick={() => { resetLoanForm(); showLoanDialog = true; }} size="sm">+ Pinjam Buku</Button>
	</div>

	<!-- Tab filter -->
	<div class="flex gap-1 border-b border-slate-200">
		{#each [['active', 'Aktif'], ['', 'Semua'], ['returned', 'Dikembalikan']] as [val, label]}
			<button
				class="px-4 py-2 text-sm font-medium transition-colors border-b-2 {tabStatus === val ? 'border-green-700 text-green-800' : 'border-transparent text-slate-500 hover:text-slate-700'}"
				onclick={() => (tabStatus = val as 'active' | 'returned' | '')}
			>{label}</button>
		{/each}
	</div>

	<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
		<Card.Content class="p-0">
			{#if loading}
				<p class="p-6 text-sm text-slate-400">Memuat data pinjaman…</p>
			{:else if filteredLoans.length === 0}
				<p class="p-6 text-sm text-slate-400">Tidak ada data pinjaman.</p>
			{:else}
				<Table.Root>
					<Table.Header>
						<Table.Row class="bg-slate-50 text-xs">
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
						{#each filteredLoans as loan (loan.id)}
							<Table.Row class="text-sm">
								<Table.Cell>
									<p class="font-medium max-w-[140px] truncate" title={loan.book_judul}>{loan.book_judul}</p>
									<p class="text-xs text-slate-400 font-mono">{loan.book_kode}</p>
								</Table.Cell>
								<Table.Cell>
									<p class="max-w-[120px] truncate" title={loan.member_nama}>{loan.member_nama}</p>
									<p class="text-xs text-slate-400">{loan.member_nip_nis}</p>
								</Table.Cell>
								<Table.Cell class="text-xs text-slate-600">{formatDate(loan.dipinjam_at)}</Table.Cell>
								<Table.Cell>
									<span class={loan.is_overdue ? 'text-red-600 font-semibold' : 'text-slate-700'}>
										{formatDate(loan.jatuh_tempo)}
									</span>
								</Table.Cell>
								<Table.Cell>
									{#if loan.status === 'returned'}
										<Badge variant="secondary" class="text-xs">Dikembalikan</Badge>
									{:else if loan.is_overdue}
										<Badge variant="destructive" class="text-xs">Terlambat</Badge>
									{:else}
										<Badge class="bg-blue-100 text-blue-800 text-xs hover:bg-blue-100">Aktif</Badge>
									{/if}
								</Table.Cell>
								<Table.Cell class="text-xs">
									{#if loan.denda_total > 0}
										<span class={loan.denda_lunas ? 'text-slate-400 line-through' : 'text-orange-600 font-medium'}>
											{formatRupiah(loan.denda_total)}
										</span>
										{#if loan.denda_lunas}
											<span class="ml-1 text-green-600">✓</span>
										{/if}
									{:else}
										<span class="text-slate-400">-</span>
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
		<div class="flex items-center gap-2 border-b border-slate-100 pb-3">
			<span class="text-xs text-slate-500">Mode:</span>
			{#each [['beginner', 'Cepat'], ['advance', 'Lengkap']] as [val, label]}
				<button
					class="rounded-full border px-3 py-1 text-xs transition-colors {loanFormMode === val ? 'bg-green-700 text-white border-green-700' : 'border-slate-300 text-slate-600 hover:border-slate-400'}"
					onclick={() => (loanFormMode = val as 'beginner' | 'advance')}
				>{label}</button>
			{/each}
			{#if loanFormMode === 'beginner'}
				<span class="text-xs text-slate-400">— 7 hari, langsung pinjam</span>
			{:else}
				<span class="text-xs text-slate-400">— atur durasi & catatan</span>
			{/if}
		</div>

		<div class="space-y-4 py-2">
			<!-- Member type -->
			<div class="space-y-2">
				<p class="text-sm font-medium">Jenis Anggota</p>
				<div class="flex gap-2">
					{#each [['student', 'Siswa'], ['employee', 'Pegawai']] as [val, label]}
						<button
							class="rounded-full border px-3 py-1 text-xs transition-colors {fMemberType === val ? 'bg-green-700 text-white border-green-700' : 'bg-white text-slate-600 border-slate-300'}"
							onclick={() => { fMemberType = val as 'student' | 'employee'; fMemberId = ''; fMemberLabel = ''; fMemberSearch = ''; }}
						>{label}</button>
					{/each}
				</div>
			</div>

			<!-- Member search -->
			<div class="space-y-1">
				<label for="m-search" class="block text-sm font-medium">Cari Anggota <span class="text-red-500">*</span></label>
				<Input id="m-search" bind:value={fMemberSearch} placeholder={fMemberType === 'student' ? 'Nama atau NIS siswa…' : 'Nama atau NIP pegawai…'} />
				{#if fMemberSearch && !fMemberId}
					<div class="mt-1 rounded-md border border-slate-200 bg-white shadow-sm">
						{#each (fMemberType === 'student' ? filteredStudents : filteredEmployees) as m}
							<button
								class="flex w-full items-center gap-2 px-3 py-2 text-sm hover:bg-slate-50 text-left"
								onclick={() => selectMember(m.id, m.nama, (m as Student).nis ?? (m as Employee).nip)}
							>
								<span class="font-medium">{m.nama}</span>
								<span class="text-xs text-slate-400">{(m as Student).nis ?? (m as Employee).nip}</span>
							</button>
						{:else}
							<p class="px-3 py-2 text-sm text-slate-400">Tidak ditemukan</p>
						{/each}
					</div>
				{/if}
				{#if fMemberId}
					<p class="text-xs text-green-700">✓ {fMemberLabel}</p>
				{/if}
			</div>

			<!-- Book search -->
			<div class="space-y-1">
				<label for="b-search" class="block text-sm font-medium">Cari Buku <span class="text-red-500">*</span></label>
				<Input id="b-search" bind:value={fBookSearch} placeholder="Judul atau kode buku (hanya yang tersedia)…" />
				{#if fBookSearch && !fBookId}
					<div class="mt-1 rounded-md border border-slate-200 bg-white shadow-sm">
						{#each availableBooks as b}
							<button
								class="flex w-full items-center gap-2 px-3 py-2 text-sm hover:bg-slate-50 text-left"
								onclick={() => selectBook(b.id, b.kode, b.judul)}
							>
								<span class="font-mono text-xs text-slate-400">{b.kode}</span>
								<span class="font-medium">{b.judul}</span>
								<span class="ml-auto text-xs text-green-700">{b.tersedia} tersedia</span>
							</button>
						{:else}
							<p class="px-3 py-2 text-sm text-slate-400">Tidak ditemukan atau stok habis</p>
						{/each}
					</div>
				{/if}
				{#if fBookId}
					<p class="text-xs text-green-700">✓ {fBookLabel}</p>
				{/if}
			</div>

			{#if loanFormMode === 'advance'}
				<!-- Duration (advance only) -->
				<div class="space-y-1">
					<label for="due-days" class="block text-sm font-medium">Durasi Pinjam (hari)</label>
					<Input id="due-days" type="number" min="1" max="30" bind:value={fDueDays} />
					<p class="text-xs text-slate-400">Maksimum 30 hari. Default: 7 hari.</p>
				</div>
			{:else}
				<p class="text-xs text-slate-500 bg-slate-50 rounded px-3 py-2">Durasi pinjam: <strong>7 hari</strong>. Pilih mode Lengkap untuk mengubah.</p>
			{/if}
		</div>
		<Dialog.Footer>
			<Button variant="outline" onclick={() => (showLoanDialog = false)}>Batal</Button>
			<Button onclick={submitLoan} disabled={busy || !fMemberId || !fBookId}>
				{busy ? 'Memproses…' : 'Pinjamkan'}
			</Button>
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
			<div class="space-y-2 text-sm text-slate-700">
				<p><span class="font-medium">Buku:</span> {returnLoan.book_judul}</p>
				<p><span class="font-medium">Anggota:</span> {returnLoan.member_nama}</p>
				<p><span class="font-medium">Jatuh Tempo:</span> {formatDate(returnLoan.jatuh_tempo)}</p>
				{#if estimatedDenda > 0}
					<div class="rounded-md bg-orange-50 border border-orange-200 p-3">
						<p class="font-medium text-orange-700">Denda Keterlambatan</p>
						<p class="text-orange-600 text-lg font-bold">{formatRupiah(estimatedDenda)}</p>
					</div>
				{:else}
					<p class="text-green-700">Pengembalian tepat waktu — tidak ada denda.</p>
				{/if}
			</div>
		{/if}
		<Dialog.Footer>
			<Button variant="outline" onclick={() => { returnLoan = null; showReturnDialog = false; }}>Batal</Button>
			<Button onclick={submitReturn} disabled={busy}>{busy ? 'Memproses…' : 'Konfirmasi Kembalikan'}</Button>
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
			<Button onclick={submitLunas} disabled={busy}>{busy ? 'Menyimpan…' : 'Tandai Lunas'}</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
