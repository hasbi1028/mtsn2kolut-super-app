<script lang="ts">
	import ReviewBankSoalPage from '../_components/ReviewBankSoalPage.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
</script>

<ReviewBankSoalPage {data} />
