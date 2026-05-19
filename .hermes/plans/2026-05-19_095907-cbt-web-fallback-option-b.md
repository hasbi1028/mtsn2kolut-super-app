# Plan Aman — CBT Web/PWA Fallback Opsi B

Tanggal: 2026-05-19 09:59 WITA
Repo: `/home/servermtsn2kolut/mtsn2kolut-super-app`
Status: Plan only, belum implementasi.

## 1. Tujuan

Membuat jalur cadangan CBT berbasis web/PWA yang aman dijalankan sebagai fallback darurat untuk siswa, tanpa menggantikan Flutter Android/Windows sebagai jalur resmi utama.

Prinsip utama:

- Flutter app tetap primary exam client.
- Web/PWA fallback hanya aktif untuk sesi yang diizinkan operator/admin.
- Semua akses tetap lewat Go Core API dan BFF SvelteKit; tidak ada akses DB langsung dari frontend.
- Web fallback tidak diiklankan sebagai anti-cheat setara aplikasi native.
- Setiap penggunaan fallback web harus terlihat di panel pengawas dan tercatat di audit/telemetry.

## 2. Konteks Saat Ini

Dari repo saat ini:

- Flutter CBT client ada di `apps/mobile`.
- Runtime student client memakai endpoint `/api/exam/*`:
  - login,
  - status,
  - heartbeat,
  - event,
  - answer,
  - submit.
- Web Admin SvelteKit sudah punya BFF route `apps/web-admin/src/routes/api/exam/[...path]/+server.ts`.
- Halaman operator rilis aplikasi siswa ada di `apps/web-admin/src/routes/asesmen/aplikasi-siswa/release/+page.svelte`.
- Arsitektur repo mewajibkan PostgreSQL hanya dimiliki `services/core-api`; SvelteKit hanya BFF/proxy.
- Anti-cheat Flutter native lebih kuat dari web karena ada MethodChannel/EventChannel Android/Windows. Web fallback hanya bisa best-effort: visibility/focus/fullscreen/heartbeat/audit.

## 3. Definisi Aman untuk Dijalankan

Plan ini aman jika implementasi mengikuti batasan berikut:

1. Fallback web default OFF.
2. Hanya admin/operator berwenang yang bisa mengaktifkan fallback per kegiatan/sesi/ruang.
3. Peserta web fallback tetap butuh token ujian dan token ruang.
4. Pengawas melihat label jelas: `Web Fallback` / `Browser Darurat`.
5. Client type dikirim dan dicatat: `web_fallback`.
6. Event web tidak langsung menghukum/lock secara agresif; hanya memberi sinyal ke pengawas kecuali pelanggaran berat seperti device mismatch/token mismatch.
7. Tidak ada jawaban/token ditampilkan di log frontend/backend.
8. Local pending answer hanya untuk pemulihan koneksi dan dibersihkan setelah submit/keluar sah.
9. Deploy dilakukan bertahap: backend flag + telemetry dulu, UI fallback terbatas, panel pengawas, lalu pilot.

## 4. Opsi Implementasi yang Dipilih

Opsi B secara produk:

- Primary: Flutter Android/Windows.
- Backup: Web/PWA fallback darurat.

Implementasi teknis yang disarankan:

- Web fallback dibuat di SvelteKit `apps/web-admin`, bukan Flutter Web tahap awal.
- Alasan:
  - lebih ringan,
  - cepat dipasang di arsitektur sekarang,
  - BFF `/api/exam/*` sudah tersedia,
  - tidak perlu refactor `dart:io`/`HttpClient` Flutter mobile untuk web,
  - anti-cheat web tetap lemah, jadi tidak perlu memaksa Flutter Web sebelum kebutuhan terbukti.

## 5. Desain Produk / UX

### 5.1 Route publik siswa

Usulan route:

- `/ujian`
- atau `/asesmen/ujian-siswa`
- atau `/cbt-siswa`

Rekomendasi: `/ujian`

Alasan:

- pendek untuk diketik siswa,
- tidak terasa sebagai halaman admin,
- mudah dibuat QR,
- cocok untuk fallback darurat.

### 5.2 Struktur halaman web fallback

Halaman 1 — Login Ujian:

- Header sederhana: logo/nama madrasah, judul `CBT Browser Darurat`.
- Notice kuning: `Gunakan hanya jika diarahkan pengawas. Aplikasi CBT tetap jalur utama.`
- Input:
  - alamat server tidak perlu ditampilkan jika memakai domain yang sama,
  - token ujian,
  - token ruang,
  - tombol `Masuk Ujian`.
