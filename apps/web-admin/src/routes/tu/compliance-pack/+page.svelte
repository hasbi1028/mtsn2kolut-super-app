<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import PrinterIcon from '@lucide/svelte/icons/printer';
	import RefreshCcwIcon from '@lucide/svelte/icons/refresh-ccw';
	import { onMount } from 'svelte';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import * as Table from '$lib/components/ui/table';
	import {
		dateValue,
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
		type EvidenceItem,
		type InventoryItem,
		type LetterDisposition,
		type TUComplianceData,
		type WorkPlanItem,
	} from '$lib/tu/compliance-data';
	import { schoolAddressLine } from '$lib/school-profile';

	type BadgeVariant = 'default' | 'secondary' | 'destructive' | 'outline';
	type RegisterRow = {
		key: string;
		area: string;
		title: string;
		date?: string;
		status: string;
		note: string;
	};

	let packPromise = $state<Promise<TUComplianceData> | null>(null);
	let snapshot = $state<TUComplianceData | null>(null);
	let generatedAt = $state(new Date());
	let refreshBusy = $state(false);

	function load() {
		generatedAt = new Date();
		const promise = fetchTUComplianceData();
		packPromise = promise;
		promise.then((data) => (snapshot = data)).catch(() => {});
		return promise;
	}

	async function refreshPack() {
		refreshBusy = true;
		try {
			await load();
		} catch {
			// AsyncContent owns the visible error state for the print pack body.
		} finally {
			refreshBusy = false;
		}
	}

	function retryPack(reset?: () => void) {
		reset?.();
		load();
	}

	function packErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Paket kepatuhan TU belum dapat dimuat. Coba lagi untuk mengambil register terbaru.';
	}

	function statusVariant(value: string): BadgeVariant {
		if (['blocked', 'gap', 'disposed', 'cancelled', 'rusak'].includes(value)) return 'destructive';
		if (['done', 'selesai', 'verified', 'issued', 'active', 'baik'].includes(value)) return 'default';
		if (['borrowed', 'ditindaklanjuti', 'collected', 'in_progress', 'perawatan'].includes(value)) return 'secondary';
		return 'outline';
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

	function formatGeneratedAt(value: Date) {
		return new Intl.DateTimeFormat('id-ID', {
			timeZone: 'Asia/Makassar',
			day: '2-digit',
			month: 'long',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
		}).format(value);
	}

	function buildRegisterRows(data: TUComplianceData): RegisterRow[] {
		return [
			...latestByDate(data.incomingLetters, (item) => item.tanggal_terima, 8).map((item) => ({
				key: `incoming-${item.id}`,
				area: 'Surat Masuk',
				title: `${item.nomor_agenda || item.nomor_surat} - ${item.perihal}`,
				date: item.tanggal_terima,
				status: item.status,
				note: item.asal,
			})),
			...latestByDate(data.outgoingLetters, (item) => item.tanggal_surat, 8).map((item) => ({
				key: `outgoing-${item.id}`,
				area: 'Surat Keluar',
				title: `${item.nomor_surat} - ${item.perihal}`,
				date: item.tanggal_surat,
				status: 'issued',
				note: item.tujuan,
			})),
			...latestByDate(data.certificates, (item) => item.tanggal_surat, 6).map((item) => ({
				key: `certificate-${item.id}`,
				area: 'Surat Keterangan',
				title: `${item.nomor_surat} - ${item.student_name}`,
				date: item.tanggal_surat,
				status: item.status,
				note: item.template_name,
			})),
		]
			.sort((a, b) => dateValue(b.date) - dateValue(a.date))
			.slice(0, 18);
	}

	function evidenceNeeds(items: EvidenceItem[]) {
		return items.filter((item) => item.status === 'needed' || item.status === 'gap' || !item.evidence_url).slice(0, 10);
	}

	function printPack() {
		window.print();
	}

	function exportPackCsv() {
		if (!snapshot) return;
		const openDispositions = snapshot.dispositions.filter(isOpenDisposition);
		const archiveAttention = snapshot.archiveDocuments.filter((item) => isArchiveRetentionAttention(item));
		const inventoryAttention = snapshot.inventoryItems.filter(isInventoryAttention);
		const workPlanAttention = snapshot.workPlanItems.filter(isWorkPlanAttention);
		const rows = [
			['Bagian', 'Nomor/ID', 'Uraian', 'Tanggal', 'Status', 'Catatan'],
			...buildRegisterRows(snapshot).map((row) => [row.area, row.key, row.title, formatDate(row.date), statusLabel(row.status), row.note]),
			...openDispositions.map((item) => [
				'Disposisi Terbuka',
				item.nomor_agenda,
				item.letter_perihal,
				formatDate(item.disposed_at),
				statusLabel(item.status),
				item.assignee_name || 'Belum ada penerima',
			]),
			...archiveAttention.map((item) => [
				'Arsip Retensi',
				item.archive_number,
				item.title,
				formatDate(item.retention_until),
				statusLabel(item.status),
				item.storage_location || 'Lokasi belum diisi',
			]),
			...inventoryAttention.map((item) => [
				'Inventaris',
				item.kode,
				item.nama,
				'',
				statusLabel(item.kondisi),
				`${item.jumlah_baik}/${item.jumlah_total} ${item.satuan} - ${item.lokasi || 'Lokasi belum diisi'}`,
			]),
			...workPlanAttention.map((item) => [
				'RKT/RKJM',
				item.activity_code,
				item.activity_name,
				formatDate(item.end_date || item.start_date),
				statusLabel(item.status),
				`${item.program_code} - ${item.progress_percent}%`,
			]),
		];
		downloadCsv(`paket-kepatuhan-tu-${new Date().toISOString().slice(0, 10)}.csv`, rows);
	}

	function dispositionNote(item: LetterDisposition) {
		return item.assignee_name || item.disposed_by_name || 'Belum ada penerima';
	}

	function archiveNote(item: ArchiveDocument) {
		return `${item.category_code} - ${item.storage_location || 'Lokasi belum diisi'} - ${formatBytes(item.file_size)}`;
	}

	function inventoryNote(item: InventoryItem) {
		return `${item.lokasi || 'Lokasi belum diisi'} - layak ${item.jumlah_baik}/${item.jumlah_total} ${item.satuan}`;
	}

	function handleRenderError(error: unknown) {
		console.error('TU compliance pack render failed', error);
	}

	onMount(() => {
		void load();
	});
</script>

<svelte:head><title>Paket Kepatuhan TU - MTsN 2 Kolaka Utara</title></svelte:head>

<div class="space-y-6">
	<div class="no-print flex flex-col gap-3 rounded-md border border-border bg-card px-4 py-3 lg:flex-row lg:items-center lg:justify-between">
		<div>
			<p class="text-sm font-semibold text-foreground">Paket Kepatuhan Tata Usaha</p>
			<p class="text-sm text-muted-foreground">Register administrasi untuk pemeriksaan kepala madrasah dan staf.</p>
		</div>
		<div class="flex flex-wrap gap-2">
			<Button href="/tu" variant="outline" size="sm">
				<ArrowLeftIcon class="mr-2 size-4" />
				Dashboard TU
			</Button>
			<Button variant="outline" size="sm" onclick={exportPackCsv}>
				<DownloadIcon class="mr-2 size-4" />
				Export CSV
			</Button>
			<LoadingButton variant="outline" size="sm" onclick={() => void refreshPack()} loading={refreshBusy} loadingLabel="Memuat...">
				<RefreshCcwIcon class="mr-2 size-4" />
				Refresh
			</LoadingButton>
			<Button size="sm" onclick={printPack}>
				<PrinterIcon class="mr-2 size-4" />
				Cetak
			</Button>
		</div>
	</div>

	<AsyncContent promise={packPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<Skeleton class="h-28 w-full" />
				<div class="grid gap-3 md:grid-cols-3">
					{#each Array.from({ length: 6 }) as _, index (`tu-pack-metric-skeleton-${index}`)}
						<Skeleton class="h-24 w-full" />
					{/each}
				</div>
				<Skeleton class="h-72 w-full" />
				<Skeleton class="h-72 w-full" />
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel message={packErrorMessage(error)} onRetry={() => retryPack(reset)} />
		{/snippet}

		{#snippet children(value)}
			{@const data = value as TUComplianceData}
			{@const registerRows = buildRegisterRows(data)}
			{@const openIncoming = data.incomingLetters.filter(isOpenIncoming)}
			{@const openDispositions = data.dispositions.filter(isOpenDisposition)}
			{@const archiveAttention = data.archiveDocuments.filter((item) => isArchiveRetentionAttention(item))}
			{@const inventoryAttention = data.inventoryItems.filter(isInventoryAttention)}
			{@const workPlanAttention = data.workPlanItems.filter(isWorkPlanAttention)}
			{@const workPlan = workPlanTotals(data.workPlanItems)}
			{@const evidenceAttention = evidenceNeeds(data.evidenceItems)}
			{@const schoolProfile = data.schoolProfile}

			<main class="print-root mx-auto max-w-5xl space-y-6 bg-card text-foreground">
				<section class="print-section rounded-md border border-border bg-card p-5">
					<div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
						<div>
							<p class="text-xs font-semibold uppercase tracking-[0.16em] text-primary">{schoolProfile.ministry_line}</p>
							<h1 class="mt-2 text-2xl font-bold text-foreground">Paket Kepatuhan Tata Usaha</h1>
							<p class="mt-1 text-sm text-muted-foreground">{schoolProfile.name}</p>
							<p class="mt-1 max-w-2xl text-xs text-muted-foreground">{schoolAddressLine(schoolProfile) || schoolProfile.office_line}</p>
						</div>
						<div class="rounded-md border border-border px-4 py-3 text-sm">
							<p class="font-medium text-foreground">Waktu cetak</p>
							<p class="text-muted-foreground">{formatGeneratedAt(generatedAt)} WITA</p>
						</div>
					</div>
				</section>

				<section class="print-section rounded-md border border-border bg-card p-5">
					<h2 class="text-base font-semibold text-foreground">Ringkasan Kontrol</h2>
					<div class="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
						<div class="rounded-md border border-border p-4">
							<p class="text-xs text-muted-foreground">Surat masuk terbuka</p>
							<p class="mt-2 text-2xl font-semibold">{openIncoming.length}</p>
							<p class="mt-1 text-xs text-muted-foreground">{data.incomingLetters.length} total register</p>
						</div>
						<div class="rounded-md border border-border p-4">
							<p class="text-xs text-muted-foreground">Disposisi terbuka</p>
							<p class="mt-2 text-2xl font-semibold">{openDispositions.length}</p>
							<p class="mt-1 text-xs text-muted-foreground">{data.dispositions.length} total disposisi</p>
						</div>
						<div class="rounded-md border border-border p-4">
							<p class="text-xs text-muted-foreground">Arsip digital</p>
							<p class="mt-2 text-2xl font-semibold">{data.archiveStats.total_documents}</p>
							<p class="mt-1 text-xs text-muted-foreground">{formatBytes(data.archiveStats.total_file_size)} tersimpan</p>
						</div>
						<div class="rounded-md border border-border p-4">
							<p class="text-xs text-muted-foreground">Inventaris perhatian</p>
							<p class="mt-2 text-2xl font-semibold">{inventoryAttention.length}</p>
							<p class="mt-1 text-xs text-muted-foreground">{numberValue(data.inventoryStats.total_unit)} unit tercatat</p>
						</div>
						<div class="rounded-md border border-border p-4">
							<p class="text-xs text-muted-foreground">RKT/RKJM terkendala</p>
							<p class="mt-2 text-2xl font-semibold">{workPlan.blocked}</p>
							<p class="mt-1 text-xs text-muted-foreground">Realisasi {formatCurrency(workPlan.realization)}</p>
						</div>
						<div class="rounded-md border border-border p-4">
							<p class="text-xs text-muted-foreground">Bukti mutu perhatian</p>
							<p class="mt-2 text-2xl font-semibold">{evidenceAttention.length}</p>
							<p class="mt-1 text-xs text-muted-foreground">Anggaran {formatCurrency(workPlan.budget)}</p>
						</div>
					</div>
				</section>

				<section class="print-section rounded-md border border-border bg-card">
					<div class="border-b border-border px-5 py-4">
						<h2 class="text-base font-semibold text-foreground">Register Surat Terbaru</h2>
					</div>
					<div class="overflow-x-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Jenis</Table.Head>
									<Table.Head>Uraian</Table.Head>
									<Table.Head>Tanggal</Table.Head>
									<Table.Head>Status</Table.Head>
									<Table.Head>Catatan</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each registerRows as row (row.key)}
									<Table.Row>
										<Table.Cell class="whitespace-nowrap text-sm">{row.area}</Table.Cell>
										<Table.Cell class="min-w-64 text-sm font-medium text-foreground">{row.title}</Table.Cell>
										<Table.Cell class="whitespace-nowrap text-sm">{formatDate(row.date)}</Table.Cell>
										<Table.Cell><Badge variant={statusVariant(row.status)}>{statusLabel(row.status)}</Badge></Table.Cell>
										<Table.Cell class="text-sm text-muted-foreground">{row.note || '-'}</Table.Cell>
									</Table.Row>
								{:else}
									<Table.Row><Table.Cell colspan={5} class="text-center text-sm text-muted-foreground">Belum ada register surat.</Table.Cell></Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</div>
				</section>

				<section class="print-section rounded-md border border-border bg-card">
					<div class="border-b border-border px-5 py-4">
						<h2 class="text-base font-semibold text-foreground">Disposisi Belum Selesai</h2>
					</div>
					<div class="overflow-x-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Agenda</Table.Head>
									<Table.Head>Perihal</Table.Head>
									<Table.Head>Penerima</Table.Head>
									<Table.Head>Tanggal</Table.Head>
									<Table.Head>Status</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each latestByDate(openDispositions, (item) => item.disposed_at, 12) as item (item.id)}
									<Table.Row>
										<Table.Cell class="whitespace-nowrap text-sm">{item.nomor_agenda || '-'}</Table.Cell>
										<Table.Cell>
											<p class="text-sm font-medium text-foreground">{item.letter_perihal}</p>
											<p class="text-xs text-muted-foreground">{item.letter_asal}</p>
										</Table.Cell>
										<Table.Cell class="text-sm">{dispositionNote(item)}</Table.Cell>
										<Table.Cell class="whitespace-nowrap text-sm">{formatDate(item.disposed_at)}</Table.Cell>
										<Table.Cell><Badge variant={statusVariant(item.status)}>{statusLabel(item.status)}</Badge></Table.Cell>
									</Table.Row>
								{:else}
									<Table.Row><Table.Cell colspan={5} class="text-center text-sm text-muted-foreground">Tidak ada disposisi terbuka.</Table.Cell></Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</div>
				</section>

				<section class="print-section grid gap-6 lg:grid-cols-2">
					<div class="rounded-md border border-border bg-card">
						<div class="border-b border-border px-5 py-4">
							<h2 class="text-base font-semibold text-foreground">Arsip dan Retensi</h2>
						</div>
						<div class="overflow-x-auto">
							<Table.Root>
								<Table.Header>
									<Table.Row>
										<Table.Head>Dokumen</Table.Head>
										<Table.Head>Retensi</Table.Head>
										<Table.Head>Status</Table.Head>
									</Table.Row>
								</Table.Header>
								<Table.Body>
									{#each latestByDate(archiveAttention.length ? archiveAttention : data.archiveDocuments, (item) => item.retention_until || item.received_date, 10) as item (item.id)}
										<Table.Row>
											<Table.Cell>
												<p class="text-sm font-medium text-foreground">{item.title}</p>
												<p class="text-xs text-muted-foreground">{archiveNote(item)}</p>
											</Table.Cell>
											<Table.Cell class="whitespace-nowrap text-sm">{formatDate(item.retention_until)}</Table.Cell>
											<Table.Cell><Badge variant={statusVariant(item.status)}>{statusLabel(item.status)}</Badge></Table.Cell>
										</Table.Row>
									{:else}
										<Table.Row><Table.Cell colspan={3} class="text-center text-sm text-muted-foreground">Belum ada arsip.</Table.Cell></Table.Row>
									{/each}
								</Table.Body>
							</Table.Root>
						</div>
					</div>

					<div class="rounded-md border border-border bg-card">
						<div class="border-b border-border px-5 py-4">
							<h2 class="text-base font-semibold text-foreground">Inventaris Perhatian</h2>
						</div>
						<div class="overflow-x-auto">
							<Table.Root>
								<Table.Header>
									<Table.Row>
										<Table.Head>Barang</Table.Head>
										<Table.Head>Kondisi</Table.Head>
										<Table.Head>Catatan</Table.Head>
									</Table.Row>
								</Table.Header>
								<Table.Body>
									{#each latestByDate(inventoryAttention, (item) => item.id, 10) as item (item.id)}
										<Table.Row>
											<Table.Cell>
												<p class="text-sm font-medium text-foreground">{item.kode} - {item.nama}</p>
												<p class="text-xs text-muted-foreground">{item.kategori}</p>
											</Table.Cell>
											<Table.Cell><Badge variant={statusVariant(item.kondisi)}>{statusLabel(item.kondisi)}</Badge></Table.Cell>
											<Table.Cell class="text-sm text-muted-foreground">{inventoryNote(item)}</Table.Cell>
										</Table.Row>
									{:else}
										<Table.Row><Table.Cell colspan={3} class="text-center text-sm text-muted-foreground">Tidak ada inventaris perhatian.</Table.Cell></Table.Row>
									{/each}
								</Table.Body>
							</Table.Root>
						</div>
					</div>
				</section>

				<section class="print-section rounded-md border border-border bg-card">
					<div class="border-b border-border px-5 py-4">
						<h2 class="text-base font-semibold text-foreground">RKT/RKJM dan Bukti Mutu</h2>
					</div>
					<div class="overflow-x-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Kegiatan/Bukti</Table.Head>
									<Table.Head>SNP</Table.Head>
									<Table.Head>Unit</Table.Head>
									<Table.Head>Progres</Table.Head>
									<Table.Head>Status</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each latestByDate(workPlanAttention, (item) => item.end_date || item.start_date, 8) as item (item.id)}
									<Table.Row>
										<Table.Cell>
											<p class="text-sm font-medium text-foreground">{item.activity_code} - {item.activity_name}</p>
											<p class="text-xs text-muted-foreground">{item.program_code} - {item.program_name}</p>
										</Table.Cell>
										<Table.Cell class="text-sm">{item.program_snp_standard || '-'}</Table.Cell>
										<Table.Cell class="text-sm">{item.responsible_employee_name || item.owner_unit_name || '-'}</Table.Cell>
										<Table.Cell class="whitespace-nowrap text-sm">{item.progress_percent}%</Table.Cell>
										<Table.Cell><Badge variant={statusVariant(item.status)}>{statusLabel(item.status)}</Badge></Table.Cell>
									</Table.Row>
								{/each}
								{#each evidenceAttention as item (item.id)}
									<Table.Row>
										<Table.Cell>
											<p class="text-sm font-medium text-foreground">{item.title}</p>
											<p class="text-xs text-muted-foreground">{item.program_code || item.source_module || '-'}</p>
										</Table.Cell>
										<Table.Cell class="text-sm">{item.snp_standard || '-'}</Table.Cell>
										<Table.Cell class="text-sm">{item.owner_unit_name || '-'}</Table.Cell>
										<Table.Cell class="whitespace-nowrap text-sm">Bukti</Table.Cell>
										<Table.Cell><Badge variant={statusVariant(item.status)}>{statusLabel(item.status)}</Badge></Table.Cell>
									</Table.Row>
								{/each}
								{#if workPlanAttention.length === 0 && evidenceAttention.length === 0}
									<Table.Row><Table.Cell colspan={5} class="text-center text-sm text-muted-foreground">Tidak ada perhatian RKT/RKJM atau bukti mutu.</Table.Cell></Table.Row>
								{/if}
							</Table.Body>
						</Table.Root>
					</div>
				</section>
			</main>
		{/snippet}
	</AsyncContent>
</div>

<style>
	@media print {
		:global(body) {
			background: white;
		}

		:global(body *) {
			visibility: hidden;
		}

		.no-print {
			display: none !important;
		}

		.print-root,
		.print-root * {
			visibility: visible;
		}

		.print-root {
			position: absolute;
			left: 0;
			top: 0;
			width: 100%;
			max-width: none;
			margin: 0;
			padding: 0;
		}

		.print-section {
			break-inside: avoid;
			box-shadow: none !important;
		}
	}
</style>
