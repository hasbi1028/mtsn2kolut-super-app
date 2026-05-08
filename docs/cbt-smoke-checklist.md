# CBT Smoke Checklist - Phase 5 Rehearsal and Rollout

Status: Phase 27 Operator Rehearsal Workflow Completion checklist, 2026-05-08.

Checklist ini dipakai untuk Phase 5 - Rehearsal, Rollout, and Post-Exam Review sebelum ujian besar dan untuk post-exam review setelah sesi selesai. Alur yang diuji adalah Bank Soal -> Asesmen Persiapan -> Pelaksanaan/Pengawasan -> Flutter APK -> Hasil/Post-exam review.

Phase 14 Operator Rehearsal and Proctor Evidence mengunci alur final: Bank Soal to Asesmen Persiapan to Pelaksanaan/Pengawasan to Flutter APK to Hasil/Post-exam review. Phase 27 Operator Rehearsal Workflow Completion memperluasnya menjadi checklist evidence final dengan proctor evidence, role/scope/token boundary, event/audit evidence, go/no-go rehearsal, backup/DR context, security control-alignment notes, dan mobile RC package reference tanpa deploy, migrasi, live DB write, atau route runtime baru.

Dokumen ini adalah `docs/cbt-smoke-checklist.md`. Gunakan bersama `docs/cbt-proposal-integration-phase-5.md`, `docs/exam-api.md`, dan `apps/mobile/RELEASE_CHECKLIST.md`.

Checklist ini mempertahankan guard operasional lama: event member role dan subject scope, export/detail Bank Soal admin/guru/guru non-penulis, berita acara/minutes token visibility, duplicate `(room_id, seat_no)`, serta endpoint/security boundary checks.

## Phase 14 Operator Rehearsal and Proctor Evidence

- [ ] Bank Soal to Asesmen Persiapan to Pelaksanaan/Pengawasan to Flutter APK to Hasil/Post-exam review selesai dari route canonical.
- [ ] proctor evidence memuat sesi, ruang, peserta uji, perangkat, heartbeat/status, warning, pengawas, operator, dan reviewer.
- [ ] role/scope/token boundary diverifikasi untuk admin/panitia, guru, pengawas/proktor, token visibility, kunci jawaban, dan rubrik.
- [ ] event/audit evidence tersedia untuk Bank Soal, Asesmen Persiapan, Pelaksanaan/Pengawasan, Flutter APK exam events, dan Hasil/Post-exam review.
- [ ] go/no-go rehearsal dicatat dengan operator, pengawas, reviewer, rollback owner, blockers, accepted operational notes, dan follow-up owner.

## Phase 27-30 Evidence References

- [ ] Phase 27 Operator Rehearsal Workflow Completion evidence bundle location dicatat.
- [ ] Phase 28 backup/DR evidence reference dicatat tanpa restore over live DB.
- [ ] Phase 29 security and ISO-control alignment evidence dicatat sebagai control alignment, not certification.
- [ ] Phase 30 mobile RC build/release package evidence mencatat APK SHA-256 hash/status bila build tersedia.

## Boundary Wajib

- [ ] Tidak deploy, tidak PM2 restart, dan tidak mengganti process manager dari checklist ini.
- [ ] Tidak menjalankan `make db-migrate`, migrasi live, atau ad hoc `ALTER TABLE` dari checklist ini.
- [ ] Tidak membuat public SvelteKit route tree `/api/cbt/**` baru; namespace itu hanya compatibility/deprecated path yang sudah ada.
- [ ] Tidak membawa PocketBase, SQLite, atau Alpine menjadi runtime CBT.
- [ ] Flutter berbicara langsung ke `services/core-api` melalui `/api/exam/*`, bukan melalui SvelteKit BFF.
- [ ] Rehearsal memakai data uji atau sesi yang memang disetujui operator; jangan memakai token ujian aktif tanpa izin panitia.
- [ ] BYOD tidak setara kiosk penuh; device-owner bukan baseline untuk perangkat siswa pribadi.

## 1. Preflight Operasional

Catat semua hasil preflight sebelum membuka rehearsal.

