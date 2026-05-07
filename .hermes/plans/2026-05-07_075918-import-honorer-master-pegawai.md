# Plan: Import Data Honorer ke Master Pegawai

## Goal
Memasukkan 8 data honorer yang diberikan ke master pegawai (`employees`) secara aman, tanpa mengganggu data ASN/PNS/PPPK dan tanpa membuat akun PUSAKA untuk honorer.

## Data dari user (hasil normalisasi awal)

| No | Nama | JK | Tempat Lahir | Tanggal Lahir | Jenis Pegawai |
|---:|---|---|---|---|---|
| 1 | KM.Muh.Tang, S.Ag. | L | Jerae | 1968-04-24 | honorer |
| 2 | Nurilmi, S.Pd | P | Olo-oloho | 2000-10-03 | honorer |
| 3 | Jayanti Jufri, S.Pd | P | Olo-oloho | 1988-08-15 | honorer |
| 4 | Nurunnisaa Alimah A, S.Pd | P | Majene | 2000-01-03 | honorer |
| 5 | Mirnawati, S.Pd | P | Olo-oloho | 2001-04-14 | honorer |
| 6 | Asniar, S.Pd | P | Olo-oloho | 2000-12-18 | honorer |
| 7 | KM. Muhammad Yani, S.Pd | L | Mallengngeng | 1992-11-14 | honorer |
| 8 | Sri Hardianti, S.Or | P | Mikuasi | 1997-05-06 | honorer |

## Asumsi
- Field inti di `employees` yang relevan saat ini: `nip`, `nama`, `unit_kerja`, `employment_type`, `tanggal_lahir`, `is_active`.
- `employment_type` mendukung `honorer`.
- Karena honorer kemungkinan tidak punya NIP ASN, kita perlu memilih strategi aman untuk kolom `nip` yang unique/not-null.
- Data tempat lahir dan jenis kelamin belum terlihat pada query `employees.sql`, jadi perlu cek apakah kolomnya sudah ada di schema/handler/UI. Jika belum ada, jangan paksa masuk tanpa migration.
- Honorer tidak eligible untuk PUSAKA berdasarkan query saat ini: `pusaka_eligible = employment_type IN ('pns','pppk')`. Jadi tidak perlu insert `pusaka_accounts`.

## Open question sebelum eksekusi
1. Untuk kolom `nip`, apakah honorer sudah punya nomor identitas resmi/NIK/NUPTK? Jika belum, saya sarankan gunakan kode internal deterministic seperti `HON-YYYYMMDD-001` atau `HON-<slug>-<tanggal_lahir>` agar unik dan mudah dilacak.
2. Apakah `tempat_lahir` dan `jenis_kelamin` sudah wajib tampil di master pegawai? Jika belum ada kolomnya, ada dua opsi:
   - Opsi cepat: simpan hanya `nama`, `tanggal_lahir`, `employment_type=honorer`, `unit_kerja`, `is_active` dulu.
   - Opsi lengkap: tambah kolom `jenis_kelamin` dan `tempat_lahir` lewat migration + UI/API.

## Recommended approach
Gunakan pendekatan staged dan aman:

### Tahap 1 — Preflight schema dan data existing
- Cek schema `employees` aktual di migration/schema DB.
- Cek apakah kolom `jenis_kelamin` dan `tempat_lahir` sudah ada.
- Cek constraints kolom `nip` dan `employment_type`.
- Cek kemungkinan duplikat berdasarkan nama + tanggal lahir.

### Tahap 2 — Normalisasi dan mapping data
- Bersihkan input yang pecah baris:
  - `Ja\nya\nnti` → `Jayanti`
  - `Nuru\nnn\nis\naa` → `Nurunnisaa`
  - `S\n.P\nd` → `S.Pd`
  - `14 - 04 - 2001` → `2001-04-14`
- Standardisasi tempat lahir:
  - `Olo-oloho`
  - `Mallengngeng`
  - `Mikuasi`
- Standardisasi gender:
  - `L`
  - `P`
- `employment_type='honorer'`
- `is_active=true`
- `unit_kerja` default: `MTsN 2 Kolaka Utara` atau sesuai value existing yang dipakai aplikasi.

### Tahap 3 — Pilih strategi `nip` untuk honorer
Prioritas:
1. Jika user punya NIK/NUPTK/NIP non-ASN: gunakan nomor itu.
2. Jika tidak ada: generate kode internal unik, contoh:
   - `HON-19680424-001` untuk KM.Muh.Tang
   - `HON-20001003-001` untuk Nurilmi

