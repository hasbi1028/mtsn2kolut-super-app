<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import PasswordInput from '$lib/components/PasswordInput.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { confirmAction } from '$lib/confirm-dialog';
	import { readClientApiData, readClientJson } from '$lib/client/api';
	import { trackInternalAnalyticsEvent } from '$lib/analytics/internal-analytics';
	import {
		employeeAccountGenerationCSV,
		fetchRBACMatrix,
		fetchUserProfileCandidates,
		generateEmployeeAccounts,
		previewEmployeeAccountGeneration,
		resetUserPassword,
		updateUserProfileLink,
		updateUserRoles,
		type RBACMatrix,
		type RBACRole,
		type EmployeeAccountGenerationResult,
		type UserProfileCandidate
	} from '$lib/client/rbac-users';

	type User = {
		id: string; username: string; display_name?: string | null; roles: string[];
		employee_id: string | null; student_id: string | null; parent_id: string | null;
		profile_nama: string | null;
		is_active: boolean;
		last_login_at?: string | null;
		deleted_at?: string | null;
		created_at: string;
	};
	type Rombel = { id: string; name: string; code?: string | null; is_active?: boolean; total_students?: number; };
	type CandidateMode = 'employee' | 'student' | 'parent' | 'none';

	type UsersOverview = {
		users: User[];
		rombels: Rombel[];
		rbac: RBACMatrix;
	};

	let users = $state<User[]>([]);
	let rombels = $state<Rombel[]>([]);
	let rbac = $state<RBACMatrix>({ roles: [], permissions: [], role_permissions: {} });
	let usersPromise = $state<Promise<UsersOverview> | null>(null);
	let usersRequestId = 0;
	let showForm = $state(false);

	let fUsername = $state('');
	let fPassword = $state('');
	let fDisplayName = $state('');
	let fRoles = $state<string[]>(['guru']);
	let fEmpId = $state('');
	let fStuId = $state('');
	let fParId = $state('');
	let fBusy = $state(false);
	let actionBusy = $state<string | null>(null);
	let employeeGeneration = $state<EmployeeAccountGenerationResult | null>(null);
	let generationBusy = $state<string | null>(null);
	let candidateClassId = $state('');
	let candidateSearch = $state('');
	let includeLinkedCandidates = $state(false);
	let profileCandidates = $state<UserProfileCandidate[]>([]);
	let profileCandidatesBusy = $state(false);
	let profileCandidatesError = $state('');
	let profileDisplayNameAutofill = $state('');
	let profileCandidatesRequestId = 0;
	let formHint = $derived.by(() => {
		if (fRoles.includes('siswa')) return 'Akun siswa wajib ditautkan ke satu profil siswa.';
		if (fRoles.includes('ortu')) return 'Akun orang tua wajib ditautkan ke satu profil orang tua.';
		if (fRoles.includes('guru') || fRoles.includes('staf') || fRoles.includes('kesiswaan')) return 'Akun guru/staf/kesiswaan wajib ditautkan ke satu profil pegawai.';
		return 'Akun admin murni boleh tanpa tautan profil.';
	});


	const candidateMode = $derived.by<CandidateMode>(() => {
		if (fRoles.includes('siswa')) return 'student';
		if (fRoles.includes('ortu')) return 'parent';
		if (fRoles.includes('guru') || fRoles.includes('staf') || fRoles.includes('kesiswaan')) return 'employee';
		return 'none';
	});
	const candidateRole = $derived.by(() => {
		if (candidateMode === 'student') return 'siswa';
		if (candidateMode === 'parent') return 'ortu';
		if (candidateMode === 'employee') return fRoles.includes('kesiswaan') ? 'kesiswaan' : fRoles.includes('staf') ? 'staf' : 'guru';
		return '';
	});
	const candidateRequiresClass = $derived(candidateMode === 'student' || candidateMode === 'parent');
	const candidateCanLoad = $derived(candidateMode !== 'none' && (!candidateRequiresClass || Boolean(candidateClassId)));
	const selectedCandidateId = $derived(candidateMode === 'employee' ? fEmpId : candidateMode === 'student' ? fStuId : candidateMode === 'parent' ? fParId : '');
	const selectedCandidate = $derived.by(() => profileCandidates.find((item) => item.id === selectedCandidateId));
	const profileLinkMissing = $derived(
		(candidateMode === 'employee' && !fEmpId) ||
		(candidateMode === 'student' && !fStuId) ||
		(candidateMode === 'parent' && !fParId)
	);

	const fallbackRoles = [
		{ value: 'admin', label: 'Administrator' },
		{ value: 'guru', label: 'Guru' },
		{ value: 'staf', label: 'Staf' },
		{ value: 'kesiswaan', label: 'Kesiswaan' },
		{ value: 'siswa', label: 'Siswa' },
		{ value: 'ortu', label: 'Orang Tua' },
	];
	const availableRoles = $derived.by(() => {
		const activeRoles = (rbac.roles ?? [])
			.filter((role: RBACRole) => role.is_active !== false)
			.map((role: RBACRole) => ({ value: role.code, label: role.name || role.code }));
		return activeRoles.length ? activeRoles : fallbackRoles;
	});

	function applyOverview(overview: UsersOverview) {
		users = overview.users;
		rombels = overview.rombels;
		rbac = overview.rbac;
		return overview;
	}

	async function fetchOverview(): Promise<UsersOverview> {
		const [usersRes, rombelsRes, nextRBAC] = await Promise.all([
			fetch('/api/users'),
			fetch('/api/academic/rombel'),
			fetchRBACMatrix(),
		]);
		const [nextUsers, nextRombels] = await Promise.all([
			readClientApiData<User[]>(usersRes, 'Gagal memuat data pengguna.'),
			readClientApiData<Rombel[]>(rombelsRes, 'Gagal memuat data kelas.'),
		]);
		return { users: nextUsers ?? [], rombels: nextRombels ?? [], rbac: nextRBAC };
	}

	function load() {
		const requestId = ++usersRequestId;
		users = [];
		rombels = [];
		usersPromise = fetchOverview()
			.then((overview) => {
				if (requestId === usersRequestId) return applyOverview(overview);
				return { users, rombels, rbac };
			})
			.catch((error: unknown) => {
				if (requestId === usersRequestId) throw error;
				return { users, rombels, rbac };
			});
	}

	async function refreshOverview() {
		const requestId = ++usersRequestId;
		try {
			const overview = await fetchOverview();
			if (requestId === usersRequestId) {
				applyOverview(overview);
				usersPromise = Promise.resolve(overview);
			}
		} catch (error) {
			if (requestId !== usersRequestId) return;
			usersPromise = Promise.resolve({ users, rombels, rbac });
			toast.error(overviewErrorMessage(error));
		}
	}

	function retryOverview(reset?: () => void) {
		reset?.();
		load();
	}

	function overviewErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat data pengguna.';
	}

	function handleOverviewRenderError(error: unknown) {
		console.error('Users overview render failed', error);
	}


	function clearProfileSelection() {
		fEmpId = '';
		fStuId = '';
		fParId = '';
	}

	function candidateLabel(candidate: UserProfileCandidate) {
		const suffix = candidate.identifier ? ` (${candidate.identifier})` : '';
		return `${candidate.nama}${suffix}`;
	}

	function candidateDescription(candidate: UserProfileCandidate) {
		if (candidate.profile_type === 'parent' && candidate.children?.length) {
			return `Anak: ${candidate.children.map((child) => child.nama).join(', ')}`;
		}
		return candidate.class_name ? `Kelas ${candidate.class_name}` : candidate.profile_type;
	}

	function selectProfileCandidate(candidate: UserProfileCandidate) {
		clearProfileSelection();
		if (candidate.profile_type === 'employee') fEmpId = candidate.id;
		if (candidate.profile_type === 'student') fStuId = candidate.id;
		if (candidate.profile_type === 'parent') fParId = candidate.id;
		if (!fDisplayName || fDisplayName === profileDisplayNameAutofill) fDisplayName = candidate.nama;
		profileDisplayNameAutofill = candidate.nama;
	}

	async function loadProfileCandidates() {
		const requestId = ++profileCandidatesRequestId;
		const role = candidateRole;
		const classId = candidateClassId;
		const query = candidateSearch;
		const includeLinked = includeLinkedCandidates;
		const requiresClass = candidateRequiresClass;
		if (!candidateCanLoad) {
			profileCandidates = [];
			profileCandidatesError = requiresClass ? 'Pilih kelas terlebih dahulu.' : '';
			return;
		}
		profileCandidatesBusy = true;
		profileCandidatesError = '';
		try {
			const response = await fetchUserProfileCandidates({
				role,
				class_id: requiresClass ? classId : null,
				q: query,
				include_linked: includeLinked,
				limit: 50
			});
			if (requestId !== profileCandidatesRequestId || role !== candidateRole || classId !== candidateClassId || query !== candidateSearch || includeLinked !== includeLinkedCandidates) return;
			profileCandidates = response?.candidates ?? [];
		} catch (error) {
			if (requestId !== profileCandidatesRequestId) return;
			profileCandidates = [];
			profileCandidatesError = overviewErrorMessage(error);
		} finally {
			if (requestId === profileCandidatesRequestId) profileCandidatesBusy = false;
		}
	}

	async function createUser() {
		if (!fUsername || !fPassword || fRoles.length === 0) {
			toast.error('Username, password, dan minimal satu role wajib diisi');
			return;
		}
		fBusy = true;
		try {
			const res = await fetch('/api/users', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					username: fUsername, password: fPassword,
					display_name: fDisplayName || null,
					roles: fRoles,
					employee_id: fEmpId || null,
					student_id: fStuId || null,
					parent_id: fParId || null,
				}),
			});
			await readClientJson<unknown>(res);
			fUsername = ''; fPassword = ''; fDisplayName = ''; profileDisplayNameAutofill = ''; fRoles = ['guru']; clearProfileSelection(); candidateClassId = ''; candidateSearch = ''; profileCandidates = [];
			showForm = false;
			toast.success('Pengguna berhasil dibuat');
			await refreshOverview();
		} catch (error) {
			toast.error(overviewErrorMessage(error));
		} finally { fBusy = false; }
	}

	async function deleteUser(id: string, name: string) {
		if (name === 'admin') {
			toast.error('User admin utama tidak bisa dihapus');
			return;
		}
		if (!(await confirmAction({
			title: 'Hapus Pengguna',
			message: `Hapus pengguna "${name}"?`,
			confirmLabel: 'Hapus Pengguna',
			tone: 'danger'
		}))) return;
		const res = await fetch(`/api/users/${id}`, { method: 'DELETE' });
		try {
			await readClientJson<unknown>(res);
		} catch (error) {
			toast.error(overviewErrorMessage(error));
			return;
		}
		toast.success('Pengguna berhasil dihapus');
		await refreshOverview();
	}

	async function toggleUserStatus(user: User) {
		if (user.username === 'admin' && user.is_active) {
			toast.error('Akun admin utama tidak bisa dinonaktifkan');
			return;
		}
		const next = !user.is_active;
		const actionLabel = next ? 'mengaktifkan' : 'menonaktifkan';
		if (!(await confirmAction({
			title: `${actionLabel} Akun`,
			message: `${actionLabel} akun "${user.username}"?`,
			confirmLabel: actionLabel,
			tone: 'warning'
		}))) return;
		const res = await fetch(`/api/users/${user.id}/status`, {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ is_active: next }),
		});
		try {
			await readClientJson<unknown>(res);
		} catch (error) {
			toast.error(overviewErrorMessage(error));
			return;
		}
		toast.success(next ? 'Akun diaktifkan' : 'Akun dinonaktifkan');
		await refreshOverview();
	}


	function roleLabel(roleCode: string) {
		return availableRoles.find((role) => role.value === roleCode)?.label ?? roleCode;
	}

	function formatDateTime(value?: string | null) {
		if (!value) return 'Belum pernah login';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return 'Belum pernah login';
		return date.toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short' });
	}

	async function updateRolesForUser(user: User, nextRoles: string[]) {
		const roles = Array.from(new Set(nextRoles.filter(Boolean)));
		if (roles.length === 0) {
			toast.error('Minimal satu role wajib dipilih.');
			return;
		}
		if (user.roles.includes('admin') && !roles.includes('admin')) {
			if (!(await confirmAction({
				title: 'Lepas Role Admin',
				message: `Lepas role admin dari "${user.username}"? Backend tetap akan menolak jika ini admin aktif terakhir.`,
				confirmLabel: 'Update Role',
				tone: 'warning'
			}))) return;
		}
		actionBusy = `roles:${user.id}`;
		try {
			await updateUserRoles(user.id, roles);
			toast.success('Role pengguna diperbarui');
			await refreshOverview();
		} catch (error) {
			toast.error(overviewErrorMessage(error));
		} finally {
			actionBusy = null;
		}
	}

	async function toggleExistingUserRole(user: User, role: string) {
		const current = user.roles ?? [];
		const nextRoles = current.includes(role) ? current.filter((value) => value !== role) : [...current, role];
		await updateRolesForUser(user, nextRoles);
	}

	async function resetPasswordForUser(user: User) {
		const password = window.prompt(`Password baru untuk ${user.username} (minimal 8 karakter):`);
		if (password === null) return;
		if (password.trim().length < 8) {
			toast.error('Password minimal 8 karakter.');
			return;
		}
		if (!(await confirmAction({
			title: 'Reset Password',
			message: `Reset password untuk "${user.username}"? Semua session aktif user akan dicabut.`,
			confirmLabel: 'Reset Password',
			tone: 'warning'
		}))) return;
		actionBusy = `password:${user.id}`;
		try {
			await resetUserPassword(user.id, password.trim());
			toast.success('Password berhasil direset dan session user dicabut');
		} catch (error) {
			toast.error(overviewErrorMessage(error));
		} finally {
			actionBusy = null;
		}
	}

	async function updateProfileForUser(user: User) {
		const type = window.prompt('Jenis profil: employee, student, parent, atau kosong untuk melepas tautan', user.employee_id ? 'employee' : user.student_id ? 'student' : user.parent_id ? 'parent' : '');
		if (type === null) return;
		const normalizedType = type.trim().toLowerCase();
		let profileID = '';
		if (normalizedType) {
			profileID = window.prompt('Masukkan UUID profil tujuan:', user.employee_id || user.student_id || user.parent_id || '')?.trim() ?? '';
			if (!profileID) {
				toast.error('UUID profil wajib diisi.');
				return;
			}
		}
		const payload = {
			employee_id: normalizedType === 'employee' ? profileID : null,
			student_id: normalizedType === 'student' ? profileID : null,
			parent_id: normalizedType === 'parent' ? profileID : null
		};
		if (normalizedType && !['employee', 'student', 'parent'].includes(normalizedType)) {
			toast.error('Jenis profil tidak valid. Gunakan employee, student, atau parent.');
			return;
		}
		actionBusy = `profile:${user.id}`;
		try {
			await updateUserProfileLink(user.id, payload);
			toast.success('Tautan profil diperbarui');
			await refreshOverview();
		} catch (error) {
			toast.error(overviewErrorMessage(error));
		} finally {
			actionBusy = null;
		}
	}

	function toggleRole(role: string) {
		if (fRoles.includes(role)) {
			fRoles = fRoles.filter((item) => item !== role);
		} else {
			fRoles = [...fRoles, role];
		}
		clearProfileSelection();
		profileCandidatesRequestId += 1; profileCandidatesBusy = false;
		profileCandidates = [];
		profileCandidatesError = '';
		profileDisplayNameAutofill = '';
	}


	async function previewEmployeeGeneration() {
		generationBusy = 'preview';
		try {
			employeeGeneration = await previewEmployeeAccountGeneration();
			toast.success('Preview generate akun pegawai dimuat');
		} catch (error) {
			toast.error(overviewErrorMessage(error));
		} finally {
			generationBusy = null;
		}
	}

	async function runEmployeeGeneration() {
		if (!(await confirmAction({
			title: 'Generate Akun Pegawai',
			message: 'Buat akun untuk pegawai siap generate? Username dan password awal memakai pola NPSN + 2 digit tahun lahir + nomor urut 3 digit, role default guru.',
			confirmLabel: 'Generate Akun',
			tone: 'warning'
		}))) return;
		generationBusy = 'generate';
		try {
			employeeGeneration = await generateEmployeeAccounts();
			toast.success(`${employeeGeneration.created} akun pegawai berhasil dibuat`);
			await refreshOverview();
		} catch (error) {
			toast.error(overviewErrorMessage(error));
		} finally {
			generationBusy = null;
		}
	}

	function downloadEmployeeGenerationCSV() {
		if (!employeeGeneration) return;
		const csv = employeeAccountGenerationCSV(employeeGeneration);
		const blob = new Blob([csv], { type: 'text/csv;charset=utf-8' });
		const url = URL.createObjectURL(blob);
		const anchor = document.createElement('a');
		anchor.href = url;
		anchor.download = `akun-pegawai-${new Date().toISOString().slice(0, 10)}.csv`;
		anchor.click();
		URL.revokeObjectURL(url);
	}

	onMount(() => {
		void trackInternalAnalyticsEvent('users.list_view', {
			pathname: window.location.pathname,
			metadata: { page_key: 'users' }
		});
		void load();
	});
