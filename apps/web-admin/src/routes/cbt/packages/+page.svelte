<script lang="ts">
	import { onMount } from 'svelte';
	import { SvelteMap, SvelteSet } from 'svelte/reactivity';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import OperationStatusPanel from '$lib/components/OperationStatusPanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { confirmChallenge } from '$lib/confirm-dialog';
	import { clientApiPathWithQuery, readClientApiData, readClientJson } from '$lib/client/api';

	type CbtPackage = {
		id: string; subject_id: string; subject_name: string; subject_code: string;
		title: string; description: string; duration_minutes: number;
		randomize_questions: boolean; is_active: boolean;
		question_count: number; created_at: string;
	};
	type Question = {
		id: string; subject_id: string; subject_code: string;
		code: string; question_text: string; difficulty: string; status: string;
		question_type?: string; cp_ref?: string; tp_ref?: string; kd_ref?: string;
		material_topic?: string; cognitive_level?: string; hots_flag?: boolean;
	};
	type Subject = { id: string; name: string; code: string; };
	type PackagesOverview = {
		packages: CbtPackage[];
		allQuestions: Question[];
		subjects: Subject[];
	};
	type CbtPackagesPayload = {
		packages?: CbtPackage[];
		questions?: unknown[];
		error?: string;
		message?: string;
	};
	type AcademicPayload = {
		subjects?: Subject[];
		error?: string;
		message?: string;
	};
	type QuestionListPayload = {
		items?: Question[];
		meta?: {
			total?: number;
			limit?: number;
			offset?: number;
		};
	};
	type BlueprintBucket = {
		label: string;
		count: number;
	};
	type BlueprintMatrixRow = {
		key: string;
		cp: string;
		tp: string;
		kd: string;
		topic: string;
		cognitive: string;
		count: number;
		hotsCount: number;
		types: string[];
		missing: boolean;
	};

	const questionPageSize = 100;
	const maxQuestionPages = 20;

	let packages = $state<CbtPackage[]>([]);
	let allQuestions = $state<Question[]>([]);
	let subjects = $state<Subject[]>([]);
	let packagesPromise = $state<Promise<PackagesOverview> | null>(null);
	let showForm = $state(false);

	let fSubjectId = $state('');
	let fTitle = $state('');
	let fDescription = $state('');
	let fDuration = $state(60);
	let fRandomize = $state(false);
	let fActive = $state(true);
	let fSelectedIds = new SvelteSet<string>();
	let fBusy = $state(false);
	let deleteBusyId = $state('');
	let operationState = $state<{ tone: 'success' | 'error' | 'warning' | 'info'; title: string; message: string } | null>(null);
	let packagesRequestId = 0;

	let questionPool = $derived(
		fSubjectId
			? allQuestions.filter(q => q.subject_id === fSubjectId && q.status === 'published')
			: []
	);
	let selectedQuestions = $derived(questionPool.filter((q) => fSelectedIds.has(q.id)));
	let selectedBlueprintMatrix = $derived(buildBlueprintMatrix(selectedQuestions));
	let selectedTypeBuckets = $derived(countByLabel(selectedQuestions, (q) => questionTypeLabel(q.question_type)));
	let selectedCognitiveBuckets = $derived(countByLabel(selectedQuestions, (q) => compactValue(q.cognitive_level, 'Belum level')));
	let selectedHotsCount = $derived(selectedQuestions.filter((q) => q.hots_flag).length);
	let selectedBlueprintMissingCount = $derived(selectedQuestions.filter(questionHasBlueprintGap).length);

	function toggleQuestion(id: string) {
		if (fSelectedIds.has(id)) fSelectedIds.delete(id);
		else fSelectedIds.add(id);
	}

	function handleSubjectChange() {
		fSelectedIds.clear();
	}

	function isRecord(value: unknown): value is Record<string, unknown> {
		return typeof value === 'object' && value !== null;
	}

	function parseQuestionPage(payload: unknown) {
		if (Array.isArray(payload)) return { items: payload as Question[], total: payload.length };
		if (!isRecord(payload)) return { items: [], total: 0 };
		const data = payload as QuestionListPayload;
		const items = Array.isArray(data.items) ? data.items : [];
		const total = data.meta?.total ?? items.length;
		return { items, total };
	}

	function parseSubjects(payload: AcademicPayload | unknown) {
		return isRecord(payload) && Array.isArray(payload.subjects) ? (payload.subjects as Subject[]) : [];
	}

	async function fetchQuestionsPage(offset: number) {
		const params = new URLSearchParams({
			limit: String(questionPageSize),
			offset: String(offset),
		});
		const payload = await fetch(clientApiPathWithQuery('/api/cbt/questions', params))
			.then((response) => readClientApiData<unknown>(response, 'Gagal memuat bank soal'));
		return parseQuestionPage(payload);
	}

	async function fetchAllQuestions() {
		const firstPage = await fetchQuestionsPage(0);
		const questionsById = new SvelteMap(firstPage.items.map((question) => [question.id, question]));
		let loaded = firstPage.items.length;
		let pages = 1;
		while (loaded < firstPage.total && pages < maxQuestionPages) {
			const page = await fetchQuestionsPage(loaded);
			if (page.items.length === 0) break;
			for (const question of page.items) questionsById.set(question.id, question);
			loaded += page.items.length;
			pages += 1;
		}
		return Array.from(questionsById.values());
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

	function difficultyLabel(value: string | null | undefined) {
		const labels: Record<string, string> = {
			easy: 'Mudah',
			medium: 'Sedang',
			hard: 'Sulit',
		};
		const normalized = compactValue(value, '');
		return labels[normalized] ?? (normalized || '-');
	}

	function questionHasBlueprintGap(question: Question) {
		return !compactValue(question.cp_ref, '')
			|| (!compactValue(question.tp_ref, '') && !compactValue(question.kd_ref, ''))
			|| !compactValue(question.cognitive_level, '');
	}

	function countByLabel(questions: Question[], selector: (question: Question) => string): BlueprintBucket[] {
		const counts = new SvelteMap<string, number>();
		for (const question of questions) {
			const label = selector(question);
			counts.set(label, (counts.get(label) ?? 0) + 1);
		}
		return Array.from(counts.entries())
			.map(([label, count]) => ({ label, count }))
			.sort((a, b) => b.count - a.count || a.label.localeCompare(b.label));
	}

	function buildBlueprintMatrix(questions: Question[]): BlueprintMatrixRow[] {
		const rows = new SvelteMap<string, BlueprintMatrixRow>();
		for (const question of questions) {
			const cp = compactValue(question.cp_ref, 'Belum CP');
			const tp = compactValue(question.tp_ref, 'Belum TP');
			const kd = compactValue(question.kd_ref, 'Belum KD');
			const topic = compactValue(question.material_topic, 'Belum materi');
			const cognitive = compactValue(question.cognitive_level, 'Belum level');
			const missing = questionHasBlueprintGap(question);
			const key = [cp, tp, kd, topic, cognitive].join('|');
			const row = rows.get(key) ?? {
				key,
				cp,
				tp,
				kd,
				topic,
				cognitive,
				count: 0,
				hotsCount: 0,
				types: [],
				missing,
			};
			row.count += 1;
			if (question.hots_flag) row.hotsCount += 1;
			const typeLabel = questionTypeLabel(question.question_type);
			if (!row.types.includes(typeLabel)) row.types.push(typeLabel);
			row.missing = row.missing || missing;
			rows.set(key, row);
		}
		return Array.from(rows.values()).sort((a, b) => Number(b.missing) - Number(a.missing) || b.count - a.count || a.cp.localeCompare(b.cp));
	}

	async function fetchOverview(): Promise<PackagesOverview> {
		const [packagesPayload, questionItems, academicPayload] = await Promise.all([
			fetch('/api/cbt/packages').then((response) => readClientApiData<CbtPackagesPayload>(response, 'Gagal memuat data paket')),
			fetchAllQuestions(),
			fetch('/api/academic').then((response) => readClientApiData<AcademicPayload>(response, 'Gagal memuat data akademik')),
		]);
		return {
			packages: packagesPayload.packages ?? [],
			allQuestions: questionItems,
			subjects: parseSubjects(academicPayload),
		};
	}

	function applyOverview(overview: PackagesOverview) {
		packages = overview.packages;
		allQuestions = overview.allQuestions;
		subjects = overview.subjects;
	}

	function load() {
		const requestId = ++packagesRequestId;
		packages = [];
		allQuestions = [];
		subjects = [];
		packagesPromise = fetchOverview().then((overview) => {
			if (requestId !== packagesRequestId) return { packages, allQuestions, subjects };
			applyOverview(overview);
			return overview;
		}).catch((error: unknown) => {
			if (requestId === packagesRequestId) throw error;
			return { packages, allQuestions, subjects };
		});
	}

	async function refreshPackages() {
		if (!packagesPromise) {
			load();
			return;
		}
		const requestId = ++packagesRequestId;
		try {
			const overview = await fetchOverview();
			if (requestId !== packagesRequestId) return;
			applyOverview(overview);
			packagesPromise = Promise.resolve(overview);
		} catch (error) {
			if (requestId === packagesRequestId) {
				packagesPromise = Promise.resolve({ packages, allQuestions, subjects });
				toast.error(packagesErrorMessage(error));
			}
		}
	}

	function retryPackages(reset?: () => void) {
		reset?.();
		load();
	}

	function packagesErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat data paket';
	}

	function handlePackagesRenderError(error: unknown) {
		console.error('CBT packages render failed', error);
	}

	function showToast(msg: string) {
		toast.success(msg);
	}

	function showError(msg: string) {
		toast.error(msg);
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

	function packageLegacyMutationPath(id: string) {
		return clientApiPathWithQuery('/api/cbt/packages', new URLSearchParams({ id }));
	}

	async function createPackage() {
		if (!fSubjectId || !fTitle || !fDuration) return;
		fBusy = true;
		try {
			const res = await fetch('/api/cbt/packages', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					subject_id: fSubjectId, title: fTitle, description: fDescription,
					duration_minutes: fDuration, randomize_questions: fRandomize,
					is_active: fActive, question_ids: selectedQuestions.map((question) => question.id),
				}),
			});
			await readClientJson<unknown>(res);
			fSubjectId = ''; fTitle = ''; fDescription = ''; fDuration = 60;
			fRandomize = false; fActive = true; fSelectedIds.clear();
			showForm = false;
			setOperationState('success', 'Paket Berhasil Dibuat', 'Paket ujian baru sudah tersimpan dan siap dipakai untuk sesi ujian.');
			showToast('Paket ujian berhasil dibuat');
			await refreshPackages();
		} catch (error) {
			showError(mutationErrorMessage(error, 'Gagal membuat paket. Periksa koneksi lalu coba lagi.'));
		} finally { fBusy = false; }
	}

	async function deletePackage(id: string, title: string) {
		if (!(await confirmPhrase('Hapus Paket Ujian', `Paket "${title}" akan dihapus dari daftar. Tindakan ini tidak bisa dibatalkan dari layar operator.`, 'HAPUS'))) return;
		deleteBusyId = id;
		try {
			const res = await fetch(packageLegacyMutationPath(id), { method: 'DELETE' });
			await readClientJson<unknown>(res);
			setOperationState('warning', 'Paket Dihapus', `Paket "${title}" sudah dihapus dari daftar paket ujian.`);
			showToast('Paket dihapus');
			await refreshPackages();
		} catch (error) {
			setOperationState('error', 'Paket Gagal Dihapus', 'Periksa kembali apakah paket masih dipakai oleh sesi aktif atau coba ulang beberapa saat lagi.');
			showError(mutationErrorMessage(error, 'Gagal menghapus paket. Periksa koneksi lalu coba lagi.'));
		} finally {
			deleteBusyId = '';
		}
	}

	onMount(() => {
		void load();
	});
