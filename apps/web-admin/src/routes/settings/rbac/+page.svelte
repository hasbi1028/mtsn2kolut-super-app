<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { fetchRBACMatrix, type RBACMatrix, type RBACRole, type RBACPermission } from '$lib/client/rbac-users';
	import { rolePermissionMap, permissionsByModule, isCriticalPermission } from '$lib/rbac/matrix';

	type RBACOverview = {
		matrix: RBACMatrix;
		rolePermissions: Record<string, string[]>;
		permissionsByModule: Record<string, RBACPermission[]>;
	};

	let overview = $state<RBACOverview | null>(null);
	let loading = $state(true);
	let errorMessage = $state('');
	let selectedRoleCode = $state('');

	const roles = $derived(overview?.matrix.roles ?? []);
	const permissions = $derived(overview?.matrix.permissions ?? []);
	const systemRoles = $derived(roles.filter((role: RBACRole) => role.is_system));
	const customRoles = $derived(roles.filter((role: RBACRole) => !role.is_system));
	const activePermissions = $derived(permissions.filter((permission: RBACPermission) => permission.is_active !== false));
	const selectedRole = $derived(roles.find((role: RBACRole) => role.code === selectedRoleCode) ?? roles[0] ?? null);
	const selectedRolePermissions = $derived(selectedRole ? (overview?.rolePermissions[selectedRole.code] ?? []) : []);

	async function loadRBACOverview() {
		loading = true;
		errorMessage = '';
		try {
			const matrix = await fetchRBACMatrix();
			const nextOverview = {
				matrix,
				rolePermissions: rolePermissionMap(matrix),
				permissionsByModule: permissionsByModule(matrix.permissions ?? [])
			};
			overview = nextOverview;
			if (!selectedRoleCode || !nextOverview.rolePermissions[selectedRoleCode]) {
				selectedRoleCode = matrix.roles?.[0]?.code ?? '';
			}
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Gagal memuat data RBAC.';
		} finally {
			loading = false;
		}
	}

	function rolePermissionCount(role: RBACRole) {
		return overview?.rolePermissions[role.code]?.length ?? 0;
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
		<Button variant="outline" onclick={() => void loadRBACOverview()} disabled={loading}>Muat Ulang</Button>
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

		<div class="grid gap-4 xl:grid-cols-[0.9fr_1.1fr]">
			<Card.Root class="border-border shadow-sm">
				<Card.Header>
					<Card.Title>Role</Card.Title>
					<p class="text-sm text-muted-foreground">Gunakan “Edit Info” untuk metadata role custom. Role sistem terkunci untuk info/status, tetapi permission dapat diatur dengan guard backend.</p>
				</Card.Header>
				<Card.Content class="space-y-3">
					{#each roles as role (role.code)}
						<button
							type="button"
							class={`w-full rounded-2xl border p-4 text-left transition-colors ${selectedRole?.code === role.code ? 'border-primary bg-primary/5' : 'border-border bg-card hover:bg-muted/50'}`}
							onclick={() => (selectedRoleCode = role.code)}
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
							<p class="mt-1 text-sm text-muted-foreground">Preview awal permission untuk role terpilih. Tahap berikutnya akan menambahkan checkbox editor, diff preview, dan dialog konfirmasi.</p>
						</div>
						{#if selectedRole}<Badge variant="secondary">{selectedRole.name || selectedRole.code}</Badge>{/if}
					</div>
				</Card.Header>
				<Card.Content class="space-y-5">
					{#if selectedRole}
						<div class="rounded-2xl border border-warning/30 bg-warning/10 p-4 text-sm text-warning-foreground">
							Perubahan role-permission akan memaksa user terdampak login ulang melalui session/auth-version invalidation di backend. Guard permission kritikal tetap menjadi sumber kebenaran backend.
						</div>
						<div class="grid gap-3 md:grid-cols-2">
							<div class="rounded-2xl border border-border p-4">
								<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Permission Role Ini</p>
								<p class="mt-2 text-2xl font-bold text-foreground">{selectedRolePermissions.length}</p>
							</div>
							<div class="rounded-2xl border border-border p-4">
								<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Permission Aktif</p>
								<p class="mt-2 text-2xl font-bold text-foreground">{activePermissions.length}</p>
							</div>
						</div>
						<div class="space-y-4">
							{#each Object.entries(overview.permissionsByModule) as [module, modulePermissions] (module)}
								<div class="rounded-2xl border border-border bg-card p-4">
									<div class="mb-3 flex items-center justify-between gap-2">
										<h2 class="text-sm font-semibold text-foreground">{module}</h2>
										<Badge variant="outline">{modulePermissions.length} permission</Badge>
									</div>
									<div class="flex flex-wrap gap-2">
										{#each modulePermissions as permission (permission.code)}
											{@const enabled = selectedRolePermissions.includes(permission.code)}
											<span class={`rounded-full border px-3 py-1 text-xs font-medium ${enabled ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-muted text-muted-foreground'}`} title={permission.description ?? permission.code}>
												{permission.code}{isCriticalPermission(permission.code) ? ' ⚠' : ''}
											</span>
										{/each}
									</div>
								</div>
							{/each}
						</div>
					{:else}
						<EmptyStatePanel title="Pilih role" description="Pilih salah satu role untuk melihat permission yang terpasang." compact />
					{/if}
				</Card.Content>
			</Card.Root>
		</div>
	{/if}
</section>
