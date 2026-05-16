# Swarm Review UI/UX Asesmen CBT

Tanggal: 2026-05-17
Commit direview: `92775fd feat(asesmen): formalize CBT SOP workflow`
Scope: UI/UX Asesmen CBT web-admin dan mobile CBT, bukan review security/code correctness.

## Verdict Ringkas

Arah UI/UX sudah **cukup sesuai CBT sekolah/madrasah secara umum**: sudah ada alur Kegiatan Asesmen, Bank Soal, Paket Soal, Sesi, Pengawasan, Hasil, Berita Acara, readiness, dan SOP formal. Namun untuk disebut matang seperti aplikasi CBT operasional, masih perlu penyederhanaan pada layar live dan penguatan finalisasi/arsip.

Prioritas terbesar bukan menambah fitur besar, melainkan membuat UI lebih operasional:

1. Mode pengawas sederhana saat ujian berlangsung.
2. Halaman/section arsip kegiatan yang benar-benar ada.
3. Panel pengesahan SOP formal yang terlihat di UI.
4. Konsistensi istilah: Kegiatan, Sesi, Token Ruang, Token Ujian, Verifikasi, Terbit.
5. Kurangi istilah teknis/Inggris di UI operator/guru/siswa.
6. Pisahkan token distribusi dari dokumen arsip final.

## Hasil Swarm per Spesialis

### 1. Operator/Admin CBT Madrasah

**Verdict:** Cukup selaras sebagai kerangka SOP CBT madrasah, tetapi belum utuh sebagai alur formal sampai pengesahan/finalisasi dan arsip.

Yang sudah cocok:
- Terminologi umum seperti Kegiatan Asesmen, Paket Soal, Sesi Ujian, Ruang/Pengawas/Kursi, Token/Kartu, Pengawasan Ruang, Berita Acara sudah cocok.
- SOP 10 tahap masuk akal: Draft, Pengisian Soal, Verifikasi Soal, Paket Siap, Peserta & Ruang Siap, Token & Kartu Siap, Pelaksanaan, Koreksi, Verifikasi Hasil, Final & Arsip.
- Detail Kegiatan berfungsi sebagai command center.
- Monitoring kelengkapan soal per guru/mapel/rombel sangat relevan.
- Readiness kartu/token dan pengawasan ruang sudah praktis.

Gaps utama:
- Finalisasi/pengesahan belum terasa sebagai aksi nyata di UI.
- Link `/asesmen/kegiatan/{id}/archive` tampak belum punya halaman.
- BA masih tersebar di level sesi, belum menjadi bundle arsip kegiatan.
- Token raw muncul di BA/daftar hadir sesi; perlu kontrol/masking untuk arsip final.
- Readiness logic berbasis count bisa membingungkan bila tidak membandingkan total vs lengkap.

### 2. Proktor/Pengawas Ruang

**Verdict:** Fitur live CBT cukup lengkap, tetapi belum sepenuhnya “siap tekanan ruang” untuk pengawas non-teknis.

Yang sudah kuat:
- `/asesmen/pelaksanaan` sudah menjadi launcher hari-H.
- `/asesmen/pengawasan` cocok sebagai daftar ruang.
- Panel ruang punya live stream/fallback polling, audio alert, statistik, tabel peserta, risiko, tindakan pengawas, riwayat, bukti, serah terima.
- Incident handling sudah ada: tandai, reset, peringatkan, instruksi masuk ulang, paksa kirim, eskalasi, selesai.
- Lock/unlock peserta didukung dan mobile copy 423 sudah mengarahkan siswa ke pengawas.

Gaps utama:
- Belum ada Simple Proctor Mode yang benar-benar sederhana.
- Tindakan lock/unlock belum cukup menonjol di panel ruang peserta terkunci.
- Masih memakai prompt browser untuk tindakan penting; sebaiknya dialog terstruktur.
- Istilah “Ruang dikunci”, “Peserta terkunci”, “Serah terima terkunci” berpotensi ambigu.
- Tabel live terlalu padat; pengawas butuh daftar “Butuh Tindakan” default.

### 3. Siswa / Mobile CBT

**Verdict:** Copy mobile sudah baik untuk mencegah panik, tetapi perlu penyelarasan istilah dan pengurangan istilah teknis.

Yang sudah baik:
- Pesan sering menegaskan jawaban tetap aman.
- Siswa diarahkan tetap di layar dan panggil pengawas.
- `423 locked` diposisikan sebagai akses dikunci/ditahan pengawas/sistem.
- Submit/restore/status/answer flow sudah cukup aman.

Gaps utama:
- Istilah status mobile (`Aman`, `Perlu sinkron`, `Perlu pengawas`) belum selaras dengan panduan web (`Tersambung`, `Lokal`, `Waspada`, `Gangguan`, `Menurun`).
- Tone `danger` untuk 423 bisa terasa menghukum untuk siswa MTs.
- Login screen masih punya istilah teknis: heartbeat, anti-switch, fingerprint, API, HTTP/HTTPS, heksadesimal.
- Restore failure memiliki tombol “Kembali ke Login Token” yang bisa mendorong siswa login ulang tanpa pengawas pada kondisi 409/423.
- Pesan media error menampilkan URL mentah.

### 4. Bank Soal / Guru Penyusun / Reviewer

**Verdict:** Struktur draft–verifikasi–disetujui–terbit sudah benar, tetapi istilah dan makna beberapa halaman belum sepenuhnya formal CBT.

Yang sudah baik:
- Alur Draft → Verifikasi → Revisi → Disetujui → Terbit cukup terbaca.
- Bank Soal cukup terpisah dari sesi ujian.
- Halaman verifikasi menyediakan naskah, opsi, kunci/rubrik, pembahasan, catatan keputusan, timeline.
- Mapel/KD membantu melihat kelengkapan metadata.

