# Arsitektur Keamanan dan Data ID Card Utama Siswa

Dokumen ini merancang model keamanan dan data untuk ID card utama siswa MTsN 2 Kolaka Utara yang dipakai lintas modul: portal siswa, presensi, perpustakaan, dan integrasi tambahan CBT. Prinsip utama: QR pada kartu adalah **identitas kartu**, bukan password permanen dan bukan login langsung tanpa faktor tambahan.

## 1. Prinsip Keamanan

1. **QR tidak boleh menjadi login permanen langsung.**
   - QR bisa difoto, disalin, dicetak ulang, atau tersebar dari kartu hilang.
   - QR hanya membuka konteks identitas kartu/siswa atau memulai proses verifikasi.
   - Akses portal siswa tetap butuh faktor tambahan: PIN/password/OTP/sesi perangkat.

2. **Token QR harus random, panjang, dan tidak bisa ditebak.**
   - Jangan memakai NISN, NIS, UUID berurutan, nomor induk, atau data personal sebagai isi QR.
   - Gunakan token acak minimal 128 bit, lebih baik 192/256 bit.
   - Simpan hash token di database, bukan token mentah.

3. **Kartu punya lifecycle eksplisit.**
   - Kartu aktif, dicetak, hilang, dicabut/revoked, diganti, expired, atau rusak harus tercatat.
   - Reset/reissue kartu menghasilkan token baru dan otomatis menonaktifkan token lama.

4. **Hak akses diputuskan oleh server, bukan oleh QR.**
   - QR hanya referensi lookup kartu.
   - Core API memverifikasi status kartu, status siswa, scope role pemindai, konteks modul, dan kebijakan rate limit.

5. **CBT harus memakai token ruang/sesi terpisah.**
   - ID card boleh membantu identifikasi peserta, tetapi tidak menggantikan token CBT ruang/sesi.
   - Untuk CBT: siswa login portal/aplikasi + pilih asesmen + masukkan token ruang/sesi aktif + validasi jadwal/perangkat jika diterapkan.

## 2. Komponen Sistem

- **Core API (Go):** pemilik data kartu, status, audit, validasi QR, lifecycle, dan policy authorization.
- **Web Admin / Portal Siswa (SvelteKit BFF):** UI dan proxy; tidak membaca database langsung.
- **Mobile/CBT Client:** login siswa dan asesmen; tidak menyimpan QR token sebagai credential permanen.
- **Presensi Scanner:** halaman/app untuk guru/petugas memindai kartu siswa sesuai scope kelas/kegiatan.
- **Perpustakaan Scanner:** halaman/app staf perpustakaan untuk transaksi pinjam/kembali/kunjungan.
- **PostgreSQL:** penyimpanan canonical untuk kartu, credential portal, audit, dan relasi modul.

## 3. Model Data yang Direkomendasikan

### 3.1 `student_cards`

Mewakili satu kartu fisik/digital yang pernah diterbitkan.

Kolom inti:

- `id`: UUID/ULID internal kartu.
- `student_id`: FK ke master siswa.
- `card_no`: nomor kartu tampilan/cetak, tidak harus rahasia.
- `qr_token_hash`: hash token QR, misalnya HMAC-SHA256 dengan server secret atau Argon2id jika lookup tidak perlu cepat.
- `qr_token_prefix`: 6-10 karakter prefix non-rahasia untuk troubleshooting/log tanpa membuka token penuh.
- `status`: enum `draft`, `active`, `suspended`, `lost`, `revoked`, `replaced`, `expired`, `damaged`.
- `issued_at`, `printed_at`, `activated_at`, `expires_at`.
- `revoked_at`, `revoked_by`, `revocation_reason`.
- `replaced_by_card_id`: FK ke kartu pengganti.
- `last_seen_at`: terakhir dipindai valid.
- `created_by`, `updated_by`, `created_at`, `updated_at`.

Constraint penting:

