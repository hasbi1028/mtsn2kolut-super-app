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
	<header class="rounded-[2rem] border border-emerald-100 bg-[linear-gradient(180deg,rgba(236,253,245,0.94),rgba(255,255,255,1))] px-6 py-8 shadow-sm sm:px-8 sm:py-10">
		<div class="max-w-4xl space-y-4">
			<p class="text-xs font-semibold uppercase tracking-[0.22em] text-emerald-700">{eyebrow}</p>
			<h1 class="text-3xl font-bold leading-tight text-slate-900 sm:text-5xl">{title}</h1>
			<p class="max-w-3xl text-base leading-8 text-slate-600 sm:text-lg">{description}</p>
		</div>
		<div class="mt-5 flex flex-wrap items-center gap-3 text-sm text-slate-500">
			<span class="rounded-full border border-emerald-200 bg-white px-3 py-1">{items.length} konten tayang</span>
			<span class="rounded-full border border-slate-200 bg-white px-3 py-1">Informasi resmi MTsN 2 Kolaka Utara</span>
		</div>
	</header>

	<div class="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
		{#each items as item (item.id)}
			<a href={`${basePath}/${item.slug}`} class="group overflow-hidden rounded-[1.75rem] border border-slate-200 bg-white shadow-sm transition hover:-translate-y-0.5 hover:border-emerald-200 hover:shadow-md">
				{#if item.cover_image_url}
					<img src={item.cover_image_url} alt={item.title} class="h-48 w-full object-cover sm:h-52" />
				{:else}
					<div class="flex h-48 items-center justify-center bg-[linear-gradient(135deg,#f0fdf4_0%,#ecfeff_100%)] text-sm font-semibold uppercase tracking-[0.18em] text-emerald-700 sm:h-52">
						{eyebrow}
					</div>
				{/if}
				<div class="space-y-3 p-5">
					<p class="text-xs font-medium uppercase tracking-[0.18em] text-slate-500">{fmtDate(item.published_at)}</p>
					<h2 class="text-lg font-semibold leading-8 text-slate-900 sm:text-xl">{item.title}</h2>
					<p class="text-sm leading-7 text-slate-600">{item.excerpt || 'Konten ini belum memiliki ringkasan singkat.'}</p>
					<div class="flex items-center justify-between gap-3 pt-1">
						<span class="text-sm font-semibold text-emerald-800 group-hover:text-emerald-900">Lihat detail</span>
						<span class="text-emerald-700 transition-transform group-hover:translate-x-1">→</span>
					</div>
				</div>
			</a>
		{:else}
			<div class="rounded-[1.75rem] border border-dashed border-slate-300 bg-slate-50 px-5 py-12 text-center text-sm text-slate-500 md:col-span-2 xl:col-span-3">
				Belum ada konten yang tayang saat ini.
			</div>
		{/each}
	</div>
</div>
