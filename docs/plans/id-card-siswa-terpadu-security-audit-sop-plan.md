# Rencana Keamanan, Audit, Privasi, dan SOP Operasional — ID Card Siswa Terpadu

> **Status:** Plan only. Tidak ada perubahan implementasi runtime.

**Goal:** Menjadikan ID Card Siswa Terpadu sebagai kartu identitas utama yang aman untuk portal siswa, presensi, perpustakaan, layanan madrasah, dan verifikasi CBT tanpa menjadikan QR sebagai password.

**Prinsip Utama:** QR pada kartu bisa hilang/difoto, maka QR hanya boleh menjadi *card identifier / challenge starter*, bukan credential permanen. Semua keputusan akses tetap dilakukan server berdasarkan status kartu, status siswa, role/scope pemindai, rate limit, dan audit.

**Boundary Sistem:** Core API adalah pemilik data dan policy. Web Admin/Portal hanya BFF/proxy. Database hanya diakses Core API.

---

## 1. Keputusan Keamanan Non-Negotiable

1. **QR bukan password dan bukan auto-login.**
   - Scan QR portal harus lanjut ke PIN/password.
   - Scan QR presensi/perpustakaan/CBT harus dilakukan oleh aktor login yang punya scope.
2. **QR tidak berisi data pribadi.**
   - Dilarang memasukkan NISN mentah, NIK, KK, alamat, nomor HP orang tua, password, PIN, token CBT, atau biodata lengkap ke QR.
   - QR berisi URL/format token random, misalnya `https://mtsn2kolut.sch.id/s/idc/{token}` atau `mt2k-idc:v1:{token}`.
3. **Token kartu harus revocable.**
   - Token dibuat dengan CSPRNG minimal 128 bit, ideal 192/256 bit.
   - DB menyimpan hash/HMAC token, bukan token mentah.
   - Reissue/cetak ulang kartu karena hilang/bocor selalu revoke token lama dan generate token baru.
4. **CBT memakai token ruang/sesi terpisah.**
   - ID card hanya verifikasi identitas peserta.
   - Masuk ujian tetap butuh login siswa + jadwal/assignment valid + token ruang/sesi aktif.
5. **Presensi anti-titip adalah kombinasi teknis + SOP.**
   - Karena QR bisa difoto, presensi siswa tidak boleh self-scan bebas tanpa pengawasan.
   - Scanner menampilkan foto/nama/kelas untuk verifikasi visual petugas.
   - Ada rate limit, idempotency, audit lokasi/perangkat, dan deteksi anomali.

---

## 2. Data Minimal dan Privasi

### Data yang boleh dicetak di kartu

- Nama madrasah dan logo.
- Judul: `Kartu Identitas Siswa`.
- Foto siswa.
- Nama siswa.
- Kelas/rombel atau angkatan/tahun berlaku.
- Nomor kartu publik / NIS internal jika memang diperlukan operasional.
- Masa berlaku.
- QR verifikasi kartu.
- Instruksi jika kartu ditemukan: kembalikan ke madrasah / kontak resmi.
- Peringatan: `QR bukan password. Jangan membagikan foto kartu.`

### Data yang tidak boleh dicetak / tidak boleh masuk QR

- PIN/password.
- Token CBT, token ruang, atau token sesi ujian.
- NIK/KK.
- Alamat lengkap.
- Nomor HP orang tua/wali.
- Tanggal lahir lengkap jika tidak wajib.
- Raw database ID yang berurutan atau mudah ditebak.

### Public verify minimal data

Jika QR dibuka oleh publik/tanpa login, tampilkan hanya:

- Status valid/tidak valid kartu.
- Nama singkat atau inisial + kelas umum bila diperlukan.
- Foto kecil opsional hanya jika kebijakan madrasah menyetujui; default lebih aman: tidak tampilkan foto publik.
- Instruksi: “Untuk layanan/login, siswa harus masuk dengan PIN.”

Jangan tampilkan:

- NISN lengkap.
- Riwayat presensi/perpustakaan/CBT.
- Kontak keluarga.
- Alamat.
- Status pelanggaran/rahasia siswa.

---

## 3. Model Lifecycle Kartu

Status kartu yang direkomendasikan:

