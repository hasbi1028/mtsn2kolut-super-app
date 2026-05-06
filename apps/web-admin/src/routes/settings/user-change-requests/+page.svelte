<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import CheckIcon from '@lucide/svelte/icons/check';
	import ChevronDownIcon from '@lucide/svelte/icons/chevron-down';
	import ChevronUpIcon from '@lucide/svelte/icons/chevron-up';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import RefreshCcwIcon from '@lucide/svelte/icons/refresh-ccw';
	import SearchIcon from '@lucide/svelte/icons/search';
	import XIcon from '@lucide/svelte/icons/x';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
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

	type ReviewStatus = 'approved' | 'rejected';

	type PendingCountPayload = {
		pending?: number;
		count?: number;
	};

	let requestsPromise = $state<Promise<AccountChangeRequest[]> | null>(null);
	let requests = $state<AccountChangeRequest[]>([]);
	let statusFilter = $state('pending');
	let profileFilter = $state('all');
	let fieldFilter = $state('all');
	let searchDraft = $state('');
	let appliedSearch = $state('');
	let pendingSummary = $state(0);
	let pendingSummaryBusy = $state(false);
	let reviewNotes = $state<Record<string, string>>({});
	let stagedReviews = $state<Record<string, ReviewStatus | undefined>>({});
	let expandedRows = $state<Record<string, boolean>>({});
	let reviewBusy = $state<string | null>(null);
	let refreshBusy = $state(false);
	let exportBusy = $state(false);

	const canReviewProfileChanges = $derived(Boolean(
		page.data.user?.role === 'admin'
		|| page.data.user?.roles?.includes('admin')
		|| page.data.user?.permissions?.includes('profile_changes.review')
	));
	const visiblePendingCount = $derived(requests.filter((request) => request.status === 'pending').length);
	const activeFilterCount = $derived([
		statusFilter !== 'pending',
		profileFilter !== 'all',
		fieldFilter !== 'all',
		appliedSearch.trim() !== ''
	].filter(Boolean).length);

	const profileOptions = [
		{ value: 'all', label: 'Semua profil' },
		{ value: 'employee', label: 'Pegawai' },
		{ value: 'student', label: 'Siswa' },
		{ value: 'parent', label: 'Orang Tua' }
	];

	const fieldOptions = [
		{ value: 'all', label: 'Semua field' },
		{ value: 'nama', label: 'Nama resmi' },
		{ value: 'tanggal_lahir', label: 'Tanggal lahir' },
		{ value: 'parent_name', label: 'Nama orang tua/wali' }
	];

	function buildFilterParams(includePaging = true) {
		const params = new URLSearchParams();
		if (statusFilter && statusFilter !== 'all') params.set('status', statusFilter);
		if (profileFilter && profileFilter !== 'all') params.set('profile_type', profileFilter);
		if (fieldFilter && fieldFilter !== 'all') params.set('field', fieldFilter);
		if (appliedSearch.trim()) params.set('search', appliedSearch.trim());
		if (includePaging) params.set('per_page', '100');
		return params;
	}

	async function fetchRequests(): Promise<AccountChangeRequest[]> {
		const res = await fetch(clientApiPathWithQuery('/api/users/change-requests', buildFilterParams()));
		return await readClientApiData<AccountChangeRequest[]>(res, 'Gagal memuat permintaan perubahan data');
	}

	async function fetchPendingSummary() {
		if (!canReviewProfileChanges) return;
		pendingSummaryBusy = true;
		try {
			const res = await fetch(clientApiPathWithQuery('/api/users/change-requests/pending-count', new URLSearchParams()));
			const payload = await readClientApiData<PendingCountPayload>(res, 'Gagal memuat jumlah permintaan');
			const next = Number(payload.pending ?? payload.count ?? 0);
			pendingSummary = Number.isFinite(next) ? next : 0;
		} catch {
			pendingSummary = 0;
		} finally {
			pendingSummaryBusy = false;
		}
	}

	function applyRequests(rows: AccountChangeRequest[]) {
		requests = rows;
		reviewNotes = Object.fromEntries(rows.map((request) => [request.id, request.review_note ?? '']));
		stagedReviews = {};
		return rows;
	}

	function loadRequests() {
		requestsPromise = fetchRequests().then(applyRequests);
		void fetchPendingSummary();
	}

	async function refreshRequests(showFailureToast = false) {
		refreshBusy = true;
		try {
			const rows = await fetchRequests();
			applyRequests(rows);
			requestsPromise = Promise.resolve(rows);
			await fetchPendingSummary();
		} catch (error) {
			requestsPromise = Promise.resolve(requests);
			if (showFailureToast) toast.error(accountErrorMessage(error, 'Gagal menyegarkan permintaan'));
		} finally {
			refreshBusy = false;
		}
	}

	function applyFilters() {
		appliedSearch = searchDraft.trim();
		loadRequests();
	}

	function resetFilters() {
		statusFilter = 'pending';
		profileFilter = 'all';
		fieldFilter = 'all';
		searchDraft = '';
		appliedSearch = '';
		loadRequests();
	}

	function retryRequests(reset?: () => void) {
		reset?.();
		loadRequests();
	}

	function handleRenderError(error: unknown, reset: () => void) {
		console.error('User change requests render failed', error);
		reset();
	}

	function stageReview(request: AccountChangeRequest, status: ReviewStatus) {
		const note = (reviewNotes[request.id] ?? '').trim();
		if (status === 'rejected' && !note) {
			toast.error('Catatan wajib diisi saat menolak permintaan');
			return;
		}
		stagedReviews = { ...stagedReviews, [request.id]: status };
	}

	function clearStagedReview(requestID: string) {
		stagedReviews = { ...stagedReviews, [requestID]: undefined };
	}

	function stagedReviewStatus(requestID: string): ReviewStatus {
		return stagedReviews[requestID] === 'rejected' ? 'rejected' : 'approved';
	}

	async function reviewRequest(request: AccountChangeRequest, status: ReviewStatus) {
		if (!canReviewProfileChanges) {
			toast.error('Akun ini tidak memiliki izin review perubahan data');
			return;
		}
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
			clearStagedReview(request.id);
		}
	}

	async function exportRequests() {
		exportBusy = true;
		try {
			const res = await fetch(clientApiPathWithQuery('/api/users/change-requests/export', buildFilterParams(false)));
			if (!res.ok) {
				const payload = await res.json().catch(() => null) as { error?: string; message?: string } | null;
				throw new Error(payload?.error || payload?.message || 'Export permintaan perubahan data gagal');
			}
			const blob = await res.blob();
			const url = URL.createObjectURL(blob);
			const anchor = document.createElement('a');
			anchor.href = url;
			anchor.download = exportFilename(res.headers.get('content-disposition'));
			document.body.appendChild(anchor);
			anchor.click();
			anchor.remove();
			URL.revokeObjectURL(url);
			toast.success('Export CSV disiapkan');
		} catch (error) {
			toast.error(accountErrorMessage(error, 'Export permintaan perubahan data gagal'));
		} finally {
			exportBusy = false;
		}
	}

	function exportFilename(disposition: string | null) {
		const match = /filename="?([^";]+)"?/i.exec(disposition ?? '');
		return match?.[1] ?? 'profile-change-requests.csv';
	}

	function statusBadgeVariant(status: string) {
		if (status === 'approved') return 'secondary';
		if (status === 'rejected') return 'destructive';
		return 'outline';
	}

	function requestFieldLabel(request: AccountChangeRequest) {
		return request.field_label?.trim() || officialFieldLabel(request.field_key);
	}

	function valueLabel(value: string | null | undefined) {
		return value?.trim() || '—';
	}

	function requesterLabel(request: AccountChangeRequest) {
		return request.requester_display_name || request.requester_username || request.requester_user_id || 'Pengguna';
	}

	function targetIDLabel(request: AccountChangeRequest) {
		return request.target_employee_id || request.target_student_id || request.target_parent_id || '—';
	}

	function toggleExpanded(requestID: string) {
		expandedRows = { ...expandedRows, [requestID]: !expandedRows[requestID] };
	}

	onMount(() => {
		loadRequests();
	});
