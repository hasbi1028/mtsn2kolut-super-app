<script lang="ts">
	type Tab = 'ruang' | 'peringatan' | 'peserta';
	type HelpItem = { id: number; text: string };

	let activeTab = $state<Tab>('ruang');
	let audioOn = $state(true);
	let helpQueue = $state<HelpItem[]>([]);
	let checkedIds = $state<Set<string>>(new Set());
	let warningSent = $state<Set<string>>(new Set());

	const room = {
		name: 'Ruang R01',
		exam: 'Gladi CBT Madrasah',
		token: 'R01-A7K9',
		total: 30,
		online: 27,
		attention: 3,
		submitted: 4
	};

	const participants = [
		{ id: 'p1', name: 'Ahmad F.', seat: 1, status: 'mengerjakan', alert: 'Keluar halaman 2x', level: 'Merah' },
		{ id: 'p2', name: 'Nur Aisyah', seat: 2, status: 'online', alert: 'Koneksi tidak stabil', level: 'Kuning' },
		{ id: 'p3', name: 'Rizky P.', seat: 3, status: 'minta selesai', alert: 'Menunggu konfirmasi selesai', level: 'Kuning' },
		{ id: 'p4', name: 'Siti R.', seat: 4, status: 'selesai', alert: '', level: 'Aman' }
	];

	let attentionList = $derived(participants.filter((item) => item.alert && !checkedIds.has(item.id)));

	function addHelp(context: string) {
		const latest = attentionList[0];
		const text = `${context}: ${room.exam}, ${room.name}, peserta perlu dicek ${attentionList.length}, kasus terakhir ${latest ? `${latest.name} - ${latest.alert}` : 'tidak ada'}.`;
		helpQueue = [{ id: Date.now(), text }, ...helpQueue].slice(0, 4);
	}

	function markChecked(id: string) {
		checkedIds = new Set([...checkedIds, id]);
	}

	function sendWarning(id: string) {
		warningSent = new Set([...warningSent, id]);
	}
</script>

<svelte:head>
	<title>Portal Pengawasan Ruang</title>
	<meta name="description" content="Portal pengawas CBT sederhana untuk ruang, peringatan, dan peserta." />
</svelte:head>

