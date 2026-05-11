# Audit Istilah Teknis UI Semua Modul Non-Akademik

Audit awal ini memindai `apps/web-admin/src/routes/**/+page.svelte`, mengecualikan modul Akademik yang sudah dibersihkan. Hasil adalah kandidat user-facing, bukan keputusan final; setiap modul tetap perlu review manual sebelum patch.

- Total kandidat: **268**
- Modul terdampak: **21**
- Total route Svelte dipindai: **103**

## Ringkasan per modul

- Asesmen: 121
- Pengaturan & Akun: 51
- Tata Kelola: 29
- Dokumen: 9
- Nilai/Rapor: 9
- Portal: 9
- TU/Surat: 9
- Inventaris: 5
- Pusaka/Kehadiran: 5
- Dashboard Umum: 3
- Bank Soal: 3
- Perpustakaan: 3
- Academic Alias: 2
- Jurnal: 2
- Orang Tua: 2
- Pegawai: 1
- Jadwal Umum: 1
- Kesiswaan: 1
- Notifikasi: 1
- Siswa: 1
- Website: 1

## Ringkasan per istilah

- sync/queue/retry/failed: 68
- token: 46
- role/RBAC/permission: 44
- slug/SEO/publish/draft: 41
- event: 17
- item/asset/stock/loan: 17
- proctoring: 13
- deploy/release/build/version: 10
- owner/briefing: 8
- template: 7
- matrix: 6
- endpoint/API: 5
- session: 5
- compliance/evidence: 5
- payload/JSON: 4
- error: 3
- readiness: 3
- workflow/cycle: 2
- package: 1

## Contoh kandidat per modul


### Asesmen

- `asesmen/+page.svelte:48` — token — “Mulai dari persiapan ujian, sesi dan token, pemantauan hari-H, lalu hasil.”
- `asesmen/+page.svelte:73` — token — “Paket, kegiatan, jadwal, peserta, ruang, dan token siap sebelum hari ujian.”
- `asesmen/+page.svelte:99` — token — “Lihat jalur paket, kegiatan, sesi, dan token yang disiapkan untuk ujian.”
- `asesmen/+page.svelte:122` — token — “Atur Sesi & Token”
- `asesmen/+page.svelte:123` — token — “Kelola sesi, kartu ujian, dan token dari pusat kegiatan CBT.”
- `asesmen/aplikasi-siswa/+page.svelte:68` — endpoint/API — “Sebelum sesi, pastikan APK terpasang dan API base URL yang dipakai benar.”
- `asesmen/aplikasi-siswa/+page.svelte:88` — proctoring — “Buka rekap ruang/proctoring untuk dashboard live.”
- `asesmen/aplikasi-siswa/+page.svelte:120` — deploy/release/build/version, matrix — “Dibuka setelah kebutuhan monitoring terpenuhi: matrix perangkat dan release checklist.”
- `asesmen/aplikasi-siswa/+page.svelte:123` — matrix — “Matriks Perangkat”
- `asesmen/aplikasi-siswa/+page.svelte:128` — deploy/release/build/version — “Release Checklist”
- `asesmen/aplikasi-siswa/matrix/+page.svelte:65` — token — “Login token berhasil pada koneksi yang dipakai siswa.”
- `asesmen/aplikasi-siswa/release/+page.svelte:33` — event, error — “Perubahan event warning atau error code sudah ditinjau dampaknya ke restore flow mobile.”

### Pengaturan & Akun

- `settings/+page.svelte:211` — sync/queue/retry/failed — “Settings overview render failed”
- `settings/account/+page.svelte:250` — sync/queue/retry/failed — “Account overview render failed”
- `settings/analytics/+page.svelte:57` — endpoint/API — “Core API”
- `settings/analytics/+page.svelte:63` — role/RBAC/permission — “Semua role”
- `settings/analytics/+page.svelte:203` — sync/queue/retry/failed — “Internal analytics render failed”
- `settings/analytics/+page.svelte:446` — event — “Agregat event public website.”
- `settings/audit-logs/+page.svelte:159` — sync/queue/retry/failed — “Audit logs render failed”
- `settings/rbac/+page.svelte:148` — role/RBAC/permission — “Gagal memuat data RBAC.”
- `settings/rbac/+page.svelte:179` — matrix, role/RBAC/permission — “Role sistem tidak dapat diedit metadata/statusnya. Permission tetap dapat diatur melalui matrix dengan guard backend.”
- `settings/rbac/+page.svelte:201` — role/RBAC/permission — “Kode role dan nama role wajib diisi.”
- `settings/rbac/+page.svelte:205` — role/RBAC/permission — “Role sistem tidak dapat diedit metadata/statusnya.”
- `settings/rbac/+page.svelte:223` — role/RBAC/permission — “Gagal menyimpan metadata role.”

