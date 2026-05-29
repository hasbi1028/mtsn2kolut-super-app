<script lang="ts">
	import {
		commandMetrics,
		excludedSurfaces,
		legacyMenuGroups,
		portalPreviews,
		sourceNotes,
		tableRows,
		workAreas,
		type StatusTone
	} from './asesmen-prototype.model';

	const toneClass: Record<StatusTone, string> = {
		blue: 'border-blue-200 bg-blue-50 text-blue-900',
		green: 'border-emerald-200 bg-emerald-50 text-emerald-900',
		amber: 'border-amber-200 bg-amber-50 text-amber-950',
		red: 'border-rose-200 bg-rose-50 text-rose-900',
		slate: 'border-slate-200 bg-slate-50 text-slate-700'
	};
</script>

<svelte:head>
	<title>Prototype CBT Super App — MTsN 2 Kolut</title>
</svelte:head>

<div class="space-y-4">
	<section class="overflow-hidden rounded-[1.75rem] border border-slate-200 bg-white shadow-sm">
		<div class="grid gap-0 xl:grid-cols-[16rem_1fr]">
			<aside class="border-b border-slate-200 bg-slate-950 p-4 text-white xl:border-b-0 xl:border-r">
				<p class="text-[11px] font-black uppercase tracking-[0.24em] text-blue-200">MTSN 2 KOLAKA UTARA</p>
				<h1 class="mt-2 text-2xl font-black tracking-tight">Prototype CBT</h1>
				<p class="mt-1 text-xs font-semibold uppercase tracking-[0.18em] text-slate-300">Pusat Data</p>

				<nav class="mt-4 grid gap-3" aria-label="Struktur menu CBT prototype">
					{#each legacyMenuGroups as group (group.label)}
						<div>
							<p class="text-[10px] font-black uppercase tracking-[0.18em] text-slate-400">{group.label}</p>
							<div class="mt-1.5 grid gap-1">
								{#each group.items as item (item)}
									<span class="rounded-lg px-2 py-1.5 text-xs font-bold text-slate-100 hover:bg-white/10">{item}</span>
								{/each}
							</div>
						</div>
					{/each}
				</nav>
			</aside>

			<div class="min-w-0 bg-slate-50 p-3 sm:p-4">
				<div class="rounded-[1.5rem] border border-slate-200 bg-white p-4 shadow-sm">
					<div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
						<div class="min-w-0">
							<p class="text-[11px] font-black uppercase tracking-[0.22em] text-blue-700">PROTOTYPE UI — belum terhubung backend</p>
							<h2 class="mt-1 text-2xl font-black tracking-tight text-slate-950 sm:text-4xl">Dashboard Utama · Command Center CBT</h2>
							<p class="mt-2 max-w-3xl text-sm leading-6 text-slate-600">
								Dibuat ulang dari rujukan <strong>cbt.mtsn2kolut.sch.id</strong> dan repo lokal <strong>cbt-ujian</strong>. Tombol di halaman ini hanya rancangan alur. Belum menjalankan aksi data.
							</p>
						</div>
						<div class="flex flex-wrap gap-1.5 text-[11px] font-black">
							<span class="rounded-full border border-blue-200 bg-blue-50 px-3 py-1 text-blue-800">frontend-only</span>
							<span class="rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1 text-emerald-800">tanpa API</span>
							<span class="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-slate-700">Bank Soal dikecualikan</span>
						</div>
					</div>

					<div class="mt-4 grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
						{#each commandMetrics as metric (metric.label)}
							<div class={`rounded-2xl border p-3 ${toneClass[metric.tone]}`}>
								<p class="text-[10px] font-black uppercase tracking-[0.16em] opacity-75">{metric.label}</p>
								<p class="mt-1 text-3xl font-black tracking-tight">{metric.value}</p>
								<p class="mt-1 text-xs font-semibold leading-5 opacity-80">{metric.note}</p>
							</div>
						{/each}
					</div>

					<div class="mt-4 rounded-2xl border border-dashed border-slate-300 bg-slate-50 p-3">
						<p class="text-[11px] font-black uppercase tracking-[0.18em] text-slate-500">Sumber prototype baru</p>
						<div class="mt-2 grid gap-1.5 text-xs font-semibold text-slate-700">
							{#each sourceNotes as note (note)}
								<p class="rounded-lg border border-slate-200 bg-white px-2.5 py-1.5">{note}</p>
							{/each}
						</div>
					</div>
				</div>
			</div>
		</div>
	</section>

	<section class="grid gap-3 xl:grid-cols-[1fr_20rem]">
		<div class="space-y-3">
			<section class="rounded-[1.5rem] border border-slate-200 bg-white p-3 shadow-sm">
				<div class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
					<div>
						<p class="text-[11px] font-black uppercase tracking-[0.18em] text-blue-700">Kesiapan Pelaksanaan</p>
						<h2 class="text-xl font-black text-slate-950">Alur dibuat mengikuti pola CBT lama</h2>
						<p class="mt-1 text-sm leading-6 text-slate-600">Bukan Micro Workflow List lagi. Prototype ini memakai bahasa dan struktur CBT lama: dashboard, ruang ujian, jadwal sesi, proctoring live, dan rekap nilai.</p>
					</div>
					<span class="rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1 text-[11px] font-black text-emerald-800">dibuat ulang</span>
				</div>

				<div class="mt-3 grid gap-2">
					{#each workAreas as area (area.code)}
						<article class="rounded-2xl border border-slate-200 bg-slate-50 p-3">
							<div class="flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
								<div class="min-w-0">
									<p class="text-[11px] font-black uppercase tracking-[0.16em] text-blue-700">{area.code} · {area.status}</p>
									<h3 class="mt-0.5 text-base font-black text-slate-950">{area.label}</h3>
									<p class="mt-1 text-sm leading-6 text-slate-600">{area.description}</p>
								</div>
								<button class="w-fit rounded-xl bg-blue-700 px-3 py-2 text-xs font-black text-white shadow-sm" type="button">{area.cta}</button>
							</div>
							<div class="mt-2 flex flex-wrap gap-1.5">
								{#each area.items as item (item)}
									<span class="rounded-lg border border-slate-200 bg-white px-2.5 py-1 text-xs font-bold text-slate-700">{item}</span>
								{/each}
							</div>
						</article>
					{/each}
				</div>
			</section>

			<section class="rounded-[1.5rem] border border-slate-200 bg-white p-3 shadow-sm">
				<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
					<div>
						<p class="text-[11px] font-black uppercase tracking-[0.18em] text-slate-500">Tabel operasional</p>
						<h2 class="text-xl font-black text-slate-950">Contoh daftar kerja seperti CBT lama</h2>
					</div>
					<div class="flex flex-wrap gap-1.5">
						<button class="rounded-lg border border-slate-300 bg-white px-2.5 py-1.5 text-[11px] font-black text-slate-700" type="button">FILTER</button>
						<button class="rounded-lg bg-slate-950 px-2.5 py-1.5 text-[11px] font-black text-white" type="button">EXPORT</button>
						<button class="rounded-lg bg-blue-700 px-2.5 py-1.5 text-[11px] font-black text-white" type="button">TAMBAH</button>
					</div>
				</div>

				<div class="mt-3 overflow-x-auto">
					<table class="min-w-[720px] w-full border-collapse text-left text-sm">
						<thead class="text-[11px] uppercase tracking-[0.14em] text-slate-500">
							<tr class="border-b border-slate-200">
								<th class="py-2 pr-3">Nama</th>
								<th class="py-2 pr-3">Keterangan</th>
								<th class="py-2 pr-3">Status</th>
								<th class="py-2 pr-3 text-right">Aksi</th>
							</tr>
						</thead>
						<tbody>
							{#each tableRows as row (row.name)}
								<tr class="border-b border-slate-100 last:border-b-0">
									<td class="py-3 pr-3 font-black text-slate-950">{row.name}</td>
									<td class="py-3 pr-3 text-slate-600">{row.meta}</td>
									<td class="py-3 pr-3"><span class={`rounded-full border px-2 py-1 text-[11px] font-black ${toneClass[row.tone]}`}>{row.status}</span></td>
									<td class="py-3 text-right">
										<div class="flex justify-end gap-1.5">
											{#each row.actions as action (action)}
												<button class="rounded-lg border border-slate-300 bg-white px-2 py-1 text-[11px] font-black text-slate-700" type="button">{action}</button>
											{/each}
										</div>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
				<p class="mt-3 text-xs font-bold text-slate-500">HAL 1 / 1 (TOTAL 4) · contoh statis untuk review UI</p>
			</section>
		</div>

		<aside class="space-y-3">
			<section class="rounded-[1.5rem] border border-slate-200 bg-white p-3 shadow-sm">
				<p class="text-[11px] font-black uppercase tracking-[0.18em] text-blue-700">Portal QR dari repo cbt-ujian</p>
				<div class="mt-2 grid gap-2">
					{#each portalPreviews as portal (portal.label)}
						<article class="rounded-2xl border border-slate-200 bg-slate-50 p-3">
							<p class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-500">{portal.label} · {portal.route}</p>
							<h3 class="mt-1 text-sm font-black text-slate-950">{portal.title}</h3>
							<p class="mt-1 text-xs leading-5 text-slate-600">{portal.description}</p>
							<ol class="mt-2 space-y-1 text-xs font-semibold text-slate-700">
								{#each portal.steps as step, index (step)}
									<li>{index + 1}. {step}</li>
								{/each}
							</ol>
							<div class="mt-2 flex flex-wrap gap-1.5">
								{#each portal.primaryActions as action (action)}
									<span class="rounded-lg bg-slate-950 px-2 py-1 text-[10px] font-black text-white">{action}</span>
								{/each}
							</div>
						</article>
					{/each}
				</div>
			</section>

			<section class="rounded-[1.5rem] border border-dashed border-slate-300 bg-slate-50 p-3 shadow-sm">
				<p class="text-[11px] font-black uppercase tracking-[0.18em] text-slate-500">Tidak dibawa ke prototype utama</p>
				<ul class="mt-2 space-y-1.5 text-xs leading-5 text-slate-600">
					{#each excludedSurfaces as item (item)}
						<li class="rounded-lg border border-slate-200 bg-white px-2.5 py-1.5">{item}</li>
					{/each}
				</ul>
			</section>
		</aside>
	</section>
</div>
