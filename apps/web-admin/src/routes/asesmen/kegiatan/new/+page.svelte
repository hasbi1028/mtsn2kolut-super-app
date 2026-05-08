<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { toast } from '$lib/components/ui/sonner';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { readClientApiData, readClientJson } from '$lib/client/api';

	type AcademicYear = { id: string; name: string; is_active: boolean };
	type AcademicPayload = { years?: AcademicYear[] };
	type FormData = { years: AcademicYear[] };

	const gradeOptions = ['VII', 'VIII', 'IX'];
	const typeLabel: Record<string, string> = {
		ulangan: 'Ulangan',
		uts: 'UTS',
		uas: 'UAS',
		uam: 'UAM',
		tryout: 'Try Out',
		lainnya: 'Lainnya'
	};
	const scopeLabel: Record<string, string> = { class: 'Per Kelas', grade: 'Per Tingkat', school: 'Seluruh Sekolah' };
	const statusLabel: Record<string, string> = { draft: 'Draft', active: 'Aktif', finished: 'Selesai' };

	let years = $state<AcademicYear[]>([]);
	let formPromise = $state<Promise<FormData> | null>(null);
	let fTitle = $state('');
	let fType = $state('uts');
	let fScope = $state('grade');
	let fYearId = $state('');
	let fStatus = $state('draft');
	let fTargetLevels = $state<string[]>([]);
	let fBusy = $state(false);

	let canSubmit = $derived(!fBusy && fTitle.trim() !== '' && fType !== '' && fYearId !== '');

	function isRecord(value: unknown): value is Record<string, unknown> {
		return typeof value === 'object' && value !== null;
	}

	function parseYears(payload: unknown): AcademicYear[] {
		return isRecord(payload) && Array.isArray((payload as AcademicPayload).years) ? (payload as AcademicPayload).years ?? [] : [];
	}

	async function fetchFormData(): Promise<FormData> {
		const academicData = await fetch('/api/academic').then((response) => readClientApiData<unknown>(response, 'Gagal memuat tahun ajaran'));
		return { years: parseYears(academicData) };
	}

	function applyFormData(data: FormData) {
		years = data.years;
		if (!fYearId && years.length > 0) fYearId = years.find((year) => year.is_active)?.id ?? years[0].id;
	}

	function loadForm() {
		formPromise = fetchFormData().then((data) => {
			applyFormData(data);
			return data;
		});
	}

	function retryForm(reset?: () => void) {
		reset?.();
		loadForm();
	}

	function errorMessage(error: unknown, fallback: string) {
		if (error instanceof Error && error.message.trim() && !error.message.toLowerCase().includes('fetch')) return error.message;
		return fallback;
	}

	function handleRenderError(error: unknown) {
		console.error('CBT event create render failed', error);
	}

	function toggleTargetLevel(level: string, checked: boolean) {
		if (checked) {
			fTargetLevels = Array.from(new Set([...fTargetLevels, level])).sort();
			return;
		}
		fTargetLevels = fTargetLevels.filter((item) => item !== level);
	}

	async function createEvent() {
		if (!canSubmit) return;
		fBusy = true;
		try {
			const res = await fetch('/api/asesmen/events', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					title: fTitle,
					exam_type: fType,
					scope: fScope,
					target_levels: fTargetLevels,
					academic_year_id: fYearId,
					status: fStatus
				})
			});
			await readClientJson<unknown>(res);
			toast.success('Kegiatan berhasil dibuat');
			await goto(resolve('/asesmen/kegiatan'));
		} catch (error) {
			toast.error(errorMessage(error, 'Gagal membuat kegiatan ujian. Periksa koneksi lalu coba lagi.'));
		} finally {
			fBusy = false;
		}
	}

	onMount(loadForm);
</script>

