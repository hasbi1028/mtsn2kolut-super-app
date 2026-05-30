# Rebuild Asesmen/CBT From Zero — Web-First Simple Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task after user approval. Keep the first deliverable prototype-only. Do not recreate the old complex Asesmen module.

**Goal:** Membangun ulang modul Asesmen/CBT dari nol dengan alur sederhana, familiar untuk operator MTsN 2 Kolaka Utara, dan aman untuk pengguna gaptek.

**Architecture:** Bank Soal tetap berdiri sendiri sebagai sumber soal. Asesmen/CBT baru dibangun sebagai lapisan ujian web-first: admin/panitia mengatur ujian lewat checklist ringkas, peserta memakai portal web-mobile `/ujian`, dan pengawas memakai portal web-mobile `/pengawas-ujian`. Implementasi dimulai dari prototype UI tanpa backend, lalu dipromosikan bertahap setelah disetujui.

**Tech Stack:** SvelteKit 2 + Svelte 5 + Tailwind v4 + shadcn-svelte untuk frontend; Go Chi + sqlc + PostgreSQL untuk backend tahap produksi; PM2 adapter-node untuk deploy web-admin.

---

## 1. Keputusan Produk

### 1.1 Prinsip utama

1. **Sederhana dulu, lengkap nanti.** Default UI hanya menampilkan pekerjaan harian yang benar-benar dipakai.
2. **Bank Soal bukan Asesmen.** Bank Soal tetap modul terpisah: buat, tinjau, publish, impor, cetak, arsip soal.
3. **Web-first untuk tahun awal.** Peserta dan pengawas memakai web-mobile yang terasa seperti aplikasi HP, bukan Flutter/APK sebagai jalur utama.
4. **Familiar seperti CBT lama.** Gunakan istilah yang sudah dikenal: `Ruang Ujian`, `Jadwal Sesi`, `Kartu Peserta`, `Lembar Pengawas`, `Rekap Nilai`.
5. **Pengawas tidak perlu login admin.** Pengawas ruang masuk lewat QR+PIN Lembar Pengawas Ruang.
6. **Admin tetap memegang kendali teknis.** Unlock/reset/force submit/audit teknis tidak muncul di portal pengawas sederhana.
7. **Jangan bangun ulang kompleksitas lama.** Hindari banyak tab event/session/package/proctoring/SOP/approval pada tahap awal.

### 1.2 Target permukaan baru

| Permukaan | Route | Pengguna | Fungsi |
|---|---|---|---|
| Panitia/Admin | `/asesmen` | Admin/operator/panitia | Checklist ujian 4 langkah |
| Peserta | `/ujian` | Siswa | Masuk QR/PIN, kerjakan soal, kumpulkan |
| Pengawas | `/pengawas-ujian` | Pengawas ruang | Mulai ujian, pantau peserta, lapor admin |
| Prototype | `/asesmen/prototype` | Reviewer internal | Validasi alur sebelum produksi |

---

## 2. Batas Modul

### 2.1 Dipertahankan dari sistem sekarang

- Bank Soal dan semua API `/api/bank-soal/*`.
- Asset soal dan kompatibilitas file asset yang masih dipakai Bank Soal.
- Data siswa, rombel, guru, mapel dari modul akademik/kesiswaan.
- Core auth/RBAC Super App untuk admin/panitia.

### 2.2 Dibangun ulang, bukan restore dari lama

- Kegiatan ujian.
- Jadwal/sesi ujian.
- Ruang ujian.
- Peserta ujian.
- Kartu Peserta QR+PIN.
- Lembar Pengawas Ruang QR+PIN.
- Portal `/ujian`.
- Portal `/pengawas-ujian`.
- Rekap hasil minimal.

### 2.3 Tidak dibuat dulu

- Anti-cheat berat/kiosk mode.
- Flutter/APK sebagai jalur utama.
- Approval/SOP bertingkat.
- Proctoring detail admin-grade.
- Event archive/finalization kompleks.
- Mode laporan besar.
- Integrasi nilai/rapor otomatis sebelum alur ujian dasar stabil.

---

## 3. Model Alur UI yang Direkomendasikan

