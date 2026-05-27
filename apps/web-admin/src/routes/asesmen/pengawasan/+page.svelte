<script lang="ts">
	import ActivityIcon from '@lucide/svelte/icons/activity';
	import SearchIcon from '@lucide/svelte/icons/search';
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { readClientApiData } from '$lib/client/api';

	type ProctorRoom = {
		id: string;
		session_id: string;
		room_name: string;
		room_token: string;
		status: string;
		is_locked: boolean;
		session_title: string;
		session_status: string;
		scheduled_start: string;
		scheduled_end: string;
		package_title: string;
		duration_minutes: number;
		school_room_code: string;
		school_room_name: string;
		school_room_building: string;
		school_room_location_note: string;
		participant_count: number;
		submitted_count: number;
		online_count: number;
		suspicious_count: number;
		missing_seat_count: number;
		proctor_count: number;
		proctor_names: string;
		actor_role: string;
	};

	type ProctorRoomStats = {
		total: number;
		active: number;
		scheduled: number;
		attention: number;
		participants: number;
	};

	const statusOptions = [
		{ value: 'all', label: 'Semua status' },
		{ value: 'active', label: 'Aktif' },
		{ value: 'scheduled', label: 'Terjadwal' },
		{ value: 'draft', label: 'Konsep' },
		{ value: 'finished', label: 'Selesai' },
	];
	let roomsPromise = $state<Promise<ProctorRoom[]> | null>(null);
	let rooms = $state<ProctorRoom[]>([]);
	let query = $state('');
	let statusFilter = $state('all');

	let filteredRooms = $derived.by(() => {
		const needle = query.trim().toLowerCase();
		return rooms.filter((room) => {
			if (statusFilter !== 'all' && room.session_status !== statusFilter) return false;
			if (!needle) return true;
			return [
				room.room_name,
				room.session_title,
				room.package_title,
				room.school_room_code,
				room.school_room_name,
				room.proctor_names,
				room.room_token,
			].some((value) => value.toLowerCase().includes(needle));
		});
	});

	let stats = $derived.by<ProctorRoomStats>(() => {
		let active = 0;
		let scheduled = 0;
		let attention = 0;
		let participants = 0;
		for (const room of rooms) {
			if (room.session_status === 'active') active += 1;
			if (room.session_status === 'scheduled') scheduled += 1;
			if (room.suspicious_count > 0 || room.missing_seat_count > 0 || room.proctor_count === 0) attention += 1;
			participants += room.participant_count;
		}
		return { total: rooms.length, active, scheduled, attention, participants };
	});

	onMount(() => {
		loadRooms();
	});

	function loadRooms() {
		roomsPromise = fetchRooms();
	}

	async function fetchRooms() {
		const response = await fetch('/api/asesmen/proctoring/my-rooms');
		const payload = await readClientApiData<ProctorRoom[]>(response, 'Gagal memuat ruang saya');
		rooms = Array.isArray(payload) ? payload : [];
		return rooms;
	}

	function retryRooms(reset?: () => void) {
		reset?.();
		loadRooms();
	}

	function roomsErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat ruang saya';
	}

	function handleRenderError(error: unknown) {
		console.error('Ruang pengawasan ujian belum dapat ditampilkan', error);
	}

	function fmtDate(value: string | null | undefined) {
		if (!value) return '—';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value;
		return date.toLocaleString('id-ID', {
			timeZone: 'Asia/Makassar',
			day: '2-digit',
			month: 'short',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
		}) + ' WITA';
	}

	function statusLabel(status: string) {
		const labels: Record<string, string> = {
			draft: 'Konsep',
			scheduled: 'Terjadwal',
			active: 'Aktif',
			finished: 'Selesai',
			cancelled: 'Dibatalkan',
		};
		return labels[status] ?? status;
	}

	function statusClass(status: string) {
		const classes: Record<string, string> = {
			active: 'border-primary/20 bg-primary/10 text-primary',
			scheduled: 'border-accent bg-accent/60 text-accent-foreground',
			draft: 'border-border bg-muted/50 text-muted-foreground',
			finished: 'border-border bg-muted text-muted-foreground',
			cancelled: 'border-destructive/30 bg-destructive/10 text-destructive',
		};
		return classes[status] ?? 'border-border bg-card text-muted-foreground';
	}

	function roomRoleLabel(role: string) {
		const labels: Record<string, string> = {
			utama: 'Pengawas utama',
			pendamping: 'Pengawas pendamping',
			cadangan: 'Pengawas cadangan',
		};
		return labels[role] ?? (role ? role : 'Operator');
	}

	function roomLocation(room: ProctorRoom) {
		const parts = [room.school_room_code, room.school_room_name, room.school_room_building, room.school_room_location_note]
			.map((item) => item.trim())
			.filter(Boolean);
		return parts.length > 0 ? parts.join(' · ') : 'Ruang manual sesi';
	}

	function readinessText(room: ProctorRoom) {
		const issues = [];
		if (room.proctor_count === 0) issues.push('pengawas belum ada');
		if (room.missing_seat_count > 0) issues.push(`${room.missing_seat_count} meja belum lengkap`);
		if (room.suspicious_count > 0) issues.push(`${room.suspicious_count} peserta atensi`);
		return issues.length > 0 ? issues.join(', ') : 'Hijau · Aman';
	}

	function simpleSignalLabel(room: ProctorRoom) {
		if (room.suspicious_count > 0) return 'Merah · Butuh bantuan';
		if (room.missing_seat_count > 0 || room.proctor_count === 0 || room.online_count < room.participant_count) return 'Kuning · Perlu dicek';
		return 'Hijau · Aman';
	}

	function simpleSignalClass(room: ProctorRoom) {
		if (room.suspicious_count > 0) return 'border-destructive/30 bg-destructive/10 text-destructive';
		if (room.missing_seat_count > 0 || room.proctor_count === 0 || room.online_count < room.participant_count) return 'border-warning/30 bg-warning/10 text-warning';
		return 'border-primary/20 bg-primary/10 text-primary';
	}

	function attentionClass(room: ProctorRoom) {
		if (room.suspicious_count > 0) return 'border-destructive/30 bg-destructive/10 text-destructive';
		if (room.missing_seat_count > 0 || room.proctor_count === 0) return 'border-warning/30 bg-warning/10 text-warning';
		return 'border-primary/20 bg-primary/10 text-primary';
	}
