# Baseline Audit Asesmen CBT Formal SOP

Tanggal: 2026-05-17

## Batas Audit

Audit ini dilakukan dari source repository saja. Tidak ada deploy, restart PM2, akses database production, pembacaan credential, atau penampilan secret. Runtime count production untuk kegiatan, sesi, paket, peserta, insiden, dan audit log tidak diambil karena sesi ini tidak boleh membaca credential dan tidak ada approval eksplisit untuk query production.

## Source Snapshot

- Backend canonical Asesmen route sudah tersedia di `services/core-api/cmd/api/main.go`.
- BFF Asesmen berada di `apps/web-admin/src/routes/api/asesmen/**` dan mayoritas re-export shared proxy dari `apps/web-admin/src/lib/server/cbt-backend-proxy/**`.
- UI utama Asesmen berada di `apps/web-admin/src/routes/asesmen/**`.
- Bank Soal tetap standalone di `apps/web-admin/src/routes/bank-soal/**` dan API `/api/bank-soal/*`.
- Mobile CBT berada di `apps/mobile/lib/src/**`.

## Temuan Baseline

- Detail Kegiatan Asesmen sudah memiliki wizard kesiapan berbasis data existing, tetapi belum memakai model SOP 10 tahap formal.
- Backend sudah memiliki aggregate `CbtEventOverview` dan readiness existing yang cukup untuk membangun SOP readiness tanpa schema change.
- Backend canonical `/api/asesmen/sessions/{id}`, `DELETE /api/asesmen/sessions/{id}`, dan `/api/asesmen/sessions/{id}/participants/{pid}/answers` sudah terdaftar, tetapi BFF alias perlu disejajarkan agar UI tidak drift.
- Pengawasan sesi/ruang sudah memiliki data heartbeat, risiko, lock anti-cheat, tindakan pengawas, handover, dan rekap insiden.
- Berita acara sesi sudah tersedia dan dapat diperkaya dengan ringkasan hadir/kirim/insiden/tanda tangan tanpa schema change.
- Approval/finalisasi formal belum memiliki table/API khusus; perlu migration additive.

## Risiko Operasional

- Perubahan harus additive dan backward-compatible karena CBT runtime dan legacy `/api/cbt/*` masih hidup.
- Pengesahan formal tidak boleh langsung memblokir workflow lama sebelum ada keputusan produksi.
- Arsip ZIP kegiatan belum tersedia di backend; tahap aman pertama adalah checklist arsip dan export dokumen existing.

## Status Production Action

- Deploy: tidak dilakukan.
- Restart PM2: tidak dilakukan.
- Migration production: tidak dijalankan.
- Secret/credential: tidak dibaca dan tidak ditampilkan.
