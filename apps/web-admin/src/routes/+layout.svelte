<script lang="ts">
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
	let isPublicSite = $derived(isPublicSitePath(page.url.pathname, Boolean(data.user)));
	let branding = $derived(data.branding ?? defaultBranding);

	$effect(() => {
		if (typeof window === 'undefined') return;
		if (!data.user || isLogin || isPublicSite) return;
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

<!-- Force light theme for admin, removing dark mode background logic -->
<div data-theme="cerberus" class="min-h-screen w-full max-w-full overflow-x-hidden bg-surface-50 text-surface-900">
	{#if isLogin || isMaintenancePage}
		{@render children()}
	{:else if isPublicSite}
		<PublicSiteShell user={data.user} branding={branding}>
			{@render children()}
		</PublicSiteShell>
	{:else}
		<div class="flex min-h-screen w-full max-w-full">
			<Sidebar user={data.user} account={data.account} branding={branding} />
			<div class="flex-1 min-w-0 w-full max-w-full pt-14 lg:pt-0 lg:pl-64">
				<main class="w-full max-w-none px-4 pb-24 pt-4 sm:px-6 lg:px-8 lg:py-6 page-enter">
					<MaintenanceBanner status={data.maintenanceStatus} user={data.user} />
					{@render children()}
				</main>
			</div>
		</div>
	{/if}
</div>