</script>

<svelte:head>
	<title>Pengawasan Ruang — MTsN 2 Kolut</title>
</svelte:head>

<AsyncContent promise={roomsPromise} onerror={handleRenderError}>
	{#snippet pending()}
		<div class="space-y-5 p-4 md:p-6">
			<div class="space-y-2">
				<Skeleton class="h-8 w-72" />
				<Skeleton class="h-4 w-96" />
			</div>
			<div class="grid gap-3 md:grid-cols-5">
				{#each Array.from({ length: 5 }) as _, index (`proctor-room-stat-${index}`)}
					<Skeleton class="h-24 rounded-lg" />
				{/each}
			</div>
			<div class="grid gap-3 xl:grid-cols-2">
				{#each Array.from({ length: 4 }) as _, index (`proctor-room-card-${index}`)}
					<Skeleton class="h-56 rounded-lg" />
				{/each}
			</div>
		</div>
	{/snippet}

	{#snippet failed(error, reset)}
		<div class="p-4 md:p-6">
			<RecoveryPanel
				title="Ruang Saya Belum Tersaji"
				message={roomsErrorMessage(error)}
				onRetry={() => retryRooms(reset)}
			/>
		</div>
	{/snippet}

	{#snippet children(value)}
		{@const loadedRooms = value as ProctorRoom[]}
		<div class="space-y-5 p-4 md:p-6">
			<section class="flex flex-col gap-4 border-b border-primary/20 pb-5 lg:flex-row lg:items-end lg:justify-between">
				<div class="max-w-3xl space-y-2">
					<p class="text-xs font-semibold uppercase tracking-[0.24em] text-primary">Ujian Digital / Ruang Saya</p>
					<h1 class="text-2xl font-semibold tracking-tight text-foreground md:text-3xl">Ruang Saya</h1>
					<p class="text-sm leading-6 text-muted-foreground">
						Halaman sederhana untuk pengawas: pilih ruang, lihat kode ruang, pantau label hijau/kuning/merah, lalu tekan tombol besar Mulai Ujian.
					</p>
				</div>
				<div class="flex flex-wrap gap-2">
					<Button href={resolve('/asesmen/aplikasi-siswa')} variant="outline">Panduan Perangkat</Button>
				</div>
			</section>

			<section class="grid gap-3 md:grid-cols-5">
				<div class="border border-primary/20 bg-card p-4 shadow-sm">
					<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Ruang</p>
					<p class="mt-1 text-2xl font-bold text-foreground">{stats.total}</p>
				</div>
				<div class="border border-primary/20 bg-card p-4 shadow-sm">
					<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Aktif</p>
					<p class="mt-1 text-2xl font-bold text-primary">{stats.active}</p>
				</div>
				<div class="border border-primary/20 bg-card p-4 shadow-sm">
					<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Terjadwal</p>
					<p class="mt-1 text-2xl font-bold text-accent-foreground">{stats.scheduled}</p>
				</div>
				<div class="border border-primary/20 bg-card p-4 shadow-sm">
					<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Peserta</p>
					<p class="mt-1 text-2xl font-bold text-foreground">{stats.participants}</p>
				</div>
				<div class="border border-primary/20 bg-card p-4 shadow-sm">
					<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Atensi</p>
					<p class="mt-1 text-2xl font-bold text-warning">{stats.attention}</p>
				</div>
			</section>

			<section class="grid gap-3 border border-border bg-card p-4 shadow-sm lg:grid-cols-[minmax(0,1fr)_220px]">
				<div>
					<label for="proctor-room-search" class="mb-1 block text-xs font-semibold uppercase tracking-wide text-muted-foreground">Cari ruang / sesi / kode ruang</label>
					<div class="relative">
						<SearchIcon class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
						<input
							id="proctor-room-search"
							bind:value={query}
							class="h-10 w-full rounded-md border border-border bg-card pl-9 pr-3 text-sm outline-none focus:border-ring focus:ring-2 focus:ring-ring"
							placeholder="Cari ruang, paket, kode ruang, atau pengawas"
						/>
					</div>
				</div>
				<div>
					<label for="proctor-room-status" class="mb-1 block text-xs font-semibold uppercase tracking-wide text-muted-foreground">Status sesi</label>
					<select
						id="proctor-room-status"
						bind:value={statusFilter}
						class="h-10 w-full rounded-md border border-border bg-card px-3 text-sm outline-none focus:border-ring focus:ring-2 focus:ring-ring"
					>
						{#each statusOptions as option (option.value)}
							<option value={option.value}>{option.label}</option>
						{/each}
					</select>
				</div>
			</section>

			{#if loadedRooms.length === 0}
				<EmptyStatePanel
					title="Belum Ada Ruang Saya"
					description="Ruang akan muncul setelah panitia menetapkan pengawas untuk sesi ujian."
				/>
			{:else if filteredRooms.length === 0}
				<EmptyStatePanel
					title="Filter Tidak Menemukan Ruang"
					description="Ubah kata kunci atau status sesi untuk melihat ruang pengawas yang lain."
				/>
			{:else}
				<section class="grid gap-3 xl:grid-cols-2">
					{#each filteredRooms as room (room.id)}
						<article class="border border-primary/20 bg-card p-4 shadow-sm">
							<div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
								<div class="min-w-0 space-y-2">
									<div class="flex flex-wrap items-center gap-2">
										<Badge variant="outline" class={simpleSignalClass(room)}>{simpleSignalLabel(room)}</Badge>
										<Badge variant="outline" class={statusClass(room.session_status)}>{statusLabel(room.session_status)}</Badge>
										<Badge variant="outline" class={attentionClass(room)}>{readinessText(room)}</Badge>
										{#if room.actor_role}
											<Badge variant="outline" class="border-primary/20 text-primary">{roomRoleLabel(room.actor_role)}</Badge>
										{/if}
									</div>
									<div>
										<h2 class="text-lg font-semibold text-foreground">{room.room_name}</h2>
										<p class="text-sm text-muted-foreground">{room.session_title}</p>
										<p class="mt-1 text-xs text-muted-foreground">{room.package_title} · {fmtDate(room.scheduled_start)}</p>
									</div>
								</div>
								<div class="rounded-md border border-primary/20 bg-primary/10 px-3 py-2 text-center">
									<p class="text-[10px] font-semibold uppercase tracking-[0.18em] text-primary">Kode Ruang</p>
									<p class="font-mono text-lg font-bold tracking-[0.16em] text-primary">{room.room_token || '—'}</p>
								</div>
							</div>

							<div class="mt-4 grid gap-2 text-sm sm:grid-cols-4">
								<div class="border border-border bg-muted/50 p-3">
									<p class="text-[11px] uppercase tracking-wide text-muted-foreground">Peserta</p>
									<p class="font-semibold text-foreground">{room.participant_count}</p>
								</div>
								<div class="border border-border bg-muted/50 p-3">
									<p class="text-[11px] uppercase tracking-wide text-muted-foreground">Terhubung</p>
									<p class="font-semibold text-primary">{room.online_count}</p>
								</div>
								<div class="border border-border bg-muted/50 p-3">
									<p class="text-[11px] uppercase tracking-wide text-muted-foreground">Kirim</p>
									<p class="font-semibold text-foreground">{room.submitted_count}</p>
								</div>
								<div class="border border-border bg-muted/50 p-3">
									<p class="text-[11px] uppercase tracking-wide text-muted-foreground">Atensi</p>
									<p class="font-semibold text-warning">{room.suspicious_count}</p>
								</div>
							</div>

							<div class="mt-4 grid gap-2 text-xs text-muted-foreground md:grid-cols-2">
								<p><span class="font-semibold text-foreground">Lokasi:</span> {roomLocation(room)}</p>
								<p><span class="font-semibold text-foreground">Pengawas:</span> {room.proctor_names || 'Belum ditugaskan'}</p>
								<p><span class="font-semibold text-foreground">Durasi:</span> {room.duration_minutes} menit</p>
								<p><span class="font-semibold text-foreground">Selesai:</span> {fmtDate(room.scheduled_end)}</p>
							</div>

							<div class="mt-4 flex flex-wrap gap-2">
								<Button class="h-12 px-5 text-base font-bold" href={resolve(`/asesmen/sesi/${room.session_id}/rooms/${room.id}/proctoring`)}>
									<ActivityIcon class="mr-2 size-4" />
									Mulai Ujian
								</Button>
							</div>
						</article>
					{/each}
				</section>

			{/if}
		</div>
	{/snippet}
</AsyncContent>
