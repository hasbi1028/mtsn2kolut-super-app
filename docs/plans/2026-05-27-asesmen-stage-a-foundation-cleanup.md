# Asesmen Stage A Foundation Cleanup Implementation Plan

> **For Hermes:** implement Stage A only. Do not jump ahead to Dokumen & Cetak Center, public portal redesign, or monster-file refactors from later stages.

**Goal:** lock the simplified Asesmen role/route contract, finalize Stage A navigation, and align route-access so guru/pengawas/staf only see or enter the right surfaces.

**Architecture:** keep all backend routes and advanced pages alive, but tighten the visible web-admin surface and its access contract. Stage A is a foundation pass: contract freeze, sidebar/IA cleanup, and route-access enforcement. No destructive API deletion and no deep UI rewrites yet.

**Tech Stack:** SvelteKit web-admin, Vitest sidebar/route-access tests, route-access helpers in `apps/web-admin/src/lib/server/route-access.ts`.

---

## Scope freeze

### In scope for Stage A
- Save the Stage A contract in-repo.
- Keep main Asesmen sidebar to the simplified five-route surface:
  - `/asesmen/ringkas`
  - `/asesmen/persiapan`
  - `/asesmen/pelaksanaan`
  - `/asesmen/ruang-saya`
  - `/asesmen/hasil`
- Keep advanced surfaces out of the sidebar:
  - `/asesmen`
  - `/asesmen/panitia`
  - `/asesmen/pengawasan`
  - `/asesmen/aplikasi-siswa`
  - `/asesmen/aplikasi-siswa/matrix`
  - `/asesmen/aplikasi-siswa/release`
  - `/asesmen/non-tes`
  - `/asesmen/kegiatan*`
  - `/asesmen/paket*`
  - `/asesmen/sesi*`
- Align route-access with the simplified contract.
- Add/update tests proving the contract.

### Explicitly out of scope for Stage A
- Add `Dokumen & Cetak` launcher or new `/asesmen/dokumen` route.
- Rationalize `/asesmen` launcher behavior.
- Rewrite `/ujian` or `/pengawas-ujian` UX.
- Refactor giant Svelte pages.
- Delete legacy `/api/cbt/*` or old advanced routes.

## Final Stage A route contract

### Admin / panitia / operator main surface
- `/asesmen/ringkas`
- `/asesmen/persiapan`
- `/asesmen/pelaksanaan`
- `/asesmen/ruang-saya`
- `/asesmen/hasil`
- `/asesmen/panitia` stays available but not shown in the sidebar.

### Guru / pengawas field surface
- `/asesmen/pelaksanaan`
- `/asesmen/ruang-saya`
- `/asesmen/hasil` only if permission allows.
- `/asesmen/aplikasi-siswa` may stay reachable by direct link later, but it is not a sidebar item in Stage A.

### Advanced admin-only or panitia-only surfaces kept alive but hidden
- `/asesmen/kegiatan*`
- `/asesmen/paket*`
- `/asesmen/sesi*`
- `/asesmen/pengawasan`
- `/asesmen/aplikasi-siswa/matrix`
- `/asesmen/aplikasi-siswa/release`
- `/asesmen/non-tes`

## Permission contract to enforce in Stage A

### Operator/panitia-oriented
- `/asesmen/ringkas` → `asesmen.operator | asesmen.event_manage | asesmen.package_manage`
- `/asesmen/persiapan` → `asesmen.operator | asesmen.event_manage | asesmen.package_manage`
- `/asesmen/panitia` → `asesmen.operator | asesmen.event_manage | asesmen.package_manage`
- `/asesmen/aplikasi-siswa/matrix` → same operator/panitia permission family
- `/asesmen/aplikasi-siswa/release` → same operator/panitia permission family

### Proctor-oriented
- `/asesmen/pelaksanaan` → `asesmen.proctor`
- `/asesmen/ruang-saya` → `asesmen.proctor`
- `/asesmen/pengawasan` → `asesmen.proctor`

### Results
- `/asesmen/hasil` → `asesmen.result_read`

## Files Stage A may modify
- `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts`
- `apps/web-admin/src/lib/components/sidebar/sidebar-config.test.ts`
- `apps/web-admin/src/lib/server/route-access.ts`
- `apps/web-admin/src/lib/server/route-access.test.ts`
- this plan file

## Verification gates

```bash
npm --prefix apps/web-admin run test:unit -- src/lib/components/sidebar/sidebar-config.test.ts src/lib/server/route-access.test.ts
npm --prefix apps/web-admin run check
rm -rf apps/web-admin/build && npm --prefix apps/web-admin run build
git diff --check
npm run security:scan-secrets:staged
npm run security:scan-secrets
```

Protected-route smoke expectations after deploy:

```bash
/login                     # 200
/asesmen/ringkas          # 302 when logged out
/asesmen/persiapan        # 302 when logged out
/asesmen/ruang-saya       # 302 when logged out
/asesmen/panitia          # 302 when logged out
/asesmen/aplikasi-siswa   # 302 when logged out
/asesmen/pengawasan       # 302 when logged out
```

## Done definition for Stage A
- Sidebar contract is stable and tested.
- `ringkas` / `persiapan` no longer fall through to broad `asesmen.read` access.
- Hidden advanced surfaces stay hidden from navigation.
- Route-access matches the intended role surface well enough to start Stage B.
