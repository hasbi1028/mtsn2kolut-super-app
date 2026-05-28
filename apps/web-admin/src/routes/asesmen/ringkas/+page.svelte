<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import {
		countRunningSessions,
		countUnassignedParticipants,
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
	let userRoles = $derived(page.data.user?.roles ?? (page.data.user?.role ? [page.data.user.role] : []));
	let userPermissions = $derived((page.data.user?.permissions ?? []).map((permission) => permission.trim()).filter(Boolean));
	let canOpenResults = $derived(userRoles.includes('admin') || userPermissions.includes('asesmen.result_read'));
	let documentHubHref = $derived(latestEventDocumentHubHref(sessions));
	let workflowReadiness = $derived(summarizeWorkflowReadiness(sessions));

	onMount(() => {
		void loadData();
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
	<title>Ringkasan Ujian — MTsN 2 Kolaka Utara</title>
</svelte:head>

<main class="min-h-dvh bg-slate-50 px-4 py-5 text-slate-950 md:px-6">
	<section class="mx-auto max-w-6xl space-y-4">
		<header class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
			<div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
				<div>
					<p class="text-xs font-bold uppercase tracking-[0.22em] text-emerald-700">Ujian Digital</p>
					<h1 class="mt-1 text-2xl font-black tracking-tight md:text-3xl">Ringkasan Ujian</h1>
					<p class="mt-1 max-w-2xl text-sm text-slate-600">Pusat kerja panitia: mulai dari persiapan, cetak dokumen, pelaksanaan ruang, lalu hasil.</p>
				</div>
				<div class="flex flex-wrap gap-2 text-sm font-bold">
					<a class="rounded-xl bg-emerald-700 px-3 py-2 text-white" href="/asesmen/persiapan">Persiapan</a>
					<a class="rounded-xl border border-slate-300 bg-white px-3 py-2" href={documentHubHref}>Dokumen & Cetak</a>
					<a class="rounded-xl border border-slate-300 bg-white px-3 py-2" href="/asesmen/pelaksanaan">Pelaksanaan</a>
					{#if canOpenResults}
						<a class="rounded-xl border border-slate-300 bg-white px-3 py-2" href="/asesmen/hasil">Hasil</a>
					{/if}
				</div>
			</div>
		</header>

		{#if errorMessage}<div class="rounded-2xl border border-red-200 bg-red-50 p-3 text-sm font-semibold text-red-800">{errorMessage}</div>{/if}
		{#if loading}
			<div class="rounded-2xl border border-slate-200 bg-white p-6 text-sm text-slate-600">Memuat ringkasan ujian...</div>
		{:else}
			<section class="grid gap-3 md:grid-cols-4">
				<div class="rounded-2xl border border-slate-200 bg-white p-4">
					<p class="text-xs font-bold uppercase tracking-[0.18em] text-slate-500">Persiapan</p>
					<p class="mt-2 text-3xl font-black">{sessions.length}</p>
					<p class="text-sm text-slate-600">sesi ujian tercatat</p>
				</div>
				<div class="rounded-2xl border border-slate-200 bg-white p-4">
					<p class="text-xs font-bold uppercase tracking-[0.18em] text-slate-500">Pelaksanaan</p>
					<p class="mt-2 text-3xl font-black">{runningSessions}</p>
					<p class="text-sm text-slate-600">sesi sedang berjalan</p>
				</div>
				<div class="rounded-2xl border border-slate-200 bg-white p-4">
					<p class="text-xs font-bold uppercase tracking-[0.18em] text-slate-500">Kesiapan</p>
					<p class="mt-2 text-3xl font-black">{unassignedParticipantCount}</p>
					<p class="text-sm text-slate-600">peserta belum ditempatkan</p>
				</div>
				<div class="rounded-2xl border border-slate-200 bg-white p-4">
					<p class="text-xs font-bold uppercase tracking-[0.18em] text-slate-500">Dokumen</p>
					<p class="mt-2 text-3xl font-black">Pusat</p>
					<p class="text-sm text-slate-600">kartu, lembar pengawas, arsip</p>
				</div>
			</section>

			<section class="grid gap-4 lg:grid-cols-[1.1fr_0.9fr]">
				<div class="space-y-4">
					<div class="rounded-2xl border border-slate-200 bg-white p-4">
						<div class="flex items-center justify-between gap-3">
							<div>
								<h2 class="text-lg font-black">Sesi terdekat</h2>
								<p class="text-sm text-slate-600">Pantau sesi terdekat tanpa masuk ke halaman teknis kecuali diperlukan.</p>
							</div>
							<a class="rounded-xl border border-slate-300 px-3 py-2 text-sm font-bold" href="/asesmen/pelaksanaan">Pelaksanaan</a>
						</div>
						<div class="mt-3 divide-y divide-slate-100">
							{#each latestSessions as session (session.id)}
								<article class="flex items-center justify-between gap-3 py-3">
									<div class="min-w-0">
										<p class="truncate font-bold">{session.title ?? 'Sesi Ujian'}</p>
										<p class="text-xs text-slate-500">{session.package_title ?? 'Paket ujian'} · {fmtDate(session.scheduled_start)}</p>
									</div>
									<span class="shrink-0 rounded-full bg-slate-100 px-2 py-1 text-[11px] font-bold">{statusLabel(session.status ?? session.session_status)}</span>
								</article>
							{/each}
							{#if latestSessions.length === 0}<p class="py-4 text-sm text-slate-500">Belum ada sesi ujian.</p>{/if}
						</div>
					</div>

					<div class="rounded-2xl border border-slate-200 bg-white p-4">
						<h2 class="text-lg font-black">Paket siap ujian</h2>
						<div class="mt-3 divide-y divide-slate-100">
							{#each latestPackages as pkg (pkg.id)}
								<article class="flex items-center justify-between gap-3 py-3">
									<div class="min-w-0">
										<p class="truncate font-bold">{pkg.title ?? 'Paket Ujian'}</p>
										<p class="text-xs text-slate-500">{pkg.subject ?? 'Mapel'} · {pkg.level ?? 'Tingkat'} · {pkg.question_count ?? 0} soal</p>
									</div>
									<span class="shrink-0 rounded-full bg-slate-100 px-2 py-1 text-[11px] font-bold">Siap dipilih</span>
								</article>
							{/each}
							{#if latestPackages.length === 0}<p class="py-4 text-sm text-slate-500">Belum ada paket ujian.</p>{/if}
						</div>
					</div>
				</div>

				<aside class="rounded-2xl border border-emerald-200 bg-white p-4 shadow-sm">
					<p class="text-xs font-bold uppercase tracking-[0.18em] text-emerald-700">Status alur</p>
					<div class="mt-2 flex flex-wrap items-center gap-2">
						<h2 class="text-xl font-black">Arah kerja berikutnya</h2>
						<span class={`rounded-full px-2.5 py-1 text-xs font-bold ${readinessClass()}`}>{readinessLabel()}</span>
					</div>
					<p class="mt-2 text-sm leading-6 text-slate-600">
						Ringkasan ini tidak menyimpan perubahan. Pembagian ruang, peserta, token, dan status sesi dikerjakan dari Persiapan atau Mode Lengkap agar keputusan teknis tidak tersebar.
					</p>

					<div class="mt-4 grid gap-2">
						<a class="rounded-xl bg-emerald-700 px-3 py-3 text-center text-sm font-black text-white" href="/asesmen/persiapan">1. Persiapan</a>
						<a class="rounded-xl border border-slate-300 bg-white px-3 py-3 text-center text-sm font-bold text-slate-900" href={documentHubHref}>2. Dokumen & Cetak</a>
						<a class="rounded-xl border border-slate-300 bg-white px-3 py-3 text-center text-sm font-bold text-slate-900" href="/asesmen/pelaksanaan">3. Pelaksanaan Ujian</a>
						<a class="rounded-xl border border-slate-300 bg-white px-3 py-3 text-center text-sm font-bold text-slate-900" href="/asesmen/panitia">Mode Lengkap Panitia</a>
					</div>

					<div class="mt-4 rounded-xl border border-slate-200 bg-slate-50 p-3 text-xs leading-5 text-slate-600">
						<p class="font-bold text-slate-900">Batas sederhana:</p>
						<p>Bank Soal untuk menyusun soal. Asesmen cukup untuk menyiapkan kegiatan, mencetak dokumen, menjalankan ruang, dan menutup hasil.</p>
					</div>
				</aside>
			</section>
		{/if}
	</section>
</main>
