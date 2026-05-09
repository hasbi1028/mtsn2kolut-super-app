# Katalog Event Internal Analytics

Status: Tahap/Fase 0, kontrak allowlist. Dokumen ini menentukan event yang boleh dipakai oleh analytics internal MTsN 2 Kolaka Utara pada fase berikutnya. Tidak ada event runtime yang dibuat pada fase ini.

## Aturan Global

- **event allowlist**: hanya event yang tertulis di dokumen ini yang boleh dikirim.
- **privacy-first**: metadata harus seminimal mungkin, disanitasi, dan bernilai operasional.
- **internal-only, no third-party analytics**: event tidak boleh dikirim ke layanan analytics pihak ketiga.
- Metadata per event group adalah allowlist. Field di luar daftar harus ditolak atau dibuang oleh Core API masa depan.
- Event name memakai format `group.action`, huruf kecil, snake_case setelah titik bila perlu.
- Event harus membawa waktu server dari Core API, bukan hanya waktu browser.
- Event harus membawa hasil aman seperti `success`, `failed`, `blocked`, `forbidden`, `not_found`, atau `validation_error` bila relevan.

## Forbidden Sensitive Keys

Kontrak ini secara eksplisit melarang penyimpanan dan pengiriman field berikut dalam event analytics atau metadata.

**forbidden sensitive keys**:

- `password`, `passphrase`, `pin`, `otp`, `secret`, `api_key`, `apikey`, `client_secret`
- `token`, `access_token`, `refresh_token`, `exam_token`, `csrf_token`, `session_token`
- `cookie`, `set_cookie`, `authorization`, `auth_header`, `bearer`, `jwt`
- `nik`, `nip_full`, `nip`, `nisn_full`, `nisn`
- `device_fingerprint`, `fingerprint`, `device_id_hash_from_fingerprint`
- `pusaka_username`, `pusaka_password`, `pusaka_credential`, `credential_pusaka`
- `raw_ip`, `ip_address_raw`, `remote_addr`, `x_forwarded_for_raw`
- `raw_user_agent`, `user_agent_raw`, `ua_raw`
- `sql`, `stack_trace`, `request_body`, `response_body`, `raw_payload`

Larangan makna, bukan hanya nama field:

- Password, token, cookie, authorization header, secret, API key.
- NIK penuh, NIP penuh, NISN penuh.
- Full device fingerprint atau fingerprint yang bisa dipakai sebagai bukti identitas perangkat kuat.
- Kredensial PUSAKA dalam bentuk apa pun.
- Raw IP dan raw user agent. Jika perlu analisis teknis, gunakan bucket aman seperti `ip_scope` atau `user_agent_family` setelah disanitasi di Core API.
- Isi jawaban siswa, kunci jawaban, token ujian, isi soal lengkap, isi surat, file upload mentah, dan dokumen sensitif.

## Metadata Aman Umum

Field berikut boleh dipakai lintas group bila relevan:

- `event_name`
- `route_group`
- `module`
- `role`
- `permission_code`
- `result`
- `status_code_class`
- `duration_bucket`
- `source_surface`
- `feature_key`
- `date_bucket`
- `content_category`
- `file_type`
- `count_bucket`
- `error_category`

Field ini tetap harus dikirim sebagai kategori atau ID internal yang aman, bukan nilai mentah sensitif.

## Public Website

Tujuan: mengukur akses dan minat publik terhadap website sekolah tanpa tracker pihak ketiga dan tanpa identitas sensitif.

Event allowlist:

| Event | Kapan dikirim |
|-------|---------------|
| `public.page_view` | Halaman publik seperti `/`, `/profil`, `/berita`, `/pengumuman`, `/ppdb`, dan `/kontak` selesai dimuat. |
| `public.cta_click` | Pengunjung menekan CTA publik seperti PPDB, kontak, pengumuman, atau tautan profil. |
| `public.download` | Pengunjung mengunduh file publik yang memang boleh diakses umum. |
| `public.search` | Pengunjung memakai pencarian publik bila fitur pencarian tersedia. |
| `public.form_start` | Pengunjung mulai formulir publik yang disetujui. |
| `public.form_submit` | Pengunjung mengirim formulir publik yang disetujui, hanya status dan kategori. |

Allowed metadata:

- `route_group`
- `page_type`
- `content_category`
- `content_id`
- `cta_key`
- `file_type`
- `file_category`
- `result`
- `status_code_class`
- `duration_bucket`
- `referrer_scope` seperti `internal`, `search`, `social`, `direct`, atau `other`
- `device_class` seperti `mobile`, `tablet`, atau `desktop`

Forbidden:

- raw IP, raw UA, full URL query, nama lengkap pengunjung, nomor telepon, email, NIK/NIP/NISN penuh, isi formulir, cookie, token, dan identifier perangkat.

## Auth

Tujuan: memahami kesehatan alur autentikasi tanpa menyimpan rahasia.

Event allowlist:

| Event | Kapan dikirim |
|-------|---------------|
| `auth.login_attempt` | Login dicoba, sukses/gagal/blocked. |
| `auth.login_success` | Login berhasil. |
| `auth.login_failed` | Login gagal dengan kategori aman. |
| `auth.refresh_success` | Refresh session berhasil. |
| `auth.logout` | Logout berhasil diproses. |
| `auth.logout_all` | Logout semua session diproses. |
| `auth.session_revoke` | User mencabut satu session. |
| `auth.password_change` | Password diganti, hanya status. |
| `auth.suspended_block` | Login atau akses ditolak karena akun suspended. |

Allowed metadata:

- `role`
- `result`
- `error_category`
- `status_code_class`
- `auth_flow` seperti `password`, `refresh`, `logout`, atau `session_revoke`
- `session_age_bucket`
- `device_label_present`
- `ip_scope` seperti `trusted_proxy`, `private`, `public_bucket`, atau `unknown`
- `user_agent_family`

Forbidden:

- password, OTP, token, cookie, authorization header, JWT, session token, raw IP, raw UA, device fingerprint, username bila tidak disetujui sebagai ID internal aman, dan pesan error internal mentah.

## Dashboard

Tujuan: melihat pemakaian ringkasan dashboard tanpa merekam isi data detail.

Event allowlist:

| Event | Kapan dikirim |
|-------|---------------|
| `dashboard.view` | Dashboard role-aware dibuka. |
| `dashboard.widget_view` | Widget utama dirender atau dimuat ulang. |
| `dashboard.filter_change` | Filter dashboard diganti. |
| `dashboard.export` | Export ringkasan dashboard dilakukan. |
| `dashboard.refresh` | Refresh manual atau background selesai. |

Allowed metadata:

- `role`
- `permission_code`
- `widget_key`
- `filter_key`
- `date_range_bucket`
- `result`
- `status_code_class`
- `duration_bucket`
- `export_type`

Forbidden:

- data siswa/guru mentah, daftar nama, NIK/NIP/NISN penuh, query mentah, isi response API, raw error, token, cookie, dan authorization header.

## Bank Soal

Tujuan: memahami kesehatan workflow Bank Soal dan readiness authoring tanpa menyimpan isi soal.

Event allowlist:

| Event | Kapan dikirim |
|-------|---------------|
| `bank_soal.list_view` | Daftar soal dibuka atau filter berubah. |
| `bank_soal.editor_open` | Editor tambah/edit soal dibuka. |
| `bank_soal.draft_save` | Draft lokal/API tersimpan, hanya status. |
| `bank_soal.question_create` | Soal dibuat. |
| `bank_soal.question_update` | Soal diperbarui. |
| `bank_soal.import_start` | Impor dimulai. |
| `bank_soal.import_complete` | Impor selesai dengan ringkasan hasil. |
| `bank_soal.review_decision` | Reviewer memutuskan publish/revisi/tolak. |
| `bank_soal.readiness_check` | Readiness score dihitung atau direfresh. |
| `bank_soal.asset_upload` | Aset soal diunggah, hanya tipe/hasil. |

Allowed metadata:

- `role`
- `subject_id`
- `grade_level`
- `question_type`
- `authoring_mode` seperti `beginner` atau `advance`
- `workflow_status`
- `review_decision`
- `readiness_bucket`
- `import_result_bucket`
- `asset_type`
- `file_type`
- `result`
- `status_code_class`
- `duration_bucket`

Forbidden:

- stem soal, stimulus, kunci jawaban, rubrik penuh, jawaban siswa, file asset mentah, URL asset dengan token, NIK/NIP/NISN penuh, token/cookie/authorization, raw request body, dan raw error.

## Asesmen

Tujuan: memantau workflow asesmen tanpa membuka jawaban, token ujian, atau data peserta sensitif.

Event allowlist:

| Event | Kapan dikirim |
|-------|---------------|
| `asesmen.hub_view` | Hub asesmen dibuka. |
| `asesmen.package_create` | Paket dibuat. |
| `asesmen.package_update` | Paket diperbarui. |
| `asesmen.event_create` | Kegiatan/event dibuat. |
| `asesmen.session_create` | Sesi dibuat. |
| `asesmen.session_update` | Sesi diperbarui. |
| `asesmen.proctoring_view` | Panel pengawasan dibuka. |
| `asesmen.participant_action` | Aksi peserta seperti assign room, reset status aman, atau validasi status. |
| `asesmen.result_view` | Halaman hasil dibuka. |
| `asesmen.result_export` | Export hasil agregat dilakukan. |
| `asesmen.non_test_sync` | Sinkronisasi penilaian non-tes ke rapor dijalankan. |

Allowed metadata:

- `role`
- `route_group`
- `assessment_kind` seperti `cbt`, `non_test`, atau `mixed`
- `subject_id`
- `grade_level`
- `scope_type`
- `session_status`
- `participant_count_bucket`
- `room_count_bucket`
- `action_key`
- `result`
- `status_code_class`
- `duration_bucket`
- `export_type`

