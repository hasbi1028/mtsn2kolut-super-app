<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { readClientApiData } from '$lib/client/api';

	type ParentPortalData = {
		parent: { nama: string; phone: string; address: string };
		children: Array<{ id: string; nama: string; nis: string; class_name: string }>;
		timetable: Array<{
			id: string;
			student_id: string;
			student_name: string;
			day_of_week: number;
			start_time: string;
			end_time: string;
			subject_name: string;
			teacher_name: string;
			room_label: string;
		}>;
	};

	let portalPromise = $state<Promise<ParentPortalData> | null>(null);
	let portalData = $state<ParentPortalData | null>(null);

	const dayLabels: Record<number, string> = {
		1: 'Senin',
		2: 'Selasa',
		3: 'Rabu',
		4: 'Kamis',
		5: 'Jumat',
		6: 'Sabtu'
	};

	function applyPortalData(data: ParentPortalData) {
		portalData = data;
		return data;
	}

	async function fetchPortalData() {
		const res = await fetch('/api/portal/parent/me');
		return readClientApiData<ParentPortalData>(res, 'Gagal memuat portal orang tua.');
	}

	function loadPortal() {
		portalPromise = fetchPortalData().then(applyPortalData);
	}

	function retryPortal(reset?: () => void) {
		reset?.();
		loadPortal();
	}

	function portalErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Data portal orang tua belum dapat dimuat.';
	}

	function handleRenderError(error: unknown) {
		console.error('Parent portal render failed', error);
	}

	function fmtTime(value: string) {
		return value?.slice(0, 5) || '—';
	}

	onMount(() => {
		loadPortal();
	});
</script>

<svelte:head><title>Portal Orang Tua — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6 p-6">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-foreground">Portal Orang Tua</h1>
			<p class="mt-1 text-sm text-muted-foreground">Data anak, jadwal, dan informasi akademik keluarga</p>
		</div>
		<Button variant="outline" onclick={loadPortal}>Refresh</Button>
	</div>

	<AsyncContent promise={portalPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<div class="grid gap-3 md:grid-cols-3">
				{#each ['Profil Wali', 'Anak Terhubung', 'Jadwal'] as label (label)}
					<div class="rounded-2xl border border-border bg-muted/50 px-4 py-4">
						<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-muted-foreground">{label}</p>
						<Skeleton class="mt-3 h-8 w-32" />
						<Skeleton class="mt-2 h-4 w-48" />
					</div>
				{/each}
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel title="Portal Orang Tua Belum Tersaji" message={portalErrorMessage(error)} onRetry={() => retryPortal(reset)} />
		{/snippet}

		{#snippet children(value)}
			{@const overview = value as ParentPortalData}
			<div class="grid gap-3 md:grid-cols-3">
				<div class="rounded-2xl border border-primary/20 bg-primary/10 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-primary">Wali</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{overview.parent.nama}</p>
					<p class="text-sm text-muted-foreground">{overview.parent.phone || 'Nomor HP belum tersedia'}</p>
				</div>
				<div class="rounded-2xl border border-accent bg-accent/60 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-accent-foreground">Anak Terhubung</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{overview.children.length}</p>
					<p class="text-sm text-muted-foreground">siswa yang tertaut dengan akun orang tua</p>
				</div>
				<div class="rounded-2xl border border-warning/30 bg-warning/10 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-warning">Jadwal</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{overview.timetable.length}</p>
					<p class="text-sm text-muted-foreground">slot pelajaran semua anak tertaut</p>
				</div>
			</div>

			<Card.Root class="overflow-hidden border-border shadow-sm">
				<Card.Header class="pb-3">
					<Card.Title class="text-base">Anak Terhubung</Card.Title>
				</Card.Header>
				<Card.Content class="p-0">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>Nama Siswa</Table.Head>
								<Table.Head>NIS</Table.Head>
								<Table.Head>Kelas</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each overview.children as child (child.id)}
								<Table.Row>
									<Table.Cell class="font-medium">{child.nama}</Table.Cell>
									<Table.Cell class="font-mono text-sm">{child.nis || '—'}</Table.Cell>
									<Table.Cell>{child.class_name || '—'}</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={3} class="p-4">
										<EmptyStatePanel compact title="Belum ada anak tertaut" description="Hubungi operator madrasah jika relasi keluarga belum sesuai." />
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			</Card.Root>

			<Card.Root class="overflow-hidden border-border shadow-sm">
				<Card.Header class="pb-3">
					<Card.Title class="text-base">Jadwal Anak</Card.Title>
				</Card.Header>
				<Card.Content class="p-0">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>Anak</Table.Head>
								<Table.Head>Hari</Table.Head>
								<Table.Head>Waktu</Table.Head>
								<Table.Head>Mata Pelajaran</Table.Head>
								<Table.Head>Guru</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each overview.timetable as item (`${item.student_id}-${item.id}`)}
								<Table.Row>
									<Table.Cell class="font-medium">{item.student_name}</Table.Cell>
									<Table.Cell>{dayLabels[item.day_of_week] ?? item.day_of_week}</Table.Cell>
									<Table.Cell class="font-mono text-sm">{fmtTime(item.start_time)}-{fmtTime(item.end_time)}</Table.Cell>
									<Table.Cell>{item.subject_name}</Table.Cell>
									<Table.Cell>{item.teacher_name}</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={5} class="p-4">
										<EmptyStatePanel compact title="Jadwal belum tersedia" description="Jadwal anak akan tampil setelah kelas dan mapel terhubung." />
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			</Card.Root>
		{/snippet}
	</AsyncContent>
</div>
