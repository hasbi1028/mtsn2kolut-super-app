import { describe, expect, it } from 'vitest';

import pageSource from './+page.svelte?raw';

describe('/students account generation actions', () => {
	it('renders student account generation controls and safety copy', () => {
		expect(pageSource).toContain('Akun Siswa');
		expect(pageSource).toContain('Preview akun siswa');
		expect(pageSource).toContain('Generate akun siswa');
		expect(pageSource).toContain('Password hanya tampil sekali. Simpan/unduh hasil generate sekarang.');
		expect(pageSource).toContain('Akun wajib mengganti password saat login pertama.');
		expect(pageSource).toContain('Data resmi tetap dikunci dan perubahan melalui approval.');
	});

	it('shows account status columns and respects permission-aware action state', () => {
		expect(pageSource).toContain('Status akun');
		expect(pageSource).toContain('Username');
		expect(pageSource).toContain('Role');
		expect(pageSource).toContain('student_accounts.manage');
		expect(pageSource).toContain('canManageStudentAccounts');
	});
});
