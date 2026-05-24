import type { SidebarFolderItem, SidebarNavGroup, SidebarNavItem, SidebarNavNode } from './sidebar-config';
import { evaluateSidebarItemAccess } from '$lib/rbac/ui-policy';
import { isSidebarFolder } from './sidebar-tree';

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

function folderAccessAllowed(
	folder: SidebarFolderItem,
	userRoles: readonly string[] | undefined,
	userPermissions: readonly string[] | undefined
) {
	if (!folder.permissions?.length && !folder.roleFallbacks?.length && !folder.allowAuthenticatedFallback) return true;
	return evaluateSidebarItemAccess(
		{
			href: `#${folder.id}`,
			label: folder.label,
			icon: folder.icon,
			permissions: folder.permissions ?? [],
			roleFallbacks: folder.roleFallbacks,
			allowAuthenticatedFallback: folder.allowAuthenticatedFallback,
			pinnable: false
		},
		userRoles ?? [],
		userPermissions ?? []
	).allowed;
}

function filterNodeByAccess(
	node: SidebarNavNode,
	userRoles: readonly string[] | undefined,
	userPermissions: readonly string[] | undefined
): SidebarNavNode | null {
	if (!isSidebarFolder(node)) {
		return itemAllowedByAccess(node, userRoles, userPermissions) ? node : null;
	}
	if (!folderAccessAllowed(node, userRoles, userPermissions)) return null;
	const children = node.children
		.map((child) => filterNodeByAccess(child, userRoles, userPermissions))
		.filter((child): child is SidebarNavNode => !!child);
	if (children.length === 0) return null;
	return { ...node, children };
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
				.filter((item): item is SidebarNavNode => !!item)
		}))
		.filter((group) => group.items.length > 0);
}
