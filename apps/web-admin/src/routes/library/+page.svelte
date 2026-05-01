<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';

	interface Stats {
		total_judul: number;
		total_eksemplar: number;
		total_tersedia: number;
		sedang_dipinjam: number;
		terlambat: number;
		denda_belum_lunas: number;
	}

	interface LoanRow {
		id: string;
		book_judul: string;
		book_kode: string;
		member_nama: string;
		member_nip_nis: string;
		member_type: string;
		dipinjam_at: string;
		jatuh_tempo: string;
		is_overdue: boolean;
		denda_total: number;
		denda_lunas: boolean;
		status: string;
	}

	interface LibraryOverview {
		stats: Stats;
		activeLoans: LoanRow[];
		overdueLoans: LoanRow[];
	}

	type ApiEnvelope<T> = {
		data?: T;
		error?: string;
		message?: string;
	};

	let overviewPromise = $state<Promise<LibraryOverview> | null>(null);

	function formatDate(iso: string) {
		if (!iso) return '-';
		return new Date(iso).toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' });
	}

	function formatRupiah(n: number) {
		return `Rp ${n.toLocaleString('id-ID')}`;
	}

	function isRecord(value: unknown): value is Record<string, unknown> {
		return typeof value === 'object' && value !== null;
	}

	function apiErrorMessage(payload: unknown) {
		if (!isRecord(payload)) return '';
		const error = payload.error;
		if (typeof error === 'string' && error.trim()) return error;
		const message = payload.message;
		if (typeof message === 'string' && message.trim()) return message;
		return '';
	}

	async function readApi<T>(response: Response, fallbackMessage: string): Promise<T> {
		const payload = (await response.json().catch(() => null)) as ApiEnvelope<T> | T | null;
		const message = apiErrorMessage(payload);
		if (!response.ok) {
			throw new Error(message || fallbackMessage);
		}
		if (isRecord(payload) && typeof payload.error === 'string' && payload.error.trim()) {
			throw new Error(payload.error);
		}
		if (isRecord(payload) && 'data' in payload) {
			const envelope = payload as ApiEnvelope<T>;
			if (envelope.data === undefined) throw new Error(fallbackMessage);
			return envelope.data;
		}
		if (payload === null) throw new Error(fallbackMessage);
		return payload as T;
	}

	async function fetchOverview(): Promise<LibraryOverview> {
		const [statsRes, loansRes] = await Promise.all([
			fetch('/api/library/stats'),
			fetch('/api/library/loans?status=active'),
		]);
		const [stats, loans] = await Promise.all([
			readApi<Stats>(statsRes, 'Gagal memuat statistik perpustakaan.'),
			readApi<LoanRow[]>(loansRes, 'Gagal memuat daftar pinjaman aktif.'),
		]);
		const rows = loans ?? [];
		return {
			stats,
			activeLoans: rows.slice(0, 10),
			overdueLoans: rows
				.filter((loan) => loan.is_overdue)
				.sort((a, b) => new Date(a.jatuh_tempo).getTime() - new Date(b.jatuh_tempo).getTime()),
		};
	}

	function load() {
		overviewPromise = fetchOverview();
	}

	function retryOverview(reset?: () => void) {
		reset?.();
		load();
	}

	function overviewErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat ringkasan perpustakaan. Coba lagi untuk mengambil statistik dan daftar pinjaman terbaru.';
	}

	function statCards(stats: Stats) {
		return [
			{ label: 'Total Judul', value: stats.total_judul, color: 'text-slate-700' },
			{ label: 'Total Eksemplar', value: stats.total_eksemplar, color: 'text-slate-700' },
			{ label: 'Tersedia', value: stats.total_tersedia, color: 'text-green-700' },
			{ label: 'Dipinjam', value: stats.sedang_dipinjam, color: 'text-blue-700' },
			{ label: 'Terlambat', value: stats.terlambat, color: stats.terlambat ? 'text-red-600' : 'text-slate-700' },
			{ label: 'Denda Belum Lunas', value: stats.denda_belum_lunas, color: stats.denda_belum_lunas ? 'text-orange-600' : 'text-slate-700' },
		];
	}

	function handleOverviewRenderError(error: unknown) {
		console.error('Library overview render failed', error);
	}

	onMount(() => {
		void load();
	});
</script>

