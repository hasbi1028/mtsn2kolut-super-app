# SOP Portal Ujian Web — runtime resmi tahun ini

Status: copy operasional web-first. Route `/ujian` diposisikan sebagai **Portal Ujian Web** resmi untuk tahun ini; istilah `Browser Darurat` hanya tersisa sebagai nama legacy/internal pada kontrak lama.

## Prinsip

- `/ujian` adalah jalur operasional resmi siswa tahun ini.
- Portal Ujian Web dipakai dengan pengawasan ruang, kartu/QR/PIN saat rollout siap, dan panel pengawas web.
- Flutter Android/Windows disimpan sebagai source/artifact nonaktif untuk arsip dan tahap lanjutan; jangan jadikan instruksi utama siswa tahun ini.
- Jika toggle lama `allow_web_fallback` masih dipakai, anggap sebagai kontrol kompatibilitas internal untuk mengizinkan Portal Ujian Web per ruang.
- Hanya operator/admin/pengawas berwenang yang boleh membuka/menutup akses web dari panel pengawas ruang bila kontrol ruang masih aktif.
- Wajib alasan aktivasi/nonaktivasi agar audit ruang jelas selama masa transisi.
- Anti-cheat browser hanya best-effort: pindah tab, kehilangan fokus, keluar fullscreen, dan koneksi buruk dicatat sebagai telemetry; tidak diklaim setara aplikasi native.

## Kapan dipakai

Gunakan Portal Ujian Web untuk:

1. Ujian resmi tahun ini melalui web sekolah.
2. Perangkat siswa/lab yang memiliki browser modern.
3. Alur kartu/QR/PIN siswa dan portal pengawasan web.
4. Simulasi/pilot terbatas dengan pengawasan langsung.

Jangan mengarahkan siswa memasang Flutter APK sebagai default tahun ini kecuali ada keputusan operasional baru.

## Alur operator/pengawas

1. Buka panel pengawas ruang: `/asesmen/sesi/{session_id}/rooms/{room_id}/proctoring`.
2. Pastikan peserta, ruang, dan token ruang benar.
3. Pada kartu kontrol web lama (`Browser Darurat /ujian` bila label belum diganti), aktifkan akses Portal Ujian Web untuk ruang sesuai SOP.
4. Isi alasan, contoh: `Portal Ujian Web dibuka untuk ruang sesuai runtime resmi tahun ini`.
5. Berikan URL `/ujian` dan credential yang berlaku (QR+PIN/kartu ujian; token lama hanya saat transisi) kepada peserta.
6. Awasi peserta secara fisik. Jika panel masih menandai `client_type = web_fallback`, baca sebagai peserta Portal Ujian Web.
7. Setelah kondisi normal atau ujian selesai, klik `Nonaktifkan` dan isi alasan penutupan.

## Alur peserta

1. Buka `/ujian` atas arahan pengawas atau dari QR kartu ujian.
2. Masukkan credential yang berlaku untuk sesi (QR+PIN/kartu ujian saat tersedia; token ujian/token ruang hanya pada masa transisi).
3. Klik tombol masuk Portal Ujian Web.
4. Klik `Masuk Fullscreen` bila browser mengizinkan.
5. Jawab soal. Jika koneksi putus, jangan tutup browser; jawaban tertunda disimpan pada `sessionStorage` browser dan dikirim ulang otomatis.
6. Submit hanya saat indikator jawaban lokal/pending kosong.
7. Setelah submit sukses, lapor pengawas sebelum menutup browser.

## Catatan penyimpanan lokal

- Token/jawaban tidak disimpan di `localStorage` permanen.
- Fallback menggunakan `sessionStorage` yang terikat sesi tab/browser.
- Data sesi dibersihkan setelah submit sukses.
- Jika browser/tab ditutup sebelum sinkron, data lokal bisa hilang; pengawas wajib mengarahkan peserta agar tidak menutup browser saat koneksi buruk.

## Telemetry yang dicatat

- `web_fallback_used`
- `web_visibility_hidden` / `web_visibility_visible`
- `web_focus_lost` / `web_focus_restored`
- `web_fullscreen_exit` / `web_fullscreen_restored`
- `web_pending_answer_saved` / `web_pending_answer_flushed`
- `web_connection_degraded` / `web_connection_restored`

## Rollback cepat

1. Nonaktifkan Browser Darurat dari panel ruang.
2. Jika perlu darurat backend: set `allow_web_fallback = false` untuk ruang terkait melalui prosedur DBA/operator yang disetujui.
3. Arahkan peserta ke perangkat/browser cadangan atau prosedur manual yang disetujui; Flutter Android/Windows tetap opsi tahap lanjutan/nonaktif kecuali diputuskan ulang.
4. Catat alasan rollback pada berita acara ruang.

## Acceptance check

- Login `/ujian` ditolak saat room `allow_web_fallback = false`.
- Login `/ujian` diterima saat room `allow_web_fallback = true`, token peserta valid, dan token ruang benar.
- Panel pengawas menampilkan status Browser Darurat dan jumlah peserta web fallback.
- Jawaban dapat disimpan, pending saat offline/kendala, lalu flush saat koneksi pulih.
- Submit ditahan jika masih ada jawaban pending lokal.
- Tidak ada restart/deploy produksi sebelum build dan migrasi disetujui.
