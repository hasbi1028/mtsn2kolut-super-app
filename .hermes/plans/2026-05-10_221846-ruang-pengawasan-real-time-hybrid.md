# Plan: Ruang Pengawasan Real-time Opsi Hybrid

## Tujuan
Meningkatkan fitur pengawasan CBT agar pengawas wajib membuka ruang pengawasan di web app dan menerima notifikasi kecurangan secara real-time, dengan mode **Hybrid**:

1. **Mode Ruang** untuk pengawas kelas/lab: fokus pada satu ruang yang ditugaskan.
2. **Mode Command Center** untuk admin/operator asesmen: memantau semua ruang dalam satu sesi.
3. **Transport Hybrid**: polling cepat sebagai MVP + SSE real-time sebagai peningkatan utama, dengan fallback polling jika SSE terputus.

## Konteks saat ini
Berdasarkan inspeksi repo:

- Sudah ada halaman daftar ruang pengawas:
  - `apps/web-admin/src/routes/asesmen/pengawasan/+page.svelte`
  - API: `/api/asesmen/proctoring/my-rooms`
- Sudah ada halaman dashboard pengawas per ruang:
  - `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`
  - API snapshot: `/api/asesmen/sessions/:id/rooms/:rid/proctoring`
  - Saat ini refresh otomatis memakai polling `15 detik`.
- Sudah ada command/action pengawas:
  - reset akses peserta
  - force submit
  - flag peserta
  - export evidence CSV
  - handover ruang
- Sudah ada endpoint command center/sesi:
  - `/api/asesmen/sessions/:id/proctoring`
  - `/api/asesmen/sessions/:id/proctoring/events`
- APK sudah mengirim event anti-cheat ke backend dan backend sudah memiliki status:
  - `risk_level`: normal/warning/high/locked
  - `violation_count`
  - `risk_score`
  - `locked_at`
  - event seperti `anti_cheat_violation`, `app_switch`, `screenshot_attempt`

## Prinsip implementasi

- Jangan mengandalkan APK saja. Web pengawas harus menjadi pusat observasi real-time.
- Real-time harus tetap aman jika SSE gagal: polling tetap jalan sebagai fallback.
- Pengawas ruang hanya melihat ruang yang ditugaskan; admin/operator bisa lihat semua ruang.
- Alert harus terlihat jelas, tapi tidak mengganggu aksi pengawas.
- Semua aksi sensitif tetap dicatat sebagai event/audit.
- Token peserta jangan ditampilkan di export evidence atau toast notifikasi.

## Opsi Hybrid yang dipilih

### UX Hybrid

1. **Mode Ruang**
   - URL existing:
     - `/asesmen/sesi/[id]/rooms/[rid]/proctoring`
   - Dipakai pengawas ruang.
   - Fokus satu ruang, satu daftar peserta, satu feed event.

2. **Mode Command Center**
   - URL baru/lanjutan:
     - `/asesmen/sesi/[id]/proctoring`
   - Dipakai admin/operator asesmen.
   - Memantau semua ruang pada sesi tersebut.
   - Menampilkan ringkasan per ruang dan daftar alert lintas ruang.

3. **Entry point**
   - Menu existing:
     - `/asesmen/pengawasan`
   - Detail sesi:
     - tombol `Buka Command Center`
     - tombol per ruang `Buka Ruang Pengawasan`

### Transport Hybrid

1. **Phase 1 MVP cepat**
   - Turunkan polling dashboard ruang dari 15 detik menjadi 3–5 detik saat tab aktif.
   - Tambah dedup toast notifikasi untuk event baru.
   - Tidak perlu SSE dulu agar cepat dipakai.

2. **Phase 2 real-time SSE**
   - Tambah endpoint stream SSE.
   - Web app subscribe event live.
   - Jika SSE putus, fallback polling otomatis.

3. **Phase 3 optimasi**
   - Command Center lintas ruang.
   - Audio alert opsional.
   - Ack/tandai diperiksa.
   - Export evidence per sesi/ruang.

## File dan area yang kemungkinan berubah

### Web Admin — halaman UI

1. `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`
   - Ubah polling 15 detik menjadi mode hybrid:
     - polling aktif 3–5 detik saat SSE belum connected.
     - polling lebih lambat/disabled saat SSE connected.
   - Tambah fungsi `mergeDashboardPayload()` agar update snapshot tidak reset seluruh state yang tidak perlu.
   - Tambah `lastSeenEventId` / `seenEventIds` untuk dedup notifikasi.
   - Tambah toast dan highlight baris peserta saat event baru masuk.
   - Tambah panel `Live Alert` di bagian atas/kanan.