### 3.1 Admin/Panitia: Checklist Ujian 4 Langkah

Halaman `/asesmen` produksi nanti hanya berisi empat area utama:

1. **Siapkan Ujian**
   - Nama ujian.
   - Mapel/kelas/rombel.
   - Pilih paket dari Bank Soal.
   - Tentukan tanggal dan jam.

2. **Atur Peserta & Ruang**
   - Pilih kelas/rombel.
   - Buat ruang otomatis.
   - Preview pembagian peserta.
   - Simpan pembagian.

3. **Cetak Kartu & Lembar Pengawas**
   - Kartu Peserta Ujian: QR + PIN.
   - Lembar Pengawas Ruang: QR + PIN per ruang/sesi.
   - Daftar hadir sederhana.

4. **Pelaksanaan & Hasil**
   - Status ruang: belum mulai, berlangsung, selesai, perlu dicek.
   - Jumlah peserta masuk/sedang mengerjakan/selesai.
   - Koreksi PG otomatis.
   - Koreksi esai/manual nanti tahap lanjut.
   - Rekap nilai.

### 3.2 Peserta: `/ujian`

Flow wajib sederhana:

1. Scan QR atau masuk kode kartu.
2. Masukkan PIN.
3. Konfirmasi identitas: nama, kelas, ruang, mapel.
4. Jika belum dimulai: tampilkan `Ujian belum dimulai — tunggu pengawas`.
5. Jika sudah dimulai: tampilkan soal satu per satu.
6. Navigasi bawah: `Sebelumnya`, `Ragu-ragu`, `Berikutnya`, `Kumpulkan`.
7. Status simpan: `Tersimpan`, `Menunggu koneksi`, `Aman di perangkat`.

### 3.3 Pengawas: `/pengawas-ujian`

Flow wajib sederhana:

1. Scan QR Lembar Pengawas Ruang.
2. Masukkan PIN ruang.
3. Masuk ke ruang/sesi yang tertera di lembar.
4. Tombol utama: `Mulai Ujian`.
5. Tab bawah:
   - `Ruang`: status dan aksi utama.
   - `Peringatan`: peserta yang perlu dicek.
   - `Peserta`: daftar peserta dan status.
6. Aksi aman:
   - `Sudah Dicek`.
   - `Beri Peringatan`.
   - `Hubungi Admin`.
7. Aksi sensitif seperti unlock/reset/force submit tetap milik admin.

---

## 4. Tahapan Implementasi

## Tahap 1 — Prototype UI Only

**Tujuan:** Membuktikan alur dan rasa UI sebelum backend dibuat.

**Scope:**
- Create: `apps/web-admin/src/routes/asesmen/prototype/+page.svelte`
- Create: `apps/web-admin/src/routes/asesmen/prototype/page-ux.test.ts`
- Optional create: `apps/web-admin/src/routes/ujian-demo-prototype` tidak perlu jika mock bisa ditampilkan di prototype.

**Batas aman:**
- Tidak ada backend.
- Tidak ada migration.
- Tidak ada BFF.
- Tidak ada route-access production.
- Tidak ada sidebar production kecuali user meminta preview live.
- Tidak ada `fetch()`.
- Tidak ada panggilan `/api/*`.

**Isi prototype:**
- Label jelas: `PROTOTYPE UI — belum terhubung backend`.
- Kode review:
  - A0 — Command Center CBT Sederhana.
  - A1 — Siapkan Ujian.
  - A2 — Atur Peserta & Ruang.
  - A3 — Cetak Kartu & Lembar Pengawas.
  - A4 — Pelaksanaan Hari-H.
  - A5 — Hasil & Rekap.
  - A9 — Mode Lengkap Admin, kecil/sekunder.
- Panel khusus: `Mengikuti CBT lama`.
- Struktur menu familiar: `Ruang Ujian`, `Jadwal Sesi`, `Paket Ujian`, `Proctoring Live`, `Rekap Nilai`.
- Mock mobile preview peserta dan pengawas.

**Verification:**

```bash
npm --prefix apps/web-admin run test:unit -- --run src/routes/asesmen/prototype/page-ux.test.ts
npm --prefix apps/web-admin run check
rm -rf apps/web-admin/build
npm --prefix apps/web-admin run build
git diff --check
```