- [ ] Environment dicatat: staging/produksi, tanggal, jam WITA, operator, pengawas, dan ruang uji.
- [ ] Commit backend, web-admin, dan APK Flutter yang diuji dicatat.
- [ ] Health check backend hijau dan response tidak menunjukkan koneksi database bermasalah.
- [ ] Audit log dapat dibaca oleh role berwenang dan mutasi rehearsal terbaru masuk sebagai event/audit yang wajar.
- [ ] Server log backend dapat diakses oleh operator teknis, tidak menunjukkan panic berulang, error database, atau raw secret/token.
- [ ] Server log frontend dapat diakses bila perlu untuk investigasi BFF, tidak menunjukkan proxy loop atau auth cookie error berulang.
- [ ] Backup PostgreSQL terbaru sudah ada, timestamp-nya sesuai jendela release yang disetujui, dan lokasi/penanggung jawab restore dicatat.
- [ ] Backup path, latest symlink, checksum, dan hasil `sha256sum -c` atau alasan skipped dicatat.
- [ ] Jika dump custom PostgreSQL tersedia, hasil `pg_restore --list` atau alasan skipped dicatat.
- [ ] Do not restore over live DB; restore rehearsal hanya boleh ke isolated scratch database/offline target dengan owner eksplisit.
- [ ] Migration state sudah diverifikasi sebagai sesuai release yang diuji; bila ada migration pending atau status tidak jelas, hentikan rehearsal dan eskalasi.
- [ ] Tidak ada perintah migrasi, deploy, PM2 restart, atau rollback database yang dijalankan sebagai bagian checklist ini.

## 2. Akun, Data Uji, dan Scope

- [ ] Akun `admin` atau panitia tersedia untuk setup penuh.
- [ ] Akun `guru` tersedia untuk authoring/review soal sesuai mapel.
- [ ] Akun pengawas/proktor tersedia untuk dashboard pelaksanaan.
- [ ] Minimal dua peserta uji tersedia pada kelas/scope yang benar.
- [ ] Minimal satu ruang dan seat peserta dapat diverifikasi.
- [ ] Ada minimal satu soal objektif dan satu soal uraian untuk menguji kunci, rubrik, scoring, dan status submit.
- [ ] Operator memahami bahwa data rehearsal yang menyentuh siswa nyata tetap harus diperlakukan sebagai data sekolah.

## 3. Event Member Role dan Subject Scope

Verifikasi assignment operasional sebelum Bank Soal/Pelaksanaan diuji.

- [ ] `panitia` terdaftar sebagai pengelola event tanpa `subject_id`.
- [ ] `pembuat_soal` terdaftar sesuai mapel yang ditugaskan.
- [ ] `reviewer` terdaftar sesuai mapel yang direview.
- [ ] `proktor` terdaftar atau ditetapkan operasional tanpa `subject_id`.
- [ ] `pengawas` terdaftar atau ditetapkan operasional tanpa `subject_id`.
- [ ] `korektor` terdaftar sesuai mapel yang memiliki essay/uraian.
- [ ] Tidak ada assignment dobel untuk kombinasi event, user, role, dan subject yang sama.
- [ ] Role `panitia`, `proktor`, dan `pengawas` tidak membawa `subject_id`; subject scope hanya dipakai untuk `pembuat_soal`, `reviewer`, dan `korektor`.

## 4. Bank Soal

Jalankan dari route canonical `/bank-soal/*`.

- [ ] Buat soal baru melalui `/bank-soal/tambah` atau pilih soal rehearsal yang sudah ada.
- [ ] Bila memakai impor, gunakan `/bank-soal/impor`; jangan edit database langsung.
- [ ] Verifikasi/review soal melalui `/bank-soal/verifikasi`.
- [ ] Pastikan rich content, KaTeX, RTL/Arab bila relevan, media, dan preview tampil.
- [ ] Admin export menampilkan label `Export CSV`, berisi soal sesuai filter admin, dan kolom kunci/rubrik hanya sesuai hak admin.
- [ ] Guru penulis export menampilkan label `Export Soal Saya`, hanya berisi soal milik/scope guru tersebut, dan tidak membuka soal guru lain.
- [ ] Admin dapat melihat kunci jawaban/rubrik pada detail soal sesuai kewenangan.
- [ ] Guru penulis dapat melihat kunci/rubrik soal sendiri sesuai workflow edit/review yang berlaku.
- [ ] Guru non-penulis tidak melihat kunci jawaban atau rubrik soal di luar scope, termasuk melalui detail dan export.

Kriteria lulus:

- [ ] Soal masuk paket dari workflow web-admin yang sah.
- [ ] Kunci jawaban hanya terlihat untuk role/scope yang berwenang.
- [ ] Tidak ada fetch baru ke public `/api/cbt/**` untuk alur Bank Soal.

## 5. Asesmen Persiapan

Jalankan dari `/asesmen/persiapan` dan route canonical terkait.

