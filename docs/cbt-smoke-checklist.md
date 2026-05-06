# CBT Smoke Checklist — Readiness, Visibility, Runtime, and Security

Status: sinkron per 2026-05-06. Gunakan checklist ini sebelum ujian besar, setelah deploy backend/frontend CBT, setelah migration CBT, dan sebelum sesi aktif dibuka untuk siswa.

Checklist ini melengkapi test otomatis dan `docs/cbt-operator-runbook.md`. Fokusnya adalah memastikan setup ujian siap, batas role admin/guru/event member benar, token/kunci jawaban tidak bocor, Flutter bisa menjalankan alur kritis, dan runtime submit/scoring tidak regress.

## Phase 1 — Readiness Setup

### 1. Deploy dan Migration

- [ ] Backend Go sudah deploy dan health check hijau.
- [ ] Migration CBT terbaru dijalankan dari backend path, bukan dari frontend/worker.
- [ ] Backup PostgreSQL sudah dibuat sebelum migration dengan `make ops-backup` atau script backup setara.
- [ ] Tidak ada jendela ujian aktif saat migration berjalan.
- [ ] Frontend web-admin sudah deploy dari commit yang sama atau lebih baru dari backend.
- [ ] Bookmark lama `/cbt/questions`, `/cbt/soal*`, dan `/bank-soal/komposer|import|review` redirect ke `/bank-soal/*` baru sesuai mode aman yang dipertahankan.
- [ ] `make ops-health` atau health check manual sudah hijau setelah restart.

Catatan token migration:

- [ ] Operator memahami bahwa `060_cbt_exam_token_hardening.sql` tidak merotasi token yang sudah ada.
- [ ] Cetak ulang kartu hanya dilakukan setelah regenerasi/repair token eksplisit, atau setelah perubahan data kartu seperti ruang/seat.
- [ ] Jika migration 060 gagal karena duplikasi token, hanya token sesi draft/scheduled yang terdampak yang diregenerasi/diperbaiki, lalu kartu terdampak dicetak ulang.

### 2. Data Uji dan Akun Role

- [ ] Tersedia akun `admin` sebagai panitia/operator.
- [ ] Tersedia `guru_a` sebagai pembuat minimal satu soal.
- [ ] Tersedia `guru_b` sebagai guru lain pada mapel/scope berbeda atau bukan penulis soal `guru_a`.
- [ ] Tersedia akun/penetapan operasional untuk proktor atau pengawas jika dashboard ruang diuji.
- [ ] Tersedia minimal dua peserta uji dari kelas/scope yang benar.
- [ ] Tersedia minimal satu ruang uji dengan seat yang bisa diverifikasi.
- [ ] Tersedia minimal satu soal objektif dan satu soal essay/uraian untuk menguji kunci, rubrik, grading, dan skor.

### 3. Event Member Role dan Subject Scope

Verifikasi role dari migration `062_cbt_event_members_question_scope.sql`:

- [ ] `panitia` terdaftar sebagai pengelola event tanpa `subject_id`.
- [ ] `pembuat_soal` terdaftar sesuai mapel yang ditugaskan.
- [ ] `reviewer` terdaftar sesuai mapel yang direview.
- [ ] `proktor` terdaftar atau ditetapkan operasional tanpa `subject_id`.
- [ ] `pengawas` terdaftar atau ditetapkan operasional tanpa `subject_id`.
- [ ] `korektor` terdaftar sesuai mapel yang memiliki essay/uraian.
- [ ] Tidak ada assignment dobel untuk kombinasi event, user, role, dan subject yang sama.
- [ ] Role `panitia`, `proktor`, dan `pengawas` tidak membawa `subject_id`; subject scope hanya dipakai untuk `pembuat_soal`, `reviewer`, dan `korektor`.

### 4. Master Data, Event, Paket, dan Sesi

- [ ] Tahun ajaran, kelas, mapel, siswa, guru, dan assignment guru-mapel sudah benar.
- [ ] Event CBT sudah dibuat dan memuat timeline yang sesuai.
- [ ] Soal sudah diinput melalui `/bank-soal/tambah` atau diimpor melalui `/bank-soal/impor`, bukan melalui edit database manual.
- [ ] Soal yang masuk paket sudah direview/published sesuai workflow.
- [ ] Paket ujian berisi soal sesuai mapel, kelas/grade/custom scope, bobot, dan jumlah soal yang disepakati.
- [ ] Sesi ujian memakai paket final, jadwal WITA benar, dan belum aktif sebelum token/room/seat siap.
- [ ] Peserta sesi sesuai daftar panitia dan tidak dobel.

### 5. Ruang, Seat, Proktor, dan Kartu

