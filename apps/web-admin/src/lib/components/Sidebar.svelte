<script lang="ts">
	import { page } from '$app/state';

	let { user }: { user?: { id: string; username: string; role: string; employee_id?: string } } = $props();
	let open = $state(false);

	type NavItem = { href: string; label: string; icon: string; roles?: string[] };
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
				{ href: '/students',  label: 'Siswa',         icon: 'users' },
			],
		},
		{
			group: 'CBT',
			items: [
				{ href: '/cbt/events',    label: 'Kegiatan Ujian', icon: 'calendar',  roles: ['admin'] },
				{ href: '/cbt/questions', label: 'Bank Soal',      icon: 'file-text' },
				{ href: '/cbt/packages',  label: 'Paket Ujian',    icon: 'package'   },
				{ href: '/cbt/sessions',  label: 'Sesi Ujian',     icon: 'play'      },
			],
		},
		{
			group: 'Operasional',
			items: [
				{ href: '/employees', label: 'Pegawai', icon: 'user-check', roles: ['admin'] },
			],
		},
		{
			group: 'PUSAKA',
			items: [
				{ href: '/pusaka',           label: 'Kontrol & Monitor',   icon: 'server',  roles: ['admin'] },
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

	const nav = $derived(
		allNav.map(g => ({
			...g,
			items: g.items.filter(i => !i.roles || (user?.role && i.roles.includes(user.role)))
		})).filter(g => g.items.length > 0)
	);

	function isActive(href: string) {
		if (href === '/') return page.url.pathname === '/';
		return page.url.pathname.startsWith(href);
	}

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
<header class="sticky top-0 z-10 flex h-14 items-center gap-3 border-b border-slate-200 bg-white px-4 lg:hidden">
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
	class="fixed inset-y-0 left-0 z-30 flex w-60 flex-col border-r border-slate-200 bg-white
	       transition-transform duration-200
	       {open ? 'translate-x-0' : '-translate-x-full'}
	       lg:translate-x-0"
>
	<!-- Brand -->
	<div class="flex h-14 shrink-0 items-center gap-2.5 border-b border-slate-200 px-4">
		<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-green-700 text-white text-xs font-bold shrink-0">
			MTs
		</div>
		<div class="min-w-0">
			<p class="truncate text-sm font-semibold text-slate-800">MTSN 2 Kolut</p>
			<p class="truncate text-xs text-slate-400">Kolaka Utara</p>
		</div>
	</div>

	<!-- Nav -->
	<nav class="flex-1 overflow-y-auto py-3 px-3 space-y-4">
		{#each nav as section}
			<div>
				<p class="mb-1 px-2 text-[10px] font-semibold uppercase tracking-wider text-slate-400">
					{section.group}
				</p>
				<ul class="space-y-0.5">
					{#each section.items as item}
						<li>
							<a
								href={item.href}
								onclick={() => (open = false)}
								class="flex items-center gap-2.5 rounded-md px-2 py-1.5 text-sm font-medium transition-colors
								       {isActive(item.href)
								         ? 'bg-green-50 text-green-800'
								         : 'text-slate-600 hover:bg-green-50/60 hover:text-slate-800'}"
							>
								{@render SidebarIcon({ name: item.icon, active: isActive(item.href) })}
								{item.label}
							</a>
						</li>
					{/each}
				</ul>
			</div>
		{/each}
	</nav>

	<!-- Footer -->
	<div class="shrink-0 border-t border-slate-200 p-3">
		{#if user}
			<button
				onclick={logout}
				class="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm text-slate-500
				       hover:bg-red-50 hover:text-red-600 transition-colors"
			>
				<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
						d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
				</svg>
				Keluar
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
		{/if}
	</svg>
{/snippet}
