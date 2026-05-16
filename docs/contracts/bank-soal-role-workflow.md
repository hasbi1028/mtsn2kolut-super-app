# Bank Soal Role Workflow Contract

Dokumen ini menjadi kontrak operasional untuk workflow Bank Soal setelah Sprint 5. Tujuannya memastikan penulisan, review, approval, pemakaian paket, audit, dan revisi soal berjalan aman untuk kegiatan resmi.

## Role matrix

- **Admin**
  - Melihat dan mengelola semua soal.
  - Boleh override scope reviewer/approver jika diperlukan.
  - Boleh melihat kunci jawaban dan rubric semua soal.
  - Wajib memakai catatan audit yang jelas saat override.

- **Pembuat soal / guru**
  - Membuat dan mengedit draft/revisi miliknya sendiri.
  - Submit soal untuk review.
  - Melihat soal miliknya dan hasil keputusan reviewer/approver.
  - Tidak boleh melihat draft/submitted milik guru lain kecuali memiliki scope review/approval.

- **Reviewer**
  - Melihat soal dalam scope mapel/tingkat/event yang ditugaskan.
  - Menandai layak review (`mark_reviewed`) atau meminta revisi (`request_revision`).
  - Tidak boleh review soal sendiri kecuali admin override.

- **Approver / penerbit**
  - Melihat soal dalam scope approval.
  - Menyetujui (`approve`) dan/atau menerbitkan (`publish`) soal yang sudah layak.
  - Tidak boleh approve/publish soal sendiri kecuali admin override.

- **Operator paket CBT**
  - Hanya memakai soal yang aman untuk paket resmi: workflow `approved`/`published` atau status legacy `published`.
  - Tidak mendapat akses jawaban/rubric kecuali memiliki izin review/approval/admin yang sesuai.

## Workflow transition matrix

- `draft` → `submitted`
  - Action: `submit_for_review`.
  - Actor: author/admin.
  - Catatan: soal masuk antrean reviewer.

- `submitted` / legacy `review` → `revision_needed`
  - Action: `request_revision`.
  - Actor: reviewer scope/admin.
  - Catatan: wajib isi alasan revisi bila dipakai resmi.

- `submitted` / legacy `review` → `reviewed`
  - Action: `mark_reviewed`.
  - Actor: reviewer scope/admin.
  - Catatan: soal siap approval.

- `reviewed` → `approved`
  - Action: `approve`.
  - Actor: approver scope/admin.
  - Catatan: soal aman untuk disiapkan ke paket.

- `approved` → `published`
  - Action: `publish`.
  - Actor: approver/publisher/admin.
  - Catatan: soal resmi terbit.

- Status aktif → `archived`
  - Action: `archive`.
  - Actor: admin/role yang berwenang.
  - Catatan: tidak boleh dipakai untuk paket baru.

## Visibility matrix

- **Draft/revisi sendiri**: author dan admin.
- **Submitted/review**: reviewer scope, approver scope bila relevan, author, admin.
- **Reviewed**: approver scope, reviewer scope terkait, author, admin.
- **Approved/published**: operator paket, reviewer/approver scope, author, admin.
- **Archived**: admin dan pengguna berizin baca penuh.

Kunci jawaban dan rubric harus tetap diredaksi untuk pengguna yang tidak punya hak review/approval/admin/author terkait.

## Package usage rule

- Paket resmi hanya boleh mengambil soal yang:
  - workflow `approved` atau `published`, atau
  - status publikasi legacy `published`.
- Paket tidak boleh mengambil soal `draft`, `submitted`, `review`, `revision_needed`, `rejected`, atau `archived`.
- Paket/sesi historis tetap mengacu ke `question_id` lama. Perubahan pada soal yang sudah published/dipakai tidak boleh in-place edit.

## Version-safe edit policy

- Soal yang sudah published atau sudah dipakai paket/sesi bersifat read-only.
- Edit lanjutan harus membuat revisi/versi baru.
- Version history ditampilkan di detail soal.
- Paket/sesi lama tetap menunjuk ke versi lama agar hasil ujian dan audit tidak berubah.

## Audit trail

Setiap perubahan workflow dicatat sebagai workflow event, minimal berisi:

- `question_id`
- actor user/username/display name bila ada
- `from_status`
- `to_status`
- `action`
- note/catatan
- metadata perubahan reviewer/approver/status publikasi bila ada
- `created_at`

UI detail soal menampilkan timeline workflow: action, transisi status, actor, waktu, note, dan metadata penting.

## Emergency admin override SOP

1. Pastikan alasan override terkait operasional resmi, bukan bypass kebiasaan.
2. Cek status soal, pemilik, scope mapel/tingkat/event, dan apakah sudah dipakai paket/sesi.
3. Jika soal sudah dipakai/published, jangan edit langsung; buat revisi baru.
4. Isi catatan workflow/audit yang menjelaskan alasan override.
5. Setelah override, lakukan smoke check daftar soal, detail soal, dan paket terkait.
6. Laporkan perubahan tanpa credential/secrets.
