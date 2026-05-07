<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import FilePenLineIcon from '@lucide/svelte/icons/file-pen-line';
	import HistoryIcon from '@lucide/svelte/icons/history';
	import UploadIcon from '@lucide/svelte/icons/upload';
	import KeyRoundIcon from '@lucide/svelte/icons/key-round';
	import LogOutIcon from '@lucide/svelte/icons/log-out';
	import MailIcon from '@lucide/svelte/icons/mail';
	import MapPinIcon from '@lucide/svelte/icons/map-pin';
	import PhoneIcon from '@lucide/svelte/icons/phone';
	import RefreshCcwIcon from '@lucide/svelte/icons/refresh-ccw';
	import SaveIcon from '@lucide/svelte/icons/save';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Textarea } from '$lib/components/ui/textarea';
	import { toast } from '$lib/components/ui/sonner';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import { clearCbtComposerDrafts } from '$lib/client/cbt-drafts';
	import { clientApiPath, readClientApiData, readClientJson } from '$lib/client/api';
	import {
		accountAvatarUrl,
		accountDisplayName,
		accountErrorMessage,
		accountInitials,
		changeRequestCanCancel,
		changeRequestStatusLabel,
		contactFieldEditable,
		formatAccountDateTime,
		hasEditableContact,
		isCurrentSession,
		linkedProfileLabel,
		normalizeAccountChangeHistory,
		normalizeChangeRequests,
		normalizeAccountSessions,
		officialChangeFieldOptions,
		officialFieldLabel,
		preferenceItemCount,
		profileHistoryActionLabel,
		profileHistoryFieldLabel,
		profileHistoryStatusLabel,
		profileTypeLabel,
		roleLabel,
		sessionTitle,
		type AccountChangeHistoryItem,
		type AccountChangeRequest,
		type AccountIdentity,
		type AuthSession,
		type OfficialChangeFieldOption,
		type SidebarPreferences
	} from '$lib/client/account';
	import { confirmAction } from '$lib/confirm-dialog';

	type AccountOverview = {
		account: AccountIdentity;
		sessions: AuthSession[];
		preferences: SidebarPreferences | null;
		changeFields: OfficialChangeFieldOption[];
		changeRequests: AccountChangeRequest[];
		changeHistory: AccountChangeHistoryItem[];
	};

	let overviewPromise = $state<Promise<AccountOverview> | null>(null);
	let account = $state<AccountIdentity | null>(null);
	let sessions = $state<AuthSession[]>([]);
	let preferences = $state<SidebarPreferences | null>(null);
	let changeFields = $state<OfficialChangeFieldOption[]>([]);
	let changeRequests = $state<AccountChangeRequest[]>([]);
	let changeHistory = $state<AccountChangeHistoryItem[]>([]);
	let labelDrafts = $state<Record<string, string>>({});
	let contactForm = $state({ phone: '', email: '', address: '' });
	let changeRequestForm = $state({ field_key: '', requested_value: '', reason: '' });
	let changeRequestLoading = $state(false);
	let cancelChangeRequestLoading = $state<string | null>(null);
	let contactLoading = $state(false);
	let avatarFile = $state<File | null>(null);
	let avatarInput = $state<HTMLInputElement | null>(null);
	let avatarUploadLoading = $state(false);
	let avatarDeleteLoading = $state(false);
	let pwForm = $state({ current: '', next: '', confirm: '' });
	let pwLoading = $state(false);
	let logoutAllLoading = $state(false);
	let revokeSessionLoading = $state<string | null>(null);
	let renameSessionLoading = $state<string | null>(null);
	let refreshLoading = $state(false);

	const currentSessionId = $derived(page.data.user?.session_id ?? '');
	const requiresPasswordChange = $derived(Boolean(page.data.user?.must_change_password || account?.must_change_password));
	const pinnedCount = $derived(preferenceItemCount(preferences?.pinned_items));
	const recentCount = $derived(preferenceItemCount(preferences?.recent_items));
	const canEditContact = $derived(hasEditableContact(account?.contact));
	const officialChangeFields = $derived(officialChangeFieldOptions(changeFields));
	const canRequestOfficialChange = $derived(officialChangeFields.length > 0);
	const selectedChangeField = $derived(officialChangeFields.find((field) => field.field_key === changeRequestForm.field_key));
	const changeRequestValueType = $derived(selectedChangeField?.value_type === 'date' ? 'date' : 'text');
	const currentAvatarUrl = $derived(accountAvatarUrl(account));
	const currentAvatarInitials = $derived(accountInitials(account));
	const selectedAvatarLabel = $derived(avatarFile ? `${avatarFile.name} (${formatAvatarFileSize(avatarFile.size)})` : '');

	const infoPanelClass = 'rounded-lg border border-border bg-card/70 p-4';
	const compactPanelClass = 'rounded-lg border border-border bg-muted/20 px-4 py-3 text-sm text-muted-foreground';
	const emptyPanelClass = 'rounded-lg border border-dashed border-border bg-muted/20 px-4 py-4 text-sm text-muted-foreground';
	const labelClass = 'text-xs font-medium uppercase tracking-[0.16em] text-muted-foreground';
	const valueClass = 'mt-2 text-sm font-semibold text-foreground';
	const noteClass = 'mt-2 rounded-md bg-muted/50 px-3 py-2 text-xs text-muted-foreground';
	const chevronClass = 'mx-2 text-muted-foreground/50';

	function applyOverview(overview: AccountOverview) {
		account = overview.account;
		sessions = overview.sessions;
		preferences = overview.preferences;
		changeFields = overview.changeFields;
		changeRequests = overview.changeRequests;
		changeHistory = overview.changeHistory;
		contactForm = contactFormFromAccount(overview.account);
		resetChangeRequestForm(overview.changeFields);
		labelDrafts = Object.fromEntries(
			overview.sessions.map((session) => [session.id, session.device_label?.trim() ?? ''])
		);
	}

	function contactFormFromAccount(identity: AccountIdentity) {
		return {
			phone: identity.contact?.phone ?? '',
			email: identity.contact?.email ?? '',
			address: identity.contact?.address ?? ''
		};
	}

	function resetChangeRequestForm(fields: OfficialChangeFieldOption[] = changeFields) {
		const options = officialChangeFieldOptions(fields);
		changeRequestForm = { field_key: options[0]?.field_key ?? '', requested_value: '', reason: '' };
	}

	function currentOverview(): AccountOverview | null {
		if (!account) return null;
		return { account, sessions, preferences, changeFields, changeRequests, changeHistory };
	}

	async function fetchPreferences() {
		try {
			const res = await fetch('/api/auth/preferences/sidebar');
			return await readClientApiData<SidebarPreferences>(res, 'Gagal memuat preferensi tampilan');
		} catch {
			return null;
		}
	}

	async function fetchChangeHistory() {
		const res = await fetch('/api/auth/account/change-history?per_page=30');
		const rows = await readClientApiData<AccountChangeHistoryItem[]>(res, 'Gagal memuat riwayat perubahan profil');
		return normalizeAccountChangeHistory(rows);
	}

	async function fetchChangeRequestFields() {
		const res = await fetch('/api/auth/account/change-request-fields');
		const rows = await readClientApiData<OfficialChangeFieldOption[]>(res, 'Gagal memuat daftar field perubahan data');
		return officialChangeFieldOptions(rows);
	}

	async function fetchOverview(): Promise<AccountOverview> {
		const accountData = await fetch('/api/auth/account').then((response) =>
			readClientApiData<AccountIdentity>(response, 'Gagal memuat identitas akun')
		);
		if (page.data.user?.must_change_password || accountData.must_change_password) {
			const sessionData = await fetch('/api/auth/sessions').then((response) =>
				readClientApiData<AuthSession[]>(response, 'Gagal memuat sesi aktif')
			);
			return {
				account: accountData,
				sessions: normalizeAccountSessions(sessionData),
				preferences: null,
				changeFields: [],
				changeRequests: [],
				changeHistory: []
			};
		}

		const [sessionData, preferenceData, changeFieldData, changeRequestData, changeHistoryData] = await Promise.all([
			fetch('/api/auth/sessions').then((response) =>
				readClientApiData<AuthSession[]>(response, 'Gagal memuat sesi aktif')
			),
			fetchPreferences(),
			fetchChangeRequestFields(),
			fetch('/api/auth/account/change-requests').then((response) =>
				readClientApiData<AccountChangeRequest[]>(response, 'Gagal memuat permintaan perubahan data')
			),
			fetchChangeHistory()
		]);

		return {
			account: accountData,
			sessions: normalizeAccountSessions(sessionData),
			preferences: preferenceData,
			changeFields: changeFieldData,
			changeRequests: normalizeChangeRequests(changeRequestData),
			changeHistory: changeHistoryData
		};
	}

	async function refreshChangeHistory() {
		try {
			changeHistory = await fetchChangeHistory();
			const current = currentOverview();
			if (current) overviewPromise = Promise.resolve(current);
		} catch {
			// Keep the visible account state stable; manual refresh can retry the history card.
		}
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

	async function saveContact() {
		if (!account?.contact || !hasEditableContact(account.contact)) {
			toast.error('Kontak pribadi belum tersedia untuk akun ini');
			return;
		}

		const payload: Record<string, string> = {};
		if (contactFieldEditable(account.contact, 'phone')) payload.phone = contactForm.phone;
		if (contactFieldEditable(account.contact, 'email')) payload.email = contactForm.email;
		if (contactFieldEditable(account.contact, 'address')) payload.address = contactForm.address;

		contactLoading = true;
		try {
			const res = await fetch('/api/auth/account', {
				method: 'PATCH',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify(payload)
			});
			const updated = await readClientApiData<AccountIdentity>(res, 'Gagal menyimpan kontak pribadi');
			account = updated;
			contactForm = contactFormFromAccount(updated);
			const current = currentOverview();
			if (current) overviewPromise = Promise.resolve(current);
			void refreshChangeHistory();
			toast.success('Kontak pribadi berhasil disimpan.');
		} catch (error) {
			toast.error(accountErrorMessage(error, 'Gagal menyimpan kontak pribadi'));
		} finally {
			contactLoading = false;
		}
	}

	async function submitChangeRequest() {
		if (!account || !canRequestOfficialChange) {
			toast.error('Akun belum tertaut ke profil resmi yang dapat diajukan perubahannya');
			return;
		}
		const selectedField = officialChangeFields.find((field) => field.field_key === changeRequestForm.field_key);
		if (!selectedField || !changeRequestForm.requested_value.trim() || !changeRequestForm.reason.trim()) {
			toast.error('Field resmi, nilai baru, dan alasan wajib diisi');
			return;
		}

		changeRequestLoading = true;
		try {
			const res = await fetch('/api/auth/account/change-requests', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({
					profile_type: selectedField.profile_type,
					field_key: selectedField.field_key,
					requested_value: changeRequestForm.requested_value,
					reason: changeRequestForm.reason
				})
			});
			const created = await readClientApiData<AccountChangeRequest>(res, 'Gagal mengirim permintaan perubahan data');
			changeRequests = [created, ...changeRequests];
			const current = currentOverview();
			if (current) overviewPromise = Promise.resolve(current);
			resetChangeRequestForm();
			void refreshChangeHistory();
			toast.success('Permintaan perubahan data resmi dikirim.');
		} catch (error) {
			toast.error(accountErrorMessage(error, 'Gagal mengirim permintaan perubahan data'));
		} finally {
			changeRequestLoading = false;
		}
	}

	async function cancelChangeRequest(request: AccountChangeRequest) {
		if (!changeRequestCanCancel(request)) return;
		if (!(await confirmAction({
			title: 'Batalkan Permintaan',
			message: `Batalkan permintaan perubahan ${officialFieldLabel(request.field_key)}?`,
			confirmLabel: 'Batalkan Permintaan',
			tone: 'warning'
		}))) return;

		cancelChangeRequestLoading = request.id;
		try {
			const res = await fetch(clientApiPath`/api/auth/account/change-requests/${request.id}/cancel`, {
				method: 'POST'
			});
			const updated = await readClientApiData<AccountChangeRequest>(res, 'Gagal membatalkan permintaan');
			changeRequests = changeRequests.map((item) => (item.id === updated.id ? updated : item));
			const current = currentOverview();
			if (current) overviewPromise = Promise.resolve(current);
			void refreshChangeHistory();
			toast.success('Permintaan perubahan data dibatalkan.');
		} catch (error) {
			toast.error(accountErrorMessage(error, 'Gagal membatalkan permintaan'));
		} finally {
			cancelChangeRequestLoading = null;
		}
	}

	function changeRequestBadgeVariant(status: string) {
		if (status === 'approved') return 'secondary';
		if (status === 'rejected') return 'destructive';
		return 'outline';
	}

	function profileHistoryBadgeVariant(status: string) {
		if (status === 'completed' || status === 'approved') return 'secondary';
		if (status === 'rejected') return 'destructive';
		return 'outline';
	}

	function handleAvatarFileChange(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		avatarFile = input.files?.[0] ?? null;
	}

	function formatAvatarFileSize(size: number) {
		if (size >= 1024 * 1024) return `${(size / (1024 * 1024)).toFixed(1)} MB`;
		return `${Math.max(1, Math.round(size / 1024))} KB`;
	}

	function validateAvatarFile(file: File) {
		const allowedTypes = new Set(['image/jpeg', 'image/png', 'image/webp']);
		if (!allowedTypes.has(file.type)) {
			toast.error('Foto profil harus berupa JPG, PNG, atau WebP.');
			return false;
		}
		if (file.size > 2 * 1024 * 1024) {
			toast.error('Ukuran foto profil maksimal 2 MB.');
			return false;
		}
		return true;
	}

	function clearAvatarInput() {
		avatarFile = null;
		if (avatarInput) avatarInput.value = '';
	}

	async function uploadAvatar() {
		if (!avatarFile) {
			toast.error('Pilih foto profil terlebih dahulu.');
			return;
		}
		if (!validateAvatarFile(avatarFile)) return;

		const form = new FormData();
		form.set('file', avatarFile);
		avatarUploadLoading = true;
		try {
			const res = await fetch('/api/auth/account/avatar', {
				method: 'POST',
				body: form
			});
			const updated = await readClientApiData<AccountIdentity>(res, 'Gagal mengunggah foto profil');
			account = updated;
			contactForm = contactFormFromAccount(updated);
			clearAvatarInput();
			const current = currentOverview();
			if (current) overviewPromise = Promise.resolve(current);
			void refreshChangeHistory();
			toast.success('Foto profil berhasil diperbarui.');
		} catch (error) {
			toast.error(accountErrorMessage(error, 'Gagal mengunggah foto profil'));
		} finally {
			avatarUploadLoading = false;
		}
	}

	async function deleteAvatar() {
		if (!currentAvatarUrl) {
			toast.error('Foto profil belum tersedia.');
			return;
		}
		if (!(await confirmAction({
			title: 'Hapus Foto Profil',
			message: 'Foto profil akun Anda akan dihapus dari profil tertaut.',
			confirmLabel: 'Hapus Foto',
			tone: 'warning'
		}))) return;

		avatarDeleteLoading = true;
		try {
			const res = await fetch('/api/auth/account/avatar', { method: 'DELETE' });
			const updated = await readClientApiData<AccountIdentity>(res, 'Gagal menghapus foto profil');
			account = updated;
			contactForm = contactFormFromAccount(updated);
			clearAvatarInput();
			const current = currentOverview();
			if (current) overviewPromise = Promise.resolve(current);
			void refreshChangeHistory();
			toast.success('Foto profil berhasil dihapus.');
		} catch (error) {
			toast.error(accountErrorMessage(error, 'Gagal menghapus foto profil'));
		} finally {
			avatarDeleteLoading = false;
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
			<h1 class="text-2xl font-semibold text-foreground">Akun Saya</h1>
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
								<div class={infoPanelClass}>
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
							<div class={infoPanelClass}>
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
			{#if requiresPasswordChange}
			<div class="mx-auto max-w-3xl space-y-6">
				<Card.Root class="border-warning/30 bg-warning/10">
					<Card.Header class="pb-3">
						<Card.Title class="text-base">Ganti Password Pertama</Card.Title>
						<Card.Description>Akun dengan password sementara hanya dapat membuka halaman ini dan logout sampai password diganti.</Card.Description>
					</Card.Header>
					<Card.Content>
						<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
							<div class={infoPanelClass}>
								<p class={labelClass}>Username</p>
								<p class={valueClass}>{account.username}</p>
							</div>
							<div class={infoPanelClass}>
								<p class={labelClass}>Profil tertaut</p>
								<p class={valueClass}>{linkedProfileLabel(account)}</p>
								<p class="mt-1 text-xs text-muted-foreground">{profileTypeLabel(account.profile_type)}</p>
							</div>
						</div>
					</Card.Content>
				</Card.Root>

				<Card.Root>
					<Card.Header class="pb-3">
						<Card.Title class="text-base">Ubah Password</Card.Title>
						<Card.Description>Gunakan password baru minimal 8 karakter. Setelah tersimpan, login ulang untuk membuka portal.</Card.Description>
					</Card.Header>
					<Card.Content>
						<div class="grid grid-cols-1 gap-4 md:grid-cols-3">
							<div>
								<label for="account-pw-current-required" class="mb-1.5 block text-sm font-medium">Password Saat Ini</label>
								<Input id="account-pw-current-required" type="password" bind:value={pwForm.current} autocomplete="current-password" />
							</div>
							<div>
								<label for="account-pw-next-required" class="mb-1.5 block text-sm font-medium">Password Baru</label>
								<Input id="account-pw-next-required" type="password" bind:value={pwForm.next} autocomplete="new-password" />
							</div>
							<div>
								<label for="account-pw-confirm-required" class="mb-1.5 block text-sm font-medium">Konfirmasi Password</label>
								<Input id="account-pw-confirm-required" type="password" bind:value={pwForm.confirm} autocomplete="new-password" />
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
						<Card.Title class="text-base">Logout</Card.Title>
						<Card.Description>Keluar dari sesi saat ini jika perlu bantuan operator sebelum mengganti password.</Card.Description>
					</Card.Header>
					<Card.Content>
						<LoadingButton variant="outline" onclick={async () => {
							await fetch('/api/auth/logout', { method: 'POST' }).catch(() => undefined);
							clearCbtComposerDrafts();
							window.location.href = '/login';
						}}>
							<LogOutIcon class="size-4" />
							Logout
						</LoadingButton>
					</Card.Content>
				</Card.Root>
			</div>
			{:else}
			<div class="grid grid-cols-1 gap-6 xl:grid-cols-[minmax(0,1fr)_360px]">
				<div class="space-y-6">
					<Card.Root>
						<Card.Header class="pb-3">
							<Card.Title class="text-base">Identitas Login</Card.Title>
							<Card.Description>Data resmi tetap dikelola oleh modul master terkait.</Card.Description>
						</Card.Header>
						<Card.Content>
							<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
								<div class={infoPanelClass}>
									<p class={labelClass}>Nama tampil</p>
									<p class={valueClass}>{accountDisplayName(account)}</p>
								</div>
								<div class={infoPanelClass}>
									<p class={labelClass}>Username</p>
									<p class={valueClass}>{account.username}</p>
								</div>
								<div class={infoPanelClass}>
									<p class={labelClass}>Role</p>
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
								<div class={infoPanelClass}>
									<p class={labelClass}>Profil tertaut</p>
									<p class={valueClass}>{linkedProfileLabel(account)}</p>
									<p class="mt-1 text-xs text-muted-foreground">{profileTypeLabel(account.profile_type)}</p>
								</div>
								<div class={infoPanelClass}>
									<p class={labelClass}>Login terakhir</p>
									<p class={valueClass}>{formatAccountDateTime(account.last_login_at)}</p>
								</div>
								<div class={infoPanelClass}>
									<p class={labelClass}>Status akun</p>
									<p class="mt-2 text-sm font-semibold {account.is_active === false ? 'text-destructive' : 'text-primary'}">
										{account.is_active === false ? 'Nonaktif' : 'Aktif'}
									</p>
								</div>
							</div>
						</Card.Content>
					</Card.Root>

					<Card.Root>
						<Card.Header class="pb-3">
							<Card.Title class="text-base">Permintaan Perubahan Data Resmi</Card.Title>
							<Card.Description>Data resmi tertentu dikoreksi melalui persetujuan admin.</Card.Description>
						</Card.Header>
						<Card.Content class="space-y-5">
							{#if canRequestOfficialChange}
								<div class="grid grid-cols-1 gap-4 md:grid-cols-[220px_minmax(0,1fr)]">
									<div>
										<label for="account-change-field" class="mb-1.5 block text-sm font-medium">Field Resmi</label>
										<select
											id="account-change-field"
											class="h-10 w-full rounded-md border border-input bg-background px-3 text-sm text-foreground"
											bind:value={changeRequestForm.field_key}
										>
											{#each officialChangeFields as field (field.field_key)}
												<option value={field.field_key}>{field.label}</option>
											{/each}
										</select>
									</div>
									<div>
										<label for="account-change-value" class="mb-1.5 block text-sm font-medium">Nilai Baru</label>
										<Input
											id="account-change-value"
											type={changeRequestValueType}
											bind:value={changeRequestForm.requested_value}
											maxlength={200}
											placeholder={changeRequestForm.field_key === 'tanggal_lahir' ? 'YYYY-MM-DD' : 'Tulis data resmi yang benar'}
										/>
									</div>
									<div class="md:col-span-2">
										<label for="account-change-reason" class="mb-1.5 block text-sm font-medium">Alasan / Rujukan Dokumen</label>
										<Textarea id="account-change-reason" rows={3} maxlength={1000} bind:value={changeRequestForm.reason} />
									</div>
								</div>
								<LoadingButton
									onclick={() => void submitChangeRequest()}
									loading={changeRequestLoading}
									loadingLabel="Mengirim..."
									disabled={!changeRequestForm.field_key || !changeRequestForm.requested_value.trim() || !changeRequestForm.reason.trim()}
								>
									<FilePenLineIcon class="size-4" />
									Kirim Permintaan
								</LoadingButton>
							{:else}
								<div class={compactPanelClass}>
									Akun tanpa profil pegawai, siswa, atau orang tua belum dapat mengajukan perubahan data resmi.
								</div>
							{/if}

							<div class="space-y-3">
								<div class="flex items-center justify-between gap-3">
									<p class="text-sm font-semibold text-foreground">Riwayat Permintaan</p>
									<Badge variant="outline">{changeRequests.length} permintaan</Badge>
								</div>
								{#if changeRequests.length === 0}
									<p class={emptyPanelClass}>Belum ada permintaan perubahan data resmi.</p>
								{:else}
									<div class="space-y-2">
										{#each changeRequests as request (request.id)}
											<div class={infoPanelClass}>
												<div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
													<div class="min-w-0 flex-1">
														<div class="flex flex-wrap items-center gap-2">
															<p class="text-sm font-semibold text-foreground">{officialFieldLabel(request.field_key)}</p>
															<Badge variant={changeRequestBadgeVariant(request.status)}>{changeRequestStatusLabel(request.status)}</Badge>
														</div>
														<p class="mt-2 text-sm text-muted-foreground">
															<span class="font-medium">Saat ini:</span> {request.current_value || '—'}
															<span class={chevronClass}>→</span>
															<span class="font-medium">Usulan:</span> {request.requested_value || '—'}
														</p>
														<p class="mt-1 text-xs text-muted-foreground">{request.reason}</p>
														{#if request.review_note}
															<p class={noteClass}>Catatan review: {request.review_note}</p>
														{/if}
														<p class="mt-2 text-[11px] text-muted-foreground">Diajukan: {formatAccountDateTime(request.created_at)}</p>
													</div>
													{#if changeRequestCanCancel(request)}
														<LoadingButton
															variant="outline"
															size="sm"
															onclick={() => void cancelChangeRequest(request)}
															loading={cancelChangeRequestLoading === request.id}
															loadingLabel="Membatalkan..."
														>
															Batalkan
														</LoadingButton>
													{/if}
												</div>
											</div>
										{/each}
									</div>
								{/if}
							</div>
						</Card.Content>
					</Card.Root>

					<Card.Root>
						<Card.Header class="pb-3">
							<div class="flex items-start justify-between gap-3">
								<div>
									<Card.Title class="flex items-center gap-2 text-base">
										<HistoryIcon class="size-4 text-muted-foreground" />
										Riwayat Perubahan Profil
									</Card.Title>
									<Card.Description>Aktivitas kontak, foto profil, dan permintaan perubahan data resmi.</Card.Description>
								</div>
								<Badge variant="outline">{changeHistory.length} aktivitas</Badge>
							</div>
						</Card.Header>
						<Card.Content>
							{#if changeHistory.length === 0}
								<p class={emptyPanelClass}>Belum ada riwayat perubahan profil.</p>
							{:else}
								<div class="space-y-2">
									{#each changeHistory as item, index (`${item.created_at}-${item.action}-${item.field_key}-${index}`)}
										<div class={infoPanelClass}>
											<div class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
												<div class="min-w-0">
													<div class="flex flex-wrap items-center gap-2">
														<p class="text-sm font-semibold text-foreground">{profileHistoryActionLabel(item.action)}</p>
														<Badge variant={profileHistoryBadgeVariant(item.status)}>{profileHistoryStatusLabel(item.status)}</Badge>
													</div>
													<p class="mt-1 text-sm text-muted-foreground">{profileHistoryFieldLabel(item.field_key)}</p>
													{#if item.reviewer_username}
														<p class="mt-1 text-xs text-muted-foreground">Reviewer: {item.reviewer_username}</p>
													{/if}
													{#if item.review_note}
														<p class={noteClass}>Catatan review: {item.review_note}</p>
													{/if}
												</div>
												<p class="shrink-0 text-xs text-muted-foreground">{formatAccountDateTime(item.created_at)}</p>
											</div>
										</div>
									{/each}
								</div>
							{/if}
						</Card.Content>
					</Card.Root>

					<Card.Root>
						<Card.Header class="pb-3">
							<Card.Title class="text-base">Kontak Pribadi</Card.Title>
							<Card.Description>Kontak profil tertaut yang aman diperbarui sendiri.</Card.Description>
						</Card.Header>
						<Card.Content>
							{#if canEditContact}
								<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
									{#if contactFieldEditable(account.contact, 'phone')}
										<div>
											<label for="account-contact-phone" class="mb-1.5 flex items-center gap-1.5 text-sm font-medium">
												<PhoneIcon class="size-4 text-muted-foreground" />
												Nomor HP/WA
											</label>
											<Input id="account-contact-phone" bind:value={contactForm.phone} autocomplete="tel" maxlength={40} placeholder="08xxxxxxxxxx" />
										</div>
									{/if}
									{#if contactFieldEditable(account.contact, 'email')}
										<div>
											<label for="account-contact-email" class="mb-1.5 flex items-center gap-1.5 text-sm font-medium">
												<MailIcon class="size-4 text-muted-foreground" />
												Email
											</label>
											<Input id="account-contact-email" type="email" bind:value={contactForm.email} autocomplete="email" maxlength={254} placeholder="nama@example.id" />
										</div>
									{/if}
									{#if contactFieldEditable(account.contact, 'address')}
										<div class="md:col-span-2">
											<label for="account-contact-address" class="mb-1.5 flex items-center gap-1.5 text-sm font-medium">
												<MapPinIcon class="size-4 text-muted-foreground" />
												Alamat Kontak
											</label>
											<Textarea id="account-contact-address" bind:value={contactForm.address} rows={3} maxlength={500} />
										</div>
									{/if}
								</div>
								<LoadingButton class="mt-4" onclick={() => void saveContact()} loading={contactLoading} loadingLabel="Menyimpan...">
									<SaveIcon class="size-4" />
									Simpan Kontak
								</LoadingButton>
							{:else}
								<div class={compactPanelClass}>
									Kontak pribadi belum tersedia untuk akun tanpa profil tertaut.
								</div>
							{/if}
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
										<div class={infoPanelClass}>
											<div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
												<div class="min-w-0 flex-1 space-y-2">
													<div class="flex flex-wrap items-center gap-2">
														<p class="text-sm font-semibold text-foreground">{sessionTitle(session)}</p>
														{#if isCurrentSession(session, currentSessionId)}
															<Badge variant="secondary">Perangkat Ini</Badge>
														{/if}
													</div>
													<div class="grid grid-cols-1 gap-2 text-xs text-muted-foreground sm:grid-cols-2">
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
														<p class="line-clamp-2 text-[11px] text-muted-foreground">{session.user_agent}</p>
													{/if}
													<div class="pt-2">
														<label for={`account-session-label-${session.id}`} class="mb-1 block text-xs font-medium text-muted-foreground">Nama perangkat</label>
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
							<Card.Title class="text-base">Foto Profil</Card.Title>
							<Card.Description>JPG, PNG, atau WebP maksimal 2 MB.</Card.Description>
						</Card.Header>
						<Card.Content class="space-y-4">
							<div class="flex items-center gap-4">
								{#if currentAvatarUrl}
									<img src={currentAvatarUrl} alt={`Foto profil ${accountDisplayName(account)}`} class="size-20 rounded-lg border border-border object-cover" />
								{:else}
									<div class="flex size-20 items-center justify-center rounded-lg border border-primary/25 bg-primary/10 text-xl font-semibold text-primary">
										{currentAvatarInitials}
									</div>
								{/if}
								<div class="min-w-0">
									<p class="truncate text-sm font-semibold text-foreground">{accountDisplayName(account)}</p>
									<p class="mt-1 text-xs text-muted-foreground">{currentAvatarUrl ? 'Foto profil aktif' : 'Belum ada foto profil'}</p>
								</div>
							</div>

							<div>
								<label for="account-avatar-file" class="mb-1.5 block text-sm font-medium">Pilih Foto</label>
								<Input
									id="account-avatar-file"
									type="file"
									accept="image/jpeg,image/png,image/webp"
									bind:ref={avatarInput}
									onchange={handleAvatarFileChange}
								/>
								{#if selectedAvatarLabel}
									<p class="mt-1 text-xs text-muted-foreground">{selectedAvatarLabel}</p>
								{/if}
							</div>

							<div class="flex flex-col gap-2 sm:flex-row">
								<LoadingButton
									variant="secondary"
									onclick={() => void uploadAvatar()}
									loading={avatarUploadLoading}
									loadingLabel="Mengunggah..."
									disabled={!avatarFile || avatarDeleteLoading}
								>
									<UploadIcon class="size-4" />
									Unggah Foto
								</LoadingButton>
								<LoadingButton
									variant="outline"
									onclick={() => void deleteAvatar()}
									loading={avatarDeleteLoading}
									loadingLabel="Menghapus..."
									disabled={!currentAvatarUrl || avatarUploadLoading}
								>
									<Trash2Icon class="size-4" />
									Hapus Foto
								</LoadingButton>
							</div>
							<p class="text-xs text-muted-foreground">Foto tersimpan di storage lokal backend dan hanya mengubah profil yang tertaut ke akun ini.</p>
						</Card.Content>
					</Card.Root>

					<Card.Root>
						<Card.Header class="pb-3">
							<Card.Title class="text-base">Preferensi Tampilan</Card.Title>
							<Card.Description>Tema aplikasi dan personalisasi sidebar tersimpan di browser dan akun.</Card.Description>
						</Card.Header>
						<Card.Content class="space-y-3">
							<div class={`flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between ${compactPanelClass}`}>
								<div>
									<span class="text-sm font-medium text-foreground">Tema aplikasi</span>
									<p class="mt-1 text-xs text-muted-foreground">Ikuti sistem, terang, atau gelap.</p>
								</div>
								<ThemeToggle />
							</div>
							<div class={`flex items-center justify-between ${compactPanelClass}`}>
								<span class="text-sm text-muted-foreground">Akses cepat</span>
								<span class="text-sm font-semibold text-foreground">{preferences ? `${pinnedCount} item` : 'Belum tersedia'}</span>
							</div>
							<div class={`flex items-center justify-between ${compactPanelClass}`}>
								<span class="text-sm text-muted-foreground">Riwayat navigasi</span>
								<span class="text-sm font-semibold text-foreground">{preferences ? `${recentCount} item` : 'Belum tersedia'}</span>
							</div>
							<div class={`flex items-center justify-between ${compactPanelClass}`}>
								<span class="text-sm text-muted-foreground">Preferensi lain</span>
								<span class="text-sm font-semibold text-muted-foreground">Belum tersedia</span>
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
			{/if}
			{:else}
				<RecoveryPanel title="Akun Belum Tersaji" message="Data akun belum tersedia. Coba muat ulang halaman." onRetry={() => void refreshOverview(true)} />
			{/if}
		{/snippet}
	</AsyncContent>
</div>
