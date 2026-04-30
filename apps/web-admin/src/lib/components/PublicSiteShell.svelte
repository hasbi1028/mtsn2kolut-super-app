<script lang="ts">
	import { page } from '$app/state';

	let { children, user } = $props<{
		children: import('svelte').Snippet;
		user?: { id?: string };
	}>();

	const navItems = [
		{ href: '/', label: 'Beranda' },
		{ href: '/profil', label: 'Profil' },
		{ href: '/berita', label: 'Berita' },
		{ href: '/pengumuman', label: 'Pengumuman' },
		{ href: '/ppdb', label: 'PPDB' },
		{ href: '/kontak', label: 'Kontak' },
	];

	let mobileOpen = $state(false);

	const footerGroups = [
		{
			title: 'Jelajahi',
			links: [
				{ href: '/profil', label: 'Profil Madrasah' },
				{ href: '/berita', label: 'Berita' },
				{ href: '/pengumuman', label: 'Pengumuman' },
			],
		},
		{
			title: 'Layanan',
			links: [
				{ href: '/ppdb', label: 'PPDB' },
				{ href: '/kontak', label: 'Kontak Resmi' },
				{ href: '/login', label: 'Login Admin' },
			],
		},
	];

	function isActive(href: string) {
		if (href === '/') return page.url.pathname === '/';
		return page.url.pathname === href || page.url.pathname.startsWith(`${href}/`);
	}
</script>

<div class="min-h-screen bg-[linear-gradient(180deg,#f7faf7_0%,#f9fafb_22%,#ffffff_100%)]">
	<header class="sticky top-0 z-30 border-b border-emerald-100/80 bg-white/90 backdrop-blur">
		<div class="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-3 sm:px-6">
			<a href="/" class="flex items-center gap-3">
				<div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-[oklch(0.38_0.13_145)] text-sm font-bold text-white shadow-sm">
					MTs
				</div>
				<div>
					<p class="text-sm font-semibold text-slate-900 sm:text-base">MTs Negeri 2 Kolaka Utara</p>
					<p class="text-xs text-emerald-700/80">Website Resmi Madrasah</p>
				</div>
			</a>

			<nav class="hidden items-center gap-1 lg:flex">
				{#each navItems as item (item.href)}
					<a
						href={item.href}
						class={`rounded-full px-4 py-2 text-sm font-medium transition-colors ${
							isActive(item.href)
								? 'bg-emerald-50 text-emerald-800'
								: 'text-slate-600 hover:bg-slate-100 hover:text-slate-900'
						}`}
					>
						{item.label}
					</a>
				{/each}
			</nav>

			<div class="hidden items-center gap-2 lg:flex">
				{#if user}
					<a href="/" class="rounded-full border border-emerald-200 px-4 py-2 text-sm font-medium text-emerald-800 hover:bg-emerald-50">
						Dashboard
					</a>
				{:else}
					<a href="/login" class="rounded-full border border-slate-200 px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50">
						Login Admin
					</a>
				{/if}
				<a href="/ppdb" class="rounded-full bg-[oklch(0.38_0.13_145)] px-4 py-2 text-sm font-semibold text-white shadow-sm hover:brightness-105">
					Daftar PPDB
				</a>
			</div>

			<button
				type="button"
				class="inline-flex rounded-xl border border-slate-200 p-2 text-slate-600 lg:hidden"
				aria-label="Buka navigasi"
				onclick={() => (mobileOpen = !mobileOpen)}
			>
				<svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
				</svg>
			</button>
		</div>

		{#if mobileOpen}
			<div class="border-t border-emerald-100 bg-white lg:hidden">
				<div class="mx-auto flex max-w-6xl flex-col gap-1 px-4 py-3 sm:px-6">
					{#each navItems as item (item.href)}
						<a
							href={item.href}
							onclick={() => (mobileOpen = false)}
							class={`rounded-xl px-3 py-2 text-sm font-medium ${
								isActive(item.href)
									? 'bg-emerald-50 text-emerald-800'
									: 'text-slate-700 hover:bg-slate-100'
							}`}
						>
							{item.label}
						</a>
					{/each}
					<div class="mt-2 flex gap-2">
						{#if user}
							<a href="/" class="flex-1 rounded-xl border border-emerald-200 px-3 py-2 text-center text-sm font-medium text-emerald-800">
								Dashboard
							</a>
						{:else}
							<a href="/login" class="flex-1 rounded-xl border border-slate-200 px-3 py-2 text-center text-sm font-medium text-slate-700">
								Login
							</a>
						{/if}
						<a href="/ppdb" class="flex-1 rounded-xl bg-[oklch(0.38_0.13_145)] px-3 py-2 text-center text-sm font-semibold text-white">
							PPDB
						</a>
					</div>
				</div>
			</div>
		{/if}
	</header>

	<main class="mx-auto min-h-[calc(100vh-210px)] max-w-6xl px-4 py-8 sm:px-6 sm:py-10">
		{@render children()}
	</main>

	<footer class="border-t border-emerald-100 bg-white">
		<div class="mx-auto max-w-6xl space-y-6 px-4 py-8 sm:px-6 sm:py-10">
			<div class="rounded-[2rem] border border-emerald-100 bg-[linear-gradient(135deg,rgba(236,253,245,0.92),rgba(255,255,255,1))] px-6 py-6 shadow-sm sm:px-8">
				<div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
					<div class="max-w-3xl">
						<p class="text-xs font-semibold uppercase tracking-[0.22em] text-emerald-700">Layanan Publik Madrasah</p>
						<h2 class="mt-2 text-2xl font-semibold text-slate-900">Akses informasi sekolah dan PPDB dari satu tempat</h2>
						<p class="mt-2 text-sm leading-7 text-slate-600">
							Gunakan website ini untuk membaca informasi resmi sekolah, mengikuti pengumuman terbaru, dan memulai proses pendaftaran calon siswa.
						</p>
					</div>
					<div class="flex flex-wrap gap-2">
						<a href="/ppdb" class="rounded-full bg-[oklch(0.38_0.13_145)] px-4 py-2 text-sm font-semibold text-white shadow-sm hover:brightness-105">
							Buka PPDB
						</a>
						<a href="/kontak" class="rounded-full border border-emerald-200 px-4 py-2 text-sm font-semibold text-emerald-800 hover:bg-emerald-50">
							Hubungi Sekolah
						</a>
					</div>
				</div>
			</div>

			<div class="grid gap-8 sm:grid-cols-[1.2fr,0.8fr,0.8fr]">
				<div>
					<p class="text-base font-semibold text-slate-900">MTs Negeri 2 Kolaka Utara</p>
					<p class="mt-2 max-w-2xl text-sm leading-7 text-slate-600">
						Website resmi madrasah untuk informasi sekolah, berita kegiatan, pengumuman, dan layanan PPDB yang mudah diakses masyarakat.
					</p>
				</div>

				{#each footerGroups as group (group.title)}
					<div class="space-y-3">
						<p class="text-sm font-semibold text-slate-900">{group.title}</p>
						<div class="grid gap-2 text-sm text-slate-600">
							{#each group.links as link (link.href)}
								<a href={link.href} class="hover:text-emerald-800">{link.label}</a>
							{/each}
						</div>
					</div>
				{/each}
			</div>

			<div class="border-t border-slate-200 pt-4 text-xs text-slate-500">
				Informasi pada website ini dikelola oleh MTs Negeri 2 Kolaka Utara dan diperbarui melalui panel editorial sekolah.
			</div>
		</div>
	</footer>
</div>