- `draft`: kartu dibuat, belum dicetak/aktif.
- `printed`: token sudah dipakai untuk render/cetak, belum aktif jika madrasah memakai aktivasi terpisah.
- `active`: kartu dapat dipakai.
- `suspended`: dinonaktifkan sementara untuk investigasi.
- `lost`: dilaporkan hilang.
- `damaged`: fisik rusak, perlu ganti.
- `revoked`: dicabut permanen.
- `replaced`: diganti oleh kartu baru.
- `expired`: masa berlaku habis / siswa lulus/pindah/nonaktif.

### Matrix aksi lifecycle

- `draft -> printed`: operator cetak kartu; token mentah hanya tersedia saat proses cetak.
- `printed -> active`: operator aktivasi setelah kartu diterima siswa.
- `active -> lost`: laporan hilang; token langsung ditolak.
- `active -> damaged`: kartu rusak; layanan lama boleh ditolak atau tetap aktif sementara sesuai SOP.
- `active/suspended/lost/damaged -> revoked`: dicabut permanen.
- `lost/damaged/revoked/expired -> replaced`: kartu baru diterbitkan; token baru; kartu lama tidak dapat dipakai.
- `active -> expired`: otomatis saat siswa keluar/lulus atau masa berlaku habis.
- `suspended -> active`: hanya jika investigasi selesai dan alasan dicatat.

### Aturan satu kartu aktif

- Untuk kartu utama, hanya boleh ada satu kartu `active` per siswa.
- Saat kartu pengganti diaktifkan, kartu lama otomatis menjadi `replaced/revoked`.
- Scan kartu lama harus menghasilkan `denied` + audit, bukan silent fail.

---

## 4. SOP Operasional Lifecycle

### SOP A — Generate Kartu Massal

1. Operator `kesiswaan/admin` memilih angkatan/kelas/siswa.
2. Sistem menolak siswa tanpa master data minimum: nama, rombel, status aktif, foto jika diwajibkan.
3. Sistem membuat `student_cards` status `draft` dengan token random.
4. Sistem mencatat audit `generated` berisi actor, scope, jumlah, filter, timestamp.
5. Gate sebelum lanjut: daftar siswa, jumlah kartu, dan duplikasi kartu aktif ditinjau operator.

**Risk:** generate ganda untuk siswa yang sama.
**Gate:** unique constraint satu active/printable utama per siswa + preview dampak.
**Test:** generate dua kali untuk siswa sama tidak membuat dua kartu aktif.

### SOP B — Print Kartu

1. Operator memilih batch `draft/printed`.
2. Sistem membuat file cetak; raw token hanya muncul di renderer cetak sekali/terbatas.
3. Status berubah `printed`, `printed_at`, `printed_by` terisi.
4. Audit `printed` menyimpan batch ID dan jumlah; tidak menyimpan raw token.
5. File cetak lama diberi TTL/akses terbatas; jangan menjadi arsip publik permanen.

**Risk:** PDF/PNG kartu bocor.
**Gate:** akses print hanya role `student_card.print`, watermark batch internal, TTL link download.
**Test:** user tanpa permission print tidak bisa download ulang file batch.

### SOP C — Aktivasi Kartu

1. Setelah kartu dibagikan, operator aktivasi per siswa/batch.
2. Sistem memeriksa tidak ada kartu aktif lain; jika ada, lama menjadi `replaced`.
3. Status menjadi `active`, audit `activated` menyimpan actor dan alasan/batch.
4. Siswa diberi instruksi mengganti/set PIN portal jika belum ada.

**Risk:** kartu belum diterima tetapi sudah aktif.
**Gate:** opsi madrasah: aktivasi massal hanya setelah tanda terima/BA distribusi.
**Test:** kartu `printed` belum aktif ditolak untuk login/presensi jika policy mengharuskan active.

### SOP D — Kartu Hilang / QR Difoto / Diduga Bocor

1. Operator menerima laporan dari siswa/wali/guru.
2. Operator cari siswa, lihat kartu aktif, klik `Tandai Hilang/Bocor`.
3. Wajib isi alasan dan sumber laporan.
4. Sistem mengubah status menjadi `lost` atau `revoked`.
5. Semua scan token lama ditolak.
6. Jika PIN juga diduga bocor, set `pin_reset_required`.
7. Jika perlu, operator jalankan `Reissue Kartu`.

