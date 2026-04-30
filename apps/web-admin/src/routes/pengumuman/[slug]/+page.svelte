<script lang="ts">
	import PublicContentPage from '$lib/components/public/PublicContentPage.svelte';
	type PageContent = {
		title: string;
		excerpt: string;
		content_html: string;
		cover_image_url: string;
		published_at: string | null;
	};
	let { data }: { data: { announcement: PageContent } } = $props();
</script>

<svelte:head><title>{data.announcement.title} — Pengumuman MTsN 2 Kolaka Utara</title></svelte:head>

<PublicContentPage content={data.announcement} eyebrow="Pengumuman Resmi" />
