<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';

	type CbtEvent = {
		id: string; title: string; exam_type: string; scope: string;
		target_levels: string[];
		academic_year_id: string; academic_year_name: string;
		status: string; created_at: string; session_count: number;
	};
	type AcademicYear = { id: string; name: string; is_active: boolean; };

	let events = $state<CbtEvent[]>([]);
	let years = $state<AcademicYear[]>([]);
	let loading = $state(true);
	let error = $state('');
	let showForm = $state(false);
	let editId = $state<string | null>(null);

	let fTitle = $state('');
	let fType = $state('uts');
	let fScope = $state('grade');
	let fYearId = $state('');
	let fStatus = $state('draft');
	let fTargetLevels = $state<string[]>([]);
	let fBusy = $state(false);
	const gradeOptions = ['VII', 'VIII', 'IX'];

	const typeLabel: Record<string, string> = {
		ulangan: 'Ulangan', uts: 'UTS', uas: 'UAS', uam: 'UAM', tryout: 'Try Out', lainnya: 'Lainnya'
	};
	const scopeLabel: Record<string, string> = { class: 'Per Kelas', grade: 'Per Tingkat', school: 'Seluruh Sekolah' };
	const statusLabel: Record<string, string> = { draft: 'Draft', active: 'Aktif', finished: 'Selesai' };

	async function load() {
		try {
			const [eRes, aRes] = await Promise.all([
				fetch('/api/cbt/events'),
				fetch('/api/academic'),
			]);
			const eJson = await eRes.json();
			const aJson = await aRes.json();
			events = eJson.data ?? eJson ?? [];
			years = (aJson.data ?? aJson)?.years ?? [];
			if (years.length > 0 && !fYearId) {
				fYearId = years.find(y => y.is_active)?.id || years[0].id;
			}
		} catch {
			error = 'Gagal memuat data event';
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

	function resetForm() {
		fTitle = ''; fType = 'uts'; fScope = 'grade'; fStatus = 'draft';
		fTargetLevels = [];
		editId = null; showForm = false;
	}

	function openEdit(e: CbtEvent) {
		fTitle = e.title;
		fType = e.exam_type;
		fScope = e.scope;
		fYearId = e.academic_year_id;
		fStatus = e.status;
		fTargetLevels = [...(e.target_levels ?? [])];
		editId = e.id;
		showForm = true;
	}

	function toggleTargetLevel(level: string, checked: boolean) {
		if (checked) {
			fTargetLevels = Array.from(new Set([...fTargetLevels, level])).sort();
			return;
		}
		fTargetLevels = fTargetLevels.filter((item) => item !== level);
	}

	async function saveEvent() {
		if (!fTitle || !fType || !fYearId) return;
		fBusy = true;
		try {
			const method = editId ? 'PUT' : 'POST';
			const path = editId ? `/api/cbt/events/${editId}` : '/api/cbt/events';
			const res = await fetch(path, {
				method,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					title: fTitle, exam_type: fType, scope: fScope,
					target_levels: fTargetLevels,
					academic_year_id: fYearId, status: fStatus,
				}),
			});
			if (!res.ok) { const j = await res.json(); showError(j.error ?? 'Gagal'); return; }
			showToast(editId ? 'Event diperbarui' : 'Event berhasil dibuat');
			resetForm();
			await load();
		} finally { fBusy = false; }
	}

	async function deleteEvent(id: string) {
		if (!confirm('Hapus kegiatan ini? Sesi di dalamnya tidak akan terhapus tapi relasinya dilepas.')) return;
		await fetch(`/api/cbt/events/${id}`, { method: 'DELETE' });
		showToast('Event dihapus');
		await load();
	}

	onMount(load);
</script>

