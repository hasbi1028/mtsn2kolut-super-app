import { describe, expect, it } from 'vitest';
import {
	diffPermissions,
	filterPermissions,
	isCriticalPermission,
	normalizePermissionDraft,
	permissionsByModule,
	rolePermissionMap
} from './matrix';

const roles = [
	{ code: 'guru', name: 'Guru', is_system: true, is_active: true },
	{ code: 'operator', name: 'Operator', is_system: false, is_active: true }
];

const permissions = [
	{ code: 'roles.manage', name: 'Kelola Role', module: 'rbac', is_active: true },
	{ code: 'bank_soal.read', name: 'Baca Bank Soal', module: 'bank_soal', is_active: true },
	{ code: 'bank_soal.create', name: 'Buat Bank Soal', module: 'bank_soal', is_active: true },
	{ code: 'users.read', name: 'Baca Pengguna', module: '', is_active: true }
];

describe('rolePermissionMap', () => {
	it('normalizes array rows into deterministic role permission map', () => {
		expect(
			rolePermissionMap({
				roles,
				role_permissions: [
					{ role_code: 'guru', permission_code: 'bank_soal.create' },
					{ role_code: 'guru', permission_code: 'bank_soal.read' },
					{ role_code: 'guru', permission_code: 'bank_soal.read' },
					{ role_code: 'operator', permission_code: 'users.read' }
				]
			})
		).toEqual({
			guru: ['bank_soal.create', 'bank_soal.read'],
			operator: ['users.read']
		});
	});
});

describe('permissionsByModule', () => {
	it('groups permissions by explicit module or inferred code prefix', () => {
		expect(Object.keys(permissionsByModule(permissions))).toEqual(['bank_soal', 'rbac', 'users']);
		expect(permissionsByModule(permissions).bank_soal.map((permission) => permission.code)).toEqual([
			'bank_soal.create',
			'bank_soal.read'
		]);
	});
});

describe('diffPermissions', () => {
	it('reports added and removed permissions in sorted unique order', () => {
		expect(diffPermissions(['users.read', 'bank_soal.read', 'bank_soal.read'], ['roles.manage', 'bank_soal.read'])).toEqual({
			added: ['roles.manage'],
			removed: ['users.read']
		});
	});
});

describe('isCriticalPermission', () => {
	it('marks critical admin permissions', () => {
		expect(isCriticalPermission('roles.manage')).toBe(true);
		expect(isCriticalPermission('users.manage_roles')).toBe(true);
		expect(isCriticalPermission('bank_soal.read')).toBe(false);
	});
});

describe('normalizePermissionDraft', () => {
	it('returns a sorted unique draft list without empty values', () => {
		expect(normalizePermissionDraft([' bank_soal.read ', '', 'roles.manage', 'bank_soal.read'])).toEqual([
			'bank_soal.read',
			'roles.manage'
		]);
	});
});

describe('filterPermissions', () => {
	it('filters by module and text query across code, name, and description', () => {
		expect(filterPermissions(permissions, { module: 'bank_soal', query: 'buat' }).map((permission) => permission.code)).toEqual([
			'bank_soal.create'
		]);
		expect(filterPermissions(permissions, { module: 'all', query: 'pengguna' }).map((permission) => permission.code)).toEqual([
			'users.read'
		]);
	});
});
