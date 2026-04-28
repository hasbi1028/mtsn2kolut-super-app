<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Badge } from '$lib/components/ui/badge';

	type Question = {
		id: string; subject_id: string; subject_name: string; subject_code: string;
		code: string; question_text: string;
		option_a: string; option_b: string; option_c: string; option_d: string; option_e: string;
		answer_key: string; explanation: string;
		difficulty: string; status: string; created_at: string;
	};
	type Subject = { id: string; name: string; code: string; };

	let questions = $state<Question[]>([]);
	let subjects = $state<Subject[]>([]);
	let loading = $state(true);
	let error = $state('');
	let toast = $state('');
	let search = $state('');
	let filterSubject = $state('');
	let filterStatus = $state('');
	let showForm = $state(false);
	let editId = $state<string | null>(null);

	let fSubjectId = $state('');
	let fCode = $state('');
	let fText = $state('');
	let fA = $state('');
	let fB = $state('');
	let fC = $state('');
	let fD = $state('');
	let fE = $state('');
	let fAnswer = $state('A');
	let fExplanation = $state('');
	let fDifficulty = $state('medium');
	let fStatus = $state('draft');
	let fBusy = $state(false);

	const difficultyLabel: Record<string, string> = { easy: 'Mudah', medium: 'Sedang', hard: 'Sulit' };
	const statusLabel: Record<string, string> = { draft: 'Draft', published: 'Aktif', archived: 'Arsip' };

	function diffBadgeClass(d: string) {
		if (d === 'easy') return 'bg-emerald-100 text-emerald-700 border-emerald-200';
		if (d === 'hard') return 'bg-red-100 text-red-700 border-red-200';
		return 'bg-amber-100 text-amber-700 border-amber-200';
	}

	function statusBadgeClass(s: string) {
		if (s === 'published') return 'bg-green-100 text-green-800 border-green-200';
		if (s === 'archived') return 'bg-slate-100 text-slate-500 border-slate-200';
		return 'bg-slate-100 text-slate-600 border-slate-200';
	}

	let filtered = $derived(questions.filter(q => {
		if (filterSubject && q.subject_id !== filterSubject) return false;
		if (filterStatus && q.status !== filterStatus) return false;
		if (search && !q.question_text.toLowerCase().includes(search.toLowerCase()) && !q.code.toLowerCase().includes(search.toLowerCase())) return false;
		return true;
	}));

	async function load() {
		try {
			const [qRes, aRes] = await Promise.all([
				fetch('/api/cbt/questions'),
				fetch('/api/academic'),
			]);
			const qJson = await qRes.json();
			const aJson = await aRes.json();
			if (qJson.error) { error = qJson.error; return; }
			questions = qJson.data ?? qJson ?? [];
			subjects = (aJson.data ?? aJson)?.subjects ?? [];
		} catch {
			error = 'Gagal memuat data soal';
		} finally {
			loading = false;
		}
	}

	function showToast(msg: string) {
		toast = msg;
		setTimeout(() => (toast = ''), 3000);
	}

	function resetForm() {
		fSubjectId = ''; fCode = ''; fText = ''; fA = ''; fB = ''; fC = ''; fD = ''; fE = '';
		fAnswer = 'A'; fExplanation = ''; fDifficulty = 'medium'; fStatus = 'draft';
		editId = null;
		showForm = false;
	}

	function openEdit(q: Question) {
		fSubjectId = q.subject_id;
		fCode = q.code;
		fText = q.question_text;
		fA = q.option_a;
		fB = q.option_b;
		fC = q.option_c;
		fD = q.option_d;
		fE = q.option_e;
		fAnswer = q.answer_key;
		fExplanation = q.explanation;
		fDifficulty = q.difficulty;
		fStatus = q.status;
		editId = q.id;
		showForm = true;
		window.scrollTo({ top: 0, behavior: 'smooth' });
	}

	async function saveQuestion() {
		if (!fSubjectId || !fText || !fA || !fB || !fC || !fD || !fAnswer) return;
		fBusy = true;
		try {
			const method = editId ? 'PUT' : 'POST';
			const path = editId ? `/api/cbt/questions/${editId}` : '/api/cbt/questions';
			const res = await fetch(path, {
				method,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					subject_id: fSubjectId, code: fCode,
					question_text: fText, option_a: fA, option_b: fB, option_c: fC, option_d: fD, option_e: fE,
					answer_key: fAnswer, explanation: fExplanation, difficulty: fDifficulty, status: fStatus,
				}),
			});
			if (!res.ok) { const j = await res.json(); showToast(j.error ?? 'Gagal'); return; }
			showToast(editId ? 'Soal diperbarui' : 'Soal berhasil ditambahkan');
			resetForm();
			await load();
		} finally { fBusy = false; }
	}

	async function deleteQuestion(id: string) {
		if (!confirm('Hapus soal ini?')) return;
		await fetch(`/api/cbt/questions?id=${id}`, { method: 'DELETE' });
		showToast('Soal dihapus');
		await load();
	}

	onMount(load);
