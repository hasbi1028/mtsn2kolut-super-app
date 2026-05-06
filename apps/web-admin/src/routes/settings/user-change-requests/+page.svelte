<script lang="ts">
	import { onMount } from 'svelte';
	import CheckIcon from '@lucide/svelte/icons/check';
	import RefreshCcwIcon from '@lucide/svelte/icons/refresh-ccw';
	import XIcon from '@lucide/svelte/icons/x';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Textarea } from '$lib/components/ui/textarea';
	import { toast } from '$lib/components/ui/sonner';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { clientApiPath, clientApiPathWithQuery, readClientApiData } from '$lib/client/api';
	import {
		accountErrorMessage,
		changeRequestTimelineItems,
		changeRequestStatusLabel,
		formatAccountDateTime,
		officialFieldLabel,
		profileTypeLabel,
		type AccountChangeRequest
	} from '$lib/client/account';

	let requestsPromise = $state<Promise<AccountChangeRequest[]> | null>(null);
	let requests = $state<AccountChangeRequest[]>([]);
	let statusFilter = $state('pending');
	let reviewNotes = $state<Record<string, string>>({});
	let reviewBusy = $state<string | null>(null);
	let refreshBusy = $state(false);

	const pendingCount = $derived(requests.filter((request) => request.status === 'pending').length);

	async function fetchRequests(): Promise<AccountChangeRequest[]> {
		const params = new URLSearchParams();
		if (statusFilter) params.set('status', statusFilter);
		params.set('per_page', '100');
		const res = await fetch(clientApiPathWithQuery('/api/users/change-requests', params));
		return await readClientApiData<AccountChangeRequest[]>(res, 'Gagal memuat permintaan perubahan data');
	}

	function applyRequests(rows: AccountChangeRequest[]) {
		requests = rows;
		reviewNotes = Object.fromEntries(rows.map((request) => [request.id, request.review_note ?? '']));
		return rows;
	}

	function loadRequests() {
		requestsPromise = fetchRequests().then(applyRequests);
	}

	async function refreshRequests(showFailureToast = false) {
		refreshBusy = true;
		try {
			const rows = await fetchRequests();
			applyRequests(rows);
			requestsPromise = Promise.resolve(rows);
		} catch (error) {
			requestsPromise = Promise.resolve(requests);
			if (showFailureToast) toast.error(accountErrorMessage(error, 'Gagal menyegarkan permintaan'));
		} finally {
			refreshBusy = false;
		}
	}

	function retryRequests(reset?: () => void) {
		reset?.();
		loadRequests();
	}

	function handleRenderError(error: unknown, reset: () => void) {
		console.error('User change requests render failed', error);
		reset();
	}

	async function reviewRequest(request: AccountChangeRequest, status: 'approved' | 'rejected') {
		const note = (reviewNotes[request.id] ?? '').trim();
		if (status === 'rejected' && !note) {
			toast.error('Catatan wajib diisi saat menolak permintaan');
			return;
		}
		reviewBusy = `${status}:${request.id}`;
		try {
			const res = await fetch(clientApiPath`/api/users/change-requests/${request.id}`, {
				method: 'PATCH',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ status, review_note: note })
			});
			await readClientApiData<AccountChangeRequest>(res, 'Gagal memproses review permintaan');
			toast.success(status === 'approved' ? 'Permintaan disetujui' : 'Permintaan ditolak');
			await refreshRequests();
		} catch (error) {
			toast.error(accountErrorMessage(error, 'Gagal memproses review permintaan'));
		} finally {
			reviewBusy = null;
		}
	}

	function statusBadgeVariant(status: string) {
		if (status === 'approved') return 'secondary';
		if (status === 'rejected') return 'destructive';
		return 'outline';
	}

	onMount(() => {
		loadRequests();
	});
</script>

