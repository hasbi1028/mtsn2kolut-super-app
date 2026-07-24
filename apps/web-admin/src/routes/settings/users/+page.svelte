<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import PasswordInput from '$lib/components/PasswordInput.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import { TablePagination } from '$lib/components/ui/pagination';
	import { toast } from '$lib/components/ui/sonner';
	import { confirmAction } from '$lib/confirm-dialog';
	import { readClientApiData, readClientJson } from '$lib/client/api';
	import { trackInternalAnalyticsEvent } from '$lib/analytics/internal-analytics';
	import { displayName } from '$lib/utils/display-name';
	import { DEFAULT_PAGE_SIZE_OPTIONS, clampPage, paginateItems, type PaginationChange } from '$lib/utils/pagination';
	import {
		employeeAccountGenerationCSV,
		fetchRBACMatrix,
		fetchUserProfileCandidates,
		generateEmployeeAccounts,
		generateParentAccounts,
		generateStudentAccounts,
		parentAccountGenerationCSV,
		previewEmployeeAccountGeneration,
		previewParentAccounts,
		previewStudentAccounts,
		resetUserPassword,
		studentAccountGenerationCSV,
		updateUserProfileLink,
		updateUserRoles,
		type EmployeeAccountGenerationResult,
		type ParentAccountGenerationResult,
		type RBACMatrix,
		type RBACRole,
		type StudentAccountGenerationResult,
		type UserProfileCandidate
	} from '$lib/client/rbac-users';

	type User = {
		id: string;
		username: string;
		display_name?: string | null;
		roles: string[];
		employee_id: string | null;
		student_id: string | null;
		parent_id: string | null;
		profile_nama: string | null;
		is_active: boolean;
		last_login_at?: string | null;
		deleted_at?: string | null;
		created_at: string;
	};
	type Rombel = { id: string; name: string; code?: string | null; is_active?: boolean; total_students?: number };
	type UsersOverview = { users: User[]; rombels: Rombel[]; rbac: RBACMatrix };
	type AccountTab = 'ringkasan' | 'pegawai' | 'siswa' | 'ortu' | 'admin' | 'generate' | 'audit';
	type AccountType = 'employee' | 'student' | 'parent' | 'admin';
	type CandidateMode = 'employee' | 'student' | 'parent' | 'none';
	type GenerationAudience = 'employee' | 'student' | 'parent';

	const fallbackRoles = [
		{ value: 'admin', label: 'Administrator' },
		{ value: 'guru', label: 'Guru' },
		{ value: 'staf', label: 'Staf' },
		{ value: 'kesiswaan', label: 'Kesiswaan' },
		{ value: 'siswa', label: 'Siswa' },
		{ value: 'ortu', label: 'Orang Tua' }
	];

	let users = $state<User[]>([]);
	let rombels = $state<Rombel[]>([]);
	let rbac = $state<RBACMatrix>({ roles: [], permissions: [], role_permissions: {} });
	let usersPromise = $state<Promise<UsersOverview> | null>(null);
	let usersRequestId = 0;
	let activeTab = $state<AccountTab>('ringkasan');
	let searchQuery = $state('');
	let roleFilter = $state('all');
	let statusFilter = $state<'all' | 'active' | 'inactive' | 'unlinked' | 'never-login' | 'multi-role'>('all');
	let usersPage = $state(1);
	let usersPageSize = $state<number>(DEFAULT_PAGE_SIZE_OPTIONS[0]);
	let generationPage = $state(1);
	let generationPageSize = $state<number>(DEFAULT_PAGE_SIZE_OPTIONS[0]);
	let selectedUserIds = $state<string[]>([]);
	let selectedUser = $state<User | null>(null);
	let createPanelOpen = $state(false);
	let createStep = $state(1);
	let accountType = $state<AccountType>('employee');
	let fUsername = $state('');
	let fPassword = $state('');
	let fDisplayName = $state('');
	let fRoles = $state<string[]>(['guru']);
	let fEmpId = $state('');
	let fStuId = $state('');
	let fParId = $state('');
	let fBusy = $state(false);
	let actionBusy = $state<string | null>(null);
	let candidateClassId = $state('');
	let candidateSearch = $state('');
	let includeLinkedCandidates = $state(false);
	let profileCandidates = $state<UserProfileCandidate[]>([]);
	let profileCandidatesBusy = $state(false);
	let profileCandidatesError = $state('');
	let profileDisplayNameAutofill = $state('');
	let profileCandidatesRequestId = 0;
	let generationAudience = $state<GenerationAudience>('employee');
	let employeeGeneration = $state<EmployeeAccountGenerationResult | null>(null);
	let studentGeneration = $state<StudentAccountGenerationResult | null>(null);
	let parentGeneration = $state<ParentAccountGenerationResult | null>(null);
	let generationBusy = $state<string | null>(null);
	let generationSearch = $state('');

	const availableRoles = $derived.by(() => {
		const activeRoles = (rbac.roles ?? [])
			.filter((role: RBACRole) => role.is_active !== false)
			.map((role: RBACRole) => ({ value: role.code, label: role.name || role.code }));
		return activeRoles.length ? activeRoles : fallbackRoles;
	});
	const employeeUsers = $derived(users.filter((user) => isEmployeeUser(user)));
	const studentUsers = $derived(users.filter((user) => user.roles?.includes('siswa') || Boolean(user.student_id)));
	const parentUsers = $derived(users.filter((user) => user.roles?.includes('ortu') || Boolean(user.parent_id)));
	const adminUsers = $derived(users.filter((user) => user.roles?.includes('admin')));
	const problemUsers = $derived(users.filter((user) => accountHealth(user).length > 0));
	const visibleUsers = $derived.by(() => filterUsers(usersForActiveTab()));
	const safeUsersPage = $derived(clampPage(usersPage, visibleUsers.length, usersPageSize));
	const paginatedUsers = $derived(paginateItems(visibleUsers, safeUsersPage, usersPageSize));
	const selectedUsers = $derived(users.filter((user) => selectedUserIds.includes(user.id)));
	const allVisibleSelected = $derived(paginatedUsers.length > 0 && paginatedUsers.every((user) => selectedUserIds.includes(user.id)));
	const currentGeneration = $derived.by(() => {
		if (generationAudience === 'employee') return employeeGeneration;
		if (generationAudience === 'student') return studentGeneration;
		return parentGeneration;
	});
	const visibleGenerationRows = $derived.by(() => filterGenerationRows(currentGenerationRows()));
	const safeGenerationPage = $derived(clampPage(generationPage, visibleGenerationRows.length, generationPageSize));
	const paginatedGenerationRows = $derived(paginateItems(visibleGenerationRows, safeGenerationPage, generationPageSize));
	const candidateMode = $derived.by<CandidateMode>(() => {
		if (accountType === 'employee') return 'employee';
		if (accountType === 'student') return 'student';
		if (accountType === 'parent') return 'parent';
		return 'none';
	});
	const candidateRole = $derived.by(() => {
		if (accountType === 'student') return 'siswa';
		if (accountType === 'parent') return 'ortu';
		if (accountType === 'employee') return fRoles.includes('kesiswaan') ? 'kesiswaan' : fRoles.includes('staf') ? 'staf' : 'guru';
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

	function applyOverview(overview: UsersOverview) {
		users = overview.users;
		rombels = overview.rombels;
		rbac = overview.rbac;
		selectedUserIds = selectedUserIds.filter((id) => overview.users.some((user) => user.id === id));
		if (selectedUser) selectedUser = overview.users.find((user) => user.id === selectedUser?.id) ?? null;
		return overview;
	}

	async function fetchOverview(): Promise<UsersOverview> {
		const [usersRes, rombelsRes, nextRBAC] = await Promise.all([
			fetch('/api/users'),
			fetch('/api/academic/rombel'),
			fetchRBACMatrix()
		]);
		const [nextUsers, nextRombels] = await Promise.all([
			readClientApiData<User[]>(usersRes, 'Gagal memuat data pengguna.'),
			readClientApiData<Rombel[]>(rombelsRes, 'Gagal memuat data kelas.')
		]);
		return { users: nextUsers ?? [], rombels: nextRombels ?? [], rbac: nextRBAC };
	}

	function load() {
		const requestId = ++usersRequestId;
		usersPromise = fetchOverview()
			.then((overview) => (requestId === usersRequestId ? applyOverview(overview) : { users, rombels, rbac }))
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
			toast.error(errorMessage(error));
		}
	}

	function errorMessage(error: unknown) {
		return error instanceof Error && error.message.trim() ? error.message : 'Operasi pengguna gagal.';
	}

	function handleOverviewRenderError(error: unknown) {
		console.error('Users workflow render failed', error);
	}

	function userDisplayLabel(user: User) {
		return displayName({ display_name: user.display_name, nama: user.profile_nama, username: user.username }, 'Pengguna');
	}

	function usernameLabel(user: User) {
		return user.username?.trim() ? `@${user.username.trim()}` : 'Username belum tercatat';
	}

	function roleLabel(roleCode: string) {
		return availableRoles.find((role) => role.value === roleCode)?.label ?? roleCode;
	}

	function isEmployeeUser(user: User) {
		return Boolean(user.employee_id) || (user.roles ?? []).some((role) => ['guru', 'staf', 'kesiswaan'].includes(role));
	}

	function profileType(user: User) {
		if (user.employee_id) return 'Pegawai';
		if (user.student_id) return 'Siswa';
		if (user.parent_id) return 'Orang tua';
		return 'Belum tertaut';
	}

	function accountHealth(user: User) {
		const items: string[] = [];
		if (!user.is_active) items.push('Nonaktif');
		if (!user.profile_nama && !user.roles?.includes('admin')) items.push('Belum tertaut profil');
		if (!user.last_login_at) items.push('Belum pernah login');
		if ((user.roles ?? []).length > 1) items.push('Multi-peran');
		if (user.roles?.includes('admin')) items.push('Akses sensitif');
		return items;
	}

	function usersForActiveTab() {
		if (activeTab === 'pegawai') return employeeUsers;
		if (activeTab === 'siswa') return studentUsers;
		if (activeTab === 'ortu') return parentUsers;
		if (activeTab === 'admin') return adminUsers;
		if (activeTab === 'ringkasan') return users;
		return users;
	}

	function filterUsers(source: User[]) {
		const query = searchQuery.trim().toLowerCase();
		return source.filter((user) => {
			const roles = user.roles ?? [];
			const matchesQuery = !query || [user.username, user.display_name ?? '', user.profile_nama ?? '', roles.join(' ')].join(' ').toLowerCase().includes(query);
			const matchesRole = roleFilter === 'all' || roles.includes(roleFilter);
			const matchesStatus =
				statusFilter === 'all' ||
				(statusFilter === 'active' && user.is_active) ||
				(statusFilter === 'inactive' && !user.is_active) ||
				(statusFilter === 'unlinked' && !user.profile_nama && !roles.includes('admin')) ||
				(statusFilter === 'never-login' && !user.last_login_at) ||
				(statusFilter === 'multi-role' && roles.length > 1);
			return matchesQuery && matchesRole && matchesStatus;
		});
	}

	function setTab(tab: AccountTab) {
		activeTab = tab;
		selectedUserIds = [];
		usersPage = 1;
	}

	function toggleSelectVisible() {
		if (allVisibleSelected) {
			selectedUserIds = selectedUserIds.filter((id) => !paginatedUsers.some((user) => user.id === id));
		} else {
			selectedUserIds = Array.from(new Set([...selectedUserIds, ...paginatedUsers.map((user) => user.id)]));
		}
	}

	function handleUserPagination(change: PaginationChange) {
		usersPage = change.reason === 'limit' ? 1 : change.page;
		usersPageSize = change.limit;
	}

	function handleGenerationPagination(change: PaginationChange) {
		generationPage = change.reason === 'limit' ? 1 : change.page;
		generationPageSize = change.limit;
	}

	function handleUserSearchInput(event: Event) {
		searchQuery = (event.currentTarget as HTMLInputElement).value;
		usersPage = 1;
	}

	function setRoleFilter(event: Event) {
		roleFilter = (event.currentTarget as HTMLSelectElement).value;
		usersPage = 1;
	}

	function setStatusFilter(event: Event) {
		statusFilter = (event.currentTarget as HTMLSelectElement).value as typeof statusFilter;
		usersPage = 1;
	}

	function handleGenerationSearchInput(event: Event) {
		generationSearch = (event.currentTarget as HTMLInputElement).value;
		generationPage = 1;
	}

	function toggleUserSelection(user: User) {
		selectedUserIds = selectedUserIds.includes(user.id)
			? selectedUserIds.filter((id) => id !== user.id)
			: [...selectedUserIds, user.id];
	}

	function openCreate(type: AccountType) {
		accountType = type;
		createStep = 1;
		createPanelOpen = true;
		resetCreateDraft(type);
	}

	function resetCreateDraft(type = accountType) {
		fUsername = '';
		fPassword = '';
		fDisplayName = '';
		fRoles = type === 'student' ? ['siswa'] : type === 'parent' ? ['ortu'] : type === 'admin' ? ['admin'] : ['guru'];
		fEmpId = '';
		fStuId = '';
		fParId = '';
		candidateClassId = '';
		candidateSearch = '';
		profileCandidates = [];
		profileCandidatesError = '';
		profileDisplayNameAutofill = '';
	}

	function setAccountType(type: AccountType) {
		accountType = type;
		resetCreateDraft(type);
	}

	function clearProfileSelection() {
		fEmpId = '';
		fStuId = '';
		fParId = '';
	}

	function toggleDraftRole(role: string) {
		if (fRoles.includes(role)) fRoles = fRoles.filter((item) => item !== role);
		else fRoles = [...fRoles, role];
		clearProfileSelection();
		profileCandidatesRequestId += 1;
		profileCandidates = [];
	}

	function candidateLabel(candidate: UserProfileCandidate) {
		return `${candidate.nama}${candidate.identifier ? ` (${candidate.identifier})` : ''}`;
	}

	function candidateDescription(candidate: UserProfileCandidate) {
		if (candidate.profile_type === 'parent' && candidate.children?.length) return `Anak: ${candidate.children.map((child) => child.nama).join(', ')}`;
		return candidate.class_name ? `Kelas ${candidate.class_name}` : candidate.profile_type;
	}

	function selectProfileCandidate(candidate: UserProfileCandidate) {
		clearProfileSelection();
		if (candidate.profile_type === 'employee') fEmpId = candidate.id;
		if (candidate.profile_type === 'student') fStuId = candidate.id;
		if (candidate.profile_type === 'parent') fParId = candidate.id;
		if (!fDisplayName || fDisplayName === profileDisplayNameAutofill) fDisplayName = candidate.nama;
		profileDisplayNameAutofill = candidate.nama;
		if (!fUsername) fUsername = makeUsernameSuggestion(candidate);
	}

	function makeUsernameSuggestion(candidate: UserProfileCandidate) {
		return (candidate.identifier || candidate.nama || '').toLowerCase().replace(/[^a-z0-9]/g, '').slice(0, 32);
	}

	async function loadProfileCandidates() {
		const requestId = ++profileCandidatesRequestId;
		if (!candidateCanLoad) {
			profileCandidatesError = candidateRequiresClass ? 'Pilih kelas terlebih dahulu.' : '';
			profileCandidates = [];
			return;
		}
		profileCandidatesBusy = true;
		profileCandidatesError = '';
		try {
			const response = await fetchUserProfileCandidates({
				role: candidateRole,
				class_id: candidateRequiresClass ? candidateClassId : null,
				q: candidateSearch,
				include_linked: includeLinkedCandidates,
				limit: 80
			});
			if (requestId !== profileCandidatesRequestId) return;
			profileCandidates = response?.candidates ?? [];
		} catch (error) {
			if (requestId !== profileCandidatesRequestId) return;
			profileCandidates = [];
			profileCandidatesError = errorMessage(error);
		} finally {
			if (requestId === profileCandidatesRequestId) profileCandidatesBusy = false;
		}
	}

	async function createUser() {
		if (!fUsername || !fPassword || fRoles.length === 0) {
			toast.error('Username, password, dan minimal satu peran wajib diisi.');
			return;
		}
		if (profileLinkMissing) {
			toast.error('Akun jenis ini wajib ditautkan ke profil resmi.');
			return;
		}
		fBusy = true;
		try {
			const res = await fetch('/api/users', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					username: fUsername.trim(),
					password: fPassword,
					display_name: fDisplayName || null,
					roles: fRoles,
					employee_id: fEmpId || null,
					student_id: fStuId || null,
					parent_id: fParId || null
				})
			});
			await readClientJson<unknown>(res);
			toast.success('Pengguna berhasil dibuat.');
			createPanelOpen = false;
			resetCreateDraft();
			await refreshOverview();
		} catch (error) {
			toast.error(errorMessage(error));
		} finally {
			fBusy = false;
		}
	}

	async function updateRolesForUser(user: User, nextRoles: string[]) {
		const roles = Array.from(new Set(nextRoles.filter(Boolean)));
		if (!roles.length) {
			toast.error('Minimal satu peran wajib dipilih.');
			return;
		}
		if ((roles.includes('admin') || user.roles.includes('admin')) && !(await confirmAction({
			title: 'Perubahan Role Sensitif',
			message: `Perbarui role untuk ${userDisplayLabel(user)}? Role admin/akses sensitif harus dibatasi.`,
			confirmLabel: 'Perbarui Role',
			tone: 'warning'
		}))) return;
		actionBusy = `roles:${user.id}`;
		try {
			await updateUserRoles(user.id, roles);
			toast.success('Peran pengguna diperbarui.');
			await refreshOverview();
		} catch (error) {
			toast.error(errorMessage(error));
		} finally {
			actionBusy = null;
		}
	}

	async function toggleExistingUserRole(user: User, role: string) {
		const current = user.roles ?? [];
		await updateRolesForUser(user, current.includes(role) ? current.filter((item) => item !== role) : [...current, role]);
	}

	async function resetPasswordForUser(user: User) {
		const password = window.prompt(`Password baru untuk ${userDisplayLabel(user)} (minimal 8 karakter):`);
		if (password === null) return;
		if (password.trim().length < 8) {
			toast.error('Password minimal 8 karakter.');
			return;
		}
		if (!(await confirmAction({
			title: 'Reset Password',
			message: `Reset password untuk ${userDisplayLabel(user)}? Session aktif user akan dicabut.`,
			confirmLabel: 'Reset Password',
			tone: 'warning'
		}))) return;
		actionBusy = `password:${user.id}`;
		try {
			await resetUserPassword(user.id, password.trim());
			toast.success('Password berhasil direset.');
		} catch (error) {
			toast.error(errorMessage(error));
		} finally {
			actionBusy = null;
		}
	}

	async function forcePasswordChangeForUser(user: User) {
		if (!(await confirmAction({
			title: 'Wajib Ganti Password',
			message: `Paksa ${userDisplayLabel(user)} mengganti password saat login berikutnya?`,
			confirmLabel: 'Wajibkan',
			tone: 'warning'
		}))) return;
		actionBusy = `force-password:${user.id}`;
		try {
			const res = await fetch(`/api/users/${user.id}/force-password-change`, { method: 'POST' });
			await readClientJson<unknown>(res);
			toast.success('User diwajibkan mengganti password.');
		} catch (error) {
			toast.error(errorMessage(error));
		} finally {
			actionBusy = null;
		}
	}

	async function updateProfileForUser(user: User) {
		const type = window.prompt('Jenis profil: employee, student, parent, atau kosong untuk melepas tautan', user.employee_id ? 'employee' : user.student_id ? 'student' : user.parent_id ? 'parent' : '');
		if (type === null) return;
		const normalizedType = type.trim().toLowerCase();
		if (normalizedType && !['employee', 'student', 'parent'].includes(normalizedType)) {
			toast.error('Jenis profil tidak valid.');
			return;
		}
		let profileID = '';
		if (normalizedType) {
			profileID = window.prompt('Masukkan UUID profil tujuan:', user.employee_id || user.student_id || user.parent_id || '')?.trim() ?? '';
			if (!profileID) {
				toast.error('ID profil wajib diisi.');
				return;
			}
		}
		actionBusy = `profile:${user.id}`;
		try {
			await updateUserProfileLink(user.id, {
				employee_id: normalizedType === 'employee' ? profileID : null,
				student_id: normalizedType === 'student' ? profileID : null,
				parent_id: normalizedType === 'parent' ? profileID : null
			});
			toast.success('Tautan profil diperbarui.');
			await refreshOverview();
		} catch (error) {
			toast.error(errorMessage(error));
		} finally {
			actionBusy = null;
		}
	}

	async function toggleUserStatus(user: User) {
		if (user.username === 'admin' && user.is_active) {
			toast.error('Akun admin utama tidak bisa dinonaktifkan.');
			return;
		}
		const next = !user.is_active;
		if (!(await confirmAction({
			title: next ? 'Aktifkan Akun' : 'Nonaktifkan Akun',
			message: `${next ? 'Aktifkan' : 'Nonaktifkan'} akun ${userDisplayLabel(user)}?`,
			confirmLabel: next ? 'Aktifkan' : 'Nonaktifkan',
			tone: 'warning'
		}))) return;
		actionBusy = `status:${user.id}`;
		try {
			const res = await fetch(`/api/users/${user.id}/status`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ is_active: next })
			});
			await readClientJson<unknown>(res);
			toast.success(next ? 'Akun diaktifkan.' : 'Akun dinonaktifkan.');
			await refreshOverview();
		} catch (error) {
			toast.error(errorMessage(error));
		} finally {
			actionBusy = null;
		}
	}

	async function deleteUser(user: User) {
		if (user.username === 'admin') {
			toast.error('User admin utama tidak bisa dihapus.');
			return;
		}
		if (!(await confirmAction({
			title: 'Hapus/Nonaktifkan Pengguna',
			message: `Hapus pengguna ${userDisplayLabel(user)}? Backend akan memproses sesuai aturan lifecycle aman.`,
			confirmLabel: 'Hapus Pengguna',
			tone: 'danger'
		}))) return;
		try {
			const res = await fetch(`/api/users/${user.id}`, { method: 'DELETE' });
			if (res.status !== 204) await readClientJson<unknown>(res);
			toast.success('Pengguna berhasil diproses.');
			await refreshOverview();
		} catch (error) {
			toast.error(errorMessage(error));
		}
	}

	async function previewGeneration() {
		generationBusy = `preview:${generationAudience}`;
		try {
			if (generationAudience === 'employee') employeeGeneration = await previewEmployeeAccountGeneration();
			if (generationAudience === 'student') studentGeneration = await previewStudentAccounts();
			if (generationAudience === 'parent') parentGeneration = await previewParentAccounts();
			toast.success('Preview generate akun dimuat.');
		} catch (error) {
			toast.error(errorMessage(error));
		} finally {
			generationBusy = null;
		}
	}

	async function runGeneration() {
		const label = generationAudience === 'employee' ? 'pegawai' : generationAudience === 'student' ? 'siswa' : 'orang tua';
		if (!(await confirmAction({
			title: `Generate Akun ${label}`,
			message: `Buat akun ${label} yang berstatus siap? Pastikan preview sudah dicek dan CSV hasil generate langsung diunduh untuk arsip resmi.`,
			confirmLabel: 'Generate Akun',
			tone: 'warning'
		}))) return;
		generationBusy = `generate:${generationAudience}`;
		try {
			if (generationAudience === 'employee') employeeGeneration = await generateEmployeeAccounts();
			if (generationAudience === 'student') studentGeneration = await generateStudentAccounts();
			if (generationAudience === 'parent') parentGeneration = await generateParentAccounts();
			toast.success(`Generate akun ${label} selesai.`);
			await refreshOverview();
		} catch (error) {
			toast.error(errorMessage(error));
		} finally {
			generationBusy = null;
		}
	}

	function currentGenerationRows() {
		if (generationAudience === 'employee') return employeeGeneration?.items ?? [];
		if (generationAudience === 'student') return studentGeneration?.candidates ?? [];
		return parentGeneration?.candidates ?? [];
	}

	function filterGenerationRows(rows: Array<Record<string, unknown>>) {
		const query = generationSearch.trim().toLowerCase();
		if (!query) return rows;
		return rows.filter((row) => Object.values(row).join(' ').toLowerCase().includes(query));
	}

	function downloadGenerationCSV() {
		let csv = '';
		if (generationAudience === 'employee' && employeeGeneration) csv = employeeAccountGenerationCSV(employeeGeneration);
		if (generationAudience === 'student' && studentGeneration) csv = studentAccountGenerationCSV(studentGeneration);
		if (generationAudience === 'parent' && parentGeneration) csv = parentAccountGenerationCSV(parentGeneration);
		if (!csv) return;
		const blob = new Blob([csv], { type: 'text/csv;charset=utf-8' });
		const url = URL.createObjectURL(blob);
		const anchor = document.createElement('a');
		anchor.href = url;
		anchor.download = `generate-akun-${generationAudience}-${new Date().toISOString().slice(0, 10)}.csv`;
		anchor.click();
		URL.revokeObjectURL(url);
	}

	function formatDateTime(value?: string | null) {
		if (!value) return 'Belum pernah login';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return 'Belum pernah login';
		return date.toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short' });
	}

	onMount(() => {
		void trackInternalAnalyticsEvent('users.list_view', {
			pathname: window.location.pathname,
			metadata: { page_key: 'users-list' }
		});
		void trackInternalAnalyticsEvent('users.workflow_view', { pathname: window.location.pathname, metadata: { page_key: 'users-workflow' } });
		load();
	});
