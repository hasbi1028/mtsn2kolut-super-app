<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import {
		fetchParentPortalChildProfile,
		fetchParentPortalChildResults,
		fetchParentPortalChildSchedule,
		fetchParentPortalChildren,
		fetchParentPortalPreviewParents,
		type ParentPortalChild,
		type ParentPortalChildProfilePayload,
		type ParentPortalChildResultsPayload,
		type ParentPortalChildSchedulePayload,
		type ParentPortalPreviewParent
	} from '$lib/client/parent-portal';

	type ChildOverview = {
		child: ParentPortalChild;
		profile: ParentPortalChildProfilePayload;
		schedule: ParentPortalChildSchedulePayload;
		results: ParentPortalChildResultsPayload;
	};

	type ParentPortalOverview = {
		children: ParentPortalChild[];
		details: ChildOverview[];
	};

	let portalPromise = $state<Promise<ParentPortalOverview> | null>(null);
	let portalData = $state<ParentPortalOverview | null>(null);
	let previewParents = $state<ParentPortalPreviewParent[]>([]);
	let selectedPreviewParentID = $state('');
	let previewLoading = $state(false);
	let previewError = $state('');
	let selectedChildID = $state('');

	const dayLabels: Record<number, string> = {
		1: 'Senin',
		2: 'Selasa',
		3: 'Rabu',
		4: 'Kamis',
		5: 'Jumat',
		6: 'Sabtu'
	};

	const currentUser = $derived(page.data.user);
	const isParentPortalUser = $derived(Boolean(currentUser?.roles?.includes('orang_tua') || currentUser?.roles?.includes('parent') || currentUser?.role === 'orang_tua' || currentUser?.role === 'parent'));
	const canPreviewParentPortal = $derived(Boolean(
		currentUser?.roles?.includes('admin') ||
		currentUser?.roles?.includes('kesiswaan') ||
		currentUser?.permissions?.includes('parents.manage')
	));
	const isPreviewMode = $derived(canPreviewParentPortal && !isParentPortalUser);
	const parentDisplayName = $derived(isPreviewMode
		? (previewParents.find((parent) => parent.id === selectedPreviewParentID)?.nama || 'Pilih orang tua/wali')
		: (page.data.account?.profile_nama || page.data.user?.username || 'Orang tua/wali'));
	const parentPhone = $derived(isPreviewMode
		? (previewParents.find((parent) => parent.id === selectedPreviewParentID)?.phone || '')
		: (page.data.account?.contact?.phone ?? ''));
	const selectedChild = $derived.by(() => {
		const details = portalData?.details ?? [];
		return details.find((item) => item.child.id === selectedChildID) ?? details[0] ?? null;
	});
	const totalSchedules = $derived((portalData?.details ?? []).reduce((sum, item) => sum + item.schedule.schedule.length, 0));
	const totalResults = $derived((portalData?.details ?? []).reduce((sum, item) => sum + item.results.results.length, 0));

	function applyPortalData(data: ParentPortalOverview) {
		portalData = data;
		if (!selectedChildID || !data.children.some((child) => child.id === selectedChildID)) {
			selectedChildID = data.children[0]?.id ?? '';
		}
		return data;
	}

	async function fetchPortalData(parentID = ''): Promise<ParentPortalOverview> {
		const childrenPayload = await fetchParentPortalChildren(fetch, parentID);
		const details = await Promise.all(childrenPayload.children.map(async (child) => {
			const [profile, schedule, results] = await Promise.all([
				fetchParentPortalChildProfile(child.id, fetch, parentID),
				fetchParentPortalChildSchedule(child.id, fetch, parentID),
				fetchParentPortalChildResults(child.id, fetch, parentID)
			]);
			return { child, profile, schedule, results };
		}));
		return { children: childrenPayload.children, details };
	}

	async function loadPreviewParents() {
		previewLoading = true;
		previewError = '';
		try {
			const payload = await fetchParentPortalPreviewParents();
			previewParents = payload.parents;
			if (!selectedPreviewParentID && payload.parents.length > 0) {
				selectedPreviewParentID = payload.parents[0].id;
			}
			if (selectedPreviewParentID) loadPortal(selectedPreviewParentID);
		} catch (error) {
			previewError = portalErrorMessage(error);
		} finally {
			previewLoading = false;
		}
	}

	function loadPortal(parentID = selectedPreviewParentID) {
		portalPromise = fetchPortalData(isPreviewMode ? parentID : '').then(applyPortalData);
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
		return value?.slice(0, 5) || '-';
	}

	function scoreText(value: unknown) {
		if (value === null || value === undefined) return '-';
		if (typeof value === 'number' || typeof value === 'string') return String(value);
		if (typeof value === 'object' && 'Float64' in value) {
			const maybeValue = (value as { Float64?: number }).Float64;
			return typeof maybeValue === 'number' ? String(maybeValue) : '-';
		}
		return '-';
	}

	function selectPreviewParent() {
		portalData = null;
		selectedChildID = '';
		if (selectedPreviewParentID) loadPortal(selectedPreviewParentID);
	}

	onMount(() => {
		if (isPreviewMode) {
			void loadPreviewParents();
		} else {
			loadPortal('');
		}
	});
