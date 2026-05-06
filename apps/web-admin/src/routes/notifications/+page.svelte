<script lang="ts">
	import BellIcon from '@lucide/svelte/icons/bell';
	import CheckCircle2Icon from '@lucide/svelte/icons/check-circle-2';
	import ExternalLinkIcon from '@lucide/svelte/icons/external-link';
	import RefreshCcwIcon from '@lucide/svelte/icons/refresh-ccw';
	import { base, resolve } from '$app/paths';
	import { onMount } from 'svelte';
	import type { BadgeVariant } from '$lib/components/ui/badge';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { toast } from '$lib/components/ui/sonner';
	import { readClientJson } from '$lib/client/api';

	interface AppNotification {
		id: string;
		category: string;
		title: string;
		body: string;
		entity_type: string;
		entity_id: string;
		link_path: string;
		read_at: string | null;
		created_at: string;
	}

	interface NotificationOverview {
		items: AppNotification[];
		unread: number;
	}

	let unreadOnly = $state(false);
	let refreshBusy = $state(false);
	let markBusyId = $state('');
	let notificationsPromise = $state<Promise<NotificationOverview> | null>(null);
	let notificationsRequestId = 0;

	function refreshNotifications() {
		const requestId = ++notificationsRequestId;
		refreshBusy = true;
		const next = loadNotifications();
		notificationsPromise = next.finally(() => {
			if (requestId === notificationsRequestId) {
				refreshBusy = false;
			}
		});
	}

	async function loadNotifications(): Promise<NotificationOverview> {
		const params = new URLSearchParams({ limit: '60' });
		if (unreadOnly) params.set('unread_only', 'true');
		const [items, count] = await Promise.all([
			fetchJson<AppNotification[]>(`/api/notifications?${params.toString()}`),
			fetchJson<{ unread: number }>('/api/notifications/unread-count')
		]);
		return { items, unread: count.unread };
	}

	async function fetchJson<T>(path: string, init?: RequestInit): Promise<T> {
		const response = await fetch(`${base}${path}`, init);
		return readClientJson<T>(response);
	}

	async function markRead(id: string) {
		markBusyId = id;
		try {
			await fetchJson<AppNotification>(`/api/notifications/${id}/read`, { method: 'POST' });
			toast.success('Notifikasi ditandai sudah dibaca');
			refreshNotifications();
		} catch (error) {
			toast.error(errorMessage(error));
		} finally {
			markBusyId = '';
		}
	}

	async function markAllRead() {
		markBusyId = 'all';
		try {
			const result = await fetchJson<{ marked: number }>('/api/notifications/read-all', { method: 'POST' });
			toast.success(`${result.marked} notifikasi ditandai sudah dibaca`);
			refreshNotifications();
		} catch (error) {
			toast.error(errorMessage(error));
		} finally {
			markBusyId = '';
		}
	}

	function categoryLabel(category: string): string {
		if (category === 'document_cycle_late') return 'Lewat Tenggat';
		if (category === 'document_cycle_reminder') return 'Pengingat Dokumen';
		return 'Sistem';
	}

	function categoryVariant(category: string): BadgeVariant {
		if (category === 'document_cycle_late') return 'destructive';
		if (category === 'document_cycle_reminder') return 'secondary';
		return 'outline';
	}

	function formatDateTime(value?: string): string {
		if (!value) return '-';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value;
		return new Intl.DateTimeFormat('id-ID', {
			day: '2-digit',
			month: 'short',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
			timeZone: 'Asia/Makassar'
		}).format(date);
	}

	function internalHref(linkPath: string): string {
		if (!linkPath) return '';
		const [pathname, query = ''] = linkPath.split('?');
		const suffix = query ? `?${query}` : '';
		return `${resolve(pathname as '/')}${suffix}`;
	}

	function errorMessage(error: unknown): string {
		return error instanceof Error ? error.message : 'Operasi gagal diproses';
	}

	function handleRenderError(error: unknown) {
		console.error('Notifications render failed', error);
	}

	function retryNotifications(reset?: () => void) {
		reset?.();
		refreshNotifications();
	}

	onMount(() => {
		refreshNotifications();
	});
</script>