</script>

<svelte:head><title>Pusat Akun & Hak Akses — MTsN 2 Kolut</title></svelte:head>

<div class="space-y-6 p-6">
	<!-- Breadcrumb -->
	<div class="flex items-center gap-2 text-sm text-base-content/70">
		<a href={resolve('/')} class="hover:text-base-content">Beranda</a>
		<span>/</span>
		<a href={resolve('/settings')} class="hover:text-base-content">Pengaturan</a>
		<span>/</span>
		<span class="text-base-content font-medium">Manajemen Pengguna</span>
	</div>

	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<p class="text-xs font-semibold uppercase tracking-[0.24em] text-primary">Pusat Akun & Hak Akses</p>
			<h1 class="mt-2 text-2xl font-semibold text-base-content">Manajemen Pengguna</h1>
			<p class="mt-1 max-w-3xl text-sm text-base-content/70">Kelola akun pegawai, siswa, orang tua, admin, generate akun massal, dan role/permission dari workflow yang lebih aman untuk operasional madrasah.</p>
		</div>
		<div class="flex flex-wrap gap-2">
			<Button variant="outline" href="/settings/user-change-requests">Permintaan Data</Button>
			<Button variant="outline" href="/settings/rbac">Manajemen Hak Akses</Button>
			<Button onclick={() => openCreate(activeTab === 'siswa' ? 'student' : activeTab === 'ortu' ? 'parent' : activeTab === 'admin' ? 'admin' : 'employee')}>+ Tambah Akun</Button>
		</div>
	</div>

	<AsyncContent promise={usersPromise} onerror={handleOverviewRenderError}>
		{#snippet failed(error, reset)}
			<RecoveryPanel title="Data pengguna belum tersaji" message={errorMessage(error)} onRetry={() => { reset?.(); load(); }} />
		{/snippet}

		{#snippet pending()}
			<div class="grid gap-3 md:grid-cols-4">
				{#each ['Total Akun', 'Akun Bermasalah', 'Siswa', 'Orang Tua'] as label}
					<div class="rounded-2xl border bg-base-200/30 p-4"><p class="text-xs text-base-content/70">{label}</p><p class="mt-2 text-2xl font-semibold">…</p></div>
				{/each}
			</div>
		{/snippet}

		{#snippet children()}
			<div class="grid gap-3 md:grid-cols-6">
				<div class="rounded-2xl border border-primary/20 bg-primary/10 p-4"><p class="text-xs font-semibold uppercase text-primary">Total</p><p class="mt-2 text-2xl font-semibold">{users.length}</p><p class="text-xs text-base-content/70">akun terdaftar</p></div>
				<div class="rounded-2xl border bg-base-100 p-4"><p class="text-xs font-semibold uppercase text-base-content/70">Pegawai</p><p class="mt-2 text-2xl font-semibold">{employeeUsers.length}</p><p class="text-xs text-base-content/70">guru/staf/kesiswaan</p></div>
				<div class="rounded-2xl border bg-base-100 p-4"><p class="text-xs font-semibold uppercase text-base-content/70">Siswa</p><p class="mt-2 text-2xl font-semibold">{studentUsers.length}</p><p class="text-xs text-base-content/70">portal siswa</p></div>
				<div class="rounded-2xl border bg-base-100 p-4"><p class="text-xs font-semibold uppercase text-base-content/70">Orang Tua</p><p class="mt-2 text-2xl font-semibold">{parentUsers.length}</p><p class="text-xs text-base-content/70">portal wali</p></div>
				<div class="rounded-2xl border border-warning/40 bg-warning/10 p-4"><p class="text-xs font-semibold uppercase text-warning">Perlu Cek</p><p class="mt-2 text-2xl font-semibold">{problemUsers.length}</p><p class="text-xs text-base-content/70">akun nonaktif/aneh</p></div>
				<div class="rounded-2xl border border-destructive/30 bg-destructive/10 p-4"><p class="text-xs font-semibold uppercase text-destructive">Admin</p><p class="mt-2 text-2xl font-semibold">{adminUsers.length}</p><p class="text-xs text-base-content/70">akses sensitif</p></div>
			</div>
		{/snippet}
	</AsyncContent>

	<div role="tablist" class="tabs tabs-box bg-base-100 overflow-x-auto flex-nowrap">
		{#each [
			{ id: 'ringkasan', label: 'Ringkasan' },
			{ id: 'pegawai', label: 'Akun Pegawai' },
			{ id: 'siswa', label: 'Akun Siswa' },
			{ id: 'ortu', label: 'Akun Orang Tua' },
			{ id: 'admin', label: 'Admin & Hak Akses' },
			{ id: 'generate', label: 'Generate Akun' },
			{ id: 'audit', label: 'Audit & Permintaan' }
		] as tab}
			<button role="tab" class="tab {activeTab === tab.id ? 'tab-active font-semibold' : ''}" onclick={() => setTab(tab.id as AccountTab)}>{tab.label}</button>
		{/each}
	</div>

	{#if createPanelOpen}
		<Card.Root class="border-primary/30 shadow-sm">
			<Card.Header class="flex flex-row items-start justify-between gap-3">
				<div><Card.Title>Tambah Akun Workflow</Card.Title><p class="mt-1 text-sm text-base-content/70">Step {createStep}/4 — pilih jenis akun, profil resmi, role, lalu credential.</p></div>
				<Button variant="outline" onclick={() => (createPanelOpen = false)}>Tutup</Button>
			</Card.Header>
			<Card.Content class="space-y-5">
				<div class="grid gap-3 md:grid-cols-4">
					{#each [
						{ id: 'employee', title: 'Pegawai', desc: 'Guru/staf/kesiswaan' },
						{ id: 'student', title: 'Siswa', desc: 'Akun portal siswa' },
						{ id: 'parent', title: 'Orang Tua', desc: 'Akun wali/orang tua' },
						{ id: 'admin', title: 'Admin', desc: 'Akun internal sensitif' }
					] as item}
						<button class={`rounded-2xl border p-4 text-left ${accountType === item.id ? 'border-primary bg-primary/10' : 'border-base-300 bg-base-100 hover:bg-base-200/40'}`} onclick={() => setAccountType(item.id as AccountType)}>
							<p class="font-semibold">{item.title}</p><p class="text-xs text-base-content/70">{item.desc}</p>
						</button>
					{/each}
				</div>

				<div class="grid gap-5 lg:grid-cols-[1fr_1.1fr]">
					<div class="space-y-4 rounded-2xl border bg-base-200/20 p-4">
						<p class="font-semibold">Profil Resmi</p>
						{#if candidateMode === 'none'}
							<p class="text-sm text-base-content/70">Akun admin murni boleh tanpa profil, tetapi sebaiknya tetap memakai nama tampil yang jelas.</p>
						{:else}
						{#if candidateRequiresClass}
							<p class="text-xs text-base-content/70">
								{candidateMode === 'student'
									? 'Siswa per kelas — Pilih kelas terlebih dahulu untuk menarik siswa'
									: 'Ortu per kelas anak — Pilih kelas anak untuk menarik orang tua/wali terkait'}
							</p>
							<select class="w-full rounded-md border bg-base-200 px-3 py-2 text-sm" bind:value={candidateClassId} onchange={() => { clearProfileSelection(); profileCandidates = []; }}>
									<option value="">Pilih rombel</option>
									{#each rombels.filter((item) => item.is_active !== false) as kelas}
										<option value={kelas.id}>{kelas.name || kelas.code} {kelas.total_students ? `(${kelas.total_students} siswa)` : ''}</option>
									{/each}
								</select>
							{/if}
							<div class="flex gap-2"><Input bind:value={candidateSearch} placeholder="Cari nama/NIP/NISN/orang tua" /><Button variant="outline" onclick={() => void loadProfileCandidates()} disabled={!candidateCanLoad || profileCandidatesBusy}>{profileCandidatesBusy ? 'Memuat…' : 'Tarik Data'}</Button></div>
							<label class="flex items-center gap-2 text-xs text-base-content/70"><input type="checkbox" bind:checked={includeLinkedCandidates} /> Tampilkan yang sudah tertaut</label>
							{#if profileCandidatesError}<p class="text-xs text-destructive">{profileCandidatesError}</p>{/if}
							<div class="max-h-72 space-y-2 overflow-auto">
								{#each profileCandidates as candidate}
									<button class={`w-full rounded-xl border p-3 text-left text-sm ${selectedCandidateId === candidate.id ? 'border-primary bg-primary/10' : 'bg-base-100 hover:bg-base-200/50'}`} onclick={() => selectProfileCandidate(candidate)}>
										<div class="flex items-center justify-between gap-2"><span class="font-medium">{candidateLabel(candidate)}</span><Badge variant={candidate.is_linked ? 'outline' : 'secondary'}>{candidate.is_linked ? 'Sudah punya akun' : 'Belum punya akun'}</Badge></div>
										<p class="mt-1 text-xs text-base-content/70">{candidateDescription(candidate)}</p>
									</button>
								{:else}
									<EmptyStatePanel compact title="Belum ada kandidat" description="Pilih scope lalu klik Tarik Data." />
								{/each}
							</div>
						{/if}
					</div>
					<div class="space-y-4 rounded-2xl border bg-base-100 p-4">
						<div><label for="create-username" class="mb-1 block text-xs text-base-content/70">Username</label><Input id="create-username" bind:value={fUsername} placeholder="username unik" /></div>
						<div><label for="create-display-name" class="mb-1 block text-xs text-base-content/70">Nama Tampil</label><Input id="create-display-name" bind:value={fDisplayName} placeholder="nama tampil" /></div>
						<div><label for="create-password" class="mb-1 block text-xs text-base-content/70">Password Awal</label><PasswordInput id="create-password" bind:value={fPassword} placeholder="minimal 8 karakter" /></div>
						<div>
							<p class="mb-2 text-xs text-base-content/70">Role Akses</p>
							<div class="flex flex-wrap gap-2">
								{#each availableRoles as role}
									<button class={`rounded-full border px-3 py-1 text-xs ${fRoles.includes(role.value) ? 'border-success bg-success text-background' : 'bg-base-100 text-base-content/70'}`} onclick={() => toggleDraftRole(role.value)}>{role.label}</button>
								{/each}
							</div>
							{#if fRoles.includes('admin')}<p class="mt-2 text-xs text-destructive">Role admin adalah akses sensitif. Gunakan terbatas.</p>{/if}
						</div>
						<div class="flex flex-wrap gap-2"><LoadingButton loading={fBusy} onclick={() => void createUser()} disabled={fBusy || !fUsername || !fPassword || fRoles.length === 0 || profileLinkMissing}>Simpan Akun</LoadingButton><Button variant="outline" onclick={() => resetCreateDraft()}>Reset Form</Button></div>
						{#if selectedCandidate}<p class="text-xs text-success">Profil terpilih: {selectedCandidate.nama}</p>{/if}
					</div>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	{#if activeTab === 'generate'}
		<Card.Root>
			<Card.Header>
				<div class="flex flex-wrap items-start justify-between gap-3">
					<div><Card.Title>Generate Akun Massal</Card.Title><p class="mt-1 text-sm text-base-content/70">Preview dulu, cek status, lalu generate dan download CSV credential hasil generate.</p></div>
					<div class="flex flex-wrap gap-2"><Button variant="outline" onclick={() => void previewGeneration()} disabled={generationBusy !== null}>Preview</Button><Button onclick={() => void runGeneration()} disabled={generationBusy !== null || !currentGeneration || currentGeneration.ready === 0}>Generate Akun</Button><Button variant="outline" onclick={downloadGenerationCSV} disabled={!currentGeneration}>Download CSV</Button></div>
				</div>
			</Card.Header>
			<Card.Content class="space-y-4">
				<div class="flex flex-wrap gap-2">
					{#each [{ id: 'employee', label: 'Pegawai' }, { id: 'student', label: 'Siswa' }, { id: 'parent', label: 'Orang Tua' }] as item}
						<button class={`rounded-xl border px-3 py-2 text-sm ${generationAudience === item.id ? 'border-primary bg-primary/10 text-primary' : 'bg-base-100 text-base-content/70'}`} onclick={() => { generationAudience = item.id as GenerationAudience; generationPage = 1; }}>{item.label}</button>
					{/each}
				</div>
				<div class="grid gap-3 md:grid-cols-5">
					{#each ['total', 'ready', 'created', 'skipped', 'failed'] as key}
						<div class="rounded-xl border bg-base-100 p-3"><p class="text-[11px] uppercase text-base-content/70">{key}</p><p class="text-xl font-semibold">{currentGeneration?.[key as 'total' | 'ready' | 'created' | 'skipped' | 'failed'] ?? 0}</p></div>
					{/each}
				</div>
				<Input value={generationSearch} oninput={handleGenerationSearchInput} placeholder="Cari hasil preview/generate" />
				<div class="max-h-[520px] overflow-auto rounded-2xl border">
					<Table.Root>
						<Table.Header><Table.Row class="bg-base-200/50"><Table.Head>Nama</Table.Head><Table.Head>Username</Table.Head><Table.Head>Password Awal</Table.Head><Table.Head>Status</Table.Head><Table.Head>Keterangan</Table.Head></Table.Row></Table.Header>
						<Table.Body>
							{#each paginatedGenerationRows as row}
								<Table.Row><Table.Cell class="font-medium">{String(row.nama ?? '')}</Table.Cell><Table.Cell class="font-mono text-xs">{String(row.username ?? row.generated_username ?? '—')}</Table.Cell><Table.Cell class="font-mono text-xs">{String(row.password ?? row.temporary_password ?? 'ditampilkan setelah generate')}</Table.Cell><Table.Cell><Badge variant={row.status === 'ready' || row.status === 'created' ? 'secondary' : row.status === 'failed' ? 'destructive' : 'outline'}>{String(row.status ?? '—')}</Badge></Table.Cell><Table.Cell class="text-xs text-base-content/70">{String(row.message ?? row.reason ?? '')}</Table.Cell></Table.Row>
							{:else}
								<Table.Row><Table.Cell colspan={5}><EmptyStatePanel compact title="Belum ada preview" description="Klik Preview untuk memuat kandidat akun." /></Table.Cell></Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</div>
				<TablePagination
					page={safeGenerationPage}
					limit={generationPageSize}
					total={visibleGenerationRows.length}
					itemLabel="kandidat"
					ariaLabel="Navigasi halaman kandidat generate akun"
					onchange={handleGenerationPagination}
				/>
			</Card.Content>
		</Card.Root>
	{:else if activeTab === 'audit'}
		<div class="grid gap-4 md:grid-cols-3">
			<Card.Root><Card.Header><Card.Title>Permintaan Data Resmi</Card.Title></Card.Header><Card.Content><p class="text-sm text-base-content/70">Review perubahan data akun/profil dari pengguna portal.</p><Button class="mt-4" href="/settings/user-change-requests">Buka Permintaan</Button></Card.Content></Card.Root>
			<Card.Root><Card.Header><Card.Title>Audit Logs</Card.Title></Card.Header><Card.Content><p class="text-sm text-base-content/70">Pantau aktivitas perubahan akses, reset password, dan tindakan admin.</p><Button class="mt-4" variant="outline" href="/settings/audit-logs">Buka Audit</Button></Card.Content></Card.Root>
			<Card.Root><Card.Header><Card.Title>Role & Permission</Card.Title></Card.Header><Card.Content><p class="text-sm text-base-content/70">Kelola role template dan permission matrix terpisah dari lifecycle akun.</p><Button class="mt-4" variant="outline" href="/settings/rbac">Buka RBAC</Button></Card.Content></Card.Root>
		</div>
	{:else}
		<Card.Root>
			<Card.Header class="space-y-4">
				<div class="flex flex-wrap items-start justify-between gap-3">
					<div><Card.Title>{activeTab === 'pegawai' ? 'Workflow Akun Pegawai' : activeTab === 'siswa' ? 'Workflow Akun Siswa' : activeTab === 'ortu' ? 'Workflow Akun Orang Tua' : activeTab === 'admin' ? 'Admin & Hak Akses Sensitif' : 'Daftar Semua Akun'}</Card.Title><p class="mt-1 text-sm text-base-content/70">Klik baris untuk melihat panel detail. Gunakan filter untuk menemukan akun bermasalah.</p></div>
					<div class="flex flex-wrap gap-2"><Button variant="outline" onclick={() => void refreshOverview()}>Muat Ulang</Button><Button onclick={() => openCreate(activeTab === 'siswa' ? 'student' : activeTab === 'ortu' ? 'parent' : activeTab === 'admin' ? 'admin' : 'employee')}>Tambah Sesuai Tab</Button></div>
				</div>
				<div class="grid gap-3 md:grid-cols-[1fr_180px_210px]">
					<input type="text" value={searchQuery} oninput={handleUserSearchInput} placeholder="Cari nama, username, profil, role…" class="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring" />
					<select class="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring" value={roleFilter} onchange={setRoleFilter}><option value="all">Semua role</option>{#each availableRoles as role}<option value={role.value}>{role.label}</option>{/each}</select>
					<select class="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring" value={statusFilter} onchange={setStatusFilter}><option value="all">Semua status</option><option value="active">Aktif</option><option value="inactive">Nonaktif</option><option value="unlinked">Belum tertaut profil</option><option value="never-login">Belum pernah login</option><option value="multi-role">Multi-peran</option></select>
				</div>
				{#if selectedUserIds.length > 0}
					<div class="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-primary/30 bg-primary/10 p-3"><p class="text-sm font-medium">{selectedUserIds.length} akun dipilih</p><div class="flex flex-wrap gap-2"><Button variant="outline" size="sm" onclick={() => (selectedUserIds = [])}>Bersihkan</Button><Button variant="outline" size="sm" onclick={() => (activeTab = 'generate')}>Ke Generate Akun</Button></div></div>
				{/if}
			</Card.Header>
			<Card.Content class="p-0">
				<div class="overflow-x-auto">
					<Table.Root>
						<Table.Header><Table.Row class="bg-base-200/50"><Table.Head class="w-10"><input type="checkbox" checked={allVisibleSelected} onchange={toggleSelectVisible} /></Table.Head><Table.Head>Pengguna</Table.Head><Table.Head>Jenis Profil</Table.Head><Table.Head>Role</Table.Head><Table.Head>Health</Table.Head><Table.Head>Last Login</Table.Head><Table.Head>Aksi</Table.Head></Table.Row></Table.Header>
						<Table.Body>
							{#each paginatedUsers as user}
								<Table.Row class={selectedUser?.id === user.id ? 'bg-primary/5' : ''}>
									<Table.Cell><input type="checkbox" checked={selectedUserIds.includes(user.id)} onchange={() => toggleUserSelection(user)} /></Table.Cell>
									<Table.Cell><button class="text-left" onclick={() => (selectedUser = user)}><div class="font-medium">{userDisplayLabel(user)}</div><div class="text-xs text-base-content/70">{usernameLabel(user)}</div></button></Table.Cell>
									<Table.Cell><Badge variant="outline">{profileType(user)}</Badge><div class="mt-1 text-xs text-base-content/70">{user.profile_nama || '—'}</div></Table.Cell>
									<Table.Cell><div class="flex max-w-md flex-wrap gap-1">{#each user.roles ?? [] as role}<Badge variant={role === 'admin' ? 'destructive' : 'secondary'}>{roleLabel(role)}</Badge>{/each}</div></Table.Cell>
									<Table.Cell><div class="flex max-w-sm flex-wrap gap-1">{#each accountHealth(user) as item}<Badge variant={item === 'Akses sensitif' || item === 'Nonaktif' ? 'destructive' : 'outline'}>{item}</Badge>{:else}<Badge variant="secondary">Sehat</Badge>{/each}</div></Table.Cell>
									<Table.Cell class="text-xs text-base-content/70">{formatDateTime(user.last_login_at)}</Table.Cell>
									<Table.Cell><Button variant="outline" size="sm" onclick={() => (selectedUser = user)}>Detail</Button></Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row><Table.Cell colspan={7}><div class="alert alert-info m-4 shadow-sm"><span>Tidak ada pengguna ditemukan. Ubah filter atau tambah akun baru.</span></div></Table.Cell></Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</div>
			</Card.Content>
			<div class="border-t border-base-300 p-3">
				<TablePagination
					page={safeUsersPage}
					limit={usersPageSize}
					total={visibleUsers.length}
					itemLabel="akun"
					ariaLabel="Navigasi halaman daftar pengguna"
					onchange={handleUserPagination}
				/>
			</div>
		</Card.Root>
	{/if}

	{#if selectedUser}
		{@const detailUser = selectedUser as User}
		<Card.Root class="border-primary/30">
			<Card.Header class="flex flex-row items-start justify-between gap-3">
				<div><Card.Title>Detail Akun: {userDisplayLabel(detailUser)}</Card.Title><p class="mt-1 text-sm text-base-content/70">{usernameLabel(detailUser)} · {profileType(detailUser)} · dibuat {new Date(detailUser.created_at).toLocaleDateString('id-ID')}</p></div>
				<Button variant="outline" onclick={() => (selectedUser = null)}>Tutup Detail</Button>
			</Card.Header>
			<Card.Content class="grid gap-5 lg:grid-cols-[1fr_1fr]">
				<div class="space-y-4">
					<div class="rounded-2xl border bg-base-200/20 p-4"><p class="text-sm font-semibold">Status Akun</p><div class="mt-3 flex flex-wrap gap-2">{#each accountHealth(detailUser) as item}<Badge variant={item === 'Akses sensitif' || item === 'Nonaktif' ? 'destructive' : 'outline'}>{item}</Badge>{:else}<Badge variant="secondary">Tidak ada masalah utama</Badge>{/each}</div></div>
					<div class="rounded-2xl border bg-base-100 p-4"><p class="text-sm font-semibold">Profil Terhubung</p><p class="mt-2 text-sm text-base-content/70">{detailUser.profile_nama || 'Belum ditautkan'} — {profileType(detailUser)}</p><Button class="mt-3" variant="outline" size="sm" disabled={actionBusy === `profile:${detailUser.id}`} onclick={() => void updateProfileForUser(detailUser)}>Ubah Tautan Profil</Button></div>
				</div>
				<div class="space-y-4">
					<div class="rounded-2xl border bg-base-100 p-4"><p class="text-sm font-semibold">Role Akses</p><div class="mt-3 flex flex-wrap gap-2">{#each availableRoles as role}<button class={`rounded-full border px-3 py-1 text-xs ${detailUser.roles?.includes(role.value) ? 'border-success bg-success text-background' : 'bg-base-100 text-base-content/70'}`} disabled={actionBusy === `roles:${detailUser.id}`} onclick={() => void toggleExistingUserRole(detailUser, role.value)}>{role.label}</button>{/each}</div></div>
					<div class="rounded-2xl border bg-base-100 p-4"><p class="text-sm font-semibold">Keamanan & Lifecycle</p><div class="mt-3 flex flex-wrap gap-2"><Button variant="outline" size="sm" disabled={actionBusy === `password:${detailUser.id}`} onclick={() => void resetPasswordForUser(detailUser)}>Reset Password</Button><Button variant="outline" size="sm" disabled={actionBusy === `force-password:${detailUser.id}`} onclick={() => void forcePasswordChangeForUser(detailUser)}>Wajib Ganti PW</Button><Button variant="outline" size="sm" disabled={actionBusy === `status:${detailUser.id}`} onclick={() => void toggleUserStatus(detailUser)}>{detailUser.is_active ? 'Nonaktifkan' : 'Aktifkan'}</Button><Button variant="ghost" size="sm" class="text-destructive hover:bg-destructive/10 hover:text-destructive" onclick={() => void deleteUser(detailUser)}>Hapus</Button></div></div>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}
</div>
