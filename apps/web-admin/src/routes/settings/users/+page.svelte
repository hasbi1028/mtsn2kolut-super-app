<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { confirmAction } from '$lib/confirm-dialog';
	import { readClientApiData, readClientJson } from '$lib/client/api';
	import {
		createRBACPermission,
		createRBACRole,
		fetchRBACMatrix,
		resetUserPassword,
		setRBACPermissionActive,
		setRBACRoleActive,
		updateRBACPermission,
		updateRBACRole,
		updateUserProfileLink,
		updateUserRoles,
		type RBACMatrix,
		type RBACRole
	} from '$lib/client/rbac-users';

	type User = {
		id: string; username: string; roles: string[];
		employee_id: string | null; student_id: string | null; parent_id: string | null;
		profile_nama: string | null;
		is_active: boolean;
		created_at: string;
	};
	type Employee = { id: string; nama: string; nip: string; };
	type Student = { id: string; nama: string; nis: string; };
	type Parent = { id: string; nama: string; phone: string; };

	type UsersOverview = {
		users: User[];
		employees: Employee[];
		students: Student[];
		parents: Parent[];
		rbac: RBACMatrix;
	};

	let users = $state<User[]>([]);
	let employees = $state<Employee[]>([]);
	let students = $state<Student[]>([]);
	let parents = $state<Parent[]>([]);
	let rbac = $state<RBACMatrix>({ roles: [], permissions: [], role_permissions: {} });
	let usersPromise = $state<Promise<UsersOverview> | null>(null);
	let usersRequestId = 0;
	let showForm = $state(false);

	let fUsername = $state('');
	let fPassword = $state('');
	let fRoles = $state<string[]>(['guru']);
	let fEmpId = $state('');
	let fStuId = $state('');
	let fParId = $state('');
	let fBusy = $state(false);
	let actionBusy = $state<string | null>(null);
	let formHint = $derived.by(() => {
		if (fRoles.includes('siswa')) return 'Akun siswa wajib ditautkan ke satu profil siswa.';
		if (fRoles.includes('ortu')) return 'Akun orang tua wajib ditautkan ke satu profil orang tua.';
		if (fRoles.includes('guru') || fRoles.includes('staf') || fRoles.includes('kesiswaan')) return 'Akun guru/staf/kesiswaan wajib ditautkan ke satu profil pegawai.';
		return 'Akun admin murni boleh tanpa tautan profil.';
	});

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
		employees = overview.employees;
		students = overview.students;
		parents = overview.parents;
		rbac = overview.rbac;
		return overview;
	}

	async function fetchOverview(): Promise<UsersOverview> {
		const [usersRes, employeesRes, studentsRes, parentsRes, nextRBAC] = await Promise.all([
			fetch('/api/users'),
			fetch('/api/employees'),
			fetch('/api/students'),
			fetch('/api/parents'),
			fetchRBACMatrix(),
		]);
		const [nextUsers, nextEmployees, nextStudents, nextParents] = await Promise.all([
			readClientApiData<User[]>(usersRes, 'Gagal memuat data pengguna.'),
			readClientApiData<Employee[]>(employeesRes, 'Gagal memuat data pegawai.'),
			readClientApiData<Student[]>(studentsRes, 'Gagal memuat data siswa.'),
			readClientApiData<Parent[]>(parentsRes, 'Gagal memuat data orang tua.'),
		]);
		return {
			users: nextUsers ?? [],
			employees: nextEmployees ?? [],
			students: nextStudents ?? [],
			parents: nextParents ?? [],
			rbac: nextRBAC,
		};
	}

	function load() {
		const requestId = ++usersRequestId;
		users = [];
		employees = [];
		students = [];
		parents = [];
		usersPromise = fetchOverview()
			.then((overview) => {
				if (requestId === usersRequestId) return applyOverview(overview);
				return { users, employees, students, parents, rbac };
			})
			.catch((error: unknown) => {
				if (requestId === usersRequestId) throw error;
				return { users, employees, students, parents, rbac };
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
			usersPromise = Promise.resolve({ users, employees, students, parents, rbac });
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
					roles: fRoles,
					employee_id: fEmpId || null,
					student_id: fStuId || null,
					parent_id: fParId || null,
				}),
			});
			await readClientJson<unknown>(res);
			fUsername = ''; fPassword = ''; fRoles = ['guru']; fEmpId = ''; fStuId = ''; fParId = '';
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
			fRoles = fRoles.filter(r => r !== role);
		} else {
			fRoles = [...fRoles, role];
		}
	}


	async function createRolePrompt() {
		const code = window.prompt('Kode role baru (contoh: operator_cbt):')?.trim();
		if (!code) return;
		const name = window.prompt('Nama role:', code)?.trim();
		if (!name) {
			toast.error('Nama role wajib diisi.');
			return;
		}
		const description = window.prompt('Deskripsi role:', '')?.trim() ?? '';
		actionBusy = 'rbac:create-role';
		try {
			await createRBACRole({ code, name, description });
			toast.success('Role RBAC berhasil dibuat');
			await refreshOverview();
		} catch (error) {
			toast.error(overviewErrorMessage(error));
		} finally {
			actionBusy = null;
		}
	}

	async function editRolePrompt(role: RBACRole) {
		if (role.is_system) {
			toast.error('Role sistem tidak boleh diubah dari UI.');
			return;
		}
		const name = window.prompt(`Nama role ${role.code}:`, role.name || role.code)?.trim();
		if (!name) return;
		const description = window.prompt('Deskripsi role:', role.description || '')?.trim() ?? '';
		actionBusy = `rbac:role:${role.code}`;
		try {
			await updateRBACRole(role.code, { name, description });
			toast.success('Role RBAC diperbarui');
			await refreshOverview();
		} catch (error) {
			toast.error(overviewErrorMessage(error));
		} finally {
			actionBusy = null;
		}
	}

	async function toggleRoleStatus(role: RBACRole) {
		if (role.is_system && role.is_active !== false) {
			toast.error('Role sistem tidak boleh dinonaktifkan.');
			return;
		}
		const next = role.is_active === false;
		if (!(await confirmAction({
			title: next ? 'Aktifkan Role' : 'Nonaktifkan Role',
			message: `${next ? 'Aktifkan' : 'Nonaktifkan'} role "${role.code}"? Backend akan menolak jika role masih dipakai user aktif.`,
			confirmLabel: next ? 'Aktifkan' : 'Nonaktifkan',
			tone: 'warning'
		}))) return;
		actionBusy = `rbac:role-status:${role.code}`;
		try {
			await setRBACRoleActive(role.code, next);
			toast.success(next ? 'Role diaktifkan' : 'Role dinonaktifkan');
			await refreshOverview();
		} catch (error) {
			toast.error(overviewErrorMessage(error));
		} finally {
			actionBusy = null;
		}
	}

	async function createPermissionPrompt() {
		const code = window.prompt('Kode permission baru (contoh: reports.view):')?.trim();
		if (!code) return;
		const [defaultModule = '', defaultAction = ''] = code.split('.');
		const module = window.prompt('Module permission:', defaultModule)?.trim();
		const action = window.prompt('Action permission:', defaultAction)?.trim();
		if (!module || !action) {
			toast.error('Module dan action permission wajib diisi.');
			return;
		}
		const description = window.prompt('Deskripsi permission:', '')?.trim() ?? '';
		actionBusy = 'rbac:create-permission';
		try {
			await createRBACPermission({ code, module, action, description });
			toast.success('Permission RBAC berhasil dibuat');
			await refreshOverview();
		} catch (error) {
			toast.error(overviewErrorMessage(error));
		} finally {
			actionBusy = null;
		}
	}

	async function editPermissionPrompt(permission: { code: string; module?: string; action?: string; description?: string }) {
		const module = window.prompt(`Module permission ${permission.code}:`, permission.module || permission.code.split('.')[0] || '')?.trim();
		const action = window.prompt('Action permission:', permission.action || permission.code.split('.')[1] || '')?.trim();
		if (!module || !action) return;
		const description = window.prompt('Deskripsi permission:', permission.description || '')?.trim() ?? '';
		actionBusy = `rbac:permission:${permission.code}`;
		try {
			await updateRBACPermission(permission.code, { module, action, description });
			toast.success('Permission RBAC diperbarui');
			await refreshOverview();
		} catch (error) {
			toast.error(overviewErrorMessage(error));
		} finally {
			actionBusy = null;
		}
	}

	async function togglePermissionStatus(permission: { code: string; is_active?: boolean }) {
		const next = permission.is_active === false;
		if (!(await confirmAction({
			title: next ? 'Aktifkan Permission' : 'Nonaktifkan Permission',
			message: `${next ? 'Aktifkan' : 'Nonaktifkan'} permission "${permission.code}"? Permission kritikal dilindungi backend.`,
			confirmLabel: next ? 'Aktifkan' : 'Nonaktifkan',
			tone: 'warning'
		}))) return;
		actionBusy = `rbac:permission-status:${permission.code}`;
		try {
			await setRBACPermissionActive(permission.code, next);
			toast.success(next ? 'Permission diaktifkan' : 'Permission dinonaktifkan');
			await refreshOverview();
		} catch (error) {
			toast.error(overviewErrorMessage(error));
		} finally {
			actionBusy = null;
		}
	}

	onMount(() => {
		void load();
	});
