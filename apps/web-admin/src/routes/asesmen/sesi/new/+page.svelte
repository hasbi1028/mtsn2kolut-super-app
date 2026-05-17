<script lang="ts">
	import { onMount } from 'svelte';
	import { SvelteMap } from 'svelte/reactivity';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { toast } from '$lib/components/ui/sonner';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import OperationStatusPanel from '$lib/components/OperationStatusPanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { clientApiPath, clientApiPathWithQuery, readClientApiData, readClientJson } from '$lib/client/api';

	type CbtPackage = { id: string; title: string; subject_code: string; subject_name: string; question_count: number; is_active: boolean; event_id?: string | null };
	type PackageQuestion = { package_id: string; question_id: string; question_type?: string; status?: string; cp_ref?: string; tp_ref?: string; kd_ref?: string; cognitive_level?: string; hots_flag?: boolean };
	type SchoolClass = { id: string; name: string; code: string; level: string };
	type EventContext = { id: string; title: string; status: string; target_levels?: string[]; academic_year_name?: string };
	type CbtPackagesPayload = { packages?: CbtPackage[]; questions?: unknown[] };
	type AcademicPayload = { classes?: SchoolClass[] };
	type PackageQualitySummary = { questions: PackageQuestion[]; typeBuckets: { label: string; count: number }[]; hotsCount: number; missingCount: number; unpublishedCount: number; totalCount: number };
	type FormData = { packages: CbtPackage[]; packageQuestions: PackageQuestion[]; classes: SchoolClass[]; eventContext: EventContext | null };

	const eventId = page.url.searchParams.get('event_id') ?? '';

	let packages = $state<CbtPackage[]>([]);
	let packageQuestions = $state<PackageQuestion[]>([]);
	let classes = $state<SchoolClass[]>([]);
	let eventContext = $state<EventContext | null>(null);
	let formPromise = $state<Promise<FormData> | null>(null);
	let hiddenEventPackageCount = $state(0);
	let browserTimeZone = $state('');
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
	let operationState = $state<{ tone: 'success' | 'error' | 'warning' | 'info'; title: string; message: string } | null>(null);

	let selectedPackage = $derived(packages.find((pkg) => pkg.id === fPackageId) ?? null);
	let selectedPackageQuality = $derived(packageQualitySummary(fPackageId));
	let sessionReadinessIssues = $derived(buildSessionReadinessIssues());
	let canCreateSession = $derived(!fBusy && sessionReadinessIssues.length === 0);
	let browserTimeZoneMismatch = $derived(browserTimeZone !== '' && browserTimeZone !== 'Asia/Makassar');
	let listHref = $derived(`${resolve('/asesmen/sesi')}${eventId ? `?event_id=${encodeURIComponent(eventId)}` : ''}`);

	function isRecord(value: unknown): value is Record<string, unknown> {
		return typeof value === 'object' && value !== null;
	}

	function parsePackageQuestions(payload: CbtPackagesPayload | unknown): PackageQuestion[] {
		return isRecord(payload) && Array.isArray((payload as CbtPackagesPayload).questions) ? ((payload as CbtPackagesPayload).questions as PackageQuestion[]) : [];
	}

	function strictEventPackages(items: CbtPackage[]) {
		return eventId ? items.filter((pkg) => pkg.event_id === eventId) : items;
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
			essay: 'Essay'
		};
		const normalized = compactValue(value, '');
		return labels[normalized] ?? (normalized ? normalized.replaceAll('_', ' ') : 'Belum tipe');
	}

	function packageQuestionHasBlueprintGap(question: PackageQuestion) {
		return !compactValue(question.cp_ref, '') || (!compactValue(question.tp_ref, '') && !compactValue(question.kd_ref, '')) || !compactValue(question.cognitive_level, '');
	}

	function countByLabel<T>(items: T[], selector: (item: T) => string) {
		const counts = new SvelteMap<string, number>();
		for (const item of items) {
			const label = selector(item);
			counts.set(label, (counts.get(label) ?? 0) + 1);
		}
		return Array.from(counts.entries()).map(([label, count]) => ({ label, count })).sort((a, b) => b.count - a.count || a.label.localeCompare(b.label));
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
			totalCount: pkg?.question_count ?? questions.length
		};
	}

	function packageQualityIssues(packageID: string) {
		const issues: string[] = [];
		const pkg = packages.find((item) => item.id === packageID);
		const quality = packageQualitySummary(packageID);
		if (pkg && !pkg.is_active) issues.push('paket nonaktif');
		if (packageID && quality.totalCount === 0) issues.push('paket kosong');
		if (packageID && quality.questions.length > 0 && quality.unpublishedCount > 0) issues.push(`${quality.unpublishedCount} belum terbit`);
		return issues;
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

	async function fetchFormData(): Promise<FormData> {
		const packageParams = new URLSearchParams();
		if (eventId) packageParams.set('event_id', eventId);
		const [packagePayload, academicPayload, context] = await Promise.all([
			fetch(clientApiPathWithQuery('/api/asesmen/packages', packageParams)).then((response) => readClientApiData<CbtPackagesPayload>(response, 'Gagal memuat paket ujian')),
			fetch('/api/academic').then((response) => readClientApiData<AcademicPayload>(response, 'Gagal memuat data akademik')),
			fetchEventContext()
		]);
		return {
			packages: packagePayload.packages ?? [],
			packageQuestions: parsePackageQuestions(packagePayload),
			classes: academicPayload.classes ?? [],
			eventContext: context
		};
	}

	function loadForm() {
		packages = [];
		packageQuestions = [];
		classes = [];
		hiddenEventPackageCount = 0;
		formPromise = fetchFormData().then((data) => {
			hiddenEventPackageCount = hiddenPackageCount(data.packages);
			packages = strictEventPackages(data.packages);
			packageQuestions = data.packageQuestions;
			classes = data.classes;
			eventContext = data.eventContext;
			return data;
		});
	}

	function retryForm(reset?: () => void) {
		reset?.();
		loadForm();
	}

	function errorMessage(error: unknown, fallback: string) {
		if (error instanceof Error && error.message.trim() && !error.message.toLowerCase().includes('fetch')) return error.message;
		return fallback;
	}

	function handleRenderError(error: unknown) {
		console.error('Form sesi ujian belum dapat ditampilkan', error);
	}

	async function createSession() {
		if (sessionReadinessIssues.length > 0) {
			operationState = { tone: 'warning', title: 'Sesi Belum Siap', message: sessionReadinessIssues[0] ?? 'Lengkapi sesi ujian terlebih dahulu.' };
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
					scheduled_start: toRFC3339(fStart),
					scheduled_end: toRFC3339(fEnd),
					status: 'draft'
				})
			});
			await readClientJson<unknown>(res);
			toast.success('Sesi ujian berhasil dibuat');
			await goto(eventId ? `${resolve('/asesmen/sesi')}?event_id=${encodeURIComponent(eventId)}` : resolve('/asesmen/sesi'));
		} catch (error) {
			toast.error(errorMessage(error, 'Gagal membuat sesi ujian. Periksa koneksi lalu coba lagi.'));
		} finally {
			fBusy = false;
		}
	}

	onMount(() => {
		detectBrowserTimeZone();
		loadForm();
	});
