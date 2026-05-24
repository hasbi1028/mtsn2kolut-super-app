import type { SidebarFlatItem, SidebarNavGroup, SidebarNavItem, SidebarNavNode } from './sidebar-config';
import { matchesSidebarPath } from './sidebar-active';

export function isSidebarFolder(node: SidebarNavNode): node is Extract<SidebarNavNode, { kind: 'folder' }> {
	return node.kind === 'folder';
}

export function isSidebarLeaf(node: SidebarNavNode): node is SidebarNavItem {
	return !isSidebarFolder(node);
}

function flattenNodes(nodes: readonly SidebarNavNode[], group: string, ancestors: string[] = []): SidebarFlatItem[] {
	return nodes.flatMap((node) => {
		if (isSidebarFolder(node)) {
			return flattenNodes(node.children, group, [...ancestors, node.label]);
		}
		return [{ ...node, group, ancestors, breadcrumb: [group, ...ancestors, node.label] }];
	});
}

export function flattenSidebarNavGroups(groups: readonly SidebarNavGroup[]): SidebarFlatItem[] {
	return groups.flatMap((group) => flattenNodes(group.items, group.group));
}

export function findSidebarActiveTrail(pathname: string, groups: readonly SidebarNavGroup[]) {
	return flattenSidebarNavGroups(groups)
		.filter((item) => matchesSidebarPath(pathname, item.href))
		.sort((a, b) => b.href.length - a.href.length)[0] ?? null;
}

export function sidebarBreadcrumbLabel(item: Pick<SidebarFlatItem, 'group' | 'ancestors' | 'label'>) {
	return [item.group, ...(item.ancestors ?? []), item.label].filter(Boolean).join(' › ');
}
