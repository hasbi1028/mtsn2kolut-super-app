<script lang="ts">
	import { onMount } from 'svelte';
	import { SvelteMap } from 'svelte/reactivity';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import OperationStatusPanel from '$lib/components/OperationStatusPanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { confirmChallenge } from '$lib/confirm-dialog';
	import { clientApiPath, clientApiPathWithQuery, readClientApiData, readClientJson } from '$lib/client/api';

	type ExamSession = {
		id: string; package_id: string; package_title: string;
		class_id: string; class_name: string; class_code: string;
		scope_type: string; scope_ref: string; mix_policy: string; assignment_mode: string;
		allow_cross_grade: boolean; is_special_event: boolean;
		title: string; scheduled_start: string; scheduled_end: string;
		status: string; participant_count: number; created_at: string;
		event_id?: string | null;
		room_count: number; total_capacity: number;
		assigned_participant_count: number; unassigned_participant_count: number;
		missing_seat_count: number; rooms_without_proctor: number; proctor_assignment_count: number;
	};
	type CbtPackage = {
		id: string; title: string; subject_code: string; subject_name: string;
		question_count: number; is_active: boolean;
		event_id?: string | null;
	};
	type PackageQuestion = {
		package_id: string; question_id: string;
		question_type?: string; status?: string; cp_ref?: string; tp_ref?: string; kd_ref?: string;
		cognitive_level?: string; hots_flag?: boolean;
	};
	type PackageQualitySummary = {
		questions: PackageQuestion[];
		typeBuckets: { label: string; count: number }[];
		hotsCount: number;
		missingCount: number;
		unpublishedCount: number;
		totalCount: number;
	};
	type NextSessionAction = {
		label: string;
		href?: string;
		kind: 'enroll' | 'link';
		tone: 'primary' | 'warning' | 'danger';
	};
	type SessionReadinessFilter =
		| 'all'
		| 'not_ready'
		| 'ready'
		| 'needs_participants'
		| 'needs_rooms'
		| 'needs_proctors';
	type ReadinessBoardTone = 'danger' | 'success' | 'warning' | 'info';
	type ReadinessBoardConfig = {
		filter: Exclude<SessionReadinessFilter, 'all'>;
		label: string;
		helper: string;
		tone: ReadinessBoardTone;
	};
	type SessionScheduleFilter = 'all' | 'today' | 'upcoming' | 'overdue';
	type SessionScheduleState = 'today' | 'upcoming' | 'overdue' | 'running_window' | 'closed' | 'unknown';
	type ScheduleBoardTone = 'danger' | 'success' | 'info';
	type ScheduleBoardConfig = {
		filter: Exclude<SessionScheduleFilter, 'all'>;
		label: string;
		helper: string;
		tone: ScheduleBoardTone;
	};
	type ScheduleQuickAction = {
		label: string;
		kind: 'status' | 'schedule' | 'link';
		status?: 'cancelled' | 'finished';
		href?: string;
		tone: NextSessionAction['tone'];
	};
	type SchoolClass = { id: string; name: string; code: string; level: string; };
	type EventContext = { id: string; title: string; status: string; target_levels?: string[]; academic_year_name?: string; };
	type SessionsOverview = {
		sessions: ExamSession[];
		packages: CbtPackage[];
		packageQuestions: PackageQuestion[];
		classes: SchoolClass[];
	};
	type CbtPackagesPayload = {
		packages?: CbtPackage[];
		questions?: unknown[];
		error?: string;
		message?: string;
	};
	type AcademicPayload = {
		classes?: SchoolClass[];
		error?: string;
		message?: string;
	};

	let sessions = $state<ExamSession[]>([]);
	let packages = $state<CbtPackage[]>([]);
	let packageQuestions = $state<PackageQuestion[]>([]);
	let classes = $state<SchoolClass[]>([]);
	let sessionsPromise = $state<Promise<SessionsOverview> | null>(null);
	let showForm = $state(false);
	let readinessFilter = $state<SessionReadinessFilter>('all');
	let scheduleFilter = $state<SessionScheduleFilter>('all');

	let fPackageId = $state('');
	let fScopeType = $state('class');
	let fClassId = $state('');
	let fGradeLevel = $state('VII');
	let fMixPolicy = $state('same_class');
	let fAssignmentMode = $state('random_balanced');
	let fAllowCrossGrade = $state(false);
	let fIsSpecialEvent = $state(false);
	let fTitle = $state('');
	let fStart = $state('');
	let fEnd = $state('');
	let fBusy = $state(false);
	let statusBusyId = $state('');
	let deleteBusyId = $state('');
	let operationState = $state<{ tone: 'success' | 'error' | 'warning' | 'info'; title: string; message: string } | null>(null);
	let sessionsRequestId = 0;
	let hiddenEventSessionCount = $state(0);
	let hiddenEventPackageCount = $state(0);
	let eventContext = $state<EventContext | null>(null);
	let browserTimeZone = $state('');
	const eventId = page.url.searchParams.get('event_id') ?? '';

	// Enroll modal
	let enrollSession = $state<ExamSession | null>(null);
	let enrollScopeType = $state('class');
	let enrollClassId = $state('');
	let enrollGradeLevel = $state('VII');
	let enrollBusy = $state(false);
	let scheduleSession = $state<ExamSession | null>(null);
	let scheduleStart = $state('');
	let scheduleEnd = $state('');
	let scheduleBusy = $state(false);
	let selectedPackage = $derived(packages.find((pkg) => pkg.id === fPackageId) ?? null);
	let selectedPackageQuality = $derived(packageQualitySummary(fPackageId));
	let sessionReadinessIssues = $derived(buildSessionReadinessIssues());
	let canCreateSession = $derived(!fBusy && sessionReadinessIssues.length === 0);
	let browserTimeZoneMismatch = $derived(browserTimeZone !== '' && browserTimeZone !== 'Asia/Makassar');
	let createSessionHref = $derived(`${resolve('/asesmen/sesi/new')}${eventId ? `?event_id=${encodeURIComponent(eventId)}` : ''}`);

	const statusLabel: Record<string, string> = {
		draft: 'Draft', scheduled: 'Terjadwal', active: 'Berlangsung',
		finished: 'Selesai', cancelled: 'Dibatalkan',
	};
	const readinessFilters: SessionReadinessFilter[] = [
		'all',
		'not_ready',
		'ready',
		'needs_participants',
		'needs_rooms',
		'needs_proctors',
	];
	const readinessBoardConfigs: ReadinessBoardConfig[] = [
		{ filter: 'not_ready', label: 'Belum Siap', helper: 'Masih punya kendala operasional', tone: 'danger' },
		{ filter: 'ready', label: 'Siap Mulai', helper: 'Paket, peserta, ruang, pengawas siap', tone: 'success' },
		{ filter: 'needs_participants', label: 'Butuh Peserta', helper: 'Peserta belum didaftarkan', tone: 'warning' },
		{ filter: 'needs_rooms', label: 'Butuh Ruang', helper: 'Ruang, kursi, atau kapasitas belum rapi', tone: 'warning' },
		{ filter: 'needs_proctors', label: 'Butuh Pengawas', helper: 'Ruang ujian belum lengkap pengawas', tone: 'warning' },
	];
	const scheduleFilters: SessionScheduleFilter[] = ['all', 'today', 'upcoming', 'overdue'];
	const scheduleBoardConfigs: ScheduleBoardConfig[] = [
		{ filter: 'today', label: 'Hari Ini', helper: 'Perlu dipantau hari ini', tone: 'success' },
		{ filter: 'upcoming', label: 'Akan Datang', helper: 'Masih sebelum tanggal mulai', tone: 'info' },
		{ filter: 'overdue', label: 'Lewat Jadwal', helper: 'Belum selesai setelah waktu tutup', tone: 'danger' },
	];
	const witaDateFormatter = new Intl.DateTimeFormat('en-CA', {
		timeZone: 'Asia/Makassar',
		year: 'numeric',
		month: '2-digit',
		day: '2-digit',
	});

	function statusClass(s: string) {
		if (s === 'active') return 'bg-primary/15 text-primary border-primary/20';
		if (s === 'finished') return 'bg-muted text-muted-foreground border-border';
		if (s === 'cancelled') return 'bg-destructive/15 text-destructive border-destructive/30';
		if (s === 'scheduled') return 'bg-success/15 text-success border-success/20';
		return 'bg-warning/15 text-warning border-warning/30';
	}

	function fmtDt(iso: string) {
		if (!iso) return '—';
		return new Date(iso).toLocaleString('id-ID', {
			timeZone: 'Asia/Makassar', year: 'numeric', month: 'short',
			day: 'numeric', hour: '2-digit', minute: '2-digit',
		}) + ' WITA';
	}

	function toRFC3339(localDt: string): string {
		if (!localDt) return '';
		return new Date(localDt).toISOString();
	}

	function detectBrowserTimeZone() {
		try {
			browserTimeZone = Intl.DateTimeFormat().resolvedOptions().timeZone || '';
		} catch {
			browserTimeZone = '';
		}
	}

	function toLocalDateTimeInput(iso: string) {
		if (!iso) return '';
		const date = new Date(iso);
		if (Number.isNaN(date.getTime())) return '';
		const local = new Date(date.getTime() - date.getTimezoneOffset() * 60_000);
		return local.toISOString().slice(0, 16);
	}

	function scopeSummary(session: ExamSession) {
		if (session.scope_type === 'grade') return `Tingkat ${session.scope_ref || '—'}`;
		if (session.scope_type === 'school') return 'Seluruh sekolah';
		if (session.scope_type === 'custom') return session.scope_ref || 'Peserta khusus';
		return session.class_code || session.class_name || 'Per kelas';
	}

	function mixPolicyLabel(value: string) {
		if (value === 'same_class') return 'Tetap per kelas';
		if (value === 'mixed_scope') return 'Campur lintas cakupan';
		return 'Campur dalam tingkat';
	}

	function adaptiveMixPolicy(scopeType: string) {
		if (scopeType === 'class') return 'same_class';
		if (scopeType === 'grade') return 'same_grade';
		return 'mixed_scope';
	}

	function updateScopeType(value: string) {
		fScopeType = value;
		fMixPolicy = adaptiveMixPolicy(value);
	}

	function sessionActionLabel(status: string) {
		if (status === 'draft') return 'Draft';
		if (status === 'scheduled') return 'Terjadwal';
		if (status === 'active') return 'Berlangsung';
		if (status === 'finished') return 'Selesai';
		return 'Dibatalkan';
	}

	function isRecord(value: unknown): value is Record<string, unknown> {
		return typeof value === 'object' && value !== null;
	}

	function parsePackageQuestions(payload: CbtPackagesPayload | unknown) {
		return isRecord(payload) && Array.isArray(payload.questions) ? (payload.questions as PackageQuestion[]) : [];
	}

	function strictEventSessions(items: ExamSession[]) {
		return eventId ? items.filter((session) => session.event_id === eventId) : items;
	}

	function strictEventPackages(items: CbtPackage[]) {
		return eventId ? items.filter((pkg) => pkg.event_id === eventId) : items;
	}

	function hiddenSessionCount(items: ExamSession[]) {
		return eventId ? items.filter((session) => session.event_id !== eventId).length : 0;
	}

	function hiddenPackageCount(items: CbtPackage[]) {
		return eventId ? items.filter((pkg) => pkg.event_id !== eventId).length : 0;
	}

	function compactValue(value: string | number | null | undefined, fallback: string) {
		const text = value === null || value === undefined ? '' : String(value).trim();
		return text || fallback;
	}

	function questionTypeLabel(value: string | null | undefined) {
		const labels: Record<string, string> = {
			multiple_choice: 'PG',
			multiple_answer: 'PG Kompleks',
			true_false: 'Benar/Salah',
			agree_disagree: 'Setuju/Tidak',
			matching: 'Menjodohkan',
			short_answer: 'Isian',
			essay: 'Essay',
		};
		const normalized = compactValue(value, '');
		return labels[normalized] ?? (normalized ? normalized.replaceAll('_', ' ') : 'Belum tipe');
	}

	function packageQuestionHasBlueprintGap(question: PackageQuestion) {
		return !compactValue(question.cp_ref, '')
			|| (!compactValue(question.tp_ref, '') && !compactValue(question.kd_ref, ''))
			|| !compactValue(question.cognitive_level, '');
	}

	function countByLabel<T>(items: T[], selector: (item: T) => string) {
		const counts = new SvelteMap<string, number>();
		for (const item of items) {
			const label = selector(item);
			counts.set(label, (counts.get(label) ?? 0) + 1);
		}
		return Array.from(counts.entries())
			.map(([label, count]) => ({ label, count }))
			.sort((a, b) => b.count - a.count || a.label.localeCompare(b.label));
	}

	function packageQualitySummary(packageID: string): PackageQualitySummary {
		const pkg = packages.find((item) => item.id === packageID);
		const questions = packageQuestions.filter((question) => question.package_id === packageID);
		return {
			questions,
			typeBuckets: countByLabel(questions, (question) => questionTypeLabel(question.question_type)),
			hotsCount: questions.filter((question) => question.hots_flag).length,
			missingCount: questions.filter(packageQuestionHasBlueprintGap).length,
			unpublishedCount: questions.filter((question) => question.status !== 'published').length,
			totalCount: pkg?.question_count ?? questions.length,
		};
	}

	function packageQualityIssues(packageID: string) {
		const issues: string[] = [];
		const pkg = packages.find((item) => item.id === packageID);
		const quality = packageQualitySummary(packageID);
		if (pkg && !pkg.is_active) issues.push('paket nonaktif');
		if (packageID && quality.totalCount === 0) issues.push('paket kosong');
		if (packageID && quality.questions.length > 0 && quality.unpublishedCount > 0) {
			issues.push(`${quality.unpublishedCount} belum terbit`);
		}
		return issues;
	}

	function sessionOperationalIssues(session: ExamSession) {
		const issues: string[] = [];
		if (session.participant_count === 0) issues.push('belum ada peserta');
		if (session.room_count === 0) issues.push('belum ada ruang');
		if (session.participant_count > 0 && session.total_capacity < session.participant_count) issues.push('kapasitas kurang');
		if (session.unassigned_participant_count > 0) issues.push(`${session.unassigned_participant_count} belum ruang`);
		if (session.missing_seat_count > 0) issues.push(`${session.missing_seat_count} belum meja`);
		if (session.rooms_without_proctor > 0) issues.push(`${session.rooms_without_proctor} ruang tanpa pengawas`);
		return issues;
	}

	function sessionRowReadinessIssues(session: ExamSession) {
		return [...packageQualityIssues(session.package_id), ...sessionOperationalIssues(session)];
	}

	function sessionNeedsParticipants(session: ExamSession) {
		return session.participant_count === 0;
	}

	function sessionNeedsRooms(session: ExamSession) {
		return session.participant_count > 0 && (
			session.room_count === 0
			|| session.total_capacity < session.participant_count
			|| session.unassigned_participant_count > 0
			|| session.missing_seat_count > 0
		);
	}

	function sessionNeedsProctors(session: ExamSession) {
		return session.room_count > 0 && session.rooms_without_proctor > 0;
	}

	function sessionHasBlockingIssues(session: ExamSession) {
		return packageQualityIssues(session.package_id).length > 0 || sessionOperationalIssues(session).length > 0;
	}

	function sessionMatchesReadinessFilter(session: ExamSession, filter: SessionReadinessFilter) {
		if (filter === 'not_ready') return sessionHasBlockingIssues(session);
		if (filter === 'ready') return !sessionHasBlockingIssues(session);
		if (filter === 'needs_participants') return sessionNeedsParticipants(session);
		if (filter === 'needs_rooms') return sessionNeedsRooms(session);
		if (filter === 'needs_proctors') return sessionNeedsProctors(session);
		return true;
	}

	function readinessFilterLabel(filter: SessionReadinessFilter) {
		const labels: Record<SessionReadinessFilter, string> = {
			all: 'Semua',
			not_ready: 'Belum Siap',
			ready: 'Siap Mulai',
			needs_participants: 'Butuh Peserta',
			needs_rooms: 'Butuh Ruang',
			needs_proctors: 'Butuh Pengawas',
		};
		return labels[filter];
	}

	function buildReadinessFilterOptions(items: ExamSession[]) {
		return readinessFilters.map((filter) => ({
			filter,
			label: readinessFilterLabel(filter),
			count: items.filter((session) => sessionMatchesReadinessFilter(session, filter)).length,
		}));
	}

	function buildReadinessBoardCards(items: ExamSession[]) {
		return readinessBoardConfigs.map((card) => ({
			...card,
			count: items.filter((session) => sessionMatchesReadinessFilter(session, card.filter)).length,
		}));
	}

	function witaDateKey(value: string | Date) {
		const date = value instanceof Date ? value : new Date(value);
		if (Number.isNaN(date.getTime())) return '';
		return witaDateFormatter.format(date);
	}

	function sessionIsClosed(session: ExamSession) {
		return session.status === 'finished' || session.status === 'cancelled';
	}

	function sessionScheduleState(session: ExamSession): SessionScheduleState {
		const now = Date.now();
		const start = new Date(session.scheduled_start);
		const end = new Date(session.scheduled_end);
		const startTime = start.getTime();
		const endTime = end.getTime();
		if (Number.isNaN(startTime)) return 'unknown';
		if (!sessionIsClosed(session) && !Number.isNaN(endTime) && endTime < now) return 'overdue';
		if (witaDateKey(start) === witaDateKey(new Date())) return 'today';
		if (startTime > now) return 'upcoming';
		if (sessionIsClosed(session)) return 'closed';
		if (Number.isNaN(endTime) || endTime >= now) return 'running_window';
		return 'unknown';
	}

	function sessionMatchesScheduleFilter(session: ExamSession, filter: SessionScheduleFilter) {
		if (filter === 'today') return sessionScheduleState(session) === 'today';
		if (filter === 'upcoming') return sessionScheduleState(session) === 'upcoming';
		if (filter === 'overdue') return sessionScheduleState(session) === 'overdue';
		return true;
	}

	function scheduleStateLabel(state: SessionScheduleState) {
		const labels: Record<SessionScheduleState, string> = {
			today: 'Hari ini',
			upcoming: 'Akan datang',
			overdue: 'Lewat jadwal',
			running_window: 'Dalam rentang',
			closed: 'Riwayat',
			unknown: 'Jadwal belum pasti',
		};
		return labels[state];
	}

	function scheduleStateClass(state: SessionScheduleState) {
		if (state === 'overdue') return 'border-destructive/30 bg-destructive/10 text-destructive';
		if (state === 'today' || state === 'running_window') return 'border-primary/20 bg-primary/10 text-primary';
		if (state === 'upcoming') return 'border-accent bg-accent/60 text-accent-foreground';
		return 'border-border bg-muted/50 text-muted-foreground';
	}

	function scheduleQuickAction(session: ExamSession, state: SessionScheduleState): ScheduleQuickAction | null {
		if (state !== 'overdue') return null;
		if (session.status === 'active') return { label: 'Selesaikan', kind: 'status', status: 'finished', tone: 'primary' };
		if (session.status === 'scheduled' || session.status === 'draft') return { label: 'Ubah Jadwal', kind: 'schedule', tone: 'warning' };
		return { label: 'Buka Detail', kind: 'link', href: resolve(`/asesmen/sesi/${session.id}`), tone: 'warning' };
	}

	function buildScheduleBoardCards(items: ExamSession[]) {
		return scheduleBoardConfigs.map((card) => ({
			...card,
			count: items.filter((session) => sessionMatchesScheduleFilter(session, card.filter)).length,
		}));
	}

	function readinessBoardCardClass(tone: ReadinessBoardTone, active: boolean) {
		const base = 'rounded-lg border p-3 text-left shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/30';
		if (active) return `${base} border-primary/20 bg-primary/10 text-primary`;
		if (tone === 'danger') return `${base} border-destructive/30 bg-card hover:border-destructive/30 hover:bg-destructive/10`;
		if (tone === 'success') return `${base} border-primary/20 bg-card hover:border-primary/20 hover:bg-primary/10`;
		if (tone === 'warning') return `${base} border-warning/30 bg-card hover:border-warning/30 hover:bg-warning/10`;
		return `${base} border-accent bg-card hover:border-accent hover:bg-accent/60`;
	}

	function scheduleBoardCardClass(tone: ScheduleBoardTone, active: boolean) {
		const base = 'rounded-lg border px-3 py-2 text-left transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/30';
		if (active) return `${base} border-primary/20 bg-primary/10 text-primary`;
		if (tone === 'danger') return `${base} border-destructive/30 bg-card hover:border-destructive/30 hover:bg-destructive/10`;
		if (tone === 'success') return `${base} border-primary/20 bg-card hover:border-primary/20 hover:bg-primary/10`;
		return `${base} border-accent bg-card hover:border-accent hover:bg-accent/60`;
	}

	function isSessionReadinessFilter(value: string | null): value is SessionReadinessFilter {
		return readinessFilters.includes(value as SessionReadinessFilter);
	}

	function isSessionScheduleFilter(value: string | null): value is SessionScheduleFilter {
		return scheduleFilters.includes(value as SessionScheduleFilter);
	}

	function replaceCurrentUrl(url: URL) {
		const query = url.searchParams.toString();
		window.history.replaceState({}, '', `${url.pathname}${query ? `?${query}` : ''}${url.hash}`);
	}

	function setSessionReadinessFilter(filter: SessionReadinessFilter) {
		readinessFilter = filter;
		if (typeof window === 'undefined') return;
		const url = new URL(window.location.href);
		if (filter === 'all') url.searchParams.delete('readiness');
		else url.searchParams.set('readiness', filter);
		replaceCurrentUrl(url);
	}

	function setSessionScheduleFilter(filter: SessionScheduleFilter) {
		scheduleFilter = filter;
		if (typeof window === 'undefined') return;
		const url = new URL(window.location.href);
		if (filter === 'all') url.searchParams.delete('schedule');
		else url.searchParams.set('schedule', filter);
		replaceCurrentUrl(url);
	}

	function initSessionReadinessFilterFromQuery() {
		if (typeof window === 'undefined') return;
		const url = new URL(window.location.href);
		const requestedFilter = url.searchParams.get('readiness');
		if (isSessionReadinessFilter(requestedFilter)) {
			readinessFilter = requestedFilter;
			return;
		}
		if (requestedFilter !== null) {
			url.searchParams.delete('readiness');
			replaceCurrentUrl(url);
		}
	}

	function initSessionScheduleFilterFromQuery() {
		if (typeof window === 'undefined') return;
		const url = new URL(window.location.href);
		const requestedFilter = url.searchParams.get('schedule');
		if (isSessionScheduleFilter(requestedFilter)) {
			scheduleFilter = requestedFilter;
			return;
		}
		if (requestedFilter !== null) {
			url.searchParams.delete('schedule');
			replaceCurrentUrl(url);
		}
	}

	function nextSessionAction(session: ExamSession): NextSessionAction {
		const packageIssues = packageQualityIssues(session.package_id);
		if (packageIssues.length > 0) {
			return { label: 'Rapikan Paket', href: `${resolve('/asesmen/paket')}${eventId ? `?event_id=${eventId}` : ''}`, kind: 'link', tone: 'danger' };
		}
		if (session.participant_count === 0) {
			return { label: 'Daftarkan Peserta', kind: 'enroll', tone: 'warning' };
		}
		if (session.room_count === 0 || session.total_capacity < session.participant_count) {
			return { label: 'Atur Ruang', href: resolve(`/asesmen/sesi/${session.id}?tab=ruangan`), kind: 'link', tone: 'warning' };
		}
		if (session.unassigned_participant_count > 0) {
			return { label: 'Acak Ruang', href: resolve(`/asesmen/sesi/${session.id}?tab=ruangan`), kind: 'link', tone: 'warning' };
		}
		if (session.missing_seat_count > 0) {
			return { label: 'Atur Nomor Meja', href: resolve(`/asesmen/sesi/${session.id}?tab=ruangan`), kind: 'link', tone: 'warning' };
		}
		if (session.rooms_without_proctor > 0) {
			return { label: 'Tetapkan Pengawas', href: resolve(`/asesmen/sesi/${session.id}?tab=ruangan`), kind: 'link', tone: 'warning' };
		}
		return { label: session.status === 'scheduled' ? 'Siap Mulai' : 'Lihat Detail', href: resolve(`/asesmen/sesi/${session.id}`), kind: 'link', tone: 'primary' };
	}

	function nextActionClass(tone: NextSessionAction['tone']) {
		if (tone === 'danger') return 'border-destructive/30 bg-destructive/10 text-destructive hover:bg-destructive/15';
		if (tone === 'warning') return 'border-warning/30 bg-warning/10 text-warning hover:bg-warning/15';
		return 'border-primary/20 bg-primary/10 text-primary hover:bg-primary/15';
	}

	function buildSessionReadinessIssues() {
		const issues: string[] = [];
		if (!fPackageId) issues.push('Pilih paket soal');
		for (const issue of packageQualityIssues(fPackageId)) {
			if (issue === 'paket nonaktif') issues.push('Paket soal tidak aktif');
			else if (issue === 'paket kosong') issues.push('Paket belum memiliki soal');
			else issues.push(`${issue} di paket soal`);
		}
		if (!fTitle.trim()) issues.push('Isi nama sesi');
		if (!fStart) issues.push('Isi jadwal mulai');
		if (!fEnd) issues.push('Isi jadwal selesai');
		if (fStart && fEnd && new Date(fEnd) <= new Date(fStart)) issues.push('Jadwal selesai harus setelah mulai');
		if (fScopeType === 'class' && !fClassId) issues.push('Pilih kelas peserta');
		if (fScopeType === 'grade' && !fGradeLevel) issues.push('Pilih tingkat peserta');
		if (fAllowCrossGrade && !fIsSpecialEvent) issues.push('Lintas tingkat hanya boleh untuk sesi khusus');
		return issues;
	}

	async function fetchEventContext() {
		if (!eventId) return null;
		try {
			return await fetch(clientApiPath`/api/asesmen/events/${eventId}`).then((response) => readClientApiData<EventContext>(response, 'Gagal memuat konteks kegiatan'));
		} catch {
			return null;
		}
	}

	async function fetchOverview(): Promise<SessionsOverview> {
		const sessionParams = new URLSearchParams();
		const packageParams = new URLSearchParams();
		if (eventId) {
			sessionParams.set('event_id', eventId);
			packageParams.set('event_id', eventId);
		}
		const [nextSessions, packagePayload, academicPayload] = await Promise.all([
			fetch(clientApiPathWithQuery('/api/asesmen/sessions', sessionParams)).then((response) => readClientApiData<ExamSession[]>(response, 'Gagal memuat data sesi')),
			fetch(clientApiPathWithQuery('/api/asesmen/packages', packageParams)).then((response) => readClientApiData<CbtPackagesPayload>(response, 'Gagal memuat paket ujian')),
			fetch('/api/academic').then((response) => readClientApiData<AcademicPayload>(response, 'Gagal memuat data akademik')),
		]);
		return {
			sessions: Array.isArray(nextSessions) ? nextSessions : [],
			packages: packagePayload.packages ?? [],
			packageQuestions: parsePackageQuestions(packagePayload),
			classes: academicPayload.classes ?? [],
		};
	}

	function applyOverview(overview: SessionsOverview) {
		hiddenEventSessionCount = hiddenSessionCount(overview.sessions);
		hiddenEventPackageCount = hiddenPackageCount(overview.packages);
		sessions = strictEventSessions(overview.sessions);
		packages = strictEventPackages(overview.packages);
		packageQuestions = overview.packageQuestions;
		classes = overview.classes;
	}

	function loadInitial() {
		const requestId = ++sessionsRequestId;
		sessions = [];
		packages = [];
		packageQuestions = [];
		classes = [];
		hiddenEventSessionCount = 0;
		hiddenEventPackageCount = 0;
		void fetchEventContext().then((context) => { eventContext = context; });
		sessionsPromise = fetchOverview().then((overview) => {
			if (requestId !== sessionsRequestId) return { sessions, packages, packageQuestions, classes };
			applyOverview(overview);
			return overview;
		}).catch((error: unknown) => {
			if (requestId === sessionsRequestId) throw error;
			return { sessions, packages, packageQuestions, classes };
		});
	}

	async function refreshSessions() {
		if (!sessionsPromise) {
			loadInitial();
			return;
		}
		const requestId = ++sessionsRequestId;
		try {
			const overview = await fetchOverview();
			if (requestId !== sessionsRequestId) return;
			applyOverview(overview);
			sessionsPromise = Promise.resolve(overview);
		} catch (error) {
			if (requestId === sessionsRequestId) {
				sessionsPromise = Promise.resolve({ sessions, packages, packageQuestions, classes });
				toast.error(sessionsErrorMessage(error));
			}
		}
	}

	function retrySessions(reset?: () => void) {
		reset?.();
		loadInitial();
	}

	function sessionsErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat data sesi';
	}

	function handleSessionsRenderError(error: unknown) {
		console.error('Daftar sesi ujian belum dapat ditampilkan', error);
	}

	function showToast(msg: string, ok = true) {
		if (ok) toast.success(msg);
		else toast.error(msg);
	}

	function mutationErrorMessage(error: unknown, fallback: string) {
		if (error instanceof Error && error.message.trim() && !error.message.toLowerCase().includes('fetch')) return error.message;
		return fallback;
	}

	function setOperationState(
		tone: 'success' | 'error' | 'warning' | 'info',
		title: string,
		message: string,
	) {
		operationState = { tone, title, message };
	}

	function confirmPhrase(title: string, detail: string, challenge: string) {
		return confirmChallenge({
			title,
			message: detail,
			challenge,
			confirmLabel: 'Konfirmasi',
			tone: 'danger'
		});
	}

	function sessionLegacyMutationPath(id: string) {
		return clientApiPathWithQuery('/api/asesmen/sessions', new URLSearchParams({ id }));
	}

	async function createSession() {
		if (sessionReadinessIssues.length > 0) {
			setOperationState('warning', 'Sesi Belum Siap', sessionReadinessIssues[0] ?? 'Lengkapi sesi ujian terlebih dahulu.');
			return;
		}
		fBusy = true;
		try {
			const scopeRef = fScopeType === 'class' ? fClassId : fScopeType === 'grade' ? fGradeLevel : '';
			const res = await fetch('/api/asesmen/sessions', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					...(eventId ? { event_id: eventId } : {}),
					package_id: fPackageId,
					class_id: fScopeType === 'class' ? fClassId : '',
					scope_type: fScopeType,
					scope_ref: scopeRef,
					mix_policy: fMixPolicy,
					assignment_mode: fAssignmentMode,
					allow_cross_grade: fAllowCrossGrade,
					is_special_event: fIsSpecialEvent,
					title: fTitle,
					scheduled_start: toRFC3339(fStart), scheduled_end: toRFC3339(fEnd),
					status: 'draft',
				}),
			});
			await readClientJson<unknown>(res);
			fPackageId = ''; fScopeType = 'class'; fClassId = ''; fGradeLevel = 'VII';
			fMixPolicy = 'same_class'; fAssignmentMode = 'random_balanced';
			fAllowCrossGrade = false; fIsSpecialEvent = false;
			fTitle = ''; fStart = ''; fEnd = '';
			showForm = false;
			setOperationState('success', 'Sesi Tersimpan Sebagai Draft', 'Sesi baru sudah dibuat. Daftarkan peserta dan cek ruang sebelum menjadwalkan atau memulai sesi.');
			showToast('Sesi ujian berhasil dibuat');
			await refreshSessions();
		} catch (error) {
			showToast(mutationErrorMessage(error, 'Gagal membuat sesi ujian. Periksa koneksi lalu coba lagi.'), false);
		} finally { fBusy = false; }
	}

	async function deleteSession(id: string, title: string) {
		if (!(await confirmPhrase('Hapus Sesi Ujian', `Sesi "${title}" hanya boleh dihapus jika masih Draft. Penghapusan akan membuang konfigurasi sesi dari daftar operator.`, 'HAPUS'))) return;
		deleteBusyId = id;
		try {
			const res = await fetch(sessionLegacyMutationPath(id), { method: 'DELETE' });
			await readClientJson<unknown>(res);
			setOperationState('warning', 'Sesi Dihapus', `Sesi "${title}" sudah dihapus dari daftar sesi ujian.`);
			showToast('Sesi dihapus');
			await refreshSessions();
		} catch (error) {
			setOperationState('error', 'Sesi Gagal Dihapus', 'Hanya sesi berstatus Draft yang dapat dihapus. Ubah alur kerja sesi atau periksa statusnya terlebih dahulu.');
			showToast(mutationErrorMessage(error, 'Gagal menghapus sesi ujian. Periksa koneksi lalu coba lagi.'), false);
		} finally {
			deleteBusyId = '';
		}
	}

	async function updateStatus(id: string, status: string) {
		const challenge = status === 'cancelled' ? 'BATALKAN' : status === 'finished' ? 'SELESAI' : '';
		if (challenge && !(await confirmPhrase('Konfirmasi Perubahan Status', `Perubahan ini akan mengubah status sesi menjadi "${statusLabel[status] ?? status}" dan memengaruhi operasi ujian berikutnya.`, challenge))) {
			return;
		}
		statusBusyId = id;
		try {
			const res = await fetch(clientApiPath`/api/asesmen/sessions/${id}/status`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ status }),
			});
			await readClientJson<unknown>(res);
			setOperationState('success', 'Status Sesi Diperbarui', `Sesi sekarang berstatus "${statusLabel[status] ?? status}". Pastikan langkah operator berikutnya sudah sesuai.`);
			showToast('Status diperbarui');
			await refreshSessions();
		} catch (error) {
			const message = mutationErrorMessage(error, 'Gagal mengubah status sesi. Periksa kesiapan ruang, peserta, kursi, dan pengawas lalu coba lagi.');
			setOperationState('error', 'Status Gagal Diperbarui', message);
			showToast(message, false);
		} finally {
			statusBusyId = '';
		}
	}

	async function enrollParticipants() {
		if (!enrollSession) return;
		if (enrollScopeType === 'class' && !enrollClassId) return;
		if (enrollScopeType === 'grade' && !enrollGradeLevel) return;
		enrollBusy = true;
		try {
			const payload =
				enrollScopeType === 'class'
					? { scope_type: 'class', class_id: enrollClassId }
					: enrollScopeType === 'grade'
						? { scope_type: 'grade', level: enrollGradeLevel }
						: { scope_type: 'school' };
			const res = await fetch(clientApiPath`/api/asesmen/sessions/${enrollSession.id}/enroll`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(payload),
			});
			await readClientJson<unknown>(res);
			setOperationState('success', 'Peserta Berhasil Didaftarkan', `Kelompok peserta untuk sesi "${enrollSession.title}" sudah masuk. Lanjutkan ke pengaturan ruangan jika diperlukan.`);
			showToast(`Peserta berhasil didaftarkan ke sesi "${enrollSession.title}"`);
			enrollSession = null;
			enrollScopeType = 'class';
			enrollClassId = '';
			enrollGradeLevel = 'VII';
			await refreshSessions();
		} catch (error) {
			showToast(mutationErrorMessage(error, 'Gagal mendaftarkan siswa. Periksa koneksi lalu coba lagi.'), false);
		} finally { enrollBusy = false; }
	}

	function openScheduleEditor(session: ExamSession) {
		scheduleSession = session;
		scheduleStart = toLocalDateTimeInput(session.scheduled_start);
		scheduleEnd = toLocalDateTimeInput(session.scheduled_end);
	}

	function closeScheduleEditor() {
		scheduleSession = null;
		scheduleStart = '';
		scheduleEnd = '';
	}

	async function updateSchedule() {
		if (!scheduleSession) return;
		if (!scheduleStart || !scheduleEnd) {
			setOperationState('warning', 'Jadwal Belum Lengkap', 'Isi jadwal mulai dan selesai sebelum menyimpan perubahan jadwal.');
			return;
		}
		if (new Date(scheduleEnd) <= new Date(scheduleStart)) {
			setOperationState('warning', 'Jadwal Tidak Valid', 'Jadwal selesai harus setelah jadwal mulai.');
			return;
		}
		scheduleBusy = true;
		try {
			const res = await fetch(clientApiPath`/api/asesmen/sessions/${scheduleSession.id}/schedule`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					scheduled_start: toRFC3339(scheduleStart),
					scheduled_end: toRFC3339(scheduleEnd),
				}),
			});
			await readClientJson<unknown>(res);
			setOperationState('success', 'Jadwal Sesi Diperbarui', `Jadwal "${scheduleSession.title}" sudah digeser. Cek kembali kesiapan ruang dan pengawas sebelum sesi dimulai.`);
			showToast('Jadwal sesi diperbarui');
			closeScheduleEditor();
			await refreshSessions();
		} catch (error) {
			const message = mutationErrorMessage(error, 'Gagal mengubah jadwal sesi. Periksa jadwal lalu coba lagi.');
			setOperationState('error', 'Jadwal Gagal Diperbarui', message);
			showToast(message, false);
		} finally {
			scheduleBusy = false;
		}
	}

	onMount(() => {
		detectBrowserTimeZone();
		initSessionReadinessFilterFromQuery();
		initSessionScheduleFilterFromQuery();
		void loadInitial();
	});
