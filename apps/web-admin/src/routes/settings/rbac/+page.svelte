<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import {
		createRBACRole,
		fetchRBACMatrix,
		setRBACRoleActive,
		updateRBACRole,
		updateRBACRolePermissions,
		type RBACMatrix,
		type RBACRole,
		type RBACPermission
	} from '$lib/client/rbac-users';
	import {
		diffPermissions,
		filterPermissions,
		isCriticalPermission,
		normalizePermissionDraft,
		permissionsByModule,
		rolePermissionMap
	} from '$lib/rbac/matrix';
	import {
		buildRoleMetadataDraft,
		canEditRoleMetadata,
		normalizeRoleCode,
		roleMetadataChanged,
		sanitizeRoleMetadataPayload,
		type RoleMetadataDraft
	} from '$lib/rbac/roles';

	type RBACOverview = {
		matrix: RBACMatrix;
		rolePermissions: Record<string, string[]>;
		permissionsByModule: Record<string, RBACPermission[]>;
	};

	let overview = $state<RBACOverview | null>(null);
	let loading = $state(true);
	let saving = $state(false);
	let errorMessage = $state('');
	let mutationMessage = $state('');
	let mutationError = $state('');
	let selectedRoleCode = $state('');
	let draftPermissions = $state<string[]>([]);
	let permissionSearch = $state('');
	let moduleFilter = $state('all');
	let confirmOpen = $state(false);
	let roleFormMode = $state<'create' | 'edit'>('create');
	let roleDraft = $state<RoleMetadataDraft>({ code: '', name: '', description: '' });
	let roleFormError = $state('');
	let roleActionLoading = $state(false);
	let roleStatusTarget = $state('');

	const roles = $derived(overview?.matrix.roles ?? []);
	const permissions = $derived(overview?.matrix.permissions ?? []);
	const systemRoles = $derived(roles.filter((role: RBACRole) => role.is_system));
	const customRoles = $derived(roles.filter((role: RBACRole) => !role.is_system));
	const activePermissions = $derived(permissions.filter((permission: RBACPermission) => permission.is_active !== false));
	const selectedRole = $derived(roles.find((role: RBACRole) => role.code === selectedRoleCode) ?? roles[0] ?? null);
	const selectedRolePermissions = $derived(selectedRole ? (overview?.rolePermissions[selectedRole.code] ?? []) : []);
	const normalizedDraftPermissions = $derived(normalizePermissionDraft(draftPermissions));
	const permissionDiff = $derived(diffPermissions(selectedRolePermissions, normalizedDraftPermissions));
	const hasPermissionChanges = $derived(permissionDiff.added.length > 0 || permissionDiff.removed.length > 0);
	const criticalChangedPermissions = $derived(
		[...permissionDiff.added, ...permissionDiff.removed].filter((permission) => isCriticalPermission(permission))
	);
	const visiblePermissionsByModule = $derived(
		permissionsByModule(filterPermissions(activePermissions, { module: moduleFilter, query: permissionSearch }))
	);
	const moduleOptions = $derived(Object.keys(overview?.permissionsByModule ?? {}));
	const roleDraftPayload = $derived(sanitizeRoleMetadataPayload(roleDraft));
	const canSaveRoleDraft = $derived(
		roleDraftPayload.code.length > 0 &&
		roleDraftPayload.name.length > 0 &&
		(roleFormMode === 'create' || (canEditRoleMetadata(selectedRole) && roleMetadataChanged(selectedRole, roleDraft)))
	);

	function buildOverview(matrix: RBACMatrix): RBACOverview {
		return {
			matrix,
			rolePermissions: rolePermissionMap(matrix),
			permissionsByModule: permissionsByModule(matrix.permissions ?? [])
		};
	}

	function setDraftFromRole(roleCode: string, rolePermissions = overview?.rolePermissions) {
		draftPermissions = normalizePermissionDraft(rolePermissions?.[roleCode] ?? []);
	}

	async function loadRBACOverview() {
		loading = true;
		errorMessage = '';
		try {
			const matrix = await fetchRBACMatrix();
			const nextOverview = buildOverview(matrix);
			overview = nextOverview;
			const nextRoleCode = selectedRoleCode && nextOverview.rolePermissions[selectedRoleCode]
				? selectedRoleCode
				: (matrix.roles?.[0]?.code ?? '');
			selectedRoleCode = nextRoleCode;
			setDraftFromRole(nextRoleCode, nextOverview.rolePermissions);
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Gagal memuat data RBAC.';
		} finally {
			loading = false;
		}
	}

	function rolePermissionCount(role: RBACRole) {
		return overview?.rolePermissions[role.code]?.length ?? 0;
	}

	function selectRole(role: RBACRole) {
		selectedRoleCode = role.code;
		setDraftFromRole(role.code);
		mutationMessage = '';
		mutationError = '';
		roleFormError = '';
		confirmOpen = false;
		roleFormMode = canEditRoleMetadata(role) ? 'edit' : 'create';
		roleDraft = canEditRoleMetadata(role) ? buildRoleMetadataDraft(role) : { code: '', name: '', description: '' };
	}

	function startCreateRole() {
		roleFormMode = 'create';
		roleDraft = { code: '', name: '', description: '' };
		roleFormError = '';
		mutationMessage = '';
		mutationError = '';
	}

	function startEditRole(role: RBACRole) {
		if (!canEditRoleMetadata(role)) {
			roleFormError = 'Role sistem tidak dapat diedit metadata/statusnya. Permission tetap dapat diatur melalui matrix dengan guard backend.';
			return;
		}
		selectedRoleCode = role.code;
		setDraftFromRole(role.code);
		roleFormMode = 'edit';
		roleDraft = buildRoleMetadataDraft(role);
		roleFormError = '';
		mutationMessage = '';
		mutationError = '';
	}

	function updateRoleCodeDraft(value: string) {
		roleDraft = { ...roleDraft, code: normalizeRoleCode(value) };
	}

	async function saveRoleMetadata() {
		roleFormError = '';
		mutationMessage = '';
		mutationError = '';
		const payload = sanitizeRoleMetadataPayload(roleDraft);
		if (!payload.code || !payload.name) {
			roleFormError = 'Kode role dan nama role wajib diisi.';
			return;
		}
		if (roleFormMode === 'edit' && !canEditRoleMetadata(selectedRole)) {
			roleFormError = 'Role sistem tidak dapat diedit metadata/statusnya.';
			return;
		}
		roleActionLoading = true;
		try {
			if (roleFormMode === 'create') {
				await createRBACRole(payload);
				selectedRoleCode = payload.code;
				mutationMessage = `Role ${payload.name} berhasil dibuat. Atur permission role pada matrix akses sebelum digunakan.`;
			} else if (selectedRole) {
				await updateRBACRole(selectedRole.code, { name: payload.name, description: payload.description });
				mutationMessage = `Metadata role ${payload.name} berhasil diperbarui.`;
			}
			await loadRBACOverview();
			roleFormMode = 'edit';
			const refreshed = roles.find((role: RBACRole) => role.code === selectedRoleCode);
			roleDraft = buildRoleMetadataDraft(refreshed ?? selectedRole);
		} catch (error) {
			roleFormError = error instanceof Error ? error.message : 'Gagal menyimpan metadata role.';
		} finally {
			roleActionLoading = false;
		}
	}

	async function toggleRoleActive(role: RBACRole) {
		if (!canEditRoleMetadata(role)) {
			roleFormError = 'Role sistem tidak dapat dinonaktifkan dari UI.';
			return;
		}
		roleFormError = '';
		mutationMessage = '';
		mutationError = '';
		roleStatusTarget = role.code;
		try {
			await setRBACRoleActive(role.code, role.is_active === false);
			mutationMessage = `Status role ${role.name || role.code} berhasil diperbarui.`;
			selectedRoleCode = role.code;
			await loadRBACOverview();
		} catch (error) {
			roleFormError = error instanceof Error ? error.message : 'Gagal memperbarui status role.';
		} finally {
			roleStatusTarget = '';
		}
	}

	function hasDraftPermission(code: string) {
		return normalizedDraftPermissions.includes(code);
	}

	function togglePermission(code: string, checked: boolean) {
		const current = new Set(normalizedDraftPermissions);
		if (checked) current.add(code);
		else current.delete(code);
		draftPermissions = normalizePermissionDraft(Array.from(current));
		mutationMessage = '';
		mutationError = '';
	}

	function resetPermissionDraft() {
		if (!selectedRole) return;
		setDraftFromRole(selectedRole.code);
		mutationError = '';
		mutationMessage = 'Perubahan draft dikembalikan ke permission aktif saat ini.';
	}

	async function confirmSaveRolePermissions() {
		if (!selectedRole || !hasPermissionChanges) return;
		saving = true;
		mutationError = '';
		mutationMessage = '';
		try {
			const roleCode = selectedRole.code;
			await updateRBACRolePermissions(roleCode, normalizedDraftPermissions);
			mutationMessage = `Permission role ${selectedRole.name || roleCode} berhasil disimpan. User dengan role ini akan diminta login ulang sesuai guard backend.`;
			confirmOpen = false;
			selectedRoleCode = roleCode;
			await loadRBACOverview();
		} catch (error) {
			mutationError = error instanceof Error ? error.message : 'Gagal menyimpan permission role.';
		} finally {
			saving = false;
		}
	}

	onMount(() => {
		void loadRBACOverview();
	});
