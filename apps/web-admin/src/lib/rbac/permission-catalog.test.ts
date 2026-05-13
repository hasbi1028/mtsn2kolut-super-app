import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { describe, expect, it } from 'vitest';

import { dashboardNavItem, sidebarNavGroups } from '$lib/components/sidebar/sidebar-config';
import { requiredPermissionsForPath } from '$lib/server/route-access';
import { DASHBOARD_WIDGETS } from '$lib/rbac/dashboard-policy';
import { RBAC_PERMISSION_CATALOG, permissionCatalogCodes, permissionLabel } from './permission-catalog';

const repoRoot = resolve(__dirname, '../../../../..');
const migration = [
	readFileSync(resolve(repoRoot, 'services/core-api/db/migrations/069_dynamic_rbac_foundation.sql'), 'utf8'),
	readFileSync(resolve(repoRoot, 'services/core-api/db/migrations/076_profile_change_review_permission.sql'), 'utf8'),
	readFileSync(resolve(repoRoot, 'services/core-api/db/migrations/079_journal_timetable_slot_scope.sql'), 'utf8'),
	readFileSync(resolve(repoRoot, 'services/core-api/db/migrations/081_student_parent_account_portal.sql'), 'utf8'),
	readFileSync(resolve(repoRoot, 'services/core-api/db/migrations/082_employee_rbac_permissions.sql'), 'utf8'),
	readFileSync(resolve(repoRoot, 'services/core-api/db/migrations/084_internal_analytics_permissions.sql'), 'utf8'),
	readFileSync(resolve(repoRoot, 'services/core-api/db/migrations/095_backup_center_permissions.sql'), 'utf8'),
	readFileSync(resolve(repoRoot, 'services/core-api/db/migrations/096_backup_restore_plan_permission.sql'), 'utf8')
].join('\n');
const docs = readFileSync(resolve(repoRoot, 'docs/rbac-permission-catalog.md'), 'utf8');

function seededPermissionCodes() {
	return Array.from(migration.matchAll(/\('([a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)+)'\s*,\s*'([^']+)'\s*,\s*'([^']+)'\s*,\s*'([^']+)'\)/g)).map((match) => match[1]);
}

describe('RBAC permission catalog stabilization', () => {
	it('keeps the frontend catalog in lockstep with seeded permissions', () => {
		const seeded = seededPermissionCodes().sort();
		expect(seeded.length).toBeGreaterThan(50);
		expect(permissionCatalogCodes()).toEqual(seeded);
		expect(RBAC_PERMISSION_CATALOG.every((permission) => permission.code === `${permission.module}.${permission.action}`)).toBe(true);
		expect(permissionLabel('roles.manage')).toContain('roles.manage');
		expect(permissionLabel('analytics.read')).toContain('dashboard analytics internal');
		expect(permissionLabel('analytics.export')).toContain('laporan analytics agregat');
		expect(permissionLabel('analytics.security_read')).toContain('sinyal keamanan analytics');
		expect(permissionLabel('backup.read')).toContain('status dan daftar backup');
		expect(permissionLabel('backup.download')).toContain('Mengunduh file backup');
		expect(permissionLabel('backup.create')).toContain('backup PostgreSQL manual');
		expect(permissionLabel('backup.restore_plan')).toContain('SOP/perintah restore manual');
	});

	it('documents every seeded permission in the operator catalog', () => {
		for (const code of seededPermissionCodes()) {
			expect(docs).toContain(`\`${code}\``);
		}
		expect(docs).toContain('Legacy role fallback');
		expect(docs).toContain('Tahap 9');
	});

	it('only references seeded permissions from route guards and sidebar metadata', () => {
		const seeded = new Set(seededPermissionCodes());
		const routeSamples = [
			['/settings/users', 'GET'],
			['/api/rbac/roles', 'POST'],
			['/api/rbac/permissions/reports.view/status', 'PATCH'],
			['/api/users/user-1/reset-password', 'POST'],
			['/api/bank-soal/questions', 'POST'],
			['/api/academic/rombel/class-1/timetable-slots/slot-1/journal-session', 'POST'],
			['/api/asesmen/packages', 'POST'],
			['/api/tu/archives/documents', 'POST'],
			['/api/pusaka/settings', 'PUT'],
			['/api/employees/employee-1', 'PUT'],
			['/api/system/backups/status', 'GET'],
			['/api/system/backups/pusaka_20260513_000001.dump/validate-restore', 'POST'],
			['/api/system/backups/pusaka_20260513_000001.dump/restore-command', 'POST'],
			['/api/system/backups/pusaka_20260513_000001.dump/download', 'GET']
		] as const;
		const routePermissions = routeSamples.flatMap(([path, method]) => requiredPermissionsForPath(path, method));
		const sidebarPermissions = [
			...dashboardNavItem.permissions,
			...sidebarNavGroups.flatMap((group) => group.items).flatMap((item) => item.permissions)
		];
		const dashboardPermissions = DASHBOARD_WIDGETS.flatMap((widget) => widget.permissions);
		const unknown = [...new Set([...routePermissions, ...sidebarPermissions, ...dashboardPermissions])].filter((code) => !seeded.has(code));
		expect(unknown).toEqual([]);
	});
});
