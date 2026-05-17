<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { cn } from '$lib/utils';

	type HeaderAction = {
		label: string;
		href?: string;
		disabled?: boolean;
		onclick?: () => void;
	};

	let {
		eyebrow = '',
		title,
		subtitle = '',
		context = '',
		primaryAction,
		secondaryAction,
		actions,
		meta,
		class: className = '',
	}: {
		eyebrow?: string;
		title: string;
		subtitle?: string;
		context?: string;
		primaryAction?: HeaderAction;
		secondaryAction?: HeaderAction;
		actions?: Snippet;
		meta?: Snippet;
		class?: string;
	} = $props();
</script>

<section class={cn('rounded-xl border border-border bg-card p-4 shadow-sm md:p-5', className)}>
	<div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
		<div class="min-w-0 space-y-2">
			{#if eyebrow || context}
				<div class="flex flex-wrap items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.18em]">
					{#if eyebrow}
						<span class="text-primary">{eyebrow}</span>
					{/if}
					{#if context}
						<span class="rounded-full border border-border bg-muted/50 px-2.5 py-1 text-muted-foreground">{context}</span>
					{/if}
				</div>
			{/if}
			<div>
				<h1 class="text-2xl font-semibold tracking-tight text-foreground md:text-3xl">{title}</h1>
				{#if subtitle}
					<p class="mt-2 max-w-3xl text-sm leading-6 text-muted-foreground">{subtitle}</p>
				{/if}
			</div>
			{#if meta}
				<div class="flex flex-wrap items-center gap-2">
					{@render meta()}
				</div>
			{/if}
		</div>

		{#if primaryAction || secondaryAction || actions}
			<div class="flex shrink-0 flex-wrap gap-2 lg:justify-end">
				{#if secondaryAction}
					<Button
						variant="outline"
						href={secondaryAction.href}
						disabled={secondaryAction.disabled}
						onclick={secondaryAction.onclick}
					>
						{secondaryAction.label}
					</Button>
				{/if}
				{#if actions}
					{@render actions()}
				{/if}
				{#if primaryAction}
					<Button
						href={primaryAction.href}
						disabled={primaryAction.disabled}
						onclick={primaryAction.onclick}
					>
						{primaryAction.label}
					</Button>
				{/if}
			</div>
		{/if}
	</div>
</section>
