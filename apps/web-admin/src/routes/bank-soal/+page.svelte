<script lang="ts">
	import BankSoalHealthDashboard from './_components/BankSoalHealthDashboard.svelte';
	import BankSoalListPage from './_components/BankSoalListPage.svelte';

	type PageData = {
		user?: {
			role?: string;
			roles?: string[];
			permissions?: string[];
		};
	};

	let { data }: { data: PageData } = $props();
</script>

<div class="space-y-6">
	<BankSoalHealthDashboard {data} />
	<BankSoalListPage {data} mode="dashboard" />
</div>
