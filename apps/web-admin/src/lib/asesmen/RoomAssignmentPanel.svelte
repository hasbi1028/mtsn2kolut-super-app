<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import {
		buildRoomAssignmentPayload,
		canSaveRoomAssignment,
		cloneRoomAssignments,
		roomAssignmentModeHint,
		roomAssignmentModeLabel,
		roomAssignmentModeOptions,
		roomAssignmentOptions,
		sortRoomAssignments,
		type RoomAssignmentMixPolicy,
		type RoomAssignmentMode,
		type RoomAssignmentPreview,
		type RoomAssignmentSeat,
		type RoomAssignmentSessionRow
	} from '$lib/asesmen/room-assignment';

	type ApiEnvelope<T> = { data?: T; items?: T; error?: string; message?: string } | T;

	let loadingSessions = $state(false);
	let roomWorking = $state(false);
	let roomError = $state('');
	let roomSuccess = $state('');
	let sessions = $state<RoomAssignmentSessionRow[]>([]);
	let selectedSessionId = $state('');
	let mixPolicy = $state<RoomAssignmentMixPolicy>('mixed_scope');
	let assignmentMode = $state<RoomAssignmentMode>('random_balanced');
	let preview = $state<RoomAssignmentPreview | null>(null);
	let draftAssignments = $state<RoomAssignmentSeat[]>([]);

	const selectedSession = $derived(sessions.find((session) => session.id === selectedSessionId) ?? sessions[0]);
	const assignmentPayload = $derived(buildRoomAssignmentPayload(mixPolicy, assignmentMode, draftAssignments));
	const canSavePreview = $derived(canSaveRoomAssignment(preview));
	const hasManualDraft = $derived(assignmentMode === 'manual' && draftAssignments.length > 0);
	const roomOptions = $derived(preview?.rooms ?? []);

	onMount(() => {
		void loadSessions();
	});

	async function loadSessions() {
		loadingSessions = true;
		roomError = '';
		try {
			const response = await fetchWithTimeout('/api/asesmen/sessions');
			sessions = await readJson<RoomAssignmentSessionRow[]>(response);
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

	function resetPreviewState(message = '') {
		preview = null;
		draftAssignments = [];
		roomSuccess = message;
	}

	function syncDraftAssignments(next: RoomAssignmentPreview | null) {
		draftAssignments = sortRoomAssignments(cloneRoomAssignments(next?.assignments ?? []));
		preview = next;
	}

	function setAssignmentMode(next: RoomAssignmentMode) {
		assignmentMode = next;
		if (next === 'manual') {
			roomSuccess = 'Mode manual aktif. Ubah ruang atau nomor kursi lalu tinjau ulang sebelum simpan.';
			if (draftAssignments.length === 0 && preview?.assignments?.length) {
				draftAssignments = cloneRoomAssignments(preview.assignments);
			}
			return;
		}
		if (preview?.assignments?.length) {
			draftAssignments = cloneRoomAssignments(preview.assignments);
		}
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
			const next = await readJson<RoomAssignmentPreview>(response);
			syncDraftAssignments(next);
			roomSuccess = assignmentMode === 'manual'
				? 'Pratinjau manual siap. Silakan ubah ruang atau kursi bila perlu.'
				: 'Pratinjau pembagian otomatis siap diperiksa.';
		} catch (error) {
			roomError = error instanceof Error ? error.message : 'Pratinjau pembagian ruang gagal.';
		} finally {
			roomWorking = false;
		}
	}

	async function applyRooms() {
		if (!selectedSession?.id) return;
		if (assignmentMode === 'manual' && draftAssignments.length === 0) {
			roomError = 'Buat pratinjau otomatis dulu sebelum menyusun manual.';
			return;
		}
		if (assignmentMode === 'random_balanced' && !preview) {
			roomError = 'Buat pratinjau otomatis dulu sebelum menyimpan.';
			return;
		}
		if (assignmentMode === 'random_balanced' && !canSavePreview) {
			roomError = 'Masih ada peserta yang belum ditempatkan.';
			return;
		}
		const confirmation = assignmentMode === 'manual'
			? 'Simpan pembagian manual ini? Pembagian lama pada sesi ini akan diganti.'
			: 'Simpan pembagian ruang otomatis ini? Pembagian lama pada sesi ini akan diganti.';
		if (!window.confirm(confirmation)) return;
		roomWorking = true;
		roomError = '';
		roomSuccess = '';
		try {
			const savePayload = assignmentMode === 'manual'
				? assignmentPayload
				: { ...assignmentPayload, assignments: preview?.assignments?.length ? preview.assignments : undefined };
			const response = await fetchWithTimeout(`/api/asesmen/sessions/${encodeURIComponent(selectedSession.id)}/rooms/assignment`, {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify(savePayload)
			});
			const next = await readJson<RoomAssignmentPreview>(response);
			syncDraftAssignments(next);
			roomSuccess = assignmentMode === 'manual'
				? 'Pembagian manual dan nomor kursi berhasil disimpan.'
				: 'Pembagian otomatis dan nomor kursi berhasil disimpan.';
			await loadSessions();
		} catch (error) {
			roomError = error instanceof Error ? error.message : 'Pembagian ruang gagal disimpan.';
		} finally {
			roomWorking = false;
		}
	}

	function updateDraftAssignment(participantId: string, patch: Partial<RoomAssignmentSeat>) {
		draftAssignments = draftAssignments.map((assignment) => {
			if (assignment.participant_id !== participantId) return assignment;
			const nextRoom = patch.room_id ? roomOptions.find((room) => room.room_id === patch.room_id) : undefined;
			const next = { ...assignment, ...patch };
			if (nextRoom) next.room_name = nextRoom.room_name;
			if (!patch.room_id && patch.room_id !== undefined) next.room_name = '';
			return next;
		});
	}

	function fmtDate(value?: string) {
		if (!value) return 'Belum dijadwalkan';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value;
		return date.toLocaleString('id-ID', { timeZone: 'Asia/Makassar', day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' }) + ' WITA';
	}
</script>

<section aria-labelledby="room-assignment-title" class="rounded-2xl border border-border bg-card p-4 shadow-sm">
	<div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
		<div class="max-w-2xl space-y-2">
			<Badge class="border-primary/20 bg-primary/10 text-primary" variant="outline">Ruang & Kursi</Badge>
			<h2 id="room-assignment-title" class="text-xl font-semibold tracking-tight text-foreground">Pembagian Ruang</h2>
			<p class="text-sm leading-6 text-muted-foreground">
				Pilih sesi, lalu gunakan mode otomatis atau manual. Otomatis cocok untuk pembagian cepat; manual cocok untuk koreksi kecil sebelum disimpan.
			</p>
		</div>
		<Button href={resolve('/asesmen/sesi')} variant="outline" size="sm">Kelola Sesi Detail</Button>
	</div>

	<div class="mt-4 flex flex-wrap gap-2">
		{#each roomAssignmentModeOptions as option}
			<button
				type="button"
				class={`rounded-xl border px-3 py-2 text-sm font-bold transition ${assignmentMode === option.value ? 'border-primary bg-primary/10 text-primary' : 'border-border bg-background text-foreground hover:border-primary/30'}`}
				onclick={() => setAssignmentMode(option.value)}
			>
				{option.title}
			</button>
		{/each}
	</div>
	<p class="mt-2 text-xs text-muted-foreground">{roomAssignmentModeHint(assignmentMode)}</p>

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
					<select id="room-assignment-session" class="min-h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={selectedSessionId} onchange={() => resetPreviewState()}>
						{#each sessions as session (session.id)}
							<option value={session.id}>{session.title ?? 'Sesi Ujian'} · {fmtDate(session.scheduled_start)}</option>
						{/each}
					</select>
				</label>

				<div class="space-y-2">
					<p class="text-sm font-medium text-foreground">Mode pembagian</p>
					{#each roomAssignmentOptions as option}
						<label class={`flex gap-3 rounded-lg border p-3 text-sm ${mixPolicy === option.value ? 'border-primary/40 bg-primary/10' : 'border-border bg-background'}`}>
							<input type="radio" bind:group={mixPolicy} value={option.value} onchange={() => resetPreviewState()} />
							<span><b>{option.title}</b><br /><span class="text-xs text-muted-foreground">{option.desc}</span></span>
						</label>
					{/each}
				</div>

				<div class="grid grid-cols-2 gap-2">
					<button class="min-h-10 rounded-md bg-primary px-3 text-sm font-semibold text-primary-foreground disabled:opacity-60" disabled={roomWorking || !selectedSession} onclick={previewRooms}>{roomWorking ? 'Memproses...' : assignmentMode === 'manual' ? 'Buat Pratinjau Manual' : 'Acak Otomatis & Pratinjau'}</button>
					<button class="min-h-10 rounded-md bg-foreground px-3 text-sm font-semibold text-background disabled:opacity-50" disabled={roomWorking || (assignmentMode === 'manual' ? !hasManualDraft : !canSavePreview)} onclick={applyRooms}>{assignmentMode === 'manual' ? 'Simpan Manual' : 'Simpan Otomatis'}</button>
				</div>

				{#if assignmentMode === 'manual'}
					<div class="rounded-lg border border-dashed border-border bg-muted/20 p-3 text-xs leading-5 text-muted-foreground">
						<p class="font-semibold text-foreground">Alur manual singkat</p>
						<p>1. Buat pratinjau otomatis dulu. 2. Ubah ruang atau nomor kursi. 3. Tinjau ulang bila perlu. 4. Simpan manual.</p>
					</div>
				{/if}
			</div>

			<div class="rounded-lg border border-border bg-muted/30 p-3">
				{#if preview}
					<p class="text-sm font-semibold text-foreground">Pratinjau {roomAssignmentModeLabel(preview.summary.mix_policy)} · {preview.summary.assigned_count}/{preview.summary.participant_count} peserta</p>
					<p class="mt-1 text-xs text-muted-foreground">{preview.summary.room_count} ruang · kapasitas {preview.summary.capacity_total} kursi · belum ditempatkan {preview.summary.unassigned_count}</p>
					{#if preview.warnings?.length}
						<ul class="mt-2 space-y-1 text-xs font-medium text-warning">
							{#each preview.warnings as warning}
								<li>{warning}</li>
							{/each}
						</ul>
					{/if}

					{#if assignmentMode === 'manual'}
						<div class="mt-3 space-y-2">
							<div class="flex items-center justify-between gap-2">
								<p class="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Editor manual</p>
								<Button variant="outline" size="xs" class="border-primary/20 text-primary hover:bg-primary/10" onclick={previewRooms}>Tinjau Manual</Button>
							</div>
							<div class="max-h-[28rem] space-y-2 overflow-y-auto pr-1">
								{#each sortRoomAssignments(draftAssignments) as assignment (assignment.participant_id)}
									<div class="rounded-lg border border-border bg-background p-3 text-xs">
										<div class="flex items-start justify-between gap-2">
											<div class="min-w-0">
												<p class="font-semibold text-foreground">{assignment.participant_name ?? 'Peserta'}</p>
												<p class="text-muted-foreground">{assignment.participant_nis ?? '-'}{assignment.participant_class ? ` · ${assignment.participant_class}` : ''}</p>
											</div>
										</div>
										<div class="mt-2 grid gap-2 sm:grid-cols-[1.3fr_0.7fr]">
											<label class="space-y-1">
												<span class="text-[11px] font-semibold uppercase tracking-[0.16em] text-muted-foreground">Ruang</span>
												<select class="w-full rounded-md border border-input bg-background px-2 py-2 text-xs" value={assignment.room_id} onchange={(event) => updateDraftAssignment(assignment.participant_id, { room_id: (event.currentTarget as HTMLSelectElement).value })}>
													{#each roomOptions as room}
														<option value={room.room_id}>{room.room_name} · {room.participant_count}/{room.capacity}</option>
													{/each}
												</select>
											</label>
											<label class="space-y-1">
												<span class="text-[11px] font-semibold uppercase tracking-[0.16em] text-muted-foreground">Kursi</span>
												<input class="w-full rounded-md border border-input bg-background px-2 py-2 text-xs" type="number" min="1" value={assignment.seat_no} oninput={(event) => updateDraftAssignment(assignment.participant_id, { seat_no: Number((event.currentTarget as HTMLInputElement).value || 0) })} />
											</label>
										</div>
									</div>
								{/each}
							</div>
						</div>
					{:else}
						<div class="mt-3 grid gap-2 sm:grid-cols-2">
							{#each preview.rooms as room (room.room_id)}
								<div class="rounded-lg border border-border bg-background p-3 text-xs">
									<p class="font-semibold text-foreground">{room.room_name}</p>
									<p class="text-muted-foreground">{room.participant_count}/{room.capacity} peserta</p>
									<p class="mt-1 text-muted-foreground">{Object.entries(room.levels).map(([key, value]) => `${key}: ${value}`).join(', ') || 'Kosong'}</p>
								</div>
							{/each}
						</div>
					{/if}
				{:else}
					<p class="text-sm font-medium text-foreground">Pratinjau belum dibuat.</p>
					<p class="mt-1 text-sm leading-6 text-muted-foreground">Gunakan tombol pratinjau untuk menyiapkan pembagian otomatis atau manual.</p>
				{/if}
			</div>
		</div>
	{/if}
</section>