**Risk:** foto QR lama masih dipakai teman.
**Gate:** revoke langsung, audit denied untuk scan berikutnya, notifikasi ke kesiswaan jika scan lama berulang.
**Test:** token kartu lost tidak bisa dipakai untuk portal, presensi, perpustakaan, CBT identity.

### SOP E — Kartu Rusak

1. Operator tandai `damaged` dengan catatan fisik.
2. Jika kartu masih dipegang siswa dan QR belum bocor, madrasah boleh pilih:
   - `reissue` langsung; atau
   - `suspend` sementara sampai cetak baru selesai.
3. Kartu pengganti memakai token baru.
4. Kartu lama menjadi `replaced/revoked` saat pengganti aktif.

**Risk:** kartu rusak tapi token lama masih bisa dipakai tanpa kontrol.
**Gate:** policy jelas: damaged tidak boleh dipakai setelah reissue aktif.
**Test:** setelah reissue, scan kartu lama denied.

### SOP F — Revoke Manual

1. Hanya `admin/kesiswaan` berpermission `student_card.revoke`.
2. Wajib isi reason: hilang, bocor, salah cetak, siswa nonaktif, penyalahgunaan.
3. Revoke permanen tidak bisa dibatalkan; jika perlu kartu lagi, harus reissue.
4. Audit menyimpan old_status, new_status, actor, reason.

**Risk:** revoke salah siswa.
**Gate:** konfirmasi dua langkah dengan nama, kelas, foto, nomor kartu.
**Test:** revoke butuh reason dan permission; audit immutable.

### SOP G — Reissue / Cetak Ulang Aman

1. Operator pilih `Reissue` dari profil siswa/kartu.
2. Sistem membuat kartu baru status `draft/printed` dengan token baru.
3. Kartu lama langsung `revoked` atau tetap `active` sampai aktivasi baru sesuai policy madrasah; rekomendasi untuk hilang/bocor: revoke langsung.
4. Saat kartu baru aktif, kartu lama `replaced`.
5. Reset PIN dilakukan terpisah hanya jika diperlukan.

**Risk:** operator menganggap reset PIN sama dengan ganti QR.
**Gate:** UI memisahkan tombol `Reset PIN` dan `Ganti/Reissue Kartu`.
**Test:** reset PIN tidak mengubah qr_token_hash; reissue mengubah qr_token_hash dan revoke lama.

### SOP H — Expired / Alumni / Siswa Pindah

1. Scheduler atau operator menandai kartu `expired` saat siswa lulus/pindah/nonaktif.
2. Kartu expired tidak boleh untuk presensi, perpustakaan aktif, atau CBT.
3. Public verify boleh menampilkan “Kartu tidak aktif/masa berlaku habis” tanpa detail sensitif.
4. Audit mencatat sumber: otomatis/scheduler/operator.

**Risk:** alumni masih meminjam buku atau ikut scan layanan aktif.
**Gate:** semua endpoint scan cek status siswa + status kartu.
**Test:** siswa nonaktif + kartu active tetap ditolak karena student status inactive.

---

## 5. Portal Login dan PIN Reset

### Flow portal aman

1. Siswa scan QR atau input username/NIS.
2. Server resolve kartu dan membuat challenge pendek 2–5 menit.
3. UI menampilkan identitas minimal untuk konfirmasi.
4. Siswa memasukkan PIN/password.
5. Server cek status kartu, status siswa, PIN, rate limit.
6. Session portal dibuat jika semua valid.

### Rate limit PIN

- Per kartu: contoh 5 gagal / 15 menit.
- Per siswa: contoh 10 gagal / jam.
- Per IP/device: contoh 30 gagal / jam.
- Lock sementara bertahap: 5 menit, 15 menit, 1 jam.
- Audit semua gagal PIN tanpa menyimpan PIN.

### SOP Reset PIN

1. Siswa/wali lapor lupa PIN.
2. Operator verifikasi identitas offline: kartu fisik + foto/kelas/wali kelas atau prosedur madrasah.
3. Operator klik `Reset PIN`.
4. Sistem membuat PIN sementara atau flag `pin_reset_required`.
5. Siswa wajib mengganti PIN pada login pertama.
6. Reset PIN tidak mengganti QR kecuali kartu hilang/bocor.

