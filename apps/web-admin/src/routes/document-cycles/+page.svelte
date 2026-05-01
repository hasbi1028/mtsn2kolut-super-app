<script lang="ts">
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import BellIcon from '@lucide/svelte/icons/bell';
	import CalendarClockIcon from '@lucide/svelte/icons/calendar-clock';
	import CheckCircle2Icon from '@lucide/svelte/icons/check-circle-2';
	import ClipboardListIcon from '@lucide/svelte/icons/clipboard-list';
	import FileCheck2Icon from '@lucide/svelte/icons/file-check-2';
	import FileWarningIcon from '@lucide/svelte/icons/file-warning';
	import PencilIcon from '@lucide/svelte/icons/pencil';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import RefreshCcwIcon from '@lucide/svelte/icons/refresh-ccw';
	import RotateCwIcon from '@lucide/svelte/icons/rotate-cw';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import { base } from '$app/paths';
	import { onMount } from 'svelte';
	import type { BadgeVariant } from '$lib/components/ui/badge';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import * as Tabs from '$lib/components/ui/tabs';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { confirmAction } from '$lib/confirm-dialog';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Textarea } from '$lib/components/ui/textarea';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { toast } from '$lib/components/ui/sonner';

	interface DocumentCycleStats {
		active_catalogs: number;
		total_obligations: number;
		not_started_obligations: number;
		draft_obligations: number;
		waiting_verification_obligations: number;
		completed_obligations: number;
		overdue_obligations: number;
		due_soon_obligations: number;
		no_pic_obligations: number;
	}

	interface UnitOption {
		id: string;
		code: string;
		name: string;
	}

	interface EmployeeOption {
		id: string;
		nip: string;
		nama: string;
		unit_kerja: string;
	}

	interface DocumentOption {
		id: string;
		title: string;
		period_label?: string;
		doc_type?: string;
	}

	interface EvidenceOption {
		id: string;
		title: string;
		status?: string;
	}

	interface ArchiveOption {
		id: string;
		title: string;
		archive_number?: string;
	}

	interface DocumentCycleCatalog {
		id: string;
		code: string;
		title: string;
		frequency: string;
		snp_standard: string;
		regulation_ref: string;
		default_owner_unit_id?: string;
		default_owner_unit_name: string;
		default_responsible_employee_id?: string;
		default_responsible_employee_name: string;
		default_responsible_employee_nip: string;
		default_verifier_employee_id?: string;
		default_verifier_employee_name: string;
		default_verifier_employee_nip: string;
		deadline_days_after_period: number;
		reminder_days_before_due: number;
		description: string;
		is_active: boolean;
		sort_order: number;
	}

	interface DocumentCycleObligation {
		id: string;
		catalog_id: string;
		catalog_code: string;
		catalog_title: string;
		frequency: string;
		snp_standard: string;
		regulation_ref: string;
		period_year: number;
		period_label: string;
		period_start: string;
		period_end: string;
		due_date: string;
		reminder_date: string;
		owner_unit_id?: string;
		owner_unit_name: string;
		responsible_employee_id?: string;
		responsible_employee_name: string;
		responsible_employee_nip: string;
		verifier_employee_id?: string;
		verifier_employee_name: string;
		verifier_employee_nip: string;
		status: string;
		governance_document_id?: string;
		governance_document_title: string;
		evidence_item_id?: string;
		evidence_item_title: string;
		archive_document_id?: string;
		archive_document_title: string;
		notes: string;
		verification_notes: string;
		completed_at?: string;
		is_overdue: boolean;
		is_due_soon: boolean;
	}

	interface DashboardData {
		stats: DocumentCycleStats;
		catalogs: DocumentCycleCatalog[];
		obligations: DocumentCycleObligation[];
		units: UnitOption[];
		employees: EmployeeOption[];
		documents: DocumentOption[];
		evidenceItems: EvidenceOption[];
		archiveDocuments: ArchiveOption[];
	}

	interface CatalogForm {
		code: string;
		title: string;
		frequency: string;
		snp_standard: string;
		regulation_ref: string;
		default_owner_unit_id: string;
		default_responsible_employee_id: string;
		default_verifier_employee_id: string;
		deadline_days_after_period: number;
		reminder_days_before_due: number;
		description: string;
		is_active: boolean;
		sort_order: number;
	}

	interface ObligationForm {
		due_date: string;
		reminder_date: string;
		owner_unit_id: string;
		responsible_employee_id: string;
		verifier_employee_id: string;
		governance_document_id: string;
		evidence_item_id: string;
		archive_document_id: string;
		notes: string;
		verification_notes: string;
	}

	interface ApiErrorPayload {
		error?: string;
	}

	const FREQUENCIES: Array<[string, string]> = [
		['', 'Semua frekuensi'],
		['daily', 'Harian'],
		['weekly', 'Mingguan'],
		['monthly', 'Bulanan'],
		['quarterly', 'Triwulan'],
		['semester', 'Semester'],
		['annual', 'Tahunan'],
		['four_year', '4 Tahunan'],
		['five_year', '5 Tahunan']
	];

	const CATALOG_FREQUENCIES = FREQUENCIES.filter(([value]) => value !== '');

	const STATUSES: Array<[string, string]> = [
		['', 'Semua status'],
		['not_started', 'Belum Mulai'],
		['draft', 'Sedang Dibuat'],
		['waiting_verification', 'Menunggu Verifikasi'],
		['completed', 'Selesai']
	];

	const SNP_STANDARDS: Array<[string, string]> = [
		['', 'Tidak dipetakan'],
		['skl', 'SKL'],
		['isi', 'Standar Isi'],
		['proses', 'Standar Proses'],
		['penilaian', 'Standar Penilaian'],
		['ptk', 'PTK'],
		['sarpras', 'Sarpras'],
		['pengelolaan', 'Pengelolaan'],
		['pembiayaan', 'Pembiayaan']
	];

	let activeTab = $state('monitoring');
	let periodYear = $state(new Date().getFullYear());
	let statusFilter = $state('');
	let frequencyFilter = $state('');
	let search = $state('');
	let reminderOnly = $state(false);
	let refreshBusy = $state(false);
	let generateBusy = $state(false);
	let catalogBusy = $state(false);
	let obligationBusy = $state(false);
	let actionBusyId = $state('');
	let dashboardPromise = $state<Promise<DashboardData> | null>(null);
	let snapshot = $state<DashboardData | null>(null);
	let editingCatalogId = $state('');
	let selectedObligationId = $state('');
	let catalogForm = $state<CatalogForm>(emptyCatalogForm());
	let obligationForm = $state<ObligationForm>(emptyObligationForm());

	const selectedObligation = $derived(
		(snapshot?.obligations ?? []).find((item) => item.id === selectedObligationId)
	);

	function emptyCatalogForm(): CatalogForm {
		return {
			code: '',
			title: '',
			frequency: 'monthly',
			snp_standard: '',
			regulation_ref: '',
			default_owner_unit_id: '',
			default_responsible_employee_id: '',
			default_verifier_employee_id: '',
			deadline_days_after_period: 5,
			reminder_days_before_due: 3,
			description: '',
			is_active: true,
			sort_order: 0
		};
	}

	function emptyObligationForm(): ObligationForm {
		return {
			due_date: '',
			reminder_date: '',
			owner_unit_id: '',
			responsible_employee_id: '',
			verifier_employee_id: '',
			governance_document_id: '',
			evidence_item_id: '',
			archive_document_id: '',
			notes: '',
			verification_notes: ''
		};
	}

	function refreshData() {
		refreshBusy = true;
		const next = loadDashboard();
		dashboardPromise = next
			.then((data) => {
				snapshot = data;
				return data;
			})
			.finally(() => {
				refreshBusy = false;
			});
	}

	async function loadDashboard(): Promise<DashboardData> {
		const obligationParams = new URLSearchParams({ period_year: String(periodYear) });
		if (statusFilter) obligationParams.set('status', statusFilter);
		if (frequencyFilter) obligationParams.set('frequency', frequencyFilter);
		if (search.trim()) obligationParams.set('search', search.trim());
		if (reminderOnly) obligationParams.set('reminder_only', 'true');

		const commonYear = new URLSearchParams({ period_year: String(periodYear) });
		const [stats, catalogs, obligations, units, employees, documents, evidenceItems, archiveDocuments] = await Promise.all([
			fetchJson<DocumentCycleStats>(`/api/document-cycles/stats?period_year=${periodYear}`),
			fetchJson<DocumentCycleCatalog[]>('/api/document-cycles/catalogs?active_only=true'),
			fetchJson<DocumentCycleObligation[]>(`/api/document-cycles/obligations?${obligationParams.toString()}`),
			fetchJson<UnitOption[]>('/api/governance/units'),
			fetchJson<EmployeeOption[]>('/api/governance/employee-options'),
			fetchJson<DocumentOption[]>(`/api/governance/documents?${commonYear.toString()}`),
			fetchJson<EvidenceOption[]>(`/api/governance/evidence-items?${commonYear.toString()}`),
			fetchJson<ArchiveOption[]>('/api/tu/archives/documents')
		]);

		return { stats, catalogs, obligations, units, employees, documents, evidenceItems, archiveDocuments };
	}

	async function fetchJson<T>(path: string, init?: RequestInit): Promise<T> {
		const response = await fetch(`${base}${path}`, init);
		if (!response.ok) {
			let message = `HTTP ${response.status}`;
			try {
				const payload = (await response.json()) as ApiErrorPayload;
				message = payload.error ?? message;
			} catch {
				message = 'Backend tidak dapat dihubungi';
			}
			throw new Error(message);
		}
		return (await response.json()) as T;
	}

	async function generateYear() {
		generateBusy = true;
		try {
			const result = await fetchJson<{ period_year: number; generated: number }>('/api/document-cycles/obligations/generate-year', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ period_year: periodYear })
			});
			toast.success(`${result.generated} jadwal dokumen tahun ${result.period_year} dibuat/dipastikan`);
			refreshData();
		} catch (error) {
			toast.error(errorMessage(error));
		} finally {
			generateBusy = false;
		}
	}

	async function saveCatalog() {
		catalogBusy = true;
		try {
			const payload = {
				...catalogForm,
				deadline_days_after_period: Number(catalogForm.deadline_days_after_period),
				reminder_days_before_due: Number(catalogForm.reminder_days_before_due),
				sort_order: Number(catalogForm.sort_order)
			};
			if (editingCatalogId) {
				await fetchJson<DocumentCycleCatalog>(`/api/document-cycles/catalogs/${editingCatalogId}`, {
					method: 'PUT',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(payload)
				});
				toast.success('Katalog siklus dokumen diperbarui');
			} else {
				await fetchJson<DocumentCycleCatalog>('/api/document-cycles/catalogs', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(payload)
				});
				toast.success('Katalog siklus dokumen ditambahkan');
			}
			resetCatalogForm();
			refreshData();
		} catch (error) {
			toast.error(errorMessage(error));
		} finally {
			catalogBusy = false;
		}
	}

	async function deleteCatalog(catalog: DocumentCycleCatalog) {
		if (!(await confirmAction({
			title: 'Hapus Katalog Siklus',
			message: `Hapus katalog ${catalog.code}? Katalog yang sudah memiliki jadwal tidak bisa dihapus.`,
			confirmLabel: 'Hapus Katalog',
			tone: 'danger'
		}))) return;
		try {
			await fetchJson<null>(`/api/document-cycles/catalogs/${catalog.id}`, { method: 'DELETE' });
			toast.success('Katalog dihapus');
			refreshData();
		} catch (error) {
			toast.error(errorMessage(error));
		}
	}

	async function saveObligation() {
		if (!selectedObligation) return;
		obligationBusy = true;
		try {
			await fetchJson<DocumentCycleObligation>(`/api/document-cycles/obligations/${selectedObligation.id}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(obligationForm)
			});
			toast.success('Monitoring dokumen diperbarui');
			refreshData();
		} catch (error) {
			toast.error(errorMessage(error));
		} finally {
			obligationBusy = false;
		}
	}

	async function updateObligationStatus(obligation: DocumentCycleObligation, status: string) {
		actionBusyId = `${obligation.id}:${status}`;
		try {
			await fetchJson<DocumentCycleObligation>(`/api/document-cycles/obligations/${obligation.id}/status`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ status, notes: statusNote(status) })
			});
			toast.success(`Status ${obligation.catalog_code} menjadi ${statusLabel(status)}`);
			refreshData();
		} catch (error) {
			toast.error(errorMessage(error));
		} finally {
			actionBusyId = '';
		}
	}

	function editCatalog(catalog: DocumentCycleCatalog) {
		editingCatalogId = catalog.id;
		catalogForm = {
			code: catalog.code,
			title: catalog.title,
			frequency: catalog.frequency,
			snp_standard: catalog.snp_standard,
			regulation_ref: catalog.regulation_ref,
			default_owner_unit_id: catalog.default_owner_unit_id ?? '',
			default_responsible_employee_id: catalog.default_responsible_employee_id ?? '',
			default_verifier_employee_id: catalog.default_verifier_employee_id ?? '',
			deadline_days_after_period: catalog.deadline_days_after_period,
			reminder_days_before_due: catalog.reminder_days_before_due,
			description: catalog.description,
			is_active: catalog.is_active,
			sort_order: catalog.sort_order
		};
	}

	function resetCatalogForm() {
		editingCatalogId = '';
		catalogForm = emptyCatalogForm();
	}

	function editObligation(obligation: DocumentCycleObligation) {
		selectedObligationId = obligation.id;
		obligationForm = {
			due_date: dateInput(obligation.due_date),
			reminder_date: dateInput(obligation.reminder_date),
			owner_unit_id: obligation.owner_unit_id ?? '',
			responsible_employee_id: obligation.responsible_employee_id ?? '',
			verifier_employee_id: obligation.verifier_employee_id ?? '',
			governance_document_id: obligation.governance_document_id ?? '',
			evidence_item_id: obligation.evidence_item_id ?? '',
			archive_document_id: obligation.archive_document_id ?? '',
			notes: obligation.notes,
			verification_notes: obligation.verification_notes
		};
	}

	function attentionItems(data: DashboardData) {
		return data.obligations.filter(
			(item) => item.status === 'waiting_verification' || item.is_overdue || item.is_due_soon || !item.responsible_employee_name
		);
	}

	function statCards(data: DashboardData) {
		return [
			{
				label: 'Katalog Aktif',
				value: data.stats.active_catalogs,
				detail: 'Template siklus dokumen',
				icon: ClipboardListIcon,
				className: 'text-emerald-700'
			},
			{
				label: 'Total Jadwal',
				value: data.stats.total_obligations,
				detail: `Tahun ${periodYear}`,
				icon: CalendarClockIcon,
				className: 'text-emerald-700'
			},
			{
				label: 'Sedang Dibuat',
				value: data.stats.draft_obligations,
				detail: `${data.stats.not_started_obligations} belum mulai`,
				icon: FileWarningIcon,
				className: 'text-amber-700'
			},
			{
				label: 'Menunggu Verifikasi',
				value: data.stats.waiting_verification_obligations,
				detail: 'Perlu review kepala',
				icon: BellIcon,
				className: 'text-sky-700'
			},
			{
				label: 'Selesai',
				value: data.stats.completed_obligations,
				detail: 'Dokumen sudah ditutup',
				icon: CheckCircle2Icon,
				className: 'text-emerald-700'
			},
			{
				label: 'Lewat Tempo',
				value: data.stats.overdue_obligations,
				detail: 'Butuh tindak lanjut',
				icon: AlertTriangleIcon,
				className: 'text-red-700'
			},
			{
				label: 'Masuk Pengingat',
				value: data.stats.due_soon_obligations,
				detail: 'Reminder aktif',
				icon: BellIcon,
				className: 'text-amber-700'
			},
			{
				label: 'Belum Ada PIC',
				value: data.stats.no_pic_obligations,
				detail: 'Perlu penanggung jawab',
				icon: FileCheck2Icon,
				className: 'text-slate-700'
			}
		];
	}

	function statusNote(status: string): string {
		if (status === 'draft') return 'Dokumen mulai disusun.';
		if (status === 'waiting_verification') return 'Dokumen diajukan untuk verifikasi kepala madrasah.';
		if (status === 'completed') return 'Dokumen selesai dan siap diarsipkan.';
		return 'Status monitoring dokumen diperbarui.';
	}

	function statusLabel(status: string): string {
		return STATUSES.find(([value]) => value === status)?.[1] ?? status;
	}

	function frequencyLabel(frequency: string): string {
		return FREQUENCIES.find(([value]) => value === frequency)?.[1] ?? frequency;
	}

	function snpLabel(snp: string): string {
		return SNP_STANDARDS.find(([value]) => value === snp)?.[1] ?? 'Tidak dipetakan';
	}

	function statusVariant(status: string): BadgeVariant {
		if (status === 'completed') return 'default';
		if (status === 'waiting_verification') return 'secondary';
		if (status === 'draft') return 'outline';
		return 'ghost';
	}

	function attentionLabel(item: DocumentCycleObligation): string {
		if (item.is_overdue) return 'Lewat tempo';
		if (item.is_due_soon) return 'Masuk pengingat';
		if (item.status === 'waiting_verification') return 'Menunggu verifikasi';
		if (!item.responsible_employee_name) return 'Belum ada PIC';
		return statusLabel(item.status);
	}

	function formatDate(value?: string): string {
		if (!value) return '-';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value;
		return new Intl.DateTimeFormat('id-ID', {
			day: '2-digit',
			month: 'short',
			year: 'numeric',
			timeZone: 'Asia/Makassar'
		}).format(date);
	}

	function dateInput(value?: string): string {
		if (!value) return '';
		return value.slice(0, 10);
	}

	function errorMessage(error: unknown): string {
		return error instanceof Error ? error.message : 'Operasi gagal diproses';
	}

	function handleRenderError(error: unknown) {
		console.error('Document cycle render failed', error);
	}

	function retryDashboard(reset?: () => void) {
		reset?.();
		refreshData();
	}

	onMount(() => {
		refreshData();
	});
</script>

<svelte:head><title>Siklus Dokumen - MTsN 2 Kolaka Utara</title></svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
		<div>
			<div class="mb-2 inline-flex items-center gap-2 rounded-full border border-emerald-100 bg-emerald-50 px-3 py-1 text-xs font-medium text-emerald-800">
				<CalendarClockIcon class="size-3.5" />
				Modul Mandiri
			</div>
			<h1 class="text-xl font-semibold text-slate-900">Siklus Dokumen</h1>
			<p class="mt-1 max-w-3xl text-sm text-slate-500">
				Monitoring dokumen harian, mingguan, bulanan, SKP, Perkin, IKU, RKT, RKJM, Renstra, dan bukti 8 SNP.
			</p>
		</div>
		<div class="flex flex-wrap items-end gap-2">
			<div class="w-28">
				<label for="period-year" class="text-xs font-medium text-slate-600">Tahun</label>
				<Input id="period-year" type="number" min="2000" bind:value={periodYear} />
			</div>
			<LoadingButton variant="outline" onclick={() => void refreshData()} loading={refreshBusy} loadingLabel="Memuat">
				<RefreshCcwIcon class="mr-2 size-4" />
				Refresh
			</LoadingButton>
			<LoadingButton onclick={() => void generateYear()} loading={generateBusy} loadingLabel="Membuat">
				<RotateCwIcon class="mr-2 size-4" />
				Generate Tahun
			</LoadingButton>
		</div>
	</div>

	<AsyncContent promise={dashboardPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
				{#each Array.from({ length: 8 }) as _, index (`document-cycle-stat-skeleton-${index}`)}
					<Card.Root class="border-slate-200">
						<Card.Content class="p-4">
							<Skeleton class="h-4 w-28" />
							<Skeleton class="mt-3 h-8 w-16" />
							<Skeleton class="mt-2 h-3 w-32" />
						</Card.Content>
					</Card.Root>
				{/each}
			</div>
			<Card.Root class="border-slate-200">
				<Card.Content class="space-y-3 p-4">
					{#each Array.from({ length: 8 }) as _, index (`document-cycle-table-skeleton-${index}`)}
						<Skeleton class="h-10 w-full" />
					{/each}
				</Card.Content>
			</Card.Root>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel message={errorMessage(error)} onRetry={() => retryDashboard(reset)} />
		{/snippet}

		{#snippet children(value)}
			{@const data = value as DashboardData}
			{@const attention = attentionItems(data)}

			<div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
				{#each statCards(data) as card (card.label)}
					{@const Icon = card.icon}
					<Card.Root class="border-slate-200">
						<Card.Content class="p-4">
							<div class="flex items-start justify-between gap-3">
								<div>
									<p class="text-xs text-slate-500">{card.label}</p>
									<p class="mt-2 text-2xl font-semibold text-slate-900">{card.value}</p>
									<p class="mt-1 text-xs text-slate-500">{card.detail}</p>
								</div>
								<Icon class={`size-5 ${card.className}`} />
							</div>
						</Card.Content>
					</Card.Root>
				{/each}
			</div>

			<div class="grid gap-6 xl:grid-cols-[1fr_380px]">
				<div class="space-y-6">
					<Card.Root class="border-slate-200">
						<Card.Header class="pb-3">
							<div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
								<div>
									<Card.Title class="text-base">Monitoring Dokumen</Card.Title>
									<Card.Description>Status penyusunan, verifikasi, pengingat, dan arsip dokumen periodik.</Card.Description>
								</div>
								<div class="flex flex-wrap gap-2">
									<Input class="w-full sm:w-64" placeholder="Cari kode, nama, periode, catatan" bind:value={search} />
									<select id="status-filter" bind:value={statusFilter} class="h-9 rounded-md border border-input bg-background px-3 text-sm">
										{#each STATUSES as [value, label] (value)}
											<option {value}>{label}</option>
										{/each}
									</select>
									<select id="frequency-filter" bind:value={frequencyFilter} class="h-9 rounded-md border border-input bg-background px-3 text-sm">
										{#each FREQUENCIES as [value, label] (value)}
											<option {value}>{label}</option>
										{/each}
									</select>
									<label class="inline-flex h-9 items-center gap-2 rounded-md border border-input px-3 text-sm text-slate-700">
										<input type="checkbox" bind:checked={reminderOnly} class="size-4 accent-emerald-700" />
										Pengingat saja
									</label>
									<Button variant="outline" onclick={() => void refreshData()}>
										<RefreshCcwIcon class="mr-2 size-4" />
										Terapkan
									</Button>
								</div>
							</div>
						</Card.Header>
						<Card.Content class="p-0">
							<div class="overflow-x-auto">
								<Table.Root>
									<Table.Header>
										<Table.Row>
											<Table.Head>Dokumen</Table.Head>
											<Table.Head>Periode</Table.Head>
											<Table.Head>PIC</Table.Head>
											<Table.Head>Jadwal</Table.Head>
											<Table.Head>Status</Table.Head>
											<Table.Head>Aksi</Table.Head>
										</Table.Row>
									</Table.Header>
									<Table.Body>
										{#each data.obligations as item (item.id)}
											<Table.Row>
												<Table.Cell class="min-w-72">
													<div class="flex items-start gap-3">
														<div class="mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-md bg-emerald-50 text-emerald-700">
															<FileCheck2Icon class="size-4" />
														</div>
														<div>
															<p class="text-sm font-medium text-slate-900">{item.catalog_title}</p>
															<p class="text-xs text-slate-500">{item.catalog_code} · {frequencyLabel(item.frequency)} · {snpLabel(item.snp_standard)}</p>
															{#if item.regulation_ref}
																<p class="mt-1 text-xs text-slate-500">{item.regulation_ref}</p>
															{/if}
														</div>
													</div>
												</Table.Cell>
												<Table.Cell class="whitespace-nowrap">
													<p class="text-sm text-slate-900">{item.period_label}</p>
													<p class="text-xs text-slate-500">{formatDate(item.period_start)} - {formatDate(item.period_end)}</p>
												</Table.Cell>
												<Table.Cell class="min-w-56">
													<p class="text-sm text-slate-900">{item.responsible_employee_name || 'Belum ditentukan'}</p>
													<p class="text-xs text-slate-500">{item.owner_unit_name || 'Tanpa unit'}{item.responsible_employee_nip ? ` · ${item.responsible_employee_nip}` : ''}</p>
												</Table.Cell>
												<Table.Cell class="whitespace-nowrap">
													<p class={item.is_overdue ? 'text-sm font-medium text-red-700' : 'text-sm text-slate-900'}>Jatuh tempo {formatDate(item.due_date)}</p>
													<p class="text-xs text-slate-500">Pengingat {formatDate(item.reminder_date)}</p>
												</Table.Cell>
												<Table.Cell>
													<div class="flex flex-col gap-1">
														<Badge variant={statusVariant(item.status)}>{statusLabel(item.status)}</Badge>
														{#if item.is_overdue || item.is_due_soon}
															<Badge variant={item.is_overdue ? 'destructive' : 'outline'}>{attentionLabel(item)}</Badge>
														{/if}
													</div>
												</Table.Cell>
												<Table.Cell>
													<div class="flex min-w-72 flex-wrap gap-1.5">
														<Button size="sm" variant="outline" onclick={() => editObligation(item)}>
															<PencilIcon class="mr-2 size-3.5" />
															Detail
														</Button>
														<Button size="sm" variant="outline" disabled={actionBusyId === `${item.id}:draft`} onclick={() => void updateObligationStatus(item, 'draft')}>Draft</Button>
														<Button size="sm" variant="outline" disabled={actionBusyId === `${item.id}:waiting_verification`} onclick={() => void updateObligationStatus(item, 'waiting_verification')}>Verifikasi</Button>
														<Button size="sm" disabled={actionBusyId === `${item.id}:completed`} onclick={() => void updateObligationStatus(item, 'completed')}>Selesai</Button>
													</div>
												</Table.Cell>
											</Table.Row>
										{:else}
											<Table.Row>
												<Table.Cell colspan={6} class="py-10 text-center text-sm text-slate-500">
													Belum ada jadwal dokumen untuk filter ini. Jalankan Generate Tahun untuk membuat kewajiban dari katalog aktif.
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
					<Card.Root class="border-amber-200 bg-amber-50/40">
						<Card.Header class="pb-2">
							<div class="flex items-center gap-2">
								<BellIcon class="size-4 text-amber-700" />
								<Card.Title class="text-base text-slate-900">Perhatian Kepala Madrasah</Card.Title>
							</div>
							<Card.Description>Dokumen yang lewat tempo, masuk pengingat, menunggu verifikasi, atau belum punya PIC.</Card.Description>
						</Card.Header>
						<Card.Content class="space-y-3">
							{#each attention.slice(0, 8) as item (item.id)}
								<button type="button" class="w-full rounded-md border border-amber-100 bg-white p-3 text-left shadow-sm transition hover:border-amber-300" onclick={() => editObligation(item)}>
									<div class="flex items-start justify-between gap-3">
										<div>
											<p class="text-sm font-medium text-slate-900">{item.catalog_title}</p>
											<p class="text-xs text-slate-500">{item.period_label} · {item.responsible_employee_name || 'Belum ada PIC'}</p>
										</div>
										<Badge variant={item.is_overdue ? 'destructive' : 'outline'}>{attentionLabel(item)}</Badge>
									</div>
									<p class="mt-2 text-xs text-slate-500">Jatuh tempo {formatDate(item.due_date)}</p>
								</button>
							{:else}
								<div class="rounded-md border border-emerald-100 bg-white p-4 text-sm text-emerald-800">
									Tidak ada dokumen yang perlu perhatian khusus pada filter saat ini.
								</div>
							{/each}
						</Card.Content>
					</Card.Root>

					<Card.Root class="border-slate-200">
						<Card.Header class="pb-2">
							<Card.Title class="text-base">Detail Monitoring</Card.Title>
							<Card.Description>Pilih dokumen dari tabel untuk mengatur PIC, pengingat, tautan bukti, dan catatan verifikasi.</Card.Description>
						</Card.Header>
						<Card.Content>
							{#if selectedObligation}
								<form class="space-y-3" onsubmit={(event) => { event.preventDefault(); void saveObligation(); }}>
									<div>
										<p class="text-sm font-medium text-slate-900">{selectedObligation.catalog_title}</p>
										<p class="text-xs text-slate-500">{selectedObligation.period_label}</p>
									</div>
									<div class="grid gap-3 sm:grid-cols-2">
										<div>
											<label for="obligation-reminder" class="text-sm font-medium">Tanggal Pengingat</label>
											<Input id="obligation-reminder" type="date" bind:value={obligationForm.reminder_date} />
										</div>
										<div>
											<label for="obligation-due" class="text-sm font-medium">Jatuh Tempo</label>
											<Input id="obligation-due" type="date" bind:value={obligationForm.due_date} />
										</div>
									</div>
									<div>
										<label for="obligation-unit" class="text-sm font-medium">Unit Pemilik</label>
										<select id="obligation-unit" bind:value={obligationForm.owner_unit_id} class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm">
											<option value="">Tanpa unit</option>
											{#each data.units as unit (unit.id)}
												<option value={unit.id}>{unit.name}</option>
											{/each}
										</select>
									</div>
									<div class="grid gap-3 sm:grid-cols-2">
										<div>
											<label for="obligation-pic" class="text-sm font-medium">PIC Penyusun</label>
											<select id="obligation-pic" bind:value={obligationForm.responsible_employee_id} class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm">
												<option value="">Belum ditentukan</option>
												{#each data.employees as employee (employee.id)}
													<option value={employee.id}>{employee.nama}</option>
												{/each}
											</select>
										</div>
										<div>
											<label for="obligation-verifier" class="text-sm font-medium">Verifikator</label>
											<select id="obligation-verifier" bind:value={obligationForm.verifier_employee_id} class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm">
												<option value="">Belum ditentukan</option>
												{#each data.employees as employee (employee.id)}
													<option value={employee.id}>{employee.nama}</option>
												{/each}
											</select>
										</div>
									</div>
									<div>
										<label for="obligation-doc" class="text-sm font-medium">Dokumen Tata Kelola</label>
										<select id="obligation-doc" bind:value={obligationForm.governance_document_id} class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm">
											<option value="">Belum ditautkan</option>
											{#each data.documents as document (document.id)}
												<option value={document.id}>{document.title}</option>
											{/each}
										</select>
									</div>
									<div>
										<label for="obligation-evidence" class="text-sm font-medium">Evidence 8 SNP</label>
										<select id="obligation-evidence" bind:value={obligationForm.evidence_item_id} class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm">
											<option value="">Belum ditautkan</option>
											{#each data.evidenceItems as evidence (evidence.id)}
												<option value={evidence.id}>{evidence.title}</option>
											{/each}
										</select>
									</div>
									<div>
										<label for="obligation-archive" class="text-sm font-medium">Arsip Digital</label>
										<select id="obligation-archive" bind:value={obligationForm.archive_document_id} class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm">
											<option value="">Belum ditautkan</option>
											{#each data.archiveDocuments as archive (archive.id)}
												<option value={archive.id}>{archive.title}</option>
											{/each}
										</select>
									</div>
									<div>
										<label for="obligation-notes" class="text-sm font-medium">Catatan Penyusunan</label>
										<Textarea id="obligation-notes" rows={3} bind:value={obligationForm.notes} />
									</div>
									<div>
										<label for="obligation-verification-notes" class="text-sm font-medium">Catatan Verifikasi</label>
										<Textarea id="obligation-verification-notes" rows={3} bind:value={obligationForm.verification_notes} />
									</div>
									<LoadingButton type="submit" loading={obligationBusy} loadingLabel="Menyimpan">
										Simpan Detail
									</LoadingButton>
								</form>
							{:else}
								<div class="rounded-md border border-dashed border-slate-200 p-4 text-sm text-slate-500">
									Pilih dokumen dari tabel monitoring untuk mengubah jadwal, PIC, atau tautan bukti.
								</div>
							{/if}
						</Card.Content>
					</Card.Root>
				</div>
			</div>

			<Tabs.Root bind:value={activeTab} class="space-y-4">
				<Tabs.List>
					<Tabs.Trigger value="monitoring">Monitoring</Tabs.Trigger>
					<Tabs.Trigger value="catalog">Katalog</Tabs.Trigger>
				</Tabs.List>
				<Tabs.Content value="monitoring">
					<div class="rounded-md border border-slate-200 bg-white p-4 text-sm text-slate-600">
						Alur status modul: Belum Mulai -> Sedang Dibuat -> Menunggu Verifikasi -> Selesai. Pengingat dihitung dari tanggal pengingat dan jatuh tempo setiap kewajiban dokumen.
					</div>
				</Tabs.Content>
				<Tabs.Content value="catalog" class="space-y-6">
					<Card.Root class="border-slate-200">
						<Card.Header class="pb-2">
							<Card.Title class="text-base">{editingCatalogId ? 'Edit Katalog Siklus' : 'Tambah Katalog Siklus'}</Card.Title>
							<Card.Description>Template ini menjadi dasar generator kewajiban dokumen tahunan.</Card.Description>
						</Card.Header>
						<Card.Content>
							<form class="space-y-4" onsubmit={(event) => { event.preventDefault(); void saveCatalog(); }}>
								<div class="grid gap-3 md:grid-cols-4">
									<div>
										<label for="catalog-code" class="text-sm font-medium">Kode</label>
										<Input id="catalog-code" bind:value={catalogForm.code} />
									</div>
									<div class="md:col-span-2">
										<label for="catalog-title" class="text-sm font-medium">Nama Dokumen</label>
										<Input id="catalog-title" bind:value={catalogForm.title} />
									</div>
									<div>
										<label for="catalog-frequency" class="text-sm font-medium">Frekuensi</label>
										<select id="catalog-frequency" bind:value={catalogForm.frequency} class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm">
											{#each CATALOG_FREQUENCIES as [value, label] (value)}
												<option {value}>{label}</option>
											{/each}
										</select>
									</div>
									<div>
										<label for="catalog-snp" class="text-sm font-medium">SNP</label>
										<select id="catalog-snp" bind:value={catalogForm.snp_standard} class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm">
											{#each SNP_STANDARDS as [value, label] (value)}
												<option {value}>{label}</option>
											{/each}
										</select>
									</div>
									<div>
										<label for="catalog-deadline" class="text-sm font-medium">Jatuh Tempo + Hari</label>
										<Input id="catalog-deadline" type="number" min="0" max="365" bind:value={catalogForm.deadline_days_after_period} />
									</div>
									<div>
										<label for="catalog-reminder" class="text-sm font-medium">Pengingat - Hari</label>
										<Input id="catalog-reminder" type="number" min="0" max="60" bind:value={catalogForm.reminder_days_before_due} />
									</div>
									<div>
										<label for="catalog-order" class="text-sm font-medium">Urutan</label>
										<Input id="catalog-order" type="number" bind:value={catalogForm.sort_order} />
									</div>
									<div class="md:col-span-2">
										<label for="catalog-unit" class="text-sm font-medium">Default Unit</label>
										<select id="catalog-unit" bind:value={catalogForm.default_owner_unit_id} class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm">
											<option value="">Tanpa unit</option>
											{#each data.units as unit (unit.id)}
												<option value={unit.id}>{unit.name}</option>
											{/each}
										</select>
									</div>
									<div>
										<label for="catalog-pic" class="text-sm font-medium">Default PIC</label>
										<select id="catalog-pic" bind:value={catalogForm.default_responsible_employee_id} class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm">
											<option value="">Belum ditentukan</option>
											{#each data.employees as employee (employee.id)}
												<option value={employee.id}>{employee.nama}</option>
											{/each}
										</select>
									</div>
									<div>
										<label for="catalog-verifier" class="text-sm font-medium">Default Verifikator</label>
										<select id="catalog-verifier" bind:value={catalogForm.default_verifier_employee_id} class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm">
											<option value="">Belum ditentukan</option>
											{#each data.employees as employee (employee.id)}
												<option value={employee.id}>{employee.nama}</option>
											{/each}
										</select>
									</div>
									<div class="md:col-span-2">
										<label for="catalog-regulation" class="text-sm font-medium">Rujukan Regulasi</label>
										<Input id="catalog-regulation" bind:value={catalogForm.regulation_ref} />
									</div>
									<div class="md:col-span-2">
										<label class="mt-6 inline-flex items-center gap-2 text-sm text-slate-700">
											<input type="checkbox" bind:checked={catalogForm.is_active} class="size-4 accent-emerald-700" />
											Katalog aktif untuk generator
										</label>
									</div>
									<div class="md:col-span-4">
										<label for="catalog-description" class="text-sm font-medium">Deskripsi</label>
										<Textarea id="catalog-description" rows={3} bind:value={catalogForm.description} />
									</div>
								</div>
								<div class="flex flex-wrap gap-2">
									<LoadingButton type="submit" loading={catalogBusy} loadingLabel="Menyimpan">
										<PlusIcon class="mr-2 size-4" />
										{editingCatalogId ? 'Simpan Perubahan' : 'Tambah Katalog'}
									</LoadingButton>
									<Button type="button" variant="outline" onclick={resetCatalogForm}>Reset</Button>
								</div>
							</form>
						</Card.Content>
					</Card.Root>

					<Card.Root class="border-slate-200">
						<Card.Header class="pb-2">
							<Card.Title class="text-base">Daftar Katalog</Card.Title>
							<Card.Description>Dokumen dari siklus MTsN yang menjadi sumber jadwal monitoring.</Card.Description>
						</Card.Header>
						<Card.Content class="p-0">
							<div class="overflow-x-auto">
								<Table.Root>
									<Table.Header>
										<Table.Row>
											<Table.Head>Kode</Table.Head>
											<Table.Head>Dokumen</Table.Head>
											<Table.Head>Frekuensi</Table.Head>
											<Table.Head>Default PIC</Table.Head>
											<Table.Head>Pengingat</Table.Head>
											<Table.Head>Aksi</Table.Head>
										</Table.Row>
									</Table.Header>
									<Table.Body>
										{#each data.catalogs as catalog (catalog.id)}
											<Table.Row>
												<Table.Cell class="font-medium">{catalog.code}</Table.Cell>
												<Table.Cell class="min-w-80">
													<p class="text-sm font-medium text-slate-900">{catalog.title}</p>
													<p class="text-xs text-slate-500">{snpLabel(catalog.snp_standard)} · {catalog.regulation_ref || 'Tanpa rujukan khusus'}</p>
												</Table.Cell>
												<Table.Cell>
													<Badge variant="outline">{frequencyLabel(catalog.frequency)}</Badge>
												</Table.Cell>
												<Table.Cell>
													<p class="text-sm text-slate-900">{catalog.default_responsible_employee_name || 'Belum ditentukan'}</p>
													<p class="text-xs text-slate-500">{catalog.default_owner_unit_name || 'Tanpa unit'}</p>
												</Table.Cell>
												<Table.Cell class="text-sm">
													<p>Tempo +{catalog.deadline_days_after_period} hari</p>
													<p class="text-xs text-slate-500">Ingat -{catalog.reminder_days_before_due} hari</p>
												</Table.Cell>
												<Table.Cell>
													<div class="flex gap-2">
														<Button size="sm" variant="outline" onclick={() => editCatalog(catalog)}>
															<PencilIcon class="mr-2 size-3.5" />
															Edit
														</Button>
														<Button size="sm" variant="outline" onclick={() => void deleteCatalog(catalog)}>
															<Trash2Icon class="mr-2 size-3.5" />
															Hapus
														</Button>
													</div>
												</Table.Cell>
											</Table.Row>
										{/each}
									</Table.Body>
								</Table.Root>
							</div>
						</Card.Content>
					</Card.Root>
				</Tabs.Content>
			</Tabs.Root>
		{/snippet}
	</AsyncContent>
</div>