- [ ] Paket dibuat dari soal final/reviewed.
- [ ] Kegiatan/event dibuat dengan timeline WITA yang benar.
- [ ] Sesi ujian memakai paket final, jadwal benar, durasi benar, dan status belum dibuka sebelum siap.
- [ ] Peserta, ruang, seat, proktor, dan pengawas sesuai daftar panitia.
- [ ] Jika memakai seat, semua `seat_no` bernilai positif dan tidak ada duplikasi `(room_id, seat_no)` dalam ruang yang sama.
- [ ] Kartu/token dicetak setelah perubahan token/room/seat terakhir.
- [ ] Token ujian yang dibagikan untuk uji adalah token backend terbaru.
- [ ] Data persiapan tidak dibuat melalui PocketBase, SQLite, import runtime liar, atau edit SQL manual.

Kriteria lulus:

- [ ] Sesi rehearsal siap dibuka tanpa data peserta dobel.
- [ ] Kartu/token, ruang, dan seat cocok dengan web-admin.
- [ ] Tidak ada duplikasi `(room_id, seat_no)` pada ruang yang sama.
- [ ] Audit log mencatat mutasi setup penting.

## 6. Pelaksanaan dan Pengawasan

Jalankan dari `/asesmen/pelaksanaan` dan `/asesmen/pengawasan`.

- [ ] Dashboard pengawas menampilkan sesi, ruang, peserta, status login, heartbeat, dan submit.
- [ ] Pengawas dapat membedakan status koneksi Flutter: `Tersambung`, `Lokal`, `Gangguan`, `Waspada`, dan `Menurun`.
- [ ] Simulasikan app-switch/resume pada satu perangkat uji.
- [ ] Simulasikan network disturbance singkat bila aman.
- [ ] Pastikan warning/event BYOD terlihat di dashboard atau evidence backend.
- [ ] Admin dapat melihat/mencetak token peserta pada berita acara/minutes untuk kebutuhan operasional sah.
- [ ] Guru/pengawas hanya melihat peserta sesuai scope sesi/ruang dan token peserta tidak tampil atau bernilai kosong bila role tidak berwenang.
- [ ] Pastikan token peserta tidak bocor ke role yang tidak berwenang.

Kriteria lulus:

- [ ] Status peserta berubah tanpa menghapus jawaban lokal yang belum sync.
- [ ] Warning BYOD muncul sebagai evidence, bukan klaim kiosk penuh.
- [ ] Pengawas punya prosedur eskalasi bila perangkat tetap `Menurun`.
- [ ] Berita acara/minutes menjaga boundary token: admin operasional boleh mencetak, role lain hanya sesuai kewenangan.

## 7. Flutter APK dan Perangkat Nyata

Gunakan `apps/mobile/RELEASE_CHECKLIST.md` sebagai checklist build/release. Flutter SDK bisa tidak tersedia di host yang menjalankan checklist ini; bila begitu, catat sebagai blocker validasi mobile dan jalankan quality gate di mesin lain/CI sebelum APK dipakai untuk ujian resmi.

- [ ] APK yang diuji dicatat versi, commit, signing mode, dan `API_BASE_URL`.
- [ ] Phase 30 Mobile RC Build and Release Package evidence mencatat RC identifier, version name/code, commit hash, signing status, APK path, APK file size, APK SHA-256 hash, `API_BASE_URL`, dan build log/status.
- [ ] Build/hash evidence tidak dipakai sebagai real-device PASS.
- [ ] real-device PASS claimed only after physical Android operator test pada minimum two Android vendors.
- [ ] Minimal dua perangkat nyata lintas vendor Android dipakai.
- [ ] Hasil perangkat dicatat di `apps/mobile/DEVICE_TEST_MATRIX.md`.
- [ ] Login token berhasil pada perangkat pertama.
- [ ] Device mismatch untuk token yang sudah terikat menghasilkan error terkendali.
- [ ] Heartbeat berhasil dan timestamp kontak terakhir berubah.
- [ ] Answer save untuk pilihan ganda dan uraian berhasil.
- [ ] Restore setelah app ditutup/buka lagi melewati status check.
- [ ] Network disturbance memunculkan warning yang dapat dipahami siswa/pengawas.
- [ ] Submit final berhasil sekali pada koneksi yang sehat.
- [ ] Duplicate submit ditolak sebagai konflik terkendali.
- [ ] Payload/event tidak memuat token ujian mentah, password, atau answer key.
- [ ] Asset soal hanya terbuka untuk JWT role berwenang atau exam-token peserta aktif sesuai paketnya.
- [ ] Endpoint `answer`, `submit`, `heartbeat`, dan `event` menolak request tanpa konteks peserta yang sah.