Forbidden:

- exam token, answer key, jawaban siswa, essay text, nilai individu mentah di analytics, NISN penuh, device fingerprint, raw telemetry payload, raw IP, raw UA, token/cookie/authorization, dan data konflik perangkat mentah.

## PUSAKA

Tujuan: memantau kesehatan automation PUSAKA sebagai subsystem terbatas tanpa menyimpan kredensial.

Event allowlist:

| Event | Kapan dikirim |
|-------|---------------|
| `pusaka.dashboard_view` | Dashboard PUSAKA dibuka. |
| `pusaka.job_claimed` | Job diklaim worker, hanya ringkasan. |
| `pusaka.job_completed` | Job selesai. |
| `pusaka.job_failed` | Job gagal dengan kategori aman. |
| `pusaka.job_retried` | Job dijadwalkan ulang. |
| `pusaka.stale_recovered` | Backend memulihkan job running yang stale. |
| `pusaka.employee_scope_update` | Eligibilitas/akun PUSAKA pegawai diubah. |
| `pusaka.settings_update` | Pengaturan PUSAKA diubah. |
| `pusaka.worker_heartbeat` | Heartbeat worker diterima sebagai agregat. |

Allowed metadata:

- `role`
- `job_type`
- `job_status`
- `retry_count_bucket`
- `worker_status`
- `consumer_count_bucket`
- `stale_age_bucket`
- `employee_status_type` seperti `PNS`, `PPPK`, atau `other_bucket`
- `result`
- `status_code_class`
- `duration_bucket`
- `error_category`

Forbidden:

- username PUSAKA, password PUSAKA, credential PUSAKA, screenshot path publik, log mentah dengan rahasia, NIP penuh, cookie/session PUSAKA, raw browser storage, raw IP, raw UA, token, API key, dan detail halaman PUSAKA yang berisi identitas sensitif.

## Users Dan RBAC

Tujuan: memahami aktivitas administrasi user dan permission tanpa menggandakan audit resmi.

Event allowlist:

| Event | Kapan dikirim |
|-------|---------------|
| `users.list_view` | Halaman user dibuka. |
| `users.create` | User dibuat, hanya status dan role target. |
| `users.update` | User diperbarui, hanya kategori perubahan. |
| `users.reset_password` | Reset password diproses, tanpa password. |
| `users.suspend_change` | Status suspended berubah. |
| `rbac.roles_view` | Halaman role/RBAC dibuka. |
| `rbac.permission_update` | Permission role diperbarui, ringkasan aman. |
| `rbac.permission_denied` | Akses ditolak oleh permission guard. |

Allowed metadata:

- `role`
- `target_role`
- `change_category`
- `permission_code`
- `permission_module`
- `result`
- `status_code_class`
- `duration_bucket`
- `denied_route_group`

Forbidden:

- password baru/sementara, hash password, token reset, NIK/NIP/NISN penuh, nama lengkap target bila tidak perlu, email/telepon target bila tidak diagregasi, raw permission payload berisi daftar besar tidak disanitasi, cookie, authorization, dan JWT.

## Security

Tujuan: memberi sinyal keamanan operasional yang sudah disanitasi untuk admin yang punya `analytics.security_read`.

Event allowlist:

| Event | Kapan dikirim |
|-------|---------------|
| `security.rate_limited` | Request kena rate limit. |
| `security.forbidden` | Guard menolak akses. |
| `security.validation_rejected` | Payload ditolak karena validasi. |
| `security.sensitive_key_rejected` | Metadata analytics mengandung forbidden sensitive keys dan ditolak. |
| `security.suspicious_pattern` | Pola aman terdeteksi, misalnya banyak gagal login dalam bucket waktu. |
| `security.export_requested` | Export data sensitif/analytics diminta. |
| `security.internal_error_class` | Kelas error internal terhitung tanpa detail mentah. |

Allowed metadata:

- `route_group`
- `module`
- `role`
- `permission_code`
- `result`
- `status_code_class`
- `rate_limit_bucket`
- `validation_rule`
- `sensitive_key_category`
- `error_category`
- `ip_scope`
- `user_agent_family`
- `duration_bucket`

Forbidden:

- raw IP, raw UA, full URL query, stack trace, SQL, request body, response body, password, token, cookie, authorization, secret/API key, NIK/NIP/NISN penuh, credential PUSAKA, device fingerprint, dan payload yang sedang ditolak.

## Aturan Perubahan Katalog

- Event baru harus ditambahkan ke katalog ini sebelum implementasi.
- Setiap event baru harus menyebut allowed metadata dan forbidden fields.
- Review perubahan harus memastikan tidak ada data sensitif penuh masuk analytics.
- Test docs guard harus diperbarui bila kontrak global berubah.
- Implementasi Core API masa depan wajib punya test untuk:
  - menerima event allowlisted.
  - menolak event di luar allowlist.
  - menolak forbidden sensitive keys.
  - menjaga separation from audit_logs.