- Hanya boleh ada satu kartu `active` per siswa untuk jenis kartu utama.
- `qr_token_hash` unik.
- Kartu `revoked/lost/replaced/expired` tidak boleh dipakai untuk aksi baru.

### 3.2 `student_card_events`

Audit lifecycle kartu, append-only.

Kolom inti:

- `id`.
- `card_id`, `student_id`.
- `event_type`: `created`, `printed`, `activated`, `scanned`, `scan_denied`, `suspended`, `marked_lost`, `revoked`, `reissued`, `pin_reset`, `portal_login_started`, `portal_login_success`, `portal_login_failed`, `attendance_used`, `library_used`, `cbt_identity_used`.
- `actor_user_id`: admin/guru/staf/siswa, nullable untuk scan publik terbatas.
- `actor_role_snapshot`: role/permission saat kejadian.
- `module`: `portal`, `attendance`, `library`, `cbt`, `admin`.
- `context`: JSONB, contoh `class_id`, `room_id`, `session_id`, `device_id`, `ip`, `user_agent`, `reason`.
- `result`: `allowed`, `denied`, `warning`.
- `created_at`.

Aturan:

- Jangan simpan QR token mentah, PIN, password, atau token CBT di audit.
- Audit perubahan status harus menyimpan old/new status dan alasan.
- Event denied tetap dicatat untuk deteksi kartu hilang/disalahgunakan.

### 3.3 `student_portal_credentials`

Credential portal siswa terpisah dari QR.

Kolom inti:

- `student_id`.
- `login_identifier`: bisa NIS/NISN/username internal.
- `pin_hash` atau `password_hash`: Argon2id/bcrypt.
- `pin_set_at`, `pin_reset_required`.
- `failed_attempts`, `locked_until`.
- `last_login_at`, `last_login_ip`.
- `created_at`, `updated_at`.

Rekomendasi:

- PIN minimal 6 digit atau password yang lebih kuat; untuk siswa MTs, PIN 6 digit + rate limit lebih realistis.
- Setelah reset oleh admin, siswa wajib ganti PIN saat login pertama.
- Reset PIN tidak mengubah token QR kecuali kartu juga dicurigai bocor/hilang.

### 3.4 `student_card_scan_sessions` opsional

Dipakai untuk flow QR portal agar QR tidak menjadi credential permanen.

Contoh flow:

- Scan QR membuka halaman `/s/card/{public_nonce}` atau endpoint validasi.
- Server membuat challenge/session pendek, misalnya 2-5 menit.
- Siswa tetap memasukkan PIN.
- Setelah PIN valid, dibuat session portal normal.

Kolom:

- `id`, `card_id`, `student_id`.
- `purpose`: `portal_login`, `attendance`, `library`, `cbt_identity`.
- `challenge_hash`.
- `expires_at`, `used_at`.
- `ip`, `user_agent`, `device_fingerprint` opsional.
- `created_at`.

## 4. Format QR Token

Rekomendasi isi QR:

```text
https://mtsn2kolut.sch.id/s/idc/{token}
```

atau untuk scanner internal:

```text
mt2k-idc:v1:{token}
```

Token:

- Dibuat dengan CSPRNG.
- Minimal 22 karakter base64url untuk 128 bit; lebih baik 32-43 karakter.
- Tidak mengandung NISN/NIS/nama/tanggal lahir.
- Token mentah hanya muncul saat pembuatan/cetak, lalu tidak ditampilkan lagi.
- Database menyimpan `HMAC_SHA256(server_secret, token)` agar lookup cepat dan token mentah tidak bocor dari DB.

Validasi:

1. Parse token dan versi.
2. Hash/HMAC token.
3. Lookup kartu by hash.
4. Pastikan status kartu `active`.
5. Pastikan siswa aktif/terdaftar.
6. Evaluasi module policy dan role scope pemindai.
7. Catat audit allowed/denied.

## 5. Lifecycle Kartu

### 5.1 Penerbitan