<svelte:head><title>Permintaan Perubahan Data - MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6 p-6">
	<div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
		<div>
			<h1 class="text-2xl font-semibold text-slate-800">Permintaan Perubahan Data Resmi</h1>
			<p class="mt-1 text-sm text-slate-500">Review perubahan identitas resmi yang diajukan dari halaman Akun Saya.</p>
		</div>
		<div class="flex flex-wrap gap-2">
			<Button variant="outline" href="/settings/users">Manajemen User</Button>
			<Button variant="outline" onclick={() => void refreshRequests(true)} disabled={refreshBusy}>
				<RefreshCcwIcon class={`size-4 ${refreshBusy ? 'animate-spin' : ''}`} />
				Refresh
			</Button>
		</div>
	</div>

	<Card.Root>
		<Card.Header class="flex flex-col gap-3 pb-3 md:flex-row md:items-center md:justify-between">
			<div>
				<Card.Title class="text-base">Antrian Review</Card.Title>
				<Card.Description>{pendingCount} permintaan menunggu dari filter saat ini.</Card.Description>
			</div>
			<div>
				<label for="change-request-status" class="mb-1 block text-xs font-medium text-slate-600">Status</label>
				<select
					id="change-request-status"
					class="h-10 rounded-md border border-input bg-background px-3 text-sm"
					bind:value={statusFilter}
					onchange={() => loadRequests()}
				>
					<option value="pending">Menunggu</option>
					<option value="all">Semua</option>
					<option value="approved">Disetujui</option>
					<option value="rejected">Ditolak</option>
					<option value="cancelled">Dibatalkan</option>
				</select>
			</div>
		</Card.Header>
		<Card.Content>
			<AsyncContent promise={requestsPromise} onerror={handleRenderError}>
				{#snippet pending()}
					<div class="space-y-3">
						{#each Array.from({ length: 4 }) as _, index (`change-request-skeleton-${index}`)}
							<div class="rounded-lg border border-slate-200 p-4">
								<Skeleton class="h-5 w-48" />
								<Skeleton class="mt-2 h-4 w-80" />
								<Skeleton class="mt-3 h-16 w-full" />
							</div>
						{/each}
					</div>
				{/snippet}

				{#snippet failed(error, reset)}
					<RecoveryPanel title="Permintaan Belum Tersaji" message={accountErrorMessage(error, 'Gagal memuat permintaan perubahan data')} onRetry={() => retryRequests(reset)} />
				{/snippet}

				{#snippet children(rows)}
					{@const items = rows as AccountChangeRequest[]}
					{#if items.length === 0}
						<p class="rounded-lg border border-dashed border-slate-200 px-4 py-8 text-center text-sm text-slate-500">Belum ada permintaan pada filter ini.</p>
					{:else}
						<div class="hidden overflow-x-auto lg:block">
							<Table.Root>
								<Table.Header>
									<Table.Row class="bg-slate-50">
										<Table.Head>Pemohon</Table.Head>
										<Table.Head>Profil</Table.Head>
										<Table.Head>Perubahan</Table.Head>
										<Table.Head>Status</Table.Head>
										<Table.Head>Timeline</Table.Head>
										<Table.Head>Catatan Review</Table.Head>
										<Table.Head>Aksi</Table.Head>
									</Table.Row>
								</Table.Header>
								<Table.Body>
									{#each items as request (request.id)}
										<Table.Row>
											<Table.Cell>
												<div class="font-medium text-slate-900">{request.requester_display_name || request.requester_username || 'Pengguna'}</div>
												<div class="text-xs text-slate-500">{request.requester_username || request.requester_user_id || '—'}</div>
											</Table.Cell>
											<Table.Cell>
												<div class="text-sm text-slate-700">{profileTypeLabel(request.profile_type)}</div>
												<div class="text-xs text-slate-500">{request.profile_nama || '—'}</div>
											</Table.Cell>
											<Table.Cell class="max-w-md">
												<div class="font-medium text-slate-900">{officialFieldLabel(request.field_key)}</div>
												<div class="mt-1 text-xs text-slate-500">{request.current_value || '—'} → {request.requested_value || '—'}</div>
												<div class="mt-1 text-xs text-slate-500">{request.reason}</div>
												<div class="mt-1 text-[11px] text-slate-400">{formatAccountDateTime(request.created_at)}</div>
											</Table.Cell>
											<Table.Cell><Badge variant={statusBadgeVariant(request.status)}>{changeRequestStatusLabel(request.status)}</Badge></Table.Cell>
											<Table.Cell class="min-w-56">
												<div class="space-y-2">
													{#each changeRequestTimelineItems(request) as timeline, index (`timeline-${request.id}-${index}`)}
														<div class="border-l border-slate-200 pl-3">
															<p class="text-xs font-medium text-slate-700">{timeline.label}</p>
															<p class="text-[11px] text-slate-500">
																{formatAccountDateTime(timeline.at)}
																{#if timeline.actor}
																	<span class="text-slate-300"> · </span>{timeline.actor}
																{/if}
															</p>
															{#if timeline.note}
																<p class="mt-1 text-[11px] text-slate-500">{timeline.note}</p>
															{/if}
														</div>
													{/each}
												</div>
											</Table.Cell>
											<Table.Cell class="min-w-64">
												<Textarea
													rows={3}
													maxlength={1000}
													bind:value={reviewNotes[request.id]}
													disabled={request.status !== 'pending'}
													placeholder="Catatan admin"
												/>
											</Table.Cell>
											<Table.Cell>
												{#if request.status === 'pending'}
													<div class="flex flex-col gap-2">
														<LoadingButton
															size="sm"
															onclick={() => void reviewRequest(request, 'approved')}
															loading={reviewBusy === `approved:${request.id}`}
															loadingLabel="Menyetujui..."
														>
															<CheckIcon class="size-3.5" />
															Setujui
														</LoadingButton>
														<LoadingButton
															size="sm"
															variant="outline"
															onclick={() => void reviewRequest(request, 'rejected')}
															loading={reviewBusy === `rejected:${request.id}`}
															loadingLabel="Menolak..."
														>
															<XIcon class="size-3.5" />
															Tolak
														</LoadingButton>
													</div>
												{:else}
													<p class="text-xs text-slate-500">{request.reviewer_username ? `Direview oleh ${request.reviewer_username}` : 'Sudah diproses'}</p>
												{/if}
											</Table.Cell>
										</Table.Row>
									{/each}
								</Table.Body>
							</Table.Root>
						</div>

						<div class="grid gap-3 lg:hidden">
							{#each items as request (request.id)}
								<div class="rounded-lg border border-slate-200 p-4">
									<div class="flex flex-wrap items-center justify-between gap-2">
										<div>
											<p class="font-medium text-slate-900">{request.requester_display_name || request.requester_username || 'Pengguna'}</p>
											<p class="text-xs text-slate-500">{profileTypeLabel(request.profile_type)} · {request.profile_nama || '—'}</p>
										</div>
										<Badge variant={statusBadgeVariant(request.status)}>{changeRequestStatusLabel(request.status)}</Badge>
									</div>
									<p class="mt-3 text-sm font-medium text-slate-900">{officialFieldLabel(request.field_key)}</p>
									<p class="mt-1 text-sm text-slate-600">{request.current_value || '—'} → {request.requested_value || '—'}</p>
									<p class="mt-1 text-xs text-slate-500">{request.reason}</p>
									<div class="mt-3 space-y-2">
										{#each changeRequestTimelineItems(request) as timeline, index (`mobile-timeline-${request.id}-${index}`)}
											<div class="border-l border-slate-200 pl-3">
												<p class="text-xs font-medium text-slate-700">{timeline.label}</p>
												<p class="text-[11px] text-slate-500">
													{formatAccountDateTime(timeline.at)}
													{#if timeline.actor}
														<span class="text-slate-300"> · </span>{timeline.actor}
													{/if}
												</p>
												{#if timeline.note}
													<p class="mt-1 text-[11px] text-slate-500">{timeline.note}</p>
												{/if}
											</div>
										{/each}
									</div>
									{#if request.status === 'pending'}
										<div class="mt-3 space-y-2">
											<label for={`mobile-note-${request.id}`} class="block text-xs font-medium text-slate-600">Catatan Review</label>
											<Textarea id={`mobile-note-${request.id}`} rows={3} maxlength={1000} bind:value={reviewNotes[request.id]} />
											<div class="flex flex-wrap gap-2">
												<LoadingButton size="sm" onclick={() => void reviewRequest(request, 'approved')} loading={reviewBusy === `approved:${request.id}`} loadingLabel="Menyetujui...">
													<CheckIcon class="size-3.5" />
													Setujui
												</LoadingButton>
												<LoadingButton size="sm" variant="outline" onclick={() => void reviewRequest(request, 'rejected')} loading={reviewBusy === `rejected:${request.id}`} loadingLabel="Menolak...">
													<XIcon class="size-3.5" />
													Tolak
												</LoadingButton>
											</div>
										</div>
									{/if}
								</div>
							{/each}
						</div>
					{/if}
				{/snippet}
			</AsyncContent>
		</Card.Content>
	</Card.Root>
</div>
