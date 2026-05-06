<script lang="ts">
	import { resolve } from '$app/paths';

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
			featuredPosts: WebsiteContent[];
			announcements: WebsiteContent[];
			profil: WebsiteContent | null;
			ppdbInfo: WebsiteContent | null;
		};
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

<svelte:head>
	<title>MTs Negeri 2 Kolaka Utara — Website Resmi</title>
	<meta name="description" content="Website resmi MTs Negeri 2 Kolaka Utara untuk informasi sekolah, berita kegiatan, pengumuman, dan PPDB." />
</svelte:head>

	<div class="space-y-10">
		<section class="relative overflow-hidden rounded-[1.75rem] border border-border bg-card px-5 py-7 shadow-sm sm:rounded-[2rem] sm:px-10 sm:py-14">
			<div class="relative grid gap-6 lg:grid-cols-[1.15fr,0.85fr] lg:gap-8">
				<div class="space-y-4 sm:space-y-5">
					<p class="inline-flex rounded-full bg-primary/15 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.18em] text-primary sm:text-xs">
						Website Resmi Madrasah
					</p>
					<div class="space-y-3">
						<h1 class="max-w-3xl text-3xl font-bold leading-[1.08] text-foreground sm:text-5xl sm:leading-tight">
							<span class="sm:hidden">Belajar, bertumbuh, dan berprestasi.</span>
							<span class="hidden sm:inline">Belajar, bertumbuh, dan berprestasi bersama MTs Negeri 2 Kolaka Utara.</span>
						</h1>
						<p class="max-w-2xl text-sm leading-7 text-muted-foreground sm:text-lg">
							Pusat informasi sekolah untuk profil madrasah, berita kegiatan, pengumuman resmi, dan layanan PPDB yang terhubung langsung dengan sistem.
						</p>
					</div>
					<div class="flex flex-wrap gap-3">
						<a href={resolve('/ppdb')} class="inline-flex w-full justify-center rounded-full bg-primary px-5 py-3 text-sm font-semibold text-primary-foreground shadow-sm hover:brightness-105 sm:w-auto">
							Lihat Info & Daftar PPDB
						</a>
						<a href={resolve('/berita')} class="inline-flex w-full justify-center rounded-full border border-primary/20 px-5 py-3 text-sm font-semibold text-primary hover:bg-primary/10 sm:w-auto">
							Baca Berita Madrasah
						</a>
					</div>
					<div class="grid gap-3 sm:grid-cols-3">
						<div class="rounded-2xl border border-border bg-muted/50 px-4 py-3 shadow-sm sm:py-4">
							<p class="text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">Informasi Resmi</p>
							<p class="mt-2 hidden text-sm leading-7 text-foreground sm:block">Berita, pengumuman, dan profil sekolah tersedia dalam satu alur publik yang rapi.</p>
						</div>
						<div class="rounded-2xl border border-border bg-muted/50 px-4 py-3 shadow-sm sm:py-4">
							<p class="text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">Layanan Publik</p>
							<p class="mt-2 hidden text-sm leading-7 text-foreground sm:block">PPDB awal bisa diakses langsung tanpa perlu masuk ke dashboard internal.</p>
						</div>
						<div class="rounded-2xl border border-border bg-muted/50 px-4 py-3 shadow-sm sm:py-4">
							<p class="text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">Akses Mudah</p>
							<p class="mt-2 hidden text-sm leading-7 text-foreground sm:block">Tampilan disusun agar nyaman dibaca di ponsel maupun layar desktop.</p>
						</div>
					</div>
				</div>

				<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-1">
				<div class="rounded-3xl border border-border bg-card p-5 shadow-sm">
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-primary">Profil Singkat</p>
					<p class="mt-3 text-sm leading-7 text-muted-foreground">
						{home?.profil?.excerpt || 'Lengkapi profil sekolah untuk menampilkan sejarah singkat, visi, misi, dan identitas madrasah di sini.'}
					</p>
					<a href={resolve('/profil')} class="mt-4 inline-flex text-sm font-semibold text-primary hover:text-primary">Buka profil →</a>
				</div>
				<div class="rounded-3xl border border-border bg-card p-5 shadow-sm">
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-warning">Info PPDB</p>
					<p class="mt-3 text-sm leading-7 text-muted-foreground">
						{home?.ppdbInfo?.excerpt || 'Gunakan website ini untuk menampilkan jalur pendaftaran, jadwal, syarat, dan FAQ PPDB dengan lebih jelas.'}
					</p>
					<a href={resolve('/ppdb')} class="mt-4 inline-flex text-sm font-semibold text-primary hover:text-primary">Buka PPDB →</a>
				</div>
			</div>
		</div>
	</section>

	<section class="grid gap-4 md:grid-cols-3">
		<div class="rounded-3xl border border-border bg-card p-5 shadow-sm">
			<p class="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Identitas</p>
			<p class="mt-3 text-lg font-semibold text-foreground">Madrasah Tsanawiyah Negeri</p>
			<p class="mt-2 text-sm leading-7 text-muted-foreground">Laman publik ini menjadi pusat informasi yang mudah diakses masyarakat, calon siswa, dan wali.</p>
		</div>
		<div class="rounded-3xl border border-border bg-card p-5 shadow-sm">
			<p class="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Informasi Resmi</p>
			<p class="mt-3 text-lg font-semibold text-foreground">Berita dan Pengumuman</p>
			<p class="mt-2 text-sm leading-7 text-muted-foreground">Semua informasi sekolah dikelola dari panel admin dan diterbitkan melalui alur editorial yang rapi.</p>
		</div>
		<div class="rounded-3xl border border-border bg-card p-5 shadow-sm">
			<p class="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Layanan</p>
			<p class="mt-3 text-lg font-semibold text-foreground">PPDB Online</p>
			<p class="mt-2 text-sm leading-7 text-muted-foreground">Calon peserta didik dapat melihat informasi dan langsung mendaftar melalui website resmi sekolah.</p>
		</div>
	</section>

	{#if home?.featuredPosts?.length}
		<section class="rounded-[2rem] border border-border bg-card p-6 shadow-sm">
			<div class="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-primary">Sorotan Utama</p>
					<h2 class="mt-2 text-2xl font-semibold text-foreground">Berita pilihan dari madrasah</h2>
					<p class="mt-2 max-w-2xl text-sm leading-7 text-muted-foreground">Konten unggulan ini dipilih untuk menonjolkan capaian, agenda, atau informasi yang paling layak dibaca pengunjung terlebih dahulu.</p>
				</div>
				<a href={resolve('/berita')} class="text-sm font-semibold text-primary hover:text-primary">Jelajahi semua berita →</a>
			</div>

			<div class="mt-6 grid gap-4 lg:grid-cols-2">
				{#each home.featuredPosts as post (post.id)}
					<a href={resolve(`/berita/${post.slug}`)} class="group overflow-hidden rounded-[1.75rem] border border-primary/20 bg-card shadow-sm transition hover:-translate-y-0.5 hover:border-primary/20">
						{#if post.cover_image_url}
							<div class="aspect-[16/9] overflow-hidden bg-primary/10">
								<img src={post.cover_image_url} alt={post.title} class="h-full w-full object-cover transition duration-500 group-hover:scale-[1.03]" />
							</div>
						{/if}
						<div class="space-y-3 p-5">
							<div class="flex items-center justify-between gap-3">
								<span class="rounded-full bg-primary/15 px-2.5 py-1 text-[11px] font-semibold uppercase tracking-[0.14em] text-primary">
									Unggulan
								</span>
								<span class="text-xs font-medium uppercase tracking-[0.16em] text-muted-foreground">{fmtDate(post.published_at)}</span>
							</div>
							<h3 class="text-xl font-semibold text-foreground">{post.title}</h3>
							<p class="text-sm leading-7 text-muted-foreground">{post.excerpt || 'Konten unggulan ini belum memiliki ringkasan singkat.'}</p>
							<div class="flex items-center justify-between gap-3 pt-1">
								<span class="text-sm font-semibold text-primary">Baca sorotan ini</span>
								<span class="text-primary transition-transform group-hover:translate-x-1">→</span>
							</div>
						</div>
					</a>
				{/each}
			</div>
		</section>
	{/if}

	<section class="grid gap-6 lg:grid-cols-[1.1fr,0.9fr]">
		<div class="rounded-[2rem] border border-border bg-card p-6 shadow-sm">
			<div class="flex items-center justify-between gap-3">
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Berita Terbaru</p>
					<h2 class="mt-2 text-2xl font-semibold text-foreground">Kegiatan dan kabar madrasah</h2>
				</div>
				<a href={resolve('/berita')} class="text-sm font-semibold text-primary hover:text-primary">Semua berita →</a>
			</div>

			<div class="mt-6 grid gap-4">
				{#if home?.posts?.length}
					{#each home.posts as post (post.id)}
						<a href={resolve(`/berita/${post.slug}`)} class="group rounded-2xl border border-border bg-muted/50 p-4 transition hover:border-primary/20 hover:bg-primary/10">
							<p class="text-xs font-medium uppercase tracking-[0.16em] text-muted-foreground">{fmtDate(post.published_at)}</p>
							<h3 class="mt-2 text-lg font-semibold text-foreground">{post.title}</h3>
							<p class="mt-2 text-sm leading-7 text-muted-foreground">{post.excerpt || 'Belum ada ringkasan berita.'}</p>
							<div class="mt-3 flex items-center justify-between gap-3">
								<span class="text-sm font-semibold text-primary">Lihat berita</span>
								<span class="text-primary transition-transform group-hover:translate-x-1">→</span>
							</div>
						</a>
					{/each}
				{:else}
					<div class="rounded-2xl border border-dashed border-border bg-muted/50 px-4 py-8 text-center text-sm text-muted-foreground">
						Belum ada berita yang tayang.
					</div>
				{/if}
			</div>
		</div>

		<div class="rounded-[2rem] border border-warning/30 bg-card p-6 shadow-sm">
			<div class="flex items-center justify-between gap-3">
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-warning">Pengumuman</p>
					<h2 class="mt-2 text-2xl font-semibold text-foreground">Info penting sekolah</h2>
				</div>
				<a href={resolve('/pengumuman')} class="text-sm font-semibold text-primary hover:text-primary">Semua pengumuman →</a>
			</div>

			<div class="mt-6 space-y-3">
				{#if home?.announcements?.length}
					{#each home.announcements as item (item.id)}
						<a href={resolve(`/pengumuman/${item.slug}`)} class="block rounded-2xl border border-warning/30 bg-card p-4 transition hover:border-warning/30 hover:bg-warning/10">
							<div class="flex items-start justify-between gap-3">
								<div>
									<h3 class="text-base font-semibold text-foreground">{item.title}</h3>
									<p class="mt-2 text-sm leading-7 text-muted-foreground">{item.excerpt || 'Belum ada ringkasan pengumuman.'}</p>
									<p class="mt-3 text-sm font-semibold text-primary">Buka pengumuman</p>
								</div>
								<span class="rounded-full bg-warning/15 px-2.5 py-1 text-[11px] font-semibold uppercase tracking-[0.14em] text-warning">
									{fmtDate(item.published_at)}
								</span>
							</div>
						</a>
					{/each}
				{:else}
					<div class="rounded-2xl border border-dashed border-warning/30 bg-card px-4 py-8 text-center text-sm text-muted-foreground">
						Belum ada pengumuman yang tayang.
					</div>
				{/if}
			</div>
		</div>
	</section>
</div>
