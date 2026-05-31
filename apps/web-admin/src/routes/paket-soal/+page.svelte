<script lang="ts">
	import { onMount } from 'svelte';
	import { clientApiPathWithQuery, readClientApiData } from '$lib/client/api';
	import {
		buildPackageReadinessCsv,
		filterPackagesByReadiness,
		packageReadinessBadge,
		packageReadiness
	} from '$lib/asesmen/package-readiness';
	import type { PackageReadinessFilter, PackageReadinessStatus } from '$lib/asesmen/package-readiness';

	type PackageOption = {
		id: string;
		event_id?: string;
		subject_id: string;
		subject_code?: string;
		subject_name: string;
		title: string;
		description?: string;
		duration_minutes: number;
		question_count: number;
		session_count: number;
		locked?: boolean;
		snapshot_version?: number;
		readiness?: PackageReadinessStatus;
	};

	type SubjectOption = {
		id?: string;
		subject_id?: string;
		code?: string;
		subject_code?: string;
		name?: string;
		subject_name?: string;
		total?: number;
	};

	type QuestionPoolItem = {
		id: string;
		subject_id: string;
		subject_name?: string;
		subject_code?: string;
		code?: string;
		question_text?: string;
		question_type?: string;
		difficulty?: string;
		status?: string;
		workflow_status?: string;
		target_level?: string | null;
		material_topic?: string;
		cognitive_level?: string;
		hots_flag?: boolean;
		author_display_name?: string;
		author_username?: string;
		package_count?: number;
	};

	type QuestionsPayload = QuestionPoolItem[] | { items?: QuestionPoolItem[] };
	type SubjectsPayload = SubjectOption[] | { subjects?: SubjectOption[] };

	type PackageReadiness = PackageReadinessStatus;

	type PackageDetailRow = PackageOption & {
		randomize_questions?: boolean;
		randomize_options?: boolean;
		source_mode?: string;
		draw_pg_count?: number;
		draw_essay_count?: number;
		random_seed?: string;
		is_active?: boolean;
		locked_at?: unknown;
		lock_reason?: string;
	};

	type PackageQuestion = {
		question_id: string | Record<string, unknown>;
		question_code?: string;
		question_text?: string;
		question_type?: string;
		position?: number;
		points?: number;
		status?: string;
		workflow_status?: string;
		target_level?: string | Record<string, unknown> | null;
		cp_ref?: string;
		tp_ref?: string;
		kd_ref?: string;
		material_topic?: string;
		cognitive_level?: string;
		hots_flag?: boolean;
	};

	type PackageDetail = {
		package: PackageDetailRow;
		questions?: PackageQuestion[];
		readiness?: PackageReadiness;
	};

	let packages = $state<PackageOption[]>([]);
	let subjectOptions = $state<SubjectOption[]>([]);
	let subjectPublishedCounts = $state<Record<string, number>>({});
	let questionPool = $state<QuestionPoolItem[]>([]);
	let selectedQuestionIds = $state<Set<string>>(new Set());
	let loading = $state(true);
	let loadingSubjects = $state(false);
	let loadingPool = $state(false);
	let savingPackage = $state(false);
	let showBuilder = $state(false);
	let builderMode = $state<'create' | 'edit'>('create');
	let activePackageId = $state('');
	let activeDetail = $state<PackageDetail | null>(null);
	let loadingDetail = $state(false);
	let packageActionBusy = $state<'lock' | 'clone' | null>(null);
	let error = $state('');
	let builderError = $state('');
	let builderNotice = $state('');
	let search = $state('');
	let subjectFilter = $state('all');
	let readinessFilter = $state<PackageReadinessFilter>('all');
	let poolSearch = $state('');
	let poolType = $state('all');
	let poolLevel = $state('all');
	let draft = $state({
		title: '',
		description: '',
		subject_id: '',
		duration_minutes: 90,
		randomize_questions: true,
		randomize_options: true,
		source_mode: 'teacher_class',
		draw_pg_count: 0,
		draw_essay_count: 0,
		random_seed: '',
		is_active: true
	});

	const subjects = $derived(
		Array.from(new Map(packages.map((item) => [item.subject_id, item])).values())
			.sort((a, b) => a.subject_name.localeCompare(b.subject_name))
	);
	const normalizedSubjects = $derived(
		Array.from(new Map([
			...subjectOptions.map((item) => [subjectId(item), item] as const),
			...subjects.map((item) => [item.subject_id, {
				id: item.subject_id,
				code: item.subject_code,
				name: item.subject_name
			} as SubjectOption] as const)
		].filter(([id]) => Boolean(id))).values())
			.sort((a, b) => subjectName(a).localeCompare(subjectName(b)))
	);
	const searchedPackages = $derived(packages.filter((item) => {
		const query = search.trim().toLowerCase();
		const haystack = `${item.title} ${item.subject_name} ${item.subject_code ?? ''} ${item.description ?? ''}`.toLowerCase();
		const matchesSearch = !query || haystack.includes(query);
		const matchesSubject = subjectFilter === 'all' || item.subject_id === subjectFilter;
		return matchesSearch && matchesSubject;
	}));
	const filteredPackages = $derived(filterPackagesByReadiness(searchedPackages, readinessFilter));
	const availablePool = $derived(questionPool.filter((item) => {
		const query = poolSearch.trim().toLowerCase();
		const haystack = `${item.code ?? ''} ${item.question_text ?? ''} ${item.material_topic ?? ''} ${item.author_display_name ?? ''}`.toLowerCase();
		const matchesSearch = !query || haystack.includes(query);
		const matchesType = poolType === 'all' || item.question_type === poolType;
		const matchesLevel = poolLevel === 'all' || (item.target_level ?? '') === poolLevel;
		return matchesSearch && matchesType && matchesLevel;
	}));
	const allSelected = $derived(availablePool.length > 0 && availablePool.every((item) => selectedQuestionIds.has(item.id)));
	const someSelected = $derived(selectedQuestionIds.size > 0 && !allSelected);
	const selectedQuestions = $derived(questionPool.filter((item) => selectedQuestionIds.has(item.id)));
	const selectedPgCount = $derived(selectedQuestions.filter((item) => isPgType(item.question_type)).length);
	const selectedEssayCount = $derived(selectedQuestions.filter((item) => !isPgType(item.question_type)).length);
	const totalQuestions = $derived(packages.reduce((total, item) => total + Number(item.question_count ?? 0), 0));
	const emptyCount = $derived(packages.filter((item) => Number(packageReadiness(item).question_count ?? 0) === 0).length);
	const readyCount = $derived(packages.filter((item) => packageReadiness(item).ready && !packageReadiness(item).locked).length);
	const gapCount = $derived(packages.filter((item) => Number(packageReadiness(item).metadata_gap_count ?? 0) > 0).length);
	const lockedCount = $derived(packages.filter((item) => packageReadiness(item).locked || item.locked).length);
	const usedCount = $derived(packages.filter((item) => Number(packageReadiness(item).session_count ?? item.session_count ?? 0) > 0).length);
	const activeReadiness = $derived(activeDetail?.readiness ?? null);
	const activeLocked = $derived(Boolean(activeReadiness?.locked || hasValue(activeDetail?.package?.locked_at) || activeDetail?.package?.locked));

	onMount(() => {
		void loadPackages();
		void loadSubjects();
		void loadSubjectQuestionCounts();
	});

	async function loadPackages() {
		loading = true;
		error = '';
		try {
			const response = await fetch('/api/asesmen/package-options');
			const optionItems = await readClientApiData<PackageOption[]>(response);
			const detailResults = await Promise.allSettled(
				optionItems.map((item) =>
					fetch(`/api/asesmen/packages/${encodeURIComponent(item.id)}`)
						.then((detailResponse) => readClientApiData<PackageDetail>(detailResponse))
				)
			);
			packages = optionItems.map((item, index) => {
				const result = detailResults[index];
				const readiness = result?.status === 'fulfilled' ? result.value.readiness : undefined;
				return {
					...item,
					readiness,
					locked: readiness?.locked ?? item.locked,
					session_count: readiness?.session_count ?? item.session_count,
					question_count: readiness?.question_count ?? item.question_count
				};
			});
			if (detailResults.some((result) => result.status === 'rejected')) {
				error = 'Daftar paket terbuka, tetapi sebagian ringkasan kesiapan belum lengkap. Klik Refresh untuk mencoba lagi.';
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'Daftar Paket Soal belum dapat dibuka.';
			packages = [];
		} finally {
			loading = false;
		}
	}

	async function loadSubjects() {
		loadingSubjects = true;
		try {
			const response = await fetch('/api/bank-soal/soal-support/subjects');
			const payload = await readClientApiData<SubjectsPayload>(response);
			subjectOptions = Array.isArray(payload) ? payload : payload.subjects ?? [];
		} catch {
			subjectOptions = [];
		} finally {
			loadingSubjects = false;
		}
	}

	async function loadSubjectQuestionCounts() {
		try {
			const params = new URLSearchParams({
				status: 'published',
				workflow_status: 'published',
				limit: '500',
				sort: 'newest'
			});
			const response = await fetch(clientApiPathWithQuery('/api/bank-soal/questions', params));
			const payload = await readClientApiData<QuestionsPayload>(response);
			const items = Array.isArray(payload) ? payload : payload.items ?? [];
			const next: Record<string, number> = {};
			for (const item of items) {
				if (!item.subject_id) continue;
				next[item.subject_id] = (next[item.subject_id] ?? 0) + 1;
			}
			subjectPublishedCounts = next;
		} catch {
			subjectPublishedCounts = {};
		}
	}

	async function loadQuestionPool(subjectId = draft.subject_id, clearSelection = true) {
		builderError = '';
		builderNotice = '';
		questionPool = [];
		if (clearSelection) selectedQuestionIds = new Set();
		if (!subjectId) {
			builderError = 'Pilih mata pelajaran dulu.';
			return;
		}
		loadingPool = true;
		try {
			const params = new URLSearchParams({
				subject_id: subjectId,
				workflow_status: 'published',
				status: 'published',
				limit: '200',
				sort: 'newest'
			});
			const response = await fetch(clientApiPathWithQuery('/api/bank-soal/questions', params));
			const payload = await readClientApiData<QuestionsPayload>(response);
			questionPool = Array.isArray(payload) ? payload : payload.items ?? [];
			if (questionPool.length === 0) {
				builderNotice = 'Belum ada soal terbit untuk mapel ini. Terbitkan soal dulu di Bank Soal.';
			}
		} catch (e) {
			builderError = e instanceof Error ? e.message : 'Pool Bank Soal belum dapat dibuka.';
		} finally {
			loadingPool = false;
		}
	}

	async function openBuilder() {
		if (normalizedSubjects.length === 0) {
			await loadSubjects();
		}
		if (Object.keys(subjectPublishedCounts).length === 0) {
			await loadSubjectQuestionCounts();
		}
		builderMode = 'create';
		activePackageId = '';
		activeDetail = null;
		showBuilder = true;
		builderError = '';
		builderNotice = '';
		draft = {
			title: '',
			description: '',
			subject_id: defaultBuilderSubjectId(),
			duration_minutes: 90,
			randomize_questions: true,
			randomize_options: true,
			source_mode: 'teacher_class',
			draw_pg_count: 0,
			draw_essay_count: 0,
			random_seed: '',
			is_active: true
		};
		selectedQuestionIds = new Set();
		if (draft.subject_id) void loadQuestionPool(draft.subject_id, true);
	}

	function closeBuilder() {
		if (savingPackage || packageActionBusy) return;
		showBuilder = false;
	}

	async function openPackageDetail(item: PackageOption) {
		builderMode = 'edit';
		activePackageId = item.id;
		showBuilder = true;
		loadingDetail = true;
		builderError = '';
		builderNotice = '';
		questionPool = [];
		selectedQuestionIds = new Set();
		try {
			const response = await fetch(`/api/asesmen/packages/${encodeURIComponent(item.id)}`);
			const detail = await readClientApiData<PackageDetail>(response);
			activeDetail = detail;
			const pkg = detail.package;
			draft = {
				title: pkg.title ?? item.title,
				description: pkg.description ?? item.description ?? '',
				subject_id: uuidValue(pkg.subject_id) || item.subject_id,
				duration_minutes: Number(pkg.duration_minutes ?? item.duration_minutes ?? 90),
				randomize_questions: Boolean(pkg.randomize_questions),
				randomize_options: Boolean(pkg.randomize_options),
				source_mode: pkg.source_mode || 'teacher_class',
				draw_pg_count: Number(pkg.draw_pg_count ?? detail.readiness?.target_pg_count ?? 0),
				draw_essay_count: Number(pkg.draw_essay_count ?? detail.readiness?.target_essay_count ?? 0),
				random_seed: pkg.random_seed ?? '',
				is_active: pkg.is_active ?? true
			};
			selectedQuestionIds = new Set((detail.questions ?? []).map((q) => uuidValue(q.question_id)).filter(Boolean));
			await loadQuestionPool(draft.subject_id, false);
		} catch (e) {
			builderError = e instanceof Error ? e.message : 'Detail paket belum dapat dibuka.';
		} finally {
			loadingDetail = false;
		}
	}

	function toggleQuestion(id: string) {
		const next = new Set(selectedQuestionIds);
		if (next.has(id)) next.delete(id); else next.add(id);
		selectedQuestionIds = next;
	}

	function toggleSelectAll() {
		if (allSelected) {
			selectedQuestionIds = new Set();
			return;
		}
		const next = new Set(selectedQuestionIds);
		for (const item of availablePool) next.add(item.id);
		selectedQuestionIds = next;
	}

	async function submitPackage() {
		builderError = '';
		builderNotice = '';
		const title = draft.title.trim();
		if (!title) {
			builderError = 'Judul paket wajib diisi.';
			return;
		}
		if (!draft.subject_id) {
			builderError = 'Mata pelajaran wajib dipilih.';
			return;
		}
		if (selectedQuestionIds.size === 0) {
			builderError = 'Pilih minimal 1 soal dari Bank Soal.';
			return;
		}
		if (builderMode === 'edit' && activeLocked) {
			builderError = 'Paket sudah terkunci. Gunakan Clone/Revisi untuk mengubah.';
			return;
		}
		savingPackage = true;
		try {
			const questionIds = Array.from(selectedQuestionIds);
			const weights = Object.fromEntries(questionIds.map((id) => [id, questionPoint(id)]));
			const payload = {
				event_id: '',
				subject_id: draft.subject_id,
				title,
				description: draft.description.trim(),
				duration_minutes: Number(draft.duration_minutes),
				randomize_questions: draft.randomize_questions,
				randomize_options: draft.randomize_options,
				source_mode: draft.source_mode || 'teacher_class',
				draw_pg_count: Number(draft.draw_pg_count || selectedPgCount),
				draw_essay_count: Number(draft.draw_essay_count || selectedEssayCount),
				random_seed: draft.random_seed.trim(),
				is_active: draft.is_active,
				question_ids: questionIds,
				question_weights: weights
			};
			if (builderMode === 'edit' && activePackageId) {
				await readClientApiData<PackageDetail>(await fetch(`/api/asesmen/packages/${encodeURIComponent(activePackageId)}`, {
					method: 'PUT',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(payload)
				}));
				const detail = await readClientApiData<PackageDetail>(await fetch(`/api/asesmen/packages/${encodeURIComponent(activePackageId)}/questions`, {
					method: 'PUT',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ question_ids: questionIds, question_weights: weights })
				}));
				activeDetail = detail;
				builderNotice = `Paket “${title}” diperbarui dengan ${questionIds.length} soal.`;
			} else {
				await readClientApiData<unknown>(await fetch('/api/asesmen/packages', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(payload)
				}));
				builderNotice = `Paket “${title}” tersimpan dengan ${questionIds.length} soal.`;
				draft = { ...draft, title: '', description: '', draw_pg_count: 0, draw_essay_count: 0 };
				selectedQuestionIds = new Set();
			}
			await loadPackages();
		} catch (e) {
			builderError = e instanceof Error ? e.message : 'Paket belum dapat disimpan.';
		} finally {
			savingPackage = false;
		}
	}

	async function lockActivePackage() {
		if (!activePackageId || packageActionBusy) return;
		if (!activeReadiness?.ready) {
			builderError = 'Paket belum siap dikunci. Lengkapi target soal dan metadata dulu.';
			return;
		}
		const ok = window.confirm('Kunci paket ini? Setelah dikunci, isi soal tidak bisa diedit dan perubahan harus lewat clone/revisi.');
		if (!ok) return;
		packageActionBusy = 'lock';
		builderError = '';
		try {
			await readClientApiData<unknown>(await fetch(`/api/asesmen/packages/${encodeURIComponent(activePackageId)}/lock`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ reason: 'Dikunci dari modul Paket Soal' })
			}));
			builderNotice = 'Paket berhasil dikunci dan snapshot dibuat.';
			await openPackageDetail({ ...(activeDetail?.package ?? {}), id: activePackageId } as PackageOption);
			await loadPackages();
		} catch (e) {
			builderError = e instanceof Error ? e.message : 'Paket belum dapat dikunci.';
		} finally {
			packageActionBusy = null;
		}
	}

	async function cloneActivePackage() {
		if (!activePackageId || packageActionBusy) return;
		const title = window.prompt('Judul paket revisi/clone:', `${draft.title} - Revisi`);
		if (title === null) return;
		packageActionBusy = 'clone';
		builderError = '';
		try {
			const detail = await readClientApiData<PackageDetail>(await fetch(`/api/asesmen/packages/${encodeURIComponent(activePackageId)}/clone`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ title })
			}));
			builderNotice = `Clone paket “${detail.package.title}” berhasil dibuat.`;
			await loadPackages();
			await openPackageDetail({ ...detail.package, id: uuidValue(detail.package.id) } as PackageOption);
		} catch (e) {
			builderError = e instanceof Error ? e.message : 'Clone paket belum dapat dibuat.';
		} finally {
			packageActionBusy = null;
		}
	}

	function exportReadinessCsv() {
		const csv = buildPackageReadinessCsv(filteredPackages);
		const blob = new Blob([csv], { type: 'text/csv;charset=utf-8' });
		const url = URL.createObjectURL(blob);
		const link = document.createElement('a');
		const stamp = new Date().toISOString().slice(0, 10);
		link.href = url;
		link.download = `kesiapan-paket-soal-${stamp}.csv`;
		document.body.appendChild(link);
		link.click();
		link.remove();
		URL.revokeObjectURL(url);
	}

	function statusLabel(item: PackageOption) {
		return packageReadinessBadge(item).label;
	}

	function statusClass(item: PackageOption) {
		const tone = packageReadinessBadge(item).tone;
		if (tone === 'slate') return 'border-slate-300 bg-slate-100 text-slate-700';
		if (tone === 'amber') return 'border-amber-300 bg-amber-50 text-amber-700';
		if (tone === 'sky') return 'border-sky-300 bg-sky-50 text-sky-700';
		if (tone === 'emerald') return 'border-emerald-300 bg-emerald-50 text-emerald-700';
		return 'border-orange-300 bg-orange-50 text-orange-700';
	}

	function subjectId(item: SubjectOption) {
		return item.subject_id ?? item.id ?? '';
	}

	function subjectName(item: SubjectOption) {
		return item.subject_name ?? item.name ?? 'Mapel tanpa nama';
	}

	function subjectCode(item: SubjectOption) {
		return item.subject_code ?? item.code ?? '-';
	}

	function subjectPublishedCount(id: string) {
		return subjectPublishedCounts[id] ?? 0;
	}

	function subjectOptionLabel(item: SubjectOption) {
		const id = subjectId(item);
		const count = subjectPublishedCount(id);
		return `${subjectCode(item)} · ${subjectName(item)} (${count} soal terbit)`;
	}

	function defaultBuilderSubjectId() {
		return subjectId(normalizedSubjects.find((item) => subjectPublishedCount(subjectId(item)) > 0) ?? normalizedSubjects[0] ?? {} as SubjectOption);
	}

	function questionTypeLabel(value = '') {
		const normalized = value.toLowerCase();
		if (normalized === 'multiple_choice' || normalized === 'pg') return 'PG';
		if (normalized === 'essay') return 'Essay';
		if (normalized === 'true_false') return 'Benar/Salah';
		if (normalized === 'short_answer') return 'Isian';
		return value || '-';
	}

	function isPgType(value = '') {
		const normalized = value.toLowerCase();
		return normalized === 'multiple_choice' || normalized === 'pg';
	}

	function hasMetadataGap(item: QuestionPoolItem) {
		return !item.target_level || !item.cognitive_level || !item.material_topic;
	}

	function hasPackageQuestionMetadataGap(item: PackageQuestion) {
		return !textValue(item.target_level) || !item.cognitive_level || !item.material_topic || (!item.tp_ref && !item.kd_ref);
	}

	function uuidValue(value: unknown): string {
		if (typeof value === 'string') return value;
		if (value && typeof value === 'object') {
			const record = value as Record<string, unknown>;
			if (typeof record.String === 'string') return record.String;
			if (typeof record.string === 'string') return record.string;
		}
		return '';
	}

	function textValue(value: unknown): string {
		if (typeof value === 'string') return value;
		if (value && typeof value === 'object') {
			const record = value as Record<string, unknown>;
			if (typeof record.String === 'string') return record.String;
			if (typeof record.string === 'string') return record.string;
		}
		return '';
	}

	function hasValue(value: unknown): boolean {
		if (!value) return false;
		if (typeof value === 'string') return value.trim() !== '';
		if (typeof value === 'object') {
			const record = value as Record<string, unknown>;
			if (typeof record.Valid === 'boolean') return record.Valid;
			if (typeof record.valid === 'boolean') return record.valid;
		}
		return true;
	}

	function questionPoint(id: string) {
		const existing = (activeDetail?.questions ?? []).find((item) => uuidValue(item.question_id) === id);
		return Number(existing?.points ?? 1);
	}

	function readinessLabel(readiness: PackageReadiness | null) {
		if (!readiness) return 'Belum dicek';
		if (readiness.locked) return 'Terkunci';
		if (readiness.ready) return 'Siap dikunci';
		if (readiness.status === 'kosong') return 'Kosong';
		return 'Perlu dilengkapi';
	}

	function readinessClass(readiness: PackageReadiness | null) {
		if (readiness?.locked) return 'border-slate-300 bg-slate-100 text-slate-700';
		if (readiness?.ready) return 'border-emerald-300 bg-emerald-50 text-emerald-700';
		if (readiness?.status === 'kosong') return 'border-amber-300 bg-amber-50 text-amber-700';
		return 'border-orange-300 bg-orange-50 text-orange-700';
	}