</script>

<svelte:head><title>Permintaan Perubahan Data - MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6 p-6">
	<div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
		<div>
			<h1 class="text-2xl font-semibold text-foreground">Permintaan Perubahan Data Resmi</h1>
			<p class="mt-1 text-sm text-muted-foreground">Review perubahan identitas resmi yang diajukan dari halaman Akun Saya.</p>
		</div>
		<div class="flex flex-wrap gap-2">
			<Button variant="outline" href="/settings/users">Manajemen User</Button>
			<LoadingButton variant="outline" onclick={() => void exportRequests()} loading={exportBusy} loadingLabel="Export..." label="Export CSV">
				<DownloadIcon class="size-4" />
				Export CSV
			</LoadingButton>
			<Button variant="outline" onclick={() => void refreshRequests(true)} disabled={refreshBusy}>
				<RefreshCcwIcon class={`size-4 ${refreshBusy ? 'animate-spin' : ''}`} />
				Refresh
			</Button>
		</div>
	</div>

	<div class="grid gap-3 md:grid-cols-3">
		<Card.Root>
			<Card.Header class="pb-2">
				<Card.Description>Menunggu Review</Card.Description>
				<Card.Title class="text-2xl text-warning">
					{#if pendingSummaryBusy}
						<span class="text-base text-muted-foreground">Memuat...</span>
					{:else}
						{pendingSummary}
					{/if}
				</Card.Title>
			</Card.Header>
		</Card.Root>
		<Card.Root>
			<Card.Header class="pb-2">
				<Card.Description>Tampil di Filter Ini</Card.Description>
				<Card.Title class="text-2xl text-foreground">{requests.length}</Card.Title>
			</Card.Header>
		</Card.Root>
		<Card.Root>
			<Card.Header class="pb-2">
				<Card.Description>Pending di Tabel Ini</Card.Description>
				<Card.Title class="text-2xl text-foreground">{visiblePendingCount}</Card.Title>
			</Card.Header>
		</Card.Root>
	</div>

	<Card.Root>
		<Card.Header class="space-y-4 pb-3">
			<div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
				<div>
					<Card.Title class="text-base">Antrian Review</Card.Title>
					<Card.Description>
						{activeFilterCount > 0 ? `${activeFilterCount} filter aktif.` : 'Menampilkan permintaan yang menunggu review.'}
					</Card.Description>
				</div>
				{#if !canReviewProfileChanges}
					<Badge variant="destructive">Tidak punya izin review</Badge>
				{/if}
			</div>

			<form class="grid gap-3 lg:grid-cols-[minmax(130px,0.8fr)_minmax(140px,0.8fr)_minmax(150px,0.9fr)_minmax(220px,1.4fr)_auto]" onsubmit={(event) => { event.preventDefault(); applyFilters(); }}>
				<div>
					<label for="change-request-status" class="mb-1 block text-xs font-medium text-muted-foreground">Status</label>
					<select
						id="change-request-status"
						class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm"
						bind:value={statusFilter}
						onchange={applyFilters}
					>
						<option value="pending">Menunggu</option>
						<option value="all">Semua</option>
						<option value="approved">Disetujui</option>
						<option value="rejected">Ditolak</option>
						<option value="cancelled">Dibatalkan</option>
					</select>
				</div>
				<div>
					<label for="change-request-profile" class="mb-1 block text-xs font-medium text-muted-foreground">Profil</label>
					<select
						id="change-request-profile"
						class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm"
						bind:value={profileFilter}
						onchange={applyFilters}
					>
						{#each profileOptions as option (option.value)}
							<option value={option.value}>{option.label}</option>
						{/each}
					</select>
				</div>
				<div>
					<label for="change-request-field" class="mb-1 block text-xs font-medium text-muted-foreground">Field</label>
					<select
						id="change-request-field"
						class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm"
						bind:value={fieldFilter}
						onchange={applyFilters}
					>
						{#each fieldOptions as option (option.value)}
							<option value={option.value}>{option.label}</option>
						{/each}
					</select>
				</div>
				<div>
					<label for="change-request-search" class="mb-1 block text-xs font-medium text-muted-foreground">Cari</label>
					<Input id="change-request-search" bind:value={searchDraft} maxlength={120} placeholder="Nama, username, alasan, atau catatan" />
				</div>
				<div class="flex items-end gap-2">
					<Button type="submit" variant="secondary">
						<SearchIcon class="size-4" />
						Cari
					</Button>
					<Button type="button" variant="outline" onclick={resetFilters}>Reset</Button>
				</div>
			</form>
		</Card.Header>
		<Card.Content>
			<AsyncContent promise={requestsPromise} onerror={handleRenderError}>
				{#snippet pending()}
					<div class="space-y-3">
						{#each Array.from({ length: 4 }) as _, index (`change-request-skeleton-${index}`)}
							<div class="rounded-lg border border-border p-4">
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
						<div class="rounded-lg border border-dashed border-border px-4 py-8 text-center">
							<p class="text-sm font-medium text-foreground">Belum ada permintaan pada filter ini.</p>
							<p class="mt-1 text-xs text-muted-foreground">Ubah filter atau reset untuk melihat riwayat lain.</p>
							<Button class="mt-4" type="button" variant="outline" onclick={resetFilters}>Reset Filter</Button>
						</div>
					{:else}
						<div class="hidden overflow-x-auto lg:block">
							<Table.Root>
								<Table.Header>
									<Table.Row class="bg-muted/50">
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
												<div class="font-medium text-foreground">{requesterLabel(request)}</div>
												<div class="text-xs text-muted-foreground">{request.requester_username || request.requester_user_id || '—'}</div>
											</Table.Cell>
											<Table.Cell>
												<div class="text-sm text-foreground">{profileTypeLabel(request.profile_type)}</div>
												<div class="text-xs text-muted-foreground">{request.profile_nama || '—'}</div>
											</Table.Cell>
											<Table.Cell class="max-w-md">
												<div class="font-medium text-foreground">{requestFieldLabel(request)}</div>
												<div class="mt-1 grid grid-cols-[90px_1fr] gap-1 text-xs text-muted-foreground">
													<span>Saat ini</span><span class="break-words">{valueLabel(request.current_value)}</span>
													<span>Diajukan</span><span class="break-words font-medium text-foreground">{valueLabel(request.requested_value)}</span>
												</div>
												<div class="mt-2 text-xs text-muted-foreground">{request.reason}</div>
											</Table.Cell>
											<Table.Cell><Badge variant={statusBadgeVariant(request.status)}>{changeRequestStatusLabel(request.status)}</Badge></Table.Cell>
											<Table.Cell class="min-w-56">
												<div class="space-y-2">
													{#each changeRequestTimelineItems(request) as timeline, index (`timeline-${request.id}-${index}`)}
														<div class="border-l border-border pl-3">
															<p class="text-xs font-medium text-foreground">{timeline.label}</p>
															<p class="text-[11px] text-muted-foreground">
																{formatAccountDateTime(timeline.at)}
																{#if timeline.actor}
																	<span class="text-muted-foreground"> · </span>{timeline.actor}
																{/if}
															</p>
															{#if timeline.note}
																<p class="mt-1 text-[11px] text-muted-foreground">{timeline.note}</p>
															{/if}
														</div>
													{/each}
												</div>
											</Table.Cell>
											<Table.Cell class="min-w-64">
												<label for={`review-note-${request.id}`} class="mb-1 block text-xs font-medium text-muted-foreground">Catatan admin</label>
												<Textarea
													id={`review-note-${request.id}`}
													rows={3}
													maxlength={1000}
													bind:value={reviewNotes[request.id]}
													disabled={request.status !== 'pending' || !canReviewProfileChanges}
													placeholder="Catatan admin"
												/>
											</Table.Cell>
											<Table.Cell>
												<div class="flex flex-col gap-2">
													<Button type="button" size="sm" variant="outline" onclick={() => toggleExpanded(request.id)}>
														{#if expandedRows[request.id]}
															<ChevronUpIcon class="size-3.5" />
															Tutup
														{:else}
															<ChevronDownIcon class="size-3.5" />
															Detail
														{/if}
													</Button>
													{#if request.status === 'pending' && canReviewProfileChanges}
														<div class="flex flex-col gap-2">
															<Button size="sm" type="button" onclick={() => stageReview(request, 'approved')}>
																<CheckIcon class="size-3.5" />
																Setujui
															</Button>
															<Button size="sm" type="button" variant="outline" onclick={() => stageReview(request, 'rejected')}>
																<XIcon class="size-3.5" />
																Tolak
															</Button>
														</div>
														{#if stagedReviews[request.id]}
															<div class="rounded-md border border-warning/30 bg-warning/10 p-2 text-xs text-warning">
																<p>Konfirmasi {stagedReviews[request.id] === 'approved' ? 'persetujuan' : 'penolakan'} request ini.</p>
																<div class="mt-2 flex flex-wrap gap-2">
																	<LoadingButton
																		size="sm"
																		onclick={() => void reviewRequest(request, stagedReviewStatus(request.id))}
																		loading={reviewBusy === `${stagedReviews[request.id]}:${request.id}`}
																		loadingLabel="Memproses..."
																		label="Konfirmasi"
																	/>
																	<Button type="button" size="sm" variant="outline" onclick={() => clearStagedReview(request.id)}>Batal</Button>
																</div>
															</div>
														{/if}
													{:else if request.status !== 'pending'}
														<p class="text-xs text-muted-foreground">{request.reviewer_username ? `Direview oleh ${request.reviewer_username}` : 'Sudah diproses'}</p>
													{/if}
												</div>
											</Table.Cell>
										</Table.Row>
										{#if expandedRows[request.id]}
											<Table.Row class="bg-muted/50">
												<Table.Cell colspan={7}>
													<div class="grid gap-4 p-3 md:grid-cols-3">
														<div>
															<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Identitas Pemohon</p>
															<p class="mt-1 text-sm font-medium text-foreground">{requesterLabel(request)}</p>
															<p class="text-xs text-muted-foreground">Username: {request.requester_username || '—'}</p>
															<p class="text-xs text-muted-foreground">User ID: {request.requester_user_id || '—'}</p>
														</div>
														<div>
															<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Profil dan Field</p>
															<p class="mt-1 text-sm font-medium text-foreground">{profileTypeLabel(request.profile_type)} · {request.profile_nama || '—'}</p>
															<p class="text-xs text-muted-foreground">Target ID: {targetIDLabel(request)}</p>
															<p class="text-xs text-muted-foreground">Field: {requestFieldLabel(request)} ({request.field_key})</p>
														</div>
														<div>
															<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Status Review</p>
															<p class="mt-1 text-sm font-medium text-foreground">{changeRequestStatusLabel(request.status)}</p>
															<p class="text-xs text-muted-foreground">Reviewer: {request.reviewer_username || '—'}</p>
															<p class="text-xs text-muted-foreground">Catatan: {request.review_note || '—'}</p>
														</div>
														<div class="md:col-span-2">
															<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Nilai</p>
															<div class="mt-2 grid gap-2 md:grid-cols-2">
																<div class="rounded-md border border-border bg-card p-3">
																	<p class="text-xs text-muted-foreground">Current value</p>
																	<p class="mt-1 break-words text-sm text-foreground">{valueLabel(request.current_value)}</p>
																</div>
																<div class="rounded-md border border-primary/20 bg-card p-3">
																	<p class="text-xs text-muted-foreground">Requested value</p>
																	<p class="mt-1 break-words text-sm font-medium text-primary">{valueLabel(request.requested_value)}</p>
																</div>
															</div>
															<p class="mt-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">Alasan</p>
															<p class="mt-1 whitespace-pre-wrap break-words text-sm text-foreground">{request.reason || '—'}</p>
														</div>
														<div>
															<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Timestamps</p>
															<div class="mt-1 space-y-1 text-xs text-muted-foreground">
																<p>Diajukan: {formatAccountDateTime(request.created_at)}</p>
																<p>Direview: {formatAccountDateTime(request.reviewed_at)}</p>
																<p>Diperbarui: {formatAccountDateTime(request.updated_at)}</p>
															</div>
														</div>
													</div>
												</Table.Cell>
											</Table.Row>
										{/if}
									{/each}
								</Table.Body>
							</Table.Root>
						</div>

						<div class="grid gap-3 lg:hidden">
							{#each items as request (request.id)}
								<div class="rounded-lg border border-border p-4">
									<div class="flex flex-wrap items-center justify-between gap-2">
										<div>
											<p class="font-medium text-foreground">{requesterLabel(request)}</p>
											<p class="text-xs text-muted-foreground">{profileTypeLabel(request.profile_type)} · {request.profile_nama || '—'}</p>
										</div>
										<Badge variant={statusBadgeVariant(request.status)}>{changeRequestStatusLabel(request.status)}</Badge>
									</div>
									<p class="mt-3 text-sm font-medium text-foreground">{requestFieldLabel(request)}</p>
									<div class="mt-2 grid grid-cols-[88px_1fr] gap-1 text-sm text-muted-foreground">
										<span>Saat ini</span><span class="break-words">{valueLabel(request.current_value)}</span>
										<span>Diajukan</span><span class="break-words font-medium text-foreground">{valueLabel(request.requested_value)}</span>
									</div>
									<p class="mt-2 whitespace-pre-wrap break-words text-xs text-muted-foreground">{request.reason}</p>

									<Button class="mt-3" type="button" size="sm" variant="outline" onclick={() => toggleExpanded(request.id)}>
										{#if expandedRows[request.id]}
											<ChevronUpIcon class="size-3.5" />
											Tutup Detail
										{:else}
											<ChevronDownIcon class="size-3.5" />
											Detail
										{/if}
									</Button>

									{#if expandedRows[request.id]}
										<div class="mt-3 space-y-3 rounded-md bg-muted/50 p-3">
											<div>
												<p class="text-xs font-semibold text-muted-foreground">Pemohon</p>
												<p class="text-sm text-foreground">{requesterLabel(request)}</p>
												<p class="text-xs text-muted-foreground">{request.requester_username || request.requester_user_id || '—'}</p>
											</div>
											<div>
												<p class="text-xs font-semibold text-muted-foreground">Profil dan Target</p>
												<p class="text-sm text-foreground">{profileTypeLabel(request.profile_type)} · {request.profile_nama || '—'}</p>
												<p class="text-xs text-muted-foreground">Target ID: {targetIDLabel(request)}</p>
											</div>
											<div>
												<p class="text-xs font-semibold text-muted-foreground">Review</p>
												<p class="text-sm text-foreground">{request.reviewer_username || '—'}</p>
												<p class="text-xs text-muted-foreground">{request.review_note || 'Belum ada catatan'}</p>
											</div>
											<div class="text-xs text-muted-foreground">
												<p>Diajukan: {formatAccountDateTime(request.created_at)}</p>
												<p>Direview: {formatAccountDateTime(request.reviewed_at)}</p>
												<p>Diperbarui: {formatAccountDateTime(request.updated_at)}</p>
											</div>
										</div>
									{/if}

									<div class="mt-3 space-y-2">
										{#each changeRequestTimelineItems(request) as timeline, index (`mobile-timeline-${request.id}-${index}`)}
											<div class="border-l border-border pl-3">
												<p class="text-xs font-medium text-foreground">{timeline.label}</p>
												<p class="text-[11px] text-muted-foreground">
													{formatAccountDateTime(timeline.at)}
													{#if timeline.actor}
														<span class="text-muted-foreground"> · </span>{timeline.actor}
													{/if}
												</p>
												{#if timeline.note}
													<p class="mt-1 text-[11px] text-muted-foreground">{timeline.note}</p>
												{/if}
											</div>
										{/each}
									</div>
									{#if request.status === 'pending' && canReviewProfileChanges}
										<div class="mt-3 space-y-2">
											<label for={`mobile-note-${request.id}`} class="block text-xs font-medium text-muted-foreground">Catatan Review</label>
											<Textarea id={`mobile-note-${request.id}`} rows={3} maxlength={1000} bind:value={reviewNotes[request.id]} />
											<div class="flex flex-wrap gap-2">
												<Button size="sm" type="button" onclick={() => stageReview(request, 'approved')}>
													<CheckIcon class="size-3.5" />
													Setujui
												</Button>
												<Button size="sm" type="button" variant="outline" onclick={() => stageReview(request, 'rejected')}>
													<XIcon class="size-3.5" />
													Tolak
												</Button>
											</div>
											{#if stagedReviews[request.id]}
												<div class="rounded-md border border-warning/30 bg-warning/10 p-2 text-xs text-warning">
													<p>Konfirmasi {stagedReviews[request.id] === 'approved' ? 'persetujuan' : 'penolakan'} request ini.</p>
													<div class="mt-2 flex flex-wrap gap-2">
														<LoadingButton
															size="sm"
															onclick={() => void reviewRequest(request, stagedReviewStatus(request.id))}
															loading={reviewBusy === `${stagedReviews[request.id]}:${request.id}`}
															loadingLabel="Memproses..."
															label="Konfirmasi"
														/>
														<Button type="button" size="sm" variant="outline" onclick={() => clearStagedReview(request.id)}>Batal</Button>
													</div>
												</div>
											{/if}
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
