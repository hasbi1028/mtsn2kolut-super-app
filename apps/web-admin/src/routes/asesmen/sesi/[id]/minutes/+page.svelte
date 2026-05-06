<script lang="ts">
	import PrinterIcon from '@lucide/svelte/icons/printer';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { Button } from '$lib/components/ui/button';
	import { fetchSchoolProfile, schoolAddressLine, type SchoolProfile } from '$lib/school-profile';
	import { readClientApiData } from '$lib/client/api';

	type SessionInfo = {
		title: string;
		package_title: string;
		class_code: string;
		scope_type?: string;
		scope_ref?: string;
		scheduled_start: string;
		scheduled_end: string;
		status: string;
	};
	type Participant = {
		id: string;
		nis: string;
		nama: string;
		gender: string;
		room_name: string;
		seat_no: number | null;
		token: string;
		submitted_at: string | null;
	};
	type Room = {
		id: string;
		room_name: string;
		capacity: number;
		participant_count: number;
	};
	type MinutesPayload = {
		session?: SessionInfo | null;
		participants?: Participant[];
		rooms?: Room[];
	};
	type MinutesDetail = {
		session: SessionInfo;
		participants: Participant[];
		rooms: Room[];
	};
	type MinutesPrintData = {
		schoolProfile: SchoolProfile;
		detail: MinutesDetail;
	};

	const sessionId = page.params.id;
	let minutesPromise = $state<Promise<MinutesPrintData> | null>(null);

	async function fetchMinutes(): Promise<MinutesDetail> {
		const response = await fetch(`/api/asesmen/sessions/${sessionId}/minutes`);
		const payload = await readClientApiData<MinutesPayload>(response, 'Gagal memuat berita acara sesi');
		if (!payload.session) throw new Error('Data sesi tidak ditemukan');
		return {
			session: payload.session,
			participants: payload.participants ?? [],
			rooms: payload.rooms ?? [],
		};
	}

	async function fetchMinutesPrintData(): Promise<MinutesPrintData> {
		const [schoolProfile, detail] = await Promise.all([fetchSchoolProfile(), fetchMinutes()]);
		return { schoolProfile, detail };
	}

	function loadMinutes() {
		minutesPromise = fetchMinutesPrintData();
	}

	function retryMinutes(reset?: () => void) {
		reset?.();
		loadMinutes();
	}

	function minutesErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat berita acara sesi';
	}

	function handleMinutesRenderError(error: unknown) {
		console.error('CBT session minutes render failed', error);
	}

	function fmtDt(value: string) {
		return new Date(value).toLocaleString('id-ID', {
			timeZone: 'Asia/Makassar',
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		}) + ' WITA';
	}

	onMount(() => {
		void loadMinutes();
	});
</script>

<svelte:head>
	<title>Berita Acara Sesi</title>
</svelte:head>

