import type { SidebarNavItem } from '$lib/components/sidebar/sidebar-config';

export function matchesSidebarPath(pathname: string, href: string) {
	if (href === '/') return pathname === '/';
	return pathname === href || pathname.startsWith(`${href}/`);
}

export function findActiveSidebarHref(pathname: string, items: Pick<SidebarNavItem, 'href'>[]) {
	return items
		.filter((item) => matchesSidebarPath(pathname, item.href))
		.sort((a, b) => b.href.length - a.href.length)[0]?.href ?? null;
}