### Tata Kelola

- `governance/+page.svelte:440` — slug/SEO/publish/draft — “Draft”
- `governance/+page.svelte:875` — sync/queue/retry/failed — “Governance overview render failed”
- `governance/+page.svelte:1420` — item/asset/stock/loan — “Item RKT”
- `governance/actions/+page.svelte:462` — compliance/evidence, sync/queue/retry/failed — “Compliance actions render failed”
- `governance/actions/+page.svelte:582` — item/asset/stock/loan — “Tandai selesai tanpa tautan/item bukti? Catatan bukti masih bisa ditambahkan lewat Edit.”
- `governance/actions/+page.svelte:612` — item/asset/stock/loan — “Tambahkan URL/lokasi bukti atau pilih item bukti sebelum menandai selesai.”
- `governance/actions/+page.svelte:813` — item/asset/stock/loan — “Buat item RKT/RKJM tahunan agar program memiliki kegiatan operasional.”
- `governance/actions/+page.svelte:816` — item/asset/stock/loan — “Tambahkan tautan bukti, arsip, atau item bukti mutu yang dapat diverifikasi.”
- `governance/actions/calendar/+page.svelte:284` — sync/queue/retry/failed — “Governance action calendar render failed”
- `governance/actions/calendar/+page.svelte:410` — item/asset/stock/loan — “Tidak ada item”
- `governance/actions/evidence-briefing/+page.svelte:80` — owner/briefing — “Briefing bukti tindak lanjut belum dapat dimuat.”
- `governance/actions/evidence-briefing/+page.svelte:240` — item/asset/stock/loan — “Status sudah menunggu bukti dan perlu validasi lokasi/item bukti.”

### Dokumen

- `document-cycles/+page.svelte:307` — event — “Semua event”
- `document-cycles/+page.svelte:446` — compliance/evidence — “evidence governance”
- `document-cycles/+page.svelte:743` — slug/SEO/publish/draft — “Mulai Draft”
- `document-cycles/+page.svelte:744` — slug/SEO/publish/draft — “Koreksi Draft”
- `document-cycles/+page.svelte:914` — compliance/evidence — “Evidence”
- `document-cycles/+page.svelte:944` — event — “Jenis Event”
- `document-cycles/+page.svelte:1007` — workflow/cycle, sync/queue/retry/failed — “Document cycle render failed”
- `document-cycles/verifikasi/+page.svelte:145` — slug/SEO/publish/draft — “Dikembalikan ke draft dari antrian verifikasi.”
- `document-cycles/verifikasi/+page.svelte:189` — sync/queue/retry/failed — “Document verification queue render failed”

### Nilai/Rapor

- `grades/+page.svelte:333` — slug/SEO/publish/draft — “Masih Ada Draft”
- `grades/+page.svelte:396` — slug/SEO/publish/draft — “Masih Ada Draft”
- `grades/+page.svelte:750` — sync/queue/retry/failed — “Grade overview render failed”
- `grades/+page.svelte:877` — slug/SEO/publish/draft — “Komponen dikembalikan ke draft”
- `grades/+page.svelte:1747` — slug/SEO/publish/draft — “Draft”
- `grades/+page.svelte:1762` — slug/SEO/publish/draft — “Kembalikan ke Draft”
- `grades/rapor/+page.svelte:141` — sync/queue/retry/failed — “Rapor school profile load failed”
- `grades/rapor/+page.svelte:177` — sync/queue/retry/failed — “Rapor assignments render failed”
- `grades/rapor/+page.svelte:181` — sync/queue/retry/failed — “Rapor detail render failed”

### Portal

- `portal/orang-tua/+page.svelte:93` — sync/queue/retry/failed — “Parent portal render failed”
- `portal/siswa/+page.svelte:91` — sync/queue/retry/failed — “Student portal render failed”
- `portal/siswa/+page.svelte:143` — token — “Token ruang minimal 4 karakter.”
- `portal/siswa/+page.svelte:153` — token — “Token ruang tidak sesuai. Pastikan Anda berada di ruang yang benar.”
- `portal/siswa/+page.svelte:299` — token — “Buka Token”
- `portal/siswa/+page.svelte:389` — token — “Buka Token”
- `portal/siswa/cbt/[participant_id]/+page.svelte:49` — token — “Token ujian belum dapat dibuka.”
- `portal/siswa/cbt/[participant_id]/+page.svelte:67` — token — “Token dapat dibuka”
- `portal/siswa/cbt/[participant_id]/+page.svelte:115` — token — “Buka Token Ujian”

### TU/Surat

