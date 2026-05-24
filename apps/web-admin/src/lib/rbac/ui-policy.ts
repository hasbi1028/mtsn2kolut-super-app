import { dashboardNavItem, sidebarNavGroups, type SidebarNavGroup, type SidebarNavItem } from '$lib/components/sidebar/sidebar-config';
import { flattenSidebarNavGroups } from '$lib/components/sidebar/sidebar-tree';
import { DASHBOARD_WIDGETS, dashboardWidgetEvaluation, type DashboardWidgetDefinition } from './dashboard-policy';
import { evaluateUIPolicyAccess, type UIPolicyAccessEvaluation } from './access-policy';

export { evaluateUIPolicyAccess } from './access-policy';

export type MenuPolicyPreviewItem = SidebarNavItem & {
	group: string;
	evaluation: UIPolicyAccessEvaluation;
};

export type DashboardPolicyPreviewItem = DashboardWidgetDefinition & {
	evaluation: UIPolicyAccessEvaluation;
};

export type UIPolicyPreview = {
	visibleMenuItems: MenuPolicyPreviewItem[];
	hiddenMenuItems: MenuPolicyPreviewItem[];
	visibleDashboardWidgets: DashboardPolicyPreviewItem[];
	hiddenDashboardWidgets: DashboardPolicyPreviewItem[];
	menuGroups: SidebarNavGroup[];
};

export function evaluateSidebarItemAccess(item: SidebarNavItem, roles: readonly string[], permissions: readonly string[]) {
	return evaluateUIPolicyAccess({
		permissions: item.permissions,
		roleFallbacks: item.roleFallbacks,
		allowAuthenticatedFallback: item.allowAuthenticatedFallback
	}, { roles, permissions });
}

export function buildUIPolicyPreview(roleCode: string, draftPermissions: readonly string[]): UIPolicyPreview {
	const roles = roleCode ? [roleCode] : [];
	const baseMenuItems = [
		{ ...dashboardNavItem, group: 'Akses Cepat', ancestors: [], breadcrumb: ['Akses Cepat', dashboardNavItem.label] },
		...flattenSidebarNavGroups(sidebarNavGroups)
	];
	const menuItems = baseMenuItems.map((item) => ({
		...item,
		evaluation: evaluateSidebarItemAccess(item, roles, draftPermissions)
	}));
	const dashboardWidgets = DASHBOARD_WIDGETS.map((widget) => ({
		...widget,
		evaluation: dashboardWidgetEvaluation(widget, { roles, permissions: draftPermissions })
	}));
	const visibleMenuItems = menuItems.filter((item) => item.evaluation.allowed);
	const hiddenMenuItems = menuItems.filter((item) => !item.evaluation.allowed);

	return {
		visibleMenuItems,
		hiddenMenuItems,
		visibleDashboardWidgets: dashboardWidgets.filter((widget) => widget.evaluation.allowed),
		hiddenDashboardWidgets: dashboardWidgets.filter((widget) => !widget.evaluation.allowed),
		menuGroups: sidebarNavGroups
			.map((group) => ({
				...group,
				items: visibleMenuItems.filter((item) => item.group === group.group)
			}))
			.filter((group) => group.items.length > 0)
	};
}
