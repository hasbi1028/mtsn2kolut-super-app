<script lang="ts">
	import { advancedPrototypeLinks, prototypeLanes, type PrototypeStatus } from './asesmen-prototype.model';

	const statusCopy: Record<PrototypeStatus, string> = {
		siap: 'Siap',
		'perlu-dicek': 'Perlu dicek',
		menunggu: 'Menunggu data'
	};

	const statusClass: Record<PrototypeStatus, string> = {
		siap: 'border-emerald-200 bg-emerald-50 text-emerald-800',
		'perlu-dicek': 'border-amber-200 bg-amber-50 text-amber-900',
		menunggu: 'border-slate-200 bg-slate-50 text-slate-600'
	};
</script>

<svelte:head>
	<title>Prototype UI Asesmen — MTsN 2 Kolut</title>
</svelte:head>

<div class="space-y-4">
	<section class="rounded-2xl border border-border bg-card px-4 py-3 shadow-sm">
		<div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
			<div class="min-w-0">
				<p class="text-[11px] font-black uppercase tracking-[0.22em] text-primary">PROTOTYPE UI — belum terhubung backend</p>
				<h1 class="mt-1 text-2xl font-black tracking-tight text-foreground">A0 · Asesmen / Ujian Digital</h1>
				<p class="mt-2 max-w-3xl text-sm leading-6 text-muted-foreground">
					Alur sederhana untuk menyiapkan, menjalankan, mengawasi, dan menutup ujian. Tombol di halaman ini hanya rancangan alur. Belum menjalankan aksi data.
				</p>
			</div>
			<div class="flex shrink-0 flex-wrap gap-2 text-[11px] font-bold text-muted-foreground">
				<span class="rounded-full border border-border bg-muted/40 px-3 py-1">Micro Workflow List</span>
				<span class="rounded-full border border-border bg-muted/40 px-3 py-1">5 langkah utama</span>
				<span class="rounded-full border border-border bg-muted/40 px-3 py-1">frontend-only</span>
				<span class="rounded-full border border-border bg-muted/40 px-3 py-1">tidak mengubah data</span>
			</div>
		</div>
	</section>

	<section class="grid gap-3 lg:grid-cols-[12.5rem_1fr]">
		<aside class="h-fit rounded-2xl border border-border bg-muted/25 p-3 shadow-sm lg:sticky lg:top-3">
			<p class="text-[11px] font-black uppercase tracking-[0.18em] text-muted-foreground">Alur</p>
			<nav class="mt-3 grid gap-1.5" aria-label="Navigasi prototype Asesmen">
				{#each prototypeLanes as lane (lane.id)}
					<a class="rounded-xl border border-border bg-card px-3 py-2 text-sm font-bold text-foreground hover:border-primary/40 hover:bg-primary/5" href={`#${lane.id}`}>
						{lane.code} · {lane.title}
					</a>
				{/each}
			</nav>
			<div id="a9-mode-lengkap" class="mt-3 rounded-xl border border-dashed border-border bg-card p-3">
				<p class="text-[11px] font-black uppercase tracking-[0.16em] text-muted-foreground">A9 · Mode Lengkap</p>
				<p class="mt-1 text-xs leading-5 text-muted-foreground">Untuk panitia/operator yang perlu membuka halaman teknis lama.</p>
				<div class="mt-2 flex flex-wrap gap-1.5">
					{#each advancedPrototypeLinks as label (label)}
						<span class="rounded-full border border-border bg-muted/30 px-2 py-1 text-[11px] font-semibold text-foreground">{label}</span>
					{/each}
				</div>
			</div>
		</aside>

		<div class="space-y-2.5">
			{#each prototypeLanes as lane (lane.id)}
				<article id={lane.id} class="scroll-mt-3 rounded-2xl border border-border bg-card p-3 shadow-sm">
					<div class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
						<div class="min-w-0">
							<p class="text-[11px] font-black uppercase tracking-[0.18em] text-primary">{lane.code} · {lane.audience}</p>
							<h2 class="mt-0.5 text-lg font-black text-foreground">{lane.title}</h2>
							<p class="mt-1 text-sm leading-6 text-muted-foreground">{lane.description}</p>
						</div>
						<span class={`w-fit shrink-0 rounded-full border px-2.5 py-1 text-[11px] font-black ${statusClass[lane.status]}`}>{statusCopy[lane.status]}</span>
					</div>

					<div class="mt-3 flex flex-wrap gap-1.5">
						{#each lane.items as item (item)}
							<span class="rounded-lg border border-border bg-muted/35 px-2.5 py-1 text-xs font-semibold text-foreground">{item}</span>
						{/each}
					</div>

					<div class="mt-3 flex flex-wrap gap-2">
						<button class="rounded-xl bg-primary px-3 py-2 text-xs font-black text-primary-foreground shadow-sm" type="button">{lane.primaryAction}</button>
						{#if lane.secondaryAction}
							<button class="rounded-xl border border-border bg-background px-3 py-2 text-xs font-black text-foreground" type="button">{lane.secondaryAction}</button>
						{/if}
					</div>
				</article>
			{/each}
		</div>
	</section>
</div>
