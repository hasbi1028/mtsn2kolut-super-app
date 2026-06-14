import type { SidebarFlatItem, SidebarNavGroup, SidebarNavItem } from './sidebar-config';
import { matchesSidebarPath } from './sidebar-active';

export function isSidebarLeaf(node: SidebarNavItem): node is SidebarNavItem {
	return true;
}

function withNumber(label: string, section?: string) {
	return section ? `${section} ${label}` : label;
}

function flattenNodes(
	nodes: readonly SidebarNavItem[],
	group: string,
	ancestors: string[] = [],
	groupSection?: string,
	ancestorSections: string[] = []
): SidebarFlatItem[] {
	return nodes.map((node) => {
		return {
			...node,
			group,
			groupSection,
			ancestors,
			ancestorSections,
			breadcrumb: [group, ...ancestors, node.label],
			section: node.section,
			numberedLabel: withNumber(node.label, node.section)
		};
	});
}

export function flattenSidebarNavGroups(groups: readonly SidebarNavGroup[]): SidebarFlatItem[] {
	return groups.flatMap((group) => flattenNodes(group.items, group.group, [], group.section));
}

export function findSidebarActiveTrail(pathname: string, groups: readonly SidebarNavGroup[]) {
	return flattenSidebarNavGroups(groups)
		.filter((item) => matchesSidebarPath(pathname, item.href))
		.sort((a, b) => b.href.length - a.href.length)[0] ?? null;
}

export function sidebarBreadcrumbLabel(item: Pick<SidebarFlatItem, 'group' | 'ancestors' | 'label'>) {
	return [item.group, ...(item.ancestors ?? []), item.label].filter(Boolean).join(' › ');
}

export function sidebarNumberedBreadcrumbLabel(
	item: Pick<SidebarFlatItem, 'group' | 'groupSection' | 'ancestors' | 'ancestorSections' | 'label' | 'section'>
) {
	const labels = [
		withNumber(item.group, item.groupSection),
		...(item.ancestors ?? []).map((label, index) => withNumber(label, item.ancestorSections?.[index])),
		withNumber(item.label, item.section)
	];
	return labels.filter(Boolean).join(' › ');
}

function numberNodes(nodes: readonly SidebarNavItem[], parentSection: string): SidebarNavItem[] {
	return nodes.map((node, index) => {
		const section = node.section ?? `${parentSection}.${index + 1}`;
		const numberedLabel = withNumber(node.label, section);
		return { ...node, section, numberedLabel };
	});
}

export function numberSidebarNavGroups(groups: readonly SidebarNavGroup[]): SidebarNavGroup[] {
	return groups.map((group, index) => {
		const section = group.section ?? `${index + 1}`;
		return {
			...group,
			section,
			numberedLabel: withNumber(group.group, section),
			items: numberNodes(group.items, section)
		};
	});
}

export function numberedLabel(label: string, section?: string) {
	return withNumber(label, section);
}
