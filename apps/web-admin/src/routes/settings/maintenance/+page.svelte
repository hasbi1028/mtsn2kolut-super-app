<script lang="ts">
	import { onMount } from 'svelte';
	import { maintenanceModuleLabel, maintenanceModules, type MaintenanceMode, type MaintenanceSeverity, type MaintenanceWindow } from '$lib/maintenance/modules';

	type MaintenanceListResponse = { items: MaintenanceWindow[] };
	type AuditLog = { id: string; action: string; reason?: string; created_at: string; actor_user_id?: string; maintenance_id?: string };
	type AuditResponse = { items: AuditLog[] };
	type ChecklistItem = { key: string; label: string; ok: boolean; severity: string; message: string };
	type HealthSummary = {
		core_api?: { status: string; server_time: string };
		database?: { connected: boolean; db_time?: string; message?: string };
		backup?: { available: boolean; health: string; backup_count: number; warnings: string[]; latest_backup_at?: string; last_run_at?: string; last_run_success?: boolean };
		disk?: { health: string; backup_dir_size_bytes: number; warnings: string[] };
		cbt?: { active_session_count: number; health: string; message: string };
		checklist?: ChecklistItem[];
	};

	const defaultMessage = 'Sistem sedang dalam maintenance. Silakan coba kembali beberapa saat lagi.';
	const defaultBypassRoles = ['admin'];

	let loading = $state(true);
	let saving = $state(false);
	let errorMessage = $state('');
	let successMessage = $state('');
	let windows = $state<MaintenanceWindow[]>([]);
	let auditLogs = $state<AuditLog[]>([]);
	let health = $state<HealthSummary | null>(null);
	let selectedModules = $state<string[]>(['global']);
	let form = $state({
		title: 'Maintenance Sistem',
		message: defaultMessage,
		mode: 'global' as MaintenanceMode,
		starts_at: '',
		ends_at: '',
		is_active: false,
		allow_admin_bypass: true,
		bypass_roles: defaultBypassRoles.join(', '),
		severity: 'warning' as MaintenanceSeverity,
		reason: ''
	});

	let activeWindow = $derived(windows.find((item) => item.is_active));
	let scheduledWindows = $derived(windows.filter((item) => item.status === 'scheduled'));

	onMount(() => {
		void reloadAll();
	});

	async function reloadAll() {
		loading = true;
		errorMessage = '';
		try {
			await Promise.all([loadWindows(), loadHealth(), loadAuditLogs()]);
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Gagal memuat data maintenance';
		} finally {
			loading = false;
		}
	}

	async function loadWindows() {
		const payload = await apiGet<MaintenanceListResponse>('/api/system/maintenance/windows');
		windows = payload.items ?? [];
	}

	async function loadHealth() {
		health = await apiGet<HealthSummary>('/api/system/maintenance/health-summary');
	}

	async function loadAuditLogs() {
		const payload = await apiGet<AuditResponse>('/api/system/maintenance/audit-logs?limit=30');
		auditLogs = payload.items ?? [];
	}

	async function createWindow() {
		saving = true;
		clearMessages();
		try {
			const payload = formPayload();
			const created = await apiJson<MaintenanceWindow>('/api/system/maintenance/windows', 'POST', payload);
			successMessage = created.is_active ? 'Maintenance aktif dibuat.' : 'Jadwal maintenance dibuat.';
			resetReason();
			await reloadAll();
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Gagal membuat maintenance';
		} finally {
			saving = false;
		}
	}

	async function activateWindow(id: string) {
		saving = true;
		clearMessages();
		try {
			await apiJson(`/api/system/maintenance/windows/${id}/activate`, 'POST', { reason: form.reason || 'Aktivasi dari Maintenance Center' });
			successMessage = 'Maintenance diaktifkan.';
			resetReason();
			await reloadAll();
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Gagal mengaktifkan maintenance';
		} finally {
			saving = false;
		}
	}

	async function deactivateWindow(id: string) {
		saving = true;
		clearMessages();
		try {
			await apiJson(`/api/system/maintenance/windows/${id}/deactivate`, 'POST', { reason: form.reason || 'Maintenance selesai' });
			successMessage = 'Maintenance dinonaktifkan.';
			resetReason();
			await reloadAll();
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Gagal menonaktifkan maintenance';
		} finally {
			saving = false;
		}
	}

	function formPayload() {
		return {
			title: form.title.trim(),
			message: form.message.trim(),
			mode: form.mode,
			affected_modules: form.mode === 'global' ? ['global'] : selectedModules.filter((item) => item !== 'global'),
			starts_at: form.starts_at ? new Date(form.starts_at).toISOString() : null,
			ends_at: form.ends_at ? new Date(form.ends_at).toISOString() : null,
			is_active: form.is_active,
			allow_admin_bypass: form.allow_admin_bypass,
			bypass_roles: form.bypass_roles.split(',').map((item) => item.trim()).filter(Boolean),
			severity: form.severity,
			reason: form.reason || 'Dibuat dari Maintenance Center'
		};
	}

	function toggleModule(code: string, checked: boolean) {
		if (checked) {
			selectedModules = Array.from(new Set([...selectedModules, code]));
			return;
		}
		selectedModules = selectedModules.filter((item) => item !== code);
	}

	function clearMessages() {
		errorMessage = '';
		successMessage = '';
	}

	function resetReason() {
		form.reason = '';
	}

	async function apiGet<T>(url: string): Promise<T> {
		const response = await fetch(url);
		return unwrap<T>(response);
	}

	async function apiJson<T = unknown>(url: string, method: string, body: unknown): Promise<T> {
		const response = await fetch(url, {
			method,
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(body)
		});
		return unwrap<T>(response);
	}

	async function unwrap<T>(response: Response): Promise<T> {
		const payload = await response.json().catch(() => ({}));
		if (!response.ok) {
			throw new Error(payload?.error ?? payload?.message ?? `HTTP ${response.status}`);
		}
		return (payload?.data ?? payload) as T;
	}

	function formatDate(value?: string) {
		if (!value) return '-';
		return new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium', timeStyle: 'short', timeZone: 'Asia/Makassar' }).format(new Date(value));
	}

	function statusBadge(status?: string) {
		if (status === 'active_now') return 'bg-red-100 text-red-700 border-red-200';
		if (status === 'scheduled') return 'bg-blue-100 text-blue-700 border-blue-200';
		if (status === 'ended') return 'bg-slate-100 text-slate-600 border-slate-200';
		return 'bg-emerald-100 text-emerald-700 border-emerald-200';
	}
