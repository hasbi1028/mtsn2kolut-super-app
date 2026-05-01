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
		total_jenis: number;
		total_unit: number;
		total_layak: number;
		perlu_restok: number;
		perlu_perawatan: number;
	}

	interface ItemRow {
		id: string;
		kode: string;
		nama: string;
		kategori: string;
		lokasi: string;
		kondisi: string;
		satuan: string;
		jumlah_total: number;
		jumlah_baik: number;
		min_stock: number;
		catatan: string;
	}

	interface InventoryOverview {
		stats: Stats;
		items: ItemRow[];
	}

	type ApiEnvelope<T> = {
		data?: T;
		error?: string;
		message?: string;
	};

	let inventoryPromise = $state<Promise<InventoryOverview> | null>(null);

	function lowStockItems(items: ItemRow[]) {
		return items
			.filter((item) => item.jumlah_baik <= item.min_stock)
			.sort((a, b) => a.jumlah_baik - b.jumlah_baik)
			.slice(0, 8);
	}

	function attentionItems(items: ItemRow[]) {
		return items
			.filter((item) => item.kondisi !== 'baik')
			.sort((a, b) => a.kondisi.localeCompare(b.kondisi))
			.slice(0, 8);
	}

	function locationRollup(items: ItemRow[]) {
		return Array.from(
			items
				.reduce(
					(map, item) => {
						const key = item.lokasi.trim() || 'Belum diatur';
						const current = map.get(key) ?? { lokasi: key, totalJenis: 0, totalUnit: 0, totalLayak: 0 };
						current.totalJenis += 1;
						current.totalUnit += item.jumlah_total;
						current.totalLayak += item.jumlah_baik;
						map.set(key, current);
						return map;
					},
					new Map<string, { lokasi: string; totalJenis: number; totalUnit: number; totalLayak: number }>()
				)
				.values()
		)
			.sort((a, b) => b.totalUnit - a.totalUnit)
			.slice(0, 8)
	}

	function conditionLabel(value: string) {
		if (value === 'perlu-perawatan') return 'Perlu Perawatan';
		if (value === 'rusak') return 'Rusak';
		return 'Baik';
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

	async function fetchInventoryOverview(): Promise<InventoryOverview> {
		const [statsRes, itemsRes] = await Promise.all([
			fetch('/api/inventory/stats'),
			fetch('/api/inventory/items'),
		]);
		const [stats, items] = await Promise.all([
			readApi<Stats>(statsRes, 'Gagal memuat statistik inventaris.'),
			readApi<ItemRow[]>(itemsRes, 'Gagal memuat daftar barang inventaris.'),
		]);
		return { stats, items: items ?? [] };
	}

	function load() {
		inventoryPromise = fetchInventoryOverview();
	}

	function retryInventory(reset?: () => void) {
		reset?.();
		load();
	}

	function inventoryErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) {
			return error.message;
		}
		return 'Gagal memuat ringkasan inventaris. Coba lagi untuk mengambil statistik dan daftar barang terbaru.';
	}

	function statCards(stats: Stats) {
		return [
			{ label: 'Total Jenis', value: stats.total_jenis, color: 'text-slate-700' },
			{ label: 'Total Unit', value: stats.total_unit, color: 'text-slate-700' },
			{ label: 'Unit Layak', value: stats.total_layak, color: 'text-green-700' },
			{ label: 'Perlu Restok', value: stats.perlu_restok, color: stats.perlu_restok ? 'text-amber-700' : 'text-slate-700' },
			{ label: 'Perlu Perawatan', value: stats.perlu_perawatan, color: stats.perlu_perawatan ? 'text-red-700' : 'text-slate-700' },
		];
	}

	function handleInventoryRenderError(error: unknown) {
		console.error('Inventory render failed', error);
	}

	onMount(() => {
		void load();
	});
</script>

