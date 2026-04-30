<script lang="ts">
	type WebsiteContent = {
		title: string;
		excerpt: string;
		content_html: string;
		cover_image_url: string;
		published_at: string | null;
	};

	let { content, eyebrow }: { content: WebsiteContent; eyebrow: string } = $props();

	function fmtDate(value: string | null) {
		if (!value) return '';
		return new Date(value).toLocaleDateString('id-ID', {
			timeZone: 'Asia/Makassar',
			year: 'numeric',
			month: 'long',
			day: 'numeric',
		});
	}
</script>

<article class="mx-auto max-w-4xl space-y-8">
	<header class="space-y-4">
		<p class="text-xs font-semibold uppercase tracking-[0.18em] text-emerald-700">{eyebrow}</p>
		<h1 class="text-4xl font-bold leading-tight text-slate-900 sm:text-5xl">{content.title}</h1>
		{#if content.excerpt}
			<p class="max-w-3xl text-lg leading-8 text-slate-600">{content.excerpt}</p>
		{/if}
		{#if content.published_at}
			<p class="text-sm text-slate-500">{fmtDate(content.published_at)}</p>
		{/if}
	</header>

	{#if content.cover_image_url}
		<img src={content.cover_image_url} alt={content.title} class="h-[260px] w-full rounded-[2rem] object-cover shadow-sm sm:h-[360px]" />
	{/if}

	<div class="max-w-none space-y-6 text-base leading-8 text-slate-700 [&_a]:font-medium [&_a]:text-emerald-800 [&_a]:underline-offset-4 hover:[&_a]:text-emerald-900 [&_a:hover]:underline [&_h1]:text-4xl [&_h1]:font-bold [&_h1]:text-slate-900 [&_h2]:text-3xl [&_h2]:font-semibold [&_h2]:text-slate-900 [&_h3]:text-2xl [&_h3]:font-semibold [&_h3]:text-slate-900 [&_img]:rounded-[1.5rem] [&_img]:shadow-sm [&_li]:leading-8 [&_p]:leading-8 [&_strong]:text-slate-900">
		{@html content.content_html}
	</div>
</article>
