<script lang="ts">
	import { onMount } from 'svelte';
	import PublicSiteShell from '$lib/components/PublicSiteShell.svelte';
	import AdminBreadcrumb from '$lib/components/breadcrumb/AdminBreadcrumb.svelte';
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
	let isExamFallbackPage = $derived(false);
	let pwaRegistrationStarted = $state(false);
	let isPublicSite = $derived(isPublicSitePath(page.url.pathname, Boolean(data.user)));
	let branding = $derived(data.branding ?? defaultBranding);
	let desktopSidebarExpanded = $state(true);
	let desktopSidebarWidth = $state(240);
	const SIDEBAR_EXPANDED_STORAGE_KEY_PREFIX = 'sidebar:desktop-expanded';
	const SIDEBAR_WIDTH_STORAGE_KEY_PREFIX = 'sidebar:desktop-width';
	const SIDEBAR_MIN_WIDTH = 220;
	const SIDEBAR_MAX_WIDTH = 360;

	function sidebarExpandedStorageKey() {
		return `${SIDEBAR_EXPANDED_STORAGE_KEY_PREFIX}:${data.user?.id ?? 'anon'}`;
	}

	function sidebarWidthStorageKey() {
		return `${SIDEBAR_WIDTH_STORAGE_KEY_PREFIX}:${data.user?.id ?? 'anon'}`;
	}

	function clampSidebarWidth(value: number) {
		return Math.min(SIDEBAR_MAX_WIDTH, Math.max(SIDEBAR_MIN_WIDTH, Math.round(value)));
	}

	onMount(() => {
		const raw = window.localStorage.getItem(sidebarExpandedStorageKey());
		if (raw === '0') desktopSidebarExpanded = false;
		if (raw === '1') desktopSidebarExpanded = true;

		const rawWidth = Number(window.localStorage.getItem(sidebarWidthStorageKey()));
		if (Number.isFinite(rawWidth)) {
			desktopSidebarWidth = clampSidebarWidth(rawWidth);
		}
	});

	$effect(() => {
		if (typeof window === 'undefined') return;
		window.localStorage.setItem(sidebarExpandedStorageKey(), desktopSidebarExpanded ? '1' : '0');
	});

	$effect(() => {
		if (typeof window === 'undefined') return;
		window.localStorage.setItem(sidebarWidthStorageKey(), String(clampSidebarWidth(desktopSidebarWidth)));
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

{#if isLogin || isMaintenancePage || isExamFallbackPage}
	{@render children()}
{:else if isPublicSite}
	<PublicSiteShell user={data.user} branding={branding}>
		{@render children()}
	</PublicSiteShell>
{:else}
	<div
		class="flex min-h-screen bg-background text-foreground"
		style={`--sidebar-width: ${clampSidebarWidth(desktopSidebarWidth)}px;`}
	>
		<Sidebar
			bind:desktopExpanded={desktopSidebarExpanded}
			bind:desktopWidth={desktopSidebarWidth}
			user={data.user}
			account={data.account}
			branding={branding}
		/>
		<!--
			--sidebar-width — adjustable expanded fixed sidebar width on desktop
			lg:pl-[5.5rem]  — offset for collapsed fixed sidebar on desktop
			pt-14           — offset for fixed 56px mobile topbar (rendered by Sidebar)
			lg:pt-0         — no offset needed on desktop (topbar hidden)
		-->
		<div class={`flex-1 min-w-0 pt-14 transition-[padding] duration-200 lg:pt-0 ${desktopSidebarExpanded ? 'lg:pl-[var(--sidebar-width)]' : 'lg:pl-[5.5rem]'}`}>
			<main class={`w-full px-2.5 py-4 sm:px-4 sm:py-6 lg:px-6 ${desktopSidebarExpanded ? 'lg:mx-auto lg:max-w-[1100px]' : 'lg:max-w-[1380px]'}`}>
				<MaintenanceBanner status={data.maintenanceStatus} user={data.user} />
				<AdminBreadcrumb user={data.user} />
				{@render children()}
			</main>
		</div>
	</div>
{/if}
