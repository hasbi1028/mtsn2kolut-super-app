# Threat Model Anti-Cheat BYOD Android

Status: draft operasional, 2026-05-17. Dokumen ini melengkapi `docs/flutter-anti-cheat-roadmap.md` dan `apps/mobile/BYOD_TRIAL_PROCEDURE.md`.

## Ringkasan Eksekutif

Model BYOD berarti peserta memakai HP pribadi. Aplikasi tidak memiliki kontrol penuh seperti perangkat sekolah yang dikelola MDM/device-owner. Karena itu tujuan anti-cheat realistis adalah:

- mencegah kecurangan oportunistik dengan friksi yang jelas;
- mendeteksi sinyal risiko penting untuk pengawas;
- menjaga jawaban tetap aman saat koneksi/app lifecycle bermasalah;
- memberi dasar audit pascaujian;
- mengomunikasikan batasan secara jujur kepada sekolah, pengawas, orang tua, dan peserta.

Anti-cheat BYOD **bukan jaminan kiosk penuh**. Keputusan disipliner tidak boleh hanya berdasarkan satu sinyal teknis lemah; harus digabung dengan observasi pengawas, pola jawaban, log server, dan kebijakan sekolah.

## Asumsi dan Aset yang Dilindungi

Asumsi:

- Platform utama: Android BYOD dengan APK Flutter.
- Backend tetap menjadi source of truth untuk token, waktu, status sesi, jawaban tersinkron, submit, skor, dan audit.
- Pengawas berada di ruang ujian dan dapat melakukan verifikasi operasional.
- Sekolah belum menyediakan MDM/perangkat terkelola untuk baseline BYOD.

Aset:

- kerahasiaan soal dan media soal selama ujian;
- integritas jawaban dan status submit;
- identitas peserta/token sesi;
- log audit yang akurat dan minim data pribadi;
- kelancaran ujian bagi peserta jujur.

## Prinsip Desain

- **Server authoritative**: timer, eligibility submit, skor, dan audit tidak bergantung pada klaim client saja.
- **Risk-based, bukan absolut**: sinyal seperti app switch/root/ADB menaikkan risiko, tidak otomatis menjadi bukti tunggal.
- **Minim data pribadi**: jangan kumpulkan screenshot, daftar aplikasi lengkap, IMEI, kontak, lokasi, atau data perangkat yang tidak diperlukan.
- **Fail safe untuk jawaban**: saat koneksi buruk, simpan lokal terenkripsi/terproteksi dan blokir submit jika sinkron belum aman.
- **Human-in-the-loop**: pengawas diberi ringkasan yang dapat ditindaklanjuti, bukan noise teknis mentah.
- **Transparan**: peserta diberi tahu perilaku yang dilarang dan konsekuensi warning.

## Matriks Ancaman, Mitigasi, dan Residual Risk

### 1. App switch / app background

- Ancaman: peserta keluar ke browser/chat/catatan lalu kembali ke ujian.
- Mitigasi teknis:
  - deteksi lifecycle `app_backgrounded` dan `app_resumed`;
  - tampilkan resume gate/warning sebelum lanjut;
  - refresh status ke server saat resume;
  - kirim event dengan timestamp, sequence, durasi background jika tersedia;
  - rate limit agar event tidak membanjiri backend.
- Mitigasi operasional:
  - pengawas menginstruksikan peserta tidak menekan Home/Recent/Back;
  - peserta dengan warning berulang ditempatkan di observasi lebih ketat;
  - kebijakan ambang misalnya 1 kali = peringatan, berulang = catatan audit/intervensi.
- Residual risk: tinggi-sedang. App switch dapat terdeteksi, tetapi isi aplikasi lain tidak diketahui dan event bisa hilang pada perangkat bermasalah.

### 2. Split screen / multi-window / picture-in-picture

- Ancaman: peserta membuka ujian berdampingan dengan catatan/browser/chat.
- Mitigasi teknis:
  - set mode portrait dan non-resizable bila memungkinkan pada Android manifest;
  - deteksi perubahan ukuran window/focus dan kirim event `window_focus_lost`/`multi_window_suspected`;
  - tutup atau blokir interaksi saat app tidak full focus;
  - uji pada vendor populer karena perilaku multi-window berbeda.
- Mitigasi operasional:
  - pengawas memeriksa layar peserta saat login dan selama ujian;
  - larang penggunaan fitur split screen dalam tata tertib.
- Residual risk: sedang. Android/vendor dapat membatasi kemampuan app mendeteksi semua mode multi-window pada BYOD.

### 3. Screenshot

