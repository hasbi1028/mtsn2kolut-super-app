<script lang="ts">
	import { onMount } from 'svelte';
	import { readClientApiData } from '$lib/client/api';

	type PackageOption = {
		id: string;
		event_id?: string;
		subject_id: string;
		subject_code?: string;
		subject_name: string;
		title: string;
		description?: string;
		duration_minutes: number;
		question_count: number;
		session_count: number;
		locked?: boolean;
		snapshot_version?: number;
	};

	let packages = $state<PackageOption[]>([]);
	let loading = $state(true);
	let error = $state('');
	let search = $state('');
	let subjectFilter = $state('all');
	let readinessFilter = $state<'all' | 'empty' | 'ready' | 'locked' | 'used'>('all');

	const subjects = $derived(
		Array.from(new Map(packages.map((item) => [item.subject_id, item])).values())
			.sort((a, b) => a.subject_name.localeCompare(b.subject_name))
	);
	const filteredPackages = $derived(packages.filter((item) => {
		const query = search.trim().toLowerCase();
		const haystack = `${item.title} ${item.subject_name} ${item.subject_code ?? ''} ${item.description ?? ''}`.toLowerCase();
		const matchesSearch = !query || haystack.includes(query);
		const matchesSubject = subjectFilter === 'all' || item.subject_id === subjectFilter;
		const matchesReadiness = readinessFilter === 'all'
			|| (readinessFilter === 'empty' && Number(item.question_count ?? 0) === 0)
			|| (readinessFilter === 'ready' && Number(item.question_count ?? 0) > 0 && !item.locked)
			|| (readinessFilter === 'locked' && Boolean(item.locked))
			|| (readinessFilter === 'used' && Number(item.session_count ?? 0) > 0);
		return matchesSearch && matchesSubject && matchesReadiness;
	}));
	const totalQuestions = $derived(packages.reduce((total, item) => total + Number(item.question_count ?? 0), 0));
	const emptyCount = $derived(packages.filter((item) => Number(item.question_count ?? 0) === 0).length);
	const lockedCount = $derived(packages.filter((item) => item.locked).length);
	const usedCount = $derived(packages.filter((item) => Number(item.session_count ?? 0) > 0).length);

	onMount(() => {
		void loadPackages();
	});

	async function loadPackages() {
		loading = true;
		error = '';
		try {
			const response = await fetch('/api/asesmen/package-options');
			packages = await readClientApiData<PackageOption[]>(response);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Daftar Paket Soal belum dapat dibuka.';
			packages = [];
		} finally {
			loading = false;
		}
	}

	function statusLabel(item: PackageOption) {
		if (item.locked) return 'Terkunci';
		if (Number(item.question_count ?? 0) === 0) return 'Kosong';
		if (Number(item.session_count ?? 0) > 0) return 'Dipakai';
		return 'Siap dirakit';
	}

	function statusClass(item: PackageOption) {
		if (item.locked) return 'border-slate-300 bg-slate-100 text-slate-700';
		if (Number(item.question_count ?? 0) === 0) return 'border-amber-300 bg-amber-50 text-amber-700';
		if (Number(item.session_count ?? 0) > 0) return 'border-sky-300 bg-sky-50 text-sky-700';
		return 'border-emerald-300 bg-emerald-50 text-emerald-700';
	}
</script>

<svelte:head>
	<title>Paket Soal | MTsN 2 Kolut</title>
</svelte:head>

