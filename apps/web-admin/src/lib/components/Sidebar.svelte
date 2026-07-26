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
					// If roles is explicitly set and user doesn't match → hide (even if permission matches)
					if (item.roles && !item.roles.some((r: string) => userRoles.includes(r))) return false;
					if (item.permissions && item.permissions.some((p: string) => userPermissions.includes(p))) return true;
					if (item.roles && item.roles.some((r: string) => userRoles.includes(r))) return true;
					if (item.allowAuthenticatedFallback && user) return true;
					return false;
				});
				return { ...group, items };
			})
			.filter((group) => group.items.length > 0);
	});

	// Icon map — font/emoji based (instead of SVG)
	const iconFontMap: Record<string, string> = {
		home: '⌂',
		clock: '⏰',
		users: '👥',
		calendar: '📅',
		'bar-chart': '📊',
		send: '📨',
		'refresh-cw': '🔄',
		user: '👤',
		building: '🏢',
		shield: '🛡',
		image: '🖼',
		database: '💾',
		'file-text': '📄',
		menu: '☰',
		'x-circle': '✕',
	};

	const getIconFont = (iconName: string) => iconFontMap[iconName] ?? '○';
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
						{@const pathParts = href.split('/').filter(Boolean)}
						{@const isActive = page.url.pathname === href || (pathParts.length >= 2 && page.url.pathname.startsWith(href + '/'))}
						{@const iconFont = getIconFont(item.icon)}

						<a
							href={href}
							class="admin-nav-item group flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-semibold transition-all lg:px-4 lg:py-2 {isActive
								? 'bg-primary/5 text-primary border-l-2 border-primary rounded-l-none'
								: 'text-muted-foreground hover:bg-muted hover:text-foreground border-l-2 border-transparent'}"
							onclick={() => (isMobileMenuOpen = false)}
						>
							<span class="icon flex h-6 w-6 shrink-0 items-center justify-center text-sm leading-none {isActive ? 'text-primary' : 'text-muted-foreground'}">
								{iconFont}
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
			data-sveltekit-reload
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
