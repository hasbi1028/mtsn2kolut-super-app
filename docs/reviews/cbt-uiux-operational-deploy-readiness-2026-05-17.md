# CBT UI/UX Operational Sprints — Deploy Readiness

Tanggal: 2026-05-17
Scope: `.hermes/plans/2026-05-17_cbt-uiux-operational-sprints.md`

## Ringkasan

Sprint UX CBT 0–5 diselesaikan secara additive dan backward-compatible:

- Simple Proctor Mode untuk pengawas ruang.
- Terminology cleanup CBT sekolah/madrasah.
- Arsip Kegiatan Asesmen dan panel pengesahan SOP visible.
- Token peserta dimasking pada Berita Acara/arsip final; token lengkap tetap tersedia di kartu ujian untuk distribusi terbatas.
- Mobile CBT copy diselaraskan agar tidak membuat peserta panik.
- Readiness deploy didokumentasikan.

## Production Safety

- SvelteKit tetap BFF/proxy ke Go core-api.
- Tidak ada akses DB langsung dari frontend.
- Tidak ada credential/secret yang disimpan atau ditampilkan dalam perubahan ini.
- Perubahan UI bersifat additive; route lama tetap hidup.
- Approval formal memakai API approval existing dari sprint sebelumnya.

## Deploy Sequence

Karena commit sebelumnya memiliki migration `113_cbt_approval_records.sql`, urutan production harus:

1. Jalankan migration database melalui mekanisme migration resmi repo.
2. Build core-api.
3. Restart PM2 `mtsn2kolut-core-api`.
4. Build web-admin.
5. Restart PM2 `mtsn2kolut-web-admin` segera setelah build agar tidak terjadi stale chunk.
6. Health check backend.
7. Smoke test UI:
   - `/asesmen`
   - `/asesmen/kegiatan`
   - detail kegiatan asesmen
   - archive kegiatan `/asesmen/kegiatan/{id}/archive`
   - panel pengawasan ruang
   - BA sesi

## Rollback Notes

- UI rollback: checkout/revert commit UX terbaru lalu rebuild/restart web-admin.
- Backend rollback jika migration sudah diterapkan harus mengikuti SOP DB; migration additive sehingga normalnya aman dibiarkan.
- Jangan rollback database tanpa backup/approval eksplisit.