</script>

<svelte:head><title>{eventId ? 'Buat Sesi Kegiatan Ujian' : 'Buat Sesi Ujian'} — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
	<section class="rounded-3xl border border-primary/20 bg-gradient-to-br from-primary/10 via-card to-primary/10 p-6 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
			<div class="max-w-3xl space-y-2">
				<p class="text-xs font-semibold uppercase tracking-[0.2em] text-primary">Jadwal dan Token CBT</p>
				<h1 class="text-3xl font-semibold tracking-tight text-foreground">{eventId ? 'Buat Sesi Kegiatan' : 'Buat Sesi Ujian'}</h1>
				<p class="text-sm leading-6 text-muted-foreground">Buat konsep sesi dari paket siap pakai, lalu lanjutkan ke peserta, ruang, pengawas, dan token dari daftar sesi.</p>
			</div>
			<div class="flex flex-wrap gap-2">
				{#if eventId}
					<Button href={resolve(`/asesmen/kegiatan/${eventId}`)} variant="outline">Kembali ke Kegiatan</Button>
				{/if}
				<Button href={listHref} variant="outline">Batal</Button>
			</div>
		</div>
	</section>

	{#if eventId}
		<div class="rounded-xl border border-success/20 bg-success/10 p-4 text-sm text-success">
			<div class="flex flex-wrap items-start justify-between gap-3">
				<div>
					<p class="font-semibold">Sesi untuk kegiatan: {eventContext?.title ?? eventId}</p>
					<p class="mt-1 text-success">Data pembuatan sesi otomatis tertaut ke kegiatan ini. Paket umum atau kegiatan lain disembunyikan dari pilihan sesi ini.</p>
				</div>
				<Button href={resolve(`/asesmen/paket/new?event_id=${eventId}`)} variant="outline" size="sm">Buat Paket Kegiatan</Button>
			</div>
		</div>
	{:else}
		<div class="rounded-xl border border-warning/30 bg-warning/10 px-4 py-3 text-sm text-warning">Anda sedang membuat sesi global. Dari Kegiatan CBT, gunakan tombol sesi agar pembuatan otomatis tertaut ke kegiatan.</div>
	{/if}

	{#if operationState}
		<OperationStatusPanel {...operationState} />
	{/if}

	<AsyncContent promise={formPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<Card.Root class="border-border shadow-sm">
				<Card.Content class="grid gap-3 p-6 sm:grid-cols-2">
					<Skeleton class="h-10" />
					<Skeleton class="h-10" />
					<Skeleton class="h-10 sm:col-span-2" />
					<Skeleton class="h-24 sm:col-span-2" />
				</Card.Content>
			</Card.Root>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel title="Form Sesi Belum Siap" message={errorMessage(error, 'Gagal memuat data sesi')} onRetry={() => retryForm(reset)} />
		{/snippet}

		{#snippet children()}
			<Card.Root class="border-border shadow-sm">
				<Card.Header>
					<Card.Title class="text-base">Buat Sesi Ujian Baru</Card.Title>
					<Card.Description>Form pembuatan ini memakai layanan sistem sesi dan pemeriksaan kelayakan paket dari daftar sesi lama.</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-4">
					<div class="grid gap-3 sm:grid-cols-2">
						<div>
							<label for="session-package" class="mb-1 block text-xs text-muted-foreground">Paket Soal <span class="text-destructive">*</span></label>
							<select id="session-package" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fPackageId}>
								<option value="">-- Pilih Paket --</option>
								{#each packages as pkg (pkg.id)}
									<option value={pkg.id}>{pkg.title} ({pkg.subject_code}){pkg.is_active ? '' : ' - nonaktif'}</option>
								{/each}
							</select>
							{#if hiddenEventPackageCount > 0}<p class="mt-1 text-[11px] text-accent-foreground">{hiddenEventPackageCount} paket umum/kegiatan lain disembunyikan dari pilihan sesi kegiatan ini.</p>{/if}
						</div>
						{#if fPackageId}
							{@const quality = selectedPackageQuality}
							<div class="rounded-lg border border-border bg-muted/50 px-3 py-2 sm:col-span-2">
								<div class="flex flex-wrap items-start justify-between gap-3">
									<div><p class="text-xs font-semibold uppercase tracking-[0.16em] text-foreground">Quality Gate Paket</p><p class="mt-1 text-sm font-medium text-foreground">{selectedPackage?.title ?? 'Paket dipilih'}</p></div>
									<div class="flex flex-wrap gap-1.5">
										<Badge variant="outline" class="bg-card text-xs">{quality.totalCount} soal</Badge>
										{#each quality.typeBuckets.slice(0, 3) as bucket (bucket.label)}<Badge variant="outline" class="bg-card text-xs">{bucket.label}: {bucket.count}</Badge>{/each}
										{#if quality.hotsCount > 0}<Badge class="border-warning/30 bg-warning/10 text-xs text-warning">{quality.hotsCount} HOTS</Badge>{/if}
										{#if quality.missingCount > 0}<Badge class="border-warning/30 bg-warning/10 text-xs text-warning">{quality.missingCount} metadata kurang</Badge>{/if}
										{#if quality.unpublishedCount > 0}<Badge class="border-destructive/30 bg-destructive/10 text-xs text-destructive">{quality.unpublishedCount} belum terbit</Badge>{/if}
									</div>
								</div>
								{#if selectedPackage && !selectedPackage.is_active}<p class="mt-2 text-xs font-medium text-destructive">Paket nonaktif tidak boleh dijadikan sesi ujian.</p>{:else if quality.totalCount === 0}<p class="mt-2 text-xs font-medium text-destructive">Paket ini belum memiliki soal, sehingga sesi tidak bisa dibuat.</p>{:else if quality.unpublishedCount > 0}<p class="mt-2 text-xs font-medium text-destructive">Rapikan paket dulu. Flutter hanya menyajikan soal terbit.</p>{:else if quality.missingCount > 0}<p class="mt-2 text-xs font-medium text-warning">Sesi masih boleh dibuat, tetapi {quality.missingCount} soal belum lengkap CP/TP/KD atau level kognitif.</p>{:else}<p class="mt-2 text-xs font-medium text-primary">Paket siap dipakai untuk draft sesi CBT.</p>{/if}
							</div>
						{/if}
						<div>
							<label for="session-scope" class="mb-1 block text-xs text-muted-foreground">Cakupan peserta <span class="text-destructive">*</span></label>
							<select id="session-scope" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" value={fScopeType} onchange={(event) => updateScopeType(event.currentTarget.value)}>
								<option value="class">Per kelas</option>
								<option value="grade">Per tingkat</option>
								<option value="school">Seluruh sekolah</option>
							</select>
						</div>
						{#if fScopeType === 'class'}
							<div>
								<label for="session-class" class="mb-1 block text-xs text-muted-foreground">Kelas <span class="text-destructive">*</span></label>
								<select id="session-class" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fClassId}>
									<option value="">-- Pilih Kelas --</option>
									{#each classes as schoolClass (schoolClass.id)}<option value={schoolClass.id}>{schoolClass.code} — {schoolClass.name}</option>{/each}
								</select>
							</div>
						{:else if fScopeType === 'grade'}
							<div>
								<label for="session-grade" class="mb-1 block text-xs text-muted-foreground">Tingkat <span class="text-destructive">*</span></label>
								<select id="session-grade" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fGradeLevel}><option value="VII">VII</option><option value="VIII">VIII</option><option value="IX">IX</option></select>
							</div>
						{:else}
							<div class="rounded-md border border-primary/20 bg-primary/10 px-3 py-2 text-sm text-primary">Semua siswa aktif di sekolah dapat menjadi peserta sesi ini.</div>
						{/if}
						<div>
							<label for="session-mix-policy" class="mb-1 block text-xs text-muted-foreground">Mix policy</label>
							<select id="session-mix-policy" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fMixPolicy}><option value="same_class">Tetap per kelas</option><option value="same_grade">Campur dalam tingkat</option><option value="mixed_scope">Campur lintas cakupan</option></select>
							<p class="mt-1 text-[11px] text-muted-foreground">Saat cakupan berubah, opsi disetel otomatis lalu tetap bisa disesuaikan operator.</p>
						</div>
						<div>
							<label for="session-assignment-mode" class="mb-1 block text-xs text-muted-foreground">Mode alokasi ruangan</label>
							<select id="session-assignment-mode" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fAssignmentMode}><option value="random_balanced">Acak seimbang</option><option value="manual">Manual</option><option value="random_by_gender">Acak per gender</option><option value="random_by_accommodation">Acak akomodasi khusus</option></select>
						</div>
						<div class="sm:col-span-2">
							<label for="session-title" class="mb-1 block text-xs text-muted-foreground">Nama Sesi <span class="text-destructive">*</span></label>
							<Input id="session-title" placeholder="mis: UTS Matematika VII A - Semester 1 2025" bind:value={fTitle} />
						</div>
						<div><label for="session-start" class="mb-1 block text-xs text-muted-foreground">Mulai <span class="text-destructive">*</span></label><Input id="session-start" type="datetime-local" bind:value={fStart} /></div>
						<div><label for="session-end" class="mb-1 block text-xs text-muted-foreground">Selesai <span class="text-destructive">*</span></label><Input id="session-end" type="datetime-local" bind:value={fEnd} /></div>
						<div class="rounded-md border border-primary/20 bg-primary/10 px-3 py-2 text-xs leading-5 text-primary sm:col-span-2">
							<p class="font-semibold">Jadwal sesi dicatat dan ditampilkan sebagai WITA (Asia/Makassar).</p>
							{#if browserTimeZoneMismatch}<p class="text-warning">Zona waktu browser terdeteksi {browserTimeZone}. Samakan perangkat operator ke Asia/Makassar sebelum menyimpan agar input tidak bergeser.</p>{:else}<p>Pastikan jam mulai dan selesai mengikuti waktu sekolah/WITA sebelum sesi dijadwalkan.</p>{/if}
						</div>
						<div class="grid gap-3 sm:col-span-2 sm:grid-cols-2">
							<label class="flex items-center gap-2 rounded-md border border-input px-3 py-2 text-sm text-foreground"><input type="checkbox" bind:checked={fIsSpecialEvent} class="size-4 accent-primary" /> Tandai sebagai sesi khusus</label>
							<label class="flex items-center gap-2 rounded-md border border-input px-3 py-2 text-sm text-foreground"><input type="checkbox" bind:checked={fAllowCrossGrade} class="size-4 accent-primary" /> Izinkan lintas tingkat</label>
						</div>
					</div>

					<div class="flex flex-wrap gap-2">
						<LoadingButton disabled={!canCreateSession} onclick={() => void createSession()} loading={fBusy} loadingLabel="Menyimpan...">Buat Sesi</LoadingButton>
						<Button href={listHref} variant="outline">Batal</Button>
					</div>
					{#if sessionReadinessIssues.length > 0}<div class="rounded-md border border-warning/30 bg-warning/10 px-3 py-2 text-xs text-warning"><span class="font-semibold">Belum siap dibuat:</span> {sessionReadinessIssues.join(', ')}</div>{/if}
				</Card.Content>
			</Card.Root>
		{/snippet}
	</AsyncContent>
</div>
