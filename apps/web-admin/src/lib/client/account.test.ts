import { describe, expect, it } from 'vitest';

import {
	accountDisplayName,
	accountAvatarUrl,
	accountInitials,
	accountErrorMessage,
	changeRequestTimelineItems,
	changeRequestCanCancel,
	changeRequestStatusLabel,
	contactFieldEditable,
	editableContactFields,
	formatAccountDateTime,
	hasEditableContact,
	isCurrentSession,
	linkedProfileLabel,
	normalizeAccountChangeHistory,
	normalizeChangeRequests,
	normalizeAccountSessions,
	officialChangeFieldOptions,
	officialFieldLabel,
	preferenceItemCount,
	profileHistoryActionLabel,
	profileHistoryFieldLabel,
	profileHistoryStatusLabel,
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
		expect(accountAvatarUrl({
			id: 'u1',
			username: 'guru.ipa',
			roles: ['guru'],
			photo_url: '/api/auth/account/avatar/a.jpg'
		})).toBe('/api/auth/account/avatar/a.jpg');
		expect(accountInitials({
			id: 'u1',
			username: 'guru.ipa',
			display_name: 'Guru IPA',
			roles: ['guru']
		})).toBe('GI');
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

	it('maps official change request fields and statuses conservatively', () => {
		expect(officialFieldLabel('nama')).toBe('Nama resmi');
		expect(officialFieldLabel('nip')).toBe('nip');
		expect(changeRequestStatusLabel('pending')).toBe('Menunggu');
		expect(changeRequestCanCancel({ status: 'pending' })).toBe(true);
		expect(changeRequestCanCancel({ status: 'approved' })).toBe(false);
		expect(officialChangeFieldOptions([
			{ profile_type: 'student', field_key: 'nama', label: 'Nama dari backend', value_type: 'text', self_requestable: true, is_active: true },
			{ profile_type: 'student', field_key: 'tanggal_lahir', label: '', value_type: 'date', self_requestable: true, is_active: true },
			{ profile_type: 'student', field_key: 'nisn', label: 'NISN', value_type: 'text', self_requestable: false, is_active: true }
		])).toEqual([
			{ profile_type: 'student', field_key: 'nama', label: 'Nama dari backend', value_type: 'text', self_requestable: true, is_active: true },
			{ profile_type: 'student', field_key: 'tanggal_lahir', label: 'Tanggal lahir', value_type: 'date', self_requestable: true, is_active: true }
		]);
		expect(officialChangeFieldOptions(null)).toEqual([]);
		expect(normalizeChangeRequests(null)).toEqual([]);
		expect(normalizeChangeRequests([{ id: 'r1', profile_type: 'employee', field_key: 'nama', current_value: 'A', requested_value: 'B', reason: 'dokumen', status: 'pending' }])).toHaveLength(1);
	});

	it('maps profile change history and request timeline labels safely', () => {
		expect(profileHistoryActionLabel('contact_update')).toBe('Kontak pribadi diperbarui');
		expect(profileHistoryActionLabel('unknown_action')).toBe('unknown_action');
		expect(profileHistoryFieldLabel('avatar')).toBe('Foto profil');
		expect(profileHistoryFieldLabel('tanggal_lahir')).toBe('Tanggal lahir');
		expect(profileHistoryStatusLabel('completed')).toBe('Selesai');
		expect(normalizeAccountChangeHistory(null)).toEqual([]);
		expect(normalizeAccountChangeHistory([
			{ action: 'avatar_delete', field_key: 'avatar', status: 'completed', created_at: '2026-05-07T00:00:00Z' },
			{ action: '', field_key: 'ignored', status: 'completed' },
			{ action: 'bad', field_key: 'bad', status: '' }
		])).toHaveLength(1);

		const timeline = changeRequestTimelineItems({
			id: 'r1',
			profile_type: 'employee',
			field_key: 'nama',
			current_value: 'A',
			requested_value: 'B',
			reason: 'dokumen',
			status: 'approved',
			requester_display_name: 'Guru IPA',
			reviewer_username: 'admin',
			review_note: 'Sesuai dokumen',
			created_at: '2026-05-06T00:00:00Z',
			reviewed_at: '2026-05-07T00:00:00Z'
		});
		expect(timeline.map((item) => item.label)).toEqual(['Diajukan', 'Review Disetujui']);
		expect(timeline[1].actor).toBe('admin');
		expect(timeline[1].note).toBe('Sesuai dokumen');
	});

	it('keeps mutation error messages controlled', () => {
		expect(accountErrorMessage(new Error('Gagal'), 'Fallback')).toBe('Gagal');
		expect(accountErrorMessage('', 'Fallback')).toBe('Fallback');
	});
});
