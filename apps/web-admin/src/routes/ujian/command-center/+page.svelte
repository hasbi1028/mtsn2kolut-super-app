<script lang="ts">
	import { examSlots, getFlatAssignments, getSupervisorLoad, getSupervisorName, getWarnings, roomCodes, toCsv } from './uas-genap-2026.model';

	const assignments = getFlatAssignments();
	const loads = getSupervisorLoad();
	const warnings = getWarnings();
	const totalRooms = assignments.length;
	const readySlots = examSlots.filter((slot) => slot.status !== 'needs_review').length;
	const reviewSlots = examSlots.length - readySlots;
	let selectedDate = $state('semua');

	const dates = $derived(Array.from(new Map(examSlots.map((slot) => [slot.date, `${slot.day}, ${formatDate(slot.date)}`])).entries()));
	const filteredSlots = $derived(selectedDate === 'semua' ? examSlots : examSlots.filter((slot) => slot.date === selectedDate));
	const minLoad = $derived(Math.min(...loads.map((item) => item.count)));
	const maxLoad = $derived(Math.max(...loads.map((item) => item.count)));

	function formatDate(date: string) {
		return new Intl.DateTimeFormat('id-ID', { day: '2-digit', month: 'long', year: 'numeric' }).format(new Date(`${date}T00:00:00+08:00`));
	}

	function downloadCsv() {
		const blob = new Blob([toCsv()], { type: 'text/csv;charset=utf-8' });
		const url = URL.createObjectURL(blob);
		const anchor = document.createElement('a');
		anchor.href = url;
		anchor.download = 'jadwal-pengawas-uas-genap-2026.csv';
		anchor.click();
		URL.revokeObjectURL(url);
	}
</script>

<svelte:head>
	<title>Command Center Ujian — UAS Genap 2026</title>
</svelte:head>

