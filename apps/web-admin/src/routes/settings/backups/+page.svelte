<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button';
	import { toast } from '$lib/components/ui/sonner';
	import { readClientApiData } from '$lib/client/api';

	type BackupFile = {
		id: string;
		name: string;
		kind: 'scheduled' | 'manual-pre-change' | 'manual' | string;
		size_bytes: number;
		created_at: string;
		sha256?: string;
		sha256_available: boolean;
		is_latest: boolean;
		downloadable: boolean;
	};

	type BackupList = {
		items: BackupFile[];
		meta: { total: number };
	};

	type BackupStatus = {
		timer_name: string;
		service_name: string;
		timer_enabled: boolean;
		timer_active: boolean;
		schedule: string;
		timezone: string;
		last_run_at?: string;
		last_run_success?: boolean;
		next_run_at?: string;
		latest_backup?: BackupFile;
		retention_days: number;
		backup_count: number;
		backup_dir_size_bytes: number;
		health: 'ok' | 'warning' | 'error' | string;
		warnings: string[];
	};

	type RestoreValidation = {
		backup_id: string;
		valid: boolean;
		object_count: number;
		preview: string[];
		checked_at: string;
		command: string;
		warnings: string[];
	};

	type RestoreCommand = {
		backup_id: string;
		generated_at: string;
		safety_level: string;
		warnings: string[];
		preflight_steps: string[];
		commands: string[];
		rollback_note: string;
	};

	type OffsiteStatus = {
		configured: boolean;
		provider?: string;
		target_label?: string;
		source: string;
		last_sync_at?: string;
		last_sync_success?: boolean;
		remote_backup_count: number;
		remote_size_bytes: number;
		health: 'ok' | 'warning' | 'error' | string;
		warnings: string[];
	};

	let loading = $state(true);
	let refreshing = $state(false);
	let errorMessage = $state('');
	let status = $state<BackupStatus | null>(null);
	let offsite = $state<OffsiteStatus | null>(null);
	let backups = $state<BackupFile[]>([]);
	let manualBackupRunning = $state(false);
	let manualBackupReason = $state('');
	let restoreBusyID = $state('');
	let restoreValidation = $state<RestoreValidation | null>(null);
	let restoreCommand = $state<RestoreCommand | null>(null);

	const isAdmin = $derived(Boolean(page.data.user?.roles?.includes('admin') || page.data.user?.role === 'admin'));
	const permissions = $derived(page.data.user?.permissions ?? []);
	const canDownload = $derived(isAdmin || permissions.includes('backup.download'));
	const canCreate = $derived(isAdmin || permissions.includes('backup.create'));
	const canRestorePlan = $derived(isAdmin || permissions.includes('backup.restore_plan'));

	onMount(() => {
		void loadBackups();
	});

	async function loadBackups(showToast = false) {
		if (showToast) refreshing = true;
		else loading = true;
		errorMessage = '';
		try {
			const [statusData, offsiteData, listData] = await Promise.all([
				fetch('/api/system/backups/status').then((response) =>
					readClientApiData<BackupStatus>(response, 'Gagal memuat status backup')
				),
				fetch('/api/system/backups/offsite').then((response) =>
					readClientApiData<OffsiteStatus>(response, 'Gagal memuat status offsite')
				),
				fetch('/api/system/backups').then((response) =>
					readClientApiData<BackupList>(response, 'Gagal memuat daftar backup')
				)
			]);
			status = statusData;
			offsite = offsiteData;
			backups = Array.isArray(listData.items) ? listData.items : [];
			if (showToast) toast.success('Status backup diperbarui');
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Backup Center belum dapat dimuat.';
			if (showToast) toast.error(errorMessage);
		} finally {
			loading = false;
			refreshing = false;
		}
	}


	async function runManualBackup() {
		if (!canCreate || manualBackupRunning) return;
		const confirmed = window.confirm('Buat backup database PostgreSQL sekarang? Proses ini aman, tetapi dapat memakan waktu beberapa menit.');
		if (!confirmed) return;
		manualBackupRunning = true;
		try {
			const data = await fetch('/api/system/backups', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ reason: manualBackupReason })
			}).then((response) => readClientApiData<{ status: string; output?: string; error?: string }>(response, 'Gagal menjalankan backup manual'));
			if (data.status === 'success') {
				toast.success('Backup manual berhasil dibuat');
				manualBackupReason = '';
				await loadBackups(false);
			} else {
				toast.error(data.error || 'Backup manual belum berhasil');
			}
		} catch (error) {
			toast.error(error instanceof Error ? error.message : 'Backup manual gagal dijalankan');
		} finally {
			manualBackupRunning = false;
		}
	}

	async function validateRestore(backup: BackupFile) {
		if (!canRestorePlan || restoreBusyID) return;
		restoreBusyID = backup.id;
		restoreValidation = null;
		restoreCommand = null;
		try {
			const data = await fetch(`/api/system/backups/${encodeURIComponent(backup.id)}/validate-restore`, {
				method: 'POST'
			}).then((response) => readClientApiData<RestoreValidation>(response, 'Gagal memvalidasi metadata restore'));
			restoreValidation = data;
			if (data.valid) toast.success(`Validasi backup OK: ${data.object_count} objek terbaca`);
			else toast.error('Validasi backup belum aman dipakai');
		} catch (error) {
			toast.error(error instanceof Error ? error.message : 'Validasi restore gagal');
		} finally {
			restoreBusyID = '';
		}
	}

	async function generateRestoreCommand(backup: BackupFile) {
		if (!canRestorePlan || restoreBusyID) return;
		restoreBusyID = backup.id;
		restoreCommand = null;
		try {
			const data = await fetch(`/api/system/backups/${encodeURIComponent(backup.id)}/restore-command`, {
				method: 'POST'
			}).then((response) => readClientApiData<RestoreCommand>(response, 'Gagal membuat SOP restore'));
			restoreCommand = data;
			toast.success('SOP restore manual dibuat');
		} catch (error) {
			toast.error(error instanceof Error ? error.message : 'Gagal membuat SOP restore');
		} finally {
			restoreBusyID = '';
		}
	}

	function statusLabel(value: string | undefined) {
		if (value === 'ok') return 'Sehat';
		if (value === 'warning') return 'Perlu perhatian';
		if (value === 'error') return 'Bermasalah';
		return value || 'Tidak diketahui';
	}

	function statusBadgeClass(value: string | undefined) {
		if (value === 'ok') return 'bg-success/10 text-success border border-success/20';
		if (value === 'warning') return 'bg-warning/10 text-warning border border-warning/20';
		if (value === 'error') return 'bg-destructive/10 text-destructive border border-destructive/20';
		return 'bg-transparent text-muted-foreground border border-border';
	}

	function kindLabel(kind: string) {
		switch (kind) {
			case 'scheduled': return 'Harian';
			case 'manual-pre-change': return 'Pra-perubahan';
			case 'manual': return 'Manual';
			default: return kind || 'Backup';
		}
	}

	function formatBytes(value: number | undefined | null) {
		const bytes = Number(value ?? 0);
		if (!Number.isFinite(bytes) || bytes <= 0) return '0 B';
		const units = ['B', 'KB', 'MB', 'GB', 'TB'];
		let size = bytes;
		let index = 0;
		while (size >= 1024 && index < units.length - 1) {
			size /= 1024;
			index += 1;
		}
		return `${size.toLocaleString('id-ID', { maximumFractionDigits: index === 0 ? 0 : 1 })} ${units[index]}`;
	}

	function formatDate(value: string | undefined | null) {
		if (!value) return '—';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value;
		return new Intl.DateTimeFormat('id-ID', {
			timeZone: 'Asia/Makassar',
			dateStyle: 'medium',
			timeStyle: 'short'
		}).format(date) + ' WITA';
	}

	function downloadHref(id: string) {
		return `/api/system/backups/${encodeURIComponent(id)}/download`;
	}
