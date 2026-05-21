# Laporan Implementasi ID Card Siswa Terpadu — Opsi 3

Tanggal: 2026-05-22 WITA

## Ringkasan

Implementasi full wave ID Card Siswa Terpadu MTsN 2 Kolut telah ditambahkan ke Super App tanpa deploy/restart otomatis.

Cakupan utama:

- Backend foundation: migration, sqlc query, service, handler, route, RBAC permission.
- Admin UI/BFF: menu Kesiswaan → Kartu Siswa, daftar kartu, generate, revoke/lost/damaged/suspended/reissue, riwayat dan audit.
- Template kartu: desain Alternatif A hijau-emas formal madrasah, preview depan-belakang, mode cetak CR80/A4 via print stylesheet.
- Public verify: `/s/idc/[token]` dan BFF `/api/public/student-cards/verify/[token]` dengan data terbatas.
- Portal siswa QR+PIN: start challenge, verifikasi PIN, set cookie session via BFF jika akun siswa aktif tersedia.
- Integrasi scan: endpoint dan halaman petugas untuk presensi kegiatan, lookup perpustakaan, dan validasi CBT tanpa mengembalikan token ujian/ruang.
- Audit: event kartu, audit log kartu, audit public verify, portal login, scan presensi/perpustakaan/CBT.

## Prinsip keamanan yang diterapkan

- QR memakai token acak; server menyimpan hash HMAC, bukan raw token.
- Raw token hanya dikembalikan saat generate/reissue agar dapat dicetak menjadi QR; list/detail/audit disaring agar tidak membocorkan `token_hash`, `pin_hash`, atau `password_hash`.
- Public verify hanya menampilkan status valid/tidak valid dan identitas terbatas.
- Portal siswa tetap wajib QR + PIN dan memakai lockout PIN bertahap.
- CBT scan hanya validasi identitas/kepesertaan; tidak membuka atau mengembalikan token ujian/ruang.
- Kartu hilang/rusak/revoke/reissue mengubah lifecycle dan kartu reissue memakai token baru.

## Test gate

Berhasil dijalankan:

- `cd services/core-api && go test ./...`
- `cd services/core-api && go build ./cmd/api`
- `cd apps/web-admin && npm run check`
- `cd apps/web-admin && npm run build`

Catatan build frontend:

- Vite menampilkan warning lama: `NODE_ENV=production is not supported in the .env file`. Build tetap sukses.
- Tidak ada deploy/restart PM2 dalam implementasi ini.

## Catatan operasional sebelum deploy

1. Jalankan migrasi database sesuai prosedur produksi.
2. Pastikan permission RBAC baru diberikan ke role yang tepat:
   - `id_cards.read`
   - `id_cards.manage`
   - `id_cards.print`
   - `id_cards.scan`
   - `id_cards.audit`
3. Pastikan siswa memiliki akun aktif dan PIN portal sebelum memakai login QR+PIN.
4. Untuk kartu yang tercetak, gunakan QR dari hasil generate/reissue terbaru.
5. Jika kartu hilang, ubah status ke `lost` atau `revoked`, lalu reissue agar token lama mati.