<svelte:head><title>Kegiatan Ujian (Events) — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-slate-800">Kegiatan Ujian (Events)</h1>
			<p class="text-sm text-slate-500 mt-1">Grup besar untuk sesi-sesi ujian (mis: UTS, UAS)</p>
		</div>
		<Button onclick={() => { if (showForm) resetForm(); else showForm = true; }}>
			{showForm ? 'Batal' : '+ Buat Kegiatan'}
		</Button>
	</div>

	{#if showForm}
		<Card.Root>
			<Card.Header class="pb-2">
				<Card.Title class="text-base">{editId ? 'Edit Kegiatan' : 'Tambah Kegiatan Baru'}</Card.Title>
			</Card.Header>
			<Card.Content class="space-y-4">
				<div class="grid gap-3 sm:grid-cols-2">
					<div class="sm:col-span-2">
						<label for="e-title" class="text-xs text-slate-500 mb-1 block">Judul Kegiatan <span class="text-red-500">*</span></label>
						<Input id="e-title" placeholder="mis: UTS Semester Ganjil 2025/2026" bind:value={fTitle} />
					</div>
					<div>
						<label for="e-year" class="text-xs text-slate-500 mb-1 block">Tahun Ajaran <span class="text-red-500">*</span></label>
						<select id="e-year" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fYearId}>
							{#each years as y}
								<option value={y.id}>{y.name} {y.is_active ? '(Aktif)' : ''}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="e-type" class="text-xs text-slate-500 mb-1 block">Jenis Ujian</label>
						<select id="e-type" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fType}>
							{#each Object.entries(typeLabel) as [val, label]}
								<option value={val}>{label}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="e-scope" class="text-xs text-slate-500 mb-1 block">Cakupan (Scope)</label>
						<select id="e-scope" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fScope}>
							{#each Object.entries(scopeLabel) as [val, label]}
								<option value={val}>{label}</option>
							{/each}
						</select>
					</div>
					<fieldset class="sm:col-span-2">
						<legend class="mb-2 block text-xs text-slate-500">Tingkat yang diikutkan</legend>
						<div class="grid gap-2 sm:grid-cols-3">
							{#each gradeOptions as level (level)}
								<label class="flex items-center gap-2 rounded-md border border-input px-3 py-2 text-sm text-slate-700">
									<input
										type="checkbox"
										checked={fTargetLevels.includes(level)}
										onchange={(event) => toggleTargetLevel(level, (event.currentTarget as HTMLInputElement).checked)}
										class="size-4 accent-emerald-700"
									/>
									<span>Tingkat {level}</span>
								</label>
							{/each}
						</div>
						<p class="mt-2 text-xs text-slate-500">
							Kosong berarti mengikuti scope biasa. Isi ini untuk kasus seperti UAS genap yang hanya berlaku bagi tingkat tertentu.
						</p>
					</fieldset>
					<div>
						<label for="e-status" class="text-xs text-slate-500 mb-1 block">Status</label>
						<select id="e-status" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fStatus}>
							{#each Object.entries(statusLabel) as [val, label]}
								<option value={val}>{label}</option>
							{/each}
						</select>
					</div>
				</div>

				<div class="flex gap-2">
					<Button disabled={fBusy || !fTitle || !fYearId} onclick={saveEvent}>
						{fBusy ? 'Menyimpan...' : (editId ? 'Perbarui' : 'Simpan Kegiatan')}
					</Button>
					<Button variant="outline" onclick={resetForm}>Batal</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	{#if loading}
		<p class="text-sm text-slate-500">Memuat data...</p>
	{:else}
		<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
			<Card.Content class="p-0 overflow-x-auto">
				<div class="hidden overflow-x-auto lg:block">
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Judul Kegiatan</Table.Head>
							<Table.Head>Tipe</Table.Head>
							<Table.Head>Scope</Table.Head>
							<Table.Head>Tingkat</Table.Head>
							<Table.Head class="text-center">Sesi</Table.Head>
							<Table.Head>Status</Table.Head>
							<Table.Head class="text-right">Aksi</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each events as e (e.id)}
							<Table.Row>
								<Table.Cell>
									<div class="font-medium text-slate-800">{e.title}</div>
									<div class="text-xs text-slate-500">{e.academic_year_name}</div>
								</Table.Cell>
								<Table.Cell>
									<Badge variant="outline" class="text-xs capitalize">{typeLabel[e.exam_type] ?? e.exam_type}</Badge>
								</Table.Cell>
								<Table.Cell class="text-sm text-slate-600">{scopeLabel[e.scope] ?? e.scope}</Table.Cell>
								<Table.Cell class="text-sm text-slate-600">{e.target_levels?.length ? e.target_levels.join(', ') : 'Semua sesuai scope'}</Table.Cell>
								<Table.Cell class="text-center">
									<Badge variant="secondary">{e.session_count} Sesi</Badge>
								</Table.Cell>
								<Table.Cell>
									<Badge class={e.status === 'active' ? 'bg-emerald-100 text-emerald-700 border-emerald-200' : 'bg-slate-100 text-slate-600'}>
										{statusLabel[e.status] ?? e.status}
									</Badge>
								</Table.Cell>
								<Table.Cell class="text-right">
									<div class="flex gap-2 justify-end">
										<Button variant="outline" size="sm" onclick={() => openEdit(e)}>Edit</Button>
										<Button variant="destructive" size="sm" onclick={() => deleteEvent(e.id)}>Hapus</Button>
									</div>
								</Table.Cell>
							</Table.Row>
						{:else}
							<Table.Row>
								<Table.Cell colspan={7} class="text-center text-slate-400 py-12">
									Belum ada kegiatan ujian.
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
				</div>

				<div class="grid gap-3 p-4 lg:hidden">
					{#each events as e (e.id)}
						<div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
							<div class="flex items-start justify-between gap-3">
								<div class="min-w-0">
									<p class="text-sm font-semibold text-slate-900">{e.title}</p>
									<p class="mt-1 text-xs text-slate-500">{e.academic_year_name}</p>
								</div>
								<Badge class={e.status === 'active' ? 'bg-emerald-100 text-emerald-700 border-emerald-200' : 'bg-slate-100 text-slate-600'}>
									{statusLabel[e.status] ?? e.status}
								</Badge>
							</div>
							<div class="mt-3 flex flex-wrap items-center gap-2">
								<Badge variant="outline" class="text-xs capitalize">{typeLabel[e.exam_type] ?? e.exam_type}</Badge>
								<Badge variant="outline" class="text-xs">{scopeLabel[e.scope] ?? e.scope}</Badge>
								{#if e.target_levels?.length}
									<Badge variant="outline" class="text-xs">{e.target_levels.join(', ')}</Badge>
								{/if}
								<Badge variant="secondary">{e.session_count} Sesi</Badge>
							</div>
							<div class="mt-4 grid grid-cols-2 gap-2">
								<Button variant="outline" size="sm" onclick={() => openEdit(e)}>Edit</Button>
								<Button variant="destructive" size="sm" onclick={() => deleteEvent(e.id)}>Hapus</Button>
							</div>
						</div>
					{:else}
						<div class="rounded-2xl border border-dashed border-slate-300 bg-slate-50 px-4 py-10 text-center text-sm text-slate-500">
							Belum ada kegiatan ujian.
						</div>
					{/each}
				</div>
			</Card.Content>
		</Card.Root>
	{/if}
</div>
