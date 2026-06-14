<script lang="ts">
	import { onMount } from 'svelte';
	import ChevronDownIcon from '@lucide/svelte/icons/chevron-down';
	import LogOutIcon from '@lucide/svelte/icons/log-out';
	import SettingsIcon from '@lucide/svelte/icons/settings';
	import ShieldCheckIcon from '@lucide/svelte/icons/shield-check';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Separator } from '$lib/components/ui/separator';
	import { clearCbtComposerDrafts } from '$lib/client/cbt-drafts';
	import {
		accountAvatarUrl,
		accountDisplayName,
		accountInitials,
		roleLabel,
		type AccountIdentity
	} from '$lib/client/account';
	import { cn } from '$lib/utils';

	type AccountMenuUser = {
		id: string;
		username: string;
		role?: string;
		roles?: string[];
	};

	type MenuSide = 'top' | 'bottom';
	type MenuAlign = 'start' | 'end';
	type NameMode = boolean | 'responsive';

	let {
		user,
		account = null,
		showName = false,
		menuSide = 'bottom',
		align = 'end',
		menuId = 'account-menu',
		class: rootClass = '',
		buttonClass = '',
		menuClass = '',
		includeThemeToggle = true
	}: {
		user?: AccountMenuUser;
		account?: AccountIdentity | null;
		showName?: NameMode;
		menuSide?: MenuSide;
		align?: MenuAlign;
		menuId?: string;
		class?: string;
		buttonClass?: string;
		menuClass?: string;
		includeThemeToggle?: boolean;
	} = $props();

	let root = $state<HTMLDivElement | null>(null);
	let trigger = $state<HTMLButtonElement | null>(null);
	let open = $state(false);
	let avatarFailed = $state(false);
	let logoutLoading = $state(false);

	const effectiveAccount = $derived(accountFromProps(account, user));
	const displayName = $derived(accountDisplayName(effectiveAccount));
	const compactName = $derived(shortDisplayName(displayName));
	const username = $derived(effectiveAccount?.username?.trim() || user?.username?.trim() || 'akun');
	const activeRoles = $derived(activeRoleCodes(effectiveAccount, user));
	const initials = $derived(accountInitials(effectiveAccount));
	const avatarUrl = $derived(accountAvatarUrl(effectiveAccount));
	const canShowAvatar = $derived(avatarUrl && !avatarFailed);
	const nameClass = $derived(
		showName === 'responsive'
			? 'hidden min-w-0 text-left sm:block'
			: showName
				? 'min-w-0 text-left'
				: 'sr-only'
	);
	const chevronClass = $derived(showName === 'responsive' ? 'hidden size-3.5 text-muted-foreground sm:block' : showName ? 'size-3.5 text-muted-foreground' : 'hidden');
	const triggerSizeClass = $derived(showName === false ? 'size-9 justify-center p-0' : 'h-9 min-w-0 px-1.5 pr-2');
	const triggerClass = $derived(cn(
		'inline-flex max-w-full items-center gap-2 rounded-lg border border-border/70 bg-background/70 text-left text-foreground shadow-sm transition-colors hover:bg-muted focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 focus-visible:outline-none aria-expanded:bg-muted aria-expanded:text-foreground',
		triggerSizeClass,
		buttonClass
	));
	const positionClass = $derived(cn(
		menuSide === 'top' ? 'bottom-full mb-2' : 'top-full mt-2',
		align === 'start' ? 'left-0' : 'right-0'
	));
	const panelClass = $derived(cn(
	    'absolute z-50 w-72 overflow-hidden rounded-xl border border-surface-300/60 dark:border-surface-600/60 bg-surface-50-950 p-1.5 text-popover-foreground shadow-2xl ring-1 ring-black/10 dark:ring-white/10 backdrop-blur-md',
	    positionClass,
	    menuClass
	));

	$effect(() => {
		avatarUrl;
		avatarFailed = false;
	});

	function accountFromProps(
		accountValue: AccountIdentity | null | undefined,
		userValue: AccountMenuUser | undefined
	): AccountIdentity | null {
		if (accountValue) {
			return {
				...accountValue,
				username: accountValue.username?.trim() || userValue?.username?.trim() || 'akun',
				roles: activeRoleCodes(accountValue, userValue)
			};
		}
		if (!userValue) return null;
		return {
			id: userValue.id,
			username: userValue.username?.trim() || 'akun',
			display_name: userValue.username?.trim() || 'Akun',
			roles: activeRoleCodes(null, userValue)
		};
	}

	function activeRoleCodes(
		accountValue: Pick<AccountIdentity, 'roles'> | null | undefined,
		userValue: AccountMenuUser | undefined
	) {
		const source = accountValue?.roles?.length
			? accountValue.roles
			: userValue?.roles?.length
				? userValue.roles
				: userValue?.role
					? [userValue.role]
					: [];
		return Array.from(new Set(source.map((role) => role.trim()).filter(Boolean)));
	}

	function shortDisplayName(value: string) {
		const normalized = value.trim();
		if (!normalized) return 'Akun';
		if (normalized.length <= 20) return normalized;
		const parts = normalized.split(/\s+/).filter(Boolean);
		if (parts.length >= 2) return `${parts[0]} ${parts[parts.length - 1]}`;
		return normalized;
	}

	function closeMenu(restoreFocus = false) {
		open = false;
		if (restoreFocus) trigger?.focus();
	}

	function toggleMenu() {
		open = !open;
	}

	async function logout() {
		if (logoutLoading) return;
		logoutLoading = true;
		try {
			await fetch('/api/auth/logout', { method: 'POST' });
		} finally {
			clearCbtComposerDrafts();
			location.href = '/login';
		}
	}

	onMount(() => {
		const handlePointerDown = (event: PointerEvent) => {
			if (!open || !root) return;
			const target = event.target;
			if (target instanceof Node && !root.contains(target)) {
				closeMenu();
			}
		};
		const handleKeydown = (event: KeyboardEvent) => {
			if (!open || event.key !== 'Escape') return;
			event.preventDefault();
			closeMenu(true);
		};

		document.addEventListener('pointerdown', handlePointerDown);
		window.addEventListener('keydown', handleKeydown);
		return () => {
			document.removeEventListener('pointerdown', handlePointerDown);
			window.removeEventListener('keydown', handleKeydown);
		};
	});