<div class="min-h-screen bg-slate-50 text-slate-950" style="color-scheme: light;">
	<div class="mx-auto flex min-h-screen w-full max-w-md flex-col px-4 py-4">
		<header class="rounded-3xl border border-emerald-200 bg-white p-4 shadow-sm">
			<p class="text-xs font-black uppercase tracking-[0.2em] text-emerald-700">Portal Pengawasan Ruang</p>
			<div class="mt-1 flex items-start justify-between gap-3">
				<div>
					<h1 class="text-2xl font-black">{room.name}</h1>
					<p class="text-sm font-semibold text-slate-600">{room.exam}</p>
				</div>
				<span class="rounded-2xl bg-emerald-50 px-3 py-2 text-sm font-black text-emerald-900">{room.token}</span>
			</div>
		</header>

		<main class="mt-4 flex-1 pb-24">
			{#if activeTab === 'ruang'}
				<section class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
					<h2 class="text-lg font-black">Status ruang</h2>
					<div class="mt-4 grid grid-cols-2 gap-2 text-center text-xs font-black">
						<div class="rounded-2xl bg-slate-100 p-3 text-slate-700">Total<br /><span class="text-2xl">{room.total}</span></div>
						<div class="rounded-2xl bg-emerald-50 p-3 text-emerald-900">Online<br /><span class="text-2xl">{room.online}</span></div>
						<div class="rounded-2xl bg-amber-50 p-3 text-amber-900">Perlu dicek<br /><span class="text-2xl">{attentionList.length}</span></div>
						<div class="rounded-2xl bg-blue-50 p-3 text-blue-900">Selesai<br /><span class="text-2xl">{room.submitted}</span></div>
					</div>
					<div class="mt-4 grid gap-2">
						<button class="rounded-2xl bg-emerald-700 px-4 py-3 text-sm font-black text-white">Mulai Ujian</button>
						<button class="rounded-2xl border border-slate-300 px-4 py-3 text-sm font-black text-slate-700" onclick={() => addHelp('Mohon bantuan admin ruang')}>Hubungi Admin</button>
						<button class="rounded-2xl border border-slate-300 px-4 py-3 text-sm font-black text-slate-700" onclick={() => (audioOn = !audioOn)}>Audio {audioOn ? 'ON' : 'OFF'}</button>
						<button class="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm font-black text-red-800">Tutup Ujian</button>
					</div>
				</section>
			{:else if activeTab === 'peringatan'}
				<section class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
					<div class="flex items-center justify-between gap-3">
						<h2 class="text-lg font-black">Peringatan</h2>
						<button class="rounded-2xl border border-slate-300 px-3 py-2 text-xs font-black text-slate-700" onclick={() => addHelp('Peringatan peserta butuh bantuan')}>Hubungi Admin</button>
					</div>
					<div class="mt-3 flex gap-2 overflow-x-auto text-xs font-black">
						<span class="rounded-full bg-slate-100 px-3 py-1 text-slate-700">Semua</span>
						<span class="rounded-full bg-red-50 px-3 py-1 text-red-800">Merah</span>
						<span class="rounded-full bg-amber-50 px-3 py-1 text-amber-900">Kuning</span>
						<span class="rounded-full bg-emerald-50 px-3 py-1 text-emerald-900">Belum Dicek</span>
					</div>
					<div class="mt-4 divide-y divide-slate-200 overflow-hidden rounded-2xl border border-slate-200">
						{#each attentionList as item (item.id)}
							<div class="p-3">
								<div class="flex items-start justify-between gap-3">
									<div>
										<p class="font-black text-slate-950">{item.name} · Kursi {item.seat}</p>
										<p class="text-sm font-semibold text-slate-600">{item.alert}</p>
									</div>
									<span class={`rounded-full px-2 py-1 text-[11px] font-black ${item.level === 'Merah' ? 'bg-red-50 text-red-800' : 'bg-amber-50 text-amber-900'}`}>{item.level}</span>
								</div>
								<div class="mt-3 flex flex-wrap gap-2">
									<button class="rounded-xl bg-emerald-700 px-3 py-2 text-xs font-black text-white" onclick={() => markChecked(item.id)}>Sudah Dicek</button>
									<button class="rounded-xl border border-amber-300 bg-amber-50 px-3 py-2 text-xs font-black text-amber-900" onclick={() => sendWarning(item.id)}>{warningSent.has(item.id) ? 'Peringatan Terkirim' : 'Beri Peringatan'}</button>
								</div>
							</div>
						{/each}
						{#if attentionList.length === 0}
							<p class="p-4 text-sm font-semibold text-slate-500">Tidak ada peringatan aktif. Tetap pantau peserta.</p>
						{/if}
					</div>
				</section>
			{:else}
				<section class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
					<h2 class="text-lg font-black">Peserta</h2>
					<div class="mt-4 divide-y divide-slate-200 overflow-hidden rounded-2xl border border-slate-200">
						{#each participants as item (item.id)}
							<div class="grid grid-cols-[3rem_1fr_auto] items-center gap-3 p-3 text-sm">
								<span class="rounded-xl bg-slate-100 px-2 py-1 text-center text-xs font-black text-slate-700">{item.seat}</span>
								<div class="min-w-0">
									<p class="truncate font-black text-slate-950">{item.name}</p>
									<p class="truncate text-xs font-bold text-slate-500">{item.alert || 'Aman'} </p>
								</div>
								<span class="rounded-full bg-slate-100 px-2 py-1 text-[11px] font-black text-slate-700">{item.status}</span>
							</div>
						{/each}
					</div>
				</section>
			{/if}

			{#if helpQueue.length > 0}
				<section class="mt-4 rounded-3xl border border-blue-200 bg-blue-50 p-4 text-blue-950 shadow-sm">
					<h2 class="text-sm font-black">Antrian bantuan admin</h2>
					<div class="mt-2 space-y-2">
						{#each helpQueue as item (item.id)}
							<p class="rounded-2xl bg-white/80 p-3 text-xs font-bold leading-5">{item.text}</p>
						{/each}
					</div>
				</section>
			{/if}
		</main>

		<nav class="fixed right-4 bottom-4 left-4 mx-auto grid max-w-md grid-cols-3 gap-2 rounded-3xl border border-slate-200 bg-white p-2 shadow-xl">
			<button class={`rounded-2xl px-3 py-3 text-xs font-black ${activeTab === 'ruang' ? 'bg-emerald-700 text-white' : 'text-slate-700'}`} onclick={() => (activeTab = 'ruang')}>Ruang</button>
			<button class={`rounded-2xl px-3 py-3 text-xs font-black ${activeTab === 'peringatan' ? 'bg-emerald-700 text-white' : 'text-slate-700'}`} onclick={() => (activeTab = 'peringatan')}>Peringatan</button>
			<button class={`rounded-2xl px-3 py-3 text-xs font-black ${activeTab === 'peserta' ? 'bg-emerald-700 text-white' : 'text-slate-700'}`} onclick={() => (activeTab = 'peserta')}>Peserta</button>
		</nav>
	</div>
</div>