- `tu/+page.svelte:172` — sync/queue/retry/failed — “TU dashboard render failed”
- `tu/arsip/+page.svelte:494` — sync/queue/retry/failed — “Archive render failed”
- `tu/compliance-pack/+page.svelte:208` — compliance/evidence, sync/queue/retry/failed — “TU compliance pack render failed”
- `tu/disposisi/+page.svelte:84` — sync/queue/retry/failed — “TU dispositions render failed”
- `tu/surat-keluar/+page.svelte:128` — sync/queue/retry/failed — “TU outgoing letters render failed”
- `tu/surat-keterangan/+page.svelte:130` — template — “Gagal memuat template surat”
- `tu/surat-keterangan/+page.svelte:202` — sync/queue/retry/failed — “TU student certificate render failed”
- `tu/surat-masuk/+page.svelte:119` — sync/queue/retry/failed — “TU incoming letters render failed”
- `tu/surat-masuk/+page.svelte:123` — sync/queue/retry/failed — “TU incoming letter dispositions render failed”

### Inventaris

- `inventory/+page.svelte:185` — sync/queue/retry/failed — “Inventory render failed”
- `inventory/+page.svelte:367` — item/asset/stock/loan — “Semua item masih berada di atas batas minimum stok yang dicatat.”
- `inventory/items/+page.svelte:91` — item/asset/stock/loan — “Min Stock”
- `inventory/items/+page.svelte:245` — sync/queue/retry/failed — “Inventory items render failed”
- `inventory/items/+page.svelte:249` — sync/queue/retry/failed — “Inventory history render failed”

### Pusaka/Kehadiran

- `pusaka/+page.svelte:161` — sync/queue/retry/failed — “PUSAKA overview render failed”
- `pusaka/antrian/+page.svelte:145` — sync/queue/retry/failed — “Job queue render failed”
- `pusaka/employees/+page.svelte:102` — sync/queue/retry/failed — “PUSAKA employees render failed”
- `pusaka/kehadiran/+page.svelte:175` — sync/queue/retry/failed — “PUSAKA attendance render failed”
- `pusaka/summary/+page.svelte:109` — sync/queue/retry/failed — “PUSAKA attendance summary render failed”

### Dashboard Umum

- `+page.svelte:153` — role/RBAC/permission — “Shortcut penyusunan dan pengelolaan Bank Soal sesuai permission yang aktif pada akun ini.”
- `+page.svelte:158` — role/RBAC/permission — “Shortcut penyusunan dan pengelolaan Bank Soal sesuai permission yang aktif pada akun ini.”
- `+page.svelte:289` — error — “Dashboard boundary error”

### Bank Soal

- `bank-soal/pengaturan/+page.svelte:34` — slug/SEO/publish/draft — “Draft”
- `bank-soal/pengaturan/+page.svelte:51` — template, error — “Gunakan template resmi, jalankan preview dry-run, validasi error, baru lakukan import final.”
- `bank-soal/pengaturan/+page.svelte:61` — template — “Preview dry-run dan import final dari Word/Excel/template.”

### Perpustakaan

- `library/+page.svelte:99` — sync/queue/retry/failed — “Library overview render failed”
- `library/books/+page.svelte:121` — sync/queue/retry/failed — “Library books render failed”
- `library/loans/+page.svelte:200` — sync/queue/retry/failed — “Library loans render failed”

### Academic Alias

- `academic/+page.svelte:241` — sync/queue/retry/failed — “Academic render failed”
- `academic/+page.svelte:838` — matrix — “Isi jadwal terlebih dahulu atau longgarkan filter agar matriks mingguan bisa ditampilkan.”

### Jurnal

- `journal/+page.svelte:129` — sync/queue/retry/failed — “Journal overview render failed”
- `journal/[id]/+page.svelte:119` — sync/queue/retry/failed — “Journal detail render failed”

### Orang Tua

- `parents/+page.svelte:186` — sync/queue/retry/failed — “Parents overview render failed”
- `parents/+page.svelte:190` — sync/queue/retry/failed — “Linked students render failed”

### Jadwal Umum

- `jadwal/+page.svelte:242` — sync/queue/retry/failed — “Schedule render failed”

### Kesiswaan

- `kesiswaan/+page.svelte:584` — sync/queue/retry/failed — “Kesiswaan render failed”

### Notifikasi

- `notifications/+page.svelte:132` — sync/queue/retry/failed — “Notifications render failed”

### Pegawai

- `employees/+page.svelte:96` — sync/queue/retry/failed — “Employees render failed”

### Website

- `website/announcements/+page.svelte:10` — workflow/cycle, slug/SEO/publish/draft — “Kelola pengumuman resmi sekolah dengan workflow draft ke publish.”

### Siswa

- `students/+page.svelte:178` — sync/queue/retry/failed — “Students overview render failed”
