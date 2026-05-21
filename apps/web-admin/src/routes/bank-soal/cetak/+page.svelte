<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import PrinterIcon from '@lucide/svelte/icons/printer';
	import RefreshCcwIcon from '@lucide/svelte/icons/refresh-ccw';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import RichContent from '$lib/components/RichContent.svelte';
	import { clientApiPathWithQuery, readClientApiData } from '$lib/client/api';
	import type { AccountIdentity } from '$lib/client/account';
	import type { AuthUser } from '$lib/server/auth';

	type PageData = {
		user?: AuthUser;
		account?: AccountIdentity | null;
	};

	type Subject = {
		id: string;
		name: string;
		code?: string;
	};

	type OptionItem = {
		label?: string;
		text?: string;
		html?: string;
		latex?: string;
		match_label?: string;
		match_text?: string;
		match_html?: string;
		is_distractor?: boolean;
	};

	type Question = {
		id: string;
		subject_id?: string;
		subject_name?: string;
		subject_code?: string;
		code?: string;
		question_text?: string;
		question_type?: string;
		stem_html?: string;
		stem_latex?: string;
		stimulus_html?: string;
		options?: OptionItem[];
		option_a?: string;
		option_b?: string;
		option_c?: string;
		option_d?: string;
		option_e?: string;
		answer_key?: string;
		explanation?: string;
		explanation_html?: string;
		rubric_html?: string;
		difficulty?: string;
		status?: string;
		workflow_status?: string;
		author_username?: string;
		author_display_name?: string;
		academic_phase?: string;
		target_level?: string | null;
		cp_ref?: string;
		tp_ref?: string;
		kd_ref?: string;
		indicator_ref?: string;
		material_topic?: string;
		cognitive_level?: string;
		hots_flag?: boolean;
		created_at?: string;
		updated_at?: string;
	};

	type QuestionListResponse = {
		items?: Question[];
		meta?: { total?: number; limit?: number; offset?: number };
	};

	type AcademicPayload = {
		subjects?: Subject[];
	};

	let { data }: { data: PageData } = $props();

	let loading = $state(true);
	let error = $state('');
	let questions = $state<Question[]>([]);
	let subjects = $state<Subject[]>([]);
	let subjectFilter = $state('');
	let workflowFilter = $state('');
	let typeFilter = $state('');
	let dateFrom = $state('');
	let dateTo = $state('');
	let includeAnswer = $state(false);
	let includeExplanation = $state(false);
	let includeMetadata = $state(true);

	const listHref = resolve('/bank-soal/daftar');
	const printTitle = 'Cetak Soal Saya';

	const workflowOptions = [
		{ value: '', label: 'Semua status' },
		{ value: 'draft', label: 'Draft' },
		{ value: 'submitted', label: 'Diajukan' },
		{ value: 'review', label: 'Review' },
		{ value: 'revision_needed', label: 'Perlu Revisi' },
		{ value: 'reviewed', label: 'Sudah Direview' },
		{ value: 'approved', label: 'Disetujui' },
		{ value: 'published', label: 'Terbit' },
		{ value: 'rejected', label: 'Ditolak' },
		{ value: 'archived', label: 'Arsip' }
	];

	const questionTypeOptions = [
		{ value: '', label: 'Semua tipe' },
		{ value: 'multiple_choice', label: 'Pilihan Ganda' },
		{ value: 'multiple_answer', label: 'Pilihan Kompleks' },
		{ value: 'true_false', label: 'Benar/Salah' },
		{ value: 'agree_disagree', label: 'Setuju/Tidak Setuju' },
		{ value: 'matching', label: 'Menjodohkan' },
		{ value: 'short_answer', label: 'Isian Singkat' },
		{ value: 'essay', label: 'Esai' }
	];

	let currentUsername = $derived(data.user?.username ?? data.account?.username ?? '');
	let teacherName = $derived(
		data.account?.profile_nama || data.account?.display_name || data.user?.username || 'Guru'
	);
	let filteredQuestions = $derived(
		questions.filter((question) => {
			if (subjectFilter && question.subject_id !== subjectFilter) return false;
			if (workflowFilter && (question.workflow_status ?? question.status ?? '') !== workflowFilter) return false;
			if (typeFilter && question.question_type !== typeFilter) return false;
			if (dateFrom && normalizeDate(question.created_at) < dateFrom) return false;
			if (dateTo && normalizeDate(question.created_at) > dateTo) return false;
			return true;
		})
	);
	let selectedSubjectName = $derived(
		subjectFilter ? subjects.find((subject) => subject.id === subjectFilter)?.name || 'Mapel dipilih' : 'Semua mapel'
	);

	onMount(() => {
		void loadPrintData();
	});

	async function loadPrintData() {
		if (!currentUsername) {
			error = 'Akun belum terbaca. Silakan login ulang sebelum mencetak soal.';
			loading = false;
			return;
		}
		loading = true;
		error = '';
		try {
			const params = new URLSearchParams({
				limit: '2000',
				offset: '0',
				author_username: currentUsername,
				sort: 'newest'
			});
			const [questionPayload, subjectPayload] = await Promise.all([
				fetch(clientApiPathWithQuery('/api/bank-soal/questions', params)).then((response) =>
					readClientApiData<QuestionListResponse>(response, 'Gagal memuat soal untuk dicetak')
				),
				fetch('/api/bank-soal/soal-support/subjects').then((response) =>
					readClientApiData<AcademicPayload>(response, 'Gagal memuat daftar mapel')
				)
			]);
			questions = questionPayload.items ?? [];
			subjects = subjectPayload.subjects ?? [];
		} catch (e) {
			error = e instanceof Error ? e.message : 'Data cetak belum dapat dimuat';
		} finally {
			loading = false;
		}
	}

	function normalizeDate(value: string | undefined): string {
		if (!value) return '';
		return value.slice(0, 10);
	}

	function formatDate(value: string | undefined): string {
		if (!value) return '-';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value;
		return new Intl.DateTimeFormat('id-ID', {
			day: '2-digit',
			month: 'short',
			year: 'numeric'
		}).format(date);
	}

	function formatDateTime(value: Date = new Date()): string {
		return new Intl.DateTimeFormat('id-ID', {
			day: '2-digit',
			month: 'long',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		}).format(value);
	}

	function compactText(value: string | undefined | null, fallback = '-'): string {
		const trimmed = (value ?? '').trim();
		return trimmed || fallback;
	}

	function escapeHtml(value: string): string {
		return value
			.replace(/&/g, '&amp;')
			.replace(/</g, '&lt;')
			.replace(/>/g, '&gt;')
			.replace(/"/g, '&quot;')
			.replace(/'/g, '&#39;')
			.replace(/\n/g, '<br>');
	}

	function richQuestionHtml(question: Question): string {
		return compactText(
			question.stem_html || question.stem_latex || question.question_text,
			'<em>Isi soal belum tersedia</em>'
		);
	}

	function richText(value: string | undefined | null): string {
		const trimmed = (value ?? '').trim();
		if (!trimmed) return '';
		if (/[<][a-zA-Z/][\s\S]*[>]/.test(trimmed) || trimmed.includes('$')) return trimmed;
		return `<p>${escapeHtml(trimmed)}</p>`;
	}

	function questionTypeLabel(value: string | undefined): string {
		return questionTypeOptions.find((option) => option.value === value)?.label ?? compactText(value, 'Tipe belum dipilih');
	}

	function workflowLabel(value: string | undefined): string {
		return workflowOptions.find((option) => option.value === value)?.label ?? compactText(value, 'Belum ada status');
	}

	function difficultyLabel(value: string | undefined): string {
		if (value === 'easy') return 'Mudah';
		if (value === 'medium') return 'Sedang';
		if (value === 'hard') return 'Sulit';
		return compactText(value, 'Kesulitan belum diisi');
	}

	function subjectLabel(question: Question): string {
		const code = compactText(question.subject_code, '');
		const name = compactText(question.subject_name, '');
		if (code && name) return `${code} - ${name}`;
		return name || code || 'Mapel belum tersedia';
	}

	function levelLabel(question: Question): string {
		return compactText(question.target_level || question.academic_phase, 'Tingkat belum diisi');
	}

	function optionRows(question: Question) {
		const rows = Array.isArray(question.options)
			? question.options
					.map((option, index) => ({
						label: compactText(option.label || String.fromCharCode(65 + index), String.fromCharCode(65 + index)),
						html: compactText(option.html || option.latex || option.text, '')
					}))
					.filter((option) => option.html)
			: [];
		if (rows.length > 0) return rows;
		return [
			['A', question.option_a],
			['B', question.option_b],
			['C', question.option_c],
			['D', question.option_d],
			['E', question.option_e]
		]
			.map(([label, html]) => ({ label: String(label), html: compactText(html, '') }))
			.filter((option) => option.html);
	}

	function metadataRows(question: Question) {
		return [
			['Materi', question.material_topic],
			['Level Kognitif', question.cognitive_level],
			['Kesulitan', difficultyLabel(question.difficulty)],
			['HOTS', question.hots_flag ? 'Ya' : 'Tidak'],
			['CP', question.cp_ref],
			['TP', question.tp_ref],
			['KD', question.kd_ref],
			['Indikator', question.indicator_ref]
		].filter(([, value]) => compactText(value as string | undefined, '') !== '');
	}

	function printPage() {
		window.print();
	}
</script>

<svelte:head>
	<title>Cetak Soal Saya - Bank Soal - MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="print-shell space-y-5">
	<section class="no-print rounded-xl border border-border bg-card p-4 shadow-sm md:p-5">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
			<div>
				<p class="text-[11px] font-semibold uppercase tracking-[0.18em] text-primary">Bank Soal</p>
				<h1 class="mt-1 text-2xl font-semibold tracking-tight text-foreground md:text-3xl">Cetak Soal Saya</h1>
				<p class="mt-2 max-w-3xl text-sm leading-6 text-muted-foreground">
					Cetak soal yang Bapak/Ibu input sendiri. Kunci jawaban, pembahasan, dan metadata bisa ditampilkan sesuai kebutuhan.
				</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<Button href={listHref} variant="outline"><ArrowLeftIcon class="size-4" />Kembali</Button>
				<Button variant="outline" onclick={() => void loadPrintData()} disabled={loading}>
					<RefreshCcwIcon class="size-4" />Muat Ulang
				</Button>
				<Button onclick={printPage} disabled={loading || filteredQuestions.length === 0}>
					<PrinterIcon class="size-4" />Print / Simpan PDF
				</Button>
			</div>
		</div>

		<div class="mt-4 grid gap-3 border-t border-border/60 pt-4 sm:grid-cols-2 lg:grid-cols-4">
			<label class="space-y-1.5 text-sm">
				<span class="text-xs font-semibold text-muted-foreground">Mapel</span>
				<select bind:value={subjectFilter} class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm">
					<option value="">Semua mapel</option>
					{#each subjects as subject (subject.id)}
						<option value={subject.id}>{subject.code ? `${subject.code} - ${subject.name}` : subject.name}</option>
					{/each}
				</select>
			</label>
			<label class="space-y-1.5 text-sm">
				<span class="text-xs font-semibold text-muted-foreground">Status</span>
				<select bind:value={workflowFilter} class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm">
					{#each workflowOptions as option (option.value)}
						<option value={option.value}>{option.label}</option>
					{/each}
				</select>
			</label>
			<label class="space-y-1.5 text-sm">
				<span class="text-xs font-semibold text-muted-foreground">Tipe</span>
				<select bind:value={typeFilter} class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm">
					{#each questionTypeOptions as option (option.value)}
						<option value={option.value}>{option.label}</option>
					{/each}
				</select>
			</label>
			<div class="grid grid-cols-2 gap-2">
				<label class="space-y-1.5 text-sm">
					<span class="text-xs font-semibold text-muted-foreground">Dari</span>
					<input type="date" bind:value={dateFrom} class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm" />
				</label>
				<label class="space-y-1.5 text-sm">
					<span class="text-xs font-semibold text-muted-foreground">Sampai</span>
					<input type="date" bind:value={dateTo} class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm" />
				</label>
			</div>
		</div>

		<div class="mt-4 flex flex-wrap gap-3 border-t border-border/60 pt-4 text-sm">
			<label class="inline-flex items-center gap-2 rounded-md border border-border px-3 py-2">
				<input type="checkbox" bind:checked={includeMetadata} />
				<span>Tampilkan metadata</span>
			</label>
			<label class="inline-flex items-center gap-2 rounded-md border border-border px-3 py-2">
				<input type="checkbox" bind:checked={includeAnswer} />
				<span>Tampilkan kunci jawaban</span>
			</label>
			<label class="inline-flex items-center gap-2 rounded-md border border-border px-3 py-2">
				<input type="checkbox" bind:checked={includeExplanation} />
				<span>Tampilkan pembahasan/rubrik</span>
			</label>
		</div>
	</section>

	{#if error}
		<div class="no-print rounded-xl border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive">{error}</div>
	{:else if loading}
		<div class="no-print rounded-xl border border-border bg-card p-6 text-sm text-muted-foreground">Memuat soal untuk cetak...</div>
	{:else}
		<section class="print-document rounded-xl border border-border bg-white p-6 text-slate-950 shadow-sm">
			<header class="border-b border-slate-300 pb-4 text-center">
				<p class="text-sm font-semibold uppercase tracking-[0.18em]">MTsN 2 Kolaka Utara</p>
				<h2 class="mt-1 text-xl font-bold">{printTitle}</h2>
				<p class="mt-1 text-sm text-slate-600">{teacherName} · {selectedSubjectName} · Dicetak {formatDateTime()}</p>
			</header>

			<div class="mt-4 grid gap-2 text-sm sm:grid-cols-2 lg:grid-cols-4">
				<div><span class="font-semibold">Jumlah soal:</span> {filteredQuestions.length}</div>
				<div><span class="font-semibold">Status:</span> {workflowOptions.find((option) => option.value === workflowFilter)?.label ?? 'Semua status'}</div>
				<div><span class="font-semibold">Tipe:</span> {questionTypeOptions.find((option) => option.value === typeFilter)?.label ?? 'Semua tipe'}</div>
				<div><span class="font-semibold">Kunci:</span> {includeAnswer ? 'Ditampilkan' : 'Tidak ditampilkan'}</div>
			</div>

			{#if filteredQuestions.length === 0}
				<div class="mt-8 rounded-lg border border-dashed border-slate-300 p-6 text-center text-sm text-slate-600">
					Belum ada soal sesuai filter cetak.
				</div>
			{:else}
				<ol class="mt-6 space-y-6">
					{#each filteredQuestions as question, index (question.id)}
						<li class="question-item break-inside-avoid rounded-lg border border-slate-300 p-4">
							<div class="flex flex-wrap items-center gap-2 text-xs text-slate-600">
								<Badge variant="outline" class="border-slate-300 text-slate-700">{compactText(question.code, `Soal ${index + 1}`)}</Badge>
								<span>{subjectLabel(question)}</span>
								<span>·</span>
								<span>{levelLabel(question)}</span>
								<span>·</span>
								<span>{questionTypeLabel(question.question_type)}</span>
								<span>·</span>
								<span>{workflowLabel(question.workflow_status ?? question.status)}</span>
								<span>·</span>
								<span>Dibuat {formatDate(question.created_at)}</span>
							</div>

							<div class="mt-3 flex gap-3">
								<div class="shrink-0 font-bold">{index + 1}.</div>
								<div class="min-w-0 flex-1">
									{#if question.stimulus_html}
										<div class="mb-3 rounded-md border border-slate-200 bg-slate-50 p-3 text-sm">
											<RichContent html={question.stimulus_html} />
										</div>
									{/if}
									<RichContent html={richQuestionHtml(question)} class="text-[15px] leading-7" />

									{#if optionRows(question).length > 0}
										<ol class="mt-3 space-y-2">
											{#each optionRows(question) as option (option.label)}
												<li class="flex gap-2 text-sm">
													<span class="font-semibold">{option.label}.</span>
													<RichContent html={richText(option.html)} class="min-w-0 flex-1" />
												</li>
											{/each}
										</ol>
									{:else if question.question_type === 'essay' || question.question_type === 'short_answer'}
										<div class="mt-4 space-y-2 text-sm text-slate-500">
											<div class="border-b border-dotted border-slate-400 pb-5">Jawaban:</div>
											<div class="border-b border-dotted border-slate-400 pb-5"></div>
										</div>
									{/if}

									{#if includeAnswer && question.answer_key}
										<div class="mt-3 rounded-md border border-emerald-300 bg-emerald-50 p-2 text-sm"><span class="font-semibold">Kunci:</span> {question.answer_key}</div>
									{/if}

									{#if includeExplanation && (question.explanation_html || question.explanation || question.rubric_html)}
										<div class="mt-3 rounded-md border border-slate-200 bg-slate-50 p-3 text-sm">
											<p class="mb-1 font-semibold">Pembahasan / Rubrik</p>
											{#if question.explanation_html || question.explanation}
												<RichContent html={richText(question.explanation_html || question.explanation)} />
											{/if}
											{#if question.rubric_html}
												<RichContent html={question.rubric_html} />
											{/if}
										</div>
									{/if}

									{#if includeMetadata && metadataRows(question).length > 0}
										<dl class="mt-3 grid gap-x-4 gap-y-1 rounded-md border border-slate-200 bg-slate-50 p-3 text-xs sm:grid-cols-[9rem_1fr]">
											{#each metadataRows(question) as row (`${question.id}-${row[0]}`)}
												<dt class="font-semibold text-slate-600">{row[0]}</dt>
												<dd>{row[1]}</dd>
											{/each}
										</dl>
									{/if}
								</div>
							</div>
						</li>
					{/each}
				</ol>
			{/if}
		</section>
	{/if}
</div>

<style>
	@media print {
		:global(body) {
			background: #fff !important;
		}

		.no-print,
		:global(header),
		:global(nav),
		:global(aside) {
			display: none !important;
		}

		.print-shell {
			margin: 0 !important;
			padding: 0 !important;
		}

		.print-document {
			border: 0 !important;
			box-shadow: none !important;
			padding: 0 !important;
		}

		.question-item {
			page-break-inside: avoid;
			break-inside: avoid;
		}
	}
</style>
