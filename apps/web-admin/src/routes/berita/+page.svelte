<script lang="ts">
	import PublicContentList from '$lib/components/public/PublicContentList.svelte';
	type ListItem = {
		id: string;
		title: string;
		slug: string;
		excerpt: string;
		cover_image_url: string;
		published_at: string | null;
	};
	let { data }: { data: { items: ListItem[] } } = $props();
</script>

<svelte:head><title>Berita — MTsN 2 Kolaka Utara</title></svelte:head>

<PublicContentList
	title="Berita Madrasah"
	description="Rangkuman kegiatan, capaian, dan kabar terbaru dari MTs Negeri 2 Kolaka Utara."
	items={data.items}
	basePath="/berita"
	eyebrow="Berita"
/>
