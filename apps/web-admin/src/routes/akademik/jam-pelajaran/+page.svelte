<script lang="ts">
	import { onMount } from 'svelte';
	import { toast } from '$lib/components/ui/sonner';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { readClientApiData } from '$lib/client/api';

	type LessonPeriod = {
		id: string;
		academic_year_id: string;
		academic_year_name: string;
		day_of_week: number;
		period_number: number;
		start_time: string;
		end_time: string;
		activity_type: string;
		label: string;
		is_counted_as_lesson: boolean;
	};

	type LessonPeriodPayload = {
		active_academic_year_id: string;
		active_academic_year_name: string;
		items: LessonPeriod[];
	};

	const dayLabels: Record<number, string> = { 1: 'Senin', 2: 'Selasa', 3: 'Rabu', 4: 'Kamis', 5: 'Jumat', 6: 'Sabtu' };
	const activityTypes = ['pelajaran', 'istirahat', 'upacara', 'pembiasaan', 'kokurikuler', 'ekstrakurikuler', 'lainnya'];

	let payload = $state<LessonPeriodPayload | null>(null);
	let loading = $state(true);
	let saving = $state(false);
	let editingId = $state('');
	let form = $state({ day_of_week: 1, period_number: 1, start_time: '07:30', end_time: '08:10', activity_type: 'pelajaran', label: '', is_counted_as_lesson: true });

	const items = $derived(payload?.items ?? []);
	const groupedItems = $derived(Object.entries(dayLabels).map(([day, label]) => ({ day: Number(day), label, items: items.filter((item) => item.day_of_week === Number(day)) })));

	function normalizeTime(value: string) {
		return value?.slice(0, 5) ?? '';
	}

	function resetForm() {
		editingId = '';
		form = { day_of_week: 1, period_number: 1, start_time: '07:30', end_time: '08:10', activity_type: 'pelajaran', label: '', is_counted_as_lesson: true };
	}

	async function load() {
		loading = true;
		try {
			const response = await fetch('/api/academic/lesson-periods');
			payload = await readClientApiData<LessonPeriodPayload>(response, 'Gagal memuat jam pelajaran');
		} finally {
			loading = false;
		}
	}

	function edit(item: LessonPeriod) {
		editingId = item.id;
		form = {
			day_of_week: item.day_of_week,
			period_number: item.period_number,
			start_time: normalizeTime(item.start_time),
			end_time: normalizeTime(item.end_time),
			activity_type: item.activity_type,
			label: item.label,
			is_counted_as_lesson: item.is_counted_as_lesson,
		};
	}

	async function save() {
		if (!payload?.active_academic_year_id) return toast.error('Tahun ajaran aktif belum tersedia');
		saving = true;
		try {
			const body = { ...form, academic_year_id: payload.active_academic_year_id };
			const response = await fetch(editingId ? `/api/academic/lesson-periods/${editingId}` : '/api/academic/lesson-periods', {
				method: editingId ? 'PUT' : 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(body),
			});
			await readClientApiData(response, 'Gagal menyimpan jam pelajaran');
			toast.success('Jam pelajaran tersimpan');
			resetForm();
			await load();
		} finally {
			saving = false;
		}
	}

	async function remove(item: LessonPeriod) {
		if (!confirm(`Hapus jam pelajaran ${dayLabels[item.day_of_week]} ke-${item.period_number}?`)) return;
		const response = await fetch(`/api/academic/lesson-periods/${item.id}`, { method: 'DELETE' });
		if (!response.ok) await readClientApiData(response, 'Gagal menghapus jam pelajaran');
		toast.success('Jam pelajaran dihapus');
		await load();
	}

	onMount(load);
</script>

<svelte:head><title>Jam Pelajaran — Akademik</title></svelte:head>