- [ ] Setiap peserta uji punya ruang yang benar.
- [ ] Jika memakai seat, semua `seat_no` bernilai positif.
- [ ] Tidak ada duplikasi `(room_id, seat_no)` dalam ruang yang sama.
- [ ] Proktor/pengawas menerima daftar ruang, daftar hadir, denah/seat, kontak eskalasi, dan prosedur reset akses.
- [ ] Token baru yang dibagikan untuk uji adalah 32 karakter hex.
- [ ] Kartu/token dicetak dari data terbaru setelah perubahan token/room/seat terakhir.

### 6. Flutter Release APK dan BYOD Trial Setup

- [ ] APK release dibangun dengan HTTPS `API_BASE_URL`, misalnya `make mobile-release-apk API_BASE_URL=https://api.sekolah.example`.
- [ ] `flutter analyze` atau `make check-mobile` lulus.
- [ ] `flutter test` atau `make test-mobile` lulus.
- [ ] Permission Android `INTERNET` terverifikasi pada manifest/APK.
- [ ] APK signing sesuai prosedur release internal sekolah dan versi build dicatat.
- [ ] Minimal dua sampai tiga perangkat nyata lintas vendor tersedia untuk trial.
- [ ] `apps/mobile/DEVICE_TEST_MATRIX.md` disiapkan untuk mencatat hasil perangkat.
- [ ] Pengawas memahami label koneksi `Tersambung`, `Sinkron`, `Cek Ulang`, `Lokal`, `Gangguan`, `Waspada`, dan `Menurun`.

## Phase 2 — Runtime and Security Smoke

### 7. Bank Soal: Label dan Export Admin

1. Login sebagai `admin`.
2. Buka `/bank-soal`.
3. Pastikan tombol export tertulis `Export CSV`.
4. Jalankan export dengan filter aktif yang kecil, misalnya satu mapel.
5. Buka CSV hasil download.

Kriteria lulus:

- [ ] CSV berhasil diunduh.
- [ ] CSV berisi soal sesuai filter admin.
- [ ] Kolom kunci/jawaban terisi sesuai data soal.
- [ ] Tidak ada error toast atau response 403.

### 8. Bank Soal: Label dan Export Guru

1. Login sebagai `guru_a`.
2. Buka `/bank-soal`.
3. Pastikan tombol export tertulis `Export Soal Saya`.
4. Jalankan export pada filter yang sama.
5. Buka CSV hasil download.

Kriteria lulus:

- [ ] CSV berhasil diunduh.
- [ ] CSV hanya berisi soal dengan penulis `guru_a` atau scope yang memang diizinkan.
- [ ] Soal milik `guru_b` atau guru lain tidak ikut keluar bila tidak termasuk scope.
- [ ] Toast sukses menyebut `soal saya`.
- [ ] Tidak ada kunci/token lintas scope yang ikut terbuka.

### 9. Bank Soal: Kunci Jawaban Detail

Admin:

1. Login sebagai `admin`.
2. Buka detail/edit soal milik `guru_a`.

Kriteria lulus:

- [ ] Kunci jawaban/rubrik terlihat sesuai data soal.

Guru penulis:

1. Login sebagai `guru_a`.
2. Buka soal yang ditulis oleh `guru_a`.

Kriteria lulus:

- [ ] Kunci jawaban soal sendiri terlihat.
- [ ] Edit tetap mengikuti guard soal terkunci/published sesuai workflow.

Guru non-penulis:

1. Login sebagai `guru_b`.
2. Buka detail soal milik `guru_a` dari route yang masih bisa dibaca oleh scope guru.

Kriteria lulus:

- [ ] Kunci jawaban kosong/tidak terbuka.
- [ ] Metadata yang aman tetap terbaca bila memang guru punya akses baca.
- [ ] Tidak ada cara melihat kunci melalui export guru.

### 10. Berita Acara / Minutes: Token Peserta

Admin:

1. Login sebagai `admin`.
2. Buka detail sesi uji.
3. Buka/cetak berita acara atau endpoint minutes.

Kriteria lulus:

- [ ] Token peserta terlihat untuk kebutuhan cetak operasional.
- [ ] Peserta, ruang, dan seat tampil sesuai sesi.

Guru/pengawas:

1. Login sebagai guru/pengawas yang hanya berhak pada sesi/ruang terkait.
2. Buka berita acara atau data peserta sesi.

Kriteria lulus:

- [ ] Token peserta tidak tampil atau bernilai kosong.
- [ ] Peserta yang tampil sesuai scope guru/pengawas, bukan semua peserta lintas sesi/mapel.

### 11. Flutter Token Login dan BYOD Runtime

