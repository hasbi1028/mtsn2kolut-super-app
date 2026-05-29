<script lang="ts">
	import {
		cardPrintSteps,
		commandMetrics,
		excludedSurfaces,
		legacyMenuGroups,
		mobileScreens,
		portalPreviews,
		productionLanes,
		approvalChecklist,
		backendWaves,
		readinessCards,
		resultSummaries,
		roomBlueprints,
		sourceNotes,
		tableRows,
		todayAgenda,
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
	<title>Blueprint Prototype CBT Super App — MTsN 2 Kolut</title>
</svelte:head>

<div class="space-y-4">
	<section class="overflow-hidden rounded-[1.75rem] border border-slate-200 bg-white shadow-sm">
		<div class="grid gap-0 xl:grid-cols-[16rem_1fr]">
			<aside class="border-b border-slate-200 bg-slate-950 p-4 text-white xl:border-b-0 xl:border-r">
				<p class="text-[11px] font-black uppercase tracking-[0.24em] text-blue-200">MTSN 2 KOLAKA UTARA</p>
				<h1 class="mt-2 text-2xl font-black tracking-tight">Blueprint CBT</h1>
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
							<h2 class="mt-1 text-2xl font-black tracking-tight text-slate-950 sm:text-4xl">Blueprint Final CBT Web-First</h2>
							<p class="mt-2 max-w-3xl text-sm leading-6 text-slate-600">
								Lanjutan tahap ini membuat prototype lebih konkret sebagai blueprint final: dashboard CBT, ruang, kartu QR, portal siswa, portal pengawas, proctoring, dan rekap nilai. Tombol di halaman ini hanya rancangan alur. Belum menjalankan aksi data.
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
						<p class="text-[11px] font-black uppercase tracking-[0.18em] text-slate-500">Sumber blueprint baru</p>
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
						<p class="text-[11px] font-black uppercase tracking-[0.18em] text-blue-700">01 · Dashboard CBT</p>
						<h2 class="text-xl font-black text-slate-950">Kesiapan, agenda hari ini, dan kekurangan operasional</h2>
						<p class="mt-1 text-sm leading-6 text-slate-600">Halaman pertama harus langsung menjawab: apa yang siap, apa yang kurang, sesi mana hari ini, dan ruang mana perlu dicek.</p>
					</div>
					<span class="rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1 text-[11px] font-black text-emerald-800">blueprint final</span>
				</div>
				<div class="mt-3 grid gap-2 md:grid-cols-2 xl:grid-cols-4">
					{#each readinessCards as card (card.label)}
						<div class={`rounded-2xl border p-3 ${toneClass[card.tone]}`}>
							<div class="flex items-start justify-between gap-2">
								<p class="text-xs font-black uppercase tracking-[0.12em] opacity-75">{card.label}</p>
								<span class="rounded-full bg-white/70 px-2 py-0.5 text-[10px] font-black">{card.percent}%</span>
							</div>
							<div class="mt-2 h-2 rounded-full bg-white/70"><div class="h-2 rounded-full bg-current" style={`width: ${card.percent}%`}></div></div>
							<p class="mt-2 text-sm font-black">{card.status}</p>
							<p class="mt-1 text-xs font-semibold leading-5 opacity-80">{card.detail}</p>
						</div>
					{/each}
				</div>
				<div class="mt-3 grid gap-2 lg:grid-cols-2">
					{#each todayAgenda as agenda (agenda.time)}
						<div class="rounded-2xl border border-slate-200 bg-slate-50 p-3">
							<div class="flex items-start justify-between gap-3">
								<div><p class="text-xs font-black text-blue-700">{agenda.time}</p><h3 class="text-sm font-black text-slate-950">{agenda.title}</h3></div>
								<span class={`rounded-full border px-2 py-1 text-[10px] font-black ${toneClass[agenda.tone]}`}>{agenda.status}</span>
							</div>
							<p class="mt-1 text-xs leading-5 text-slate-600">{agenda.detail}</p>
						</div>
					{/each}
				</div>
			</section>

			<section class="rounded-[1.5rem] border border-slate-200 bg-white p-3 shadow-sm">
				<p class="text-[11px] font-black uppercase tracking-[0.18em] text-blue-700">02 · Ruang Ujian & Kartu QR</p>
				<h2 class="text-xl font-black text-slate-950">Ruang, kartu peserta, dan lembar pengawas dibuat satu alur</h2>
				<div class="mt-3 grid gap-3 lg:grid-cols-[1.1fr_0.9fr]">
					<div class="overflow-x-auto rounded-2xl border border-slate-200">
						<table class="w-full min-w-[620px] border-collapse text-left text-sm">
							<thead class="bg-slate-50 text-[11px] uppercase tracking-[0.14em] text-slate-500"><tr><th class="p-3">Ruang</th><th class="p-3">Pengawas</th><th class="p-3">Peserta</th><th class="p-3">Status</th></tr></thead>
							<tbody>
								{#each roomBlueprints as room (room.room)}
									<tr class="border-t border-slate-100"><td class="p-3 font-black text-slate-950">{room.room}<span class="block text-xs font-semibold text-slate-500">{room.capacity}</span></td><td class="p-3 text-slate-600">{room.proctor}</td><td class="p-3 text-slate-600">{room.students}</td><td class="p-3"><span class={`rounded-full border px-2 py-1 text-[11px] font-black ${toneClass[room.tone]}`}>{room.status}</span></td></tr>
								{/each}
							</tbody>
						</table>
					</div>
					<div class="grid gap-2">
						{#each cardPrintSteps as step (step.label)}
							<div class="rounded-2xl border border-slate-200 bg-slate-50 p-3">
								<p class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-500">{step.badge}</p>
								<h3 class="mt-1 text-sm font-black text-slate-950">{step.label}</h3>
								<p class="mt-1 text-xs leading-5 text-slate-600">{step.detail}</p>
							</div>
						{/each}
					</div>
				</div>
			</section>

			<section class="rounded-[1.5rem] border border-slate-200 bg-white p-3 shadow-sm">
				<p class="text-[11px] font-black uppercase tracking-[0.18em] text-blue-700">03 · Portal mobile-web</p>
				<h2 class="text-xl font-black text-slate-950">Siswa dan pengawas terasa seperti aplikasi HP, bukan dashboard admin</h2>
				<div class="mt-3 grid gap-3 lg:grid-cols-2">
					{#each mobileScreens as screen (screen.label)}
						<article class="rounded-[1.4rem] border border-slate-200 bg-slate-50 p-3">
							<div class="mx-auto max-w-[18rem] rounded-[2rem] border-4 border-slate-900 bg-white p-2 shadow-sm">
								<div class="rounded-[1.5rem] bg-slate-50 p-3">
									<p class="text-[10px] font-black uppercase tracking-[0.16em] text-blue-700">{screen.label}</p>
									<h3 class="mt-1 text-sm font-black text-slate-950">{screen.title}</h3>
									<div class="mt-3 grid gap-1.5">
										{#each screen.sections as section (section)}
											<div class="rounded-xl border border-slate-200 bg-white px-2.5 py-2 text-xs font-bold text-slate-700">{section}</div>
										{/each}
									</div>
									<div class="mt-3 grid grid-cols-3 gap-1">
										{#each screen.bottomActions as action (action)}
											<span class="rounded-lg bg-slate-950 px-1.5 py-2 text-center text-[10px] font-black text-white">{action}</span>
										{/each}
									</div>
								</div>
							</div>
						</article>
					{/each}
				</div>
			</section>

			<section class="rounded-[1.5rem] border border-slate-200 bg-white p-3 shadow-sm">
				<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
					<div><p class="text-[11px] font-black uppercase tracking-[0.18em] text-slate-500">04 · Proctoring Live & Rekap Nilai</p><h2 class="text-xl font-black text-slate-950">Tabel operasional seperti CBT lama, tetapi aksi sensitif tetap admin</h2></div>
					<div class="flex flex-wrap gap-1.5"><button class="rounded-lg border border-slate-300 bg-white px-2.5 py-1.5 text-[11px] font-black text-slate-700" type="button">FILTER</button><button class="rounded-lg bg-slate-950 px-2.5 py-1.5 text-[11px] font-black text-white" type="button">EXPORT</button><button class="rounded-lg bg-blue-700 px-2.5 py-1.5 text-[11px] font-black text-white" type="button">TAMBAH</button></div>
				</div>
				<div class="mt-3 overflow-x-auto">
					<table class="w-full min-w-[720px] border-collapse text-left text-sm"><thead class="text-[11px] uppercase tracking-[0.14em] text-slate-500"><tr class="border-b border-slate-200"><th class="py-2 pr-3">Nama</th><th class="py-2 pr-3">Keterangan</th><th class="py-2 pr-3">Status</th><th class="py-2 pr-3 text-right">Aksi</th></tr></thead><tbody>{#each tableRows as row (row.name)}<tr class="border-b border-slate-100 last:border-b-0"><td class="py-3 pr-3 font-black text-slate-950">{row.name}</td><td class="py-3 pr-3 text-slate-600">{row.meta}</td><td class="py-3 pr-3"><span class={`rounded-full border px-2 py-1 text-[11px] font-black ${toneClass[row.tone]}`}>{row.status}</span></td><td class="py-3 text-right"><div class="flex justify-end gap-1.5">{#each row.actions as action (action)}<button class="rounded-lg border border-slate-300 bg-white px-2 py-1 text-[11px] font-black text-slate-700" type="button">{action}</button>{/each}</div></td></tr>{/each}</tbody></table>
				</div>
				<div class="mt-3 grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
					{#each resultSummaries as item (item.label)}
						<div class={`rounded-2xl border p-3 ${toneClass[item.tone]}`}><p class="text-[10px] font-black uppercase tracking-[0.16em] opacity-75">{item.label}</p><p class="mt-1 text-2xl font-black">{item.value}</p><p class="mt-1 text-xs font-semibold leading-5 opacity-80">{item.detail}</p></div>
					{/each}
				</div>
				<p class="mt-3 text-xs font-bold text-slate-500">HAL 1 / 1 (TOTAL 4) · contoh statis untuk review UI</p>
			</section>

			<section class="rounded-[1.5rem] border border-slate-200 bg-white p-3 shadow-sm">
				<div class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
					<div>
						<p class="text-[11px] font-black uppercase tracking-[0.18em] text-blue-700">05 · Peta Promosi Produksi</p>
						<h2 class="text-xl font-black text-slate-950">Dari prototype ke produksi dibuat bertahap, bukan sekali lompat</h2>
						<p class="mt-1 text-sm leading-6 text-slate-600">Bagian ini memperjelas keputusan setelah Bapak setujui tampilan: mana yang masuk admin Super App, mana yang menjadi portal publik, dan mana yang tetap Mode Rinci Admin.</p>
					</div>
					<span class="w-fit rounded-full border border-blue-200 bg-blue-50 px-3 py-1 text-[11px] font-black text-blue-800">rencana rollout</span>
				</div>
				<div class="mt-3 grid gap-2">
					{#each productionLanes as lane (lane.lane)}
						<article class="rounded-2xl border border-slate-200 bg-slate-50 p-3">
							<div class="flex flex-col gap-2 lg:flex-row lg:items-start lg:justify-between">
								<div class="min-w-0">
									<p class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-500">{lane.target}</p>
									<h3 class="text-sm font-black text-slate-950">{lane.lane}</h3>
									<p class="mt-1 text-xs leading-5 text-slate-600">{lane.action}</p>
									<p class="mt-1 text-xs font-semibold leading-5 text-slate-500">{lane.note}</p>
								</div>
								<span class={`w-fit rounded-full border px-2 py-1 text-[10px] font-black ${toneClass[lane.tone]}`}>{lane.status}</span>
							</div>
						</article>
					{/each}
				</div>
			</section>

			<section class="rounded-[1.5rem] border border-slate-200 bg-white p-3 shadow-sm">
				<p class="text-[11px] font-black uppercase tracking-[0.18em] text-blue-700">06 · Checklist Persetujuan</p>
				<h2 class="text-xl font-black text-slate-950">Yang perlu Bapak nilai sebelum saya promosi ke halaman produksi</h2>
				<div class="mt-3 grid gap-2 lg:grid-cols-2">
					{#each approvalChecklist as item (item.label)}
						<div class="rounded-2xl border border-slate-200 bg-slate-50 p-3">
							<div class="flex items-start gap-2">
								<span class="mt-0.5 grid h-5 w-5 shrink-0 place-items-center rounded-full bg-emerald-100 text-xs font-black text-emerald-800">✓</span>
								<div>
									<h3 class="text-sm font-black text-slate-950">{item.label}</h3>
									<p class="mt-1 text-xs leading-5 text-slate-600">{item.detail}</p>
									<p class="mt-1 text-[10px] font-black uppercase tracking-[0.14em] text-slate-400">Reviewer: {item.owner}</p>
								</div>
							</div>
						</div>
					{/each}
				</div>
			</section>

			<section class="rounded-[1.5rem] border border-slate-200 bg-white p-3 shadow-sm">
				<p class="text-[11px] font-black uppercase tracking-[0.18em] text-blue-700">07 · Gelombang Backend Nanti</p>
				<h2 class="text-xl font-black text-slate-950">Integrasi data dilakukan setelah UI disetujui</h2>
				<div class="mt-3 overflow-x-auto rounded-2xl border border-slate-200">
					<table class="w-full min-w-[720px] border-collapse text-left text-sm">
						<thead class="bg-slate-50 text-[11px] uppercase tracking-[0.14em] text-slate-500"><tr><th class="p-3">Wave</th><th class="p-3">Cakupan</th><th class="p-3">Safety gate</th></tr></thead>
						<tbody>
							{#each backendWaves as wave (wave.phase)}
								<tr class="border-t border-slate-100"><td class="p-3 font-black text-slate-950">{wave.phase}</td><td class="p-3 text-slate-600">{wave.scope}</td><td class="p-3 text-slate-600">{wave.safeGate}</td></tr>
							{/each}
						</tbody>
					</table>
				</div>
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
							<ol class="mt-2 space-y-1 text-xs font-semibold text-slate-700">{#each portal.steps as step, index (step)}<li>{index + 1}. {step}</li>{/each}</ol>
							<div class="mt-2 flex flex-wrap gap-1.5">{#each portal.primaryActions as action (action)}<span class="rounded-lg bg-slate-950 px-2 py-1 text-[10px] font-black text-white">{action}</span>{/each}</div>
						</article>
					{/each}
				</div>
			</section>

			<section class="rounded-[1.5rem] border border-dashed border-slate-300 bg-slate-50 p-3 shadow-sm">
				<p class="text-[11px] font-black uppercase tracking-[0.18em] text-slate-500">Tidak dibawa ke prototype utama</p>
				<ul class="mt-2 space-y-1.5 text-xs leading-5 text-slate-600">{#each excludedSurfaces as item (item)}<li class="rounded-lg border border-slate-200 bg-white px-2.5 py-1.5">{item}</li>{/each}</ul>
			</section>
		</aside>
	</section>
</div>