- Ancaman: peserta mengambil gambar soal untuk dibagikan.
- Mitigasi teknis:
  - aktifkan `FLAG_SECURE` di semua screen ujian;
  - event `screenshot_attempt` jika platform/plugin bisa menangkap percobaan, tanpa menyimpan gambar;
  - watermark ringan di UI: nama peserta, sesi, timestamp parsial agar foto eksternal lebih mudah ditelusuri.
- Mitigasi operasional:
  - larang kamera/second device di meja;
  - pengawas melakukan inspeksi visual acak;
  - konsekuensi jelas jika soal difoto/disebarkan.
- Residual risk: sedang-tinggi. `FLAG_SECURE` membantu mencegah screenshot OS, tetapi tidak mencegah foto menggunakan perangkat lain.

### 4. Screen recording / casting / mirroring

- Ancaman: peserta merekam layar atau menyiarkan ke perangkat lain.
- Mitigasi teknis:
  - `FLAG_SECURE` untuk menghambat screen recording/casting standar;
  - deteksi display eksternal/cast jika API tersedia;
  - warning saat fokus/display berubah.
- Mitigasi operasional:
  - pengawas melarang earphone/perangkat tambahan yang tidak perlu;
  - meja bersih dan posisi layar terlihat pengawas.
- Residual risk: tinggi. Recording eksternal dengan kamera kedua tidak bisa dicegah oleh aplikasi BYOD.

### 5. Root, Magisk, custom ROM, emulator, hook framework

- Ancaman: peserta memodifikasi OS/app, melewati deteksi, memalsukan event, atau mengambil data lokal.
- Mitigasi teknis:
  - gunakan Play Integrity/SafetyNet-like check bila distribusi memungkinkan;
  - deteksi indikator root/emulator/debuggable/hooking sebagai risk signal;
  - simpan token/jawaban pending dengan storage aman dan enkripsi;
  - integrity check APK/signature dan TLS pinning bila operasional siap;
  - server jangan percaya event client sebagai bukti tunggal.
- Mitigasi operasional:
  - daftar perangkat berisiko tinggi dapat diminta memakai perangkat pinjaman/ruang khusus;
  - tata tertib menyebut root/custom ROM dapat menyebabkan pemeriksaan tambahan.
- Residual risk: tinggi. Pada perangkat yang dikontrol penuh oleh peserta, deteksi root/custom ROM dapat dibypass.

### 6. Developer options / ADB / USB debugging

- Ancaman: peserta memakai ADB untuk inspect app, inject input, copy local storage, atau automate jawaban.
- Mitigasi teknis:
  - deteksi USB debugging/developer options jika API tersedia dan kirim risk event;
  - jangan simpan token/answer plaintext;
  - obfuscation/minify release build;
  - matikan debug logging dan pastikan build release non-debuggable.
- Mitigasi operasional:
  - instruksikan peserta menonaktifkan Developer Options/USB debugging sebelum ujian;
  - pengawas melarang kabel USB terhubung selain charger yang disetujui.
- Residual risk: sedang-tinggi. Sebagian status developer option tidak selalu dapat dibaca stabil di semua versi Android.

### 7. Overlay / floating window / accessibility abuse

- Ancaman: aplikasi lain menampilkan jawaban di atas layar ujian, auto-clicker, keyboard/clipboard berbahaya, atau accessibility service membaca layar.
- Mitigasi teknis:
  - deteksi overlay obscured touch (`filterTouchesWhenObscured`) untuk tombol penting;
  - blokir submit/aksi penting saat touch obscured;
  - monitor focus loss dan keyboard/clipboard abuse sebatas yang legal;
  - hindari menyalin soal ke clipboard; bersihkan clipboard jika fitur app menyentuhnya.
- Mitigasi operasional:
  - minta peserta menutup floating apps/chat heads/auto-clicker;
  - pengawas cek ikon overlay/floating bubble sebelum mulai.
- Residual risk: tinggi. BYOD tidak dapat melarang semua overlay/accessibility tanpa kontrol perangkat.

### 8. Remote access / remote control / screen sharing app

- Ancaman: pihak luar mengendalikan HP peserta atau melihat layar dari jarak jauh.
- Mitigasi teknis:
  - `FLAG_SECURE` mengurangi screen capture oleh banyak aplikasi;
  - deteksi aplikasi remote populer hanya jika legal, transparan, dan tidak mengumpulkan daftar app penuh;
  - event focus/display/cast mencurigakan.
- Mitigasi operasional:
  - larang aplikasi remote control aktif selama ujian;
  - pengawas memeriksa tidak ada notifikasi remote/cast;
  - jaringan ujian dapat memblokir domain remote access umum bila memakai Wi-Fi sekolah, dengan caveat tidak berlaku untuk data seluler.
