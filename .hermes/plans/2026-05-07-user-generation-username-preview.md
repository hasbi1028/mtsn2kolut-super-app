# User Generation Username + Preview Fix Implementation Plan

> **For Hermes:** Implement bertahap, validasi setiap tahap, jangan tampilkan credential/secret/password/hash.

**Goal:** Perbaiki generate akun pegawai agar username memakai format `NPSN + 2 digit tahun lahir + nomor urut 3 digit`, nomor urut dihitung per tahun lahir, NPSN diambil dari profil/pengaturan sekolah, dan fitur preview menampilkan hasil yang sama dengan generate final.

**Architecture:** Logika username harus berada di backend Go/core-api sebagai source of truth. Frontend `/settings/users` hanya menampilkan preview dari API dan menjalankan generate berdasarkan hasil backend. NPSN tidak boleh hardcode permanen di query; gunakan school profile/settings dengan fallback aman ke nilai existing bila setting belum tersedia.

**Tech Stack:** Go 1.26, sqlc, PostgreSQL, SvelteKit 2/Svelte 5, Vitest, PM2.

---

## Aturan final yang disepakati

Username akun pegawai dibuat dengan format:

```text
<NPSN><YY><SEQ>
```

Contoh:

```text
4040603190001
```

Keterangan:

- `NPSN`: ambil dari pengaturan/profil sekolah, contoh saat ini `40406031`.
- `YY`: 2 digit tahun lahir dari `employees.tanggal_lahir`, bukan tanggal/bulan, bukan 4 digit tahun.
- `SEQ`: nomor urut 3 digit per tahun lahir, mulai `001`.

Contoh:

- Tahun lahir 1990, belum ada username prefix `4040603190` → `4040603190001`.
- Tahun lahir 1990, sudah ada `4040603190001` → berikutnya `4040603190002`.
- Tahun lahir 1985 punya urutan sendiri → `4040603185001`, `4040603185002`, dst.

Preview dan generate wajib memakai logika yang sama.

---

## Task 1: Audit sumber NPSN dan flow generate existing

**Objective:** Pastikan sumber data NPSN/settings dan file yang perlu diubah sebelum patch.

**Files:**

- Inspect: `services/core-api/db/queries/settings.sql` atau query school profile terkait
- Inspect: `services/core-api/internal/service/user_generation.go`
- Inspect: `services/core-api/db/queries/users.sql`
- Inspect: `services/core-api/internal/handler/user.go`
- Inspect: `apps/web-admin/src/routes/settings/users/+page.svelte`
- Inspect: `apps/web-admin/src/lib/client/rbac-users.ts`

**Steps:**

1. Cari query/service school profile yang menyimpan NPSN.
2. Cari method preview dan generate akun pegawai.
3. Catat response shape preview yang dipakai frontend.
4. Pastikan tidak ada credential/secret yang tercetak.

**Verification:**

```bash
git status --short
```

Expected: belum ada perubahan sebelum patch.

---

## Task 2: Tambah test backend untuk format username baru

**Objective:** Kunci aturan username `NPSN + YY + SEQ` dengan test sebelum implementasi.

**Files:**

- Modify/Create test: `services/core-api/internal/service/user_generation_test.go`

**Test cases wajib:**

1. `tanggal_lahir = 1990-05-07`, NPSN `40406031`, belum ada prefix → username `4040603190001`.
2. Ada existing username `4040603190001` → kandidat berikutnya `4040603190002`.
3. Dua pegawai dalam batch dengan tahun lahir sama → preview menghasilkan `...001`, `...002` berurutan.
4. Tahun lahir beda punya sequence sendiri.
5. `tanggal_lahir` kosong → status `missing_birth_date`, username kosong/null.
6. Pegawai sudah punya akun aktif/belum deleted → status `already_linked`.
7. User deleted (`deleted_at IS NOT NULL`) tidak boleh membuat pegawai dianggap already linked, tetapi username historis deleted tetap perlu dipertimbangkan atau dilewati dengan keputusan eksplisit.