<svelte:head><title>Buat Kegiatan CBT — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
	<section class="rounded-3xl border border-primary/20 bg-gradient-to-br from-primary/10 via-card to-primary/10 p-6 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
			<div class="max-w-3xl space-y-2">
				<p class="text-xs font-semibold uppercase tracking-[0.2em] text-primary">Create Flow CBT</p>
				<h1 class="text-3xl font-semibold tracking-tight text-foreground">Buat Kegiatan CBT</h1>
				<p class="text-sm leading-6 text-muted-foreground">Isi identitas kegiatan sekali, lalu lanjutkan ke paket, sesi, peserta, dan token dari daftar kegiatan.</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<Button href={resolve('/asesmen/persiapan')} variant="outline">Persiapan CBT</Button>
				<Button href={resolve('/asesmen/kegiatan')} variant="outline">Batal</Button>
			</div>
		</div>
	</section>

	<AsyncContent promise={formPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<Card.Root class="border-border shadow-sm">
				<Card.Content class="grid gap-3 p-6 sm:grid-cols-2">
					<Skeleton class="h-10 sm:col-span-2" />
					<Skeleton class="h-10" />
					<Skeleton class="h-10" />
					<Skeleton class="h-24 sm:col-span-2" />
				</Card.Content>
			</Card.Root>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel title="Form Kegiatan Belum Siap" message={errorMessage(error, 'Gagal memuat tahun ajaran')} onRetry={() => retryForm(reset)} />
		{/snippet}

		{#snippet children()}
			<Card.Root class="border-border shadow-sm">
				<Card.Header>
					<Card.Title class="text-base">Identitas Kegiatan</Card.Title>
					<Card.Description>Payload mengikuti endpoint <code>/api/asesmen/events</code> yang dipakai form lama.</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-4">
					<div class="grid gap-3 sm:grid-cols-2">
						<div class="sm:col-span-2">
							<label for="event-title" class="mb-1 block text-xs text-muted-foreground">Judul Kegiatan <span class="text-destructive">*</span></label>
							<Input id="event-title" placeholder="mis: UTS Semester Ganjil 2025/2026" bind:value={fTitle} />
						</div>
						<div>
							<label for="event-year" class="mb-1 block text-xs text-muted-foreground">Tahun Ajaran <span class="text-destructive">*</span></label>
							<select id="event-year" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fYearId}>
								<option value="">-- Pilih --</option>
								{#each years as year (year.id)}
									<option value={year.id}>{year.name} {year.is_active ? '(Aktif)' : ''}</option>
								{/each}
							</select>
						</div>
						<div>
							<label for="event-type" class="mb-1 block text-xs text-muted-foreground">Jenis Ujian</label>
							<select id="event-type" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fType}>
								{#each Object.entries(typeLabel) as [value, label] (value)}
									<option value={value}>{label}</option>
								{/each}
							</select>
						</div>
						<div>
							<label for="event-scope" class="mb-1 block text-xs text-muted-foreground">Cakupan</label>
							<select id="event-scope" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fScope}>
								{#each Object.entries(scopeLabel) as [value, label] (value)}
									<option value={value}>{label}</option>
								{/each}
							</select>
						</div>
						<div>
							<label for="event-status" class="mb-1 block text-xs text-muted-foreground">Status Awal</label>
							<select id="event-status" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fStatus}>
								{#each Object.entries(statusLabel) as [value, label] (value)}
									<option value={value}>{label}</option>
								{/each}
							</select>
						</div>
						<fieldset class="sm:col-span-2">
							<legend class="mb-2 block text-xs text-muted-foreground">Tingkat yang diikutkan</legend>
							<div class="grid gap-2 sm:grid-cols-3">
								{#each gradeOptions as level (level)}
									<label class="flex items-center gap-2 rounded-md border border-input px-3 py-2 text-sm text-foreground">
										<input type="checkbox" checked={fTargetLevels.includes(level)} onchange={(event) => toggleTargetLevel(level, event.currentTarget.checked)} class="size-4 accent-primary" />
										<span>Tingkat {level}</span>
									</label>
								{/each}
							</div>
							<p class="mt-2 text-xs text-muted-foreground">Kosong berarti mengikuti cakupan biasa. Isi untuk kegiatan yang hanya berlaku bagi tingkat tertentu.</p>
						</fieldset>
					</div>

					<div class="flex flex-wrap gap-2">
						<LoadingButton disabled={!canSubmit} onclick={() => void createEvent()} loading={fBusy} loadingLabel="Menyimpan...">Simpan Kegiatan</LoadingButton>
						<Button href={resolve('/asesmen/kegiatan')} variant="outline">Batal</Button>
					</div>
				</Card.Content>
			</Card.Root>
		{/snippet}
	</AsyncContent>
</div>