**Risk:** social engineering minta reset PIN.
**Gate:** reason wajib, audit reset, opsi persetujuan kesiswaan untuk mass reset, notifikasi ke wali kelas/orang tua jika tersedia.
**Test:** reset PIN oleh role tanpa permission ditolak; setelah reset, login pertama wajib ganti PIN.

---

## 6. Presensi Anti Titip

### Policy

- Presensi QR hanya lewat akun petugas/guru/piket/kesiswaan yang login.
- Guru hanya bisa scan kelas/jadwal/sesi yang menjadi scope-nya.
- Petugas piket/kesiswaan bisa lintas kelas sesuai permission.
- Scanner wajib menampilkan foto, nama, kelas, dan status kartu agar petugas bisa mencocokkan fisik siswa.

### Flow presensi kelas/kegiatan

1. Petugas login.
2. Pilih konteks: kelas, mapel, kegiatan, gerbang, atau sesi.
3. Scan QR.
4. Server validasi token kartu, status kartu, status siswa, scope petugas, window waktu.
5. Sistem membuat record presensi idempotent.
6. Scan ulang dalam window sama tidak menggandakan; tampilkan “sudah tercatat”.
7. Audit mencatat actor, context, result, device, IP/user-agent.

### Anti-titip controls

- Verifikasi visual foto siswa di layar scanner.
- Rate limit scan per petugas/device.
- Device registration untuk scanner resmi jika memungkinkan.
- Deteksi kartu sama discan di dua lokasi/sesi berdekatan.
- Deteksi petugas scan siswa di luar scope.
- Laporan anomali harian: scan sangat cepat, scan luar jadwal, scan kartu lost/revoked.

**Tests:**

- Scan foto QR dari siswa luar kelas oleh guru tanpa scope harus denied.
- Scan siswa yang sama dua kali dalam window sama tidak membuat duplikat.
- Kartu revoked/lost menghasilkan denied audit.
- Petugas dengan permission global tapi tanpa konteks valid tidak bisa mencatat presensi kelas sembarang.

---

## 7. Perpustakaan

### Policy

- Hanya staf/admin/peran library yang boleh scan transaksi perpustakaan.
- QR hanya lookup anggota; transaksi tetap butuh konfirmasi petugas.
- Pinjam buku tetap memvalidasi status anggota, batas pinjam, denda, dan ketersediaan buku.

### Flow pinjam/kembali

1. Staf login modul perpustakaan.
2. Scan kartu siswa.
3. Server cek kartu active + siswa active + status anggota.
4. Scan barcode buku.
5. Server cek ketersediaan/riwayat/limit.
6. Petugas konfirmasi transaksi.
7. Audit `library_used` dan audit transaksi buku.

### Gate privasi

- Staf library hanya melihat data minimum untuk transaksi: nama, kelas, foto, status anggota, pinjaman/denda relevan.
- Tidak melihat data CBT, nilai, catatan BK, atau data keluarga.

**Tests:**

- Role guru biasa tidak bisa memakai endpoint library scan.
- Kartu active tapi siswa nonaktif ditolak untuk pinjam.
- Kunjungan harian idempotent per siswa/window.

---

## 8. CBT: Token Ruang Terpisah

### Policy

- ID card bukan token ujian.
- ID card dapat dipakai pengawas/proktor untuk validasi identitas peserta dan daftar hadir CBT.
- Masuk ujian tetap memerlukan token ruang/sesi CBT yang punya TTL, scope event/session/room, dan rotasi.

### Flow CBT aman

1. Siswa login aplikasi/portal dengan PIN/password.
2. Sistem menampilkan event CBT yang assigned.
3. Proktor/pengawas membagikan token ruang/sesi aktif.
4. Siswa memasukkan token ruang.
5. Server validasi event, jadwal, room assignment, token TTL, status attempt, dan kebijakan perangkat/IP.
6. Pengawas dapat scan kartu untuk mencocokkan peserta dengan daftar hadir/ruang.
7. Audit `cbt_identity_used` mencatat proktor/pengawas, event, session, room, result.

### Gate CBT

