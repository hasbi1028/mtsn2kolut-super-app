<script lang="ts">
	import PrinterIcon from '@lucide/svelte/icons/printer';
	import { onDestroy, onMount } from 'svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import OperationStatusPanel from '$lib/components/OperationStatusPanel.svelte';
	import { confirmAction } from '$lib/confirm-dialog';
	import { clientApiPath, readClientApiData, readClientJson } from '$lib/client/api';

	type RoomDashboard = {
		id: string;
		session_id: string;
		room_name: string;
		capacity: number;
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
	};
	type RoomProctor = {
		id: string;
		exam_room_id: string;
		employee_id: string;
		nip: string;
		nama: string;
		role: string;
		assigned_at: string;
	};
	type ProctoringRow = {
		participant_id: string;
		student_id: string;
		nis: string;
		nama: string;
		token: string;
		room_id: string | null;
		room_name: string;
		seat_no?: number | null;
		submitted_at: string | null;
		last_heartbeat: string | null;
		app_switch_count: number;
		screenshot_attempt: number;
		suspicious_flag: boolean;
		answered_count: number;
		score: string | null;
	};
	type ProctoringEvent = {
		id: string;
		participant_id: string;
		student_id: string;
		nis: string;
		nama: string;
		room_id: string | null;
		room_name: string;
		event_type: string;
		event_data: unknown;
		created_at: string;
	};
	type DashboardPayload = {
		room: RoomDashboard;
		proctors: RoomProctor[];
		participants: ProctoringRow[];
		events: ProctoringEvent[];
	};

	const sessionId = page.params.id ?? '';
	const roomId = page.params.rid ?? '';

	let dashboardPromise = $state<Promise<DashboardPayload> | null>(null);
	let room = $state<RoomDashboard | null>(null);
	let proctors = $state<RoomProctor[]>([]);
	let participants = $state<ProctoringRow[]>([]);
	let events = $state<ProctoringEvent[]>([]);
	let refreshBusy = $state(false);
	let backgroundBusy = $state(false);
	let actionBusyId = $state('');
	let operationState = $state<{ tone: 'success' | 'error' | 'warning' | 'info'; title: string; message: string } | null>(null);
	let interval: ReturnType<typeof setInterval> | undefined;

	let participantStats = $derived.by(() => {
		let online = 0;
		let stale = 0;
		let offline = 0;
		let submitted = 0;
		for (const row of participants) {
			const state = heartbeatState(row);
			if (state === 'online') online += 1;
			else if (state === 'stale') stale += 1;
			else offline += 1;
			if (row.submitted_at) submitted += 1;
		}
		return { online, stale, offline, submitted };
	});

	onMount(() => {
		dashboardPromise = loadDashboard();
		interval = setInterval(() => {
			void refreshDashboard(true);
		}, 15000);
	});

	onDestroy(() => {
		if (interval) clearInterval(interval);
	});

	async function loadDashboard() {
		const res = await fetch(clientApiPath`/api/cbt/sessions/${sessionId}/rooms/${roomId}/proctoring`);
		const payload = await readClientApiData<DashboardPayload>(res, 'Gagal memuat dashboard pengawas ruang');
		room = payload.room;
		proctors = Array.isArray(payload.proctors) ? payload.proctors : [];
		participants = Array.isArray(payload.participants) ? payload.participants : [];
		events = Array.isArray(payload.events) ? payload.events : [];
		return payload;
	}

	async function refreshDashboard(background = false) {
		if (background) backgroundBusy = true;
		else refreshBusy = true;
		try {
			await loadDashboard();
		} catch (error) {
			if (!background) {
				const message = detailErrorMessage(error);
				operationState = { tone: 'error', title: 'Refresh Gagal', message };
				toast.error(message);
			}
		} finally {
			refreshBusy = false;
			backgroundBusy = false;
		}
	}

	function detailErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Operasi belum berhasil. Periksa koneksi lalu coba lagi.';
	}

	function fmtDate(value: string | null | undefined) {
		if (!value) return '—';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value;
		return date.toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short' });
	}

	function fmtScore(value: string | null) {
		if (!value) return '—';
		const parsed = Number(value);
		if (Number.isNaN(parsed)) return value;
		return parsed.toFixed(2);
	}

	function minutesSince(value: string | null) {
		if (!value) return Number.POSITIVE_INFINITY;
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return Number.POSITIVE_INFINITY;
		return (Date.now() - date.getTime()) / 60000;
	}

	function heartbeatState(row: ProctoringRow): 'submitted' | 'online' | 'stale' | 'offline' {
		if (row.submitted_at) return 'submitted';
		const minutes = minutesSince(row.last_heartbeat);
		if (minutes <= 2) return 'online';
		if (minutes <= 7) return 'stale';
		return 'offline';
	}

	function heartbeatLabel(row: ProctoringRow) {
		const state = heartbeatState(row);
		if (state === 'submitted') return 'Submit';
		if (state === 'online') return 'Online';
		if (state === 'stale') return 'Waspada';
		return 'Offline';
	}

	function heartbeatClass(row: ProctoringRow) {
		const state = heartbeatState(row);
		if (state === 'submitted') return 'border-slate-300 bg-slate-100 text-slate-600';
		if (state === 'online') return 'border-emerald-300 bg-emerald-50 text-emerald-700';
		if (state === 'stale') return 'border-amber-300 bg-amber-50 text-amber-700';
		return 'border-red-300 bg-red-50 text-red-700';
	}

	function eventLabel(type: string) {
		const labels: Record<string, string> = {
			login: 'Login',
			heartbeat: 'Heartbeat',
			app_switch: 'Keluar Aplikasi',
			screenshot_attempt: 'Percobaan Screenshot',
			resume: 'Kembali Ujian',
			answer_save: 'Simpan Jawaban',
			submit: 'Submit',
			proctor_reset_access: 'Reset Akses',
			proctor_force_submit: 'Paksa Submit',
		};
		return labels[type] ?? type.replaceAll('_', ' ');
	}

	async function flagParticipant(row: ProctoringRow, flag: boolean) {
		actionBusyId = `flag-${row.participant_id}`;
		try {
			const res = await fetch(clientApiPath`/api/cbt/sessions/${sessionId}/rooms/${roomId}/participants/${row.participant_id}/flag`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ flag }),
			});
			await readClientJson<unknown>(res);
			operationState = {
				tone: flag ? 'warning' : 'success',
				title: flag ? 'Peserta Ditandai' : 'Tanda Dibersihkan',
				message: `${row.nama} ${flag ? 'masuk daftar atensi pengawas.' : 'tidak lagi ditandai sebagai atensi.'}`,
			};
			await refreshDashboard(true);
		} catch (error) {
			toast.error(detailErrorMessage(error));
		} finally {
			actionBusyId = '';
		}
	}

	async function resetAccess(row: ProctoringRow) {
		if (!(await confirmAction({
			title: 'Reset Akses Peserta',
			message: `Reset perangkat dan heartbeat ${row.nama}. Gunakan ini saat siswa perlu login ulang di ruang ini.`,
			confirmLabel: 'Reset Akses',
			tone: 'warning',
		}))) return;
		actionBusyId = `reset-${row.participant_id}`;
		try {
			const res = await fetch(clientApiPath`/api/cbt/sessions/${sessionId}/rooms/${roomId}/participants/${row.participant_id}/reset-access`, { method: 'POST' });
			await readClientJson<unknown>(res);
			operationState = { tone: 'success', title: 'Akses Direset', message: `${row.nama} dapat login ulang setelah diverifikasi pengawas.` };
			await refreshDashboard(true);
		} catch (error) {
			toast.error(detailErrorMessage(error));
		} finally {
			actionBusyId = '';
		}
	}

	async function forceSubmit(row: ProctoringRow) {
		if (!(await confirmAction({
			title: 'Paksa Submit Peserta',
			message: `Paksa submit jawaban ${row.nama}. Tindakan ini dipakai hanya saat ujian ruang sudah harus ditutup.`,
			confirmLabel: 'Paksa Submit',
			tone: 'danger',
		}))) return;
		actionBusyId = `submit-${row.participant_id}`;
		try {
			const res = await fetch(clientApiPath`/api/cbt/sessions/${sessionId}/rooms/${roomId}/participants/${row.participant_id}/force-submit`, { method: 'POST' });
			await readClientJson<unknown>(res);
			operationState = { tone: 'warning', title: 'Peserta Disubmit', message: `${row.nama} sudah dipaksa submit dari dashboard ruang.` };
			await refreshDashboard(true);
		} catch (error) {
			toast.error(detailErrorMessage(error));
		} finally {
			actionBusyId = '';
		}
	}

	function handleRenderError(error: unknown) {
		console.error('CBT room proctoring dashboard render failed', error);
	}
