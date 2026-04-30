<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';

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

	let stats = $state<Stats | null>(null);
	let activeLoans = $state<LoanRow[]>([]);
	let overdueLoans = $state<LoanRow[]>([]);
	let loading = $state(true);

	function formatDate(iso: string) {
		if (!iso) return '-';
		return new Date(iso).toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' });
	}

	function formatRupiah(n: number) {
		return `Rp ${n.toLocaleString('id-ID')}`;
	}

	async function load() {
		loading = true;
		try {
			const [sRes, lRes] = await Promise.all([
				fetch('/api/library/stats'),
				fetch('/api/library/loans?status=active'),
			]);
			const sJson = await sRes.json();
			const lJson = await lRes.json();
			stats = sJson.data ?? sJson;
			const loans: LoanRow[] = lJson.data ?? lJson ?? [];
			activeLoans = loans.slice(0, 10);
			overdueLoans = loans.filter((l: LoanRow) => l.is_overdue).sort((a: LoanRow, b: LoanRow) =>
				new Date(a.jatuh_tempo).getTime() - new Date(b.jatuh_tempo).getTime()
			);
		} finally {
			loading = false;
		}
	}

	onMount(load);
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

	<!-- Stats cards -->
	<div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
		{#each [
			{ label: 'Total Judul', value: stats?.total_judul ?? '-', color: 'text-slate-700' },
			{ label: 'Total Eksemplar', value: stats?.total_eksemplar ?? '-', color: 'text-slate-700' },
			{ label: 'Tersedia', value: stats?.total_tersedia ?? '-', color: 'text-green-700' },
			{ label: 'Dipinjam', value: stats?.sedang_dipinjam ?? '-', color: 'text-blue-700' },
			{ label: 'Terlambat', value: stats?.terlambat ?? '-', color: stats?.terlambat ? 'text-red-600' : 'text-slate-700' },
			{ label: 'Denda Belum Lunas', value: stats?.denda_belum_lunas ?? '-', color: stats?.denda_belum_lunas ? 'text-orange-600' : 'text-slate-700' },
		] as card}
			<Card.Root class="border-slate-200">
				<Card.Content class="p-4">
					<p class="text-xs text-slate-500">{card.label}</p>
					{#if loading}
						<Skeleton class="mt-2 h-8 w-16" />
					{:else}
						<p class="mt-1 text-2xl font-bold {card.color}">{card.value}</p>
					{/if}
				</Card.Content>
			</Card.Root>
		{/each}
	</div>

	<div class="grid gap-6 lg:grid-cols-2">
		<!-- Active loans today -->
		<Card.Root class="border-slate-200">
			<Card.Header class="pb-2">
				<Card.Title class="text-sm font-medium text-slate-700">Pinjaman Aktif Terbaru</Card.Title>
			</Card.Header>
			<Card.Content class="p-0">
				{#if loading}
					<div class="space-y-3 p-4">
						{#each Array.from({ length: 5 }) as _, index (`active-loan-skeleton-${index}`)}
							<div class="grid gap-3 sm:grid-cols-3 sm:items-center">
								<Skeleton class="h-5 w-32" />
								<Skeleton class="h-5 w-28" />
								<Skeleton class="h-5 w-24" />
							</div>
						{/each}
					</div>
				{:else if activeLoans.length === 0}
					<p class="px-4 pb-4 text-sm text-slate-400">Tidak ada pinjaman aktif.</p>
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
							{#each activeLoans as loan (loan.id)}
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

		<!-- Overdue loans -->
		<Card.Root class="border-slate-200">
			<Card.Header class="pb-2">
				<Card.Title class="text-sm font-medium text-slate-700">
					Keterlambatan
					{#if overdueLoans.length > 0}
						<Badge variant="destructive" class="ml-2 text-xs">{overdueLoans.length}</Badge>
					{/if}
				</Card.Title>
			</Card.Header>
			<Card.Content class="p-0">
				{#if loading}
					<div class="space-y-3 p-4">
						{#each Array.from({ length: 5 }) as _, index (`overdue-loan-skeleton-${index}`)}
							<div class="grid gap-3 sm:grid-cols-3 sm:items-center">
								<Skeleton class="h-5 w-32" />
								<Skeleton class="h-5 w-28" />
								<Skeleton class="h-5 w-24" />
							</div>
						{/each}
					</div>
				{:else if overdueLoans.length === 0}
					<p class="px-4 pb-4 text-sm text-slate-400">Tidak ada keterlambatan.</p>
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
							{#each overdueLoans as loan (loan.id)}
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
</div>