</script>

<svelte:head><title>Bank Soal CBT — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-slate-800">Bank Soal CBT</h1>
			<p class="text-sm text-slate-500 mt-1">Kelola soal pilihan ganda untuk ujian berbasis komputer</p>
		</div>
		<Button onclick={() => { if (showForm) resetForm(); else showForm = true; }}>
			{showForm ? 'Batal' : '+ Tambah Soal'}
		</Button>
	</div>

	{#if toast}
		<div class="rounded-md bg-emerald-50 border border-emerald-200 px-4 py-3 text-sm text-emerald-800">{toast}</div>
	{/if}
	{#if error}
		<div class="rounded-md bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-800">{error}</div>
	{/if}

	{#if showForm}
		<Card.Root>
			<Card.Header class="pb-2">
				<Card.Title class="text-base">{editId ? 'Edit Soal' : 'Tambah Soal Baru'}</Card.Title>
			</Card.Header>
			<Card.Content class="space-y-4">
				<div class="grid gap-3 sm:grid-cols-3">
					<div>
						<label for="q-subject" class="text-xs text-slate-500 mb-1 block">Mata Pelajaran <span class="text-red-500">*</span></label>
						<select id="q-subject" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fSubjectId}>
							<option value="">-- Pilih --</option>
							{#each subjects as s}
								<option value={s.id}>{s.code} — {s.name}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="q-code" class="text-xs text-slate-500 mb-1 block">Kode Soal</label>
						<Input id="q-code" placeholder="mis: MTK-001" bind:value={fCode} />
					</div>
					<div>
						<label for="q-diff" class="text-xs text-slate-500 mb-1 block">Tingkat Kesulitan</label>
						<select id="q-diff" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fDifficulty}>
							<option value="easy">Mudah</option>
							<option value="medium">Sedang</option>
							<option value="hard">Sulit</option>
						</select>
					</div>
				</div>

				<div>
					<label for="q-text" class="text-xs text-slate-500 mb-1 block">Teks Soal <span class="text-red-500">*</span></label>
					<Textarea id="q-text" placeholder="Tulis pertanyaan di sini..." rows={3} bind:value={fText} />
				</div>

				<div class="grid gap-3 sm:grid-cols-2">
					<div>
						<label for="q-oa" class="text-xs text-slate-500 mb-1 block">Opsi A <span class="text-red-500">*</span></label>
						<Input id="q-oa" placeholder="Jawaban opsi A" bind:value={fA} />
					</div>
					<div>
						<label for="q-ob" class="text-xs text-slate-500 mb-1 block">Opsi B <span class="text-red-500">*</span></label>
						<Input id="q-ob" placeholder="Jawaban opsi B" bind:value={fB} />
					</div>
					<div>
						<label for="q-oc" class="text-xs text-slate-500 mb-1 block">Opsi C <span class="text-red-500">*</span></label>
						<Input id="q-oc" placeholder="Jawaban opsi C" bind:value={fC} />
					</div>
					<div>
						<label for="q-od" class="text-xs text-slate-500 mb-1 block">Opsi D <span class="text-red-500">*</span></label>
						<Input id="q-od" placeholder="Jawaban opsi D" bind:value={fD} />
					</div>
					<div>
						<label for="q-oe" class="text-xs text-slate-500 mb-1 block">Opsi E (opsional)</label>
						<Input id="q-oe" placeholder="Jawaban opsi E" bind:value={fE} />
					</div>
				</div>

				<div class="grid gap-3 sm:grid-cols-2">
					<div>
						<label for="q-ans" class="text-xs text-slate-500 mb-1 block">Kunci Jawaban <span class="text-red-500">*</span></label>
						<select id="q-ans" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fAnswer}>
							<option value="A">A</option>
							<option value="B">B</option>
							<option value="C">C</option>
							<option value="D">D</option>
							<option value="E">E</option>
						</select>
					</div>
					<div>
						<label for="q-stat" class="text-xs text-slate-500 mb-1 block">Status</label>
						<select id="q-stat" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fStatus}>
							<option value="draft">Draft</option>
							<option value="published">Aktif/Published</option>
							<option value="archived">Arsip</option>
						</select>
					</div>
				</div>

				<div>
					<label for="q-exp" class="text-xs text-slate-500 mb-1 block">Penjelasan (opsional)</label>
					<Textarea id="q-exp" placeholder="Pembahasan jawaban..." rows={2} bind:value={fExplanation} />
				</div>

				<div class="flex gap-2">
					<Button disabled={fBusy || !fSubjectId || !fText || !fA || !fB || !fC || !fD} onclick={saveQuestion}>
						{fBusy ? 'Menyimpan...' : (editId ? 'Perbarui Soal' : 'Simpan Soal')}
					</Button>
					<Button variant="outline" onclick={resetForm}>Batal</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	{#if loading}
		<p class="text-sm text-slate-500">Memuat data...</p>
	{:else}
		<Card.Root>
			<Card.Header class="pb-3">
				<div class="flex flex-wrap items-center gap-3">
					<Card.Title class="text-base">Daftar Soal ({questions.length})</Card.Title>
					<Input placeholder="Cari soal..." bind:value={search} class="max-w-xs" />
					<select class="rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={filterSubject}>
						<option value="">Semua Mapel</option>
						{#each subjects as s}
							<option value={s.id}>{s.code}</option>
						{/each}
					</select>
					<select class="rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={filterStatus}>
						<option value="">Semua Status</option>
						<option value="draft">Draft</option>
						<option value="published">Aktif</option>
						<option value="archived">Arsip</option>
					</select>
				</div>
			</Card.Header>
			<Card.Content class="p-0 overflow-x-auto">
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head class="w-24">Kode</Table.Head>
							<Table.Head>Soal</Table.Head>
							<Table.Head>Mapel</Table.Head>
							<Table.Head>Kunci</Table.Head>
							<Table.Head>Kesulitan</Table.Head>
							<Table.Head>Status</Table.Head>
							<Table.Head class="text-right">Aksi</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each filtered as q}
							<Table.Row>
								<Table.Cell class="font-mono text-xs text-slate-500">{q.code || '—'}</Table.Cell>
								<Table.Cell class="max-w-xs">
									<p class="text-sm line-clamp-2">{q.question_text}</p>
								</Table.Cell>
								<Table.Cell>
									<Badge variant="outline" class="text-xs">{q.subject_code}</Badge>
								</Table.Cell>
								<Table.Cell>
									<Badge class="bg-green-100 text-green-800 border-green-200 font-mono">{q.answer_key}</Badge>
								</Table.Cell>
								<Table.Cell>
									<Badge class={diffBadgeClass(q.difficulty)}>{difficultyLabel[q.difficulty] ?? q.difficulty}</Badge>
								</Table.Cell>
								<Table.Cell>
									<Badge class={statusBadgeClass(q.status)}>{statusLabel[q.status] ?? q.status}</Badge>
								</Table.Cell>
								<Table.Cell class="text-right">
									<div class="flex gap-2 justify-end">
										<Button variant="outline" size="sm" onclick={() => openEdit(q)}>Edit</Button>
										<Button variant="destructive" size="sm" onclick={() => deleteQuestion(q.id)}>Hapus</Button>
									</div>
								</Table.Cell>
							</Table.Row>
						{:else}
							<Table.Row>
								<Table.Cell colspan={7} class="text-center text-slate-400 py-8">
									{search || filterSubject || filterStatus ? 'Tidak ada soal yang cocok' : 'Belum ada soal'}
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</Card.Content>
		</Card.Root>
	{/if}
</div>
