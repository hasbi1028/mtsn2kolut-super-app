import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { describe, expect, it } from 'vitest';

import { sidebarNavGroups } from '$lib/components/sidebar/sidebar-config';
import { requiredPermissionsForPath } from '$lib/server/route-access';
import { RBAC_PERMISSION_CATALOG, permissionCatalogCodes, permissionLabel } from './permission-catalog';

const repoRoot = resolve(__dirname, '../../../../..');
const migration = readFileSync(resolve(repoRoot, 'services/core-api/db/migrations/069_dynamic_rbac_foundation.sql'), 'utf8');
const docs = readFileSync(resolve(repoRoot, 'docs/rbac-permission-catalog.md'), 'utf8');

function seededPermissionCodes() {
	return Array.from(migration.matchAll(/\('([a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)+)'\s*,\s*'([^']+)'\s*,\s*'([^']+)'\s*,\s*'([^']+)'\)/g)).map((match) => match[1]);
}

describe('RBAC permission catalog stabilization', () => {
	it('keeps the frontend catalog in lockstep with migration 069 seed permissions', () => {
		const seeded = seededPermissionCodes().sort();
		expect(seeded.length).toBeGreaterThan(50);
		expect(permissionCatalogCodes()).toEqual(seeded);
		expect(RBAC_PERMISSION_CATALOG.every((permission) => permission.code === `${permission.module}.${permission.action}`)).toBe(true);
		expect(permissionLabel('roles.manage')).toContain('roles.manage');
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
			['/api/asesmen/packages', 'POST'],
			['/api/tu/archives/documents', 'POST'],
			['/api/pusaka/settings', 'PUT']
		] as const;
		const routePermissions = routeSamples.flatMap(([path, method]) => requiredPermissionsForPath(path, method));
		const sidebarPermissions = sidebarNavGroups.flatMap((group) => group.items).flatMap((item) => item.permissions ?? []);
		const unknown = [...new Set([...routePermissions, ...sidebarPermissions])].filter((code) => !seeded.has(code));
		expect(unknown).toEqual([]);
	});
});
