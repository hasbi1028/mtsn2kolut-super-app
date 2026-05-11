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

<section class="rounded-xl border border-border bg-card/80 p-3 shadow-sm">
	<div class="grid gap-3 lg:grid-cols-[minmax(0,1fr)_minmax(0,44rem)] lg:items-end">
		<div class="min-w-0">
			<p class="text-[11px] font-bold uppercase tracking-wider text-muted-foreground">Konteks opsional</p>
			<h2 class="mt-1 truncate text-sm font-semibold text-foreground">{selectedEventTitle}</h2>
			<p class="mt-1 text-xs text-muted-foreground">Kegiatan membantu target kebutuhan; daftar soal tetap menjadi bank soal pakai ulang.</p>
		</div>
		<div class="grid gap-2 md:grid-cols-3">
			<div>
				<label for="event-context" class="mb-1 block text-xs font-medium text-muted-foreground">Konteks kegiatan (opsional)</label>
				<select
					id="event-context"
					value={selectedEventId}
					onchange={(event) => onEventChange((event.currentTarget as HTMLSelectElement).value)}
					class="h-9 w-full rounded-md border border-border bg-card px-2.5 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
				>
					<option value="">Tanpa kegiatan - bank pakai ulang</option>
					{#each events as event (event.id)}
						<option value={event.id}>{event.title}{event.status ? ` · ${event.status}` : ''}</option>
					{/each}
				</select>
			</div>
			<div>
				<label for="subject-context" class="mb-1 block text-xs font-medium text-muted-foreground">Mapel</label>
				<select
					id="subject-context"
					value={filterSubject}
					onchange={(event) => onSubjectChange((event.currentTarget as HTMLSelectElement).value)}
					class="h-9 w-full rounded-md border border-border bg-card px-2.5 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
				>
					<option value="">Semua Mapel</option>
					{#each subjects as subject (subject.id)}
						<option value={subject.id}>{subject.name}</option>
					{/each}
				</select>
			</div>
			<div>
				<span id="role-context-label" class="mb-1 block text-xs font-medium text-muted-foreground">Peran saya</span>
				<div aria-describedby="role-context-label" class="flex h-9 items-center rounded-md border border-primary/20 bg-primary/10 px-2.5 text-sm font-semibold text-primary">
					{roleLabel}
				</div>
			</div>
		</div>
	</div>
	{#if !selectedEventId}
		<div class="mt-3 rounded-lg border border-dashed border-border bg-muted/50 px-3 py-2 text-sm text-foreground">
			<p class="font-semibold">Mode bank pakai ulang aktif.</p>
			<p class="mt-1 text-xs leading-5 text-muted-foreground">Pilih kegiatan hanya saat perlu melihat target kebutuhan atau penugasan event tertentu.</p>
		</div>
	{/if}
</section>
