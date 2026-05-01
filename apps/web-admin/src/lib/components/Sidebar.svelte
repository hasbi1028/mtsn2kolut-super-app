<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import SidebarCommandPalette from '$lib/components/sidebar/SidebarCommandPalette.svelte';
	import SidebarIcon from '$lib/components/sidebar/SidebarIcon.svelte';
	import SidebarNavSection from '$lib/components/sidebar/SidebarNavSection.svelte';
	import SidebarQuickAccess from '$lib/components/sidebar/SidebarQuickAccess.svelte';
	import { fetchSidebarAttention } from '$lib/components/sidebar/sidebar-attention';
	import {
		defaultPinnedByRole,
		sidebarNavGroups,
		type SidebarNavItem
	} from '$lib/components/sidebar/sidebar-config';

	let {
		user,
		desktopExpanded = $bindable(true)
	}: {
		user?: { id: string; username: string; role: string; roles?: string[]; employee_id?: string };
		desktopExpanded?: boolean;
	} = $props();
	let open = $state(false);
	let commandOpen = $state(false);
	let inventoryAttention = $state(0);
	let libraryAttention = $state(0);
	let pusakaAttention = $state(0);
	let attentionRefreshInFlight = $state<Promise<void> | null>(null);
	let lastAttentionLoadedAt = $state(0);

	const userRoles = $derived(user?.roles || (user?.role ? [user.role] : []));
	const resolveNavHref = resolve as unknown as (href: string) => string;

	const nav = $derived(
		sidebarNavGroups
			.map((g) => ({
				...g,
				items: g.items.filter((i) => {
					if (!i.roles) return true;
					return i.roles.some((r) => userRoles.includes(r));
				}),
			}))
			.filter((g) => g.items.length > 0)
	);

	const PINNED_STORAGE_KEY_PREFIX = 'sidebar:pinned-items';
	const RECENT_STORAGE_KEY_PREFIX = 'sidebar:recent-items';
	const RECENT_LIMIT = 6;
	const ATTENTION_REFRESH_INTERVAL_MS = 60_000;
	let openGroups = $state<string[]>([]);
	let pinnedItems = $state<string[]>([]);
	let recentItems = $state<string[]>([]);
	let pinnedLoaded = false;

	const activeGroup = $derived(
		nav.find((section) => section.items.some((item) => isActive(item.href)))?.group ?? 'Utama'
	);

	const visibleNavItems = $derived(
		nav.flatMap((section) =>
			section.items.map((item) => ({
				...item,
				group: section.group,
			}))
		)
	);

	const visibleNavHrefSet = $derived(new Set(visibleNavItems.map((item) => item.href)));
	const pinnableNavHrefSet = $derived(
		new Set(
			visibleNavItems
				.filter((item) => item.href !== '/' && item.pinnable !== false)
				.map((item) => item.href)
		)
	);

	const quickAccess = $derived.by(() => {
		const orderedHrefs = Array.from(new Set(['/', ...pinnedItems.filter((href) => href !== '/')]));
		return orderedHrefs
			.map((href) => visibleNavItems.find((item) => item.href === href))
			.filter((item): item is (typeof visibleNavItems)[number] => !!item);
	});

	const commandItems = $derived.by(() => {
		const flattened = visibleNavItems.map((item) => ({
			...item,
			pinned: item.href === '/' || pinnedItems.includes(item.href),
		}));
		return flattened.filter(
			(item, index) => flattened.findIndex((candidate) => candidate.href === item.href) === index
		);
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
		if (href === '/') {
			return;
		}
		if (isPinned(href)) {
			pinnedItems = pinnedItems.filter((value) => value !== href);
			return;
		}
		pinnedItems = [...pinnedItems, href];
	}

	function pinButtonLabel(item: SidebarNavItem) {
		return isPinned(item.href) ? `Lepas ${item.label} dari akses cepat` : `Pin ${item.label} ke akses cepat`;
	}

	function canMovePinned(href: string, direction: -1 | 1) {
		const index = pinnedItems.indexOf(href);
		if (index === -1) return false;
		const nextIndex = index + direction;
		return nextIndex >= 0 && nextIndex < pinnedItems.length;
	}

	function movePinned(href: string, direction: -1 | 1) {
		const index = pinnedItems.indexOf(href);
		if (index === -1) return;
		const nextIndex = index + direction;
		if (nextIndex < 0 || nextIndex >= pinnedItems.length) return;
		const next = [...pinnedItems];
		const [value] = next.splice(index, 1);
		next.splice(nextIndex, 0, value);
		pinnedItems = next;
	}

	function railTooltip(item: SidebarNavItem, group: string) {
		return `${group} · ${item.label}`;
	}

	function storageScope() {
		return user?.id?.trim() || 'anon';
	}

	function pinnedStorageKey() {
		return `${PINNED_STORAGE_KEY_PREFIX}:${storageScope()}`;
	}

	function recentStorageKey() {
		return `${RECENT_STORAGE_KEY_PREFIX}:${storageScope()}`;
	}

	function dedupeHrefs(values: string[]) {
		return Array.from(new Set(values));
	}

	function normalizePinnedHrefs(values: unknown) {
		if (!Array.isArray(values)) return [];
		return dedupeHrefs(
			values.filter(
				(value): value is string =>
					typeof value === 'string' && value !== '/' && pinnableNavHrefSet.has(value)
			)
		);
	}

	function normalizeRecentHrefs(values: unknown) {
		if (!Array.isArray(values)) return [];
		return dedupeHrefs(
			values.filter(
				(value): value is string => typeof value === 'string' && visibleNavHrefSet.has(value)
			)
		).slice(0, RECENT_LIMIT);
	}

	function defaultPinnedItemsForUser() {
		for (const role of userRoles) {
			const defaults = normalizePinnedHrefs(defaultPinnedByRole[role] ?? []);
			if (defaults.length > 0) {
				return defaults;
			}
		}
		return [];
	}

	function navBadge(href: string) {
		if (href === '/inventory/items') return inventoryAttention;
		if (href === '/library/loans') return libraryAttention;
		if (href === '/pusaka/antrian') return pusakaAttention;
		return 0;
	}

	function groupBadge(group: string) {
		if (group === 'Inventaris') return inventoryAttention;
		if (group === 'Perpustakaan') return libraryAttention;
		if (group === 'PUSAKA') return pusakaAttention;
		return 0;
	}

	async function loadSidebarAttention() {
		const attention = await fetchSidebarAttention(fetch, userRoles);
		inventoryAttention = attention.inventory;
		libraryAttention = attention.library;
		pusakaAttention = attention.pusaka;
	}

	async function refreshSidebarAttention(force = false) {
		if (attentionRefreshInFlight) {
			return attentionRefreshInFlight;
		}
		if (!force && Date.now() - lastAttentionLoadedAt < ATTENTION_REFRESH_INTERVAL_MS) {
			return;
		}
		attentionRefreshInFlight = (async () => {
			await loadSidebarAttention();
			lastAttentionLoadedAt = Date.now();
		})().finally(() => {
			attentionRefreshInFlight = null;
		});
		return attentionRefreshInFlight;
	}

	function openCommandPalette() {
		commandOpen = true;
	}

	function rememberRecent(href: string) {
		recentItems = [href, ...recentItems.filter((value) => value !== href)].slice(0, RECENT_LIMIT);
	}

	async function runCommand(href: string) {
		commandOpen = false;
		open = false;
		rememberRecent(href);
		await goto(resolveNavHref(href));
	}

	function loadPinnedItems() {
		if (typeof window === 'undefined') return;
		const raw = window.localStorage.getItem(pinnedStorageKey());
		let nextPinnedItems: string[] = [];
		if (raw) {
			try {
				const parsed = JSON.parse(raw);
				nextPinnedItems = normalizePinnedHrefs(parsed);
			} catch {
				nextPinnedItems = [];
			}
		}

		if (nextPinnedItems.length === 0) {
			nextPinnedItems = defaultPinnedItemsForUser();
		}

		pinnedItems = nextPinnedItems;
		pinnedLoaded = true;
	}

	function loadRecentItems() {
		if (typeof window === 'undefined') return;
		const raw = window.localStorage.getItem(recentStorageKey());
		if (!raw) {
			recentItems = [];
			return;
		}
		try {
			const parsed = JSON.parse(raw);
			recentItems = normalizeRecentHrefs(parsed);
		} catch {
			recentItems = [];
		}
	}

	onMount(() => {
		loadPinnedItems();
		loadRecentItems();
		void refreshSidebarAttention(true);
		const handleKeydown = (event: KeyboardEvent) => {
			if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'k') {
				event.preventDefault();
				openCommandPalette();
				return;
			}
			if (event.key === '/' && !event.ctrlKey && !event.metaKey && !event.altKey) {
				const target = event.target;
				if (
					target instanceof HTMLInputElement ||
					target instanceof HTMLTextAreaElement ||
					target instanceof HTMLSelectElement ||
					(target instanceof HTMLElement && target.isContentEditable)
				) {
					return;
				}
				event.preventDefault();
				openCommandPalette();
			}
		};
		const handleVisibilityChange = () => {
			if (document.visibilityState === 'visible') {
				void refreshSidebarAttention();
			}
		};
		const handleWindowFocus = () => {
			void refreshSidebarAttention();
		};
		window.addEventListener('keydown', handleKeydown);
		window.addEventListener('focus', handleWindowFocus);
		document.addEventListener('visibilitychange', handleVisibilityChange);
		const interval = window.setInterval(() => {
			void refreshSidebarAttention();
		}, ATTENTION_REFRESH_INTERVAL_MS);
		return () => {
			window.removeEventListener('keydown', handleKeydown);
			window.removeEventListener('focus', handleWindowFocus);
			document.removeEventListener('visibilitychange', handleVisibilityChange);
			window.clearInterval(interval);
		};
	});

	$effect(() => {
		page.url.pathname;
		if (typeof window === 'undefined') return;
		void refreshSidebarAttention();
	});

	$effect(() => {
		if (!pinnedLoaded || typeof window === 'undefined') return;
		window.localStorage.setItem(pinnedStorageKey(), JSON.stringify(normalizePinnedHrefs(pinnedItems)));
	});

	$effect(() => {
		if (!pinnedLoaded || typeof window === 'undefined') return;
		window.localStorage.setItem(recentStorageKey(), JSON.stringify(normalizeRecentHrefs(recentItems)));
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
	<button
		type="button"
		class="inline-flex h-8 items-center rounded-lg border border-slate-200 px-3 text-xs font-medium text-slate-500 hover:bg-slate-50"
		onclick={openCommandPalette}
	>
		Cari menu
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
		<div class={desktopExpanded ? 'block' : 'block lg:hidden'}>
			<button
				type="button"
				class="flex w-full items-center justify-between rounded-xl border border-slate-200 bg-slate-50 px-3 py-2 text-left text-sm text-slate-600 transition-colors hover:bg-white hover:text-slate-800"
				onclick={openCommandPalette}
				aria-label="Buka pencarian menu"
			>
				<span>Cari menu atau modul…</span>
				<span class="rounded-md border border-slate-200 bg-white px-1.5 py-0.5 text-[10px] font-semibold text-slate-400">Ctrl K</span>
			</button>
		</div>
		<SidebarQuickAccess
			items={quickAccess}
			{desktopExpanded}
			{isActive}
			{rememberRecent}
			{togglePin}
			{isPinned}
			{pinButtonLabel}
			{canMovePinned}
			{movePinned}
			{railTooltip}
			closeMobile={() => (open = false)}
		/>

		{#each nav as section (section.group)}
			<SidebarNavSection
				{section}
				{desktopExpanded}
				{isGroupOpen}
				{toggleGroup}
				{groupBadge}
				{navBadge}
				{isActive}
				{rememberRecent}
				{togglePin}
				{isPinned}
				{pinButtonLabel}
				{railTooltip}
				closeMobile={() => (open = false)}
			/>
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

<SidebarCommandPalette
	bind:open={commandOpen}
	items={commandItems}
	recentHrefs={recentItems}
	runCommand={runCommand}
	clearRecent={() => (recentItems = [])}
	{isActive}
/>
