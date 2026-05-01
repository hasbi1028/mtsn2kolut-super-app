<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';

	let {
		user,
		desktopExpanded = $bindable(true)
	}: {
		user?: { id: string; username: string; role: string; roles?: string[]; employee_id?: string };
		desktopExpanded?: boolean;
	} = $props();
	let open = $state(false);

	type NavItem = { href: string; label: string; icon: string; roles?: string[]; pinnable?: boolean };
	type NavGroup = { group: string; items: NavItem[] };

	const allNav: NavGroup[] = [
		{
			group: 'Utama',
			items: [
				{ href: '/',         label: 'Dashboard',    icon: 'grid' },
			],
		},
		{
			group: 'Akademik',
			items: [
				{ href: '/academic',  label: 'Data Akademik', icon: 'book-open', roles: ['admin'] },
				{ href: '/jadwal',        label: 'Jadwal',       icon: 'calendar', roles: ['guru', 'siswa', 'ortu'] },
				{ href: '/grades',        label: 'Nilai',        icon: 'clipboard', roles: ['admin', 'guru'] },
				{ href: '/grades/rapor',  label: 'Cetak Rapor',  icon: 'printer',   roles: ['admin', 'guru'] },
				{ href: '/journal',       label: 'Jurnal Kelas', icon: 'journal',   roles: ['admin', 'guru'] },
				{ href: '/students',  label: 'Siswa',         icon: 'users' },
				{ href: '/parents',   label: 'Orang Tua',     icon: 'user-group', roles: ['admin', 'staf'] },
			],
		},
		{
			group: 'CBT',
			items: [
				{ href: '/cbt/events',    label: 'Kegiatan Ujian', icon: 'calendar',  roles: ['admin'] },
				{ href: '/cbt/byod',      label: 'Panduan BYOD',   icon: 'activity',  roles: ['admin', 'guru'] },
				{ href: '/cbt/questions', label: 'Bank Soal',      icon: 'file-text' },
				{ href: '/cbt/soal',      label: 'Komposer Soal',  icon: 'pen-tool'  },
				{ href: '/cbt/packages',  label: 'Paket Ujian',    icon: 'package'   },
				{ href: '/cbt/sessions',  label: 'Sesi Ujian',     icon: 'play'      },
			],
		},
		{
			group: 'Operasional',
			items: [
				{ href: '/employees', label: 'Master Pegawai', icon: 'user-check', roles: ['admin'] },
			],
		},
		{
			group: 'Perpustakaan',
			items: [
				{ href: '/library',       label: 'Dashboard',    icon: 'book-open', roles: ['admin', 'staf'] },
				{ href: '/library/books', label: 'Katalog Buku', icon: 'book',      roles: ['admin', 'staf'] },
				{ href: '/library/loans', label: 'Peminjaman',   icon: 'repeat',    roles: ['admin', 'staf'] },
			],
		},
		{
			group: 'Inventaris',
			items: [
				{ href: '/inventory',       label: 'Dashboard',        icon: 'package', roles: ['admin', 'staf'] },
				{ href: '/inventory/items', label: 'Daftar Barang',    icon: 'layers',  roles: ['admin', 'staf'] },
			],
		},
		{
			group: 'Website',
			items: [
				{ href: '/website',               label: 'Website Publik', icon: 'globe', roles: ['admin'] },
				{ href: '/website/posts',         label: 'Berita',         icon: 'file-text', roles: ['admin'] },
				{ href: '/website/announcements', label: 'Pengumuman',     icon: 'clipboard', roles: ['admin'] },
				{ href: '/website/pages',         label: 'Halaman Publik', icon: 'book-open', roles: ['admin'] },
			],
		},
		{
			group: 'PUSAKA',
			items: [
				{ href: '/pusaka',           label: 'Kontrol & Monitor',   icon: 'server',  roles: ['admin'] },
				{ href: '/pusaka/employees', label: 'Pegawai PUSAKA',      icon: 'user-check', roles: ['admin'] },
				{ href: '/pusaka/kehadiran', label: 'Data Kehadiran',       icon: 'clock',   roles: ['admin'] },
				{ href: '/pusaka/summary',   label: 'Ringkasan Kehadiran',  icon: 'layers',  roles: ['admin'] },
				{ href: '/pusaka/antrian',   label: 'Antrian Job',          icon: 'activity',roles: ['admin'] },
			],
		},
		{
			group: 'Sistem',
			items: [
				{ href: '/settings/users', label: 'Manajemen User', icon: 'users', roles: ['admin'] },
				{ href: '/settings/audit-logs', label: 'Audit Trail', icon: 'file-text', roles: ['admin'] },
				{ href: '/settings', label: 'Pengaturan', icon: 'settings', roles: ['admin'] },
			],
		},
	];

	const userRoles = $derived(user?.roles || (user?.role ? [user.role] : []));

	const nav = $derived(
		allNav
			.map((g) => ({
				...g,
				items: g.items.filter((i) => {
					if (!i.roles) return true;
					return i.roles.some((r) => userRoles.includes(r));
				}),
			}))
			.filter((g) => g.items.length > 0)
	);

	const PINNED_STORAGE_KEY = 'sidebar:pinned-items';
	let openGroups = $state<string[]>([]);
	let pinnedItems = $state<string[]>([]);
	let pinnedLoaded = false;

	const defaultPinnedByRole: Record<string, string[]> = {
		admin: ['/', '/cbt/sessions', '/grades', '/settings'],
		guru: ['/', '/cbt/questions', '/grades', '/jadwal'],
		staf: ['/', '/inventory', '/library']
	};

	const activeGroup = $derived(
		nav.find((section) => section.items.some((item) => isActive(item.href)))?.group ?? 'Utama'
	);

	const quickAccess = $derived.by(() => {
		const hrefs = new Set<string>(['/', ...pinnedItems]);
		const flattened = nav.flatMap((section) =>
			section.items.map((item) => ({
				...item,
				group: section.group,
			}))
		);
		return flattened.filter((item, index) => hrefs.has(item.href) && flattened.findIndex((candidate) => candidate.href === item.href) === index);
	});

	function isActive(href: string) {
		if (href === '/') return page.url.pathname === '/';
		return page.url.pathname.startsWith(href);
	}

	function isGroupOpen(group: string) {
		return group === 'Utama' || group === activeGroup || openGroups.includes(group);
	}

	function toggleGroup(group: string) {
		if (group === 'Utama' || group === activeGroup) {
			return;
		}
		if (isGroupOpen(group)) {
			openGroups = openGroups.filter((value) => value !== group);
			return;
		}
		openGroups = [...openGroups, group];
	}

	function isPinned(href: string) {
		return pinnedItems.includes(href);
	}

	function togglePin(href: string) {
		if (isPinned(href)) {
			pinnedItems = pinnedItems.filter((value) => value !== href);
			return;
		}
		pinnedItems = [...pinnedItems, href];
	}

	function pinButtonLabel(item: NavItem) {
		return isPinned(item.href) ? `Lepas ${item.label} dari akses cepat` : `Pin ${item.label} ke akses cepat`;
	}

	function railTooltip(item: NavItem, group: string) {
		return `${group} · ${item.label}`;
	}

	function loadPinnedItems() {
		if (typeof window === 'undefined') return;
		const raw = window.localStorage.getItem(PINNED_STORAGE_KEY);
		if (raw) {
			try {
				const parsed = JSON.parse(raw);
				if (Array.isArray(parsed)) {
					pinnedItems = parsed.filter((value): value is string => typeof value === 'string');
				}
			} catch {
				pinnedItems = [];
			}
		}

		if (pinnedItems.length === 0) {
			for (const role of userRoles) {
				const defaults = defaultPinnedByRole[role];
				if (defaults?.length) {
					pinnedItems = defaults;
					break;
				}
			}
		}

		pinnedLoaded = true;
	}

	onMount(() => {
		loadPinnedItems();
	});

	$effect(() => {
		if (!pinnedLoaded || typeof window === 'undefined') return;
		window.localStorage.setItem(PINNED_STORAGE_KEY, JSON.stringify(pinnedItems));
	});

	async function logout() {
		await fetch('/api/auth/logout', { method: 'POST' });
		location.href = '/login';
	}