1. Admin/kesiswaan membuat kartu untuk siswa.
2. Server generate token QR random.
3. Status awal `draft` atau langsung `active` sesuai SOP.
4. Token mentah dikirim ke renderer cetak sekali.
5. Audit `created` dan `printed`.

### 5.2 Aktivasi

- Jika perlu kontrol cetak massal, kartu baru `draft/printed` belum dapat dipakai sampai operator menekan aktifkan.
- Saat aktivasi, kartu aktif lama siswa otomatis menjadi `replaced` jika ada.

### 5.3 Kartu Hilang/Difoto/Bocor

1. Operator pilih siswa/kartu lalu `Tandai Hilang/Bocor`.
2. Status kartu lama menjadi `lost` atau `revoked`.
3. Semua scan kartu lama langsung ditolak.
4. Sistem opsional membuat kartu pengganti dengan token baru.
5. Portal siswa tidak otomatis terkunci kecuali ada indikasi PIN ikut bocor; jika ya, set `pin_reset_required`.
6. Audit wajib menyimpan alasan dan actor.

### 5.4 Reset/Reissue

Jenis reset harus dibedakan:

- **Reset PIN portal:** hanya mengganti credential portal; QR tetap sama.
- **Reissue kartu:** membuat token QR baru dan revoke/replace kartu lama.
- **Suspend sementara:** kartu tidak bisa dipakai sampai diaktifkan kembali, tanpa ganti token.

### 5.5 Expiry dan Alumni

- Kartu bisa expired otomatis saat siswa lulus/pindah/nonaktif.
- Kartu alumni boleh tetap menjadi identitas read-only jika diperlukan, tetapi tidak boleh melakukan presensi, peminjaman aktif, atau CBT.

## 6. Portal Login Siswa

Flow aman yang direkomendasikan:

### Opsi A: Login Manual Utama

1. Siswa buka portal.
2. Masukkan NIS/NISN/username.
3. Masukkan PIN/password.
4. Server validasi credential, status siswa, rate limit.
5. Session portal dibuat.

### Opsi B: QR sebagai Shortcut Identitas, Bukan Password

1. Siswa scan QR kartu.
2. Portal menampilkan nama/inisial/kelas secara terbatas atau langsung form PIN.
3. Siswa memasukkan PIN.
4. Server validasi: kartu aktif + siswa aktif + PIN benar.
5. Session portal dibuat.

Larangan:

- Jangan membuat QR sebagai bearer token yang langsung login tanpa PIN.
- Jangan membuat session jangka panjang hanya karena QR discan.
- Jangan menampilkan data sensitif setelah scan QR sebelum PIN valid.

Kontrol tambahan:

- Rate limit per kartu, IP, device, dan student.
- Lock sementara setelah gagal PIN berulang.
- Notifikasi/audit jika kartu discan dari lokasi/perangkat mencurigakan.

## 7. CBT: Token Ruang/Sesi Tetap Terpisah

ID card hanya membantu identifikasi peserta, bukan membuka ujian penuh.

Flow CBT yang disarankan:

1. Siswa login portal/aplikasi dengan PIN/password.
2. Siswa memilih kegiatan CBT yang tersedia.
3. Sistem memastikan siswa terdaftar di event/session.
4. Proktor/pengawas memberikan **token ruang/sesi** yang berubah per sesi/ruang/waktu.
5. Siswa memasukkan token ruang.
6. Server validasi:
   - event aktif dan jadwal valid,
   - siswa assigned ke sesi/ruang,
   - token ruang cocok dan belum expired,
   - status attempt memenuhi aturan,
   - perangkat/IP/lockdown policy jika diterapkan.
7. Baru masuk ujian.

Pemakaian ID card dalam CBT:

- Pengawas bisa scan kartu untuk mencocokkan peserta dengan daftar hadir CBT.
- Scan kartu dapat mempercepat pencarian peserta, bukan menggantikan token ruang.
- Audit `cbt_identity_used` mencatat pengawas, sesi, ruang, hasil validasi.

