<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import KeyRoundIcon from '@lucide/svelte/icons/key-round';
	import LogOutIcon from '@lucide/svelte/icons/log-out';
	import RefreshCcwIcon from '@lucide/svelte/icons/refresh-ccw';
	import SaveIcon from '@lucide/svelte/icons/save';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { toast } from '$lib/components/ui/sonner';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { clearCbtComposerDrafts } from '$lib/client/cbt-drafts';
	import { clientApiPath, readClientApiData, readClientJson } from '$lib/client/api';
	import {
		accountDisplayName,
		accountErrorMessage,
		formatAccountDateTime,
		isCurrentSession,
		linkedProfileLabel,
		normalizeAccountSessions,
		preferenceItemCount,
		profileTypeLabel,
		roleLabel,
		sessionTitle,
		type AccountIdentity,
		type AuthSession,
		type SidebarPreferences
	} from '$lib/client/account';
	import { confirmAction } from '$lib/confirm-dialog';

	type AccountOverview = {
		account: AccountIdentity;
		sessions: AuthSession[];
		preferences: SidebarPreferences | null;
	};

	let overviewPromise = $state<Promise<AccountOverview> | null>(null);
	let account = $state<AccountIdentity | null>(null);
	let sessions = $state<AuthSession[]>([]);
	let preferences = $state<SidebarPreferences | null>(null);
	let labelDrafts = $state<Record<string, string>>({});
	let pwForm = $state({ current: '', next: '', confirm: '' });
	let pwLoading = $state(false);
	let logoutAllLoading = $state(false);
	let revokeSessionLoading = $state<string | null>(null);
	let renameSessionLoading = $state<string | null>(null);
	let refreshLoading = $state(false);

	const currentSessionId = $derived(page.data.user?.session_id ?? '');
	const pinnedCount = $derived(preferenceItemCount(preferences?.pinned_items));
	const recentCount = $derived(preferenceItemCount(preferences?.recent_items));

	function applyOverview(overview: AccountOverview) {
		account = overview.account;
		sessions = overview.sessions;
		preferences = overview.preferences;
		labelDrafts = Object.fromEntries(
			overview.sessions.map((session) => [session.id, session.device_label?.trim() ?? ''])
		);
	}

	function currentOverview(): AccountOverview | null {
		if (!account) return null;
		return { account, sessions, preferences };
	}

	async function fetchPreferences() {
		try {
			const res = await fetch('/api/auth/preferences/sidebar');
			return await readClientApiData<SidebarPreferences>(res, 'Gagal memuat preferensi tampilan');
		} catch {
			return null;
		}
	}

	async function fetchOverview(): Promise<AccountOverview> {
		const [accountData, sessionData, preferenceData] = await Promise.all([
			fetch('/api/auth/account').then((response) =>
				readClientApiData<AccountIdentity>(response, 'Gagal memuat identitas akun')
			),
			fetch('/api/auth/sessions').then((response) =>
				readClientApiData<AuthSession[]>(response, 'Gagal memuat sesi aktif')
			),
			fetchPreferences()
		]);

		return {
			account: accountData,
			sessions: normalizeAccountSessions(sessionData),
			preferences: preferenceData
		};
	}

	function loadOverview() {
		overviewPromise = fetchOverview().then((overview) => {
			applyOverview(overview);
			return overview;
		});
		return overviewPromise;
	}

	async function refreshOverview(showFailureToast = false) {
		refreshLoading = true;
		try {
			const overview = await fetchOverview();
			applyOverview(overview);
			overviewPromise = Promise.resolve(overview);
		} catch (error) {
			const current = currentOverview();
			if (current) overviewPromise = Promise.resolve(current);
			if (showFailureToast) {
				toast.error(accountErrorMessage(error, 'Gagal menyegarkan data akun'));
			}
		} finally {
			refreshLoading = false;
		}
	}

	function retryOverview(reset?: () => void) {
		reset?.();
		loadOverview();
	}

	function handleOverviewRenderError(error: unknown, reset: () => void) {
		console.error('Account overview render failed', error);
		reset();
	}

	async function changePassword() {
		if (!pwForm.current || !pwForm.next) {
			toast.error('Password saat ini dan password baru wajib diisi');
			return;
		}
		if (pwForm.next !== pwForm.confirm) {
			toast.error('Konfirmasi password tidak cocok');
			return;
		}
		if (pwForm.next.length < 8) {
			toast.error('Password baru minimal 8 karakter');
			return;
		}

		pwLoading = true;
		try {
			const res = await fetch('/api/auth/change-password', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ current_password: pwForm.current, new_password: pwForm.next })
			});
			await readClientJson<unknown>(res);
			pwForm = { current: '', next: '', confirm: '' };
			toast.success('Password berhasil diubah. Silakan login ulang.');
			await fetch('/api/auth/logout', { method: 'POST' }).catch(() => undefined);
			clearCbtComposerDrafts();
			window.setTimeout(() => {
				window.location.href = '/login';
			}, 500);
		} catch (error) {
			toast.error(accountErrorMessage(error, 'Gagal mengubah password'));
		} finally {
			pwLoading = false;
		}
	}

	async function logoutAllSessions() {
		if (!(await confirmAction({
			title: 'Keluar dari Semua Sesi',
			message: 'Semua sesi login akun Anda akan diakhiri, termasuk perangkat ini.',
			confirmLabel: 'Keluar Semua',
			tone: 'danger'
		}))) return;

		logoutAllLoading = true;
		try {
			const res = await fetch('/api/auth/logout-all', { method: 'POST' });
			await readClientJson<unknown>(res);
			clearCbtComposerDrafts();
			window.location.href = '/login';
		} catch (error) {
			toast.error(accountErrorMessage(error, 'Gagal mengakhiri semua sesi'));
		} finally {
			logoutAllLoading = false;
		}
	}

	async function revokeSession(sessionId: string) {
		const target = sessions.find((session) => session.id === sessionId);
		if (!(await confirmAction({
			title: 'Akhiri Sesi',
			message: `Akhiri ${target ? sessionTitle(target) : 'sesi ini'}?`,
			confirmLabel: 'Akhiri Sesi',
			tone: isCurrentSession({ id: sessionId }, currentSessionId) ? 'danger' : 'warning'
		}))) return;

		revokeSessionLoading = sessionId;
		try {
			const res = await fetch(clientApiPath`/api/auth/sessions/${sessionId}`, { method: 'DELETE' });
			await readClientJson<unknown>(res);

			if (sessionId === currentSessionId) {
				await fetch('/api/auth/logout', { method: 'POST' }).catch(() => undefined);
				clearCbtComposerDrafts();
				window.location.href = '/login';
				return;
			}

			sessions = sessions.filter((session) => session.id !== sessionId);
			const current = currentOverview();
			if (current) overviewPromise = Promise.resolve(current);
			toast.success('Sesi berhasil diakhiri.');
		} catch (error) {
			toast.error(accountErrorMessage(error, 'Gagal mengakhiri sesi'));
		} finally {
			revokeSessionLoading = null;
		}
	}

	async function renameSession(sessionId: string) {
		const deviceLabel = (labelDrafts[sessionId] ?? '').trim();
		if (!deviceLabel) {
			toast.error('Nama perangkat tidak boleh kosong');
			return;
		}

		renameSessionLoading = sessionId;
		try {
			const res = await fetch(clientApiPath`/api/auth/sessions/${sessionId}`, {
				method: 'PATCH',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ device_label: deviceLabel })
			});
			await readClientJson<unknown>(res);
			sessions = sessions.map((session) =>
				session.id === sessionId ? { ...session, device_label: deviceLabel } : session
			);
			const current = currentOverview();
			if (current) overviewPromise = Promise.resolve(current);
			toast.success('Nama perangkat berhasil disimpan.');
		} catch (error) {
			toast.error(accountErrorMessage(error, 'Gagal menyimpan nama perangkat'));
		} finally {
			renameSessionLoading = null;
		}
	}

	onMount(() => {
		void loadOverview();
	});
