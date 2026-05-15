# Paket Soal List — Opsi C Full Paket Operations

Tanggal: 2026-05-15

## Scope
Menerapkan operasi lengkap pada halaman `/asesmen/paket` tanpa migrasi DB:

1. Summary cards: total, kosong, kurang, siap target, locked, dipakai sesi.
2. Search/filter/sort:
   - search paket/mapel/deskripsi
   - mapel
   - readiness
   - aktif/nonaktif
   - lock
   - pemakaian sesi
   - sort operasional
3. Mode Kesiapan UTS:
   - target 20 PG + 5 Essay
   - progress 0/25
   - badge kosong/kurang/siap/locked
4. Export CSV readiness.
5. Bulk selection:
   - select all hasil filter
   - clear selection
   - bulk export selected
   - bulk lock paket yang siap dan belum locked
   - bulk set aktif/nonaktif
6. Quick actions per paket:
   - Detail/Edit
   - Isi Soal
   - Blueprint
   - Lock jika siap
   - Clone/Revisi via detail page
   - Hapus dipindah sebagai aksi destructive sekunder.

## Guardrails
- SvelteKit tetap BFF/proxy; tidak akses DB langsung.
- Bulk lock hanya untuk paket yang `ready` dan belum locked.
- Bulk active/nonactive memakai endpoint metadata existing per paket; locked package tidak diubah.
- Tidak menyentuh untracked seed/plan lama.
- Validasi wajib: npm check/build, sqlc, go test/build.
- Deploy setelah build dan restart PM2 langsung setelah build web-admin.
