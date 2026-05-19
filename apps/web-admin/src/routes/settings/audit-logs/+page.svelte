<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { TablePagination } from '$lib/components/ui/pagination';
	import { readClientApiData } from '$lib/client/api';
	import { displayName } from '$lib/utils/display-name';
	import { DEFAULT_PAGE_SIZE_OPTIONS, type PaginationChange } from '$lib/utils/pagination';

	type AuditLog = {
		id: string;
		user_id: string | null;
		username: string | null;
		user_display_name?: string | null;
		action: string;
		entity_type: string;
		entity_id: string;
		metadata: unknown;
		created_at: string;
	};

	let logs = $state<AuditLog[]>([]);
	let logsPromise = $state<Promise<AuditLog[]> | null>(null);
	let logsRequestId = 0;
	let fetching = $state(false);
	let page = $state(1);
	let perPage = $state<number>(DEFAULT_PAGE_SIZE_OPTIONS[0]);
	let scopeFilter = $state<'all' | 'auth'>('all');
	let search = $state('');
	let searchTimer: ReturnType<typeof setTimeout> | null = null;

	const filteredLogs = $derived.by(() => filterAuditLogs(logs));
	const totalEstimate = $derived.by(() => {
		const offset = (page - 1) * perPage;
		const hasNext = logs.length >= perPage;
		return offset + filteredLogs.length + (hasNext ? perPage : 0);
	});

	function filterAuditLogs(logRows: AuditLog[]) {
		const q = search.trim().toLowerCase();
		return logRows.filter((log) => {
			if (scopeFilter === 'auth' && !isAuthLog(log)) return false;
			if (!q) return true;
			return [
				log.user_display_name ?? '',
				log.username ?? '',
				log.action,
				log.entity_type,
				log.entity_id,
				metaPath(log.metadata),
				authSummary(log)
			]
				.join(' ')
				.toLowerCase()
				.includes(q);
		});
	}

	function methodColor(action: string) {
		if (action.startsWith('AUTH_')) return 'bg-warning/15 text-warning border-warning/30';
		if (action === 'POST') return 'bg-primary/15 text-primary border-primary/20';
		if (action === 'PUT' || action === 'PATCH') return 'bg-accent text-accent-foreground border-accent';
		if (action === 'DELETE') return 'bg-destructive/15 text-destructive border-destructive/30';
		return 'bg-muted text-foreground border-border';
	}

	function fmtDt(iso: string) {
		if (!iso) return '—';
		return new Date(iso).toLocaleString('id-ID', {
			timeZone: 'Asia/Makassar',
			year: 'numeric', month: 'short', day: 'numeric',
			hour: '2-digit', minute: '2-digit', second: '2-digit',
		}) + ' WITA';
	}

	function isRecord(value: unknown): value is Record<string, unknown> {
		return typeof value === 'object' && value !== null;
	}

	function parseMetadata(meta: unknown): Record<string, unknown> | null {
		if (!meta) return null;
		try {
			const parsed = typeof meta === 'string' ? JSON.parse(meta) : meta;
			return isRecord(parsed) ? parsed : null;
		} catch {
			return null;
		}
	}

	function metaStatus(meta: unknown): number | null {
		const parsed = parseMetadata(meta);
		const status = parsed?.status;
		return typeof status === 'number' ? status : null;
	}

	function metaPath(meta: unknown): string {
		const parsed = parseMetadata(meta);
		const path = parsed?.path;
		return typeof path === 'string' ? path : '';
	}

	function isAuthLog(log: AuditLog): boolean {
		return log.action.startsWith('AUTH_') || log.entity_type === 'auth_session';
	}

	function authSummary(log: AuditLog): string {
		if (!isAuthLog(log)) return '';
		const parsed = parseMetadata(log.metadata);
		if (!parsed) return '';
		return [
			parsed.device_label,
			parsed.scope,
			parsed.username
		]
			.filter((value): value is string => typeof value === 'string' && value.trim().length > 0)
			.join(' • ');
	}

	function actorLabel(log: AuditLog) {
		return displayName({
			display_name: log.user_display_name,
			username: log.username
		}, 'Sistem');
	}

	function auditTargetLabel(log: AuditLog) {
		return authSummary(log) || metaPath(log.metadata) || log.entity_type || '—';
	}

	function auditTechnicalTitle(log: AuditLog) {
		const parts = [
			log.user_id ? `User ID internal: ${log.user_id}` : '',
			log.entity_id ? `Target ID internal: ${log.entity_id}` : ''
		].filter(Boolean);
		return parts.join('\n');
	}

	async function fetchLogs(pageNumber: number): Promise<AuditLog[]> {
		const res = await fetch(`/api/users/audit-logs?page=${pageNumber}&per_page=${perPage}`);
		return readClientApiData<AuditLog[]>(res, 'Gagal memuat riwayat aktivitas');
	}

	function load(pageNumber = page, nextPerPage = perPage) {
		const requestId = ++logsRequestId;
		page = pageNumber;
		perPage = nextPerPage;
		logs = [];
		fetching = true;
		logsPromise = fetchLogs(pageNumber)
			.then((nextLogs) => {
				if (requestId === logsRequestId) {
					logs = nextLogs ?? [];
				}
				return logs;
			})
			.catch((error: unknown) => {
				if (requestId === logsRequestId) throw error;
				return logs;
			})
			.finally(() => {
				if (requestId === logsRequestId) {
					fetching = false;
				}
			});
	}

	function setScopeFilter(nextScope: 'all' | 'auth') {
		scopeFilter = nextScope;
		load(1);
	}

	function handleSearchInput(event: Event) {
		search = (event.currentTarget as HTMLInputElement).value;
		if (searchTimer) clearTimeout(searchTimer);
		searchTimer = setTimeout(() => load(1), 250);
	}

	function handlePagination(change: PaginationChange) {
		load(change.reason === 'limit' ? 1 : change.page, change.limit);
	}

	function retryLogs(reset?: () => void) {
		reset?.();
		load(page);
	}

	function auditErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat riwayat aktivitas';
	}

	function handleLogsRenderError(error: unknown) {
		console.error('Audit logs render failed', error);
	}

	onMount(() => {
		void load();
		return () => {
			if (searchTimer) clearTimeout(searchTimer);
		};
	});
