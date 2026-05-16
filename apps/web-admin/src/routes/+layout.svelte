<script lang="ts">
	import { onMount } from 'svelte';
	import PublicSiteShell from '$lib/components/PublicSiteShell.svelte';
	import GlobalConfirmDialog from '$lib/components/GlobalConfirmDialog.svelte';
	import MaintenanceBanner from '$lib/components/maintenance/MaintenanceBanner.svelte';
	import RouteProgress from '$lib/components/RouteProgress.svelte';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import { Sonner } from '$lib/components/ui/sonner';
	import { isPublicSitePath } from '$lib/routes/public-policy';
	import { defaultBranding, versionedAsset } from '$lib/branding';
	import '../app.css';
	import { navigating, page } from '$app/state';

	let { children, data } = $props();
	let isLogin = $derived(page.url.pathname === '/login');
	let isMaintenancePage = $derived(page.url.pathname === '/maintenance');
	let pwaRegistrationStarted = $state(false);
	let isPublicSite = $derived(isPublicSitePath(page.url.pathname, Boolean(data.user)));
	let branding = $derived(data.branding ?? defaultBranding);
	let desktopSidebarExpanded = $state(true);
	const SIDEBAR_EXPANDED_STORAGE_KEY_PREFIX = 'sidebar:desktop-expanded';

	function sidebarExpandedStorageKey() {
		return `${SIDEBAR_EXPANDED_STORAGE_KEY_PREFIX}:${data.user?.id ?? 'anon'}`;
	}

	onMount(() => {
		const raw = window.localStorage.getItem(sidebarExpandedStorageKey());
		if (raw === '0') desktopSidebarExpanded = false;
		if (raw === '1') desktopSidebarExpanded = true;
	});

	$effect(() => {
		if (typeof window === 'undefined') return;
		window.localStorage.setItem(sidebarExpandedStorageKey(), desktopSidebarExpanded ? '1' : '0');
	});

	$effect(() => {
		if (typeof window === 'undefined' || pwaRegistrationStarted) return;
		if (!data.user || isLogin || isPublicSite) return;
		pwaRegistrationStarted = true;
		void import('$lib/client/pwa').then(({ registerWebAdminPwa }) => registerWebAdminPwa({
			isAuthenticated: Boolean(data.user),
			isLogin,
			isPublicSite,
			location: window.location,
			serviceWorker: navigator.serviceWorker
		})).catch(() => undefined);
	});
</script>

<svelte:head>
	<link rel="icon" href={versionedAsset(branding.mark_url, branding.version)} />
	<link rel="icon" href={versionedAsset(branding.favicon_url, branding.version)} sizes="any" />
	<link rel="apple-touch-icon" href={versionedAsset(branding.apple_touch_icon_url, branding.version)} />
	<link rel="manifest" href={versionedAsset('/manifest.webmanifest', branding.version)} />
	<meta name="theme-color" content={branding.theme_color} />
</svelte:head>

<Sonner />
<GlobalConfirmDialog />
<RouteProgress active={!!navigating.to} />

{#if isLogin || isMaintenancePage}
	{@render children()}
{:else if isPublicSite}
	<PublicSiteShell user={data.user} branding={branding}>
		{@render children()}
	</PublicSiteShell>
{:else}
	<div class="flex min-h-screen bg-background text-foreground">
		<Sidebar bind:desktopExpanded={desktopSidebarExpanded} user={data.user} account={data.account} branding={branding} />
		<!--
			lg:pl-60      — offset for expanded fixed sidebar on desktop
			lg:pl-[5.5rem] — offset for collapsed fixed sidebar on desktop
			pt-14     — offset for fixed 56px mobile topbar (rendered by Sidebar)
			lg:pt-0   — no offset needed on desktop (topbar hidden)
		-->
		<div class={`flex-1 min-w-0 pt-14 transition-[padding] duration-200 lg:pt-0 ${desktopSidebarExpanded ? 'lg:pl-60' : 'lg:pl-[5.5rem]'}`}>
			<main class={`w-full px-2.5 py-4 sm:px-4 sm:py-6 lg:px-6 ${desktopSidebarExpanded ? 'lg:mx-auto lg:max-w-[1100px]' : 'lg:max-w-[1380px]'}`}>
				<MaintenanceBanner status={data.maintenanceStatus} user={data.user} />
				{@render children()}
			</main>
		</div>
	</div>
{/if}
