<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { clientApiPathWithQuery, readClientApiData } from '$lib/client/api';
	import { canCreateBankSoal, type BankSoalAccessUser } from '$lib/bank-soal/access';

	type Subject = { id: string; name: string; code?: string };
	type Question = {
		id: string;
		subject_id?: string;
		subject_name?: string;
		subject_code?: string;
		academic_phase?: string;
		target_level?: string | null;
		grade_level?: number | null;
		cp_ref?: string;
		tp_ref?: string;
		kd_ref?: string;
		indicator_ref?: string;
		material_topic?: string;
		cognitive_level?: string;
		workflow_status?: string;
		status?: string;
	};
	type SubjectPayload = { subjects?: Subject[] };
	type QuestionListResponse = { items?: Question[]; meta?: { total?: number } };
	type Payload = { subjects: Subject[]; questions: Question[] };
	type PageData = { user?: BankSoalAccessUser };
	type CoverageRow = { key: string; subject: string; total: number; approved: number; kdCount: number; topicCount: number; missingKd: number; levels: string[]; grades: string[] };

	let { data }: { data: PageData } = $props();

	let promise = $state<Promise<Payload> | null>(null);
	let subjects = $state<Subject[]>([]);
	let questions = $state<Question[]>([]);
	let search = $state('');

	function subjectName(question: Question) {
		return question.subject_name || question.subject_code || (question.subject_id ? subjects.find((item) => item.id === question.subject_id)?.name : '') || 'Tanpa Mapel';
	}

	async function fetchPayload(): Promise<Payload> {
		const params = new URLSearchParams({ limit: '300', offset: '0' });
		const [subjectPayload, questionPayload] = await Promise.all([
			fetch('/api/bank-soal/soal-support/subjects').then((response) => readClientApiData<SubjectPayload>(response, 'Gagal memuat mapel')),
			fetch(clientApiPathWithQuery('/api/bank-soal/questions', params)).then((response) => readClientApiData<QuestionListResponse>(response, 'Gagal memuat sebaran soal'))
		]);
		return { subjects: subjectPayload.subjects ?? [], questions: questionPayload.items ?? [] };
	}

	function load() {
		promise = fetchPayload().then((payload) => {
			subjects = payload.subjects;
			questions = payload.questions;
			return payload;
		});
	}

	function rowFor(subject: string, group: Question[]): CoverageRow {
		const kd = new Set(group.map((item) => item.kd_ref?.trim()).filter(Boolean));
		const topics = new Set(group.map((item) => item.material_topic?.trim()).filter(Boolean));
		const levels = Array.from(new Set(group.map((item) => item.cognitive_level?.trim()).filter(Boolean))) as string[];
		const grades = Array.from(new Set(group.map((item) => item.target_level ? `Tingkat ${item.target_level}` : item.grade_level ? `Kelas ${item.grade_level}` : item.academic_phase?.trim()).filter(Boolean))) as string[];
		return {
			key: subject,
			subject,
			total: group.length,
			approved: group.filter((item) => ['approved', 'published'].includes(item.workflow_status ?? item.status ?? '')).length,
			kdCount: kd.size,
			topicCount: topics.size,
			missingKd: group.filter((item) => !item.kd_ref?.trim() && !item.cp_ref?.trim() && !item.tp_ref?.trim()).length,
			levels,
			grades
		};
	}

	let coverageRows = $derived.by(() => {
		const map = new Map<string, Question[]>();
		for (const question of questions) {
			const key = subjectName(question);
			if (!map.has(key)) map.set(key, []);
			map.get(key)?.push(question);
		}
		for (const subject of subjects) {
			const key = subject.name || subject.code || subject.id;
			if (!map.has(key)) map.set(key, []);
		}
		return Array.from(map.entries()).map(([subject, group]) => rowFor(subject, group)).sort((a, b) => b.total - a.total || a.subject.localeCompare(b.subject, 'id'));
	});
	let filteredRows = $derived(coverageRows.filter((row) => row.subject.toLowerCase().includes(search.trim().toLowerCase())));
	let totalMapped = $derived(coverageRows.reduce((sum, row) => sum + row.total, 0));
	let totalKd = $derived(coverageRows.reduce((sum, row) => sum + row.kdCount, 0));
	let missingKd = $derived(coverageRows.reduce((sum, row) => sum + row.missingKd, 0));
	let sparseRows = $derived(coverageRows.filter((row) => row.total === 0 || row.missingKd > 0).slice(0, 8));
	let canCreate = $derived(canCreateBankSoal(data.user));

	onMount(load);
</script>

<svelte:head><title>Mapel & KD - Bank Soal</title></svelte:head>

