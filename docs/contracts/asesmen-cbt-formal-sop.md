# Kontrak Formal SOP Asesmen CBT

Tanggal baseline: 2026-05-17

Dokumen ini menjadi acuan istilah dan alur formal Asesmen CBT di MTs Negeri 2 Kolaka Utara. Tujuannya menjaga UI operator/guru/pengawas/siswa tetap natural, formal, dan tidak membawa istilah teknis yang tidak perlu.

## Glosarium Resmi

| Konsep | Istilah UI resmi | Catatan |
| --- | --- | --- |
| Modul | Asesmen CBT | Nama modul asesmen berbasis komputer. |
| Event | Kegiatan Asesmen | Dipakai untuk induk workflow panitia. |
| Package | Paket Soal | Bank Soal tetap domain sumber soal. |
| Session | Sesi Ujian | Jadwal dan ruang pelaksanaan ujian. |
| Room | Ruang Ujian | Ruang fisik/panel pengawas. |
| Proctoring | Pengawasan Ruang | Pantauan pengawas/proktor. |
| Code/token | Token Ujian, Token Ruang | "Kode" hanya boleh muncul sebagai transisi bila perlu. |
| Review soal | Verifikasi Soal | Dipakai untuk telaah/approval soal. |
| Published | Terbit | Status soal siap digunakan. |
| Pool | Kumpulan Soal | Hindari istilah "pool" di UI operator. |
| Draw | Ambil Acak | Untuk pengambilan soal acak. |
| Restore | Pulihkan Sesi | Untuk memulihkan sesi siswa. |
| Fingerprint | Penanda Perangkat | Petunjuk teknis BYOD, bukan bukti identitas kuat. |

## Istilah Yang Dihindari Di UI Operator

Istilah berikut tidak dipakai sebagai teks utama UI operator/guru/siswa, kecuali di dokumentasi teknis, route, kode, nama field API, atau test yang memang mengunci kontrak:

- `event`
- `endpoint`
- `payload`
- `pool`
- `draw`
- `seed`
- `stale`
- `fingerprint`
- `restore`
- `builder`
- `proctoring`

Padanan UI:

- `endpoint` -> layanan sistem
- `payload` -> data permintaan/data sistem
- `pool soal` -> kumpulan soal
- `draw` -> ambil acak
- `seed` -> kode acak opsional
- `locked` -> dikunci/terkunci
- `stale` -> kontak server terlalu lama
- `fingerprint` -> penanda perangkat
- `restore` -> pulihkan sesi
- `builder` -> penyusun/editor
- `proctoring` -> pengawasan ruang

## Tahap SOP Kegiatan Asesmen

| Urutan | Key | Label | Tujuan |
| --- | --- | --- | --- |
| 1 | `draft` | Draft | Kegiatan dibuat dan identitas awal tersedia. |
| 2 | `question_authoring` | Pengisian Soal | Guru melengkapi soal sesuai kebutuhan. |
| 3 | `question_verification` | Telaah/Verifikasi Soal | Soal ditelaah sampai siap terbit. |
| 4 | `package_ready` | Paket Siap | Paket soal aktif dan berisi soal. |
| 5 | `participants_rooms_ready` | Peserta & Ruang Siap | Peserta, ruang, kursi, dan pengawas siap. |
| 6 | `tokens_cards_ready` | Token & Kartu Siap | Token ujian/ruang dan kartu ujian siap. |
| 7 | `execution` | Pelaksanaan | Ujian berlangsung dan dipantau pengawas. |
| 8 | `grading` | Koreksi | Nilai otomatis/manual diselesaikan. |
| 9 | `result_verification` | Verifikasi Hasil | Hasil dicek sebelum finalisasi. |
| 10 | `final_archive` | Final & Arsip | Kegiatan ditutup dan dokumen arsip tersedia. |

Status tahap:

- `ready`: Siap
- `warning`: Perlu Perhatian
- `blocked`: Belum Siap
- `running`: Sedang Berjalan

## Pengesahan Formal

Pengesahan yang boleh dicatat secara additive:

- `package` -> Paket Soal
- `participants_rooms` -> Peserta & Ruang
- `tokens_cards` -> Token & Kartu
- `results` -> Hasil
- `final_archive` -> Final & Arsip

Pengesahan tidak boleh menampilkan token raw, answer key, password, API key, atau connection string. Audit metadata boleh menyimpan jenis pengesahan, entity, actor user id, waktu, dan catatan tanpa data rahasia.

## Berita Acara dan Arsip

Dokumen formal kegiatan minimal mencakup:

- Kartu ujian.
- Daftar hadir.
- Berita acara sesi/ruang.
- Rekap hasil.
- Rekap insiden dan tindakan pengawas.
- Audit ringkas tanpa token raw.
- Tanda tangan pengawas, operator, ketua panitia, dan kepala madrasah bila diperlukan.

## Pengecualian Teknis

Nama route, nama field API, nama database column, test contract, dan dokumentasi developer boleh tetap memakai istilah teknis seperti `event_id`, `payload`, atau `fingerprint` selama tidak menjadi copy utama UI pengguna.
