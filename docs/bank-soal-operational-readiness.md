# Bank Soal Operational Readiness

Dokumen ini merangkum hasil integrasi bertahap Bank Soal standalone pada web-admin MTsN 2 Kolaka Utara.

## Route UI aktif

| Area | Route | Fungsi | Data/API utama |
| --- | --- | --- | --- |
| Dashboard | `/bank-soal` | KPI, distribusi, aktivitas, shortcut operasional | `/api/bank-soal/summary`, `/api/bank-soal/questions` |
| Daftar soal | `/bank-soal/daftar` | Search, filter, pagination, aksi per soal | `/api/bank-soal/questions` |
| Komposer | `/bank-soal/tambah` | Buat/edit soal, autosave, preview siswa, draft/review | `/api/bank-soal/questions`, workflow/timeline aliases |
| Import | `/bank-soal/impor` | Template, upload, preview dry-run, import final | `/api/bank-soal/template`, `/api/bank-soal/import-legacy` |
| Review | `/bank-soal/verifikasi` | Queue review, approve/revisi, timeline, catatan reviewer | `/api/bank-soal/questions`, workflow/timeline aliases |
| Analisis butir | `/bank-soal/analisis-butir` | Distribusi tipe/status/HOTS dan tindak lanjut | `/api/bank-soal/questions`, `/api/bank-soal/summary` |
| Mapel & KD | `/bank-soal/mapel-kd` | Coverage mapel, KD/CP/TP, metadata belum lengkap | `/api/bank-soal/soal-support/subjects`, `/api/bank-soal/questions` |
| Pengaturan & SOP | `/bank-soal/pengaturan` | SOP operasional, workflow, standar kualitas, integrasi | `/api/bank-soal/summary`, `/api/bank-soal/soal-support/subjects` |

## Namespace API aktif

Fitur Bank Soal baru wajib memakai namespace aktif:

- `/api/bank-soal/*`
- `/bank-soal/*`

Jangan menambahkan fitur baru pada namespace legacy/retired:

- `/api/cbt`
- `/cbt`
- `/bank-soal/komposer`
- `/bank-soal/import`
- `/bank-soal/review`

## Checklist operasional sebelum periode asesmen aktif

1. Dashboard `/bank-soal` terbuka dan KPI tampil tanpa error JS.
2. Daftar `/bank-soal/daftar` bisa search/filter dan membuka edit soal.
3. Komposer `/bank-soal/tambah` menyimpan draft, autosave lokal, dan bisa ajukan review.
4. Import `/bank-soal/impor` bisa download template, dry-run, lalu import final.
5. Review `/bank-soal/verifikasi` bisa approve dan minta revisi dengan catatan.
6. Analisis `/bank-soal/analisis-butir` menampilkan distribusi status, tipe, dan level kognitif.
7. Mapel & KD `/bank-soal/mapel-kd` menampilkan coverage metadata kurikulum.
8. Pengaturan `/bank-soal/pengaturan` menjadi referensi SOP workflow dan standar kualitas.
9. Paket asesmen tetap berada di `/asesmen/paket`, bukan di Bank Soal storage.
10. Guard test akses memastikan semua route final Bank Soal tetap bisa diakses role guru/admin dan tidak ikut admin-only Asesmen.
11. Jalankan validasi:

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run test:unit
cd services/core-api && go test ./...
npm --prefix apps/web-admin run build
deploy/scripts/health-check.sh bank-soal
git diff --check && git diff --cached --check
```

## Catatan desain

Integrasi Bank Soal mengikuti blueprint SCS tanpa mengimpor runtime prototype React/CDN/Babel. Implementasi produksi tetap native SvelteKit dan memakai endpoint live yang tersedia.
