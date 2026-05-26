<script lang="ts">
	import type { Snippet } from 'svelte';
	import * as Table from '$lib/components/ui/table';
	import { cn } from '$lib/utils';

	type Column = {
		key: string;
		label: string;
		class?: string;
		headClass?: string;
	};

	let {
		title = '',
		description = '',
		columns,
		rows,
		rowKey,
		cell,
		actions,
		mobile,
		empty,
		emptyTitle = 'Belum ada data.',
		emptyDescription = '',
		class: className = '',
		tableClass = 'min-w-[640px]',
		showMobile = true,
	}: {
		title?: string;
		description?: string;
		columns: Column[];
		rows: unknown[];
		rowKey?: (row: unknown, index: number) => string;
		cell: Snippet<[unknown, Column, number]>;
		actions?: Snippet<[unknown, number]>;
		mobile?: Snippet<[unknown, number]>;
		empty?: Snippet;
		emptyTitle?: string;
		emptyDescription?: string;
		class?: string;
		tableClass?: string;
		showMobile?: boolean;
	} = $props();

	const keyFor = (row: unknown, index: number) => rowKey?.(row, index) ?? String(index);
</script>

<section class={cn('overflow-hidden rounded-[1.15rem] border border-[var(--gold)]/20 bg-card shadow-sm shadow-primary/5', className)}>
	{#if title || description}
		<div class="border-b border-[var(--gold)]/15 bg-primary/5 px-3 py-2">
			{#if title}
				<h2 class="font-[var(--font-display)] text-base font-semibold text-foreground">{title}</h2>
			{/if}
			{#if description}
				<p class="mt-0.5 text-xs leading-5 text-muted-foreground">{description}</p>
			{/if}
		</div>
	{/if}

	{#if rows.length > 0}
		<div class="hidden overflow-x-auto md:block">
			<Table.Root class={cn('text-xs', tableClass)}>
				<Table.Header>
					<Table.Row class="bg-primary/5 hover:bg-primary/5">
						{#each columns as column (column.key)}
							<Table.Head class={cn('h-8 border-b border-[var(--gold)]/15 px-3 text-[11px] uppercase tracking-[0.16em] text-muted-foreground', column.headClass)}>
								{column.label}
							</Table.Head>
						{/each}
						{#if actions}
							<Table.Head class="h-8 w-32 px-3 text-right text-[11px] uppercase tracking-wide text-muted-foreground">Aksi</Table.Head>
						{/if}
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each rows as row, index (keyFor(row, index))}
						<Table.Row class="transition-colors hover:bg-primary/5">
							{#each columns as column (column.key)}
								<Table.Cell class={cn('px-3 py-2 align-top', column.class)}>
									{@render cell(row, column, index)}
								</Table.Cell>
							{/each}
							{#if actions}
								<Table.Cell class="px-3 py-2 text-right align-top">
									<div class="flex flex-wrap justify-end gap-1.5">
										{@render actions(row, index)}
									</div>
								</Table.Cell>
							{/if}
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</div>

		{#if showMobile}
			<div class="divide-y divide-border md:hidden">
				{#each rows as row, index (keyFor(row, index))}
					<div class="p-3">
						{#if mobile}
							{@render mobile(row, index)}
						{:else}
							<div class="space-y-2 text-xs">
								{#each columns as column (column.key)}
									<div>
										<p class="text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">{column.label}</p>
										<div class="mt-0.5">{@render cell(row, column, index)}</div>
									</div>
								{/each}
								{#if actions}
									<div class="flex flex-wrap gap-1.5 pt-1">{@render actions(row, index)}</div>
								{/if}
							</div>
						{/if}
					</div>
				{/each}
			</div>
		{/if}
	{:else}
		<div class="p-4 text-center text-sm text-muted-foreground">
			{#if empty}
				{@render empty()}
			{:else}
				<p class="font-medium text-foreground">{emptyTitle}</p>
				{#if emptyDescription}
					<p class="mt-1 text-xs">{emptyDescription}</p>
				{/if}
			{/if}
		</div>
	{/if}
</section>
