import { describe, expect, it } from 'vitest';

import pageSource from './+page.svelte?raw';

describe('/parents account generation actions', () => {
	it('renders parent account generation controls and safety copy', () => {
		expect(pageSource).toContain('Akun Orang Tua/Wali');
		expect(pageSource).toContain('Preview akun orang tua');
		expect(pageSource).toContain('Generate akun orang tua');
		expect(pageSource).toContain('Password hanya tampil sekali. Simpan/unduh hasil generate sekarang.');
		expect(pageSource).toContain('Akun wajib mengganti password saat login pertama.');
		expect(pageSource).toContain('Data resmi tetap dikunci dan perubahan melalui approval.');
	});

	it('shows account status columns and respects permission-aware action state', () => {
		expect(pageSource).toContain('Status akun');
		expect(pageSource).toContain('Username');
		expect(pageSource).toContain('Anak terhubung');
		expect(pageSource).toContain('parent_accounts.manage');
		expect(pageSource).toContain('canManageParentAccounts');
	});
});