- Info perangkat/browser ringkas.
- Bantuan error yang tidak panik: token salah, sesi belum mulai, perangkat tidak sesuai, koneksi bermasalah.

Halaman 2 — Shell Ujian:

- Header sticky:
  - nama siswa,
  - sesi/ruang,
  - sisa waktu,
  - status koneksi,
  - label `Browser Darurat`.
- Area soal:
  - stem/rich text,
  - media gambar/audio jika tersedia,
  - pilihan/essay/matching sesuai payload.
- Navigasi soal:
  - nomor soal,
  - status terjawab/belum/lokal belum terkirim.
- Footer/aksi:
  - `Simpan Jawaban`,
  - `Sebelumnya/Berikutnya`,
  - `Kirim Ujian` dengan konfirmasi kuat.

Halaman 3 — Selesai:

- Ringkasan submit.
- Instruksi kembali ke pengawas.
- Bersihkan snapshot lokal setelah submit sukses.

### 5.3 Mobile behavior

- Layout satu kolom.
- Header ringkas agar tidak memakan layar.
- Navigasi nomor soal bisa berupa drawer/bottom sheet.
- Tombol submit tidak selalu terlihat; taruh di panel konfirmasi agar tidak tersentuh tidak sengaja.

### 5.4 Copy yang harus jelas

Gunakan istilah:

- `Aplikasi CBT` untuk Flutter native.
- `Browser Darurat` untuk fallback web.
- `Pengawasan wajib` untuk kondisi web fallback.
- `Jawaban tersimpan lokal` saat offline/pending.
- `Belum terkirim ke server` jika pending.

Hindari copy seperti:

- `Web sama aman dengan aplikasi`.
- `Anti-cheat browser aktif penuh`.
- `Mode aman sempurna`.

## 6. Perubahan Backend/Core API yang Dibutuhkan

### 6.1 Flag kebijakan fallback

Tambahkan kebijakan di domain kegiatan/sesi/ruang, minimal salah satu:

- level kegiatan/event: `allow_web_fallback`
- level sesi: `allow_web_fallback`
- level ruang: `allow_web_fallback`

Rekomendasi tahap awal:

- mulai dari level sesi atau room/session assignment, karena fallback biasanya kondisi ruang/perangkat.
- Jika lebih cepat, level event boleh dulu, lalu granular room menyusul.

Data tambahan opsional:

- `web_fallback_enabled_at`
- `web_fallback_enabled_by`
- `web_fallback_reason`
- `web_fallback_disabled_at`

### 6.2 Validasi login exam

Saat `/api/exam/login` menerima client web:

- payload tambahkan:
  - `client_type: web_fallback`
  - `browser_fingerprint`
  - `user_agent`
- backend cek:
  - token ujian valid,
  - token ruang valid,
  - sesi berjalan/diizinkan,
  - fallback web aktif untuk sesi/ruang/event tersebut.
- Jika fallback tidak aktif, balas error aman:
  - `Mode browser belum diizinkan. Gunakan aplikasi CBT atau hubungi pengawas.`

### 6.3 Event taxonomy web fallback

Whitelist event, jangan terima event bebas untuk keputusan risk.

Event minimal:

- `web_fallback_used`
- `web_visibility_hidden`
- `web_visibility_visible`
- `web_focus_lost`
- `web_focus_restored`
- `web_fullscreen_exit`
- `web_fullscreen_restored`
- `web_pending_answer_saved`
- `web_pending_answer_flushed`
- `web_connection_degraded`
- `web_connection_restored`

Severity awal:

- visibility/focus/fullscreen: warning, butuh debounce.
- pending/offline: technical, bukan kecurangan.
- device mismatch/token mismatch: high.

### 6.4 Risk policy konservatif

Jangan auto-lock siswa web hanya karena satu focus lost/visibility hidden.

Rekomendasi:

- focus lost < 2 detik: catat ringan / tidak eskalasi.
- repeated focus lost: `Waspada`.
- fullscreen exit berulang: `Butuh Tindakan`.
- network offline: `Gangguan Teknis`, bukan cheating.
- device mismatch: `Butuh Tindakan`.

## 7. Perubahan Web Admin/SvelteKit

### 7.1 Route fallback siswa

Files kemungkinan:

- `apps/web-admin/src/routes/ujian/+page.svelte`
- `apps/web-admin/src/routes/ujian/+layout.svelte` jika perlu layout publik tanpa sidebar admin.
- `apps/web-admin/src/routes/ujian/_components/*` untuk komponen shell.
- `apps/web-admin/src/lib/exam/*` untuk client helper/model jika belum ada.

Pastikan route ini:

- tidak memakai layout admin/sidebar,
- tidak butuh session admin/guru,
- hanya mengakses BFF `/api/exam/*`,
- tidak membaca DB langsung.

### 7.2 BFF exam proxy

Cek dan mungkin sesuaikan:

- `apps/web-admin/src/routes/api/exam/[...path]/+server.ts`

Kebutuhan:

- teruskan header/request body dengan aman,
- jangan log token/jawaban,
- jaga CORS/CSRF sesuai SvelteKit same-origin,
- dukung client type web fallback.

### 7.3 Panel operator/admin untuk mengaktifkan fallback

Files kemungkinan tergantung letak UI asesmen saat ini:

- `apps/web-admin/src/routes/asesmen/kegiatan/[id]/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`

UI minimal:

- Toggle: `Izinkan Browser Darurat`.
- Field alasan wajib saat mengaktifkan.
- Badge status di sesi/ruang.
- Audit note: siapa dan kapan mengaktifkan.

### 7.4 Panel pengawas

Tambahkan indikator di proctoring:

- label client: `Aplikasi Android`, `Aplikasi Windows`, `Browser Darurat`.
- filter: `Peserta Browser Darurat`.
- timeline event web.
- kelompok status:
  - `Butuh Tindakan`,
  - `Waspada`,
  - `Gangguan Teknis`,
  - `Normal`,
  - `Selesai`.

## 8. Local Persistence di Web Fallback

Tujuan: menjaga jawaban saat koneksi putus, bukan mode offline penuh.

Rekomendasi:

- Gunakan IndexedDB jika implementasi cukup; fallback sessionStorage boleh untuk tahap pilot minimal.
- Simpan:
  - exam token hash/identifier, bukan token mentah jika memungkinkan,
  - question answer map,
  - pending queue,
  - last server contact,
  - current question index.
- Bersihkan saat:
  - submit sukses,
  - operator reset akses,
  - siswa logout/keluar dengan konfirmasi,
  - sesi expired.

Catatan keamanan:

- Browser storage tidak seaman secure storage Flutter.
- Jangan simpan answer key atau data sensitif lain.
- Jangan expose payload internal di console log.

## 9. PWA Capability

Tahap awal tidak wajib offline-first penuh.

Yang boleh ditambahkan:

- manifest PWA sederhana,
- icon madrasah,
- install prompt opsional,
- caching static asset saja.

Yang jangan dilakukan tahap awal:

- cache payload soal lengkap secara agresif,
- service worker yang menyimpan response exam API tanpa strategi keamanan jelas,
- full offline exam tanpa desain conflict resolution.

## 10. File yang Kemungkinan Berubah

Backend:

- `services/core-api/db/migrations/*_cbt_web_fallback.sql`
- `services/core-api/db/queries/*.sql` sesuai domain asesmen/exam
- `services/core-api/internal/repository/postgres/*` generated sqlc
- `services/core-api/internal/handler/*exam*` atau handler asesmen terkait
- `services/core-api/internal/service/*exam*` atau service asesmen terkait
- `services/core-api/internal/domain/*` untuk types/policy
- tests backend terkait exam login/event/proctoring

Frontend web-admin:

- `apps/web-admin/src/routes/ujian/+page.svelte`
- `apps/web-admin/src/routes/ujian/+layout.svelte`
- `apps/web-admin/src/routes/ujian/_components/*`
- `apps/web-admin/src/routes/api/exam/[...path]/+server.ts`
- `apps/web-admin/src/routes/asesmen/kegiatan/[id]/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`
- `apps/web-admin/src/lib/*` helper/model exam if needed
- tests Svelte/unit route as appropriate

Docs/runbooks:

- `docs/contracts/asesmen-cbt-formal-sop.md`
- `docs/cbt-operator-runbook.md`
- `docs/exam-api.md` if present/active
- new `docs/cbt-web-fallback-sop.md`

## 11. Tahapan Implementasi Aman

### Phase 0 — Discovery & Contract Audit

Tujuan: memastikan payload `/api/exam/*` cukup untuk web fallback.

Langkah:

1. Audit endpoint exam existing:
   - login,
   - status,
   - heartbeat,
   - event,
   - answer,
   - submit.
2. Catat payload pertanyaan yang wajib didukung:
   - pilihan ganda,
   - essay,
   - matching,
   - rich text,
   - gambar,
   - audio.
3. Audit panel pengawas existing untuk tempat indikator client type.
4. Buat daftar test fixture berdasarkan data simulasi CBT existing.

Deliverable:

- catatan kontrak singkat,
- daftar gap sebelum coding.

### Phase 1 — Backend Policy Flag & Audit

