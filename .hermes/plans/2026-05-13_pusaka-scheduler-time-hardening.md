# PUSAKA Scheduler Time Hardening Implementation Plan

> **For Hermes:** Use systematic-debugging + test-driven-development for implementation. Do not deploy/restart until explicitly requested, unless user says patch sampai selesai.

**Goal:** Fix PUSAKA scheduler so attendance jobs are generated only inside the intended schedule window, resistant to server clock jumps, RTC/NTP corrections, and late scheduler ticks.

**Architecture:** Keep SvelteKit as BFF/UI only. Core scheduling rules live in Go `core-api` scheduler service and sqlc queries. Database timestamps remain source-of-record, but scheduler must compute due windows from local WITA schedule time, not from enqueue time.

**Tech Stack:** Go core-api, PostgreSQL/sqlc, SvelteKit web-admin, PM2, Linux timedatectl/NTP.

---

## Evidence Summary

Observed issue:
- UI/server current time around `15:24–15:29 WITA`.
- Some PUSAKA jobs show `created_at = 16:39 WITA` and `not_before = 16:54/16:57 WITA`.
- DB query showed `jobs_created_in_future = 70`.
- Some rows have impossible ordering: `created_at 16:39` but `updated_at 15:19`, proving system/DB clock moved forward and later corrected backward.
- `timedatectl` warning: `RTC in local TZ: yes`; RTC not aligned with OS local time.

Code root cause:
- `services/core-api/internal/service/scheduler.go` currently sets employee job delay as:
  - `notBefore = now + randomDelay`
- It should be:
  - `notBefore = localDate + run_time + randomDelay`
- It also lacks an expired-window guard, so if scheduler sees an already-past schedule due to clock jump or late tick, it can still enqueue it.

Recommended behavior:
- For each employee schedule, compute:
  - `base = today WITA + schedule.run_time`
  - `latest = base + random_window_minutes + grace`
  - `not_before = base + random_delay`
- If `now > latest`, skip job for that day and mark schedule processed for the day, so it does not enqueue late.
- Add diagnostics so skipped expired windows are visible.

---

## Recommendation

### Recommended patch policy

Use **scheduled-base + expired guard + small grace period**:

```text
base_time   = tanggal WITA hari ini + run_time
random_pick = 0..random_window_minutes
not_before  = base_time + random_pick
expire_at   = base_time + random_window_minutes + grace

if now > expire_at:
  skip enqueue for today
else:
  create queued job with not_before
```

Suggested grace:
- Start with hardcoded `3 minutes` in code as `defaultEmployeeScheduleGrace = 3 * time.Minute`.
- Later can be setting-backed if needed.

Rationale:
- Prevents `14:33 + 28m` becoming `16:39 + 28m`.
- Protects against server clock jump, late scheduler, PM2 restart, NTP correction.
- Safer to skip an expired presensi window than to auto-presensi much later than intended.

### Server time recommendation

Fix host time configuration after code patch:

```bash
sudo timedatectl set-local-rtc 0
sudo timedatectl set-timezone Asia/Makassar
timedatectl
```

Acceptance:
- `RTC in local TZ: no`
- `System clock synchronized: yes`
- `Time zone: Asia/Makassar (WITA, +0800)`

Do this carefully because it changes OS/RTC time handling.

---

## Implementation Plan

### Task 1: Add scheduler time helper functions

**Objective:** Centralize parsing and calculation of schedule window to avoid scattered time logic.

**Files:**
- Modify: `services/core-api/internal/service/scheduler.go`
- Test: `services/core-api/internal/service/scheduler_test.go`

**Implementation notes:**
Add helper functions near scheduler constants:

```go
const defaultEmployeeScheduleGrace = 3 * time.Minute

type employeeScheduleWindow struct {
    Base      time.Time
    Latest    time.Time
    NotBefore time.Time
    Expired   bool
}

func parseScheduleLocalTime(localDate time.Time, runTime string, loc *time.Location) (time.Time, error) {
    parsed, err := time.ParseInLocation("15:04", runTime, loc)
    if err != nil {
        return time.Time{}, err
    }
    return time.Date(localDate.Year(), localDate.Month(), localDate.Day(), parsed.Hour(), parsed.Minute(), 0, 0, loc), nil
}

func buildEmployeeScheduleWindow(localNow time.Time, runTime string, randomWindowMinutes int32, randomDelayMinutes int32, grace time.Duration, loc *time.Location) (employeeScheduleWindow, error) {
    base, err := parseScheduleLocalTime(localNow, runTime, loc)
    if err != nil {
        return employeeScheduleWindow{}, err
    }
    if randomWindowMinutes < 0 {
        randomWindowMinutes = 0
    }
    if randomDelayMinutes < 0 {
        randomDelayMinutes = 0
    }
    if randomDelayMinutes > randomWindowMinutes {
        randomDelayMinutes = randomWindowMinutes
    }
    latest := base.Add(time.Duration(randomWindowMinutes) * time.Minute).Add(grace)
    return employeeScheduleWindow{
        Base:      base,
        Latest:    latest,
        NotBefore: base.Add(time.Duration(randomDelayMinutes) * time.Minute),
        Expired:   localNow.After(latest),
    }, nil
}
```

**Tests:**
- `14:33 + 28m + 3m`, `now=14:50` => not expired.
- `14:33 + 28m + 3m`, `now=16:39` => expired.
- `not_before` uses base schedule time, not `now`.

**Command:**

```bash
cd services/core-api
go test ./internal/service -run 'TestSchedulerEmployeeScheduleWindow' -count=1
```

---

### Task 2: Change employee scheduler enqueue logic

**Objective:** Use scheduled-base `not_before` and skip expired windows.

**Files:**
- Modify: `services/core-api/internal/service/scheduler.go`
- Test: `services/core-api/internal/service/scheduler_test.go`

**Current code to replace:**

```go
notBefore := pgtype.Timestamptz{}
if es.RandomWindowMinutes > 0 {
    delayMinutes := rand.Int31n(int32(es.RandomWindowMinutes) + 1)
    _ = notBefore.Scan(now.Add(time.Duration(delayMinutes) * time.Minute))
}
_, createErr := s.jobs.CreateWithDelay(ctx, es.EmployeeID, string(es.RunType), maxAttempts, notBefore)
```

**New behavior:**

```go
var delayMinutes int32
if es.RandomWindowMinutes > 0 {
    delayMinutes = rand.Int31n(int32(es.RandomWindowMinutes) + 1)
}
window, windowErr := buildEmployeeScheduleWindow(localNow, es.RunTime, int32(es.RandomWindowMinutes), delayMinutes, defaultEmployeeScheduleGrace, s.loc)
if windowErr != nil {
    _ = s.store.ResetEmployeeScheduleEnqueueState(ctx, db.ResetEmployeeScheduleEnqueueStateParams{ID: es.ID, LastEnqueuedForDate: today})
    slog.Error("scheduler: invalid employee schedule time", "employee_id", es.EmployeeID, "run_type", es.RunType, "run_time", es.RunTime, "error", windowErr)
    continue
}
if window.Expired {
    result.Skipped++
    slog.Warn("scheduler: skipped expired employee schedule window", "employee_id", es.EmployeeID, "run_type", es.RunType, "run_time", es.RunTime, "base", window.Base.Format(time.RFC3339), "latest", window.Latest.Format(time.RFC3339), "now", localNow.Format(time.RFC3339))
    continue
}
notBefore := pgtype.Timestamptz{}
if es.RandomWindowMinutes > 0 || window.NotBefore.After(localNow) {
    _ = notBefore.Scan(window.NotBefore)
}
_, createErr := s.jobs.CreateWithDelay(ctx, es.EmployeeID, string(es.RunType), maxAttempts, notBefore)
```

