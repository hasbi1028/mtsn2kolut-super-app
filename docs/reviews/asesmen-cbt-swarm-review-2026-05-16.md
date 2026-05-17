# Review Swarm Modul Asesmen CBT — Naturalitas, Formalitas, dan Rekomendasi

Tanggal: 2026-05-16
Repo: `/home/servermtsn2kolut/mtsn2kolut-super-app`
Metode: 5 agent swarm read-only

## Ringkasan Eksekutif

Modul Asesmen CBT sudah cukup matang secara teknis dan operasional. Alur hari pelaksanaan kuat: sesi, peserta, token, ruang, pengawasan, autosave, anti-cheat, force submit, handover, hasil, dan analisis sudah tersedia. Namun agar terasa lebih natural dan formal untuk lingkungan madrasah/Kemenag, perlu pembenahan pada nomenklatur UI, workflow SOP terpadu, approval/finalisasi, berita acara/arsip, dan mode sederhana untuk guru/proktor/pengawas.

Skor konsolidasi: **82/100**

- Copy/wording natural-formal: 70/100
- Workflow operasional: 82/100
- Technical/API/security/readiness: 82/100
- Static UX QA web-admin: 78/100
- Mobile CBT siswa: 82/100

## Pendapat Utama

Modul ini bukan sekadar CRUD asesmen; secara teknis sudah mendekati sistem ujian digital lengkap. Yang belum sepenuhnya selesai adalah lapisan “bahasa institusi” dan “SOP formal”. Fitur banyak dan kuat, tetapi beberapa istilah masih terasa campuran antara bahasa developer, bahasa CBT, dan bahasa operator. Untuk pemakaian madrasah, modul perlu lebih memandu pengguna seperti alur panitia resmi: rencana kegiatan → pengisian/telaah soal → paket disahkan → peserta/ruang/token siap → pelaksanaan → koreksi → verifikasi hasil → finalisasi/arsip.

## Temuan Konsolidasi Prioritas

### P1 — Standarisasi istilah UI

Masalah:
- “Asesmen”, “Ujian”, “CBT” dipakai bergantian.
- “Event” masih muncul di UI; lebih formal sebagai “Kegiatan”.
- “Kode ujian” dan “token ujian/ruang” belum seragam.
- “Review” dan “verifikasi” di Bank Soal bercampur.
- Istilah teknis seperti endpoint, payload, pool, draw, seed, stale, fingerprint, restore, builder masih muncul di beberapa area.

Rekomendasi glossary:
- Modul: **Asesmen CBT**
- Event: **Kegiatan Asesmen**
- Package: **Paket Soal**
- Session: **Sesi Ujian**
- Room: **Ruang Ujian**
- Proctoring: **Pengawasan Ruang**
- Token/code: **Token Ujian** dan **Token Ruang**
- Results: **Hasil Asesmen**
- Review soal: **Verifikasi Soal**
- Published: **Terbit**
- Draft: pilih konsisten antara **Draft** atau **Konsep**
- Pool: **Kumpulan Soal**
- Draw: **Ambil Acak**
- Restore: **Pulihkan Sesi**
- Fingerprint: **Penanda Perangkat**

### P1 — Buat “Pusat Kegiatan Asesmen” sebagai alur SOP terpadu

Saat ini fitur sudah ada, tetapi tersebar di paket, kegiatan, sesi, kartu, pengawasan, hasil, dan bank soal. Kegiatan belum terasa sebagai induk workflow.

Rekomendasi status formal kegiatan:
1. Draft
2. Pengisian Soal
3. Telaah/Verifikasi Soal
4. Paket Siap
5. Peserta & Ruang Siap
6. Token & Kartu Siap
7. Pelaksanaan
8. Koreksi
9. Verifikasi Hasil
10. Final & Arsip

Setiap tahap perlu checklist, owner/peran, status blocking/warning, dan next action.

### P1 — Finalisasi dan arsip belum cukup formal

Fondasi sudah ada: operational recap, minutes, audit logs, handover lock, finalize overdue. Namun belum menjadi closing workflow resmi.

Rekomendasi closing checklist:
- Semua peserta selesai/ditandai tidak hadir.
- Semua esai/manual grading selesai.
- Semua insiden ruang ditutup.
- Semua force submit/reset/unlock tercatat.
- Hasil dihitung ulang dan diverifikasi guru/panitia.
- Handover ruang dikunci.
- Sesi/kegiatan dikunci final.
- Dokumen arsip diunduh/terbit.

### P1 — Contract BFF `/api/asesmen` perlu disinkronkan dengan backend

Agent teknis menemukan mismatch antara route backend dan BFF, antara lain backend punya beberapa route yang tidak terlihat di BFF, misalnya:
- `GET /api/asesmen/sessions/{id}`
- `DELETE /api/asesmen/sessions/{id}`
- `GET /api/asesmen/sessions/{id}/participants/{pid}/answers`
- beberapa endpoint enroll/answer legacy

Sebaliknya ada BFF stream/custom route yang tidak mirror langsung sebagai backend native.

