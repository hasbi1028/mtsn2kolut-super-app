<script lang="ts">
	let { home }: {
		home?: {
			posts: any[];
			featuredPosts: any[];
			announcements: any[];
			profil: any | null;
			ppdbInfo: any | null;
		};
	} = $props();
</script>

<svelte:head>
	<title>MTs Negeri 2 Kolaka Utara — Dashboard</title>
	<meta name="description" content="Sistem informasi MTs Negeri 2 Kolaka Utara untuk pengelolaan pegawai, kehadiran, dan pengaturan." />
</svelte:head>

	<div class="space-y-10">
		<section class="relative overflow-hidden rounded-[1.75rem] border border-emerald-100 bg-[linear-gradient(135deg,oklch(0.98_0.02_145)_0%,white_55%,oklch(0.95_0.05_145)_100%)] px-5 py-7 shadow-sm sm:rounded-[2rem] sm:px-10 sm:py-14">
			<div class="absolute inset-y-0 right-0 hidden w-1/3 bg-[radial-gradient(circle_at_top,rgba(19,106,67,0.12),transparent_60%)] lg:block"></div>
			<div class="relative grid gap-6 lg:grid-cols-[1.15fr,0.85fr] lg:gap-8">
				<div class="space-y-4 sm:space-y-5">
					<p class="inline-flex rounded-full bg-emerald-100 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.18em] text-emerald-800 sm:text-xs">
						Sistem Informasi Madrasah
					</p>
					<div class="space-y-3">
						<h1 class="max-w-3xl text-3xl font-bold leading-[1.08] text-slate-900 sm:text-5xl sm:leading-tight">
							<span class="sm:hidden">Selamat datang di MTsN 2 Kolaka Utara.</span>
							<span class="hidden sm:inline">Selamat datang di Sistem Informasi MTsN 2 Kolaka Utara.</span>
						</h1>
						<p class="max-w-2xl text-sm leading-7 text-slate-600 sm:text-lg">
							Kelola data pegawai, kehadiran, dan pengaturan sistem dalam satu platform terpadu.
						</p>
					</div>
				</div>
			</div>
		</section>

		<section class="grid gap-4 md:grid-cols-3">
			<div class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
				<p class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">Pegawai</p>
				<p class="mt-3 text-lg font-semibold text-slate-900">Master Data Pegawai</p>
				<p class="mt-2 text-sm leading-7 text-slate-600">Kelola data pegawai, profil, dan informasi kepegawaian.</p>
			</div>
			<div class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
				<p class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">Kehadiran</p>
				<p class="mt-3 text-lg font-semibold text-slate-900">Sistem PUSAKA</p>
				<p class="mt-2 text-sm leading-7 text-slate-600">Monitor kehadiran pegawai secara real-time.</p>
			</div>
			<div class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
				<p class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">Pengaturan</p>
				<p class="mt-3 text-lg font-semibold text-slate-900">Pengelolaan Sistem</p>
				<p class="mt-2 text-sm leading-7 text-slate-600">Atur pengguna, hak akses, dan konfigurasi sistem.</p>
			</div>
		</section>
	</div>
