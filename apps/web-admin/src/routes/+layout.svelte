<script lang="ts">
	import { onMount } from 'svelte';
	import favicon from '$lib/assets/favicon.svg';
	import PublicSiteShell from '$lib/components/PublicSiteShell.svelte';
	import RouteProgress from '$lib/components/RouteProgress.svelte';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import { Sonner } from '$lib/components/ui/sonner';
	import '../app.css';
	import { navigating, page } from '$app/state';

	let { children, data } = $props();
	let isLogin = $derived(page.url.pathname === '/login');
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
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<Sonner />
<RouteProgress active={!!navigating.to} />

{#if isLogin}
	{@render children()}
{:else if isPublicSite}
	<PublicSiteShell user={data.user}>
		{@render children()}
	</PublicSiteShell>
{:else}
	<div class="flex min-h-screen bg-slate-50">
		<Sidebar bind:desktopExpanded={desktopSidebarExpanded} user={data.user} />
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