**Review checkpoint:** User menilai apakah alur, istilah, dan tampilan sudah cocok sebelum masuk Tahap 2.

---

## Tahap 2 — Production Shell Minimal

**Tujuan:** Menghidupkan route produksi kosong/aman setelah prototype disetujui.

**Files target:**
- Create: `apps/web-admin/src/routes/asesmen/+page.svelte`
- Modify: `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts`
- Modify: `apps/web-admin/src/lib/server/route-access.ts`
- Modify tests terkait sidebar/route-access.

**Isi:**
- Sidebar hanya satu entry: `Asesmen Ujian`.
- `/asesmen` menampilkan checklist 4 langkah dengan data mock/empty state aman.
- Belum ada mutasi data.
- Tombol produksi yang belum aktif diberi label `Belum aktif` atau `Tahap berikutnya`.

**Verification:**

```bash
npm --prefix apps/web-admin run test:unit -- --run src/lib/components/sidebar/sidebar-config.test.ts src/lib/server/route-access.test.ts
npm --prefix apps/web-admin run check
rm -rf apps/web-admin/build
npm --prefix apps/web-admin run build
```

---

## Tahap 3 — Data Model Minimal Backend

**Tujuan:** Membuat fondasi database paling kecil untuk ujian.

**Entitas minimal:**

1. `assessment_exams`
   - id
   - title
   - subject_id
   - grade_level
   - status: draft, ready, running, finished, archived
   - starts_at, ends_at
   - created_by, created_at, updated_at

2. `assessment_sessions`
   - id
   - exam_id
   - title
   - starts_at, ends_at
   - status

3. `assessment_rooms`
   - id
   - session_id
   - code
   - name
   - capacity
   - status

4. `assessment_participants`
   - id
   - session_id
   - room_id
   - student_id
   - status
   - started_at, submitted_at

5. `assessment_access_cards`
   - id
   - card_type: participant/proctor
   - session_id
   - room_id nullable for participant
   - participant_id nullable
   - token_hash
   - pin_hash
   - status
   - failed_attempts
   - expires_at

**Catatan:** Gunakan nama baru `assessment_*` agar tidak bergantung ke tabel CBT lama. Jangan drop data lama tanpa permintaan eksplisit.

**Backend files:**
- Create migration in `services/core-api/db/migrations/`.
- Create queries in `services/core-api/db/queries/assessment_*.sql`.
- Generate sqlc.
- Create service + handler minimal.
- Register routes in `services/core-api/cmd/api/main.go`.

**API minimal admin:**
- `GET /api/asesmen/exams`
- `POST /api/asesmen/exams`
- `GET /api/asesmen/exams/{id}`
- `PATCH /api/asesmen/exams/{id}`
- `POST /api/asesmen/exams/{id}/prepare-rooms`
- `POST /api/asesmen/exams/{id}/issue-cards`

**Verification:**

```bash
cd services/core-api
/home/servermtsn2kolut/.local/go/bin/go test ./internal/handler ./internal/service -count=1 -timeout 180s
/home/servermtsn2kolut/.local/go/bin/go build -o bin/api ./cmd/api
```

---

## Tahap 4 — Admin BFF + UI Real Data

**Tujuan:** Menghubungkan `/asesmen` ke backend minimal.

**BFF files:**
- `apps/web-admin/src/routes/api/asesmen/exams/+server.ts`
- `apps/web-admin/src/routes/api/asesmen/exams/[id]/+server.ts`
- `apps/web-admin/src/routes/api/asesmen/exams/[id]/prepare-rooms/+server.ts`
- `apps/web-admin/src/routes/api/asesmen/exams/[id]/issue-cards/+server.ts`

**UI behavior:**
- List ujian.
- Create ujian draft.
- Checklist status.
- Empty state ramah operator.
- Tidak ada tab kompleks.

**Verification:**

```bash
npm --prefix apps/web-admin run check
rm -rf apps/web-admin/build
npm --prefix apps/web-admin run build
```

---

## Tahap 5 — Portal Peserta `/ujian`