Tujuan: fallback tidak bisa dipakai sembarang.

Langkah:

1. Tambah migration additive untuk flag fallback.
2. Tambah query sqlc dan service policy.
3. Tambah endpoint admin/operator untuk enable/disable fallback dengan reason.
4. Tambah audit actor/time/reason.
5. Update exam login agar menolak `web_fallback` jika flag OFF.

Validasi:

- sqlc generate,
- backend tests,
- login web_fallback OFF harus ditolak,
- login Flutter/native tetap tidak berubah.

### Phase 2 — Web Fallback Login + Shell Minimal

Tujuan: siswa bisa login dan melihat/mengerjakan soal via browser saat fallback aktif.

Langkah:

1. Buat public route `/ujian` tanpa sidebar admin.
2. Buat login form token ujian + token ruang.
3. Kirim `client_type=web_fallback` dan browser fingerprint sederhana.
4. Render shell soal minimal dari payload login.
5. Implement answer state lokal dan save answer ke backend.
6. Implement status koneksi dan pending queue sederhana.
7. Implement submit dengan konfirmasi.

Validasi:

- route `/ujian` dapat dibuka tanpa login admin.
- fallback OFF: login ditolak dengan pesan jelas.
- fallback ON: login berhasil.
- jawab soal tersimpan.
- submit berhasil.
- refresh browser masih bisa recover state secara aman.

### Phase 3 — Browser Telemetry Best-Effort

Tujuan: pengawas mendapatkan sinyal risiko browser tanpa false-positive agresif.

Langkah:

1. Tambah event visibility/focus/fullscreen.
2. Debounce event agar tidak spam.
3. Tambah heartbeat web.
4. Kirim event `web_fallback_used` setelah login sukses.
5. Backend whitelist event dan severity.

Validasi:

- pindah tab tercatat warning.
- kembali fokus tercatat restored.
- event tidak dobel berlebihan.
- network/pending dianggap teknis.

### Phase 4 — Proctor Panel Integration

Tujuan: penggunaan fallback terlihat pada hari-H.

Langkah:

1. Tambah client type di participant card/table.
2. Tambah filter `Browser Darurat`.
3. Tambah timeline event web.
4. Tambah badge status teknis vs risiko.
5. Tambah aksi pengawas jika perlu:
   - catat insiden,
   - verifikasi siswa,
   - reset binding sesuai policy existing.

Validasi:

- peserta web muncul dengan label benar.
- event focus/visibility muncul.
- pengawas dapat membedakan gangguan koneksi vs indikator kecurangan.

### Phase 5 — SOP, Pilot, dan Rollout Terbatas

Tujuan: fallback dipakai aman secara operasional.

Langkah:

1. Tulis SOP `docs/cbt-web-fallback-sop.md`.
2. Tambah checklist pengawas:
   - kapan boleh aktifkan fallback,
   - cara memberi URL/QR,
   - cara memantau peserta web,
   - cara mencatat insiden.
3. Pilot 5–10 perangkat campuran.
4. Uji gangguan koneksi, refresh browser, pindah tab, submit.
5. Baru aktifkan untuk simulasi kelas kecil.

Validasi:

- operator bisa menjelaskan kapan fallback dipakai.
- pengawas bisa melihat peserta fallback.
- tidak ada data jawaban hilang dalam skenario refresh/koneksi putus ringan.

## 12. Test & Verification Plan

### Backend

Jalankan setelah implementasi backend:

```bash
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./...
go build -o /tmp/core-api-cbt-web-fallback-check ./cmd/api
```

Test yang harus ada:

- login web fallback ditolak saat flag OFF.
- login web fallback diterima saat flag ON.
- login Flutter/native tidak terpengaruh.
- event type web fallback hanya whitelist.
- reason/actor audit tersimpan saat toggle fallback.
- risk policy tidak auto-lock karena satu focus lost singkat.

### Frontend

