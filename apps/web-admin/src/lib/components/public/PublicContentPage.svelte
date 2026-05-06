<script lang="ts">
	type WebsiteContent = {
		title: string;
		excerpt: string;
		content_html: string;
		cover_image_url: string;
		published_at: string | null;
	};

	type SidePanel = {
		eyebrow?: string;
		title: string;
		tone?: 'default' | 'emerald';
		lines: string[];
	};

	let {
		content,
		eyebrow,
		sidePanels = [],
	}: {
		content: WebsiteContent;
		eyebrow: string;
		sidePanels?: SidePanel[];
	} = $props();

	function fmtDate(value: string | null) {
		if (!value) return '';
		return new Date(value).toLocaleDateString('id-ID', {
			timeZone: 'Asia/Makassar',
			year: 'numeric',
			month: 'long',
			day: 'numeric',
		});
	}

	const panels = $derived.by<SidePanel[]>(() => {
		if (sidePanels.length > 0) return sidePanels;
		return [
			{
				eyebrow: 'Ringkasan',
				title: 'Informasi Utama',
				lines: [
					`Kategori: ${eyebrow}`,
					content.published_at ? `Tanggal tayang: ${fmtDate(content.published_at)}` : '',
					'Sumber: Website resmi MTsN 2 Kolaka Utara',
				].filter(Boolean),
			},
			{
				eyebrow: 'Catatan Baca',
				title: 'Panduan Singkat',
				tone: 'emerald',
				lines: [
					'Informasi ini disiapkan untuk warga madrasah dan masyarakat umum.',
					'Jika memuat jadwal atau ketentuan, gunakan tanggal tayang sebagai acuan terbaru.',
				],
			},
		];
	});
</script>

<article class="mx-auto max-w-6xl space-y-8">
	<header class="space-y-5 rounded-[2rem] border border-border bg-card px-6 py-8 shadow-sm sm:px-8 sm:py-10">
		<p class="text-xs font-semibold uppercase tracking-[0.22em] text-primary">{eyebrow}</p>
		<h1 class="max-w-4xl text-3xl font-bold leading-tight text-foreground sm:text-5xl">{content.title}</h1>
		{#if content.excerpt}
			<p class="max-w-3xl text-base leading-8 text-muted-foreground sm:text-lg">{content.excerpt}</p>
		{/if}
		<div class="flex flex-wrap items-center gap-3 text-sm text-muted-foreground">
			{#if content.published_at}
				<span class="rounded-full border border-primary/20 bg-card px-3 py-1">Tayang {fmtDate(content.published_at)}</span>
			{/if}
			<span class="rounded-full border border-border bg-card px-3 py-1">Informasi resmi MTsN 2 Kolaka Utara</span>
		</div>
	</header>

	{#if content.cover_image_url}
		<div class="overflow-hidden rounded-[2rem] border border-border bg-card shadow-sm">
			<img src={content.cover_image_url} alt={content.title} class="h-[240px] w-full object-cover sm:h-[360px] lg:h-[420px]" />
		</div>
	{/if}

	<div class="grid gap-8 lg:grid-cols-[minmax(0,1fr)_18rem] lg:items-start">
		<div class="rounded-[2rem] border border-border bg-card px-6 py-6 shadow-sm sm:px-8 sm:py-8">
			<div class="max-w-none space-y-6 text-base leading-8 text-foreground [&_a]:font-medium [&_a]:text-primary [&_a]:underline-offset-4 hover:[&_a]:text-primary [&_a:hover]:underline [&_blockquote]:rounded-2xl [&_blockquote]:border-l-4 [&_blockquote]:border-primary/20 [&_blockquote]:bg-primary/10 [&_blockquote]:px-5 [&_blockquote]:py-4 [&_blockquote]:text-foreground [&_h1]:text-4xl [&_h1]:font-bold [&_h1]:text-foreground [&_h2]:mt-10 [&_h2]:text-3xl [&_h2]:font-semibold [&_h2]:text-foreground [&_h3]:mt-8 [&_h3]:text-2xl [&_h3]:font-semibold [&_h3]:text-foreground [&_img]:rounded-[1.5rem] [&_img]:shadow-sm [&_li]:leading-8 [&_p]:leading-8 [&_strong]:text-foreground">
				{@html content.content_html}
			</div>
		</div>

		<aside class="space-y-4 lg:sticky lg:top-24">
			{#each panels as panel (panel.title)}
				<div class={`rounded-[1.75rem] px-5 py-5 shadow-sm ${
					panel.tone === 'emerald'
						? 'border border-primary/20 bg-primary/10'
						: 'border border-border bg-card'
				}`}>
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-primary">
						{panel.eyebrow || 'Ringkasan'}
					</p>
					<h2 class="mt-2 text-base font-semibold text-foreground">{panel.title}</h2>
					<div class={`mt-4 space-y-3 text-sm leading-7 ${
						panel.tone === 'emerald' ? 'text-foreground' : 'text-muted-foreground'
					}`}>
						{#each panel.lines as line (line)}
							<p>{line}</p>
						{/each}
					</div>
				</div>
			{/each}
		</aside>
	</div>
</article>
