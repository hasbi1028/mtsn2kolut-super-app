import { describe, expect, it } from 'vitest';

import { buildUIPolicyPreview, evaluateUIPolicyAccess } from './ui-policy';

describe('UI policy access evaluation', () => {
	it('uses permissions first and keeps admin as the only generic role override', () => {
		const policy = { permissions: ['bank_soal.create'], roleFallbacks: ['siswa'] };

		expect(evaluateUIPolicyAccess(policy, { roles: ['guru'], permissions: [] })).toMatchObject({
			allowed: false,
			reason: 'missing_permission'
		});
		expect(evaluateUIPolicyAccess(policy, { roles: ['guru'], permissions: ['bank_soal.create'] })).toMatchObject({
			allowed: true,
			reason: 'permission'
		});
		expect(evaluateUIPolicyAccess(policy, { roles: ['admin'], permissions: [] })).toMatchObject({
			allowed: true,
			reason: 'admin'
		});
		expect(evaluateUIPolicyAccess(policy, { roles: ['siswa'], permissions: [] })).toMatchObject({
			allowed: true,
			reason: 'role_fallback'
		});
	});

	it('builds menu and dashboard previews from draft permissions', () => {
		const preview = buildUIPolicyPreview('guru', ['bank_soal.read', 'bank_soal.create']);

		expect(preview.visibleMenuItems.map((item) => item.href)).toEqual([
			'/',
			'/bank-soal',
			'/bank-soal/tambah',
			'/bank-soal/impor',
			'/settings/account'
		]);
		expect(preview.hiddenMenuItems.map((item) => item.href)).not.toContain('/asesmen/pelaksanaan');
		expect(preview.visibleDashboardWidgets.map((widget) => widget.id)).toEqual([
			'bank-soal-overview',
			'bank-soal-authoring'
		]);
		expect(preview.hiddenDashboardWidgets.map((widget) => widget.id)).toContain('teacher-assessment');
	});
});