</script>

<svelte:head><title>Manajemen Pengguna — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6 p-6">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-foreground">Manajemen Pengguna</h1>
			<p class="text-sm text-muted-foreground mt-1">Kelola akun akses sistem dengan RBAC terpadu</p>
		</div>
		<div class="flex flex-wrap gap-2">
			<Button variant="outline" href="/settings/user-change-requests">Permintaan Data Resmi</Button>
			<Button onclick={() => (showForm = !showForm)}>
				{showForm ? 'Batal' : '+ Tambah Pengguna'}
			</Button>
		</div>
	</div>

	<AsyncContent promise={usersPromise} onerror={handleOverviewRenderError}>
		{#snippet pending()}
			<div class="grid gap-3 md:grid-cols-4">
				{#each ['Total Akun', 'Akun Aktif', 'Multi-Role', 'Terhubung Profil'] as label (label)}
					<div class="rounded-2xl border border-border bg-muted/50 px-4 py-4">
						<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-muted-foreground">{label}</p>
						<Skeleton class="mt-3 h-8 w-16" />
						<Skeleton class="mt-2 h-4 w-44" />
					</div>
				{/each}
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Data Pengguna Belum Tersaji"
				message={overviewErrorMessage(error)}
				onRetry={() => retryOverview(reset)}
			/>
		{/snippet}

		{#snippet children(value)}
			{@const overview = value as UsersOverview}
			<div class="grid gap-3 md:grid-cols-4">
				<div class="rounded-2xl border border-primary/20 bg-primary/10 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-primary">Total Akun</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{overview.users.length}</p>
					<p class="text-sm text-muted-foreground">akun yang sudah dapat masuk ke sistem</p>
				</div>
				<div class="rounded-2xl border border-accent bg-accent/60 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-accent-foreground">Akun Aktif</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{overview.users.filter((item) => item.is_active).length}</p>
					<p class="text-sm text-muted-foreground">akun yang saat ini masih aktif digunakan</p>
				</div>
				<div class="rounded-2xl border border-warning/30 bg-warning/10 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-warning">Multi-Role</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{overview.users.filter((item) => (item.roles ?? []).length > 1).length}</p>
					<p class="text-sm text-muted-foreground">akun yang memegang lebih dari satu role</p>
				</div>
				<div class="rounded-2xl border border-accent bg-accent/60 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-accent-foreground">Terhubung Profil</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{overview.users.filter((item) => item.profile_nama).length}</p>
					<p class="text-sm text-muted-foreground">akun yang sudah terkait dengan entitas sekolah</p>
				</div>
			</div>
		{/snippet}
	</AsyncContent>

	{#if showForm}
		<Card.Root>
			<Card.Header class="pb-2"><Card.Title class="text-base">Tambah Akun Baru</Card.Title></Card.Header>
			<Card.Content class="space-y-4">
				<div class="grid gap-4 sm:grid-cols-2">
					<div class="space-y-3">
						<div>
							<label for="u-name" class="text-xs text-muted-foreground mb-1 block">Username</label>
							<Input id="u-name" bind:value={fUsername} placeholder="Gunakan nama akun yang mudah dikenali" />
						</div>
						<div>
							<label for="u-display" class="text-xs text-muted-foreground mb-1 block">Display Name</label>
							<Input id="u-display" bind:value={fDisplayName} placeholder="Nama tampil pengguna" />
						</div>
						<div>
							<label for="u-pass" class="text-xs text-muted-foreground mb-1 block">Password</label>
							<PasswordInput id="u-pass" bind:value={fPassword} placeholder="Minimal 8 karakter" />
						</div>
						<div>
							<p class="mb-2 block text-xs text-muted-foreground">Peran Akses (boleh pilih lebih dari satu)</p>
							<div class="flex flex-wrap gap-2">
								{#each availableRoles as r (r.value)}
									<button
										class={`px-3 py-1 text-xs rounded-full border transition-colors ${fRoles.includes(r.value) ? 'bg-success text-background border-success' : 'bg-card text-muted-foreground border-border'}`}
										onclick={() => toggleRole(r.value)}
									>
										{r.label}
									</button>
								{/each}
							</div>
							<p class="mt-2 text-xs text-muted-foreground">{formHint}</p>
						</div>
					</div>

					<div class="space-y-3 rounded-2xl border bg-muted/20 p-4">
						<p class="text-sm font-semibold text-foreground">Tarik Data Profil</p>
						{#if candidateMode === 'employee'}
							<p class="text-xs text-muted-foreground">Guru only — tarik data pegawai/guru aktif.</p>
						{:else if candidateMode === 'student'}
							<p class="text-xs text-muted-foreground">Siswa per kelas — Pilih kelas terlebih dahulu untuk menarik siswa.</p>
						{:else if candidateMode === 'parent'}
							<p class="text-xs text-muted-foreground">Ortu per kelas anak — Pilih kelas anak untuk menarik orang tua/wali terkait.</p>
						{:else}
							<p class="text-xs text-muted-foreground">Role admin murni tidak wajib ditautkan ke profil.</p>
						{/if}

						{#if candidateRequiresClass}
							<div>
								<label for="candidate-class" class="mb-1 block text-xs text-muted-foreground">Kelas/Rombel</label>
								<select id="candidate-class" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={candidateClassId} onchange={() => { clearProfileSelection(); profileCandidatesRequestId += 1; profileCandidatesBusy = false; profileCandidates = []; }}>
									<option value="">-- Pilih Kelas --</option>
									{#each rombels.filter((item) => item.is_active !== false) as kelas (kelas.id)}
										<option value={kelas.id}>{kelas.name || kelas.code} {kelas.total_students ? `(${kelas.total_students} siswa)` : ''}</option>
									{/each}
								</select>
							</div>
						{/if}

						<div class="flex gap-2">
							<Input bind:value={candidateSearch} placeholder={candidateMode === 'parent' ? 'Cari nama ortu/anak' : 'Cari nama/NIP/NISN'} />
							<Button variant="outline" onclick={() => void loadProfileCandidates()} disabled={!candidateCanLoad || profileCandidatesBusy}>{profileCandidatesBusy ? 'Memuat...' : 'Tarik Data'}</Button>
						</div>
						<label class="flex items-center gap-2 text-xs text-muted-foreground"><input type="checkbox" bind:checked={includeLinkedCandidates} onchange={() => { profileCandidatesRequestId += 1; profileCandidatesBusy = false; profileCandidates = []; }} /> Tampilkan yang sudah tertaut</label>

						{#if profileCandidatesError}
							<p class="text-xs text-destructive">{profileCandidatesError}</p>
						{:else if selectedCandidate}
							<p class="text-xs text-success">Terpilih: {selectedCandidate.nama}</p>
						{/if}

						<div class="max-h-64 space-y-2 overflow-auto">
							{#each profileCandidates as candidate (candidate.id)}
								<button class={`w-full rounded-xl border p-3 text-left text-sm transition ${selectedCandidateId === candidate.id ? 'border-primary bg-primary/10' : 'border-border bg-card hover:bg-muted/60'}`} onclick={() => selectProfileCandidate(candidate)}>
									<div class="flex items-center justify-between gap-2"><span class="font-medium">{candidateLabel(candidate)}</span><Badge variant={candidate.is_linked ? 'outline' : 'secondary'}>{candidate.is_linked ? 'Sudah punya akun' : 'Belum punya akun'}</Badge></div>
									<p class="mt-1 text-xs text-muted-foreground">{candidateDescription(candidate)}</p>
								</button>
							{/each}
						</div>
					</div>
				</div>
				<div class="flex gap-2">
					<LoadingButton onclick={() => void createUser()} loading={fBusy} disabled={fBusy || !fUsername || !fPassword || fRoles.length === 0 || profileLinkMissing}>
						Simpan Pengguna
					</LoadingButton>
					<Button variant="outline" onclick={() => (showForm = false)}>Batal</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	<Card.Root class="border-primary/20 bg-primary/10 shadow-sm">
		<Card.Header class="flex flex-row items-start justify-between gap-3 pb-2">
			<div>
				<Card.Title class="text-base">Generate Akun dari Data Pegawai</Card.Title>
				<p class="mt-1 text-sm text-muted-foreground">Username/password awal: NPSN profil madrasah + 2 digit tahun lahir + nomor urut 3 digit. Role default: guru.</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<Button variant="outline" onclick={() => void previewEmployeeGeneration()} disabled={generationBusy !== null}>Preview</Button>
				<Button onclick={() => void runEmployeeGeneration()} disabled={generationBusy !== null || !employeeGeneration || employeeGeneration.ready === 0}>Generate Akun</Button>
				<Button variant="outline" onclick={downloadEmployeeGenerationCSV} disabled={!employeeGeneration}>Download CSV</Button>
			</div>
		</Card.Header>
		<Card.Content class="space-y-4">
			<div class="grid gap-3 md:grid-cols-6">
				<div class="rounded-xl border bg-card p-3"><p class="text-[11px] uppercase tracking-wide text-muted-foreground">NPSN Prefix</p><p class="font-mono text-xl font-semibold">{employeeGeneration?.npsn ?? '—'}</p></div>
				<div class="rounded-xl border bg-card p-3"><p class="text-[11px] uppercase tracking-wide text-muted-foreground">Total Pegawai</p><p class="text-xl font-semibold">{employeeGeneration?.total ?? 0}</p></div>
				<div class="rounded-xl border bg-card p-3"><p class="text-[11px] uppercase tracking-wide text-muted-foreground">Siap Dibuat</p><p class="text-xl font-semibold text-primary">{employeeGeneration?.ready ?? 0}</p></div>
				<div class="rounded-xl border bg-card p-3"><p class="text-[11px] uppercase tracking-wide text-muted-foreground">Dibuat</p><p class="text-xl font-semibold text-accent-foreground">{employeeGeneration?.created ?? 0}</p></div>
				<div class="rounded-xl border bg-card p-3"><p class="text-[11px] uppercase tracking-wide text-muted-foreground">Dilewati</p><p class="text-xl font-semibold text-warning">{employeeGeneration?.skipped ?? 0}</p></div>
				<div class="rounded-xl border bg-card p-3"><p class="text-[11px] uppercase tracking-wide text-muted-foreground">Gagal</p><p class="text-xl font-semibold text-destructive">{employeeGeneration?.failed ?? 0}</p></div>
			</div>
			{#if generationBusy}
				<p class="text-sm text-muted-foreground">Memproses {generationBusy === 'preview' ? 'preview' : 'generate'} akun pegawai...</p>
			{:else if !employeeGeneration}
				<EmptyStatePanel title="Belum Ada Preview" description="Klik Preview untuk melihat pegawai aktif yang siap dibuatkan akun otomatis." />
			{:else}
				<div class="max-h-96 overflow-auto rounded-2xl border bg-card">
					<Table.Root>
						<Table.Header>
							<Table.Row class="bg-muted/50">
								<Table.Head>Pegawai</Table.Head>
								<Table.Head>Tanggal Lahir</Table.Head>
								<Table.Head>Username</Table.Head>
								<Table.Head>Password Awal</Table.Head>
								<Table.Head>Status</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each employeeGeneration.items as item (item.employee_id)}
								<Table.Row>
									<Table.Cell><div class="font-medium">{item.nama}</div><div class="text-xs text-muted-foreground">{item.nip}</div></Table.Cell>
									<Table.Cell>{item.tanggal_lahir || '—'}</Table.Cell>
									<Table.Cell class="font-mono text-xs">{item.username || '—'}</Table.Cell>
									<Table.Cell class="font-mono text-xs">{item.password || (item.status === 'ready' ? 'ditampilkan setelah generate' : '—')}</Table.Cell>
									<Table.Cell><Badge variant={item.status === 'created' || item.status === 'ready' ? 'secondary' : item.status === 'failed' ? 'destructive' : 'outline'}>{item.status}</Badge><div class="mt-1 text-xs text-muted-foreground">{item.message}</div></Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</div>
			{/if}
		</Card.Content>
	</Card.Root>

	<Card.Root class="border-border bg-muted/20 shadow-sm">
		<Card.Header class="flex flex-row items-start justify-between gap-3 pb-2">
			<div>
				<Card.Title class="text-base">Manajemen RBAC</Card.Title>
				<p class="mt-1 text-sm text-muted-foreground">Halaman ini fokus ke akun pengguna dan assignment role. Editor role, permission, matrix akses, diff perubahan, dan guard permission kritikal dipusatkan di halaman RBAC khusus.</p>
			</div>
			<Button variant="outline" href="/settings/rbac">Buka Manajemen RBAC</Button>
		</Card.Header>
		<Card.Content>
			<div class="grid gap-3 md:grid-cols-3">
				<div class="rounded-2xl border border-border bg-card p-4">
					<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Role Aktif</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{rbac.roles.filter((role) => role.is_active !== false).length}</p>
					<p class="text-xs text-muted-foreground">dipakai sebagai pilihan assignment pengguna</p>
				</div>
				<div class="rounded-2xl border border-border bg-card p-4">
					<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Total Permission</p>
					<p class="mt-2 text-2xl font-semibold text-foreground">{rbac.permissions.length}</p>
					<p class="text-xs text-muted-foreground">dikelola melalui katalog permission RBAC</p>
				</div>
				<div class="rounded-2xl border border-border bg-card p-4">
					<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Scope Halaman Ini</p>
					<p class="mt-2 text-sm font-semibold text-foreground">User lifecycle</p>
					<p class="text-xs text-muted-foreground">buat akun, tautkan profil, reset password, status akun, dan assignment role</p>
				</div>
			</div>
		</Card.Content>
	</Card.Root>

	<AsyncContent promise={usersPromise} onerror={handleOverviewRenderError}>
		{#snippet pending()}
			<Card.Root class="overflow-hidden border-border shadow-sm">
				<Card.Content class="p-0">
				<div class="space-y-3 p-6">
					{#each Array.from({ length: 5 }) as _, index (`user-skeleton-${index}`)}
						<div class="grid gap-3 md:grid-cols-[1fr_1.1fr_1fr_0.7fr_0.8fr_auto] md:items-center">
							<Skeleton class="h-5 w-28" />
							<Skeleton class="h-5 w-36" />
							<Skeleton class="h-5 w-40" />
							<Skeleton class="h-6 w-20" />
							<Skeleton class="h-5 w-24" />
							<Skeleton class="h-9 w-40 justify-self-end" />
						</div>
					{/each}
				</div>
				</Card.Content>
			</Card.Root>
		{/snippet}

		{#snippet children(value)}
			{@const overview = value as UsersOverview}
	<Card.Root class="overflow-hidden border-border shadow-sm">
		<Card.Content class="p-0">
				<div class="hidden overflow-x-auto lg:block">
				<Table.Root>
					<Table.Header>
						<Table.Row class="bg-muted/50">
							<Table.Head>Username / Display Name</Table.Head>
							<Table.Head>Role Dinamis</Table.Head>
							<Table.Head>Profil Terhubung</Table.Head>
							<Table.Head>Status</Table.Head>
							<Table.Head>Last Login</Table.Head>
							<Table.Head>Dibuat</Table.Head>
							<Table.Head>Aksi</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each overview.users as u (u.id)}
							<Table.Row>
								<Table.Cell class="font-medium">
									<div>{u.username}</div>
									<div class="mt-1 text-xs font-normal text-muted-foreground">{u.display_name || u.profile_nama || '—'}</div>
								</Table.Cell>
								<Table.Cell>
									<div class="flex max-w-md flex-wrap gap-1.5">
										{#each availableRoles as r (r.value)}
											<button
												class={`rounded-full border px-2.5 py-1 text-[10px] font-semibold uppercase transition-colors ${u.roles?.includes(r.value) ? 'border-success bg-success text-background' : 'border-border bg-card text-muted-foreground hover:border-success/20 hover:text-success'}`}
												disabled={actionBusy === `roles:${u.id}`}
												title={`Toggle role ${r.label}`}
												onclick={() => void toggleExistingUserRole(u, r.value)}
											>
												{roleLabel(r.value)}
											</button>
										{/each}
									</div>
								</Table.Cell>
								<Table.Cell class="text-sm text-muted-foreground">
									<div>{u.profile_nama || '—'}</div>
									<div class="mt-1 text-[11px] text-muted-foreground">
										{u.employee_id ? 'Pegawai' : u.student_id ? 'Siswa' : u.parent_id ? 'Orang tua' : 'Belum ditautkan'}
									</div>
								</Table.Cell>
								<Table.Cell>
									<Badge variant={u.is_active ? 'outline' : 'destructive'}>
										{u.is_active ? 'Aktif' : 'Nonaktif'}
									</Badge>
								</Table.Cell>
								<Table.Cell class="text-xs text-muted-foreground">{formatDateTime(u.last_login_at)}</Table.Cell>
								<Table.Cell class="text-xs text-muted-foreground">{new Date(u.created_at).toLocaleDateString()}</Table.Cell>
								<Table.Cell class="text-right">
									<div class="flex flex-wrap justify-end gap-2">
										<Button
											variant="outline"
											size="sm"
											disabled={actionBusy === `password:${u.id}`}
											onclick={() => void resetPasswordForUser(u)}
										>
											Reset PW
										</Button>
										<Button
											variant="outline"
											size="sm"
											disabled={actionBusy === `profile:${u.id}`}
											onclick={() => void updateProfileForUser(u)}
										>
											Profil
										</Button>
										<Button
											variant="outline"
											size="sm"
											onclick={() => toggleUserStatus(u)}
										>
											{u.is_active ? 'Nonaktifkan' : 'Aktifkan'}
										</Button>
										<Button variant="ghost" size="sm" onclick={() => deleteUser(u.id, u.username)}
											class="text-destructive hover:text-destructive hover:bg-destructive/10">Hapus</Button>
									</div>
								</Table.Cell>
							</Table.Row>
						{:else}
							<Table.Row>
								<Table.Cell colspan={7} class="p-4">
									<EmptyStatePanel
										compact
										title="Belum ada data pengguna"
										description="Tambahkan akun pertama untuk mulai menghubungkan pegawai, siswa, atau orang tua ke akses sistem."
									/>
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
				</div>

				<div class="grid gap-3 p-4 lg:hidden">
					{#each overview.users as u (u.id)}
						<div class="rounded-2xl border border-border bg-card p-4 shadow-sm">
							<div class="flex items-start justify-between gap-3">
								<div class="min-w-0">
									<p class="text-sm font-semibold text-foreground">{u.username}</p>
									<p class="mt-1 text-xs text-muted-foreground">{u.display_name || u.profile_nama || '—'}</p>
									<div class="mt-1 flex flex-wrap gap-1">
										{#each availableRoles as r (r.value)}
											<button
												class={`rounded-full border px-2 py-0.5 text-[10px] font-semibold uppercase ${u.roles?.includes(r.value) ? 'border-success bg-success text-background' : 'border-border bg-card text-muted-foreground'}`}
												disabled={actionBusy === `roles:${u.id}`}
												onclick={() => void toggleExistingUserRole(u, r.value)}
											>
												{roleLabel(r.value)}
											</button>
										{/each}
									</div>
									<p class="mt-2 text-xs text-muted-foreground">{u.profile_nama || 'Tidak terhubung profil'}</p>
									<p class="mt-1 text-xs text-muted-foreground">Last login: {formatDateTime(u.last_login_at)}</p>
								</div>
								<Badge variant={u.is_active ? 'outline' : 'destructive'}>{u.is_active ? 'Aktif' : 'Nonaktif'}</Badge>
							</div>
							<div class="mt-4 flex flex-wrap gap-2">
								<Button
									variant="outline"
									size="sm"
									class="flex-1 justify-center"
									disabled={actionBusy === `password:${u.id}`}
									onclick={() => void resetPasswordForUser(u)}
								>
									Reset PW
								</Button>
								<Button
									variant="outline"
									size="sm"
									class="flex-1 justify-center"
									disabled={actionBusy === `profile:${u.id}`}
									onclick={() => void updateProfileForUser(u)}
								>
									Profil
								</Button>
								<Button
									variant="outline"
									size="sm"
									class="flex-1 justify-center"
									onclick={() => toggleUserStatus(u)}
								>
									{u.is_active ? 'Nonaktifkan' : 'Aktifkan'}
								</Button>
								<Button variant="ghost" size="sm" onclick={() => deleteUser(u.id, u.username)}
									class="flex-1 justify-center text-destructive hover:text-destructive hover:bg-destructive/10">Hapus</Button>
							</div>
						</div>
					{:else}
						<EmptyStatePanel
							title="Belum ada data pengguna"
							description="Tambahkan akun pertama agar role sekolah dan akses portal bisa mulai dikelola dari panel ini."
						/>
					{/each}
				</div>
		</Card.Content>
	</Card.Root>
		{/snippet}
	</AsyncContent>
</div>