</script>

<svelte:head><title>Akun Saya - MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
		<div>
			<h1 class="text-2xl font-semibold text-slate-800">Akun Saya</h1>
			<p class="mt-1 text-sm text-muted-foreground">Identitas login, password, sesi perangkat, dan preferensi pribadi.</p>
		</div>
		<Button variant="outline" onclick={() => void refreshOverview(true)} disabled={refreshLoading}>
			<RefreshCcwIcon class={`size-4 ${refreshLoading ? 'animate-spin' : ''}`} />
			Refresh
		</Button>
	</div>

	<AsyncContent promise={overviewPromise} onerror={handleOverviewRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<Card.Root>
					<Card.Header class="pb-3">
						<Skeleton class="h-5 w-32" />
						<Skeleton class="h-4 w-72" />
					</Card.Header>
					<Card.Content>
						<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
							{#each Array.from({ length: 4 }) as _, index (`account-field-skeleton-${index}`)}
								<div class="rounded-lg border border-slate-200 p-4">
									<Skeleton class="h-4 w-24" />
									<Skeleton class="mt-3 h-5 w-44" />
								</div>
							{/each}
						</div>
					</Card.Content>
				</Card.Root>
				<Card.Root>
					<Card.Header class="pb-3">
						<Skeleton class="h-5 w-40" />
						<Skeleton class="h-4 w-80" />
					</Card.Header>
					<Card.Content class="space-y-3">
						{#each Array.from({ length: 2 }) as _, index (`account-session-skeleton-${index}`)}
							<div class="rounded-lg border border-slate-200 p-4">
								<Skeleton class="h-5 w-40" />
								<Skeleton class="mt-2 h-4 w-56" />
							</div>
						{/each}
					</Card.Content>
				</Card.Root>
			</div>
		{/snippet}
		{#snippet failed(error, reset)}
			<RecoveryPanel title="Akun Belum Tersaji" message={accountErrorMessage(error, 'Gagal memuat data akun')} onRetry={() => retryOverview(reset)} />
		{/snippet}
		{#snippet children(_overview)}
			{#if account}
			<div class="grid grid-cols-1 gap-6 xl:grid-cols-[minmax(0,1fr)_360px]">
				<div class="space-y-6">
					<Card.Root>
						<Card.Header class="pb-3">
							<Card.Title class="text-base">Identitas Login</Card.Title>
							<Card.Description>Data resmi tetap dikelola oleh modul master terkait.</Card.Description>
						</Card.Header>
						<Card.Content>
							<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
								<div class="rounded-lg border border-slate-200 p-4">
									<p class="text-xs font-medium uppercase tracking-[0.16em] text-slate-500">Nama tampil</p>
									<p class="mt-2 text-sm font-semibold text-slate-900">{accountDisplayName(account)}</p>
								</div>
								<div class="rounded-lg border border-slate-200 p-4">
									<p class="text-xs font-medium uppercase tracking-[0.16em] text-slate-500">Username</p>
									<p class="mt-2 text-sm font-semibold text-slate-900">{account.username}</p>
								</div>
								<div class="rounded-lg border border-slate-200 p-4">
									<p class="text-xs font-medium uppercase tracking-[0.16em] text-slate-500">Role</p>
									<div class="mt-2 flex flex-wrap gap-2">
										{#if account.roles.length === 0}
											<Badge variant="outline">Belum ada role</Badge>
										{:else}
											{#each account.roles as role (role)}
												<Badge variant="secondary">{roleLabel(role)}</Badge>
											{/each}
										{/if}
									</div>
								</div>
								<div class="rounded-lg border border-slate-200 p-4">
									<p class="text-xs font-medium uppercase tracking-[0.16em] text-slate-500">Profil tertaut</p>
									<p class="mt-2 text-sm font-semibold text-slate-900">{linkedProfileLabel(account)}</p>
									<p class="mt-1 text-xs text-slate-500">{profileTypeLabel(account.profile_type)}</p>
								</div>
								<div class="rounded-lg border border-slate-200 p-4">
									<p class="text-xs font-medium uppercase tracking-[0.16em] text-slate-500">Login terakhir</p>
									<p class="mt-2 text-sm font-semibold text-slate-900">{formatAccountDateTime(account.last_login_at)}</p>
								</div>
								<div class="rounded-lg border border-slate-200 p-4">
									<p class="text-xs font-medium uppercase tracking-[0.16em] text-slate-500">Status akun</p>
									<p class="mt-2 text-sm font-semibold {account.is_active === false ? 'text-red-700' : 'text-green-700'}">
										{account.is_active === false ? 'Nonaktif' : 'Aktif'}
									</p>
								</div>
							</div>
						</Card.Content>
					</Card.Root>

					<Card.Root>
						<Card.Header class="pb-3">
							<Card.Title class="text-base">Ubah Password</Card.Title>
							<Card.Description>Gunakan password baru minimal 8 karakter.</Card.Description>
						</Card.Header>
						<Card.Content>
							<div class="grid grid-cols-1 gap-4 md:grid-cols-3">
								<div>
									<label for="account-pw-current" class="mb-1.5 block text-sm font-medium">Password Saat Ini</label>
									<Input id="account-pw-current" type="password" bind:value={pwForm.current} autocomplete="current-password" />
								</div>
								<div>
									<label for="account-pw-next" class="mb-1.5 block text-sm font-medium">Password Baru</label>
									<Input id="account-pw-next" type="password" bind:value={pwForm.next} autocomplete="new-password" />
								</div>
								<div>
									<label for="account-pw-confirm" class="mb-1.5 block text-sm font-medium">Konfirmasi Password</label>
									<Input id="account-pw-confirm" type="password" bind:value={pwForm.confirm} autocomplete="new-password" />
								</div>
							</div>
							<LoadingButton class="mt-4" onclick={() => void changePassword()} loading={pwLoading} loadingLabel="Menyimpan...">
								<KeyRoundIcon class="size-4" />
								Simpan Password
							</LoadingButton>
						</Card.Content>
					</Card.Root>

					<Card.Root>
						<Card.Header class="pb-3">
							<Card.Title class="text-base">Sesi Aktif</Card.Title>
							<Card.Description>Daftar perangkat yang masih memiliki sesi login aktif.</Card.Description>
						</Card.Header>
						<Card.Content>
							{#if sessions.length === 0}
								<p class="text-sm text-muted-foreground">Belum ada sesi aktif tercatat.</p>
							{:else}
								<div class="space-y-3">
									{#each sessions as session (session.id)}
										<div class="rounded-lg border border-slate-200 p-4">
											<div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
												<div class="min-w-0 flex-1 space-y-2">
													<div class="flex flex-wrap items-center gap-2">
														<p class="text-sm font-semibold text-slate-900">{sessionTitle(session)}</p>
														{#if isCurrentSession(session, currentSessionId)}
															<Badge variant="secondary">Perangkat Ini</Badge>
														{/if}
													</div>
													<div class="grid grid-cols-1 gap-2 text-xs text-slate-500 sm:grid-cols-2">
														<p>Terakhir aktif: {formatAccountDateTime(session.last_used_at)}</p>
														<p>Berlaku sampai: {formatAccountDateTime(session.expires_at)}</p>
														{#if session.ip_address}
															<p>IP: {session.ip_address}</p>
														{/if}
														{#if session.created_at}
															<p>Dibuat: {formatAccountDateTime(session.created_at)}</p>
														{/if}
													</div>
													{#if session.user_agent}
														<p class="line-clamp-2 text-[11px] text-slate-400">{session.user_agent}</p>
													{/if}
													<div class="pt-2">
														<label for={`account-session-label-${session.id}`} class="mb-1 block text-xs font-medium text-slate-600">Nama perangkat</label>
														<div class="flex flex-col gap-2 sm:flex-row">
															<Input
																id={`account-session-label-${session.id}`}
																bind:value={labelDrafts[session.id]}
																maxlength={60}
																placeholder="Mis. Laptop Ruang Guru"
															/>
															<LoadingButton
																variant="secondary"
																size="sm"
																onclick={() => void renameSession(session.id)}
																loading={renameSessionLoading === session.id}
																loadingLabel="Menyimpan..."
															>
																<SaveIcon class="size-3.5" />
																Simpan
															</LoadingButton>
														</div>
													</div>
												</div>
												<LoadingButton
													variant="outline"
													size="sm"
													onclick={() => void revokeSession(session.id)}
													loading={revokeSessionLoading === session.id}
													loadingLabel="Memproses..."
												>
													<Trash2Icon class="size-3.5" />
													Akhiri Sesi
												</LoadingButton>
											</div>
										</div>
									{/each}
								</div>
							{/if}
						</Card.Content>
					</Card.Root>
				</div>

				<div class="space-y-6">
					<Card.Root>
						<Card.Header class="pb-3">
							<Card.Title class="text-base">Preferensi Tampilan</Card.Title>
							<Card.Description>Preferensi yang tersedia saat ini mengikuti pengaturan sidebar.</Card.Description>
						</Card.Header>
						<Card.Content class="space-y-3">
							<div class="flex items-center justify-between rounded-lg border border-slate-200 px-4 py-3">
								<span class="text-sm text-slate-600">Akses cepat</span>
								<span class="text-sm font-semibold text-slate-900">{preferences ? `${pinnedCount} item` : 'Belum tersedia'}</span>
							</div>
							<div class="flex items-center justify-between rounded-lg border border-slate-200 px-4 py-3">
								<span class="text-sm text-slate-600">Riwayat navigasi</span>
								<span class="text-sm font-semibold text-slate-900">{preferences ? `${recentCount} item` : 'Belum tersedia'}</span>
							</div>
							<div class="flex items-center justify-between rounded-lg border border-slate-200 px-4 py-3">
								<span class="text-sm text-slate-600">Preferensi lain</span>
								<span class="text-sm font-semibold text-slate-500">Belum tersedia</span>
							</div>
						</Card.Content>
					</Card.Root>

					<Card.Root>
						<Card.Header class="pb-3">
							<Card.Title class="text-base">Keluar dari Semua Perangkat</Card.Title>
							<Card.Description>Gunakan saat akun perlu dikunci ulang dari seluruh perangkat.</Card.Description>
						</Card.Header>
						<Card.Content>
							<LoadingButton variant="destructive" onclick={() => void logoutAllSessions()} loading={logoutAllLoading} loadingLabel="Memproses...">
								<LogOutIcon class="size-4" />
								Keluar Semua Sesi
							</LoadingButton>
						</Card.Content>
					</Card.Root>
				</div>
			</div>
			{:else}
				<RecoveryPanel title="Akun Belum Tersaji" message="Data akun belum tersedia. Coba muat ulang halaman." onRetry={() => void refreshOverview(true)} />
			{/if}
		{/snippet}
	</AsyncContent>
</div>
