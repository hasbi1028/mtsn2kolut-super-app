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
	const startedForms = new Set<string>();

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
				{ href: '/login', label: 'Masuk' }
			]
		}
	] as const;

	function isActive(href: string) {
		if (href === '/') return page.url.pathname === '/';
		return page.url.pathname === href || page.url.pathname.startsWith(`${href}/`);
	}

	type PublicAnalyticsOptions = {
		pathname?: string;
		result?: string;
		metadata?: Record<string, unknown>;
	};

	function runWhenIdle(task: () => void) {
		if (typeof window === 'undefined') return;
		const requestIdleCallback = window.requestIdleCallback ?? ((callback: IdleRequestCallback) => window.setTimeout(() => callback({ didTimeout: false, timeRemaining: () => 0 } as IdleDeadline), 1200));
		requestIdleCallback(task, { timeout: 2500 });
	}

	function trackPublicAnalyticsEventDeferred(eventName: string, options: PublicAnalyticsOptions = {}) {
		runWhenIdle(() => {
			void import('$lib/analytics/public-analytics').then(({ trackPublicAnalyticsEvent }) =>
				trackPublicAnalyticsEvent(eventName, options)
			).catch(() => false);
		});
	}

	function trackPublicPageViewDeferred(pathname: string, metadata: Record<string, unknown>) {
		runWhenIdle(() => {
			void import('$lib/analytics/public-analytics').then(({ trackPublicPageView }) =>
				trackPublicPageView(pathname, metadata)
			).catch(() => false);
		});
	}

	afterNavigate(() => {
		mobileOpen = false;
		trackPublicPageViewDeferred(page.url.pathname, {
			page_key: publicPageKey(page.url.pathname),
			page_kind: publicPageKind(page.url.pathname),
			device_class: publicDeviceClass()
		});
	});

	function publicPageKey(pathname: string) {
		const clean = pathname.split(/[?#]/, 1)[0] ?? '/';
		const parts = clean.split('/').filter(Boolean);
		if (parts.length === 0) return 'home';
		if (parts[0] === 'berita' && parts.length > 1) return 'berita_detail';
		if (parts[0] === 'pengumuman' && parts.length > 1) return 'pengumuman_detail';
		return safePublicToken(parts[0]);
	}

	function publicPageKind(pathname: string) {
		const key = publicPageKey(pathname);
		if (key.endsWith('_detail')) return 'detail';
		if (key === 'berita' || key === 'pengumuman') return 'list';
		if (key === 'ppdb') return 'form';
		return 'page';
	}

	function publicDeviceClass() {
		if (typeof window === 'undefined') return 'unknown';
		if (window.matchMedia('(max-width: 640px)').matches) return 'mobile';
		if (window.matchMedia('(max-width: 1024px)').matches) return 'tablet';
		return 'desktop';
	}

	function handlePublicClick(event: MouseEvent) {
		const target = event.target;
		if (!(target instanceof Element)) return;
		const trigger = target.closest('a,button');
		if (!(trigger instanceof HTMLElement)) return;
		const href = trigger instanceof HTMLAnchorElement ? trigger.getAttribute('href') || '' : '';
		const sourceComponent = trigger.closest('header') ? 'header' : trigger.closest('footer') ? 'footer' : 'content';
		const metadata = {
			page_key: publicPageKey(page.url.pathname),
			cta_key: safePublicToken(trigger.dataset.analyticsKey || linkKindFromHref(href) || trigger.getAttribute('type') || 'action'),
			cta_group: safePublicToken(trigger.dataset.analyticsGroup || sourceComponent),
			link_kind: linkKindFromHref(href),
			source_component: sourceComponent
		};
		if (isDownloadHref(href) || (trigger instanceof HTMLAnchorElement && trigger.hasAttribute('download'))) {
			trackPublicAnalyticsEventDeferred('public.download', {
				pathname: page.url.pathname,
				metadata: {
					...metadata,
					file_kind: fileKindFromHref(href),
					download_kind: trigger.hasAttribute('download') ? 'explicit' : 'file_link'
				}
			});
			return;
		}
		trackPublicAnalyticsEventDeferred('public.cta_click', { pathname: page.url.pathname, metadata });
	}

	function handlePublicFocusIn(event: FocusEvent) {
		const target = event.target;
		if (!(target instanceof Element)) return;
		const form = target.closest('form');
		if (!(form instanceof HTMLFormElement)) return;
		const formKey = safePublicToken(form.dataset.analyticsForm || form.getAttribute('name') || publicPageKey(page.url.pathname));
		if (startedForms.has(formKey)) return;
		startedForms.add(formKey);
		trackPublicAnalyticsEventDeferred('public.form_start', {
			pathname: page.url.pathname,
			metadata: { form_key: formKey, page_key: publicPageKey(page.url.pathname), source_component: 'form' }
		});
	}

	function handlePublicSubmit(event: SubmitEvent) {
		const target = event.target;
		if (!(target instanceof HTMLFormElement)) return;
		const formKey = safePublicToken(target.dataset.analyticsForm || target.getAttribute('name') || publicPageKey(page.url.pathname));
		trackPublicAnalyticsEventDeferred('public.form_submit', {
			pathname: page.url.pathname,
			metadata: { form_key: formKey, page_key: publicPageKey(page.url.pathname), result: 'started', source_component: 'form' }
		});
	}

	function handlePublicChange(event: Event) {
		const target = event.target;
		if (!(target instanceof HTMLInputElement)) return;
		if (target.type !== 'search' && !target.dataset.publicSearch) return;
		trackPublicAnalyticsEventDeferred('public.search', {
			pathname: page.url.pathname,
			metadata: {
				page_key: publicPageKey(page.url.pathname),
				search_length_bucket: searchLengthBucket(target.value),
				search_bucket: 'public_site'
			}
		});
	}

	function safePublicToken(value: string) {
		return (value || 'unknown').trim().toLowerCase().replace(/[-/\s]+/g, '_').replace(/[^a-z0-9_]/g, '').replace(/^_+|_+$/g, '').slice(0, 64) || 'unknown';
	}

	function linkKindFromHref(href: string) {
		if (!href) return 'button';
		if (href.startsWith('mailto:')) return 'email';
		if (href.startsWith('tel:')) return 'phone';
		if (/^https?:\/\//i.test(href)) return 'external';
		const first = href.split(/[?#]/, 1)[0]?.split('/').filter(Boolean)[0] ?? 'home';
		return safePublicToken(first);
	}

	function isDownloadHref(href: string) {
		return /\.(pdf|docx?|xlsx?|pptx?|zip|jpg|jpeg|png|webp)$/i.test(href.split(/[?#]/, 1)[0] ?? '');
	}

	function fileKindFromHref(href: string) {
		const clean = href.split(/[?#]/, 1)[0] ?? '';
		const ext = clean.split('.').pop()?.toLowerCase() ?? '';
		if (ext === 'pdf') return 'pdf';
		if (['doc', 'docx'].includes(ext)) return 'document';
		if (['xls', 'xlsx'].includes(ext)) return 'spreadsheet';
		if (['jpg', 'jpeg', 'png', 'webp'].includes(ext)) return 'image';
		if (ext === 'zip') return 'archive';
		return 'file';
	}

	function searchLengthBucket(value: string) {
		const len = value.trim().length;
		if (len === 0) return 'empty';
		if (len <= 10) return '1_10';
		if (len <= 20) return '11_20';
		return 'gt_20';
	}
</script>

<svelte:window onclick={handlePublicClick} onfocusin={handlePublicFocusIn} onsubmit={handlePublicSubmit} onchange={handlePublicChange} />

<div class="min-h-screen bg-[linear-gradient(180deg,#f7faf7_0%,#f9fafb_22%,#ffffff_100%)]">
	<header class="sticky top-0 z-30 border-b border-emerald-100/80 bg-white/90 backdrop-blur">
		<div class="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-3 sm:px-6">
				<a href={resolve('/')} class="flex items-center gap-3">
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
							href={resolve(item.href)}
							aria-current={isActive(item.href) ? 'page' : undefined}
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
						<a href={resolve('/')} class="rounded-full border border-emerald-200 px-4 py-2 text-sm font-medium text-emerald-800 hover:bg-emerald-50">
							Dashboard
						</a>
					{:else}
						<a href={resolve('/login')} class="rounded-full border border-slate-200 px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50">
							Masuk
						</a>
					{/if}
					<a href={resolve('/ppdb')} class="rounded-full bg-[oklch(0.38_0.13_145)] px-4 py-2 text-sm font-semibold text-white shadow-sm hover:brightness-105">
						Daftar PPDB
					</a>
				</div>

			<button
					type="button"
					class="inline-flex rounded-xl border border-slate-200 p-2 text-slate-600 transition hover:bg-slate-50 lg:hidden"
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
				<div class="border-t border-emerald-100 bg-white lg:hidden">
					<nav id={mobileMenuId} aria-label="Navigasi website mobile" class="mx-auto flex max-w-6xl flex-col gap-2 px-4 py-3 sm:px-6">
						<div class="mb-1 rounded-2xl border border-emerald-100 bg-emerald-50 px-3 py-2">
							<p class="text-[10px] font-semibold uppercase tracking-[0.2em] text-emerald-700">Menu Website</p>
							<p class="mt-1 text-xs text-emerald-900">Akses halaman publik dan layanan PPDB MTsN 2 Kolaka Utara.</p>
						</div>
						{#each navItems as item (item.href)}
							<a
								href={resolve(item.href)}
								aria-current={isActive(item.href) ? 'page' : undefined}
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
								<a href={resolve('/')} class="flex-1 rounded-xl border border-emerald-200 px-3 py-2 text-center text-sm font-medium text-emerald-800">
									Dashboard
								</a>
							{:else}
								<a href={resolve('/login')} class="flex-1 rounded-xl border border-slate-200 px-3 py-2 text-center text-sm font-medium text-slate-700">
									Login
								</a>
							{/if}
							<a href={resolve('/ppdb')} class="flex-1 rounded-xl bg-[oklch(0.38_0.13_145)] px-3 py-2 text-center text-sm font-semibold text-white">
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
							<a href={resolve('/ppdb')} class="rounded-full bg-[oklch(0.38_0.13_145)] px-4 py-2 text-sm font-semibold text-white shadow-sm hover:brightness-105">
								Buka PPDB
							</a>
							<a href={resolve('/kontak')} class="rounded-full border border-emerald-200 px-4 py-2 text-sm font-semibold text-emerald-800 hover:bg-emerald-50">
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
									<a href={resolve(link.href)} class="hover:text-emerald-800">{link.label}</a>
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
