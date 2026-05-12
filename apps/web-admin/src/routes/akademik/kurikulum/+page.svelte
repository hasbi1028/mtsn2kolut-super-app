<script lang="ts">
	import { onMount } from 'svelte';
	import { RefreshCw } from '@lucide/svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { clientApiPathWithQuery, readClientApiData } from '$lib/client/api';

	type CurriculumProfile = {
		id: string;
		code: string;
		name: string;
		regulation_reference: string;
		education_level: string;
		status: string;
		notes: string;
	};

	type CurriculumAllocation = {
		id: string;
		subject_code: string;
		subject_name: string;
		level: string;
		subject_group: string;
		intra_annual_hours: number;
		koku_annual_hours: number;
		total_annual_hours: number;
		intra_weekly_hours: number;
		koku_weekly_hours: number;
		total_weekly_hours: number;
		lesson_minutes: number;
		display_order: number;
		counts_for_schedule: boolean;
		counts_for_report: boolean;
		counts_for_assessment: boolean;
		counts_for_ranking: boolean;
		notes: string;
	};

	type CurriculumSummary = {
		level: string;
		subject_count: number;
		intra_annual_hours: number;
		koku_annual_hours: number;
		total_annual_hours: number;
		intra_weekly_hours: number;
		koku_weekly_hours: number;
		total_weekly_hours: number;
		compliance_status: string;
		status_label: string;
	};

	type ClassAssignment = {
		id: string;
		class_name: string;
		class_level: string;
		curriculum_profile_name: string;
		is_active: boolean;
	};

	type CurriculumOverview = {
		profiles: CurriculumProfile[];
		active_profile: CurriculumProfile | null;
		allocations: CurriculumAllocation[];
		summary_by_level: CurriculumSummary[];
		class_assignments: ClassAssignment[];
	};

	const levels = ['VII', 'VIII', 'IX'];
	let selectedLevel = $state('VII');
	let selectedProfileId = $state('');
	let curriculumPromise = $state<Promise<CurriculumOverview> | null>(null);
	let refreshBusy = $state(false);


	async function fetchCurriculum() {
		const params = new URLSearchParams();
		if (selectedProfileId) params.set('profile_id', selectedProfileId);
		if (selectedLevel) params.set('level', selectedLevel);
		return readClientApiData<CurriculumOverview>(
			await fetch(clientApiPathWithQuery('/api/academic/curriculum', params))
		);
	}

	function load() {
		curriculumPromise = fetchCurriculum();
		return curriculumPromise;
	}

	async function refresh() {
		refreshBusy = true;
		try {
			await load();
		} finally {
			refreshBusy = false;
		}
	}

	function changeLevel(level: string) {
		selectedLevel = level;
		load();
	}

	function changeProfile(event: Event) {
		selectedProfileId = (event.currentTarget as HTMLSelectElement).value;
		load();
	}

	function groupLabel(value: string) {
		return {
			wajib: 'Wajib',
			pilihan: 'Pilihan',
			muatan_lokal: 'Muatan Lokal',
			layanan: 'Layanan',
			kokurikuler: 'Kokurikuler',
			kegiatan: 'Kegiatan'
		}[value] ?? value;
	}

	function yesNo(value: boolean) {
		return value ? 'Ya' : 'Tidak';
	}

	onMount(load);
</script>

