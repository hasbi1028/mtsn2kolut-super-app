<script lang="ts">
	import { afterNavigate } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { defaultBranding, versionedAsset, type BrandingSettings } from '$lib/branding';

	let { children, user, branding = defaultBranding } = $props<{
		children: import('svelte').Snippet;
		user?: { id?: string };
		branding?: BrandingSettings;
	}>();

	const navItems = [
		{ href: '/', label: 'Dashboard' },
		{ href: '/login', label: 'Masuk' }
	] as const;

	let mobileOpen = $state(false);
	const mobileMenuId = 'public-site-mobile-menu';
	const startedForms = new Set<string>();

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
		return safePublicToken(parts[0]);
	}

	function publicPageKind(pathname: string) {
		const key = publicPageKey(pathname);
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

<div class="min-h-screen" style={`--brand-primary: ${branding.primary_color}; --brand-primary-soft: color-mix(in srgb, ${branding.primary_color} 12%, white)`}>
	<header class="sticky top-0 z-30 border-b border-emerald-100/80 bg-white/90 backdrop-blur">
		<div class="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-3 sm:px-6">
			<a href={resolve('/')} class="flex items-center gap-3">
				<span class="flex h-11 w-11 items-center justify-center overflow-hidden rounded-2xl bg-white p-1.5 shadow-sm ring-1 ring-emerald-100">
					<img src={versionedAsset(branding.mark_url, branding.version)} alt={`Logo ${branding.short_name}`} class="h-full w-full object-contain" />
				</span>
				<div>
					<p class="text-sm font-extrabold text-slate-900 tracking-tight sm:text-base">{branding.app_name}</p>
					<p class="text-[10px] font-extrabold uppercase tracking-widest text-emerald-600">{branding.tagline}</p>
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
					<nav id={mobileMenuId} aria-label="Navigasi website mobile" class="mx-auto max-w-6xl px-4 py-4 sm:px-6">
						<!-- Menu Header -->
						<div class="mb-3 flex items-center justify-between">
							<div>
								<p class="text-[10px] font-extrabold uppercase tracking-widest text-emerald-600">Navigasi</p>
								<p class="mt-0.5 text-xs font-semibold text-slate-500">Akses dashboard dan login.</p>
							</div>
							<button
								type="button"
								class="flex h-8 w-8 items-center justify-center rounded-xl border border-slate-200 text-slate-400 transition hover:bg-slate-50"
								onclick={() => (mobileOpen = false)}
								aria-label="Tutup menu"
							>
								<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
								</svg>
							</button>
						</div>

						<!-- Nav Items -->
						<div class="space-y-1.5">
							{#each navItems as item (item.href)}
								<a
									href={resolve(item.href)}
									aria-current={isActive(item.href) ? 'page' : undefined}
									onclick={() => (mobileOpen = false)}
									class={`flex items-center gap-3 rounded-2xl px-4 py-3 text-sm font-bold transition-all ${
										isActive(item.href)
											? 'bg-emerald-50 text-emerald-700 ring-1 ring-emerald-100'
											: 'text-slate-600 hover:bg-slate-50'
									}`}
								>
									{#if item.href === '/'}
										<div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl {isActive(item.href) ? 'bg-emerald-100 text-emerald-600' : 'bg-slate-100 text-slate-400'}">
											<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-4 0a1 1 0 01-1-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 01-1 1" />
											</svg>
										</div>
									{:else}
										<div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl {isActive(item.href) ? 'bg-emerald-100 text-emerald-600' : 'bg-slate-100 text-slate-400'}">
											<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" />
											</svg>
										</div>
									{/if}
									<span>{item.label}</span>
									<svg class="ml-auto h-4 w-4 {isActive(item.href) ? 'text-emerald-400' : 'text-slate-300'}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
									</svg>
								</a>
							{/each}
						</div>

						<!-- Mobile CTA -->
						<a href={resolve('/login')} class="mt-4 flex w-full items-center justify-center gap-2 rounded-2xl bg-emerald-600 px-4 py-3 text-sm font-extrabold text-white shadow-lg shadow-emerald-500/25 transition hover:bg-emerald-700">
							Masuk ke Sistem
							<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6" />
							</svg>
						</a>
					</nav>
				</div>
			{/if}
	</header>

	<main class="mx-auto min-h-[calc(100vh-210px)] max-w-6xl px-4 py-8 sm:px-6 sm:py-10">
		{@render children()}
	</main>

	<footer class="border-t border-emerald-100 bg-white">
		<div class="mx-auto max-w-6xl space-y-6 px-4 py-8 sm:px-6 sm:py-10">
			<div class="grid gap-8 sm:grid-cols-[1.2fr,0.8fr]">
				<div>
					<p class="text-base font-extrabold tracking-tight text-slate-900">{branding.app_name}</p>
					<p class="mt-2 max-w-2xl text-sm leading-7 text-slate-500 font-medium">
						Sistem informasi madrasah untuk pengelolaan pegawai, kehadiran, dan pengaturan sistem.
					</p>
				</div>
				<div class="space-y-3">
					<p class="text-xs font-extrabold uppercase tracking-widest text-slate-400">Menu</p>
					<div class="grid gap-2 text-sm text-slate-600 font-medium">
						<a href={resolve('/')} class="hover:text-emerald-700 transition-colors">Dashboard</a>
						<a href={resolve('/login')} class="hover:text-emerald-700 transition-colors">Masuk</a>
					</div>
				</div>
			</div>

			<div class="border-t border-slate-100 pt-4 text-xs text-slate-400 font-medium">
				Informasi pada website ini dikelola oleh MTs Negeri 2 Kolaka Utara.
			</div>
		</div>
	</footer>
</div>
