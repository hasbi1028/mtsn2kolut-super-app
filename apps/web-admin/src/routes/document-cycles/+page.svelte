<script lang="ts">
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import BellIcon from '@lucide/svelte/icons/bell';
	import CalendarClockIcon from '@lucide/svelte/icons/calendar-clock';
	import CheckCircle2Icon from '@lucide/svelte/icons/check-circle-2';
	import ClipboardListIcon from '@lucide/svelte/icons/clipboard-list';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import ExternalLinkIcon from '@lucide/svelte/icons/external-link';
	import FileCheck2Icon from '@lucide/svelte/icons/file-check-2';
	import FileWarningIcon from '@lucide/svelte/icons/file-warning';
	import FolderArchiveIcon from '@lucide/svelte/icons/folder-archive';
	import Link2Icon from '@lucide/svelte/icons/link-2';
	import NetworkIcon from '@lucide/svelte/icons/network';
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
	import { readClientJson } from '$lib/client/api';

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
		linked_archive_obligations: number;
		linked_evidence_obligations: number;
		linked_governance_document_obligations: number;
		linked_compliance_action_obligations: number;
		external_tracker_obligations: number;
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

	interface WorkPlanOption {
		id: string;
		activity_code: string;
		activity_name: string;
		status: string;
	}

	interface PerformanceTargetOption {
		id: string;
		title: string;
		employee_name: string;
		status: string;
	}

	interface ComplianceActionOption {
		id: string;
		title: string;
		status: string;
		priority: string;
		due_date?: string;
	}

	interface DocumentCycleCatalog {
		id: string;
		code: string;
		title: string;
		frequency: string;
		domain_area: string;
		external_system: string;
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
		domain_area: string;
		external_system: string;
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
		work_plan_item_id?: string;
		work_plan_item_code: string;
		work_plan_item_name: string;
		performance_target_id?: string;
		performance_target_title: string;
		evidence_item_id?: string;
		evidence_item_title: string;
		compliance_action_id?: string;
		compliance_action_title: string;
		archive_document_id?: string;
		archive_document_title: string;
		notes: string;
		verification_notes: string;
		completed_at?: string;
		is_overdue: boolean;
		is_due_soon: boolean;
	}

	interface DocumentCycleEvent {
		id: string;
		obligation_id: string;
		event_type: string;
		from_status: string;
		to_status: string;
		notes: string;
		actor_username: string;
		created_at: string;
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
		workPlanItems: WorkPlanOption[];
		performanceTargets: PerformanceTargetOption[];
		complianceActions: ComplianceActionOption[];
	}

	interface CatalogForm {
		code: string;
		title: string;
		frequency: string;
		domain_area: string;
		external_system: string;
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
		domain_area: string;
		external_system: string;
		owner_unit_id: string;
		responsible_employee_id: string;
		verifier_employee_id: string;
		governance_document_id: string;
		work_plan_item_id: string;
		performance_target_id: string;
		evidence_item_id: string;
		compliance_action_id: string;
		archive_document_id: string;
		notes: string;
		verification_notes: string;
	}

	interface ExternalTrackerRow {
		system: string;
		label: string;
		total: number;
		completed: number;
		waiting: number;
		draft: number;
		overdue: number;
		linkedEvidence: number;
		linkedArchive: number;
		status: string;
	}

	interface ConnectionScore {
		done: number;
		total: number;
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

	const DOMAIN_AREAS: Array<[string, string]> = [
		['', 'Semua bidang'],
		['tu', 'TU'],
		['kesiswaan', 'Kesiswaan'],
		['kurikulum', 'Kurikulum'],
		['sarpras', 'Sarpras'],
		['governance', 'Governance'],
		['keuangan', 'Keuangan'],
		['eksternal', 'Eksternal']
	];

	const CATALOG_DOMAIN_AREAS = DOMAIN_AREAS.filter(([value]) => value !== '');

	const EXTERNAL_SYSTEMS: Array<[string, string]> = [
		['', 'Tidak terkait portal eksternal'],
		['skp_bkn', 'SKP BKN / e-Kinerja'],
		['emis', 'EMIS'],
		['sipka', 'SIPKA'],
		['simak_bmn', 'SIMAK-BMN'],
		['rkam_bos', 'RKAM / BOS'],
		['perkin', 'Perkin'],
		['iku', 'IKU'],
		['lakip_lkj', 'LAKIP / LKj'],
		['edm', 'EDM']
	];

	const EXTERNAL_TRACKERS = EXTERNAL_SYSTEMS.filter(([value]) => value !== '');

	const STATUSES: Array<[string, string]> = [
		['', 'Semua status'],
		['not_started', 'Belum Mulai'],
		['draft', 'Sedang Dibuat'],
		['waiting_verification', 'Menunggu Verifikasi'],
		['completed', 'Selesai']
	];

	const AUDIT_EVENT_TYPES: Array<[string, string]> = [
		['', 'Semua event'],
		['generated', 'Dibuat Generator'],
		['updated', 'Detail Diperbarui'],
		['status_changed', 'Status Berubah'],
		['monitoring_note', 'Catatan Monitoring'],
		['reminder_sent', 'Reminder PIC'],
		['created', 'Dibuat']
	];

	const EXTERNAL_CHECKLIST_STATUSES: Array<[string, string]> = [
		['', 'Semua checklist'],
		['not_input', 'Belum input'],
		['in_progress', 'Sedang proses'],
		['submitted', 'Sudah input'],
		['needs_revision', 'Perlu revisi'],
		['done', 'Selesai']
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
	let domainAreaFilter = $state('');
	let externalSystemFilter = $state('');
	let externalChecklistStatusFilter = $state('');
	let auditEventTypeFilter = $state('');
	let auditActorFilter = $state('');
	let search = $state('');
	let reminderOnly = $state(false);
	let refreshBusy = $state(false);
	let generateBusy = $state(false);
	let catalogBusy = $state(false);
	let obligationBusy = $state(false);
	let actionBusyId = $state('');
	let dashboardPromise = $state<Promise<DashboardData> | null>(null);
	let eventsPromise = $state<Promise<DocumentCycleEvent[]> | null>(null);
	let snapshot = $state<DashboardData | null>(null);
	let dashboardRequestId = 0;
	let editingCatalogId = $state('');
	let selectedObligationId = $state('');
	let initialSelectedObligationId = $state('');
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
			domain_area: 'governance',
			external_system: '',
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
			domain_area: '',
			external_system: '',
			owner_unit_id: '',
			responsible_employee_id: '',
			verifier_employee_id: '',
			governance_document_id: '',
			work_plan_item_id: '',
			performance_target_id: '',
			evidence_item_id: '',
			compliance_action_id: '',
			archive_document_id: '',
			notes: '',
			verification_notes: ''
		};
	}

	function refreshData() {
		const requestId = ++dashboardRequestId;
		refreshBusy = true;
		const next = loadDashboard();
		dashboardPromise = next
			.then((data) => {
				if (requestId === dashboardRequestId) {
					snapshot = data;
					selectInitialObligation(data);
				}
				return data;
			})
			.finally(() => {
				if (requestId === dashboardRequestId) {
					refreshBusy = false;
				}
			});
	}

	async function loadDashboard(): Promise<DashboardData> {
		const obligationParams = new URLSearchParams({ period_year: String(periodYear) });
		if (statusFilter) obligationParams.set('status', statusFilter);
		if (frequencyFilter) obligationParams.set('frequency', frequencyFilter);
		if (domainAreaFilter) obligationParams.set('domain_area', domainAreaFilter);
		if (externalSystemFilter) obligationParams.set('external_system', externalSystemFilter);
		if (search.trim()) obligationParams.set('search', search.trim());
		if (reminderOnly) obligationParams.set('reminder_only', 'true');

		const commonYear = new URLSearchParams({ period_year: String(periodYear) });
		const [stats, catalogs, obligations] = await Promise.all([
			fetchJson<DocumentCycleStats>(`/api/document-cycles/stats?period_year=${periodYear}`),
			fetchJson<DocumentCycleCatalog[]>('/api/document-cycles/catalogs?active_only=true'),
			fetchJson<DocumentCycleObligation[]>(`/api/document-cycles/obligations?${obligationParams.toString()}`)
		]);

		const [units, employees, documents, evidenceItems, archiveDocuments, workPlanItems, performanceTargets, complianceActions] = await Promise.all([
			fetchOptionalJson<UnitOption[]>('/api/governance/units', [], 'unit governance'),
			fetchOptionalJson<EmployeeOption[]>('/api/governance/employee-options', [], 'opsi pegawai'),
			fetchOptionalJson<DocumentOption[]>(`/api/governance/documents?${commonYear.toString()}`, [], 'dokumen governance'),
			fetchOptionalJson<EvidenceOption[]>(`/api/governance/evidence-items?${commonYear.toString()}`, [], 'evidence governance'),
			fetchOptionalJson<ArchiveOption[]>('/api/tu/archives/documents', [], 'arsip TU'),
			fetchOptionalJson<WorkPlanOption[]>(`/api/governance/work-plan-items?${commonYear.toString()}`, [], 'RKT/RKJM'),
			fetchOptionalJson<PerformanceTargetOption[]>(`/api/governance/performance-targets?${commonYear.toString()}`, [], 'IKU/Perkin/SKP'),
			fetchOptionalJson<ComplianceActionOption[]>(`/api/governance/compliance-actions?${commonYear.toString()}`, [], 'tindak lanjut kepatuhan')
		]);

		return { stats, catalogs, obligations, units, employees, documents, evidenceItems, archiveDocuments, workPlanItems, performanceTargets, complianceActions };
	}

	async function fetchJson<T>(path: string, init?: RequestInit): Promise<T> {
		const response = await fetch(`${base}${path}`, init);
		return readClientJson<T>(response);
	}

	async function fetchOptionalJson<T>(path: string, fallback: T, label: string): Promise<T> {
		try {
			return await fetchJson<T>(path);
		} catch (error) {
			console.warn(`Data opsional siklus dokumen gagal dimuat: ${label}`, error);
			return fallback;
		}
	}

	function loadEvents(obligationId: string) {
		const params = new URLSearchParams();
		if (auditEventTypeFilter) params.set('event_type', auditEventTypeFilter);
		if (auditActorFilter.trim()) params.set('actor', auditActorFilter.trim());
		const suffix = params.toString() ? `?${params.toString()}` : '';
		eventsPromise = fetchJson<DocumentCycleEvent[]>(`/api/document-cycles/obligations/${obligationId}/events${suffix}`);
	}

	function retryEvents(reset?: () => void) {
		reset?.();
		if (selectedObligationId) loadEvents(selectedObligationId);
	}

	function applyAuditFilters(reset?: () => void) {
		reset?.();
		if (selectedObligationId) loadEvents(selectedObligationId);
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
			loadEvents(selectedObligation.id);
			refreshData();
		} catch (error) {
			toast.error(errorMessage(error));
		} finally {
			obligationBusy = false;
		}
	}

	async function updateObligationStatus(obligation: DocumentCycleObligation, status: string) {
		const gaps = status === 'completed' ? completionIssues(obligation) : [];
		if (gaps.length > 0) {
			toast.error(`Belum bisa diselesaikan. Lengkapi ${gaps.join(', ')}.`);
			return;
		}
		actionBusyId = `${obligation.id}:${status}`;
		try {
			await fetchJson<DocumentCycleObligation>(`/api/document-cycles/obligations/${obligation.id}/status`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ status, notes: statusNote(status) })
			});
			toast.success(`Status ${obligation.catalog_code} menjadi ${statusLabel(status)}`);
			if (selectedObligationId === obligation.id) loadEvents(obligation.id);
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
			domain_area: catalog.domain_area,
			external_system: catalog.external_system,
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
		loadEvents(obligation.id);
		obligationForm = {
			due_date: dateInput(obligation.due_date),
			reminder_date: dateInput(obligation.reminder_date),
			domain_area: obligation.domain_area,
			external_system: obligation.external_system,
			owner_unit_id: obligation.owner_unit_id ?? '',
			responsible_employee_id: obligation.responsible_employee_id ?? '',
			verifier_employee_id: obligation.verifier_employee_id ?? '',
			governance_document_id: obligation.governance_document_id ?? '',
			work_plan_item_id: obligation.work_plan_item_id ?? '',
			performance_target_id: obligation.performance_target_id ?? '',
			evidence_item_id: obligation.evidence_item_id ?? '',
			compliance_action_id: obligation.compliance_action_id ?? '',
			archive_document_id: obligation.archive_document_id ?? '',
			notes: obligation.notes,
			verification_notes: obligation.verification_notes
		};
	}

	function selectObligation(obligation: DocumentCycleObligation, focusDetail = true) {
		editObligation(obligation);
		updateSelectedObligationUrl(obligation.id);
		if (focusDetail) {
			window.setTimeout(() => {
				document.getElementById('document-cycle-detail-panel')?.scrollIntoView({ behavior: 'smooth', block: 'start' });
			}, 0);
		}
	}

	function selectInitialObligation(data: DashboardData) {
		if (!initialSelectedObligationId || selectedObligationId) return;
		const requested = data.obligations.find((item) => item.id === initialSelectedObligationId);
		if (!requested) return;
		initialSelectedObligationId = '';
		editObligation(requested);
	}

	function updateSelectedObligationUrl(obligationId: string) {
		const url = new URL(window.location.href);
		url.searchParams.set('selected_obligation', obligationId);
		window.history.replaceState({}, '', url);
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
				detail: `${data.stats.external_tracker_obligations} jadwal eksternal`,
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
				detail: `${data.stats.linked_archive_obligations} sudah berarsip`,
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
				detail: `${data.stats.linked_evidence_obligations} punya bukti SNP`,
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

	function draftActionLabel(status: string): string {
		if (status === 'not_started') return 'Mulai Draft';
		return 'Koreksi Draft';
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

	function domainAreaLabel(area: string): string {
		return DOMAIN_AREAS.find(([value]) => value === area)?.[1] ?? area;
	}

	function externalSystemLabel(system: string): string {
		return EXTERNAL_SYSTEMS.find(([value]) => value === system)?.[1] ?? 'Tidak terkait portal eksternal';
	}

	function statusVariant(status: string): BadgeVariant {
		if (status === 'completed') return 'default';
		if (status === 'waiting_verification') return 'secondary';
		if (status === 'draft') return 'outline';
		return 'ghost';
	}

	function trackerStatusLabel(status: string): string {
		if (status === 'not_input') return 'Belum input';
		if (status === 'in_progress') return 'Sedang proses';
		if (status === 'submitted') return 'Sudah input';
		if (status === 'needs_revision') return 'Perlu revisi';
		return 'Selesai';
	}

	function trackerStatusVariant(status: string): BadgeVariant {
		if (status === 'done') return 'default';
		if (status === 'needs_revision') return 'destructive';
		if (status === 'submitted') return 'secondary';
		if (status === 'in_progress') return 'outline';
		return 'ghost';
	}

	function eventTypeLabel(type: string): string {
		if (type === 'generated') return 'Dibuat Generator';
		if (type === 'updated') return 'Detail Diperbarui';
		if (type === 'status_changed') return 'Status Berubah';
		if (type === 'monitoring_note') return 'Catatan Monitoring';
		if (type === 'reminder_sent') return 'Reminder PIC';
		if (type === 'created') return 'Dibuat';
		return type;
	}

	function eventVariant(type: string): BadgeVariant {
		if (type === 'status_changed') return 'secondary';
		if (type === 'generated' || type === 'created' || type === 'reminder_sent') return 'outline';
		return 'default';
	}

	function completionIssues(item: DocumentCycleObligation): string[] {
		const issues: string[] = [];
		if (!item.responsible_employee_id) issues.push('PIC penyusun');
		if (!item.verifier_employee_id) issues.push('verifikator');
		if (!item.archive_document_id) issues.push('arsip digital');
		return issues;
	}

	function canTransitionStatus(item: DocumentCycleObligation, status: string): boolean {
		if (item.status === status) return false;
		if (item.status === 'not_started') return status === 'draft';
		if (item.status === 'draft') return status === 'not_started' || status === 'waiting_verification';
		if (item.status === 'waiting_verification') return status === 'draft' || status === 'completed';
		if (item.status === 'completed') return status === 'draft';
		return false;
	}

	function connectionScore(item: DocumentCycleObligation): ConnectionScore {
		const checks = [
			Boolean(item.governance_document_id),
			Boolean(item.work_plan_item_id),
			Boolean(item.performance_target_id),
			Boolean(item.evidence_item_id),
			Boolean(item.compliance_action_id),
			Boolean(item.archive_document_id)
		];
		return { done: checks.filter(Boolean).length, total: checks.length };
	}

	function linkedEvidenceCount(data: DashboardData): number {
		return data.obligations.filter((item) => item.evidence_item_id || item.archive_document_id).length;
	}

	function externalTrackerRows(data: DashboardData): ExternalTrackerRow[] {
		return EXTERNAL_TRACKERS.map(([system, label]) => {
			const items = data.obligations.filter((item) => item.external_system === system);
			const completed = items.filter((item) => item.status === 'completed').length;
			const waiting = items.filter((item) => item.status === 'waiting_verification').length;
			const draft = items.filter((item) => item.status === 'draft').length;
			const overdue = items.filter((item) => item.is_overdue).length;
			const linkedEvidence = items.filter((item) => item.evidence_item_id).length;
			const linkedArchive = items.filter((item) => item.archive_document_id).length;
			let status = 'not_input';
			if (items.length > 0 && completed === items.length) status = 'done';
			else if (overdue > 0) status = 'needs_revision';
			else if (waiting > 0) status = 'submitted';
			else if (draft > 0 || linkedEvidence > 0 || linkedArchive > 0) status = 'in_progress';
			return { system, label, total: items.length, completed, waiting, draft, overdue, linkedEvidence, linkedArchive, status };
		});
	}

	function externalChecklistItems(data: DashboardData): DocumentCycleObligation[] {
		return data.obligations.filter((item) => item.external_system);
	}

	function filteredExternalChecklistItems(data: DashboardData): DocumentCycleObligation[] {
		const items = externalChecklistItems(data);
		if (!externalChecklistStatusFilter) return items;
		return items.filter((item) => obligationTrackerStatus(item) === externalChecklistStatusFilter);
	}

	function obligationTrackerStatus(item: DocumentCycleObligation): string {
		if (item.status === 'completed') return 'done';
		if (item.is_overdue) return 'needs_revision';
		if (item.status === 'waiting_verification') return 'submitted';
		if (item.status === 'draft' || item.evidence_item_id || item.archive_document_id || item.compliance_action_id) return 'in_progress';
		return 'not_input';
	}

	function focusExternalTracker(system: string) {
		externalSystemFilter = system;
		externalChecklistStatusFilter = '';
		activeTab = 'external';
		refreshData();
	}

	function csvEscape(value: string | number | boolean | null | undefined) {
		const text = String(value ?? '');
		if (/[",\n]/.test(text)) return `"${text.replaceAll('"', '""')}"`;
		return text;
	}

	function downloadCsv(filename: string, rows: Array<Array<string | number | boolean | null | undefined>>) {
		const csv = rows.map((row) => row.map(csvEscape).join(',')).join('\n');
		const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
		const url = URL.createObjectURL(blob);
		const link = document.createElement('a');
		link.href = url;
		link.download = filename;
		link.click();
		URL.revokeObjectURL(url);
	}

	function exportExternalChecklistCsv(items: DocumentCycleObligation[]) {
		const rows = [
			[
				'Portal',
				'Status Checklist',
				'Status Siklus Dokumen',
				'Kode Dokumen',
				'Nama Dokumen',
				'Periode',
				'Bidang',
				'SNP',
				'Regulasi',
				'PIC',
				'Verifikator',
				'Jatuh Tempo',
				'Evidence',
				'Arsip',
				'Tindak Lanjut',
				'Catatan Verifikasi'
			],
			...items.map((item) => [
				externalSystemLabel(item.external_system),
				trackerStatusLabel(obligationTrackerStatus(item)),
				statusLabel(item.status),
				item.catalog_code,
				item.catalog_title,
				item.period_label,
				domainAreaLabel(item.domain_area),
				snpLabel(item.snp_standard),
				item.regulation_ref,
				item.responsible_employee_name || 'Belum ada PIC',
				item.verifier_employee_name || 'Belum ada verifikator',
				formatDate(item.due_date),
				item.evidence_item_title,
				item.archive_document_title,
				item.compliance_action_title,
				item.verification_notes
			])
		];
		downloadCsv(`checklist-kepatuhan-eksternal-${periodYear}.csv`, rows);
		toast.success('Checklist kepatuhan eksternal diekspor');
	}

	function exportAuditCsv(obligation: DocumentCycleObligation, events: DocumentCycleEvent[]) {
		const rows = [
			['Waktu', 'Jenis Event', 'Aktor', 'Status Awal', 'Status Akhir', 'Catatan'],
			...events.map((event) => [
				formatDateTime(event.created_at),
				eventTypeLabel(event.event_type),
				event.actor_username || 'Sistem',
				event.from_status ? statusLabel(event.from_status) : '',
				event.to_status ? statusLabel(event.to_status) : '',
				event.notes
			])
		];
		downloadCsv(`audit-siklus-${obligation.catalog_code}-${obligation.period_year}.csv`, rows);
		toast.success('Riwayat audit diekspor');
	}

	function exportSelectedAuditCsv(events: DocumentCycleEvent[]) {
		if (!selectedObligation) return;
		exportAuditCsv(selectedObligation, events);
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

	function formatDateTime(value?: string): string {
		if (!value) return '-';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value;
		return new Intl.DateTimeFormat('id-ID', {
			day: '2-digit',
			month: 'short',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
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

	function applyInitialFiltersFromUrl() {
		const params = new URLSearchParams(window.location.search);
		const requestedYear = Number(params.get('period_year'));
		if (Number.isInteger(requestedYear) && requestedYear >= 2000 && requestedYear <= 2100) {
			periodYear = requestedYear;
		}

		const requestedStatus = params.get('status') ?? '';
		if (STATUSES.some(([value]) => value === requestedStatus)) statusFilter = requestedStatus;

		const requestedFrequency = params.get('frequency') ?? '';
		if (FREQUENCIES.some(([value]) => value === requestedFrequency)) frequencyFilter = requestedFrequency;

		const requestedDomain = params.get('domain_area') ?? '';
		if (DOMAIN_AREAS.some(([value]) => value === requestedDomain)) domainAreaFilter = requestedDomain;

		const requestedExternal = params.get('external_system') ?? '';
		if (EXTERNAL_SYSTEMS.some(([value]) => value === requestedExternal)) externalSystemFilter = requestedExternal;

		const requestedExternalStatus = params.get('external_status') ?? '';
		if (EXTERNAL_CHECKLIST_STATUSES.some(([value]) => value === requestedExternalStatus)) externalChecklistStatusFilter = requestedExternalStatus;

		const requestedTab = params.get('tab') ?? '';
		if (['monitoring', 'external', 'connections', 'catalog'].includes(requestedTab)) activeTab = requestedTab;

		const requestedSearch = params.get('search')?.trim() ?? '';
		if (requestedSearch) search = requestedSearch;

		const requestedReminderOnly = (params.get('reminder_only') ?? '').toLowerCase();
		reminderOnly = requestedReminderOnly === 'true' || requestedReminderOnly === '1';

		initialSelectedObligationId = params.get('selected_obligation') ?? '';
	}

	onMount(() => {
		applyInitialFiltersFromUrl();
		refreshData();
	});
</script>

<svelte:head><title>Siklus Dokumen - MTsN 2 Kolaka Utara</title></svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
		<div>
			<div class="mb-2 inline-flex items-center gap-2 rounded-full border border-emerald-100 bg-emerald-50 px-3 py-1 text-xs font-medium text-emerald-800">
				<CalendarClockIcon class="size-3.5" />
				Radar Dokumen
			</div>
			<h1 class="text-xl font-semibold text-slate-900">Siklus Dokumen Madrasah</h1>
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
									<Input class="w-full sm:w-64" aria-label="Cari dokumen siklus" placeholder="Cari kode, nama, periode, catatan" bind:value={search} />
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
									<select id="domain-filter" bind:value={domainAreaFilter} class="h-9 rounded-md border border-input bg-background px-3 text-sm">
										{#each DOMAIN_AREAS as [value, label] (value)}
											<option {value}>{label}</option>
										{/each}
									</select>
									<select id="external-filter" bind:value={externalSystemFilter} class="h-9 rounded-md border border-input bg-background px-3 text-sm">
										<option value="">Semua tracker</option>
										{#each EXTERNAL_TRACKERS as [value, label] (value)}
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
											{@const completionGaps = completionIssues(item)}
											<Table.Row>
												<Table.Cell class="min-w-72">
													<div class="flex items-start gap-3">
														<div class="mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-md bg-emerald-50 text-emerald-700">
															<FileCheck2Icon class="size-4" />
														</div>
														<div>
															<p class="text-sm font-medium text-slate-900">{item.catalog_title}</p>
															<p class="text-xs text-slate-500">{item.catalog_code} · {domainAreaLabel(item.domain_area)} · {frequencyLabel(item.frequency)} · {snpLabel(item.snp_standard)}</p>
															{#if item.regulation_ref}
																<p class="mt-1 text-xs text-slate-500">{item.regulation_ref}</p>
															{/if}
															{#if item.external_system}
																<Badge variant="outline" class="mt-2">{externalSystemLabel(item.external_system)}</Badge>
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
														<Button size="sm" variant="outline" onclick={() => selectObligation(item)}>
															<PencilIcon class="mr-2 size-3.5" />
															Detail
														</Button>
														<Button size="sm" variant="outline" disabled={actionBusyId === `${item.id}:draft` || !canTransitionStatus(item, 'draft')} onclick={() => void updateObligationStatus(item, 'draft')}>Draft</Button>
														<Button size="sm" variant="outline" disabled={actionBusyId === `${item.id}:waiting_verification` || !canTransitionStatus(item, 'waiting_verification')} onclick={() => void updateObligationStatus(item, 'waiting_verification')}>Verifikasi</Button>
														<Button
															size="sm"
															disabled={actionBusyId === `${item.id}:completed` || !canTransitionStatus(item, 'completed') || completionGaps.length > 0}
															title={completionGaps.length > 0 ? `Lengkapi ${completionGaps.join(', ')}` : 'Tandai selesai'}
															onclick={() => void updateObligationStatus(item, 'completed')}
														>Selesai</Button>
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
								<button type="button" class="w-full rounded-md border border-amber-100 bg-white p-3 text-left shadow-sm transition hover:border-amber-300" onclick={() => selectObligation(item)}>
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
							<div class="flex items-center gap-2">
								<ExternalLinkIcon class="size-4 text-emerald-700" />
								<Card.Title class="text-base">Tracker Kepatuhan Eksternal</Card.Title>
							</div>
							<Card.Description>Checklist internal untuk portal resmi; status dan bukti disimpan di madrasah.</Card.Description>
						</Card.Header>
						<Card.Content class="space-y-3">
							{#each externalTrackerRows(data) as tracker (tracker.system)}
								<div class="rounded-md border border-slate-200 bg-white p-3">
									<div class="flex items-start justify-between gap-3">
										<div>
											<p class="text-sm font-medium text-slate-900">{tracker.label}</p>
											<p class="text-xs text-slate-500">
												{tracker.total} jadwal · {tracker.completed} selesai · {tracker.linkedEvidence + tracker.linkedArchive} bukti
											</p>
										</div>
										<Badge variant={trackerStatusVariant(tracker.status)}>{trackerStatusLabel(tracker.status)}</Badge>
									</div>
									{#if tracker.overdue > 0}
										<p class="mt-2 text-xs font-medium text-red-700">{tracker.overdue} lewat tempo/perlu revisi</p>
									{/if}
									<Button class="mt-3" type="button" variant="outline" size="sm" disabled={tracker.total === 0} onclick={() => focusExternalTracker(tracker.system)}>
										<ClipboardListIcon class="mr-2 size-3.5" />
										Checklist
									</Button>
								</div>
							{/each}
						</Card.Content>
					</Card.Root>

					<div id="document-cycle-detail-panel">
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
									<div class="grid gap-3 rounded-md border border-slate-200 bg-white p-3 text-sm sm:grid-cols-2">
										<div>
											<p class="text-xs font-medium text-slate-500">Status</p>
											<Badge variant={statusVariant(selectedObligation.status)} class="mt-1">{statusLabel(selectedObligation.status)}</Badge>
										</div>
										<div>
											<p class="text-xs font-medium text-slate-500">Periode</p>
											<p class="mt-1 text-slate-800">{selectedObligation.period_label} · {formatDate(selectedObligation.period_start)} - {formatDate(selectedObligation.period_end)}</p>
										</div>
										<div>
											<p class="text-xs font-medium text-slate-500">PIC Penyusun</p>
											<p class="mt-1 text-slate-800">{selectedObligation.responsible_employee_name || 'Belum ditentukan'}</p>
										</div>
										<div>
											<p class="text-xs font-medium text-slate-500">Verifikator</p>
											<p class="mt-1 text-slate-800">{selectedObligation.verifier_employee_name || 'Belum ditentukan'}</p>
										</div>
										<div>
											<p class="text-xs font-medium text-slate-500">SNP / Regulasi</p>
											<p class="mt-1 text-slate-800">{snpLabel(selectedObligation.snp_standard)} · {selectedObligation.regulation_ref || 'Tanpa rujukan khusus'}</p>
										</div>
										<div>
											<p class="text-xs font-medium text-slate-500">Arsip Resmi</p>
											<p class="mt-1 text-slate-800">{selectedObligation.archive_document_title || 'Belum ditautkan'}</p>
										</div>
										<div class="sm:col-span-2">
											<p class="text-xs font-medium text-slate-500">Catatan Verifikasi</p>
											<p class="mt-1 text-slate-800">{selectedObligation.verification_notes || 'Belum ada catatan verifikasi'}</p>
										</div>
									</div>
									{#if completionIssues(selectedObligation).length > 0}
										<div class="rounded-md border border-amber-200 bg-amber-50 p-3 text-sm text-amber-900">
											<p class="font-medium">Belum siap ditandai selesai.</p>
											<p class="mt-1">Lengkapi {completionIssues(selectedObligation).join(', ')} lalu simpan detail sebelum finalisasi.</p>
										</div>
									{:else}
										<div class="rounded-md border border-emerald-100 bg-emerald-50 p-3 text-sm text-emerald-800">
											Dokumen sudah memiliki PIC, verifikator, dan tautan bukti yang bisa ditelusuri.
										</div>
									{/if}
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
									<div class="grid gap-3 sm:grid-cols-2">
										<div>
											<label for="obligation-domain" class="text-sm font-medium">Bidang</label>
											<select id="obligation-domain" bind:value={obligationForm.domain_area} class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm">
												{#each CATALOG_DOMAIN_AREAS as [value, label] (value)}
													<option {value}>{label}</option>
												{/each}
											</select>
										</div>
										<div>
											<label for="obligation-external-system" class="text-sm font-medium">Tracker Eksternal</label>
											<select id="obligation-external-system" bind:value={obligationForm.external_system} class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm">
												{#each EXTERNAL_SYSTEMS as [value, label] (value)}
													<option {value}>{label}</option>
												{/each}
											</select>
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
										<label for="obligation-work-plan" class="text-sm font-medium">RKT/RKJM/RKAM</label>
										<select id="obligation-work-plan" bind:value={obligationForm.work_plan_item_id} class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm">
											<option value="">Belum ditautkan</option>
											{#each data.workPlanItems as item (item.id)}
												<option value={item.id}>{item.activity_code} - {item.activity_name}</option>
											{/each}
										</select>
									</div>
									<div>
										<label for="obligation-performance" class="text-sm font-medium">SKP / Target Kinerja</label>
										<select id="obligation-performance" bind:value={obligationForm.performance_target_id} class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm">
											<option value="">Belum ditautkan</option>
											{#each data.performanceTargets as target (target.id)}
												<option value={target.id}>{target.title} · {target.employee_name}</option>
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
										<label for="obligation-compliance-action" class="text-sm font-medium">Tindak Lanjut Kepatuhan</label>
										<select id="obligation-compliance-action" bind:value={obligationForm.compliance_action_id} class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm">
											<option value="">Belum ditautkan</option>
											{#each data.complianceActions as action (action.id)}
												<option value={action.id}>{action.title}</option>
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
									<div class="rounded-md border border-emerald-100 bg-emerald-50/50 p-3">
										<div class="mb-3">
											<p class="text-xs font-medium text-emerald-900">Aksi Cepat</p>
											<p class="text-xs text-emerald-800">Status saat ini: {statusLabel(selectedObligation.status)}</p>
										</div>
										<div class="flex flex-wrap gap-2">
											<Button href="/governance" variant="outline" size="sm">
												<Link2Icon class="mr-2 size-3.5" />
												Tambah Bukti
											</Button>
											<Button href="/tu/arsip" variant="outline" size="sm">
												<FolderArchiveIcon class="mr-2 size-3.5" />
												Tautkan Arsip
											</Button>
											{#if canTransitionStatus(selectedObligation, 'not_started')}
												<Button
													type="button"
													variant="outline"
													size="sm"
													disabled={actionBusyId === `${selectedObligation.id}:not_started`}
													onclick={() => void updateObligationStatus(selectedObligation, 'not_started')}
												>
													<RotateCwIcon class="mr-2 size-3.5" />
													Kembalikan Belum Mulai
												</Button>
											{/if}
											{#if canTransitionStatus(selectedObligation, 'draft')}
												<Button
													type="button"
													variant="outline"
													size="sm"
													disabled={actionBusyId === `${selectedObligation.id}:draft`}
													onclick={() => void updateObligationStatus(selectedObligation, 'draft')}
												>
													<FileWarningIcon class="mr-2 size-3.5" />
													{draftActionLabel(selectedObligation.status)}
												</Button>
											{/if}
											{#if canTransitionStatus(selectedObligation, 'waiting_verification')}
												<Button
													type="button"
													variant="outline"
													size="sm"
													disabled={actionBusyId === `${selectedObligation.id}:waiting_verification`}
													onclick={() => void updateObligationStatus(selectedObligation, 'waiting_verification')}
												>
													<BellIcon class="mr-2 size-3.5" />
													Ajukan Verifikasi
												</Button>
											{/if}
											{#if canTransitionStatus(selectedObligation, 'completed')}
												<Button
													type="button"
													size="sm"
													disabled={actionBusyId === `${selectedObligation.id}:completed` || completionIssues(selectedObligation).length > 0}
													title={completionIssues(selectedObligation).length > 0 ? `Lengkapi ${completionIssues(selectedObligation).join(', ')}` : 'Tandai selesai'}
													onclick={() => void updateObligationStatus(selectedObligation, 'completed')}
												>
													<CheckCircle2Icon class="mr-2 size-3.5" />
													Tandai Selesai
												</Button>
											{/if}
										</div>
									</div>
									<div class="rounded-md border border-slate-200 bg-slate-50/70 p-3">
										<div class="mb-3 flex flex-col gap-3">
											<div>
												<p class="text-sm font-medium text-slate-900">Riwayat Audit</p>
												<p class="text-xs text-slate-500">Jejak perubahan status, update detail, generator, dan reminder PIC.</p>
											</div>
											<div class="grid gap-2 sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto_auto] sm:items-end">
												<div>
													<label for="audit-event-type-filter" class="text-xs font-medium text-slate-600">Jenis Event</label>
													<select id="audit-event-type-filter" bind:value={auditEventTypeFilter} class="mt-1 h-9 w-full rounded-md border border-input bg-background px-3 text-sm">
														{#each AUDIT_EVENT_TYPES as [value, label] (value)}
															<option {value}>{label}</option>
														{/each}
													</select>
												</div>
												<div>
													<label for="audit-actor-filter" class="text-xs font-medium text-slate-600">Aktor</label>
													<Input id="audit-actor-filter" class="mt-1" placeholder="username atau system" bind:value={auditActorFilter} />
												</div>
												<Button type="button" variant="outline" size="sm" onclick={() => applyAuditFilters()}>
													<RefreshCcwIcon class="mr-2 size-3.5" />
													Terapkan
												</Button>
												<Button type="button" variant="outline" size="sm" onclick={() => { auditEventTypeFilter = ''; auditActorFilter = ''; applyAuditFilters(); }}>
													Reset
												</Button>
											</div>
										</div>
										<AsyncContent promise={eventsPromise} onerror={handleRenderError}>
											{#snippet pending()}
												<div class="space-y-2">
													{#each Array.from({ length: 3 }) as _, index (`document-cycle-event-skeleton-${index}`)}
														<div class="rounded-md border border-slate-200 bg-white p-3">
															<Skeleton class="h-4 w-32" />
															<Skeleton class="mt-2 h-3 w-full" />
														</div>
													{/each}
												</div>
											{/snippet}
											{#snippet failed(error, reset)}
												<RecoveryPanel compact title="Riwayat Audit Belum Tersaji" message={errorMessage(error)} onRetry={() => retryEvents(reset)} />
											{/snippet}
											{#snippet children(events)}
												{@const currentEvents = events as DocumentCycleEvent[]}
												<div class="mb-3 flex items-center justify-between gap-3">
													<p class="text-xs text-slate-500">{currentEvents.length} event audit ditampilkan</p>
													<Button type="button" variant="outline" size="sm" disabled={currentEvents.length === 0} onclick={() => exportSelectedAuditCsv(currentEvents)}>
														<DownloadIcon class="mr-2 size-3.5" />
														Export CSV
													</Button>
												</div>
												{#if currentEvents.length === 0}
													<div class="rounded-md border border-dashed border-slate-200 bg-white p-3 text-sm text-slate-500">
														Belum ada riwayat audit untuk dokumen ini.
													</div>
												{:else}
													<div class="max-h-80 space-y-2 overflow-y-auto pr-1">
														{#each currentEvents as event (event.id)}
															<div class="rounded-md border border-slate-200 bg-white p-3">
																<div class="flex items-start justify-between gap-3">
																	<Badge variant={eventVariant(event.event_type)}>{eventTypeLabel(event.event_type)}</Badge>
																	<p class="text-xs text-slate-500">{formatDateTime(event.created_at)}</p>
																</div>
																{#if event.from_status || event.to_status}
																	<p class="mt-2 text-xs text-slate-600">
																		{event.from_status ? statusLabel(event.from_status) : '-'} -> {event.to_status ? statusLabel(event.to_status) : '-'}
																	</p>
																{/if}
																<p class="mt-2 text-sm text-slate-700">{event.notes || 'Tanpa catatan.'}</p>
																<p class="mt-2 text-xs text-slate-500">{event.actor_username ? `oleh ${event.actor_username}` : 'oleh sistem'}</p>
															</div>
														{/each}
													</div>
												{/if}
											{/snippet}
										</AsyncContent>
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
			</div>

			<Tabs.Root bind:value={activeTab} class="space-y-4">
				<Tabs.List>
					<Tabs.Trigger value="monitoring">Monitoring</Tabs.Trigger>
					<Tabs.Trigger value="external">Kepatuhan Eksternal</Tabs.Trigger>
					<Tabs.Trigger value="connections">Peta Keterhubungan</Tabs.Trigger>
					<Tabs.Trigger value="catalog">Katalog</Tabs.Trigger>
				</Tabs.List>
				<Tabs.Content value="monitoring">
					<div class="rounded-md border border-slate-200 bg-white p-4 text-sm text-slate-600">
						Alur status modul: Belum Mulai -> Sedang Dibuat -> Menunggu Verifikasi -> Selesai. Tahun {periodYear} memiliki {linkedEvidenceCount(data)} jadwal dengan bukti atau arsip tertaut.
					</div>
				</Tabs.Content>
				<Tabs.Content value="external" class="space-y-4">
					{@const allExternalItems = externalChecklistItems(data)}
					{@const externalItems = filteredExternalChecklistItems(data)}
					<div class="grid gap-3 md:grid-cols-4">
						<Card.Root class="border-slate-200">
							<Card.Content class="p-4">
								<p class="text-xs text-slate-500">Ditampilkan</p>
								<p class="mt-2 text-2xl font-semibold text-slate-900">{externalItems.length}</p>
								<p class="mt-1 text-xs text-slate-500">dari {allExternalItems.length} checklist</p>
							</Card.Content>
						</Card.Root>
						<Card.Root class="border-slate-200">
							<Card.Content class="p-4">
								<p class="text-xs text-slate-500">Selesai</p>
								<p class="mt-2 text-2xl font-semibold text-slate-900">{externalItems.filter((item) => obligationTrackerStatus(item) === 'done').length}</p>
							</Card.Content>
						</Card.Root>
						<Card.Root class="border-slate-200">
							<Card.Content class="p-4">
								<p class="text-xs text-slate-500">Perlu Revisi</p>
								<p class="mt-2 text-2xl font-semibold text-slate-900">{externalItems.filter((item) => obligationTrackerStatus(item) === 'needs_revision').length}</p>
							</Card.Content>
						</Card.Root>
						<Card.Root class="border-slate-200">
							<Card.Content class="p-4">
								<p class="text-xs text-slate-500">Ada Bukti</p>
								<p class="mt-2 text-2xl font-semibold text-slate-900">{externalItems.filter((item) => item.evidence_item_id || item.archive_document_id).length}</p>
							</Card.Content>
						</Card.Root>
					</div>
					<div class="flex flex-wrap items-center gap-2 rounded-md border border-slate-200 bg-white p-3">
						<label for="external-checklist-status" class="text-sm font-medium text-slate-700">Status checklist</label>
						<select id="external-checklist-status" bind:value={externalChecklistStatusFilter} class="h-9 rounded-md border border-input bg-background px-3 text-sm">
							{#each EXTERNAL_CHECKLIST_STATUSES as [value, label] (value)}
								<option {value}>{label}</option>
							{/each}
						</select>
						{#if externalChecklistStatusFilter}
							<Button type="button" variant="outline" size="sm" onclick={() => { externalChecklistStatusFilter = ''; }}>
								Reset Status
							</Button>
						{/if}
					</div>
					<Card.Root class="border-slate-200">
						<Card.Header class="pb-2">
							<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
								<div>
									<div class="flex items-center gap-2">
										<ExternalLinkIcon class="size-4 text-emerald-700" />
										<Card.Title class="text-base">Checklist Kepatuhan Eksternal</Card.Title>
									</div>
									<Card.Description>Status internal untuk SKP, EMIS, SIPKA, SIMAK-BMN, RKAM/BOS, Perkin, IKU, LAKIP/LKj, dan EDM.</Card.Description>
								</div>
								<Button type="button" variant="outline" size="sm" disabled={externalItems.length === 0} onclick={() => exportExternalChecklistCsv(externalItems)}>
									<DownloadIcon class="mr-2 size-3.5" />
									Export CSV
								</Button>
							</div>
						</Card.Header>
						<Card.Content class="p-0">
							<div class="overflow-x-auto">
								<Table.Root>
									<Table.Header>
										<Table.Row>
											<Table.Head>Portal</Table.Head>
											<Table.Head>Dokumen</Table.Head>
											<Table.Head>Status Checklist</Table.Head>
											<Table.Head>PIC & Tenggat</Table.Head>
											<Table.Head>Bukti Dukung</Table.Head>
											<Table.Head>Aksi</Table.Head>
										</Table.Row>
									</Table.Header>
									<Table.Body>
										{#each externalItems as item (item.id)}
											{@const trackerStatus = obligationTrackerStatus(item)}
											<Table.Row>
												<Table.Cell class="min-w-52">
													<Badge variant="outline">{externalSystemLabel(item.external_system)}</Badge>
													<p class="mt-2 text-xs text-slate-500">{domainAreaLabel(item.domain_area)}</p>
												</Table.Cell>
												<Table.Cell class="min-w-72">
													<p class="text-sm font-medium text-slate-900">{item.catalog_title}</p>
													<p class="text-xs text-slate-500">{item.period_label} · {item.catalog_code}</p>
												</Table.Cell>
												<Table.Cell>
													<Badge variant={trackerStatusVariant(trackerStatus)}>{trackerStatusLabel(trackerStatus)}</Badge>
													<p class="mt-2 text-xs text-slate-500">Alur dokumen: {statusLabel(item.status)}</p>
												</Table.Cell>
												<Table.Cell class="min-w-56">
													<p class="text-sm text-slate-900">{item.responsible_employee_name || 'Belum ada PIC'}</p>
													<p class={item.is_overdue ? 'mt-1 text-xs font-medium text-red-700' : 'mt-1 text-xs text-slate-500'}>Jatuh tempo {formatDate(item.due_date)}</p>
												</Table.Cell>
												<Table.Cell class="min-w-64">
													<div class="flex flex-wrap gap-1.5">
														<Badge variant={item.evidence_item_id ? 'default' : 'outline'}>Evidence</Badge>
														<Badge variant={item.archive_document_id ? 'default' : 'outline'}>Arsip</Badge>
														<Badge variant={item.compliance_action_id ? 'default' : 'outline'}>Aksi</Badge>
													</div>
													<p class="mt-2 text-xs text-slate-500">{item.archive_document_title || item.evidence_item_title || item.compliance_action_title || 'Belum ada bukti tertaut'}</p>
												</Table.Cell>
												<Table.Cell>
													<Button size="sm" variant="outline" onclick={() => selectObligation(item)}>
														<PencilIcon class="mr-2 size-3.5" />
														Detail
													</Button>
												</Table.Cell>
											</Table.Row>
										{:else}
											<Table.Row>
												<Table.Cell colspan={6} class="py-10 text-center text-sm text-slate-500">
													Belum ada checklist eksternal pada filter saat ini.
												</Table.Cell>
											</Table.Row>
										{/each}
									</Table.Body>
								</Table.Root>
							</div>
						</Card.Content>
					</Card.Root>
				</Tabs.Content>
				<Tabs.Content value="connections" class="space-y-4">
					<div class="grid gap-3 md:grid-cols-4">
						<Card.Root class="border-slate-200">
							<Card.Content class="p-4">
								<p class="text-xs text-slate-500">Dokumen Tata Kelola</p>
								<p class="mt-2 text-2xl font-semibold text-slate-900">{data.stats.linked_governance_document_obligations}</p>
							</Card.Content>
						</Card.Root>
						<Card.Root class="border-slate-200">
							<Card.Content class="p-4">
								<p class="text-xs text-slate-500">Evidence 8 SNP</p>
								<p class="mt-2 text-2xl font-semibold text-slate-900">{data.stats.linked_evidence_obligations}</p>
							</Card.Content>
						</Card.Root>
						<Card.Root class="border-slate-200">
							<Card.Content class="p-4">
								<p class="text-xs text-slate-500">Arsip Digital</p>
								<p class="mt-2 text-2xl font-semibold text-slate-900">{data.stats.linked_archive_obligations}</p>
							</Card.Content>
						</Card.Root>
						<Card.Root class="border-slate-200">
							<Card.Content class="p-4">
								<p class="text-xs text-slate-500">Tindak Lanjut</p>
								<p class="mt-2 text-2xl font-semibold text-slate-900">{data.stats.linked_compliance_action_obligations}</p>
							</Card.Content>
						</Card.Root>
					</div>
					<Card.Root class="border-slate-200">
						<Card.Header class="pb-2">
							<div class="flex items-center gap-2">
								<NetworkIcon class="size-4 text-emerald-700" />
								<Card.Title class="text-base">Peta Keterhubungan Dokumen</Card.Title>
							</div>
							<Card.Description>Jejak dari siklus dokumen ke arsip, evidence, RKT/RKJM, SKP, dan tindak lanjut.</Card.Description>
						</Card.Header>
						<Card.Content class="p-0">
							<div class="overflow-x-auto">
								<Table.Root>
									<Table.Header>
										<Table.Row>
											<Table.Head>Dokumen</Table.Head>
											<Table.Head>Bidang</Table.Head>
											<Table.Head>Keterhubungan</Table.Head>
											<Table.Head>Jejak Utama</Table.Head>
											<Table.Head>Aksi</Table.Head>
										</Table.Row>
									</Table.Header>
									<Table.Body>
										{#each data.obligations as item (item.id)}
											{@const score = connectionScore(item)}
											<Table.Row>
												<Table.Cell class="min-w-72">
													<p class="text-sm font-medium text-slate-900">{item.catalog_title}</p>
													<p class="text-xs text-slate-500">{item.period_label} · {item.catalog_code}</p>
												</Table.Cell>
												<Table.Cell>
													<Badge variant="outline">{domainAreaLabel(item.domain_area)}</Badge>
													{#if item.external_system}
														<Badge variant="secondary" class="mt-1">{externalSystemLabel(item.external_system)}</Badge>
													{/if}
												</Table.Cell>
												<Table.Cell class="min-w-72">
													<div class="flex flex-wrap gap-1.5">
														<Badge variant={item.governance_document_id ? 'default' : 'outline'}>Dokumen</Badge>
														<Badge variant={item.work_plan_item_id ? 'default' : 'outline'}>RKT/RKJM</Badge>
														<Badge variant={item.performance_target_id ? 'default' : 'outline'}>SKP</Badge>
														<Badge variant={item.evidence_item_id ? 'default' : 'outline'}>Evidence</Badge>
														<Badge variant={item.compliance_action_id ? 'default' : 'outline'}>Aksi</Badge>
														<Badge variant={item.archive_document_id ? 'default' : 'outline'}>Arsip</Badge>
													</div>
													<p class="mt-2 text-xs text-slate-500">{score.done}/{score.total} tautan terisi</p>
												</Table.Cell>
												<Table.Cell class="min-w-80 text-xs text-slate-600">
													<p>Dokumen: {item.governance_document_title || '-'}</p>
													<p>RKT/RKJM: {item.work_plan_item_name || '-'}</p>
													<p>SKP: {item.performance_target_title || '-'}</p>
													<p>Evidence: {item.evidence_item_title || '-'}</p>
													<p>Arsip: {item.archive_document_title || '-'}</p>
												</Table.Cell>
												<Table.Cell>
													<Button size="sm" variant="outline" onclick={() => selectObligation(item)}>
														<PencilIcon class="mr-2 size-3.5" />
														Detail
													</Button>
												</Table.Cell>
											</Table.Row>
										{/each}
									</Table.Body>
								</Table.Root>
							</div>
						</Card.Content>
					</Card.Root>
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
										<label for="catalog-domain" class="text-sm font-medium">Bidang</label>
										<select id="catalog-domain" bind:value={catalogForm.domain_area} class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm">
											{#each CATALOG_DOMAIN_AREAS as [value, label] (value)}
												<option {value}>{label}</option>
											{/each}
										</select>
									</div>
									<div>
										<label for="catalog-external-system" class="text-sm font-medium">Tracker Eksternal</label>
										<select id="catalog-external-system" bind:value={catalogForm.external_system} class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm">
											{#each EXTERNAL_SYSTEMS as [value, label] (value)}
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
													<p class="text-xs text-slate-500">{domainAreaLabel(catalog.domain_area)} · {snpLabel(catalog.snp_standard)} · {catalog.regulation_ref || 'Tanpa rujukan khusus'}</p>
													{#if catalog.external_system}
														<Badge variant="outline" class="mt-2">{externalSystemLabel(catalog.external_system)}</Badge>
													{/if}
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
