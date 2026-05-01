<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import {
		defaultSchoolProfile,
		fetchSchoolProfile,
		schoolAddressLine,
		type SchoolProfile,
	} from '$lib/school-profile';

	type Assignment = {
		id: string;
		class_name: string;
		class_code: string;
		subject_name: string;
		subject_code: string;
		teacher_name: string;
	};
	type GradeComponent = {
		id: string;
		title: string;
		category: string;
		weight: number;
		max_score: number;
		is_published: boolean;
	};
	type GradeSummary = {
		student_id: string;
		nis: string;
		nisn: string;
		nama: string;
		component_count: number;
		filled_count: number;
		final_score: number;
	};
	type RaporDetail = {
		assignmentId: string;
		components: GradeComponent[];
		summary: GradeSummary[];
	};
	type GradeOverview = {
		assignments?: Assignment[];
		components?: GradeComponent[];
		summary?: GradeSummary[];
		error?: string;
		message?: string;
	};
	type ApiEnvelope<T> = {
		data?: T;
		error?: string;
		message?: string;
	};

	const categoryLabel: Record<string, string> = {
		assignment: 'Tugas', quiz: 'Kuis', midterm: 'UTS', final: 'UAS',
		project: 'Proyek', practice: 'Praktik', attitude: 'Sikap',
		attendance: 'Kehadiran', other: 'Lainnya',
	};

	let assignments = $state<Assignment[]>([]);
	let assignmentsPromise = $state<Promise<Assignment[]> | null>(null);
	let selectedId = $state('');
	let components = $state<GradeComponent[]>([]);
	let summary = $state<GradeSummary[]>([]);
	let raporPromise = $state<Promise<RaporDetail> | null>(null);
	let schoolProfile = $state<SchoolProfile | null>(null);

	let selectedAssignment = $derived(assignments.find(a => a.id === selectedId) ?? null);
	let activeSchoolProfile = $derived(schoolProfile ?? defaultSchoolProfile);
	let validScores = $derived(summary.filter((s) => s.final_score >= 0).map((s) => s.final_score));
	let completedCount = $derived(summary.filter((s) => s.final_score >= 75).length);
	let classAverage = $derived(
		validScores.length > 0 ? (validScores.reduce((a, b) => a + b, 0) / validScores.length).toFixed(1) : '—'
	);

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

	async function fetchAssignments() {
		const overview = await fetch('/api/grades').then((response) => readApi<GradeOverview>(response, 'Gagal memuat daftar kelas dan mata pelajaran'));
		return overview.assignments ?? [];
	}

	async function fetchRapor(assignmentId: string): Promise<RaporDetail> {
		const overview = await fetch(`/api/grades?assignment_id=${assignmentId}&published_only=true`)
			.then((response) => readApi<GradeOverview>(response, 'Gagal memuat data rapor'));
		return {
			assignmentId,
			components: overview.components ?? [],
			summary: (overview.summary ?? []).sort((a, b) => a.nama.localeCompare(b.nama, 'id')),
		};
	}

	function loadRapor(assignmentId: string) {
		components = [];
		summary = [];
		raporPromise = fetchRapor(assignmentId).then((detail) => {
			if (selectedId === assignmentId) {
				components = detail.components;
				summary = detail.summary;
			}
			return detail;
		});
	}

	function loadAssignments() {
		assignments = [];
		assignmentsPromise = fetchAssignments().then((rows) => {
			assignments = rows;
			if (!selectedId) {
				const requestedId = new URL(window.location.href).searchParams.get('assignment_id') ?? '';
				if (requestedId && rows.some((item) => item.id === requestedId)) {
					selectedId = requestedId;
					loadRapor(requestedId);
				}
			}
			return rows;
		});
	}

	async function loadSchoolProfile() {
		try {
			schoolProfile = await fetchSchoolProfile();
		} catch (error) {
			console.error('Rapor school profile load failed', error);
		}
	}

	function selectAssignment(assignmentId: string) {
		selectedId = assignmentId;
		if (!assignmentId) {
			components = [];
			summary = [];
			raporPromise = null;
			return;
		}
		loadRapor(assignmentId);
	}

	function retryAssignments(reset?: () => void) {
		reset?.();
		loadAssignments();
	}

	function retryRapor(reset?: () => void) {
		reset?.();
		if (selectedId) loadRapor(selectedId);
	}

	function assignmentsErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat daftar kelas dan mata pelajaran';
	}

	function raporErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat data rapor';
	}

	function handleAssignmentsRenderError(error: unknown) {
		console.error('Rapor assignments render failed', error);
	}

	function handleRaporRenderError(error: unknown) {
		console.error('Rapor detail render failed', error);
	}

	function scoreClass(score: number) {
		if (score < 0) return 'text-slate-400';
		if (score >= 80) return 'text-emerald-700 font-semibold';
		if (score >= 65) return 'text-amber-700';
		return 'text-red-700';
	}

	function fmt(score: number) {
		if (score < 0) return '—';
		return score.toFixed(1);
	}

	onMount(() => {
		loadAssignments();
		void loadSchoolProfile();
	});
