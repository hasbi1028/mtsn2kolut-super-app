import { describe, expect, it } from 'vitest';
import {
	buildPermissionMetadataDraft,
	canTogglePermissionStatus,
	filterPermissionCatalog,
	normalizePermissionSegment,
	permissionMetadataChanged,
	sanitizePermissionMetadataPayload
} from './permissions';

const permission = {
	code: 'bank_soal.publish',
	module: 'bank_soal',
	action: 'publish',
	description: 'Publikasi soal',
	is_active: true
};

describe('rbac permission metadata helpers', () => {
	it('normalizes module/action segments for safe permission codes', () => {
		expect(normalizePermissionSegment(' Bank Soal ')).toBe('bank_soal');
		expect(normalizePermissionSegment('Manage-Roles')).toBe('manage_roles');
		expect(normalizePermissionSegment('__Review!!Queue__')).toBe('review_queue');
	});

	it('builds and sanitizes permission metadata payloads', () => {
		expect(buildPermissionMetadataDraft(permission)).toEqual({
			module: 'bank_soal',
			action: 'publish',
			description: 'Publikasi soal'
		});
		expect(sanitizePermissionMetadataPayload({ module: ' Bank Soal ', action: ' Publish ', description: '  Publish soal  ' })).toEqual({
			code: 'bank_soal.publish',
			module: 'bank_soal',
			action: 'publish',
			description: 'Publish soal'
		});
	});

	it('detects metadata changes without treating code as editable on update', () => {
		expect(permissionMetadataChanged(permission, buildPermissionMetadataDraft(permission))).toBe(false);
		expect(permissionMetadataChanged(permission, { module: 'bank_soal', action: 'archive', description: 'Publikasi soal' })).toBe(true);
		expect(permissionMetadataChanged(permission, { module: 'bank_soal', action: 'publish', description: '' })).toBe(true);
	});

	it('filters permission catalog by module/status/query', () => {
		const catalog = [
			permission,
			{ code: 'users.manage', module: 'users', action: 'manage', description: 'Kelola user', is_active: false },
			{ code: 'roles.read', module: 'roles', action: 'read', description: 'Baca role', is_active: true }
		];
		expect(filterPermissionCatalog(catalog, { module: 'users', status: 'inactive', query: 'kelola' }).map((item) => item.code)).toEqual(['users.manage']);
		expect(filterPermissionCatalog(catalog, { module: 'all', status: 'active', query: 'role' }).map((item) => item.code)).toEqual(['roles.read']);
	});

	it('guards critical permission status toggles in UI helper', () => {
		expect(canTogglePermissionStatus({ code: 'bank_soal.read' })).toBe(true);
		expect(canTogglePermissionStatus({ code: 'roles.manage' })).toBe(false);
		expect(canTogglePermissionStatus({ code: 'users.manage_roles' })).toBe(false);
	});
});
