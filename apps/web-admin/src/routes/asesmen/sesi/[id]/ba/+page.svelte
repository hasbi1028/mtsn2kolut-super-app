<script lang="ts">
	import SessionPrintPage from '../_SessionPrintPage.svelte';

	let { data }: { data: { id: string } } = $props();
</script>

<SessionPrintPage id={data.id} kind="ba" />
