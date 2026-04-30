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

<svelte:head><title>Pengumuman — MTsN 2 Kolaka Utara</title></svelte:head>

<PublicContentList
	title="Pengumuman Sekolah"
	description="Informasi penting, jadwal, dan pemberitahuan resmi untuk warga sekolah dan masyarakat."
	items={data.items}
	basePath="/pengumuman"
	eyebrow="Pengumuman"
/>
