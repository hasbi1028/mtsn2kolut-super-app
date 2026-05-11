<script lang="ts">
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import ArchiveIcon from '@lucide/svelte/icons/archive';
	import ClipboardListIcon from '@lucide/svelte/icons/clipboard-list';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import PackageIcon from '@lucide/svelte/icons/package';
	import PrinterIcon from '@lucide/svelte/icons/printer';
	import RefreshCcwIcon from '@lucide/svelte/icons/refresh-ccw';
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import * as Table from '$lib/components/ui/table';
	import {
		downloadCsv,
		fetchTUComplianceData,
		formatBytes,
		formatCurrency,
		formatDate,
		isArchiveRetentionAttention,
		isInventoryAttention,
		isOpenDisposition,
		isOpenIncoming,
		isWorkPlanAttention,
		latestByDate,
		numberValue,
		statusLabel,
		type ArchiveDocument,
		type InventoryItem,
		type TUComplianceData,
		type WorkPlanItem,
	} from '$lib/tu/compliance-data';

	type BadgeVariant = 'default' | 'secondary' | 'destructive' | 'outline';
	type AttentionHref = '/tu/disposisi' | '/tu/arsip' | '/inventory' | '/inventory/items' | '/governance';
	type AttentionItem = {
		key: string;
		title: string;
		meta: string;
		level: 'warning' | 'danger';
		href: AttentionHref;
	};

	let dashboardPromise = $state<Promise<TUComplianceData> | null>(null);
	let snapshot = $state<TUComplianceData | null>(null);
	let refreshBusy = $state(false);

	function load() {
		const promise = fetchTUComplianceData();
		dashboardPromise = promise;
		promise.then((data) => (snapshot = data)).catch(() => {});
		return promise;
	}

	async function refreshDashboard() {
		refreshBusy = true;
		try {
			await load();
		} catch {
			// AsyncContent owns the visible error state for the dashboard body.
		} finally {
			refreshBusy = false;
		}
	}

	function retryDashboard(reset?: () => void) {
		reset?.();
		load();
	}

	function dashboardErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Data dashboard TU belum dapat dimuat. Coba lagi untuk mengambil ringkasan terbaru.';
	}

	function statusVariant(value: string): BadgeVariant {
		if (['blocked', 'gap', 'disposed', 'cancelled'].includes(value)) return 'destructive';
		if (['done', 'selesai', 'verified', 'issued', 'active'].includes(value)) return 'default';
		if (['borrowed', 'ditindaklanjuti', 'collected', 'in_progress'].includes(value)) return 'secondary';
		return 'outline';
	}

	function buildAttention(data: TUComplianceData): AttentionItem[] {
		const items: AttentionItem[] = [];
		for (const disposition of data.dispositions.filter(isOpenDisposition).slice(0, 5)) {
			items.push({
				key: `disposition-${disposition.id}`,
				title: disposition.letter_perihal,
				meta: `${statusLabel(disposition.status)} - ${disposition.assignee_name || 'Belum ada penerima'}`,
				level: disposition.status === 'terkirim' ? 'danger' : 'warning',
				href: '/tu/disposisi',
			});
		}
		for (const item of data.inventoryItems.filter(isInventoryAttention).slice(0, 4)) {
			items.push({
				key: `inventory-${item.id}`,
				title: item.nama,
				meta: `${item.kondisi} - layak ${item.jumlah_baik}/${item.jumlah_total} ${item.satuan}`,
				level: item.kondisi === 'rusak' ? 'danger' : 'warning',
				href: '/inventory/items',
			});
		}
		for (const item of data.workPlanItems.filter(isWorkPlanAttention).slice(0, 4)) {
			items.push({
				key: `workplan-${item.id}`,
				title: item.activity_name,
				meta: `${item.program_code} - ${statusLabel(item.status)} - ${item.progress_percent}%`,
				level: item.status === 'blocked' ? 'danger' : 'warning',
				href: '/governance',
			});
		}
		for (const item of data.archiveDocuments.filter((archive) => isArchiveRetentionAttention(archive)).slice(0, 3)) {
			items.push({
				key: `archive-${item.id}`,
				title: item.title,
				meta: `Retensi ${formatDate(item.retention_until)} - ${item.storage_location || 'lokasi belum diisi'}`,
				level: 'warning',
				href: '/tu/arsip',
			});
		}
		return items.slice(0, 12);
	}

	function workPlanTotals(items: WorkPlanItem[]) {
		return items.reduce(
			(total, item) => ({
				budget: total.budget + numberValue(item.budget_amount),
				realization: total.realization + numberValue(item.realization_amount),
				blocked: total.blocked + (item.status === 'blocked' ? 1 : 0),
			}),
			{ budget: 0, realization: 0, blocked: 0 }
		);
	}

	function inventoryAttentionLabel(item: InventoryItem) {
		if (item.kondisi !== 'baik') return statusLabel(item.kondisi);
		if (item.jumlah_baik <= item.min_stock) return 'Perlu restok';
		return 'Baik';
	}

	function archiveMeta(item: ArchiveDocument) {
		const location = item.storage_location || 'Lokasi belum diisi';
		return `${item.category_code} - ${location} - ${formatBytes(item.file_size)}`;
	}

	function exportSummaryCsv() {
		if (!snapshot) return;
		const openIncoming = snapshot.incomingLetters.filter(isOpenIncoming).length;
		const openDispositions = snapshot.dispositions.filter(isOpenDisposition).length;
		const inventoryAttention = snapshot.inventoryItems.filter(isInventoryAttention).length;
		const workPlan = workPlanTotals(snapshot.workPlanItems);
		const rows = [
			['Area', 'Nilai', 'Catatan'],
			['Surat masuk terbuka', openIncoming, 'Status baru/didisposisi'],
			['Surat keluar', snapshot.outgoingLetters.length, 'Total register'],
			['Surat keterangan', snapshot.certificates.length, 'Total register'],
			['Disposisi terbuka', openDispositions, 'Belum selesai'],
			['Arsip dokumen', snapshot.archiveStats.total_documents, `${snapshot.archiveStats.documents_this_year} tahun ini`],
			['Inventaris perlu perhatian', inventoryAttention, 'Rusak/perawatan/restok'],
			['RKT/RKJM terkendala', workPlan.blocked, `Anggaran ${workPlan.budget}, realisasi ${workPlan.realization}`],
		];
		downloadCsv(`dashboard-tu-${new Date().toISOString().slice(0, 10)}.csv`, rows);
	}

	function handleRenderError(error: unknown) {
		console.error('TU dashboard render failed', error);
	}

	onMount(() => {
		void load();
	});
