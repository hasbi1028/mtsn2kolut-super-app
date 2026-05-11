<script lang="ts">
	import SoalWorkspacePage from '../_components/SoalWorkspacePage.svelte';

	type PageData = {
		user?: {
			role?: string;
			roles?: string[];
		};
	};

	let { data }: { data: PageData } = $props();
</script>

<svelte:head>
	<title>Penyusun soal Soal - MTsN 2 Kolaka Utara</title>
</svelte:head>

<SoalWorkspacePage {data} routeMode="composer" />