</script>

<svelte:head><title>Paket Ujian CBT — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-slate-800">Paket Ujian CBT</h1>
			<p class="text-sm text-slate-500 mt-1">Buat dan kelola paket soal untuk sesi ujian</p>
		</div>
		<LoadingButton onclick={() => (showForm = !showForm)}>
			{showForm ? 'Batal' : '+ Buat Paket'}
		</LoadingButton>
	</div>

	{#if operationState}
		<OperationStatusPanel {...operationState} />
	{/if}

	{#if showForm}
		<Card.Root>
			<Card.Header class="pb-2">
				<Card.Title class="text-base">Buat Paket Ujian Baru</Card.Title>
			</Card.Header>
			<Card.Content class="space-y-4">
				<div class="grid gap-3 sm:grid-cols-2">
					<div>
						<label for="package-subject-id" class="text-xs text-slate-500 mb-1 block">Mata Pelajaran <span class="text-red-500">*</span></label>
						<select id="package-subject-id" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fSubjectId} onchange={handleSubjectChange}>
							<option value="">-- Pilih --</option>
								{#each subjects as s (s.id)}
								<option value={s.id}>{s.code} — {s.name}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="package-title" class="text-xs text-slate-500 mb-1 block">Nama Paket <span class="text-red-500">*</span></label>
						<Input id="package-title" placeholder="mis: UTS Matematika Sem 1 2025" bind:value={fTitle} />
					</div>
					<div>
						<label for="package-duration" class="text-xs text-slate-500 mb-1 block">Durasi (menit) <span class="text-red-500">*</span></label>
						<Input id="package-duration" type="number" min={10} max={300} bind:value={fDuration} />
					</div>
					<div class="flex items-end gap-4 pb-1">
						<label class="flex items-center gap-2 text-sm">
							<input type="checkbox" bind:checked={fRandomize} class="rounded" />
							Acak urutan soal
						</label>
						<label class="flex items-center gap-2 text-sm">
							<input type="checkbox" bind:checked={fActive} class="rounded" />
							Paket aktif
						</label>
					</div>
				</div>

				<div>
					<label for="package-description" class="text-xs text-slate-500 mb-1 block">Deskripsi (opsional)</label>
					<Textarea id="package-description" placeholder="Keterangan paket ujian..." rows={2} bind:value={fDescription} />
				</div>

				{#if fSubjectId}
					<div>
						<div class="text-xs text-slate-500 mb-2 block">
							Pilih Soal dari Bank ({questionPool.length} soal tersedia)
							{#if selectedQuestions.length > 0}
								— <span class="text-green-700 font-medium">{selectedQuestions.length} dipilih</span>
							{/if}
						</div>
						{#if questionPool.length === 0}
							<p class="text-sm text-slate-400 py-4 text-center border rounded-md">
								Belum ada soal berstatus "Terbit" untuk mata pelajaran ini
							</p>
						{:else}
							<div class="border rounded-md max-h-64 overflow-y-auto">
									{#each questionPool as q (q.id)}
									<label class="flex items-start gap-3 px-3 py-2 hover:bg-slate-50 cursor-pointer border-b last:border-b-0">
										<input type="checkbox" checked={fSelectedIds.has(q.id)} onchange={() => toggleQuestion(q.id)} class="mt-0.5 rounded" />
										<div class="flex-1 min-w-0">
											<p class="text-sm line-clamp-1">{q.question_text}</p>
											<div class="flex flex-wrap gap-1 mt-0.5">
												{#if q.code}
													<span class="text-xs text-slate-400 font-mono">{q.code}</span>
												{/if}
												<Badge variant="outline" class="text-xs py-0">{questionTypeLabel(q.question_type)}</Badge>
												<Badge variant="outline" class="text-xs py-0">{difficultyLabel(q.difficulty)}</Badge>
												{#if q.cognitive_level}
													<Badge variant="secondary" class="text-xs py-0">{q.cognitive_level}</Badge>
												{/if}
												{#if q.hots_flag}
													<Badge class="border-amber-200 bg-amber-50 text-amber-700 text-xs py-0">HOTS</Badge>
												{/if}
											</div>
										</div>
									</label>
								{/each}
							</div>
						{/if}
						{#if selectedQuestions.length > 0}
							<div class="mt-3 rounded-lg border border-emerald-100 bg-emerald-50/50 p-3">
								<div class="flex flex-wrap items-start justify-between gap-2">
									<div>
										<p class="text-xs font-semibold uppercase tracking-[0.18em] text-emerald-900">Blueprint Paket Sementara</p>
										<p class="mt-1 text-xs text-emerald-800">
											{selectedQuestions.length} soal dipilih, {selectedBlueprintMatrix.length} kombinasi CP/TP/KD, {selectedHotsCount} HOTS.
										</p>
									</div>
									{#if selectedBlueprintMissingCount > 0}
										<Badge class="border-amber-200 bg-amber-50 text-amber-700">{selectedBlueprintMissingCount} perlu metadata</Badge>
									{:else}
										<Badge class="border-emerald-200 bg-white text-emerald-700">Blueprint lengkap</Badge>
									{/if}
								</div>

								<div class="mt-3 grid gap-3 xl:grid-cols-[0.75fr_1.25fr]">
									<div class="space-y-2 text-xs">
										<div>
											<p class="mb-1 font-semibold text-slate-600">Bentuk soal</p>
											<div class="flex flex-wrap gap-1">
												{#each selectedTypeBuckets as bucket (bucket.label)}
													<Badge variant="outline" class="bg-white">{bucket.label}: {bucket.count}</Badge>
												{/each}
											</div>
										</div>
										<div>
											<p class="mb-1 font-semibold text-slate-600">Level kognitif</p>
											<div class="flex flex-wrap gap-1">
												{#each selectedCognitiveBuckets as bucket (bucket.label)}
													<Badge variant={bucket.label === 'Belum level' ? 'secondary' : 'outline'} class="bg-white">{bucket.label}: {bucket.count}</Badge>
												{/each}
											</div>
										</div>
										<p class="rounded-md border border-emerald-100 bg-white px-2 py-1.5 text-slate-600">
											Gunakan panel ini untuk mencegah paket terlalu menumpuk di satu CP/KD sebelum sesi ujian dibuat.
										</p>
									</div>

									<div class="overflow-hidden rounded-md border border-emerald-100 bg-white">
										<div class="grid grid-cols-[1fr_1fr_1fr_0.8fr_0.5fr] gap-2 border-b bg-emerald-50 px-2 py-1.5 text-[11px] font-semibold uppercase tracking-[0.12em] text-emerald-900">
											<span>CP</span>
											<span>TP</span>
											<span>KD / Materi</span>
											<span>Level</span>
											<span class="text-right">Soal</span>
										</div>
										<div class="max-h-40 overflow-y-auto">
											{#each selectedBlueprintMatrix as row (row.key)}
												<div class="grid grid-cols-[1fr_1fr_1fr_0.8fr_0.5fr] gap-2 border-b px-2 py-1.5 text-xs last:border-b-0 {row.missing ? 'bg-amber-50/70' : ''}">
													<span class="min-w-0 truncate font-medium text-slate-700" title={row.cp}>{row.cp}</span>
													<span class="min-w-0 truncate text-slate-600" title={row.tp}>{row.tp}</span>
													<span class="min-w-0 truncate text-slate-600" title={`${row.kd} / ${row.topic}`}>{row.kd} / {row.topic}</span>
													<span class="min-w-0 truncate text-slate-600" title={row.types.join(', ')}>{row.cognitive}</span>
													<span class="text-right font-semibold text-slate-800">
														{row.count}
														{#if row.hotsCount > 0}
															<span class="text-amber-600">/{row.hotsCount}</span>
														{/if}
													</span>
												</div>
											{/each}
										</div>
									</div>
								</div>
							</div>
						{/if}
					</div>
				{/if}

				<div class="flex gap-2">
					<LoadingButton disabled={fBusy || !fSubjectId || !fTitle || !fDuration} onclick={() => void createPackage()} loading={fBusy} loadingLabel="Menyimpan...">
						{`Buat Paket${selectedQuestions.length > 0 ? ` (${selectedQuestions.length} soal)` : ''}`}
					</LoadingButton>
					<LoadingButton variant="outline" onclick={() => (showForm = false)}>Batal</LoadingButton>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	<AsyncContent promise={packagesPromise} onerror={handlePackagesRenderError}>
		{#snippet pending()}
			<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
				<Card.Content class="space-y-3 p-6">
					{#each Array.from({ length: 5 }) as _, index (`package-row-skeleton-${index}`)}
						<div class="grid gap-3 lg:grid-cols-[1.2fr_0.6fr_0.5fr_0.5fr_0.5fr_0.6fr_auto] lg:items-center">
							<Skeleton class="h-5 w-40" />
							<Skeleton class="h-6 w-16" />
							<Skeleton class="h-5 w-14" />
							<Skeleton class="h-5 w-12" />
							<Skeleton class="h-6 w-12" />
							<Skeleton class="h-6 w-16" />
							<Skeleton class="h-9 w-20 justify-self-end" />
						</div>
					{/each}
				</Card.Content>
			</Card.Root>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Paket Ujian Belum Tersaji"
				message={packagesErrorMessage(error)}
				onRetry={() => retryPackages(reset)}
			/>
		{/snippet}

		{#snippet children(value)}
			{@const overview = value as PackagesOverview}
			{@const currentPackages = overview.packages}
		<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
			<Card.Header class="pb-2">
				<Card.Title class="text-base">Daftar Paket ({currentPackages.length})</Card.Title>
			</Card.Header>
			<Card.Content class="p-0">
				<div class="hidden overflow-x-auto lg:block">
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Nama Paket</Table.Head>
							<Table.Head>Mapel</Table.Head>
							<Table.Head>Durasi</Table.Head>
							<Table.Head>Jml Soal</Table.Head>
							<Table.Head>Acak</Table.Head>
							<Table.Head>Status</Table.Head>
							<Table.Head></Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each currentPackages as p (p.id)}
							<Table.Row>
								<Table.Cell class="font-medium">{p.title}</Table.Cell>
								<Table.Cell>
									<Badge variant="outline" class="text-xs">{p.subject_code}</Badge>
								</Table.Cell>
								<Table.Cell class="text-slate-600">{p.duration_minutes} mnt</Table.Cell>
								<Table.Cell>
									<span class="font-mono text-sm">{p.question_count}</span>
								</Table.Cell>
								<Table.Cell>
									{#if p.randomize_questions}
										<Badge class="bg-emerald-100 text-emerald-700 border-emerald-200 text-xs">Ya</Badge>
									{:else}
										<Badge variant="secondary" class="text-xs">Tidak</Badge>
									{/if}
								</Table.Cell>
								<Table.Cell>
									{#if p.is_active}
										<Badge class="bg-emerald-100 text-emerald-700 border-emerald-200">Aktif</Badge>
									{:else}
										<Badge variant="secondary">Nonaktif</Badge>
									{/if}
								</Table.Cell>
								<Table.Cell>
									<LoadingButton
										variant="destructive"
										size="xs"
										onclick={() => deletePackage(p.id, p.title)}
										loading={deleteBusyId === p.id}
										disabled={deleteBusyId !== '' && deleteBusyId !== p.id}
										loadingLabel="Menghapus..."
									>
										Hapus
									</LoadingButton>
								</Table.Cell>
							</Table.Row>
						{:else}
							<Table.Row>
								<Table.Cell colspan={7} class="text-center text-slate-400 py-8">Belum ada paket ujian</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
				</div>

				<div class="grid gap-3 p-4 lg:hidden">
					{#each currentPackages as p (p.id)}
						<div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
							<div class="flex items-start justify-between gap-3">
								<div class="min-w-0">
									<p class="text-sm font-semibold text-slate-900">{p.title}</p>
									<p class="mt-1 text-xs text-slate-500">{p.subject_name} ({p.subject_code})</p>
								</div>
								{#if p.is_active}
									<Badge class="bg-emerald-100 text-emerald-700 border-emerald-200">Aktif</Badge>
								{:else}
									<Badge variant="secondary">Nonaktif</Badge>
								{/if}
							</div>
							<div class="mt-3 flex flex-wrap items-center gap-2">
								<Badge variant="outline" class="text-xs">{p.duration_minutes} menit</Badge>
								<Badge variant="secondary">{p.question_count} soal</Badge>
								{#if p.randomize_questions}
									<Badge class="bg-emerald-50 text-emerald-700 border-emerald-200 text-xs">Acak</Badge>
								{/if}
							</div>
							{#if p.description}
								<p class="mt-3 text-sm text-slate-600">{p.description}</p>
							{/if}
							<div class="mt-4">
								<LoadingButton
									variant="destructive"
									size="sm"
									class="w-full"
									onclick={() => deletePackage(p.id, p.title)}
									loading={deleteBusyId === p.id}
									disabled={deleteBusyId !== '' && deleteBusyId !== p.id}
									loadingLabel="Menghapus..."
								>
									Hapus
								</LoadingButton>
							</div>
						</div>
					{:else}
						<div class="rounded-2xl border border-dashed border-slate-300 bg-slate-50 px-4 py-10 text-center text-sm text-slate-500">
							Belum ada paket ujian
						</div>
					{/each}
				</div>
			</Card.Content>
		</Card.Root>
		{/snippet}
	</AsyncContent>
</div>