</script>

<div bind:this={root} class={cn('relative inline-flex min-w-0', rootClass)}>
	<button
		bind:this={trigger}
		type="button"
		class={triggerClass}
		aria-label={`Buka menu akun ${displayName}`}
		aria-haspopup="menu"
		aria-expanded={open}
		aria-controls={menuId}
		title={displayName}
		onclick={toggleMenu}
	>
		<span class="flex size-8 shrink-0 items-center justify-center overflow-hidden rounded-full border border-border bg-primary/10 text-xs font-semibold text-primary">
			{#if canShowAvatar}
				<img src={avatarUrl} alt="" class="size-full object-cover" onerror={() => (avatarFailed = true)} />
			{:else}
				<span aria-hidden="true">{initials}</span>
			{/if}
		</span>
		<span class={nameClass}>
			<span class="block truncate text-sm font-semibold leading-tight text-foreground">{compactName}</span>
			<span class="block truncate text-[11px] leading-tight text-muted-foreground">@{username}</span>
		</span>
		<ChevronDownIcon class={chevronClass} />
	</button>

	{#if open}
		<div id={menuId} class={panelClass} role="menu" aria-label="Menu akun">
			<div class="flex items-start gap-3 rounded-md px-2 py-2.5" role="group" aria-label="Akun aktif">
				<span class="flex size-10 shrink-0 items-center justify-center overflow-hidden rounded-full border border-border bg-primary/10 text-sm font-semibold text-primary">
					{#if canShowAvatar}
						<img src={avatarUrl} alt="" class="size-full object-cover" onerror={() => (avatarFailed = true)} />
					{:else}
						<span aria-hidden="true">{initials}</span>
					{/if}
				</span>
				<div class="min-w-0">
					<p class="truncate text-sm font-semibold text-popover-foreground">{displayName}</p>
					<p class="truncate text-xs text-muted-foreground">@{username}</p>
				</div>
			</div>

			<div class="px-2 pb-2" role="group" aria-label="Peran aktif">
				<div class="mb-1 flex items-center gap-1.5 text-[11px] font-medium uppercase tracking-[0.14em] text-muted-foreground">
					<ShieldCheckIcon class="size-3.5" />
					<span>Peran aktif</span>
				</div>
				<div class="flex flex-wrap gap-1.5">
					{#if activeRoles.length > 0}
						{#each activeRoles as role (role)}
							<Badge variant="secondary" class="border border-border bg-secondary text-secondary-foreground">{roleLabel(role)}</Badge>
						{/each}
					{:else}
						<Badge variant="outline">Pengguna</Badge>
					{/if}
				</div>
			</div>

			<Separator orientation="horizontal" class="my-1" />

			<a
				href="/settings/account"
				role="menuitem"
				class="flex items-center gap-2 rounded-md px-2 py-2 text-sm text-popover-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:bg-accent focus-visible:text-accent-foreground focus-visible:outline-none"
				onclick={() => closeMenu()}
			>
				<SettingsIcon class="size-4" />
				<span class="min-w-0">
					<span class="block font-medium">Profil Saya</span>
					<span class="block text-xs text-muted-foreground">Pengaturan Akun</span>
				</span>
			</a>

			{#if includeThemeToggle}
				<div class="py-1" role="none">
					<ThemeToggle
						expanded={true}
						variant="ghost"
						size="sm"
						class="w-full justify-start px-2 text-popover-foreground hover:bg-accent hover:text-accent-foreground"
					/>
				</div>
			{/if}

			<Separator orientation="horizontal" class="my-1" />

			<button
				type="button"
				role="menuitem"
				class="flex w-full items-center gap-2 rounded-md px-2 py-2 text-left text-sm font-medium text-destructive transition-colors hover:bg-destructive/10 focus-visible:bg-destructive/10 focus-visible:outline-none disabled:pointer-events-none disabled:opacity-60"
				disabled={logoutLoading}
				onclick={() => void logout()}
			>
				<LogOutIcon class="size-4" />
				<span>{logoutLoading ? 'Keluar...' : 'Logout'}</span>
			</button>
		</div>
	{/if}
</div>
