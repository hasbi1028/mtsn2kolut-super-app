# User Display Name Policy

Goal: UI seluruh aplikasi tidak menampilkan UUID/internal ID atau username teknis bila tersedia nama pengguna/profil yang lebih manusiawi.

## Source of truth

Untuk tampilan manusia, gunakan urutan fallback:

1. `display_name` dari API bila tersedia.
2. Nama profil terkait:
   - `employees.nama`
   - `students.nama`
   - `parents.nama`
3. `users.display_name`
4. `username`
5. ID internal hanya untuk debug/admin teknis dan harus disembunyikan dari tampilan utama.

## API contract

Endpoint yang mengembalikan actor/user harus mengekspos pasangan field:

- `*_username`: identifier teknis/stabil untuk filter, request, audit, dan permission.
- `*_display_name`: label manusia untuk UI.

Contoh:

```json
{
  "author_username": "guru.mtk",
  "author_display_name": "Nama Guru Matematika"
}
```

UI wajib menampilkan `*_display_name`, bukan `*_username`, kecuali layar admin/debug memang membutuhkan username.

## Rollout priority

1. CBT/Bank Soal/Asesmen:
   - pembuat soal
   - reviewer
   - approver
   - panitia/pengawas/session users
2. Settings/Audit Logs:
   - actor user
   - reviewer user
   - change request user
3. TU/Governance/Arsip:
   - created_by / updated_by / assigned_to
4. Analytics/Notification/Inventory:
   - user_id hanya backend/audit, UI pakai nama.

## Implemented

- `GET /api/bank-soal/questions` now returns display names for list/scoped question rows:
  - `author_display_name`
  - `reviewer_display_name`
  - `approver_display_name`
- Paket Builder `/asesmen/paket/[id]` displays author filter and badges with `author_display_name` while retaining `author_username` internally for filtering.
