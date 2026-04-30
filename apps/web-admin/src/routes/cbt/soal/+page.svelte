<script lang="ts">
	import { onMount } from 'svelte';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import EditorWrapper from '$lib/components/EditorWrapper.svelte';
	import { renderRichMathHtml } from '$lib/utils/render-rich-math';

	// ── Types ─────────────────────────────────────────────────────────────────
	type Subject = { id: string; name: string; code: string };
	type OptionItem = { label: string; text?: string; html?: string; latex?: string };
	type Question = {
		id: string;
		subject_id: string;
		subject_name: string;
		subject_code: string;
		code: string;
		question_type: string;
		stem_html: string;
		question_text: string;
		options: OptionItem[];
		answer_key: string;
		difficulty: string;
		status: string;
		workflow_status: string;
		authoring_mode?: string;
		author_username: string;
		created_at: string;
	};
	type QuestionListResponse = {
		items: Question[];
		meta?: { total: number; limit: number; offset: number };
	};
	type Template = {
		id: string;
		label: string;
		desc: string;
		stem: string;
		options: string[];
		answerKey: 'A' | 'B' | 'C' | 'D';
		isRtl?: boolean;
	};

	// ── Constants ─────────────────────────────────────────────────────────────
	const PAGE_SIZE = 15;
	const ANSWER_LABELS = ['A', 'B', 'C', 'D'] as const;
	const WORKFLOW_LABEL: Record<string, string> = {
		draft: 'Draft',
		review: 'Ditinjau',
		approved: 'Disetujui',
		rejected: 'Revisi',
	};
	const DIFFICULTY_LABEL: Record<string, string> = { easy: 'Mudah', medium: 'Sedang', hard: 'Sulit' };

	const DRAFT_KEY = (id: string | null) => `mtsn2-soal-komposer:${id ?? 'new'}`;

	const templates: Template[] = [
		{
			id: 'standar',
			label: 'Soal Standar',
			desc: 'PG biasa',
			stem: '<p>Perhatikan pernyataan berikut!</p><p><br></p><p>Berdasarkan pernyataan di atas, yang paling tepat adalah...</p>',
			options: ['', '', '', ''],
			answerKey: 'A',
		},
		{
			id: 'stimulus',
			label: 'Dengan Stimulus',
			desc: 'Teks + pertanyaan',
			stem: '<p><strong>Bacalah teks berikut dengan saksama!</strong></p><p><br></p><p>[Teks/wacana/stimulus di sini]</p><p><br></p><p>Berdasarkan teks di atas, apa yang dimaksud dengan...?</p>',
			options: ['', '', '', ''],
			answerKey: 'A',
		},
		{
			id: 'matematika',
			label: 'Matematika',
			desc: 'Rumus LaTeX',
			stem: '<p>Diketahui $a = 3$ dan $b = 4$. Tentukan nilai dari:</p><p>$$a^2 + b^2 = \\ldots$$</p>',
			options: ['25', '12', '7', '49'],
			answerKey: 'A',
		},
		{
			id: 'cerita',
			label: 'Soal Cerita',
			desc: 'Narasi + hitung',
			stem: '<p>Pak Ahmad memiliki [jumlah] [barang]. Setelah [kejadian], sisanya menjadi [jumlah baru]. Berapa [yang ditanya]?</p>',
			options: ['', '', '', ''],
			answerKey: 'A',
		},
		{
			id: 'grafik',
			label: 'Grafik / Data',
			desc: 'Interpretasi data',
			stem: '<p>Perhatikan grafik/tabel berikut!</p><p><br></p><p>[Sisipkan gambar grafik di sini menggunakan tombol 🖼]</p><p><br></p><p>Berdasarkan data tersebut, kesimpulan yang paling tepat adalah...</p>',
			options: ['', '', '', ''],
			answerKey: 'A',
		},
		{
			id: 'arab-mufradat',
			label: 'Arab: Mufradat',
			desc: 'Kosakata Arab (RTL)',
			stem: '<p>ما معنى الكلمة التالية ؟</p><p><br></p><p><strong>[الكلمة]</strong></p>',
			options: ['', '', '', ''],
			answerKey: 'A',
			isRtl: true,
		},
		{
			id: 'arab-qiraah',
			label: 'Arab: Qiraah',
			desc: 'Bacaan Arab (RTL)',
			stem: '<p>اقرأ النص الآتي ثم أجب عن الأسئلة!</p><p><br></p><p>[النص هنا]</p><p><br></p><p>ما الموضوع الرئيسي للنص ؟</p>',
			options: ['', '', '', ''],
			answerKey: 'A',
			isRtl: true,
		},
		{
			id: 'dalil',
			label: 'Dalil & Tafsir',
			desc: 'Ayat/hadis + analisis',
			stem: '<p>Perhatikan dalil berikut!</p><p><br></p><p><em>[Tulis ayat atau hadis di sini]</em></p><p><br></p><p>Kandungan pokok dari dalil di atas adalah...</p>',
			options: ['', '', '', ''],
			answerKey: 'A',
		},
		{
			id: 'sains',
			label: 'Percobaan Sains',
			desc: 'Data eksperimen',
			stem: '<p>Perhatikan hasil percobaan berikut!</p><p><br></p><p>[Tabel / data percobaan]</p><p><br></p><p>Berdasarkan data tersebut, variabel yang paling berpengaruh adalah...</p>',
			options: ['', '', '', ''],
			answerKey: 'A',
		},
	];

	// ── Page state ─────────────────────────────────────────────────────────────
	let loading = $state(true);
	let pageError = $state('');
	let questions = $state<Question[]>([]);
	let subjects = $state<Subject[]>([]);
	let totalItems = $state(0);
	let currentPage = $state(1);

	let search = $state('');
	let filterSubject = $state('');
	let filterWorkflow = $state('');

	// ── Composer state ─────────────────────────────────────────────────────────
	let showComposer = $state(false);
	let editingId = $state<string | null>(null);
	let composerBusy = $state(false);
	let draftStatus = $state('');
	let showTemplates = $state(true);
	let draftTimer: ReturnType<typeof setTimeout> | null = null;
	let lastDraftSig = '';

	// ── Form fields ────────────────────────────────────────────────────────────
	let fSubjectId = $state('');
	let fStem = $state('');
	let fOptions = $state(['', '', '', '']);
	let fAnswerKey = $state<string>('A');
	let fWeight = $state(1);
	let fDifficulty = $state('medium');
	let fIsRtl = $state(false);

	// ── Derived ────────────────────────────────────────────────────────────────
	let pageCount = $derived(Math.max(1, Math.ceil(totalItems / PAGE_SIZE)));

	function getPlainText(html: string): string {
		if (!html) return '';
		if (typeof document === 'undefined') return html.replace(/<[^>]+>/g, '');
		const d = document.createElement('div');
		d.innerHTML = html;
		return (d.textContent ?? '').replace(/\s+/g, ' ').trim();
	}

	let stemText = $derived(getPlainText(fStem));
	let hasImage = $derived(fStem.includes('<img'));

	let readinessChecks = $derived({
		subject: !!fSubjectId,
		stem: stemText.length >= 5 || hasImage,
		optA: fOptions[0].trim().length > 0,
		optB: fOptions[1].trim().length > 0,
		optC: fOptions[2].trim().length > 0,
		optD: fOptions[3].trim().length > 0,
		answerKey: !!fAnswerKey,
		weight: Number.isFinite(fWeight) && fWeight >= 1,
	});

	let passedChecks = $derived(Object.values(readinessChecks).filter(Boolean).length);
	let totalChecks = $derived(Object.keys(readinessChecks).length);
	let readinessScore = $derived(Math.round((passedChecks / totalChecks) * 100));
	let canSave = $derived(readinessScore === 100 && !composerBusy);

	let qualitySignals = $derived.by(() => {
		const filled = fOptions.map((o) => o.trim()).filter(Boolean);
		const unique = new Set(filled);
		const lengths = filled.map((o) => o.length);
		const lengthRange =
			lengths.length >= 2 ? Math.max(...lengths) - Math.min(...lengths) : 0;
		return [
			{
				label: 'Stem cukup jelas',
				status: stemText.length >= 35 ? 'good' : 'warn',
				desc: `${stemText.length} / 35 karakter minimum`,
			},
			{
				label: 'Distraktor bervariasi',
				status: unique.size === filled.length ? 'good' : 'warn',
				desc: unique.size < filled.length ? 'Ada opsi yang duplikat' : 'Semua opsi unik',
			},
			{
				label: 'Panjang opsi seimbang',
				status: lengthRange <= 80 ? 'good' : 'warn',
				desc: `Selisih panjang: ${lengthRange} karakter`,
			},
			{
				label: 'Media pendukung',
				status: hasImage || stemText.length >= 70 ? 'good' : 'warn',
				desc: hasImage ? 'Ada gambar' : `${stemText.length} / 70 karakter`,
			},
		];
	});

	let validationIssues = $derived.by(() => {
		const issues: string[] = [];
		if (!readinessChecks.subject) issues.push('Pilih mata pelajaran');
		if (!readinessChecks.stem) issues.push('Isi soal minimal 5 karakter');
		if (!readinessChecks.optA || !readinessChecks.optB || !readinessChecks.optC || !readinessChecks.optD)
			issues.push('Semua opsi (A–D) wajib diisi');
		if (!readinessChecks.answerKey) issues.push('Pilih kunci jawaban');
		if (!readinessChecks.weight) issues.push('Bobot nilai minimal 1');
		return issues;
	});

	let previewStem = $derived(renderRichMathHtml(fStem));
	let previewOptions = $derived(fOptions.map((o) => (o ? renderRichMathHtml(o) : '')));

	// ── Draft autosave ─────────────────────────────────────────────────────────
	$effect(() => {
		if (!showComposer) return;
		const sig = JSON.stringify({
			fSubjectId,
			fStem,
			opts: fOptions.map((o) => o),
			fAnswerKey,
			fWeight,
			fDifficulty,
			fIsRtl,
		});
		if (draftTimer) clearTimeout(draftTimer);
		draftTimer = setTimeout(() => {
			if (sig === lastDraftSig) return;
			lastDraftSig = sig;
			try {
				localStorage.setItem(
					DRAFT_KEY(editingId),
					JSON.stringify({
						subjectId: fSubjectId,
						stem: fStem,
						options: [...fOptions],
						answerKey: fAnswerKey,
						weight: fWeight,
						difficulty: fDifficulty,
						isRtl: fIsRtl,
						savedAt: new Date().toISOString(),
					})
				);
				draftStatus = 'Draft tersimpan otomatis';
			} catch {
				/* ignore storage errors */
			}
		}, 700);
		return () => {
			if (draftTimer) clearTimeout(draftTimer);
		};
	});

	function restoreDraft(): boolean {
		try {
			const raw = localStorage.getItem(DRAFT_KEY(editingId));
			if (!raw) return false;
			const d = JSON.parse(raw) as {
				subjectId?: string;
				stem?: string;
				options?: string[];
				answerKey?: string;
				weight?: number;
				difficulty?: string;
				isRtl?: boolean;
			};
			fSubjectId = d.subjectId ?? '';
			fStem = d.stem ?? '';
			fOptions = d.options ?? ['', '', '', ''];
			while (fOptions.length < 4) fOptions = [...fOptions, ''];
			fAnswerKey = d.answerKey ?? 'A';
			fWeight = d.weight ?? 1;
			fDifficulty = d.difficulty ?? 'medium';
			fIsRtl = d.isRtl ?? false;
			draftStatus = 'Draft lokal dipulihkan';
			return true;
		} catch {
			return false;
		}
	}

	function clearDraft() {
		try {
			localStorage.removeItem(DRAFT_KEY(editingId));
		} catch {
			/* ignore */
		}
		lastDraftSig = '';
		draftStatus = '';
	}

	// ── API ────────────────────────────────────────────────────────────────────
	async function load(page = currentPage) {
		loading = true;
		pageError = '';
		try {
			const params = new URLSearchParams();
			params.set('limit', String(PAGE_SIZE));
			params.set('offset', String((page - 1) * PAGE_SIZE));
			if (search.trim()) params.set('q', search.trim());
			if (filterSubject) params.set('subject_id', filterSubject);
			if (filterWorkflow) params.set('workflow_status', filterWorkflow);

			const [qRes, aRes] = await Promise.all([
				fetch(`/api/cbt/questions?${params.toString()}`),
				fetch('/api/academic'),
			]);
			const qJson = (await qRes.json().catch(() => ({}))) as QuestionListResponse & {
				error?: string;
			};
			const aJson = await aRes.json().catch(() => ({}));

			if (!qRes.ok) {
				pageError = qJson.error ?? 'Gagal memuat soal';
				return;
			}
			questions = qJson.items ?? [];
			totalItems = qJson.meta?.total ?? questions.length;
			currentPage = page;
			subjects = ((aJson.data ?? aJson)?.subjects ?? []) as Subject[];
		} catch (e) {
			pageError = (e as Error).message || 'Gagal memuat soal';
		} finally {
			loading = false;
		}
	}

	function resetForm() {
		fSubjectId = '';
		fStem = '';
		fOptions = ['', '', '', ''];
		fAnswerKey = 'A';
		fWeight = 1;
		fDifficulty = 'medium';
		fIsRtl = false;
		draftStatus = '';
		lastDraftSig = '';
	}

	function openCreate() {
		editingId = null;
		resetForm();
		showTemplates = true;
		showComposer = true;
		// Delay to let state settle before restoring
		setTimeout(() => {
			if (!restoreDraft()) draftStatus = '';
		}, 50);
	}

	async function openEdit(q: Question) {
		composerBusy = true;
		try {
			const res = await fetch(`/api/cbt/questions/${q.id}`);
			const detail = (await res.json().catch(() => q)) as Question;
			const d = res.ok ? detail : q;

			editingId = d.id;
			resetForm();
			fSubjectId = d.subject_id ?? '';
			fStem = d.stem_html || d.question_text || '';
			const opts = d.options?.length
				? d.options.map((o) => o.text || o.html || o.latex || '')
				: ['', '', '', ''];
			while (opts.length < 4) opts.push('');
			fOptions = opts;
			fAnswerKey = d.answer_key || 'A';
			fDifficulty = d.difficulty || 'medium';
			showTemplates = false;
			showComposer = true;
		} catch {
			editingId = q.id;
			resetForm();
			fSubjectId = q.subject_id;
			fStem = q.stem_html || q.question_text || '';
			fAnswerKey = q.answer_key || 'A';
			showComposer = true;
		} finally {
			composerBusy = false;
		}
	}

	function applyTemplate(t: Template) {
		fStem = t.stem;
		fOptions = [...t.options];
		while (fOptions.length < 4) fOptions = [...fOptions, ''];
		fAnswerKey = t.answerKey;
		fIsRtl = t.isRtl ?? false;
		draftStatus = `Template "${t.label}" diterapkan`;
		showTemplates = false;
	}

	function closeComposer() {
		showComposer = false;
		editingId = null;
	}

	async function saveQuestion() {
		if (!canSave) return;
		composerBusy = true;
		try {
			const payload = {
				authoring_mode: 'beginner',
				subject_id: fSubjectId,
				question_text: getPlainText(fStem),
				question_type: 'multiple_choice',
				stem_html: fStem,
				options: fOptions.map((text, i) => ({ label: ANSWER_LABELS[i], text })),
				answer_key: fAnswerKey,
				difficulty: fDifficulty,
				status: 'draft',
				workflow_status: 'draft',
			};

			const url = editingId ? `/api/cbt/questions/${editingId}` : '/api/cbt/questions';
			const method = editingId ? 'PUT' : 'POST';
			const res = await fetch(url, {
				method,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(payload),
			});
			const data = (await res.json().catch(() => ({}))) as { error?: string };

			if (!res.ok) {
				toast.error(data.error ?? 'Gagal menyimpan soal');
				return;
			}

			clearDraft();
			toast.success(editingId ? 'Soal berhasil diperbarui' : 'Soal berhasil dibuat');
			closeComposer();
			await load(1);
		} catch (e) {
			toast.error((e as Error).message || 'Gagal menyimpan soal');
		} finally {
			composerBusy = false;
		}
	}

	async function deleteQuestion(id: string) {
		if (!confirm('Hapus soal ini? Tindakan tidak bisa dibatalkan.')) return;
		try {
			const res = await fetch(`/api/cbt/questions?id=${id}`, { method: 'DELETE' });
			if (!res.ok) {
				const d = (await res.json().catch(() => ({}))) as { error?: string };
				toast.error(d.error ?? 'Gagal menghapus soal');
				return;
			}
			toast.success('Soal dihapus');
			await load(currentPage);
		} catch (e) {
			toast.error((e as Error).message || 'Gagal menghapus soal');
		}
	}

	async function uploadImageInEditor(file: File): Promise<string> {
		const form = new FormData();
		form.set('file', file);
		form.set('purpose', 'general');
		const res = await fetch('/api/cbt/assets', { method: 'POST', body: form });
		const payload = (await res.json().catch(() => ({}))) as { url?: string; error?: string };
		if (!res.ok) throw new Error(payload.error ?? 'Upload gambar gagal');
		return payload.url ?? '';
	}

	let searchTimer: ReturnType<typeof setTimeout> | null = null;
	function onSearchInput(e: Event) {
		search = (e.target as HTMLInputElement).value;
		if (searchTimer) clearTimeout(searchTimer);
		searchTimer = setTimeout(() => load(1), 400);
	}

	function stemPreview(q: Question): string {
		const t = getPlainText(q.stem_html || q.question_text || '');
		return t.length > 90 ? t.slice(0, 90) + '…' : t || '(kosong)';
	}

	function workflowClass(status: string): string {
		const map: Record<string, string> = {
			draft: 'bg-slate-100 text-slate-600',
			review: 'bg-yellow-100 text-yellow-700',
			approved: 'bg-green-100 text-green-700',
			rejected: 'bg-red-100 text-red-600',
		};
		return map[status] ?? 'bg-slate-100 text-slate-500';
	}

	onMount(() => load());
