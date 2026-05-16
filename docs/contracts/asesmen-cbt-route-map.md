# Peta Kontrak Route Asesmen CBT

Tanggal baseline: 2026-05-17

Dokumen ini memetakan route backend Go, BFF SvelteKit, UI, dan legacy compatibility. Semua alias publik sudah diputuskan eksplisit dalam baseline ini.

## Ringkasan Keputusan

| Area | Keputusan |
| --- | --- |
| `/api/asesmen/*` | Kontrak BFF/UI utama untuk modul Asesmen CBT. |
| `/api/bank-soal/*` | Kontrak BFF/UI utama untuk Bank Soal. Bank Soal tetap standalone. |
| `/api/cbt/*` | Legacy backend compatibility tetap hidup, tetapi client web-admin baru memakai `/api/asesmen/*` atau `/api/bank-soal/*`. |
| Stream pengawasan | BFF-only custom stream karena SvelteKit menjaga koneksi browser dan meneruskan data dari backend. |
| Session detail/delete/answers | Public UI/BFF alias ditambahkan agar tidak drift dari backend canonical Asesmen. |

## Backend dan BFF Asesmen

| Method | Route | Status | Catatan |
| --- | --- | --- | --- |
| GET | `/api/asesmen/readiness` | public-ui-used | Hub readiness kegiatan. |
| GET | `/api/asesmen/events` | public-ui-used | Daftar Kegiatan Asesmen. |
| POST | `/api/asesmen/events` | public-ui-used | Admin membuat kegiatan. |
| GET | `/api/asesmen/events/{id}` | public-ui-used | Detail kegiatan. |
| PUT | `/api/asesmen/events/{id}` | public-ui-used | Update identitas kegiatan. |
| PATCH | `/api/asesmen/events/{id}/status` | public-ui-used | Update status kegiatan. |
| DELETE | `/api/asesmen/events/{id}` | public-ui-used | Hapus kegiatan sesuai guard backend. |
| GET | `/api/asesmen/events/{id}/overview` | public-ui-used | Ringkasan kesiapan. |
| GET | `/api/asesmen/events/{id}/readiness` | public-ui-used | Legacy readiness shape dengan selected event. |
| GET | `/api/asesmen/events/{id}/sop-readiness` | public-ui-used | Readiness SOP 10 tahap. |
| GET | `/api/asesmen/events/{id}/question-completeness` | public-ui-used | Kelengkapan soal per target. |
| GET/PUT | `/api/asesmen/events/{id}/question-requirements` | public-ui-used | Target kebutuhan soal. |
| GET | `/api/asesmen/events/{id}/packages` | public-ui-used | Paket terkait kegiatan. |
| GET | `/api/asesmen/events/{id}/sessions` | public-ui-used | Sesi terkait kegiatan. |
| GET/POST | `/api/asesmen/events/{id}/members` | public-ui-used | Tim/panitia kegiatan. |
| PUT/DELETE | `/api/asesmen/events/{id}/members/{member_id}` | public-ui-used | Mutasi anggota kegiatan. |
| GET/PUT | `/api/asesmen/events/{id}/question-targets` | public-ui-used | Target mapel kegiatan. |
| DELETE | `/api/asesmen/events/{id}/question-targets/{subject_id}` | public-ui-used | Hapus target mapel. |
| GET | `/api/asesmen/events/{id}/results` | public-ui-used | Rekap hasil kegiatan. |
| GET | `/api/asesmen/events/{id}/exam-cards` | public-ui-used | Kartu ujian kegiatan. |
| GET | `/api/asesmen/packages` | public-ui-used | Daftar Paket Soal. |
| POST | `/api/asesmen/packages` | public-ui-used | Buat paket. |
| GET | `/api/asesmen/packages/readiness` | public-ui-used | Readiness paket. |
| GET/PUT/DELETE | `/api/asesmen/packages/{id}` | public-ui-used | Detail/update/hapus paket. |
| PUT | `/api/asesmen/packages/{id}/questions` | public-ui-used | Susun soal paket. |
| POST | `/api/asesmen/packages/{id}/clone` | public-ui-used | Duplikasi paket. |
| POST | `/api/asesmen/packages/{id}/lock` | public-ui-used | Kunci paket. |
| GET | `/api/asesmen/sessions` | public-ui-used | Daftar Sesi Ujian. |
| POST | `/api/asesmen/sessions` | public-ui-used | Buat sesi. |
| GET | `/api/asesmen/sessions/{id}` | public-ui-used | Alias BFF ditambahkan. |
| DELETE | `/api/asesmen/sessions/{id}` | public-ui-used | Alias BFF ditambahkan; backend guard tetap sumber kebenaran. |
| GET | `/api/asesmen/sessions/{id}/participants` | public-ui-used | Daftar peserta. |
| POST | `/api/asesmen/sessions/{id}/enroll` | public-ui-used | Enroll peserta. |
| POST | `/api/asesmen/sessions/{id}/generate-tokens` | public-ui-used | Generate token; token raw tetap dikendalikan backend. |
| GET | `/api/asesmen/sessions/{id}/participants/{pid}/answers` | public-ui-used | Alias BFF ditambahkan untuk admin/guru sesuai guard backend. |
| GET | `/api/asesmen/sessions/{id}/rooms` | public-ui-used | Daftar ruang. |
| GET | `/api/asesmen/sessions/{id}/rooms/readiness` | public-ui-used | Kesiapan ruang. |
| GET | `/api/asesmen/sessions/{id}/minutes` | public-ui-used | Berita acara sesi. |
| GET | `/api/asesmen/sessions/{id}/operational-recap` | public-ui-used | Rekap operasional. |
| GET | `/api/asesmen/sessions/{id}/results` | public-ui-used | Hasil sesi. |
| GET | `/api/asesmen/sessions/{id}/item-analysis` | public-ui-used | Analisis butir. |
| GET | `/api/asesmen/sessions/{id}/proctoring` | public-ui-used | Pengawasan sesi. |
| GET | `/api/asesmen/sessions/{id}/proctoring/events` | public-ui-used | Riwayat pengawasan. |
| GET | `/api/asesmen/sessions/{id}/proctoring/stream` | bff-only-custom | Stream browser dari BFF. |
| GET | `/api/asesmen/proctoring/my-rooms` | public-ui-used | Ruang pengawas aktif. |
| GET | `/api/asesmen/sessions/{id}/rooms/{rid}/proctoring` | public-ui-used | Panel ruang. |
| GET | `/api/asesmen/sessions/{id}/rooms/{rid}/proctoring/stream` | bff-only-custom | Stream browser dari BFF. |
| GET/PUT/POST | `/api/asesmen/sessions/{id}/rooms/{rid}/handover*` | public-ui-used | Serah terima ruang. |
| POST | `/api/asesmen/sessions/{id}/rooms/{rid}/participants/{pid}/*` | public-ui-used | Tindakan pengawas. |
| GET | `/api/asesmen/approvals` | public-ui-used | Pengesahan formal additive. |
| POST | `/api/asesmen/approvals` | public-ui-used | Tambah/ulang pengesahan formal. |
| POST | `/api/asesmen/approvals/{id}/revoke` | public-ui-used | Cabut pengesahan formal. |

## Bank Soal

Route Bank Soal aktif memakai `/api/bank-soal/questions/*`, `/api/bank-soal/assets/*`, dan `/api/bank-soal/soal-support/*`. Route lama `/api/cbt/questions/*` tetap legacy compatibility untuk backend/BFF lama, bukan target client baru.

## Legacy Compatibility

| Prefix | Status | Catatan |
| --- | --- | --- |
| `/api/cbt/events/*` | legacy-compat | Dipertahankan di backend untuk klien lama. |
| `/api/cbt/packages/*` | legacy-compat | Dipertahankan di backend. |
| `/api/cbt/sessions/*` | legacy-compat | Dipertahankan di backend. |
| `/api/cbt/questions/*` | legacy-compat | Deprecated; UI baru memakai `/api/bank-soal/questions/*`. |

## Catatan Keamanan

- Authenticated BFF meneruskan JWT user ke backend.
- `X-Internal-Key` tidak menjadi fallback diam-diam untuk route Asesmen yang sudah login.
- Route yang menampilkan token tetap mengikuti masking/redaksi backend.
- Route finalisasi/approval tidak membawa token raw, answer key, password, API key, atau connection string.
