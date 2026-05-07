<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { RefreshCw } from '@lucide/svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { readClientApiData } from '$lib/client/api';

	type RombelSummary = {
		id: string;
		code: string;
		name: string;
		level: string;
		is_active: boolean;
		academic_year_name: string;
		homeroom_teacher_name: string;
		total_students: number;
		total_subject_teachers: number;
		total_subject_assignments: number;
		total_timetable_slots: number;
	};

	let rombels = $state<RombelSummary[]>([]);
	let rombelPromise = $state<Promise<RombelSummary[]> | null>(null);
	let refreshBusy = $state(false);
	let requestId = 0;

	const summary = $derived.by(() => ({
		activeClasses: rombels.filter((item) => item.is_active).length,
		totalStudents: rombels.reduce((sum, item) => sum + item.total_students, 0),
		missingHomeroom: rombels.filter((item) => !item.homeroom_teacher_name).length,
	}));

	async function fetchRombels(): Promise<RombelSummary[]> {
		return await fetch('/api/academic/rombel').then((response) =>
			readClientApiData<RombelSummary[]>(response, 'Gagal memuat data rombel')
		);
	}

	function loadRombels() {
		const current = ++requestId;
		rombelPromise = fetchRombels().then((items) => {
			if (current === requestId) {
				rombels = items;
			}
			return items;
		});
	}

	async function refreshRombels() {
		refreshBusy = true;
		try {
			loadRombels();
			await rombelPromise;
		} finally {
			refreshBusy = false;
		}
	}

	function rombelErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Data rombel belum dapat dimuat. Periksa koneksi backend lalu coba lagi.';
	}

	function handleRenderError(error: unknown) {
		console.error('Rombel list render failed', error);
	}

	onMount(loadRombels);
</script>

<svelte:head>
	<title>Rombel | MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
		<div>
			<p class="text-sm font-medium text-primary">Akademik</p>
			<h1 class="text-2xl font-semibold tracking-normal text-foreground">Rombel</h1>
			<p class="mt-1 max-w-2xl text-sm text-muted-foreground">
				Daftar kelas belajar beserta wali kelas, jumlah siswa, guru mapel, dan slot jadwal.
			</p>
		</div>
		<Button variant="outline" onclick={() => void refreshRombels()} disabled={refreshBusy} aria-label="Muat ulang rombel">
			<RefreshCw class={`mr-2 size-4 ${refreshBusy ? 'animate-spin' : ''}`} />
			Muat Ulang
		</Button>
	</div>

	<AsyncContent promise={rombelPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<div class="grid gap-3 md:grid-cols-3">
					<Skeleton class="h-24 w-full" />
					<Skeleton class="h-24 w-full" />
					<Skeleton class="h-24 w-full" />
				</div>
				<Skeleton class="h-80 w-full" />
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Rombel Belum Tersaji"
				message={rombelErrorMessage(error)}
				onRetry={() => {
					reset?.();
					loadRombels();
				}}
			/>
		{/snippet}

		<div class="grid gap-3 md:grid-cols-3">
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Description>Rombel Aktif</Card.Description>
					<Card.Title class="text-2xl">{summary.activeClasses}</Card.Title>
				</Card.Header>
			</Card.Root>
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Description>Total Siswa Aktif</Card.Description>
					<Card.Title class="text-2xl">{summary.totalStudents}</Card.Title>
				</Card.Header>
			</Card.Root>
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Description>Belum Ada Wali Kelas</Card.Description>
					<Card.Title class="text-2xl">{summary.missingHomeroom}</Card.Title>
				</Card.Header>
			</Card.Root>
		</div>

		<Card.Root>
			<Card.Header>
				<Card.Title class="text-base">Daftar Rombel</Card.Title>
				<Card.Description>Gunakan detail rombel untuk melihat siswa, orang tua, wali kelas, guru mapel, dan jadwal.</Card.Description>
			</Card.Header>
			<Card.Content class="p-0">
				<div class="overflow-x-auto">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>Kelas</Table.Head>
								<Table.Head>Tahun Ajaran</Table.Head>
								<Table.Head>Wali Kelas</Table.Head>
								<Table.Head class="text-right">Siswa</Table.Head>
								<Table.Head class="text-right">Guru Mapel</Table.Head>
								<Table.Head class="text-right">Jadwal</Table.Head>
								<Table.Head>Status</Table.Head>
								<Table.Head></Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each rombels as rombel (rombel.id)}
								<Table.Row>
									<Table.Cell>
										<div class="space-y-0.5">
											<p class="font-medium text-foreground">{rombel.name}</p>
											<p class="text-xs text-muted-foreground">{rombel.code} · Tingkat {rombel.level}</p>
										</div>
									</Table.Cell>
									<Table.Cell class="text-muted-foreground">{rombel.academic_year_name}</Table.Cell>
									<Table.Cell>
										{#if rombel.homeroom_teacher_name}
											{rombel.homeroom_teacher_name}
										{:else}
											<span class="text-muted-foreground">Belum ditetapkan</span>
										{/if}
									</Table.Cell>
									<Table.Cell class="text-right tabular-nums">{rombel.total_students}</Table.Cell>
									<Table.Cell class="text-right tabular-nums">{rombel.total_subject_teachers}</Table.Cell>
									<Table.Cell class="text-right tabular-nums">{rombel.total_timetable_slots}</Table.Cell>
									<Table.Cell>
										{#if rombel.is_active}
											<Badge class="border-primary/20 bg-primary/15 text-primary">Aktif</Badge>
										{:else}
											<Badge variant="secondary">Nonaktif</Badge>
										{/if}
									</Table.Cell>
									<Table.Cell class="text-right">
										<Button variant="outline" size="sm" href={resolve(`/akademik/rombel/${rombel.id}` as '/')}>
											Detail
										</Button>
									</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={8} class="p-4">
										<EmptyStatePanel
											compact
											title="Belum ada rombel"
											description="Rombel mengikuti data kelas akademik yang sudah dibuat di Data Akademik."
										/>
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</div>
			</Card.Content>
		</Card.Root>
	</AsyncContent>
</div>
