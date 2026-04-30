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
	let { data }: { data: { post: PageContent } } = $props();
	const seoTitle = $derived(data.post.meta_title || data.post.title);
	const seoDesc = $derived(data.post.meta_description || data.post.excerpt);
</script>

<svelte:head>
	<title>{seoTitle} — Berita MTsN 2 Kolaka Utara</title>
	{#if seoDesc}<meta name="description" content={seoDesc} />{/if}
	<meta property="og:title" content={seoTitle} />
	{#if seoDesc}<meta property="og:description" content={seoDesc} />{/if}
	{#if data.post.cover_image_url}<meta property="og:image" content={data.post.cover_image_url} />{/if}
	<meta property="og:type" content="article" />
</svelte:head>

<PublicContentPage content={data.post} eyebrow="Berita Madrasah" />
