<script lang="ts">
	import BankSoalListPage from '../_components/BankSoalListPage.svelte';

	type PageData = {
		user?: {
			role?: string;
			roles?: string[];
		};
	};

	let { data }: { data: PageData } = $props();
</script>

<BankSoalListPage {data} mode="list" />
