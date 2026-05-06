<script lang="ts">
	import { afterNavigate } from '$app/navigation';
	import { resolve } from '$app/paths';
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
		{ href: '/kontak', label: 'Kontak' }
	] as const;

	let mobileOpen = $state(false);
	const mobileMenuId = 'public-site-mobile-menu';

	const footerGroups = [
		{
			title: 'Jelajahi',
			links: [
				{ href: '/profil', label: 'Profil Madrasah' },
				{ href: '/berita', label: 'Berita' },
				{ href: '/pengumuman', label: 'Pengumuman' }
			]
		},
		{
			title: 'Layanan',
			links: [
				{ href: '/ppdb', label: 'PPDB' },
				{ href: '/kontak', label: 'Kontak Resmi' },
				{ href: '/login', label: 'Login Admin' }
			]
		}
	] as const;

	function isActive(href: string) {
		if (href === '/') return page.url.pathname === '/';
		return page.url.pathname === href || page.url.pathname.startsWith(`${href}/`);
	}

	afterNavigate(() => {
		mobileOpen = false;
	});
</script>

<div class="min-h-screen bg-background text-foreground">
	<header class="sticky top-0 z-30 border-b border-primary/20 bg-card/90 backdrop-blur">
		<div class="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-3 sm:px-6">
				<a href={resolve('/')} class="flex items-center gap-3">
					<div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary text-sm font-bold text-primary-foreground shadow-sm">
						MTs
				</div>
				<div>
					<p class="text-sm font-semibold text-foreground sm:text-base">MTs Negeri 2 Kolaka Utara</p>
					<p class="text-xs text-primary/80">Website Resmi Madrasah</p>
				</div>
			</a>

			<nav class="hidden items-center gap-1 lg:flex">
				{#each navItems as item (item.href)}
						<a
							href={resolve(item.href)}
							aria-current={isActive(item.href) ? 'page' : undefined}
							class={`rounded-full px-4 py-2 text-sm font-medium transition-colors ${
								isActive(item.href)
								? 'bg-primary/10 text-primary'
								: 'text-muted-foreground hover:bg-muted hover:text-foreground'
						}`}
					>
						{item.label}
					</a>
				{/each}
			</nav>

				<div class="hidden items-center gap-2 lg:flex">
					{#if user}
						<a href={resolve('/')} class="rounded-full border border-primary/20 px-4 py-2 text-sm font-medium text-primary hover:bg-primary/10">
							Dashboard
						</a>
					{:else}
						<a href={resolve('/login')} class="rounded-full border border-border px-4 py-2 text-sm font-medium text-foreground hover:bg-muted/50">
							Login Admin
						</a>
					{/if}
					<a href={resolve('/ppdb')} class="rounded-full bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground shadow-sm hover:brightness-105">
						Daftar PPDB
					</a>
				</div>

			<button
					type="button"
					class="inline-flex rounded-xl border border-border p-2 text-muted-foreground transition hover:bg-muted/50 lg:hidden"
					aria-controls={mobileMenuId}
					aria-expanded={mobileOpen}
					aria-label={mobileOpen ? 'Tutup navigasi' : 'Buka navigasi'}
					onclick={() => (mobileOpen = !mobileOpen)}
				>
					<svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						{#if mobileOpen}
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						{:else}
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
						{/if}
					</svg>
				</button>
			</div>

			{#if mobileOpen}
				<div class="border-t border-primary/20 bg-card lg:hidden">
					<nav id={mobileMenuId} aria-label="Navigasi website mobile" class="mx-auto flex max-w-6xl flex-col gap-2 px-4 py-3 sm:px-6">
						<div class="mb-1 rounded-2xl border border-primary/20 bg-primary/10 px-3 py-2">
							<p class="text-[10px] font-semibold uppercase tracking-[0.2em] text-primary">Menu Website</p>
							<p class="mt-1 text-xs text-primary">Akses halaman publik dan layanan PPDB MTsN 2 Kolaka Utara.</p>
						</div>
						{#each navItems as item (item.href)}
							<a
								href={resolve(item.href)}
								aria-current={isActive(item.href) ? 'page' : undefined}
								onclick={() => (mobileOpen = false)}
								class={`rounded-xl px-3 py-2 text-sm font-medium ${
								isActive(item.href)
									? 'bg-primary/10 text-primary'
									: 'text-foreground hover:bg-muted'
							}`}
						>
							{item.label}
						</a>
						{/each}
						<div class="mt-2 flex gap-2">
							{#if user}
								<a href={resolve('/')} class="flex-1 rounded-xl border border-primary/20 px-3 py-2 text-center text-sm font-medium text-primary">
									Dashboard
								</a>
							{:else}
								<a href={resolve('/login')} class="flex-1 rounded-xl border border-border px-3 py-2 text-center text-sm font-medium text-foreground">
									Login
								</a>
							{/if}
							<a href={resolve('/ppdb')} class="flex-1 rounded-xl bg-primary px-3 py-2 text-center text-sm font-semibold text-primary-foreground">
								PPDB
							</a>
						</div>
					</nav>
				</div>
			{/if}
	</header>

	<main class="mx-auto min-h-[calc(100vh-210px)] max-w-6xl px-4 py-8 sm:px-6 sm:py-10">
		{@render children()}
	</main>

	<footer class="border-t border-primary/20 bg-card">
		<div class="mx-auto max-w-6xl space-y-6 px-4 py-8 sm:px-6 sm:py-10">
			<div class="rounded-[2rem] border border-border bg-card px-6 py-6 shadow-sm sm:px-8">
				<div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
					<div class="max-w-3xl">
						<p class="text-xs font-semibold uppercase tracking-[0.22em] text-primary">Layanan Publik Madrasah</p>
						<h2 class="mt-2 text-2xl font-semibold text-foreground">Akses informasi sekolah dan PPDB dari satu tempat</h2>
						<p class="mt-2 text-sm leading-7 text-muted-foreground">
							Gunakan website ini untuk membaca informasi resmi sekolah, mengikuti pengumuman terbaru, dan memulai proses pendaftaran calon siswa.
						</p>
					</div>
					<div class="flex flex-wrap gap-2">
							<a href={resolve('/ppdb')} class="rounded-full bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground shadow-sm hover:brightness-105">
								Buka PPDB
							</a>
							<a href={resolve('/kontak')} class="rounded-full border border-primary/20 px-4 py-2 text-sm font-semibold text-primary hover:bg-primary/10">
								Hubungi Sekolah
							</a>
					</div>
				</div>
			</div>

			<div class="grid gap-8 sm:grid-cols-[1.2fr,0.8fr,0.8fr]">
				<div>
					<p class="text-base font-semibold text-foreground">MTs Negeri 2 Kolaka Utara</p>
					<p class="mt-2 max-w-2xl text-sm leading-7 text-muted-foreground">
						Website resmi madrasah untuk informasi sekolah, berita kegiatan, pengumuman, dan layanan PPDB yang mudah diakses masyarakat.
					</p>
				</div>

				{#each footerGroups as group (group.title)}
					<div class="space-y-3">
						<p class="text-sm font-semibold text-foreground">{group.title}</p>
						<div class="grid gap-2 text-sm text-muted-foreground">
								{#each group.links as link (link.href)}
									<a href={resolve(link.href)} class="hover:text-primary">{link.label}</a>
								{/each}
						</div>
					</div>
				{/each}
			</div>

			<div class="border-t border-border pt-4 text-xs text-muted-foreground">
				Informasi pada website ini dikelola oleh MTs Negeri 2 Kolaka Utara dan diperbarui melalui panel editorial sekolah.
			</div>
		</div>
	</footer>
</div>
