<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import CalendarDaysIcon from '@lucide/svelte/icons/calendar-days';
	import ClipboardCheckIcon from '@lucide/svelte/icons/clipboard-check';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import PrinterIcon from '@lucide/svelte/icons/printer';
	import RefreshCcwIcon from '@lucide/svelte/icons/refresh-ccw';
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { downloadCsv } from '$lib/governance/print-pack-data';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import * as Table from '$lib/components/ui/table';
	import { Textarea } from '$lib/components/ui/textarea';
	import { toast } from '$lib/components/ui/sonner';
	import { confirmAction } from '$lib/confirm-dialog';

	type ApiEnvelope<T> = {
		data?: T;
		error?: string;
		message?: string;
	};

	interface Stats {
		compliance_actions?: number;
		open_compliance_actions?: number;
		completed_compliance_actions?: number;
		critical_compliance_actions?: number;
	}

	interface ComplianceActionRow {
		id: string;
		period_year: number;
		source_type: string;
		source_ref_id?: string;
		snp_standard: string;
		program_id?: string;
		document_id?: string;
		performance_target_id?: string;
		evidence_item_id?: string;
		owner_unit_id?: string;
		responsible_employee_id?: string;
		title: string;
		description: string;
		priority: string;
		status: string;
		due_date?: string;
		completed_at?: string;
		follow_up_notes: string;
		evidence_url: string;
		updated_at: string;
		program_code?: string;
		program_name?: string;
		document_title?: string;
		performance_target_title?: string;
		evidence_item_title?: string;
		owner_unit_name?: string;
		responsible_employee_name?: string;
		responsible_employee_nip?: string;
	}

	interface ProgramRow {
		id: string;
		period_year: number;
		code: string;
		name: string;
		source_document_id?: string;
		owner_unit_id?: string;
		responsible_employee_id?: string;
		snp_standard: string;
		iku_code: string;
		status: string;
		progress_percent: number;
		evidence_url: string;
		owner_unit_name?: string;
		responsible_employee_name?: string;
	}

	interface DocumentRow {
		id: string;
		title: string;
		doc_type: string;
		period_year: number;
	}

	interface WorkPlanItemRow {
		id: string;
		program_id: string;
		status: string;
		evidence_item_id?: string;
		evidence_url: string;
	}

	interface PerformanceTargetRow {
		id: string;
		program_id?: string;
		status: string;
		evidence_url: string;
		title: string;
	}

	interface EvidenceItemRow {
		id: string;
		program_id?: string;
		performance_target_id?: string;
		title: string;
		status: string;
	}

	interface UnitRow {
		id: string;
		name: string;
	}

	interface EmployeeOption {
		id: string;
		nama: string;
		nip: string;
		unit_kerja: string;
	}

	interface ComplianceOverview {
		stats: Stats;
		actions: ComplianceActionRow[];
		programs: ProgramRow[];
		documents: DocumentRow[];
		workPlanItems: WorkPlanItemRow[];
		performanceTargets: PerformanceTargetRow[];
		evidenceItems: EvidenceItemRow[];
		units: UnitRow[];
		employees: EmployeeOption[];
	}

	interface ComplianceActionForm {
		period_year: number;
		source_type: string;
		source_ref_id: string;
		snp_standard: string;
		program_id: string;
		document_id: string;
		performance_target_id: string;
		evidence_item_id: string;
		owner_unit_id: string;
		responsible_employee_id: string;
		title: string;
		description: string;
		priority: string;
		status: string;
		due_date: string;
		completed_at: string;
		follow_up_notes: string;
		evidence_url: string;
	}

	interface GapSuggestion {
		key: string;
		program: ProgramRow;
		gap: string;
		title: string;
		description: string;
		priority: string;
	}

	interface RegisterFocus {
		kind: 'risk' | 'owner' | 'snp';
		value: string;
		label: string;
	}

	interface EscalationRow {
		key: string;
		label: string;
		unitLabel: string;
		openCount: number;
		criticalCount: number;
		overdueCount: number;
		waitingEvidenceCount: number;
		nextDueDate?: string;
	}

	interface SnpEscalationRow {
		code: string;
		label: string;
		openCount: number;
		criticalCount: number;
		overdueCount: number;
		waitingEvidenceCount: number;
	}

	interface OwnerFilterOption {
		key: string;
		label: string;
		unitLabel: string;
	}

	interface QuickStatusOption {
		status: string;
		label: string;
		variant: 'default' | 'secondary' | 'outline' | 'destructive';
	}

	interface EvidenceCaptureForm {
		evidence_item_id: string;
		evidence_url: string;
		follow_up_notes: string;
	}

	type BadgeVariant = 'default' | 'secondary' | 'destructive' | 'outline';

	const SOURCE_TYPES = [
		['alignment_gap', 'Gap Peta IKU'],
		['snp_gap', 'Gap 8 SNP'],
		['audit', 'Review/Audit'],
		['document', 'Dokumen'],
		['program', 'Program'],
		['performance', 'SKP/Kinerja'],
		['evidence', 'Bukti Mutu'],
		['manual', 'Manual'],
	] as const;

	const PRIORITIES = [
		['low', 'Rendah'],
		['medium', 'Sedang'],
		['high', 'Tinggi'],
		['urgent', 'Mendesak'],
	] as const;

	const STATUSES = [
		['open', 'Terbuka'],
		['in_progress', 'Dikerjakan'],
		['waiting_evidence', 'Menunggu Bukti'],
		['done', 'Selesai'],
		['cancelled', 'Dibatalkan'],
	] as const;

	const SNP_STANDARDS = [
		['', 'Tidak dikaitkan'],
		['skl', 'SKL'],
		['isi', 'Isi'],
		['proses', 'Proses'],
		['penilaian', 'Penilaian'],
		['ptk', 'PTK'],
		['sarpras', 'Sarpras'],
		['pengelolaan', 'Pengelolaan'],
		['pembiayaan', 'Pembiayaan'],
	] as const;

	let overviewPromise = $state<Promise<ComplianceOverview> | null>(null);
	let actions = $state<ComplianceActionRow[]>([]);
	let programs = $state<ProgramRow[]>([]);
	let documents = $state<DocumentRow[]>([]);
	let workPlanItems = $state<WorkPlanItemRow[]>([]);
	let performanceTargets = $state<PerformanceTargetRow[]>([]);
	let evidenceItems = $state<EvidenceItemRow[]>([]);
	let units = $state<UnitRow[]>([]);
	let employees = $state<EmployeeOption[]>([]);
	let stats = $state<Stats>({});
	let search = $state('');
	let statusFilter = $state('');
	let priorityFilter = $state('');
	let yearFilter = $state('');
	let sourceFilter = $state('');
	let snpFilter = $state('');
	let ownerFilter = $state('');
	let registerFocus = $state<RegisterFocus | null>(null);
	let dialogOpen = $state(false);
	let editingId = $state<string | null>(null);
	let busy = $state(false);
	let refreshBusy = $state(false);
	let quickStatusBusy = $state<Record<string, boolean>>({});
	let actionForm = $state<ComplianceActionForm>(emptyActionForm());
	let evidenceDialogOpen = $state(false);
	let evidenceBusy = $state(false);
	let evidenceAction = $state<ComplianceActionRow | null>(null);
	let evidenceForm = $state<EvidenceCaptureForm>(emptyEvidenceForm());

	const filteredActions = $derived.by(() => {
		const q = search.trim().toLowerCase();
		return actions.filter((action) => {
			const matchesSearch =
				!q ||
				[
					action.title,
					action.description,
					action.follow_up_notes,
					action.program_code ?? '',
					action.program_name ?? '',
					action.document_title ?? '',
					action.owner_unit_name ?? '',
					action.responsible_employee_name ?? '',
				].some((value) => value.toLowerCase().includes(q));
			const matchesStatus = !statusFilter || action.status === statusFilter;
			const matchesPriority = !priorityFilter || action.priority === priorityFilter;
			const matchesYear = !yearFilter || String(action.period_year) === yearFilter;
			const matchesSource = !sourceFilter || action.source_type === sourceFilter;
			const matchesSnp = !snpFilter || (action.snp_standard || '') === snpFilter;
			const matchesOwner = !ownerFilter || actionOwnerKey(action) === ownerFilter;
			const matchesFocus = !registerFocus || matchesRegisterFocus(action, registerFocus);
			return matchesSearch && matchesStatus && matchesPriority && matchesYear && matchesSource && matchesSnp && matchesOwner && matchesFocus;
		});
	});

	const actionYears = $derived.by(() => buildActionYears());
	const ownerFilterOptions = $derived.by(() => buildOwnerFilterOptions());
	const activeFilterCount = $derived.by(() => {
		let count = 0;
		if (search.trim()) count += 1;
		if (statusFilter) count += 1;
		if (priorityFilter) count += 1;
		if (yearFilter) count += 1;
		if (sourceFilter) count += 1;
		if (snpFilter) count += 1;
		if (ownerFilter) count += 1;
		if (registerFocus) count += 1;
		return count;
	});
	const gapSuggestions = $derived.by(() => buildGapSuggestions().slice(0, 8));
	const openCount = $derived(actions.filter((action) => isOpenAction(action)).length);
	const overdueCount = $derived(actions.filter((action) => isOverdue(action)).length);
	const criticalOpenCount = $derived(actions.filter((action) => isOpenAction(action) && isCriticalAction(action)).length);
	const waitingEvidenceCount = $derived(actions.filter((action) => isOpenAction(action) && action.status === 'waiting_evidence').length);
	const unassignedCount = $derived(actions.filter((action) => isOpenAction(action) && actionOwnerKey(action) === 'unassigned').length);
	const escalationRows = $derived.by(() => buildEscalationRows().slice(0, 6));
	const snpEscalationRows = $derived.by(() => buildSnpEscalationRows());

	function emptyActionForm(): ComplianceActionForm {
		return {
			period_year: new Date().getFullYear(),
			source_type: 'manual',
			source_ref_id: '',
			snp_standard: '',
			program_id: '',
			document_id: '',
			performance_target_id: '',
			evidence_item_id: '',
			owner_unit_id: '',
			responsible_employee_id: '',
			title: '',
			description: '',
			priority: 'medium',
			status: 'open',
			due_date: '',
			completed_at: '',
			follow_up_notes: '',
			evidence_url: '',
		};
	}

	function emptyEvidenceForm(): EvidenceCaptureForm {
		return {
			evidence_item_id: '',
			evidence_url: '',
			follow_up_notes: '',
		};
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
		if (!response.ok) throw new Error(message || fallbackMessage);
		if (isRecord(payload) && typeof payload.error === 'string' && payload.error.trim()) throw new Error(payload.error);
		if (isRecord(payload) && 'data' in payload) {
			const envelope = payload as ApiEnvelope<T>;
			if (envelope.data === undefined) throw new Error(fallbackMessage);
			return envelope.data;
		}
		if (payload === null) throw new Error(fallbackMessage);
		return payload as T;
	}

	async function fetchOverview(): Promise<ComplianceOverview> {
		const [statsData, actionData, programData, documentData, workPlanData, performanceData, evidenceData, unitData, employeeData] =
			await Promise.all([
				fetch('/api/governance/stats').then((response) => readApi<Stats>(response, 'Gagal memuat statistik.')),
				fetch('/api/governance/compliance-actions').then((response) => readApi<ComplianceActionRow[]>(response, 'Gagal memuat tindak lanjut.')),
				fetch('/api/governance/programs').then((response) => readApi<ProgramRow[]>(response, 'Gagal memuat program.')),
				fetch('/api/governance/documents').then((response) => readApi<DocumentRow[]>(response, 'Gagal memuat dokumen.')),
				fetch('/api/governance/work-plan-items').then((response) => readApi<WorkPlanItemRow[]>(response, 'Gagal memuat RKT/RKJM.')),
				fetch('/api/governance/performance-targets').then((response) => readApi<PerformanceTargetRow[]>(response, 'Gagal memuat SKP.')),
				fetch('/api/governance/evidence-items').then((response) => readApi<EvidenceItemRow[]>(response, 'Gagal memuat bukti mutu.')),
				fetch('/api/governance/units').then((response) => readApi<UnitRow[]>(response, 'Gagal memuat unit kerja.')),
				fetch('/api/governance/employee-options').then((response) => readApi<EmployeeOption[]>(response, 'Gagal memuat pegawai.')),
			]);
		return {
			stats: statsData ?? {},
			actions: actionData ?? [],
			programs: programData ?? [],
			documents: documentData ?? [],
			workPlanItems: workPlanData ?? [],
			performanceTargets: performanceData ?? [],
			evidenceItems: evidenceData ?? [],
			units: unitData ?? [],
			employees: employeeData ?? [],
		};
	}

	function applyOverview(next: ComplianceOverview) {
		stats = next.stats;
		actions = next.actions;
		programs = next.programs;
		documents = next.documents;
		workPlanItems = next.workPlanItems;
		performanceTargets = next.performanceTargets;
		evidenceItems = next.evidenceItems;
		units = next.units;
		employees = next.employees;
	}

	function currentOverview(): ComplianceOverview {
		return { stats, actions, programs, documents, workPlanItems, performanceTargets, evidenceItems, units, employees };
	}

	function loadOverview() {
		overviewPromise = fetchOverview().then((next) => {
			applyOverview(next);
			return next;
		});
		return overviewPromise;
	}

	async function refreshOverview() {
		try {
			const next = await fetchOverview();
			applyOverview(next);
			overviewPromise = Promise.resolve(next);
		} catch (error) {
			overviewPromise = Promise.resolve(currentOverview());
			toast.error(errorMessage(error));
		}
	}

	async function refreshOverviewAction() {
		refreshBusy = true;
		try {
			await refreshOverview();
		} finally {
			refreshBusy = false;
		}
	}

	function retryOverview(reset?: () => void) {
		reset?.();
		loadOverview();
	}

	function errorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Data tindak lanjut belum dapat dimuat.';
	}

	function handleRenderError(error: unknown, reset: () => void) {
		console.error('Compliance actions render failed', error);
		reset();
	}

	function openCreate() {
		editingId = null;
		actionForm = emptyActionForm();
		dialogOpen = true;
	}

	function openEdit(action: ComplianceActionRow) {
		editingId = action.id;
		actionForm = actionFormFromRow(action);
		dialogOpen = true;
	}

	function openSuggestion(suggestion: GapSuggestion) {
		editingId = null;
		actionForm = {
			...emptyActionForm(),
			period_year: suggestion.program.period_year,
			source_type: 'alignment_gap',
			source_ref_id: suggestion.program.id,
			snp_standard: suggestion.program.snp_standard,
			program_id: suggestion.program.id,
			owner_unit_id: suggestion.program.owner_unit_id ?? '',
			responsible_employee_id: suggestion.program.responsible_employee_id ?? '',
			title: suggestion.title,
			description: suggestion.description,
			priority: suggestion.priority,
			due_date: defaultDueDate(suggestion.priority),
		};
		dialogOpen = true;
	}

	function openEvidenceCapture(action: ComplianceActionRow) {
		evidenceAction = action;
		evidenceForm = {
			evidence_item_id: action.evidence_item_id ?? '',
			evidence_url: action.evidence_url,
			follow_up_notes: action.follow_up_notes,
		};
		evidenceDialogOpen = true;
	}

	async function saveAction() {
		busy = true;
		try {
			const res = await fetch(editingId ? `/api/governance/compliance-actions/${editingId}` : '/api/governance/compliance-actions', {
				method: editingId ? 'PUT' : 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(actionForm),
			});
			const payload = await res.json().catch(() => null);
			if (!res.ok) {
				toast.error(apiErrorMessage(payload) || 'Gagal menyimpan tindak lanjut.');
				return;
			}
			toast.success(editingId ? 'Tindak lanjut diperbarui.' : 'Tindak lanjut ditambahkan.');
			dialogOpen = false;
			await refreshOverview();
		} finally {
			busy = false;
		}
	}

	async function deleteAction(action: ComplianceActionRow) {
		if (!(await confirmAction({
			title: 'Hapus Tindak Lanjut',
			message: `Hapus tindak lanjut "${action.title}"?`,
			confirmLabel: 'Hapus Tindak Lanjut',
			tone: 'danger'
		}))) return;
		const res = await fetch(`/api/governance/compliance-actions/${action.id}`, { method: 'DELETE' });
		if (!res.ok) {
			const payload = await res.json().catch(() => null);
			toast.error(apiErrorMessage(payload) || 'Gagal menghapus tindak lanjut.');
			return;
		}
		toast.success('Tindak lanjut dihapus.');
		await refreshOverview();
	}

	function actionFormFromRow(action: ComplianceActionRow, overrides: Partial<ComplianceActionForm> = {}): ComplianceActionForm {
		return {
			period_year: action.period_year,
			source_type: action.source_type,
			source_ref_id: action.source_ref_id ?? '',
			snp_standard: action.snp_standard,
			program_id: action.program_id ?? '',
			document_id: action.document_id ?? '',
			performance_target_id: action.performance_target_id ?? '',
			evidence_item_id: action.evidence_item_id ?? '',
			owner_unit_id: action.owner_unit_id ?? '',
			responsible_employee_id: action.responsible_employee_id ?? '',
			title: action.title,
			description: action.description,
			priority: action.priority,
			status: action.status,
			due_date: dateValue(action.due_date),
			completed_at: action.completed_at ?? '',
			follow_up_notes: action.follow_up_notes,
			evidence_url: action.evidence_url,
			...overrides,
		};
	}

	function quickStatusOptions(action: ComplianceActionRow): QuickStatusOption[] {
		if (action.status === 'done') return [{ status: 'in_progress', label: 'Buka Ulang', variant: 'outline' }];
		if (action.status === 'cancelled') return [{ status: 'open', label: 'Aktifkan', variant: 'outline' }];
		const options: QuickStatusOption[] = [];
		if (action.status === 'open') options.push({ status: 'in_progress', label: 'Mulai', variant: 'secondary' });
		if (action.status !== 'waiting_evidence') options.push({ status: 'waiting_evidence', label: 'Bukti', variant: 'outline' });
		options.push({ status: 'done', label: 'Selesai', variant: 'default' });
		return options;
	}

	async function quickUpdateStatus(action: ComplianceActionRow, nextStatus: string) {
		if (nextStatus === 'done' && !action.evidence_url.trim() && !action.evidence_item_id) {
			const proceed = await confirmAction({
				title: 'Selesaikan Tanpa Bukti',
				message: 'Tandai selesai tanpa tautan/item bukti? Catatan bukti masih bisa ditambahkan lewat Edit.',
				confirmLabel: 'Tandai Selesai',
				tone: 'warning'
			});
			if (!proceed) return;
		}
		quickStatusBusy = { ...quickStatusBusy, [action.id]: true };
		try {
			const body = actionFormFromRow(action, {
				status: nextStatus,
				completed_at: nextStatus === 'done' ? action.completed_at ?? '' : '',
			});
			const res = await fetch(`/api/governance/compliance-actions/${action.id}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(body),
			});
			const payload = await res.json().catch(() => null);
			if (!res.ok) {
				toast.error(apiErrorMessage(payload) || 'Gagal memperbarui status tindak lanjut.');
				return;
			}
			toast.success(`Status diperbarui menjadi ${statusLabel(nextStatus)}.`);
			await refreshOverview();
		} finally {
			quickStatusBusy = { ...quickStatusBusy, [action.id]: false };
		}
	}

	async function saveEvidenceCapture(markDone: boolean) {
		if (!evidenceAction) return;
		if (markDone && !evidenceForm.evidence_url.trim() && !evidenceForm.evidence_item_id) {
			toast.error('Tambahkan URL/lokasi bukti atau pilih item bukti sebelum menandai selesai.');
			return;
		}
		evidenceBusy = true;
		try {
			const body = actionFormFromRow(evidenceAction, {
				evidence_item_id: evidenceForm.evidence_item_id,
				evidence_url: evidenceForm.evidence_url,
				follow_up_notes: evidenceForm.follow_up_notes,
				status: markDone ? 'done' : evidenceAction.status,
				completed_at: markDone ? evidenceAction.completed_at ?? '' : evidenceAction.completed_at ?? '',
			});
			const res = await fetch(`/api/governance/compliance-actions/${evidenceAction.id}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(body),
			});
			const payload = await res.json().catch(() => null);
			if (!res.ok) {
				toast.error(apiErrorMessage(payload) || 'Gagal menyimpan bukti tindak lanjut.');
				return;
			}
			toast.success(markDone ? 'Bukti dicatat dan tindak lanjut diselesaikan.' : 'Bukti tindak lanjut dicatat.');
			evidenceDialogOpen = false;
			evidenceAction = null;
			evidenceForm = emptyEvidenceForm();
			await refreshOverview();
		} finally {
			evidenceBusy = false;
		}
	}

	function isOpenAction(action: ComplianceActionRow) {
		return !['done', 'cancelled'].includes(action.status);
	}

	function isCriticalAction(action: ComplianceActionRow) {
		return ['high', 'urgent'].includes(action.priority);
	}

	function numericDate(value?: string) {
		if (!value) return Number.MAX_SAFE_INTEGER;
		const time = new Date(value).getTime();
		return Number.isFinite(time) ? time : Number.MAX_SAFE_INTEGER;
	}

	function actionOwnerKey(action: ComplianceActionRow) {
		if (action.responsible_employee_id) return `employee:${action.responsible_employee_id}`;
		if (action.owner_unit_id) return `unit:${action.owner_unit_id}`;
		return 'unassigned';
	}

	function actionOwnerLabel(action: ComplianceActionRow) {
		if (action.responsible_employee_name) return action.responsible_employee_name;
		if (action.responsible_employee_id) return employees.find((employee) => employee.id === action.responsible_employee_id)?.nama ?? 'PIC tanpa nama';
		if (action.owner_unit_name) return action.owner_unit_name;
		if (action.owner_unit_id) return units.find((unit) => unit.id === action.owner_unit_id)?.name ?? 'Unit tanpa nama';
		return 'Belum ada PIC';
	}

	function actionOwnerUnitLabel(action: ComplianceActionRow) {
		if (action.owner_unit_name) return action.owner_unit_name;
		if (action.responsible_employee_nip) return action.responsible_employee_nip;
		return actionOwnerKey(action) === 'unassigned' ? 'Perlu penugasan' : 'Unit belum dicatat';
	}

	function buildActionYears() {
		const years: number[] = [];
		for (const action of actions) {
			if (!years.includes(action.period_year)) years.push(action.period_year);
		}
		return years.sort((a, b) => b - a);
	}

	function buildOwnerFilterOptions(): OwnerFilterOption[] {
		const rows: OwnerFilterOption[] = [];
		for (const action of actions) {
			const key = actionOwnerKey(action);
			if (rows.some((row) => row.key === key)) continue;
			rows.push({
				key,
				label: actionOwnerLabel(action),
				unitLabel: actionOwnerUnitLabel(action),
			});
		}
		return rows.sort((a, b) => a.label.localeCompare(b.label) || a.unitLabel.localeCompare(b.unitLabel));
	}

	function buildEscalationRows(): EscalationRow[] {
		const grouped: EscalationRow[] = [];
		for (const action of actions.filter(isOpenAction)) {
			const key = actionOwnerKey(action);
			let existing = grouped.find((row) => row.key === key);
			if (!existing) {
				existing = {
					key,
					label: actionOwnerLabel(action),
					unitLabel: actionOwnerUnitLabel(action),
					openCount: 0,
					criticalCount: 0,
					overdueCount: 0,
					waitingEvidenceCount: 0,
					nextDueDate: undefined,
				};
				grouped.push(existing);
			}
			existing.openCount += 1;
			if (isCriticalAction(action)) existing.criticalCount += 1;
			if (isOverdue(action)) existing.overdueCount += 1;
			if (action.status === 'waiting_evidence') existing.waitingEvidenceCount += 1;
			if (numericDate(action.due_date) < numericDate(existing.nextDueDate)) existing.nextDueDate = action.due_date;
		}
		return grouped.sort(
			(a, b) =>
				b.overdueCount - a.overdueCount ||
				b.criticalCount - a.criticalCount ||
				b.openCount - a.openCount ||
				numericDate(a.nextDueDate) - numericDate(b.nextDueDate) ||
				a.label.localeCompare(b.label)
		);
	}

	function buildSnpEscalationRows(): SnpEscalationRow[] {
		const grouped: SnpEscalationRow[] = [];
		for (const action of actions.filter(isOpenAction)) {
			const code = action.snp_standard || '';
			let existing = grouped.find((row) => row.code === code);
			if (!existing) {
				existing = {
					code,
					label: snpLabel(code),
					openCount: 0,
					criticalCount: 0,
					overdueCount: 0,
					waitingEvidenceCount: 0,
				};
				grouped.push(existing);
			}
			existing.openCount += 1;
			if (isCriticalAction(action)) existing.criticalCount += 1;
			if (isOverdue(action)) existing.overdueCount += 1;
			if (action.status === 'waiting_evidence') existing.waitingEvidenceCount += 1;
		}
		return grouped.sort(
			(a, b) =>
				b.overdueCount - a.overdueCount ||
				b.criticalCount - a.criticalCount ||
				b.openCount - a.openCount ||
				a.label.localeCompare(b.label)
		);
	}

	function matchesRegisterFocus(action: ComplianceActionRow, focus: RegisterFocus) {
		if (focus.kind === 'owner') return isOpenAction(action) && actionOwnerKey(action) === focus.value;
		if (focus.kind === 'snp') return isOpenAction(action) && (action.snp_standard || '') === focus.value;
		if (focus.value === 'overdue') return isOverdue(action);
		if (focus.value === 'critical') return isOpenAction(action) && isCriticalAction(action);
		if (focus.value === 'unassigned') return isOpenAction(action) && actionOwnerKey(action) === 'unassigned';
		if (focus.value === 'waiting_evidence') return isOpenAction(action) && action.status === 'waiting_evidence';
		return true;
	}

	function setRegisterFocus(focus: RegisterFocus) {
		registerFocus = focus;
	}

	function clearRegisterFocus() {
		registerFocus = null;
	}

	function clearRegisterFilters() {
		search = '';
		statusFilter = '';
		priorityFilter = '';
		yearFilter = '';
		sourceFilter = '';
		snpFilter = '';
		ownerFilter = '';
		registerFocus = null;
	}

	function focusVariant(kind: RegisterFocus['kind'], value: string): 'default' | 'outline' {
		return registerFocus?.kind === kind && registerFocus.value === value ? 'default' : 'outline';
	}

	function buildGapSuggestions(): GapSuggestion[] {
		const documentsByID = new Set(documents.map((document) => document.id));
		const suggestions: GapSuggestion[] = [];
		for (const program of programs) {
			const linkedWorkPlan = workPlanItems.filter((item) => item.program_id === program.id);
			const linkedPerformance = performanceTargets.filter((target) => target.program_id === program.id);
			const linkedPerformanceIDs = new Set(linkedPerformance.map((target) => target.id));
			const linkedWorkPlanEvidenceIDs = new Set(linkedWorkPlan.map((item) => item.evidence_item_id ?? '').filter(Boolean));
			const linkedEvidence = evidenceItems.filter(
				(item) =>
					item.program_id === program.id ||
					(item.performance_target_id ? linkedPerformanceIDs.has(item.performance_target_id) : false) ||
					linkedWorkPlanEvidenceIDs.has(item.id)
			);
			if (!program.iku_code.trim()) suggestions.push(gapSuggestion(program, 'IKU belum diisi', 'Lengkapi kode IKU dan indikator yang menjadi ukuran program.', 'high'));
			if (!program.source_document_id || !documentsByID.has(program.source_document_id)) {
				suggestions.push(gapSuggestion(program, 'Dokumen sumber belum dikaitkan', 'Hubungkan program ke dokumen Renstra/RKJM/RKT/Perkin/IKU yang sah.', 'medium'));
			}
			if (linkedWorkPlan.length === 0) suggestions.push(gapSuggestion(program, 'Belum turun ke RKT/RKJM', 'Buat item RKT/RKJM tahunan agar program memiliki kegiatan operasional.', 'high'));
			if (linkedPerformance.length === 0) suggestions.push(gapSuggestion(program, 'Belum turun ke SKP', 'Turunkan program ke target kinerja pegawai sebagai mirror SKP internal.', 'medium'));
			const hasEvidence = program.evidence_url.trim() || linkedWorkPlan.some((item) => item.evidence_url.trim()) || linkedPerformance.some((target) => target.evidence_url.trim()) || linkedEvidence.length > 0;
			if (!hasEvidence) suggestions.push(gapSuggestion(program, 'Bukti mutu belum tersedia', 'Tambahkan tautan bukti, arsip, atau item bukti mutu yang dapat diverifikasi.', 'high'));
			if (program.status === 'blocked' || linkedWorkPlan.some((item) => item.status === 'blocked') || linkedPerformance.some((target) => target.status === 'blocked')) {
				suggestions.push(gapSuggestion(program, 'Ada status terkendala', 'Catat keputusan tindak lanjut untuk membuka hambatan program/RKT/SKP.', 'urgent'));
			}
		}
		const activeActionKeys = new Set(actions.filter((action) => !['done', 'cancelled'].includes(action.status)).map((action) => `${action.program_id ?? action.source_ref_id}:${action.title}`));
		return suggestions.filter((suggestion) => !activeActionKeys.has(`${suggestion.program.id}:${suggestion.title}`));
	}

	function gapSuggestion(program: ProgramRow, gap: string, description: string, priority: string): GapSuggestion {
		return {
			key: `${program.id}-${gap}`,
			program,
			gap,
			title: `${gap}: ${program.code} ${program.name}`,
			description,
			priority,
		};
	}

	function defaultDueDate(priority: string) {
		const days = priority === 'urgent' ? 3 : priority === 'high' ? 7 : 14;
		return new Date(Date.now() + days * 24 * 60 * 60 * 1000).toISOString().slice(0, 10);
	}

	function isOverdue(action: ComplianceActionRow) {
		if (!action.due_date || !isOpenAction(action)) return false;
		return new Date(action.due_date).getTime() < new Date(new Date().toISOString().slice(0, 10)).getTime();
	}

	function dueLabel(value?: string) {
		if (!value) return 'Tanpa tenggat';
		const today = new Date(new Date().toISOString().slice(0, 10)).getTime();
		const due = new Date(value.slice(0, 10)).getTime();
		if (!Number.isFinite(due)) return 'Tenggat tidak valid';
		const days = Math.ceil((due - today) / (24 * 60 * 60 * 1000));
		if (days < 0) return `${Math.abs(days)} hari lewat`;
		if (days === 0) return 'Hari ini';
		if (days === 1) return 'Besok';
		return `${days} hari lagi`;
	}

	function dateValue(value?: string) {
		if (!value) return '';
		return value.slice(0, 10);
	}

	function formatDate(value?: string) {
		if (!value) return '-';
		return new Intl.DateTimeFormat('id-ID', {
			timeZone: 'Asia/Makassar',
			day: '2-digit',
			month: 'short',
			year: 'numeric',
		}).format(new Date(value));
	}

	function sourceLabel(value: string) {
		return SOURCE_TYPES.find(([key]) => key === value)?.[1] ?? value;
	}

	function priorityLabel(value: string) {
		return PRIORITIES.find(([key]) => key === value)?.[1] ?? value;
	}

	function statusLabel(value: string) {
		return STATUSES.find(([key]) => key === value)?.[1] ?? value;
	}

	function snpLabel(value: string) {
		return SNP_STANDARDS.find(([key]) => key === value)?.[1] ?? 'Tidak dikaitkan';
	}

	function priorityVariant(value: string): BadgeVariant {
		if (value === 'urgent' || value === 'high') return 'destructive';
		if (value === 'medium') return 'secondary';
		return 'outline';
	}

	function statusVariant(value: string): BadgeVariant {
		if (value === 'done') return 'default';
		if (value === 'cancelled') return 'outline';
		if (value === 'waiting_evidence') return 'secondary';
		return 'destructive';
	}

	function linkedLabel(action: ComplianceActionRow) {
		const parts = [
			action.program_code || action.program_name ? `Program: ${[action.program_code, action.program_name].filter(Boolean).join(' · ')}` : '',
			action.document_title ? `Dokumen: ${action.document_title}` : '',
			action.performance_target_title ? `SKP: ${action.performance_target_title}` : '',
			action.evidence_item_title ? `Bukti: ${action.evidence_item_title}` : '',
		].filter(Boolean);
		return parts.join(' | ') || 'Belum dikaitkan';
	}

	function openEvidenceUrl(url: string) {
		const trimmed = url.trim();
		if (!trimmed) return;
		window.open(trimmed, '_blank', 'noopener,noreferrer');
	}

	function exportActionsCsv() {
		const rows = [
			['Tahun', 'Judul', 'Sumber', 'SNP', 'Prioritas', 'Status', 'Tenggat', 'PIC', 'Unit', 'Kaitan', 'Uraian', 'Catatan', 'Bukti'],
			...filteredActions.map((action) => [
				action.period_year,
				action.title,
				sourceLabel(action.source_type),
				snpLabel(action.snp_standard),
				priorityLabel(action.priority),
				statusLabel(action.status),
				formatDate(action.due_date),
				action.responsible_employee_name || '',
				action.owner_unit_name || '',
				linkedLabel(action),
				action.description,
				action.follow_up_notes,
				action.evidence_url,
			]),
		];
		downloadCsv(`tindak-lanjut-kepatuhan-${new Date().toISOString().slice(0, 10)}.csv`, rows);
	}

	onMount(() => {
		void loadOverview();
	});