**Important:** For zero random window, decide behavior:
- If schedule due and not expired, immediate job is okay.
- `not_before` may remain invalid for immediate jobs.

**Tests:**
- Existing `TestSchedulerTickEmployeeSchedulesRandomWindowSetsNotBefore` must change expectation from `now..now+15m` to `base..base+15m`.
- Add regression: when `now=16:39`, `run_time=14:33`, `window=28`, job is not created and `Skipped` increments.
- Add regression: when `now=14:40`, `run_time=14:33`, `window=28`, job created with `not_before <= 15:01`.

**Command:**

```bash
cd services/core-api
go test ./internal/service -run 'TestSchedulerTickEmployeeSchedules' -count=1
```

---

### Task 3: Add diagnostic API/UI fields for due time vs created time

**Objective:** Make future debugging easier by showing planned execution time (`not_before`) in the PUSAKA queue UI.

**Files:**
- Modify: `services/core-api/db/queries/jobs.sql`
- Generated: `services/core-api/internal/repository/postgres/jobs.sql.go`
- Modify: `apps/web-admin/src/routes/api/pusaka/jobs/+server.ts`
- Modify: `apps/web-admin/src/routes/pusaka/antrian/+page.svelte`
- Test: `apps/web-admin/src/lib/server/proxy-routes.test.ts` if existing coverage applies.

**Backend query:**
Currently `ListJobs` returns `next_retry_at, created_at, updated_at` but not `not_before` in list rows. Include:

```sql
j.next_retry_at, j.created_at, j.updated_at, j.not_before
```

Update generated sqlc via:

```bash
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
```

**BFF/UI:**
- Add `not_before` to job payload.
- Add UI column/card label:
  - `Dibuat`
  - `Mulai diproses setelah` / `Not before`
  - `Diperbarui`
- This distinguishes “created time” from scheduled execution time.

**Acceptance:**
- Queue UI no longer hides key timing fields.
- For scheduled-random jobs, operator can see why job is not yet claimed.

---

### Task 4: Add clock health diagnostics endpoint/card

**Objective:** Surface OS/DB clock anomalies before they break scheduler.

**Files:**
- Modify or create handler/service under core-api health/admin area if existing pattern available.
- Candidate existing files:
  - `services/core-api/internal/handler/health.go`
  - `services/core-api/internal/service/health.go` if present
  - web-admin PUSAKA dashboard/settings page.

**Minimum viable check:**
- DB: `SELECT now()`
- API server: `time.Now()`
- Difference in seconds.
- Add status warning if abs(diff) > 5 seconds.

**Optional shell-only check:**
- Document `timedatectl` manual check in admin SOP, because Go API may not safely run shell commands.

**Acceptance:**
- Admin can detect DB/API clock drift.
- PUSAKA page warns if drift is dangerous.

---

### Task 5: Clean up current bad future jobs safely

**Objective:** Remove or mark failed any stale/future PUSAKA jobs created by clock jump so worker does not run invalid late presensi.

**Requires explicit approval before production SQL.**

**Pre-check SQL:**

```sql
SELECT id, run_type, status, created_at, not_before, updated_at, error_message
FROM jobs
WHERE created_at > NOW()
   OR not_before > NOW() + INTERVAL '10 minutes'
ORDER BY created_at DESC;
```

**Safe cleanup option:**
Only for active queued/running jobs that are invalid future jobs:

```sql
UPDATE jobs
SET status = 'failed',
    error_message = 'Dibatalkan otomatis: timestamp job tidak valid akibat koreksi jam server',
    next_retry_at = NULL,
    updated_at = NOW()
WHERE status IN ('queued', 'running')
  AND (created_at > NOW() OR not_before > NOW() + INTERVAL '10 minutes');
```

**Do not delete history** unless explicitly requested. Mark failed is auditable.

---

### Task 6: Fix OS RTC/NTP configuration

**Objective:** Prevent system time from jumping due to RTC local timezone mismatch.

**Command:**