Catatan: strategi kode internal sebaiknya dipakai hanya jika kolom `nip` memang wajib dan belum tersedia kolom identitas honorer terpisah.

### Tahap 4 — Implement import
Ada dua opsi implementasi:

#### Opsi A — Insert langsung via migration seed/idempotent
Cocok jika data ini dianggap seed/master resmi.
- Tambah migration baru misalnya `077_seed_honorer_employees.sql`.
- Gunakan `INSERT ... ON CONFLICT (nip) DO UPDATE` agar idempotent.
- Tidak membuat `pusaka_accounts`.

#### Opsi B — Import via endpoint/backend service existing
Cocok jika ingin lewat aplikasi/audit route.
- Gunakan API create employee bila sudah ada endpointnya.
- Jalankan script smoke/import sekali.
- Tetap hindari mencetak secret/token.

Rekomendasi saya: **Opsi A** jika data honorer ini master resmi dan harus ikut deploy konsisten; **Opsi B** jika user ingin semua perubahan tercatat sebagai aksi aplikasi/audit.

### Tahap 5 — Jika kolom tempat lahir/gender belum ada
Jika user ingin data lengkap tersimpan:
- Tambahkan migration untuk kolom:
  - `jenis_kelamin text CHECK (jenis_kelamin IN ('L','P'))`
  - `tempat_lahir text`
- Update sqlc queries `employees.sql`.
- Update Go DTO/service/handler employee.
- Update UI master pegawai agar kolom tampil/editable.
- Jalankan `sqlc generate` dan test backend/frontend.

Jika user hanya butuh masuk ke master pegawai sekarang:
- Tunda kolom ini, masukkan `tanggal_lahir` dan `nama` dulu.

## Files likely to change
Jika hanya import seed:
- `services/core-api/db/migrations/077_seed_honorer_employees.sql`

Jika perlu support gender/tempat lahir lengkap:
- `services/core-api/db/migrations/077_employee_birthplace_gender.sql`
- `services/core-api/db/migrations/078_seed_honorer_employees.sql`
- `services/core-api/db/queries/employees.sql`
- `services/core-api/internal/handler/*employee*`
- `services/core-api/internal/service/*employee*`
- `apps/web-admin/src/routes/employees/+page.svelte`
- BFF/API client files jika ada route proxy khusus employee.

## Validation plan
Backend:
```bash
cd services/core-api
PATH=$HOME/go/bin:/usr/local/go/bin:/usr/bin:/bin:$PATH sqlc generate -f db/sqlc.yaml
PATH=/usr/local/go/bin:/usr/bin:/bin:$PATH go test ./...
PATH=/usr/local/go/bin:/usr/bin:/bin:$PATH go build -o bin/api ./cmd/api
```

Frontend jika UI employees berubah:
```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run test:unit
npm --prefix apps/web-admin run build
git diff --check && git diff --cached --check
```

DB/live verification:
- Apply migration.
- Query count honorer bertambah 8.
- Pastikan `employment_type='honorer'`.
- Pastikan tidak ada `pusaka_accounts` baru untuk honorer.
- Pastikan halaman master pegawai menampilkan honorer.

Deploy:
```bash
pm2 restart mtsn2kolut-core-api --update-env
pm2 restart mtsn2kolut-web-admin --update-env
curl -s -o /dev/null -w 'core=%{http_code}\n' http://127.0.0.1:8080/health
curl -s -o /dev/null -w 'web=%{http_code}\n' http://127.0.0.1:8021/
```

## Risks
- Data input dari user mengalami line-break/OCR noise; nama perlu diverifikasi sebelum insert final.
- `KM.Muh.Tang` mungkin perlu spasi/format gelar yang benar: `KM. Muh. Tang, S.Ag.` atau tetap seperti input.
- Jika kolom `nip` wajib, memakai kode internal honorer harus disepakati supaya tidak bentrok dengan kebijakan username/generation berikutnya.
- Jika gender/tempat lahir belum ada di schema, menyimpan data lengkap butuh migration dan UI/API update, bukan sekadar seed.

## Proposed execution after approval
1. Saya cek schema aktual `employees` dan endpoint master pegawai.
2. Saya konfirmasi strategi `nip` honorer.
3. Saya buat migration idempotent untuk 8 honorer.
4. Jika perlu, saya tambahkan kolom `jenis_kelamin` dan `tempat_lahir` dulu.
5. Saya jalankan validasi backend/frontend.
6. Saya apply migration live, restart PM2, health/smoke check.
7. Saya commit dengan pesan seperti `feat(employees): seed honorer master data`.
