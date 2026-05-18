<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { clientApiPath, clientApiPathWithQuery, readClientApiData } from '$lib/client/api';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import { toast } from '$lib/components/ui/sonner';

	type PackageRow = {
		id: string; event_id?: string | null; subject_id: string; subject_name?: string; subject_code?: string;
		title: string; description: string; duration_minutes: number; randomize_questions: boolean; randomize_options?: boolean;
		source_mode?: string; draw_pg_count?: number; draw_essay_count?: number; random_seed?: string; is_active: boolean;
		locked_at?: string | null; lock_reason?: string | null; snapshot_version?: number; session_count?: number;
	};
	type PackageQuestion = {
		question_id: string; position: number; points: number; question_code?: string; question_text?: string;
		question_type?: string; difficulty?: string; status?: string; workflow_status?: string; target_level?: string; cp_ref?: string; tp_ref?: string; kd_ref?: string;
		material_topic?: string; cognitive_level?: string; hots_flag?: boolean;
	};
	type Readiness = {
		status: string; target_pg_count: number; target_essay_count: number; missing_pg_count: number; missing_essay_count: number;
		question_count: number; pg_count: number; essay_count: number; total_points: number; published_count: number; unpublished_count: number;
		metadata_gap_count: number; session_count: number; locked: boolean; ready: boolean;
	};
	type DetailPayload = { package: PackageRow; questions: PackageQuestion[]; readiness: Readiness };
	type PoolQuestion = {
		id: string; subject_id: string; code?: string; question_text?: string; question_type?: string; status?: string; event_id?: string | null;
		target_level?: string; difficulty?: string; workflow_status?: string; cp_ref?: string; tp_ref?: string; kd_ref?: string; material_topic?: string; cognitive_level?: string; hots_flag?: boolean;
		author_username?: string; author_display_name?: string; created_at?: string; updated_at?: string; package_count?: number;
	};
	type QuestionListPayload = { items?: PoolQuestion[]; meta?: { total?: number } };

	const packageId = page.params.id ?? '';
	let detailPromise = $state<Promise<DetailPayload> | null>(null);
	let detail = $state<DetailPayload | null>(null);
	let pool = $state<PoolQuestion[]>([]);
	let selectedPool = $state(new Set<string>());
	let busy = $state('');
	type PackageTab = 'questions' | 'pool' | 'blueprint' | 'lock';
	let activeTab = $state<PackageTab>('questions');

	function tabFromHash(hash: string): PackageTab | null {
		const value = hash.replace(/^#/, '').toLowerCase();
		if (value === 'soal' || value === 'questions' || value === 'isi-soal') return 'questions';
		if (value === 'tambah-soal' || value === 'bank-soal' || value === 'pool') return 'pool';
		if (value === 'blueprint' || value === 'kisi-kisi' || value === 'mutu') return 'blueprint';
		if (value === 'lock' || value === 'kunci' || value === 'revisi') return 'lock';
		return null;
	}

	function syncTabFromHash() {
		const tab = tabFromHash(window.location.hash);
		if (tab) activeTab = tab;
	}

	function selectTab(tab: PackageTab) {
		activeTab = tab;
		const hashByTab: Record<PackageTab, string> = {
			questions: '#soal',
			pool: '#tambah-soal',
			blueprint: '#blueprint',
			lock: '#kunci',
		};
		window.history.replaceState(null, '', hashByTab[tab]);
	}

	let title = $state('');
	let description = $state('');
	let duration = $state(60);
	let randomizeQuestions = $state(false);
	let randomizeOptions = $state(false);
	let active = $state(true);
	let targetPg = $state(20);
	let targetEssay = $state(5);
	let rows = $state<PackageQuestion[]>([]);
	let poolSearch = $state('');
	let poolLevel = $state('all');
	let poolType = $state('all');
	let poolStatus = $state('published');
	let poolCognitive = $state('all');
	let poolHots = $state('all');
	let poolDifficulty = $state('all');
	let poolMetadata = $state('all');
	let poolTopic = $state('all');
	let poolCurriculumSearch = $state('');
	let poolAuthor = $state('all');
	let poolCreatedFrom = $state('');
	let poolCreatedTo = $state('');
	let poolUsage = $state('all');
	let poolSort = $state('metadata_first');

	let isLocked = $derived(Boolean(detail?.package.locked_at || detail?.readiness.locked));
	let poolForSubject = $derived(pool.filter((q) => q.subject_id === detail?.package.subject_id));
	let availablePool = $derived(sortPoolQuestions(poolForSubject.filter(matchesPoolFilters).filter((q) => !rows.some((row) => row.question_id === q.id))));
	let poolAuthorOptions = $derived(Array.from(
		new Map(poolForSubject
			.map((q) => [String(q.author_username ?? '').trim(), String(q.author_display_name ?? q.author_username ?? '').trim()] as const)
			.filter(([username]) => Boolean(username)))
		.entries()
	).sort((a, b) => a[1].localeCompare(b[1])));
	let selectedQuestions = $derived(availablePool.filter((q) => selectedPool.has(q.id)));
	let missingLabel = $derived(detail ? `${Math.max(0, targetPg - countType(rows, 'multiple_choice'))} PG + ${Math.max(0, targetEssay - countType(rows, 'essay'))} Essay kurang` : '');

	function typeLabel(value?: string) {
		const map: Record<string, string> = { multiple_choice: 'PG', essay: 'Essay', true_false: 'Benar/Salah', short_answer: 'Isian' };
		return map[value ?? ''] ?? (value || 'Lainnya');
	}

	function countType(items: PackageQuestion[], type: string) {
		return items.filter((item) => item.question_type === type).length;
	}

	function hasMetadataGap(item: PackageQuestion | PoolQuestion) {
		return !String(item.target_level ?? '').trim() || !String(item.cp_ref ?? '').trim() || !(String(item.tp_ref ?? '').trim() || String(item.kd_ref ?? '').trim()) || !String(item.cognitive_level ?? '').trim();
	}

	function inferLevelFromText(text?: string) {
		const upper = String(text ?? '').toUpperCase();
		if (/\bVII\b/.test(upper)) return 'VII';
		if (/\bVIII\b/.test(upper)) return 'VIII';
		if (/\bIX\b/.test(upper)) return 'IX';
		return 'all';
	}

	function difficultyLabel(value?: string) {
		return { easy: 'Mudah', medium: 'Sedang', hard: 'Sulit' }[value ?? ''] ?? (value || 'Belum');
	}

	function statusLabel(value?: string) {
		return { published: 'Terbit', draft: 'Konsep', review: 'Review', archived: 'Arsip', rejected: 'Ditolak' }[value ?? ''] ?? (value || 'Belum');
	}

	function userDisplayName(item: Pick<PoolQuestion, 'author_display_name' | 'author_username'>) {
		return String(item.author_display_name || item.author_username || '').trim();
	}

	function parseDateOnly(value: string, endOfDay = false) {
		if (!value) return null;
		const parsed = new Date(`${value}T${endOfDay ? '23:59:59.999' : '00:00:00.000'}`);
		return Number.isNaN(parsed.getTime()) ? null : parsed;
	}

	function dateLabel(value?: string) {
		if (!value) return 'Tanggal kosong';
		const parsed = new Date(value);
		if (Number.isNaN(parsed.getTime())) return value;
		return new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium' }).format(parsed);
	}

	function matchesPoolFilters(item: PoolQuestion) {
		const term = poolSearch.trim().toLowerCase();
		if (term && !`${item.code ?? ''} ${item.question_text ?? ''} ${item.material_topic ?? ''}`.toLowerCase().includes(term)) return false;
		if (poolLevel !== 'all' && item.target_level !== poolLevel) return false;
		if (poolType !== 'all' && item.question_type !== poolType) return false;
		if (poolStatus !== 'all' && item.status !== poolStatus) return false;
		if (poolCognitive !== 'all' && item.cognitive_level !== poolCognitive) return false;
		if (poolHots === 'hots' && !item.hots_flag) return false;
		if (poolHots === 'non_hots' && item.hots_flag) return false;
		if (poolDifficulty !== 'all' && item.difficulty !== poolDifficulty) return false;
		if (poolMetadata === 'complete' && hasMetadataGap(item)) return false;
		if (poolMetadata === 'gap' && !hasMetadataGap(item)) return false;
		if (poolTopic !== 'all' && String(item.material_topic ?? '') !== poolTopic) return false;
		if (poolAuthor !== 'all' && String(item.author_username ?? '') !== poolAuthor) return false;
		if (poolUsage === 'unused' && Number(item.package_count ?? 0) > 0) return false;
		if (poolUsage === 'used' && Number(item.package_count ?? 0) === 0) return false;
		const createdAt = item.created_at ? new Date(item.created_at) : null;
		const from = parseDateOnly(poolCreatedFrom);
		const to = parseDateOnly(poolCreatedTo, true);
		if (from && (!createdAt || createdAt < from)) return false;
		if (to && (!createdAt || createdAt > to)) return false;
		const cur = poolCurriculumSearch.trim().toLowerCase();
		if (cur && !`${item.cp_ref ?? ''} ${item.tp_ref ?? ''} ${item.kd_ref ?? ''}`.toLowerCase().includes(cur)) return false;
		return true;
	}

	function sortPoolQuestions(items: PoolQuestion[]) {
		return [...items].sort((a, b) => {
			if (poolSort === 'newest') return new Date(b.created_at ?? 0).getTime() - new Date(a.created_at ?? 0).getTime();
			if (poolSort === 'oldest') return new Date(a.created_at ?? 0).getTime() - new Date(b.created_at ?? 0).getTime();
			if (poolSort === 'author') return userDisplayName(a).localeCompare(userDisplayName(b)) || String(a.code ?? '').localeCompare(String(b.code ?? ''));
			if (poolSort === 'unused_first') return Number(a.package_count ?? 0) - Number(b.package_count ?? 0) || String(a.code ?? '').localeCompare(String(b.code ?? ''));
			if (poolSort === 'hots_first') return Number(Boolean(b.hots_flag)) - Number(Boolean(a.hots_flag));
			if (poolSort === 'difficulty') return String(a.difficulty ?? '').localeCompare(String(b.difficulty ?? '')) || String(a.code ?? '').localeCompare(String(b.code ?? ''));
			if (poolSort === 'type') return String(a.question_type ?? '').localeCompare(String(b.question_type ?? '')) || String(a.code ?? '').localeCompare(String(b.code ?? ''));
			if (poolSort === 'code') return String(a.code ?? '').localeCompare(String(b.code ?? ''));
			return Number(hasMetadataGap(a)) - Number(hasMetadataGap(b)) || String(a.code ?? '').localeCompare(String(b.code ?? ''));
		});
	}

	function syncForm(payload: DetailPayload) {
		detail = payload;
		title = payload.package.title;
		description = payload.package.description ?? '';
		duration = payload.package.duration_minutes || 60;
		randomizeQuestions = Boolean(payload.package.randomize_questions);
		randomizeOptions = Boolean(payload.package.randomize_options);
		active = Boolean(payload.package.is_active);
		targetPg = payload.package.draw_pg_count || 20;
		targetEssay = payload.package.draw_essay_count || 5;
		rows = [...payload.questions].sort((a, b) => a.position - b.position);
		if (poolLevel === 'all') poolLevel = inferLevelFromText(payload.package.title);
		selectedPool = new Set();
	}

	async function loadDetail() {
		const payload = await fetch(clientApiPath`/api/asesmen/packages/${packageId}`).then((res) => readClientApiData<DetailPayload>(res));
		syncForm(payload);
		return payload;
	}

	async function fetchQuestionPool(payload: DetailPayload) {
		const params = new URLSearchParams({ limit: '1000', offset: '0', scope: payload.package.event_id ? 'event_pool' : 'global' });
		params.set('subject_id', payload.package.subject_id);
		if (payload.package.event_id) params.set('event_id', payload.package.event_id);
		const q = await fetch(clientApiPathWithQuery('/api/bank-soal/questions', params)).then((res) => readClientApiData<QuestionListPayload | PoolQuestion[]>(res));
		pool = Array.isArray(q) ? q : (q.items ?? []);
	}

	async function refresh() {
		detailPromise = loadDetail();
		const payload = await detailPromise;
		await fetchQuestionPool(payload);
		return payload;
	}

	onMount(() => {
		syncTabFromHash();
		window.addEventListener('hashchange', syncTabFromHash);
		void refresh();
		return () => window.removeEventListener('hashchange', syncTabFromHash);
	});

	async function saveMetadata() {
		if (!detail || isLocked) return;
		busy = 'metadata';
		try {
			const payload = await fetch(clientApiPath`/api/asesmen/packages/${packageId}`, {
				method: 'PUT', headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ title, description, duration_minutes: duration, randomize_questions: randomizeQuestions, randomize_options: randomizeOptions, source_mode: detail.package.source_mode || 'teacher_class', draw_pg_count: targetPg, draw_essay_count: targetEssay, random_seed: detail.package.random_seed || '', is_active: active })
			}).then((res) => readClientApiData<DetailPayload>(res));
			syncForm(payload); toast.success('Metadata paket tersimpan');
		} catch (e) { toast.error((e as Error).message); } finally { busy = ''; }
	}

	async function saveQuestions() {
		if (isLocked) return;
		busy = 'questions';
		try {
			const payload = await fetch(clientApiPath`/api/asesmen/packages/${packageId}/questions`, {
				method: 'PUT', headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ questions: rows.map((row) => ({ question_id: row.question_id, points: row.points || 1 })) })
			}).then((res) => readClientApiData<DetailPayload>(res));
			syncForm(payload); await fetchQuestionPool(payload); toast.success('Isi soal paket tersimpan');
		} catch (e) { toast.error((e as Error).message); } finally { busy = ''; }
	}

	function addSelected() {
		const next = [...rows];
		for (const question of selectedQuestions) {
			next.push({ question_id: question.id, position: next.length + 1, points: 1, question_code: question.code, question_text: question.question_text, question_type: question.question_type, status: question.status, cp_ref: question.cp_ref, tp_ref: question.tp_ref, kd_ref: question.kd_ref, material_topic: question.material_topic, cognitive_level: question.cognitive_level, hots_flag: question.hots_flag, target_level: question.target_level, difficulty: question.difficulty });
		}
		rows = next; selectedPool = new Set(); activeTab = 'questions';
	}

	function autofillTarget() {
		const next = [...rows];
		const needPg = Math.max(0, targetPg - countType(next, 'multiple_choice'));
		const needEssay = Math.max(0, targetEssay - countType(next, 'essay'));
		const picked = [...availablePool.filter((q) => q.question_type === 'multiple_choice').slice(0, needPg), ...availablePool.filter((q) => q.question_type === 'essay').slice(0, needEssay)];
		for (const q of picked) next.push({ question_id: q.id, position: next.length + 1, points: 1, question_code: q.code, question_text: q.question_text, question_type: q.question_type, status: q.status, cp_ref: q.cp_ref, tp_ref: q.tp_ref, kd_ref: q.kd_ref, material_topic: q.material_topic, cognitive_level: q.cognitive_level, hots_flag: q.hots_flag, target_level: q.target_level, difficulty: q.difficulty });
		rows = next;
		toast.info(`${picked.length} soal ditambahkan ke draft paket`);
	}

	function removeRow(index: number) { rows = rows.filter((_, i) => i !== index).map((row, i) => ({ ...row, position: i + 1 })); }
	function moveRow(index: number, direction: -1 | 1) { const target = index + direction; if (target < 0 || target >= rows.length) return; const copy = [...rows]; [copy[index], copy[target]] = [copy[target], copy[index]]; rows = copy.map((row, i) => ({ ...row, position: i + 1 })); }
	function setPoints(index: number, value: number) { rows = rows.map((row, i) => i === index ? { ...row, points: Math.max(1, Math.min(100, Math.round(value || 1))) } : row); }
	function togglePool(id: string) { const next = new Set(selectedPool); next.has(id) ? next.delete(id) : next.add(id); selectedPool = next; }

	async function clonePackage() {
		if (!detail) return; const newTitle = window.prompt('Nama paket revisi/clone', `${detail.package.title} - Revisi`); if (!newTitle) return;
		busy = 'clone';
		try { const payload = await fetch(clientApiPath`/api/asesmen/packages/${packageId}/clone`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ title: newTitle }) }).then((res) => readClientApiData<DetailPayload>(res)); toast.success('Paket revisi dibuat'); await goto(`/asesmen/paket/${payload.package.id}`); }
		catch (e) { toast.error((e as Error).message); } finally { busy = ''; }
	}

	async function lockPackage() {
		if (!detail || !window.confirm('Kunci paket dan buat snapshot? Setelah terkunci, paket tidak bisa diedit langsung.')) return;
		busy = 'lock';
		try { await fetch(clientApiPath`/api/asesmen/packages/${packageId}/lock`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ reason: 'Dikunci dari Paket Builder' }) }).then((res) => readClientApiData(res)); toast.success('Paket terkunci dan snapshot dibuat'); await refresh(); }
		catch (e) { toast.error((e as Error).message); } finally { busy = ''; }
	}
