# Paket Metadata Filter + Tingkat Soal

Tanggal: 2026-05-15

## Tujuan
Menerapkan rekomendasi filter metadata untuk Paket Builder sampai operasional, dengan metadata tingkat soal resmi untuk mencegah soal salah tingkat.

## Scope
1. DB/backend:
   - Tambah kolom `target_level` pada `cbt_questions` dengan nilai `VII|VIII|IX` atau kosong untuk legacy.
   - Validasi backend saat create/update soal.
   - Expose target_level pada query/list/detail/pool paket.
   - Update readiness/quality agar metadata gap menghitung target_level kosong.
2. Bank Soal:
   - Tambah input tingkat soal di form tambah/edit.
   - Tampilkan badge tingkat di daftar/kartu.
   - Filter tingkat bila pola UI existing memungkinkan.
3. Paket Builder:
   - Panel Tambah dari Bank Soal memiliki filter:
     - tingkat, jenis soal, status, level kognitif, HOTS, kesulitan, metadata lengkap/kurang, materi, CP/TP/KD search, sort.
   - Default tingkat mengikuti judul paket bila mengandung VII/VIII/IX; tetap bisa manual.
   - Blueprint/mutu menampilkan gap tingkat soal.
4. Validasi:
   - sqlc generate
   - Go tests/build
   - npm check/build
   - deploy/restart PM2 karena user meminta terapkan sampai selesai.

## Guardrail
- Tidak menampilkan credential.
- Migration aman dan nullable untuk data legacy.
- Paket locked tetap tidak bisa diedit.
- Pool soal tetap default published/terbit untuk paket resmi.
