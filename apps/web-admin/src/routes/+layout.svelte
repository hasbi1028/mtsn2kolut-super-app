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
	import { fade } from 'svelte/transition';

	let { children, data } = $props();
	let isLogin = $derived(page.url.pathname === '/login');
	let isMaintenancePage = $derived(page.url.pathname === '/maintenance');
	let isPublicSite = $derived(isPublicSitePath(page.url.pathname, Boolean(data.user)));
	let branding = $derived(data.branding ?? defaultBranding);
	let isMobileMenuOpen = $state(false);

	const user = $derived(data.user);

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

	$effect(() => {
		if (navigating.to) isMobileMenuOpen = false;
	});
</script>

<svelte:head>
	<link rel="icon" href={versionedAsset(branding.mark_url, branding.version)} />
	<link rel="icon" href={versionedAsset(branding.favicon_url, branding.version)} sizes="any" />
	<link rel="apple-touch-icon" href={versionedAsset(branding.apple_touch_icon_url, branding.version)} />
	<link rel="manifest" href={versionedAsset('/manifest.webmanifest', branding.version)} />
	<meta name="theme-color" content={branding.theme_color} />
</svelte:head>

<svelte:body class:overflow-hidden={isMobileMenuOpen} />

<Sonner />
<GlobalConfirmDialog />
<RouteProgress active={!!navigating.to} />

{#if isLogin || isMaintenancePage}
	<div class="min-h-screen bg-base-200 text-base-content">
		{@render children()}
	</div>
{:else if isPublicSite}
	<div class="min-h-screen bg-base-200 text-base-content">
		<PublicSiteShell user={data.user} branding={branding}>
			{@render children()}
		</PublicSiteShell>
	</div>
{:else}
	<!-- ══ CBT-style Layout ══ -->
	<div class="admin-layout-ui flex min-h-screen bg-base-200 text-base-content">

		<!-- MOBILE HEADER -->
		<header class="mobile-header fixed top-0 right-0 left-0 z-[110] flex h-14 items-center justify-between border-b border-base-300 bg-base-100 px-4 shadow-sm lg:hidden">
			<button
				class="rounded-md p-2 text-base-content hover:bg-base-200"
				onclick={() => (isMobileMenuOpen = !isMobileMenuOpen)}
				aria-label="Menu"
				aria-expanded={isMobileMenuOpen}
			>
				<svg class="h-5 w-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
					{#if isMobileMenuOpen}
						<path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
					{:else}
						<path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16" />
					{/if}
				</svg>
			</button>
			<div class="brand-mobile flex min-w-0 flex-1 items-center justify-center gap-2 font-black tracking-tighter text-base-content uppercase italic">
				<img
					src={versionedAsset(branding.mark_url, branding.version)}
					alt="Logo"
					class="h-8 w-8 rounded-lg border border-base-300 bg-base-200/80 object-contain p-1 shadow-sm"
				/>
				<span class="truncate">{branding.short_name}</span>
			</div>
			<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary text-xs font-bold text-primary-content shadow-sm">
				{user?.username?.substring(0, 2).toUpperCase() || '??'}
			</div>
		</header>

		<!-- SIDEBAR -->
		<Sidebar
			{user}
			{branding}
			bind:isMobileMenuOpen
		/>

		{#if isMobileMenuOpen}
			<div
				class="sidebar-overlay fixed inset-0 z-[115] bg-base-200/80 backdrop-blur-sm lg:hidden"
				transition:fade={{ duration: 200 }}
				onclick={() => (isMobileMenuOpen = false)}
				role="presentation"
			></div>
		{/if}

		<!-- MAIN CONTENT -->
		<main class="admin-main flex min-h-screen min-w-0 flex-1 flex-col">
			<!-- Desktop Top Header -->
			<header class="admin-header z-10 hidden h-16 shrink-0 items-center justify-between border-b border-base-300 bg-base-100 px-8 lg:flex">
				<h1 class="page-title text-xl font-black tracking-tight text-base-content">
					Super App Command Center
				</h1>
				<div class="header-tools flex items-center gap-4">
					<div class="text-sm font-semibold text-base-content/70">
						{new Date().toLocaleDateString('id-ID', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric', timeZone: 'Asia/Makassar' })}
					</div>
				</div>
			</header>

			<div class="admin-content bg-base-200 pt-14 lg:flex-1 lg:pt-0">
				<MaintenanceBanner status={data.maintenanceStatus} user={data.user} />
				<div class="content-wrapper mx-auto w-full max-w-[1280px] px-4 py-5 sm:px-5 md:px-8 lg:px-10 lg:py-10">
					{#if navigating.to}
						<div class="space-y-6">
							<div class="skeleton h-8 w-56"></div>
							<div class="grid grid-cols-4 gap-4">
								<div class="skeleton h-24 rounded-xl"></div>
								<div class="skeleton h-24 rounded-xl"></div>
								<div class="skeleton h-24 rounded-xl"></div>
								<div class="skeleton h-24 rounded-xl"></div>
							</div>
							<div class="space-y-3">
								<div class="skeleton h-8 w-48"></div>
								<div class="skeleton h-14 w-full"></div>
								<div class="skeleton h-14 w-full"></div>
								<div class="skeleton h-14 w-full"></div>
								<div class="skeleton h-14 w-full"></div>
								<div class="skeleton h-14 w-full"></div>
							</div>
						</div>
					{:else}
						{@render children()}
					{/if}
				</div>
			</div>
		</main>

	</div>
{/if}
