<script lang="ts">
	import PublicContentPage from '$lib/components/public/PublicContentPage.svelte';
	type PageContent = {
		title: string;
		excerpt: string;
		content_html: string;
		cover_image_url: string;
		meta_title: string;
		meta_description: string;
		published_at: string | null;
	};
	let { data }: { data: { announcement: PageContent } } = $props();
	const seoTitle = $derived(data.announcement.meta_title || data.announcement.title);
	const seoDesc = $derived(data.announcement.meta_description || data.announcement.excerpt);
</script>

<svelte:head>
	<title>{seoTitle} — Pengumuman MTsN 2 Kolaka Utara</title>
	{#if seoDesc}<meta name="description" content={seoDesc} />{/if}
	<meta property="og:title" content={seoTitle} />
	{#if seoDesc}<meta property="og:description" content={seoDesc} />{/if}
	{#if data.announcement.cover_image_url}<meta property="og:image" content={data.announcement.cover_image_url} />{/if}
	<meta property="og:type" content="article" />
</svelte:head>

<PublicContentPage content={data.announcement} eyebrow="Pengumuman Resmi" />