<svelte:head><title>Perpustakaan — MTSN 2 Kolaka Utara</title></svelte:head>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-lg font-semibold text-slate-800">Perpustakaan</h1>
			<p class="text-sm text-slate-500">Ringkasan operasional perpustakaan hari ini</p>
		</div>
		<div class="flex gap-2">
			<Button href="/library/books" variant="outline" size="sm">Katalog Buku</Button>
			<Button href="/library/loans" size="sm">Peminjaman</Button>
		</div>
	</div>

	<AsyncContent promise={overviewPromise} onerror={handleOverviewRenderError}>
		{#snippet pending()}
			<div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
				{#each ['Total Judul', 'Total Eksemplar', 'Tersedia', 'Dipinjam', 'Terlambat', 'Denda Belum Lunas'] as label (label)}
					<Card.Root class="border-slate-200">
						<Card.Content class="p-4">
							<p class="text-xs text-slate-500">{label}</p>
							<Skeleton class="mt-2 h-8 w-16" />
						</Card.Content>
					</Card.Root>
				{/each}
			</div>

			<div class="grid gap-6 lg:grid-cols-2">
				{#each ['active-loan-skeleton', 'overdue-loan-skeleton'] as key (key)}
					<Card.Root class="border-slate-200">
						<Card.Header class="pb-2">
							<Skeleton class="h-5 w-44" />
						</Card.Header>
						<Card.Content class="p-0">
							<div class="space-y-3 p-4">
								{#each Array.from({ length: 5 }) as _, index (`${key}-${index}`)}
									<div class="grid gap-3 sm:grid-cols-3 sm:items-center">
										<Skeleton class="h-5 w-32" />
										<Skeleton class="h-5 w-28" />
										<Skeleton class="h-5 w-24" />
									</div>
								{/each}
							</div>
						</Card.Content>
					</Card.Root>
				{/each}
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Data Perpustakaan Belum Tersaji"
				message={overviewErrorMessage(error)}
				onRetry={() => retryOverview(reset)}
			/>
		{/snippet}

		{#snippet children(value)}
			{@const overview = value as LibraryOverview}
			<div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
				{#each statCards(overview.stats) as card (card.label)}
					<Card.Root class="border-slate-200">
						<Card.Content class="p-4">
							<p class="text-xs text-slate-500">{card.label}</p>
							<p class="mt-1 text-2xl font-bold {card.color}">{card.value}</p>
						</Card.Content>
					</Card.Root>
				{/each}
			</div>

			<div class="grid gap-6 lg:grid-cols-2">
				<Card.Root class="border-slate-200">
					<Card.Header class="pb-2">
						<Card.Title class="text-sm font-medium text-slate-700">Pinjaman Aktif Terbaru</Card.Title>
					</Card.Header>
					<Card.Content class="p-0">
						{#if overview.activeLoans.length === 0}
							<div class="p-4">
								<EmptyStatePanel
									compact
									title="Tidak ada pinjaman aktif"
									description="Belum ada buku yang sedang dipinjam. Operasional pinjaman baru akan muncul di panel ini."
								/>
							</div>
						{:else}
							<Table.Root>
								<Table.Header>
									<Table.Row class="bg-slate-50 text-xs">
										<Table.Head>Buku</Table.Head>
										<Table.Head>Anggota</Table.Head>
										<Table.Head>Jatuh Tempo</Table.Head>
									</Table.Row>
								</Table.Header>
								<Table.Body>
									{#each overview.activeLoans as loan (loan.id)}
										<Table.Row class="text-sm">
											<Table.Cell class="font-medium">
												<span class="block truncate max-w-[160px]" title={loan.book_judul}>{loan.book_judul}</span>
												<span class="text-xs text-slate-400">{loan.book_kode}</span>
											</Table.Cell>
											<Table.Cell>
												<span class="block truncate max-w-[120px]" title={loan.member_nama}>{loan.member_nama}</span>
												<span class="text-xs text-slate-400">{loan.member_nip_nis}</span>
											</Table.Cell>
											<Table.Cell>
												<span class={loan.is_overdue ? 'text-red-600 font-medium' : ''}>
													{formatDate(loan.jatuh_tempo)}
												</span>
											</Table.Cell>
										</Table.Row>
									{/each}
								</Table.Body>
							</Table.Root>
						{/if}
					</Card.Content>
				</Card.Root>

				<Card.Root class="border-slate-200">
					<Card.Header class="pb-2">
						<Card.Title class="text-sm font-medium text-slate-700">
							Keterlambatan
							{#if overview.overdueLoans.length > 0}
								<Badge variant="destructive" class="ml-2 text-xs">{overview.overdueLoans.length}</Badge>
							{/if}
						</Card.Title>
					</Card.Header>
					<Card.Content class="p-0">
						{#if overview.overdueLoans.length === 0}
							<div class="p-4">
								<EmptyStatePanel
									compact
									title="Tidak ada keterlambatan"
									description="Semua pinjaman aktif masih dalam batas waktu. Panel ini akan menyorot denda dan jatuh tempo yang lewat."
								/>
							</div>
						{:else}
							<Table.Root>
								<Table.Header>
									<Table.Row class="bg-slate-50 text-xs">
										<Table.Head>Buku</Table.Head>
										<Table.Head>Anggota</Table.Head>
										<Table.Head>Jatuh Tempo</Table.Head>
									</Table.Row>
								</Table.Header>
								<Table.Body>
									{#each overview.overdueLoans as loan (loan.id)}
										<Table.Row class="text-sm">
											<Table.Cell class="font-medium">
												<span class="block truncate max-w-[160px]" title={loan.book_judul}>{loan.book_judul}</span>
												<span class="text-xs text-slate-400">{loan.book_kode}</span>
											</Table.Cell>
											<Table.Cell>
												<span class="block truncate max-w-[120px]" title={loan.member_nama}>{loan.member_nama}</span>
												<span class="text-xs text-slate-400">{loan.member_nip_nis}</span>
											</Table.Cell>
											<Table.Cell class="text-red-600 font-medium text-xs">
												{formatDate(loan.jatuh_tempo)}
												{#if loan.denda_total > 0}
													<span class="block text-orange-600">{formatRupiah(loan.denda_total)}</span>
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
		{/snippet}
	</AsyncContent>
</div>
