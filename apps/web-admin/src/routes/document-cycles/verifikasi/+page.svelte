<script lang="ts">
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import BellIcon from '@lucide/svelte/icons/bell';
	import CalendarClockIcon from '@lucide/svelte/icons/calendar-clock';
	import CheckCircle2Icon from '@lucide/svelte/icons/check-circle-2';
	import FileWarningIcon from '@lucide/svelte/icons/file-warning';
	import PencilIcon from '@lucide/svelte/icons/pencil';
	import RefreshCcwIcon from '@lucide/svelte/icons/refresh-ccw';
	import { base, resolve } from '$app/paths';
	import { onMount } from 'svelte';
	import type { PageData } from './$types';
	import type { BadgeVariant } from '$lib/components/ui/badge';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import * as Table from '$lib/components/ui/table';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { toast } from '$lib/components/ui/sonner';
	import { readClientJson } from '$lib/client/api';

	let { data }: { data: PageData } = $props();

	interface DocumentCycleObligation {
		id: string;
		catalog_code: string;
		catalog_title: string;
		frequency: string;
		domain_area: string;
		external_system: string;
		snp_standard: string;
		period_year: number;
		period_label: string;
		period_start: string;
		period_end: string;
		due_date: string;
		reminder_date: string;
		owner_unit_name: string;
		responsible_employee_id?: string;
		responsible_employee_name: string;
		responsible_employee_nip: string;
		verifier_employee_id?: string;
		verifier_employee_name: string;
		status: string;
		archive_document_id?: string;
		archive_document_title: string;
		evidence_item_id?: string;
		evidence_item_title: string;
		compliance_action_id?: string;
		compliance_action_title: string;
		verification_notes: string;
		is_overdue: boolean;
		is_due_soon: boolean;
	}

	const DOMAIN_AREAS: Array<[string, string]> = [
		['tu', 'TU'],
		['kesiswaan', 'Kesiswaan'],
		['kurikulum', 'Kurikulum'],
		['sarpras', 'Sarpras'],
		['governance', 'Governance'],
		['keuangan', 'Keuangan'],
		['eksternal', 'Eksternal']
	];

	const FREQUENCIES: Array<[string, string]> = [
		['daily', 'Harian'],
		['weekly', 'Mingguan'],
		['monthly', 'Bulanan'],
		['quarterly', 'Triwulan'],
		['semester', 'Semester'],
		['annual', 'Tahunan'],
		['four_year', '4 Tahunan'],
		['five_year', '5 Tahunan']
	];

	let periodYear = $state(new Date().getFullYear());
	let includeAll = $state(false);
	let refreshBusy = $state(false);
	let actionBusyId = $state('');
	let queuePromise = $state<Promise<DocumentCycleObligation[]> | null>(null);
	let queueRequestId = 0;

	const isAdmin = $derived((data.user?.roles ?? []).includes('admin') || data.user?.role === 'admin');

	function refreshQueue() {
		const requestId = ++queueRequestId;
		refreshBusy = true;
		const next = loadQueue();
		queuePromise = next.finally(() => {
			if (requestId === queueRequestId) {
				refreshBusy = false;
			}
		});
	}

	async function loadQueue(): Promise<DocumentCycleObligation[]> {
		const params = new URLSearchParams({ period_year: String(periodYear) });
		if (includeAll && isAdmin) params.set('all', 'true');
		return fetchJson<DocumentCycleObligation[]>(`/api/document-cycles/verification-queue?${params.toString()}`);
	}

	async function fetchJson<T>(path: string, init?: RequestInit): Promise<T> {
		const response = await fetch(`${base}${path}`, init);
		return readClientJson<T>(response);
	}

	async function updateObligationStatus(obligation: DocumentCycleObligation, status: string) {
		if (status === 'completed') {
			const gaps = completionIssues(obligation);
			if (gaps.length > 0) {
				toast.error(`Belum bisa diselesaikan. Lengkapi ${gaps.join(', ')}.`);
				return;
			}
		}

		actionBusyId = `${obligation.id}:${status}`;
		try {
			await fetchJson<DocumentCycleObligation>(`/api/document-cycles/obligations/${obligation.id}/status`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ status, notes: statusNote(status) })
			});
			toast.success(`${obligation.catalog_code} diperbarui menjadi ${statusLabel(status)}`);
			refreshQueue();
		} catch (error) {
			toast.error(errorMessage(error));
		} finally {
			actionBusyId = '';
		}
	}

	function completionIssues(item: DocumentCycleObligation): string[] {
		const issues: string[] = [];
		if (!item.responsible_employee_id) issues.push('PIC penyusun');
		if (!item.verifier_employee_id) issues.push('verifikator');
		if (!item.archive_document_id) issues.push('arsip digital');
		return issues;
	}

	function statusNote(status: string): string {
		if (status === 'draft') return 'Dikembalikan ke draft dari antrian verifikasi.';
		if (status === 'completed') return 'Dokumen diverifikasi dan ditandai selesai dari antrian verifikasi.';
		return 'Status monitoring dokumen diperbarui dari antrian verifikasi.';
	}

	function statusLabel(status: string): string {
		if (status === 'waiting_verification') return 'Menunggu Verifikasi';
		if (status === 'completed') return 'Selesai';
		if (status === 'draft') return 'Sedang Dibuat';
		return 'Belum Mulai';
	}

	function statusVariant(status: string): BadgeVariant {
		if (status === 'completed') return 'default';
		if (status === 'waiting_verification') return 'secondary';
		if (status === 'draft') return 'outline';
		return 'ghost';
	}

	function domainAreaLabel(area: string): string {
		return DOMAIN_AREAS.find(([value]) => value === area)?.[1] ?? area;
	}

	function frequencyLabel(frequency: string): string {
		return FREQUENCIES.find(([value]) => value === frequency)?.[1] ?? frequency;
	}

	function formatDate(value?: string): string {
		if (!value) return '-';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value;
		return new Intl.DateTimeFormat('id-ID', {
			day: '2-digit',
			month: 'short',
			year: 'numeric',
			timeZone: 'Asia/Makassar'
		}).format(date);
	}

	function errorMessage(error: unknown): string {
		return error instanceof Error ? error.message : 'Operasi gagal diproses';
	}

	function handleRenderError(error: unknown) {
		console.error('Document verification queue render failed', error);
	}

	function retryQueue(reset?: () => void) {
		reset?.();
		refreshQueue();
	}

	onMount(() => {
		refreshQueue();
	});