</script>

<svelte:head><title>Tindak Lanjut Tata Kelola - MTsN 2 Kolaka Utara</title></svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
		<div>
			<h1 class="text-lg font-semibold text-slate-800">Tindak Lanjut Kepatuhan</h1>
			<p class="text-sm text-slate-500">Catat PIC, tenggat, bukti, dan status dari gap Renstra/IKU/RKT/SKP/8 SNP.</p>
		</div>
		<div class="flex flex-wrap gap-2">
			<Button href={resolve('/governance')} variant="outline" size="sm">
				<ArrowLeftIcon class="mr-2 size-4" />
				Tata Kelola
			</Button>
			<LoadingButton variant="outline" size="sm" loading={refreshBusy} loadingLabel="Memuat..." onclick={() => void refreshOverviewAction()}>
				<RefreshCcwIcon class="mr-2 size-4" />
				Refresh
			</LoadingButton>
			<Button variant="outline" size="sm" onclick={exportActionsCsv}>
				<DownloadIcon class="mr-2 size-4" />
				Export CSV
			</Button>
			<Button href={resolve('/governance/print-pack')} variant="outline" size="sm">
				<PrinterIcon class="mr-2 size-4" />
				Paket Cetak
			</Button>
			<Button href={resolve('/governance/actions/meeting-pack')} variant="outline" size="sm">
				<ClipboardCheckIcon class="mr-2 size-4" />
				Paket Rapat
			</Button>
			<Button href={resolve('/governance/actions/owner-briefing')} variant="outline" size="sm">
				<ClipboardCheckIcon class="mr-2 size-4" />
				Briefing PIC
			</Button>
			<Button href={resolve('/governance/actions/snp-briefing')} variant="outline" size="sm">
				<ClipboardCheckIcon class="mr-2 size-4" />
				Briefing SNP
			</Button>
			<Button href={resolve('/governance/actions/evidence-briefing')} variant="outline" size="sm">
				<ClipboardCheckIcon class="mr-2 size-4" />
				Briefing Bukti
			</Button>
			<Button href={resolve('/governance/actions/calendar')} variant="outline" size="sm">
				<CalendarDaysIcon class="mr-2 size-4" />
				Kalender
			</Button>
			<Button size="sm" onclick={openCreate}>
				<ClipboardCheckIcon class="mr-2 size-4" />
				Tambah Tindak Lanjut
			</Button>
		</div>
	</div>

	<AsyncContent promise={overviewPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<div class="grid gap-3 md:grid-cols-4">
					{#each Array.from({ length: 4 }) as _, index (`compliance-stat-skeleton-${index}`)}
						<Skeleton class="h-24 w-full rounded-lg" />
					{/each}
				</div>
				<Skeleton class="h-64 w-full rounded-lg" />
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel title="Tindak Lanjut Belum Tersaji" message={errorMessage(error)} onRetry={() => retryOverview(reset)} />
		{/snippet}

		{#snippet children(value)}
			{@const loaded = value as ComplianceOverview}
			{@const loadedStats = loaded.stats}
			<div class="grid gap-3 md:grid-cols-4">
				{#each [
					{ label: 'Total Tindak Lanjut', value: loadedStats.compliance_actions ?? actions.length, note: 'Semua periode' },
					{ label: 'Masih Terbuka', value: loadedStats.open_compliance_actions ?? openCount, note: 'Perlu gerak PIC' },
					{ label: 'Prioritas Tinggi', value: loadedStats.critical_compliance_actions ?? criticalOpenCount, note: 'High/urgent' },
					{ label: 'Lewat Tenggat', value: overdueCount, note: 'Butuh eskalasi' },
				] as item (item.label)}
					<Card.Root class="border-slate-200">
						<Card.Content class="p-4">
							<p class="text-xs text-slate-500">{item.label}</p>
							<p class="mt-1 text-2xl font-semibold text-emerald-800">{item.value}</p>
							<p class="text-xs text-slate-500">{item.note}</p>
						</Card.Content>
					</Card.Root>
				{/each}
			</div>

			<Card.Root class="border-slate-200">
				<Card.Content class="flex flex-col gap-3 p-4 lg:flex-row lg:items-center lg:justify-between">
					<div>
						<p class="text-sm font-medium text-slate-800">Filter Cepat Eskalasi</p>
						<p class="text-xs text-slate-500">
							{registerFocus ? `Fokus aktif: ${registerFocus.label}` : 'Pakai untuk rapat tindak lanjut singkat kepala/staf.'}
						</p>
					</div>
					<div class="flex flex-wrap gap-2">
						<Button size="sm" variant={focusVariant('risk', 'overdue')} onclick={() => setRegisterFocus({ kind: 'risk', value: 'overdue', label: 'Lewat tenggat' })}>
							Lewat Tenggat ({overdueCount})
						</Button>
						<Button size="sm" variant={focusVariant('risk', 'critical')} onclick={() => setRegisterFocus({ kind: 'risk', value: 'critical', label: 'Prioritas tinggi/mendesak' })}>
							Prioritas Tinggi ({criticalOpenCount})
						</Button>
						<Button size="sm" variant={focusVariant('risk', 'unassigned')} onclick={() => setRegisterFocus({ kind: 'risk', value: 'unassigned', label: 'Belum ada PIC' })}>
							Tanpa PIC ({unassignedCount})
						</Button>
						<Button size="sm" variant={focusVariant('risk', 'waiting_evidence')} onclick={() => setRegisterFocus({ kind: 'risk', value: 'waiting_evidence', label: 'Menunggu bukti' })}>
							Menunggu Bukti ({waitingEvidenceCount})
						</Button>
						{#if registerFocus}
							<Button size="sm" variant="outline" onclick={clearRegisterFocus}>Lepas Fokus</Button>
						{/if}
					</div>
				</Card.Content>
			</Card.Root>

			<div class="grid gap-4 xl:grid-cols-[360px_1fr]">
				<div class="space-y-4">
					<Card.Root class="border-slate-200">
						<Card.Header class="pb-2">
							<Card.Title class="text-sm font-medium text-slate-700">Saran dari Gap</Card.Title>
							<Card.Description>Diambil dari program yang belum lengkap pemetaan IKU/RKT/SKP/bukti.</Card.Description>
						</Card.Header>
						<Card.Content class="space-y-2">
							{#if gapSuggestions.length === 0}
								<EmptyStatePanel compact title="Tidak ada saran baru" description="Semua gap utama sudah punya tindak lanjut aktif atau belum ada program yang perlu dipetakan." />
							{:else}
								{#each gapSuggestions as suggestion (suggestion.key)}
									<div class="rounded-md border border-slate-200 px-3 py-2">
										<p class="text-sm font-medium text-slate-900">{suggestion.gap}</p>
										<p class="mt-1 text-xs text-slate-500">{suggestion.program.code} · {suggestion.program.name}</p>
										<div class="mt-2 flex items-center justify-between gap-2">
											<Badge variant={priorityVariant(suggestion.priority)}>{priorityLabel(suggestion.priority)}</Badge>
											<Button size="sm" variant="outline" onclick={() => openSuggestion(suggestion)}>Jadikan</Button>
										</div>
									</div>
								{/each}
							{/if}
						</Card.Content>
					</Card.Root>

					<Card.Root class="border-slate-200">
						<Card.Header class="pb-2">
							<Card.Title class="text-sm font-medium text-slate-700">Eskalasi PIC/Unit</Card.Title>
							<Card.Description>Urutan unit atau PIC yang punya beban tindak lanjut aktif.</Card.Description>
						</Card.Header>
						<Card.Content class="space-y-2">
							{#if escalationRows.length === 0}
								<EmptyStatePanel compact title="Tidak ada beban aktif" description="Semua tindak lanjut sudah selesai/dibatalkan atau belum ada register." />
							{:else}
								{#each escalationRows as row (row.key)}
									<div class="rounded-md border border-slate-200 px-3 py-2">
										<div class="flex items-start justify-between gap-2">
											<div>
												<p class="text-sm font-medium text-slate-900">{row.label}</p>
												<p class="text-xs text-slate-500">{row.unitLabel}</p>
											</div>
											<Button size="sm" variant={focusVariant('owner', row.key)} onclick={() => setRegisterFocus({ kind: 'owner', value: row.key, label: row.label })}>Fokus</Button>
										</div>
										<div class="mt-2 flex flex-wrap gap-1">
											<Badge variant="secondary">{row.openCount} terbuka</Badge>
											{#if row.criticalCount > 0}<Badge variant="destructive">{row.criticalCount} tinggi</Badge>{/if}
											{#if row.overdueCount > 0}<Badge variant="destructive">{row.overdueCount} lewat</Badge>{/if}
											{#if row.waitingEvidenceCount > 0}<Badge variant="outline">{row.waitingEvidenceCount} bukti</Badge>{/if}
										</div>
										<p class="mt-2 text-xs text-slate-500">Tenggat terdekat: {dueLabel(row.nextDueDate)}</p>
									</div>
								{/each}
							{/if}
						</Card.Content>
					</Card.Root>

					<Card.Root class="border-slate-200">
						<Card.Header class="pb-2">
							<Card.Title class="text-sm font-medium text-slate-700">Sebaran 8 SNP</Card.Title>
							<Card.Description>Standar yang masih punya tindak lanjut aktif.</Card.Description>
						</Card.Header>
						<Card.Content class="space-y-2">
							{#if snpEscalationRows.length === 0}
								<EmptyStatePanel compact title="Tidak ada beban SNP" description="Belum ada tindak lanjut aktif yang dikaitkan ke standar." />
							{:else}
								{#each snpEscalationRows as row (row.code)}
									<div class="flex items-center justify-between gap-3 rounded-md border border-slate-200 px-3 py-2">
										<div>
											<p class="text-sm font-medium text-slate-900">{row.label}</p>
											<p class="text-xs text-slate-500">
												{row.openCount} terbuka · {row.criticalCount} tinggi · {row.overdueCount} lewat · {row.waitingEvidenceCount} tunggu bukti
											</p>
										</div>
										<Button size="sm" variant={focusVariant('snp', row.code)} onclick={() => setRegisterFocus({ kind: 'snp', value: row.code, label: row.label })}>Fokus</Button>
									</div>
								{/each}
							{/if}
						</Card.Content>
					</Card.Root>
				</div>

				<Card.Root class="border-slate-200">
					<Card.Header class="pb-2">
						<Card.Title class="text-sm font-medium text-slate-700">Register Tindak Lanjut</Card.Title>
						<Card.Description>Daftar kerja korektif dari gap tata kelola madrasah.</Card.Description>
					</Card.Header>
					<Card.Content class="space-y-4">
						<div class="space-y-2">
							<div class="grid gap-2 md:grid-cols-2 xl:grid-cols-4">
								<div class="md:col-span-2 xl:col-span-4">
									<label for="action-search" class="sr-only">Cari tindak lanjut</label>
									<Input id="action-search" placeholder="Cari judul, program, PIC, catatan, atau bukti" bind:value={search} />
								</div>
								<div>
									<label for="action-status-filter" class="sr-only">Filter status tindak lanjut</label>
									<select id="action-status-filter" bind:value={statusFilter} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
										<option value="">Semua status</option>
										{#each STATUSES as [value, label] (value)}<option {value}>{label}</option>{/each}
									</select>
								</div>
								<div>
									<label for="action-priority-filter" class="sr-only">Filter prioritas tindak lanjut</label>
									<select id="action-priority-filter" bind:value={priorityFilter} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
										<option value="">Semua prioritas</option>
										{#each PRIORITIES as [value, label] (value)}<option {value}>{label}</option>{/each}
									</select>
								</div>
								<div>
									<label for="action-year-filter" class="sr-only">Filter tahun tindak lanjut</label>
									<select id="action-year-filter" bind:value={yearFilter} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
										<option value="">Semua tahun</option>
										{#each actionYears as year (year)}<option value={String(year)}>{year}</option>{/each}
									</select>
								</div>
								<div>
									<label for="action-source-filter" class="sr-only">Filter sumber tindak lanjut</label>
									<select id="action-source-filter" bind:value={sourceFilter} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
										<option value="">Semua sumber</option>
										{#each SOURCE_TYPES as [value, label] (value)}<option {value}>{label}</option>{/each}
									</select>
								</div>
								<div>
									<label for="action-snp-filter" class="sr-only">Filter standar SNP tindak lanjut</label>
									<select id="action-snp-filter" bind:value={snpFilter} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
										<option value="">Semua SNP</option>
										{#each SNP_STANDARDS as [value, label] (value)}<option {value}>{label}</option>{/each}
									</select>
								</div>
								<div class="md:col-span-2 xl:col-span-3">
									<label for="action-owner-filter" class="sr-only">Filter PIC atau unit tindak lanjut</label>
									<select id="action-owner-filter" bind:value={ownerFilter} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
										<option value="">Semua PIC/unit</option>
										{#each ownerFilterOptions as owner (owner.key)}
											<option value={owner.key}>{owner.label} · {owner.unitLabel}</option>
										{/each}
									</select>
								</div>
							</div>
							<div class="flex flex-col gap-2 text-xs text-slate-500 sm:flex-row sm:items-center sm:justify-between">
								<p>
									Menampilkan {filteredActions.length} dari {actions.length} tindak lanjut{registerFocus ? ` · ${registerFocus.label}` : ''}.
								</p>
								{#if activeFilterCount > 0}
									<Button size="sm" variant="outline" onclick={clearRegisterFilters}>Reset Filter ({activeFilterCount})</Button>
								{/if}
							</div>
						</div>

						{#if filteredActions.length === 0}
							<EmptyStatePanel title="Belum ada tindak lanjut" description="Tambahkan tindak lanjut manual atau jadikan saran gap sebagai pekerjaan yang punya PIC dan tenggat." />
						{:else}
							<div class="overflow-x-auto">
								<Table.Root>
									<Table.Header>
										<Table.Row>
											<Table.Head>Tindak Lanjut</Table.Head>
											<Table.Head>Kaitan</Table.Head>
											<Table.Head>PIC</Table.Head>
											<Table.Head>Status</Table.Head>
											<Table.Head>Tenggat</Table.Head>
											<Table.Head class="w-56">Aksi</Table.Head>
										</Table.Row>
									</Table.Header>
									<Table.Body>
										{#each filteredActions as action (action.id)}
											<Table.Row>
												<Table.Cell class="min-w-[280px]">
													<p class="text-sm font-medium text-slate-900">{action.title}</p>
													<p class="text-xs text-slate-500">{action.period_year} · {sourceLabel(action.source_type)} · {snpLabel(action.snp_standard)}</p>
													{#if action.description}
														<p class="mt-1 text-xs text-slate-500">{action.description}</p>
													{/if}
												</Table.Cell>
												<Table.Cell class="min-w-[230px] text-sm text-slate-700">{linkedLabel(action)}</Table.Cell>
												<Table.Cell class="min-w-[170px]">
													<p class="text-sm">{action.responsible_employee_name || '-'}</p>
													<p class="text-xs text-slate-500">{action.owner_unit_name || action.responsible_employee_nip || 'Belum ditetapkan'}</p>
												</Table.Cell>
												<Table.Cell>
													<div class="flex flex-col gap-1">
														<Badge variant={statusVariant(action.status)}>{statusLabel(action.status)}</Badge>
														<Badge variant={priorityVariant(action.priority)}>{priorityLabel(action.priority)}</Badge>
													</div>
												</Table.Cell>
												<Table.Cell class="whitespace-nowrap text-sm">
													<span class={isOverdue(action) ? 'font-medium text-red-700' : 'text-slate-700'}>{formatDate(action.due_date)}</span>
													{#if action.evidence_url}
														<Button size="sm" variant="outline" onclick={() => openEvidenceUrl(action.evidence_url)}>Bukti</Button>
													{/if}
												</Table.Cell>
												<Table.Cell>
													<div class="flex flex-wrap gap-1">
														{#each quickStatusOptions(action) as option (option.status)}
															<LoadingButton
																size="sm"
																variant={option.variant}
																loading={quickStatusBusy[action.id] ?? false}
																loadingLabel="..."
																onclick={() => void quickUpdateStatus(action, option.status)}
															>
																{option.label}
															</LoadingButton>
														{/each}
														<Button size="sm" variant="outline" onclick={() => openEvidenceCapture(action)}>Catat Bukti</Button>
														<Button size="sm" variant="outline" onclick={() => openEdit(action)}>Edit</Button>
														<Button size="sm" variant="destructive" onclick={() => void deleteAction(action)}>Hapus</Button>
													</div>
												</Table.Cell>
											</Table.Row>
										{/each}
									</Table.Body>
								</Table.Root>
							</div>
						{/if}
					</Card.Content>
				</Card.Root>
			</div>
		{/snippet}
	</AsyncContent>
</div>

<Dialog.Root bind:open={dialogOpen}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>{editingId ? 'Edit Tindak Lanjut' : 'Tambah Tindak Lanjut'}</Dialog.Title>
			<Dialog.Description>Tetapkan sumber gap, PIC, tenggat, status, dan bukti penyelesaian.</Dialog.Description>
		</Dialog.Header>

		<div class="mt-4 space-y-4">
			<div class="grid gap-3 sm:grid-cols-2">
				<div>
					<label for="action-year" class="text-sm font-medium">Tahun</label>
					<Input id="action-year" type="number" bind:value={actionForm.period_year} />
				</div>
				<div>
					<label for="action-source" class="text-sm font-medium">Sumber</label>
					<select id="action-source" bind:value={actionForm.source_type} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
						{#each SOURCE_TYPES as [value, label] (value)}<option {value}>{label}</option>{/each}
					</select>
				</div>
				<div class="sm:col-span-2">
					<label for="action-title" class="text-sm font-medium">Judul Tindak Lanjut</label>
					<Input id="action-title" bind:value={actionForm.title} />
				</div>
				<div>
					<label for="action-priority" class="text-sm font-medium">Prioritas</label>
					<select id="action-priority" bind:value={actionForm.priority} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
						{#each PRIORITIES as [value, label] (value)}<option {value}>{label}</option>{/each}
					</select>
				</div>
				<div>
					<label for="action-status" class="text-sm font-medium">Status</label>
					<select id="action-status" bind:value={actionForm.status} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
						{#each STATUSES as [value, label] (value)}<option {value}>{label}</option>{/each}
					</select>
				</div>
				<div>
					<label for="action-snp" class="text-sm font-medium">Standar SNP</label>
					<select id="action-snp" bind:value={actionForm.snp_standard} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
						{#each SNP_STANDARDS as [value, label] (value)}<option {value}>{label}</option>{/each}
					</select>
				</div>
				<div>
					<label for="action-due" class="text-sm font-medium">Tenggat</label>
					<Input id="action-due" type="date" bind:value={actionForm.due_date} />
				</div>
				<div>
					<label for="action-program" class="text-sm font-medium">Program</label>
					<select id="action-program" bind:value={actionForm.program_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
						<option value="">Tanpa program</option>
						{#each programs as program (program.id)}<option value={program.id}>{program.code} · {program.name}</option>{/each}
					</select>
				</div>
				<div>
					<label for="action-document" class="text-sm font-medium">Dokumen</label>
					<select id="action-document" bind:value={actionForm.document_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
						<option value="">Tanpa dokumen</option>
						{#each documents as document (document.id)}<option value={document.id}>{document.title}</option>{/each}
					</select>
				</div>
				<div>
					<label for="action-performance" class="text-sm font-medium">Target SKP</label>
					<select id="action-performance" bind:value={actionForm.performance_target_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
						<option value="">Tanpa target</option>
						{#each performanceTargets as target (target.id)}<option value={target.id}>{target.title}</option>{/each}
					</select>
				</div>
				<div>
					<label for="action-evidence-item" class="text-sm font-medium">Bukti Mutu</label>
					<select id="action-evidence-item" bind:value={actionForm.evidence_item_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
						<option value="">Tanpa bukti</option>
						{#each evidenceItems as item (item.id)}<option value={item.id}>{item.title}</option>{/each}
					</select>
				</div>
				<div>
					<label for="action-unit" class="text-sm font-medium">Unit Pemilik</label>
					<select id="action-unit" bind:value={actionForm.owner_unit_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
						<option value="">Tanpa unit</option>
						{#each units as unit (unit.id)}<option value={unit.id}>{unit.name}</option>{/each}
					</select>
				</div>
				<div>
					<label for="action-employee" class="text-sm font-medium">PIC Pegawai</label>
					<select id="action-employee" bind:value={actionForm.responsible_employee_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
						<option value="">Tanpa PIC</option>
						{#each employees as employee (employee.id)}<option value={employee.id}>{employee.nama} · {employee.nip}</option>{/each}
					</select>
				</div>
				<div class="sm:col-span-2">
					<label for="action-description" class="text-sm font-medium">Uraian Gap/Tindak Lanjut</label>
					<Textarea id="action-description" bind:value={actionForm.description} />
				</div>
				<div class="sm:col-span-2">
					<label for="action-follow-up" class="text-sm font-medium">Catatan Tindak Lanjut</label>
					<Textarea id="action-follow-up" bind:value={actionForm.follow_up_notes} />
				</div>
				<div class="sm:col-span-2">
					<label for="action-evidence-url" class="text-sm font-medium">URL/Lokasi Bukti</label>
					<Input id="action-evidence-url" bind:value={actionForm.evidence_url} />
				</div>
			</div>
		</div>

		<Dialog.Footer>
			<Button variant="outline" onclick={() => (dialogOpen = false)}>Batal</Button>
			<LoadingButton loading={busy} onclick={() => void saveAction()}>Simpan</LoadingButton>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<Dialog.Root bind:open={evidenceDialogOpen}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>Catat Bukti Tindak Lanjut</Dialog.Title>
			<Dialog.Description>
				{#if evidenceAction}
					{evidenceAction.title}
				{:else}
					Catat tautan, item bukti, dan catatan penyelesaian.
				{/if}
			</Dialog.Description>
		</Dialog.Header>

		<div class="mt-4 space-y-4">
			<div>
				<label for="evidence-capture-item" class="text-sm font-medium">Item Bukti Mutu</label>
				<select id="evidence-capture-item" bind:value={evidenceForm.evidence_item_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
					<option value="">Tanpa item bukti</option>
					{#each evidenceItems as item (item.id)}<option value={item.id}>{item.title}</option>{/each}
				</select>
			</div>
			<div>
				<label for="evidence-capture-url" class="text-sm font-medium">URL/Lokasi Bukti</label>
				<Input id="evidence-capture-url" bind:value={evidenceForm.evidence_url} placeholder="Tautan arsip, folder, dokumen, atau lokasi fisik" />
			</div>
			<div>
				<label for="evidence-capture-notes" class="text-sm font-medium">Catatan Tindak Lanjut</label>
				<Textarea id="evidence-capture-notes" bind:value={evidenceForm.follow_up_notes} rows={4} placeholder="Ringkas hasil tindak lanjut, keputusan rapat, atau kondisi bukti." />
			</div>
			{#if evidenceAction}
				<div class="rounded-md border border-slate-200 bg-slate-50 px-3 py-2 text-xs text-slate-600">
					<p>Status saat ini: {statusLabel(evidenceAction.status)} · Prioritas: {priorityLabel(evidenceAction.priority)}</p>
					<p>Kaitan: {linkedLabel(evidenceAction)}</p>
				</div>
			{/if}
		</div>

		<Dialog.Footer>
			<Button variant="outline" onclick={() => (evidenceDialogOpen = false)}>Batal</Button>
			<LoadingButton loading={evidenceBusy} variant="outline" onclick={() => void saveEvidenceCapture(false)}>Simpan Bukti</LoadingButton>
			<LoadingButton loading={evidenceBusy} onclick={() => void saveEvidenceCapture(true)}>Simpan & Selesai</LoadingButton>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
