<script lang="ts">
	import SoalWorkspacePage from '../../_components/SoalWorkspacePage.svelte';

	type PageData = {
		questionId: string;
		user?: {
			role?: string;
			roles?: string[];
			permissions?: string[];
		};
	};

	let { data }: { data: PageData } = $props();
</script>

<svelte:head>
	<title>Detail Soal - MTsN 2 Kolaka Utara</title>
</svelte:head>

<SoalWorkspacePage {data} routeMode="composer" questionId={data.questionId} />