</script>

<svelte:head><title>Paket Builder — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-5 p-4 md:p-6">
	<AsyncContent promise={detailPromise}>
		{#snippet children()}
		{#if detail}
			<div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-primary">Paket Builder</p>
					<h1 class="text-2xl font-semibold tracking-tight">{detail.package.title}</h1>
					<p class="text-sm text-muted-foreground">{detail.package.subject_name ?? detail.package.subject_code} · {detail.readiness.session_count} sesi memakai paket ini</p>
				</div>
				<div class="flex flex-wrap gap-2">
					<Badge class={detail.readiness.ready ? 'bg-primary/15 text-primary border-primary/20' : 'bg-warning/10 text-warning border-warning/30'}>{detail.readiness.status}</Badge>
					{#if isLocked}<Badge class="bg-destructive/10 text-destructive border-destructive/30">Terkunci v{detail.package.snapshot_version ?? 1}</Badge>{/if}
					<a class="rounded-md border px-3 py-2 text-sm hover:bg-muted" href="/asesmen/paket">Kembali</a>
				</div>
			</div>

			<div class="grid gap-4 md:grid-cols-4">
				<Card.Root><Card.Content class="p-4"><p class="text-xs text-muted-foreground">PG</p><p class="text-2xl font-semibold">{countType(rows, 'multiple_choice')}/{targetPg}</p></Card.Content></Card.Root>
				<Card.Root><Card.Content class="p-4"><p class="text-xs text-muted-foreground">Essay</p><p class="text-2xl font-semibold">{countType(rows, 'essay')}/{targetEssay}</p></Card.Content></Card.Root>
				<Card.Root><Card.Content class="p-4"><p class="text-xs text-muted-foreground">Total Poin</p><p class="text-2xl font-semibold">{rows.reduce((s, r) => s + (r.points || 0), 0)}</p></Card.Content></Card.Root>
				<Card.Root><Card.Content class="p-4"><p class="text-xs text-muted-foreground">Gap</p><p class="text-sm font-semibold">{missingLabel}</p></Card.Content></Card.Root>
			</div>

			<Card.Root>
				<Card.Header><Card.Title class="text-base">Identitas Paket</Card.Title><Card.Description>{isLocked ? 'Paket terkunci. Buat revisi jika perlu mengubah.' : 'Edit identitas, durasi, pengacakan, dan target kisi-kisi.'}</Card.Description></Card.Header>
				<Card.Content class="grid gap-3 md:grid-cols-3">
					<label class="space-y-1 md:col-span-2"><span class="text-xs text-muted-foreground">Nama Paket</span><Input bind:value={title} disabled={isLocked} /></label>
					<label class="space-y-1"><span class="text-xs text-muted-foreground">Durasi menit</span><Input type="number" bind:value={duration} min="1" max="360" disabled={isLocked} /></label>
					<label class="space-y-1 md:col-span-3"><span class="text-xs text-muted-foreground">Deskripsi</span><Textarea bind:value={description} disabled={isLocked} /></label>
					<label class="space-y-1"><span class="text-xs text-muted-foreground">Target PG</span><Input type="number" bind:value={targetPg} min="0" disabled={isLocked} /></label>
					<label class="space-y-1"><span class="text-xs text-muted-foreground">Target Essay</span><Input type="number" bind:value={targetEssay} min="0" disabled={isLocked} /></label>
					<div class="flex flex-wrap items-center gap-4 text-sm">
						<label><input type="checkbox" bind:checked={randomizeQuestions} disabled={isLocked} /> Acak soal</label>
						<label><input type="checkbox" bind:checked={randomizeOptions} disabled={isLocked} /> Acak opsi</label>
						<label><input type="checkbox" bind:checked={active} disabled={isLocked} /> Aktif</label>
					</div>
					<div class="md:col-span-3 flex gap-2">
						<LoadingButton onclick={saveMetadata} loading={busy === 'metadata'} disabled={isLocked}>Simpan Identitas</LoadingButton>
						<LoadingButton variant="outline" onclick={clonePackage} loading={busy === 'clone'}>Buat Revisi/Salinan</LoadingButton>
					</div>
				</Card.Content>
			</Card.Root>

			<div class="flex flex-wrap gap-2">
				{#each [['questions','Soal Dalam Paket'], ['pool','Tambah dari Bank Soal'], ['blueprint','Kisi-kisi & Mutu'], ['lock','Kunci & Salinan']] as tab}
					<button class={`rounded-md border px-3 py-2 text-sm ${activeTab === tab[0] ? 'bg-primary text-primary-foreground' : 'hover:bg-muted'}`} onclick={() => selectTab(tab[0] as PackageTab)}>{tab[1]}</button>
				{/each}
			</div>

			{#if activeTab === 'questions'}
				<Card.Root><Card.Header><Card.Title class="text-base">Soal Dalam Paket ({rows.length})</Card.Title></Card.Header><Card.Content class="space-y-3">
					<div class="flex flex-wrap gap-2"><LoadingButton onclick={saveQuestions} loading={busy === 'questions'} disabled={isLocked}>Simpan Isi Soal</LoadingButton><button class="rounded-md border px-3 py-2 text-sm" disabled={isLocked} onclick={autofillTarget}>Autofill Target dari Soal Terbit</button></div>
					<div class="overflow-x-auto"><Table.Root><Table.Header><Table.Row><Table.Head>No</Table.Head><Table.Head>Soal</Table.Head><Table.Head>Bentuk</Table.Head><Table.Head>Bobot</Table.Head><Table.Head>Mutu</Table.Head><Table.Head>Aksi</Table.Head></Table.Row></Table.Header><Table.Body>
						{#each rows as row, i (row.question_id)}<Table.Row><Table.Cell>{i + 1}</Table.Cell><Table.Cell><p class="font-medium">{row.question_code || row.question_id}</p><p class="line-clamp-2 text-xs text-muted-foreground">{row.question_text}</p></Table.Cell><Table.Cell>{typeLabel(row.question_type)}</Table.Cell><Table.Cell><Input class="w-20" type="number" value={row.points} min="1" max="100" disabled={isLocked} oninput={(e) => setPoints(i, Number((e.currentTarget as HTMLInputElement).value))} /></Table.Cell><Table.Cell>{#if hasMetadataGap(row)}<Badge class="bg-warning/10 text-warning border-warning/30">Metadata kurang</Badge>{:else}<Badge variant="outline">OK</Badge>{/if}</Table.Cell><Table.Cell><div class="flex gap-1"><button class="rounded border px-2 py-1 text-xs" disabled={isLocked || i === 0} onclick={() => moveRow(i, -1)}>↑</button><button class="rounded border px-2 py-1 text-xs" disabled={isLocked || i === rows.length - 1} onclick={() => moveRow(i, 1)}>↓</button><button class="rounded border px-2 py-1 text-xs text-destructive" disabled={isLocked} onclick={() => removeRow(i)}>Hapus</button></div></Table.Cell></Table.Row>{:else}<Table.Row><Table.Cell colspan={6} class="py-8 text-center text-muted-foreground">Belum ada soal dalam paket.</Table.Cell></Table.Row>{/each}
					</Table.Body></Table.Root></div>
				</Card.Content></Card.Root>
			{:else if activeTab === 'pool'}
				<Card.Root>
					<Card.Header>
						<Card.Title class="text-base">Tambah dari Bank Soal</Card.Title>
						<Card.Description>{availablePool.length} soal tersedia sesuai filter untuk mapel ini. Default filter status = Terbit agar aman untuk paket resmi.</Card.Description>
					</Card.Header>
					<Card.Content class="space-y-4">
						<div class="rounded-2xl border bg-muted/30 p-3">
							<div class="grid gap-3 md:grid-cols-2 lg:grid-cols-5">
								<Input placeholder="Cari kode, teks soal, materi..." bind:value={poolSearch} />
								<select class="rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={poolLevel}>
									<option value="all">Semua tingkat</option><option value="VII">VII</option><option value="VIII">VIII</option><option value="IX">IX</option>
								</select>
								<select class="rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={poolType}>
									<option value="all">Semua jenis</option><option value="multiple_choice">Pilihan Ganda</option><option value="essay">Essay</option><option value="true_false">Benar/Salah</option><option value="short_answer">Isian</option>
								</select>
								<select class="rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={poolStatus}>
									<option value="published">Terbit saja</option><option value="all">Semua status</option><option value="draft">Konsep</option><option value="review">Verifikasi</option><option value="archived">Arsip</option>
								</select>
								<select class="rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={poolMetadata}>
									<option value="all">Semua identitas</option><option value="complete">Identitas lengkap</option><option value="gap">Identitas kurang</option>
								</select>
								<select class="rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={poolCognitive}>
									<option value="all">Semua level</option><option value="C1">C1</option><option value="C2">C2</option><option value="C3">C3</option><option value="C4">C4</option><option value="C5">C5</option><option value="C6">C6</option>
								</select>
								<select class="rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={poolHots}>
									<option value="all">Semua HOTS</option><option value="hots">HOTS</option><option value="non_hots">Non-HOTS</option>
								</select>
								<select class="rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={poolDifficulty}>
									<option value="all">Semua kesulitan</option><option value="easy">Mudah</option><option value="medium">Sedang</option><option value="hard">Sulit</option>
								</select>
								<Input placeholder="Cari CP/TP/KD..." bind:value={poolCurriculumSearch} />
								<select class="rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={poolAuthor}>
									<option value="all">Semua pembuat</option>
									{#each poolAuthorOptions as [username, displayName]}<option value={username}>{displayName}</option>{/each}
								</select>
								<label class="space-y-1 text-xs text-muted-foreground"><span>Dari tanggal</span><Input type="date" bind:value={poolCreatedFrom} /></label>
								<label class="space-y-1 text-xs text-muted-foreground"><span>Sampai tanggal</span><Input type="date" bind:value={poolCreatedTo} /></label>
								<select class="rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={poolUsage}>
									<option value="all">Semua pemakaian</option><option value="unused">Belum dipakai paket</option><option value="used">Sudah dipakai paket</option>
								</select>
								<select class="rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={poolSort}>
									<option value="metadata_first">Identitas lengkap dulu</option><option value="newest">Terbaru dibuat</option><option value="oldest">Terlama dibuat</option><option value="author">Pembuat A-Z</option><option value="unused_first">Belum dipakai dulu</option><option value="hots_first">HOTS dulu</option><option value="difficulty">Kesulitan</option><option value="type">Jenis soal</option><option value="code">Kode A-Z</option>
								</select>
							</div>
							{#if poolStatus !== 'published'}<p class="mt-2 text-xs text-warning">Mode pemeriksaan: soal belum Terbit bisa dilihat, tetapi layanan sistem tetap menolak jika dimasukkan ke paket resmi.</p>{/if}
						</div>
						<button class="rounded-md bg-primary px-3 py-2 text-sm text-primary-foreground disabled:opacity-50" disabled={isLocked || selectedQuestions.length === 0} onclick={addSelected}>Tambah {selectedQuestions.length} Soal</button>
						<div class="grid gap-2 md:grid-cols-2">
							{#each availablePool as q (q.id)}
								<label class="rounded-xl border p-3 text-sm">
									<div class="flex gap-2">
										<input type="checkbox" checked={selectedPool.has(q.id)} disabled={isLocked || q.status !== 'published'} onchange={() => togglePool(q.id)} />
										<div class="min-w-0 flex-1">
											<p class="font-semibold">{q.code || q.id} · {typeLabel(q.question_type)}</p>
											<p class="line-clamp-2 text-muted-foreground">{q.question_text}</p>
											<div class="mt-2 flex flex-wrap gap-1">
												{#if q.target_level}<Badge variant="outline">Tingkat {q.target_level}</Badge>{:else}<Badge class="bg-warning/10 text-warning border-warning/30">Tingkat kosong</Badge>{/if}
												<Badge variant="outline">{statusLabel(q.status)}</Badge>
												<Badge variant="outline">{difficultyLabel(q.difficulty)}</Badge>
												{#if q.cognitive_level}<Badge variant="outline">{q.cognitive_level}</Badge>{/if}
												{#if q.hots_flag}<Badge variant="outline">HOTS</Badge>{/if}
												{#if userDisplayName(q)}<Badge variant="outline">Pembuat: {userDisplayName(q)}</Badge>{/if}
												{#if q.created_at}<Badge variant="outline">{dateLabel(q.created_at)}</Badge>{/if}
												{#if Number(q.package_count ?? 0) > 0}<Badge variant="outline">Dipakai {q.package_count} paket</Badge>{:else}<Badge variant="outline">Belum dipakai</Badge>{/if}
												{#if hasMetadataGap(q)}<Badge class="bg-warning/10 text-warning border-warning/30">Identitas kurang</Badge>{/if}
											</div>
											{#if q.material_topic}<p class="mt-1 text-xs text-muted-foreground">Materi: {q.material_topic}</p>{/if}
										</div>
									</div>
								</label>
							{:else}<p class="text-sm text-muted-foreground">Belum ada soal sesuai filter yang bisa ditambahkan.</p>{/each}
						</div>
					</Card.Content>
				</Card.Root>
			{:else if activeTab === 'blueprint'}
				<Card.Root><Card.Header><Card.Title class="text-base">Kisi-kisi & Mutu</Card.Title></Card.Header><Card.Content class="grid gap-3 md:grid-cols-3"><div class="rounded-xl border p-4"><p class="text-xs text-muted-foreground">Status</p><p class="text-lg font-semibold">{detail.readiness.status}</p></div><div class="rounded-xl border p-4"><p class="text-xs text-muted-foreground">Identitas kurang</p><p class="text-lg font-semibold">{rows.filter(hasMetadataGap).length}</p></div><div class="rounded-xl border p-4"><p class="text-xs text-muted-foreground">HOTS</p><p class="text-lg font-semibold">{rows.filter((r) => r.hots_flag).length}</p></div><div class="md:col-span-3 text-sm text-muted-foreground">Distribusi: PG {countType(rows, 'multiple_choice')}, Essay {countType(rows, 'essay')}. Lengkapi CP/TP/KD dan level kognitif di Bank Soal untuk menutup kekurangan identitas soal.</div></Card.Content></Card.Root>
			{:else}
				<Card.Root><Card.Header><Card.Title class="text-base">Kunci / Salinan Kondisi / Revisi</Card.Title><Card.Description>Paket terkunci tidak bisa diedit langsung; gunakan salin/revisi agar riwayat perubahan tetap aman.</Card.Description></Card.Header><Card.Content class="space-y-3"><p class="text-sm">Status: {isLocked ? `Terkunci (${detail.package.locked_at})` : 'Belum terkunci'}</p>{#if detail.package.lock_reason}<p class="text-sm text-muted-foreground">Alasan: {detail.package.lock_reason}</p>{/if}<div class="flex gap-2"><LoadingButton onclick={lockPackage} loading={busy === 'lock'} disabled={isLocked}>Kunci + Simpan Salinan Kondisi</LoadingButton><LoadingButton variant="outline" onclick={clonePackage} loading={busy === 'clone'}>Buat Revisi/Salinan</LoadingButton></div></Card.Content></Card.Root>
			{/if}
		{/if}
		{/snippet}
	</AsyncContent>
</div>
