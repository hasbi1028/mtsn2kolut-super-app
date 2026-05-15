<script lang="ts">
	import PublishBankSoalPage from '../_components/PublishBankSoalPage.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
</script>

<PublishBankSoalPage {data} />
