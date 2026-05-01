<script lang="ts">
	import { onMount } from 'svelte';
	import { SvelteSet } from 'svelte/reactivity';
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

	type CbtPackage = {
		id: string; subject_id: string; subject_name: string; subject_code: string;
		title: string; description: string; duration_minutes: number;
		randomize_questions: boolean; is_active: boolean;
		question_count: number; created_at: string;
	};
	type Question = {
		id: string; subject_id: string; subject_code: string;
		code: string; question_text: string; difficulty: string; status: string;
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
	type ApiEnvelope<T> = {
		data?: T;
		error?: string;
		message?: string;
	};

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

	let questionPool = $derived(
		fSubjectId
			? allQuestions.filter(q => q.subject_id === fSubjectId && q.status === 'published')
			: []
	);

	function toggleQuestion(id: string) {
		if (fSelectedIds.has(id)) fSelectedIds.delete(id);
		else fSelectedIds.add(id);
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

	function parseQuestions(payload: unknown) {
		return Array.isArray(payload) ? (payload as Question[]) : [];
	}

	function parseSubjects(payload: AcademicPayload | unknown) {
		return isRecord(payload) && Array.isArray(payload.subjects) ? (payload.subjects as Subject[]) : [];
	}

	async function fetchOverview(): Promise<PackagesOverview> {
		const [packagesPayload, questionsPayload, academicPayload] = await Promise.all([
			fetch('/api/cbt/packages').then((response) => readApi<CbtPackagesPayload>(response, 'Gagal memuat data paket')),
			fetch('/api/cbt/questions').then((response) => readApi<unknown>(response, 'Gagal memuat bank soal')),
			fetch('/api/academic').then((response) => readApi<AcademicPayload>(response, 'Gagal memuat data akademik')),
		]);
		return {
			packages: packagesPayload.packages ?? [],
			allQuestions: parseQuestions(questionsPayload),
			subjects: parseSubjects(academicPayload),
		};
	}

	function applyOverview(overview: PackagesOverview) {
		packages = overview.packages;
		allQuestions = overview.allQuestions;
		subjects = overview.subjects;
	}

	function load() {
		packages = [];
		allQuestions = [];
		subjects = [];
		packagesPromise = fetchOverview().then((overview) => {
			applyOverview(overview);
			return overview;
		});
	}

	async function refreshPackages() {
		if (!packagesPromise) {
			load();
			return;
		}
		try {
			const overview = await fetchOverview();
			applyOverview(overview);
			packagesPromise = Promise.resolve(overview);
		} catch (error) {
			packagesPromise = Promise.resolve({ packages, allQuestions, subjects });
			toast.error(packagesErrorMessage(error));
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

	async function responseErrorMessage(response: Response, fallback: string) {
		const payload = await response.json().catch(() => null);
		return apiErrorMessage(payload) || fallback;
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
					is_active: fActive, question_ids: Array.from(fSelectedIds),
				}),
			});
			if (!res.ok) { showError(await responseErrorMessage(res, 'Gagal membuat paket')); return; }
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
			const res = await fetch(`/api/cbt/packages?id=${id}`, { method: 'DELETE' });
			if (!res.ok) {
				setOperationState('error', 'Paket Gagal Dihapus', 'Periksa kembali apakah paket masih dipakai oleh sesi aktif atau coba ulang beberapa saat lagi.');
				showError('Gagal menghapus paket');
				return;
			}
			setOperationState('warning', 'Paket Dihapus', `Paket "${title}" sudah dihapus dari daftar paket ujian.`);
			showToast('Paket dihapus');
			await refreshPackages();
		} catch (error) {
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
						<select id="package-subject-id" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fSubjectId}>
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
							{#if fSelectedIds.size > 0}
								— <span class="text-green-700 font-medium">{fSelectedIds.size} dipilih</span>
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
											<div class="flex gap-1 mt-0.5">
												{#if q.code}
													<span class="text-xs text-slate-400 font-mono">{q.code}</span>
												{/if}
										<Badge variant="outline" class="text-xs py-0">{q.difficulty === 'easy' ? 'Mudah' : q.difficulty === 'medium' ? 'Sedang' : q.difficulty === 'hard' ? 'Sulit' : q.difficulty}</Badge>
											</div>
										</div>
									</label>
								{/each}
							</div>
						{/if}
					</div>
				{/if}

				<div class="flex gap-2">
					<LoadingButton disabled={fBusy || !fSubjectId || !fTitle || !fDuration} onclick={() => void createPackage()} loading={fBusy} loadingLabel="Menyimpan...">
						{`Buat Paket${fSelectedIds.size > 0 ? ` (${fSelectedIds.size} soal)` : ''}`}
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