Jalankan setelah implementasi web-admin:

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
git diff --check
```

Test yang harus ada:

- route `/ujian` render login tanpa admin session.
- form token validation.
- fallback OFF menampilkan pesan aman.
- fallback ON render exam shell.
- save answer memanggil API benar.
- pending queue muncul saat mock network gagal.
- submit confirmation tidak mudah tersentuh tidak sengaja.
- visibility/focus event debounce.

### Browser smoke manual

Skenario minimal:

1. Buka `/ujian` dari HP/laptop.
2. Login dengan token aktif saat fallback OFF: harus ditolak.
3. Aktifkan fallback via operator.
4. Login ulang: berhasil.
5. Jawab 2 soal.
6. Matikan koneksi sebentar: status jadi lokal/waspada, jawaban pending tidak hilang.
7. Nyalakan koneksi: pending terkirim.
8. Pindah tab: panel pengawas mencatat warning.
9. Submit: selesai dan snapshot lokal dibersihkan.
10. Panel pengawas menunjukkan peserta `Browser Darurat`.

## 13. Deployment Plan Aman

Tidak deploy otomatis tanpa approval eksplisit.

Jika nanti diimplementasikan, urutan deploy aman:

1. Backup DB sebelum migration.
2. Apply migration additive.
3. Build core-api.
4. Restart PM2 `mtsn2kolut-core-api`.
5. Health check core API.
6. Build web-admin.
7. Restart PM2 `mtsn2kolut-web-admin` segera setelah build untuk hindari stale chunks.
8. Smoke local protected/public routes:
   - `/ujian`,
   - `/asesmen`,
   - proctoring room page,
   - `/api/exam/*` expected behavior.
9. Pilot hanya pada sesi simulasi.

Rollback:

- UI rollback: revert web route changes, rebuild/restart web-admin.
- Backend flag default OFF membuat fitur inert.
- Migration additive sebaiknya tidak di-drop tanpa SOP backup/approval.

## 14. Risiko dan Mitigasi

### Risiko: Web fallback disalahartikan sebagai mode aman penuh

Mitigasi:

- copy UI jelas: `Browser Darurat`, `pengawasan wajib`.
- SOP menyebut anti-cheat web terbatas.

### Risiko: Siswa pindah tab lalu false positive

Mitigasi:

- debounce focus/visibility.
- tidak auto-lock dari satu event ambigu.
- pengawas verifikasi manual.

### Risiko: Jawaban pending hilang karena refresh/browser storage

Mitigasi:

- snapshot lokal terstruktur.
- indikator pending jelas.
- test refresh/network.

### Risiko: Token/jawaban bocor di log

Mitigasi:

- audit BFF/backend logging.
- redaksi field sensitif.
- jangan console.log payload exam.

### Risiko: Route publik membuka surface baru

Mitigasi:

- tetap token + room token.
- rate limit login exam jika belum ada.
- fallback flag default OFF.
- no DB direct from SvelteKit.

### Risiko: Service worker/PWA cache menyimpan data ujian

Mitigasi:

- tahap awal jangan cache API exam.
- cache static asset saja.
- no offline-first full exam sebelum ada desain keamanan.

## 15. Open Questions Sebelum Implementasi

1. Route final yang dipilih: `/ujian`, `/cbt-siswa`, atau lainnya?
2. Fallback flag mau level event, sesi, atau ruang?
3. Apakah fallback boleh dipakai untuk semua jenis soal termasuk audio/matching pada phase awal?
4. Apakah perlu QR khusus per sesi/ruang untuk membuka `/ujian`?
5. Apakah device binding web memakai fingerprint browser saja atau digabung dengan user agent/IP?
6. Apakah operator boleh mengaktifkan fallback sendiri, atau harus admin?
7. Berapa lama snapshot lokal browser boleh bertahan setelah sesi selesai?

## 16. Rekomendasi Keputusan Awal

Untuk implementasi pertama yang paling aman:

- Route: `/ujian`.
- Flag: level sesi/ruang jika struktur backend mendukung; jika tidak, level event dulu dengan default OFF.
- Client type: `web_fallback`.
- Storage: IndexedDB/session-scoped; jika terlalu berat, sessionStorage pilot dengan warning dan cleanup.
- PWA: manifest + static caching saja; jangan cache exam API.
- Risk: warning only untuk focus/visibility, high untuk device/token mismatch.
- Rollout: simulasi kecil dulu, bukan ujian resmi besar.

## 17. Acceptance Criteria

Fitur dianggap siap pilot jika:

- Flutter/native login tetap berjalan seperti sebelumnya.
- Web fallback tidak bisa login saat flag OFF.
- Admin/operator bisa mengaktifkan fallback dengan reason dan audit.
- Siswa bisa login via `/ujian` saat flag ON.
- Jawaban bisa tersimpan, pending saat offline ringan, dan terkirim saat koneksi pulih.
- Submit berhasil dan membersihkan snapshot lokal.
- Panel pengawas menampilkan peserta `Browser Darurat` dan event penting.
- Test backend/frontend lulus.
- SOP fallback tersedia dan dipahami pengawas.

## 18. Catatan Eksekusi

Plan ini belum melakukan perubahan kode. Setelah user menyetujui, implementasi sebaiknya dilakukan bertahap per phase dengan review evidence di tiap fase sebelum deploy produksi.