</script>

<svelte:head>
	<title>Paket Soal | MTsN 2 Kolut</title>
</svelte:head>

<div class="space-y-5 pb-16">
	<section class="rounded-2xl border border-border bg-card p-5 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
			<div class="space-y-2">
				<p class="text-xs font-semibold tracking-[0.22em] text-muted-foreground uppercase">Modul Mandiri</p>
				<h1 class="text-2xl font-bold tracking-tight text-foreground md:text-3xl">Paket Soal</h1>
				<p class="max-w-3xl text-sm leading-6 text-muted-foreground">
					Dapur perakitan paket dari Bank Soal sebelum dipakai di Asesmen/CBT. Paket dapat dipakai ulang untuk beberapa rombel, kegiatan, simulasi, atau ujian susulan.
				</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<a class="inline-flex h-10 items-center justify-center rounded-md border px-4 text-sm font-semibold text-foreground hover:bg-muted" href="/bank-soal">Buka Bank Soal</a>
				<a class="inline-flex h-10 items-center justify-center rounded-md border px-4 text-sm font-semibold text-foreground hover:bg-muted" href="/asesmen">Pakai di Asesmen</a>
				<button type="button" class="inline-flex h-10 items-center justify-center rounded-md border px-4 text-sm font-semibold text-foreground hover:bg-muted" onclick={() => void loadPackages()} disabled={loading}>{loading ? 'Memuat…' : 'Refresh'}</button>
				<button type="button" class="inline-flex h-10 items-center justify-center rounded-md bg-primary px-4 text-sm font-semibold text-primary-foreground shadow-sm transition hover:bg-primary/90" onclick={() => void openBuilder()}>Buat Paket</button>
			</div>
		</div>
	</section>

	{#if error}
		<p class="rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive" role="alert">{error}</p>
	{/if}

	<section class="grid gap-3 md:grid-cols-5">
		<div class="rounded-xl border bg-background p-4"><p class="text-xs font-medium text-muted-foreground">Total paket</p><p class="mt-1 text-2xl font-bold">{packages.length}</p></div>
		<div class="rounded-xl border bg-background p-4"><p class="text-xs font-medium text-muted-foreground">Total soal tertaut</p><p class="mt-1 text-2xl font-bold">{totalQuestions}</p></div>
		<div class="rounded-xl border bg-background p-4"><p class="text-xs font-medium text-muted-foreground">Siap dikunci</p><p class="mt-1 text-2xl font-bold">{readyCount}</p></div>
		<div class="rounded-xl border bg-background p-4"><p class="text-xs font-medium text-muted-foreground">Kosong/gap</p><p class="mt-1 text-2xl font-bold">{emptyCount}/{gapCount}</p></div>
		<div class="rounded-xl border bg-background p-4"><p class="text-xs font-medium text-muted-foreground">Terkunci/dipakai</p><p class="mt-1 text-2xl font-bold">{lockedCount}/{usedCount}</p></div>
	</section>

	<section class="rounded-2xl border bg-card shadow-sm">
		<div class="border-b border-border px-4 py-3">
			<h2 class="text-base font-semibold text-foreground">Alur Opsi 4</h2>
			<p class="text-xs leading-5 text-muted-foreground">Modul ini berdiri sendiri, tapi tetap menjadi jembatan antara Bank Soal dan Asesmen.</p>
		</div>
		<div class="grid gap-3 p-4 md:grid-cols-3">
			<div class="rounded-xl border bg-background p-3 text-sm"><p class="font-semibold text-foreground">1. Bank Soal</p><p class="mt-1 text-xs leading-5 text-muted-foreground">Guru/admin membuat, mereview, dan menerbitkan soal.</p></div>
			<div class="rounded-xl border border-primary/30 bg-primary/5 p-3 text-sm"><p class="font-semibold text-primary">2. Paket Soal</p><p class="mt-1 text-xs leading-5 text-muted-foreground">Panitia merakit paket, validasi kesiapan, lock, clone/revisi.</p></div>
			<div class="rounded-xl border bg-background p-3 text-sm"><p class="font-semibold text-foreground">3. Asesmen/CBT</p><p class="mt-1 text-xs leading-5 text-muted-foreground">Kegiatan ujian memilih paket siap untuk rombel/sesi.</p></div>
		</div>
	</section>

	<section class="rounded-2xl border bg-card shadow-sm">
		<div class="border-b border-border px-4 py-3">
			<div class="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
				<div><h2 class="text-base font-semibold text-foreground">Daftar Paket</h2><p class="text-xs text-muted-foreground">Dibaca dari paket aktif yang sudah ada di backend.</p></div>
				<div class="grid gap-2 sm:grid-cols-2 lg:min-w-[50rem] lg:grid-cols-[1.2fr_1fr_1fr_auto]">
					<input class="min-w-0 rounded-md border bg-background px-3 py-2 text-sm" placeholder="Cari paket/mapel..." bind:value={search} />
					<select class="min-w-0 rounded-md border bg-background px-3 py-2 text-sm" bind:value={subjectFilter}>
						<option value="all">Semua mapel</option>
						{#each subjects as subject}<option value={subject.subject_id}>{subject.subject_name}</option>{/each}
					</select>
					<select class="min-w-0 rounded-md border bg-background px-3 py-2 text-sm" bind:value={readinessFilter}>
						<option value="all">Semua status</option>
						<option value="empty">Kosong</option>
						<option value="ready">Siap dikunci</option>
						<option value="kurang_pg">Kurang PG</option>
						<option value="kurang_essay">Kurang Essay</option>
						<option value="metadata_gap">Metadata Gap</option>
						<option value="locked">Terkunci</option>
						<option value="used">Dipakai</option>
					</select>
					<button type="button" class="rounded-md border bg-background px-3 py-2 text-sm font-semibold text-foreground hover:bg-muted disabled:opacity-60" onclick={exportReadinessCsv} disabled={filteredPackages.length === 0}>Export CSV</button>
				</div>
			</div>
		</div>
		{#if loading}
			<p class="px-4 py-8 text-center text-sm text-muted-foreground">Memuat paket soal…</p>
		{:else if filteredPackages.length === 0}
			<div class="px-4 py-8 text-center"><p class="text-sm font-semibold text-foreground">Belum ada paket sesuai filter.</p><p class="mt-1 text-xs text-muted-foreground">Klik Buat Paket untuk merakit paket dari soal terbit.</p></div>
		{:else}
			<div class="overflow-x-auto">
				<table class="min-w-[980px] w-full text-left text-sm">
					<thead class="border-b bg-muted/40 text-xs text-muted-foreground">
						<tr><th class="px-4 py-3 font-semibold">Paket</th><th class="px-4 py-3 font-semibold">Mapel</th><th class="px-4 py-3 font-semibold">Komposisi</th><th class="px-4 py-3 font-semibold">Gap</th><th class="px-4 py-3 font-semibold">Durasi</th><th class="px-4 py-3 font-semibold">Pemakaian</th><th class="px-4 py-3 font-semibold">Status</th><th class="px-4 py-3 font-semibold text-right">Aksi</th></tr>
					</thead>
					<tbody class="divide-y">
						{#each filteredPackages as item (item.id)}
							<tr class="hover:bg-muted/30">
								<td class="px-4 py-3"><p class="font-semibold text-foreground">{item.title}</p><p class="mt-1 max-w-md truncate text-xs text-muted-foreground">{item.description || 'Belum ada deskripsi.'}</p></td>
								<td class="px-4 py-3"><p class="font-medium text-foreground">{item.subject_name}</p><p class="text-xs text-muted-foreground">{item.subject_code || '-'}</p></td>
								<td class="px-4 py-3 text-xs text-muted-foreground"><p class="font-semibold text-foreground">{packageReadiness(item).question_count ?? item.question_count} soal</p><p>PG {packageReadiness(item).pg_count ?? 0}/{packageReadiness(item).target_pg_count ?? 20} · Essay {packageReadiness(item).essay_count ?? 0}/{packageReadiness(item).target_essay_count ?? 5}</p></td>
								<td class="px-4 py-3 text-xs text-muted-foreground"><p>Metadata: {packageReadiness(item).metadata_gap_count ?? 0}</p><p>Belum terbit: {packageReadiness(item).unpublished_count ?? 0}</p></td>
								<td class="px-4 py-3 text-muted-foreground">{item.duration_minutes || 0} menit</td>
								<td class="px-4 py-3 text-muted-foreground">{item.session_count || 0} sesi</td>
								<td class="px-4 py-3"><span class={`rounded-full border px-2.5 py-1 text-[11px] font-semibold ${statusClass(item)}`}>{statusLabel(item)}</span></td>
								<td class="px-4 py-3 text-right"><div class="flex justify-end gap-2"><button type="button" class="rounded-md border px-3 py-2 text-xs font-semibold text-foreground hover:bg-muted" onclick={() => void openPackageDetail(item)}>Detail/Edit</button><a class="rounded-md border px-3 py-2 text-xs font-semibold text-foreground hover:bg-muted" href={`/asesmen?paket=${item.id}`}>Gunakan</a></div></td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</section>

	<section class="rounded-2xl border border-dashed bg-muted/30 p-4">
		<h2 class="text-sm font-semibold text-foreground">Status iterasi ini</h2>
		<ul class="mt-2 list-disc space-y-1 pl-5 text-sm leading-6 text-muted-foreground">
			<li>Modul mandiri membaca paket aktif dan menjadi jembatan Bank Soal → Paket Soal → Asesmen.</li>
			<li>Builder membuat paket langsung dari soal terbit, sehingga tidak membuat paket kosong yang belum valid.</li>
			<li>Detail/edit, validasi kesiapan, lock/snapshot, dan clone/revisi tersedia di aksi Detail/Edit.</li>
			<li>Daftar paket sekarang bisa difilter Siap/Kurang PG/Kurang Essay/Metadata Gap/Terkunci dan diekspor CSV untuk audit sebelum gladi.</li>
		</ul>
	</section>

	{#if showBuilder}
		<div class="fixed inset-0 z-50 flex justify-end" role="dialog" aria-modal="true" aria-labelledby="package-builder-title">
			<button type="button" class="absolute inset-0 bg-slate-950/35 backdrop-blur-[1px]" aria-label="Tutup builder paket" onclick={closeBuilder}></button>
			<aside class="relative flex h-full w-full max-w-5xl flex-col border-l border-border bg-card shadow-2xl">
				<div class="border-b border-border px-5 py-4">
					<div class="flex items-start justify-between gap-3">
						<div class="space-y-1">
							<p class="text-xs font-semibold tracking-[0.18em] text-muted-foreground uppercase">{builderMode === 'edit' ? 'Detail Paket' : 'Builder Paket'}</p>
							<h2 id="package-builder-title" class="text-lg font-bold text-foreground">{builderMode === 'edit' ? 'Detail/Edit Paket Soal' : 'Buat Paket dari Bank Soal'}</h2>
							<p class="text-xs leading-5 text-muted-foreground">{builderMode === 'edit' ? 'Edit metadata/soal selama belum terkunci, cek kesiapan, lalu lock atau clone.' : 'Pilih mapel, ambil soal terbit, lalu simpan sebagai paket siap dipakai Asesmen.'}</p>
						</div>
						<button type="button" class="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-full border text-sm font-bold text-muted-foreground hover:bg-muted" aria-label="Tutup" onclick={closeBuilder}>×</button>
					</div>
				</div>
				<div class="min-h-0 flex-1 overflow-y-auto px-5 py-4">
					<div class="grid gap-4 lg:grid-cols-[22rem_1fr]">
						<section class="space-y-4 rounded-xl border bg-background p-4">
							<h3 class="text-sm font-semibold text-foreground">Metadata Paket</h3>
							<label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Judul paket</span><input class="w-full rounded-md border bg-card px-3 py-2 text-sm" placeholder="Contoh: Paket UAS IPA VII" bind:value={draft.title} disabled={activeLocked} /></label>
							<label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Mata pelajaran</span><select class="w-full rounded-md border bg-card px-3 py-2 text-sm" bind:value={draft.subject_id} onchange={() => void loadQuestionPool(draft.subject_id, true)} disabled={loadingSubjects || builderMode === 'edit' || activeLocked}>
								<option value="">Pilih mapel</option>
								{#each normalizedSubjects as subject}<option value={subjectId(subject)}>{subjectOptionLabel(subject)}</option>{/each}
							</select></label>
							<label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Deskripsi</span><textarea class="min-h-20 w-full rounded-md border bg-card px-3 py-2 text-sm" placeholder="Opsional" bind:value={draft.description} disabled={activeLocked}></textarea></label>
							<div class="grid gap-3 sm:grid-cols-3 lg:grid-cols-1">
								<label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Durasi menit</span><input type="number" min="1" max="360" class="w-full rounded-md border bg-card px-3 py-2 text-sm" bind:value={draft.duration_minutes} disabled={activeLocked} /></label>
								<label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Target PG</span><input type="number" min="0" class="w-full rounded-md border bg-card px-3 py-2 text-sm" bind:value={draft.draw_pg_count} disabled={activeLocked} /></label>
								<label class="space-y-1.5"><span class="text-xs font-medium text-muted-foreground">Target Essay</span><input type="number" min="0" class="w-full rounded-md border bg-card px-3 py-2 text-sm" bind:value={draft.draw_essay_count} disabled={activeLocked} /></label>
							</div>
							<label class="flex items-center gap-2 text-sm"><input type="checkbox" bind:checked={draft.randomize_questions} disabled={activeLocked} /> <span>Acak urutan soal</span></label>
							<label class="flex items-center gap-2 text-sm"><input type="checkbox" bind:checked={draft.randomize_options} disabled={activeLocked} /> <span>Acak opsi PG</span></label>
							<div class="rounded-lg border bg-muted/40 p-3 text-xs leading-5 text-muted-foreground">
								<p class="font-semibold text-foreground">Dipilih: {selectedQuestionIds.size} soal</p>
								<p>PG: {selectedPgCount} · Non-PG/Essay: {selectedEssayCount}</p>
								<p>Target kosong otomatis memakai jumlah soal yang dipilih.</p>
							</div>
							{#if builderMode === 'edit'}
								<div class="rounded-lg border bg-card p-3 text-xs leading-5">
									<div class="flex items-center justify-between gap-2"><p class="font-semibold text-foreground">Validasi Kesiapan</p><span class={`rounded-full border px-2 py-1 font-semibold ${readinessClass(activeReadiness)}`}>{readinessLabel(activeReadiness)}</span></div>
									<p class="mt-2 text-muted-foreground">PG {activeReadiness?.pg_count ?? selectedPgCount}/{activeReadiness?.target_pg_count ?? (draft.draw_pg_count || 20)} · Essay {activeReadiness?.essay_count ?? selectedEssayCount}/{activeReadiness?.target_essay_count ?? (draft.draw_essay_count || 5)}</p>
									<p class="text-muted-foreground">Gap metadata: {activeReadiness?.metadata_gap_count ?? 0} · Belum terbit: {activeReadiness?.unpublished_count ?? 0} · Poin: {activeReadiness?.total_points ?? selectedQuestionIds.size}</p>
									{#if activeLocked}<p class="mt-1 font-medium text-slate-700">Terkunci: perubahan harus lewat Clone/Revisi.</p>{/if}
								</div>
							{/if}
							{#if builderError}<p class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive" role="alert">{builderError}</p>{/if}
							{#if builderNotice}<p class="rounded-md border border-emerald-300 bg-emerald-50 px-3 py-2 text-sm text-emerald-800" role="status">{builderNotice}</p>{/if}
							<div class="flex flex-wrap gap-2"><button type="button" class="rounded-md border px-4 py-2 text-sm font-semibold hover:bg-muted" onclick={() => void loadQuestionPool(draft.subject_id, builderMode === 'create')} disabled={loadingPool || !draft.subject_id}>{loadingPool ? 'Memuat…' : 'Muat Soal'}</button><button type="button" class="rounded-md bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground disabled:opacity-60" onclick={() => void submitPackage()} disabled={savingPackage || activeLocked}>{savingPackage ? 'Menyimpan…' : builderMode === 'edit' ? 'Simpan Edit' : 'Simpan Paket'}</button>{#if builderMode === 'edit'}<button type="button" class="rounded-md border px-4 py-2 text-sm font-semibold hover:bg-muted disabled:opacity-60" onclick={() => void lockActivePackage()} disabled={packageActionBusy !== null || activeLocked || !activeReadiness?.ready}>{packageActionBusy === 'lock' ? 'Mengunci…' : 'Lock Paket'}</button><button type="button" class="rounded-md border px-4 py-2 text-sm font-semibold hover:bg-muted disabled:opacity-60" onclick={() => void cloneActivePackage()} disabled={packageActionBusy !== null}>{packageActionBusy === 'clone' ? 'Clone…' : 'Clone/Revisi'}</button>{/if}</div>
						</section>

						{#if builderMode === 'edit'}
							<section class="space-y-3 rounded-xl border bg-background p-4 lg:col-span-2">
								<div class="flex items-center justify-between gap-2"><h3 class="text-sm font-semibold text-foreground">Soal Saat Ini</h3><p class="text-xs text-muted-foreground">{activeDetail?.questions?.length ?? 0} soal tersimpan</p></div>
								<div class="overflow-x-auto rounded-lg border">
									<table class="min-w-[760px] w-full text-left text-sm">
										<thead class="border-b bg-muted/40 text-xs text-muted-foreground"><tr><th class="px-3 py-2">Urut</th><th class="px-3 py-2">Kode & Soal</th><th class="px-3 py-2">Bentuk</th><th class="px-3 py-2">Level</th><th class="px-3 py-2">Poin</th><th class="px-3 py-2">Mutu</th></tr></thead>
										<tbody class="divide-y">
											{#if loadingDetail}
												<tr><td colspan="6" class="px-3 py-6 text-center text-muted-foreground">Memuat detail…</td></tr>
											{:else if !activeDetail?.questions?.length}
												<tr><td colspan="6" class="px-3 py-6 text-center text-muted-foreground">Belum ada soal tersimpan.</td></tr>
											{:else}
												{#each activeDetail.questions as q (uuidValue(q.question_id))}
													<tr><td class="px-3 py-2 text-xs text-muted-foreground">{q.position}</td><td class="px-3 py-2"><p class="font-semibold text-foreground">{q.question_code || 'Tanpa kode'}</p><p class="line-clamp-1 max-w-xl text-xs text-muted-foreground">{q.question_text}</p></td><td class="px-3 py-2 text-xs text-muted-foreground">{questionTypeLabel(q.question_type)}</td><td class="px-3 py-2"><span class="rounded-full border px-2 py-1 text-[11px] font-semibold">{textValue(q.target_level) || 'Kosong'}</span></td><td class="px-3 py-2 text-xs text-muted-foreground">{q.points ?? 1}</td><td class="px-3 py-2"><span class={`rounded-full border px-2 py-1 text-[11px] font-semibold ${hasPackageQuestionMetadataGap(q) ? 'border-amber-300 bg-amber-50 text-amber-700' : 'border-emerald-300 bg-emerald-50 text-emerald-700'}`}>{hasPackageQuestionMetadataGap(q) ? 'Gap' : 'OK'}</span></td></tr>
												{/each}
											{/if}
										</tbody>
									</table>
								</div>
							</section>
						{/if}
						<section class="space-y-3 rounded-xl border bg-background p-4">
							<div class="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
								<div><h3 class="text-sm font-semibold text-foreground">Pool Bank Soal Terbit</h3><p class="text-xs text-muted-foreground">Hanya soal terbit sesuai mapel yang bisa dimasukkan paket.</p></div>
								<div class="grid gap-2 sm:grid-cols-3 lg:min-w-[34rem]">
									<input class="min-w-0 rounded-md border bg-card px-3 py-2 text-sm" placeholder="Cari kode/teks/pembuat..." bind:value={poolSearch} />
									<select class="min-w-0 rounded-md border bg-card px-3 py-2 text-sm" bind:value={poolType}><option value="all">Semua bentuk</option><option value="multiple_choice">PG</option><option value="essay">Essay</option><option value="true_false">Benar/Salah</option><option value="short_answer">Isian</option></select>
									<select class="min-w-0 rounded-md border bg-card px-3 py-2 text-sm" bind:value={poolLevel}><option value="all">Semua tingkat</option><option value="VII">VII</option><option value="VIII">VIII</option><option value="IX">IX</option></select>
								</div>
							</div>
							<div class="overflow-x-auto rounded-lg border">
								<table class="min-w-[860px] w-full text-left text-sm">
									<thead class="border-b bg-muted/40 text-xs text-muted-foreground">
										<tr><th class="w-12 px-3 py-3"><input type="checkbox" checked={allSelected} indeterminate={someSelected} onchange={toggleSelectAll} aria-label="Pilih semua soal terlihat" disabled={activeLocked} /></th><th class="px-3 py-3 font-semibold">Kode & Soal</th><th class="px-3 py-3 font-semibold">Bentuk</th><th class="px-3 py-3 font-semibold">Tingkat</th><th class="px-3 py-3 font-semibold">Pembuat</th><th class="px-3 py-3 font-semibold">Mutu</th><th class="px-3 py-3 font-semibold">Pakai</th></tr>
									</thead>
									<tbody class="divide-y">
										{#if loadingPool}
											<tr><td colspan="7" class="px-3 py-8 text-center text-muted-foreground">Memuat soal…</td></tr>
										{:else if availablePool.length === 0}
											<tr><td colspan="7" class="px-3 py-8 text-center text-muted-foreground">Belum ada soal sesuai filter.</td></tr>
										{:else}
											{#each availablePool as item (item.id)}
												<tr class={selectedQuestionIds.has(item.id) ? 'bg-muted/50' : 'hover:bg-muted/30'}>
													<td class="px-3 py-3"><input type="checkbox" checked={selectedQuestionIds.has(item.id)} onchange={() => toggleQuestion(item.id)} aria-label={`Pilih soal ${item.code || item.id}`} disabled={activeLocked} /></td>
													<td class="px-3 py-3"><p class="font-semibold text-foreground">{item.code || 'Tanpa kode'}</p><p class="mt-1 line-clamp-1 max-w-xl text-xs text-muted-foreground">{item.question_text || 'Teks soal kosong'}</p></td>
													<td class="px-3 py-3 text-xs text-muted-foreground">{questionTypeLabel(item.question_type)}</td>
													<td class="px-3 py-3"><span class="rounded-full border px-2 py-1 text-[11px] font-semibold">{item.target_level || 'Kosong'}</span></td>
													<td class="max-w-32 truncate px-3 py-3 text-xs text-muted-foreground">{item.author_display_name || item.author_username || '-'}</td>
													<td class="px-3 py-3"><span class={`rounded-full border px-2 py-1 text-[11px] font-semibold ${hasMetadataGap(item) ? 'border-amber-300 bg-amber-50 text-amber-700' : 'border-emerald-300 bg-emerald-50 text-emerald-700'}`}>{hasMetadataGap(item) ? 'Gap' : 'OK'}</span></td>
													<td class="px-3 py-3 text-xs text-muted-foreground">{Number(item.package_count ?? 0) > 0 ? `Dipakai ${item.package_count}` : 'Baru'}</td>
												</tr>
											{/each}
										{/if}
									</tbody>
								</table>
							</div>
							<p class="text-xs text-muted-foreground">{selectedQuestionIds.size} dari {availablePool.length} soal terlihat dipilih.</p>
						</section>
					</div>
				</div>
			</aside>
		</div>
	{/if}
</div>
