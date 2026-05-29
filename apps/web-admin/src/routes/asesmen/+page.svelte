<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { AssessmentPhaseHeader, AssessmentTaskCard } from '$lib/components/asesmen';
	import {
		buildAsesmenWorkflowPhaseCards,
		countRunningSessions,
		countUnassignedParticipants,
		deriveAsesmenWorkflowAccess,
		latestEventDocumentHubHref,
		summarizeWorkflowReadiness,
		workflowReadinessClass
	} from '$lib/asesmen/workflow-hub';

	type ApiEnvelope<T> = { data?: T; items?: T; error?: string; message?: string } | T;
	type SessionRow = {
		id: string;
		title?: string;
		status?: string;
		session_status?: string;
		scheduled_start?: string;
		scheduled_end?: string;
		package_title?: string;
		room_count?: number;
		participant_count?: number;
		unassigned_participant_count?: number;
		event_id?: string;
		event_title?: string;
	};
	type PackageRow = { id: string; title?: string; subject?: string; level?: string; question_count?: number; status?: string };

	let loading = $state(true);
	let errorMessage = $state('');
	let sessions = $state<SessionRow[]>([]);
	let packages = $state<PackageRow[]>([]);

	let latestSessions = $derived(sessions.slice(0, 5));
	let latestPackages = $derived(packages.slice(0, 5));
	let unassignedParticipantCount = $derived(countUnassignedParticipants(sessions));
	let runningSessions = $derived(countRunningSessions(sessions));
	let workflowAccess = $derived(deriveAsesmenWorkflowAccess(page.data.user));
	let canOpenPreparation = $derived(workflowAccess.canOpenPreparation);
	let canLoadDashboardStats = $derived(workflowAccess.canLoadDashboardStats);
	let documentHubHref = $derived(latestEventDocumentHubHref(sessions));
	let workflowReadiness = $derived(summarizeWorkflowReadiness(sessions));
	let phaseCards = $derived(buildAsesmenWorkflowPhaseCards(workflowAccess, documentHubHref, resolve));

	onMount(() => {
		if (canLoadDashboardStats) {
			void loadData();
		} else {
			loading = false;
		}
	});

	async function loadData() {
		loading = true;
		errorMessage = '';
		try {
			const [sessionResult, packageResult] = await Promise.allSettled([
				fetchWithTimeout('/api/asesmen/sessions').then((response) => readJson<SessionRow[]>(response)),
				fetchWithTimeout('/api/asesmen/packages').then((response) => readJson<PackageRow[]>(response))
			]);
			const errors: string[] = [];
			if (sessionResult.status === 'fulfilled') {
				sessions = Array.isArray(sessionResult.value) ? sessionResult.value : [];
			} else {
				sessions = [];
				errors.push(`Sesi: ${friendlyLoadError(sessionResult.reason)}`);
			}
			if (packageResult.status === 'fulfilled') {
				packages = Array.isArray(packageResult.value) ? packageResult.value : [];
			} else {
				packages = [];
				errors.push(`Paket: ${friendlyLoadError(packageResult.reason)}`);
			}
			if (errors.length > 0) {
				errorMessage = `Sebagian data belum termuat. ${errors.join(' · ')}`;
			}
		} finally {
			loading = false;
		}
	}

	async function fetchWithTimeout(url: string, init: RequestInit = {}, timeoutMs = 10000) {
		const controller = new AbortController();
		const timer = window.setTimeout(() => controller.abort(), timeoutMs);
		try {
			return await fetch(url, { ...init, signal: controller.signal });
		} finally {
			window.clearTimeout(timer);
		}
	}

	function friendlyLoadError(error: unknown) {
		if (error instanceof DOMException && error.name === 'AbortError') return 'koneksi terlalu lama, coba muat ulang';
		return error instanceof Error ? error.message : 'gagal dimuat';
	}

	async function readJson<T>(response: Response): Promise<T> {
		const body = (await response.json().catch(() => ({}))) as ApiEnvelope<T>;
		if (!response.ok) {
			const message = (body as { error?: string; message?: string }).error ?? (body as { message?: string }).message ?? 'Permintaan gagal.';
			throw new Error(message);
		}
		if (body && typeof body === 'object' && 'data' in body) return (body as { data: T }).data;
		if (body && typeof body === 'object' && 'items' in body) return (body as { items: T }).items;
		return body as T;
	}

	function fmtDate(value?: string) {
		if (!value) return 'Belum dijadwalkan';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value;
		return date.toLocaleString('id-ID', { timeZone: 'Asia/Makassar', day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' }) + ' WITA';
	}

	function statusLabel(value?: string) {
		if (value === 'active') return 'Berjalan';
		if (value === 'scheduled') return 'Terjadwal';
		if (value === 'finished') return 'Selesai';
		if (value === 'cancelled') return 'Dibatalkan';
		return 'Draft';
	}

	function readinessLabel() {
		return workflowReadiness.label;
	}

	function readinessClass() {
		return workflowReadinessClass(workflowReadiness.tone);
	}
</script>

<svelte:head>
	<title>Command Center CBT — MTsN 2 Kolaka Utara</title>
</svelte:head>

<main class="min-h-dvh bg-background px-4 py-5 text-foreground md:px-6">
	<section class="mx-auto max-w-6xl space-y-4">
		<AssessmentPhaseHeader
			code="7.0"
			badge="CBT Web"
			title="Command Center CBT"
			description="Alur sederhana untuk panitia: siapkan ujian, jalankan ruang, buka portal peserta, lalu tutup hasil."
			primaryAction={canOpenPreparation ? { label: 'Mulai Persiapan', href: resolve('/asesmen/persiapan') } : undefined}
			secondaryActions={[{ label: 'Portal Peserta', href: resolve('/asesmen/aplikasi-siswa'), variant: 'outline' }]}
		/>

		{#if errorMessage}<div class="rounded-2xl border border-destructive/25 bg-destructive/10 p-3 text-sm font-semibold text-destructive">{errorMessage}</div>{/if}
		{#if loading}
			<div class="rounded-2xl border border-border bg-card p-6 text-sm text-muted-foreground">Memuat ringkasan ujian...</div>
		{:else}
			<section aria-labelledby="asesmen-phase-title" class="space-y-3">
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.22em] text-primary">Alur utama</p>
					<h2 id="asesmen-phase-title" class="mt-1 text-xl font-semibold tracking-tight text-foreground">Jalur kerja panitia</h2>
				</div>
				<div class="grid gap-3 lg:grid-cols-2">
					{#each phaseCards as item (item.code)}
						<AssessmentTaskCard code={item.code} title={item.title} description={item.description} href={item.href} cta={item.cta} tone={item.tone ?? 'default'} />
					{/each}
				</div>
			</section>

			<section class="grid gap-3 md:grid-cols-3">
				<div class="rounded-2xl border border-border bg-card p-4">
					<p class="text-xs font-bold uppercase tracking-[0.18em] text-muted-foreground">Sesi CBT</p>
					<p class="mt-2 text-3xl font-black">{sessions.length}</p>
					<p class="text-sm text-muted-foreground">sesi tercatat</p>
				</div>
				<div class="rounded-2xl border border-border bg-card p-4">
					<p class="text-xs font-bold uppercase tracking-[0.18em] text-muted-foreground">Berjalan</p>
					<p class="mt-2 text-3xl font-black">{runningSessions}</p>
					<p class="text-sm text-muted-foreground">sesi sedang berjalan</p>
				</div>
				<div class="rounded-2xl border border-border bg-card p-4">
					<p class="text-xs font-bold uppercase tracking-[0.18em] text-muted-foreground">Belum ditempatkan</p>
					<p class="mt-2 text-3xl font-black">{unassignedParticipantCount}</p>
					<p class="text-sm text-muted-foreground">peserta perlu ruang</p>
				</div>
			</section>

			<section class="grid gap-4 lg:grid-cols-[1.1fr_0.9fr]">
				<div class="space-y-4">
					<div class="rounded-2xl border border-border bg-card p-4">
						<div class="flex items-center justify-between gap-3">
							<div>
								<h2 class="text-lg font-black">Sesi CBT terdekat</h2>
								<p class="text-sm text-muted-foreground">Pantau jadwal dan status ruang tanpa membuka panel teknis kecuali diperlukan.</p>
							</div>
							</div>
						<div class="mt-3 divide-y divide-border">
							{#each latestSessions as session (session.id)}
								<article class="flex items-center justify-between gap-3 py-3">
									<div class="min-w-0">
										<p class="truncate font-bold">{session.title ?? 'Sesi Ujian'}</p>
										<p class="text-xs text-muted-foreground">{session.package_title ?? 'Paket ujian'} · {fmtDate(session.scheduled_start)}</p>
									</div>
									<span class="shrink-0 rounded-full bg-muted px-2 py-1 text-[11px] font-bold text-muted-foreground">{statusLabel(session.status ?? session.session_status)}</span>
								</article>
							{/each}
							{#if latestSessions.length === 0}<p class="py-4 text-sm text-muted-foreground">Belum ada sesi ujian.</p>{/if}
						</div>
					</div>

					<div class="rounded-2xl border border-border bg-card p-4">
						<h2 class="text-lg font-black">Paket siap dipakai</h2>
						<div class="mt-3 divide-y divide-border">
							{#each latestPackages as pkg (pkg.id)}
								<article class="flex items-center justify-between gap-3 py-3">
									<div class="min-w-0">
										<p class="truncate font-bold">{pkg.title ?? 'Paket Ujian'}</p>
										<p class="text-xs text-muted-foreground">{pkg.subject ?? 'Mapel'} · {pkg.level ?? 'Tingkat'} · {pkg.question_count ?? 0} soal</p>
									</div>
									<span class="shrink-0 rounded-full bg-muted px-2 py-1 text-[11px] font-bold text-muted-foreground">Siap dipilih</span>
								</article>
							{/each}
							{#if latestPackages.length === 0}<p class="py-4 text-sm text-muted-foreground">Belum ada paket ujian.</p>{/if}
						</div>
					</div>
				</div>

				<aside class="rounded-2xl border border-primary/25 bg-card p-4 shadow-sm">
					<p class="text-xs font-bold uppercase tracking-[0.18em] text-primary">Ringkas</p>
					<div class="mt-2 flex flex-wrap items-center gap-2">
						<h2 class="text-xl font-black">Apa yang perlu dilakukan</h2>
						<span class={`rounded-full px-2.5 py-1 text-xs font-bold ${readinessClass()}`}>{readinessLabel()}</span>
					</div>
					<p class="mt-2 text-sm leading-6 text-muted-foreground">
						Gunakan halaman ini sebagai pintu utama CBT. Menu teknis seperti kegiatan, paket, sesi, kartu, dan arsip tetap ada di halaman detail, bukan di sidebar harian.
					</p>
					<p class="mt-4 rounded-xl border border-border bg-muted/40 px-3 py-2 text-xs leading-5 text-muted-foreground">
						Bank Soal tetap untuk menyusun soal. CBT web dipakai untuk portal peserta, pengawasan ruang, dokumen, dan hasil.
					</p>

				</aside>
			</section>
		{/if}
	</section>
</main>
