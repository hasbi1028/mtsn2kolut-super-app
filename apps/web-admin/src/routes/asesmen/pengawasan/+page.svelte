<script lang="ts">
	import ActivityIcon from '@lucide/svelte/icons/activity';
	import PrinterIcon from '@lucide/svelte/icons/printer';
	import SearchIcon from '@lucide/svelte/icons/search';
	import ShieldCheckIcon from '@lucide/svelte/icons/shield-check';
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import * as Table from '$lib/components/ui/table';
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
		{ value: 'draft', label: 'Draft' },
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
		const payload = await readClientApiData<ProctorRoom[]>(response, 'Gagal memuat ruang pengawas');
		rooms = Array.isArray(payload) ? payload : [];
		return rooms;
	}

	function retryRooms(reset?: () => void) {
		reset?.();
		loadRooms();
	}

	function roomsErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat ruang pengawas';
	}

	function handleRenderError(error: unknown) {
		console.error('CBT proctor rooms render failed', error);
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
			draft: 'Draft',
			scheduled: 'Terjadwal',
			active: 'Aktif',
			finished: 'Selesai',
			cancelled: 'Dibatalkan',
		};
		return labels[status] ?? status;
	}

	function statusClass(status: string) {
		const classes: Record<string, string> = {
			active: 'border-emerald-300 bg-emerald-50 text-emerald-700',
			scheduled: 'border-blue-200 bg-blue-50 text-blue-700',
			draft: 'border-slate-200 bg-slate-50 text-slate-600',
			finished: 'border-slate-300 bg-slate-100 text-slate-600',
			cancelled: 'border-red-200 bg-red-50 text-red-700',
		};
		return classes[status] ?? 'border-slate-200 bg-white text-slate-600';
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
		return issues.length > 0 ? issues.join(', ') : 'Siap dipantau';
	}

	function attentionClass(room: ProctorRoom) {
		if (room.suspicious_count > 0) return 'border-red-200 bg-red-50 text-red-700';
		if (room.missing_seat_count > 0 || room.proctor_count === 0) return 'border-amber-200 bg-amber-50 text-amber-700';
		return 'border-emerald-200 bg-emerald-50 text-emerald-700';
	}
</script>

