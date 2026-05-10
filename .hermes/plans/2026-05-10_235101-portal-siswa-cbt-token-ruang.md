# Portal Siswa CBT + Token Ruang Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Siswa dapat melihat jadwal ujian CBT mendatang di Portal Siswa, mendapatkan akses token ujian secara terbatas/aman, dan APK memvalidasi token siswa + token ruang sebelum ujian dimulai.

**Architecture:** Gunakan data CBT existing (`cbt_sessions`, `cbt_exam_rooms`, `cbt_exam_participants`, event table) tanpa migrasi destruktif. Portal siswa menampilkan jadwal dan kartu ujian aman; token siswa hanya dibuka sesuai window konfigurasi dan/atau gate token ruang. APK tetap memakai token peserta sebagai credential utama, tetapi login bisa ditingkatkan untuk menerima `room_token` sebagai second factor.

**Tech Stack:** SvelteKit web-admin, Go core-api, PostgreSQL/sqlc, Flutter APK siswa.

---

## Prinsip desain

1. **Token siswa tetap individual** — tidak diganti oleh token ruang.
2. **Token ruang menjadi second factor / kunci ruang** — token siswa bocor tidak cukup untuk login tanpa token ruang.
3. **Portal siswa boleh menampilkan jadwal ujian** — tetapi token ujian tidak tampil terlalu awal.
4. **Pengawas tetap pegang kontrol** — token ruang dicetak di paket pengawas dan bisa diumumkan saat ujian dimulai.
5. **Tidak destruktif** — tidak menghapus token lama, tidak merusak APK lama sebelum transisi.

---

## Alur target setelah implementasi

### Sebelum hari ujian

Siswa login Portal Siswa dan melihat:

- nama ujian
- mapel/paket
- tanggal dan jam
- ruang
- nomor meja
- status: `Belum dibuka`

Token siswa belum ditampilkan.

### Menjelang ujian

Jika sudah masuk window, misalnya H-0 atau 15 menit sebelum mulai:

- siswa melihat tombol **Buka Kartu Ujian**
- sistem meminta token ruang atau menampilkan instruksi: “Minta token ruang ke pengawas”

### Saat ujian aktif

Siswa memakai APK CBT:

- input token siswa
- input token ruang
- APK kirim `token`, `room_token`, `device_fingerprint`
- backend validasi:
  - token siswa valid
  - peserta berada di ruang tersebut
  - token ruang cocok dengan ruang peserta
  - sesi aktif
  - waktu ujian sudah masuk window
  - device fingerprint valid

### Jika token ruang salah

APK menolak dengan pesan:

> Token ruang tidak sesuai. Pastikan Anda berada di ruang ujian yang benar dan minta token dari pengawas.

---

# Phase A — Backend CBT access contract

## Task A1: Tambahkan helper validasi token ruang di service exam

**Objective:** `service.Exam.Login` bisa memvalidasi `room_token` jika dikirim.

**Files:**
- Modify: `services/core-api/internal/service/exam.go`
- Test: `services/core-api/internal/service/exam_test.go`

**Implementation idea:**

Tambahkan parameter `roomToken` ke `Login`:

```go
Login(ctx context.Context, token, roomToken, deviceFingerprint, loginIP string) (LoginResult, error)
```

Validasi:

```go
func validateParticipantRoomToken(p db.GetParticipantByTokenRow, roomToken string) error {
    roomToken = strings.TrimSpace(roomToken)
    if roomToken == "" {
        return ErrRoomTokenRequired
    }
    if strings.TrimSpace(p.RoomToken) == "" || !strings.EqualFold(strings.TrimSpace(p.RoomToken), roomToken) {
        return ErrRoomTokenMismatch
    }
    return nil
}
```

Catatan: jika `GetParticipantByTokenRow` belum memuat `room_token`, update query sqlc.

**Tests:**

- login gagal jika `room_token` kosong ketika peserta punya ruang
- login gagal jika `room_token` salah
- login berhasil jika `room_token` cocok
- login peserta tanpa ruang tetap gagal dengan pesan ruang belum ditetapkan