<div class="space-y-6 p-4 md:p-6">
	<div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
		<div>
			<p class="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">Akademik</p>
			<h1 class="text-2xl font-semibold text-foreground">Jam Pelajaran</h1>
			<p class="text-sm text-muted-foreground">Atur pola jam belajar per hari untuk tahun ajaran aktif.</p>
		</div>
		<Badge variant="outline">{payload?.active_academic_year_name || 'Tahun ajaran aktif'}</Badge>
	</div>

	<Card.Root>
		<Card.Header>
			<Card.Title class="text-base">{editingId ? 'Ubah Jam Pelajaran' : 'Tambah Jam Pelajaran'}</Card.Title>
			<Card.Description>Gunakan istilah kegiatan madrasah seperti pelajaran, istirahat, upacara, atau pembiasaan.</Card.Description>
		</Card.Header>
		<Card.Content>
			<div class="grid gap-3 md:grid-cols-8">
				<label class="space-y-1 text-xs font-medium text-muted-foreground">Hari
					<select bind:value={form.day_of_week} class="h-10 rounded-md border border-input bg-background px-3 text-sm text-foreground md:w-full">
						{#each Object.entries(dayLabels) as [day, label]}<option value={Number(day)}>{label}</option>{/each}
					</select>
				</label>
				<label class="space-y-1 text-xs font-medium text-muted-foreground">Ke-
					<Input type="number" min="1" bind:value={form.period_number} />
				</label>
				<label class="space-y-1 text-xs font-medium text-muted-foreground">Mulai
					<Input type="time" bind:value={form.start_time} />
				</label>
				<label class="space-y-1 text-xs font-medium text-muted-foreground">Selesai
					<Input type="time" bind:value={form.end_time} />
				</label>
				<label class="space-y-1 text-xs font-medium text-muted-foreground">Kegiatan
					<select bind:value={form.activity_type} class="h-10 rounded-md border border-input bg-background px-3 text-sm text-foreground md:w-full">
						{#each activityTypes as type}<option value={type}>{type}</option>{/each}
					</select>
				</label>
				<label class="space-y-1 text-xs font-medium text-muted-foreground md:col-span-2">Label
					<Input bind:value={form.label} placeholder="Contoh: JP 1" />
				</label>
				<label class="flex items-center gap-2 pt-6 text-sm text-muted-foreground"><input type="checkbox" bind:checked={form.is_counted_as_lesson} /> Dihitung JP</label>
			</div>
			<div class="mt-4 flex gap-2">
				<Button onclick={save} disabled={saving}>{saving ? 'Menyimpan...' : 'Simpan'}</Button>
				{#if editingId}<Button variant="outline" onclick={resetForm}>Batal</Button>{/if}
			</div>
		</Card.Content>
	</Card.Root>

	{#if loading}
		<p class="text-sm text-muted-foreground">Memuat jam pelajaran...</p>
	{:else}
		{#each groupedItems as group}
			<Card.Root>
				<Card.Header><Card.Title class="text-base">{group.label}</Card.Title></Card.Header>
				<Card.Content>
					{#if group.items.length === 0}
						<p class="text-sm text-muted-foreground">Belum ada jam pelajaran.</p>
					{:else}
						<Table.Root>
							<Table.Header><Table.Row><Table.Head>Ke-</Table.Head><Table.Head>Waktu</Table.Head><Table.Head>Kegiatan</Table.Head><Table.Head>Label</Table.Head><Table.Head>JP</Table.Head><Table.Head class="text-right">Aksi</Table.Head></Table.Row></Table.Header>
							<Table.Body>{#each group.items as item}<Table.Row><Table.Cell>{item.period_number}</Table.Cell><Table.Cell>{normalizeTime(item.start_time)}–{normalizeTime(item.end_time)}</Table.Cell><Table.Cell>{item.activity_type}</Table.Cell><Table.Cell>{item.label || '-'}</Table.Cell><Table.Cell>{item.is_counted_as_lesson ? 'Ya' : 'Tidak'}</Table.Cell><Table.Cell class="text-right"><Button size="sm" variant="outline" onclick={() => edit(item)}>Ubah</Button> <Button size="sm" variant="ghost" onclick={() => remove(item)}>Hapus</Button></Table.Cell></Table.Row>{/each}</Table.Body>
						</Table.Root>
					{/if}
				</Card.Content>
			</Card.Root>
		{/each}
	{/if}
</div>
