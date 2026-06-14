<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import AccountMenu from '$lib/components/AccountMenu.svelte';
	import { filterSidebarNavGroupsByAccess } from '$lib/components/sidebar/sidebar-access';
	import { findActiveSidebarHref } from '$lib/components/sidebar/sidebar-active';
	import { flattenSidebarNavGroups, numberSidebarNavGroups } from '$lib/components/sidebar/sidebar-tree';
	import { appAttribution, defaultBranding, versionedAsset, type BrandingSettings } from '$lib/branding';
	import { sidebarNavGroups, type SidebarNavItem } from '$lib/components/sidebar/sidebar-config';
	import type { AccountIdentity } from '$lib/client/account';

	let {
		user,
		account = null,
		branding = defaultBranding
	}: {
		user?: { id: string; username: string; role: string; roles?: string[]; permissions?: string[]; employee_id?: string };
		account?: AccountIdentity | null;
		branding?: BrandingSettings;
	} = $props();

	let mobileMenuOpen = $state(false);

	const userRoles = $derived(user?.roles || (user?.role ? [user.role] : []));
	const userPermissions = $derived(user?.permissions || []);
	const numberedGroups = numberSidebarNavGroups(sidebarNavGroups);
	const nav = $derived(filterSidebarNavGroupsByAccess(numberedGroups, userRoles, userPermissions));

	const visibleNavItems = $derived(flattenSidebarNavGroups(nav));
	const activeHref = $derived(findActiveSidebarHref(page.url.pathname, visibleNavItems));

	function isActive(href: string) {
		return activeHref === href;
	}

	// Simple icon map — maps icon names to SVG viewBox and path data
	const iconMap: Record<string, { viewBox: string; path: string }> = {
		home: { viewBox: '0 0 24 24', path: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-4 0a1 1 0 01-1-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 01-1 1' },
		clock: { viewBox: '0 0 24 24', path: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z' },
		users: { viewBox: '0 0 24 24', path: 'M12 4.354a4 4 0 110 7.292 4 4 0 010-7.292zM15 21H9a2 2 0 01-2-2V12a2 2 0 012-2h6a2 2 0 012 2v7a2 2 0 01-2 2z' },
		calendar: { viewBox: '0 0 24 24', path: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z' },
		'bar-chart': { viewBox: '0 0 24 24', path: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z' },
		send: { viewBox: '0 0 24 24', path: 'M12 19l9 2-9-18-9 18 9-2zm0 0v-8' },
		'refresh-cw': { viewBox: '0 0 24 24', path: 'M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15' },
		user: { viewBox: '0 0 24 24', path: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z' },
		building: { viewBox: '0 0 24 24', path: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0H5m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4' },
		shield: { viewBox: '0 0 24 24', path: 'M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z' },
		image: { viewBox: '0 0 24 24', path: 'M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z' },
		database: { viewBox: '0 0 24 24', path: 'M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4' },
		'file-text': { viewBox: '0 0 24 24', path: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' }
	};

	// Mobile bottom nav items — simplified for teacher use
	const mobileNavItems = [
		{ label: 'Beranda', icon: 'home', href: '/' },
		{ label: 'Kehadiran', icon: 'clock', href: '/pusaka' },
		{ label: 'Pegawai', icon: 'users', href: '/employees' },
		{ label: 'Pengaturan', icon: 'user', href: '/settings/account' }
	];

	function getIconPath(iconName: string): { viewBox: string; path: string } | null {
		return iconMap[iconName] ?? null;
	}

	function navTo(href: string) {
		mobileMenuOpen = false;
		goto(resolve(href as '/'));
	}
</script>

<!-- Desktop Sidebar -->
<aside class="hidden lg:fixed lg:inset-y-0 lg:left-0 lg:z-30 lg:flex lg:w-60 lg:flex-col lg:border-r lg:border-border lg:bg-[var(--card)]">
	<!-- Brand -->
	<div class="flex h-14 shrink-0 items-center gap-2.5 border-b border-border px-4">
		<span class="flex h-8 w-8 shrink-0 items-center justify-center overflow-hidden rounded-lg shadow-sm" style={`background: ${branding.primary_color}`}>
			<img src={versionedAsset(branding.mark_url, branding.version)} alt={`Ikon ${branding.short_name}`} class="h-full w-full object-cover" />
		</span>
		<div class="min-w-0 flex-1">
			<p class="truncate text-sm font-semibold text-foreground">{branding.short_name}</p>
			<p class="truncate text-xs text-muted-foreground">{branding.tagline}</p>
		</div>
	</div>

	<!-- Navigation -->
	<nav class="flex-1 overflow-y-auto px-3 py-3 space-y-4" style="font-size: 15px;">
		{#each nav as section (section.group)}
			{#if section.items.length > 0}
				<div>
					<p class="mb-1 px-3 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">{section.group}</p>
					<div class="space-y-0.5">
						{#each section.items as item ('href' in item ? item.href : item.id)}
							{#if 'href' in item}
								{@const icon = getIconPath(item.icon)}
								<a
									href={resolve(item.href as '/')}
									class="flex w-full items-center gap-2.5 rounded-lg px-3 py-2 text-[15px] no-underline transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
									class:bg-[oklch(0.92_0.04_145)]={isActive(item.href)}
									class:text-green-700={isActive(item.href)}
									class:hover:bg-gray-100={!isActive(item.href)}
									class:text-gray-700={!isActive(item.href)}
									onclick={(e) => { e.preventDefault(); navTo(item.href); }}
								>
									{#if icon}
										<svg class="h-[18px] w-[18px] shrink-0" fill="none" stroke="currentColor" stroke-width="1.5" viewBox={icon.viewBox}>
											<path stroke-linecap="round" stroke-linejoin="round" d={icon.path} />
										</svg>
									{/if}
									<span class="truncate">{item.label}</span>
								</a>
							{/if}
						{/each}
					</div>
				</div>
			{/if}
		{/each}
	</nav>

	<!-- User Section -->
	{#if user}
		<div class="shrink-0 border-t border-border p-3">
			<AccountMenu
				{user}
				{account}
				showName={true}
				menuSide="top"
				align="start"
				menuId="desktop-sidebar-account-menu"
				class="w-full"
				buttonClass="w-full justify-start"
			/>
		</div>
	{/if}

	<!-- Attribution -->
	<div class="shrink-0 border-t border-border px-3 py-2 text-[10px] leading-4 text-muted-foreground/80">
		<p class="truncate font-medium">{appAttribution.productName}</p>
		<p class="truncate">{appAttribution.shortLabel}</p>
	</div>
</aside>

<!-- Mobile Header -->
<header class="fixed inset-x-0 top-0 z-20 flex h-14 w-full items-center gap-3 border-b border-border bg-[var(--card)] px-4 lg:hidden">
	<span class="flex h-8 w-8 shrink-0 items-center justify-center overflow-hidden rounded-lg shadow-sm" style={`background: ${branding.primary_color}`}>
		<img src={versionedAsset(branding.mark_url, branding.version)} alt={`Ikon ${branding.short_name}`} class="h-full w-full object-cover" />
	</span>
	<span class="min-w-0 flex-1 truncate text-sm font-semibold text-foreground">{branding.short_name}</span>
	{#if user}
		<AccountMenu
			{user}
			{account}
			menuId="mobile-topbar-account-menu"
			buttonClass="border-transparent bg-transparent shadow-none hover:bg-gray-100"
		/>
	{/if}
</header>

<!-- Mobile Bottom Navigation -->
<nav class="fixed inset-x-0 bottom-0 z-30 flex h-14 items-stretch border-t border-border bg-[var(--card)] lg:hidden">
	{#each mobileNavItems as item (item.href)}
		{@const icon = getIconPath(item.icon)}
		{@const active = isActive(item.href)}
		<a
			href={resolve(item.href as '/')}
			class="flex flex-1 flex-col items-center justify-center gap-0.5 text-[11px] font-medium no-underline transition-colors"
			class:text-green-700={active}
			class:text-gray-400={!active}
			onclick={(e) => { e.preventDefault(); navTo(item.href); }}
		>
			{#if icon}
				<svg class="h-5 w-5" fill="none" stroke="currentColor" stroke-width={active ? 2.5 : 1.5} viewBox={icon.viewBox}>
					<path stroke-linecap="round" stroke-linejoin="round" d={icon.path} />
				</svg>
			{/if}
			<span>{item.label}</span>
		</a>
	{/each}
</nav>

<!-- Mobile content spacer -->
<div class="h-14 lg:hidden"></div>
