import { describe, expect, it } from 'vitest';

import {
	accountDisplayName,
	accountErrorMessage,
	contactFieldEditable,
	editableContactFields,
	formatAccountDateTime,
	hasEditableContact,
	isCurrentSession,
	linkedProfileLabel,
	normalizeAccountSessions,
	preferenceItemCount,
	profileTypeLabel,
	roleLabel,
	sessionTitle
} from './account';

describe('account client helpers', () => {
	it('formats account identity labels without exposing implementation fields', () => {
		expect(accountDisplayName({
			id: 'u1',
			username: 'guru.ipa',
			display_name: '',
			profile_nama: 'Nama Guru',
			roles: ['guru']
		})).toBe('Nama Guru');
		expect(roleLabel('admin')).toBe('Administrator');
		expect(roleLabel('custom_role')).toBe('custom_role');
		expect(profileTypeLabel('employee')).toBe('Pegawai');
		expect(profileTypeLabel('')).toBe('Tidak tertaut');
		expect(linkedProfileLabel({
			id: 'u1',
			username: 'guru.ipa',
			roles: ['guru'],
			profile_type: 'employee',
			profile_nama: 'Nama Guru'
		})).toBe('Pegawai: Nama Guru');
	});

	it('formats WITA timestamps with a stable empty fallback', () => {
		expect(formatAccountDateTime(null)).toBe('Belum tercatat');
		expect(formatAccountDateTime('not-a-date')).toBe('Belum tercatat');
		expect(formatAccountDateTime('2026-05-06T00:30:00Z')).toContain('6 Mei 2026');
	});

	it('normalizes session and preference labels', () => {
		expect(sessionTitle({ id: 'abcdef123456', device_label: '' })).toBe('Sesi abcdef12');
		expect(sessionTitle({ id: 'abcdef123456', device_label: ' Laptop TU ' })).toBe('Laptop TU');
		expect(normalizeAccountSessions(null)).toEqual([]);
		expect(normalizeAccountSessions([{ id: 's1' }])).toEqual([{ id: 's1' }]);
		expect(isCurrentSession({ id: 's1' }, 's1')).toBe(true);
		expect(isCurrentSession({ id: 's1' }, 's2')).toBe(false);
		expect(preferenceItemCount(['/', '', '/settings/account'])).toBe(2);
	});

	it('keeps contact editability limited to supported fields', () => {
		const contact = { phone: '0812', email: 'guru@example.id', editable_fields: ['phone', 'email', 'role'] };
		expect(editableContactFields(contact)).toEqual(['phone', 'email']);
		expect(contactFieldEditable(contact, 'phone')).toBe(true);
		expect(contactFieldEditable(contact, 'role')).toBe(false);
		expect(hasEditableContact(contact)).toBe(true);
		expect(hasEditableContact({ editable_fields: [] })).toBe(false);
	});

	it('keeps mutation error messages controlled', () => {
		expect(accountErrorMessage(new Error('Gagal'), 'Fallback')).toBe('Gagal');
		expect(accountErrorMessage('', 'Fallback')).toBe('Fallback');
	});
});
