<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';

	type AuditLog = {
		id: string;
		user_id: string | null;
		username: string | null;
		action: string;
		entity_type: string;
		entity_id: string;
		metadata: unknown;
		created_at: string;
	};

	type ApiEnvelope<T> = {
		data?: T;
		error?: string;
		message?: string;
	};

	let logs = $state<AuditLog[]>([]);
	let logsPromise = $state<Promise<AuditLog[]> | null>(null);
	let fetching = $state(false);
	let fetchingAction = $state<'previous' | 'next' | null>(null);
	let page = $state(1);
	let scopeFilter = $state<'all' | 'auth'>('all');
	let search = $state('');
	const perPage = 50;

	const filteredLogs = $derived.by(() => filterAuditLogs(logs));

	function filterAuditLogs(logRows: AuditLog[]) {
		const q = search.trim().toLowerCase();
		return logRows.filter((log) => {
			if (scopeFilter === 'auth' && !isAuthLog(log)) return false;
			if (!q) return true;
			return [
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
		if (action.startsWith('AUTH_')) return 'bg-amber-100 text-amber-800 border-amber-200';
		if (action === 'POST') return 'bg-emerald-100 text-emerald-700 border-emerald-200';
		if (action === 'PUT' || action === 'PATCH') return 'bg-blue-100 text-blue-700 border-blue-200';
		if (action === 'DELETE') return 'bg-red-100 text-red-700 border-red-200';
		return 'bg-slate-100 text-slate-700 border-slate-200';
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

	function apiErrorMessage(payload: unknown) {
		if (!isRecord(payload)) return '';
		const error = payload.error;
		if (typeof error === 'string' && error.trim()) return error;
		const message = payload.message;
		if (typeof message === 'string' && message.trim()) return message;
		return '';
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
			parsed.revoked_session_id,
			parsed.renamed_session_id,
			parsed.username
		]
			.filter((value): value is string => typeof value === 'string' && value.trim().length > 0)
			.join(' • ');
	}

	async function readApi<T>(response: Response, fallbackMessage: string): Promise<T> {
		const payload = (await response.json().catch(() => null)) as ApiEnvelope<T> | T | null;
		const message = apiErrorMessage(payload);
		if (!response.ok) {
			throw new Error(message || fallbackMessage);
		}
		if (isRecord(payload) && typeof payload.error === 'string' && payload.error.trim()) {
			throw new Error(payload.error);
		}
		if (isRecord(payload) && 'data' in payload) {
			const envelope = payload as ApiEnvelope<T>;
			if (envelope.data === undefined) throw new Error(fallbackMessage);
			return envelope.data;
		}
		if (payload === null) throw new Error(fallbackMessage);
		return payload as T;
	}

	async function fetchLogs(pageNumber: number): Promise<AuditLog[]> {
		const res = await fetch(`/api/users/audit-logs?page=${pageNumber}&per_page=${perPage}`);
		return readApi<AuditLog[]>(res, 'Gagal memuat audit logs');
	}

	function load(pageNumber = page, action: 'previous' | 'next' | null = null) {
		page = pageNumber;
		logs = [];
		fetching = true;
		fetchingAction = action;
		logsPromise = fetchLogs(pageNumber)
			.then((nextLogs) => {
				logs = nextLogs ?? [];
				return logs;
			})
			.finally(() => {
				fetching = false;
				fetchingAction = null;
			});
	}

	function retryLogs(reset?: () => void) {
		reset?.();
		load(page);
	}

	function auditErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat audit logs';
	}

	function handleLogsRenderError(error: unknown) {
		console.error('Audit logs render failed', error);
	}

	onMount(() => {
		void load();
	});
</script>

<svelte:head><title>Audit Trail — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6 p-6">
	<div class="flex flex-col justify-between gap-4 lg:flex-row lg:items-center">
		<div>
			<h1 class="text-2xl font-bold text-[oklch(0.38_0.13_145)]">Audit Trail</h1>
			<p class="text-sm text-muted-foreground mt-1">
				Riwayat semua perubahan data oleh pengguna sistem
			</p>
		</div>
		<div class="flex flex-wrap items-center gap-2">
			<Button variant={scopeFilter === 'all' ? 'default' : 'outline'} size="sm" onclick={() => (scopeFilter = 'all')}>
				Semua
			</Button>
			<Button variant={scopeFilter === 'auth' ? 'default' : 'outline'} size="sm" onclick={() => (scopeFilter = 'auth')}>
				Auth & Session
			</Button>
			<Input bind:value={search} placeholder="Cari aksi, user, sesi..." class="w-full sm:w-64" />
			<LoadingButton
				variant="outline"
				size="sm"
				disabled={page === 1 || fetching}
				loading={fetchingAction === 'previous'}
				loadingLabel="Memuat..."
				onclick={() => load(Math.max(1, page - 1), 'previous')}
			>← Sebelumnya</LoadingButton>
			<span class="text-sm text-muted-foreground">Hal. {page}</span>
			<LoadingButton
				variant="outline"
				size="sm"
				disabled={fetching || logs.length < perPage}
				loading={fetchingAction === 'next'}
				loadingLabel="Memuat..."
				onclick={() => load(page + 1, 'next')}
			>Berikutnya →</LoadingButton>
		</div>
	</div>

	<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
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
							title="Audit Trail Belum Tersaji"
							message={auditErrorMessage(error)}
							onRetry={() => retryLogs(reset)}
						/>
					</div>
				{/snippet}

				{#snippet children(value)}
					{@const currentLogs = filterAuditLogs(value as AuditLog[])}
					{#if currentLogs.length === 0}
						<div class="p-8 text-center text-muted-foreground text-sm">Belum ada audit log.</div>
					{:else}
					<div class="hidden overflow-x-auto lg:block">
					<Table.Root>
						<Table.Header>
							<Table.Row class="bg-green-50">
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
								<Table.Row class="hover:bg-green-50/40">
									<Table.Cell class="text-xs text-muted-foreground whitespace-nowrap">{fmtDt(log.created_at)}</Table.Cell>
									<Table.Cell class="font-medium text-sm">{log.username ?? '—'}</Table.Cell>
									<Table.Cell>
										<Badge variant="outline" class="text-xs font-mono {methodColor(log.action)}">{log.action}</Badge>
									</Table.Cell>
									<Table.Cell class="text-sm">{log.entity_type}</Table.Cell>
									<Table.Cell class="text-xs text-muted-foreground font-mono truncate max-w-[300px]" title={authSummary(log) || metaPath(log.metadata) || log.entity_id}>
										{authSummary(log) || metaPath(log.metadata) || log.entity_id}
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
							<div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
								<div class="flex items-start justify-between gap-3">
									<div class="min-w-0">
										<p class="text-xs font-semibold uppercase tracking-[0.16em] text-slate-400">{fmtDt(log.created_at)}</p>
										<p class="mt-1 text-sm font-semibold text-slate-900">{log.username ?? '—'}</p>
										<p class="mt-1 text-sm text-slate-600">{log.entity_type}</p>
									</div>
									<Badge variant="outline" class="text-xs font-mono {methodColor(log.action)}">{log.action}</Badge>
								</div>
								<p class="mt-3 rounded-xl bg-slate-50 px-3 py-2 text-xs font-mono text-slate-600 break-all">
									{authSummary(log) || metaPath(log.metadata) || log.entity_id}
								</p>
								<p class="mt-3 text-xs text-slate-500">Status {metaStatus(log.metadata) ?? '—'}</p>
							</div>
						{/each}
					</div>
					{/if}
				{/snippet}
			</AsyncContent>
		</Card.Content>
	</Card.Root>
</div>