**Keputusan untuk deleted username:**

Rekomendasi: username deleted tetap tidak dipakai ulang agar audit/history aman. Jadi sequence mencari max dari semua username, termasuk deleted.

**Run:**

```bash
cd services/core-api
PATH=/usr/local/go/bin:/usr/bin:/bin:$PATH go test ./internal/service -run 'Test.*Employee.*Account|Test.*Username' -count=1
```

Expected awal: FAIL karena implementasi masih lama.

---

## Task 3: Refactor generator username di backend service

**Objective:** Pindahkan aturan username ke helper yang mudah dites dan dipakai preview/generate.

**Files:**

- Modify: `services/core-api/internal/service/user_generation.go`
- Modify if needed: `services/core-api/db/queries/users.sql`
- Regenerate: `services/core-api/internal/repository/postgres/*.go`

**Implementation outline:**

Buat helper kecil di service, contoh konsep:

```go
func birthYearSuffix(t pgtype.Date) (string, bool) {
    if !t.Valid {
        return "", false
    }
    year := t.Time.Year() % 100
    return fmt.Sprintf("%02d", year), true
}

func employeeUsernamePrefix(npsn string, tanggalLahir pgtype.Date) (string, bool) {
    yy, ok := birthYearSuffix(tanggalLahir)
    if !ok {
        return "", false
    }
    return npsn + yy, true
}
```

Untuk sequence:

- Ambil existing username dengan prefix `npsn + yy`.
- Parse 3 digit terakhir.
- Ambil max sequence.
- Untuk batch preview, maintain map `nextSeqByPrefix` agar kandidat pada batch yang sama tidak tabrakan.

Pseudo:

```go
nextSeq := maxExistingSeq[prefix] + 1
username := fmt.Sprintf("%s%03d", prefix, nextSeq)
maxExistingSeq[prefix] = nextSeq
```

**Important:** Hindari hardcode NPSN di SQL. Service menerima `npsn` dari settings/school profile.

**Run:**

```bash
cd services/core-api
PATH=$HOME/go/bin:/usr/local/go/bin:/usr/bin:/bin:$PATH sqlc generate -f db/sqlc.yaml
PATH=/usr/local/go/bin:/usr/bin:/bin:$PATH go test ./internal/service -count=1
```

Expected: service tests PASS.

---

## Task 4: Ambil NPSN dari school profile/settings

**Objective:** NPSN tidak hardcode di generator, tetapi diambil dari data profil sekolah.

**Files:**

- Inspect/Modify: `services/core-api/internal/service/setting.go` atau service setting yang relevan
- Modify: `services/core-api/internal/service/user_generation.go`
- Modify tests/fakes sesuai interface baru

**Behavior:**

1. Jika school profile punya NPSN valid, gunakan itu.
2. Jika belum ada/missing, fallback ke `40406031` untuk backward compatibility, tapi tambahkan TODO/comment agar bisa dikelola dari UI profil sekolah.
3. Validasi NPSN harus numeric dan minimal tidak kosong.

**Verification:**

```bash
cd services/core-api
PATH=/usr/local/go/bin:/usr/bin:/bin:$PATH go test ./internal/service ./internal/handler -count=1
```

Expected: PASS.

---

## Task 5: Perbaiki query preview agar tidak bug dan sama dengan generate

**Objective:** Preview harus menampilkan username final yang akan dibuat saat generate.

**Files:**

- Modify: `services/core-api/internal/service/user_generation.go`
- Modify: `services/core-api/internal/handler/user.go` bila response mapping perlu ditambah
- Modify: `services/core-api/db/queries/users.sql` bila candidate query perlu data tambahan

**Acceptance criteria preview:**

- Klik Preview menampilkan daftar kandidat.
- Kandidat dengan tanggal lahir kosong tampil status `missing_birth_date`.
- Kandidat yang sudah punya akun tampil `already_linked`.
- Kandidat siap generate tampil username `NPSN + YY + SEQ`.
- Jika dua pegawai lahir pada tahun sama, sequence preview tidak sama.
- Jika existing username sudah ada, preview mulai dari sequence berikutnya.