Token CBT:

- Disimpan dan diaudit terpisah dari `student_cards`.
- Token ruang punya TTL, scope event/session/room, dan rotasi oleh panitia/proktor.
- Jangan cetak token CBT di ID card.

## 8. Presensi

Policy presensi:

- Pemindai harus login sebagai guru/piket/kesiswaan/admin dengan permission presensi.
- Scope guru dibatasi pada kelas yang diajar/wali kelas/jadwal hari itu jika berlaku.
- Scope petugas piket/kesiswaan bisa lintas kelas sesuai permission.
- QR valid hanya jika kartu aktif dan siswa aktif.

Flow:

1. Petugas login.
2. Pilih mode presensi: kelas, kegiatan, gerbang, atau sesi tertentu.
3. Scan kartu.
4. Server cek kartu + siswa + scope petugas + window waktu.
5. Buat record presensi idempotent: scan ulang dalam window yang sama tidak menggandakan.
6. Audit `attendance_used` dengan result.

Risiko dan mitigasi:

- Foto QR dipakai teman: butuh pengawasan visual/foto siswa di scanner.
- Scan massal palsu: rate limit, device registration untuk scanner tertentu, audit anomali.
- Kartu hilang: status `lost/revoked` langsung menolak scan.

## 9. Perpustakaan

Policy perpustakaan:

- Pemindai harus login sebagai `admin`/`staf` atau permission library.
- QR membantu lookup anggota/siswa.
- Transaksi pinjam/kembali tetap butuh konfirmasi petugas.

Flow pinjam:

1. Staf login ke modul perpustakaan.
2. Scan kartu siswa.
3. Server cek kartu aktif + siswa aktif + status anggota.
4. Scan/input barcode buku.
5. Server cek ketersediaan dan limit pinjam.
6. Petugas konfirmasi transaksi.
7. Audit `library_used`.

Flow kunjungan:

- Bisa lebih cepat: scan kartu mencatat kunjungan jika kartu aktif.
- Tetap idempotent per hari/window agar scan ulang tidak menggandakan statistik.

## 10. Role Scope dan Permission

Permission yang direkomendasikan:

- `student_card.read`: melihat kartu dan status.
- `student_card.issue`: menerbitkan kartu baru.
- `student_card.print`: mencetak kartu/QR.
- `student_card.activate`: aktivasi kartu.
- `student_card.suspend`: suspend sementara.
- `student_card.revoke`: revoke permanen.
- `student_card.reissue`: ganti kartu/token.
- `student_card.reset_pin`: reset PIN portal siswa.
- `student_card.audit.read`: melihat audit.
- `student_card.scan.attendance`: memakai scan untuk presensi.
- `student_card.scan.library`: memakai scan untuk perpustakaan.
- `student_card.scan.cbt`: memakai scan untuk verifikasi identitas CBT.

Mapping awal:

- `admin`: semua permission.
- `kesiswaan`: issue/print/activate/reissue/revoke/reset PIN/audit untuk siswa.
- `staf perpustakaan`: scan library dan read terbatas identitas siswa.
- `guru/wali kelas/piket`: scan attendance sesuai scope kelas/jadwal.
- `panitia/proktor/pengawas CBT`: scan CBT sesuai event/session/room assignment.
- `siswa`: melihat status kartunya sendiri, reset/ganti PIN sendiri jika authenticated; tidak bisa revoke sendiri tanpa workflow laporan.
- `ortu`: read-only status kartu/riwayat terbatas anak jika portal ortu aktif.

Scope penting:

- Role global memberi kemampuan modul.
- Scope kelas/event/ruang menentukan objek mana yang boleh diakses.
- Backend wajib enforce scope; UI/sidebar hanya bantuan UX.

## 11. Endpoint API Konseptual

Semua endpoint berada di Core API; SvelteKit hanya proxy/BFF.

