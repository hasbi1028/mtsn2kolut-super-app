<script lang="ts">
	import { onMount } from 'svelte';
	import PublicSiteShell from '$lib/components/PublicSiteShell.svelte';
	import GlobalConfirmDialog from '$lib/components/GlobalConfirmDialog.svelte';
	import RouteProgress from '$lib/components/RouteProgress.svelte';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import { Sonner } from '$lib/components/ui/sonner';
	import '../app.css';
	import { navigating, page } from '$app/state';

	let { children, data } = $props();
	let isLogin = $derived(page.url.pathname === '/login');
	let pwaRegistrationStarted = $state(false);
	const publicExactPaths = new Set(['/', '/ppdb', '/profil', '/berita', '/pengumuman', '/kontak']);
	const publicPrefixPaths = ['/berita/', '/pengumuman/'];
	let isPublicSite = $derived.by(() => {
		const pathname = page.url.pathname;
		if (pathname === '/' && data.user) return false;
		return publicExactPaths.has(pathname) || publicPrefixPaths.some((prefix) => pathname.startsWith(prefix));
	});
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
	<link rel="icon" type="image/svg+xml" href="/brand/m2k-mark.svg" />
	<link rel="icon" href="/favicon.ico" sizes="any" />
	<link rel="icon" type="image/png" sizes="16x16" href="/favicon-16x16.png" />
	<link rel="icon" type="image/png" sizes="32x32" href="/favicon-32x32.png" />
	<link rel="apple-touch-icon" href="/apple-touch-icon.png" />
	<link rel="manifest" href="/manifest.webmanifest" />
	<meta name="theme-color" content="#166534" />
</svelte:head>

<Sonner />
<GlobalConfirmDialog />
<RouteProgress active={!!navigating.to} />

{#if isLogin}
	{@render children()}
{:else if isPublicSite}
	<PublicSiteShell user={data.user}>
		{@render children()}
	</PublicSiteShell>
{:else}
	<div class="flex min-h-screen bg-background text-foreground">
		<Sidebar bind:desktopExpanded={desktopSidebarExpanded} user={data.user} account={data.account} />
		<!--
			lg:pl-60      — offset for expanded fixed sidebar on desktop
			lg:pl-[5.5rem] — offset for collapsed fixed sidebar on desktop
			pt-14     — offset for fixed 56px mobile topbar (rendered by Sidebar)
			lg:pt-0   — no offset needed on desktop (topbar hidden)
		-->
		<div class={`flex-1 min-w-0 pt-14 transition-[padding] duration-200 lg:pt-0 ${desktopSidebarExpanded ? 'lg:pl-60' : 'lg:pl-[5.5rem]'}`}>
			<main class={`w-full px-2.5 py-4 sm:px-4 sm:py-6 lg:px-6 ${desktopSidebarExpanded ? 'lg:mx-auto lg:max-w-[1100px]' : 'lg:max-w-[1380px]'}`}>
				{@render children()}
			</main>
		</div>
	</div>
{/if}