**Commands:**

```bash
cd services/core-api
go test ./internal/service -run 'TestExamLogin.*RoomToken' -v
```

---

## Task A2: Update query `GetParticipantByToken` agar membawa room token

**Objective:** Backend login punya data `room_token` tanpa query tambahan.

**Files:**
- Modify: `services/core-api/db/queries/cbt_sessions.sql`
- Generate: `services/core-api/internal/repository/postgres/cbt_sessions.sql.go`

**Implementation idea:**

Di query participant token, join/field existing room:

```sql
r.room_token
```

Pastikan `LEFT JOIN cbt_exam_rooms r ON r.id = ep.room_id` sudah ada atau tambahkan.

**Commands:**

```bash
cd services/core-api
sqlc generate
go test ./internal/repository/... ./internal/service/... -v
```

---

## Task A3: Update handler `/api/exam/login`

**Objective:** API login menerima `room_token` dari APK.

**Files:**
- Modify: `services/core-api/internal/handler/exam.go`
- Test: `services/core-api/internal/handler/exam_test.go`

**Request body target:**

```json
{
  "token": "TOKEN-SISWA",
  "room_token": "TOKEN-RUANG",
  "device_fingerprint": "device-hash"
}
```

**Error mapping:**

- `ErrRoomTokenRequired` → HTTP 400
- `ErrRoomTokenMismatch` → HTTP 403
- `ErrExamRoomRequired` → HTTP 403

**Commands:**

```bash
cd services/core-api
go test ./internal/handler -run 'TestExamLogin.*RoomToken' -v
```

---

# Phase B — Portal Siswa CBT schedule

## Task B1: Tambahkan endpoint portal siswa untuk jadwal CBT

**Objective:** Portal siswa dapat mengambil daftar ujian CBT mendatang dan status akses token.

**Files:**
- Modify/Create backend portal student handler/service/query sesuai pola existing:
  - `services/core-api/internal/handler/student_portal*.go` atau file existing portal siswa
  - `services/core-api/internal/service/student_portal*.go`
  - `services/core-api/db/queries/*.sql`
- Modify proxy/client web-admin:
  - `apps/web-admin/src/lib/client/student-portal.ts`

**Response target:**

```ts
type StudentCbtScheduleItem = {
  participant_id: string;
  session_id: string;
  session_title: string;
  package_title: string | null;
  scheduled_start: string;
  scheduled_end: string;
  duration_minutes: number;
  room_id: string | null;
  room_name: string | null;
  seat_no: number | null;
  status: 'upcoming' | 'token_window' | 'active' | 'submitted' | 'closed' | 'locked';
  can_reveal_token: boolean;
  requires_room_token: boolean;
  token_masked: string | null;
};
```

**Rules:**

- upcoming: jadwal belum masuk window
- token_window: sudah dekat waktu mulai, kartu ujian boleh dibuka
- active: sesi aktif dan waktu valid
- submitted: sudah submit
- locked: dikunci anti-cheat

**Commands:**

```bash
cd services/core-api
go test ./... 
```

---

## Task B2: Tambahkan UI jadwal ujian di Portal Siswa

**Objective:** Siswa melihat jadwal CBT terpisah dari jadwal pelajaran mingguan.

**Files:**
- Modify: `apps/web-admin/src/routes/portal/siswa/+page.svelte`
- Modify: `apps/web-admin/src/lib/client/student-portal.ts`

**UI section:**

```text
Jadwal Ujian CBT
- Nama ujian
- Waktu
- Ruang
- Meja
- Status
- Tombol Kartu Ujian / Belum Dibuka
```

**Status labels:**

- Belum dibuka
- Siap dibuka
- Sedang berlangsung
- Selesai
- Dikunci pengawas

**Token policy:**

Jangan tampilkan token mentah di tabel utama. Tampilkan hanya masked token, misalnya:

```text
ABCD-••••-••••
```

**Commands:**

```bash
cd apps/web-admin
npm run check
```

---

## Task B3: Halaman kartu ujian siswa

