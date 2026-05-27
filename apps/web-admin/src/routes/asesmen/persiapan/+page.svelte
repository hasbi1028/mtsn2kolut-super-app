<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { MicroActionTable } from '$lib/components/ops';

	type PersiapanRoute =
		| '/asesmen/paket'
		| '/asesmen/paket/new'
		| '/asesmen/kegiatan'
		| '/asesmen/kegiatan/new'
		| '/asesmen/sesi'
		| '/asesmen/sesi/new'
		| '/asesmen/pelaksanaan'
		| '/asesmen/hasil'
		| '/asesmen/panitia'
		| '/asesmen';

	type PreparationTask = {
		step: string;
		title: string;
		description: string;
		href: PersiapanRoute;
		cta: string;
	};
	type ApiEnvelope<T> = { data?: T; items?: T; error?: string; message?: string } | T;
	type SessionRow = {
		id: string;
		title?: string;
		status?: string;
		session_status?: string;
		scheduled_start?: string;
		package_title?: string;
		unassigned_participant_count?: number;
	};
	type AssignmentPreview = {
		summary: {
			participant_count: number;
			room_count: number;
			capacity_total: number;
			assigned_count: number;
			unassigned_count: number;
			mix_policy: string;
		};
		rooms: Array<{ room_id: string; room_name: string; capacity: number; participant_count: number; levels: Record<string, number> }>;
		warnings?: string[];
	};

	const userRoles = $derived(page.data.user?.roles ?? (page.data.user?.role ? [page.data.user.role] : []));
	const userPermissions = $derived(page.data.user?.permissions ?? []);
	const canAccess = $derived(
		userRoles.includes('admin')
			|| userPermissions.includes('asesmen.operator')
			|| userPermissions.includes('asesmen.event_manage')
			|| userPermissions.includes('asesmen.package_manage')
			|| userPermissions.includes('asesmen.session_manage')
			|| userPermissions.includes('asesmen.participant_manage')
	);
	const canManageRoomAssignment = $derived(userRoles.includes('admin'));

	const adminTasks: PreparationTask[] = [
		{
			step: '01',
			title: 'Pilih Kegiatan',
			description: 'Buat atau pilih kegiatan sebagai konteks ujian.',
			href: '/asesmen/kegiatan',
			cta: 'Kelola Kegiatan'
		},
		{
			step: '02',
			title: 'Siapkan Paket',
			description: 'Pilih paket soal yang siap dipakai. Untuk Simulasi/Gladi, gunakan paket server dari Bank Soal; seed sistem Informatika dapat dijadikan pool awal bila tersedia.',
			href: '/asesmen/paket',
			cta: 'Kelola Paket'
		},
		{
			step: '03',
			title: 'Atur Sesi & Token',
			description: 'Tetapkan jadwal, ruang, peserta, dan token.',
			href: '/asesmen/sesi/new',
			cta: 'Atur Sesi'
		},
		{
			step: '04',
			title: 'Masuk Pelaksanaan',
			description: 'Pantau ruang ujian saat paket dan sesi sudah siap.',
			href: '/asesmen/pelaksanaan',
			cta: 'Ke Pelaksanaan'
		}
	];

	const tasks = $derived(adminTasks);

	const taskColumns = [
		{ key: 'step', label: '#', class: 'w-20' },
		{ key: 'task', label: 'Pekerjaan', class: 'min-w-[16rem]' },
		{ key: 'description', label: 'Catatan', class: 'min-w-[20rem]' }
	];

	let loadingSessions = $state(false);
	let roomWorking = $state(false);
	let roomError = $state('');
	let roomSuccess = $state('');
	let sessions = $state<SessionRow[]>([]);
	let selectedSessionId = $state('');
	let mixPolicy = $state<'same_class' | 'same_grade' | 'mixed_scope'>('mixed_scope');
	let preview = $state<AssignmentPreview | null>(null);

	const selectedSession = $derived(sessions.find((session) => session.id === selectedSessionId) ?? sessions[0]);
	const assignmentPayload = $derived({
		mix_policy: mixPolicy,
		assignment_mode: 'random_balanced',
		allow_cross_grade: mixPolicy === 'mixed_scope',
		is_special_event: mixPolicy === 'mixed_scope'
	});

	onMount(() => {
		if (canManageRoomAssignment) void loadSessions();
	});

	async function loadSessions() {
		loadingSessions = true;
		roomError = '';
		try {
			const response = await fetchWithTimeout('/api/asesmen/sessions');
			sessions = await readJson<SessionRow[]>(response);
			selectedSessionId = selectedSessionId || sessions[0]?.id || '';
		} catch (error) {
			sessions = [];
			roomError = friendlyLoadError(error);
		} finally {
			loadingSessions = false;
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

	function friendlyLoadError(error: unknown) {
		if (error instanceof DOMException && error.name === 'AbortError') return 'Koneksi terlalu lama, coba muat ulang.';
		return error instanceof Error ? error.message : 'Data belum dapat dimuat.';
	}

	async function previewRooms() {
		if (!selectedSession?.id) {
			roomError = 'Pilih sesi ujian terlebih dahulu.';
			return;
		}
		roomWorking = true;
		roomError = '';
		roomSuccess = '';
		try {
			const response = await fetchWithTimeout(`/api/asesmen/sessions/${encodeURIComponent(selectedSession.id)}/rooms/assignment-preview`, {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify(assignmentPayload)
			});
			preview = await readJson<AssignmentPreview>(response);
			roomSuccess = 'Pratinjau pembagian ruang siap diperiksa.';
		} catch (error) {
			roomError = error instanceof Error ? error.message : 'Pratinjau pembagian ruang gagal.';
		} finally {
			roomWorking = false;
		}
	}

	async function applyRooms() {
		if (!selectedSession?.id || !preview) return;
		const ok = window.confirm('Simpan pembagian ruang dan nomor kursi sesuai pratinjau ini? Pembagian lama pada sesi ini akan diganti.');
		if (!ok) return;
		roomWorking = true;
		roomError = '';
		roomSuccess = '';
		try {
			const response = await fetchWithTimeout(`/api/asesmen/sessions/${encodeURIComponent(selectedSession.id)}/rooms/assignment`, {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify(assignmentPayload)
			});
			preview = await readJson<AssignmentPreview>(response);
			roomSuccess = 'Pembagian ruang dan nomor kursi berhasil disimpan.';
			await loadSessions();
		} catch (error) {
			roomError = error instanceof Error ? error.message : 'Pembagian ruang gagal disimpan.';
		} finally {
			roomWorking = false;
		}
	}

	function fmtDate(value?: string) {
		if (!value) return 'Belum dijadwalkan';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value;
		return date.toLocaleString('id-ID', { timeZone: 'Asia/Makassar', day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' }) + ' WITA';
	}

	function modeLabel(value: string) {
		if (value === 'same_class') return 'Per Kelas';
		if (value === 'same_grade') return 'Campur Satu Tingkat';
		return 'Campur Lintas Tingkat';
	}
</script>

<svelte:head>
	<title>Persiapan Ujian — MTsN 2 Kolaka Utara</title>
</svelte:head>

{#if canAccess}
	<div class="space-y-6">
	<section class="rounded-2xl border border-border bg-card p-4 shadow-sm">
		<div class="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
			<div class="max-w-3xl space-y-2">
				<Badge class="border-primary/20 bg-primary/10 text-primary" variant="outline">Ujian Digital · Persiapan</Badge>
				<h1 class="text-2xl font-semibold tracking-tight text-foreground">Persiapan Ujian</h1>
				<p class="max-w-2xl text-sm leading-6 text-muted-foreground">
					Siapkan kegiatan, paket, sesi, peserta, ruang, dan kode akses sebelum hari ujian.
				</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<Button href={resolve('/asesmen/ringkas')} variant="outline" size="sm">Kembali ke Ringkasan</Button>
			</div>
		</div>
	</section>

	<section aria-labelledby="persiapan-tasks-title" class="space-y-4">
		<div>
			<p class="text-xs font-semibold uppercase tracking-[0.22em] text-primary">Daftar tugas</p>
			<h2 id="persiapan-tasks-title" class="mt-1 text-xl font-semibold tracking-tight text-foreground">Checklist Persiapan</h2>
		</div>

		<MicroActionTable
			title="Checklist Persiapan"
			description="Pilih pekerjaan yang perlu diselesaikan."
			columns={taskColumns}
			rows={tasks}
			rowKey={(row) => (row as PreparationTask).href}
			tableClass="min-w-[760px]"
		>
			{#snippet cell(row, column)}
				{@const task = row as PreparationTask}
				{#if column.key === 'step'}
					<span class="inline-flex h-7 min-w-7 items-center justify-center rounded-md border border-primary/20 bg-primary/10 px-2 text-[11px] font-semibold text-primary">{task.step}</span>
				{:else if column.key === 'task'}
					<p class="font-medium text-foreground">{task.title}</p>
				{:else if column.key === 'description'}
					<p class="max-w-2xl leading-5 text-muted-foreground">{task.description}</p>
				{/if}
			{/snippet}
			{#snippet actions(row)}
				{@const task = row as PreparationTask}
				<Button href={resolve(task.href)} variant="outline" size="xs" class="border-primary/20 text-primary hover:bg-primary/10">{task.cta}</Button>
			{/snippet}
			{#snippet mobile(row)}
				{@const task = row as PreparationTask}
				<div class="space-y-2 text-xs">
					<div class="flex items-start gap-3">
						<span class="inline-flex h-7 min-w-7 shrink-0 items-center justify-center rounded-md border border-primary/20 bg-primary/10 px-2 text-[11px] font-semibold text-primary">{task.step}</span>
						<div class="min-w-0 flex-1">
							<p class="font-medium text-foreground">{task.title}</p>
							<p class="mt-1 leading-5 text-muted-foreground">{task.description}</p>
						</div>
					</div>
					<div class="flex justify-end pt-1">
						<Button href={resolve(task.href)} variant="outline" size="xs" class="border-primary/20 text-primary hover:bg-primary/10">{task.cta}</Button>
					</div>
				</div>
			{/snippet}
		</MicroActionTable>
	</section>

	{#if canManageRoomAssignment}
		<section aria-labelledby="room-assignment-title" class="rounded-2xl border border-border bg-card p-4 shadow-sm">
			<div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
				<div class="max-w-2xl space-y-2">
					<Badge class="border-primary/20 bg-primary/10 text-primary" variant="outline">Ruang & Kursi</Badge>
					<h2 id="room-assignment-title" class="text-xl font-semibold tracking-tight text-foreground">Pembagian Ruang</h2>
					<p class="text-sm leading-6 text-muted-foreground">
						Pilih sesi, lihat pratinjau, lalu simpan pembagian kursi. Kontrol ini dipusatkan di Persiapan agar Ringkasan tetap hanya untuk membaca status.
					</p>
				</div>
				<Button href={resolve('/asesmen/sesi')} variant="outline" size="sm">Kelola Sesi Detail</Button>
			</div>

			{#if roomError}
				<p class="mt-4 rounded-lg border border-destructive/20 bg-destructive/10 px-3 py-2 text-sm font-medium text-destructive">{roomError}</p>
			{/if}
			{#if roomSuccess}
				<p class="mt-4 rounded-lg border border-primary/20 bg-primary/10 px-3 py-2 text-sm font-medium text-primary">{roomSuccess}</p>
			{/if}

			{#if loadingSessions}
				<p class="mt-4 rounded-lg border border-border bg-muted/40 px-3 py-3 text-sm text-muted-foreground">Memuat sesi ujian...</p>
			{:else if sessions.length === 0}
				<div class="mt-4 rounded-lg border border-dashed border-border bg-muted/30 px-4 py-5">
					<p class="font-medium text-foreground">Belum ada sesi untuk dibagi ke ruang.</p>
					<p class="mt-1 text-sm text-muted-foreground">Buat sesi terlebih dahulu, lalu kembali ke bagian ini.</p>
					<Button href={resolve('/asesmen/sesi/new')} class="mt-3" size="sm">Buat Sesi</Button>
				</div>
			{:else}
				<div class="mt-4 grid gap-4 lg:grid-cols-[0.85fr_1.15fr]">
					<div class="space-y-4">
						<label class="block space-y-1 text-sm font-medium text-foreground" for="room-assignment-session">
							<span>Sesi ujian</span>
							<select id="room-assignment-session" class="min-h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={selectedSessionId} onchange={() => (preview = null)}>
								{#each sessions as session (session.id)}
									<option value={session.id}>{session.title ?? 'Sesi Ujian'} · {fmtDate(session.scheduled_start)}</option>
								{/each}
							</select>
						</label>

						<div class="space-y-2">
							<p class="text-sm font-medium text-foreground">Mode pembagian</p>
							{#each [
								{ value: 'same_class', title: 'Tetap per Kelas', desc: 'Peserta tetap mengikuti rombel asal.' },
								{ value: 'same_grade', title: 'Campur Satu Tingkat', desc: 'Rombel boleh bercampur, tetapi tingkat tetap dipisah.' },
								{ value: 'mixed_scope', title: 'Campur Lintas Tingkat', desc: 'Gunakan hanya bila panitia memang memutuskan lintas tingkat.' }
							] as option}
								<label class={`flex gap-3 rounded-lg border p-3 text-sm ${mixPolicy === option.value ? 'border-primary/40 bg-primary/10' : 'border-border bg-background'}`}>
									<input type="radio" bind:group={mixPolicy} value={option.value} onchange={() => (preview = null)} />
									<span><b>{option.title}</b><br /><span class="text-xs text-muted-foreground">{option.desc}</span></span>
								</label>
							{/each}
						</div>

						<div class="grid grid-cols-2 gap-2">
							<button class="min-h-10 rounded-md bg-primary px-3 text-sm font-semibold text-primary-foreground disabled:opacity-60" disabled={roomWorking || !selectedSession} onclick={previewRooms}>{roomWorking ? 'Memproses...' : 'Lihat Pratinjau'}</button>
							<button class="min-h-10 rounded-md bg-foreground px-3 text-sm font-semibold text-background disabled:opacity-50" disabled={roomWorking || !preview || (preview.summary.unassigned_count ?? 0) > 0} onclick={applyRooms}>Simpan</button>
						</div>
					</div>

					<div class="rounded-lg border border-border bg-muted/30 p-3">
						{#if preview}
							<p class="text-sm font-semibold text-foreground">Pratinjau {modeLabel(preview.summary.mix_policy)} · {preview.summary.assigned_count}/{preview.summary.participant_count} peserta</p>
							<p class="mt-1 text-xs text-muted-foreground">{preview.summary.room_count} ruang · kapasitas {preview.summary.capacity_total} kursi · belum ditempatkan {preview.summary.unassigned_count}</p>
							{#if preview.warnings?.length}
								<ul class="mt-2 space-y-1 text-xs font-medium text-warning">
									{#each preview.warnings as warning}
										<li>{warning}</li>
									{/each}
								</ul>
							{/if}
							<div class="mt-3 grid gap-2 sm:grid-cols-2">
								{#each preview.rooms as room (room.room_id)}
									<div class="rounded-lg border border-border bg-background p-3 text-xs">
										<p class="font-semibold text-foreground">{room.room_name}</p>
										<p class="text-muted-foreground">{room.participant_count}/{room.capacity} peserta</p>
										<p class="mt-1 text-muted-foreground">{Object.entries(room.levels).map(([key, value]) => `${key}: ${value}`).join(', ') || 'Kosong'}</p>
									</div>
								{/each}
							</div>
						{:else}
							<p class="text-sm font-medium text-foreground">Pratinjau belum dibuat.</p>
							<p class="mt-1 text-sm leading-6 text-muted-foreground">Gunakan “Lihat Pratinjau” sebelum menyimpan agar pembagian peserta bisa diperiksa lebih dulu.</p>
						{/if}
					</div>
				</div>
			{/if}
		</section>
	{/if}

	<p class="rounded-lg border border-dashed border-border bg-muted/40 px-3 py-2 text-xs text-muted-foreground">
		Catatan: penyusunan soal tetap berada di modul Bank Soal. DEMO hanya contoh lokal; Simulasi/Gladi/Ujian nyata harus memakai kegiatan, paket, dan sesi server.
	</p>
</div>
{:else}
	<div class="flex min-h-[60vh] flex-col items-center justify-center gap-4 px-4 py-16 text-center">
		<div class="max-w-lg rounded-2xl border border-border bg-card p-8 shadow-sm">
			<h2 class="text-xl font-semibold text-foreground">Akses terbatas</h2>
			<p class="mt-3 text-sm leading-6 text-muted-foreground">
				Fase persiapan ujian hanya tersedia untuk panitia/operator. Silakan kembali ke ringkasan atau gunakan menu asesmen lain sesuai tugas.
			</p>
			<div class="mt-6">
				<Button href={resolve('/asesmen/ringkas')} variant="outline">Kembali ke Ringkasan</Button>
			</div>
		</div>
	</div>
{/if}
