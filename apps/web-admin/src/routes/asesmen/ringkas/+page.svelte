<script lang="ts">
	import { onMount } from 'svelte';

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
	};
	type PackageRow = { id: string; title?: string; subject?: string; level?: string; question_count?: number; status?: string };
	type AssignmentPreview = {
		summary: {
			participant_count: number;
			room_count: number;
			capacity_total: number;
			assigned_count: number;
			unassigned_count: number;
			mix_policy: string;
			assignment_mode: string;
			allow_cross_grade: boolean;
			is_special_event: boolean;
		};
		rooms: Array<{ room_id: string; room_name: string; capacity: number; participant_count: number; levels: Record<string, number>; classes: string[] }>;
		warnings?: string[];
	};

	let loading = $state(true);
	let working = $state(false);
	let errorMessage = $state('');
	let successMessage = $state('');
	let sessions = $state<SessionRow[]>([]);
	let packages = $state<PackageRow[]>([]);
	let selectedSessionId = $state('');
	let mixPolicy = $state<'same_class' | 'same_grade' | 'mixed_scope'>('mixed_scope');
	let preview = $state<AssignmentPreview | null>(null);

	let selectedSession = $derived(sessions.find((session) => session.id === selectedSessionId) ?? sessions[0]);
	let latestSessions = $derived(sessions.slice(0, 5));
	let latestPackages = $derived(packages.slice(0, 5));
	let todaySessions = $derived(sessions.filter((session) => isToday(session.scheduled_start)).slice(0, 5));
	let assignmentPayload = $derived({
		mix_policy: mixPolicy,
		assignment_mode: 'random_balanced',
		allow_cross_grade: mixPolicy === 'mixed_scope',
		is_special_event: mixPolicy === 'mixed_scope'
	});

	onMount(() => {
		void loadData();
	});

	async function loadData() {
		loading = true;
		errorMessage = '';
		successMessage = '';
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
			selectedSessionId = sessions[0]?.id ?? '';
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

	async function previewRooms() {
		if (!selectedSession?.id) {
			errorMessage = 'Pilih sesi ujian terlebih dahulu.';
			return;
		}
		working = true;
		errorMessage = '';
		successMessage = '';
		try {
			preview = await fetchWithTimeout(`/api/asesmen/sessions/${encodeURIComponent(selectedSession.id)}/rooms/assignment-preview`, {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify(assignmentPayload)
			}).then((response) => readJson<AssignmentPreview>(response));
			successMessage = 'Pratinjau pembagian ruang siap diperiksa.';
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Pratinjau pembagian ruang gagal.';
		} finally {
			working = false;
		}
	}

	async function applyRooms() {
		if (!selectedSession?.id || !preview) return;
		const ok = window.confirm('Simpan pembagian ruang dan nomor kursi sesuai pratinjau ini? Pembagian lama pada sesi ini akan diganti.');
		if (!ok) return;
		working = true;
		errorMessage = '';
		successMessage = '';
		try {
			preview = await fetchWithTimeout(`/api/asesmen/sessions/${encodeURIComponent(selectedSession.id)}/rooms/assignment`, {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify(assignmentPayload)
			}).then((response) => readJson<AssignmentPreview>(response));
			successMessage = 'Pembagian ruang dan nomor kursi berhasil disimpan.';
			await loadData();
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Pembagian ruang gagal disimpan.';
		} finally {
			working = false;
		}
	}

	function isToday(value?: string) {
		if (!value) return false;
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return false;
		const now = new Date();
		return date.toDateString() === now.toDateString();
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

	function modeLabel(value: string) {
		if (value === 'same_class') return 'Per Kelas';
		if (value === 'same_grade') return 'Campur Satu Tingkat';
		return 'Campur Lintas Tingkat';
	}
</script>

<svelte:head>
	<title>Asesmen Ringkas — MTsN 2 Kolaka Utara</title>
</svelte:head>

<main class="min-h-dvh bg-slate-50 px-4 py-5 text-slate-950 md:px-6">
	<section class="mx-auto max-w-6xl space-y-4">
		<header class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
			<div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
				<div>
					<p class="text-xs font-bold uppercase tracking-[0.22em] text-emerald-700">Asesmen/Ujian Digital</p>
					<h1 class="mt-1 text-2xl font-black tracking-tight md:text-3xl">Asesmen Ringkas</h1>
					<p class="mt-1 max-w-2xl text-sm text-slate-600">Satu halaman untuk melihat persiapan, hari ujian, hasil, dan pembagian 8 ruang tanpa masuk banyak menu.</p>
				</div>
				<div class="flex flex-wrap gap-2 text-sm font-bold">
					<a class="rounded-xl bg-emerald-700 px-3 py-2 text-white" href="/asesmen/persiapan">Persiapan</a>
					<a class="rounded-xl border border-slate-300 bg-white px-3 py-2" href="/asesmen">Mode Lengkap</a>
				</div>
			</div>
		</header>

		{#if errorMessage}<div class="rounded-2xl border border-red-200 bg-red-50 p-3 text-sm font-semibold text-red-800">{errorMessage}</div>{/if}
		{#if successMessage}<div class="rounded-2xl border border-emerald-200 bg-emerald-50 p-3 text-sm font-semibold text-emerald-800">{successMessage}</div>{/if}

		{#if loading}
			<div class="rounded-2xl border border-slate-200 bg-white p-6 text-sm text-slate-600">Memuat ringkasan asesmen...</div>
		{:else}
			<section class="grid gap-3 md:grid-cols-3">
				<div class="rounded-2xl border border-slate-200 bg-white p-4">
					<p class="text-xs font-bold uppercase tracking-[0.18em] text-slate-500">Persiapan</p>
					<p class="mt-2 text-3xl font-black">{sessions.length}</p>
					<p class="text-sm text-slate-600">sesi ujian tercatat</p>
				</div>
				<div class="rounded-2xl border border-slate-200 bg-white p-4">
					<p class="text-xs font-bold uppercase tracking-[0.18em] text-slate-500">Hari Ujian</p>
					<p class="mt-2 text-3xl font-black">{todaySessions.length}</p>
					<p class="text-sm text-slate-600">sesi hari ini</p>
				</div>
				<div class="rounded-2xl border border-slate-200 bg-white p-4">
					<p class="text-xs font-bold uppercase tracking-[0.18em] text-slate-500">Paket</p>
					<p class="mt-2 text-3xl font-black">{packages.length}</p>
					<p class="text-sm text-slate-600">paket ujian tersedia</p>
				</div>
			</section>

			<section class="grid gap-4 lg:grid-cols-[1.1fr_0.9fr]">
				<div class="space-y-4">
					<div class="rounded-2xl border border-slate-200 bg-white p-4">
						<div class="flex items-center justify-between gap-3">
							<div>
								<h2 class="text-lg font-black">Lanjutkan pekerjaan</h2>
								<p class="text-sm text-slate-600">Mulai dari sesi terdekat, lalu atur ruang dan cetak kartu.</p>
							</div>
							<a class="rounded-xl border border-slate-300 px-3 py-2 text-sm font-bold" href="/asesmen/sesi">Lihat Semua</a>
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
						<h2 class="text-lg font-black">Paket terbaru</h2>
						<div class="mt-3 divide-y divide-slate-100">
							{#each latestPackages as pkg (pkg.id)}
								<article class="flex items-center justify-between gap-3 py-3">
									<div class="min-w-0">
										<p class="truncate font-bold">{pkg.title ?? 'Paket Ujian'}</p>
										<p class="text-xs text-slate-500">{pkg.subject ?? 'Mapel'} · {pkg.level ?? 'Tingkat'} · {pkg.question_count ?? 0} soal</p>
									</div>
									<a class="shrink-0 rounded-xl border border-slate-300 px-3 py-2 text-xs font-bold" href={`/asesmen/paket/${pkg.id}`}>Buka</a>
								</article>
							{/each}
							{#if latestPackages.length === 0}<p class="py-4 text-sm text-slate-500">Belum ada paket ujian.</p>{/if}
						</div>
					</div>
				</div>

				<aside class="rounded-2xl border border-emerald-200 bg-white p-4 shadow-sm">
					<p class="text-xs font-bold uppercase tracking-[0.18em] text-emerald-700">Jadwal & Ruang</p>
					<h2 class="mt-1 text-xl font-black">Pembagian Peserta 8 Ruang</h2>
					<p class="mt-1 text-sm text-slate-600">Pilih sesi, mode campur, pratinjau, lalu simpan pembagian ruang dan nomor kursi.</p>

					<label class="mt-4 block space-y-1 text-sm font-bold">
						<span>Sesi ujian</span>
						<select class="w-full rounded-xl border border-slate-300 bg-white px-3 py-2" bind:value={selectedSessionId} onchange={() => (preview = null)}>
							{#each sessions as session (session.id)}<option value={session.id}>{session.title ?? 'Sesi Ujian'} · {fmtDate(session.scheduled_start)}</option>{/each}
						</select>
					</label>

					<div class="mt-4 grid gap-2">
						<p class="text-sm font-bold">Mode pembagian</p>
						{#each [
							{ value: 'same_class', title: 'Per Kelas', desc: 'Satu ruang berisi satu rombel.' },
							{ value: 'same_grade', title: 'Campur Satu Tingkat', desc: 'Rombel boleh bercampur, tingkat tetap dipisah.' },
							{ value: 'mixed_scope', title: 'Campur Lintas Tingkat', desc: 'VII, VIII, IX dapat bercampur. Sesi khusus.' }
						] as option}
							<label class="flex gap-3 rounded-xl border p-3 text-sm {mixPolicy === option.value ? 'border-emerald-500 bg-emerald-50' : 'border-slate-200'}">
								<input type="radio" bind:group={mixPolicy} value={option.value} onchange={() => (preview = null)} />
								<span><b>{option.title}</b><br /><span class="text-xs text-slate-600">{option.desc}</span></span>
							</label>
						{/each}
					</div>

					{#if mixPolicy === 'mixed_scope'}
						<div class="mt-4 rounded-xl border border-amber-200 bg-amber-50 p-3 text-xs font-semibold leading-5 text-amber-950">Campur lintas tingkat otomatis dikirim sebagai sesi khusus dan izin campur tingkat. Periksa pratinjau sebelum menyimpan.</div>
					{/if}

					<div class="mt-4 grid grid-cols-2 gap-2">
						<button class="min-h-12 rounded-xl bg-emerald-700 px-3 text-sm font-black text-white disabled:opacity-60" disabled={working || !selectedSession} onclick={previewRooms}>{working ? 'Memproses...' : 'Pratinjau'}</button>
						<button class="min-h-12 rounded-xl bg-slate-900 px-3 text-sm font-black text-white disabled:opacity-50" disabled={working || !preview || (preview.summary.unassigned_count ?? 0) > 0} onclick={applyRooms}>Acak & Simpan</button>
					</div>

					{#if preview}
						<div class="mt-4 rounded-2xl border border-slate-200 bg-slate-50 p-3">
							<p class="text-sm font-black">{modeLabel(preview.summary.mix_policy)} · {preview.summary.assigned_count}/{preview.summary.participant_count} peserta</p>
							<p class="mt-1 text-xs text-slate-600">{preview.summary.room_count} ruang · kapasitas {preview.summary.capacity_total} kursi · belum ditempatkan {preview.summary.unassigned_count}</p>
							{#if preview.warnings?.length}<ul class="mt-2 space-y-1 text-xs font-semibold text-amber-800">{#each preview.warnings as warning}<li>• {warning}</li>{/each}</ul>{/if}
							<div class="mt-3 grid gap-2 sm:grid-cols-2">
								{#each preview.rooms as room (room.room_id)}
									<div class="rounded-xl border border-slate-200 bg-white p-3 text-xs">
										<p class="font-black">{room.room_name}</p>
										<p class="text-slate-600">{room.participant_count}/{room.capacity} peserta</p>
										<p class="mt-1 text-slate-500">{Object.entries(room.levels).map(([k, v]) => `${k}: ${v}`).join(', ') || 'Kosong'}</p>
									</div>
								{/each}
							</div>
						</div>
					{/if}
				</aside>
			</section>
		{/if}
	</section>
</main>
