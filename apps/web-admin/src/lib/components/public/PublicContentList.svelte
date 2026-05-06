<script lang="ts">
	type WebsiteContent = {
		id: string;
		title: string;
		slug: string;
		excerpt: string;
		cover_image_url: string;
		published_at: string | null;
	};

	let {
		title,
		description,
		items,
		basePath,
		eyebrow,
	}: {
		title: string;
		description: string;
		items: WebsiteContent[];
		basePath: string;
		eyebrow: string;
	} = $props();

	function fmtDate(value: string | null) {
		if (!value) return 'Belum tayang';
		return new Date(value).toLocaleDateString('id-ID', {
			timeZone: 'Asia/Makassar',
			year: 'numeric',
			month: 'long',
			day: 'numeric',
		});
	}
</script>

<div class="space-y-8">
	<header class="rounded-[2rem] border border-border bg-card px-6 py-8 shadow-sm sm:px-8 sm:py-10">
		<div class="max-w-4xl space-y-4">
			<p class="text-xs font-semibold uppercase tracking-[0.22em] text-primary">{eyebrow}</p>
			<h1 class="text-3xl font-bold leading-tight text-foreground sm:text-5xl">{title}</h1>
			<p class="max-w-3xl text-base leading-8 text-muted-foreground sm:text-lg">{description}</p>
		</div>
		<div class="mt-5 flex flex-wrap items-center gap-3 text-sm text-muted-foreground">
			<span class="rounded-full border border-primary/20 bg-card px-3 py-1">{items.length} konten tayang</span>
			<span class="rounded-full border border-border bg-card px-3 py-1">Informasi resmi MTsN 2 Kolaka Utara</span>
		</div>
	</header>

	<div class="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
		{#each items as item (item.id)}
			<a href={`${basePath}/${item.slug}`} class="group overflow-hidden rounded-[1.75rem] border border-border bg-card shadow-sm transition hover:-translate-y-0.5 hover:border-primary/20 hover:shadow-md">
				{#if item.cover_image_url}
					<img src={item.cover_image_url} alt={item.title} class="h-48 w-full object-cover sm:h-52" />
				{:else}
					<div class="flex h-48 items-center justify-center bg-primary/10 text-sm font-semibold uppercase tracking-[0.18em] text-primary sm:h-52">
						{eyebrow}
					</div>
				{/if}
				<div class="space-y-3 p-5">
					<p class="text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">{fmtDate(item.published_at)}</p>
					<h2 class="text-lg font-semibold leading-8 text-foreground sm:text-xl">{item.title}</h2>
					<p class="text-sm leading-7 text-muted-foreground">{item.excerpt || 'Konten ini belum memiliki ringkasan singkat.'}</p>
					<div class="flex items-center justify-between gap-3 pt-1">
						<span class="text-sm font-semibold text-primary group-hover:text-primary">Lihat detail</span>
						<span class="text-primary transition-transform group-hover:translate-x-1">→</span>
					</div>
				</div>
			</a>
		{:else}
			<div class="rounded-[1.75rem] border border-dashed border-border bg-muted/50 px-5 py-12 text-center text-sm text-muted-foreground md:col-span-2 xl:col-span-3">
				Belum ada konten yang tayang saat ini.
			</div>
		{/each}
	</div>
</div>
