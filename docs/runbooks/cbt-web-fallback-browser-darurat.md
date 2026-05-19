# SOP CBT Web/PWA Fallback — Browser Darurat

Status: implementasi fitur, belum deploy produksi.

## Prinsip

- Flutter Android/Windows tetap jalur utama ujian resmi.
- `/ujian` adalah jalur cadangan darurat, bukan pengganti aplikasi native.
- Default OFF per ruang ujian.
- Hanya operator/admin/pengawas berwenang yang boleh mengaktifkan dari panel pengawas ruang.
- Wajib alasan aktivasi/nonaktivasi agar audit ruang jelas.
- Anti-cheat browser hanya best-effort: pindah tab, kehilangan fokus, keluar fullscreen, dan koneksi buruk dicatat sebagai telemetry; tidak diklaim setara aplikasi native.

## Kapan dipakai

Gunakan Browser Darurat hanya untuk:

1. Perangkat peserta tidak bisa memasang/menjalankan APK CBT.
2. Build Windows/lab bermasalah pada hari-H.
3. Perangkat cadangan hanya memiliki browser modern.
4. Simulasi/pilot terbatas dengan pengawasan langsung.

Jangan gunakan untuk ujian resmi besar sebagai jalur default.

## Alur operator/pengawas

1. Buka panel pengawas ruang: `/asesmen/sesi/{session_id}/rooms/{room_id}/proctoring`.
2. Pastikan peserta, ruang, dan token ruang benar.
3. Pada kartu `Browser Darurat /ujian`, klik `Aktifkan Darurat`.
4. Isi alasan, contoh: `APK gagal dibuka di perangkat peserta; disetujui pengawas ruang`.
5. Berikan URL `/ujian`, token ujian peserta, dan token ruang kepada peserta bermasalah saja.
6. Awasi peserta secara fisik. Panel menandai peserta `Browser Darurat` lewat `client_type = web_fallback`.
7. Setelah kondisi normal atau ujian selesai, klik `Nonaktifkan` dan isi alasan penutupan.

## Alur peserta

1. Buka `/ujian` hanya atas arahan pengawas.
2. Masukkan token ujian dan token ruang.
3. Klik `Masuk Mode Browser Darurat`.
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
3. Arahkan peserta kembali ke Flutter Android/Windows.
4. Catat alasan rollback pada berita acara ruang.

## Acceptance check

- Login `/ujian` ditolak saat room `allow_web_fallback = false`.
- Login `/ujian` diterima saat room `allow_web_fallback = true`, token peserta valid, dan token ruang benar.
- Panel pengawas menampilkan status Browser Darurat dan jumlah peserta web fallback.
- Jawaban dapat disimpan, pending saat offline/kendala, lalu flush saat koneksi pulih.
- Submit ditahan jika masih ada jawaban pending lokal.
- Tidak ada restart/deploy produksi sebelum build dan migrasi disetujui.