- Token CBT tidak pernah dicetak di kartu.
- Token CBT tidak masuk audit mentah; audit cukup token_id/prefix/context.
- Proktor hanya bisa scan peserta pada room/session assignment-nya.
- Panitia dapat melihat laporan, tapi tidak otomatis punya akses jawaban/kunci tanpa permission khusus.

**Tests:**

- QR kartu saja tidak membuka exam attempt.
- Siswa assigned ruang A tidak bisa masuk dengan token ruang B.
- Proktor ruang A tidak bisa memvalidasi peserta ruang B.
- Scan kartu revoked saat CBT menghasilkan denied dan alert.

---

## 9. Role, Permission, dan Scope

### Permission yang disarankan

- `student_card.read`
- `student_card.issue`
- `student_card.print`
- `student_card.activate`
- `student_card.suspend`
- `student_card.revoke`
- `student_card.reissue`
- `student_card.reset_pin`
- `student_card.audit.read`
- `student_card.scan.attendance`
- `student_card.scan.library`
- `student_card.scan.cbt`
- `student_card.public_verify`

### Mapping awal

- `admin`: semua permission, tetap diaudit.
- `kesiswaan`: issue, print, activate, revoke, reissue, reset PIN, audit siswa.
- `staf perpustakaan`: library scan + data minimum anggota.
- `guru/wali kelas/piket`: attendance scan sesuai kelas/jadwal/tugas.
- `panitia/proktor/pengawas CBT`: CBT identity scan sesuai event/session/room assignment.
- `siswa`: lihat status kartu sendiri setelah login; ganti PIN sendiri.
- `ortu`: lihat status terbatas anak jika portal orang tua aktif.

### Scope rules

- Permission menjawab “boleh memakai fitur apa”.
- Scope menjawab “boleh terhadap siswa/kelas/event/ruang mana”.
- Backend wajib enforce permission + scope; UI/sidebar bukan boundary keamanan.
- Semua perubahan role/scope yang memengaruhi kartu wajib audit.

---

## 10. Audit dan Monitoring

### Audit wajib

- Generate, print, activate, suspend, lost, damaged, revoke, reissue, expired.
- Reset PIN dan perubahan credential portal.
- Portal QR challenge start, PIN success/failure, lockout.
- Public verify request minimal: token prefix/hash lookup result, IP/user-agent, tanpa data sensitif.
- Semua scan allowed/denied di presensi, perpustakaan, CBT.
- Perubahan role/permission/scope.

### Field audit minimum

- `event_type`.
- `card_id`, `student_id`.
- `actor_user_id` nullable untuk public verify.
- `actor_role_snapshot`.
- `module`: portal, attendance, library, cbt, admin, public_verify.
- `context` JSONB: class/session/room/device/IP/user-agent/reason.
- `old_status`, `new_status` untuk lifecycle.
- `result`: allowed, denied, warning.
- `failure_reason` kode aman.
- `created_at`.

### Larangan audit

- Jangan log raw QR token.
- Jangan log PIN/password.
- Jangan log token CBT mentah.
- Jangan log data keluarga/alamat penuh.

### Monitoring anomali

- Kartu sama discan dari dua device/lokasi dalam waktu dekat.
- Banyak gagal PIN setelah QR scan.
- Kartu lost/revoked masih sering discan.
- Petugas scan luar scope.
- Device scan volume tidak wajar.
- Banyak public verify untuk token berbeda dari satu IP.

### Respons operasional

- Low: tampilkan di dashboard audit.
- Medium: notifikasi kesiswaan/admin, minta review.
- High: suspend challenge/kartu sementara, butuh konfirmasi admin.
- Critical: revoke kartu + reset PIN wajib + laporan insiden.

---

## 11. Rate Limit dan Abuse Controls

### Endpoint public verify / QR resolve

- Limit per IP dan per token prefix/hash.
- Response untuk token invalid harus generik: “Kartu tidak valid atau tidak aktif”.
- Jangan bedakan terlalu detail antara token tidak ada vs revoked untuk publik.

### Portal card-login

- Limit challenge start per token/IP/device.
- Challenge TTL 2–5 menit.
- Challenge one-time use.
- PIN verify lockout bertahap.

### Scanner internal