<div class="space-y-5 p-4 md:p-6">
	<section class="overflow-hidden rounded-2xl border border-primary/20 bg-card shadow-sm">
		<div class="bg-gradient-to-r from-primary/10 via-card to-warning/10 p-4 md:p-5">
			<div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
				<div>
					<p class="text-[10px] font-black uppercase tracking-[0.28em] text-primary">Advanced Bank Soal</p>
					<h1 class="mt-1 text-2xl font-black uppercase italic tracking-tight text-foreground">Mapel & KD Coverage</h1>
					<p class="mt-2 max-w-3xl text-sm leading-6 text-muted-foreground">Pantau pemerataan bank soal berdasarkan mapel, kelas/fase, KD/CP/TP, materi, dan status review agar repositori soal siap dipakai lintas asesmen.</p>
				</div>
				<div class="flex flex-wrap gap-2">
					<a href={resolve('/bank-soal')} class="rounded-md border border-border bg-card px-3 py-2 text-sm font-semibold text-foreground hover:bg-muted/50">Dashboard</a>
					{#if canCreate}
						<a href={resolve('/bank-soal/tambah')} class="rounded-md border border-primary/20 bg-primary/10 px-3 py-2 text-sm font-semibold text-primary hover:bg-primary/15">Tambah Soal</a>
					{/if}
				</div>
			</div>
		</div>
	</section>

	<AsyncContent {promise}>
		{#snippet pending()}
			<Skeleton class="h-96 w-full" />
		{/snippet}
		{#snippet failed(error, reset)}
			<RecoveryPanel title="Coverage belum tersedia" message={error instanceof Error ? error.message : 'Gagal memuat coverage mapel.'} onRetry={() => { reset?.(); load(); }} />
		{/snippet}
		{#snippet children()}
			<section class="grid gap-3 md:grid-cols-4">
				<div class="rounded-xl border border-border bg-card p-4 shadow-sm"><p class="text-[10px] font-bold uppercase tracking-wide text-muted-foreground">Mapel</p><p class="mt-2 text-3xl font-black text-foreground">{coverageRows.length}</p></div>
				<div class="rounded-xl border border-border bg-card p-4 shadow-sm"><p class="text-[10px] font-bold uppercase tracking-wide text-muted-foreground">Soal Terpetakan</p><p class="mt-2 text-3xl font-black text-foreground">{totalMapped}</p></div>
				<div class="rounded-xl border border-border bg-card p-4 shadow-sm"><p class="text-[10px] font-bold uppercase tracking-wide text-muted-foreground">KD/CP/TP unik</p><p class="mt-2 text-3xl font-black text-foreground">{totalKd}</p></div>
				<div class="rounded-xl border border-warning/30 bg-warning/10 p-4 shadow-sm"><p class="text-[10px] font-bold uppercase tracking-wide text-warning">Belum lengkap</p><p class="mt-2 text-3xl font-black text-warning">{missingKd}</p></div>
			</section>

			<section class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_22rem]">
				<div class="rounded-xl border border-border bg-card p-4 shadow-sm">
					<div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
						<div><h2 class="text-base font-bold text-foreground">Matriks Mapel & KD</h2><p class="text-xs text-muted-foreground">Sampel maksimal 300 soal terbaru dari Bank Soal.</p></div>
						<input bind:value={search} placeholder="Cari mapel..." class="h-9 rounded-md border border-border px-3 text-sm" />
					</div>
					<div class="mt-4 overflow-hidden rounded-lg border border-border">
						<div class="grid grid-cols-[minmax(10rem,1.5fr)_repeat(4,minmax(5rem,0.7fr))] bg-muted/50 px-3 py-2 text-[10px] font-bold uppercase tracking-wide text-muted-foreground">
							<span>Mapel</span><span>Soal</span><span>Review</span><span>KD</span><span>Belum KD</span>
						</div>
						{#each filteredRows as row (row.key)}
							<div class="grid grid-cols-[minmax(10rem,1.5fr)_repeat(4,minmax(5rem,0.7fr))] items-center border-t border-border px-3 py-3 text-sm">
								<div class="min-w-0"><p class="truncate font-semibold text-foreground">{row.subject}</p><p class="truncate text-xs text-muted-foreground">{row.grades.join(', ') || 'Fase/kelas belum dominan'}</p></div>
								<span>{row.total}</span><span>{row.approved}</span><span>{row.kdCount}</span><span class={row.missingKd > 0 ? 'font-semibold text-warning' : 'text-primary'}>{row.missingKd}</span>
							</div>
						{:else}
							<p class="border-t border-border p-4 text-sm text-muted-foreground">Tidak ada mapel sesuai pencarian.</p>
						{/each}
					</div>
				</div>

				<aside class="space-y-4">
					<div class="rounded-xl border border-border bg-card p-4 shadow-sm">
						<h2 class="text-base font-bold text-foreground">Perlu Dilengkapi</h2>
						<div class="mt-3 space-y-2">
							{#each sparseRows as row (row.key)}
								<div class="rounded-lg border border-warning/30 bg-warning/10 px-3 py-2 text-sm text-warning"><div class="flex justify-between gap-3"><span class="font-semibold">{row.subject}</span><span>{row.missingKd || '0'} belum KD</span></div><p class="mt-1 text-xs text-warning">{row.total === 0 ? 'Belum ada soal pada sampel.' : `${row.kdCount} KD/CP/TP · ${row.topicCount} materi`}</p></div>
							{:else}
								<p class="text-sm text-muted-foreground">Coverage mapel terlihat rapi pada sampel saat ini.</p>
							{/each}
						</div>
					</div>
					<div class="rounded-xl border border-primary/20 bg-primary/10 p-4 text-primary"><p class="text-sm font-bold">Rekomendasi</p><p class="mt-2 text-xs leading-5 text-primary">Gunakan halaman ini untuk menentukan mapel/KD prioritas sebelum membuat paket asesmen baru. Lengkapi metadata KD/CP/TP dari komposer agar analisis makin akurat.</p></div>
				</aside>
			</section>
		{/snippet}
	</AsyncContent>
</div>
