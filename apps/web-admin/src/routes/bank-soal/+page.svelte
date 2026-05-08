<script lang="ts">
	import BankSoalHealthDashboard from './_components/BankSoalHealthDashboard.svelte';

	type PageData = {
		user?: {
			role?: string;
			roles?: string[];
			permissions?: string[];
		};
	};

	let { data }: { data: PageData } = $props();
</script>

<BankSoalHealthDashboard {data} />