- Limit per actor/device/context.
- Idempotency key untuk presensi/perpustakaan.
- Device registration opsional untuk pos gerbang/perpustakaan/ruang CBT.

### Admin lifecycle

- Bulk generate/print/revoke/reissue harus punya preview dan count confirmation.
- Bulk action wajib reason dan batch audit.
- Export audit dibatasi role dan diberi watermark/operator timestamp.

---

## 12. Sprint 0 / Safe Rollout Gate

Sebelum implementasi produksi:

1. **Audit data siswa:** jumlah siswa aktif, siswa tanpa foto, siswa tanpa rombel, duplikasi NIS/NISN.
2. **Audit role:** siapa yang punya admin/kesiswaan/staf/guru/panitia/proktor.
3. **Audit modul:** presensi/perpustakaan/CBT endpoint mana yang akan memakai kartu.
4. **Legal/SOP:** putuskan data yang boleh tampil di public verify dan kartu fisik.
5. **Backup:** backup DB dan simpan checksum sebelum migrasi.
6. **Migrations additive only:** tambah tabel/kolom, jangan drop/rename/enforce keras sebelum report-only.
7. **Report-only mode:** aktifkan audit/deny preview sebelum hard enforcement untuk presensi/perpustakaan jika sudah ada alur lama.
8. **No deploy/restart produksi tanpa approval eksplisit.**

---

## 13. Tahap Implementasi yang Disarankan

### Tahap 1 — Fondasi kartu dan lifecycle

- Tabel `student_cards`, `student_card_events`, `student_card_scan_audit`.
- Token random + HMAC hash lookup.
- Generate, print, activate, revoke, lost, damaged, reissue, expired.
- UI admin/kesiswaan untuk daftar kartu dan aksi lifecycle.
- Audit immutable.

**Exit gate:** kartu bisa diterbitkan, dicetak, diaktifkan, dicabut, diganti; kartu lama ditolak; audit lengkap.

### Tahap 2 — Portal siswa + PIN

- Credential portal terpisah dari kartu.
- QR challenge + PIN verify.
- Reset PIN SOP dan UI.
- Rate limit/lockout.

**Exit gate:** QR tanpa PIN tidak login; reset PIN tidak mengganti QR; reissue tidak otomatis reset PIN kecuali dipilih.

### Tahap 3 — Presensi scanner

- Scanner petugas login + konteks kelas/kegiatan.
- Scope guru/piket/kesiswaan.
- Idempotency dan foto verifikasi.
- Anomaly report dasar.

**Exit gate:** anti-titip minimal berjalan: scope, visual verification, idempotency, audit denied.

### Tahap 4 — Perpustakaan

- Member scan, visit scan, borrow/return integration.
- Data minimum untuk staf library.
- Audit transaksi dan scan.

**Exit gate:** staf library tidak melihat data di luar kebutuhan, transaksi butuh konfirmasi.

### Tahap 5 — CBT identity integration

- Proctor/pengawas scan identity per room/session.
- Token ruang/sesi tetap terpisah.
- Audit CBT identity dan denied alerts.

**Exit gate:** kartu tidak bisa membuka ujian tanpa token ruang; proktor terbatas pada assignment.

### Tahap 6 — Monitoring, reporting, dan hardening

- Dashboard lost/revoked/suspicious.
- Export audit terbatas.
- Alert scan anomali.
- Rotasi secret/token versioning jika diperlukan.

**Exit gate:** admin dapat menjawab siapa melakukan apa, kapan, dari perangkat mana, dan kenapa ditolak/diizinkan.

---

## 14. Risk Register

### High

- **QR difoto dipakai untuk presensi titip.**
  - Mitigasi: petugas-scanned only, foto di scanner, scope kelas/jadwal, rate limit, anomaly report.
  - Test: siswa luar scope / kartu revoked / scan cepat massal.

- **QR menjadi password de facto.**
  - Mitigasi: PIN wajib untuk portal; token ruang CBT wajib untuk ujian; QR hanya lookup/challenge.
  - Test: scan QR tidak membuat session tanpa PIN.

- **Kartu hilang tetap aktif.**
  - Mitigasi: SOP lost/revoke cepat, denied audit, reissue token baru.
  - Test: scan token lama denied semua modul.