</script>

<svelte:head><title>Portal Orang Tua - MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6 p-6">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-foreground">Portal Orang Tua</h1>
			<p class="mt-1 text-sm text-muted-foreground">Data anak, jadwal, dan hasil asesmen yang tertaut ke akun wali.</p>
		</div>
		<Button variant="outline" onclick={() => loadPortal()}>Refresh</Button>
	</div>

	{#if isPreviewMode}
		<Card.Root class="border-primary/20 bg-primary/5 shadow-sm">
			<Card.Header class="pb-3">
				<Card.Title class="text-base">Preview Portal Orang Tua</Card.Title>
				<Card.Description>Pilih orang tua/wali untuk melihat data anak seperti yang tampil pada akun wali.</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-3">
				<div class="flex flex-col gap-3 md:flex-row md:items-end">
					<div class="flex-1 space-y-2">
						<label for="preview-parent" class="text-sm font-medium">Orang tua/wali</label>
						<select
							id="preview-parent"
							class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
							bind:value={selectedPreviewParentID}
							disabled={previewLoading || previewParents.length === 0}
						>
							{#each previewParents as parent (parent.id)}
								<option value={parent.id}>{parent.nama} · {parent.phone || 'No HP belum ada'} · {parent.linked_student_count ?? 0} anak</option>
							{/each}
						</select>
					</div>
					<Button onclick={selectPreviewParent} disabled={!selectedPreviewParentID || previewLoading}>Lihat Preview</Button>
				</div>
				{#if previewError}
					<p class="rounded-md border border-destructive/20 bg-destructive/10 px-3 py-2 text-sm text-destructive">{previewError}</p>
				{:else if previewLoading}
					<p class="text-sm text-muted-foreground">Memuat daftar orang tua...</p>
				{:else if previewParents.length === 0}
					<p class="text-sm text-muted-foreground">Belum ada data orang tua/wali untuk dipreview.</p>
				{/if}
			</Card.Content>
		</Card.Root>
	{/if}

	<AsyncContent promise={portalPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<div class="grid gap-3 md:grid-cols-3">
				{#each ['Wali', 'Anak Terhubung', 'Jadwal & Hasil'] as label (label)}
					<div class="rounded-lg border border-border bg-muted/50 px-4 py-4">
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
			{@const overview = value as ParentPortalOverview}
			<div class="grid gap-3 md:grid-cols-3">
				<div class="rounded-lg border border-primary/20 bg-primary/10 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-primary">Wali</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{parentDisplayName}</p>
					<p class="text-sm text-muted-foreground">{parentPhone || 'Nomor HP belum tersedia'}</p>
				</div>
				<div class="rounded-lg border border-accent bg-accent/60 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-accent-foreground">Anak Terhubung</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{overview.children.length}</p>
					<p class="text-sm text-muted-foreground">relasi resmi dari data `parent_students`</p>
				</div>
				<div class="rounded-lg border border-warning/30 bg-warning/10 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-warning">Akademik</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{totalSchedules + totalResults}</p>
					<p class="text-sm text-muted-foreground">{totalSchedules} jadwal, {totalResults} hasil</p>
				</div>
			</div>

			<Card.Root class="overflow-hidden border-border shadow-sm">
				<Card.Header class="pb-3">
					<Card.Title class="text-base">Anak Terhubung</Card.Title>
					<Card.Description>Pilih anak untuk melihat profil, jadwal, dan hasil asesmen.</Card.Description>
				</Card.Header>
				<Card.Content class="p-0">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>Nama Siswa</Table.Head>
								<Table.Head>NIS</Table.Head>
								<Table.Head>Kelas</Table.Head>
								<Table.Head>Relasi</Table.Head>
								<Table.Head class="w-32">Aksi</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each overview.children as child (child.id)}
								<Table.Row>
									<Table.Cell class="font-medium">{child.nama}</Table.Cell>
									<Table.Cell class="font-mono text-sm">{child.nis || '-'}</Table.Cell>
									<Table.Cell>{child.class_name || '-'}</Table.Cell>
									<Table.Cell>
										<div class="flex flex-wrap gap-2">
											<Badge variant="outline">{child.relationship || 'wali'}</Badge>
											{#if child.is_primary_contact}<Badge>kontak utama</Badge>{/if}
										</div>
									</Table.Cell>
									<Table.Cell>
										<Button size="sm" variant={selectedChildID === child.id ? 'default' : 'outline'} onclick={() => { selectedChildID = child.id; }}>
											Lihat
										</Button>
									</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={5} class="p-4">
										<EmptyStatePanel compact title="Belum ada anak tertaut" description="Hubungi operator madrasah jika relasi keluarga belum sesuai." />
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			</Card.Root>

			{#if selectedChild}
				<Card.Root class="border-border shadow-sm">
					<Card.Header class="pb-3">
						<Card.Title class="text-base">Profil Anak</Card.Title>
						<Card.Description>{selectedChild.profile.student.nama}</Card.Description>
					</Card.Header>
					<Card.Content>
						<div class="grid gap-3 md:grid-cols-4">
							<div class="rounded-lg border border-border bg-muted/20 px-4 py-3">
								<p class="text-xs text-muted-foreground">NISN</p>
								<p class="mt-1 font-semibold text-foreground">{selectedChild.profile.student.nisn || '-'}</p>
							</div>
							<div class="rounded-lg border border-border bg-muted/20 px-4 py-3">
								<p class="text-xs text-muted-foreground">Kelas</p>
								<p class="mt-1 font-semibold text-foreground">{selectedChild.profile.student.class_code || selectedChild.profile.student.class_name || '-'}</p>
							</div>
							<div class="rounded-lg border border-border bg-muted/20 px-4 py-3">
								<p class="text-xs text-muted-foreground">Status</p>
								<p class="mt-1 font-semibold text-foreground">{selectedChild.profile.student.status || '-'}</p>
							</div>
							<div class="rounded-lg border border-border bg-muted/20 px-4 py-3">
								<p class="text-xs text-muted-foreground">Kontak Wali</p>
								<p class="mt-1 font-semibold text-foreground">{selectedChild.profile.student.parent_phone || '-'}</p>
							</div>
						</div>
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
									<Table.Head>Hari</Table.Head>
									<Table.Head>Waktu</Table.Head>
									<Table.Head>Mata Pelajaran</Table.Head>
									<Table.Head>Guru</Table.Head>
									<Table.Head>Ruang</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each selectedChild.schedule.schedule as item (item.id)}
									<Table.Row>
										<Table.Cell>{dayLabels[item.day_of_week] ?? item.day_of_week}</Table.Cell>
										<Table.Cell class="font-mono text-sm">{fmtTime(item.start_time)}-{fmtTime(item.end_time)}</Table.Cell>
										<Table.Cell class="font-medium">{item.subject_name}</Table.Cell>
										<Table.Cell>{item.teacher_name}</Table.Cell>
										<Table.Cell>{item.room_label || '-'}</Table.Cell>
									</Table.Row>
								{:else}
									<Table.Row>
										<Table.Cell colspan={5} class="p-4">
											<EmptyStatePanel compact title="Jadwal belum tersedia" description="Jadwal akan tampil setelah kelas dan mapel anak terhubung." />
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</Card.Content>
				</Card.Root>

				<Card.Root class="overflow-hidden border-border shadow-sm">
					<Card.Header class="pb-3">
						<Card.Title class="text-base">Hasil Asesmen Anak</Card.Title>
					</Card.Header>
					<Card.Content class="p-0">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Asesmen</Table.Head>
									<Table.Head>Paket</Table.Head>
									<Table.Head>Status</Table.Head>
									<Table.Head>Nilai</Table.Head>
									<Table.Head>Ruang</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each selectedChild.results.results as result (result.participant_id)}
									<Table.Row>
										<Table.Cell class="font-medium">{result.session_title}</Table.Cell>
										<Table.Cell>{result.package_title || '-'}</Table.Cell>
										<Table.Cell><Badge variant="outline">{result.session_status || '-'}</Badge></Table.Cell>
										<Table.Cell class="font-semibold">{scoreText(result.score)}</Table.Cell>
										<Table.Cell>{result.room_name || '-'}</Table.Cell>
									</Table.Row>
								{:else}
									<Table.Row>
										<Table.Cell colspan={5} class="p-4">
											<EmptyStatePanel compact title="Belum ada hasil" description="Hasil akan tampil setelah asesmen dinilai dan tersedia untuk orang tua." />
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</Card.Content>
				</Card.Root>
			{/if}
		{/snippet}
	</AsyncContent>
</div>
