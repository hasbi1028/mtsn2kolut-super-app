import { describe, expect, it } from 'vitest';

import { canDeleteBankSoal, canImportBankSoal, canManageBankSoalSettings, canPublishBankSoal, canReviewBankSoal } from './access';

describe('bank soal permission helpers', () => {
	it('does not treat plain guru role as reviewer/import/settings authority', () => {
		const guru = { role: 'guru', roles: ['guru'], permissions: [] };

		expect(canReviewBankSoal(guru)).toBe(false);
		expect(canImportBankSoal(guru)).toBe(false);
		expect(canManageBankSoalSettings(guru)).toBe(false);
		expect(canPublishBankSoal(guru)).toBe(false);
		expect(canDeleteBankSoal(guru)).toBe(false);
	});

	it('allows admin and granular permissions for privileged Bank Soal actions', () => {
		expect(canReviewBankSoal({ role: 'admin', roles: ['admin'], permissions: [] })).toBe(true);
		expect(canReviewBankSoal({ role: '', roles: [], permissions: ['bank_soal.review'] })).toBe(true);
		expect(canImportBankSoal({ role: '', roles: [], permissions: ['bank_soal.import'] })).toBe(true);
		expect(canManageBankSoalSettings({ role: '', roles: [], permissions: ['bank_soal.settings'] })).toBe(true);
		expect(canPublishBankSoal({ role: '', roles: [], permissions: ['bank_soal.publish'] })).toBe(true);
		expect(canDeleteBankSoal({ role: 'admin', roles: ['admin'], permissions: [] })).toBe(true);
		expect(canDeleteBankSoal({ role: '', roles: [], permissions: ['bank_soal.delete'] })).toBe(true);
	});
});
