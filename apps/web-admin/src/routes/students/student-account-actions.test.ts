import { describe, expect, it } from 'vitest';

import pageSource from './+page.svelte?raw';

describe('/students account generation actions', () => {
	it('renders student account generation controls and safety copy', () => {
		expect(pageSource).toContain('Akun Siswa');
		expect(pageSource).toContain('Pratinjau akun siswa');
		expect(pageSource).toContain('Buat akun siswa');
		expect(pageSource).toContain('Password awal memakai NISN siswa dan hanya tampil sekali.');
		expect(pageSource).toContain('Akun wajib mengganti password saat login pertama.');
		expect(pageSource).toContain('Data resmi tetap dilindungi dan perubahan melalui persetujuan.');
	});

	it('shows account status columns and respects permission-aware action state', () => {
		expect(pageSource).toContain('Status akun');
		expect(pageSource).toContain('Username');
		expect(pageSource).toContain('Role');
		expect(pageSource).toContain('student_accounts.manage');
		expect(pageSource).toContain('canManageStudentAccounts');
	});
});
