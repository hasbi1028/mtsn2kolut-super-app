<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { Skeleton } from '$lib/components/ui/skeleton';

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

	const sessionId = page.params.id;
	let session = $state<SessionInfo | null>(null);
	let participants = $state<Participant[]>([]);
	let rooms = $state<Room[]>([]);
	let loading = $state(true);

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

	onMount(async () => {
		const res = await fetch(`/api/cbt/sessions/${sessionId}/minutes`);
		if (res.ok) {
			const data = await res.json();
			session = data.session;
			participants = data.participants ?? [];
			rooms = data.rooms ?? [];
		}
		loading = false;
	});
</script>

<svelte:head>
	<title>Berita Acara Sesi</title>
</svelte:head>

{#if loading}
	<div class="mx-auto max-w-6xl space-y-6 p-6">
		<div class="space-y-2">
			<Skeleton class="h-8 w-56" />
			<Skeleton class="h-4 w-80" />
		</div>
		<Skeleton class="h-32 w-full rounded-3xl" />
		<div class="grid gap-4 md:grid-cols-3">
			{#each Array.from({ length: 3 }) as _, index (`minutes-stat-skeleton-${index}`)}
				<Skeleton class="h-28 w-full rounded-3xl" />
			{/each}
		</div>
		<Skeleton class="h-64 w-full rounded-3xl" />
		<Skeleton class="h-80 w-full rounded-3xl" />
	</div>
{:else if session}
	<div class="mx-auto max-w-6xl space-y-6 p-6 print:p-0">
		<div class="flex items-center justify-between print:hidden">
			<div>
				<h1 class="text-2xl font-semibold text-slate-900">Berita Acara Sesi Ujian</h1>
				<p class="text-sm text-slate-500">Siap dicetak untuk pengawas dan arsip madrasah.</p>
			</div>
			<button class="rounded-md border border-input px-3 py-2 text-sm font-medium" onclick={() => window.print()}>
				Cetak
			</button>
		</div>

		<section class="rounded-3xl border border-emerald-200 bg-white p-6 shadow-sm">
			<h2 class="text-xl font-semibold text-slate-900">{session.title}</h2>
			<div class="mt-3 grid gap-2 text-sm text-slate-700 md:grid-cols-2">
				<p><span class="font-medium">Paket:</span> {session.package_title}</p>
				<p><span class="font-medium">Kelas/Scope:</span> {session.class_code || session.scope_ref || session.scope_type || '—'}</p>
				<p><span class="font-medium">Mulai:</span> {fmtDt(session.scheduled_start)}</p>
				<p><span class="font-medium">Selesai:</span> {fmtDt(session.scheduled_end)}</p>
			</div>
		</section>

		<section class="grid gap-4 md:grid-cols-3">
			<div class="rounded-3xl border border-emerald-200 bg-white p-5 shadow-sm">
				<p class="text-sm text-slate-500">Total peserta</p>
				<p class="mt-2 text-3xl font-bold text-emerald-950">{participants.length}</p>
			</div>
			<div class="rounded-3xl border border-emerald-200 bg-white p-5 shadow-sm">
				<p class="text-sm text-slate-500">Sudah submit</p>
				<p class="mt-2 text-3xl font-bold text-emerald-950">{participants.filter((p) => p.submitted_at).length}</p>
			</div>
			<div class="rounded-3xl border border-emerald-200 bg-white p-5 shadow-sm">
				<p class="text-sm text-slate-500">Ruangan aktif</p>
				<p class="mt-2 text-3xl font-bold text-emerald-950">{rooms.length}</p>
			</div>
		</section>

		<section class="rounded-3xl border border-emerald-200 bg-white p-6 shadow-sm">
			<h3 class="text-lg font-semibold text-slate-900">Distribusi Ruangan</h3>
			<div class="mt-4 overflow-x-auto">
				<table class="min-w-full text-sm">
					<thead class="bg-emerald-50 text-left text-slate-600">
						<tr>
							<th class="px-3 py-2">Ruangan</th>
							<th class="px-3 py-2">Kapasitas</th>
							<th class="px-3 py-2">Terisi</th>
						</tr>
					</thead>
					<tbody>
						{#each rooms as room}
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

		<section class="rounded-3xl border border-emerald-200 bg-white p-6 shadow-sm">
			<h3 class="text-lg font-semibold text-slate-900">Daftar Hadir dan Token</h3>
			<div class="mt-4 overflow-x-auto">
				<table class="min-w-full text-sm">
					<thead class="bg-emerald-50 text-left text-slate-600">
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
						{#each participants as participant}
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
	</div>
{/if}
