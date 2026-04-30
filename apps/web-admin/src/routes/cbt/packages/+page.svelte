<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import LoadingButton from '$lib/components/LoadingButton.svelte';

	type CbtPackage = {
		id: string; subject_id: string; subject_name: string; subject_code: string;
		title: string; description: string; duration_minutes: number;
		randomize_questions: boolean; is_active: boolean;
		question_count: number; created_at: string;
	};
	type Question = {
		id: string; subject_id: string; subject_code: string;
		code: string; question_text: string; difficulty: string; status: string;
	};
	type Subject = { id: string; name: string; code: string; };

	let packages = $state<CbtPackage[]>([]);
	let allQuestions = $state<Question[]>([]);
	let subjects = $state<Subject[]>([]);
	let loading = $state(true);
	let error = $state('');
	let showForm = $state(false);

	let fSubjectId = $state('');
	let fTitle = $state('');
	let fDescription = $state('');
	let fDuration = $state(60);
	let fRandomize = $state(false);
	let fActive = $state(true);
	let fSelectedIds = $state<Set<string>>(new Set());
	let fBusy = $state(false);

	let questionPool = $derived(
		fSubjectId
			? allQuestions.filter(q => q.subject_id === fSubjectId && q.status === 'published')
			: []
	);

	function toggleQuestion(id: string) {
		const next = new Set(fSelectedIds);
		if (next.has(id)) next.delete(id);
		else next.add(id);
		fSelectedIds = next;
	}

	async function load() {
		try {
			const [pRes, qRes, aRes] = await Promise.all([
				fetch('/api/cbt/packages'),
				fetch('/api/cbt/questions'),
				fetch('/api/academic'),
			]);
			const pJson = await pRes.json();
			const qJson = await qRes.json();
			const aJson = await aRes.json();
			if (pJson.error) { error = pJson.error; return; }
			const pd = pJson.data ?? pJson;
			packages = pd.packages ?? [];
			allQuestions = qJson.data ?? qJson ?? [];
			subjects = (aJson.data ?? aJson)?.subjects ?? [];
		} catch {
			error = 'Gagal memuat data paket';
		} finally {
			loading = false;
		}
	}

	function showToast(msg: string) {
		toast.success(msg);
	}

	function showError(msg: string) {
		toast.error(msg);
	}

	async function createPackage() {
		if (!fSubjectId || !fTitle || !fDuration) return;
		fBusy = true;
		try {
			const res = await fetch('/api/cbt/packages', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					subject_id: fSubjectId, title: fTitle, description: fDescription,
					duration_minutes: fDuration, randomize_questions: fRandomize,
					is_active: fActive, question_ids: Array.from(fSelectedIds),
				}),
			});
			if (!res.ok) { const j = await res.json(); showError(j.error ?? 'Gagal'); return; }
			fSubjectId = ''; fTitle = ''; fDescription = ''; fDuration = 60;
			fRandomize = false; fActive = true; fSelectedIds = new Set();
			showForm = false;
			showToast('Paket ujian berhasil dibuat');
			await load();
		} finally { fBusy = false; }
	}

	async function deletePackage(id: string, title: string) {
		if (!confirm(`Hapus paket "${title}"?`)) return;
		await fetch(`/api/cbt/packages?id=${id}`, { method: 'DELETE' });
		showToast('Paket dihapus');
		await load();
	}

	onMount(load);
</script>