Admin/kesiswaan:

- `POST /api/student-cards` buat kartu.
- `POST /api/student-cards/{id}/print-token` hanya untuk proses cetak terkontrol; sebaiknya one-time.
- `POST /api/student-cards/{id}/activate`.
- `POST /api/student-cards/{id}/suspend`.
- `POST /api/student-cards/{id}/revoke`.
- `POST /api/students/{studentId}/cards/reissue`.
- `POST /api/students/{studentId}/portal-pin/reset`.
- `GET /api/student-cards/{id}/audit`.

Scan lintas modul:

- `POST /api/card-scans/resolve` validasi token untuk konteks tertentu.
- `POST /api/attendance/scans` catat presensi dari QR.
- `POST /api/library/member-scans` resolve anggota perpustakaan.
- `POST /api/cbt/sessions/{sessionId}/identity-scan` verifikasi peserta CBT.

Portal siswa:

- `POST /api/student-portal/login` login manual.
- `POST /api/student-portal/card-login/start` mulai challenge dari QR.
- `POST /api/student-portal/card-login/verify-pin` validasi PIN dan buat session.

## 12. Audit, Monitoring, dan Deteksi Penyalahgunaan

Wajib diaudit:

- Penerbitan, cetak, aktivasi, revoke, reissue, suspend.
- Reset PIN dan login portal berhasil/gagal.
- Semua scan denied untuk kartu inactive/revoked/lost.
- Scan valid di presensi/perpustakaan/CBT.
- Perubahan role/permission/scope yang berkaitan dengan kartu.

Sinyal anomali:

- Kartu yang sama discan dari dua lokasi/perangkat dalam waktu dekat.
- Banyak gagal PIN setelah scan QR.
- Scan kartu revoked/lost berulang.
- Petugas melakukan scan siswa di luar scope kelas/ruang.
- Volume scan terlalu tinggi dari satu device.

Tindakan otomatis opsional:

- Lock challenge portal sementara.
- Tandai kartu `suspicious`/`suspended` untuk review.
- Kirim notifikasi ke admin/kesiswaan.

## 13. Rekomendasi Implementasi Bertahap

### Tahap 1 — Fondasi Aman

- Tambah tabel `student_cards`, `student_card_events`, dan credential portal jika belum ada.
- Generate QR random + hash token.
- Lifecycle: issue, print, activate, revoke, reissue.
- Audit lifecycle.

### Tahap 2 — Portal Siswa Aman

- Login manual NIS/NISN + PIN/password.
- QR sebagai shortcut identitas + PIN, bukan login langsung.
- Rate limit dan lockout.
- Reset PIN oleh admin/kesiswaan.

### Tahap 3 — Presensi dan Perpustakaan

- Endpoint scan dengan context `attendance` dan `library`.
- Idempotency presensi/kunjungan.
- Scope role petugas.
- Audit scan allowed/denied.

### Tahap 4 — CBT Tambahan

- Integrasi scan kartu untuk verifikasi peserta oleh pengawas/proktor.
- Tetap gunakan token ruang/sesi CBT terpisah.
- Audit per event/session/room.

### Tahap 5 — Monitoring dan Hardening

- Dashboard kartu hilang/revoked/suspicious.
- Deteksi anomali scan.
- Report audit untuk kesiswaan/admin.
- Rotasi server secret HMAC dengan versi token/hash jika diperlukan.

## 14. Keputusan Desain Kunci

- QR ID card = token identitas kartu yang random dan revocable.
- Portal login = PIN/password wajib; QR hanya shortcut.
- CBT = token ruang/sesi tetap wajib dan terpisah dari QR ID card.
- Reset PIN ≠ reissue kartu.
- Reissue kartu selalu revoke/replace token lama.
- Presensi/perpustakaan menggunakan QR dengan role scope petugas dan audit.
- Backend/Core API menjadi satu-satunya boundary keamanan dan pemilik data.