- Residual risk: tinggi. Remote access canggih atau via second device sulit dibuktikan oleh app.

### 9. Second device / buku catatan / bantuan fisik

- Ancaman: peserta memakai HP kedua, smartwatch, buku, atau bantuan orang lain di luar aplikasi.
- Mitigasi teknis:
  - watermark soal untuk deterensi kebocoran;
  - randomisasi urutan soal/opsi jika pedagogis dan backend mendukung;
  - paket soal per sesi dan timer server-side.
- Mitigasi operasional:
  - meja bersih, tas dikumpulkan, smartwatch dilepas;
  - tempat duduk berjarak dan pengawas aktif berkeliling;
  - pemeriksaan sebelum ujian dan aturan izin toilet/keluar ruangan.
- Residual risk: tinggi. Ini terutama kontrol operasional, bukan teknis.

### 10. Token sharing / impersonasi

- Ancaman: token dibagikan ke orang lain atau peserta login dari perangkat lain.
- Mitigasi teknis:
  - token single-use/short-lived, terikat sesi dan peserta;
  - batasi satu sesi aktif per peserta, deteksi login ganda/perangkat berganti;
  - server-side session state, heartbeat, dan invalidasi token setelah submit;
  - jangan tampilkan/token mentah di log/event.
- Mitigasi operasional:
  - pembagian token tepat sebelum ujian;
  - verifikasi identitas oleh pengawas saat login;
  - prosedur reset token harus tercatat dan disetujui operator/pengawas.
- Residual risk: sedang. Identitas kuat tetap membutuhkan verifikasi manusia atau perangkat terkelola.

### 11. Collusion / kerja sama antar peserta

- Ancaman: peserta saling memberi jawaban via chat, suara, gestur, atau membandingkan soal.
- Mitigasi teknis:
  - randomisasi urutan soal/opsi;
  - variasi paket soal setara jika siap;
  - analitik pascaujian untuk pola jawaban identik/waktu tidak wajar;
  - proctor dashboard menandai app switch/warning berdekatan antar peserta.
- Mitigasi operasional:
  - pengaturan tempat duduk, pengawas bergerak, larangan komunikasi;
  - jadwal ujian serentak dan kontrol keluar-masuk ruang.
- Residual risk: sedang-tinggi. Kolusi fisik tidak dapat diselesaikan aplikasi saja.

### 12. Network loss / airplane mode / server tidak stabil

- Ancaman: jawaban hilang, timer berbeda, peserta sengaja offline untuk menghindari telemetry, submit tidak terpercaya.
- Mitigasi teknis:
  - local pending answer queue yang aman;
  - heartbeat, stale connection thresholds, dan status chip `Tersambung/Lokal/Waspada/Menurun`;
  - submit diblokir saat pending sync/status tidak aman;
  - server authoritative timer saat reconnect;
  - event `heartbeat_failed`, `stale_connection_attention`, `answer_sync_recovered`.
- Mitigasi operasional:
  - Wi-Fi cadangan, area ujian dengan sinyal memadai, prosedur pindah tempat/perangkat;
  - pengawas memutuskan intervensi saat status `Menurun`;
  - latihan/uji coba perangkat sebelum ujian besar.
- Residual risk: sedang. Offline sementara dapat ditangani, tetapi offline panjang mengurangi visibilitas dan bisa membutuhkan keputusan manual.

### 13. Privacy / legal / kepercayaan publik

- Ancaman: sekolah dianggap memata-matai HP pribadi, mengumpulkan data berlebihan, atau memberi sanksi dari bukti lemah.
- Mitigasi teknis:
  - data minimization: event anti-cheat hanya timestamp, tipe event, severity, session context, dan metadata teknis terbatas;
  - tidak mengumpulkan kontak, lokasi, foto, daftar aplikasi lengkap, IMEI, file pribadi, audio, atau raw screenshot;
  - retention policy untuk log ujian;
  - role-based access untuk dashboard pengawas/admin;
  - audit trail akses data.
- Mitigasi operasional:
  - informed notice/consent sesuai kebijakan sekolah;
  - dokumen tata tertib menjelaskan data yang dikumpulkan dan tidak dikumpulkan;
  - prosedur banding/validasi manual sebelum sanksi berat;
  - persetujuan orang tua bila diperlukan.
- Residual risk: sedang. Risiko reputasi tetap ada jika komunikasi buruk atau ada false positive.

## Matriks Opsi Ringkas

