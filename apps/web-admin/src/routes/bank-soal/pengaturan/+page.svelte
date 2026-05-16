<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { readClientApiData } from '$lib/client/api';
	import {
		canAssignBankSoalReviewer,
		canCreateBankSoal,
		canImportBankSoal,
		canReviewBankSoal,
		type BankSoalAccessUser
	} from '$lib/bank-soal/access';

	type SummaryResponse = {
		counts?: Partial<Record<'all' | 'total' | 'draft' | 'review' | 'revision' | 'approved' | 'published' | 'package_usage', number>>;
		by_subject?: Array<{ subject_name?: string; subject_code?: string; total?: number }>;
		by_cognitive_level?: Array<{ cognitive_level?: string; total?: number }>;
	};
	type Subject = { id: string; name: string; code?: string };
	type SubjectPayload = { subjects?: Subject[] };
	type UserOption = { id: string; username: string; display_name?: string; profile_nama?: string; roles?: string[]; is_active?: boolean; employee_id?: string };
	type ReviewerScope = {
		id: string;
		user_id: string;
		username: string;
		user_display_name: string;
		user_roles?: string[];
		subject_id?: string | null;
		subject_name?: string;
		subject_code?: string;
		grade_level?: number | null;
		can_review: boolean;
		can_approve: boolean;
		assigned_by_display_name?: string;
		updated_at?: string;
	};
	type ReviewerScopePayload = { items?: ReviewerScope[] };
	type Payload = {
		summary: SummaryResponse;
		subjects: Subject[];
		users: UserOption[];
		scopes: ReviewerScope[];
	};
	type PageData = { user?: BankSoalAccessUser };

	const workflowSteps = [
		{ label: 'Draft', desc: 'Guru menyusun metadata, naskah, opsi/kunci, dan pembahasan sebelum diajukan.' },
		{ label: 'Review', desc: 'Reviewer memeriksa substansi, konstruksi, bahasa, kunci/rubrik, dan kesesuaian KD/CP/TP.' },
		{ label: 'Revisi', desc: 'Soal dikembalikan jika perlu perbaikan. Catatan reviewer wajib jelas dan bisa ditindaklanjuti.' },
		{ label: 'Approved', desc: 'Soal lolos review dan siap dipakai untuk paket asesmen internal.' },
		{ label: 'Terbit', desc: 'Soal tersedia untuk pemakaian paket dan menjadi bagian bank soal pakai ulang.' }
	];
	const qualityRules = [
		'Isi metadata mapel, kelas/fase, KD/CP/TP, materi, level kognitif, dan kesulitan sebelum review.',
		'Naskah soal wajib jelas, bebas ambigu, dan tidak bergantung pada informasi di luar stimulus.',
		'Soal pilihan wajib memiliki kunci benar; essay/isian wajib memiliki rubrik atau jawaban acuan.',
		'Pembahasan dianjurkan untuk semua tipe soal agar bank soal bisa dipakai ulang untuk remedial/pengayaan.',
		'Gunakan flag HOTS hanya jika soal menuntut analisis, evaluasi, atau kreasi; bukan sekadar narasi panjang.',
		'Soal yang sudah dipakai paket/jawaban tidak diedit sembarang; lakukan duplikasi/revisi versi bila perlu perubahan besar.'
	];
	const integrations = [
		{ name: 'Daftar Soal', path: resolve('/bank-soal/daftar'), desc: 'Pencarian, filter, pagination, dan aksi per soal.', required: 'read' },
		{ name: 'Penyusun soal', path: resolve('/bank-soal/tambah'), desc: 'Pembuatan/edit soal dengan pratinjau siswa.', required: 'create' },
		{ name: 'Review', path: resolve('/bank-soal/verifikasi'), desc: 'Antrean verifikasi, catatan reviewer, approve/revisi.', required: 'review' },
		{ name: 'Import', path: resolve('/bank-soal/impor'), desc: 'Pratinjau cek data dan impor final.', required: 'import' },
		{ name: 'Asesmen Paket', path: resolve('/asesmen/paket'), desc: 'Pemakaian soal terbit ke paket asesmen.', required: 'read' }
	] as const;

	let { data }: { data: PageData } = $props();
	let promise = $state<Promise<Payload> | null>(null);
	let summary = $state<SummaryResponse>({});
	let subjects = $state<Subject[]>([]);
	let users = $state<UserOption[]>([]);
	let scopes = $state<ReviewerScope[]>([]);
	let errorMessage = $state('');
	let saving = $state(false);
	let selectedUserID = $state('');
	let selectedSubjectID = $state('');
	let selectedGradeLevel = $state('');
	let canReview = $state(true);
	let canApprove = $state(false);

	let canAssignReviewer = $derived(canAssignBankSoalReviewer(data.user));
	let totalQuestions = $derived(summary.counts?.total ?? summary.counts?.all ?? 0);
	let readyQuestions = $derived((summary.counts?.approved ?? 0) + (summary.counts?.published ?? 0));
	let pendingReview = $derived(summary.counts?.review ?? 0);
	let completionRate = $derived(totalQuestions > 0 ? Math.round((readyQuestions / totalQuestions) * 100) : 0);
	let subjectCoverage = $derived(summary.by_subject?.length ?? subjects.length ?? 0);
	let cognitiveCoverage = $derived(summary.by_cognitive_level?.filter((item) => (item.total ?? 0) > 0).length ?? 0);
	let reviewerUserOptions = $derived(users.filter((user) => user.is_active !== false && (user.employee_id || user.roles?.includes('guru') || user.roles?.includes('admin'))));
	let operationalStatus = $derived([
		{ label: 'Kesiapan Bank Soal', value: `${completionRate}%`, desc: `${readyQuestions} dari ${totalQuestions} soal approved/published` },
		{ label: 'Antrean Verifikasi', value: pendingReview, desc: 'Soal menunggu keputusan reviewer' },
		{ label: 'Scope Reviewer', value: scopes.length, desc: 'Scope review/approve manual aktif' },
		{ label: 'Coverage Mapel', value: subjectCoverage, desc: 'Mapel muncul pada ringkasan/sumber akademik' },
		{ label: 'Level Kognitif', value: cognitiveCoverage, desc: 'Kategori Bloom/C-level berisi soal' }
	]);
	let visibleIntegrations = $derived(integrations.filter((item) => {
		if (item.required === 'create') return canCreateBankSoal(data.user);
		if (item.required === 'review') return canReviewBankSoal(data.user);
		if (item.required === 'import') return canImportBankSoal(data.user);
		return true;
	}));

	function displayUser(user: UserOption) {
		return user.display_name || user.profile_nama || user.username;
	}
	function scopeSubjectLabel(scope: ReviewerScope) {
		if (!scope.subject_id) return 'Semua mapel';
		return scope.subject_code ? `${scope.subject_name} (${scope.subject_code})` : scope.subject_name || 'Mapel tidak ditemukan';
	}
	function scopeGradeLabel(grade?: number | null) {
		return grade ? `Kelas ${grade}` : 'Semua tingkat';
	}
	function resetScopeForm() {
		selectedUserID = '';
		selectedSubjectID = '';
		selectedGradeLevel = '';
		canReview = true;
		canApprove = false;
	}
	function editScope(scope: ReviewerScope) {
		selectedUserID = scope.user_id;
		selectedSubjectID = scope.subject_id ?? '';
		selectedGradeLevel = scope.grade_level ? String(scope.grade_level) : '';
		canReview = scope.can_review;
		canApprove = scope.can_approve;
	}

	async function fetchPayload(): Promise<Payload> {
		const [summaryPayload, subjectPayload, scopePayload, userPayload] = await Promise.all([
			fetch('/api/bank-soal/summary').then((response) => readClientApiData<SummaryResponse>(response, 'Gagal memuat ringkasan Bank Soal')),
			fetch('/api/bank-soal/soal-support/subjects').then((response) => readClientApiData<SubjectPayload>(response, 'Gagal memuat mapel')),
			canAssignReviewer ? fetch('/api/bank-soal/reviewer-scopes').then((response) => readClientApiData<ReviewerScopePayload>(response, 'Gagal memuat scope reviewer')) : Promise.resolve({ items: [] }),
			canAssignReviewer ? fetch('/api/users').then((response) => readClientApiData<UserOption[]>(response, 'Gagal memuat user reviewer')) : Promise.resolve([])
		]);
		return {
			summary: summaryPayload ?? {},
			subjects: subjectPayload.subjects ?? [],
			scopes: scopePayload.items ?? [],
			users: userPayload ?? []
		};
	}
	function load() {
		errorMessage = '';
		promise = fetchPayload().then((payload) => {
			summary = payload.summary;
			subjects = payload.subjects;
			scopes = payload.scopes;
			users = payload.users;
			return payload;
		});
	}
	async function saveScope() {
		if (!selectedUserID) {
			errorMessage = 'Pilih user/guru reviewer terlebih dahulu.';
			return;
		}
		saving = true;
		errorMessage = '';
		try {
			const response = await fetch('/api/bank-soal/reviewer-scopes', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					user_id: selectedUserID,
					subject_id: selectedSubjectID || null,
					grade_level: selectedGradeLevel ? Number(selectedGradeLevel) : null,
					can_review: canReview,
					can_approve: canApprove
				})
			});
			await readClientApiData(response, 'Gagal menyimpan scope reviewer');
			resetScopeForm();
			load();
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Gagal menyimpan scope reviewer.';
		} finally {
			saving = false;
		}
	}
	async function deleteScope(scope: ReviewerScope) {
		if (!confirm(`Hapus scope ${scope.user_display_name} untuk ${scopeSubjectLabel(scope)}?`)) return;
		saving = true;
		errorMessage = '';
		try {
			const response = await fetch(`/api/bank-soal/reviewer-scopes?id=${encodeURIComponent(scope.id)}`, { method: 'DELETE' });
			if (!response.ok) await readClientApiData(response, 'Gagal menghapus scope reviewer');
			load();
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Gagal menghapus scope reviewer.';
		} finally {
			saving = false;
		}
	}

	onMount(load);
