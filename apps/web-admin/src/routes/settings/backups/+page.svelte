<script lang="ts">
	import { onMount } from 'svelte';
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

	let loading = $state(true);
	let refreshing = $state(false);
	let errorMessage = $state('');
	let status = $state<BackupStatus | null>(null);
	let backups = $state<BackupFile[]>([]);
	let manualBackupRunning = $state(false);
	let manualBackupReason = $state('');

	const isAdmin = $derived(Boolean(page.data.user?.roles?.includes('admin') || page.data.user?.role === 'admin'));
	const permissions = $derived(page.data.user?.permissions ?? []);
	const canDownload = $derived(isAdmin || permissions.includes('backup.download'));
	const canCreate = $derived(isAdmin || permissions.includes('backup.create'));

	onMount(() => {
		void loadBackups();
	});

	async function loadBackups(showToast = false) {
		if (showToast) refreshing = true;
		else loading = true;
		errorMessage = '';
		try {
			const [statusData, listData] = await Promise.all([
				fetch('/api/system/backups/status').then((response) =>
					readClientApiData<BackupStatus>(response, 'Gagal memuat status backup')
				),
				fetch('/api/system/backups').then((response) =>
					readClientApiData<BackupList>(response, 'Gagal memuat daftar backup')
				)
			]);
			status = statusData;
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

	function statusLabel(value: string | undefined) {
		if (value === 'ok') return 'Sehat';
		if (value === 'warning') return 'Perlu perhatian';
		if (value === 'error') return 'Bermasalah';
		return value || 'Tidak diketahui';
	}

	function statusBadgeClass(value: string | undefined) {
		if (value === 'ok') return 'border-emerald-200 bg-emerald-50 text-emerald-700';
		if (value === 'warning') return 'border-amber-200 bg-amber-50 text-amber-700';
		if (value === 'error') return 'border-red-200 bg-red-50 text-red-700';
		return 'border-slate-200 bg-slate-50 text-slate-600';
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
	<section class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
			<div class="space-y-2">
				<p class="text-xs font-semibold uppercase tracking-[0.24em] text-slate-500">Pengaturan Sistem</p>
				<h1 class="text-2xl font-bold text-slate-950 md:text-3xl">Backup & Restore</h1>
				<p class="max-w-3xl text-sm text-slate-600">
					Pantau backup PostgreSQL harian, cek kesehatan timer, dan unduh file backup resmi. Sprint ini bersifat read-only: belum ada restore production langsung dari aplikasi.
				</p>
			</div>
			<div class="flex flex-col gap-2 sm:flex-row sm:items-center">
				<input
					class="rounded-xl border border-slate-300 px-3 py-2 text-sm"
					placeholder="Alasan backup manual (opsional)"
					maxlength="200"
					bind:value={manualBackupReason}
					disabled={!canCreate || manualBackupRunning}
				/>
				<Button onclick={runManualBackup} disabled={!canCreate || manualBackupRunning}>
					{manualBackupRunning ? 'Membuat Backup...' : 'Buat Backup Sekarang'}
				</Button>
				<Button variant="outline" onclick={() => loadBackups(true)} disabled={refreshing || manualBackupRunning}>
					{refreshing ? 'Memuat...' : 'Refresh Status'}
				</Button>
			</div>
		</div>
	</section>

	{#if loading}
		<div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
			{#each Array(4) as _}
				<div class="h-32 animate-pulse rounded-3xl border border-slate-200 bg-slate-100"></div>
			{/each}
		</div>
	{:else if errorMessage}
		<div class="rounded-3xl border border-red-200 bg-red-50 p-5 text-sm text-red-700">
			<p class="font-semibold">Backup Center belum dapat dimuat</p>
			<p class="mt-1">{errorMessage}</p>
			<Button class="mt-4" variant="outline" onclick={() => loadBackups(true)}>Coba Lagi</Button>
		</div>
	{:else if status}
		<section class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
			<div class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
				<p class="text-xs font-semibold uppercase tracking-wide text-slate-500">Status Backup</p>
				<div class={`mt-3 inline-flex rounded-full border px-3 py-1 text-sm font-semibold ${statusBadgeClass(status.health)}`}>
					{statusLabel(status.health)}
				</div>
				<p class="mt-3 text-xs text-slate-500">Timer: {status.timer_active ? 'aktif' : 'tidak aktif'} · {status.timer_enabled ? 'enabled' : 'belum enabled'}</p>
			</div>

			<div class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
				<p class="text-xs font-semibold uppercase tracking-wide text-slate-500">Backup Terakhir</p>
				<p class="mt-3 truncate text-lg font-bold text-slate-950">{status.latest_backup?.name ?? 'Belum ada backup'}</p>
				<p class="mt-1 text-sm text-slate-600">{formatDate(status.latest_backup?.created_at)}</p>
				<p class="mt-1 text-xs text-slate-500">{formatBytes(status.latest_backup?.size_bytes)}</p>
			</div>

			<div class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
				<p class="text-xs font-semibold uppercase tracking-wide text-slate-500">Jadwal Berikutnya</p>
				<p class="mt-3 text-lg font-bold text-slate-950">{formatDate(status.next_run_at)}</p>
				<p class="mt-1 text-xs text-slate-500">{status.schedule} · {status.timezone}</p>
			</div>

			<div class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
				<p class="text-xs font-semibold uppercase tracking-wide text-slate-500">Retensi & Ukuran</p>
				<p class="mt-3 text-lg font-bold text-slate-950">{status.retention_days} hari</p>
				<p class="mt-1 text-sm text-slate-600">{status.backup_count} file · {formatBytes(status.backup_dir_size_bytes)}</p>
			</div>
		</section>

		{#if status.warnings.length > 0}
			<section class="rounded-3xl border border-amber-200 bg-amber-50 p-5 text-sm text-amber-800">
				<p class="font-semibold">Perlu perhatian</p>
				<ul class="mt-2 list-disc space-y-1 pl-5">
					{#each status.warnings as warning}
						<li>{warning}</li>
					{/each}
				</ul>
			</section>
		{/if}

		<section class="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
			<div class="flex flex-col gap-2 border-b border-slate-200 p-5 md:flex-row md:items-center md:justify-between">
				<div>
					<h2 class="text-lg font-semibold text-slate-950">Daftar Backup</h2>
					<p class="text-sm text-slate-600">Hanya file <code>.dump</code> dari direktori backup resmi yang ditampilkan.</p>
				</div>
				{#if status.latest_backup && canDownload}
					<a class="inline-flex rounded-xl border border-slate-300 px-4 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-50" href={downloadHref(status.latest_backup.id)}>
						Download Latest
					</a>
				{/if}
			</div>

			{#if backups.length === 0}
				<div class="p-8 text-center text-sm text-slate-500">Belum ada file backup PostgreSQL yang tersedia.</div>
			{:else}
				<div class="overflow-x-auto">
					<table class="w-full min-w-[760px] text-left text-sm">
						<thead class="bg-slate-50 text-xs uppercase tracking-wide text-slate-500">
							<tr>
								<th class="px-5 py-3">File</th>
								<th class="px-5 py-3">Tanggal</th>
								<th class="px-5 py-3">Jenis</th>
								<th class="px-5 py-3">Ukuran</th>
								<th class="px-5 py-3">Checksum</th>
								<th class="px-5 py-3 text-right">Aksi</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-slate-100">
							{#each backups as backup}
								<tr class="align-top hover:bg-slate-50/80">
									<td class="px-5 py-4">
										<div class="font-medium text-slate-950">{backup.name}</div>
										{#if backup.is_latest}
											<span class="mt-1 inline-flex rounded-full bg-emerald-50 px-2 py-0.5 text-xs font-semibold text-emerald-700">latest</span>
										{/if}
									</td>
									<td class="px-5 py-4 text-slate-600">{formatDate(backup.created_at)}</td>
									<td class="px-5 py-4 text-slate-600">{kindLabel(backup.kind)}</td>
									<td class="px-5 py-4 text-slate-600">{formatBytes(backup.size_bytes)}</td>
									<td class="px-5 py-4 text-slate-600">
										{#if backup.sha256_available}
											<span class="font-mono text-xs">{backup.sha256?.slice(0, 12)}…</span>
										{:else}
											<span class="text-slate-400">Belum ada</span>
										{/if}
									</td>
									<td class="px-5 py-4 text-right">
										{#if canDownload && backup.downloadable}
											<a class="inline-flex rounded-xl border border-slate-300 px-3 py-1.5 text-xs font-semibold text-slate-700 hover:bg-white" href={downloadHref(backup.id)}>
												Download
											</a>
										{:else}
											<span class="text-xs text-slate-400">Butuh izin download</span>
										{/if}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</section>
	{/if}

	<section class="grid gap-4 lg:grid-cols-2">
		<div class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
			<h2 class="text-lg font-semibold text-slate-950">Restore Production</h2>
			<p class="mt-2 text-sm text-slate-600">
				Restore database production bersifat destruktif, sehingga tombol restore langsung belum dibuka pada Sprint Backup 1. Tahap berikutnya akan menambahkan validasi restore dan generate SOP/command manual yang aman.
			</p>
			<div class="mt-4 rounded-2xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
				Tidak ada aksi restore production dari halaman ini. Tombol backup manual hanya membuat file dump baru, bukan mengubah database.
			</div>
		</div>
		<div class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
			<h2 class="text-lg font-semibold text-slate-950">Backup Offsite</h2>
			<p class="mt-2 text-sm text-slate-600">
				Backup saat ini dipantau dari server lokal. Untuk ketahanan bencana, Sprint lanjutan direkomendasikan menambah sinkronisasi ke lokasi kedua seperti Google Drive, S3-compatible storage, NAS, atau server lain.
			</p>
			<div class="mt-4 rounded-2xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800">
				Rekomendasi: aktifkan monitoring offsite setelah Backup Center dan backup manual stabil.
			</div>
		</div>
	</section>
</div>