</script>

<svelte:head><title>Manajemen Pengguna — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6 p-6">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-slate-800">Manajemen Pengguna</h1>
			<p class="text-sm text-slate-500 mt-1">Kelola akun akses sistem dengan RBAC terpadu</p>
		</div>
		<Button onclick={() => (showForm = !showForm)}>
			{showForm ? 'Batal' : '+ Tambah Pengguna'}
		</Button>
	</div>

	<AsyncContent promise={usersPromise} onerror={handleOverviewRenderError}>
		{#snippet pending()}
			<div class="grid gap-3 md:grid-cols-4">
				{#each ['Total Akun', 'Akun Aktif', 'Multi-Role', 'Terhubung Profil'] as label (label)}
					<div class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-4">
						<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-slate-500">{label}</p>
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
				<div class="rounded-2xl border border-emerald-100 bg-emerald-50 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-emerald-700">Total Akun</p>
					<p class="mt-2 text-2xl font-semibold text-slate-900">{overview.users.length}</p>
					<p class="text-sm text-slate-600">akun yang sudah dapat masuk ke sistem</p>
				</div>
				<div class="rounded-2xl border border-sky-100 bg-sky-50 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-sky-700">Akun Aktif</p>
					<p class="mt-2 text-2xl font-semibold text-slate-900">{overview.users.filter((item) => item.is_active).length}</p>
					<p class="text-sm text-slate-600">akun yang saat ini masih aktif digunakan</p>
				</div>
				<div class="rounded-2xl border border-amber-100 bg-amber-50 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-amber-700">Multi-Role</p>
					<p class="mt-2 text-2xl font-semibold text-slate-900">{overview.users.filter((item) => (item.roles ?? []).length > 1).length}</p>
					<p class="text-sm text-slate-600">akun yang memegang lebih dari satu role</p>
				</div>
				<div class="rounded-2xl border border-sky-100 bg-sky-50 px-4 py-4">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-sky-700">Terhubung Profil</p>
					<p class="mt-2 text-2xl font-semibold text-slate-900">{overview.users.filter((item) => item.profile_nama).length}</p>
					<p class="text-sm text-slate-600">akun yang sudah terkait dengan entitas sekolah</p>
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
							<label for="u-name" class="text-xs text-slate-500 mb-1 block">Username</label>
							<Input id="u-name" bind:value={fUsername} placeholder="Gunakan nama akun yang mudah dikenali" />
						</div>
						<div>
							<label for="u-pass" class="text-xs text-slate-500 mb-1 block">Password</label>
							<Input id="u-pass" type="password" bind:value={fPassword} placeholder="Minimal 8 karakter" />
						</div>
						<div>
							<p class="mb-2 block text-xs text-slate-500">Peran Akses (boleh pilih lebih dari satu)</p>
							<div class="flex flex-wrap gap-2">
								{#each availableRoles as r (r.value)}
									<button
										class={`px-3 py-1 text-xs rounded-full border transition-colors ${fRoles.includes(r.value) ? 'bg-green-700 text-white border-green-700' : 'bg-white text-slate-600 border-slate-200'}`}
										onclick={() => toggleRole(r.value)}
									>
										{r.label}
									</button>
								{/each}
							</div>
							<p class="mt-2 text-xs text-slate-500">{formHint}</p>
						</div>
					</div>

					<div class="space-y-3">
						{#if fRoles.includes('guru') || fRoles.includes('staf') || fRoles.includes('kesiswaan')}
							<div>
								<label for="u-emp" class="text-xs text-slate-500 mb-1 block">Hubungkan ke Pegawai</label>
								<select id="u-emp" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fEmpId}>
									<option value="">-- Pilih Pegawai --</option>
									{#each employees as e (e.id)}
										<option value={e.id}>{e.nama} ({e.nip})</option>
									{/each}
								</select>
							</div>
						{/if}

						{#if fRoles.includes('siswa')}
							<div>
								<label for="u-stu" class="text-xs text-slate-500 mb-1 block">Hubungkan ke Siswa</label>
								<select id="u-stu" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fStuId}>
									<option value="">-- Pilih Siswa --</option>
									{#each students as s (s.id)}
										<option value={s.id}>{s.nama} ({s.nis})</option>
									{/each}
								</select>
							</div>
						{/if}

						{#if fRoles.includes('ortu')}
							<div>
								<label for="u-par" class="text-xs text-slate-500 mb-1 block">Hubungkan ke Orang Tua</label>
								<select id="u-par" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fParId}>
									<option value="">-- Pilih Orang Tua --</option>
									{#each parents as p (p.id)}
										<option value={p.id}>{p.nama} ({p.phone})</option>
									{/each}
								</select>
							</div>
						{/if}
					</div>
				</div>
				<div class="flex gap-2">
					<LoadingButton onclick={() => void createUser()} loading={fBusy} disabled={fBusy || !fUsername || !fPassword || fRoles.length === 0}>
						Simpan Pengguna
					</LoadingButton>
					<Button variant="outline" onclick={() => (showForm = false)}>Batal</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}


	<Card.Root class="border-slate-200 shadow-sm">
		<Card.Header class="flex flex-row items-start justify-between gap-3 pb-2">
			<div>
				<Card.Title class="text-base">Manajemen Role & Permission Dinamis</Card.Title>
				<p class="mt-1 text-sm text-slate-500">CRUD role/permission via RBAC backend. Role sistem dan permission kritikal tetap dilindungi backend.</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<Button variant="outline" onclick={() => void createRolePrompt()} disabled={actionBusy === 'rbac:create-role'}>+ Role</Button>
				<Button variant="outline" onclick={() => void createPermissionPrompt()} disabled={actionBusy === 'rbac:create-permission'}>+ Permission</Button>
			</div>
		</Card.Header>
		<Card.Content class="grid gap-4 lg:grid-cols-2">
			<div class="rounded-2xl border border-slate-100 p-4">
				<div class="mb-3 flex items-center justify-between">
					<h2 class="text-sm font-semibold text-slate-800">Role</h2>
					<Badge variant="secondary">{rbac.roles.length} role</Badge>
				</div>
				<div class="max-h-80 space-y-2 overflow-auto pr-1">
					{#each rbac.roles as role (role.code)}
						<div class="rounded-xl border border-slate-100 bg-slate-50 p-3">
							<div class="flex items-start justify-between gap-2">
								<div>
									<p class="font-medium text-slate-800">{role.name || role.code}</p>
									<p class="text-xs text-slate-500">{role.code}</p>
								</div>
								<div class="flex flex-wrap gap-1">
									{#if role.is_system}<Badge variant="outline">System</Badge>{/if}
									<Badge variant={role.is_active === false ? 'destructive' : 'secondary'}>{role.is_active === false ? 'Nonaktif' : 'Aktif'}</Badge>
								</div>
							</div>
							{#if role.description}<p class="mt-2 text-xs text-slate-500">{role.description}</p>{/if}
							<div class="mt-3 flex flex-wrap gap-2">
								<Button size="sm" variant="outline" onclick={() => void editRolePrompt(role)} disabled={role.is_system || actionBusy === `rbac:role:${role.code}`}>Edit</Button>
								<Button size="sm" variant="outline" onclick={() => void toggleRoleStatus(role)} disabled={role.is_system || actionBusy === `rbac:role-status:${role.code}`}>{role.is_active === false ? 'Aktifkan' : 'Nonaktifkan'}</Button>
							</div>
						</div>
					{/each}
				</div>
			</div>
			<div class="rounded-2xl border border-slate-100 p-4">
				<div class="mb-3 flex items-center justify-between">
					<h2 class="text-sm font-semibold text-slate-800">Permission</h2>
					<Badge variant="secondary">{rbac.permissions.length} permission</Badge>
				</div>
				<div class="max-h-80 space-y-2 overflow-auto pr-1">
					{#each rbac.permissions as permission (permission.code)}
						<div class="rounded-xl border border-slate-100 bg-slate-50 p-3">
							<div class="flex items-start justify-between gap-2">
								<div>
									<p class="font-medium text-slate-800">{permission.code}</p>
									<p class="text-xs text-slate-500">{permission.module ?? '-'} / {permission.action ?? '-'}</p>
								</div>
								<Badge variant={permission.is_active === false ? 'destructive' : 'secondary'}>{permission.is_active === false ? 'Nonaktif' : 'Aktif'}</Badge>
							</div>
							{#if permission.description}<p class="mt-2 text-xs text-slate-500">{permission.description}</p>{/if}
							<div class="mt-3 flex flex-wrap gap-2">
								<Button size="sm" variant="outline" onclick={() => void editPermissionPrompt(permission)} disabled={actionBusy === `rbac:permission:${permission.code}`}>Edit</Button>
								<Button size="sm" variant="outline" onclick={() => void togglePermissionStatus(permission)} disabled={actionBusy === `rbac:permission-status:${permission.code}`}>{permission.is_active === false ? 'Aktifkan' : 'Nonaktifkan'}</Button>
							</div>
						</div>
					{/each}
				</div>
			</div>
		</Card.Content>
	</Card.Root>

	<AsyncContent promise={usersPromise} onerror={handleOverviewRenderError}>
		{#snippet pending()}
			<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
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
	<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
		<Card.Content class="p-0">
				<div class="hidden overflow-x-auto lg:block">
				<Table.Root>
					<Table.Header>
						<Table.Row class="bg-slate-50">
							<Table.Head>Username</Table.Head>
							<Table.Head>Role Dinamis</Table.Head>
							<Table.Head>Profil Terhubung</Table.Head>
							<Table.Head>Status</Table.Head>
							<Table.Head>Dibuat</Table.Head>
							<Table.Head>Aksi</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each overview.users as u (u.id)}
							<Table.Row>
								<Table.Cell class="font-medium">{u.username}</Table.Cell>
								<Table.Cell>
									<div class="flex max-w-md flex-wrap gap-1.5">
										{#each availableRoles as r (r.value)}
											<button
												class={`rounded-full border px-2.5 py-1 text-[10px] font-semibold uppercase transition-colors ${u.roles?.includes(r.value) ? 'border-green-700 bg-green-700 text-white' : 'border-slate-200 bg-white text-slate-500 hover:border-green-300 hover:text-green-700'}`}
												disabled={actionBusy === `roles:${u.id}`}
												title={`Toggle role ${r.label}`}
												onclick={() => void toggleExistingUserRole(u, r.value)}
											>
												{roleLabel(r.value)}
											</button>
										{/each}
									</div>
								</Table.Cell>
								<Table.Cell class="text-sm text-slate-600">
									<div>{u.profile_nama || '—'}</div>
									<div class="mt-1 text-[11px] text-slate-400">
										{u.employee_id ? 'Pegawai' : u.student_id ? 'Siswa' : u.parent_id ? 'Orang tua' : 'Belum ditautkan'}
									</div>
								</Table.Cell>
								<Table.Cell>
									<Badge variant={u.is_active ? 'outline' : 'destructive'}>
										{u.is_active ? 'Aktif' : 'Nonaktif'}
									</Badge>
								</Table.Cell>
								<Table.Cell class="text-xs text-slate-400">{new Date(u.created_at).toLocaleDateString()}</Table.Cell>
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
											class="text-red-600 hover:text-red-700 hover:bg-red-50">Hapus</Button>
									</div>
								</Table.Cell>
							</Table.Row>
						{:else}
							<Table.Row>
								<Table.Cell colspan={6} class="p-4">
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
						<div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
							<div class="flex items-start justify-between gap-3">
								<div class="min-w-0">
									<p class="text-sm font-semibold text-slate-900">{u.username}</p>
									<div class="mt-1 flex flex-wrap gap-1">
										{#each availableRoles as r (r.value)}
											<button
												class={`rounded-full border px-2 py-0.5 text-[10px] font-semibold uppercase ${u.roles?.includes(r.value) ? 'border-green-700 bg-green-700 text-white' : 'border-slate-200 bg-white text-slate-500'}`}
												disabled={actionBusy === `roles:${u.id}`}
												onclick={() => void toggleExistingUserRole(u, r.value)}
											>
												{roleLabel(r.value)}
											</button>
										{/each}
									</div>
									<p class="mt-2 text-xs text-slate-500">{u.profile_nama || 'Tidak terhubung profil'}</p>
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
									class="flex-1 justify-center text-red-600 hover:text-red-700 hover:bg-red-50">Hapus</Button>
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
