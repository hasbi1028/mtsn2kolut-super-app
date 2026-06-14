import type { SidebarNavGroup, SidebarNavItem } from './sidebar-config';
import { evaluateSidebarItemAccess } from '$lib/rbac/ui-policy';

function normalize(values: readonly string[] | undefined) {
	return (values ?? []).map((value) => value.trim()).filter(Boolean);
}

export function hasAnyPermission(userPermissions: readonly string[] | undefined, required: readonly string[] | undefined) {
	const needed = normalize(required);
	if (needed.length === 0) return true;
	const owned = new Set(normalize(userPermissions));
	return needed.some((permission) => owned.has(permission));
}

export function hasAnyRole(userRoles: readonly string[] | undefined, required: readonly string[] | undefined) {
	const needed = normalize(required);
	if (needed.length === 0) return true;
	const owned = new Set(normalize(userRoles));
	return needed.some((role) => owned.has(role));
}

export function itemAllowedByAccess(
	item: SidebarNavItem,
	userRoles: readonly string[] | undefined,
	userPermissions: readonly string[] | undefined
) {
	return evaluateSidebarItemAccess(item, userRoles ?? [], userPermissions ?? []).allowed;
}

function filterNodeByAccess(
	node: SidebarNavItem,
	userRoles: readonly string[] | undefined,
	userPermissions: readonly string[] | undefined
): SidebarNavItem | null {
	return itemAllowedByAccess(node, userRoles, userPermissions) ? node : null;
}

export function filterSidebarNavGroupsByAccess(
	groups: readonly SidebarNavGroup[],
	userRoles: readonly string[] | undefined,
	userPermissions: readonly string[] | undefined
): SidebarNavGroup[] {
	return groups
		.map((group) => ({
			...group,
			items: group.items
				.map((item) => filterNodeByAccess(item, userRoles, userPermissions))
				.filter((item): item is SidebarNavItem => !!item)
		}))
		.filter((group) => group.items.length > 0);
}