</script>

	<svelte:head><title>{eventId ? 'Kegiatan & Sesi Ujian' : 'Sesi Ujian'} — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<p class="text-xs font-semibold uppercase tracking-[0.16em] text-success">Kegiatan & Sesi</p>
			<h1 class="text-2xl font-semibold text-foreground">Sesi Ujian</h1>
			<p class="text-sm text-muted-foreground mt-1">Daftar sesi ujian sebagai bagian dari alur Kegiatan & Sesi{eventId ? ' untuk kegiatan ini' : ''}.</p>
		</div>
		<div class="flex flex-wrap gap-2">
			{#if eventId}
				<a href={resolve(`/asesmen/kegiatan/${eventId}`)} class="inline-flex items-center rounded-md border border-success/20 bg-success/10 px-3 py-2 text-sm font-semibold text-success hover:bg-success/15">Kembali ke Kegiatan</a>
			{/if}
			<Button href={createSessionHref}>+ Sesi</Button>
		</div>
	</div>

	{#if eventId}
		<div class="rounded-xl border border-success/20 bg-success/10 p-4 text-sm text-success">
			<div class="flex flex-wrap items-start justify-between gap-3">
				<div>
					<p class="font-semibold">Sesi untuk kegiatan: {eventContext?.title ?? eventId}</p>
					<p class="mt-1 text-success">Daftar sesi dan payload pembuatan sesi membawa <code class="rounded bg-card px-1">event_id</code>. Item global atau event lain disembunyikan agar tidak terbaca sebagai sesi kegiatan ini.</p>
				</div>
				<a href={resolve(`/asesmen/paket?event_id=${eventId}`)} class="rounded-md border border-success/20 bg-card px-3 py-2 text-sm font-semibold text-success hover:bg-success/15">Paket Kegiatan</a>
			</div>
		</div>
	{:else}
		<div class="rounded-xl border border-warning/30 bg-warning/10 px-4 py-3 text-sm text-warning">
			Anda sedang melihat sesi global. Dari Kegiatan & Sesi, gunakan tombol Sesi agar pembuatan sesi otomatis terhubung ke kegiatan.
		</div>
	{/if}

	{#if hiddenEventSessionCount > 0 || hiddenEventPackageCount > 0}
		<div class="rounded-xl border border-accent bg-accent/60 px-4 py-3 text-sm text-accent-foreground">
			<p class="font-semibold">Item di luar event disembunyikan dari layar ini.</p>
			<p class="mt-1">
				{hiddenEventSessionCount} sesi dan {hiddenEventPackageCount} paket global/event lain tidak ditampilkan karena <code class="rounded bg-card px-1">event_id</code> tidak sama dengan kegiatan aktif.
			</p>
		</div>
	{/if}

	{#if operationState}
		<OperationStatusPanel {...operationState} />
	{/if}

	{#if showForm}
		<Card.Root>
			<Card.Header class="pb-2">
				<Card.Title class="text-base">Buat Sesi Ujian Baru</Card.Title>
			</Card.Header>
			<Card.Content class="space-y-4">
				<div class="grid gap-3 sm:grid-cols-2">
					<div>
						<label for="session-package" class="text-xs text-muted-foreground mb-1 block">Paket Soal <span class="text-destructive">*</span></label>
						<select id="session-package" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fPackageId}>
							<option value="">-- Pilih Paket --</option>
							{#each packages as p (p.id)}
								<option value={p.id}>{p.title} ({p.subject_code}){p.is_active ? '' : ' - nonaktif'}</option>
							{/each}
						</select>
						{#if hiddenEventPackageCount > 0}
							<p class="mt-1 text-[11px] text-accent-foreground">{hiddenEventPackageCount} paket global/event lain disembunyikan dari pilihan sesi event ini.</p>
						{/if}
					</div>
					{#if fPackageId}
						{@const quality = selectedPackageQuality}
						<div class="sm:col-span-2 rounded-lg border border-border bg-muted/50 px-3 py-2">
							<div class="flex flex-wrap items-start justify-between gap-3">
								<div>
									<p class="text-xs font-semibold uppercase tracking-[0.16em] text-foreground">Quality Gate Paket</p>
									<p class="mt-1 text-sm font-medium text-foreground">{selectedPackage?.title ?? 'Paket dipilih'}</p>
								</div>
								<div class="flex flex-wrap gap-1.5">
									<Badge variant="outline" class="bg-card text-xs">{quality.totalCount} soal</Badge>
									{#each quality.typeBuckets.slice(0, 3) as bucket (bucket.label)}
										<Badge variant="outline" class="bg-card text-xs">{bucket.label}: {bucket.count}</Badge>
									{/each}
									{#if quality.hotsCount > 0}
										<Badge class="border-warning/30 bg-warning/10 text-warning text-xs">{quality.hotsCount} HOTS</Badge>
									{/if}
									{#if quality.missingCount > 0}
										<Badge class="border-warning/30 bg-warning/10 text-warning text-xs">{quality.missingCount} metadata kurang</Badge>
									{/if}
									{#if quality.unpublishedCount > 0}
										<Badge class="border-destructive/30 bg-destructive/10 text-destructive text-xs">{quality.unpublishedCount} belum terbit</Badge>
									{/if}
								</div>
							</div>
							{#if selectedPackage && !selectedPackage.is_active}
								<p class="mt-2 text-xs font-medium text-destructive">Paket nonaktif tidak boleh dijadikan sesi ujian.</p>
							{:else if quality.totalCount === 0}
								<p class="mt-2 text-xs font-medium text-destructive">Paket ini belum memiliki soal, sehingga sesi tidak bisa dibuat.</p>
							{:else if quality.unpublishedCount > 0}
								<p class="mt-2 text-xs font-medium text-destructive">Rapikan paket dulu. Flutter hanya menyajikan soal terbit, jadi soal belum terbit akan membuat jumlah soal sesi tidak konsisten.</p>
							{:else if quality.missingCount > 0}
								<p class="mt-2 text-xs font-medium text-warning">Sesi masih boleh dibuat, tetapi {quality.missingCount} soal belum lengkap CP/TP/KD atau level kognitif.</p>
							{:else}
								<p class="mt-2 text-xs font-medium text-primary">Paket siap dipakai untuk draft sesi CBT.</p>
							{/if}
						</div>
					{/if}
					<div>
						<label for="session-scope" class="text-xs text-muted-foreground mb-1 block">Cakupan peserta <span class="text-destructive">*</span></label>
						<select
							id="session-scope"
							class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
							value={fScopeType}
							onchange={(event) => updateScopeType((event.currentTarget as HTMLSelectElement).value)}
						>
							<option value="class">Per kelas</option>
							<option value="grade">Per tingkat</option>
							<option value="school">Seluruh sekolah</option>
						</select>
					</div>
					{#if fScopeType === 'class'}
						<div>
							<label for="session-class" class="text-xs text-muted-foreground mb-1 block">Kelas <span class="text-destructive">*</span></label>
							<select id="session-class" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fClassId}>
								<option value="">-- Pilih Kelas --</option>
								{#each classes as c (c.id)}
									<option value={c.id}>{c.code} — {c.name}</option>
								{/each}
							</select>
						</div>
					{:else if fScopeType === 'grade'}
						<div>
							<label for="session-grade" class="text-xs text-muted-foreground mb-1 block">Tingkat <span class="text-destructive">*</span></label>
							<select id="session-grade" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fGradeLevel}>
								<option value="VII">VII</option>
								<option value="VIII">VIII</option>
								<option value="IX">IX</option>
							</select>
						</div>
					{:else}
						<div class="rounded-md border border-primary/20 bg-primary/10 px-3 py-2 text-sm text-primary">
							Semua siswa aktif di sekolah dapat menjadi peserta sesi ini.
						</div>
					{/if}
					<div>
						<label for="session-mix-policy" class="text-xs text-muted-foreground mb-1 block">Mix policy</label>
						<select id="session-mix-policy" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fMixPolicy}>
							<option value="same_class">Tetap per kelas</option>
							<option value="same_grade">Campur dalam tingkat</option>
							<option value="mixed_scope">Campur lintas cakupan</option>
						</select>
						<p class="mt-1 text-[11px] text-muted-foreground">Nilai ini dikirim ke backend. Saat cakupan berubah, opsi disetel otomatis lalu tetap bisa disesuaikan operator.</p>
					</div>
					<div>
						<label for="session-assignment-mode" class="text-xs text-muted-foreground mb-1 block">Mode alokasi ruangan</label>
						<select id="session-assignment-mode" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fAssignmentMode}>
							<option value="random_balanced">Acak seimbang</option>
							<option value="manual">Manual</option>
							<option value="random_by_gender">Acak per gender</option>
							<option value="random_by_accommodation">Acak akomodasi khusus</option>
						</select>
					</div>
					<div class="sm:col-span-2">
						<label for="session-title" class="text-xs text-muted-foreground mb-1 block">Nama Sesi <span class="text-destructive">*</span></label>
						<Input id="session-title" placeholder="mis: UTS Matematika VII A - Semester 1 2025" bind:value={fTitle} />
					</div>
					<div>
						<label for="session-start" class="text-xs text-muted-foreground mb-1 block">Mulai <span class="text-destructive">*</span></label>
						<Input id="session-start" type="datetime-local" bind:value={fStart} />
					</div>
					<div>
						<label for="session-end" class="text-xs text-muted-foreground mb-1 block">Selesai <span class="text-destructive">*</span></label>
						<Input id="session-end" type="datetime-local" bind:value={fEnd} />
					</div>
					<div class="sm:col-span-2 rounded-md border border-primary/20 bg-primary/10 px-3 py-2 text-xs leading-5 text-primary">
						<p class="font-semibold">Jadwal sesi dicatat dan ditampilkan sebagai WITA (Asia/Makassar).</p>
						{#if browserTimeZoneMismatch}
							<p class="text-warning">Zona waktu browser terdeteksi {browserTimeZone}. Samakan perangkat operator ke Asia/Makassar sebelum menyimpan agar input <code class="rounded bg-card px-1">datetime-local</code> tidak bergeser.</p>
						{:else}
							<p>Pastikan jam mulai dan selesai mengikuti waktu sekolah/WITA sebelum sesi dijadwalkan.</p>
						{/if}
					</div>
					<div class="sm:col-span-2 grid gap-3 sm:grid-cols-2">
						<label class="flex items-center gap-2 rounded-md border border-input px-3 py-2 text-sm text-foreground">
							<input type="checkbox" bind:checked={fIsSpecialEvent} class="size-4 accent-primary" />
							Tandai sebagai sesi khusus
						</label>
						<label class="flex items-center gap-2 rounded-md border border-input px-3 py-2 text-sm text-foreground">
							<input type="checkbox" bind:checked={fAllowCrossGrade} class="size-4 accent-primary" />
							Izinkan lintas tingkat
						</label>
					</div>
				</div>
				<div class="flex gap-2">
					<LoadingButton
						disabled={!canCreateSession}
						onclick={() => void createSession()}
						loading={fBusy}
						loadingLabel="Menyimpan..."
					>
						Buat Sesi
					</LoadingButton>
					<Button variant="outline" onclick={() => (showForm = false)}>Batal</Button>
				</div>
				{#if sessionReadinessIssues.length > 0}
					<div class="rounded-md border border-warning/30 bg-warning/10 px-3 py-2 text-xs text-warning">
						<span class="font-semibold">Belum siap dibuat:</span>
						<span>{sessionReadinessIssues.join(', ')}</span>
					</div>
				{/if}
			</Card.Content>
		</Card.Root>
	{/if}

	<!-- Enroll modal -->
	{#if enrollSession}
		<Card.Root class="border-success/20 bg-success/10">
			<Card.Header class="pb-2">
				<Card.Title class="text-base text-success">Daftarkan Siswa ke Sesi</Card.Title>
				<p class="text-sm text-success mt-0.5">{enrollSession.title}</p>
			</Card.Header>
			<Card.Content class="space-y-3">
				<p class="text-sm text-muted-foreground">Tentukan kelompok peserta untuk sesi ini. Ruangan tetap bisa diacak terpisah setelah peserta terdaftar.</p>
				<div>
					<label for="enroll-scope" class="text-xs text-muted-foreground mb-1 block">Cakupan peserta</label>
					<select id="enroll-scope" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={enrollScopeType}>
						<option value="class">Per kelas</option>
						<option value="grade">Per tingkat</option>
						<option value="school">Seluruh sekolah</option>
					</select>
				</div>
				{#if enrollScopeType === 'class'}
					<div>
						<label for="enroll-class" class="text-xs text-muted-foreground mb-1 block">Pilih Kelas</label>
						<select id="enroll-class" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={enrollClassId}>
							<option value="">-- Pilih Kelas --</option>
							{#each classes as c (c.id)}
								<option value={c.id}>{c.code} — {c.name}</option>
							{/each}
						</select>
					</div>
				{:else if enrollScopeType === 'grade'}
					<div>
						<label for="enroll-grade" class="text-xs text-muted-foreground mb-1 block">Pilih Tingkat</label>
						<select id="enroll-grade" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={enrollGradeLevel}>
							<option value="VII">VII</option>
							<option value="VIII">VIII</option>
							<option value="IX">IX</option>
						</select>
					</div>
				{:else}
					<div class="rounded-md border border-primary/20 bg-card px-3 py-2 text-sm text-foreground">
						Semua siswa aktif di sekolah akan didaftarkan ke sesi ini.
					</div>
				{/if}
				<div class="flex gap-2">
					<LoadingButton
						disabled={enrollBusy || (enrollScopeType === 'class' && !enrollClassId) || (enrollScopeType === 'grade' && !enrollGradeLevel)}
						onclick={() => void enrollParticipants()}
						loading={enrollBusy}
						loadingLabel="Mendaftarkan..."
					>
						Daftarkan Siswa
					</LoadingButton>
					<Button variant="outline" onclick={() => { enrollSession = null; enrollScopeType = 'class'; enrollClassId = ''; enrollGradeLevel = 'VII'; }}>Batal</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	{#if scheduleSession}
		<Card.Root class="border-warning/30 bg-warning/10">
			<Card.Header class="pb-2">
				<Card.Title class="text-base text-warning">Ubah Jadwal Sesi</Card.Title>
				<p class="text-sm text-warning mt-0.5">{scheduleSession.title}</p>
			</Card.Header>
			<Card.Content class="space-y-3">
				<p class="text-sm text-muted-foreground">Geser jadwal untuk sesi draft atau terjadwal. Setelah disimpan, cek lagi kesiapan ruang dan pengawas.</p>
				<div class="rounded-md border border-warning/30 bg-card px-3 py-2 text-xs leading-5 text-warning">
					<p class="font-semibold">Jadwal sesi menggunakan WITA (Asia/Makassar).</p>
					{#if browserTimeZoneMismatch}
						<p>Browser operator saat ini terdeteksi {browserTimeZone}. Koreksi zona waktu perangkat ke Asia/Makassar sebelum menyimpan perubahan jadwal.</p>
					{:else}
						<p>Periksa ulang jam mulai dan selesai dengan jam sekolah sebelum menyimpan.</p>
					{/if}
				</div>
				<div class="grid gap-3 sm:grid-cols-2">
					<div>
						<label for="quick-schedule-start" class="text-xs text-muted-foreground mb-1 block">Mulai <span class="text-destructive">*</span></label>
						<Input id="quick-schedule-start" type="datetime-local" bind:value={scheduleStart} />
					</div>
					<div>
						<label for="quick-schedule-end" class="text-xs text-muted-foreground mb-1 block">Selesai <span class="text-destructive">*</span></label>
						<Input id="quick-schedule-end" type="datetime-local" bind:value={scheduleEnd} />
					</div>
				</div>
				<div class="flex flex-wrap gap-2">
					<LoadingButton
						disabled={scheduleBusy || !scheduleStart || !scheduleEnd}
						onclick={() => void updateSchedule()}
						loading={scheduleBusy}
						loadingLabel="Menyimpan..."
					>
						Simpan Jadwal
					</LoadingButton>
					<Button variant="outline" disabled={scheduleBusy} onclick={closeScheduleEditor}>Batal</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	<AsyncContent promise={sessionsPromise} onerror={handleSessionsRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<div class="grid gap-2 sm:grid-cols-2 xl:grid-cols-5">
					{#each Array.from({ length: 5 }) as _, index (`cbt-session-stat-skeleton-${index}`)}
						<Card.Root class="border-border">
							<Card.Content class="space-y-2 p-4">
								<Skeleton class="h-4 w-24" />
								<Skeleton class="h-7 w-16" />
							</Card.Content>
						</Card.Root>
					{/each}
				</div>
				<Card.Root class="overflow-hidden border-border shadow-sm">
					<Card.Content class="space-y-3 p-6">
						{#each Array.from({ length: 5 }) as _, index (`cbt-session-row-skeleton-${index}`)}
							<div class="grid gap-3 lg:grid-cols-[1.1fr_1fr_0.85fr_0.9fr_0.45fr_1fr_0.65fr_auto] lg:items-center">
								<Skeleton class="h-5 w-40" />
								<Skeleton class="h-5 w-32" />
								<Skeleton class="h-5 w-28" />
								<Skeleton class="h-5 w-36" />
								<Skeleton class="h-5 w-12" />
								<Skeleton class="h-8 w-44" />
								<Skeleton class="h-6 w-20" />
								<Skeleton class="h-9 w-36 justify-self-end" />
							</div>
						{/each}
					</Card.Content>
				</Card.Root>
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Sesi Ujian Belum Tersaji"
				message={sessionsErrorMessage(error)}
				onRetry={() => retrySessions(reset)}
			/>
		{/snippet}

		{#snippet children(value)}
			{@const overview = value as SessionsOverview}
			{@const eventSessions = strictEventSessions(overview.sessions)}
			{@const currentSessions = eventSessions}
			{@const hiddenSessions = hiddenSessionCount(overview.sessions)}
			{@const scheduleScopedSessions = currentSessions.filter((session) => sessionMatchesScheduleFilter(session, scheduleFilter))}
			{@const readinessScopedSessions = currentSessions.filter((session) => sessionMatchesReadinessFilter(session, readinessFilter))}
			{@const visibleSessions = currentSessions.filter((session) => sessionMatchesReadinessFilter(session, readinessFilter) && sessionMatchesScheduleFilter(session, scheduleFilter))}
			{@const readinessFilterOptions = buildReadinessFilterOptions(scheduleScopedSessions)}
			{@const readinessBoardCards = buildReadinessBoardCards(scheduleScopedSessions)}
			{@const scheduleBoardCards = buildScheduleBoardCards(readinessScopedSessions)}
		<div class="space-y-4">
			<div class="grid gap-2 sm:grid-cols-2 xl:grid-cols-5">
				{#each readinessBoardCards as card (card.filter)}
					<button
						type="button"
						class={readinessBoardCardClass(card.tone, readinessFilter === card.filter)}
						aria-pressed={readinessFilter === card.filter}
						onclick={() => setSessionReadinessFilter(card.filter)}
					>
						<div class="flex items-start justify-between gap-2">
							<div>
								<p class="text-[11px] font-semibold uppercase tracking-[0.16em] text-muted-foreground">{card.label}</p>
								<p class="mt-1 text-2xl font-semibold text-foreground">{card.count}</p>
							</div>
							{#if readinessFilter === card.filter}
								<span class="rounded bg-primary px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.12em] text-primary-foreground">Aktif</span>
							{/if}
						</div>
						<p class="mt-1 text-xs text-muted-foreground">{card.helper}</p>
					</button>
				{/each}
			</div>
			<div class="rounded-lg border border-border bg-card p-3 shadow-sm">
				<div class="flex flex-wrap items-center justify-between gap-2">
					<div>
						<p class="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Jadwal Sesi</p>
						<p class="mt-0.5 text-xs text-muted-foreground">Urutan waktu pelaksanaan CBT</p>
					</div>
					{#if scheduleFilter !== 'all'}
						<Button variant="outline" size="sm" onclick={() => setSessionScheduleFilter('all')}>
							Reset Jadwal
						</Button>
					{/if}
				</div>
				<div class="mt-3 grid gap-2 sm:grid-cols-3">
					{#each scheduleBoardCards as card (card.filter)}
						<button
							type="button"
							class={scheduleBoardCardClass(card.tone, scheduleFilter === card.filter)}
							aria-pressed={scheduleFilter === card.filter}
							onclick={() => setSessionScheduleFilter(card.filter)}
						>
							<div class="flex items-start justify-between gap-2">
								<div>
									<p class="text-xs font-semibold text-foreground">{card.label}</p>
									<p class="text-xl font-semibold text-foreground">{card.count}</p>
								</div>
								{#if scheduleFilter === card.filter}
									<span class="rounded bg-primary px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.12em] text-primary-foreground">Aktif</span>
								{/if}
							</div>
							<p class="mt-0.5 text-xs text-muted-foreground">{card.helper}</p>
						</button>
					{/each}
				</div>
			</div>
		<Card.Root class="overflow-hidden border-border shadow-sm">
			<Card.Header class="space-y-3 pb-3">
				<div class="flex flex-wrap items-center justify-between gap-2">
					<Card.Title class="text-base">Sesi dalam Kegiatan ({visibleSessions.length}/{currentSessions.length})</Card.Title>
					{#if readinessFilter !== 'all'}
						<Button variant="outline" size="sm" onclick={() => setSessionReadinessFilter('all')}>
							Reset Kesiapan
						</Button>
					{/if}
				</div>
				{#if hiddenSessions > 0}
					<Card.Description>{hiddenSessions} sesi global atau event lain disembunyikan dari daftar event ini.</Card.Description>
				{/if}
				<div class="flex flex-wrap gap-1.5">
					{#each readinessFilterOptions as option (option.filter)}
						<button
							type="button"
							class="inline-flex items-center gap-1 rounded-md border px-2.5 py-1 text-xs font-semibold transition-colors {readinessFilter === option.filter ? 'border-primary/20 bg-primary/10 text-primary' : 'border-border bg-card text-muted-foreground hover:bg-muted/50'}"
							onclick={() => setSessionReadinessFilter(option.filter)}
						>
							<span>{option.label}</span>
							<span class="rounded bg-muted px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground">{option.count}</span>
						</button>
					{/each}
				</div>
			</Card.Header>
			<Card.Content class="p-0">
				<div class="hidden overflow-x-auto lg:block">
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Sesi</Table.Head>
							<Table.Head>Paket</Table.Head>
							<Table.Head>Cakupan</Table.Head>
							<Table.Head>Jadwal Mulai</Table.Head>
							<Table.Head>Peserta</Table.Head>
							<Table.Head>Kesiapan</Table.Head>
							<Table.Head>Status</Table.Head>
							<Table.Head></Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each visibleSessions as s (s.id)}
							{@const rowPackageQuality = packageQualitySummary(s.package_id)}
							{@const rowPackageIssues = packageQualityIssues(s.package_id)}
							{@const rowOperationalIssues = sessionOperationalIssues(s)}
							{@const rowReadinessIssues = sessionRowReadinessIssues(s)}
							{@const rowNextAction = nextSessionAction(s)}
							{@const rowScheduleState = sessionScheduleState(s)}
							{@const rowScheduleQuickAction = scheduleQuickAction(s, rowScheduleState)}
							<Table.Row>
								<Table.Cell class="font-medium max-w-48">
									<p class="truncate">{s.title}</p>
								</Table.Cell>
								<Table.Cell class="max-w-44">
									<p class="truncate text-sm text-muted-foreground">{s.package_title}</p>
									<div class="mt-1 flex flex-wrap gap-1">
										<Badge variant="outline" class="bg-card text-[11px]">{rowPackageQuality.totalCount} soal</Badge>
										{#if rowPackageIssues.length > 0}
											<Badge class="border-destructive/30 bg-destructive/10 text-[11px] text-destructive">{rowPackageIssues.join(', ')}</Badge>
										{:else if rowPackageQuality.missingCount > 0}
											<Badge class="border-warning/30 bg-warning/10 text-[11px] text-warning">{rowPackageQuality.missingCount} metadata kurang</Badge>
										{:else}
											<Badge class="border-primary/20 bg-primary/10 text-[11px] text-primary">Paket siap</Badge>
										{/if}
									</div>
								</Table.Cell>
								<Table.Cell>
									<div class="space-y-1">
										<Badge variant="outline" class="text-xs">{scopeSummary(s)}</Badge>
										<p class="text-[11px] text-muted-foreground">{mixPolicyLabel(s.mix_policy)}</p>
									</div>
								</Table.Cell>
								<Table.Cell class="whitespace-nowrap">
									<div class="space-y-1">
										<p class="text-xs text-muted-foreground">{fmtDt(s.scheduled_start)}</p>
										<Badge class="{scheduleStateClass(rowScheduleState)} text-[11px]">{scheduleStateLabel(rowScheduleState)}</Badge>
										{#if rowScheduleQuickAction?.kind === 'schedule'}
											<button
												type="button"
												class="inline-flex items-center rounded-md border px-2 py-0.5 text-[11px] font-semibold transition-colors {nextActionClass(rowScheduleQuickAction.tone)}"
												onclick={() => openScheduleEditor(s)}
											>
												{rowScheduleQuickAction.label}
											</button>
										{:else if rowScheduleQuickAction?.status === 'cancelled'}
											<LoadingButton size="xs" variant="outline" onclick={() => updateStatus(s.id, 'cancelled')} loading={statusBusyId === s.id} disabled={statusBusyId !== '' && statusBusyId !== s.id} loadingLabel="Memproses...">
												{rowScheduleQuickAction.label}
											</LoadingButton>
										{:else if rowScheduleQuickAction?.status === 'finished'}
											<LoadingButton size="xs" onclick={() => updateStatus(s.id, 'finished')} loading={statusBusyId === s.id} disabled={statusBusyId !== '' && statusBusyId !== s.id} loadingLabel="Memproses...">
												{rowScheduleQuickAction.label}
											</LoadingButton>
										{:else if rowScheduleQuickAction?.href}
											<a class="inline-flex items-center rounded-md border px-2 py-0.5 text-[11px] font-semibold transition-colors {nextActionClass(rowScheduleQuickAction.tone)}" href={resolve(rowScheduleQuickAction.href as '/')}>
												{rowScheduleQuickAction.label}
											</a>
										{/if}
									</div>
								</Table.Cell>
								<Table.Cell>
									<span class="font-mono text-sm">{s.participant_count}</span>
								</Table.Cell>
								<Table.Cell>
									<div class="flex max-w-60 flex-wrap gap-1">
										{#if rowReadinessIssues.length === 0}
											<Badge class="border-primary/20 bg-primary/10 text-[11px] text-primary">Siap mulai</Badge>
										{:else}
											<Badge class="{rowPackageIssues.length > 0 ? 'border-destructive/30 bg-destructive/10 text-destructive' : 'border-warning/30 bg-warning/10 text-warning'} text-[11px]">
												{rowReadinessIssues.length} atensi
											</Badge>
										{/if}
										<Badge variant="outline" class="bg-card text-[11px]">{s.room_count} ruang / {s.total_capacity} kursi</Badge>
										<Badge variant="outline" class="bg-card text-[11px]">{s.assigned_participant_count}/{s.participant_count} ditempatkan</Badge>
										<Badge variant="outline" class="bg-card text-[11px]">{s.proctor_assignment_count} pengawas</Badge>
										{#each rowReadinessIssues.slice(0, 2) as issue (issue)}
											<span class="text-[11px] text-muted-foreground">{issue}</span>
										{/each}
										{#if rowNextAction.kind === 'enroll'}
											<button
												type="button"
												class="inline-flex items-center rounded-md border px-2 py-0.5 text-[11px] font-semibold transition-colors {nextActionClass(rowNextAction.tone)}"
												onclick={() => {
													enrollSession = s;
													enrollScopeType = s.scope_type || 'class';
													enrollClassId = s.class_id;
													enrollGradeLevel = s.scope_type === 'grade' ? s.scope_ref : 'VII';
												}}
											>
												Aksi: {rowNextAction.label}
											</button>
										{:else if rowNextAction.href}
											<a class="inline-flex items-center rounded-md border px-2 py-0.5 text-[11px] font-semibold transition-colors {nextActionClass(rowNextAction.tone)}" href={resolve(rowNextAction.href as '/')}>
												Aksi: {rowNextAction.label}
											</a>
										{/if}
									</div>
								</Table.Cell>
								<Table.Cell>
									<Badge class={statusClass(s.status)}>{statusLabel[s.status] ?? s.status}</Badge>
								</Table.Cell>
								<Table.Cell>
									<div class="flex gap-1 flex-wrap">
										{#if s.status === 'draft'}
											<Button
												size="xs"
												variant="outline"
												onclick={() => {
													enrollSession = s;
													enrollScopeType = s.scope_type || 'class';
													enrollClassId = s.class_id;
													enrollGradeLevel = s.scope_type === 'grade' ? s.scope_ref : 'VII';
												}}
											>
												Daftarkan Siswa
											</Button>
											<LoadingButton size="xs" onclick={() => updateStatus(s.id, 'scheduled')} loading={statusBusyId === s.id} disabled={(statusBusyId !== '' && statusBusyId !== s.id) || rowPackageIssues.length > 0 || rowScheduleState === 'overdue'} loadingLabel="Memproses...">
												Jadwalkan
											</LoadingButton>
											<LoadingButton size="xs" variant="destructive" onclick={() => deleteSession(s.id, s.title)} loading={deleteBusyId === s.id} disabled={deleteBusyId !== '' && deleteBusyId !== s.id} loadingLabel="Menghapus...">
												Hapus
											</LoadingButton>
										{:else if s.status === 'scheduled'}
											<LoadingButton size="xs" onclick={() => updateStatus(s.id, 'active')} loading={statusBusyId === s.id} disabled={(statusBusyId !== '' && statusBusyId !== s.id) || rowPackageIssues.length > 0 || rowOperationalIssues.length > 0 || rowScheduleState === 'overdue'} loadingLabel="Memproses...">Mulai</LoadingButton>
											<LoadingButton size="xs" variant="outline" onclick={() => updateStatus(s.id, 'cancelled')} loading={statusBusyId === s.id} disabled={statusBusyId !== '' && statusBusyId !== s.id} loadingLabel="Memproses...">Batalkan</LoadingButton>
										{:else if s.status === 'active'}
											<LoadingButton size="xs" onclick={() => updateStatus(s.id, 'finished')} loading={statusBusyId === s.id} disabled={statusBusyId !== '' && statusBusyId !== s.id} loadingLabel="Memproses...">Selesaikan</LoadingButton>
										{/if}
										{#if s.status === 'finished' || s.status === 'active'}
											<a href={resolve(`/asesmen/sesi/${s.id}`)} class="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium border border-input bg-background hover:bg-muted text-foreground transition-colors">
												Detail
											</a>
										{/if}
									</div>
								</Table.Cell>
							</Table.Row>
						{:else}
							<Table.Row>
								<Table.Cell colspan={8} class="text-center text-muted-foreground py-8">
									{currentSessions.length === 0 ? 'Belum ada sesi ujian' : 'Tidak ada sesi pada filter ini'}
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
				</div>

				<div class="grid gap-3 p-4 lg:hidden">
					{#each visibleSessions as s (s.id)}
						{@const rowPackageQuality = packageQualitySummary(s.package_id)}
						{@const rowPackageIssues = packageQualityIssues(s.package_id)}
						{@const rowOperationalIssues = sessionOperationalIssues(s)}
						{@const rowReadinessIssues = sessionRowReadinessIssues(s)}
						{@const rowNextAction = nextSessionAction(s)}
						{@const rowScheduleState = sessionScheduleState(s)}
						{@const rowScheduleQuickAction = scheduleQuickAction(s, rowScheduleState)}
						<div class="rounded-2xl border border-border bg-card p-4 shadow-sm">
							<div class="flex items-start justify-between gap-3">
								<div class="min-w-0">
									<p class="text-sm font-semibold text-foreground">{s.title}</p>
									<p class="mt-1 text-xs text-muted-foreground">{s.package_title}</p>
								</div>
								<Badge class={statusClass(s.status)}>{sessionActionLabel(s.status)}</Badge>
							</div>
							<div class="mt-3 flex flex-wrap items-center gap-2">
								<Badge variant="outline" class="text-xs">{scopeSummary(s)}</Badge>
								<Badge variant="outline" class="text-xs">{mixPolicyLabel(s.mix_policy)}</Badge>
								<Badge variant="secondary">{s.participant_count} peserta</Badge>
							</div>
							<div class="mt-2 flex flex-wrap items-center gap-1.5">
								<Badge variant="outline" class="bg-card text-xs">{rowPackageQuality.totalCount} soal</Badge>
								{#if rowPackageIssues.length > 0}
									<Badge class="border-destructive/30 bg-destructive/10 text-destructive text-xs">{rowPackageIssues.join(', ')}</Badge>
								{:else if rowPackageQuality.missingCount > 0}
									<Badge class="border-warning/30 bg-warning/10 text-warning text-xs">{rowPackageQuality.missingCount} metadata kurang</Badge>
								{:else}
									<Badge class="border-primary/20 bg-primary/10 text-primary text-xs">Paket siap</Badge>
								{/if}
							</div>
							<div class="mt-2 flex flex-wrap items-center gap-1.5">
								{#if rowReadinessIssues.length === 0}
									<Badge class="border-primary/20 bg-primary/10 text-primary text-xs">Siap mulai</Badge>
								{:else}
									<Badge class="{rowPackageIssues.length > 0 ? 'border-destructive/30 bg-destructive/10 text-destructive' : 'border-warning/30 bg-warning/10 text-warning'} text-xs">{rowReadinessIssues.length} atensi</Badge>
								{/if}
								<Badge variant="outline" class="bg-card text-xs">{s.room_count} ruang / {s.total_capacity} kursi</Badge>
								<Badge variant="outline" class="bg-card text-xs">{s.assigned_participant_count}/{s.participant_count} ditempatkan</Badge>
								<Badge variant="outline" class="bg-card text-xs">{s.proctor_assignment_count} pengawas</Badge>
								{#each rowReadinessIssues.slice(0, 2) as issue (issue)}
									<span class="text-xs text-muted-foreground">{issue}</span>
								{/each}
								{#if rowNextAction.kind === 'enroll'}
									<button
										type="button"
										class="inline-flex items-center rounded-md border px-2 py-1 text-xs font-semibold transition-colors {nextActionClass(rowNextAction.tone)}"
										onclick={() => {
											enrollSession = s;
											enrollScopeType = s.scope_type || 'class';
											enrollClassId = s.class_id;
											enrollGradeLevel = s.scope_type === 'grade' ? s.scope_ref : 'VII';
										}}
									>
										Aksi: {rowNextAction.label}
									</button>
								{:else if rowNextAction.href}
									<a class="inline-flex items-center rounded-md border px-2 py-1 text-xs font-semibold transition-colors {nextActionClass(rowNextAction.tone)}" href={resolve(rowNextAction.href as '/')}>
										Aksi: {rowNextAction.label}
									</a>
								{/if}
							</div>
							<div class="mt-3 flex flex-wrap items-center gap-2">
								<p class="text-xs text-muted-foreground">{fmtDt(s.scheduled_start)}</p>
								<Badge class="{scheduleStateClass(rowScheduleState)} text-xs">{scheduleStateLabel(rowScheduleState)}</Badge>
								{#if rowScheduleQuickAction?.kind === 'schedule'}
									<button
										type="button"
										class="inline-flex items-center rounded-md border px-2 py-1 text-xs font-semibold transition-colors {nextActionClass(rowScheduleQuickAction.tone)}"
										onclick={() => openScheduleEditor(s)}
									>
										{rowScheduleQuickAction.label}
									</button>
								{:else if rowScheduleQuickAction?.status === 'cancelled'}
									<LoadingButton size="xs" variant="outline" onclick={() => updateStatus(s.id, 'cancelled')} loading={statusBusyId === s.id} disabled={statusBusyId !== '' && statusBusyId !== s.id} loadingLabel="Memproses...">
										{rowScheduleQuickAction.label}
									</LoadingButton>
								{:else if rowScheduleQuickAction?.status === 'finished'}
									<LoadingButton size="xs" onclick={() => updateStatus(s.id, 'finished')} loading={statusBusyId === s.id} disabled={statusBusyId !== '' && statusBusyId !== s.id} loadingLabel="Memproses...">
										{rowScheduleQuickAction.label}
									</LoadingButton>
								{:else if rowScheduleQuickAction?.href}
									<a class="inline-flex items-center rounded-md border px-2 py-1 text-xs font-semibold transition-colors {nextActionClass(rowScheduleQuickAction.tone)}" href={resolve(rowScheduleQuickAction.href as '/')}>
										{rowScheduleQuickAction.label}
									</a>
								{/if}
							</div>
							<div class="mt-4 flex flex-wrap gap-2">
								{#if s.status === 'draft'}
									<Button
										size="sm"
										variant="outline"
										onclick={() => {
											enrollSession = s;
											enrollScopeType = s.scope_type || 'class';
											enrollClassId = s.class_id;
											enrollGradeLevel = s.scope_type === 'grade' ? s.scope_ref : 'VII';
										}}
									>
										Daftarkan
									</Button>
									<LoadingButton size="sm" onclick={() => updateStatus(s.id, 'scheduled')} loading={statusBusyId === s.id} disabled={(statusBusyId !== '' && statusBusyId !== s.id) || rowPackageIssues.length > 0 || rowScheduleState === 'overdue'} loadingLabel="Memproses...">Jadwalkan</LoadingButton>
									<LoadingButton size="sm" variant="destructive" onclick={() => deleteSession(s.id, s.title)} loading={deleteBusyId === s.id} disabled={deleteBusyId !== '' && deleteBusyId !== s.id} loadingLabel="Menghapus...">Hapus</LoadingButton>
								{:else if s.status === 'scheduled'}
									<LoadingButton size="sm" onclick={() => updateStatus(s.id, 'active')} loading={statusBusyId === s.id} disabled={(statusBusyId !== '' && statusBusyId !== s.id) || rowPackageIssues.length > 0 || rowOperationalIssues.length > 0 || rowScheduleState === 'overdue'} loadingLabel="Memproses...">Mulai</LoadingButton>
									<LoadingButton size="sm" variant="outline" onclick={() => updateStatus(s.id, 'cancelled')} loading={statusBusyId === s.id} disabled={statusBusyId !== '' && statusBusyId !== s.id} loadingLabel="Memproses...">Batalkan</LoadingButton>
								{:else if s.status === 'active'}
									<LoadingButton size="sm" onclick={() => updateStatus(s.id, 'finished')} loading={statusBusyId === s.id} disabled={statusBusyId !== '' && statusBusyId !== s.id} loadingLabel="Memproses...">Selesaikan</LoadingButton>
								{/if}
								{#if s.status === 'finished' || s.status === 'active'}
									<a href={resolve(`/asesmen/sesi/${s.id}`)} class="inline-flex items-center rounded-md px-3 py-1.5 text-sm font-medium border border-input bg-background hover:bg-muted text-foreground transition-colors">
										Detail
									</a>
								{/if}
							</div>
						</div>
					{:else}
						<div class="rounded-2xl border border-dashed border-border bg-muted/50 px-4 py-10 text-center text-sm text-muted-foreground">
							{currentSessions.length === 0 ? 'Belum ada sesi ujian' : 'Tidak ada sesi pada filter ini'}
						</div>
					{/each}
				</div>
			</Card.Content>
		</Card.Root>
		</div>
		{/snippet}
	</AsyncContent>
</div>
