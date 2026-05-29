<script lang="ts">
	import {
		advancedPrototypeLinks,
		hiddenFromMainFlow,
		prototypeLanes,
		prototypeMetrics,
		prototypeRooms,
		type PrototypeStatus,
		type PrototypeStepState
	} from './asesmen-prototype.model';

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

	const stepCopy: Record<PrototypeStepState, string> = {
		selesai: 'Selesai',
		lanjut: 'Lanjut',
		cek: 'Cek',
		opsional: 'Opsional'
	};

	const stepClass: Record<PrototypeStepState, string> = {
		selesai: 'border-emerald-200 bg-emerald-50 text-emerald-800',
		lanjut: 'border-sky-200 bg-sky-50 text-sky-800',
		cek: 'border-amber-200 bg-amber-50 text-amber-900',
		opsional: 'border-slate-200 bg-slate-50 text-slate-600'
	};

	const roomToneClass = {
		green: 'border-emerald-200 bg-emerald-50 text-emerald-900',
		amber: 'border-amber-200 bg-amber-50 text-amber-950',
		blue: 'border-sky-200 bg-sky-50 text-sky-900'
	};
</script>

<svelte:head>
	<title>Prototype UI Asesmen — MTsN 2 Kolut</title>
</svelte:head>

<div class="space-y-4">
	<section class="overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
		<div class="grid gap-0 lg:grid-cols-[1.35fr_0.85fr]">
			<div class="min-w-0 px-4 py-4 sm:px-5">
				<p class="text-[11px] font-black uppercase tracking-[0.22em] text-primary">PROTOTYPE UI — belum terhubung backend</p>
				<h1 class="mt-1 text-2xl font-black tracking-tight text-foreground sm:text-3xl">A0 · Asesmen / Ujian Digital</h1>
				<p class="mt-2 max-w-3xl text-sm leading-6 text-muted-foreground">
					Alur sederhana untuk menyiapkan, menjalankan, mengawasi, dan menutup ujian. Tombol di halaman ini hanya rancangan alur. Belum menjalankan aksi data.
				</p>
				<div class="mt-3 flex flex-wrap gap-2 text-[11px] font-bold text-muted-foreground">
					<span class="rounded-full border border-border bg-muted/40 px-3 py-1">Micro Workflow List</span>
					<span class="rounded-full border border-border bg-muted/40 px-3 py-1">5 langkah utama</span>
					<span class="rounded-full border border-border bg-muted/40 px-3 py-1">frontend-only</span>
					<span class="rounded-full border border-border bg-muted/40 px-3 py-1">tidak mengubah data</span>
				</div>
			</div>
			<div class="border-t border-border bg-muted/25 p-3 lg:border-l lg:border-t-0">
				<p class="text-[11px] font-black uppercase tracking-[0.18em] text-muted-foreground">Contoh ringkasan A0</p>
				<div class="mt-2 grid gap-2">
					{#each prototypeMetrics as metric (metric.label)}
						<div class="rounded-xl border border-border bg-card px-3 py-2">
							<p class="text-[11px] font-bold text-muted-foreground">{metric.label}</p>
							<p class="mt-0.5 text-sm font-black text-foreground">{metric.value}</p>
							<p class="mt-0.5 text-[11px] leading-4 text-muted-foreground">{metric.note}</p>
						</div>
					{/each}
				</div>
			</div>
		</div>
	</section>

	<section class="grid gap-3 xl:grid-cols-[13.5rem_1fr_18rem]">
		<aside class="h-fit rounded-2xl border border-border bg-muted/25 p-3 shadow-sm xl:sticky xl:top-3">
			<p class="text-[11px] font-black uppercase tracking-[0.18em] text-muted-foreground">Alur Utama</p>
			<nav class="mt-3 grid gap-1.5" aria-label="Navigasi prototype Asesmen">
				{#each prototypeLanes as lane (lane.id)}
					<a class="rounded-xl border border-border bg-card px-3 py-2 text-sm font-bold text-foreground hover:border-primary/40 hover:bg-primary/5" href={`#${lane.id}`}>
						{lane.code} · {lane.title}
					</a>
				{/each}
			</nav>
			<div id="a9-mode-lengkap" class="mt-3 rounded-xl border border-dashed border-border bg-card p-3">
				<p class="text-[11px] font-black uppercase tracking-[0.16em] text-muted-foreground">A9 · Mode Lengkap Panitia</p>
				<p class="mt-1 text-xs leading-5 text-muted-foreground">Fitur teknis tetap ada, tapi tidak menjadi pintu utama operator.</p>
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

					<div class="mt-3 grid gap-2 lg:grid-cols-[1fr_15rem]">
						<div class="rounded-xl border border-border bg-muted/20 p-2.5">
							<p class="text-[11px] font-black uppercase tracking-[0.16em] text-muted-foreground">Daftar kerja ringkas</p>
							<div class="mt-2 grid gap-1.5">
								{#each lane.checklist as item (item.label)}
									<div class="flex min-w-0 items-center justify-between gap-2 rounded-lg border border-border bg-card px-2.5 py-2">
										<span class="min-w-0 truncate text-xs font-semibold text-foreground">{item.label}</span>
										<span class={`shrink-0 rounded-full border px-2 py-0.5 text-[10px] font-black ${stepClass[item.state]}`}>{stepCopy[item.state]}</span>
									</div>
								{/each}
							</div>
						</div>

						<div class="rounded-xl border border-border bg-muted/20 p-2.5">
							<p class="text-[11px] font-black uppercase tracking-[0.16em] text-muted-foreground">Isi menu</p>
							<div class="mt-2 flex flex-wrap gap-1.5">
								{#each lane.items as item (item)}
									<span class="rounded-lg border border-border bg-card px-2.5 py-1 text-xs font-semibold text-foreground">{item}</span>
								{/each}
							</div>
							<p class="mt-2 text-xs leading-5 text-muted-foreground">{lane.operatorNote}</p>
						</div>
					</div>

					<div class="mt-3 flex flex-col gap-2 border-t border-border pt-3 sm:flex-row sm:items-center sm:justify-between">
						<div class="flex flex-wrap gap-2">
							<button class="rounded-xl bg-primary px-3 py-2 text-xs font-black text-primary-foreground shadow-sm" type="button">{lane.primaryAction}</button>
							{#if lane.secondaryAction}
								<button class="rounded-xl border border-border bg-background px-3 py-2 text-xs font-black text-foreground" type="button">{lane.secondaryAction}</button>
							{/if}
						</div>
						{#if lane.modeLengkap}
							<p class="text-xs leading-5 text-muted-foreground">Mode lengkap: {lane.modeLengkap.join(' · ')}</p>
						{/if}
					</div>
				</article>
			{/each}
		</div>

		<aside class="space-y-3">
			<section class="rounded-2xl border border-border bg-card p-3 shadow-sm">
				<p class="text-[11px] font-black uppercase tracking-[0.18em] text-muted-foreground">Preview Ruang</p>
				<p class="mt-1 text-xs leading-5 text-muted-foreground">Contoh tampilan kecil agar panitia melihat status ruang tanpa membuka konsol besar.</p>
				<div class="mt-2 grid gap-2">
					{#each prototypeRooms as room (room.code)}
						<div class={`rounded-xl border px-3 py-2 ${roomToneClass[room.tone]}`}>
							<div class="flex items-center justify-between gap-2">
								<p class="text-sm font-black">{room.code} · {room.name}</p>
								<span class="text-[11px] font-bold">{room.students}</span>
							</div>
							<p class="mt-1 text-xs font-semibold">{room.status}</p>
						</div>
					{/each}
				</div>
			</section>

			<section class="rounded-2xl border border-dashed border-border bg-muted/20 p-3 shadow-sm">
				<p class="text-[11px] font-black uppercase tracking-[0.18em] text-muted-foreground">Disembunyikan dari alur utama</p>
				<ul class="mt-2 space-y-1.5 text-xs leading-5 text-muted-foreground">
					{#each hiddenFromMainFlow as item (item)}
						<li class="rounded-lg border border-border bg-card px-2.5 py-1.5">{item}</li>
					{/each}
				</ul>
			</section>
		</aside>
	</section>
</div>