</script>

<svelte:head>
	<title>Manajemen RBAC • MTsN 2 Kolaka Utara</title>
</svelte:head>

<section class="space-y-6 p-4 md:p-6">
	<div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
		<div>
			<p class="text-sm font-semibold uppercase tracking-wide text-primary">Pengaturan Sistem</p>
			<h1 class="text-2xl font-bold text-foreground md:text-3xl">Manajemen RBAC</h1>
			<p class="mt-2 max-w-3xl text-sm text-muted-foreground">
				Kelola role, permission, dan matrix hak akses secara dinamis. Halaman ini memisahkan “Edit Info” role dari “Atur Permission” agar perubahan akses lebih jelas dan aman.
			</p>
		</div>
		<Button variant="outline" onclick={() => void loadRBACOverview()} disabled={loading || saving}>Muat Ulang</Button>
	</div>

	{#if loading}
		<div class="grid gap-4 md:grid-cols-4">
			{#each Array.from({ length: 4 }) as _, index (`rbac-summary-${index}`)}
				<Card.Root class="border-border shadow-sm">
					<Card.Content class="space-y-3 p-5">
						<Skeleton class="h-4 w-24" />
						<Skeleton class="h-8 w-16" />
					</Card.Content>
				</Card.Root>
			{/each}
		</div>
		<Card.Root class="border-border shadow-sm">
			<Card.Content class="space-y-3 p-5">
				<Skeleton class="h-5 w-44" />
				<Skeleton class="h-32 w-full" />
			</Card.Content>
		</Card.Root>
	{:else if errorMessage}
		<RecoveryPanel
			title="RBAC belum dapat dimuat"
			message={errorMessage}
			actionLabel="Coba Lagi"
			onRetry={() => void loadRBACOverview()}
		/>
	{:else if overview}
		<div class="grid gap-4 md:grid-cols-4">
			<Card.Root class="border-border shadow-sm">
				<Card.Content class="p-5">
					<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Total Role</p>
					<p class="mt-2 text-3xl font-bold text-foreground">{roles.length}</p>
				</Card.Content>
			</Card.Root>
			<Card.Root class="border-border shadow-sm">
				<Card.Content class="p-5">
					<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Total Permission</p>
					<p class="mt-2 text-3xl font-bold text-foreground">{permissions.length}</p>
				</Card.Content>
			</Card.Root>
			<Card.Root class="border-border shadow-sm">
				<Card.Content class="p-5">
					<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Role Sistem</p>
					<p class="mt-2 text-3xl font-bold text-foreground">{systemRoles.length}</p>
				</Card.Content>
			</Card.Root>
			<Card.Root class="border-border shadow-sm">
				<Card.Content class="p-5">
					<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Role Custom</p>
					<p class="mt-2 text-3xl font-bold text-foreground">{customRoles.length}</p>
				</Card.Content>
			</Card.Root>
		</div>

		{#if mutationMessage}
			<div class="rounded-2xl border border-success/30 bg-success/10 p-4 text-sm text-success-foreground">{mutationMessage}</div>
		{/if}
		{#if mutationError}
			<div class="rounded-2xl border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive">{mutationError}</div>
		{/if}

		<div class="grid gap-4 xl:grid-cols-[0.85fr_1.15fr]">
			<Card.Root class="border-border shadow-sm">
				<Card.Header>
					<Card.Title>Role</Card.Title>
					<p class="text-sm text-muted-foreground">Gunakan “Edit Info” untuk metadata role custom. Role sistem terkunci untuk info/status, tetapi permission dapat diatur dengan guard backend.</p>
				</Card.Header>
				<Card.Content class="space-y-4">
					<div class="rounded-2xl border border-border bg-muted/20 p-4">
						<div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
							<div>
								<p class="text-sm font-semibold text-foreground">{roleFormMode === 'create' ? 'Tambah Role Custom' : 'Edit Info Role Custom'}</p>
								<p class="text-xs text-muted-foreground">Role sistem tidak bisa diedit metadata/statusnya dari UI; backend tetap menjadi guard utama.</p>
							</div>
							<Button variant="outline" onclick={startCreateRole} disabled={roleActionLoading}>Role Baru</Button>
						</div>
						<div class="mt-4 grid gap-3 md:grid-cols-2">
							<div class="space-y-1">
								<label for="role-code" class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Kode Role</label>
								<Input id="role-code" value={roleDraft.code} oninput={(event) => updateRoleCodeDraft(event.currentTarget.value)} placeholder="operator_asesmen" disabled={roleFormMode === 'edit' || roleActionLoading} />
							</div>
							<div class="space-y-1">
								<label for="role-name" class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Nama Role</label>
								<Input id="role-name" bind:value={roleDraft.name} placeholder="Operator Asesmen" disabled={roleActionLoading} />
							</div>
							<div class="space-y-1 md:col-span-2">
								<label for="role-description" class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Deskripsi</label>
								<Input id="role-description" bind:value={roleDraft.description} placeholder="Ringkasan kewenangan role" disabled={roleActionLoading} />
							</div>
						</div>
						{#if roleFormError}
							<div class="mt-3 rounded-xl border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">{roleFormError}</div>
						{/if}
						<div class="mt-4 flex flex-wrap items-center justify-between gap-2">
							<p class="text-xs text-muted-foreground">Kode disimpan dalam format slug snake_case. Permission role diatur terpisah pada matrix akses.</p>
							<Button onclick={() => void saveRoleMetadata()} disabled={!canSaveRoleDraft || roleActionLoading}>{roleActionLoading ? 'Menyimpan…' : (roleFormMode === 'create' ? 'Buat Role' : 'Simpan Info Role')}</Button>
						</div>
					</div>
					{#each roles as role (role.code)}
						<button
							type="button"
							class={`w-full rounded-2xl border p-4 text-left transition-colors ${selectedRole?.code === role.code ? 'border-primary bg-primary/5' : 'border-border bg-card hover:bg-muted/50'}`}
							onclick={() => selectRole(role)}
						>
							<div class="flex items-start justify-between gap-3">
								<div>
									<p class="font-semibold text-foreground">{role.name || role.code}</p>
									<p class="text-xs text-muted-foreground">{role.code}</p>
								</div>
								<div class="flex flex-wrap justify-end gap-1">
									{#if role.is_system}<Badge variant="outline">System</Badge>{/if}
									<Badge variant={role.is_active === false ? 'destructive' : 'secondary'}>{role.is_active === false ? 'Nonaktif' : 'Aktif'}</Badge>
								</div>
							</div>
							{#if role.description}<p class="mt-2 text-sm text-muted-foreground">{role.description}</p>{/if}
							<p class="mt-3 text-xs font-medium text-muted-foreground">{rolePermissionCount(role)} permission aktif/terpasang</p>
						</button>
						<div class="flex flex-wrap justify-end gap-2 rounded-2xl border border-border/70 bg-muted/20 px-3 py-2">
							<Button size="sm" variant="outline" onclick={() => startEditRole(role)} disabled={!canEditRoleMetadata(role) || roleActionLoading || roleStatusTarget === role.code}>Edit Info</Button>
							<Button size="sm" variant={role.is_active === false ? 'secondary' : 'outline'} onclick={() => void toggleRoleActive(role)} disabled={!canEditRoleMetadata(role) || roleActionLoading || roleStatusTarget === role.code}>
								{roleStatusTarget === role.code ? 'Memproses…' : (role.is_active === false ? 'Aktifkan' : 'Nonaktifkan')}
							</Button>
						</div>
					{:else}
						<EmptyStatePanel title="Belum ada role" description="Matrix RBAC belum mengembalikan data role." compact />
					{/each}
				</Card.Content>
			</Card.Root>

			<Card.Root class="border-border shadow-sm">
				<Card.Header>
					<div class="flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
						<div>
							<Card.Title>Matrix Akses</Card.Title>
							<p class="mt-1 text-sm text-muted-foreground">Pilih role, centang permission yang sesuai, lalu simpan setelah melihat diff dan warning dampak.</p>
						</div>
						{#if selectedRole}<Badge variant="secondary">{selectedRole.name || selectedRole.code}</Badge>{/if}
					</div>
				</Card.Header>
				<Card.Content class="space-y-5">
					{#if selectedRole}
						<div class="rounded-2xl border border-warning/30 bg-warning/10 p-4 text-sm text-warning-foreground">
							Perubahan role-permission akan memaksa user terdampak login ulang melalui session/auth-version invalidation di backend. Guard permission kritikal tetap menjadi sumber kebenaran backend.
						</div>

						<div class="grid gap-3 md:grid-cols-3">
							<div class="rounded-2xl border border-border p-4">
								<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Permission Tersimpan</p>
								<p class="mt-2 text-2xl font-bold text-foreground">{selectedRolePermissions.length}</p>
							</div>
							<div class="rounded-2xl border border-border p-4">
								<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Draft Dipilih</p>
								<p class="mt-2 text-2xl font-bold text-foreground">{normalizedDraftPermissions.length}</p>
							</div>
							<div class="rounded-2xl border border-border p-4">
								<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Diff</p>
								<p class="mt-2 text-2xl font-bold text-foreground">+{permissionDiff.added.length} / -{permissionDiff.removed.length}</p>
							</div>
						</div>

						<div class="grid gap-3 md:grid-cols-[1fr_220px]">
							<Input bind:value={permissionSearch} placeholder="Cari permission, contoh: bank_soal atau pengguna" />
							<select bind:value={moduleFilter} class="h-10 rounded-md border border-input bg-background px-3 text-sm text-foreground shadow-xs focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]">
								<option value="all">Semua module</option>
								{#each moduleOptions as module (module)}
									<option value={module}>{module}</option>
								{/each}
							</select>
						</div>

						<div class="space-y-4">
							{#each Object.entries(visiblePermissionsByModule) as [module, modulePermissions] (module)}
								<div class="rounded-2xl border border-border bg-card p-4">
									<div class="mb-3 flex items-center justify-between gap-2">
										<h2 class="text-sm font-semibold text-foreground">{module}</h2>
										<Badge variant="outline">{modulePermissions.length} permission</Badge>
									</div>
									<div class="grid gap-2 md:grid-cols-2">
										{#each modulePermissions as permission (permission.code)}
											{@const enabled = hasDraftPermission(permission.code)}
											<label class={`flex cursor-pointer items-start gap-3 rounded-2xl border p-3 transition-colors ${enabled ? 'border-primary bg-primary/5' : 'border-border bg-muted/30 hover:bg-muted/60'}`}>
												<input
													type="checkbox"
													class="mt-1 h-4 w-4 rounded border-border accent-primary"
													checked={enabled}
													disabled={saving}
													onchange={(event) => togglePermission(permission.code, event.currentTarget.checked)}
												/>
												<span class="min-w-0 flex-1">
													<span class="flex flex-wrap items-center gap-2 text-sm font-semibold text-foreground">
														{permission.code}
														{#if isCriticalPermission(permission.code)}<Badge variant="destructive">Kritikal</Badge>{/if}
													</span>
													{#if permission.name || permission.description}
														<span class="mt-1 block text-xs text-muted-foreground">{permission.name || permission.description}</span>
													{/if}
												</span>
											</label>
										{/each}
									</div>
								</div>
							{:else}
								<EmptyStatePanel title="Permission tidak ditemukan" description="Ubah kata kunci pencarian atau filter module." compact />
							{/each}
						</div>

						<div class="rounded-2xl border border-border bg-muted/30 p-4">
							<div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
								<div>
									<p class="text-sm font-semibold text-foreground">Preview Perubahan</p>
									<p class="text-xs text-muted-foreground">Ditambah: {permissionDiff.added.length} · Dicabut: {permissionDiff.removed.length} · Critical: {criticalChangedPermissions.length}</p>
								</div>
								<div class="flex flex-wrap gap-2">
									<Button variant="outline" onclick={resetPermissionDraft} disabled={!hasPermissionChanges || saving}>Reset Perubahan</Button>
									<Button onclick={() => (confirmOpen = true)} disabled={!hasPermissionChanges || saving}>Simpan Permission Role</Button>
								</div>
							</div>
							{#if hasPermissionChanges}
								<div class="mt-4 grid gap-3 md:grid-cols-2">
									<div>
										<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Akan Ditambah</p>
										<div class="mt-2 flex flex-wrap gap-2">
											{#each permissionDiff.added as code (code)}<Badge variant={isCriticalPermission(code) ? 'destructive' : 'secondary'}>{code}</Badge>{:else}<span class="text-xs text-muted-foreground">Tidak ada.</span>{/each}
										</div>
									</div>
									<div>
										<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Akan Dicabut</p>
										<div class="mt-2 flex flex-wrap gap-2">
											{#each permissionDiff.removed as code (code)}<Badge variant={isCriticalPermission(code) ? 'destructive' : 'outline'}>{code}</Badge>{:else}<span class="text-xs text-muted-foreground">Tidak ada.</span>{/each}
										</div>
									</div>
								</div>
							{/if}
						</div>
					{:else}
						<EmptyStatePanel title="Pilih role" description="Pilih salah satu role untuk melihat dan mengatur permission." compact />
					{/if}
				</Card.Content>
			</Card.Root>
		</div>
	{/if}
</section>

{#if confirmOpen && selectedRole}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-background/80 p-4 backdrop-blur-sm" role="presentation">
		<div class="w-full max-w-2xl rounded-2xl border border-border bg-card p-5 shadow-xl">
			<div class="space-y-2">
				<p class="text-lg font-semibold text-foreground">Konfirmasi Simpan Permission Role</p>
				<p class="text-sm text-muted-foreground">
					Role target: <span class="font-semibold text-foreground">{selectedRole.name || selectedRole.code}</span>. Semua user dengan role ini akan diminta login ulang karena token/session dicabut oleh backend.
				</p>
			</div>
			{#if criticalChangedPermissions.length > 0}
				<div class="mt-4 rounded-2xl border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
					Perubahan menyentuh permission kritikal: {criticalChangedPermissions.join(', ')}. Backend tetap dapat menolak jika berisiko mengunci admin terakhir.
				</div>
			{/if}
			<div class="mt-4 grid gap-4 md:grid-cols-2">
				<div>
					<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Permission Ditambah</p>
					<div class="mt-2 flex flex-wrap gap-2">
						{#each permissionDiff.added as code (code)}<Badge variant={isCriticalPermission(code) ? 'destructive' : 'secondary'}>{code}</Badge>{:else}<span class="text-sm text-muted-foreground">Tidak ada permission ditambah.</span>{/each}
					</div>
				</div>
				<div>
					<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Permission Dicabut</p>
					<div class="mt-2 flex flex-wrap gap-2">
						{#each permissionDiff.removed as code (code)}<Badge variant={isCriticalPermission(code) ? 'destructive' : 'outline'}>{code}</Badge>{:else}<span class="text-sm text-muted-foreground">Tidak ada permission dicabut.</span>{/each}
					</div>
				</div>
			</div>
			<div class="mt-6 flex flex-col-reverse gap-2 md:flex-row md:justify-end">
				<Button variant="outline" onclick={() => (confirmOpen = false)} disabled={saving}>Batal</Button>
				<Button onclick={() => void confirmSaveRolePermissions()} disabled={saving}>{saving ? 'Menyimpan…' : 'Ya, Simpan Permission'}</Button>
			</div>
		</div>
	</div>
{/if}
