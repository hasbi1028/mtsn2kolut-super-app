import { describe, expect, it } from 'vitest';

import { dashboardDataAccessForUser, visibleDashboardWidgetsForUser } from './dashboard-policy';

describe('dashboard permission policy', () => {
	it('keeps student and parent portal dashboards available by portal role fallback', () => {
		expect(dashboardDataAccessForUser({ role: 'siswa', roles: ['siswa'], permissions: [] }).studentPortal).toBe(true);
		expect(dashboardDataAccessForUser({ role: 'ortu', roles: ['ortu'], permissions: [] }).parentPortal).toBe(true);
	});

	it('preserves admin operational dashboard access without enabling personal portal data fetches, even when portal permissions are present', () => {
		const adminWithPortalPermissions = {
			role: 'admin',
			roles: ['admin'],
			permissions: ['student_portal.read', 'parent_portal.read']
		};
		const access = dashboardDataAccessForUser(adminWithPortalPermissions);

		expect(access.academicStats).toBe(true);
		expect(access.studentPortal).toBe(false);
		expect(access.parentPortal).toBe(false);
		expect(visibleDashboardWidgetsForUser({ role: 'admin', roles: ['admin'], permissions: [] }).map((widget) => widget.id)).toEqual([
			'academic-overview'
		]);
	});
});
