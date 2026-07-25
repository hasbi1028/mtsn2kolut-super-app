<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { cn } from '$lib/utils.js';
	import {
		buildPageItems,
		calculatePaginationRange,
		normalizePageSize,
		offsetForPage,
		type PaginationChange,
		type PaginationReason
	} from '$lib/utils/pagination';

	type Props = {
		page?: number;
		limit?: number;
		total: number;
		pageSizeOptions?: readonly number[];
		defaultLimit?: number;
		itemLabel?: string;
		loading?: boolean;
		disabled?: boolean;
		showPageSize?: boolean;
		showSummary?: boolean;
		showFirstLast?: boolean;
		siblingCount?: number;
		ariaLabel?: string;
		class?: string;
		onchange?: (detail: PaginationChange) => void;
	};

	let {
		page = $bindable(1),
		limit = $bindable(12),
		total,
		pageSizeOptions = [12, 24, 48, 96],
		defaultLimit = pageSizeOptions[0] ?? 12,
		itemLabel = 'data',
		loading = false,
		disabled = false,
		showPageSize = true,
		showSummary = true,
		showFirstLast = true,
		siblingCount = 1,
		ariaLabel = 'Navigasi halaman tabel',
		class: className,
		onchange
	}: Props = $props();

	let range = $derived(calculatePaginationRange(total, page, limit));
	let pageItems = $derived(buildPageItems(range.page, range.pageCount, siblingCount));
	let isDisabled = $derived(disabled || loading);
	let canNavigate = $derived(!isDisabled && total > 0);
	let showNavigation = $derived(range.pageCount > 1);
	let showFooter = $derived(total > 0 || showPageSize);

	function emit(nextPage: number, nextLimit: number, reason: PaginationReason) {
		const safeLimit = normalizePageSize(nextLimit, pageSizeOptions, defaultLimit);
		const next = calculatePaginationRange(total, nextPage, safeLimit);
		page = next.page;
		limit = safeLimit;
		onchange?.({
			page: next.page,
			limit: safeLimit,
			offset: offsetForPage(next.page, safeLimit),
			reason
		});
	}

	function changePage(nextPage: number, reason: PaginationReason) {
		emit(nextPage, limit, reason);
	}

	function changeLimit(event: Event) {
		const target = event.currentTarget as HTMLSelectElement;
		emit(1, normalizePageSize(target.value, pageSizeOptions, defaultLimit), 'limit');
	}
</script>

{#if showFooter}
	<nav
		class={cn(
			'rounded-xl border border-border bg-card p-3 text-sm shadow-sm',
			'flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between',
			className
		)}
		aria-label={ariaLabel}
	>
		<div class="flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-center sm:gap-3">
			{#if showSummary}
				<p class="text-sm text-muted-foreground" aria-live="polite">
					{#if total === 0}
						Tidak ada {itemLabel}
					{:else}
						Menampilkan
						<span class="font-medium text-foreground">{range.start}–{range.end}</span>
						dari <span class="font-medium text-foreground">{total}</span> {itemLabel}
					{/if}
				</p>
			{/if}

			{#if showPageSize}
				<label class="flex items-center gap-2 text-sm text-muted-foreground">
					<span>Tampilkan</span>
					<select
						class="h-9 rounded-md border border-border bg-background px-2 text-sm text-foreground shadow-sm focus:border-ring focus:outline-none focus:ring-2 focus:ring-ring/30 disabled:cursor-not-allowed disabled:opacity-50"
						aria-label="Jumlah data per halaman"
						value={limit}
						disabled={isDisabled}
						onchange={changeLimit}
					>
						{#each pageSizeOptions as option}
							<option value={option}>{option === 0 ? 'Semua' : option}</option>
						{/each}
					</select>
					<span>per halaman</span>
				</label>
			{/if}
		</div>

		{#if showNavigation}
			<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-end">
				<div class="grid grid-cols-2 gap-2 sm:hidden">
					{#if showFirstLast}
						<Button
							variant="outline"
							size="lg"
							class="min-h-10 w-full"
							disabled={!canNavigate || !range.hasPrevious}
							aria-label="Buka halaman pertama"
							onclick={() => changePage(1, 'first')}
						>
							Awal
						</Button>
					{/if}
					<Button
						variant="outline"
						size="lg"
						class="min-h-10 w-full"
						disabled={!canNavigate || !range.hasPrevious}
						aria-label="Buka halaman sebelumnya"
						onclick={() => changePage(range.page - 1, 'previous')}
					>
						Sebelumnya
					</Button>
					<Button
						variant="outline"
						size="lg"
						class="min-h-10 w-full"
						disabled={!canNavigate || !range.hasNext}
						aria-label="Buka halaman berikutnya"
						onclick={() => changePage(range.page + 1, 'next')}
					>
						Berikutnya
					</Button>
					{#if showFirstLast}
						<Button
							variant="outline"
							size="lg"
							class="min-h-10 w-full"
							disabled={!canNavigate || !range.hasNext}
							aria-label="Buka halaman terakhir"
							onclick={() => changePage(range.pageCount, 'last')}
						>
							Akhir
						</Button>
					{/if}
				</div>
				<p class="text-center text-xs text-muted-foreground sm:hidden">
					Halaman {range.page} dari {range.pageCount}
				</p>

				<div class="hidden flex-wrap items-center justify-end gap-1 sm:flex">
					{#if showFirstLast}
						<Button
							variant="outline"
							size="sm"
							disabled={!canNavigate || !range.hasPrevious}
							aria-label="Buka halaman pertama"
							onclick={() => changePage(1, 'first')}
						>
							Awal
						</Button>
					{/if}
					<Button
						variant="outline"
						size="sm"
						disabled={!canNavigate || !range.hasPrevious}
						aria-label="Buka halaman sebelumnya"
						onclick={() => changePage(range.page - 1, 'previous')}
					>
						Sebelumnya
					</Button>

					{#each pageItems as item, index (`${item}-${index}`)}
						{#if item === 'ellipsis'}
							<span class="px-2 text-muted-foreground" aria-hidden="true">…</span>
						{:else}
							<Button
								variant={item === range.page ? 'default' : 'outline'}
								size="sm"
								class="min-w-8"
								disabled={!canNavigate || item === range.page}
								aria-current={item === range.page ? 'page' : undefined}
								aria-label={item === range.page ? `Halaman ${item}, halaman aktif` : `Buka halaman ${item}`}
								onclick={() => changePage(item, 'page')}
							>
								{item}
							</Button>
						{/if}
					{/each}

					<Button
						variant="outline"
						size="sm"
						disabled={!canNavigate || !range.hasNext}
						aria-label="Buka halaman berikutnya"
						onclick={() => changePage(range.page + 1, 'next')}
					>
						Berikutnya
					</Button>
					{#if showFirstLast}
						<Button
							variant="outline"
							size="sm"
							disabled={!canNavigate || !range.hasNext}
							aria-label="Buka halaman terakhir"
							onclick={() => changePage(range.pageCount, 'last')}
						>
							Akhir
						</Button>
					{/if}
				</div>
			</div>
		{/if}
	</nav>
{/if}
