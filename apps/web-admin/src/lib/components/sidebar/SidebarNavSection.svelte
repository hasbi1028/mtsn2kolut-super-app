<script lang="ts">
	import { resolve } from '$app/paths';
	import SidebarIcon from '$lib/components/sidebar/SidebarIcon.svelte';
	import type { SidebarNavGroup, SidebarNavItem } from '$lib/components/sidebar/sidebar-config';

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

</script>

<div>
	<button
		type="button"
		class={`mb-1 flex w-full items-center rounded-md px-2 py-1 text-left text-[10px] font-semibold uppercase tracking-wider text-slate-400 transition-colors hover:bg-slate-50 ${desktopExpanded ? 'flex' : 'flex lg:hidden'}`}
		onclick={() => toggleGroup(section.group)}
		aria-expanded={isGroupOpen(section.group)}
	>
		<span class="truncate">{section.group}</span>
		{#if groupBadge(section.group) > 0 && desktopExpanded}
			<span class="ml-2 rounded-full bg-amber-100 px-1.5 py-0.5 text-[10px] font-semibold tracking-normal text-amber-700">
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
				<li>
					<div class={`group relative flex items-center ${desktopExpanded ? 'gap-1' : 'gap-0 lg:justify-center'}`}>
						<a
							href={resolve(item.href as '/')}
							onclick={() => {
								rememberRecent(item.href);
								closeMobile();
							}}
							title={!desktopExpanded ? railTooltip(item, section.group) : undefined}
							class={`flex min-w-0 flex-1 items-center rounded-md py-1.5 text-sm font-medium transition-colors
								${desktopExpanded ? 'gap-2.5 px-2' : 'gap-2.5 px-2 lg:justify-center lg:px-0'}
								${isActive(item.href)
									? 'bg-green-50 text-green-800'
									: 'text-slate-600 hover:bg-green-50/60 hover:text-slate-800'}`}
						>
							<SidebarIcon name={item.icon} active={isActive(item.href)} />
							<span class={`truncate ${desktopExpanded ? 'inline' : 'inline lg:hidden'}`}>{item.label}</span>
							{#if navBadge(item.href) > 0 && desktopExpanded}
								<span class="ml-auto rounded-full bg-amber-100 px-1.5 py-0.5 text-[10px] font-semibold text-amber-700">
									{navBadge(item.href)}
								</span>
							{/if}
						</a>
						{#if item.pinnable !== false}
							<button
								type="button"
								class={`shrink-0 rounded-md p-1 text-slate-400 transition-colors hover:bg-amber-50 hover:text-amber-600 ${desktopExpanded ? 'inline-flex' : 'inline-flex lg:hidden'}`}
								onclick={() => togglePin(item.href)}
								aria-label={pinButtonLabel(item)}
							>
								<svg class="h-3.5 w-3.5" fill={isPinned(item.href) ? 'currentColor' : 'none'} stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 17.75l-6.172 3.245 1.179-6.872L2 9.38l6.914-1.005L12 2.11l3.086 6.265L22 9.38l-5.007 4.743 1.18 6.872z" />
								</svg>
							</button>
						{/if}
						{#if !desktopExpanded}
							<div class="pointer-events-none absolute left-full top-1/2 z-40 hidden -translate-y-1/2 rounded-md border border-slate-200 bg-white px-2 py-1 text-xs font-medium text-slate-700 shadow-sm lg:group-hover:block lg:ml-3">
								{railTooltip(item, section.group)}
							</div>
						{/if}
					</div>
				</li>
			{/each}
		</ul>
	{/if}
</div>