</script>

<svelte:head><title>Cetak Rapor — MTsN 2 Kolaka Utara</title></svelte:head>

<style>
	@media print {
		.no-print { display: none !important; }
		.print-page {
			font-size: 11pt;
			color: #000;
		}
		table { border-collapse: collapse; width: 100%; }
		th, td { border: 1px solid #000; padding: 4px 6px; }
		th { background: #f3f4f6; }
	}
</style>

<div class="space-y-6">
	<!-- Header + controls (hidden on print) -->
	<div class="no-print flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
		<div>
			<h1 class="text-2xl font-semibold text-slate-800">Cetak Rapor</h1>
			<p class="mt-1 text-sm text-muted-foreground">Pilih mata pelajaran dan kelas, lalu cetak daftar nilai siswa.</p>
		</div>
		{#if selectedId && summary.length > 0}
			<LoadingButton onclick={() => window.print()} label="Cetak / Simpan PDF" />
		{/if}
	</div>

	<!-- Assignment selector -->
	<div class="no-print">
		<Card.Root class="border-slate-200 shadow-sm">
			<Card.Content class="p-4">
				<label for="rapor-assignment" class="mb-1 block text-xs font-medium text-slate-600">Pilih Mata Pelajaran / Kelas</label>
				<AsyncContent promise={assignmentsPromise} onerror={handleAssignmentsRenderError}>
					{#snippet pending()}
						<div class="space-y-2">
							<Skeleton class="h-4 w-40" />
							<Skeleton class="h-10 w-full max-w-xl" />
						</div>
					{/snippet}

					{#snippet failed(error, reset)}
						<RecoveryPanel
							compact
							title="Pilihan Rapor Belum Tersaji"
							message={assignmentsErrorMessage(error)}
							onRetry={() => retryAssignments(reset)}
						/>
					{/snippet}

					{#snippet children(value)}
						{@const currentAssignments = value as Assignment[]}
						<select
							id="rapor-assignment"
							value={selectedId}
							onchange={(event) => selectAssignment((event.currentTarget as HTMLSelectElement).value)}
							class="w-full max-w-xl rounded-md border border-input bg-background px-3 py-2 text-sm"
						>
							<option value="">— Pilih kelas dan mata pelajaran —</option>
							{#each currentAssignments as a (a.id)}
								<option value={a.id}>{a.class_name} · {a.subject_name} ({a.teacher_name})</option>
							{/each}
						</select>
					{/snippet}
				</AsyncContent>
			</Card.Content>
		</Card.Root>
	</div>

	{#if selectedAssignment}
		<!-- Print content -->
		<div class="print-page space-y-4">
			<!-- School header -->
			<div class="border-b-2 border-slate-800 pb-3 text-center no-print:border-slate-300">
				<p class="text-sm font-semibold uppercase text-slate-800">{activeSchoolProfile.ministry_line}</p>
				<p class="text-sm font-semibold uppercase text-slate-800">{activeSchoolProfile.office_line}</p>
				<p class="mt-1 text-base font-bold uppercase tracking-wide text-slate-900">{activeSchoolProfile.name}</p>
				<p class="mt-1 text-xs text-slate-600">{schoolAddressLine(activeSchoolProfile) || 'Alamat madrasah belum diisi'}</p>
				<p class="mt-2 text-xs font-semibold uppercase tracking-[0.16em] text-slate-700">Daftar Nilai Akademik Siswa</p>
			</div>

			<!-- Class + subject info -->
			<div class="grid grid-cols-2 gap-x-8 gap-y-1 text-sm text-slate-700">
				<div class="flex gap-2">
					<span class="w-32 font-medium text-slate-900">Kelas</span>
					<span>: {selectedAssignment.class_name} ({selectedAssignment.class_code})</span>
				</div>
				<div class="flex gap-2">
					<span class="w-32 font-medium text-slate-900">Guru</span>
					<span>: {selectedAssignment.teacher_name}</span>
				</div>
				<div class="flex gap-2">
					<span class="w-32 font-medium text-slate-900">Mata Pelajaran</span>
					<span>: {selectedAssignment.subject_name} ({selectedAssignment.subject_code})</span>
				</div>
				<div class="flex gap-2">
					<span class="w-32 font-medium text-slate-900">Tanggal Cetak</span>
					<span>: {new Date().toLocaleDateString('id-ID', { year: 'numeric', month: 'long', day: 'numeric' })}</span>
				</div>
			</div>

			<AsyncContent promise={raporPromise} onerror={handleRaporRenderError}>
				{#snippet pending()}
					<div class="space-y-3 py-4">
						{#each Array.from({ length: 6 }) as _, index (`rapor-row-skeleton-${index}`)}
							<div class="grid gap-3 md:grid-cols-[0.3fr_0.8fr_0.8fr_1.4fr_0.6fr_0.6fr_0.8fr] md:items-center">
								<Skeleton class="h-5 w-6" />
								<Skeleton class="h-5 w-20" />
								<Skeleton class="h-5 w-24" />
								<Skeleton class="h-5 w-full max-w-xs" />
								<Skeleton class="h-5 w-16" />
								<Skeleton class="h-5 w-14" />
								<Skeleton class="h-5 w-20" />
							</div>
						{/each}
					</div>
				{/snippet}

				{#snippet failed(error, reset)}
					<div class="no-print">
						<RecoveryPanel
							title="Rapor Belum Tersaji"
							message={raporErrorMessage(error)}
							onRetry={() => retryRapor(reset)}
						/>
					</div>
				{/snippet}

				{#snippet children(value)}
					{@const currentRapor = value as RaporDetail}
					{@const currentComponents = currentRapor.components}
					{@const currentSummary = currentRapor.summary}

					{#if currentComponents.length > 0}
						<div class="no-print flex flex-wrap gap-2">
							{#each currentComponents as c (c.id)}
								<Badge variant="outline" class="text-xs">
									{c.title} · {categoryLabel[c.category] ?? c.category} · Bobot {c.weight}
								</Badge>
							{/each}
						</div>
					{/if}

					{#if currentComponents.length === 0}
						<p class="py-6 text-center text-sm text-slate-500">Belum ada komponen nilai yang diterbitkan untuk rapor pada kelas dan mata pelajaran ini.</p>
					{:else if currentSummary.length === 0}
						<p class="py-6 text-center text-sm text-slate-500">Belum ada data nilai dari komponen yang sudah diterbitkan.</p>
					{:else}
						<div class="overflow-x-auto">
							<table class="min-w-full border border-slate-200 text-sm">
								<thead>
									<tr class="bg-slate-50 text-left text-xs font-semibold text-slate-700">
										<th class="border border-slate-200 px-3 py-2 text-center">No</th>
										<th class="border border-slate-200 px-3 py-2">NIS</th>
										<th class="border border-slate-200 px-3 py-2">NISN</th>
										<th class="border border-slate-200 px-3 py-2">Nama Siswa</th>
										<th class="border border-slate-200 px-3 py-2 text-center">Komponen</th>
										<th class="border border-slate-200 px-3 py-2 text-center">Nilai Akhir</th>
										<th class="border border-slate-200 px-3 py-2 text-center">Keterangan</th>
									</tr>
								</thead>
								<tbody>
									{#each currentSummary as s, i (s.student_id)}
										<tr class="hover:bg-slate-50 {i % 2 === 1 ? 'bg-slate-50/40' : ''}">
											<td class="border border-slate-200 px-3 py-2 text-center text-slate-600">{i + 1}</td>
											<td class="border border-slate-200 px-3 py-2 font-mono text-xs text-slate-700">{s.nis || '—'}</td>
											<td class="border border-slate-200 px-3 py-2 font-mono text-xs text-slate-700">{s.nisn || '—'}</td>
											<td class="border border-slate-200 px-3 py-2 font-medium text-slate-900">{s.nama}</td>
											<td class="border border-slate-200 px-3 py-2 text-center text-slate-600">
												{s.filled_count}/{s.component_count}
											</td>
											<td class="border border-slate-200 px-3 py-2 text-center text-base {scoreClass(s.final_score)}">
												{fmt(s.final_score)}
											</td>
											<td class="border border-slate-200 px-3 py-2 text-center text-xs">
												{#if s.final_score < 0}
													<span class="text-slate-400">Belum dinilai</span>
												{:else if s.final_score >= 75}
													<span class="text-emerald-700">Tuntas</span>
												{:else}
													<span class="text-red-700">Belum Tuntas</span>
												{/if}
											</td>
										</tr>
									{/each}
								</tbody>
								<tfoot>
									<tr class="bg-slate-100 font-medium text-slate-800">
										<td colspan="5" class="border border-slate-200 px-3 py-2 text-right text-xs">
											Rata-rata kelas:
										</td>
										<td class="border border-slate-200 px-3 py-2 text-center">
											{classAverage}
										</td>
										<td class="border border-slate-200 px-3 py-2 text-center text-xs text-slate-600">
											{completedCount}/{currentSummary.length} tuntas
										</td>
									</tr>
								</tfoot>
							</table>
						</div>

						<!-- Signature area (print only) -->
						<div class="mt-8 hidden print:grid grid-cols-2 gap-8 text-sm text-slate-700">
							<div class="text-center">
								<p>Mengetahui,</p>
								<p>Kepala Madrasah</p>
								<div class="mt-16 border-t border-slate-700 pt-1">
									<p class="font-semibold">{activeSchoolProfile.head_name || '___________________________'}</p>
									<p class="text-xs">NIP. {activeSchoolProfile.head_nip || '___________________________'}</p>
								</div>
							</div>
							<div class="text-center">
								<p>Guru Mata Pelajaran,</p>
								<p>{selectedAssignment.subject_name}</p>
								<div class="mt-16 border-t border-slate-700 pt-1">
									<p class="font-semibold">{selectedAssignment.teacher_name}</p>
									<p class="text-xs">NIP. ___________________________</p>
								</div>
							</div>
						</div>
					{/if}
				{/snippet}
			</AsyncContent>
		</div>
	{:else if !selectedId}
		<Card.Root class="no-print border-slate-200 shadow-sm">
			<Card.Content class="py-16 text-center">
				<p class="text-sm text-slate-500">Pilih mata pelajaran dan kelas untuk melihat daftar nilai.</p>
			</Card.Content>
		</Card.Root>
	{/if}
</div>