| Dimensi | MVP BYOD | Standard BYOD | Strict BYOD |
|---|---|---|---|
| Tujuan | Pilot/rehearsal aman dasar | Produksi realistis untuk ujian sekolah | Friksi tinggi untuk ujian bernilai tinggi sementara |
| Kontrol teknis inti | Token single-use, heartbeat, lifecycle event, pending answer safe, submit gate, `FLAG_SECURE` | Semua MVP + resume risk, multi-window/focus/overlay signal, watermark, device/session continuity, integrity signal, dashboard ringkas | Semua Standard + pre-flight wajib, approval/block untuk risk tinggi, submit gate ketat, Wi-Fi filtering bila ada |
| Kontrol operasional | Briefing, meja bersih, token saat mulai, catatan manual | SOP intervensi, pengawas aktif, trial perangkat, laporan event pascaujian | Check-in lebih awal, perangkat cadangan, pengawas tambahan, prosedur banding eksplisit |
| Privasi | Paling minim data | Seimbang, masih minim dan transparan | Perlu komunikasi lebih kuat karena lebih banyak check perangkat |
| False positive | Rendah | Sedang | Sedang-tinggi |
| Biaya/kompleksitas | Rendah | Sedang | Tinggi |
| Residual risk utama | Root/overlay/remote/second device | Root canggih, second device, kolusi fisik | Second device, OS termodifikasi canggih, beban support |
| Rekomendasi penggunaan | Uji coba dan ujian risiko rendah-menengah | Baseline produksi yang direkomendasikan | Hanya jika ada fallback perangkat dan kebijakan siap |

## Opsi Implementasi BYOD

### Opsi A - MVP BYOD

Tujuan: siap uji coba lapangan dengan kontrol dasar, biaya rendah, dan klaim terbatas.

Teknis:

- release APK non-debuggable;
- login token single-use/short-lived;
- server authoritative timer/status/submit;
- heartbeat dan event lifecycle: background/resume, warning koneksi, restore, submit blocked;
- local answer pending queue yang tidak plaintext;
- `FLAG_SECURE` di screen ujian;
- back navigation/route escape ditahan;
- dashboard/log sederhana untuk pengawas/operator.

Operasional:

- briefing peserta dan pengawas;
- meja bersih, larangan second device, token dibagi saat mulai;
- prosedur reset token dan resume;
- catatan manual untuk warning berulang.

Kelebihan:

- cepat diterapkan dan relatif aman untuk pilot;
- rendah risiko privasi karena data minimal.

Kekurangan/residual risk:

- belum kuat terhadap root/overlay/remote access;
- sangat bergantung pada pengawasan fisik.

Cocok untuk:

- latihan, ujian rendah-menengah risiko, rehearsal, sekolah yang belum siap MDM.

### Opsi B - Standard BYOD

Tujuan: baseline produksi realistis untuk ujian sekolah dengan audit lebih baik.

Tambahan teknis di atas MVP:

- resume gate dengan risk scoring sederhana;
- deteksi multi-window/focus loss/obscured touch untuk aksi penting;
- event severity dan aggregation agar dashboard tidak bising;
- watermark identitas/sesi/timestamp di UI soal;
- device/session continuity check untuk login ganda/perangkat berganti;
- Play Integrity atau integrity check serupa sebagai risk signal bila distribusi mendukung;
- obfuscation/minify, certificate/signature check, dan TLS hardening sesuai kapasitas;
- proctor dashboard: ringkasan peserta berisiko, koneksi menurun, app switch berulang, login ganda.

Operasional:

- SOP ambang intervensi;
- trial perangkat per vendor;
- pengawas aktif berkeliling;
- kebijakan false positive dan hak klarifikasi;
- laporan pascaujian berisi ringkasan event, bukan data pribadi berlebihan.

Kelebihan:

- keseimbangan terbaik antara deterrence, privasi, dan kelancaran;
- cukup jujur untuk BYOD tanpa klaim kiosk.

Kekurangan/residual risk:

- masih bisa dilewati perangkat rooted/custom atau bantuan second device;
- butuh disiplin SOP dan dashboard yang rapi.

Cocok untuk:

- ujian sekolah reguler dengan risiko sedang-tinggi selama ada pengawas fisik.

### Opsi C - Strict BYOD

Tujuan: meningkatkan friksi terhadap peserta berisiko, tetap bukan kiosk penuh.

Tambahan teknis di atas Standard:

