<script lang="ts">
	import { onMount } from 'svelte';

	onMount(async () => {
		try {
			await fetch('/api/auth/logout', { method: 'POST' });
		} catch { /* best effort */ }
		// Full page redirect to clear all SvelteKit state + force layout re-check
		window.location.href = '/login';
	});
</script>

<div class="flex min-h-screen items-center justify-center bg-base-200 p-6">
	<div class="text-center space-y-4">
		<div class="mx-auto w-10 h-10 border-[3px] border-primary border-t-transparent rounded-full animate-spin"></div>
		<div>
			<p class="text-sm font-semibold text-foreground">Keluar dari sistem...</p>
			<p class="text-[11px] text-muted-foreground mt-1">dibuat oleh Hasbi Awal, S. Kom.</p>
		</div>
	</div>
</div>