</script>

<!-- Mobile overlay -->
{#if open}
	<div
		class="fixed inset-0 z-20 bg-black/40 lg:hidden"
		role="button"
		tabindex="-1"
		aria-label="Tutup menu"
		onclick={() => (open = false)}
		onkeydown={(e) => e.key === 'Escape' && (open = false)}
	></div>
{/if}

<!-- Mobile topbar -->
<header class="fixed inset-x-0 top-0 z-20 flex h-14 w-full items-center gap-3 border-b border-slate-200 bg-white px-4 lg:hidden">
	<button
		class="rounded-md p-1.5 text-slate-500 hover:bg-slate-100"
		onclick={() => (open = !open)}
		aria-label="Toggle menu"
	>
		<svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
		</svg>
	</button>
	<span class="text-sm font-semibold text-slate-700">MTSN 2 Kolaka Utara</span>
</header>

<!-- Sidebar -->
<aside
	class={`fixed inset-y-0 left-0 z-30 flex flex-col border-r border-slate-200 bg-white
	       transition-transform duration-200
	       ${open ? 'translate-x-0' : '-translate-x-full'}
	       ${desktopExpanded ? 'lg:w-60' : 'lg:w-[5.5rem]'}
	       lg:translate-x-0`}
>
	<!-- Brand -->
	<div class={`flex h-14 shrink-0 items-center border-b border-slate-200 ${desktopExpanded ? 'gap-2.5 px-4' : 'justify-center px-3'}`}>
		<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-green-700 text-white text-xs font-bold shrink-0">
			MTs
		</div>
		<div class={`min-w-0 ${desktopExpanded ? 'block' : 'block lg:hidden'}`}>
			<p class="truncate text-sm font-semibold text-slate-800">MTSN 2 Kolut</p>
			<p class="truncate text-xs text-slate-400">Kolaka Utara</p>
		</div>
		<button
			class={`ml-auto hidden rounded-md p-1.5 text-slate-500 hover:bg-slate-100 lg:inline-flex ${desktopExpanded ? '' : 'ml-0'}`}
			onclick={() => (desktopExpanded = !desktopExpanded)}
			aria-label={desktopExpanded ? 'Collapse sidebar' : 'Expand sidebar'}
		>
			<svg class={`h-4 w-4 transition-transform ${desktopExpanded ? '' : 'rotate-180'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
			</svg>
		</button>
	</div>

	<!-- Nav -->
	<nav class={`flex-1 overflow-y-auto py-3 ${desktopExpanded ? 'px-3' : 'px-2'} space-y-4`}>
		{#if quickAccess.length > 0}
			<div>
				<p class={`mb-1 px-2 text-[10px] font-semibold uppercase tracking-wider text-slate-400 ${desktopExpanded ? 'block' : 'block lg:hidden'}`}>
					Akses Cepat
				</p>
				<ul class="space-y-0.5">
					{#each quickAccess as item (item.href)}
						<li>
							<div class={`group relative flex items-center ${desktopExpanded ? 'gap-1' : 'gap-0 lg:justify-center'}`}>
								<a
									href={resolve(item.href)}
									onclick={() => (open = false)}
									title={!desktopExpanded ? railTooltip(item, item.group) : undefined}
									class={`flex min-w-0 flex-1 items-center rounded-md py-1.5 text-sm font-medium transition-colors
										${desktopExpanded ? 'gap-2.5 px-2' : 'gap-2.5 px-2 lg:justify-center lg:px-0'}
										${isActive(item.href)
											? 'bg-green-50 text-green-800'
											: 'text-slate-600 hover:bg-green-50/60 hover:text-slate-800'}`}
								>
									{@render SidebarIcon({ name: item.icon, active: isActive(item.href) })}
									<span class={`truncate ${desktopExpanded ? 'inline' : 'inline lg:hidden'}`}>{item.label}</span>
								</a>
								{#if item.pinnable !== false}
									<button
										type="button"
										class={`shrink-0 rounded-md p-1 text-amber-500 hover:bg-amber-50 hover:text-amber-600 ${desktopExpanded ? 'inline-flex' : 'inline-flex lg:hidden'}`}
										onclick={() => togglePin(item.href)}
										aria-label={pinButtonLabel(item)}
									>
										<svg class="h-3.5 w-3.5" fill={isPinned(item.href) ? 'currentColor' : 'none'} stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 17.75l-6.172 3.245 1.179-6.872L2 9.38l6.914-1.005L12 2.11l3.086 6.265L22 9.38l-5.007 4.743 1.18 6.872z" />
										</svg>
									</button>
								{/if}
								{#if !desktopExpanded}
									<div class="pointer-events-none absolute left-full top-1/2 z-40 ml-3 hidden -translate-y-1/2 rounded-md border border-slate-200 bg-white px-2 py-1 text-xs font-medium text-slate-700 shadow-sm lg:group-hover:block">
										{railTooltip(item, item.group)}
									</div>
								{/if}
							</div>
						</li>
					{/each}
				</ul>
			</div>
		{/if}

		{#each nav as section (section.group)}
			<div>
				<button
					type="button"
					class={`mb-1 flex w-full items-center rounded-md px-2 py-1 text-left text-[10px] font-semibold uppercase tracking-wider text-slate-400 transition-colors hover:bg-slate-50 ${desktopExpanded ? 'flex' : 'flex lg:hidden'}`}
					onclick={() => toggleGroup(section.group)}
					aria-expanded={isGroupOpen(section.group)}
				>
					<span class="truncate">{section.group}</span>
					<svg class={`ml-auto h-3.5 w-3.5 transition-transform ${isGroupOpen(section.group) ? 'rotate-90' : ''}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
					</svg>
				</button>
				{#if !desktopExpanded || isGroupOpen(section.group)}
					<ul class="space-y-0.5">
						{#each section.items as item (item.href)}
							<li>
								<div class={`group relative flex items-center ${desktopExpanded ? 'gap-1' : 'gap-0 lg:justify-center'}`}>
									<a
										href={resolve(item.href)}
										onclick={() => (open = false)}
										title={!desktopExpanded ? railTooltip(item, section.group) : undefined}
										class={`flex min-w-0 flex-1 items-center rounded-md py-1.5 text-sm font-medium transition-colors
											${desktopExpanded ? 'gap-2.5 px-2' : 'gap-2.5 px-2 lg:justify-center lg:px-0'}
											${isActive(item.href)
												? 'bg-green-50 text-green-800'
												: 'text-slate-600 hover:bg-green-50/60 hover:text-slate-800'}`}
									>
										{@render SidebarIcon({ name: item.icon, active: isActive(item.href) })}
										<span class={`truncate ${desktopExpanded ? 'inline' : 'inline lg:hidden'}`}>{item.label}</span>
									</a>
									{#if item.pinnable !== false}
										<button
											type="button"
											class={`shrink-0 rounded-md p-1 text-slate-400 transition-colors hover:bg-amber-50 hover:text-amber-600 ${desktopExpanded ? 'inline-flex' : 'inline-flex lg:hidden'}`}
											onclick={() => togglePin(item.href)}
											aria-label={pinButtonLabel(item)}
										>
											<svg class="h-3.5 w-3.5" fill={isPinned(item.href) ? 'currentColor' : 'none'} stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 17.75l-6.172 3.245 1.179-6.872L2 9.38l6.914-1.005L12 2.11l3.086 6.265L22 9.38l-5.007 4.743 1.18 6.872z" />
											</svg>
										</button>
									{/if}
									{#if !desktopExpanded}
										<div class="pointer-events-none absolute left-full top-1/2 z-40 hidden -translate-y-1/2 rounded-md border border-slate-200 bg-white px-2 py-1 text-xs font-medium text-slate-700 shadow-sm lg:group-hover:block lg:ml-3">
											{railTooltip(item, section.group)}
										</div>
									{/if}
								</div>
							</li>
						{/each}
					</ul>
				{/if}
			</div>
		{/each}
	</nav>

	<!-- Footer -->
	<div class="shrink-0 border-t border-slate-200 p-3">
		{#if user}
			<button
				onclick={logout}
				title={!desktopExpanded ? 'Keluar' : undefined}
				class={`flex w-full items-center rounded-md py-1.5 text-sm text-slate-500 transition-colors
				       hover:bg-red-50 hover:text-red-600 ${desktopExpanded ? 'gap-2 px-2' : 'justify-center px-0'}`}
			>
				<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
						d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
				</svg>
				<span class={desktopExpanded ? 'inline' : 'inline lg:hidden'}>Keluar</span>
			</button>
		{/if}
	</div>
</aside>

<!-- Icon helper snippet -->
{#snippet SidebarIcon({ name, active }: { name: string; active: boolean })}
	<svg class="h-4 w-4 shrink-0 {active ? 'text-green-700' : 'text-slate-400'}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
		{#if name === 'grid'}
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
		{:else if name === 'calendar'}
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
		{:else if name === 'book-open'}
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
		{:else if name === 'users'}
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
		{:else if name === 'user-group'}
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
		{:else if name === 'file-text'}
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
		{:else if name === 'package'}
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
		{:else if name === 'user-check'}
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
		{:else if name === 'clock'}
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
		{:else if name === 'layers'}
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" />
		{:else if name === 'play'}
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
		{:else if name === 'settings'}
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
		{:else if name === 'server'}
			<rect x="2" y="2" width="20" height="8" rx="2" ry="2" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" />
			<rect x="2" y="14" width="20" height="8" rx="2" ry="2" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" />
			<line x1="6" y1="6" x2="6.01" y2="6" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" />
			<line x1="6" y1="18" x2="6.01" y2="18" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" />
		{:else if name === 'activity'}
			<polyline points="22 12 18 12 15 21 9 3 6 12 2 12" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" />
		{:else if name === 'clipboard'}
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2" />
			<rect x="9" y="3" width="6" height="4" rx="1" ry="1" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" />
		{:else if name === 'pen-tool'}
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 20h9M16.5 3.5a2.121 2.121 0 013 3L7 19l-4 1 1-4L16.5 3.5z" />
		{:else if name === 'book'}
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 19.5A2.5 2.5 0 016.5 17H20" />
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6.5 2H20v20H6.5A2.5 2.5 0 014 19.5v-15A2.5 2.5 0 016.5 2z" />
		{:else if name === 'globe'}
			<circle cx="12" cy="12" r="10" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" />
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2 12h20M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z" />
		{:else if name === 'repeat'}
			<polyline points="17 1 21 5 17 9" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" />
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 11V9a4 4 0 014-4h14" />
			<polyline points="7 23 3 19 7 15" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" />
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 13v2a4 4 0 01-4 4H3" />
		{:else if name === 'journal'}
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.746 0 3.332.477 4.5 1.253v13C19.832 18.477 18.246 18 16.5 18c-1.746 0-3.332.477-4.5 1.253z" />
		{/if}
	</svg>
{/snippet}
