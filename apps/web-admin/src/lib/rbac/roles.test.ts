import { describe, expect, it } from 'vitest';
import { buildRoleMetadataDraft, canEditRoleMetadata, normalizeRoleCode, roleMetadataChanged, sanitizeRoleMetadataPayload } from './roles';

const systemRole = { code: 'admin', name: 'Administrator', description: 'System admin', is_system: true, is_active: true };
const customRole = { code: 'operator_cbt', name: 'Operator CBT', description: 'Kelola asesmen', is_system: false, is_active: true };

describe('rbac role metadata helpers', () => {
	it('normalizes custom role codes for safe creation', () => {
		expect(normalizeRoleCode(' Operator CBT ')).toBe('operator_cbt');
		expect(normalizeRoleCode('Wali-Kelas.Manage')).toBe('wali_kelas_manage');
		expect(normalizeRoleCode('__Guru__Piket__')).toBe('guru_piket');
	});

	it('builds and sanitizes role metadata payloads', () => {
		expect(buildRoleMetadataDraft(customRole)).toEqual({
			code: 'operator_cbt',
			name: 'Operator CBT',
			description: 'Kelola asesmen'
		});
		expect(sanitizeRoleMetadataPayload({ code: ' Operator CBT ', name: ' Operator CBT ', description: '  Kelola CBT  ' })).toEqual({
			code: 'operator_cbt',
			name: 'Operator CBT',
			description: 'Kelola CBT'
		});
		expect(sanitizeRoleMetadataPayload({ code: '', name: 'Role Kosong', description: '' })).toEqual({
			code: '',
			name: 'Role Kosong',
			description: ''
		});
	});

	it('guards metadata/status editing for system roles', () => {
		expect(canEditRoleMetadata(customRole)).toBe(true);
		expect(canEditRoleMetadata(systemRole)).toBe(false);
	});

	it('detects metadata changes against selected role', () => {
		expect(roleMetadataChanged(customRole, buildRoleMetadataDraft(customRole))).toBe(false);
		expect(roleMetadataChanged(customRole, { code: 'ignored', name: 'Operator Asesmen', description: 'Kelola asesmen' })).toBe(true);
		expect(roleMetadataChanged(customRole, { code: 'ignored', name: 'Operator CBT', description: '' })).toBe(true);
	});
});
