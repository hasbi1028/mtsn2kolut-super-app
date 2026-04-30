<script lang="ts">
	type WebsiteContent = {
		id: string;
		title: string;
		slug: string;
		excerpt: string;
		content_html: string;
		cover_image_url: string;
		published_at: string | null;
	};

	let { home }: {
		home?: {
			posts: WebsiteContent[];
			announcements: WebsiteContent[];
			profil: WebsiteContent | null;
			ppdbInfo: WebsiteContent | null;
		};
	} = $props();

	function fmtDate(value: string | null) {
		if (!value) return 'Belum dipublikasikan';
		return new Date(value).toLocaleDateString('id-ID', {
			timeZone: 'Asia/Makassar',
			year: 'numeric',
			month: 'long',
			day: 'numeric',
		});
	}
</script>

<svelte:head>
	<title>MTs Negeri 2 Kolaka Utara — Website Resmi</title>
	<meta name="description" content="Website resmi MTs Negeri 2 Kolaka Utara untuk informasi sekolah, berita kegiatan, pengumuman, dan PPDB." />
</svelte:head>

<div class="space-y-10">
	<section class="relative overflow-hidden rounded-[2rem] border border-emerald-100 bg-[linear-gradient(135deg,oklch(0.98_0.02_145)_0%,white_55%,oklch(0.95_0.05_145)_100%)] px-6 py-10 shadow-sm sm:px-10 sm:py-14">
		<div class="absolute inset-y-0 right-0 hidden w-1/3 bg-[radial-gradient(circle_at_top,rgba(19,106,67,0.12),transparent_60%)] lg:block"></div>
		<div class="relative grid gap-8 lg:grid-cols-[1.15fr,0.85fr]">
			<div class="space-y-5">
				<p class="inline-flex rounded-full bg-emerald-100 px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em] text-emerald-800">
					Website Resmi Madrasah
				</p>
				<div class="space-y-3">
					<h1 class="max-w-3xl text-4xl font-bold leading-tight text-slate-900 sm:text-5xl">
						Belajar, bertumbuh, dan berprestasi bersama MTs Negeri 2 Kolaka Utara.
					</h1>
					<p class="max-w-2xl text-base leading-7 text-slate-600 sm:text-lg">
						Pusat informasi sekolah untuk profil madrasah, berita kegiatan, pengumuman resmi, dan layanan PPDB yang terhubung langsung dengan sistem.
					</p>
				</div>
				<div class="flex flex-wrap gap-3">
					<a href="/ppdb" class="rounded-full bg-[oklch(0.38_0.13_145)] px-5 py-3 text-sm font-semibold text-white shadow-sm hover:brightness-105">
						Daftar PPDB Sekarang
					</a>
					<a href="/berita" class="rounded-full border border-emerald-200 px-5 py-3 text-sm font-semibold text-emerald-800 hover:bg-emerald-50">
						Lihat Berita Madrasah
					</a>
				</div>
			</div>

			<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-1">
				<div class="rounded-3xl border border-white/80 bg-white/80 p-5 shadow-sm backdrop-blur">
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-emerald-700">Profil Singkat</p>
					<p class="mt-3 text-sm leading-7 text-slate-600">
						{home?.profil?.excerpt || 'Lengkapi profil sekolah untuk menampilkan sejarah singkat, visi, misi, dan identitas madrasah di sini.'}
					</p>
					<a href="/profil" class="mt-4 inline-flex text-sm font-semibold text-emerald-800 hover:text-emerald-900">Buka profil →</a>
				</div>
				<div class="rounded-3xl border border-white/80 bg-white/80 p-5 shadow-sm backdrop-blur">
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-amber-700">Info PPDB</p>
					<p class="mt-3 text-sm leading-7 text-slate-600">
						{home?.ppdbInfo?.excerpt || 'Gunakan website ini untuk menampilkan jalur pendaftaran, jadwal, syarat, dan FAQ PPDB dengan lebih jelas.'}
					</p>
					<a href="/ppdb" class="mt-4 inline-flex text-sm font-semibold text-emerald-800 hover:text-emerald-900">Buka PPDB →</a>
				</div>
			</div>
		</div>
	</section>

	<section class="grid gap-4 md:grid-cols-3">
		<div class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
			<p class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">Identitas</p>
			<p class="mt-3 text-lg font-semibold text-slate-900">Madrasah Tsanawiyah Negeri</p>
			<p class="mt-2 text-sm leading-7 text-slate-600">Laman publik ini menjadi pusat informasi yang mudah diakses masyarakat, calon siswa, dan wali.</p>
		</div>
		<div class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
			<p class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">Informasi Resmi</p>
			<p class="mt-3 text-lg font-semibold text-slate-900">Berita dan Pengumuman</p>
			<p class="mt-2 text-sm leading-7 text-slate-600">Semua informasi sekolah dikelola dari panel admin dan dipublikasikan melalui alur draft ke publish.</p>
		</div>
		<div class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
			<p class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">Layanan</p>
			<p class="mt-3 text-lg font-semibold text-slate-900">PPDB Online</p>
			<p class="mt-2 text-sm leading-7 text-slate-600">Calon peserta didik dapat melihat informasi dan langsung mendaftar melalui website resmi sekolah.</p>
		</div>
	</section>

	<section class="grid gap-6 lg:grid-cols-[1.1fr,0.9fr]">
		<div class="rounded-[2rem] border border-slate-200 bg-white p-6 shadow-sm">
			<div class="flex items-center justify-between gap-3">
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">Berita Terbaru</p>
					<h2 class="mt-2 text-2xl font-semibold text-slate-900">Kegiatan dan kabar madrasah</h2>
				</div>
				<a href="/berita" class="text-sm font-semibold text-emerald-800 hover:text-emerald-900">Semua berita →</a>
			</div>

			<div class="mt-6 grid gap-4">
				{#if home?.posts?.length}
					{#each home.posts as post (post.id)}
						<a href={`/berita/${post.slug}`} class="rounded-2xl border border-slate-200 bg-slate-50/70 p-4 transition hover:border-emerald-200 hover:bg-emerald-50/40">
							<p class="text-xs font-medium uppercase tracking-[0.16em] text-slate-500">{fmtDate(post.published_at)}</p>
							<h3 class="mt-2 text-lg font-semibold text-slate-900">{post.title}</h3>
							<p class="mt-2 text-sm leading-7 text-slate-600">{post.excerpt || 'Belum ada ringkasan berita.'}</p>
						</a>
					{/each}
				{:else}
					<div class="rounded-2xl border border-dashed border-slate-300 bg-slate-50 px-4 py-8 text-center text-sm text-slate-500">
						Belum ada berita yang dipublikasikan.
					</div>
				{/if}
			</div>
		</div>

		<div class="rounded-[2rem] border border-amber-100 bg-[linear-gradient(180deg,#fffdf7_0%,white_100%)] p-6 shadow-sm">
			<div class="flex items-center justify-between gap-3">
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-amber-700">Pengumuman</p>
					<h2 class="mt-2 text-2xl font-semibold text-slate-900">Info penting sekolah</h2>
				</div>
				<a href="/pengumuman" class="text-sm font-semibold text-emerald-800 hover:text-emerald-900">Semua pengumuman →</a>
			</div>

			<div class="mt-6 space-y-3">
				{#if home?.announcements?.length}
					{#each home.announcements as item (item.id)}
						<a href={`/pengumuman/${item.slug}`} class="block rounded-2xl border border-amber-100 bg-white p-4 transition hover:border-amber-200 hover:bg-amber-50/50">
							<div class="flex items-start justify-between gap-3">
								<div>
									<h3 class="text-base font-semibold text-slate-900">{item.title}</h3>
									<p class="mt-2 text-sm leading-7 text-slate-600">{item.excerpt || 'Belum ada ringkasan pengumuman.'}</p>
								</div>
								<span class="rounded-full bg-amber-100 px-2.5 py-1 text-[11px] font-semibold uppercase tracking-[0.14em] text-amber-800">
									{fmtDate(item.published_at)}
								</span>
							</div>
						</a>
					{/each}
				{:else}
					<div class="rounded-2xl border border-dashed border-amber-200 bg-white px-4 py-8 text-center text-sm text-slate-500">
						Belum ada pengumuman yang dipublikasikan.
					</div>
				{/if}
			</div>
		</div>
	</section>
</div>