<svelte:head><title>Notifikasi - MTsN 2 Kolaka Utara</title></svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
		<div>
			<div class="mb-2 inline-flex items-center gap-2 rounded-full border border-primary/20 bg-primary/10 px-3 py-1 text-xs font-medium text-primary">
				<BellIcon class="size-3.5" />
				Pusat Notifikasi
			</div>
			<h1 class="text-xl font-semibold text-foreground">Notifikasi</h1>
			<p class="mt-1 max-w-3xl text-sm text-muted-foreground">Pengingat kerja dan status penting untuk akun ini.</p>
		</div>
		<div class="flex flex-wrap items-center gap-2">
			<label class="inline-flex h-9 items-center gap-2 rounded-md border border-input px-3 text-sm text-foreground">
				<input type="checkbox" bind:checked={unreadOnly} onchange={() => refreshNotifications()} class="size-4 accent-emerald-700" />
				Belum dibaca
			</label>
			<LoadingButton variant="outline" onclick={() => void refreshNotifications()} loading={refreshBusy} loadingLabel="Memuat">
				<RefreshCcwIcon class="mr-2 size-4" />
				Refresh
			</LoadingButton>
			<LoadingButton onclick={() => void markAllRead()} loading={markBusyId === 'all'} loadingLabel="Memproses">
				<CheckCircle2Icon class="mr-2 size-4" />
				Tandai Semua Dibaca
			</LoadingButton>
		</div>
	</div>

	<AsyncContent promise={notificationsPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<Card.Root class="border-border">
				<Card.Content class="space-y-3 p-4">
					{#each Array.from({ length: 5 }) as _, index (`notification-skeleton-${index}`)}
						<Skeleton class="h-20 w-full" />
					{/each}
				</Card.Content>
			</Card.Root>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel message={errorMessage(error)} onRetry={() => retryNotifications(reset)} />
		{/snippet}

		{#snippet children(value)}
			{@const overview = value as NotificationOverview}
			<div class="grid gap-3 md:grid-cols-2">
				<Card.Root class="border-border">
					<Card.Content class="p-4">
						<p class="text-xs text-muted-foreground">Belum Dibaca</p>
						<p class="mt-2 text-2xl font-semibold text-foreground">{overview.unread}</p>
					</Card.Content>
				</Card.Root>
				<Card.Root class="border-border">
					<Card.Content class="p-4">
						<p class="text-xs text-muted-foreground">Ditampilkan</p>
						<p class="mt-2 text-2xl font-semibold text-foreground">{overview.items.length}</p>
					</Card.Content>
				</Card.Root>
			</div>

			<Card.Root class="border-border">
				<Card.Header>
					<Card.Title class="text-base">Daftar Notifikasi</Card.Title>
					<Card.Description>Terurut dari yang terbaru.</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-3">
					{#each overview.items as item (item.id)}
						<div class={`rounded-md border p-4 ${item.read_at ? 'border-border bg-card' : 'border-primary/20 bg-primary/10'}`}>
							<div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
								<div class="min-w-0">
									<div class="flex flex-wrap items-center gap-2">
										<Badge variant={categoryVariant(item.category)}>{categoryLabel(item.category)}</Badge>
										<span class="text-xs text-muted-foreground">{formatDateTime(item.created_at)} WITA</span>
										{#if !item.read_at}
											<span class="rounded-full bg-primary/15 px-2 py-0.5 text-xs font-medium text-primary">Baru</span>
										{/if}
									</div>
									<p class="mt-2 text-sm font-semibold text-foreground">{item.title}</p>
									<p class="mt-1 text-sm text-muted-foreground">{item.body}</p>
								</div>
								<div class="flex shrink-0 flex-wrap gap-1.5">
									{#if item.link_path}
										<Button href={internalHref(item.link_path)} size="sm" variant="outline">
											<ExternalLinkIcon class="mr-2 size-3.5" />
											Buka
										</Button>
									{/if}
									<Button size="sm" variant="outline" disabled={Boolean(item.read_at) || markBusyId === item.id} onclick={() => void markRead(item.id)}>
										<CheckCircle2Icon class="mr-2 size-3.5" />
										Dibaca
									</Button>
								</div>
							</div>
						</div>
					{:else}
						<div class="rounded-md border border-dashed border-border bg-muted/50 p-6 text-center text-sm text-muted-foreground">
							Belum ada notifikasi pada filter ini.
						</div>
					{/each}
				</Card.Content>
			</Card.Root>
		{/snippet}
	</AsyncContent>
</div>
