# Review Swarm Asesmen/CBT — 2026-05-20 06:11 WITA

## Ringkasan Eksekutif

Review read-only menggunakan 5 perspektif agent swarm:

1. UX wording/copy formal madrasah
2. Workflow operasional panitia/guru/proktor/siswa
3. Technical/API/security architecture
4. Static UI QA web-admin
5. Mobile/student/proctoring/anti-cheat readiness

Tidak ada source code yang diubah selama review. Verifikasi ringan:

- `npm run check` di `apps/web-admin`: PASS, 0 errors, 0 warnings
- `go test ./internal/handler ./internal/service -run "Cbt|Exam|Proctor" -count=1 -timeout 120s`: PASS
- `git status --short`: bersih sebelum report ini dibuat

## Skor Konsolidasi

- Overall: 80/100
- UX wording/copy formal: 78/100
- Workflow operasional: 81/100
- Technical/API/security: 80/100
- Static UI QA: 74/100
- Mobile/proctoring/anti-cheat: 75/100

Verdict: Asesmen/CBT sudah kuat dan operasional untuk penggunaan madrasah, tetapi belum layak disebut “sangat sempurna”. Kekuatan utama ada di alur data, readiness, proctoring, token, laporan, dan pemisahan Bank Soal vs Asesmen. Gap utama ada di formal SOP enforcement, dialog aksi high-impact, copy teknis yang masih muncul, dan hardening proctoring/BYOD.

## Kekuatan Utama

- Alur besar sudah mendekati workflow formal madrasah: persiapan, paket, kegiatan, sesi, pelaksanaan, hasil, arsip.
- Kegiatan Asesmen sudah menjadi hub operasional yang baik.
- Backend menghitung readiness berbasis data nyata: paket, soal terbit, peserta, ruang, token, submit, scoring.
- Ada approval/pengesahan formal: package_ready, participants_rooms_ready, tokens_cards_ready, results_verified, final_archive, session_minutes, room_handover.
- Token peserta sudah dimasking; token ujian dibuka dengan token ruang.
- Proctoring sudah memiliki live summary, event timeline, acknowledge/action, reset/unlock/force-submit, report data.
- Mobile CBT sudah punya deteksi BYOD dasar: app background/focus, split screen, PiP, FLAG_SECURE, local lock.
- SvelteKit tetap BFF/proxy; Go core-api tetap pemilik PostgreSQL.
- Test ringan CBT/Exam/Proctor dan svelte-check lulus.

## P1 — Harus Diprioritaskan

### 1. Jadikan SOP/approval sebagai gate, bukan hanya catatan

Saat ini SOP timeline dan approval sebagian masih panduan/catatan formal. Rekomendasi:

- Terapkan state machine kegiatan asesmen:
  - draft
  - question_authoring
  - question_verification
  - package_ready
  - participants_rooms_ready
  - tokens_cards_ready
  - execution
  - grading
  - result_verification
  - final_archive
- Setiap transisi punya precondition backend dan audit log.
- `final_archive` harus memblokir perubahan paket, peserta, token, skor, dan BA kecuali ada reopen/cabut pengesahan dengan alasan.

### 2. Ganti `window.confirm/prompt` untuk aksi high-impact

Ditemukan:

- `paket/[id]/+page.svelte`: clone paket pakai `window.prompt`
- `paket/[id]/+page.svelte`: kunci paket pakai `window.confirm`
- `sesi/[id]/proctoring/+page.svelte`: buka kunci dan acknowledge pakai `window.prompt`
- `rooms/[rid]/proctoring/+page.svelte`: browser darurat pakai `window.prompt`

Rekomendasi:

- Pakai dialog shadcn/formal confirmation.
- Wajib alasan/catatan untuk unlock, reset, force-submit, browser darurat, lock paket.
- Tampilkan dampak tindakan sebelum tombol utama.

### 3. Hardening answer-key redaction

Temuan security reviewer:

- List/detail Bank Soal masih terlalu bergantung pada SQL/service redaction.
- Handler detail memanggil serializer dengan `includeAnswerKey=true` dan mengandalkan service sudah mengosongkan field.

Rekomendasi:

- Defense-in-depth di handler/serializer.
- List endpoint default redacted.
- Tes eksplisit non-admin/non-author tidak menerima answer_key.
- Jika reviewer/approver memang boleh lihat kunci, dokumentasikan sebagai exception formal.

### 4. Perbaiki BFF `/api/exam` header/body limit

Temuan:

- BFF `/api/exam/[...path]` meneruskan `x-forwarded-for` / `x-real-ip` dari client.
- Body limit memakai `text.length`, bukan byte count, dan membaca body penuh dulu.

Rekomendasi:

- Jangan forward IP header mentah dari client.
- Tetapkan canonical client IP dari trusted source.
- Gunakan streaming byte limit / `Uint8Array.byteLength`.

### 5. Verifikasi kontrak event mobile ↔ backend

Temuan:

- Mobile punya event `anti_cheat_violation`, `app_switch`, `screenshot_attempt`.
- Backend policy memakai normalized event seperti `app_switch_once`, `background_over_threshold`, `split_screen_detected`, `screenshot_attempt_valid/ambiguous`.

Rekomendasi:

- Tambahkan contract test bahwa semua event Flutter diterima dan dipetakan dengan severity/risk yang benar.
- Audio dashboard sebaiknya menggunakan `severity/category/audio_key` backend, bukan hardcoded event name.

## P2 — Penting Untuk Iterasi Berikutnya

- Formalisasi istilah/copy:
  - `Quality Gate Paket` → `Pemeriksaan Kesiapan Paket`
  - `Mix policy` → `Pengaturan Pembagian Peserta`
  - `backend` → hilangkan dari UI operator
  - `Flutter` → `Aplikasi siswa`
  - `Payload ujian terlalu besar` → `Data jawaban terlalu besar untuk dikirim...`
  - `Hasil & BA` → `Hasil & Berita Acara`
- Tingkatkan `/asesmen/hasil` menjadi pusat pasca-ujian: sesi belum dikoreksi, esai belum dinilai, peserta menggantung, hasil belum verified.
- Tambahkan final archive precondition: semua sesi selesai, semua peserta punya status akhir, esai selesai dikoreksi, incident acknowledged, BA lengkap.
- Tambahkan checklist arsip formal: SK/panitia, jadwal resmi, peserta, ruang/pengawas, kartu/token, BA sesi/ruang, rekap nilai, insiden, pengesahan final.
- Tambahkan pending/failed/retry state eksplisit di halaman proctoring/report/paket detail yang masih belum konsisten.
- Tambahkan rate limit event `/api/exam/event` selain dedup policy.
- Force submit harus meminta alasan/catatan eksplisit di UI, terutama saat pending sync.

## P3 — Polish

- Konsistensi CTA: `Kelola` → `Kelola Kegiatan`, `Monitor` → `Pantau Ujian`.
- Breadcrumb lebih konsisten di halaman detail/proctoring/print-pack.
- Audit token/hash/commit display agar selalu `break-all`, `truncate`, atau `min-w-0`.
- Mode cetak paket arsip lengkap: cover, ringkasan, jadwal, ruang, hadir, BA, insiden, hasil, pengesahan.
- Pisahkan copy teknis dan non-teknis: technical/offline/pending sync harus jelas sebagai kendala teknis, bukan otomatis kecurangan.

## Rekomendasi Sprint

1. Sprint 1 — Formalisasi bahasa & hapus istilah teknis UI.
2. Sprint 2 — Dialog aksi high-impact untuk paket/proctor/browser darurat/force-submit.
3. Sprint 3 — SOP state machine + approval gates backend.
4. Sprint 4 — Finalisasi hasil + arsip digital formal.
5. Sprint 5 — Security hardening `/api/exam`, answer-key defense-in-depth, body limits.
6. Sprint 6 — Event contract mobile-backend + audio alert berbasis policy backend.
7. Sprint 7 — Mode proktor sederhana hari-H + laporan/berita acara final.

## Kesimpulan

Asesmen/CBT sudah **baik dan cukup matang untuk operasional internal**, dengan skor sekitar **80/100**. Untuk menjadi “sangat baik/sangat siap audit”, fokus berikutnya bukan menambah banyak fitur baru, tetapi mengunci workflow formal, menyederhanakan bahasa operator/siswa, mengganti native prompt/confirm, dan memperkuat kontrak keamanan proctoring/token/event.
