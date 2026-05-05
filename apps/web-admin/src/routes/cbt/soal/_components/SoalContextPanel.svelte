<script lang="ts">
	type Subject = { id: string; name: string; code: string };
	type CbtEvent = { id: string; title: string; status?: string };

	let {
		events,
		subjects,
		selectedEventId,
		filterSubject,
		selectedEventTitle,
		roleLabel,
		onEventChange,
		onSubjectChange
	}: {
		events: CbtEvent[];
		subjects: Subject[];
		selectedEventId: string;
		filterSubject: string;
		selectedEventTitle: string;
		roleLabel: string;
		onEventChange: (eventId: string) => void;
		onSubjectChange: (subjectId: string) => void;
	} = $props();
</script>

<section class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
	<div class="grid gap-3 lg:grid-cols-[minmax(0,1fr)_minmax(0,46rem)] lg:items-end">
		<div class="min-w-0">
			<p class="text-xs font-bold uppercase tracking-wider text-green-700">Konteks Bank Soal</p>
			<h2 class="mt-1 text-base font-semibold text-slate-900">{selectedEventTitle}</h2>
			<p class="mt-1 text-xs text-slate-500">Bank Soal tetap reusable. Kegiatan hanya dipakai untuk target kebutuhan dan penugasan, bukan untuk menyaring repositori soal.</p>
		</div>
		<div class="grid gap-2 md:grid-cols-3">
			<div>
				<label for="event-context" class="mb-1 block text-xs font-medium text-slate-600">Konteks kegiatan (opsional)</label>
				<select
					id="event-context"
					value={selectedEventId}
					onchange={(event) => onEventChange((event.currentTarget as HTMLSelectElement).value)}
					class="h-9 w-full rounded-md border border-slate-200 bg-white px-2.5 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-green-500"
				>
					<option value="">Tanpa kegiatan - bank reusable</option>
					{#each events as event (event.id)}
						<option value={event.id}>{event.title}{event.status ? ` · ${event.status}` : ''}</option>
					{/each}
				</select>
			</div>
			<div>
				<label for="subject-context" class="mb-1 block text-xs font-medium text-slate-600">Mapel</label>
				<select
					id="subject-context"
					value={filterSubject}
					onchange={(event) => onSubjectChange((event.currentTarget as HTMLSelectElement).value)}
					class="h-9 w-full rounded-md border border-slate-200 bg-white px-2.5 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-green-500"
				>
					<option value="">Semua Mapel</option>
					{#each subjects as subject (subject.id)}
						<option value={subject.id}>{subject.name}</option>
					{/each}
				</select>
			</div>
			<div>
				<span id="role-context-label" class="mb-1 block text-xs font-medium text-slate-600">Peran saya</span>
				<div aria-describedby="role-context-label" class="flex h-9 items-center rounded-md border border-emerald-100 bg-emerald-50 px-2.5 text-sm font-semibold text-emerald-900">
					{roleLabel}
				</div>
			</div>
		</div>
	</div>
	{#if !selectedEventId}
		<div class="mt-3 rounded-lg border border-dashed border-emerald-200 bg-emerald-50/70 px-3 py-3 text-sm text-emerald-950">
			<p class="font-semibold">Mode bank reusable aktif.</p>
			<p class="mt-1 text-xs leading-5 text-green-800">Gunakan tanpa kegiatan untuk menyusun stok soal lintas paket. Pilih kegiatan hanya saat perlu melihat target kebutuhan atau penugasan event tertentu.</p>
		</div>
	{/if}
</section>