</script>

<!-- ── Main page ──────────────────────────────────────────────────────────── -->
<div class="pt-14 lg:pt-0 p-4 lg:p-6 space-y-4">
	<!-- Header -->
	<div class="flex flex-wrap items-start justify-between gap-3">
		<div>
			<h1 class="text-xl font-semibold text-slate-800">Komposer Soal</h1>
			<p class="text-sm text-slate-500 mt-0.5">
				Buat butir soal dengan template, preview langsung, dan simpan draft otomatis
			</p>
		</div>
		<Button onclick={openCreate} class="bg-green-700 hover:bg-green-800 text-white shrink-0">
			+ Buat Soal
		</Button>
	</div>

	<!-- Filter bar -->
	<div class="flex flex-wrap gap-2">
		<Input
			placeholder="Cari soal..."
			value={search}
			oninput={onSearchInput}
			class="w-48 text-sm h-8"
		/>
		<select
			bind:value={filterSubject}
			onchange={() => load(1)}
			class="h-8 rounded-md border border-slate-200 bg-white px-2.5 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-green-500"
		>
			<option value="">Semua Mapel</option>
			{#each subjects as s (s.id)}
				<option value={s.id}>{s.name}</option>
			{/each}
		</select>
		<select
			bind:value={filterWorkflow}
			onchange={() => load(1)}
			class="h-8 rounded-md border border-slate-200 bg-white px-2.5 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-green-500"
		>
			<option value="">Semua Status</option>
			<option value="draft">Draft</option>
			<option value="review">Ditinjau</option>
			<option value="approved">Disetujui</option>
			<option value="rejected">Revisi</option>
		</select>
		{#if totalItems > 0}
			<span class="flex items-center text-xs text-slate-400">{totalItems} soal</span>
		{/if}
	</div>

	<!-- Table -->
	<div class="rounded-lg border border-slate-200 bg-white overflow-hidden">
		<Table.Root>
			<Table.Header>
				<Table.Row class="bg-slate-50 text-xs">
					<Table.Head class="w-10 text-slate-500">#</Table.Head>
					<Table.Head class="text-slate-500">Isi Soal</Table.Head>
					<Table.Head class="w-32 text-slate-500">Mapel</Table.Head>
					<Table.Head class="w-28 hidden sm:table-cell text-slate-500">Status</Table.Head>
					<Table.Head class="w-16 text-slate-500">Kunci</Table.Head>
					<Table.Head class="w-24 text-right text-slate-500">Aksi</Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#if loading}
					{#each Array.from({ length: 6 }) as _, index (`composer-question-skeleton-${index}`)}
						<Table.Row>
							<Table.Cell><Skeleton class="h-4 w-4" /></Table.Cell>
							<Table.Cell><Skeleton class="h-5 w-full max-w-sm" /></Table.Cell>
							<Table.Cell><Skeleton class="h-5 w-24" /></Table.Cell>
							<Table.Cell class="hidden sm:table-cell"><Skeleton class="h-6 w-20" /></Table.Cell>
							<Table.Cell><Skeleton class="h-5 w-8" /></Table.Cell>
							<Table.Cell class="text-right"><Skeleton class="ml-auto h-8 w-16" /></Table.Cell>
						</Table.Row>
					{/each}
				{:else if pageError}
					<Table.Row>
						<Table.Cell colspan={6} class="py-10 text-center text-sm text-red-500">
							{pageError}
						</Table.Cell>
					</Table.Row>
				{:else if questions.length === 0}
					<Table.Row>
						<Table.Cell colspan={6} class="py-10 text-center text-sm text-slate-400">
							Belum ada soal.
							<button onclick={openCreate} class="text-green-700 underline ml-1"
								>Buat soal pertama →</button
							>
						</Table.Cell>
					</Table.Row>
				{:else}
					{#each questions as q, i (q.id)}
						<Table.Row
							class="hover:bg-slate-50 cursor-pointer"
							onclick={() => openEdit(q)}
						>
							<Table.Cell class="text-xs text-slate-400">
								{(currentPage - 1) * PAGE_SIZE + i + 1}
							</Table.Cell>
							<Table.Cell class="text-sm text-slate-700 max-w-xs">
								<div class="truncate">{stemPreview(q)}</div>
								{#if q.author_username}
									<div class="text-[10px] text-slate-400 mt-0.5">{q.author_username}</div>
								{/if}
							</Table.Cell>
							<Table.Cell class="text-xs text-slate-500 truncate max-w-[8rem]">
								{q.subject_name || q.subject_code || '-'}
							</Table.Cell>
							<Table.Cell class="hidden sm:table-cell">
								<span
									class="inline-block rounded px-1.5 py-0.5 text-[10px] font-medium {workflowClass(q.workflow_status)}"
								>
									{WORKFLOW_LABEL[q.workflow_status] ?? q.workflow_status ?? '-'}
								</span>
							</Table.Cell>
							<Table.Cell class="text-sm font-bold text-green-700">
								{q.answer_key || '-'}
							</Table.Cell>
							<Table.Cell class="text-right">
								<button
									onclick={(e) => {
										e.stopPropagation();
										deleteQuestion(q.id);
									}}
									class="text-xs text-red-400 hover:text-red-600 px-2 py-1 rounded hover:bg-red-50 transition-colors"
								>
									Hapus
								</button>
							</Table.Cell>
						</Table.Row>
					{/each}
				{/if}
			</Table.Body>
		</Table.Root>
	</div>

	<!-- Pagination -->
	{#if pageCount > 1}
		<div class="flex items-center justify-between text-sm text-slate-500">
			<span class="text-xs">{totalItems} soal total</span>
			<div class="flex items-center gap-1">
				<Button
					variant="outline"
					class="h-7 w-7 p-0 text-xs"
					disabled={currentPage <= 1}
					onclick={() => load(currentPage - 1)}
				>
					‹
				</Button>
				<span class="px-2 text-xs">Hal {currentPage} / {pageCount}</span>
				<Button
					variant="outline"
					class="h-7 w-7 p-0 text-xs"
					disabled={currentPage >= pageCount}
					onclick={() => load(currentPage + 1)}
				>
					›
				</Button>
			</div>
		</div>
	{/if}
</div>

<!-- ── Composer Dialog ─────────────────────────────────────────────────────── -->
<Dialog.Root bind:open={showComposer}>
	<Dialog.Content>
		<div class="sm:max-w-[min(94vw,88rem)] max-h-[93vh] h-[93vh] flex flex-col gap-0 p-0 overflow-hidden">
		<!-- Dialog header -->
		<div class="flex shrink-0 items-center justify-between border-b border-slate-100 px-5 py-3">
			<div>
				<h2 class="text-base font-semibold text-slate-800">
					{editingId ? 'Edit Butir Soal' : 'Buat Butir Soal Baru'}
				</h2>
				{#if draftStatus}
					<p class="text-[10px] text-slate-400 mt-0.5">{draftStatus}</p>
				{/if}
			</div>
			<button
				onclick={closeComposer}
				class="rounded-md p-1.5 text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-colors"
				aria-label="Tutup"
			>
				<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M6 18L18 6M6 6l12 12"
					/>
				</svg>
			</button>
		</div>

		<!-- Dialog body: split layout -->
		<div class="flex flex-1 overflow-hidden min-h-0">
			<!-- Left: form -->
			<div class="flex-1 overflow-y-auto p-5 space-y-5 min-w-0">
				<!-- Template strip -->
				<div>
					<div class="flex items-center justify-between mb-1.5">
						<span class="text-[10px] font-semibold uppercase tracking-wider text-slate-400">
							Template Cepat
						</span>
						<button
							type="button"
							onclick={() => (showTemplates = !showTemplates)}
							class="text-[10px] text-slate-400 hover:text-slate-600"
						>
							{showTemplates ? 'Sembunyikan' : 'Tampilkan'}
						</button>
					</div>
					{#if showTemplates}
						<div class="flex gap-2 overflow-x-auto pb-2">
							{#each templates as t (t.id)}
								<button
									type="button"
									onclick={() => applyTemplate(t)}
									class="flex-shrink-0 rounded-lg border border-slate-200 bg-white px-3 py-2 text-left hover:border-green-400 hover:bg-green-50 transition-colors min-w-[110px] max-w-[130px]"
								>
									<div class="text-xs font-medium text-slate-700 leading-tight">{t.label}</div>
									<div class="text-[10px] text-slate-400 mt-0.5">{t.desc}</div>
									{#if t.isRtl}
										<span
											class="mt-1 inline-block rounded bg-amber-100 px-1 text-[9px] font-medium text-amber-700"
											>RTL</span
										>
									{/if}
								</button>
							{/each}
						</div>
					{/if}
				</div>

				<!-- Metadata row -->
				<div class="flex flex-wrap gap-3">
					<div class="flex-1 min-w-[160px]">
						<label
							for="f-subject"
							class="mb-1 block text-xs font-medium text-slate-600"
						>
							Mata Pelajaran <span class="text-red-500">*</span>
						</label>
						<select
							id="f-subject"
							bind:value={fSubjectId}
							class="w-full rounded-md border border-slate-200 bg-white px-2.5 py-1.5 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-green-500"
						>
							<option value="">-- Pilih Mapel --</option>
							{#each subjects as s (s.id)}
								<option value={s.id}>{s.name}</option>
							{/each}
						</select>
					</div>
					<div class="w-28">
						<label
							for="f-difficulty"
							class="mb-1 block text-xs font-medium text-slate-600"
						>
							Tingkat Kesulitan
						</label>
						<select
							id="f-difficulty"
							bind:value={fDifficulty}
							class="w-full rounded-md border border-slate-200 bg-white px-2.5 py-1.5 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-green-500"
						>
							{#each Object.entries(DIFFICULTY_LABEL) as [val, lbl]}
								<option value={val}>{lbl}</option>
							{/each}
						</select>
					</div>
					<div class="w-20">
						<label for="f-weight" class="mb-1 block text-xs font-medium text-slate-600">
							Bobot
						</label>
						<Input
							id="f-weight"
							type="number"
							min="1"
							bind:value={fWeight}
							class="h-8 text-sm"
						/>
					</div>
					<div class="flex items-end pb-1">
						<label class="flex cursor-pointer items-center gap-2">
							<input type="checkbox" bind:checked={fIsRtl} class="rounded" />
							<span class="text-xs text-slate-600">Mode Arab (RTL)</span>
						</label>
					</div>
				</div>

				<!-- Stem editor -->
				<div>
					<div class="mb-1.5 block text-xs font-medium text-slate-600">
						Isi Soal (Stem) <span class="text-red-500">*</span>
					</div>
					<EditorWrapper
						bind:value={fStem}
						id="soal-stem"
						placeholder="Tuliskan pertanyaan utama. Gambar (🖼), LaTeX ($x^2$), dan tabel bisa disisipkan langsung."
						minRows={6}
						onImageUpload={uploadImageInEditor}
					/>
				</div>

				<!-- Options A–D -->
				<div>
					<div class="mb-2 flex items-baseline justify-between">
						<div class="text-xs font-medium text-slate-600">
							Pilihan Jawaban <span class="text-red-500">*</span>
						</div>
						<span class="text-[10px] text-slate-400">
							Klik lingkaran huruf untuk menandai kunci jawaban
						</span>
					</div>
					<div class="space-y-2">
						{#each fOptions as _, i}
							{@const label = ANSWER_LABELS[i]}
							{@const isAnswer = fAnswerKey === label}
							<div class="flex items-center gap-2">
								<button
									type="button"
									onclick={() => (fAnswerKey = label)}
									aria-label="Pilih opsi {label} sebagai kunci jawaban"
									class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full border-2 text-xs font-bold transition-colors
										{isAnswer
										? 'border-green-600 bg-green-600 text-white'
										: 'border-slate-300 bg-white text-slate-500 hover:border-green-400 hover:text-green-600'}"
								>
									{label}
								</button>
								<textarea
									rows="1"
									value={fOptions[i]}
									oninput={(e) =>
										(fOptions[i] = (e.target as HTMLTextAreaElement).value)}
									placeholder="Opsi {label}... (LaTeX: $rumus$)"
									dir={fIsRtl ? 'rtl' : undefined}
									class="flex-1 resize-none rounded-md border px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-green-500 transition-colors
										{isAnswer
										? 'border-green-300 bg-green-50'
										: 'border-slate-200 bg-white'}"
								></textarea>
							</div>
						{/each}
					</div>
				</div>

				<div class="h-4"></div>
			</div>

			<!-- Right: preview panel (desktop only) -->
			<div
				class="hidden w-72 shrink-0 overflow-y-auto border-l border-slate-100 bg-slate-50 p-4 lg:block xl:w-80"
			>
				<!-- Readiness score -->
				<div class="mb-4">
					<div class="mb-1 flex items-center justify-between">
						<span class="text-[10px] font-semibold uppercase tracking-wide text-slate-400"
							>Kesiapan Soal</span
						>
						<span
							class="text-sm font-bold {readinessScore === 100
								? 'text-green-700'
								: readinessScore >= 50
									? 'text-amber-600'
									: 'text-red-500'}"
						>
							{readinessScore}%
						</span>
					</div>
					<div class="h-2 overflow-hidden rounded-full bg-slate-200">
						<div
							class="h-2 rounded-full transition-all duration-300 {readinessScore === 100
								? 'bg-green-600'
								: readinessScore >= 50
									? 'bg-amber-400'
									: 'bg-red-400'}"
							style="width: {readinessScore}%"
						></div>
					</div>
					{#if validationIssues.length > 0}
						<ul class="mt-1.5 space-y-0.5">
							{#each validationIssues as issue}
								<li class="text-[10px] text-red-500">• {issue}</li>
							{/each}
						</ul>
					{:else}
						<p class="mt-1 text-[10px] text-green-600">Soal siap disimpan!</p>
					{/if}
				</div>

				<!-- Quality signals -->
				<div class="mb-4">
					<p class="mb-2 text-[10px] font-semibold uppercase tracking-wide text-slate-400">
						Sinyal Kualitas
					</p>
					<div class="space-y-2">
						{#each qualitySignals as sig}
							<div class="flex items-start gap-2">
								<span
									class="mt-0.5 shrink-0 text-sm font-bold {sig.status === 'good'
										? 'text-green-600'
										: 'text-amber-500'}"
								>
									{sig.status === 'good' ? '✓' : '!'}
								</span>
								<div>
									<div class="text-xs font-medium text-slate-700">{sig.label}</div>
									<div class="text-[10px] text-slate-400">{sig.desc}</div>
								</div>
							</div>
						{/each}
					</div>
				</div>

				<div class="my-3 border-t border-slate-200"></div>

				<!-- Student preview -->
				<div>
					<p class="mb-2 text-[10px] font-semibold uppercase tracking-wide text-slate-400">
						Preview Siswa
					</p>
					<div
						class="rounded-lg border border-slate-200 bg-white p-3 space-y-3"
						dir={fIsRtl ? 'rtl' : undefined}
					>
						{#if fStem}
							<div class="prose prose-sm max-w-none text-slate-800 latex-preview text-sm">
								{@html previewStem}
							</div>
						{:else}
							<p class="text-xs text-slate-400 italic">Isi soal belum dimasukkan</p>
						{/if}

						{#if fOptions.some((o) => o.trim())}
							<div class="space-y-1.5 border-t border-slate-100 pt-2">
								{#each fOptions as opt, i}
									{@const label = ANSWER_LABELS[i]}
									{@const isAnswer = fAnswerKey === label}
									<div
										class="flex items-baseline gap-2 text-sm {isAnswer
											? 'text-green-700 font-medium'
											: 'text-slate-700'}"
									>
										<span class="shrink-0 font-bold">{label}.</span>
										{#if opt.trim()}
											<span class="latex-preview">{@html previewOptions[i]}</span>
										{:else}
											<span class="italic text-slate-300">(kosong)</span>
										{/if}
										{#if isAnswer}
											<span class="ml-auto shrink-0 text-xs text-green-500">✓</span>
										{/if}
									</div>
								{/each}
							</div>
						{/if}
					</div>
				</div>
			</div>
		</div>

		<!-- Dialog footer -->
		<div
			class="flex shrink-0 items-center justify-between gap-3 border-t border-slate-100 bg-white px-5 py-3"
		>
			<div class="text-[10px] text-slate-400 truncate max-w-xs">
				{draftStatus}
			</div>
			<div class="flex gap-2 shrink-0">
				<Button
					variant="outline"
					class="h-8 text-sm"
					onclick={() => {
						clearDraft();
						closeComposer();
					}}
				>
					Batalkan
				</Button>
				<LoadingButton
					onclick={saveQuestion}
					disabled={!canSave}
					loading={composerBusy}
					loadingLabel="Menyimpan..."
					class="h-8 bg-green-700 text-sm text-white hover:bg-green-800 disabled:opacity-50"
				>
					{editingId ? 'Simpan Perubahan' : 'Simpan Soal'}
				</LoadingButton>
			</div>
		</div>
		</div>
	</Dialog.Content>
</Dialog.Root>

<style>
	:global(.latex-preview .latex-display) {
		overflow-x: auto;
		padding: 0.25rem 0;
	}
	:global(.latex-preview .katex-display) {
		margin: 0.5rem 0;
	}
</style>