<svelte:head>
	<title>Ruang Pengawas CBT — MTsN 2 Kolut</title>
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
				title="Ruang Pengawas Belum Tersaji"
				message={roomsErrorMessage(error)}
				onRetry={() => retryRooms(reset)}
			/>
		</div>
	{/snippet}

	{#snippet children(value)}
		{@const loadedRooms = value as ProctorRoom[]}
		<div class="space-y-5 p-4 md:p-6">
			<section class="flex flex-col gap-4 border-b border-emerald-100 pb-5 lg:flex-row lg:items-end lg:justify-between">
				<div class="max-w-3xl space-y-2">
					<p class="text-xs font-semibold uppercase tracking-[0.24em] text-emerald-700">CBT / Pengawas Ruang</p>
					<h1 class="text-2xl font-semibold tracking-tight text-slate-900 md:text-3xl">Ruang Pengawas CBT</h1>
					<p class="text-sm leading-6 text-slate-600">
						Satu layar untuk menemukan ruang ujian yang perlu dipantau, membuka dashboard live, dan kembali ke Monitoring BYOD saat butuh status guide.
					</p>
				</div>
				<div class="flex flex-wrap gap-2">
					<Button href={resolve('/asesmen/aplikasi-siswa')} variant="outline">Monitoring BYOD</Button>
					<Button href={resolve('/asesmen/sesi')} variant="outline">Sesi Ujian</Button>
				</div>
			</section>

			<section class="grid gap-3 md:grid-cols-5">
				<div class="border border-emerald-100 bg-white p-4 shadow-sm">
					<p class="text-xs font-semibold uppercase tracking-wide text-slate-500">Ruang</p>
					<p class="mt-1 text-2xl font-bold text-slate-900">{stats.total}</p>
				</div>
				<div class="border border-emerald-100 bg-white p-4 shadow-sm">
					<p class="text-xs font-semibold uppercase tracking-wide text-slate-500">Aktif</p>
					<p class="mt-1 text-2xl font-bold text-emerald-700">{stats.active}</p>
				</div>
				<div class="border border-emerald-100 bg-white p-4 shadow-sm">
					<p class="text-xs font-semibold uppercase tracking-wide text-slate-500">Terjadwal</p>
					<p class="mt-1 text-2xl font-bold text-blue-700">{stats.scheduled}</p>
				</div>
				<div class="border border-emerald-100 bg-white p-4 shadow-sm">
					<p class="text-xs font-semibold uppercase tracking-wide text-slate-500">Peserta</p>
					<p class="mt-1 text-2xl font-bold text-slate-900">{stats.participants}</p>
				</div>
				<div class="border border-emerald-100 bg-white p-4 shadow-sm">
					<p class="text-xs font-semibold uppercase tracking-wide text-slate-500">Atensi</p>
					<p class="mt-1 text-2xl font-bold text-amber-700">{stats.attention}</p>
				</div>
			</section>

			<section class="grid gap-3 border border-slate-200 bg-white p-4 shadow-sm lg:grid-cols-[minmax(0,1fr)_220px]">
				<div>
					<label for="proctor-room-search" class="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-500">Cari ruang / sesi / token</label>
					<div class="relative">
						<SearchIcon class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-slate-400" />
						<input
							id="proctor-room-search"
							bind:value={query}
							class="h-10 w-full rounded-md border border-slate-200 bg-white pl-9 pr-3 text-sm outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-100"
							placeholder="Cari ruang, paket, token, atau pengawas"
						/>
					</div>
				</div>
				<div>
					<label for="proctor-room-status" class="mb-1 block text-xs font-semibold uppercase tracking-wide text-slate-500">Status sesi</label>
					<select
						id="proctor-room-status"
						bind:value={statusFilter}
						class="h-10 w-full rounded-md border border-slate-200 bg-white px-3 text-sm outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-100"
					>
						{#each statusOptions as option (option.value)}
							<option value={option.value}>{option.label}</option>
						{/each}
					</select>
				</div>
			</section>

			{#if loadedRooms.length === 0}
				<EmptyStatePanel
					title="Belum Ada Ruang Pengawas"
					description="Ruang akan muncul setelah operator menetapkan pengawas pada tab Ruangan di detail sesi CBT."
				/>
			{:else if filteredRooms.length === 0}
				<EmptyStatePanel
					title="Filter Tidak Menemukan Ruang"
					description="Ubah kata kunci atau status sesi untuk melihat ruang pengawas yang lain."
				/>
			{:else}
				<section class="grid gap-3 xl:grid-cols-2">
					{#each filteredRooms as room (room.id)}
						<article class="border border-emerald-100 bg-white p-4 shadow-sm">
							<div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
								<div class="min-w-0 space-y-2">
									<div class="flex flex-wrap items-center gap-2">
										<Badge variant="outline" class={statusClass(room.session_status)}>{statusLabel(room.session_status)}</Badge>
										<Badge variant="outline" class={attentionClass(room)}>{readinessText(room)}</Badge>
										{#if room.actor_role}
											<Badge variant="outline" class="border-emerald-200 text-emerald-700">{roomRoleLabel(room.actor_role)}</Badge>
										{/if}
									</div>
									<div>
										<h2 class="text-lg font-semibold text-slate-900">{room.room_name}</h2>
										<p class="text-sm text-slate-600">{room.session_title}</p>
										<p class="mt-1 text-xs text-slate-500">{room.package_title} · {fmtDate(room.scheduled_start)}</p>
									</div>
								</div>
								<div class="rounded-md border border-emerald-200 bg-emerald-50 px-3 py-2 text-center">
									<p class="text-[10px] font-semibold uppercase tracking-[0.18em] text-emerald-700">Token Ruang</p>
									<p class="font-mono text-lg font-bold tracking-[0.16em] text-emerald-950">{room.room_token || '—'}</p>
								</div>
							</div>

							<div class="mt-4 grid gap-2 text-sm sm:grid-cols-4">
								<div class="border border-slate-200 bg-slate-50 p-3">
									<p class="text-[11px] uppercase tracking-wide text-slate-500">Peserta</p>
									<p class="font-semibold text-slate-900">{room.participant_count}</p>
								</div>
								<div class="border border-slate-200 bg-slate-50 p-3">
									<p class="text-[11px] uppercase tracking-wide text-slate-500">Online</p>
									<p class="font-semibold text-emerald-700">{room.online_count}</p>
								</div>
								<div class="border border-slate-200 bg-slate-50 p-3">
									<p class="text-[11px] uppercase tracking-wide text-slate-500">Submit</p>
									<p class="font-semibold text-slate-900">{room.submitted_count}</p>
								</div>
								<div class="border border-slate-200 bg-slate-50 p-3">
									<p class="text-[11px] uppercase tracking-wide text-slate-500">Atensi</p>
									<p class="font-semibold text-amber-700">{room.suspicious_count}</p>
								</div>
							</div>

							<div class="mt-4 grid gap-2 text-xs text-slate-600 md:grid-cols-2">
								<p><span class="font-semibold text-slate-700">Lokasi:</span> {roomLocation(room)}</p>
								<p><span class="font-semibold text-slate-700">Pengawas:</span> {room.proctor_names || 'Belum ditugaskan'}</p>
								<p><span class="font-semibold text-slate-700">Durasi:</span> {room.duration_minutes} menit</p>
								<p><span class="font-semibold text-slate-700">Selesai:</span> {fmtDate(room.scheduled_end)}</p>
							</div>

							<div class="mt-4 flex flex-wrap gap-2">
								<Button href={resolve(`/asesmen/sesi/${room.session_id}/rooms/${room.id}/proctoring`)}>
									<ActivityIcon class="mr-2 size-4" />
									Dashboard
								</Button>
								<Button href={resolve(`/asesmen/sesi/${room.session_id}/rooms/${room.id}/print-pack`)} variant="outline">
									<PrinterIcon class="mr-2 size-4" />
									Paket Cetak
								</Button>
								<Button href={resolve(`/asesmen/sesi/${room.session_id}`)} variant="outline">
									<ShieldCheckIcon class="mr-2 size-4" />
									Detail Sesi
								</Button>
							</div>
						</article>
					{/each}
				</section>

				<section class="border border-slate-200 bg-white shadow-sm">
					<div class="flex items-center justify-between gap-3 border-b border-slate-100 p-4">
						<div>
							<h2 class="text-base font-semibold text-slate-900">Tabel Ringkas Ruang</h2>
							<p class="text-xs text-slate-500">Tampilan padat untuk membandingkan ruang ujian dan status operasional.</p>
						</div>
						<Badge variant="outline">{filteredRooms.length} ruang</Badge>
					</div>
					<div class="overflow-x-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row class="bg-green-50">
									<Table.Head>Ruang</Table.Head>
									<Table.Head>Sesi</Table.Head>
									<Table.Head>Status</Table.Head>
									<Table.Head class="text-center">Peserta</Table.Head>
									<Table.Head class="text-center">Online</Table.Head>
									<Table.Head class="text-center">Atensi</Table.Head>
									<Table.Head>Pengawas</Table.Head>
									<Table.Head class="text-right">Aksi</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each filteredRooms as room (room.id)}
									<Table.Row>
										<Table.Cell>
											<div class="font-medium text-slate-900">{room.room_name}</div>
											<div class="font-mono text-xs text-slate-500">Token {room.room_token || '—'}</div>
										</Table.Cell>
										<Table.Cell>
											<div class="font-medium text-slate-800">{room.session_title}</div>
											<div class="text-xs text-slate-500">{fmtDate(room.scheduled_start)}</div>
										</Table.Cell>
										<Table.Cell><Badge variant="outline" class={statusClass(room.session_status)}>{statusLabel(room.session_status)}</Badge></Table.Cell>
										<Table.Cell class="text-center font-mono">{room.participant_count}</Table.Cell>
										<Table.Cell class="text-center font-mono text-emerald-700">{room.online_count}</Table.Cell>
										<Table.Cell class="text-center font-mono text-amber-700">{room.suspicious_count + room.missing_seat_count}</Table.Cell>
										<Table.Cell class="max-w-64 truncate text-sm text-slate-600">{room.proctor_names || '—'}</Table.Cell>
										<Table.Cell class="text-right">
											<div class="flex flex-wrap justify-end gap-2">
												<Button size="sm" href={resolve(`/asesmen/sesi/${room.session_id}/rooms/${room.id}/proctoring`)} variant="outline">Dashboard</Button>
												<Button size="sm" href={resolve(`/asesmen/sesi/${room.session_id}/rooms/${room.id}/print-pack`)} variant="outline">Cetak</Button>
											</div>
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</div>
				</section>
			{/if}
		</div>
	{/snippet}
</AsyncContent>