```bash
timedatectl
sudo timedatectl set-local-rtc 0
sudo timedatectl set-timezone Asia/Makassar
timedatectl
```

**Acceptance:**

```text
System clock synchronized: yes
NTP service: active
RTC in local TZ: no
Time zone: Asia/Makassar (WITA, +0800)
```

**Caution:**
- This is OS-level state change. Do after code patch or during maintenance window.
- Verify DB `now()` and OS `date` align after change.

---

### Task 7: Full validation

**Objective:** Ensure scheduler patch is safe and web-admin builds.

**Commands:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build

cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-pusaka-time-hardening ./cmd/api
```

**Expected:** all pass.

---

### Task 8: Commit before deploy

**Objective:** Keep a clean, auditable patch.

**Commit message:**

```bash
git add services/core-api apps/web-admin .hermes/plans/2026-05-13_pusaka-scheduler-time-hardening.md
git commit -m "fix(pusaka): harden scheduler time windows"
```

Do not include unrelated untracked seed scripts.

---

### Task 9: Deploy sequence only after approval

**Objective:** Apply safely to production.

**Sequence:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/services/core-api
cp -p bin/api bin/api.backup-pusaka-time-hardening-$(date +%Y%m%d-%H%M%S)
go build -o bin/api ./cmd/api
pm2 restart mtsn2kolut-core-api --update-env
curl -fsS http://127.0.0.1:8080/health

cd /home/servermtsn2kolut/mtsn2kolut-super-app
npm --prefix apps/web-admin run build
pm2 restart mtsn2kolut-web-admin --update-env
```

Then perform smoke tests:
- `/settings`
- `/pusaka/antrian`
- trigger scheduler tick manually only if safe.

---

## Acceptance Criteria

1. For schedule `14:33`, random `28`, scheduler never creates `not_before` later than `15:04` with 3-minute grace.
2. If current time is `16:39`, the same schedule is skipped, not enqueued.
3. Queue UI shows both created time and planned processing time.
4. Future `created_at > now()` active jobs are cleaned/failed only after approval.
5. OS time warning `RTC in local TZ` is resolved after maintenance command.
6. All checks pass:
   - `npm --prefix apps/web-admin run check`
   - `npm --prefix apps/web-admin run build`
   - `sqlc generate`
   - `go test ./internal/handler ./internal/service ./internal/repository/postgres`
   - `go build`

---

## Implementation Notes — 2026-05-13

- Scheduler now computes `not_before` from WITA schedule base time (`today + run_time + random_delay`), not from enqueue wall-clock time.
- Expired schedule windows are skipped with warning logs instead of enqueuing late jobs.
- Regression tests cover normal window calculation, late `16:39` skip behavior, and random-window `not_before` range from schedule base.
- PUSAKA job list query/API/UI now exposes `not_before` as "Mulai Setelah" in table/card views.
- `/health` now includes API/DB clock diagnostics: server local/UTC time, DB `now()`, DB timezone, and DB-server drift in ms.
- Production cleanup requested by user: backed up and deleted 70 jobs with `created_at` in `2026-05-13 16:39 WITA`; remaining matching jobs and future jobs are 0.
- RTC fix command was attempted, but host rejected it because interactive sudo authentication is required. Current software patch mitigates scheduler impact; OS-level acceptance still requires an operator to run `sudo timedatectl set-local-rtc 0`.
- Validation passed: web-admin check/build, sqlc generate, Go handler/service/repository tests, and backend build.

## Rollback Plan

If backend deploy causes issue:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/services/core-api
cp -p bin/api.backup-pusaka-time-hardening-YYYYMMDD-HHMMSS bin/api
pm2 restart mtsn2kolut-core-api --update-env
curl -fsS http://127.0.0.1:8080/health
```

If web-admin deploy causes issue:
- Revert commit or restore previous build artifact if available.
- Restart `mtsn2kolut-web-admin`.

Do not roll back RTC fix unless there is a host-level reason; UTC RTC is the recommended Linux configuration.
