<script lang="ts">
	import { resolve } from '$app/paths';
	import SidebarIcon from '$lib/components/sidebar/SidebarIcon.svelte';
	import type { SidebarNavItem } from '$lib/components/sidebar/sidebar-config';

	type QuickAccessItem = SidebarNavItem & { group: string };

	let {
		items,
		desktopExpanded,
		isActive,
		rememberRecent,
		togglePin,
		isPinned,
		pinButtonLabel,
		canMovePinned,
		movePinned,
		railTooltip,
		closeMobile
	}: {
		items: QuickAccessItem[];
		desktopExpanded: boolean;
		isActive: (href: string) => boolean;
		rememberRecent: (href: string) => void;
		togglePin: (href: string) => void;
		isPinned: (href: string) => boolean;
		pinButtonLabel: (item: SidebarNavItem) => string;
		canMovePinned: (href: string, direction: -1 | 1) => boolean;
		movePinned: (href: string, direction: -1 | 1) => void;
		railTooltip: (item: SidebarNavItem, group: string) => string;
		closeMobile: () => void;
	} = $props();

</script>

{#if items.length > 0}
	<div>
		<p class={`mb-1 px-2 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground ${desktopExpanded ? 'block' : 'block lg:hidden'}`}>
			Akses Cepat
		</p>
		<ul class="space-y-0.5">
			{#each items as item (item.href)}
				<li>
					<div class={`group relative flex items-center ${desktopExpanded ? 'gap-1' : 'gap-0 lg:justify-center'}`}>
						<a
							href={resolve(item.href as '/')}
							onclick={() => {
								rememberRecent(item.href);
								closeMobile();
							}}
							title={!desktopExpanded ? railTooltip(item, item.group) : undefined}
							class={`flex min-w-0 flex-1 items-center rounded-md py-1.5 text-sm font-medium transition-colors
								${desktopExpanded ? 'gap-2.5 px-2' : 'gap-2.5 px-2 lg:justify-center lg:px-0'}
								${isActive(item.href)
									? 'bg-accent text-accent-foreground'
									: 'text-muted-foreground hover:bg-muted hover:text-foreground'}`}
						>
							<SidebarIcon name={item.icon} active={isActive(item.href)} />
							<span class={`truncate ${desktopExpanded ? 'inline' : 'inline lg:hidden'}`}>{item.label}</span>
						</a>
						{#if item.pinnable !== false}
							{#if desktopExpanded && isPinned(item.href)}
								<div class="flex items-center gap-0.5">
									<button
										type="button"
										class="inline-flex rounded-md p-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:cursor-not-allowed disabled:opacity-40"
										onclick={() => movePinned(item.href, -1)}
										disabled={!canMovePinned(item.href, -1)}
										aria-label={`Naikkan ${item.label} dalam akses cepat`}
									>
										<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
										</svg>
									</button>
									<button
										type="button"
										class="inline-flex rounded-md p-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:cursor-not-allowed disabled:opacity-40"
										onclick={() => movePinned(item.href, 1)}
										disabled={!canMovePinned(item.href, 1)}
										aria-label={`Turunkan ${item.label} dalam akses cepat`}
									>
										<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
										</svg>
									</button>
								</div>
							{/if}
							<button
								type="button"
								class={`shrink-0 rounded-md p-1 text-warning hover:bg-warning/10 hover:text-warning ${desktopExpanded ? 'inline-flex' : 'inline-flex lg:hidden'}`}
								onclick={() => togglePin(item.href)}
								aria-label={pinButtonLabel(item)}
							>
								<svg class="h-3.5 w-3.5" fill={isPinned(item.href) ? 'currentColor' : 'none'} stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 17.75l-6.172 3.245 1.179-6.872L2 9.38l6.914-1.005L12 2.11l3.086 6.265L22 9.38l-5.007 4.743 1.18 6.872z" />
								</svg>
							</button>
						{/if}
						{#if !desktopExpanded}
							<div class="pointer-events-none absolute left-full top-1/2 z-40 ml-3 hidden -translate-y-1/2 rounded-md border border-border bg-popover px-2 py-1 text-xs font-medium text-popover-foreground shadow-sm lg:group-hover:block">
								{railTooltip(item, item.group)}
							</div>
						{/if}
					</div>
				</li>
			{/each}
		</ul>
	</div>
{/if}
