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
		<div class="mx-auto grid max-w-6xl gap-8 px-4 py-8 sm:grid-cols-[1.2fr,0.8fr] sm:px-6">
			<div>
				<p class="text-base font-semibold text-slate-900">MTs Negeri 2 Kolaka Utara</p>
				<p class="mt-2 max-w-2xl text-sm leading-6 text-slate-600">
					Website resmi madrasah untuk informasi sekolah, berita kegiatan, pengumuman, dan layanan PPDB.
				</p>
			</div>
			<div class="grid gap-2 text-sm text-slate-600">
				<p><span class="font-medium text-slate-900">Navigasi cepat:</span> Berita, Pengumuman, Profil, PPDB</p>
				<p><span class="font-medium text-slate-900">Kontak:</span> Lengkapi dari halaman Kontak sekolah</p>
			</div>
		</div>
	</footer>
</div>