**Objective:** Siswa punya halaman kartu ujian yang print/mobile-friendly.

**Files:**
- Create: `apps/web-admin/src/routes/portal/siswa/cbt/[participant_id]/+page.svelte`
- Add API client/helper if needed.

**Kartu menampilkan:**

- nama siswa
- NIS/NISM
- kelas
- nama ujian
- waktu ujian
- ruang
- nomor meja
- status akses
- instruksi: “Token ruang diberikan oleh pengawas saat ujian dimulai.”

**Token display:**

- Jika belum masuk window: token siswa tidak tampil.
- Jika masuk window: boleh tampilkan token siswa dengan tombol “Tampilkan token” dan warning.
- Jika memakai room-token-first flow: kartu hanya menampilkan instruksi, bukan token siswa.

**Preferred safer mode:**

Tampilkan QR/deeplink hanya setelah validasi token ruang.

---

# Phase C — Token reveal dengan validasi token ruang

## Task C1: Endpoint reveal token siswa berbasis token ruang

**Objective:** Portal siswa bisa membuka token siswa hanya jika token ruang benar.

**Endpoint target:**

```text
POST /api/student/portal/cbt/:participant_id/reveal-token
```

**Body:**

```json
{
  "room_token": "TOKEN-RUANG"
}
```

**Response:**

```json
{
  "token": "TOKEN-SISWA",
  "expires_at": "2026-05-11T00:15:00+08:00"
}
```

**Security rules:**

- authenticated student only
- participant must belong to logged-in student
- room token must match participant room
- reveal only inside configured time window
- log event:

```text
student_portal_token_reveal
```

**Files:**
- Backend handler/service/query in core-api
- Web-admin proxy route if portal calls through SvelteKit API
- Tests for ownership and wrong token room

---

## Task C2: UI modal input token ruang

**Objective:** Portal siswa meminta token ruang sebelum menampilkan token siswa.

**Files:**
- Modify: `apps/web-admin/src/routes/portal/siswa/+page.svelte`
- Modify/Create modal component if reusable.

**UX:**

1. Klik **Buka Token Ujian**
2. Modal:

```text
Masukkan Token Ruang dari pengawas
[______]
```

3. Jika benar:

```text
Token Ujian Anda: XXXX-XXXX-XXXX
Gunakan hanya di perangkat Anda sendiri.
```

4. Jika salah:

```text
Token ruang tidak sesuai. Pastikan Anda berada di ruang yang benar.
```

---

# Phase D — APK login dengan token ruang

## Task D1: Update `ExamApiClient.login`

**Objective:** APK mengirim `room_token` saat login.

**Files:**
- Modify: `apps/mobile/lib/src/exam_api.dart`
- Test: `apps/mobile/test/exam_api_test.dart`

**API method:**

```dart
Future<ExamLoginPayload> login({
  required String token,
  required String roomToken,
  required String deviceFingerprint,
})
```

**JSON:**

```dart
body: <String, Object?>{
  'token': token,
  'room_token': roomToken,
  'device_fingerprint': deviceFingerprint,
}
```

---

## Task D2: Update layar login APK

**Objective:** APK punya input token siswa + token ruang.

**Files:**
- Modify: `apps/mobile/lib/src/screens/exam_login_screen.dart`
- Test: `apps/mobile/test/widget_test.dart` or dedicated login screen test.

**Fields:**

- Token Ujian / Token Siswa
- Token Ruang
- Server URL operator setting tetap tersembunyi seperti sekarang

**Validation:**

- token siswa 8–64 karakter
- token ruang minimal 4 karakter atau sesuai generator existing
- pesan error jelas jika token ruang salah

**Error message mapping:**

HTTP 403 with room-token mismatch:

```text
Token ruang tidak sesuai. Pastikan Anda berada di ruang ujian yang benar.
```

---

# Phase E — Admin/pengawas controls

## Task E1: Tambahkan kebijakan token di pengaturan sesi

**Objective:** Admin bisa mengontrol kapan token boleh dibuka.

**Minimal no-migration option:** gunakan field settings/json existing jika ada. Jika belum ada, gunakan constant backend dulu.