</script>

<svelte:head><title>Antrian Verifikasi Dokumen - MTsN 2 Kolaka Utara</title></svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
		<div>
			<div class="mb-2 inline-flex items-center gap-2 rounded-full border border-primary/20 bg-primary/10 px-3 py-1 text-xs font-medium text-primary">
				<BellIcon class="size-3.5" />
				Antrian Verifikator
			</div>
			<h1 class="text-xl font-semibold text-foreground">Verifikasi Siklus Dokumen</h1>
			<p class="mt-1 max-w-3xl text-sm text-muted-foreground">
				Dokumen yang sedang menunggu verifikasi sesuai penugasan verifikator.
			</p>
		</div>
		<div class="flex flex-wrap items-end gap-2">
			<div class="w-28">
				<label for="verification-period-year" class="text-xs font-medium text-muted-foreground">Tahun</label>
				<Input id="verification-period-year" type="number" min="2000" bind:value={periodYear} />
			</div>
			{#if isAdmin}
				<label class="inline-flex h-9 items-center gap-2 rounded-md border border-input px-3 text-sm text-foreground">
					<input type="checkbox" bind:checked={includeAll} class="size-4 accent-emerald-700" />
					Semua verifikator
				</label>
			{/if}
			<LoadingButton variant="outline" onclick={() => void refreshQueue()} loading={refreshBusy} loadingLabel="Memuat">
				<RefreshCcwIcon class="mr-2 size-4" />
				Refresh
			</LoadingButton>
		</div>
	</div>

	<AsyncContent promise={queuePromise} onerror={handleRenderError}>
		{#snippet pending()}
			<Card.Root class="border-border">
				<Card.Content class="space-y-3 p-4">
					{#each Array.from({ length: 6 }) as _, index (`document-verification-skeleton-${index}`)}
						<Skeleton class="h-14 w-full" />
					{/each}
				</Card.Content>
			</Card.Root>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel message={errorMessage(error)} onRetry={() => retryQueue(reset)} />
		{/snippet}

		{#snippet children(value)}
			{@const queue = value as DocumentCycleObligation[]}
			<div class="grid gap-3 md:grid-cols-3">
				<Card.Root class="border-border">
					<Card.Content class="p-4">
						<p class="text-xs text-muted-foreground">Menunggu Verifikasi</p>
						<p class="mt-2 text-2xl font-semibold text-foreground">{queue.length}</p>
					</Card.Content>
				</Card.Root>
				<Card.Root class="border-border">
					<Card.Content class="p-4">
						<p class="text-xs text-muted-foreground">Lewat Tempo</p>
						<p class="mt-2 text-2xl font-semibold text-destructive">{queue.filter((item) => item.is_overdue).length}</p>
					</Card.Content>
				</Card.Root>
				<Card.Root class="border-border">
					<Card.Content class="p-4">
						<p class="text-xs text-muted-foreground">Belum Siap Selesai</p>
						<p class="mt-2 text-2xl font-semibold text-warning">{queue.filter((item) => completionIssues(item).length > 0).length}</p>
					</Card.Content>
				</Card.Root>
			</div>

			<Card.Root class="border-border">
				<Card.Header class="pb-3">
					<div class="flex items-center gap-2">
						<CalendarClockIcon class="size-4 text-primary" />
						<Card.Title class="text-base">Dokumen Menunggu Saya Verifikasi</Card.Title>
					</div>
					<Card.Description>Gunakan koreksi draft untuk mengembalikan dokumen ke PIC, atau tandai selesai setelah arsip resmi tertaut.</Card.Description>
				</Card.Header>
				<Card.Content class="p-0">
					<div class="overflow-x-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Dokumen</Table.Head>
									<Table.Head>PIC</Table.Head>
									<Table.Head>Tenggat</Table.Head>
									<Table.Head>Kesiapan</Table.Head>
									<Table.Head>Aksi</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each queue as item (item.id)}
									{@const gaps = completionIssues(item)}
									<Table.Row>
										<Table.Cell class="min-w-80">
											<div class="flex items-start gap-3">
												<div class="mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary">
													<BellIcon class="size-4" />
												</div>
												<div>
													<p class="text-sm font-medium text-foreground">{item.catalog_title}</p>
													<p class="text-xs text-muted-foreground">{item.catalog_code} · {item.period_label} · {frequencyLabel(item.frequency)}</p>
													<div class="mt-2 flex flex-wrap gap-1.5">
														<Badge variant="outline">{domainAreaLabel(item.domain_area)}</Badge>
														<Badge variant={statusVariant(item.status)}>{statusLabel(item.status)}</Badge>
													</div>
												</div>
											</div>
										</Table.Cell>
										<Table.Cell class="min-w-56">
											<p class="text-sm text-foreground">{item.responsible_employee_name || 'Belum ada PIC'}</p>
											<p class="text-xs text-muted-foreground">{item.owner_unit_name || 'Tanpa unit'}{item.responsible_employee_nip ? ` · ${item.responsible_employee_nip}` : ''}</p>
										</Table.Cell>
										<Table.Cell class="whitespace-nowrap">
											<p class={item.is_overdue ? 'text-sm font-medium text-destructive' : 'text-sm text-foreground'}>Jatuh tempo {formatDate(item.due_date)}</p>
											<p class="text-xs text-muted-foreground">Pengingat {formatDate(item.reminder_date)}</p>
											{#if item.is_overdue}
												<p class="mt-1 inline-flex items-center gap-1 text-xs font-medium text-destructive">
													<AlertTriangleIcon class="size-3" />
													Lewat tempo
												</p>
											{/if}
										</Table.Cell>
										<Table.Cell class="min-w-64">
											<div class="flex flex-wrap gap-1.5">
												<Badge variant={item.archive_document_id ? 'default' : 'outline'}>Arsip</Badge>
												<Badge variant={item.evidence_item_id ? 'default' : 'outline'}>Evidence</Badge>
												<Badge variant={item.compliance_action_id ? 'default' : 'outline'}>Aksi</Badge>
											</div>
											<p class="mt-2 text-xs text-muted-foreground">{gaps.length > 0 ? `Kurang ${gaps.join(', ')}` : item.archive_document_title || 'Siap diverifikasi'}</p>
										</Table.Cell>
										<Table.Cell>
											<div class="flex min-w-72 flex-wrap gap-1.5">
												<Button size="sm" variant="outline" href={`${resolve('/document-cycles')}?selected_obligation=${item.id}&period_year=${periodYear}`}>
													<PencilIcon class="mr-2 size-3.5" />
													Detail
												</Button>
												<Button
													size="sm"
													variant="outline"
													disabled={actionBusyId === `${item.id}:draft`}
													onclick={() => void updateObligationStatus(item, 'draft')}
												>
													<FileWarningIcon class="mr-2 size-3.5" />
													Koreksi Draft
												</Button>
												<Button
													size="sm"
													disabled={actionBusyId === `${item.id}:completed` || gaps.length > 0}
													title={gaps.length > 0 ? `Lengkapi ${gaps.join(', ')}` : 'Tandai selesai'}
													onclick={() => void updateObligationStatus(item, 'completed')}
												>
													<CheckCircle2Icon class="mr-2 size-3.5" />
													Selesai
												</Button>
											</div>
										</Table.Cell>
									</Table.Row>
								{:else}
									<Table.Row>
										<Table.Cell colspan={5} class="py-10 text-center text-sm text-muted-foreground">
											Tidak ada dokumen yang menunggu verifikasi pada tahun dan cakupan ini.
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</div>
				</Card.Content>
			</Card.Root>
		{/snippet}
	</AsyncContent>
</div>
