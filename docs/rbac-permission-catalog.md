# RBAC Permission Catalog — Tahap 9

Dokumen ini adalah katalog operasional permission dinamis MTsN 2 Kolut setelah Dynamic RBAC Tahap 9. Sumber teknis utama tetap migration `services/core-api/db/migrations/069_dynamic_rbac_foundation.sql` plus seed tambahan staged setelahnya seperti `076_profile_change_review_permission.sql`, `081_student_parent_account_portal.sql`, `082_employee_rbac_permissions.sql`, `084_internal_analytics_permissions.sql`, dan `095_backup_center_permissions.sql`; file ini mengunci daftar agar operator dan developer memakai kode permission yang sama.

## Prinsip Stabilization

- Backend tetap source of truth untuk authorization.
- Frontend hanya menyembunyikan/menampilkan menu berdasarkan permission.
- UI menu dan dashboard memakai permission-first. Fallback role generik hanya dipertahankan untuk admin dan fallback portal siswa/orang tua yang eksplisit.
- Legacy role fallback backend masih dipertahankan di beberapa handler selama transisi; jangan hapus `admin/guru/staf/kesiswaan/siswa/ortu` sebelum cleanup terpisah.
- Mutation role/permission harus diaudit dan dilindungi guard admin terakhir.

## Legacy role fallback

Legacy role fallback masih aktif untuk compatibility backend tertentu. UI baru tidak memakai role `guru` sebagai pengganti permission menu/dashboard; role fallback UI yang tersisa harus dinyatakan eksplisit, terutama `siswa`/`ortu` untuk portal. Target cleanup nanti: audit route yang masih role-only, tambah seeded permission jika perlu, lalu hapus fallback secara staged setelah tidak ada sesi/token lama yang bergantung pada role hardcoded.

## Permission Seed

### academic

- `academic.manage` — Mengelola data akademik.
- `academic.read` — Melihat data akademik.

### archives

- `archives.manage` — Mengelola arsip.
- `archives.read` — Melihat arsip.

### asesmen

- `asesmen.event_manage` — Mengelola kegiatan Asesmen.
- `asesmen.package_manage` — Mengelola paket Asesmen.
- `asesmen.participant_manage` — Mengelola peserta Asesmen.
- `asesmen.proctor` — Mengawasi pelaksanaan Asesmen.
- `asesmen.read` — Melihat modul Asesmen.
- `asesmen.result_manage` — Mengelola hasil Asesmen.
- `asesmen.result_read` — Melihat hasil Asesmen.
- `asesmen.score` — Mengoreksi/menilai Asesmen.
- `asesmen.session_manage` — Mengelola sesi Asesmen.

### audit

- `audit.read` — Melihat audit log sistem.

### analytics

- `analytics.read` — Melihat dashboard analytics internal yang hanya berisi agregat.
- `analytics.export` — Mengekspor laporan analytics agregat tanpa event mentah atau metadata sensitif.
- `analytics.security_read` — Melihat sinyal keamanan analytics yang sudah diagregasi.

### bank_soal

- `bank_soal.analytics` — Melihat analisis Bank Soal.
- `bank_soal.create` — Membuat soal Bank Soal.
- `bank_soal.delete` — Menghapus soal.
- `bank_soal.import` — Impor soal.
- `bank_soal.publish` — Mempublikasikan soal.
- `bank_soal.read` — Melihat Bank Soal.
- `bank_soal.review` — Melakukan review/verifikasi soal.
- `bank_soal.settings` — Melihat/mengelola pengaturan Bank Soal.
- `bank_soal.update` — Mengubah soal Bank Soal.

### backup

- `backup.read` — Melihat status dan daftar backup PostgreSQL.
- `backup.download` — Mengunduh file backup PostgreSQL yang tersedia.
- `backup.create` — Menjalankan backup PostgreSQL manual dari Backup Center.

### dashboard

- `dashboard.read` — Melihat dashboard utama.

### document_cycles

- `document_cycles.manage` — Mengelola siklus dokumen.
- `document_cycles.read` — Melihat siklus dokumen.

### employees

- `employees.manage` — Mengelola data master pegawai.
- `employees.read` — Melihat data master pegawai.

### governance

- `governance.manage` — Mengelola tata kelola.
- `governance.read` — Melihat tata kelola.

### grades

- `grades.manage` — Mengelola nilai/rapor.
- `grades.read` — Melihat nilai/rapor.

### inventory

- `inventory.manage` — Mengelola inventaris.
- `inventory.read` — Melihat inventaris.

### journal

- `journal.manage` — Mengelola jurnal kelas.
- `journal.manage_all` — Mengelola seluruh jurnal kelas lintas guru.
- `journal.read` — Melihat jurnal kelas.
- `journal.read_all` — Melihat seluruh jurnal kelas lintas guru.

### kesiswaan

- `kesiswaan.manage` — Mengelola modul kesiswaan.
- `kesiswaan.read` — Melihat modul kesiswaan.

### letters

- `letters.manage` — Mengelola persuratan.
- `letters.read` — Melihat persuratan.

### library

- `library.manage` — Mengelola perpustakaan.
- `library.read` — Melihat perpustakaan.

### notifications

- `notifications.read` — Melihat notifikasi.

### parent_accounts

- `parent_accounts.manage` — Membuat, reset, dan menonaktifkan akun orang tua/wali.

### parent_portal

- `parent_portal.child_attendance_read` — Melihat kehadiran anak yang terhubung.
- `parent_portal.child_grades_read` — Melihat hasil/nilai anak yang sudah dirilis.
- `parent_portal.child_profile_read` — Melihat profil anak yang terhubung.
- `parent_portal.child_schedule_read` — Melihat jadwal anak yang terhubung.
- `parent_portal.children_read` — Melihat daftar anak yang terhubung.
- `parent_portal.profile_change_request` — Mengajukan perubahan data orang tua/anak melalui approval.
- `parent_portal.read` — Mengakses portal orang tua/wali.