<AsyncContent promise={minutesPromise} onerror={handleMinutesRenderError}>
	{#snippet pending()}
		<div class="mx-auto max-w-6xl space-y-6 p-6">
			<div class="space-y-2">
				<Skeleton class="h-8 w-56" />
				<Skeleton class="h-4 w-80" />
			</div>
			<Skeleton class="h-32 w-full rounded-lg" />
			<div class="grid gap-4 md:grid-cols-3">
				{#each Array.from({ length: 3 }) as _, index (`minutes-stat-skeleton-${index}`)}
					<Skeleton class="h-28 w-full rounded-lg" />
				{/each}
			</div>
			<Skeleton class="h-64 w-full rounded-lg" />
			<Skeleton class="h-80 w-full rounded-lg" />
		</div>
	{/snippet}

	{#snippet failed(error, reset)}
		<div class="mx-auto max-w-6xl p-6">
			<RecoveryPanel
				title="Berita Acara Belum Tersaji"
				message={minutesErrorMessage(error)}
				onRetry={() => retryMinutes(reset)}
			/>
		</div>
	{/snippet}

	{#snippet children(value)}
		{@const printData = value as MinutesPrintData}
		{@const detail = printData.detail}
		{@const schoolProfile = printData.schoolProfile}
		{@const currentSession = detail.session}
		{@const currentParticipants = detail.participants}
		{@const currentRooms = detail.rooms}
	<div class="mx-auto max-w-6xl space-y-6 p-6 print:p-0">
		<div class="flex items-center justify-between print:hidden">
			<div>
				<h1 class="text-2xl font-semibold text-foreground">Berita Acara Sesi Ujian</h1>
				<p class="text-sm text-muted-foreground">Siap dicetak untuk pengawas dan arsip madrasah.</p>
			</div>
			<Button onclick={() => window.print()}>
				<PrinterIcon class="mr-2 size-4" />
				Cetak
			</Button>
		</div>

		<section class="rounded-lg border border-primary/20 bg-card p-6 text-center shadow-sm">
			<p class="text-sm font-semibold uppercase text-foreground">{schoolProfile.ministry_line}</p>
			<p class="text-sm font-semibold uppercase text-foreground">{schoolProfile.office_line}</p>
			<h1 class="mt-1 text-xl font-bold uppercase text-foreground">{schoolProfile.name}</h1>
			<p class="mt-1 text-xs leading-5 text-muted-foreground">{schoolAddressLine(schoolProfile) || 'Alamat madrasah belum diisi'}</p>
			{#if schoolProfile.nsm || schoolProfile.npsn}
				<p class="text-xs text-muted-foreground">
					{#if schoolProfile.nsm}NSM {schoolProfile.nsm}{/if}
					{#if schoolProfile.nsm && schoolProfile.npsn} · {/if}
					{#if schoolProfile.npsn}NPSN {schoolProfile.npsn}{/if}
				</p>
			{/if}
		</section>

		<section class="rounded-lg border border-primary/20 bg-card p-6 shadow-sm">
			<h2 class="text-xl font-semibold text-foreground">{currentSession.title}</h2>
			<div class="mt-3 grid gap-2 text-sm text-foreground md:grid-cols-2">
				<p><span class="font-medium">Paket:</span> {currentSession.package_title}</p>
				<p><span class="font-medium">Kelas/Scope:</span> {currentSession.class_code || currentSession.scope_ref || currentSession.scope_type || '—'}</p>
				<p><span class="font-medium">Mulai:</span> {fmtDt(currentSession.scheduled_start)}</p>
				<p><span class="font-medium">Selesai:</span> {fmtDt(currentSession.scheduled_end)}</p>
			</div>
		</section>

		<section class="grid gap-4 md:grid-cols-3">
			<div class="rounded-lg border border-primary/20 bg-card p-5 shadow-sm">
				<p class="text-sm text-muted-foreground">Total peserta</p>
				<p class="mt-2 text-3xl font-bold text-primary">{currentParticipants.length}</p>
			</div>
			<div class="rounded-lg border border-primary/20 bg-card p-5 shadow-sm">
				<p class="text-sm text-muted-foreground">Sudah submit</p>
				<p class="mt-2 text-3xl font-bold text-primary">{currentParticipants.filter((p) => p.submitted_at).length}</p>
			</div>
			<div class="rounded-lg border border-primary/20 bg-card p-5 shadow-sm">
				<p class="text-sm text-muted-foreground">Ruangan aktif</p>
				<p class="mt-2 text-3xl font-bold text-primary">{currentRooms.length}</p>
			</div>
		</section>

		<section class="rounded-lg border border-primary/20 bg-card p-6 shadow-sm">
			<h3 class="text-lg font-semibold text-foreground">Distribusi Ruangan</h3>
			<div class="mt-4 overflow-x-auto">
				<table class="min-w-full text-sm">
					<thead class="bg-primary/10 text-left text-muted-foreground">
						<tr>
							<th class="px-3 py-2">Ruangan</th>
							<th class="px-3 py-2">Kapasitas</th>
							<th class="px-3 py-2">Terisi</th>
						</tr>
					</thead>
					<tbody>
						{#each currentRooms as room (room.id)}
							<tr class="border-t">
								<td class="px-3 py-2">{room.room_name}</td>
								<td class="px-3 py-2">{room.capacity}</td>
								<td class="px-3 py-2">{room.participant_count}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</section>

		<section class="rounded-lg border border-primary/20 bg-card p-6 shadow-sm">
			<h3 class="text-lg font-semibold text-foreground">Daftar Hadir dan Token</h3>
			<div class="mt-4 overflow-x-auto">
				<table class="min-w-full text-sm">
					<thead class="bg-primary/10 text-left text-muted-foreground">
						<tr>
							<th class="px-3 py-2">NIS</th>
							<th class="px-3 py-2">Nama</th>
							<th class="px-3 py-2">Ruangan</th>
							<th class="px-3 py-2">No Meja</th>
							<th class="px-3 py-2">Token</th>
							<th class="px-3 py-2">Paraf</th>
						</tr>
					</thead>
					<tbody>
						{#each currentParticipants as participant (participant.id)}
							<tr class="border-t">
								<td class="px-3 py-2 font-mono">{participant.nis}</td>
								<td class="px-3 py-2">{participant.nama}</td>
								<td class="px-3 py-2">{participant.room_name || '—'}</td>
								<td class="px-3 py-2">{participant.seat_no ?? '—'}</td>
								<td class="px-3 py-2 font-mono">{participant.token || '—'}</td>
								<td class="px-3 py-2">&nbsp;</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</section>

		<section class="grid gap-8 rounded-lg border border-border bg-card p-6 text-sm text-foreground md:grid-cols-2">
			<div class="text-center">
				<p>Mengetahui,</p>
				<p>Kepala Madrasah</p>
				<div class="mt-16 border-t border-foreground pt-2">
					<p class="font-semibold">{schoolProfile.head_name || '........................................'}</p>
					<p>NIP. {schoolProfile.head_nip || '................................'}</p>
				</div>
			</div>
			<div class="text-center">
				<p>Pengawas Ruang,</p>
				<p>{currentSession.class_code || currentSession.scope_ref || '................................'}</p>
				<div class="mt-16 border-t border-foreground pt-2">
					<p class="font-semibold">........................................</p>
					<p>NIP. ................................</p>
				</div>
			</div>
		</section>
	</div>
	{/snippet}
</AsyncContent>
