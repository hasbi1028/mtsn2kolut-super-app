# Plan: NIP Opsional + Pegawai UID Wajib + Import 8 Honorer

## Goal
Merapikan identitas master pegawai agar:

1. `nip` **tidak wajib** karena honorer belum tentu punya NIP ASN.
2. Tetap mempertahankan `employees.id UUID PRIMARY KEY` sebagai primary key teknis yang stabil.
3. Menambahkan ID manusiawi/wajib: `pegawai_uid`, format berbasis **NPSN + tahun lahir + nomor urut**.
4. Menyesuaikan seluruh data pegawai eksisting agar punya `pegawai_uid`.
5. Memasukkan 8 data honorer ke master pegawai dengan `employment_type='honorer'` dan `nip=NULL`.

## Prinsip desain

### Yang disetujui
- `nip` tidak boleh lagi wajib.
- Semua pegawai tetap harus punya identitas internal wajib.
- Format identitas internal memakai pola sekolah:
  - NPSN MTsN 2 Kolut fallback: `40406031`
  - Tahun lahir 2 digit: `YY`
  - Nomor urut 3 digit per tahun lahir: `001`, `002`, dst.

Contoh:
- lahir 1968 → `4040603168001`
- lahir 2000 urutan pertama → `4040603100001`
- lahir 2000 urutan kedua → `4040603100002`

### Yang tidak saya ubah
- `employees.id` tetap `UUID PRIMARY KEY`.
- Tidak mengganti foreign key yang sudah berjalan ke business ID.
- Tidak membuat akun PUSAKA untuk honorer.

Alasan: business ID boleh berubah jika data lahir dikoreksi, sedangkan PK teknis harus stabil dan tidak memicu perubahan relasi `jobs`, `attendance_records`, user profile link, audit, dsb.

## Data honorer yang akan dimasukkan

| No | Nama | JK | Tempat Lahir | Tanggal Lahir | employment_type | nip |
|---:|---|---|---|---|---|---|
| 1 | KM.Muh.Tang, S.Ag. | L | Jerae | 1968-04-24 | honorer | NULL |
| 2 | Nurilmi, S.Pd | P | Olo-oloho | 2000-10-03 | honorer | NULL |
| 3 | Jayanti Jufri, S.Pd | P | Olo-oloho | 1988-08-15 | honorer | NULL |
| 4 | Nurunnisaa Alimah A, S.Pd | P | Majene | 2000-01-03 | honorer | NULL |
| 5 | Mirnawati, S.Pd | P | Olo-oloho | 2001-04-14 | honorer | NULL |
| 6 | Asniar, S.Pd | P | Olo-oloho | 2000-12-18 | honorer | NULL |
| 7 | KM. Muhammad Yani, S.Pd | L | Mallengngeng | 1992-11-14 | honorer | NULL |
| 8 | Sri Hardianti, S.Or | P | Mikuasi | 1997-05-06 | honorer | NULL |

## Tahap implementasi

### Tahap 1 — Schema migration identitas pegawai
Buat migration baru, kemungkinan nomor berikutnya setelah `076`, misalnya:

`services/core-api/db/migrations/077_employee_uid_nullable_nip_honorer.sql`

Isi utama:

1. Tambah kolom baru:
   - `pegawai_uid text`
   - `jenis_kelamin text` nullable, check `L/P` jika belum ada
   - `tempat_lahir text` nullable jika belum ada

2. Ubah `nip` menjadi nullable:
   - Drop constraint unique lama pada `nip`.
   - `ALTER TABLE employees ALTER COLUMN nip DROP NOT NULL`.
   - Ubah empty string NIP menjadi NULL jika aman.

3. Tambah unique index partial untuk NIP:
   ```sql
   CREATE UNIQUE INDEX IF NOT EXISTS uq_employees_nip_not_null
   ON employees (nip)
   WHERE nip IS NOT NULL AND nip <> '';
   ```

4. Backfill `pegawai_uid` untuk data eksisting:
   - Ambil NPSN dari setting/profil sekolah jika ada; fallback `40406031`.
   - Gunakan `tanggal_lahir` kalau tersedia.
   - Kalau `tanggal_lahir` NULL, fallback aman perlu dibuat, misalnya `00` + sequence khusus, atau lebih baik isi tanggal lahir eksisting dulu bila tersedia.
   - Sequence dihitung per `YY` dengan urutan stabil `created_at, nama, id`.

5. Jadikan `pegawai_uid` wajib dan unik:
   ```sql
   ALTER TABLE employees ALTER COLUMN pegawai_uid SET NOT NULL;
   CREATE UNIQUE INDEX IF NOT EXISTS uq_employees_pegawai_uid ON employees (pegawai_uid);
   ```

### Tahap 2 — Update query sqlc
Update:

`services/core-api/db/queries/employees.sql`

Tambahkan field:
- `pegawai_uid`
- `jenis_kelamin`
- `tempat_lahir`

Ke query:
- `ListEmployees`
- `ListActiveEmployees`
- `GetEmployee`
- `CreateEmployee`
- `UpdateEmployee`
- `ListEmployeesWithStatus`
- `ListPusakaEligibleEmployeesWithStatus`

Catatan penting:
- `ListPusakaEligibleEmployeesWithStatus` tetap hanya `pns/pppk`.
- Honorer tidak masuk employee eligible PUSAKA.

### Tahap 3 — Update backend DTO/service/handler
Update handler/service employee agar:

- `nip` boleh kosong/null.
- `pegawai_uid` wajib pada response.
- Saat create/update pegawai:
  - Jika `pegawai_uid` tidak dikirim, backend generate otomatis.
  - Format generator: `NPSN + YY + 3-digit sequence`.
  - Sequence scoped per prefix `NPSN+YY`.
  - Cegah race condition dengan transaksi/unique retry.
- Validasi `jenis_kelamin` hanya `L`, `P`, atau kosong/null.
- `employment_type='honorer'` boleh dengan `nip=NULL`.
- `pns/pppk` tetap boleh punya NIP dan NIP unik.

Kemungkinan file:
- `services/core-api/internal/handler/employee.go`
- `services/core-api/internal/service/employee*`
- `services/core-api/internal/repository/postgres/*` hasil generate sqlc

### Tahap 4 — Update frontend master pegawai
Update UI master pegawai agar jelas:

- Tampilkan `pegawai_uid` sebagai ID utama pegawai.
- Label `NIP` menjadi opsional, misalnya `NIP (opsional)`.
- Untuk honorer, form tidak memaksa NIP.
- Tambahkan/ tampilkan:
  - `Jenis Kelamin`
  - `Tempat Lahir`
  - `Tanggal Lahir`
  - `Jenis Pegawai`
- Jangan ubah modul publik/dark mode.

Kemungkinan file:
- `apps/web-admin/src/routes/employees/+page.svelte`
- client type/helper employee jika ada di `apps/web-admin/src/lib/client/*`
- BFF proxy jika DTO parsing perlu disesuaikan.

### Tahap 5 — Seed/import 8 honorer
Setelah schema dan backend siap:

- Insert/upsert 8 data honorer secara idempotent.
- `nip=NULL`.
- `employment_type='honorer'`.
- `is_active=true`.
- `pegawai_uid` digenerate dengan rule yang sama seperti existing.
- Tidak insert `pusaka_accounts`.

Strategi idempotensi:
- Hindari duplicate berdasarkan kombinasi `lower(nama) + tanggal_lahir + employment_type` atau `pegawai_uid`.
- Jika data sudah ada, update field yang aman: `nama`, `tanggal_lahir`, `jenis_kelamin`, `tempat_lahir`, `employment_type`, `is_active`.

### Tahap 6 — Validasi backend/frontend
Backend:

```bash
cd services/core-api
PATH=$HOME/go/bin:/usr/local/go/bin:/usr/bin:/bin:$PATH sqlc generate -f db/sqlc.yaml
PATH=/usr/local/go/bin:/usr/bin:/bin:$PATH go test ./...
PATH=/usr/local/go/bin:/usr/bin:/bin:$PATH go build -o bin/api ./cmd/api
```

Frontend:

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run test:unit
npm --prefix apps/web-admin run build
git diff --check && git diff --cached --check
```

### Tahap 7 — Apply live + smoke check
Urutan aman:

1. Apply migration live.
2. Restart core API.
3. Restart web-admin jika frontend berubah.
4. Health check:
   ```bash
   curl -s -o /dev/null -w 'core=%{http_code}\n' http://127.0.0.1:8080/health
   curl -s -o /dev/null -w 'web=%{http_code}\n' http://127.0.0.1:8021/
   ```
5. Smoke DB/API:
   - Semua employee punya `pegawai_uid`.
   - `pegawai_uid` unik.
   - `nip` boleh NULL.
   - NIP non-null tetap unik.
   - Honorer bertambah 8.
   - Honorer tidak punya `pusaka_accounts`.
   - PUSAKA eligible tetap hanya PNS/PPPK.

### Tahap 8 — Commit
Commit terpisah yang disarankan:

1. `feat(employees): add internal employee uid and nullable nip`
2. `feat(employees): import honorer master data`

Atau satu commit jika perubahan kecil:

`feat(employees): add internal uid and seed honorer data`

## Risiko dan mitigasi

### Risiko 1 — Pegawai eksisting ada yang `tanggal_lahir` NULL
Mitigasi:
- Preflight cek jumlah NULL.
- Jika ada, gunakan fallback prefix `00` sementara atau isi dari NIP bila bisa diekstrak.
- Karena user minta data existing disesuaikan, lebih baik semua existing diberi `pegawai_uid`; jangan blokir karena sebagian tanggal lahir kosong.

### Risiko 2 — Existing unique constraint `employees_nip_key`
Mitigasi:
- Drop constraint dengan nama constraint aktual dari DB/schema.
- Buat partial unique index pengganti.

### Risiko 3 — Handler/frontend masih menganggap `nip` string wajib
Mitigasi:
- Update type ke nullable/optional di backend response dan frontend form.
- Di UI tampilkan `—` jika NIP kosong.

### Risiko 4 — Sequence UID bentrok
Mitigasi:
- Unique index di `pegawai_uid`.
- Generator backend pakai transaksi/unique retry.
- Migration backfill menggunakan deterministic row_number per prefix.

### Risiko 5 — Data honorer typo akibat OCR/noisy text
Mitigasi:
- Pakai normalisasi yang sudah disepakati di plan ini.
- Setelah import, tampilkan ringkasan 8 data untuk verifikasi.

## Keputusan final untuk eksekusi
Saya akan eksekusi dengan desain:

- `id UUID` tetap primary key.
- Tambah `pegawai_uid` wajib unik sebagai ID pegawai yang digunakan manusia/aplikasi.
- `nip` nullable + partial unique index.
- Existing data di-backfill `pegawai_uid`.
- 8 honorer diinsert dengan `nip=NULL` dan UID sesuai rule.
- Tempat lahir dan jenis kelamin disimpan di kolom baru jika belum ada.
