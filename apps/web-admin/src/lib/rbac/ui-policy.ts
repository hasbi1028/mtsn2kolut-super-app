import { dashboardNavItem, sidebarNavGroups, type SidebarNavGroup, type SidebarNavItem } from '$lib/components/sidebar/sidebar-config';
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
	const bankSoalDraftItems: SidebarNavItem[] = [
		{ href: '/bank-soal/tambah', label: 'Tambah Soal', icon: 'file-text', permissions: ['bank_soal.create'], roles: ['admin', 'guru'] },
		{ href: '/bank-soal/mapel-kd', label: 'Mapel & KD', icon: 'layers', permissions: ['bank_soal.create'], roles: ['admin', 'guru'] }
	];
	const baseMenuItems = [
		{ ...dashboardNavItem, group: 'Akses Cepat' },
		...sidebarNavGroups.flatMap((group) => group.items.map((item) => ({ ...item, group: group.group })))
	];
	const insertAfter = baseMenuItems.findIndex((item) => item.href === '/bank-soal/daftar');
	if (insertAfter >= 0) {
		baseMenuItems.splice(insertAfter + 1, 0, ...bankSoalDraftItems.map((item) => ({ ...item, group: 'Bank Soal' })));
	} else {
		baseMenuItems.push(...bankSoalDraftItems.map((item) => ({ ...item, group: 'Bank Soal' })));
	}
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