<svelte:head><title>Inventaris — MTSN 2 Kolaka Utara</title></svelte:head>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-lg font-semibold text-slate-800">Inventaris</h1>
			<p class="text-sm text-slate-500">Ringkasan barang sekolah, stok layak pakai, dan item yang butuh perhatian.</p>
		</div>
		<div class="flex gap-2">
			<Button href="/inventory/items" size="sm">Kelola Barang</Button>
		</div>
	</div>

	<AsyncContent promise={inventoryPromise} onerror={handleInventoryRenderError}>
		{#snippet pending()}
			<div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
				{#each ['Total Jenis', 'Total Unit', 'Unit Layak', 'Perlu Restok', 'Perlu Perawatan'] as label (label)}
					<Card.Root class="border-slate-200">
						<Card.Content class="p-4">
							<p class="text-xs text-slate-500">{label}</p>
							<Skeleton class="mt-2 h-8 w-16" />
						</Card.Content>
					</Card.Root>
				{/each}
			</div>

			<div class="grid gap-6 lg:grid-cols-2">
				{#each ['inventory-stock-skeleton', 'inventory-condition-skeleton'] as key (key)}
					<Card.Root class="border-slate-200">
						<Card.Header class="pb-2">
							<Skeleton class="h-5 w-40" />
						</Card.Header>
						<Card.Content class="p-0">
							<div class="space-y-3 p-4">
								{#each Array.from({ length: 5 }) as _, index (`${key}-${index}`)}
									<div class="grid gap-3 sm:grid-cols-3 sm:items-center">
										<Skeleton class="h-5 w-32" />
										<Skeleton class="h-5 w-24" />
										<Skeleton class="h-5 w-24" />
									</div>
								{/each}
							</div>
						</Card.Content>
					</Card.Root>
				{/each}
			</div>

			<Card.Root class="border-slate-200">
				<Card.Header class="pb-2">
					<Skeleton class="h-5 w-40" />
				</Card.Header>
				<Card.Content class="p-0">
					<div class="space-y-3 p-4">
						{#each Array.from({ length: 5 }) as _, index (`inventory-location-skeleton-${index}`)}
							<div class="grid gap-3 sm:grid-cols-4 sm:items-center">
								<Skeleton class="h-5 w-32" />
								<Skeleton class="h-5 w-20" />
								<Skeleton class="h-5 w-20" />
								<Skeleton class="h-5 w-20" />
							</div>
						{/each}
					</div>
				</Card.Content>
			</Card.Root>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Data Inventaris Belum Tersaji"
				message={inventoryErrorMessage(error)}
				onRetry={() => retryInventory(reset)}
			/>
		{/snippet}

		{#snippet children(value)}
			{@const overview = value as InventoryOverview}
			{@const stockRows = lowStockItems(overview.items)}
			{@const attentionRows = attentionItems(overview.items)}
			{@const locationRows = locationRollup(overview.items)}

			<div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
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
						<Card.Title class="text-sm font-medium text-slate-700">Barang Perlu Restok</Card.Title>
					</Card.Header>
					<Card.Content class="p-0">
						{#if stockRows.length === 0}
							<div class="p-4">
								<EmptyStatePanel compact title="Tidak ada barang yang perlu restok" description="Semua item masih berada di atas batas minimum stok yang dicatat." />
							</div>
						{:else}
							<Table.Root>
								<Table.Header>
									<Table.Row class="bg-slate-50 text-xs">
										<Table.Head>Barang</Table.Head>
										<Table.Head>Lokasi</Table.Head>
										<Table.Head>Stok Layak</Table.Head>
									</Table.Row>
								</Table.Header>
								<Table.Body>
									{#each stockRows as item (item.id)}
										<Table.Row class="text-sm">
											<Table.Cell class="font-medium">
												<span class="block truncate max-w-[180px]" title={item.nama}>{item.nama}</span>
												<span class="text-xs text-slate-400">{item.kode}</span>
											</Table.Cell>
											<Table.Cell>{item.lokasi || '—'}</Table.Cell>
											<Table.Cell>
												<span class="font-medium text-amber-700">{item.jumlah_baik} {item.satuan}</span>
												<span class="ml-1 text-xs text-slate-400">min. {item.min_stock}</span>
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
						<Card.Title class="text-sm font-medium text-slate-700">Barang Perlu Perhatian</Card.Title>
					</Card.Header>
					<Card.Content class="p-0">
						{#if attentionRows.length === 0}
							<div class="p-4">
								<EmptyStatePanel compact title="Tidak ada barang bermasalah" description="Semua barang yang tercatat saat ini masih ditandai dalam kondisi baik." />
							</div>
						{:else}
							<Table.Root>
								<Table.Header>
									<Table.Row class="bg-slate-50 text-xs">
										<Table.Head>Barang</Table.Head>
										<Table.Head>Kondisi</Table.Head>
										<Table.Head>Jumlah Baik</Table.Head>
									</Table.Row>
								</Table.Header>
								<Table.Body>
									{#each attentionRows as item (item.id)}
										<Table.Row class="text-sm">
											<Table.Cell class="font-medium">
												<span class="block truncate max-w-[180px]" title={item.nama}>{item.nama}</span>
												<span class="text-xs text-slate-400">{item.kode}</span>
											</Table.Cell>
											<Table.Cell>
												<Badge variant={item.kondisi === 'rusak' ? 'destructive' : 'outline'}>
													{conditionLabel(item.kondisi)}
												</Badge>
											</Table.Cell>
											<Table.Cell>{item.jumlah_baik} / {item.jumlah_total}</Table.Cell>
										</Table.Row>
									{/each}
								</Table.Body>
							</Table.Root>
						{/if}
					</Card.Content>
				</Card.Root>
			</div>

			<Card.Root class="border-slate-200">
				<Card.Header class="pb-2">
					<Card.Title class="text-sm font-medium text-slate-700">Ringkasan per Lokasi</Card.Title>
				</Card.Header>
				<Card.Content class="p-0">
					{#if locationRows.length === 0}
						<div class="p-4">
							<EmptyStatePanel compact title="Belum ada ringkasan lokasi" description="Ringkasan per lokasi akan muncul setelah ada barang inventaris yang tercatat." />
						</div>
					{:else}
						<Table.Root>
							<Table.Header>
								<Table.Row class="bg-slate-50 text-xs">
									<Table.Head>Lokasi</Table.Head>
									<Table.Head>Jenis Barang</Table.Head>
									<Table.Head>Total Unit</Table.Head>
									<Table.Head>Unit Layak</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each locationRows as row (row.lokasi)}
									<Table.Row class="text-sm">
										<Table.Cell class="font-medium">{row.lokasi}</Table.Cell>
										<Table.Cell>{row.totalJenis}</Table.Cell>
										<Table.Cell>{row.totalUnit}</Table.Cell>
										<Table.Cell>{row.totalLayak}</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					{/if}
				</Card.Content>
			</Card.Root>
		{/snippet}
	</AsyncContent>
</div>
