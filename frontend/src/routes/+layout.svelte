<script lang="ts">
  import favicon from '$lib/assets/favicon.svg';
  import Nav from '$lib/components/Nav.svelte';
  import '../app.css';
  import { page } from '$app/state';

  let { children, data } = $props();
  let isLogin = $derived(page.url.pathname === '/login');
</script>

<svelte:head>
  <link rel="icon" href={favicon} />
</svelte:head>

{#if !isLogin}
  <Nav user={data.user} />
  <main class="page">
    {@render children()}
  </main>
{:else}
  {@render children()}
{/if}