**Tujuan:** Peserta bisa masuk QR+PIN, menunggu, mengerjakan, dan submit.

**Frontend files:**
- Create: `apps/web-admin/src/routes/ujian/+page.svelte`
- Create tests for route shell and no admin chrome.

**Backend/API minimal:**
- `POST /api/ujian/card/verify`
- `POST /api/ujian/start`
- `GET /api/ujian/status`
- `GET /api/ujian/questions`
- `POST /api/ujian/answer`
- `POST /api/ujian/submit`

**Security:**
- QR token opaque.
- PIN hashed.
- No answer key in response.
- Body limit.
- Rate limit failed PIN.
- Do not expose raw token/PIN except print/issue response.

**UI constraints:**
- Force light mode.
- `max-w-md`, `min-h-dvh`, mobile-first.
- One question per screen.
- Large touch targets.
- Friendly copy.

---

## Tahap 6 — Portal Pengawas `/pengawas-ujian`

**Tujuan:** Pengawas bisa memulai dan memantau ruang tanpa login admin.

**Frontend files:**
- Create: `apps/web-admin/src/routes/pengawas-ujian/+page.svelte`
- Create test: standalone mobile shell, no sidebar/admin chrome.

**Backend/API minimal:**
- `POST /api/pengawas-ujian/card/verify`
- `POST /api/pengawas-ujian/room/dashboard`
- `POST /api/pengawas-ujian/room/start`
- `POST /api/pengawas-ujian/room/finish`
- `POST /api/pengawas-ujian/participants/{id}/warn`
- `POST /api/pengawas-ujian/help-request`

**UI tabs:**
- `Ruang`
- `Peringatan`
- `Peserta`

**Do not include:**
- unlock public
- reset access public
- force submit public
- raw event logs
- technical risk score

---

## Tahap 7 — Hasil Minimal

**Tujuan:** Memberi rekap yang cukup untuk madrasah.

**Scope awal:**
- Jumlah peserta.
- Sudah masuk.
- Sudah submit.
- Nilai PG otomatis.
- Ekspor CSV sederhana.

**Scope nanti:**
- Koreksi esai.
- Remedial.
- Sinkron nilai ke rapor.
- Analisis butir.

---

## 5. Route dan Permission Model

### 5.1 Admin/panitia

Permission baru yang disarankan:

- `asesmen.read`
- `asesmen.manage`
- `asesmen.cards_issue`
- `asesmen.proctor_admin`
- `asesmen.results_read`
- `asesmen.results_manage`

### 5.2 Public card-gated

Routes berikut public secara halaman, tetapi action backend tetap card-gated:

- `/ujian`
- `/pengawas-ujian`
- `/api/ujian/*`
- `/api/pengawas-ujian/*`

Public bukan berarti bebas. Setiap action wajib membawa QR token + PIN/session card proof.

---

## 6. UI Copy Standard

Gunakan:

- `Asesmen Ujian`
- `Command Center CBT`
- `Siapkan Ujian`
- `Ruang Ujian`
- `Jadwal Sesi`
- `Kartu Peserta Ujian`
- `Lembar Pengawas Ruang`
- `Mulai Ujian`
- `Tutup Ujian`
- `Panggil Pengawas`
- `Hubungi Admin`
- `Rekap Nilai`

Hindari untuk user biasa:

- heartbeat
- device fingerprint
- token hash
- risk score
- SSE
- web fallback
- browser darurat
- force submit
- unlock
- incident action

---

## 7. Visual Direction

### 7.1 Admin `/asesmen`

- Compact admin normal.
- Bukan card dashboard besar.
- Gunakan checklist/action-row/table kecil.
- Maksimal 4 area utama.
- `Mode Lengkap Admin` kecil di bawah/kanan.

### 7.2 Peserta `/ujian`

- Mobile app feel.
- Light mode paksa.
- Large buttons.
- One-question-per-screen.
- Bottom navigation/action bar.
- Warning copy manusiawi.

### 7.3 Pengawas `/pengawas-ujian`

- Mobile app feel.
- Header ruang sticky.
- Bottom tabs: Ruang/Peringatan/Peserta.
- Warna status jelas: hijau, kuning, oranye, merah.
- Aksi sedikit dan aman.