Kriteria lulus:

- [ ] Flutter berbicara langsung ke `services/core-api` melalui `/api/exam/*`.
- [ ] Tidak ada runtime siswa melalui public `/api/cbt/login` atau `/api/cbt/status`.
- [ ] BYOD limitation tercatat: `FLAG_SECURE` dan telemetry adalah deterrence/evidence, bukan kiosk guarantee.

## 8. Hasil dan Post-Exam Review

Jalankan setelah peserta uji submit atau sesi ditutup sesuai prosedur.

- [ ] `/asesmen/hasil` menampilkan peserta, status submit, skor objektif, dan follow-up uraian.
- [ ] Koreksi uraian tidak menandai peserta belum submit sebagai sudah submit.
- [ ] Export hasil diuji dengan scope admin/guru yang sesuai.
- [ ] Event anti-cheat, heartbeat risk, device mismatch, dan network warning direview bersama pengawas.
- [ ] Audit log mutasi penting direview: setup sesi, perubahan token/room/seat, submit, scoring, export.
- [ ] Server log dicek ulang untuk 500, panic, request oversized, atau auth/rate-limit anomaly selama rehearsal.
- [ ] Temuan dipisah menjadi blocker ujian, catatan operasional, dan backlog produk.

Kriteria lulus:

- [ ] Data hasil dapat direview tanpa membuka kunci jawaban ke role yang tidak sah.
- [ ] Evidence post-exam cukup untuk keputusan lanjut/tunda.
- [ ] Tidak ada raw internal error detail muncul di response pengguna.

## 9. Rollback Plan

Rollback plan harus disetujui sebelum rehearsal dianggap lulus.

- [ ] Jika login token, answer save, submit, atau proctor dashboard gagal kritis: hentikan sesi baru dan jangan buka gelombang berikutnya.
- [ ] Jika APK release bermasalah: distribusikan APK sebelumnya yang sudah lulus rehearsal, lalu ulangi device test minimal.
- [ ] Jika web-admin bermasalah tetapi backend dan Flutter masih aman: tahan perubahan operasional baru, pakai prosedur pengawas yang sudah disetujui, dan eskalasi teknis.
- [ ] Jika backend runtime ujian bermasalah: pertahankan data backend, jangan hapus jawaban/token/audit, dan ambil evidence log untuk perbaikan.
- [ ] Jangan rollback database dengan ad hoc SQL. Ikuti prosedur release resmi dan backup/restore yang disetujui penanggung jawab backend.
- [ ] Komunikasikan status ke operator, pengawas, dan panitia sebelum siswa berikutnya menerima token.

## 10. Acceptance Evidence

Lampirkan evidence berikut pada tiket/release note internal.

- [ ] Tanggal, jam WITA, environment, operator, dan pengawas.
- [ ] Commit backend, web-admin, dan APK Flutter.
- [ ] Hasil health check, audit log, server log, backup, dan migration state.
- [ ] Screenshot Bank Soal, Asesmen Persiapan, Pelaksanaan/Pengawasan, dan Hasil.
- [ ] Event/paket/sesi/ruang/peserta uji yang dipakai.
- [ ] Kartu/token yang diuji dan catatan bila token/room/seat berubah.
- [ ] Device matrix ringkas untuk perangkat nyata.
- [ ] Hasil login token, heartbeat, answer save, restore, network disturbance, warning, dan submit.
- [ ] Hasil review token/kunci jawaban tidak bocor.
- [ ] Rollback plan yang dipakai bila ada temuan.
- [ ] Keputusan akhir: lulus, lulus dengan catatan, atau tunda.

Keputusan wajib tunda bila ada kebocoran token/kunci jawaban, jawaban tidak tersimpan, submit final tidak dapat dipercaya, status peserta salah, role melihat data lintas scope, atau migration state tidak jelas.

## Phase 31 production candidate smoke evidence

- [x] Backend `/health`: `200`.
- [x] Web root `http://127.0.0.1:8021/`: `200`.
- [x] Protected `/settings/rbac`: `302` unauthenticated redirect.
- [x] Protected `/api/bank-soal/summary`: `401` unauthenticated.
- [x] Core API `/api/exam/status` without token: `401`.
- [x] `/bank-soal`: `302` unauthenticated redirect.
- [x] `/asesmen`: `302` unauthenticated redirect.
- [x] `bash deploy/scripts/health-check.sh all`: PASS.
