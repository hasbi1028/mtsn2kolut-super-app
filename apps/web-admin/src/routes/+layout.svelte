<script lang="ts">
	import favicon from '$lib/assets/favicon.svg';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import '../app.css';
	import { page } from '$app/state';

	let { children, data } = $props();
	let isLogin = $derived(page.url.pathname === '/login');
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{#if !isLogin}
	<div class="flex min-h-screen bg-slate-50">
		<Sidebar user={data.user} />
		<!--
			lg:pl-60  — offset for fixed 240px sidebar on desktop
			pt-14     — offset for fixed 56px mobile topbar (rendered by Sidebar)
			lg:pt-0   — no offset needed on desktop (topbar hidden)
		-->
		<div class="flex-1 min-w-0 lg:pl-60 pt-14 lg:pt-0">
			<main class="max-w-[1100px] mx-auto px-3 py-4 sm:px-4 sm:py-6">
				{@render children()}
			</main>
		</div>
	</div>
{:else}
	{@render children()}
{/if}
