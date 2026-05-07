import type { RBACMatrix, RBACPermission } from '$lib/client/rbac-users';

export type RBACRolePermissionRow = {
	role_code?: string | null;
	permission_code?: string | null;
};

export type PermissionDiff = {
	added: string[];
	removed: string[];
};

const CRITICAL_PERMISSIONS = new Set([
	'roles.manage',
	'users.manage_roles',
	'permissions.manage',
	'rbac.manage'
]);

function sortedUnique(values: Array<string | null | undefined>) {
	return Array.from(new Set(values.map((value) => (value ?? '').trim()).filter(Boolean))).sort();
}

function inferModule(permission: Pick<RBACPermission, 'code' | 'module'>) {
	const module = (permission.module ?? '').trim();
	if (module) return module;
	return permission.code.split('.')[0] || 'lainnya';
}

export function rolePermissionMap(matrix: Pick<RBACMatrix, 'roles' | 'role_permissions'>): Record<string, string[]> {
	const result: Record<string, string[]> = {};
	for (const role of matrix.roles ?? []) {
		if (role.code) result[role.code] = [];
	}

	const rolePermissions = matrix.role_permissions ?? [];
	if (Array.isArray(rolePermissions)) {
		for (const row of rolePermissions as RBACRolePermissionRow[]) {
			const roleCode = (row.role_code ?? '').trim();
			const permissionCode = (row.permission_code ?? '').trim();
			if (!roleCode || !permissionCode) continue;
			result[roleCode] = sortedUnique([...(result[roleCode] ?? []), permissionCode]);
		}
		return result;
	}

	for (const [roleCode, permissions] of Object.entries(rolePermissions)) {
		result[roleCode] = sortedUnique(Array.isArray(permissions) ? permissions : []);
	}
	return result;
}

export function diffPermissions(before: string[], after: string[]): PermissionDiff {
	const beforeSet = new Set(sortedUnique(before));
	const afterSet = new Set(sortedUnique(after));
	return {
		added: Array.from(afterSet).filter((code) => !beforeSet.has(code)).sort(),
		removed: Array.from(beforeSet).filter((code) => !afterSet.has(code)).sort()
	};
}

export function permissionsByModule(permissions: RBACPermission[]): Record<string, RBACPermission[]> {
	const grouped: Record<string, RBACPermission[]> = {};
	for (const permission of permissions ?? []) {
		const module = inferModule(permission);
		grouped[module] = [...(grouped[module] ?? []), permission];
	}
	return Object.fromEntries(
		Object.entries(grouped)
			.sort(([a], [b]) => a.localeCompare(b))
			.map(([module, values]) => [module, [...values].sort((a, b) => a.code.localeCompare(b.code))])
	);
}

export function isCriticalPermission(code: string) {
	return CRITICAL_PERMISSIONS.has(code);
}