</script>

<svelte:head><title>Riwayat Aktivitas — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6 p-6">
	<div class="flex flex-col justify-between gap-4 lg:flex-row lg:items-center">
		<div>
			<h1 class="text-2xl font-bold text-primary">Riwayat Aktivitas</h1>
			<p class="text-sm text-muted-foreground mt-1">
				Riwayat semua perubahan data oleh pengguna sistem
			</p>
		</div>
		<div class="flex flex-wrap items-center gap-2">
			<Button variant={scopeFilter === 'all' ? 'default' : 'outline'} size="sm" onclick={() => setScopeFilter('all')}>
				Semua
			</Button>
			<Button variant={scopeFilter === 'auth' ? 'default' : 'outline'} size="sm" onclick={() => setScopeFilter('auth')}>
				Auth & Session
			</Button>
			<Input value={search} oninput={handleSearchInput} placeholder="Cari aksi, user, sesi..." class="w-full sm:w-64" />
		</div>
	</div>

	<Card.Root class="overflow-hidden border-border shadow-sm">
		<Card.Content class="p-0">
			<AsyncContent promise={logsPromise} onerror={handleLogsRenderError}>
				{#snippet pending()}
					<div class="space-y-3 p-4">
						<Skeleton class="h-10 w-full" />
						<Skeleton class="h-14 w-full" />
						<Skeleton class="h-14 w-full" />
						<Skeleton class="h-14 w-full" />
					</div>
				{/snippet}

				{#snippet failed(error, reset)}
					<div class="p-4">
						<RecoveryPanel
							compact
							title="Riwayat Aktivitas Belum Tersaji"
							message={auditErrorMessage(error)}
							onRetry={() => retryLogs(reset)}
						/>
					</div>
				{/snippet}

				{#snippet children(value)}
					{@const currentLogs = filterAuditLogs(value as AuditLog[])}
					{#if currentLogs.length === 0}
						<div class="p-8 text-center text-muted-foreground text-sm">Belum ada riwayat aktivitas.</div>
					{:else}
					<div class="hidden overflow-x-auto lg:block">
					<Table.Root>
						<Table.Header>
							<Table.Row class="bg-success/10">
								<Table.Head>Waktu</Table.Head>
								<Table.Head>Pengguna</Table.Head>
								<Table.Head>Aksi</Table.Head>
								<Table.Head>Entitas</Table.Head>
								<Table.Head>Path</Table.Head>
								<Table.Head class="text-center">Status</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each currentLogs as log (log.id)}
								<Table.Row class="hover:bg-success/10">
									<Table.Cell class="text-xs text-muted-foreground whitespace-nowrap">{fmtDt(log.created_at)}</Table.Cell>
									<Table.Cell>
										<div class="font-medium text-sm">{actorLabel(log)}</div>
										{#if log.username}
											<div class="mt-1 text-xs text-muted-foreground">@{log.username}</div>
										{/if}
									</Table.Cell>
									<Table.Cell>
										<Badge variant="outline" class="text-xs font-mono {methodColor(log.action)}">{log.action}</Badge>
									</Table.Cell>
									<Table.Cell class="text-sm">{log.entity_type}</Table.Cell>
									<Table.Cell class="max-w-[300px] truncate text-xs text-muted-foreground" title={auditTechnicalTitle(log) || auditTargetLabel(log)}>
										{auditTargetLabel(log)}
									</Table.Cell>
									<Table.Cell class="text-center text-xs font-mono">
										{metaStatus(log.metadata) ?? '—'}
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
					</div>

					<div class="grid gap-3 p-4 lg:hidden">
						{#each currentLogs as log (log.id)}
							<div class="rounded-2xl border border-border bg-card p-4 shadow-sm">
								<div class="flex items-start justify-between gap-3">
									<div class="min-w-0">
										<p class="text-xs font-semibold uppercase tracking-[0.16em] text-muted-foreground">{fmtDt(log.created_at)}</p>
										<p class="mt-1 text-sm font-semibold text-foreground">{actorLabel(log)}</p>
										{#if log.username}
											<p class="mt-1 text-xs text-muted-foreground">@{log.username}</p>
										{/if}
										<p class="mt-1 text-sm text-muted-foreground">{log.entity_type}</p>
									</div>
									<Badge variant="outline" class="text-xs font-mono {methodColor(log.action)}">{log.action}</Badge>
								</div>
								<p class="mt-3 rounded-xl bg-muted/50 px-3 py-2 text-xs font-mono text-muted-foreground break-all">
									{auditTargetLabel(log)}
								</p>
								<p class="mt-3 text-xs text-muted-foreground">Status {metaStatus(log.metadata) ?? '—'}</p>
							</div>
						{/each}
					</div>
					{/if}
				{/snippet}
			</AsyncContent>
		</Card.Content>
	</Card.Root>

	<div class="space-y-2">
		<p class="text-sm text-muted-foreground" aria-live="polite">
			Menampilkan {filteredLogs.length} aktivitas pada halaman {page}
		</p>
		<TablePagination
			{page}
			limit={perPage}
			total={totalEstimate}
			itemLabel="aktivitas"
			loading={fetching}
			showSummary={false}
			ariaLabel="Navigasi halaman riwayat aktivitas"
			onchange={handlePagination}
		/>
	</div>
</div>