</script>

<svelte:head>
	<title>Backup & Restore · MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="mx-auto flex w-full max-w-7xl flex-col gap-6 p-4 md:p-6">
	<!-- Breadcrumb -->
	<div class="flex items-center gap-2 text-sm text-base-content/70">
		<a href={resolve('/')} class="hover:text-base-content">Beranda</a>
		<span>/</span>
		<a href={resolve('/settings')} class="hover:text-base-content">Pengaturan</a>
		<span>/</span>
		<span class="text-base-content font-medium">Backup & Restore</span>
	</div>

	<!-- Header -->
	<section class="card bg-base-100 border border-base-300 shadow-sm">
		<div class="card-body">
			<div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
				<div class="space-y-2">
					<p class="badge badge-sm badge-outline uppercase tracking-wider">Pengaturan Sistem</p>
					<h1 class="text-2xl font-bold text-base-content md:text-3xl">Backup & Restore</h1>
					<p class="max-w-3xl text-sm text-base-content/70">
						Pantau backup PostgreSQL harian, cek kesehatan timer, dan unduh file backup resmi. Sprint ini bersifat read-only: belum ada restore production langsung dari aplikasi.
					</p>
				</div>
				<div class="flex flex-col gap-2 sm:flex-row sm:items-center">
					<input
						class="h-10 rounded-lg border border-input bg-background px-3 text-sm text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
						placeholder="Alasan backup manual (opsional)"
						maxlength="200"
						bind:value={manualBackupReason}
						disabled={!canCreate || manualBackupRunning}
					/>
					<button class="inline-flex items-center justify-center gap-2 rounded-lg h-10 px-4 text-sm font-bold text-primary-foreground bg-primary border border-primary hover:bg-primary/90 transition-colors disabled:opacity-40 disabled:pointer-events-none shadow-sm" onclick={runManualBackup} disabled={!canCreate || manualBackupRunning}>
						{manualBackupRunning ? 'Membuat Backup...' : 'Buat Backup Sekarang'}
					</button>
					<button class="inline-flex items-center justify-center gap-2 rounded-lg h-10 px-4 text-sm font-medium text-foreground border border-input bg-background hover:bg-accent transition-colors disabled:opacity-40 disabled:pointer-events-none" onclick={() => loadBackups(true)} disabled={refreshing || manualBackupRunning}>
						{refreshing ? 'Memuat...' : 'Refresh Status'}
					</button>
				</div>
			</div>
		</div>
	</section>

	{#if loading}
		<div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
			{#each Array(4) as _}
				<div class="card bg-base-100 border border-base-300 animate-pulse h-32"></div>
			{/each}
		</div>
		<div class="card bg-base-100 border border-base-300 shadow-sm mt-4">
			<div class="card-body p-0">
				<div class="border-b border-base-300 p-5">
					<div class="skeleton h-5 w-36"></div>
					<div class="skeleton h-4 w-56 mt-2"></div>
				</div>
				<div class="overflow-x-auto p-4">
					<div class="space-y-4">
						{#each Array(5) as _}
							<div class="flex gap-6 px-2">
								<div class="skeleton h-4 w-48"></div>
								<div class="skeleton h-4 w-32"></div>
								<div class="skeleton h-4 w-20"></div>
								<div class="skeleton h-4 w-24"></div>
								<div class="skeleton h-4 w-16"></div>
								<div class="skeleton h-4 w-20 ml-auto"></div>
							</div>
						{/each}
					</div>
				</div>
			</div>
		</div>
	{:else if errorMessage}
		<div class="alert alert-error shadow-sm">
			<div>
				<p class="font-semibold">Backup Center belum dapat dimuat</p>
				<p class="mt-1">{errorMessage}</p>
				<button class="inline-flex items-center justify-center rounded-lg h-8 px-3 text-xs font-medium border border-input bg-background text-foreground hover:bg-accent transition-colors mt-2" onclick={() => loadBackups(true)}>Coba Lagi</button>
			</div>
		</div>
	{:else if status}
		<!-- Status Cards -->
		<section class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
			<div class="card bg-base-100 border border-base-300 shadow-sm">
				<div class="card-body">
					<p class="card-title text-sm">Status Backup</p>
					<span class={statusBadgeClass(status.health)}>{statusLabel(status.health)}</span>
					<p class="text-xs text-base-content/70 mt-2">Timer: {status.timer_active ? 'aktif' : 'tidak aktif'} · {status.timer_enabled ? 'enabled' : 'belum enabled'}</p>
				</div>
			</div>

			<div class="card bg-base-100 border border-base-300 shadow-sm">
				<div class="card-body">
					<p class="card-title text-sm">Backup Terakhir</p>
					<p class="truncate text-lg font-bold text-base-content">{status.latest_backup?.name ?? 'Belum ada backup'}</p>
					<p class="text-sm text-base-content/70">{formatDate(status.latest_backup?.created_at)}</p>
					<p class="text-xs text-base-content/70">{formatBytes(status.latest_backup?.size_bytes)}</p>
				</div>
			</div>

			<div class="card bg-base-100 border border-base-300 shadow-sm">
				<div class="card-body">
					<p class="card-title text-sm">Jadwal Berikutnya</p>
					<p class="text-lg font-bold text-base-content">{formatDate(status.next_run_at)}</p>
					<p class="text-xs text-base-content/70">{status.schedule} · {status.timezone}</p>
				</div>
			</div>

			<div class="card bg-base-100 border border-base-300 shadow-sm">
				<div class="card-body">
					<p class="card-title text-sm">Retensi & Ukuran</p>
					<p class="text-lg font-bold text-base-content">{status.retention_days} hari</p>
					<p class="text-sm text-base-content/70">{status.backup_count} file · {formatBytes(status.backup_dir_size_bytes)}</p>
				</div>
			</div>
		</section>

		{#if status.warnings.length > 0}
			<section class="alert alert-warning shadow-sm">
				<div>
					<p class="font-semibold">Perlu perhatian</p>
					<ul class="mt-2 list-disc space-y-1 pl-5">
						{#each status.warnings as warning}
							<li>{warning}</li>
						{/each}
					</ul>
				</div>
			</section>
		{/if}

		<!-- Backup Table -->
		<section class="card bg-base-100 border border-base-300 shadow-sm">
			<div class="card-body p-0">
				<div class="flex flex-col gap-2 border-b border-base-300 p-5 md:flex-row md:items-center md:justify-between">
					<div>
						<h2 class="card-title">Daftar Backup</h2>
						<p class="text-sm text-base-content/70">Hanya file <code>.dump</code> dari direktori backup resmi yang ditampilkan.</p>
					</div>
					{#if status.latest_backup && canDownload}
						<a class="inline-flex items-center justify-center rounded-lg h-8 px-3 text-xs font-medium border border-input bg-background text-foreground hover:bg-accent transition-colors" href={downloadHref(status.latest_backup.id)}>
							Download Latest
						</a>
					{/if}
				</div>

				{#if backups.length === 0}
					<div class="alert alert-info m-4 shadow-sm"><span>Belum ada file backup PostgreSQL yang tersedia.</span></div>
				{:else}
					<div class="overflow-x-auto">
						<table class="w-full min-w-[980px]">
							<thead>
								<tr>
									<th>File</th>
									<th>Tanggal</th>
									<th>Jenis</th>
									<th>Ukuran</th>
									<th>Checksum</th>
									<th class="text-right">Aksi</th>
								</tr>
							</thead>
							<tbody class="divide-y divide-border">
								{#each backups as backup}
									<tr class="align-top hover:bg-base-200/50">
										<td>
											<div class="font-medium text-base-content">{backup.name}</div>
											{#if backup.is_latest}
												<span class="inline-flex items-center rounded-full px-1.5 py-0.5 text-[9px] font-semibold uppercase tracking-wider bg-success/10 text-success border border-success/20 mt-1">latest</span>
											{/if}
										</td>
										<td class="text-base-content/70">{formatDate(backup.created_at)}</td>
										<td class="text-base-content/70">{kindLabel(backup.kind)}</td>
										<td class="text-base-content/70">{formatBytes(backup.size_bytes)}</td>
										<td class="text-base-content/70">
											{#if backup.sha256_available}
												<span class="font-mono text-xs">{backup.sha256?.slice(0, 12)}…</span>
											{:else}
												<span class="text-base-content/70/60">Belum ada</span>
											{/if}
										</td>
										<td class="text-right">
											{#if canRestorePlan || (canDownload && backup.downloadable)}
												<div class="relative">
													<button class="inline-flex items-center justify-center gap-1.5 rounded-lg h-8 px-3 text-xs font-medium border border-input bg-background text-foreground hover:bg-accent transition-colors"
														onclick={(ev) => {
															const menu = ev.currentTarget.nextElementSibling as HTMLElement;
															if (menu) menu.classList.toggle('hidden');
														}}>
														Aksi
														<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
														</svg>
													</button>
													<ul class="hidden absolute right-0 z-50 mt-1 w-52 rounded-lg border border-border bg-card p-1 shadow-xl" onclick={(ev) => ev.stopPropagation()}>
														{#if canRestorePlan}
															<li><button class="flex w-full items-center gap-2 rounded-md px-3 py-1.5 text-xs font-medium text-foreground hover:bg-accent transition-colors" onclick={() => validateRestore(backup)} disabled={restoreBusyID === backup.id}>{restoreBusyID === backup.id ? 'Memeriksa...' : 'Validasi'}</button></li>
															<li><button class="flex w-full items-center gap-2 rounded-md px-3 py-1.5 text-xs font-medium text-foreground hover:bg-accent transition-colors" onclick={() => generateRestoreCommand(backup)} disabled={restoreBusyID === backup.id}>{restoreBusyID === backup.id ? 'Memeriksa...' : 'SOP Restore'}</button></li>
														{/if}
														{#if canDownload && backup.downloadable}
															<li><a class="flex w-full items-center gap-2 rounded-md px-3 py-1.5 text-xs font-medium text-foreground hover:bg-accent transition-colors" href={downloadHref(backup.id)}>Download</a></li>
														{:else}
															<li><span class="text-xs text-base-content/70 px-3 py-2 block">Butuh izin download</span></li>
														{/if}
													</ul>
												</div>
											{:else}
												<span class="text-xs text-base-content/70">Tidak ada aksi</span>
											{/if}
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</div>
		</section>
	{/if}

	{#if restoreValidation || restoreCommand}
		<section class="grid gap-4 lg:grid-cols-2">
			{#if restoreValidation}
				<div class="card bg-base-100 border border-base-300 shadow-sm">
					<div class="card-body">
						<div class="flex items-start justify-between gap-3">
							<div>
								<h2 class="card-title">Hasil Validasi Restore</h2>
								<p class="text-sm text-base-content/70">{restoreValidation.backup_id} · {formatDate(restoreValidation.checked_at)}</p>
							</div>
							<span class={restoreValidation.valid ? 'badge badge-success' : 'badge badge-error'}>{restoreValidation.valid ? 'Valid' : 'Perlu cek manual'}</span>
						</div>
						<p class="mt-3 text-sm text-base-content/70">Object terbaca: <strong>{restoreValidation.object_count}</strong></p>
						{#if restoreValidation.preview.length > 0}
							<pre class="mt-3 max-h-56 overflow-auto rounded-2xl bg-base-300 p-4 text-xs text-base-content">{restoreValidation.preview.join('\n')}</pre>
						{/if}
						{#if restoreValidation.warnings.length > 0}
							<ul class="mt-3 list-disc space-y-1 pl-5 text-sm text-warning">
								{#each restoreValidation.warnings as warning}<li>{warning}</li>{/each}
							</ul>
						{/if}
					</div>
				</div>
			{/if}

			{#if restoreCommand}
				<div class="card bg-base-100 border border-base-300 shadow-sm">
					<div class="card-body">
						<h2 class="card-title">SOP Restore Manual</h2>
						<p class="mt-1 text-sm text-base-content/70">{restoreCommand.backup_id} · {restoreCommand.safety_level}</p>
						<ul class="mt-3 list-decimal space-y-1 pl-5 text-sm text-base-content/70">
							{#each restoreCommand.preflight_steps as step}<li>{step}</li>{/each}
						</ul>
						<pre class="mt-3 max-h-72 overflow-auto rounded-2xl bg-base-300 p-4 text-xs text-base-content">{restoreCommand.commands.join('\n')}</pre>
						<div class="mt-3 alert alert-error">{restoreCommand.rollback_note}</div>
					</div>
				</div>
			{/if}
		</section>
	{/if}

	<section class="grid gap-4 lg:grid-cols-2">
		<div class="card bg-base-100 border border-base-300 shadow-sm">
			<div class="card-body">
				<h2 class="card-title">Restore Production</h2>
				<p class="mt-2 text-sm text-base-content/70">
					Restore database production bersifat destruktif, sehingga tombol restore langsung belum dibuka pada Sprint Backup 1. Tahap berikutnya akan menambahkan validasi restore dan generate SOP/command manual yang aman.
				</p>
				<div class="mt-4 alert alert-error">
					Tidak ada aksi restore production dari halaman ini. Tombol backup manual hanya membuat file dump baru, bukan mengubah database.
				</div>
			</div>
		</div>
		<div class="card bg-base-100 border border-base-300 shadow-sm">
			<div class="card-body">
				<div class="flex items-start justify-between gap-3">
					<div>
						<h2 class="card-title">Backup Offsite</h2>
						<p class="mt-1 text-sm text-base-content/70">Monitoring lokasi backup kedua: Google Drive, S3-compatible storage, NAS, rsync server, atau mount eksternal.</p>
					</div>
					{#if offsite}
						<span class={statusBadgeClass(offsite.health)}>{statusLabel(offsite.health)}</span>
					{/if}
				</div>
				{#if offsite}
					<div class="mt-4 grid gap-3 text-sm sm:grid-cols-2">
						<div class="rounded-2xl bg-base-200/30 p-3">
							<p class="text-xs font-semibold uppercase tracking-wide text-base-content/70">Status</p>
							<p class="mt-1 font-semibold text-base-content">{offsite.configured ? 'Terkonfigurasi' : 'Belum dikonfigurasi'}</p>
							<p class="mt-1 text-xs text-base-content/70">{offsite.provider || 'Provider belum diatur'} · {offsite.source}</p>
						</div>
						<div class="rounded-2xl bg-base-200/30 p-3">
							<p class="text-xs font-semibold uppercase tracking-wide text-base-content/70">Sync Terakhir</p>
							<p class="mt-1 font-semibold text-base-content">{formatDate(offsite.last_sync_at)}</p>
							<p class="mt-1 text-xs text-base-content/70">{offsite.last_sync_success === true ? 'berhasil' : offsite.last_sync_success === false ? 'gagal' : 'belum ada status'}</p>
						</div>
						<div class="rounded-2xl bg-base-200/30 p-3">
							<p class="text-xs font-semibold uppercase tracking-wide text-base-content/70">Remote Backup</p>
							<p class="mt-1 font-semibold text-base-content">{offsite.remote_backup_count} file</p>
							<p class="mt-1 text-xs text-base-content/70">{formatBytes(offsite.remote_size_bytes)}</p>
						</div>
						<div class="rounded-2xl bg-base-200/30 p-3">
							<p class="text-xs font-semibold uppercase tracking-wide text-base-content/70">Target</p>
							<p class="mt-1 truncate font-semibold text-base-content">{offsite.target_label || 'Belum ada label target'}</p>
							<p class="mt-1 text-xs text-base-content/70">Cloud secret tidak disimpan di aplikasi</p>
						</div>
					</div>
					{#if offsite.warnings.length > 0}
						<div class="mt-4 alert alert-warning">
							<ul class="list-disc space-y-1 pl-5">
								{#each offsite.warnings as warning}<li>{warning}</li>{/each}
							</ul>
						</div>
					{/if}
				{:else}
					<div class="mt-4 alert alert-warning">Status offsite belum dimuat.</div>
				{/if}
			</div>
		</div>
	</section>
</div>
