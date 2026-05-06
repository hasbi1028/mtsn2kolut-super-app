export type RBACPermissionCatalogItem = {
	code: string;
	module: string;
	action: string;
	description: string;
};

export const RBAC_PERMISSION_CATALOG = [
	{ code: 'academic.manage', module: 'academic', action: 'manage', description: 'Mengelola data akademik.' },
	{ code: 'academic.read', module: 'academic', action: 'read', description: 'Melihat data akademik.' },
	{ code: 'archives.manage', module: 'archives', action: 'manage', description: 'Mengelola arsip.' },
	{ code: 'archives.read', module: 'archives', action: 'read', description: 'Melihat arsip.' },
	{ code: 'asesmen.event_manage', module: 'asesmen', action: 'event_manage', description: 'Mengelola kegiatan Asesmen.' },
	{ code: 'asesmen.package_manage', module: 'asesmen', action: 'package_manage', description: 'Mengelola paket Asesmen.' },
	{ code: 'asesmen.participant_manage', module: 'asesmen', action: 'participant_manage', description: 'Mengelola peserta Asesmen.' },
	{ code: 'asesmen.proctor', module: 'asesmen', action: 'proctor', description: 'Mengawasi pelaksanaan Asesmen.' },
	{ code: 'asesmen.read', module: 'asesmen', action: 'read', description: 'Melihat modul Asesmen.' },
	{ code: 'asesmen.result_manage', module: 'asesmen', action: 'result_manage', description: 'Mengelola hasil Asesmen.' },
	{ code: 'asesmen.result_read', module: 'asesmen', action: 'result_read', description: 'Melihat hasil Asesmen.' },
	{ code: 'asesmen.score', module: 'asesmen', action: 'score', description: 'Mengoreksi/menilai Asesmen.' },
	{ code: 'asesmen.session_manage', module: 'asesmen', action: 'session_manage', description: 'Mengelola sesi Asesmen.' },
	{ code: 'audit.read', module: 'audit', action: 'read', description: 'Melihat audit log sistem.' },
	{ code: 'bank_soal.analytics', module: 'bank_soal', action: 'analytics', description: 'Melihat analisis Bank Soal.' },
	{ code: 'bank_soal.create', module: 'bank_soal', action: 'create', description: 'Membuat soal Bank Soal.' },
	{ code: 'bank_soal.delete', module: 'bank_soal', action: 'delete', description: 'Menghapus soal.' },
	{ code: 'bank_soal.import', module: 'bank_soal', action: 'import', description: 'Impor soal.' },
	{ code: 'bank_soal.publish', module: 'bank_soal', action: 'publish', description: 'Mempublikasikan soal.' },
	{ code: 'bank_soal.read', module: 'bank_soal', action: 'read', description: 'Melihat Bank Soal.' },
	{ code: 'bank_soal.review', module: 'bank_soal', action: 'review', description: 'Melakukan review/verifikasi soal.' },
	{ code: 'bank_soal.settings', module: 'bank_soal', action: 'settings', description: 'Melihat/mengelola pengaturan Bank Soal.' },
	{ code: 'bank_soal.update', module: 'bank_soal', action: 'update', description: 'Mengubah soal Bank Soal.' },
	{ code: 'dashboard.read', module: 'dashboard', action: 'read', description: 'Melihat dashboard utama.' },
	{ code: 'document_cycles.manage', module: 'document_cycles', action: 'manage', description: 'Mengelola siklus dokumen.' },
	{ code: 'document_cycles.read', module: 'document_cycles', action: 'read', description: 'Melihat siklus dokumen.' },
	{ code: 'governance.manage', module: 'governance', action: 'manage', description: 'Mengelola tata kelola.' },
	{ code: 'governance.read', module: 'governance', action: 'read', description: 'Melihat tata kelola.' },
	{ code: 'grades.manage', module: 'grades', action: 'manage', description: 'Mengelola nilai/rapor.' },
	{ code: 'grades.read', module: 'grades', action: 'read', description: 'Melihat nilai/rapor.' },
	{ code: 'inventory.manage', module: 'inventory', action: 'manage', description: 'Mengelola inventaris.' },
	{ code: 'inventory.read', module: 'inventory', action: 'read', description: 'Melihat inventaris.' },
	{ code: 'journal.manage', module: 'journal', action: 'manage', description: 'Mengelola jurnal kelas.' },
	{ code: 'journal.read', module: 'journal', action: 'read', description: 'Melihat jurnal kelas.' },
	{ code: 'kesiswaan.manage', module: 'kesiswaan', action: 'manage', description: 'Mengelola modul kesiswaan.' },
	{ code: 'kesiswaan.read', module: 'kesiswaan', action: 'read', description: 'Melihat modul kesiswaan.' },
	{ code: 'letters.manage', module: 'letters', action: 'manage', description: 'Mengelola persuratan.' },
	{ code: 'letters.read', module: 'letters', action: 'read', description: 'Melihat persuratan.' },
	{ code: 'library.manage', module: 'library', action: 'manage', description: 'Mengelola perpustakaan.' },
	{ code: 'library.read', module: 'library', action: 'read', description: 'Melihat perpustakaan.' },
	{ code: 'notifications.read', module: 'notifications', action: 'read', description: 'Melihat notifikasi.' },
	{ code: 'parents.manage', module: 'parents', action: 'manage', description: 'Mengelola data orang tua/wali.' },
	{ code: 'parents.read', module: 'parents', action: 'read', description: 'Melihat data orang tua/wali.' },
	{ code: 'pusaka.credentials_manage', module: 'pusaka', action: 'credentials_manage', description: 'Mengelola kredensial PUSAKA.' },
	{ code: 'pusaka.manage', module: 'pusaka', action: 'manage', description: 'Mengelola modul PUSAKA.' },
	{ code: 'pusaka.read', module: 'pusaka', action: 'read', description: 'Melihat modul PUSAKA.' },
	{ code: 'pusaka.sync', module: 'pusaka', action: 'sync', description: 'Menjalankan sinkronisasi PUSAKA.' },
	{ code: 'roles.manage', module: 'roles', action: 'manage', description: 'Mengelola role dan permission.' },
	{ code: 'roles.read', module: 'roles', action: 'read', description: 'Melihat role dan matriks permission.' },
	{ code: 'settings.account', module: 'settings', action: 'account', description: 'Mengelola pengaturan akun sendiri.' },
	{ code: 'settings.school_profile', module: 'settings', action: 'school_profile', description: 'Mengelola profil madrasah.' },
	{ code: 'students.manage', module: 'students', action: 'manage', description: 'Mengelola data siswa.' },
	{ code: 'students.read', module: 'students', action: 'read', description: 'Melihat data siswa.' },
	{ code: 'users.create', module: 'users', action: 'create', description: 'Membuat user baru.' },
	{ code: 'users.deactivate', module: 'users', action: 'deactivate', description: 'Menonaktifkan atau mengaktifkan user.' },
	{ code: 'users.manage_roles', module: 'users', action: 'manage_roles', description: 'Mengatur role user.' },
	{ code: 'users.read', module: 'users', action: 'read', description: 'Melihat daftar dan detail user.' },
	{ code: 'users.reset_password', module: 'users', action: 'reset_password', description: 'Reset password user.' },
	{ code: 'users.update', module: 'users', action: 'update', description: 'Mengubah data user.' },
	{ code: 'website.manage', module: 'website', action: 'manage', description: 'Mengelola konten website.' },
	{ code: 'website.read', module: 'website', action: 'read', description: 'Melihat konten website.' },
] as const satisfies readonly RBACPermissionCatalogItem[];

export function permissionCatalogCodes() {
	return RBAC_PERMISSION_CATALOG.map((permission) => permission.code).sort();
}

export function permissionLabel(code: string) {
	const permission = RBAC_PERMISSION_CATALOG.find((item) => item.code === code);
	if (!permission) return code;
	return `${permission.code} — ${permission.description}`;
}