<div class="space-y-5 pb-16">
	<section class="rounded-2xl border border-border bg-card p-5 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
			<div class="space-y-2">
				<p class="text-xs font-semibold tracking-[0.22em] text-muted-foreground uppercase">Modul Mandiri</p>
				<h1 class="text-2xl font-bold tracking-tight text-foreground md:text-3xl">Paket Soal</h1>
				<p class="max-w-3xl text-sm leading-6 text-muted-foreground">
					Dapur perakitan paket dari Bank Soal sebelum dipakai di Asesmen/CBT. Paket dapat dipakai ulang untuk beberapa rombel, kegiatan, simulasi, atau ujian susulan.
				</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<a class="inline-flex h-10 items-center justify-center rounded-md border px-4 text-sm font-semibold text-foreground hover:bg-muted" href="/bank-soal">Buka Bank Soal</a>
				<a class="inline-flex h-10 items-center justify-center rounded-md border px-4 text-sm font-semibold text-foreground hover:bg-muted" href="/asesmen">Pakai di Asesmen</a>
				<button type="button" class="inline-flex h-10 items-center justify-center rounded-md bg-primary px-4 text-sm font-semibold text-primary-foreground shadow-sm transition hover:bg-primary/90" onclick={() => void loadPackages()} disabled={loading}>{loading ? 'Memuat…' : 'Refresh'}</button>
			</div>
		</div>
	</section>

	{#if error}
		<p class="rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive" role="alert">{error}</p>
	{/if}

	<section class="grid gap-3 md:grid-cols-4">
		<div class="rounded-xl border bg-background p-4"><p class="text-xs font-medium text-muted-foreground">Total paket</p><p class="mt-1 text-2xl font-bold">{packages.length}</p></div>
		<div class="rounded-xl border bg-background p-4"><p class="text-xs font-medium text-muted-foreground">Total soal tertaut</p><p class="mt-1 text-2xl font-bold">{totalQuestions}</p></div>
		<div class="rounded-xl border bg-background p-4"><p class="text-xs font-medium text-muted-foreground">Kosong</p><p class="mt-1 text-2xl font-bold">{emptyCount}</p></div>
		<div class="rounded-xl border bg-background p-4"><p class="text-xs font-medium text-muted-foreground">Terkunci/dipakai</p><p class="mt-1 text-2xl font-bold">{lockedCount}/{usedCount}</p></div>
	</section>

	<section class="rounded-2xl border bg-card shadow-sm">
		<div class="border-b border-border px-4 py-3">
			<h2 class="text-base font-semibold text-foreground">Alur Opsi 4</h2>
			<p class="text-xs leading-5 text-muted-foreground">Modul ini berdiri sendiri, tapi tetap menjadi jembatan antara Bank Soal dan Asesmen.</p>
		</div>
		<div class="grid gap-3 p-4 md:grid-cols-3">
			<div class="rounded-xl border bg-background p-3 text-sm"><p class="font-semibold text-foreground">1. Bank Soal</p><p class="mt-1 text-xs leading-5 text-muted-foreground">Guru/admin membuat, mereview, dan menerbitkan soal.</p></div>
			<div class="rounded-xl border border-primary/30 bg-primary/5 p-3 text-sm"><p class="font-semibold text-primary">2. Paket Soal</p><p class="mt-1 text-xs leading-5 text-muted-foreground">Panitia merakit paket, validasi kesiapan, lock, clone/revisi.</p></div>
			<div class="rounded-xl border bg-background p-3 text-sm"><p class="font-semibold text-foreground">3. Asesmen/CBT</p><p class="mt-1 text-xs leading-5 text-muted-foreground">Kegiatan ujian memilih paket siap untuk rombel/sesi.</p></div>
		</div>
	</section>

	<section class="rounded-2xl border bg-card shadow-sm">
		<div class="border-b border-border px-4 py-3">
			<div class="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
				<div><h2 class="text-base font-semibold text-foreground">Daftar Paket</h2><p class="text-xs text-muted-foreground">Dibaca dari paket aktif yang sudah ada di backend.</p></div>
				<div class="grid gap-2 sm:grid-cols-3 lg:min-w-[42rem]">
					<input class="min-w-0 rounded-md border bg-background px-3 py-2 text-sm" placeholder="Cari paket/mapel..." bind:value={search} />
					<select class="min-w-0 rounded-md border bg-background px-3 py-2 text-sm" bind:value={subjectFilter}>
						<option value="all">Semua mapel</option>
						{#each subjects as subject}<option value={subject.subject_id}>{subject.subject_name}</option>{/each}
					</select>
					<select class="min-w-0 rounded-md border bg-background px-3 py-2 text-sm" bind:value={readinessFilter}>
						<option value="all">Semua status</option>
						<option value="empty">Kosong</option>
						<option value="ready">Siap dirakit</option>
						<option value="locked">Terkunci</option>
						<option value="used">Dipakai</option>
					</select>
				</div>
			</div>
		</div>
		{#if loading}
			<p class="px-4 py-8 text-center text-sm text-muted-foreground">Memuat paket soal…</p>
		{:else if filteredPackages.length === 0}
			<div class="px-4 py-8 text-center"><p class="text-sm font-semibold text-foreground">Belum ada paket sesuai filter.</p><p class="mt-1 text-xs text-muted-foreground">Iterasi berikutnya akan menambahkan tombol Buat Paket dan Builder Isi Soal.</p></div>
		{:else}
			<div class="overflow-x-auto">
				<table class="min-w-[820px] w-full text-left text-sm">
					<thead class="border-b bg-muted/40 text-xs text-muted-foreground">
						<tr><th class="px-4 py-3 font-semibold">Paket</th><th class="px-4 py-3 font-semibold">Mapel</th><th class="px-4 py-3 font-semibold">Soal</th><th class="px-4 py-3 font-semibold">Durasi</th><th class="px-4 py-3 font-semibold">Pemakaian</th><th class="px-4 py-3 font-semibold">Status</th><th class="px-4 py-3 font-semibold text-right">Aksi</th></tr>
					</thead>
					<tbody class="divide-y">
						{#each filteredPackages as item (item.id)}
							<tr class="hover:bg-muted/30">
								<td class="px-4 py-3"><p class="font-semibold text-foreground">{item.title}</p><p class="mt-1 max-w-md truncate text-xs text-muted-foreground">{item.description || 'Belum ada deskripsi.'}</p></td>
								<td class="px-4 py-3"><p class="font-medium text-foreground">{item.subject_name}</p><p class="text-xs text-muted-foreground">{item.subject_code || '-'}</p></td>
								<td class="px-4 py-3 font-semibold text-foreground">{item.question_count}</td>
								<td class="px-4 py-3 text-muted-foreground">{item.duration_minutes || 0} menit</td>
								<td class="px-4 py-3 text-muted-foreground">{item.session_count || 0} sesi</td>
								<td class="px-4 py-3"><span class={`rounded-full border px-2.5 py-1 text-[11px] font-semibold ${statusClass(item)}`}>{statusLabel(item)}</span></td>
								<td class="px-4 py-3 text-right"><a class="rounded-md border px-3 py-2 text-xs font-semibold text-foreground hover:bg-muted" href={`/asesmen?paket=${item.id}`}>Gunakan</a></td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</section>

	<section class="rounded-2xl border border-dashed bg-muted/30 p-4">
		<h2 class="text-sm font-semibold text-foreground">Batas iterasi ini</h2>
		<ul class="mt-2 list-disc space-y-1 pl-5 text-sm leading-6 text-muted-foreground">
			<li>Modul mandiri sudah dibuat untuk membaca paket aktif dan menjadi jembatan Bank Soal → Paket Soal → Asesmen.</li>
			<li>Belum membuat paket kosong karena endpoint backend create saat ini masih mewajibkan daftar soal.</li>
			<li>Iterasi berikutnya: form Buat Paket, Builder Isi Soal, Validasi Kesiapan, lalu Lock/Clone.</li>
		</ul>
	</section>
</div>