2. `apps/web-admin/src/routes/asesmen/sesi/[id]/proctoring/+page.svelte` *(mungkin sudah ada atau perlu dibuat jika belum lengkap)*
   - Command Center sesi.
   - Menampilkan semua ruang dan semua peserta risk/high/locked.
   - Feed event lintas ruang.
   - Filter:
     - semua
     - warning
     - high
     - locked
     - offline/stale
     - belum submit

3. `apps/web-admin/src/routes/asesmen/pengawasan/+page.svelte`
   - Tambah badge `Live` / `Butuh perhatian` lebih jelas.
   - Tambah CTA langsung:
     - `Buka Ruang Pengawasan`
     - `Command Center` untuk role admin/operator.

4. `apps/web-admin/src/lib/cbt/proctor-evidence.ts`
   - Pastikan event baru dari SSE diklasifikasi sama dengan event polling.
   - Tambah label untuk:
     - `split_screen_detected`
     - `picture_in_picture_detected`
     - `anti_cheat_local_lock`
     - `window_focus_lost`
     - `app_backgrounded`

5. Komponen opsional baru:
   - `apps/web-admin/src/lib/components/cbt/LiveProctorAlertFeed.svelte`
   - `apps/web-admin/src/lib/components/cbt/ProctorParticipantRiskBadge.svelte`
   - `apps/web-admin/src/lib/components/cbt/ProctorConnectionStatus.svelte`

### Web Admin — proxy API

Existing proxy:

- `apps/web-admin/src/routes/api/asesmen/sessions/[id]/rooms/[rid]/proctoring/+server.ts`
- `apps/web-admin/src/routes/api/asesmen/sessions/[id]/proctoring/+server.ts`
- `apps/web-admin/src/routes/api/asesmen/sessions/[id]/proctoring/events/+server.ts`

Tambahan untuk SSE:

1. Route SvelteKit proxy baru:
   - `apps/web-admin/src/routes/api/asesmen/sessions/[id]/rooms/[rid]/proctoring/stream/+server.ts`
   - `apps/web-admin/src/routes/api/asesmen/sessions/[id]/proctoring/stream/+server.ts`

2. Backend path proxy:
   - `/api/cbt/sessions/:id/rooms/:rid/proctoring/stream`
   - `/api/cbt/sessions/:id/proctoring/stream`

Catatan: jika SvelteKit proxy streaming ke Go backend bermasalah, opsi fallback adalah SSE langsung dari SvelteKit dengan polling database/backend setiap 1–2 detik. Tapi rekomendasi utama tetap Go backend sebagai sumber event.

### Core API — handler/service/repository

Kemungkinan file area:

- `services/core-api/internal/handler/*cbt*` atau handler asesmen CBT existing.
- `services/core-api/internal/service/exam.go`
- `services/core-api/internal/repository/postgres/cbt_sessions.sql.go` *(generated; jangan edit manual jika pakai sqlc)*
- `services/core-api/db/queries/cbt_sessions.sql`
- migrations bila perlu index tambahan.

Tambahan endpoint:

1. Snapshot existing tetap:
   - `GET /api/cbt/sessions/:id/rooms/:rid/proctoring`
   - `GET /api/cbt/sessions/:id/proctoring`

2. Events existing tetap:
   - `GET /api/cbt/sessions/:id/proctoring/events?after_id=...&limit=...`
   - jika belum mendukung `after_id`, tambahkan.

3. SSE baru:
   - `GET /api/cbt/sessions/:id/rooms/:rid/proctoring/stream`
   - `GET /api/cbt/sessions/:id/proctoring/stream`

Event SSE format:

```text
event: proctor_event
id: <event_id_or_timestamp>
data: {json}
```

Contoh JSON:

```json
{
  "id": "event_uuid",
  "type": "anti_cheat_violation",
  "session_id": "...",
  "room_id": "...",
  "participant_id": "...",
  "student_id": "...",
  "nis": "...",
  "nama": "Ahmad Fulan",
  "room_name": "Lab 1",
  "risk_level": "warning",
  "violation_count": 1,
  "risk_score": 25,
  "reason": "split_screen_detected",
  "created_at": "2026-05-10T22:20:00+08:00"
}
```

SSE heartbeat:

```text
event: ping
data: {"time":"..."}
```

### Database/index

Jika event query belum optimal, tambahkan migration index:

```sql
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_cbt_participant_events_participant_created
ON cbt_participant_events (participant_id, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_cbt_participant_events_created
ON cbt_participant_events (created_at DESC);
```

