<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import {
		fetchStudentPortalCbtSchedule,
		fetchStudentPortalPreviewStudents,
		fetchStudentPortalProfile,
		fetchStudentPortalResults,
		fetchStudentPortalSchedule,
		revealStudentPortalCbtToken,
		type StudentPortalCbtRevealPayload,
		type StudentPortalCbtScheduleItem,
		type StudentPortalCbtSchedulePayload,
		type StudentPortalPreviewStudent,
		type StudentPortalProfilePayload,
		type StudentPortalResultsPayload,
		type StudentPortalSchedulePayload
	} from '$lib/client/student-portal';

	type PortalOverview = {
		profile: StudentPortalProfilePayload;
		schedule: StudentPortalSchedulePayload;
		results: StudentPortalResultsPayload;
		cbt: StudentPortalCbtSchedulePayload;
	};

	let portalPromise = $state<Promise<PortalOverview> | null>(null);
	let portalData = $state<PortalOverview | null>(null);
	let previewStudents = $state<StudentPortalPreviewStudent[]>([]);
	let selectedPreviewStudentID = $state('');
	let previewLoading = $state(false);
	let previewError = $state('');
	let selectedDay = $state('all');
	let revealOpen = $state(false);
	let revealTarget = $state<StudentPortalCbtScheduleItem | null>(null);
	let roomTokenInput = $state('');
	let revealLoading = $state(false);
	let revealError = $state('');
	let revealedToken = $state<StudentPortalCbtRevealPayload | null>(null);

	const dayLabels: Record<number, string> = {
		1: 'Senin',
		2: 'Selasa',
		3: 'Rabu',
		4: 'Kamis',
		5: 'Jumat',
		6: 'Sabtu'
	};

	const visibleSchedule = $derived.by(() => {
		const schedule = portalData?.schedule.schedule ?? [];
		if (selectedDay === 'all') return schedule;
		return schedule.filter((item) => item.day_of_week === Number(selectedDay));
	});
	const cbtSchedule = $derived(portalData?.cbt.schedule ?? []);
	const currentUser = $derived(page.data.user);
	const isStudentPortalUser = $derived(Boolean(currentUser?.roles?.includes('siswa') || currentUser?.role === 'siswa'));
	const canPreviewStudentPortal = $derived(Boolean(
		currentUser?.roles?.includes('admin') ||
		currentUser?.roles?.includes('kesiswaan') ||
		currentUser?.permissions?.includes('students.manage')
	));
	const isPreviewMode = $derived(canPreviewStudentPortal && !isStudentPortalUser);

	function applyPortalData(data: PortalOverview) {
		portalData = data;
		return data;
	}

	async function fetchPortalData(studentID = ''): Promise<PortalOverview> {
		const [profile, schedule, results, cbt] = await Promise.all([
			fetchStudentPortalProfile(fetch, studentID),
			fetchStudentPortalSchedule(fetch, studentID),
			fetchStudentPortalResults(fetch, studentID),
			fetchStudentPortalCbtSchedule(fetch, studentID)
		]);
		return { profile, schedule, results, cbt };
	}

	async function loadPreviewStudents() {
		previewLoading = true;
		previewError = '';
		try {
			const payload = await fetchStudentPortalPreviewStudents();
			previewStudents = payload.students;
			if (!selectedPreviewStudentID && payload.students.length > 0) {
				selectedPreviewStudentID = payload.students[0].id;
			}
			if (selectedPreviewStudentID) loadPortal(selectedPreviewStudentID);
		} catch (error) {
			previewError = portalErrorMessage(error);
		} finally {
			previewLoading = false;
		}
	}

	function loadPortal(studentID = selectedPreviewStudentID) {
		portalPromise = fetchPortalData(isPreviewMode ? studentID : '').then(applyPortalData);
	}

	function retryPortal(reset?: () => void) {
		reset?.();
		loadPortal();
	}

	function portalErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Data portal siswa belum dapat dimuat.';
	}

	function handleRenderError(error: unknown) {
		console.error('Student portal render failed', error);
	}

	function fmtTime(value: string) {
		return value?.slice(0, 5) || '—';
	}

	function scoreText(value: unknown) {
		if (value === null || value === undefined) return '—';
		if (typeof value === 'number' || typeof value === 'string') return String(value);
		if (typeof value === 'object' && 'Float64' in value) {
			const maybeValue = (value as { Float64?: number }).Float64;
			return typeof maybeValue === 'number' ? String(maybeValue) : '—';
		}
		return '—';
	}

	function fmtDateTime(value: string) {
		if (!value) return '—';
		return new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value));
	}

	function cbtStatusLabel(status: StudentPortalCbtScheduleItem['status']) {
		const labels: Record<StudentPortalCbtScheduleItem['status'], string> = {
			upcoming: 'Belum dibuka',
			token_window: 'Siap dibuka',
			active: 'Sedang berlangsung',
			submitted: 'Selesai',
			closed: 'Ditutup',
			locked: 'Dikunci pengawas'
		};
		return labels[status] ?? status;
	}

	function cbtStatusVariant(status: StudentPortalCbtScheduleItem['status']) {
		if (status === 'active' || status === 'token_window') return 'default';
		if (status === 'locked' || status === 'closed') return 'destructive';
		return 'outline';
	}

	function openReveal(item: StudentPortalCbtScheduleItem) {
		revealTarget = item;
		roomTokenInput = '';
		revealError = '';
		revealedToken = null;
		revealOpen = true;
	}

	async function submitReveal() {
		if (!revealTarget || revealLoading) return;
		const roomToken = roomTokenInput.trim();
		if (roomToken.length < 4) {
			revealError = 'Kode ruang minimal 4 karakter.';
			return;
		}
		revealLoading = true;
		revealError = '';
		try {
			revealedToken = await revealStudentPortalCbtToken(revealTarget.participant_id, roomToken);
		} catch (error) {
			revealError = error instanceof Error && error.message.trim()
				? error.message
				: 'Kode ruang tidak sesuai. Pastikan Anda berada di ruang yang benar.';
		} finally {
			revealLoading = false;
		}
	}

	function selectPreviewStudent() {
		portalData = null;
		selectedDay = 'all';
		if (selectedPreviewStudentID) loadPortal(selectedPreviewStudentID);
	}

	onMount(() => {
		if (isPreviewMode) {
			void loadPreviewStudents();
		} else {
			loadPortal('');
		}
	});