<svelte:head><title>Paket Ujian CBT — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-slate-800">Paket Ujian CBT</h1>
			<p class="text-sm text-slate-500 mt-1">Buat dan kelola paket soal untuk sesi ujian</p>
		</div>
		<LoadingButton onclick={() => (showForm = !showForm)}>
			{showForm ? 'Batal' : '+ Buat Paket'}
		</LoadingButton>
	</div>

	{#if error}
		<div class="rounded-md bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-800">{error}</div>
	{/if}

	{#if showForm}
		<Card.Root>
			<Card.Header class="pb-2">
				<Card.Title class="text-base">Buat Paket Ujian Baru</Card.Title>
			</Card.Header>
			<Card.Content class="space-y-4">
				<div class="grid gap-3 sm:grid-cols-2">
					<div>
						<label for="package-subject-id" class="text-xs text-slate-500 mb-1 block">Mata Pelajaran <span class="text-red-500">*</span></label>
						<select id="package-subject-id" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fSubjectId}>
							<option value="">-- Pilih --</option>
								{#each subjects as s (s.id)}
								<option value={s.id}>{s.code} — {s.name}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="package-title" class="text-xs text-slate-500 mb-1 block">Nama Paket <span class="text-red-500">*</span></label>
						<Input id="package-title" placeholder="mis: UTS Matematika Sem 1 2025" bind:value={fTitle} />
					</div>
					<div>
						<label for="package-duration" class="text-xs text-slate-500 mb-1 block">Durasi (menit) <span class="text-red-500">*</span></label>
						<Input id="package-duration" type="number" min={10} max={300} bind:value={fDuration} />
					</div>
					<div class="flex items-end gap-4 pb-1">
						<label class="flex items-center gap-2 text-sm">
							<input type="checkbox" bind:checked={fRandomize} class="rounded" />
							Acak urutan soal
						</label>
						<label class="flex items-center gap-2 text-sm">
							<input type="checkbox" bind:checked={fActive} class="rounded" />
							Paket aktif
						</label>
					</div>
				</div>

				<div>
					<label for="package-description" class="text-xs text-slate-500 mb-1 block">Deskripsi (opsional)</label>
					<Textarea id="package-description" placeholder="Keterangan paket ujian..." rows={2} bind:value={fDescription} />
				</div>

				{#if fSubjectId}
					<div>
						<div class="text-xs text-slate-500 mb-2 block">
							Pilih Soal dari Bank ({questionPool.length} soal tersedia)
							{#if fSelectedIds.size > 0}
								— <span class="text-green-700 font-medium">{fSelectedIds.size} dipilih</span>
							{/if}
						</div>
						{#if questionPool.length === 0}
							<p class="text-sm text-slate-400 py-4 text-center border rounded-md">
								Belum ada soal berstatus "Terbit" untuk mata pelajaran ini
							</p>
						{:else}
							<div class="border rounded-md max-h-64 overflow-y-auto">
									{#each questionPool as q (q.id)}
									<label class="flex items-start gap-3 px-3 py-2 hover:bg-slate-50 cursor-pointer border-b last:border-b-0">
										<input type="checkbox" checked={fSelectedIds.has(q.id)} onchange={() => toggleQuestion(q.id)} class="mt-0.5 rounded" />
										<div class="flex-1 min-w-0">
											<p class="text-sm line-clamp-1">{q.question_text}</p>
											<div class="flex gap-1 mt-0.5">
												{#if q.code}
													<span class="text-xs text-slate-400 font-mono">{q.code}</span>
												{/if}
										<Badge variant="outline" class="text-xs py-0">{q.difficulty === 'easy' ? 'Mudah' : q.difficulty === 'medium' ? 'Sedang' : q.difficulty === 'hard' ? 'Sulit' : q.difficulty}</Badge>
											</div>
										</div>
									</label>
								{/each}
							</div>
						{/if}
					</div>
				{/if}

				<div class="flex gap-2">
					<LoadingButton disabled={fBusy || !fSubjectId || !fTitle || !fDuration} onclick={createPackage} loading={fBusy} loadingLabel="Menyimpan...">
						{`Buat Paket${fSelectedIds.size > 0 ? ` (${fSelectedIds.size} soal)` : ''}`}
					</LoadingButton>
					<LoadingButton variant="outline" onclick={() => (showForm = false)}>Batal</LoadingButton>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	{#if loading}
		<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
			<Card.Content class="space-y-3 p-6">
				{#each Array.from({ length: 5 }) as _, index (`package-row-skeleton-${index}`)}
					<div class="grid gap-3 lg:grid-cols-[1.2fr_0.6fr_0.5fr_0.5fr_0.5fr_0.6fr_auto] lg:items-center">
						<Skeleton class="h-5 w-40" />
						<Skeleton class="h-6 w-16" />
						<Skeleton class="h-5 w-14" />
						<Skeleton class="h-5 w-12" />
						<Skeleton class="h-6 w-12" />
						<Skeleton class="h-6 w-16" />
						<Skeleton class="h-9 w-20 justify-self-end" />
					</div>
				{/each}
			</Card.Content>
		</Card.Root>
	{:else}
		<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
			<Card.Header class="pb-2">
				<Card.Title class="text-base">Daftar Paket ({packages.length})</Card.Title>
			</Card.Header>
			<Card.Content class="p-0">
				<div class="hidden overflow-x-auto lg:block">
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Nama Paket</Table.Head>
							<Table.Head>Mapel</Table.Head>
							<Table.Head>Durasi</Table.Head>
							<Table.Head>Jml Soal</Table.Head>
							<Table.Head>Acak</Table.Head>
							<Table.Head>Status</Table.Head>
							<Table.Head></Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each packages as p (p.id)}
							<Table.Row>
								<Table.Cell class="font-medium">{p.title}</Table.Cell>
								<Table.Cell>
									<Badge variant="outline" class="text-xs">{p.subject_code}</Badge>
								</Table.Cell>
								<Table.Cell class="text-slate-600">{p.duration_minutes} mnt</Table.Cell>
								<Table.Cell>
									<span class="font-mono text-sm">{p.question_count}</span>
								</Table.Cell>
								<Table.Cell>
									{#if p.randomize_questions}
										<Badge class="bg-purple-100 text-purple-700 border-purple-200 text-xs">Ya</Badge>
									{:else}
										<Badge variant="secondary" class="text-xs">Tidak</Badge>
									{/if}
								</Table.Cell>
								<Table.Cell>
									{#if p.is_active}
										<Badge class="bg-emerald-100 text-emerald-700 border-emerald-200">Aktif</Badge>
									{:else}
										<Badge variant="secondary">Nonaktif</Badge>
									{/if}
								</Table.Cell>
								<Table.Cell>
									<LoadingButton variant="destructive" size="xs" onclick={() => deletePackage(p.id, p.title)}>Hapus</LoadingButton>
								</Table.Cell>
							</Table.Row>
						{:else}
							<Table.Row>
								<Table.Cell colspan={7} class="text-center text-slate-400 py-8">Belum ada paket ujian</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
				</div>

				<div class="grid gap-3 p-4 lg:hidden">
					{#each packages as p (p.id)}
						<div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
							<div class="flex items-start justify-between gap-3">
								<div class="min-w-0">
									<p class="text-sm font-semibold text-slate-900">{p.title}</p>
									<p class="mt-1 text-xs text-slate-500">{p.subject_name} ({p.subject_code})</p>
								</div>
								{#if p.is_active}
									<Badge class="bg-emerald-100 text-emerald-700 border-emerald-200">Aktif</Badge>
								{:else}
									<Badge variant="secondary">Nonaktif</Badge>
								{/if}
							</div>
							<div class="mt-3 flex flex-wrap items-center gap-2">
								<Badge variant="outline" class="text-xs">{p.duration_minutes} menit</Badge>
								<Badge variant="secondary">{p.question_count} soal</Badge>
								{#if p.randomize_questions}
									<Badge class="bg-emerald-50 text-emerald-700 border-emerald-200 text-xs">Acak</Badge>
								{/if}
							</div>
							{#if p.description}
								<p class="mt-3 text-sm text-slate-600">{p.description}</p>
							{/if}
							<div class="mt-4">
								<LoadingButton variant="destructive" size="sm" class="w-full" onclick={() => deletePackage(p.id, p.title)}>Hapus</LoadingButton>
							</div>
						</div>
					{:else}
						<div class="rounded-2xl border border-dashed border-slate-300 bg-slate-50 px-4 py-10 text-center text-sm text-slate-500">
							Belum ada paket ujian
						</div>
					{/each}
				</div>
			</Card.Content>
		</Card.Root>
	{/if}
</div>
