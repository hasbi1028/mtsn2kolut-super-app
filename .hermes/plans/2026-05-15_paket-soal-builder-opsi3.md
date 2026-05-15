# Paket Soal Builder Opsi 3 — Detail, Edit, Bulk Readiness

Tanggal: 2026-05-15
Repo: `/home/servermtsn2kolut/mtsn2kolut-super-app`

## Tujuan
Membuat modul Paket Soal menjadi builder lengkap dan audit-safe:

1. Paket bisa dibuka sebagai detail.
2. Metadata paket bisa diedit.
3. Soal di dalam paket bisa ditambah, dihapus, diatur bobot, dan diurutkan.
4. Paket yang sudah dipakai sesi diberi guard jelas.
5. Paket bisa dikunci/snapshot sebelum ujian.
6. Paket bisa diduplikasi/revisi untuk perubahan aman.
7. Daftar paket punya readiness 20 PG + 5 essay, filter kosong/kurang/siap/locked.
8. Bulk tools untuk UTS: readiness massal dan autofill ringan dari pool soal published.

## Data baseline production terakhir
- Total paket: 112
- Paket event: 112
- Paket kosong: 112
- Paket sudah dipakai sesi: 112
- Paket locked: 0
- Total soal Bank Soal: 31
- Soal published: 1

## Guardrail
- Jangan akses DB langsung dari SvelteKit; semua lewat BFF -> Go API.
- Jangan edit paket locked langsung.
- Paket yang sudah dipakai sesi tetapi belum locked boleh diedit dengan warning dan role manage, karena UTS saat ini masih draft/placeholder; namun UI harus menyarankan lock sebelum ujian.
- Paket yang sudah locked harus memakai duplikasi/revisi.
- Validasi subject dan status published tetap di backend.

## Sprint Implementasi

### Sprint Paket 1 — Backend kontrak editable
- Tambah query:
  - Get package detail + usage/session count.
  - Update package metadata.
  - Replace package questions.
  - Delete package question.
  - Clone package.
  - Bulk readiness overview.
- Tambah service method dengan validasi:
  - duration 1–360.
  - title wajib.
  - questions wajib mapel sama, published, event scope valid.
  - points 1–100.
  - locked package tidak bisa diubah.
- Tambah handler endpoints:
  - `GET /api/asesmen/packages/{id}`
  - `PUT /api/asesmen/packages/{id}`
  - `PUT /api/asesmen/packages/{id}/questions`
  - `POST /api/asesmen/packages/{id}/clone`
  - `POST /api/asesmen/packages/{id}/lock`
  - `GET /api/asesmen/packages/readiness`

### Sprint Paket 2 — UI detail paket
- Route `/asesmen/paket/[id]`.
- Header metadata editable.
- Tab soal dalam paket: reorder, bobot, remove.
- Tab tambah soal dari Bank Soal: filter, add.
- Readiness target 20 PG + 5 essay.
- Blueprint/mutu.
- Lock/snapshot status dan tombol.

### Sprint Paket 3 — Bulk tools
- Tambah filter readiness di daftar paket.
- Bulk readiness panel event.
- Autofill ringan: isi paket kosong/kurang dari pool soal published sesuai mapel dan scope.
- Clone/revisi paket.

## Validasi wajib
```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-paket-builder ./cmd/api
```

## Deploy jika valid
Urutan:
1. Build backend dan restart `mtsn2kolut-core-api`.
2. Build web-admin dan restart `mtsn2kolut-web-admin` segera setelah build.
3. Smoke: backend health, web local 8021, domain public, protected route 302/200.
