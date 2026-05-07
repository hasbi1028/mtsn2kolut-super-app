import type { RBACPermission, RBACPermissionInput } from '$lib/client/rbac-users';
import { isCriticalPermission } from '$lib/rbac/matrix';

export type PermissionMetadataDraft = Required<Pick<RBACPermissionInput, 'module' | 'action' | 'description'>>;

export type PermissionCatalogFilter = {
	module?: string;
	status?: 'all' | 'active' | 'inactive';
	query?: string;
};

export function normalizePermissionSegment(value: string | null | undefined) {
	return (value ?? '')
		.trim()
		.toLowerCase()
		.replace(/[^a-z0-9]+/g, '_')
		.replace(/^_+|_+$/g, '')
		.replace(/_+/g, '_');
}

function inferModule(permission: RBACPermission | null | undefined) {
	return permission?.module || permission?.code?.split('.')[0] || '';
}

function inferAction(permission: RBACPermission | null | undefined) {
	return permission?.action || permission?.code?.split('.').slice(1).join('.') || '';
}

export function buildPermissionMetadataDraft(permission: RBACPermission | null | undefined): PermissionMetadataDraft {
	return {
		module: inferModule(permission),
		action: inferAction(permission),
		description: permission?.description ?? permission?.name ?? ''
	};
}

export function sanitizePermissionMetadataPayload(input: RBACPermissionInput) {
	const module = normalizePermissionSegment(input.module);
	const action = normalizePermissionSegment(input.action);
	return {
		code: module && action ? `${module}.${action}` : '',
		module,
		action,
		description: (input.description ?? '').trim()
	};
}

export function permissionMetadataChanged(permission: RBACPermission | null | undefined, draft: RBACPermissionInput) {
	if (!permission) return false;
	const normalized = sanitizePermissionMetadataPayload(draft);
	return (
		inferModule(permission) !== normalized.module ||
		inferAction(permission) !== normalized.action ||
		(permission.description ?? permission.name ?? '') !== normalized.description
	);
}

export function filterPermissionCatalog(permissions: RBACPermission[], filter: PermissionCatalogFilter = {}) {
	const module = filter.module ?? 'all';
	const status = filter.status ?? 'all';
	const query = (filter.query ?? '').trim().toLowerCase();
	return [...(permissions ?? [])]
		.filter((permission) => module === 'all' || inferModule(permission) === module)
		.filter((permission) => {
			if (status === 'active') return permission.is_active !== false;
			if (status === 'inactive') return permission.is_active === false;
			return true;
		})
		.filter((permission) => {
			if (!query) return true;
			return [permission.code, permission.name, permission.description, inferModule(permission), inferAction(permission)]
				.join(' ')
				.toLowerCase()
				.includes(query);
		})
		.sort((a, b) => a.code.localeCompare(b.code));
}

export function canTogglePermissionStatus(permission: RBACPermission | null | undefined) {
	return Boolean(permission?.code && !isCriticalPermission(permission.code));
}
