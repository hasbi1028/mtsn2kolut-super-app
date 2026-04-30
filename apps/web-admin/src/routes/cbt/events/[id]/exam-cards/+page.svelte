<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';

	type ExamCard = {
		event_id: string;
		event_title: string;
		exam_type: string;
		event_scope: string;
		academic_year_name: string;
		session_id: string;
		session_title: string;
		scheduled_start: string;
		participant_id: string;
		token: string;
		seat_no: number | null;
		nis: string;
		student_nama: string;
		gender: string;
		class_code: string;
		room_name: string;
	};

	const eventId = page.params.id;
	let cards = $state<ExamCard[]>([]);
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
		const res = await fetch(`/api/cbt/events/${eventId}/exam-cards`);
		if (res.ok) {
			cards = await res.json();
		}
		loading = false;
	});
</script>

<svelte:head>
	<title>Kartu Ujian Event</title>
</svelte:head>

{#if loading}
	<div class="p-6 text-sm text-slate-500">Memuat kartu ujian...</div>
{:else}
	<div class="mx-auto max-w-7xl space-y-6 p-6 print:p-0">
		<div class="flex items-center justify-between print:hidden">
			<div>
				<h1 class="text-2xl font-semibold text-slate-900">Kartu Ujian Event</h1>
				<p class="text-sm text-slate-500">Cetak per peserta dengan token, ruangan, dan nomor meja.</p>
			</div>
			<button class="rounded-md border border-input px-3 py-2 text-sm font-medium" onclick={() => window.print()}>
				Cetak
			</button>
		</div>

		<div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
			{#each cards as card (card.participant_id)}
				<article class="break-inside-avoid rounded-3xl border border-emerald-200 bg-white p-5 shadow-sm print:shadow-none">
					<div class="border-b border-dashed border-emerald-200 pb-3">
						<p class="text-xs font-semibold uppercase tracking-[0.2em] text-emerald-700">MTs Negeri 2 Kolaka Utara</p>
						<h2 class="mt-2 text-lg font-semibold text-slate-900">{card.event_title}</h2>
						<p class="text-sm text-slate-500">{card.session_title}</p>
					</div>
					<div class="mt-4 space-y-2 text-sm text-slate-700">
						<p><span class="font-medium">Nama:</span> {card.student_nama}</p>
						<p><span class="font-medium">NIS:</span> {card.nis}</p>
						<p><span class="font-medium">Kelas:</span> {card.class_code || '—'}</p>
						<p><span class="font-medium">Ruangan:</span> {card.room_name || 'Belum ditentukan'}</p>
						<p><span class="font-medium">No Meja:</span> {card.seat_no ?? '—'}</p>
						<p><span class="font-medium">Jadwal:</span> {fmtDt(card.scheduled_start)}</p>
					</div>
					<div class="mt-5 rounded-2xl bg-emerald-50 px-4 py-3">
						<p class="text-xs uppercase tracking-[0.2em] text-emerald-700">Token Ujian</p>
						<p class="mt-1 font-mono text-2xl font-bold text-emerald-950">{card.token || 'Belum digenerate'}</p>
					</div>
				</article>
			{/each}
		</div>
	</div>
{/if}