</script>

<svelte:head>
	<title>Maintenance Center</title>
</svelte:head>

<div class="space-y-6">
	<section class="rounded-2xl border bg-card p-5 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
			<div>
				<p class="text-sm font-semibold uppercase tracking-wide text-muted-foreground">Pengaturan Sistem</p>
				<h1 class="text-2xl font-bold tracking-tight">Maintenance Center</h1>
				<p class="mt-2 max-w-3xl text-sm text-muted-foreground">
					Kendalikan maintenance global, module-level, read-only mode, health checklist, dan audit log tanpa membuka akses rahasia server.
				</p>
			</div>
			<button class="rounded-lg border px-4 py-2 text-sm font-semibold hover:bg-muted" onclick={reloadAll} disabled={loading}>Muat ulang</button>
		</div>
	</section>

	{#if errorMessage}
		<div class="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">{errorMessage}</div>
	{/if}
	{#if successMessage}
		<div class="rounded-xl border border-emerald-200 bg-emerald-50 p-4 text-sm text-emerald-700">{successMessage}</div>
	{/if}

	<section class="grid gap-4 md:grid-cols-4">
		<div class="rounded-2xl border bg-card p-4 shadow-sm">
			<p class="text-xs uppercase text-muted-foreground">Status</p>
			<p class={`mt-2 text-xl font-bold ${activeWindow ? 'text-red-600' : 'text-emerald-600'}`}>{activeWindow ? 'Aktif' : 'Normal'}</p>
		</div>
		<div class="rounded-2xl border bg-card p-4 shadow-sm">
			<p class="text-xs uppercase text-muted-foreground">Jadwal</p>
			<p class="mt-2 text-xl font-bold">{scheduledWindows.length}</p>
		</div>
		<div class="rounded-2xl border bg-card p-4 shadow-sm">
			<p class="text-xs uppercase text-muted-foreground">DB</p>
			<p class={`mt-2 text-xl font-bold ${health?.database?.connected ? 'text-emerald-600' : 'text-red-600'}`}>{health?.database?.connected ? 'Connected' : 'Unknown'}</p>
		</div>
		<div class="rounded-2xl border bg-card p-4 shadow-sm">
			<p class="text-xs uppercase text-muted-foreground">CBT aktif</p>
			<p class="mt-2 text-xl font-bold">{health?.cbt?.active_session_count ?? 0}</p>
		</div>
	</section>

	<section class="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
		<form class="space-y-4 rounded-2xl border bg-card p-5 shadow-sm" onsubmit={(event) => { event.preventDefault(); void createWindow(); }}>
			<div>
				<h2 class="text-lg font-semibold">Buat Maintenance Window</h2>
				<p class="text-sm text-muted-foreground">Untuk emergency lock, centang “aktifkan langsung”. Untuk maintenance terjadwal, isi waktu mulai dan selesai.</p>
			</div>
			<div class="grid gap-4 md:grid-cols-2">
				<label class="space-y-1 text-sm font-medium">Judul
					<input class="w-full rounded-lg border bg-background px-3 py-2" bind:value={form.title} required />
				</label>
				<label class="space-y-1 text-sm font-medium">Severity
					<select class="w-full rounded-lg border bg-background px-3 py-2" bind:value={form.severity}>
						<option value="info">Info</option>
						<option value="warning">Warning</option>
						<option value="critical">Critical</option>
					</select>
				</label>
			</div>
			<label class="space-y-1 text-sm font-medium">Pesan user
				<textarea class="min-h-24 w-full rounded-lg border bg-background px-3 py-2" bind:value={form.message} required></textarea>
			</label>
			<div class="grid gap-4 md:grid-cols-3">
				<label class="space-y-1 text-sm font-medium">Mode
					<select class="w-full rounded-lg border bg-background px-3 py-2" bind:value={form.mode}>
						<option value="global">Global</option>
						<option value="module">Per Modul</option>
						<option value="read_only">Read-only</option>
					</select>
				</label>
				<label class="space-y-1 text-sm font-medium">Mulai
					<input type="datetime-local" class="w-full rounded-lg border bg-background px-3 py-2" bind:value={form.starts_at} />
				</label>
				<label class="space-y-1 text-sm font-medium">Selesai
					<input type="datetime-local" class="w-full rounded-lg border bg-background px-3 py-2" bind:value={form.ends_at} />
				</label>
			</div>
			<div class="rounded-xl border p-4">
				<p class="mb-3 text-sm font-semibold">Modul terdampak</p>
				<div class="grid gap-2 sm:grid-cols-2">
					{#each maintenanceModules as module}
						<label class="flex items-center gap-2 text-sm">
							<input type="checkbox" checked={selectedModules.includes(module.code)} disabled={form.mode === 'global' && module.code !== 'global'} onchange={(event) => toggleModule(module.code, event.currentTarget.checked)} />
							<span>{module.label}</span>
						</label>
					{/each}
				</div>
			</div>
			<div class="grid gap-4 md:grid-cols-2">
				<label class="flex items-center gap-2 text-sm font-medium">
					<input type="checkbox" bind:checked={form.is_active} />
					Aktifkan langsung
				</label>
				<label class="flex items-center gap-2 text-sm font-medium">
					<input type="checkbox" bind:checked={form.allow_admin_bypass} />
					Admin bypass tetap boleh akses
				</label>
			</div>
			<label class="space-y-1 text-sm font-medium">Role bypass
				<input class="w-full rounded-lg border bg-background px-3 py-2" bind:value={form.bypass_roles} placeholder="admin, superadmin" />
			</label>
			<label class="space-y-1 text-sm font-medium">Alasan / catatan audit
				<input class="w-full rounded-lg border bg-background px-3 py-2" bind:value={form.reason} placeholder="Contoh: deploy modul CBT malam ini" />
			</label>
			<button class="rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground disabled:opacity-60" disabled={saving || loading}>Simpan Maintenance</button>
		</form>

		<div class="space-y-6">
			<section class="rounded-2xl border bg-card p-5 shadow-sm">
				<h2 class="text-lg font-semibold">Pre-maintenance Checklist</h2>
				<div class="mt-4 space-y-3">
					{#each health?.checklist ?? [] as item}
						<div class="rounded-xl border p-3">
							<div class="flex items-center justify-between gap-2">
								<p class="font-medium">{item.label}</p>
								<span class={`rounded-full px-2 py-0.5 text-xs font-semibold ${item.ok ? 'bg-emerald-100 text-emerald-700' : 'bg-amber-100 text-amber-700'}`}>{item.ok ? 'OK' : item.severity}</span>
							</div>
							<p class="mt-1 text-sm text-muted-foreground">{item.message}</p>
						</div>
					{:else}
						<p class="text-sm text-muted-foreground">Checklist belum tersedia.</p>
					{/each}
				</div>
			</section>
			<section class="rounded-2xl border bg-card p-5 shadow-sm">
				<h2 class="text-lg font-semibold">Health Summary</h2>
				<dl class="mt-4 space-y-2 text-sm">
					<div class="flex justify-between gap-4"><dt>Core API</dt><dd class="font-semibold">{health?.core_api?.status ?? '-'}</dd></div>
					<div class="flex justify-between gap-4"><dt>Backup health</dt><dd class="font-semibold">{health?.backup?.health ?? '-'}</dd></div>
					<div class="flex justify-between gap-4"><dt>Backup count</dt><dd class="font-semibold">{health?.backup?.backup_count ?? 0}</dd></div>
					<div class="flex justify-between gap-4"><dt>Disk backup dir</dt><dd class="font-semibold">{Math.round((health?.disk?.backup_dir_size_bytes ?? 0) / 1024 / 1024)} MB</dd></div>
				</dl>
			</section>
		</div>
	</section>

	<section class="rounded-2xl border bg-card p-5 shadow-sm">
		<h2 class="text-lg font-semibold">Maintenance Windows</h2>
		<div class="mt-4 overflow-x-auto">
			<table class="w-full min-w-[880px] text-left text-sm">
				<thead class="text-xs uppercase text-muted-foreground">
					<tr><th class="py-2">Judul</th><th>Mode</th><th>Modul</th><th>Mulai</th><th>Selesai</th><th>Status</th><th>Aksi</th></tr>
				</thead>
				<tbody class="divide-y">
					{#each windows as item}
						<tr>
							<td class="py-3 font-medium">{item.title}</td>
							<td>{item.mode.replace('_', ' ')}</td>
							<td>{item.affected_modules.map(maintenanceModuleLabel).join(', ') || '-'}</td>
							<td>{formatDate(item.starts_at)}</td>
							<td>{formatDate(item.ends_at)}</td>
							<td><span class={`rounded-full border px-2 py-1 text-xs font-semibold ${statusBadge(item.status)}`}>{item.status}</span></td>
							<td>
								{#if item.is_active}
									<button class="rounded-lg border border-red-200 px-3 py-1 text-xs font-semibold text-red-700 hover:bg-red-50" onclick={() => void deactivateWindow(item.id)} disabled={saving}>Nonaktifkan</button>
								{:else}
									<button class="rounded-lg border px-3 py-1 text-xs font-semibold hover:bg-muted" onclick={() => void activateWindow(item.id)} disabled={saving}>Aktifkan</button>
								{/if}
							</td>
						</tr>
					{:else}
						<tr><td colspan="7" class="py-6 text-center text-muted-foreground">Belum ada maintenance window.</td></tr>
					{/each}
				</tbody>
			</table>
		</div>
	</section>

	<section class="rounded-2xl border bg-card p-5 shadow-sm">
		<h2 class="text-lg font-semibold">Audit Log</h2>
		<div class="mt-4 space-y-3">
			{#each auditLogs as log}
				<div class="rounded-xl border p-3 text-sm">
					<div class="flex flex-wrap items-center justify-between gap-2">
						<p class="font-semibold capitalize">{log.action}</p>
						<p class="text-muted-foreground">{formatDate(log.created_at)}</p>
					</div>
					<p class="mt-1 text-muted-foreground">{log.reason || 'Tanpa alasan'}</p>
				</div>
			{:else}
				<p class="text-sm text-muted-foreground">Audit log belum tersedia.</p>
			{/each}
		</div>
	</section>
</div>