---

## 8. Acceptance Criteria

Tahap 1 diterima jika:

- Prototype tersedia di `/asesmen/prototype`.
- Tidak ada backend/fetch/API call.
- A0–A5 dan A9 terlihat jelas.
- Ada preview alur admin, peserta, pengawas.
- Istilah familiar CBT lama muncul.
- Bank Soal terlihat sebagai modul terpisah, bukan bagian Asesmen.
- `npm check` dan build pass.

Tahap produksi minimal diterima jika:

- `/asesmen` hanya satu menu sederhana di sidebar.
- `/ujian` dan `/pengawas-ujian` tidak memakai admin chrome/sidebar.
- QR+PIN tidak membocorkan token mentah.
- Kartu peserta dan lembar pengawas bisa dicetak.
- Peserta bisa submit ujian PG minimal.
- Pengawas bisa mulai/tutup ujian dan kirim bantuan ke admin.
- Bank Soal tetap aktif dan tidak regresi.

---

## 9. Verification Full Gate Sebelum Commit Produksi

Frontend:

```bash
npm --prefix apps/web-admin run test:unit -- --run
npm --prefix apps/web-admin run check
rm -rf apps/web-admin/build
npm --prefix apps/web-admin run build
```

Backend jika Go berubah:

```bash
cd services/core-api
/home/servermtsn2kolut/.local/go/bin/go test ./cmd/api ./internal/handler ./internal/service -count=1 -timeout 180s
/home/servermtsn2kolut/.local/go/bin/go build -o bin/api ./cmd/api
```

Security:

```bash
npm run security:scan-secrets:staged
npm run security:scan-secrets
```

Smoke setelah deploy:

```bash
curl -fsS http://127.0.0.1:8080/health
curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8021/asesmen
curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8021/ujian
curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8021/pengawas-ujian
curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:8080/api/bank-soal/questions
```

Expected:
- Protected admin routes may return `302` when logged out.
- Public `/ujian` and `/pengawas-ujian` should return `200` after implemented.
- Bank Soal API should return `401` when unauthenticated, not `404`.

---

## 10. Recommended Next Action

Start with:

```text
Tahap 1: Frontend-only prototype di /asesmen/prototype.
```

Do not build backend yet. The first deliverable is a reviewable UI artifact for approval.

After user approves prototype:

1. promote core flow to `/asesmen`;
2. add sidebar entry;
3. then build backend minimal;
4. then connect QR+PIN student/proctor portals.

---

## 11. Implementation Task List for Tahap 1

### Task 1: Create prototype route

**Objective:** Add safe frontend-only route for review.

**Files:**
- Create: `apps/web-admin/src/routes/asesmen/prototype/+page.svelte`
- Create: `apps/web-admin/src/routes/asesmen/prototype/page-ux.test.ts`

**Requirements:**
- Contains `PROTOTYPE UI — belum terhubung backend`.
- Contains no `fetch(`.
- Contains no `/api/` string.
- Contains A0, A1, A2, A3, A4, A5, A9.
- Shows admin checklist, student preview, proctor preview.

### Task 2: Add source-level regression test

**Objective:** Prevent prototype from accidentally becoming production/backend-connected.

**Test assertions:**
- no `fetch(`
- no `/api/`
- contains `PROTOTYPE UI`
- contains `Mengikuti CBT lama`
- contains `Ruang Ujian`, `Jadwal Sesi`, `Proctoring Live`, `Rekap Nilai`
- A9 appears after A0–A5 and is marked secondary

### Task 3: Run frontend gates

```bash
npm --prefix apps/web-admin run test:unit -- --run src/routes/asesmen/prototype/page-ux.test.ts
npm --prefix apps/web-admin run check
rm -rf apps/web-admin/build
npm --prefix apps/web-admin run build
git diff --check
```

### Task 4: Commit prototype only

```bash
git status --short
git add apps/web-admin/src/routes/asesmen/prototype/+page.svelte apps/web-admin/src/routes/asesmen/prototype/page-ux.test.ts
git commit -m "Prototype simple Asesmen CBT rebuild"
```

Do not include screenshots, reports, credentials, or unrelated files.
