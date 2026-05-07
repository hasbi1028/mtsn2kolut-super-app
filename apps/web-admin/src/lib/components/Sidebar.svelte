<script lang="ts">
	import { onMount } from 'svelte';
	import { afterNavigate, goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import AccountMenu from '$lib/components/AccountMenu.svelte';
	import SidebarCommandPalette from '$lib/components/sidebar/SidebarCommandPalette.svelte';
	import SidebarIcon from '$lib/components/sidebar/SidebarIcon.svelte';
	import SidebarNavSection from '$lib/components/sidebar/SidebarNavSection.svelte';
	import SidebarQuickAccess from '$lib/components/sidebar/SidebarQuickAccess.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import { fetchSidebarAttention } from '$lib/components/sidebar/sidebar-attention';
	import { filterSidebarNavGroupsByAccess } from '$lib/components/sidebar/sidebar-access';
	import { findActiveSidebarHref } from '$lib/components/sidebar/sidebar-active';
	import { readClientJson } from '$lib/client/api';
	import type { AccountIdentity } from '$lib/client/account';
	import {
		defaultPinnedByRole,
		sidebarNavGroups,
		type SidebarNavItem
	} from '$lib/components/sidebar/sidebar-config';

	let {
		user,
		account = null,
		desktopExpanded = $bindable(true)
	}: {
		user?: { id: string; username: string; role: string; roles?: string[]; permissions?: string[]; employee_id?: string };
		account?: AccountIdentity | null;
		desktopExpanded?: boolean;
	} = $props();
	let open = $state(false);
	let commandOpen = $state(false);
	let inventoryAttention = $state(0);
	let libraryAttention = $state(0);
	let pusakaAttention = $state(0);
	let profileChangeAttention = $state(0);
	let attentionRefreshInFlight: Promise<void> | null = null;
	let lastAttentionLoadedAt = 0;
	let remotePrefsLoaded = false;
	let sidebarPrefsSyncInFlight: Promise<void> | null = null;
	let lastSyncedSidebarPrefs = '';
	let sidebarPrefsLocalVersion = 0;

	type SidebarPreferencesPayload = {
		data?: unknown;
		pinned_items?: unknown;
		recent_items?: unknown;
	};

	const userRoles = $derived(user?.roles || (user?.role ? [user.role] : []));
	const userPermissions = $derived(user?.permissions || []);
	const nav = $derived(filterSidebarNavGroupsByAccess(sidebarNavGroups, userRoles, userPermissions));

	const PINNED_STORAGE_KEY_PREFIX = 'sidebar:pinned-items';
	const RECENT_STORAGE_KEY_PREFIX = 'sidebar:recent-items';
	const RECENT_LIMIT = 6;
	const ATTENTION_REFRESH_INTERVAL_MS = 60_000;
	let openGroups = $state<string[]>([]);
	let pinnedItems = $state<string[]>([]);
	let recentItems = $state<string[]>([]);
	let pinnedLoaded = $state(false);

	const visibleNavItems = $derived(
		nav.flatMap((section) =>
			section.items.map((item) => ({
				...item,
				group: section.group,
			}))
		)
	);

	const activeHref = $derived(findActiveSidebarHref(page.url.pathname, visibleNavItems));

	const activeGroup = $derived(
		nav.find((section) => section.items.some((item) => item.href === activeHref))?.group ?? 'Utama'
	);

	const visibleNavHrefSet = $derived(new Set(visibleNavItems.map((item) => item.href)));
	const pinnableNavHrefSet = $derived(
		new Set(
			visibleNavItems
				.filter((item) => item.href !== '/' && item.pinnable !== false)
				.map((item) => item.href)
		)
	);
	const pinnedStorageKeyValue = $derived(`${PINNED_STORAGE_KEY_PREFIX}:${storageScope()}`);
	const recentStorageKeyValue = $derived(`${RECENT_STORAGE_KEY_PREFIX}:${storageScope()}`);
	const normalizedPinnedItems = $derived(normalizePinnedHrefs(pinnedItems));
	const normalizedRecentItems = $derived(normalizeRecentHrefs(recentItems));
	const sidebarPrefsSignatureValue = $derived(JSON.stringify({
		pinned_items: normalizedPinnedItems,
		recent_items: normalizedRecentItems
	}));

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
		return activeHref === href;
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
			markSidebarPrefsChanged();
			return;
		}
		pinnedItems = [...pinnedItems, href];
		markSidebarPrefsChanged();
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
		markSidebarPrefsChanged();
	}

	function railTooltip(item: SidebarNavItem, group: string) {
		return `${group} · ${item.label}`;
	}

	function storageScope() {
		return user?.id?.trim() || 'anon';
	}

	function isRecord(value: unknown): value is Record<string, unknown> {
		return typeof value === 'object' && value !== null;
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
		if (href === '/settings/user-change-requests') return profileChangeAttention;
		return 0;
	}

	function groupBadge(group: string) {
		if (group === 'Aset & Layanan') return inventoryAttention + libraryAttention;
		if (group === 'Pegawai & PUSAKA') return pusakaAttention;
		if (group === 'Sistem') return profileChangeAttention;
		return 0;
	}

	async function loadSidebarAttention() {
		const attention = await fetchSidebarAttention(fetch, userRoles, userPermissions);
		inventoryAttention = attention.inventory;
		libraryAttention = attention.library;
		pusakaAttention = attention.pusaka;
		profileChangeAttention = attention.profileChanges;
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
		markSidebarPrefsChanged();
	}

	async function runCommand(href: string) {
		commandOpen = false;
		open = false;
		rememberRecent(href);
		await goto(resolve(href as '/'));
	}

	function loadPinnedItems() {
		if (typeof window === 'undefined') return;
		const raw = window.localStorage.getItem(pinnedStorageKeyValue);
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
		persistSidebarCache();
	}

	function loadRecentItems() {
		if (typeof window === 'undefined') return;
		const raw = window.localStorage.getItem(recentStorageKeyValue);
		if (!raw) {
			recentItems = [];
			return;
		}
		try {
			const parsed = JSON.parse(raw);
			recentItems = normalizeRecentHrefs(parsed);
			persistSidebarCache();
		} catch {
			recentItems = [];
			persistSidebarCache();
		}
	}

	function persistSidebarCache() {
		if (!pinnedLoaded || typeof window === 'undefined') return;
		window.localStorage.setItem(pinnedStorageKeyValue, JSON.stringify(normalizePinnedHrefs(pinnedItems)));
		window.localStorage.setItem(recentStorageKeyValue, JSON.stringify(normalizeRecentHrefs(recentItems)));
	}

	function markSidebarPrefsChanged() {
		sidebarPrefsLocalVersion += 1;
		persistSidebarCache();
		void queueRemoteSidebarPreferences();
	}

	function sidebarPrefsSource(payload: SidebarPreferencesPayload | null): Record<string, unknown> {
		if (!payload || typeof payload !== 'object') return {};
		if (isRecord(payload.data)) return payload.data;
		return payload as Record<string, unknown>;
	}

	async function loadRemoteSidebarPreferences() {
		if (typeof window === 'undefined') return;
		const versionAtStart = sidebarPrefsLocalVersion;
		try {
			const res = await fetch('/api/auth/preferences/sidebar');
			const payload = await readClientJson<SidebarPreferencesPayload | null>(res);
			const source = sidebarPrefsSource(payload);
			const remotePinned = normalizePinnedHrefs(source.pinned_items ?? []);
			const remoteRecent = normalizeRecentHrefs(source.recent_items ?? []);

			if (versionAtStart === sidebarPrefsLocalVersion) {
				if (remotePinned.length > 0) {
					pinnedItems = remotePinned;
				} else if (pinnedItems.length === 0) {
					pinnedItems = defaultPinnedItemsForUser();
				}
				recentItems = remoteRecent;
				persistSidebarCache();
				lastSyncedSidebarPrefs = JSON.stringify({
					pinned_items: remotePinned,
					recent_items: remoteRecent
				});
			}
		} catch {
			// Keep local cache as fallback for sidebar personalization.
		} finally {
			remotePrefsLoaded = true;
			if (versionAtStart !== sidebarPrefsLocalVersion) {
				void queueRemoteSidebarPreferences();
			}
		}
	}

	async function queueRemoteSidebarPreferences() {
		if (typeof window === 'undefined' || !remotePrefsLoaded) return;
		const signature = sidebarPrefsSignatureValue;
		if (signature === lastSyncedSidebarPrefs) return;
		if (sidebarPrefsSyncInFlight) return sidebarPrefsSyncInFlight;

		const versionAtStart = sidebarPrefsLocalVersion;
		let shouldResyncAfterFlight = false;
		sidebarPrefsSyncInFlight = fetch('/api/auth/preferences/sidebar', {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: signature
		})
			.then(async (res) => {
				if (!res.ok) {
					return;
				}
				const payload = await readClientJson<SidebarPreferencesPayload | null>(res);
				const source = sidebarPrefsSource(payload);
				if (versionAtStart === sidebarPrefsLocalVersion) {
					const syncedPinned = normalizePinnedHrefs(source.pinned_items ?? pinnedItems);
					const syncedRecent = normalizeRecentHrefs(source.recent_items ?? recentItems);
					lastSyncedSidebarPrefs = JSON.stringify({
						pinned_items: syncedPinned,
						recent_items: syncedRecent
					});
				} else {
					shouldResyncAfterFlight = true;
				}
			})
			.catch(() => {
				// Keep local cache; the next sidebar change or visibility refresh will retry.
			})
			.finally(() => {
				sidebarPrefsSyncInFlight = null;
				if (shouldResyncAfterFlight && sidebarPrefsSignatureValue !== lastSyncedSidebarPrefs) {
					void queueRemoteSidebarPreferences();
				}
			});
		return sidebarPrefsSyncInFlight;
	}

	onMount(() => {
		loadPinnedItems();
		loadRecentItems();
		void loadRemoteSidebarPreferences();
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

	afterNavigate(() => {
		if (typeof window !== 'undefined') {
			void refreshSidebarAttention();
		}
	});

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
<header class="fixed inset-x-0 top-0 z-20 flex h-14 w-full items-center gap-3 border-b border-border bg-card px-4 text-card-foreground lg:hidden">
	<button
		class="rounded-md p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground"
		onclick={() => (open = !open)}
		aria-label="Toggle menu"
	>
		<svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
		</svg>
	</button>
	<button
		type="button"
		class="inline-flex h-8 items-center rounded-lg border border-border px-3 text-xs font-medium text-muted-foreground hover:bg-muted hover:text-foreground"
		onclick={openCommandPalette}
	>
		Cari menu
	</button>
	<span class="min-w-0 flex-1 truncate text-sm font-semibold text-foreground">MTSN 2 Kolaka Utara</span>
	{#if user}
		<AccountMenu
			{user}
			{account}
			menuId="mobile-topbar-account-menu"
			buttonClass="border-transparent bg-transparent shadow-none hover:bg-muted"
		/>
	{/if}
</header>

<!-- Sidebar -->
<aside
	class={`fixed inset-y-0 left-0 z-30 flex flex-col border-r border-border bg-card text-card-foreground
	       transition-transform duration-200
	       ${open ? 'translate-x-0' : '-translate-x-full'}
	       ${desktopExpanded ? 'lg:w-60' : 'lg:w-[5.5rem]'}
	       lg:translate-x-0`}
>
	<!-- Brand -->
	<div class={`flex h-14 shrink-0 items-center border-b border-border ${desktopExpanded ? 'gap-2.5 px-4' : 'justify-center px-3'}`}>
		<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-primary text-xs font-bold text-primary-foreground">
			MTs
		</div>
		<div class={`min-w-0 ${desktopExpanded ? 'block' : 'block lg:hidden'}`}>
			<p class="truncate text-sm font-semibold text-foreground">MTSN 2 Kolut</p>
			<p class="truncate text-xs text-muted-foreground">Kolaka Utara</p>
		</div>
		<button
			class={`ml-auto hidden rounded-md p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground lg:inline-flex ${desktopExpanded ? '' : 'ml-0'}`}
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
				class="flex w-full items-center justify-between rounded-xl border border-border bg-muted/40 px-3 py-2 text-left text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
				onclick={openCommandPalette}
				aria-label="Buka pencarian menu"
			>
				<span>Cari menu atau modul…</span>
				<span class="rounded-md border border-border bg-background px-1.5 py-0.5 text-[10px] font-semibold text-muted-foreground">Ctrl K</span>
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
	<div class="shrink-0 space-y-2 border-t border-border p-3">
		<ThemeToggle expanded={desktopExpanded} variant="outline" size="sm" class={desktopExpanded ? 'w-full justify-start' : 'w-full lg:justify-center'} />
		{#if user}
			<AccountMenu
				{user}
				{account}
				showName={desktopExpanded}
				menuSide="top"
				align="start"
				menuId="desktop-sidebar-account-menu"
				class="w-full"
				buttonClass={desktopExpanded ? 'w-full justify-start' : 'w-full justify-center'}
			/>
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
