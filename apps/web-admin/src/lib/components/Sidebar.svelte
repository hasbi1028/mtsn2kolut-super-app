<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { sidebarNavGroups } from '$lib/components/sidebar/sidebar-config';
	import { versionedAsset, type BrandingSettings } from '$lib/branding';

	let {
		user,
		branding = {} as BrandingSettings,
		isMobileMenuOpen = $bindable(false)
	}: {
		user?: { id: string; username?: string; role?: string; roles?: string[]; permissions?: string[] };
		branding?: BrandingSettings;
		isMobileMenuOpen: boolean;
	} = $props();

	const userRoles = $derived(user?.roles || (user?.role ? [user.role] : []));
	const userPermissions = $derived(user?.permissions || []);

	const filteredMenus = $derived.by(() => {
		return sidebarNavGroups
			.map((group) => {
				const items = group.items.filter((item) => {
					if (userRoles.includes('admin')) return true;
					if (item.permissions && item.permissions.some((p: string) => userPermissions.includes(p))) return true;
					if (item.roles && item.roles.some((r: string) => userRoles.includes(r))) return true;
					if (item.allowAuthenticatedFallback && user) return true;
					return false;
				});
				return { ...group, items };
			})
			.filter((group) => group.items.length > 0);
	});

	// Icon map (same as original, kept for compatibility)
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
		'file-text': { viewBox: '0 0 24 24', path: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' },
		menu: { viewBox: '0 0 24 24', path: 'M4 6h16M4 12h16M4 18h16' },
		'x-circle': { viewBox: '0 0 24 24', path: 'M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z' }
	};

	const getIconPath = (iconName: string) => iconMap[iconName] ?? null;
</script>

<!-- ═══ CBT-style Sidebar ═══ -->
<aside
	id="admin-mobile-sidebar"
	class={[
		'admin-sidebar fixed top-0 bottom-0 left-0 z-[120] flex w-[min(22.5rem,calc(100vw-48px))] flex-col border-r border-border bg-card p-3 shadow-2xl transition-transform duration-300 lg:sticky lg:top-0 lg:h-screen lg:w-72 lg:self-start lg:overflow-hidden lg:p-5 lg:shadow-none',
		isMobileMenuOpen
			? 'pointer-events-auto translate-x-0'
			: 'pointer-events-none -translate-x-full',
		'lg:pointer-events-auto lg:translate-x-0'
	]}
	aria-hidden={!isMobileMenuOpen}
>
	<!-- Brand -->
	<div class="brand mb-3 flex shrink-0 items-center gap-2 px-1 pt-[max(0.75rem,env(safe-area-inset-top))] lg:mb-5 lg:px-2 lg:pt-0">
		<div class="flex min-w-0 flex-1 items-center gap-2">
			{#if branding?.mark_url}
				<div class="brand-logo flex h-10 w-10 items-center justify-center overflow-hidden rounded-xl border border-border bg-background/80 shadow-sm shadow-primary/10">
					<img src={versionedAsset(branding.mark_url, branding.version)} alt="Logo" class="h-full w-full object-contain p-1.5" />
				</div>
			{:else}
				<div class="brand-logo flex h-10 w-10 items-center justify-center rounded-xl bg-primary font-black text-primary-foreground italic shadow-sm shadow-primary/20">
					SA
				</div>
			{/if}
			<div class="min-w-0">
				<div class="brand-text truncate leading-none font-black tracking-tight text-foreground uppercase">
					Command Center
				</div>
				<div class="mt-0.5 text-[9px] font-black tracking-[0.1em] text-primary uppercase">
					{branding?.short_name || 'MTsN 2 Kolut'}
				</div>
			</div>
		</div>
		<button
			type="button"
			class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border border-border bg-background text-muted-foreground shadow-sm hover:bg-muted hover:text-foreground lg:h-8 lg:w-8 lg:hidden"
			onclick={() => (isMobileMenuOpen = false)}
			aria-label="Tutup menu"
		>
			<svg class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
			</svg>
		</button>
	</div>

	<!-- Navigation -->
	<nav class="admin-nav custom-scrollbar min-h-0 flex-1 overflow-y-auto pr-1 lg:pr-2">
		{#if filteredMenus.length > 0}
			{#each filteredMenus as group (group.group)}
				<div class="group-label mt-5 mb-2 pl-3 text-[10px] font-black tracking-widest text-muted-foreground uppercase opacity-60 first:mt-0">
					{group.group}
				</div>
				<div class="space-y-1">
					{#each group.items as item}
						{@const href = resolve(item.href as '/')}
						{@const isActive = page.url.pathname === href || (item.href !== '/' && page.url.pathname.startsWith(item.href) && item.href.length > 1)}
						{@const icon = getIconPath(item.icon)}

						<a
							href={href}
							class="admin-nav-item group flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-semibold transition-all lg:px-4 lg:py-2 {isActive
								? 'bg-primary text-primary-foreground shadow-sm shadow-primary/15'
								: 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
							onclick={() => (isMobileMenuOpen = false)}
						>
							<span class="icon flex h-7 w-7 shrink-0 items-center justify-center rounded-lg text-base leading-none {isActive ? 'bg-white/15 text-primary-foreground' : 'bg-muted/60 text-muted-foreground group-hover:text-primary'}">
								{#if icon}
									<svg class="h-4 w-4" fill="none" stroke="currentColor" stroke-width={isActive ? 2.2 : 1.6} viewBox={icon.viewBox}>
										<path stroke-linecap="round" stroke-linejoin="round" d={icon.path} />
									</svg>
								{/if}
							</span>
							<span class="label truncate">{item.label}</span>
						</a>
					{/each}
				</div>
			{/each}
		{:else}
			<div class="p-4 text-xs text-muted-foreground italic opacity-50">
				Menyiapkan menu akses...
			</div>
		{/if}
	</nav>

	<!-- Footer -->
	<div class="sidebar-footer mt-3 shrink-0 border-t border-border pt-3">
		<div class="mb-2 rounded-xl border border-border bg-muted/35 p-2.5">
			<div class="flex items-center gap-2">
				<div class="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
					<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
					</svg>
				</div>
				<div class="min-w-0">
					<p class="text-[9px] font-bold tracking-wide text-muted-foreground">Masuk sebagai</p>
					<p class="truncate text-[11px] font-black text-foreground">{user?.username || 'Anonymous'}</p>
				</div>
			</div>
		</div>
		<a
			href="/logout"
			class="logout-btn flex w-full items-center justify-center gap-2 rounded-xl border border-destructive/20 bg-destructive/5 py-2.5 text-[11px] font-bold text-destructive transition-all hover:bg-destructive hover:text-white lg:py-2"
		>
			<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
			</svg>
			Keluar
		</a>
		<div class="mt-2 rounded-xl border border-primary/10 bg-primary/5 px-2.5 py-1.5">
			<p class="text-[9px] font-black tracking-wide text-foreground">Super App {branding?.short_name || 'MTsN 2 Kolut'}</p>
			<p class="mt-0.5 text-[9px] font-medium leading-snug text-muted-foreground">v2 · by Hasbi Awal</p>
		</div>
	</div>
</aside>