- **Role terlalu luas.**
  - Mitigasi: permission + object scope; event/session/room assignment untuk CBT.
  - Test: guru/proktor/staf mencoba akses luar scope.

### Medium

- **Data pribadi bocor via public verify.**
  - Mitigasi: minimal data, generic invalid response, no NISN/alamat/kontak.
  - Test: unauthenticated response snapshot.

- **PDF/PNG batch kartu bocor.**
  - Mitigasi: TTL link, access role print, no long-lived public storage, audit download.
  - Test: old print link expired; unauthorized denied.

- **Reset PIN disalahgunakan.**
  - Mitigasi: reason wajib, identity verification SOP, audit, forced change on first login.
  - Test: reset by unauthorized denied; PIN temp requires change.

### Low/Operational

- **Kartu rusak/lupa bawa menghambat layanan.**
  - Mitigasi: fallback manual by staff with stronger identity verification and audit.
  - Test: manual override requires reason and elevated role.

- **Duplikasi kartu aktif akibat batch.**
  - Mitigasi: unique active card per student, preview duplicate, transaction guard.
  - Test: concurrent activation cannot create two active cards.

---

## 15. Acceptance Tests / Audit-Safe Test Plan

### Lifecycle tests

- Generate card creates token hash, no raw token persisted.
- Print card records audit and does not expose token outside authorized print flow.
- Activate card makes it usable and replaces old active card.
- Lost/revoked/replaced/expired card denied in all scan endpoints.
- Reissue creates new token and links old/new cards.
- Reset PIN does not change QR token.

### Privacy tests

- Public verify unauthenticated returns minimal fields only.
- QR payload does not contain NISN/NIK/KK/address/PIN/token CBT.
- Audit logs never contain raw token/PIN/password/token CBT.
- Staf library cannot read CBT/grade/BK/private family data through card scan.

### Authorization tests

- Admin/kesiswaan can lifecycle-manage cards according to permission.
- Guru can scan attendance only for scoped class/jadwal/kegiatan.
- Staf library can scan only library context.
- Proktor can scan CBT only for assigned event/session/room.
- Sidebar/UI visibility is not sufficient: direct API calls outside scope must fail.

### Rate-limit tests

- Repeated invalid public verify is throttled.
- Repeated PIN failure locks challenge/student temporarily.
- High-frequency scanner requests from one device/actor are throttled or flagged.
- Scan ulang presensi in same window is idempotent, not duplicate.

### CBT separation tests

- QR scan alone cannot start exam attempt.
- Valid card + invalid room token cannot enter exam.
- Valid token room + wrong student assignment denied.
- Revoked/lost card in CBT identity scan creates denied audit/alert.

### Operational audit tests

- Every lifecycle mutation has actor, reason, old/new state, timestamp.
- Denied scans are audited with safe failure reason.
- Audit export is permission-gated and filters by date/module/student.
- Anomaly report detects scan lost/revoked and out-of-scope scanner.

---

## 16. Dokumentasi dan SOP yang Harus Dibuat untuk Operator

1. Panduan cetak dan distribusi kartu.
2. Form/BA serah terima batch kartu.
3. SOP kartu hilang/difoto/bocor.
4. SOP kartu rusak dan cetak ulang.
5. SOP reset PIN.
6. SOP presensi scanner anti-titip.
7. SOP perpustakaan pakai kartu.
8. SOP CBT: kartu hanya verifikasi identitas, token ruang tetap wajib.
9. SOP audit insiden: siapa cek log, kapan suspend/revoke, kapan reset PIN.
10. Template komunikasi ke siswa: “QR bukan password; jangan unggah foto kartu.”

---

## 17. Ringkasan Final untuk Keputusan Madrasah

Rekomendasi final: gunakan **ID Card Siswa Terpadu** sebagai kartu identitas utama, bukan sekadar kartu CBT. Bangun bertahap: lifecycle aman → portal QR+PIN → presensi petugas-scanned → perpustakaan → CBT identity scan dengan token ruang terpisah → monitoring/anomali. Keamanan utama bukan pada merahasiakan kartu fisik, tetapi pada token yang bisa dicabut, PIN terpisah, role/scope pemindai, rate limit, audit lengkap, dan SOP operasional yang disiplin.