</script>

<svelte:head><title>Portal Siswa — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6 p-6">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-foreground">Portal Siswa</h1>
			<p class="mt-1 text-sm text-muted-foreground">Profil, jadwal, dan hasil asesmen siswa</p>
		</div>
		<Button variant="outline" onclick={() => loadPortal()}>Refresh</Button>
	</div>

	{#if isPreviewMode}
		<Card.Root class="border-primary/20 bg-primary/5 shadow-sm">
			<Card.Header class="pb-3">
				<Card.Title class="text-base">Preview Portal Siswa</Card.Title>
				<Card.Description>Pilih siswa untuk melihat portal seperti yang tampil pada akun siswa. Token ujian tidak bisa dibuka dari mode preview admin.</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-3">
				<div class="flex flex-col gap-3 md:flex-row md:items-end">
					<div class="flex-1 space-y-2">
						<label for="preview-student" class="text-sm font-medium">Siswa</label>
						<select
							id="preview-student"
							class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
							bind:value={selectedPreviewStudentID}
							disabled={previewLoading || previewStudents.length === 0}
						>
							{#each previewStudents as student (student.id)}
								<option value={student.id}>{student.nama} · {student.class_code || 'Tanpa Rombel'} · NIS {student.nis || '—'}</option>
							{/each}
						</select>
					</div>
					<Button onclick={selectPreviewStudent} disabled={!selectedPreviewStudentID || previewLoading}>Lihat Preview</Button>
				</div>
				{#if previewError}
					<p class="rounded-md border border-destructive/20 bg-destructive/10 px-3 py-2 text-sm text-destructive">{previewError}</p>
				{:else if previewLoading}
					<p class="text-sm text-muted-foreground">Memuat daftar siswa...</p>
				{:else if previewStudents.length === 0}
					<p class="text-sm text-muted-foreground">Belum ada siswa aktif untuk dipreview.</p>
				{/if}
			</Card.Content>
		</Card.Root>
	{/if}

	<AsyncContent promise={portalPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<div class="grid gap-3 md:grid-cols-3">
				{#each ['Profil', 'Jadwal', 'Hasil'] as label (label)}
					<div class="rounded-2xl border border-border bg-muted/50 px-4 py-4">
						<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-muted-foreground">{label}</p>
						<Skeleton class="mt-3 h-8 w-32" />
						<Skeleton class="mt-2 h-4 w-48" />
					</div>
				{/each}
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel title="Portal Siswa Belum Tersaji" message={portalErrorMessage(error)} onRetry={() => retryPortal(reset)} />
		{/snippet}

		{#snippet children(value)}
			{@const overview = value as PortalOverview}
			<div class="grid gap-3 md:grid-cols-3">
				<div class="rounded-2xl border border-primary/20 bg-primary/10 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-primary">Nama Siswa</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{overview.profile.student.nama}</p>
					<p class="text-sm text-muted-foreground">NIS {overview.profile.student.nis || '—'} · {overview.profile.student.class_code || 'Belum ada kelas'}</p>
				</div>
				<div class="rounded-2xl border border-accent bg-accent/60 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-accent-foreground">Jadwal Pekanan</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{overview.schedule.schedule.length}</p>
					<p class="text-sm text-muted-foreground">slot pelajaran aktif yang terhubung ke kelas</p>
				</div>
				<div class="rounded-2xl border border-warning/30 bg-warning/10 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-warning">Hasil Asesmen</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{overview.results.results.length}</p>
					<p class="text-sm text-muted-foreground">hasil yang tersedia untuk akun ini</p>
				</div>
			</div>

			<Card.Root class="overflow-hidden border-border shadow-sm">
				<Card.Header class="pb-3">
					<div class="flex flex-wrap items-center justify-between gap-3">
						<Card.Title class="text-base">Jadwal Siswa</Card.Title>
						<select class="rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={selectedDay}>
							<option value="all">Semua hari</option>
							{#each Object.entries(dayLabels) as [day, label] (day)}
								<option value={day}>{label}</option>
							{/each}
						</select>
					</div>
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
							{#each visibleSchedule as item (item.id)}
								<Table.Row>
									<Table.Cell>{dayLabels[item.day_of_week] ?? item.day_of_week}</Table.Cell>
									<Table.Cell class="font-mono text-sm">{fmtTime(item.start_time)}-{fmtTime(item.end_time)}</Table.Cell>
									<Table.Cell class="font-medium">{item.subject_name}</Table.Cell>
									<Table.Cell>{item.teacher_name}</Table.Cell>
									<Table.Cell>{item.room_label || '—'}</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={5} class="p-4">
										<EmptyStatePanel compact title="Jadwal belum tersedia" description="Jadwal akan tampil setelah kelas dan mapel siswa terhubung." />
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			</Card.Root>

			<Card.Root class="overflow-hidden border-border shadow-sm">
				<Card.Header class="pb-3">
					<div class="flex flex-wrap items-start justify-between gap-3">
						<div>
							<Card.Title class="text-base">Jadwal Ujian CBT</Card.Title>
							<Card.Description>Token siswa hanya dibuka di window ujian dengan token ruang dari pengawas.</Card.Description>
						</div>
						<Badge variant="outline">{cbtSchedule.length} sesi</Badge>
					</div>
				</Card.Header>
				<Card.Content class="p-0">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>Ujian</Table.Head>
								<Table.Head>Waktu</Table.Head>
								<Table.Head>Ruang</Table.Head>
								<Table.Head>Status</Table.Head>
								<Table.Head>Token</Table.Head>
								<Table.Head class="text-right">Aksi</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each cbtSchedule as item (item.participant_id)}
								<Table.Row>
									<Table.Cell>
										<div class="font-medium">{item.session_title}</div>
										<div class="text-xs text-muted-foreground">{item.package_title || 'Paket belum tersedia'}</div>
									</Table.Cell>
									<Table.Cell class="text-sm">
										<div>{fmtDateTime(item.scheduled_start)}</div>
										<div class="text-xs text-muted-foreground">sampai {fmtDateTime(item.scheduled_end)} · {item.duration_minutes} menit</div>
									</Table.Cell>
									<Table.Cell>
										<div>{item.room_name || 'Belum ditetapkan'}</div>
										<div class="text-xs text-muted-foreground">Meja {item.seat_no ?? '—'}</div>
									</Table.Cell>
									<Table.Cell><Badge variant={cbtStatusVariant(item.status)}>{cbtStatusLabel(item.status)}</Badge></Table.Cell>
									<Table.Cell class="font-mono text-xs">{item.token_masked || 'Belum tersedia'}</Table.Cell>
									<Table.Cell>
										<div class="flex justify-end gap-2">
											{#if !isPreviewMode}
												<Button size="sm" variant="outline" href={resolve(`/portal/siswa/cbt/${item.participant_id}`)}>Kartu</Button>
											{/if}
											<Button size="sm" disabled={isPreviewMode || !item.can_reveal_token} onclick={() => openReveal(item)}>
												{isPreviewMode ? 'Preview Saja' : item.can_reveal_token ? 'Buka Kode' : 'Belum Dibuka'}
											</Button>
										</div>
									</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={6} class="p-4">
										<EmptyStatePanel compact title="Belum ada jadwal CBT" description="Jadwal ujian akan tampil setelah peserta didaftarkan ke sesi CBT." />
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			</Card.Root>

			<Card.Root class="overflow-hidden border-border shadow-sm">
				<Card.Header class="pb-3">
					<Card.Title class="text-base">Hasil Asesmen</Card.Title>
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
							{#each overview.results.results as result (result.participant_id)}
								<Table.Row>
									<Table.Cell class="font-medium">{result.session_title}</Table.Cell>
									<Table.Cell>{result.package_title || '—'}</Table.Cell>
									<Table.Cell><Badge variant="outline">{result.session_status || '—'}</Badge></Table.Cell>
									<Table.Cell class="font-semibold">{scoreText(result.score)}</Table.Cell>
									<Table.Cell>{result.room_name || '—'}</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={5} class="p-4">
										<EmptyStatePanel compact title="Belum ada hasil" description="Hasil akan tampil setelah asesmen dinilai dan tersedia untuk siswa." />
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

<Dialog.Root bind:open={revealOpen}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>Buka Kode Ujian</Dialog.Title>
			<Dialog.Description>Masukkan token ruang yang diberikan pengawas setelah peserta siap di ruang ujian.</Dialog.Description>
		</Dialog.Header>
		<div class="mt-4 space-y-4">
			{#if revealTarget}
				<div class="rounded-md border border-border bg-muted/40 p-3 text-sm">
					<div class="font-medium">{revealTarget.session_title}</div>
					<div class="text-muted-foreground">{revealTarget.room_name || 'Ruang belum ditetapkan'} · Meja {revealTarget.seat_no ?? '—'}</div>
				</div>
			{/if}

			{#if revealedToken}
				<div class="rounded-md border border-primary/20 bg-primary/10 p-4">
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-primary">Token Ujian Anda</p>
					<p class="mt-2 break-all font-mono text-xl font-bold text-primary">{revealedToken.token}</p>
					<p class="mt-2 text-xs text-muted-foreground">Gunakan hanya di perangkat Anda sendiri. Berlaku sampai {fmtDateTime(revealedToken.expires_at)}.</p>
				</div>
			{:else}
				<div class="space-y-2">
					<label for="portal-room-token" class="text-sm font-medium">Token Ruang</label>
					<Input id="portal-room-token" bind:value={roomTokenInput} autocomplete="off" placeholder="Masukkan token ruang" />
				</div>
				{#if revealError}
					<p class="rounded-md border border-destructive/20 bg-destructive/10 px-3 py-2 text-sm text-destructive">{revealError}</p>
				{/if}
			{/if}
		</div>
		<Dialog.Footer>
			<Button variant="outline" onclick={() => (revealOpen = false)}>Tutup</Button>
			{#if !revealedToken}
				<Button onclick={() => void submitReveal()} disabled={revealLoading}>
					{revealLoading ? 'Memeriksa...' : 'Buka Kode'}
				</Button>
			{/if}
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