1. Pakai kartu ujian/token terbaru setelah regenerasi/repair eksplisit atau perubahan room/seat terakhir.
2. Login dari perangkat pertama dengan base URL HTTPS.
3. Coba login token yang sama dari perangkat kedua.
4. Buka payload soal, termasuk rich content/media/audio bila sesi uji memilikinya.
5. Lakukan heartbeat, simpan jawaban, app-switch/resume, dan status refresh.
6. Submit final dari perangkat yang valid.

Kriteria lulus:

- [ ] Token 32 karakter dapat login pada perangkat pertama.
- [ ] Token yang sudah device-bound tidak dapat dipakai perangkat lain tanpa prosedur reset yang sah.
- [ ] Payload soal tampil dan tidak membocorkan kunci jawaban.
- [ ] Heartbeat dan simpan jawaban sukses.
- [ ] Resume gate dan status koneksi menampilkan guidance yang dapat dipahami.
- [ ] Submit sukses hanya sekali.
- [ ] Hasil perangkat dicatat di `apps/mobile/DEVICE_TEST_MATRIX.md`.

### 12. Runtime: Skor, Grading, dan Status Submit

1. Pastikan ada peserta uji yang sudah mulai ujian tetapi belum final submit.
2. Dari admin/korektor berwenang, nilai satu jawaban essay peserta tersebut.
3. Jalankan hitung skor sesi jika tersedia.
4. Cek daftar peserta dan status peserta.
5. Minta peserta melanjutkan ujian atau menyimpan jawaban lagi.

Kriteria lulus:

- [ ] Grading essay tidak mengisi `submitted_at` untuk peserta yang belum submit.
- [ ] Hitung skor tidak mengisi `submitted_at` untuk peserta yang belum submit.
- [ ] Peserta belum submit masih bisa menyimpan jawaban dan final submit sesuai jadwal.
- [ ] Duplicate final submit ditolak sebagai konflik yang terkendali, bukan 500.

### 13. Runtime: Room, Seat, and Proctoring Integrity

1. Buka detail sesi/ruang sebagai admin atau proktor/pengawas berwenang.
2. Cocokkan peserta, room, seat, dan daftar hadir.
3. Pantau heartbeat/status koneksi selama perangkat uji login.
4. Simulasikan satu gangguan ringan seperti app-switch atau koneksi putus sementara bila aman.

Kriteria lulus:

- [ ] Dashboard/detail ruang menampilkan peserta sesuai scope.
- [ ] Seat tidak dobel dalam ruang.
- [ ] Status koneksi berubah secara wajar dan tidak menghapus jawaban lokal yang belum sync.
- [ ] Warning event/guidance muncul untuk kondisi BYOD yang berisiko.

### 14. Security Boundary Checks

- [ ] Guru non-penulis tidak melihat kunci jawaban soal orang lain.
- [ ] Guru/pengawas tidak melihat token peserta kecuali role/scope memang mengizinkan secara eksplisit.
- [ ] Flutter payload siswa tidak memuat kunci jawaban, token peserta lain, atau metadata admin.
- [ ] Asset soal hanya terbuka untuk JWT role berwenang atau exam-token peserta aktif sesuai paketnya.
- [ ] Login token gagal dengan status yang terkendali untuk token tidak ada, sesi tidak aktif, device mismatch, atau participant context hilang.
- [ ] Endpoint answer/submit/heartbeat/event tidak menerima request tanpa konteks peserta yang sah.

## Phase 3 — Evidence and Decision

### 15. Evidence Wajib Dicatat

Catat hasil smoke test dalam tiket/deployment note:

- tanggal dan jam uji
- environment: staging atau produksi
- commit backend/frontend/mobile APK
- akun role yang dipakai
- event, paket, sesi, ruang, dan peserta uji
- hasil preflight token/seat bila migration dijalankan
- hasil export admin dan guru
- screenshot label export admin/guru
- screenshot detail soal admin/guru penulis/guru non-penulis
- screenshot berita acara/minutes admin vs guru/pengawas
- hasil uji token Flutter, heartbeat, answer save, resume, dan submit
- hasil uji grading sebelum submit
- device matrix ringkas untuk perangkat trial
- daftar temuan dan keputusan: lulus / lulus dengan catatan / tunda deploy

### 16. Keputusan

- [ ] Lulus: boleh lanjut ujian/deploy.
- [ ] Lulus dengan catatan: boleh lanjut jika temuan tidak menyentuh token, kunci jawaban, submit, sinkron jawaban, atau akses lintas role.
- [ ] Tunda: wajib jika ada kebocoran kunci/token, peserta belum submit terkunci, token Flutter tidak bisa login/simpan/submit, seat peserta rancu, atau migration preflight gagal.
