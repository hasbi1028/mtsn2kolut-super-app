<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { clientApiPath, clientApiPathWithQuery, readClientApiData, readClientJson } from '$lib/client/api';
	import { proctorEventLabel } from '$lib/cbt/proctor-evidence';

	type ProctoringRow = {
		participant_id: string;
		nis: string;
		nama: string;
		room_id?: string | null;
		room_name: string;
		seat_no?: number | null;
		submitted_at: string | null;
		last_heartbeat: string | null;
		app_switch_count: number;
		screenshot_attempt: number;
		suspicious_flag: boolean;
		violation_count?: number;
		risk_score?: number;
		risk_level?: string;
		locked_at?: string | null;
		locked_reason?: string | null;
		recent_violation_count?: number;
		last_violation_at?: string | null;
		last_violation_reason?: string;
		answered_count: number;
		score: string | null;
	};
	type ProctoringEvent = {
		id: string;
		participant_id: string;
		student_id: string;
		nis: string;
		nama: string;
		room_id?: string | null;
		room_name: string;
		event_type: string;
		event_data: unknown;
		created_at: string;
	};
	type RoomSummary = {
		room_id: string;
		room_name: string;
		participant_count: number;
		online_count: number;
		offline_count: number;
		submitted_count: number;
		warning_count: number;
		high_count: number;
		locked_count: number;
	};

	const sessionId = page.params.id ?? '';
	let dashboardPromise = $state<Promise<ProctoringRow[]> | null>(null);
	let participants = $state<ProctoringRow[]>([]);
	let events = $state<ProctoringEvent[]>([]);
	let recentAlertEvents = $state<ProctoringEvent[]>([]);
	let seenEventIds = $state(new Set<string>());
	let liveSource = $state<EventSource | null>(null);
	let interval: ReturnType<typeof setInterval> | undefined;
	let liveMode = $state<'connecting' | 'sse' | 'polling'>('connecting');
	let filter = $state<'all' | 'warning' | 'high' | 'locked' | 'offline'>('all');
	let audioAlertsEnabled = $state(false);
	let actionBusyId = $state('');
	let hasPrimedEvents = false;

	let roomSummaries = $derived.by<RoomSummary[]>(() => {
		const map = new Map<string, RoomSummary>();
		for (const row of participants) {
			const key = row.room_id || row.room_name || 'unassigned';
			const room = map.get(key) ?? {
				room_id: row.room_id || '',
				room_name: row.room_name || 'Tanpa ruang',
				participant_count: 0,
				online_count: 0,
				offline_count: 0,
				submitted_count: 0,
				warning_count: 0,
				high_count: 0,
				locked_count: 0,
			};
			room.participant_count += 1;
			if (heartbeatState(row) === 'online') room.online_count += 1;
			if (heartbeatState(row) === 'offline') room.offline_count += 1;
			if (row.submitted_at) room.submitted_count += 1;
			if (riskLevel(row) === 'warning') room.warning_count += 1;
			if (riskLevel(row) === 'high') room.high_count += 1;
			if (riskLevel(row) === 'locked') room.locked_count += 1;
			map.set(key, room);
		}
		return [...map.values()].sort((a, b) => (b.locked_count + b.high_count + b.warning_count) - (a.locked_count + a.high_count + a.warning_count));
	});

	let filteredParticipants = $derived.by(() => participants.filter((row) => {
		if (filter === 'warning') return riskLevel(row) === 'warning';
		if (filter === 'high') return riskLevel(row) === 'high';
		if (filter === 'locked') return riskLevel(row) === 'locked';
		if (filter === 'offline') return heartbeatState(row) === 'offline';
		return true;
	}));

	onMount(() => {
		dashboardPromise = loadDashboard();
		void loadEvents();
		connectLiveStream();
		interval = setInterval(() => {
			if (typeof document !== 'undefined' && document.hidden) return;
			if (liveMode === 'sse') return;
			void loadDashboard();
			void loadEvents(true);
		}, 5000);
	});

	onDestroy(() => {
		if (interval) clearInterval(interval);
		liveSource?.close();
	});

	async function loadDashboard() {
		const res = await fetch(clientApiPath`/api/asesmen/sessions/${sessionId}/proctoring`);
		participants = await readClientApiData<ProctoringRow[]>(res, 'Gagal memuat command center pengawasan');
		return participants;
	}

	async function loadEvents(background = false) {
		try {
			const params = new URLSearchParams({ limit: '100' });
			const res = await fetch(clientApiPathWithQuery(`/api/asesmen/sessions/${sessionId}/proctoring/events`, params));
			const payload = await readClientApiData<ProctoringEvent[]>(res, 'Gagal memuat event pengawasan');
			if (hasPrimedEvents) notifyNewEvents(payload);
			else {
				primeSeenEvents(payload);
				hasPrimedEvents = true;
			}
			events = payload;
		} catch (error) {
			if (!background) toast.error(error instanceof Error ? error.message : 'Gagal memuat event pengawasan');
		}
	}

	function primeSeenEvents(items: ProctoringEvent[]) {
		const next = new Set(seenEventIds);
		for (const event of items) next.add(event.id);
		seenEventIds = next;
	}

	function notifyNewEvents(items: ProctoringEvent[]) {
		const known = new Set(seenEventIds);
		const fresh = items.filter((event) => event.id && !known.has(event.id));
		if (fresh.length === 0) return;
		for (const event of fresh) known.add(event.id);
		seenEventIds = known;
		for (const event of fresh.filter(importantEvent)) {
			recentAlertEvents = [event, ...recentAlertEvents.filter((item) => item.id !== event.id)].slice(0, 20);
			const message = `${event.room_name} · ${event.nama}: ${eventReason(event)}`;
			if (event.event_type === 'anti_cheat_violation') toast.warning(message);
			else toast.info(message);
			if (shouldPlayAlertSound(event)) playAlertSound();
		}
	}

	function handleLiveEvent(event: ProctoringEvent) {
		notifyNewEvents([event]);
		events = [event, ...events.filter((item) => item.id !== event.id)].slice(0, 100);
		void loadDashboard();
	}

	function connectLiveStream() {
		if (typeof EventSource === 'undefined') {
			liveMode = 'polling';
			return;
		}
		liveSource?.close();
		liveMode = 'connecting';
		const source = new EventSource(clientApiPath`/api/asesmen/sessions/${sessionId}/proctoring/stream`);
		liveSource = source;
		source.addEventListener('ready', () => liveMode = 'sse');
		source.addEventListener('proctor_event', (message) => {
			liveMode = 'sse';
			try {
				handleLiveEvent(JSON.parse((message as MessageEvent).data) as ProctoringEvent);
			} catch (error) {
				console.error('Invalid session proctoring event', error);
			}
		});
		source.onerror = () => {
			liveMode = 'polling';
			source.close();
			liveSource = null;
			setTimeout(() => {
				if (typeof document !== 'undefined' && document.hidden) return;
				connectLiveStream();
			}, 5000);
		};
	}

	function importantEvent(event: ProctoringEvent) {
		return ['anti_cheat_violation', 'app_switch', 'screenshot_attempt', 'proctor_force_submit', 'proctor_reset_access', 'proctor_unlock', 'proctor_acknowledge', 'proctor_incident_action', 'participant_command'].includes(event.event_type);
	}

	function eventReason(event: ProctoringEvent) {
		if (event.event_data && typeof event.event_data === 'object' && 'reason' in event.event_data) {
			const reason = (event.event_data as { reason?: unknown }).reason;
			if (typeof reason === 'string' && reason.trim()) return reason.replaceAll('_', ' ');
		}
		return proctorEventLabel(event);
	}


	function shouldPlayAlertSound(event: ProctoringEvent) {
		if (!audioAlertsEnabled) return false;
		if (event.event_type !== 'anti_cheat_violation') return event.event_type === 'app_switch' || event.event_type === 'screenshot_attempt';
		const reason = eventReason(event).toLowerCase();
		return reason.includes('locked') || reason.includes('high') || reason.includes('split') || reason.includes('switch');
	}

	function playAlertSound() {
		try {
			const Ctx = window.AudioContext || (window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
			if (!Ctx) return;
			const ctx = new Ctx();
			const oscillator = ctx.createOscillator();
			const gain = ctx.createGain();
			oscillator.frequency.setValueAtTime(880, ctx.currentTime);
			gain.gain.setValueAtTime(0.0001, ctx.currentTime);
			gain.gain.exponentialRampToValueAtTime(0.15, ctx.currentTime + 0.02);
			gain.gain.exponentialRampToValueAtTime(0.0001, ctx.currentTime + 0.28);
			oscillator.connect(gain).connect(ctx.destination);
			oscillator.start();
			oscillator.stop(ctx.currentTime + 0.3);
		} catch (error) {
			console.warn('Audio alert gagal diputar', error);
		}
	}

	async function unlockParticipant(row: ProctoringRow) {
		if (!row.room_id) return;
		const notes = window.prompt(`Catatan unlock untuk ${row.nama}`, 'Diverifikasi dari command center') ?? '';
		actionBusyId = `unlock-${row.participant_id}`;
		try {
			const res = await fetch(clientApiPath`/api/asesmen/sessions/${sessionId}/rooms/${row.room_id}/participants/${row.participant_id}/unlock`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ notes }),
			});
			await readClientJson<unknown>(res);
			toast.success(`${row.nama} sudah di-unlock`);
			await loadDashboard();
		} catch (error) {
			toast.error(error instanceof Error ? error.message : 'Gagal unlock peserta');
		} finally {
			actionBusyId = '';
		}
	}

	async function acknowledgeEvent(event: ProctoringEvent) {
		if (!event.room_id) return;
		const notes = window.prompt(`Catatan pemeriksaan untuk ${event.nama}`, eventReason(event)) ?? '';
		actionBusyId = `ack-${event.id}`;
		try {
			const res = await fetch(clientApiPath`/api/asesmen/sessions/${sessionId}/rooms/${event.room_id}/participants/${event.participant_id}/acknowledge`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ event_id: event.id, notes }),
			});
			await readClientJson<unknown>(res);
			toast.success(`${event.nama} ditandai sudah diperiksa`);
			await loadEvents(true);
		} catch (error) {
			toast.error(error instanceof Error ? error.message : 'Gagal menandai event');
		} finally {
			actionBusyId = '';
		}
	}

	function minutesSince(value: string | null | undefined) {
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

	function riskLevel(row: ProctoringRow) {
		if (row.locked_at || row.risk_level === 'locked') return 'locked';
		if (row.risk_level === 'high') return 'high';
		if (row.risk_level === 'warning') return 'warning';
		return 'normal';
	}

	function badgeClass(tone: string) {
		if (tone === 'locked' || tone === 'offline') return 'border-destructive/30 bg-destructive/10 text-destructive';
		if (tone === 'high' || tone === 'stale') return 'border-warning/30 bg-warning/10 text-warning';
		if (tone === 'warning') return 'border-accent bg-accent/60 text-accent-foreground';
		return 'border-primary/20 bg-primary/10 text-primary';
	}

	function liveModeLabel() {
		if (liveMode === 'sse') return 'Live connected';
		if (liveMode === 'connecting') return 'Menghubungkan live';
		return 'Fallback polling 5 detik';
	}

	function fmtDate(value: string | null | undefined) {
		if (!value) return '—';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value;
		return date.toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short' });
	}
</script>

<svelte:head>
	<title>Command Center Pengawasan CBT</title>
</svelte:head>

<div class="space-y-5 p-4 md:p-6">
	<div class="flex flex-col gap-3 border-b border-primary/20 pb-4 md:flex-row md:items-start md:justify-between">
		<div>
			<a href={resolve(`/asesmen/sesi/${sessionId}`)} class="text-xs font-semibold uppercase tracking-[0.18em] text-primary">Kembali ke detail sesi</a>
			<h1 class="mt-2 text-2xl font-bold tracking-tight text-foreground">Command Center Pengawasan</h1>
			<p class="text-sm text-muted-foreground">Pantau semua ruang, alert kecurangan, dan status peserta secara real-time.</p>
		</div>
		<div class="flex flex-wrap gap-2">
			<Button variant={audioAlertsEnabled ? 'default' : 'outline'} onclick={() => audioAlertsEnabled = !audioAlertsEnabled}>Audio {audioAlertsEnabled ? 'ON' : 'OFF'}</Button>
			<Button variant="outline" href={resolve(`/asesmen/sesi/${sessionId}/proctoring/report`)}>Rekap Insiden</Button>
			<Button variant="outline" onclick={() => void loadEvents(true)}>Refresh Event</Button>
		</div>
	</div>

	<AsyncContent promise={dashboardPromise}>
		{#snippet pending()}
			<div class="grid gap-3 md:grid-cols-3">
				{#each Array.from({ length: 3 }) as _, index (index)}<Skeleton class="h-28 rounded-lg" />{/each}
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel title="Command Center Belum Termuat" message={error instanceof Error ? error.message : 'Gagal memuat command center'} onRetry={() => { reset?.(); dashboardPromise = loadDashboard(); }} />
		{/snippet}

		<div class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
			<Card.Root><Card.Header><Card.Title class="text-sm">Peserta</Card.Title></Card.Header><Card.Content><div class="text-2xl font-bold">{participants.length}</div></Card.Content></Card.Root>
			<Card.Root><Card.Header><Card.Title class="text-sm">Warning/High</Card.Title></Card.Header><Card.Content><div class="text-2xl font-bold">{participants.filter((row) => riskLevel(row) === 'warning' || riskLevel(row) === 'high').length}</div></Card.Content></Card.Root>
			<Card.Root><Card.Header><Card.Title class="text-sm">Locked</Card.Title></Card.Header><Card.Content><div class="text-2xl font-bold text-destructive">{participants.filter((row) => riskLevel(row) === 'locked').length}</div></Card.Content></Card.Root>
			<Card.Root><Card.Header><Card.Title class="text-sm">Offline</Card.Title></Card.Header><Card.Content><div class="text-2xl font-bold">{participants.filter((row) => heartbeatState(row) === 'offline').length}</div></Card.Content></Card.Root>
		</div>

		{#if recentAlertEvents.length > 0}
			<Card.Root class="border-warning/30 bg-warning/5">
				<Card.Header><Card.Title class="text-base">Live Alert Lintas Ruang</Card.Title></Card.Header>
				<Card.Content class="grid gap-2 md:grid-cols-2 xl:grid-cols-3">
					{#each recentAlertEvents.slice(0, 9) as event (event.id)}
						<div class="rounded-lg border bg-card p-3 text-sm">
							<div class="font-semibold">{event.room_name} · {event.nama}</div>
							<div class="text-muted-foreground">{eventReason(event)} · {fmtDate(event.created_at)}</div>
							<div class="mt-2"><LoadingButton size="sm" variant="outline" onclick={() => void acknowledgeEvent(event)} loading={actionBusyId === `ack-${event.id}`} disabled={!event.room_id || (actionBusyId !== '' && actionBusyId !== `ack-${event.id}`)} loadingLabel="Simpan...">Tandai diperiksa</LoadingButton></div>
						</div>
					{/each}
				</Card.Content>
			</Card.Root>
		{/if}

		<Card.Root>
			<Card.Header><Card.Title>Ringkasan Ruang</Card.Title></Card.Header>
			<Card.Content class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
				{#each roomSummaries as room (room.room_id || room.room_name)}
					<div class="rounded-lg border p-3">
						<div class="flex items-start justify-between gap-2">
							<div class="font-semibold">{room.room_name}</div>
							{#if room.room_id}<Button size="sm" variant="outline" href={resolve(`/asesmen/sesi/${sessionId}/rooms/${room.room_id}/proctoring`)}>Buka</Button>{/if}
						</div>
						<div class="mt-2 flex flex-wrap gap-2 text-xs">
							<Badge variant="outline">{room.participant_count} peserta</Badge>
							<Badge variant="outline" class={badgeClass('normal')}>{room.online_count} online</Badge>
							<Badge variant="outline" class={badgeClass('warning')}>{room.warning_count} warning</Badge>
							<Badge variant="outline" class={badgeClass('high')}>{room.high_count} high</Badge>
							<Badge variant="outline" class={badgeClass('locked')}>{room.locked_count} locked</Badge>
						</div>
					</div>
				{/each}
			</Card.Content>
		</Card.Root>

		<Card.Root>
			<Card.Header class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
				<Card.Title>Peserta Bermasalah / Semua Peserta</Card.Title>
				<div class="flex flex-wrap gap-2">
					{#each ['all', 'warning', 'high', 'locked', 'offline'] as item (item)}
						<Button size="sm" variant={filter === item ? 'default' : 'outline'} onclick={() => filter = item as typeof filter}>{item}</Button>
					{/each}
				</div>
			</Card.Header>
			<Card.Content class="overflow-x-auto">
				<Table.Root>
					<Table.Header><Table.Row><Table.Head>Peserta</Table.Head><Table.Head>Ruang</Table.Head><Table.Head>Koneksi</Table.Head><Table.Head>Risk</Table.Head><Table.Head>Pelanggaran</Table.Head><Table.Head>Terakhir</Table.Head><Table.Head class="text-right">Aksi</Table.Head></Table.Row></Table.Header>
					<Table.Body>
						{#each filteredParticipants as row (row.participant_id)}
							<Table.Row>
								<Table.Cell><div class="font-semibold">{row.nama}</div><div class="text-xs text-muted-foreground">{row.nis}</div></Table.Cell>
								<Table.Cell>{row.room_name || '—'}</Table.Cell>
								<Table.Cell><Badge variant="outline" class={badgeClass(heartbeatState(row))}>{heartbeatState(row)}</Badge></Table.Cell>
								<Table.Cell><Badge variant="outline" class={badgeClass(riskLevel(row))}>{riskLevel(row)}</Badge></Table.Cell>
								<Table.Cell>{row.violation_count ?? 0} · skor {row.risk_score ?? 0}</Table.Cell>
								<Table.Cell>{row.last_violation_reason || fmtDate(row.last_violation_at)}</Table.Cell>
								<Table.Cell class="text-right"><div class="flex justify-end gap-2"><Button size="sm" variant="outline" href={row.room_id ? resolve(`/asesmen/sesi/${sessionId}/rooms/${row.room_id}/proctoring`) : undefined}>Ruang</Button><LoadingButton size="sm" variant="outline" onclick={() => void unlockParticipant(row)} loading={actionBusyId === `unlock-${row.participant_id}`} disabled={!row.locked_at || !row.room_id || (actionBusyId !== '' && actionBusyId !== `unlock-${row.participant_id}`)} loadingLabel="Unlock...">Unlock</LoadingButton></div></Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</Card.Content>
		</Card.Root>
	</AsyncContent>
</div>
