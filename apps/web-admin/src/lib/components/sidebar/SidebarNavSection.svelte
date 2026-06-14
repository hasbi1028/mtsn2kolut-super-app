<script lang="ts">
	import { resolve } from '$app/paths';
	import SidebarIcon from '$lib/components/sidebar/SidebarIcon.svelte';
	import type { SidebarNavGroup, SidebarNavItem } from '$lib/components/sidebar/sidebar-config';
	import { numberedLabel } from '$lib/components/sidebar/sidebar-tree';

	let {
		section,
		desktopExpanded,
		isGroupOpen,
		toggleGroup,
		groupBadge,
		navBadge,
		isActive,
		rememberRecent,
		togglePin,
		isPinned,
		pinButtonLabel,
		railTooltip,
		closeMobile
	}: {
		section: SidebarNavGroup;
		desktopExpanded: boolean;
		isGroupOpen: (group: string) => boolean;
		toggleGroup: (group: string) => void;
		groupBadge: (group: string) => number;
		navBadge: (href: string) => number;
		isActive: (href: string) => boolean;
		rememberRecent: (href: string) => void;
		togglePin: (href: string) => void;
		isPinned: (href: string) => boolean;
		pinButtonLabel: (item: SidebarNavItem) => string;
		railTooltip: (item: SidebarNavItem, group: string) => string;
		closeMobile: () => void;
	} = $props();

	function itemTooltip(item: SidebarNavItem, ancestors: string[] = []) {
		return [numberedLabel(section.group, section.section), ...ancestors, numberedLabel(item.label, item.section)].filter(Boolean).join(' › ');
	}
</script>

{#snippet renderItem(node: SidebarNavItem, ancestors: string[] = [], depth = 0)}
	<li>
		<div class={`group relative flex items-center ${desktopExpanded ? 'gap-1' : 'gap-0 lg:justify-center'}`}>
			<a
				href={resolve(node.href as '/')}
				onclick={() => {
					rememberRecent(node.href);
					closeMobile();
				}}
				title={!desktopExpanded ? itemTooltip(node, ancestors) : undefined}
				aria-current={isActive(node.href) ? 'page' : undefined}
				class={`relative flex min-w-0 flex-1 items-center rounded-lg py-1.5 text-sm font-medium transition-all no-underline
					${desktopExpanded ? 'gap-2.5 px-2' : 'gap-2.5 px-2 lg:justify-center lg:px-0'}
					${isActive(node.href)
						? 'bg-[var(--gold)]/15 text-foreground shadow-sm ring-1 ring-[var(--gold)]/20 before:absolute before:inset-y-1 before:left-0 before:w-1 before:rounded-full before:bg-[var(--gold)]'
						: 'text-muted-foreground hover:bg-[var(--gold)]/10 hover:text-foreground'}`}
				style={desktopExpanded ? `padding-left: ${0.5 + depth * 0.65}rem` : undefined}
			>
				<SidebarIcon name={node.icon} active={isActive(node.href)} />
				{#if node.section}
					<span class={`shrink-0 rounded bg-muted px-1.5 py-0.5 text-[10px] font-semibold text-muted-foreground ${desktopExpanded ? 'inline' : 'inline lg:hidden'}`}>{node.section}</span>
				{/if}
				<span class={`truncate ${desktopExpanded ? 'inline' : 'inline lg:hidden'}`}>{node.label}</span>
				{#if navBadge(node.href) > 0 && desktopExpanded}
					<span class="ml-auto rounded-full bg-[var(--gold)]/15 px-1.5 py-0.5 text-[10px] font-semibold text-[var(--gold)]">
						{navBadge(node.href)}
					</span>
				{/if}
			</a>
			{#if node.pinnable !== false}
				<button
					type="button"
					class={`shrink-0 rounded-md p-1 text-[var(--gold)] transition-colors hover:bg-[var(--gold)]/10 ${desktopExpanded ? 'inline-flex' : 'inline-flex lg:hidden'}`}
					onclick={() => togglePin(node.href)}
					aria-label={pinButtonLabel(node)}
				>
					<svg class="h-3.5 w-3.5" fill={isPinned(node.href) ? 'currentColor' : 'none'} stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 17.75l-6.172 3.245 1.179-6.872L2 9.38l6.914-1.005L12 2.11l3.086 6.265L22 9.38l-5.007 4.743 1.18 6.872z" />
					</svg>
				</button>
			{/if}
			{#if !desktopExpanded}
				<div class="pointer-events-none absolute left-full top-1/2 z-40 hidden -translate-y-1/2 rounded-md border border-border bg-popover px-2 py-1 text-xs font-medium text-popover-foreground shadow-sm lg:group-hover:block lg:ml-3">
					{railTooltip(node, section.group)}
				</div>
			{/if}
		</div>
	</li>
{/snippet}

<div>
	<button
		type="button"
		class={`parchment-texture page-enter mb-1 flex w-full items-center rounded-lg px-2 py-1 text-left text-[10px] font-semibold uppercase tracking-[0.18em] text-primary transition-colors hover:bg-[var(--gold)]/10 ${desktopExpanded ? 'flex' : 'flex lg:hidden'}`}
		onclick={() => toggleGroup(section.group)}
		aria-expanded={isGroupOpen(section.group)}
	>
		{#if section.section}
			<span class="mr-1 shrink-0 rounded bg-primary/10 px-1.5 py-0.5 text-[10px] tracking-normal text-primary">{section.section}</span>
		{/if}
		<span class="truncate">{section.group}</span>
		{#if groupBadge(section.group) > 0 && desktopExpanded}
			<span class="ml-2 rounded-full bg-[var(--gold)]/15 px-1.5 py-0.5 text-[10px] font-semibold tracking-normal text-[var(--gold)]">
				{groupBadge(section.group)}
			</span>
		{/if}
		<svg class={`ml-auto h-3.5 w-3.5 transition-transform ${isGroupOpen(section.group) ? 'rotate-90' : ''}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
		</svg>
	</button>
	{#if !desktopExpanded || isGroupOpen(section.group)}
		<ul class="space-y-0.5">
			{#each section.items as item (item.href)}
				{@render renderItem(item)}
			{/each}
		</ul>
	{/if}
</div>