**Policy target:**

```json
{
  "student_portal_token_window_minutes": 15,
  "require_room_token_for_exam_login": true,
  "allow_student_portal_token_reveal": true
}
```

**Default:**

- token reveal: 15 menit sebelum mulai
- room token required: true
- portal reveal enabled: false untuk sesi existing, true untuk sesi baru jika admin pilih

---

## Task E2: Perjelas print pack pengawas

**Objective:** Paket pengawas menjelaskan pemakaian token ruang.

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/print-pack/+page.svelte`

**Text target:**

```text
Token Ruang dipakai untuk membuka akses token siswa di portal dan/atau validasi login APK. Jangan diberikan sebelum peserta berada di ruang ujian.
```

Checklist tambahkan:

```text
Token ruang diumumkan hanya setelah peserta siap di ruang.
```

---

# Phase F — Audit, monitoring, and migration safety

## Task F1: Event audit token reveal dan room token mismatch

**Objective:** Pengawas/admin bisa melihat percobaan buka token dan kegagalan token ruang.

**Events:**

```text
student_portal_token_reveal
student_portal_room_token_mismatch
exam_room_token_mismatch
```

**Fields:**

- participant_id
- session_id
- room_id
- student_id
- actor/student username
- timestamp
- IP hash if available

---

## Task F2: Integrasi ke proctoring evidence

**Objective:** Event token reveal/mismatch muncul di rekap proctoring.

**Files:**
- Modify: `apps/web-admin/src/lib/cbt/proctor-evidence.ts`
- Modify report pages if needed:
  - `apps/web-admin/src/routes/asesmen/sesi/[id]/proctoring/report/+page.svelte`
  - `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/report/+page.svelte`

**Mapping:**

- token reveal → info
- room token mismatch → warning
- repeated mismatch → high risk candidate

---

# Phase G — Validation and deploy

## Commands wajib

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/services/core-api
gofmt -w internal/handler/*.go internal/service/*.go
go test ./...
```

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/apps/web-admin
npm run check
npm run build
```

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/apps/mobile
flutter test
```

## Deploy boundary

Restart hanya jika diminta/di akhir validasi:

```bash
pm2 restart mtsn2kolut-core-api --update-env
pm2 restart mtsn2kolut-web-admin --update-env
```

APK perlu build/release terpisah:

```bash
/home/servermtsn2kolut/mtsn2kolut-super-app/scripts/publish-mobile-apk.sh
```

---

# Acceptance Criteria

## Portal siswa

- Siswa melihat daftar ujian CBT mendatang.
- Siswa melihat ruang dan nomor meja.
- Token siswa tidak tampil sebelum window.
- Token siswa hanya bisa dibuka dengan token ruang yang benar.
- Siswa tidak bisa membuka token peserta lain.

## APK

- APK login meminta token siswa + token ruang.
- Login gagal jika token ruang salah.
- Login gagal jika token siswa benar tapi ruang salah.
- Login tetap mengikat perangkat pertama.
- Pesan error mudah dipahami siswa/pengawas.

## Admin/pengawas

- Token ruang tetap tampil di paket pengawas ruang.
- Paket pengawas menjelaskan kapan token ruang diberikan.
- Event token reveal/mismatch tercatat.
- Proctoring report bisa menampilkan event terkait token.

---

# Recommended implementation order

1. Backend room token validation for APK login.
2. APK login field token ruang.
3. Portal siswa CBT schedule without token reveal.
4. Portal token reveal with room token modal.
5. Print pack wording and checklist update.
6. Audit/event integration to proctoring report.
7. Build, test, commit, deploy.

---

# Commit strategy

Use small commits:

```bash
git commit -m "feat: require room token for CBT exam login"
git commit -m "feat: show CBT schedule in student portal"
git commit -m "feat: add student token reveal gate"
git commit -m "feat: update mobile CBT login room token"
git commit -m "feat: audit CBT token reveal events"
```

Final verification commit if needed:

```bash
git commit -m "test: validate CBT portal token flow"
```
