<script lang="ts">
	import { onMount } from 'svelte';
	import { trackInternalAnalyticsEvent } from '$lib/analytics/internal-analytics';
	import BankSoalHealthDashboard from './_components/BankSoalHealthDashboard.svelte';

	type PageData = {
		user?: {
			role?: string;
			roles?: string[];
			permissions?: string[];
		};
	};

	let { data }: { data: PageData } = $props();

	onMount(() => {
		void trackInternalAnalyticsEvent('bank_soal.list_view', {
			pathname: window.location.pathname,
			role: data.user?.roles?.[0] ?? data.user?.role,
			metadata: { page_key: 'bank_soal' }
		});
	});
</script>

<BankSoalHealthDashboard {data} />