**Run:**

```bash
cd services/core-api
PATH=/usr/local/go/bin:/usr/bin:/bin:$PATH go test ./... -count=1
```

Expected: PASS.

---

## Task 6: Tambah/perbaiki frontend preview UX

**Objective:** Pastikan fitur preview di `/settings/users` benar-benar work dan jelas bagi user.

**Files:**

- Modify: `apps/web-admin/src/routes/settings/users/+page.svelte`
- Modify: `apps/web-admin/src/lib/client/rbac-users.ts`
- Modify test: `apps/web-admin/src/lib/client/rbac-users.test.ts` atau page test terkait bila ada

**UI requirements:**

- Tombol Preview punya loading state.
- Jika API error, tampilkan toast error yang jelas.
- Jika preview kosong, tampilkan empty state.
- Tampilkan kolom/field:
  - Nama pegawai
  - Tanggal lahir
  - Tahun lahir/YY
  - Username hasil preview
  - Status
  - Alasan status bila tidak ready
- Tombol Generate hanya aktif kalau ada kandidat ready.
- Setelah Generate sukses, refresh list user dan preview.

**Run:**

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run test:unit -- src/lib/client/rbac-users.test.ts
```

Expected: PASS.

---

## Task 7: Browser/API smoke test flow preview → generate

**Objective:** Validasi live behavior sebelum commit final.

**Safe smoke approach:**

- Jangan tampilkan password awal di log publik.
- Jika membuat test data, gunakan pegawai dummy hanya bila aman, atau gunakan preview-only smoke untuk data live.
- Untuk generate sebenarnya, pastikan hanya kandidat yang memang seharusnya dibuat.

**Checks:**

1. Login admin ke web-admin.
2. Buka `/settings/users`.
3. Klik preview generate akun pegawai.
4. Pastikan preview tidak error.
5. Pastikan username format 13 digit: `8 digit NPSN + 2 digit YY + 3 digit SEQ`.
6. Pastikan tidak ada duplikat username dalam preview.
7. Pastikan user yang sudah ada tidak ditawarkan generate ulang.

**Optional API smoke:**

Gunakan token admin dari env secara lokal tanpa mencetak token/password.

---

## Task 8: Full validation, restart, health check, commit

**Objective:** Selesaikan sesuai standar repo.

**Commands:**

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
```

Diff:

```bash
git diff --check && git diff --cached --check
```

Restart:

```bash
pm2 restart mtsn2kolut-core-api --update-env
pm2 restart mtsn2kolut-web-admin --update-env
curl -s -o /dev/null -w "core=%{http_code}\n" http://127.0.0.1:8080/health
curl -s -o /dev/null -w "web=%{http_code}\n" http://127.0.0.1:8021/
```

Commit:

```bash
git add services/core-api apps/web-admin
 git commit -m "fix(users): generate employee usernames by birth year"
```

Expected final:

- Tests PASS.
- Build PASS.
- PM2 online.
- Core health `200`.
- Web health `200`.
- Commit hash reported ke user.

---

## Risk notes

- Jangan reuse username deleted untuk menghindari konflik audit/history.
- Jangan hardcode NPSN bila sudah tersedia di school profile/settings.
- Preview dan generate harus berbagi helper/logika backend yang sama agar tidak beda hasil.
- Jangan expose password awal/hash/token/env saat testing.
- Jika ada pegawai dengan tahun lahir sama banyak sekali, sequence 3 digit cukup sampai 999 per tahun lahir; tambahkan guard error jika overflow.

---

## Definition of Done

- Format username final: `NPSN + YY + 3-digit sequence`.
- Sequence dihitung per tahun lahir.
- NPSN diambil dari profil/settings dengan fallback aman.
- Preview work dan sama dengan hasil generate.
- Generate tidak membuat duplikat username.
- User existing/soft-deleted ditangani aman.
- Semua test/build PASS.
- PM2 restart dan health check PASS.
- Commit final dibuat.