</script>

<svelte:head><title>Pengaturan Bank Soal</title></svelte:head>

<div class="space-y-5 p-4 md:p-6">
	<section class="overflow-hidden rounded-2xl border border-primary/20 bg-card shadow-sm">
		<div class="bg-gradient-to-r from-primary/10 via-card to-muted/50 p-4 md:p-5">
			<div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
				<div>
					<p class="text-[10px] font-black uppercase tracking-[0.28em] text-primary">Governance Bank Soal</p>
					<h1 class="mt-1 text-2xl font-black uppercase italic tracking-tight text-foreground">Pengaturan & Scope Reviewer</h1>
					<p class="mt-2 max-w-3xl text-sm leading-6 text-muted-foreground">Pusat SOP Bank Soal sekaligus pengaturan reviewer/approver per mapel dan tingkat. Sprint 1 masih additive: belum mencabut permission guru dan belum enforce filtering.</p>
				</div>
				<div class="flex flex-wrap gap-2">
					<a href={resolve('/bank-soal')} class="rounded-md border border-border bg-card px-3 py-2 text-sm font-semibold text-foreground hover:bg-muted/50">Dashboard</a>
					<a href={resolve('/bank-soal/analisis-butir')} class="rounded-md border border-primary/20 bg-primary/10 px-3 py-2 text-sm font-semibold text-primary hover:bg-primary/15">Analisis</a>
				</div>
			</div>
		</div>
	</section>

	<AsyncContent {promise}>
		{#snippet pending()}
			<Skeleton class="h-80 w-full" />
		{/snippet}
		{#snippet failed(error, reset)}
			<RecoveryPanel title="Pengaturan belum bisa dimuat" message={error instanceof Error ? error.message : 'Gagal memuat data pengaturan.'} onRetry={() => { reset?.(); load(); }} />
		{/snippet}
		{#snippet children()}
			<section class="grid gap-3 md:grid-cols-2 xl:grid-cols-5">
				{#each operationalStatus as item (item.label)}
					<article class="rounded-xl border border-border bg-card p-4 shadow-sm">
						<p class="text-[10px] font-bold uppercase tracking-[0.18em] text-muted-foreground">{item.label}</p>
						<p class="mt-2 text-3xl font-black text-foreground">{item.value}</p>
						<p class="mt-1 text-xs text-muted-foreground">{item.desc}</p>
					</article>
				{/each}
			</section>

			{#if canAssignReviewer}
				<section class="grid gap-4 xl:grid-cols-[24rem_minmax(0,1fr)]">
					<form class="rounded-xl border border-border bg-card p-4 shadow-sm" onsubmit={(event) => { event.preventDefault(); saveScope(); }}>
						<div class="flex items-start justify-between gap-3">
							<div>
								<h2 class="text-base font-bold text-foreground">Tambah / Update Scope Reviewer</h2>
								<p class="mt-1 text-xs leading-5 text-muted-foreground">Scope kosong berarti semua mapel/tingkat. Data ini belum dipakai untuk hard enforcement sampai Sprint berikutnya.</p>
							</div>
							<button type="button" class="rounded-md border border-border px-3 py-1.5 text-xs font-semibold" onclick={resetScopeForm}>Reset</button>
						</div>
						<div class="mt-4 space-y-3">
							<label for="reviewer-user" class="block text-xs font-bold uppercase tracking-wide text-muted-foreground">User/Guru</label>
							<select id="reviewer-user" bind:value={selectedUserID} class="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm">
								<option value="">Pilih user reviewer</option>
								{#each reviewerUserOptions as user (user.id)}
									<option value={user.id}>{displayUser(user)} — {user.username}</option>
								{/each}
							</select>
							<label for="reviewer-subject" class="block text-xs font-bold uppercase tracking-wide text-muted-foreground">Mapel</label>
							<select id="reviewer-subject" bind:value={selectedSubjectID} class="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm">
								<option value="">Semua mapel</option>
								{#each subjects as subject (subject.id)}
									<option value={subject.id}>{subject.name}{subject.code ? ` (${subject.code})` : ''}</option>
								{/each}
							</select>
							<label for="reviewer-grade" class="block text-xs font-bold uppercase tracking-wide text-muted-foreground">Tingkat</label>
							<select id="reviewer-grade" bind:value={selectedGradeLevel} class="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm">
								<option value="">Semua tingkat</option>
								<option value="7">VII</option>
								<option value="8">VIII</option>
								<option value="9">IX</option>
							</select>
							<label class="flex items-center gap-2 rounded-lg border border-border bg-muted/40 p-3 text-sm"><input type="checkbox" bind:checked={canReview} /> Bisa review</label>
							<label class="flex items-center gap-2 rounded-lg border border-border bg-muted/40 p-3 text-sm"><input type="checkbox" bind:checked={canApprove} /> Bisa approve</label>
							{#if errorMessage}<p class="rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">{errorMessage}</p>{/if}
							<button type="submit" disabled={saving} class="w-full rounded-lg bg-primary px-4 py-2 text-sm font-bold text-primary-foreground disabled:opacity-60">{saving ? 'Menyimpan…' : 'Simpan Scope'}</button>
						</div>
					</form>

					<div class="rounded-xl border border-border bg-card p-4 shadow-sm">
						<div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
							<div>
								<h2 class="text-base font-bold text-foreground">Daftar Scope Reviewer</h2>
								<p class="mt-1 text-xs text-muted-foreground">Manual assignment reviewer/approver Bank Soal. Scope masih foundation dan belum mengubah visibilitas soal.</p>
							</div>
							<button type="button" class="rounded-md border border-border px-3 py-2 text-xs font-semibold" onclick={load}>Muat ulang</button>
						</div>
						<div class="mt-4 overflow-x-auto">
							<table class="min-w-full text-left text-sm">
								<thead class="text-xs uppercase tracking-wide text-muted-foreground">
									<tr><th class="p-2">Reviewer</th><th class="p-2">Mapel</th><th class="p-2">Tingkat</th><th class="p-2">Hak</th><th class="p-2">Aksi</th></tr>
								</thead>
								<tbody>
									{#if scopes.length === 0}
										<tr><td colspan="5" class="p-4 text-center text-muted-foreground">Belum ada scope reviewer. Tambahkan manual setelah data guru-mapel siap.</td></tr>
									{:else}
										{#each scopes as scope (scope.id)}
											<tr class="border-t border-border align-top">
												<td class="p-2"><p class="font-semibold text-foreground">{scope.user_display_name}</p><p class="text-xs text-muted-foreground">{scope.username}</p></td>
												<td class="p-2">{scopeSubjectLabel(scope)}</td>
												<td class="p-2">{scopeGradeLabel(scope.grade_level)}</td>
												<td class="p-2"><div class="flex flex-wrap gap-1">{#if scope.can_review}<span class="rounded-full bg-amber-100 px-2 py-1 text-xs font-bold text-amber-700">Review</span>{/if}{#if scope.can_approve}<span class="rounded-full bg-emerald-100 px-2 py-1 text-xs font-bold text-emerald-700">Approve</span>{/if}</div></td>
												<td class="p-2"><div class="flex gap-2"><button type="button" class="rounded-md border border-border px-2 py-1 text-xs" onclick={() => editScope(scope)}>Edit</button><button type="button" class="rounded-md border border-destructive/30 px-2 py-1 text-xs text-destructive" onclick={() => deleteScope(scope)}>Hapus</button></div></td>
											</tr>
										{/each}
									{/if}
								</tbody>
							</table>
						</div>
					</div>
				</section>
			{:else}
				<section class="rounded-xl border border-warning/30 bg-warning/10 p-4 text-sm text-warning">Akun ini belum memiliki permission <code>bank_soal.assign_reviewer</code> atau <code>bank_soal.settings</code>, sehingga pengaturan scope reviewer disembunyikan.</section>
			{/if}

			<section class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_24rem]">
				<div class="space-y-4">
					<div class="rounded-xl border border-border bg-card p-4 shadow-sm">
						<h2 class="text-base font-bold text-foreground">Alur Standar</h2>
						<div class="mt-4 grid gap-3 md:grid-cols-5">
							{#each workflowSteps as step, index (step.label)}
								<div class="rounded-xl border border-border bg-muted/50 p-3"><div class="flex items-center gap-2"><span class="flex size-7 items-center justify-center rounded-full bg-card text-xs font-black text-foreground">{index + 1}</span><p class="text-sm font-bold text-foreground">{step.label}</p></div><p class="mt-2 text-xs leading-5 text-muted-foreground">{step.desc}</p></div>
							{/each}
						</div>
					</div>
					<div class="rounded-xl border border-border bg-card p-4 shadow-sm"><h2 class="text-base font-bold text-foreground">Standar Kualitas Minimum</h2><div class="mt-4 grid gap-2 md:grid-cols-2">{#each qualityRules as rule, index (rule)}<div class="rounded-lg border border-primary/20 bg-primary/10 p-3 text-sm leading-5 text-primary"><span class="mr-2 inline-flex size-5 items-center justify-center rounded-full bg-primary text-[10px] font-black text-primary-foreground">{index + 1}</span>{rule}</div>{/each}</div></div>
				</div>
				<aside class="space-y-4">
					<div class="rounded-xl border border-border bg-card p-4 shadow-sm"><h2 class="text-base font-bold text-foreground">Integrasi Modul</h2><div class="mt-3 space-y-2">{#each visibleIntegrations as item (item.path)}<a href={item.path} class="block rounded-lg border border-border bg-muted/50 p-3 transition hover:border-primary/20 hover:bg-primary/10"><p class="text-sm font-bold text-foreground">{item.name}</p><p class="mt-1 text-xs leading-5 text-muted-foreground">{item.desc}</p></a>{/each}</div></div>
					<div class="rounded-xl border border-warning/30 bg-warning/10 p-4 text-warning"><p class="text-sm font-bold">Catatan Sprint 1</p><p class="mt-2 text-xs leading-5">Permission dan scope reviewer ditambahkan secara additive. Enforcement visibilitas dan workflow action tetap menunggu Sprint 2+.</p></div>
				</aside>
			</section>
		{/snippet}
	</AsyncContent>
</div>