<svelte:head>
	<title>Struktur Kurikulum | MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-6 p-6">
	<div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
		<div class="space-y-2">
			<p class="text-sm font-medium uppercase tracking-[0.18em] text-primary">Akademik</p>
			<h1 class="text-2xl font-semibold tracking-normal text-foreground">Struktur Kurikulum</h1>
			<p class="max-w-3xl text-sm text-muted-foreground">
				Pantau struktur Kurikulum Merdeka MTs berdasarkan KMA 1503 Tahun 2025. Halaman ini untuk pemeriksaan awal sebelum pengaturan lanjutan guru mapel, jadwal, nilai, dan rapor.
			</p>
		</div>
		<Button variant="outline" onclick={refresh} disabled={refreshBusy}>
			<RefreshCw class="size-4 {refreshBusy ? 'animate-spin' : ''}" />
			Muat ulang
		</Button>
	</div>

	<AsyncContent promise={curriculumPromise}>
		{#snippet pending()}
			<div class="grid gap-4 md:grid-cols-4">
				{#each Array.from({ length: 4 }) as _, index (`kurikulum-skeleton-${index}`)}
					<Card.Root>
						<Card.Header>
							<div class="h-4 w-24 rounded bg-muted"></div>
							<div class="h-8 w-32 rounded bg-muted"></div>
						</Card.Header>
					</Card.Root>
				{/each}
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Struktur kurikulum belum tersaji"
				message={error instanceof Error ? error.message : 'Data struktur kurikulum belum dapat dimuat.'}
				onRetry={() => { reset?.(); load(); }}
			/>
		{/snippet}

		{#snippet children(value)}
			{@const data = value as CurriculumOverview}
		{#if !data.active_profile}
			<EmptyStatePanel
				title="Struktur kurikulum belum tersedia"
				description="Profil kurikulum belum terisi. Jalankan penyesuaian struktur data terlebih dahulu."
			/>
		{:else}
			<div class="grid gap-4 md:grid-cols-4">
				<Card.Root class="md:col-span-2">
					<Card.Header>
						<Card.Description>Profil Kurikulum</Card.Description>
						<Card.Title class="text-lg">{data.active_profile.name}</Card.Title>
					</Card.Header>
					<Card.Content class="space-y-3 text-sm text-muted-foreground">
						<p>{data.active_profile.regulation_reference} · {data.active_profile.education_level}</p>
						{#if data.active_profile.notes}
							<p>{data.active_profile.notes}</p>
						{/if}
						<label class="block space-y-1">
							<span class="text-xs font-medium text-foreground">Pilih profil</span>
							<select class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" onchange={changeProfile}>
								<option value="">Profil aktif</option>
								{#each data.profiles as profile}
									<option value={profile.id} selected={profile.id === selectedProfileId}>{profile.name}</option>
								{/each}
							</select>
						</label>
					</Card.Content>
				</Card.Root>

				{#each data.summary_by_level as summary}
					<Card.Root class={summary.level === selectedLevel ? 'border-primary/50' : ''}>
						<Card.Header class="pb-2">
							<Card.Description>Kelas {summary.level}</Card.Description>
							<Card.Title class="text-2xl">{summary.total_weekly_hours} JP</Card.Title>
						</Card.Header>
						<Card.Content class="space-y-2 text-sm">
							<Badge variant={summary.compliance_status === 'sesuai' ? 'default' : 'destructive'}>{summary.status_label}</Badge>
							<p class="text-muted-foreground">{summary.intra_weekly_hours} intra + {summary.koku_weekly_hours} koku per minggu</p>
						</Card.Content>
					</Card.Root>
				{/each}
			</div>

			<div class="flex flex-wrap gap-2">
				{#each levels as level}
					<Button variant={selectedLevel === level ? 'default' : 'outline'} onclick={() => changeLevel(level)}>Kelas {level}</Button>
				{/each}
			</div>

			<section class="overflow-hidden rounded-lg border border-border bg-card shadow-sm">
				<div class="flex flex-col gap-2 border-b border-border p-4 md:flex-row md:items-center md:justify-between">
					<div>
						<h2 class="font-semibold">Alokasi JP Kelas {selectedLevel}</h2>
						<p class="text-sm text-muted-foreground">Alokasi wajib per mapel untuk pemeriksaan manual madrasah.</p>
					</div>
					<Badge variant="outline">{data.allocations.length} mapel</Badge>
				</div>
				<div class="overflow-x-auto">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head class="w-16">Urut</Table.Head>
								<Table.Head>Mapel</Table.Head>
								<Table.Head>Kelompok</Table.Head>
								<Table.Head class="text-right">Intra/tahun</Table.Head>
								<Table.Head class="text-right">Koku/tahun</Table.Head>
								<Table.Head class="text-right">Total/tahun</Table.Head>
								<Table.Head class="text-right">Total/minggu</Table.Head>
								<Table.Head>Jadwal</Table.Head>
								<Table.Head>Rapor</Table.Head>
								<Table.Head>Peringkat</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each data.allocations as row}
								<Table.Row>
									<Table.Cell>{row.display_order}</Table.Cell>
									<Table.Cell>
										<div class="font-medium">{row.subject_name}</div>
										<div class="text-xs text-muted-foreground">{row.subject_code}</div>
										{#if row.notes}<p class="mt-1 text-xs text-amber-600">{row.notes}</p>{/if}
									</Table.Cell>
									<Table.Cell>{groupLabel(row.subject_group)}</Table.Cell>
									<Table.Cell class="text-right">{row.intra_annual_hours}</Table.Cell>
									<Table.Cell class="text-right">{row.koku_annual_hours}</Table.Cell>
									<Table.Cell class="text-right font-medium">{row.total_annual_hours}</Table.Cell>
									<Table.Cell class="text-right font-medium">{row.total_weekly_hours}</Table.Cell>
									<Table.Cell>{yesNo(row.counts_for_schedule)}</Table.Cell>
									<Table.Cell>{yesNo(row.counts_for_report)}</Table.Cell>
									<Table.Cell>{yesNo(row.counts_for_ranking)}</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</div>
			</section>

			<Card.Root>
				<Card.Header>
					<Card.Title class="text-base">Rombel memakai profil ini</Card.Title>
					<Card.Description>Daftar rombel aktif yang sudah dipasang ke profil kurikulum.</Card.Description>
				</Card.Header>
				<Card.Content>
					<div class="flex flex-wrap gap-2">
						{#each data.class_assignments as item}
							<Badge variant="outline">{item.class_name} · {item.class_level}</Badge>
						{/each}
						{#if data.class_assignments.length === 0}
							<p class="text-sm text-muted-foreground">Belum ada rombel yang memakai profil ini.</p>
						{/if}
					</div>
				</Card.Content>
			</Card.Root>
		{/if}
		{/snippet}
	</AsyncContent>
</div>