Gaps utama:
- Label “Analisis Butir” belum sesuai makna CBT umum bila belum ada statistik pasca-ujian seperti kesukaran, daya pembeda, efektivitas pengecoh.
- Masih ada istilah UI teknis/Inggris: Advanced Bank Soal, Mapel & KD Coverage, Export, Penugasan Event, Published, Sprint, additive, hard enforcement.
- Kartu “Telah Diverifikasi” menghitung approved + published, tetapi kliknya berpotensi hanya memfilter approved draft.
- Tombol Setujui masih bisa terlihat walau checklist kunci/rubrik belum lengkap; perlu warning/disable sesuai tipe soal.

### 5. Accessibility / Information Architecture

**Verdict:** Beranda dan fase kerja sudah baik, tetapi detail kegiatan dan navigasi Asesmen terlalu padat/bercabang.

Yang sudah baik:
- Beranda Asesmen task-oriented: Siapkan, Jalankan, Evaluasi.
- Loading/error state memakai AsyncContent, Skeleton, RecoveryPanel, LoadingButton.
- Halaman sesi sudah punya card mobile selain tabel desktop.

Gaps utama:
- SOP timeline 10 tahap tampil penuh dan bisa overload.
- Sidebar mencampur fase kerja dan objek data: Beranda, Paket, Kegiatan, Persiapan, Aplikasi Siswa, Pelaksanaan, Hasil.
- Tab/filter belum semuanya memakai ARIA tablist/tab/tabpanel atau aria-label yang spesifik.
- Detail Kegiatan mobile berisiko scroll panjang; timeline dan tabel horizontal perlu ringkas/card mode.
- Empty states belum selalu memberi CTA kontekstual.

## Prioritas Perbaikan UI/UX

### P0 — Wajib sebelum production dianggap nyaman CBT

1. **Simple Proctor Mode**
   - Tab/toggle: Mode Sederhana vs Mode Lengkap.
   - Default saat sesi aktif: tampilkan Token Ruang besar, ringkasan peserta, daftar Butuh Tindakan, tombol tindakan aman.

2. **Archive Kegiatan benar-benar tersedia**
   - Jika link `/asesmen/kegiatan/{id}/archive` muncul, buat halaman/sectionnya.
   - Minimal: checklist dokumen, kartu ujian, daftar hadir, BA sesi/ruang, rekap hasil, rekap insiden, audit/pengesahan.

3. **Panel Pengesahan SOP di Detail Kegiatan**
   - Sahkan Paket Soal.
   - Sahkan Peserta & Ruang.
   - Sahkan Token & Kartu.
   - Sahkan Hasil.
   - Finalkan & Arsipkan.
   - Tampilkan aktor, waktu, catatan, status dicabut bila ada.

4. **Pisahkan token distribusi dan arsip final**
   - Kartu ujian boleh token raw.
   - BA/arsip final sebaiknya token masked atau opsional dengan warning.

### P1 — Penting untuk CBT umum

5. **Konsistensi istilah**
   - Kode Ruang → Token Ruang.
   - Kirim → Sudah kirim / Jawaban terkirim.
   - Pindah → Keluar aplikasi.
   - Tangkapan → Coba tangkap layar.
   - Event → Kegiatan Asesmen.
   - Published → Terbit.
   - Export → Ekspor.

6. **Timeline SOP ringkas**
   - Default hanya tampilkan status utama dan Langkah Berikutnya.
   - Detail 10 tahap dalam accordion “Lihat SOP formal”.

7. **Dialog tindakan pengawas terstruktur**
   - Ganti `window.prompt` untuk unlock/reset/incident dengan modal alasan cepat + catatan.

8. **Readiness berbasis total-vs-lengkap**
   - Token/kartu: peserta total vs token/ruang/kursi/kartu lengkap.
   - Verifikasi soal: target terpenuhi, ada draft, ada antrean review, ada kekurangan.
   - Ruang: pengawas/kursi/peserta/room token lengkap.

### P2 — Polish lanjutan

9. **IA sidebar disederhanakan**
   - Beranda Asesmen
   - Persiapan
   - Hari-H / Pengawasan
   - Hasil
   - Aplikasi Siswa
   - Paket/Kegiatan/Sesi menjadi sub-aksi Persiapan.

10. **Bank Soal: revisi label Analisis Butir**
   - Bila belum ada psikometrik, ganti ke “Pemantauan Mutu Soal” atau “Kesiapan Butir Soal”.

11. **Mobile CBT: kurangi istilah teknis**
   - Heartbeat → Status/koneksi.
   - Anti-switch → Tetap di aplikasi.
   - Fingerprint → Penanda perangkat.
   - Heksadesimal → Token dari kartu ujian.

12. **A11y tab/filter dan empty state**
   - Tab detail pakai tablist/tab/tabpanel.
   - Kartu filter punya aria-label.
   - Empty state punya CTA: Buat Kegiatan, Buat Sesi, Reset Filter, Kelola Paket.

## Kesimpulan Akhir

UI/UX yang baru sudah bergerak ke arah CBT formal yang benar dan cukup familiar bagi madrasah. Namun saat ini masih lebih kuat sebagai **dashboard admin lengkap** daripada **alat operasional CBT yang cepat dipakai saat ujian berlangsung**.

Jika Bapak ingin sprint berikutnya, rekomendasi urutannya:

1. Sprint UX CBT 1: Simple Proctor Mode + terminology cleanup.
2. Sprint UX CBT 2: Archive Kegiatan + Pengesahan SOP visible di UI.
3. Sprint UX CBT 3: Bank Soal terminology + readiness logic refinement.
4. Sprint UX CBT 4: Mobile CBT copy alignment + a11y/responsive polish.