Jika query lintas sesi sering dipakai, pertimbangkan join index existing pada `cbt_exam_participants(session_id, room_id)`.

Migration harus non-destruktif. Untuk `CREATE INDEX CONCURRENTLY`, pastikan runner migration mendukung di luar transaction; jika tidak, gunakan index biasa di jam sepi atau buat migration khusus.

## Detail UX yang direncanakan

### Mode Ruang

Bagian atas:

- Status live:
  - `Live connected`
  - `Fallback polling`
  - `Disconnected`
- Ringkasan:
  - peserta total
  - online
  - stale
  - offline
  - submitted
  - warning
  - high
  - locked

Panel alert:

- Alert terbaru di kanan/atas:
  - waktu
  - nama siswa
  - reason
  - risk level
  - tombol `Tandai diperiksa`
  - tombol `Detail`

Tabel peserta:

- Nama/NIS
- Status koneksi
- Progress jawaban
- Risk badge
- Violation count
- Last violation reason
- Last heartbeat
- Aksi:
  - flag/unflag
  - reset access
  - force submit
  - unlock jika locked *(jika endpoint tersedia/ditambahkan)*

Toast:

- Warning:
  - `Ahmad Fulan: split-screen terdeteksi (peringatan 1/3)`
- High:
  - `Ahmad Fulan: pelanggaran berulang, risiko tinggi (2/3)`
- Locked:
  - `Ahmad Fulan terkunci otomatis — perlu verifikasi pengawas`

Audio:

- Default off agar tidak mengganggu.
- Toggle `Bunyi alert` untuk pengawas.
- Bunyi hanya untuk `high` dan `locked`, bukan setiap heartbeat.

### Mode Command Center

Layout:

- Kartu ringkasan sesi.
- Grid ruang:
  - nama ruang
  - online/offline
  - warning/high/locked
  - tombol buka ruang
- Feed alert lintas ruang.
- Filter cepat:
  - semua alert
  - high/locked
  - offline/stale
  - belum submit
- Tabel peserta bermasalah lintas ruang.

Hak akses:

- Admin/operator asesmen: semua ruang.
- Pengawas ruang: redirect/terbatas ke ruang yang ditugaskan.

## Permission dan route access

Periksa/update:

- `apps/web-admin/src/lib/server/route-access.ts`
- `apps/web-admin/src/lib/server/route-access.test.ts`

Rekomendasi permission:

- `asesmen.proctor_read`
  - buka ruang pengawasan dan command center sesuai scope.
- `asesmen.proctor_action`
  - reset akses, force submit, unlock, flag.
- `asesmen.proctor_export`
  - export CSV evidence.

Mapping awal:

- admin: semua.
- operator asesmen: read/action/export.
- guru/pengawas: read/action untuk ruang yang ditugaskan.
- siswa/orang tua: tidak boleh.

## Backend authorization/scope

Pastikan backend tidak hanya mengandalkan UI:

- `GET room proctoring`: user harus admin/operator atau tercatat sebagai proctor ruang.
- `GET session command center`: hanya admin/operator atau permission global.
- action peserta: user harus admin/operator atau proctor ruang terkait.
- SSE stream: sama dengan read proctoring.

Jika scope pengawas ruang saat ini hanya ada di web proxy, pindahkan/duplikasi validasi di core-api agar aman.

## Step-by-step implementasi

### Phase 1 — MVP Polling + Notifikasi

1. Ubah polling ruang dari 15 detik ke 5 detik saat tab aktif.
2. Simpan `seenEventIds` pada halaman room proctoring.
3. Saat payload baru masuk:
   - cari event baru yang belum dilihat,
   - filter event penting: anti-cheat, app switch, screenshot, locked, reset, force submit,
   - munculkan toast,
   - tambahkan ke alert feed,
   - highlight peserta terkait 10–15 detik.
4. Tambah ringkasan warning/high/locked yang eksplisit.
5. Tambah tes unit untuk helper event dedup/format jika dipisah ke util TS.
6. Jalankan:
   - `npm run check`
   - `npm run build`

### Phase 2 — SSE Real-time

1. Tambah query backend event incremental:
   - by `session_id`
   - optional `room_id`
   - `created_at > cursor` atau `id > cursor` jika uuid/time ordering aman.
2. Tambah handler SSE di Go:
   - set header `Content-Type: text/event-stream`
   - `Cache-Control: no-cache`
   - `Connection: keep-alive`
   - kirim ping tiap 20–30 detik.
3. Implement loop:
   - setiap 1–2 detik cek event baru.
   - kirim hanya event baru.
   - stop saat context request cancelled.