<div class="mx-auto flex max-w-7xl flex-col gap-6 p-4 sm:p-6">
	<section class="rounded-3xl border border-emerald-100 bg-gradient-to-br from-emerald-50 via-white to-amber-50 p-5 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
			<div>
				<p class="text-xs font-semibold uppercase tracking-[0.25em] text-emerald-700">Command Center Ujian</p>
				<h1 class="mt-2 text-2xl font-bold text-slate-950 sm:text-3xl">Evaluasi Semester Genap TP 2025/2026</h1>
				<p class="mt-2 max-w-3xl text-sm text-slate-600">
					Jadwal sudah dikoreksi mengikuti arahan panitia: Kamis 04 Juni, Jumat 05 Juni, Sabtu 06 Juni, Senin 08 Juni, Selasa 09 Juni, dan Rabu 10 Juni 2026. Tidak ada sesi pada Minggu 07 Juni.
				</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<button class="rounded-xl bg-emerald-700 px-4 py-2 text-sm font-semibold text-white shadow hover:bg-emerald-800" type="button" onclick={downloadCsv}>Unduh CSV Jadwal</button>
				<button class="rounded-xl border border-slate-300 bg-white px-4 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-50" type="button" onclick={() => print()}>Cetak / PDF</button>
			</div>
		</div>
	</section>

	<section class="grid gap-3 sm:grid-cols-2 lg:grid-cols-5">
		<div class="rounded-2xl border bg-white p-4 shadow-sm">
			<p class="text-xs font-medium text-slate-500">Slot ujian</p>
			<p class="mt-1 text-2xl font-bold text-slate-900">{examSlots.length}</p>
		</div>
		<div class="rounded-2xl border bg-white p-4 shadow-sm">
			<p class="text-xs font-medium text-slate-500">Penempatan ruang</p>
			<p class="mt-1 text-2xl font-bold text-slate-900">{totalRooms}</p>
		</div>
		<div class="rounded-2xl border bg-white p-4 shadow-sm">
			<p class="text-xs font-medium text-slate-500">Pengawas</p>
			<p class="mt-1 text-2xl font-bold text-slate-900">{loads.length}</p>
		</div>
		<div class="rounded-2xl border bg-white p-4 shadow-sm">
			<p class="text-xs font-medium text-slate-500">Siap / review</p>
			<p class="mt-1 text-2xl font-bold text-slate-900">{readySlots}/{reviewSlots}</p>
		</div>
		<div class="rounded-2xl border bg-white p-4 shadow-sm">
			<p class="text-xs font-medium text-slate-500">Beban min–maks</p>
			<p class="mt-1 text-2xl font-bold text-slate-900">{minLoad}–{maxLoad}</p>
		</div>
	</section>

	{#if warnings.length}
		<section class="rounded-2xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-900">
			<h2 class="font-semibold">Catatan pending koreksi</h2>
			<ul class="mt-2 list-disc space-y-1 pl-5">
				{#each warnings as warning}
					<li>{warning}</li>
				{/each}
			</ul>
			<p class="mt-2 text-xs">Catatan ini tidak memblokir import awal sesuai arahan panitia; koreksi nomor 23 dapat dilakukan manual besok sebelum final/pelaksanaan.</p>
		</section>
	{/if}

	<section class="rounded-2xl border bg-white p-4 shadow-sm">
		<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
			<div>
				<h2 class="text-lg font-semibold text-slate-900">Jadwal per tanggal dan ruang</h2>
				<p class="text-sm text-slate-500">Setiap angka pengawas sudah ditampilkan dengan nama lengkap dari foto.</p>
			</div>
			<label class="text-sm text-slate-600">
				Tanggal
				<select bind:value={selectedDate} class="ml-2 rounded-xl border border-slate-300 bg-white px-3 py-2 text-sm">
					<option value="semua">Semua</option>
					{#each dates as [date, label]}
						<option value={date}>{label}</option>
					{/each}
				</select>
			</label>
		</div>

		<div class="mt-4 grid gap-4">
			{#each filteredSlots as slot}
				<article class="overflow-hidden rounded-2xl border border-slate-200">
					<header class="flex flex-col gap-2 border-b bg-slate-50 p-4 sm:flex-row sm:items-center sm:justify-between">
						<div>
							<p class="text-xs font-medium uppercase tracking-wide text-slate-500">{slot.day}, {formatDate(slot.date)} • {slot.start}–{slot.end}</p>
							<h3 class="text-lg font-semibold text-slate-900">{slot.subject}</h3>
						</div>
						<span class={`w-fit rounded-full px-3 py-1 text-xs font-semibold ${slot.status === 'needs_review' ? 'bg-amber-100 text-amber-800' : 'bg-emerald-100 text-emerald-800'}`}>{slot.status === 'needs_review' ? 'Pending review' : 'Siap'}</span>
					</header>
					<div class="grid gap-2 p-4 sm:grid-cols-2 lg:grid-cols-4">
						{#each roomCodes as room}
							{@const code = slot.supervisors[room]}
							<div class="rounded-xl border border-slate-100 bg-white p-3">
								<p class="text-xs font-semibold text-slate-500">{room}</p>
								<p class="mt-1 text-sm font-semibold text-slate-950">#{code} {getSupervisorName(code)}</p>
							</div>
						{/each}
					</div>
				</article>
			{/each}
		</div>
	</section>

	<section class="rounded-2xl border bg-white p-4 shadow-sm">
		<h2 class="text-lg font-semibold text-slate-900">Rekap beban pengawas</h2>
		<div class="mt-4 grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
			{#each loads as item}
				<div class="flex items-center justify-between rounded-xl border border-slate-100 p-3">
					<div>
						<p class="text-sm font-semibold text-slate-900">#{item.code} {item.name}</p>
						<p class="text-xs text-slate-500">Kode pengawas foto: {item.code}</p>
					</div>
					<span class="rounded-full bg-slate-100 px-3 py-1 text-sm font-bold text-slate-700">{item.count}x</span>
				</div>
			{/each}
		</div>
	</section>
</div>