### parents

- `parents.manage` — Mengelola data orang tua/wali.
- `parents.read` — Melihat data orang tua/wali.

### profile_changes

- `profile_changes.review` — Meninjau dan menyetujui permintaan perubahan data resmi profil.

Reviewer perubahan profil dapat membuka `/settings/user-change-requests`, mengambil daftar/filter, melihat pending count, dan export CSV aman melalui `/api/users/change-requests*`. Legacy role `admin` tetap menjadi fallback kompatibilitas selama migrasi RBAC.

### pusaka

- `pusaka.credentials_manage` — Mengelola kredensial PUSAKA.
- `pusaka.manage` — Mengelola modul PUSAKA.
- `pusaka.read` — Melihat modul PUSAKA.
- `pusaka.sync` — Menjalankan sinkronisasi PUSAKA.

### roles

- `roles.manage` — Mengelola role dan permission.
- `roles.read` — Melihat role dan matriks permission.

### settings

- `settings.account` — Mengelola pengaturan akun sendiri.
- `settings.school_profile` — Mengelola profil madrasah.

### student_accounts

- `student_accounts.manage` — Membuat, reset, dan menonaktifkan akun siswa.

### student_portal

- `student_portal.assessment_take` — Mengikuti asesmen sebagai siswa.
- `student_portal.grades_read` — Melihat hasil/nilai siswa sendiri yang sudah dirilis.
- `student_portal.profile_change_request` — Mengajukan perubahan data resmi siswa sendiri.
- `student_portal.profile_read` — Melihat profil siswa sendiri.
- `student_portal.read` — Mengakses portal siswa.
- `student_portal.schedule_read` — Melihat jadwal siswa sendiri.

### students

- `students.manage` — Mengelola data siswa.
- `students.read` — Melihat data siswa.

### users

- `users.create` — Membuat user baru.
- `users.deactivate` — Menonaktifkan atau mengaktifkan user.
- `users.manage_roles` — Mengatur role user.
- `users.read` — Melihat daftar dan detail user.
- `users.reset_password` — Reset password user.
- `users.update` — Mengubah data user.

### website

- `website.manage` — Mengelola konten website.
- `website.read` — Melihat konten website.

## Route Guard Audit Snapshot

Tahap 9 mempertahankan guard permission-aware plus fallback role. Berikut pola yang diaudit di backend `cmd/api/main.go` saat katalog dibuat:

- L168: `requireAdmin := mw.RequireAdmin()`
- L169: `requireCbt := mw.RequireAnyPermissionOrRole([]string{"bank_soal.read", "asesmen.read", "asesmen.score"}, "admin", "guru")`
- L170: `requireCbtOps := mw.RequireAnyPermissionOrRole([]string{"asesmen.proctor"}, "admin", "guru", "staf")`
- L171: `requireStaff := mw.RequireAnyPermissionOrRole([]string{"library.read", "library.manage", "inventory.read", "inventory.manage", "letters.read", "letters.manage", "archives.read", "archives.manage", "governance.read", "governance.manage", "document_cycles.read", "document_cycles.manage"}, "admin", "staf")`
- L172: `requireKesiswaanManage := mw.RequireAnyPermissionOrRole([]string{"students.manage", "kesiswaan.manage"}, "admin", "kesiswaan")`
- L173: `requireAcademicManage := mw.RequireAnyPermissionOrRole([]string{"academic.manage"}, "admin")`
- L174: `requireStudentsManage := mw.RequireAnyPermissionOrRole([]string{"students.manage"}, "admin", "kesiswaan")`
- L175: `requireParentsManage := mw.RequireAnyPermissionOrRole([]string{"parents.manage"}, "admin", "kesiswaan")`
- L176: `requireWebsiteManage := mw.RequireAnyPermissionOrRole([]string{"website.manage"}, "admin")`
- L177: `requirePusakaManage := mw.RequireAnyPermissionOrRole([]string{"pusaka.manage", "pusaka.sync", "pusaka.credentials_manage"}, "admin")`
- L178: `requireAsesmenPackageManage := mw.RequireAnyPermissionOrRole([]string{"asesmen.package_manage"}, "admin")`
- L179: `requireAsesmenEventManage := mw.RequireAnyPermissionOrRole([]string{"asesmen.event_manage"}, "admin")`
- L180: `requireAsesmenSessionManage := mw.RequireAnyPermissionOrRole([]string{"asesmen.session_manage"}, "admin")`
- L181: `requireAsesmenResultRead := mw.RequireAnyPermissionOrRole([]string{"asesmen.result_read", "asesmen.result_manage"}, "admin")`
- L182: `requireAsesmenProctor := mw.RequireAnyPermissionOrRole([]string{"asesmen.proctor"}, "admin")`
- L183: `requireUsersRead := mw.RequirePermission("users.read")`
- L184: `requireUsersCreate := mw.RequirePermission("users.create")`
- L185: `requireUsersDeactivate := mw.RequirePermission("users.deactivate")`
- L186: `requireUsersResetPassword := mw.RequirePermission("users.reset_password")`
- L187: `requireUsersUpdate := mw.RequirePermission("users.update")`
- L188: `requireUsersManageRoles := mw.RequirePermission("users.manage_roles")`
- L189: `requireRolesRead := mw.RequirePermission("roles.read")`
- L190: `requireRolesManage := mw.RequirePermission("roles.manage")`
- L191: `requireAuditRead := mw.RequirePermission("audit.read")`
- L192: `requireSchoolProfileSettings := mw.RequirePermission("settings.school_profile")`