Rekomendasi:
- Buat route contract map: backend route ↔ BFF alias ↔ UI usage.
- Tambahkan contract test untuk `/api/asesmen/*`, `/api/bank-soal/*`, dan legacy `/api/cbt/*` yang masih dijanjikan.
- Putuskan secara eksplisit: compatibility `/api/cbt/*` berlaku di backend saja atau juga BFF.

### P2 — Audit domain untuk aksi sensitif

Audit generic sudah ada, dan beberapa proctoring action sudah domain-specific. Namun aksi high-impact perlu audit metadata yang lebih kaya:
- Generate/regenerate token
- Enroll peserta massal
- Score/finalize session
- Package lock/clone
- Event status update
- Force submit/reset/unlock

Metadata jangan memuat token raw.

### P2 — Mode sederhana untuk pengawas/proktor

Panel pengawasan kuat, tetapi bisa terasa teknis. Perlu mode yang berbasis kasus:
- Belum masuk
- Sedang mengerjakan
- Perlu bantuan
- Jaringan bermasalah
- Terkunci anti-cheat
- Sudah selesai

Setiap status diberi tombol SOP: sinkron ulang, reset akses, unlock, force submit, catat insiden, hubungi siswa.

### P2 — Mobile CBT: pesan siswa perlu lebih sederhana

Mobile CBT kuat, tetapi istilah seperti heartbeat, anti-switch, degraded mode, fingerprint, sinkron, restore, submit masih dapat disederhanakan.

Rekomendasi cepat:
- Heartbeat aktif → Koneksi dipantau
- Anti-switch dasar → Keluar aplikasi tercatat
- Submit → Kirim jawaban
- Restore → Pulihkan sesi
- Sinkron → Terkirim ke server / belum terkirim
- Tambahkan mapping khusus HTTP 423 Locked: “Ujian dikunci sementara oleh sistem pengawasan. Tetap di tempat dan minta pengawas memeriksa akses Anda.”
- Jangan tampilkan pesan backend mentah kepada siswa.

### P2 — UX QA web-admin

Temuan penting:
- Beberapa launcher memakai role hardcoded, tidak sepenuhnya selaras dengan permission/RBAC.
- `/asesmen/non-tes` ada tetapi kurang discoverable dari hub/sidebar.
- Link `#soal` / `#blueprint` dari daftar paket tidak benar-benar membuka tab target karena detail page tidak membaca hash.
- Paket Builder detail dapat blank saat loading karena `AsyncContent` tidak punya pending snippet khusus.
- Aksi besar seperti lock/clone masih memakai native `window.confirm/prompt`, kurang formal dan kurang konsisten.

### P3 — Formalisasi hasil dan nilai

Perlu workflow nilai:
1. Koreksi otomatis
2. Koreksi esai/manual
3. Review guru mapel
4. Validasi koordinator/panitia
5. Kunci nilai
6. Sync ke nilai/rapor
7. Remedial/tindak lanjut

Tambahkan status “nilai sementara” vs “nilai final” dan audit perubahan nilai manual.

## Rekomendasi Sprint Lanjutan

### Sprint A — Bahasa & Formalisasi UI Asesmen CBT
- Terapkan glossary resmi.
- Ganti “event” menjadi “kegiatan” di UI.
- Samakan kode/token, review/verifikasi, pengawasan/proctoring.
- Hapus istilah developer dari UI operator/guru/siswa.

### Sprint B — Pusat Kegiatan Asesmen + SOP Wizard
- Dashboard kegiatan sebagai pusat kerja panitia.
- Tahapan formal dari draft sampai arsip.
- Checklist blocking/warning dan next action.
- Owner per tahap: admin/operator/proktor/guru/pengawas.

### Sprint C — Approval, Lock, dan Finalisasi Formal
- Pengesahan paket soal.
- Approval peserta/ruang/token.
- Verifikasi hasil.
- Lock nilai dan lock kegiatan.
- Audit domain high-impact.

### Sprint D — Berita Acara & Arsip Digital
- BA sesi/ruang/kegiatan lebih lengkap.
- Rekap hadir/tidak hadir/susulan.
- Insiden dan tindakan proktor.
- Lampiran hasil, audit ringkas, tanda tangan panitia.
- Export ZIP arsip kegiatan.

### Sprint E — Mode Proktor/Pengawas Sederhana
- Dashboard berbasis status peserta.
- SOP tindakan per kasus.
- Bahasa minim teknis.

### Sprint F — Mobile CBT Student Copy Polish
- Mapping error 423/locked.
- Fallback error Indonesia non-teknis.
- Rename istilah teknis.
- Dialog submit lebih jelas jika ada soal belum dijawab.

### Sprint G — Route/API Contract Hardening
- Contract map backend/BFF/UI.
- Test alias `/api/asesmen`, `/api/bank-soal`, legacy `/api/cbt`.
- Standarkan backend JSON body limit.
- Kurangi route drift.

## Kesimpulan

Modul Asesmen CBT sudah layak disebut kuat dari sisi mesin ujian dan hari pelaksanaan. Untuk naik kelas menjadi sistem asesmen madrasah yang natural dan formal, fokus berikutnya bukan menambah fitur acak, tetapi menyatukan pengalaman menjadi SOP resmi: bahasa konsisten, kegiatan sebagai pusat alur, approval jelas, finalisasi/arsip kuat, dan tampilan sederhana untuk pengawas/siswa.
