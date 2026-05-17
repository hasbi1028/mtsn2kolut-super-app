<script lang="ts">
	import { cn } from '$lib/utils';

	export type EntityTab = {
		id: string;
		label: string;
		badge?: string | number;
		description?: string;
	};

	let {
		tabs = [],
		active = $bindable(''),
		label = 'Area kerja',
		onSelect,
		class: className = '',
	}: {
		tabs?: EntityTab[];
		active?: string;
		label?: string;
		onSelect?: (id: string) => void;
		class?: string;
	} = $props();

	const visibleTabs = $derived(tabs.slice(0, 5));

	function selectTab(id: string) {
		active = id;
		onSelect?.(id);
	}
</script>

<nav aria-label={label} class={cn('overflow-x-auto rounded-lg border border-border bg-muted/50 p-1', className)}>
	<div class="flex min-w-max gap-1" role="tablist" aria-label={label}>
		{#each visibleTabs as tab (tab.id)}
			<button
				type="button"
				role="tab"
				aria-selected={active === tab.id}
				aria-controls={`${tab.id}-panel`}
				title={tab.description}
				class={cn(
					'inline-flex h-9 items-center gap-2 rounded-md px-3 text-sm font-semibold transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
					active === tab.id ? 'bg-card text-primary shadow-sm ring-1 ring-primary/20' : 'text-muted-foreground hover:bg-card hover:text-foreground'
				)}
				onclick={() => selectTab(tab.id)}
			>
				<span>{tab.label}</span>
				{#if tab.badge !== undefined && tab.badge !== ''}
					<span class="rounded-full border border-current/20 px-1.5 py-0.5 text-[10px]">{tab.badge}</span>
				{/if}
			</button>
		{/each}
	</div>
</nav>