4. Tambah proxy SvelteKit stream.
5. Tambah client `EventSource` di halaman ruang:
   - connect onMount.
   - onmessage/proctor_event merge event.
   - onerror: close dan aktifkan fallback polling.
   - reconnect exponential backoff ringan.
6. Tambah status UI `Live connected / fallback polling`.
7. Tes:
   - Go handler/service test untuk filtering event.
   - Web admin check/build.
   - Manual dengan insert event atau APK test.

### Phase 3 — Command Center

1. Buat/rapikan halaman `/asesmen/sesi/[id]/proctoring`.
2. Pakai endpoint session proctoring existing untuk snapshot.
3. Subscribe SSE session-level.
4. Render:
   - summary global,
   - room cards,
   - risk participants,
   - event feed.
5. Tambah tombol dari detail sesi dan daftar pengawasan.
6. Permission test untuk admin/operator vs pengawas ruang.

### Phase 4 — Aksi Pengawas Lanjutan

1. Tambah `unlock participant` jika belum ada:
   - endpoint room-scoped.
   - reset `locked_at`, `locked_reason`, set `risk_level` minimal `warning/high` sesuai kebijakan atau `normal` jika admin reset penuh.
   - insert event `proctor_unlock`.
2. Tambah `acknowledge event` / `tandai diperiksa`:
   - bisa berupa event baru `proctor_acknowledge` dengan `event_id`, `participant_id`, `notes`.
   - tidak wajib migration jika memakai event_data JSON.
3. Tambah catatan pengawas per peserta bila diperlukan.

## Testing dan validasi

### Automated

Web admin:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/apps/web-admin
npm run check
npm run build
```

Core API:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/services/core-api
go test ./internal/service ./internal/handler ./internal/middleware
go build -o bin/api ./cmd/api
```

Jika ada SQL/query:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/services/core-api
# jalankan generator sqlc sesuai script project bila tersedia
```

### Manual test

1. Login web admin sebagai admin/operator.
2. Buka `/asesmen/pengawasan`.
3. Buka ruang pengawasan.
4. Login APK sebagai peserta test.
5. Trigger split-screen/multi-window.
6. Pastikan di web:
   - toast muncul dalam < 3–5 detik pada Phase 1,
   - muncul instant/1–2 detik pada Phase 2 SSE,
   - risk naik warning → high → locked,
   - row peserta highlight,
   - event masuk evidence feed.
7. Uji tab browser hidden/visible:
   - saat hidden jangan boros polling,
   - saat visible langsung refresh.
8. Uji SSE disconnect:
   - matikan backend/core-api sebentar atau block stream,
   - UI berubah ke fallback polling,
   - reconnect saat backend kembali.

## Risiko dan mitigasi

1. **Cloudflared/proxy memutus SSE**
   - Mitigasi: ping SSE tiap 20–30 detik dan fallback polling.

2. **Polling terlalu berat saat banyak ruang**
   - Mitigasi: Phase 1 hanya 5 detik untuk room-level; command center bisa 10 detik; SSE mengurangi beban.

3. **Toast terlalu ramai**
   - Mitigasi: dedup event ID, throttle per peserta, audio hanya high/locked.

4. **Pengawas membuka banyak tab**
   - Mitigasi: tab hidden pause polling; SSE reconnect only visible atau tetap connected tapi no toast saat hidden.

5. **Permission bocor antar ruang**
   - Mitigasi: backend scope check wajib, route-access test, manual test akun pengawas ruang.

6. **Event ordering UUID tidak urut**
   - Mitigasi: cursor pakai `created_at + id`, bukan UUID saja.

## Deployment boundary

Untuk eksekusi nanti:

1. Jangan deploy/restart sebelum semua test lulus.
2. Jika hanya web UI Phase 1:
   - build web-admin.
   - restart PM2 web-admin jika diminta.
3. Jika backend SSE/endpoint berubah:
   - build core-api.
   - restart PM2 core-api jika diminta.
4. Jika migration index ditambahkan:
   - backup DB dulu.
   - jalankan migration di jam aman.
   - verifikasi query.

## Rekomendasi urutan final

Saya rekomendasikan eksekusi bertahap:

1. **Phase 1 sekarang**: polling 5 detik + toast/dedup/highlight di halaman ruang existing.
2. **Phase 2 berikutnya**: SSE room-level + fallback polling.
3. **Phase 3**: Command Center session-level.
4. **Phase 4**: unlock/ack/catatan pengawas lanjutan.

Dengan begitu fitur cepat bisa dipakai, tapi arsitektur tetap siap menjadi real-time penuh.