</script>

<svelte:head>
	<title>Dashboard Pengawas Ruang | CBT</title>
</svelte:head>

<div class="space-y-5 p-4 md:p-6">
	<div class="flex flex-col gap-3 border-b border-emerald-100 pb-4 md:flex-row md:items-start md:justify-between">
		<div>
			<a href={resolve(`/cbt/sessions/${sessionId}`)} class="text-xs font-semibold uppercase tracking-[0.18em] text-emerald-700">Kembali ke detail sesi</a>
			<h1 class="mt-2 text-2xl font-bold tracking-tight text-slate-900">Dashboard Pengawas Ruang</h1>
			<p class="text-sm text-slate-500">{room?.session_title ?? 'Memuat sesi'} · {room?.room_name ?? 'Memuat ruang'}</p>
		</div>
		<div class="flex flex-wrap items-center gap-2">
			{#if backgroundBusy}
				<Badge variant="outline" class="border-emerald-200 text-emerald-700">Memperbarui</Badge>
			{/if}
			<Button variant="outline" href={resolve(`/cbt/sessions/${sessionId}/rooms/${roomId}/print-pack`)}>
				<PrinterIcon class="mr-2 size-4" />
				Paket Cetak
			</Button>
			<LoadingButton variant="outline" onclick={() => void refreshDashboard()} loading={refreshBusy} loadingLabel="Memuat...">
				Refresh
			</LoadingButton>
		</div>
	</div>

	{#if operationState}
		<OperationStatusPanel {...operationState} />
	{/if}

	<AsyncContent promise={dashboardPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<div class="grid gap-4 lg:grid-cols-4">
				{#each Array.from({ length: 4 }) as _, index (index)}
					<Skeleton class="h-28 rounded-lg" />
				{/each}
			</div>
			<Skeleton class="h-96 rounded-lg" />
		{/snippet}

		{#if room}
				<div class="grid gap-3 md:grid-cols-2 xl:grid-cols-5">
					<Card.Root class="border-emerald-100">
						<Card.Header class="pb-2">
							<Card.Title class="text-sm text-slate-500">Peserta</Card.Title>
						</Card.Header>
						<Card.Content>
							<div class="text-2xl font-bold text-slate-900">{room.participant_count}</div>
							<p class="text-xs text-slate-500">Kapasitas {room.capacity}</p>
						</Card.Content>
					</Card.Root>
					<Card.Root class="border-emerald-100">
						<Card.Header class="pb-2">
							<Card.Title class="text-sm text-slate-500">Online</Card.Title>
						</Card.Header>
						<Card.Content>
							<div class="text-2xl font-bold text-emerald-700">{participantStats.online}</div>
							<p class="text-xs text-slate-500">Heartbeat 2 menit terakhir</p>
						</Card.Content>
					</Card.Root>
					<Card.Root class="border-emerald-100">
						<Card.Header class="pb-2">
							<Card.Title class="text-sm text-slate-500">Waspada / Offline</Card.Title>
						</Card.Header>
						<Card.Content>
							<div class="text-2xl font-bold text-amber-700">{participantStats.stale + participantStats.offline}</div>
							<p class="text-xs text-slate-500">Perlu dicek pengawas</p>
						</Card.Content>
					</Card.Root>
					<Card.Root class="border-emerald-100">
						<Card.Header class="pb-2">
							<Card.Title class="text-sm text-slate-500">Submit</Card.Title>
						</Card.Header>
						<Card.Content>
							<div class="text-2xl font-bold text-slate-900">{participantStats.submitted}</div>
							<p class="text-xs text-slate-500">Dari {room.participant_count} peserta</p>
						</Card.Content>
					</Card.Root>
					<Card.Root class="border-emerald-100">
						<Card.Header class="pb-2">
							<Card.Title class="text-sm text-slate-500">Atensi</Card.Title>
						</Card.Header>
						<Card.Content>
							<div class="text-2xl font-bold text-red-700">{room.suspicious_count}</div>
							<p class="text-xs text-slate-500">{room.missing_seat_count} tanpa nomor meja</p>
						</Card.Content>
					</Card.Root>
				</div>

				<Card.Root class="border-emerald-100">
					<Card.Header>
						<div class="flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
							<div>
								<Card.Title>{room.room_name}</Card.Title>
								<Card.Description>
									{room.package_title} · {fmtDate(room.scheduled_start)} - {fmtDate(room.scheduled_end)}
								</Card.Description>
							</div>
							<div class="flex flex-wrap items-center gap-2">
								<Badge variant="outline" class="border-emerald-300 text-emerald-700">Token ruang {room.room_token || '—'}</Badge>
								<Badge variant="outline">{room.session_status}</Badge>
							</div>
						</div>
					</Card.Header>
					<Card.Content class="grid gap-4 md:grid-cols-3">
						<div class="rounded-lg border border-slate-200 p-3">
							<p class="text-xs font-semibold uppercase tracking-wide text-slate-500">Lokasi</p>
							<p class="mt-1 text-sm font-medium text-slate-800">{room.school_room_code ? `${room.school_room_code} · ${room.school_room_name}` : 'Ruang manual sesi'}</p>
							<p class="text-xs text-slate-500">{room.school_room_building || room.school_room_location_note || 'Lokasi belum dicatat'}</p>
						</div>
						<div class="rounded-lg border border-slate-200 p-3">
							<p class="text-xs font-semibold uppercase tracking-wide text-slate-500">Pengawas</p>
							<div class="mt-1 space-y-1">
								{#each proctors as proctor (proctor.id)}
									<p class="text-sm text-slate-700">{proctor.nama} <span class="text-xs text-slate-400">({proctor.role})</span></p>
								{:else}
									<p class="text-sm text-amber-700">Belum ada pengawas</p>
								{/each}
							</div>
						</div>
						<div class="rounded-lg border border-slate-200 p-3">
							<p class="text-xs font-semibold uppercase tracking-wide text-slate-500">Status Ruang</p>
							<p class="mt-1 text-sm text-slate-700">Durasi paket {room.duration_minutes} menit</p>
							<p class="text-xs text-slate-500">{room.is_locked ? 'Ruang dikunci' : 'Ruang masih dapat diperbarui operator'}</p>
						</div>
					</Card.Content>
				</Card.Root>

				<div class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_380px]">
					<Card.Root class="border-emerald-100">
						<Card.Header>
							<Card.Title>Peserta Ruang</Card.Title>
							<Card.Description>Monitoring heartbeat, submit, dan tindakan pengawas terbatas pada ruang ini.</Card.Description>
						</Card.Header>
						<Card.Content class="overflow-x-auto">
							<Table.Root>
								<Table.Header>
									<Table.Row class="bg-green-50">
										<Table.Head>Peserta</Table.Head>
										<Table.Head class="text-center">Meja</Table.Head>
										<Table.Head>Status</Table.Head>
										<Table.Head class="text-center">Jawab</Table.Head>
										<Table.Head class="text-center">Switch</Table.Head>
										<Table.Head class="text-center">SS</Table.Head>
										<Table.Head class="text-center">Skor</Table.Head>
										<Table.Head class="text-right">Aksi</Table.Head>
									</Table.Row>
								</Table.Header>
								<Table.Body>
									{#each participants as row (row.participant_id)}
										<Table.Row class={row.suspicious_flag ? 'bg-red-50' : row.app_switch_count >= 3 ? 'bg-amber-50/60' : ''}>
											<Table.Cell>
												<div class="font-medium text-slate-900">{row.nama}</div>
												<div class="text-xs text-slate-500">{row.nis} · token {row.token}</div>
											</Table.Cell>
											<Table.Cell class="text-center font-mono">{row.seat_no ?? '—'}</Table.Cell>
											<Table.Cell>
												<Badge variant="outline" class={heartbeatClass(row)}>{heartbeatLabel(row)}</Badge>
												<p class="mt-1 text-[11px] text-slate-400">{fmtDate(row.last_heartbeat)}</p>
											</Table.Cell>
											<Table.Cell class="text-center font-mono">{row.answered_count}</Table.Cell>
											<Table.Cell class="text-center font-mono">{row.app_switch_count}</Table.Cell>
											<Table.Cell class="text-center font-mono">{row.screenshot_attempt}</Table.Cell>
											<Table.Cell class="text-center font-mono">{fmtScore(row.score)}</Table.Cell>
											<Table.Cell class="min-w-[260px] text-right">
												<div class="flex flex-wrap justify-end gap-2">
													<Button size="sm" variant="outline" onclick={() => void flagParticipant(row, !row.suspicious_flag)}>
														{row.suspicious_flag ? 'Bersihkan' : 'Tandai'}
													</Button>
													<LoadingButton size="sm" variant="outline" onclick={() => void resetAccess(row)} loading={actionBusyId === `reset-${row.participant_id}`} disabled={actionBusyId !== '' && actionBusyId !== `reset-${row.participant_id}`} loadingLabel="Reset...">
														Reset
													</LoadingButton>
													<LoadingButton size="sm" onclick={() => void forceSubmit(row)} loading={actionBusyId === `submit-${row.participant_id}`} disabled={!!row.submitted_at || (actionBusyId !== '' && actionBusyId !== `submit-${row.participant_id}`)} loadingLabel="Submit...">
														Submit
													</LoadingButton>
												</div>
											</Table.Cell>
										</Table.Row>
									{:else}
										<Table.Row>
											<Table.Cell colspan={8} class="py-10 text-center text-slate-400">Belum ada peserta di ruang ini</Table.Cell>
										</Table.Row>
									{/each}
								</Table.Body>
							</Table.Root>
						</Card.Content>
					</Card.Root>

					<Card.Root class="border-emerald-100">
						<Card.Header>
							<Card.Title>Log Ruang</Card.Title>
							<Card.Description>Aktivitas terakhir dari peserta ruang ini.</Card.Description>
						</Card.Header>
						<Card.Content class="space-y-3">
							{#each events.slice(0, 15) as event (event.id)}
								<div class="rounded-lg border border-slate-200 p-3">
									<div class="flex items-center justify-between gap-2">
										<p class="text-sm font-medium text-slate-800">{event.nama}</p>
										<Badge variant="outline" class="text-[11px]">{eventLabel(event.event_type)}</Badge>
									</div>
									<p class="mt-1 text-xs text-slate-500">{event.nis} · {fmtDate(event.created_at)}</p>
								</div>
							{:else}
								<p class="rounded-lg border border-dashed border-slate-200 p-5 text-center text-sm text-slate-400">Belum ada log ruang</p>
							{/each}
						</Card.Content>
					</Card.Root>
				</div>
		{/if}
	</AsyncContent>
</div>