</script>

<svelte:head><title>Dashboard TU - MTsN 2 Kolaka Utara</title></svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
		<div>
			<h1 class="text-lg font-semibold text-foreground">Dashboard Tata Usaha</h1>
			<p class="text-sm text-muted-foreground">Kontrol harian surat, disposisi, arsip, inventaris, dan bukti kerja madrasah.</p>
		</div>
		<div class="flex flex-wrap gap-2">
			<Button href={`${resolve('/document-cycles')}?domain_area=tu`} variant="outline" size="sm">
				<ClipboardListIcon class="mr-2 size-4" />
				Siklus Dokumen
			</Button>
			<Button href="/tu/compliance-pack" variant="outline" size="sm">
				<PrinterIcon class="mr-2 size-4" />
				Paket Kepatuhan
			</Button>
			<Button variant="outline" size="sm" onclick={exportSummaryCsv}>
				<DownloadIcon class="mr-2 size-4" />
				Export Ringkasan
			</Button>
			<LoadingButton size="sm" onclick={() => void refreshDashboard()} loading={refreshBusy} loadingLabel="Memuat...">
				<RefreshCcwIcon class="mr-2 size-4" />
				Refresh
			</LoadingButton>
		</div>
	</div>

	<AsyncContent promise={dashboardPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
				{#each Array.from({ length: 8 }) as _, index (`tu-dashboard-stat-skeleton-${index}`)}
					<Card.Root class="border-border">
						<Card.Content class="p-4">
							<Skeleton class="h-4 w-28" />
							<Skeleton class="mt-3 h-8 w-16" />
							<Skeleton class="mt-2 h-3 w-32" />
						</Card.Content>
					</Card.Root>
				{/each}
			</div>
			<div class="grid gap-6 xl:grid-cols-[1fr_360px]">
				<Card.Root class="border-border">
					<Card.Content class="space-y-3 p-4">
						{#each Array.from({ length: 6 }) as _, index (`tu-dashboard-main-skeleton-${index}`)}
							<Skeleton class="h-10 w-full" />
						{/each}
					</Card.Content>
				</Card.Root>
				<Card.Root class="border-border">
					<Card.Content class="space-y-3 p-4">
						{#each Array.from({ length: 5 }) as _, index (`tu-dashboard-side-skeleton-${index}`)}
							<Skeleton class="h-10 w-full" />
						{/each}
					</Card.Content>
				</Card.Root>
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel message={dashboardErrorMessage(error)} onRetry={() => retryDashboard(reset)} />
		{/snippet}

		{#snippet children(value)}
			{@const data = value as TUComplianceData}
			{@const attention = buildAttention(data)}
			{@const workPlan = workPlanTotals(data.workPlanItems)}
			{@const openIncoming = data.incomingLetters.filter(isOpenIncoming)}
			{@const openDispositions = data.dispositions.filter(isOpenDisposition)}
			{@const inventoryAttention = data.inventoryItems.filter(isInventoryAttention)}
			<div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
				{#each [
					{ label: 'Surat Masuk Terbuka', value: openIncoming.length, detail: `${data.incomingLetters.length} total register`, href: '/tu/surat-masuk', icon: FileTextIcon },
					{ label: 'Surat Keluar', value: data.outgoingLetters.length, detail: 'Register nomor surat keluar', href: '/tu/surat-keluar', icon: FileTextIcon },
					{ label: 'Disposisi Terbuka', value: openDispositions.length, detail: `${data.dispositions.length} total disposisi`, href: '/tu/disposisi', icon: ClipboardListIcon },
					{ label: 'Surat Keterangan', value: data.certificates.length, detail: 'Terbit dan batal tercatat', href: '/tu/surat-keterangan', icon: FileTextIcon },
					{ label: 'Arsip Digital', value: data.archiveStats.total_documents, detail: `${data.archiveStats.documents_this_year} dokumen tahun ini`, href: '/tu/arsip', icon: ArchiveIcon },
					{ label: 'Kapasitas Arsip', value: formatBytes(data.archiveStats.total_file_size), detail: `${data.archiveStats.active_categories} kategori aktif`, href: '/tu/arsip', icon: ArchiveIcon },
					{ label: 'Inventaris Perhatian', value: inventoryAttention.length, detail: `${numberValue(data.inventoryStats.total_unit)} unit tercatat`, href: '/inventory', icon: PackageIcon },
					{ label: 'RKT Terkendala', value: workPlan.blocked, detail: `${formatCurrency(workPlan.realization)} realisasi`, href: '/governance', icon: AlertTriangleIcon },
				] as item (item.label)}
					{@const Icon = item.icon}
					<Card.Root class="border-border">
						<Card.Content class="p-4">
							<div class="flex items-start justify-between gap-3">
								<div>
									<p class="text-xs text-muted-foreground">{item.label}</p>
									<p class="mt-2 text-2xl font-semibold text-foreground">{item.value}</p>
									<p class="mt-1 text-xs text-muted-foreground">{item.detail}</p>
								</div>
								<Icon class="size-5 text-primary" />
							</div>
							<Button href={item.href} variant="ghost" size="sm" class="mt-3 px-0 text-primary">Buka modul</Button>
						</Card.Content>
					</Card.Root>
				{/each}
			</div>

			<div class="grid gap-6 xl:grid-cols-[1fr_380px]">
				<div class="space-y-6">
					<Card.Root class="border-border">
						<Card.Header class="pb-2">
							<Card.Title class="text-base">Register Terbaru</Card.Title>
							<Card.Description>Aktivitas administrasi yang paling baru masuk lintas modul TU.</Card.Description>
						</Card.Header>
						<Card.Content class="p-0">
							<div class="overflow-x-auto">
								<Table.Root>
									<Table.Header>
										<Table.Row>
											<Table.Head>Jenis</Table.Head>
											<Table.Head>Uraian</Table.Head>
											<Table.Head>Tanggal</Table.Head>
											<Table.Head>Status</Table.Head>
										</Table.Row>
									</Table.Header>
									<Table.Body>
										{#each [
											...latestByDate(data.incomingLetters, (item) => item.tanggal_terima, 4).map((item) => ({ key: `in-${item.id}`, type: 'Surat Masuk', title: item.perihal, meta: item.asal, date: item.tanggal_terima, status: item.status })),
											...latestByDate(data.outgoingLetters, (item) => item.tanggal_surat, 4).map((item) => ({ key: `out-${item.id}`, type: 'Surat Keluar', title: item.perihal, meta: item.tujuan, date: item.tanggal_surat, status: 'issued' })),
											...latestByDate(data.certificates, (item) => item.tanggal_surat, 4).map((item) => ({ key: `cert-${item.id}`, type: 'Surat Ket.', title: item.student_name, meta: item.template_name, date: item.tanggal_surat, status: item.status })),
											...latestByDate(data.archiveDocuments, (item) => item.received_date, 4).map((item) => ({ key: `archive-${item.id}`, type: 'Arsip', title: item.title, meta: archiveMeta(item), date: item.received_date, status: item.status })),
										].sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime()).slice(0, 10) as row (row.key)}
											<Table.Row>
												<Table.Cell class="whitespace-nowrap text-sm text-muted-foreground">{row.type}</Table.Cell>
												<Table.Cell>
													<p class="text-sm font-medium text-foreground">{row.title}</p>
													<p class="text-xs text-muted-foreground">{row.meta}</p>
												</Table.Cell>
												<Table.Cell class="whitespace-nowrap text-sm">{formatDate(row.date)}</Table.Cell>
												<Table.Cell><Badge variant={statusVariant(row.status)}>{statusLabel(row.status)}</Badge></Table.Cell>
											</Table.Row>
										{/each}
									</Table.Body>
								</Table.Root>
							</div>
						</Card.Content>
					</Card.Root>

					<Card.Root class="border-border">
						<Card.Header class="pb-2">
							<Card.Title class="text-base">RKT/RKJM dan Anggaran</Card.Title>
							<Card.Description>Kontrol bukti program yang perlu terlihat saat pemeriksaan administrasi.</Card.Description>
						</Card.Header>
						<Card.Content class="p-0">
							<div class="overflow-x-auto">
								<Table.Root>
									<Table.Header>
										<Table.Row>
											<Table.Head>Kegiatan</Table.Head>
											<Table.Head>PJ</Table.Head>
											<Table.Head>Anggaran</Table.Head>
											<Table.Head>Progres</Table.Head>
										</Table.Row>
									</Table.Header>
									<Table.Body>
										{#each latestByDate(data.workPlanItems, (item) => item.end_date || item.start_date, 8) as item (item.id)}
											<Table.Row>
												<Table.Cell>
													<p class="text-sm font-medium text-foreground">{item.activity_code} - {item.activity_name}</p>
													<p class="text-xs text-muted-foreground">{item.program_code} - {item.program_name}</p>
												</Table.Cell>
												<Table.Cell class="text-sm">{item.responsible_employee_name || item.owner_unit_name || '-'}</Table.Cell>
												<Table.Cell>
													<p class="text-sm">{formatCurrency(item.budget_amount)}</p>
													<p class="text-xs text-muted-foreground">Realisasi {formatCurrency(item.realization_amount)}</p>
												</Table.Cell>
												<Table.Cell>
													<div class="flex items-center gap-2">
														<div class="h-2 w-20 rounded-full bg-muted"><div class="h-2 rounded-full bg-primary" style={`width: ${item.progress_percent}%`}></div></div>
														<span class="text-xs text-muted-foreground">{item.progress_percent}%</span>
													</div>
													<Badge variant={statusVariant(item.status)}>{statusLabel(item.status)}</Badge>
												</Table.Cell>
											</Table.Row>
										{/each}
									</Table.Body>
								</Table.Root>
							</div>
						</Card.Content>
					</Card.Root>
				</div>

				<div class="space-y-6">
					<Card.Root class="border-border">
						<Card.Header class="pb-2">
							<Card.Title class="text-base">Perlu Perhatian</Card.Title>
							<Card.Description>Daftar operasional yang belum selesai atau perlu tindak lanjut.</Card.Description>
						</Card.Header>
						<Card.Content class="space-y-3">
							{#if attention.length === 0}
								<div class="rounded-md border border-primary/20 bg-primary/10 px-4 py-3 text-sm text-primary">Tidak ada perhatian utama dari data saat ini.</div>
							{:else}
								{#each attention as item (item.key)}
										<a href={resolve(item.href)} class="block rounded-md border border-border px-4 py-3 hover:bg-muted/50">
										<div class="flex items-start justify-between gap-3">
											<div>
												<p class="text-sm font-medium text-foreground">{item.title}</p>
												<p class="mt-1 text-xs text-muted-foreground">{item.meta}</p>
											</div>
											<Badge variant={item.level === 'danger' ? 'destructive' : 'secondary'}>{item.level === 'danger' ? 'Prioritas' : 'Pantau'}</Badge>
										</div>
									</a>
								{/each}
							{/if}
						</Card.Content>
					</Card.Root>

					<Card.Root class="border-border">
						<Card.Header class="pb-2">
							<Card.Title class="text-base">Inventaris dan Arsip</Card.Title>
							<Card.Description>Barang dan dokumen yang mendukung kesiapan administrasi.</Card.Description>
						</Card.Header>
						<Card.Content class="space-y-4">
							<div>
								<p class="text-xs font-medium text-muted-foreground">Inventaris Perhatian</p>
								<div class="mt-2 space-y-2">
									{#each data.inventoryItems.filter(isInventoryAttention).slice(0, 5) as item (item.id)}
										<div class="flex items-center justify-between gap-3 rounded-md border border-border px-3 py-2">
											<div>
												<p class="text-sm font-medium text-foreground">{item.nama}</p>
												<p class="text-xs text-muted-foreground">{item.lokasi || 'Lokasi belum diisi'} - {item.jumlah_baik}/{item.jumlah_total} {item.satuan}</p>
											</div>
											<Badge variant={item.kondisi === 'rusak' ? 'destructive' : 'outline'}>{inventoryAttentionLabel(item)}</Badge>
										</div>
									{/each}
								</div>
							</div>
							<div>
								<p class="text-xs font-medium text-muted-foreground">Arsip Retensi Dekat</p>
								<div class="mt-2 space-y-2">
									{#each data.archiveDocuments.filter((item) => isArchiveRetentionAttention(item)).slice(0, 5) as item (item.id)}
										<div class="rounded-md border border-border px-3 py-2">
											<p class="text-sm font-medium text-foreground">{item.title}</p>
											<p class="text-xs text-muted-foreground">{item.category_code} - retensi {formatDate(item.retention_until)}</p>
										</div>
									{/each}
								</div>
							</div>
						</Card.Content>
					</Card.Root>
				</div>
			</div>
		{/snippet}
	</AsyncContent>
</div>