- pre-flight check wajib: OS minimum, developer option/USB debugging, root/custom ROM/emulator indikasi, overlay aktif, koneksi stabil;
- blokir atau require pengawas approval untuk risk tinggi tertentu (misalnya emulator/root jelas, USB debugging aktif, perangkat berganti saat ujian);
- stricter submit gate: pending sync nol, status server fresh, no unresolved high severity events;
- device attestation jika feasible;
- jaringan Wi-Fi sekolah dengan filtering dasar untuk domain remote/collaboration umum;
- audit export untuk komite ujian.

Operasional:

- peserta wajib datang lebih awal untuk check perangkat;
- perangkat cadangan/pinjaman disediakan untuk peserta yang gagal pre-flight;
- surat pemberitahuan lebih eksplisit tentang data dan batasan;
- pengawas tambahan untuk menangani false positive dan eskalasi.

Kelebihan:

- deterrence lebih kuat dan lebih banyak titik intervensi;
- lebih sesuai untuk ujian bernilai tinggi jika belum ada perangkat sekolah.

Kekurangan/residual risk:

- false positive meningkat;
- dukungan teknis lebih berat;
- risiko persepsi privasi lebih tinggi;
- tetap tidak mengalahkan second device dan OS yang dimodifikasi canggih.

Cocok untuk:

- ujian bernilai tinggi sementara, hanya jika sekolah siap menyediakan fallback perangkat dan prosedur banding.

## Rekomendasi Praktis

Rekomendasi baseline: **Standard BYOD** untuk produksi, dengan **MVP BYOD** sebagai tahap pilot dan **Strict BYOD** hanya untuk sesi tertentu yang risikonya lebih tinggi.

Urutan implementasi:

1. Selesaikan MVP: event lifecycle, heartbeat, local pending answer safety, submit gate, `FLAG_SECURE`, token single-use.
2. Tambahkan dashboard pengawas dan SOP intervensi.
3. Tambahkan Standard hardening: multi-window/focus/overlay signal, watermark, device/session continuity, integrity signal.
4. Jalankan rehearsal di beberapa merek/versi Android dan ukur false positive.
5. Gunakan Strict hanya bila ada perangkat cadangan dan persetujuan kebijakan privasi.

## Batasan yang Wajib Dikomunikasikan

Kepada sekolah:

- BYOD tidak sama dengan kiosk/MDM. Aplikasi tidak dapat mengunci total HP pribadi.
- Anti-cheat teknis harus dipadukan dengan pengawasan ruang, tata tertib, dan audit pascaujian.
- Sinyal seperti app switch/root/ADB adalah indikator risiko, bukan bukti tunggal kecurangan.
- Untuk ujian sangat tinggi risikonya, pertimbangkan perangkat sekolah terkelola atau laboratorium komputer.
- Sekolah perlu menyiapkan SOP reset token, perangkat gagal, jaringan buruk, banding, dan retensi log.

Kepada peserta/orang tua:

- Aplikasi mengirim log terbatas terkait ujian: login/status, heartbeat, app keluar/masuk, gangguan koneksi, resume, submit, dan peringatan keamanan tertentu.
- Aplikasi tidak mengambil kontak, galeri, lokasi, audio, file pribadi, atau screenshot mentah.
- Peserta dilarang membuka aplikasi lain, split screen, screenshot/rekam layar, remote access, memakai second device, atau membagikan token.
- Jika perangkat rooted/custom ROM/developer options/overlay aktif, peserta mungkin diminta menonaktifkan fitur, memakai perangkat lain, atau mendapat pemeriksaan pengawas.
- Jika koneksi menurun, peserta harus melapor dan tidak memaksa submit sampai status aman.

Kepada pengawas:

- Fokus pada sinyal yang dapat ditindaklanjuti: warning berulang, status `Menurun`, login ganda/perangkat berganti, app switch berulang, peserta terlihat memakai perangkat lain.
- Jangan langsung menghukum dari satu event; catat, verifikasi kondisi fisik, dan ikuti SOP.
- Pastikan submit hanya saat status sehat dan pending jawaban kosong.

## Checklist Minimum Sebelum Go-Live

- APK release non-debuggable dan `FLAG_SECURE` aktif di seluruh screen ujian.
- Token single-use/short-lived dan server authoritative timer/status/submit.
- Heartbeat, app background/resume, stale connection, restore, pending answer, submit blocked tercatat di backend.
- Jawaban lokal tidak plaintext dan tersinkron saat koneksi pulih.
- Dashboard pengawas menampilkan ringkasan risiko per peserta.
- SOP tertulis untuk app switch, koneksi buruk, perangkat gagal, reset token, dan banding.
- Pemberitahuan privasi dan tata tertib BYOD dibagikan sebelum ujian.
- Rehearsal lintas perangkat dilakukan dan false positive dicatat.
